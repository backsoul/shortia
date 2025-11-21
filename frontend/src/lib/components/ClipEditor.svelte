<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import type {
    Video,
    SuggestedClip,
    Transcript,
    SubtitleConfig,
  } from "$lib/types";
  import VideoPlayer from "./VideoPlayer.svelte";
  import SubtitleCanvas from "./SubtitleCanvas.svelte";
  import SubtitleSettings from "./SubtitleSettings.svelte";
  import ExportProgress from "./ExportProgress.svelte";
  import VideoCanvasRenderer from "./VideoCanvasRenderer.svelte";
  import TemplateGallery from "./TemplateGallery.svelte";
  import { subtitleTemplates } from "$lib/data/templates";
  import {
    exportClipWithMediaRecorder,
    convertWebMToMP4,
    convertWebMToMP4OnBackend,
    exportClipWithFFmpeg,
  } from "$lib/services/ffmpeg";

  // Props
  let {
    video,
    transcript,
    suggestedClips,
  }: {
    video: Video;
    transcript: Transcript | null;
    suggestedClips: SuggestedClip[];
  } = $props();

  // Estado
  let selectedClip = $state<SuggestedClip | null>(null);
  let customStartTime = $state(0);
  let customEndTime = $state(30);
  let clipTitle = $state("");
  let subtitles = $state<SubtitleConfig[]>([]);
  let creating = $state(false);
  let isPlaying = $state(false);
  let currentTime = $state(0);
  let currentSubtitle = $state<SubtitleConfig | null>(null);
  let transcriptContainerElement = $state<HTMLDivElement>();
  let editingSegmentId = $state<number | null>(null);
  let editingText = $state("");

  // YouTube Configuration
  let youtubeTitle = $state("");
  let youtubeDescription = $state("");
  let youtubeTags = $state<string[]>([]);
  let generatingYoutubeSuggestions = $state(false);

  // Sincronización
  let syncOffset = $state(0); // Offset en segundos para ajustar sincronización

  // UI State
  let showClips = $state(true); // Controlar visibilidad de clips sugeridos
  let showTemplateGallery = $state(false); // Controlar visibilidad de galería de plantillas

  // Segmentos filtrados para el clip actual
  let clipTranscriptSegments = $derived.by(() => {
    if (!transcript || !selectedClip || !video) return [];

    // Calcular tiempos ajustados del clip
    const videoDuration = video.duration || 0;
    const adjustedStart = Math.max(0, selectedClip.start_time);
    const adjustedEnd = Math.min(selectedClip.end_time, videoDuration);

    console.log("🎯 [Filtro] Filtrando transcripción:", {
      clipTitle: selectedClip.title,
      clipStart: selectedClip.start_time,
      clipEnd: selectedClip.end_time,
      videoDuration,
      adjustedStart,
      adjustedEnd,
      totalSegments: transcript.segments.length,
    });

    // Filtrado estricto: solo segmentos que se solapan directamente con el clip
    const filtered = transcript.segments.filter((seg) => {
      // Solo incluir segmentos que estén dentro del rango del clip o se solapen
      const segmentStartInClip =
        seg.start >= adjustedStart && seg.start <= adjustedEnd;
      const segmentEndInClip =
        seg.end >= adjustedStart && seg.end <= adjustedEnd;
      const segmentOverlapsClip =
        seg.start < adjustedEnd && seg.end > adjustedStart;

      const matches =
        segmentStartInClip || segmentEndInClip || segmentOverlapsClip;

      console.log("🔍 Evaluando segmento:", {
        segmentStart: seg.start,
        segmentEnd: seg.end,
        clipRange: `${adjustedStart}-${adjustedEnd}`,
        segmentText: seg.text.substring(0, 30) + "...",
        segmentStartInClip,
        segmentEndInClip,
        segmentOverlapsClip,
        incluido: matches,
      });

      return matches;
    });

    console.log("🎯 [Filtro] Resultado final:", {
      segmentosFiltrados: filtered.length,
      primerSegmento: filtered[0]?.text?.substring(0, 50) + "...",
      ultimoSegmento:
        filtered[filtered.length - 1]?.text?.substring(0, 50) + "...",
    });

    return filtered;
  });

  // Subtitle Configuration
  let subtitleFontSize = $state(20);
  let subtitleFontFamily = $state("Inter");
  let subtitleFontWeight = $state(600);
  let wordsPerLine = $state(4);
  let subtitleTransition = $state<
    "none" | "pop" | "fade" | "slide" | "bounce" | "elastic" | "rotate"
  >("pop");
  let subtitleBgHexColor = $state("#000000");
  let subtitleBgOpacity = $state(0.8);
  let subtitleTextColor = $state("#FFFFFF");
  let subtitleActiveTextColor = $state("#22c55e");
  let subtitleBorderRadius = $state(8);
  let subtitleShadowBlur = $state(12);
  let subtitlePosition = $state("bottom");
  let isExporting = $state(false);
  let exportProgress = $state(0);
  let exportCancelled = $state(false); // Variable de estado para cancelación
  let exportFormat = $state<"webm" | "mp4">("mp4"); // Formato de exportación
  let exportStage = $state<
    | "preparing"
    | "loading-ffmpeg"
    | "downloading-video"
    | "extracting"
    | "rendering"
    | "encoding"
    | "uploading"
    | "converting"
    | "downloading"
    | "finalizing"
    | "complete"
    | "error"
    | "safari-fallback"
    | "capturing-subtitles"
    | "processing"
  >("preparing");

  let videoPlayerRef = $state<any>(null);
  let canvasRenderer = $state<any>(null); // Referencia al VideoCanvasRenderer para exportación
  let useCanvasRenderer = $state(true); // Flag para usar nuevo renderizador (DEBE ser true para exportación)
  let isRendererReady = $state(false); // Flag para verificar si el renderizador está completamente inicializado

  // Verificar si el renderizador está listo
  $effect(() => {
    isRendererReady = !!(
      canvasRenderer &&
      typeof canvasRenderer.getCanvas === "function" &&
      typeof canvasRenderer.getVideoElement === "function"
    );
    console.log("🎬 Estado del renderizador:", {
      isRendererReady,
      canvasRenderer,
    });
  });

  // Helper para obtener el player activo
  function getActivePlayer() {
    return useCanvasRenderer ? canvasRenderer : videoPlayerRef;
  }

  // Toggle play/pause con la barra espaciadora (solo en vista principal)
  onMount(() => {
    const handleGlobalKeydown = (event: KeyboardEvent) => {
      if (event.code !== "Space" && event.key !== " ") {
        return;
      }

      const target = event.target as HTMLElement | null;
      if (
        target &&
        (target.tagName === "INPUT" ||
          target.tagName === "TEXTAREA" ||
          target.tagName === "SELECT" ||
          target.tagName === "BUTTON" ||
          target.isContentEditable)
      ) {
        return;
      }

      if (isExporting) {
        return;
      }

      const activePlayer = getActivePlayer();
      if (activePlayer && typeof activePlayer.togglePlayPause === "function") {
        event.preventDefault();
        activePlayer.togglePlayPause();
      }
    };

    window.addEventListener("keydown", handleGlobalKeydown, { passive: false });

    return () => {
      window.removeEventListener("keydown", handleGlobalKeydown);
    };
  });

  // Seleccionar el primer clip automáticamente cuando se cargan
  $effect(() => {
    if (suggestedClips.length > 0 && !selectedClip) {
      selectClip(suggestedClips[0]);
    }
  });

  // Actualizar subtítulo actual basado en tiempo
  $effect(() => {
    if (!subtitles.length) {
      console.log("⏳ [ClipEditor] No hay subtítulos generados");
      currentSubtitle = null;
      return;
    }

    // Asegurar que currentTime esté dentro del rango del clip
    if (currentTime < customStartTime || currentTime > customEndTime) {
      currentSubtitle = null;
      return;
    }

    const relativeTime = currentTime - customStartTime + syncOffset;

    // Buscar subtítulo activo con una pequeña tolerancia
    const tolerance = 0.1; // 100ms de tolerancia
    const activeSubtitle = subtitles.find(
      (sub) =>
        relativeTime >= sub.start_time - tolerance &&
        relativeTime <= sub.end_time + tolerance
    );

    console.log("🕐 [ClipEditor] Actualizando subtítulo:", {
      currentTime,
      customStartTime,
      relativeTime,
      totalSubtitles: subtitles.length,
      activeSubtitle: activeSubtitle?.text || "ninguno",
      subtitleTimes: activeSubtitle
        ? `${activeSubtitle.start_time.toFixed(2)}-${activeSubtitle.end_time.toFixed(2)}`
        : "N/A",
      segmentId: activeSubtitle?.segmentId,
      originalSegmentTime:
        activeSubtitle?.segmentId !== undefined
          ? `${activeSubtitle.originalStart?.toFixed(2)}-${activeSubtitle.originalEnd?.toFixed(2)}`
          : "N/A",
    });

    currentSubtitle = activeSubtitle || null;
  });

  // Auto-scroll del transcript cuando cambia el tiempo del video (MEJORADO)
  let lastScrolledSegmentIndex = $state<number>(-1);

  $effect(() => {
    if (!transcript || !transcriptContainerElement) return;

    // Encontrar el segmento actual basado en currentTime
    const currentSegment = transcript.segments.find(
      (seg) => currentTime >= seg.start && currentTime <= seg.end
    );

    if (!currentSegment) return;

    const segmentIndex = transcript.segments.indexOf(currentSegment);

    // Solo hacer scroll si cambiamos de segmento
    if (segmentIndex === lastScrolledSegmentIndex) return;
    lastScrolledSegmentIndex = segmentIndex;

    // Buscar el elemento del segmento en el DOM
    const segmentElement = transcriptContainerElement.querySelector(
      `[data-segment-index="${segmentIndex}"]`
    ) as HTMLElement;

    if (segmentElement && transcriptContainerElement) {
      // Scroll suave solo dentro del contenedor de transcript
      const containerRect = transcriptContainerElement.getBoundingClientRect();
      const elementRect = segmentElement.getBoundingClientRect();
      const relativeTop = elementRect.top - containerRect.top;
      const containerHeight = transcriptContainerElement.clientHeight;
      const elementHeight = segmentElement.clientHeight;

      // Centrar el elemento en el contenedor
      const targetScroll =
        transcriptContainerElement.scrollTop +
        relativeTop -
        containerHeight / 2 +
        elementHeight / 2;

      transcriptContainerElement.scrollTo({
        top: targetScroll,
        behavior: "smooth",
      });
    }
  });

  function selectClip(clip: SuggestedClip) {
    console.log("🎬 [ClipEditor] Seleccionando clip:", {
      title: clip.title,
      start: clip.start_time,
      end: clip.end_time,
      videoDuration: video?.duration,
    });

    selectedClip = clip;

    // Validar y ajustar tiempos del clip
    const videoDuration = video?.duration || 0;
    if (clip.end_time > videoDuration) {
      console.warn("⚠️ [ClipEditor] Clip fuera de rango, ajustando...");
      const clipDuration = Math.min(clip.end_time - clip.start_time, 60);
      customStartTime = Math.max(0, videoDuration - clipDuration);
      customEndTime = videoDuration;
    } else {
      customStartTime = clip.start_time;
      customEndTime = clip.end_time;
    }

    clipTitle = clip.title;

    // Reproducir desde el inicio del clip (segundo 0 del clip)
    currentTime = customStartTime;
    const player = getActivePlayer();
    if (player) {
      player.seekTo(customStartTime);
    }

    generateSubtitles();
    scrollToClipInTranscript(clip.start_time);

    // Generar sugerencias de YouTube automáticamente
    youtubeTitle = "";
    youtubeDescription = "";
    generateYoutubeSuggestions();
  }

  function generateSubtitles() {
    if (!transcript || !selectedClip) {
      console.log("⏳ [ClipEditor] No se puede generar subtítulos:", {
        hasTranscript: !!transcript,
        hasSelectedClip: !!selectedClip,
      });
      return;
    }

    console.log("📝 [ClipEditor] Generando subtítulos...", {
      customStartTime,
      customEndTime,
      totalSegments: transcript.segments.length,
      wordsPerLine,
      fontSize: subtitleFontSize,
      fontFamily: subtitleFontFamily,
      transition: subtitleTransition,
    });

    // Filtrar segmentos que se solapan con el rango del clip
    // Usar una lógica más permisiva para incluir segmentos cercanos
    const clipSegments = transcript.segments.filter((seg) => {
      // Básico: segmentos que se solapan directamente
      const overlaps = seg.start < customEndTime && seg.end > customStartTime;

      // Alternativo: buscar segmentos dentro de un margen más amplio
      const margin = 5; // 5 segundos de margen
      const nearStart =
        seg.start >= customStartTime - margin &&
        seg.start <= customEndTime + margin;
      const nearEnd =
        seg.end >= customStartTime - margin &&
        seg.end <= customEndTime + margin;

      return overlaps || nearStart || nearEnd;
    });

    // Si aún no hay segmentos, buscar los más cercanos
    if (clipSegments.length === 0) {
      console.log("🔍 [ClipEditor] Buscando segmentos más cercanos...");
      const allSegmentsWithDistance = transcript.segments.map((seg) => ({
        ...seg,
        distance: Math.min(
          Math.abs(seg.start - customStartTime),
          Math.abs(seg.end - customEndTime),
          Math.abs(
            (seg.start + seg.end) / 2 - (customStartTime + customEndTime) / 2
          )
        ),
      }));

      // Tomar los 3 segmentos más cercanos
      allSegmentsWithDistance.sort((a, b) => a.distance - b.distance);
      clipSegments.push(...allSegmentsWithDistance.slice(0, 3));
    }

    console.log("📊 [ClipEditor] Segmentos en el rango:", clipSegments.length);

    if (clipSegments.length === 0) {
      console.warn(
        "⚠️ [ClipEditor] No se encontraron segmentos para el clip:",
        {
          startTime: customStartTime,
          endTime: customEndTime,
          totalSegments: transcript.segments.length,
          firstSegment: transcript.segments[0],
          lastSegment: transcript.segments[transcript.segments.length - 1],
          videoDuration: video?.duration,
          allSegmentTimes: transcript.segments.map((s) => ({
            start: s.start,
            end: s.end,
            text: s.text.substring(0, 50),
          })),
        }
      );

      // Si el clip está fuera del rango del video, ajustarlo
      if (customEndTime > (video?.duration || 0)) {
        console.log(
          "🔧 [ClipEditor] Ajustando tiempos del clip al rango del video"
        );
        const videoDuration = video?.duration || 0;
        const clipDuration = Math.min(60, videoDuration); // Máximo 60 segundos
        customStartTime = Math.max(0, videoDuration - clipDuration);
        customEndTime = videoDuration;

        // Regenerar con los nuevos tiempos
        setTimeout(() => generateSubtitles(), 100);
        return;
      }
    }

    const wordSubtitles: SubtitleConfig[] = [];

    clipSegments.forEach((seg) => {
      // Recortar el segmento al rango del clip
      const clippedStart = Math.max(seg.start, customStartTime);
      const clippedEnd = Math.min(seg.end, customEndTime);

      // Si el segmento recortado no tiene duración, saltarlo
      if (clippedStart >= clippedEnd) return;

      // Dividir el texto en grupos de palabras según wordsPerLine
      const words = seg.text.trim().split(/\s+/);
      const segmentDuration = clippedEnd - clippedStart;

      if (wordsPerLine === 1) {
        // Modo palabra por palabra para efecto karaoke
        const timePerWord = segmentDuration / words.length;

        words.forEach((word, wordIndex) => {
          const wordStartTime = clippedStart + wordIndex * timePerWord;
          const wordEndTime = wordStartTime + timePerWord;

          const subtitle = {
            text: word,
            start_time: wordStartTime - customStartTime,
            end_time: wordEndTime - customStartTime,
            font_family: subtitleFontFamily,
            font_size: subtitleFontSize,
            font_weight: subtitleFontWeight,
            color: subtitleTextColor,
            bg_color: subtitleBgHexColor,
            bg_opacity: subtitleBgOpacity,
            position: subtitlePosition as "top" | "center" | "bottom",
            bold: subtitleFontWeight >= 600,
            italic: false,
            border_radius: subtitleBorderRadius,
            shadow_blur: subtitleShadowBlur,
            transition: subtitleTransition,
            active_text_color: subtitleActiveTextColor,
            segmentId: transcript.segments.indexOf(seg),
            originalStart: seg.start,
            originalEnd: seg.end,
            wordIndex: wordIndex,
            totalWords: words.length,
          };

          if (
            subtitle.end_time > subtitle.start_time &&
            subtitle.start_time >= 0
          ) {
            wordSubtitles.push(subtitle);
          }
        });
      } else {
        // Modo grupos de palabras
        const groupedWords: string[][] = [];
        for (let i = 0; i < words.length; i += wordsPerLine) {
          groupedWords.push(words.slice(i, i + wordsPerLine));
        }

        const timePerGroup = segmentDuration / groupedWords.length;

        groupedWords.forEach((wordGroup, groupIndex) => {
          const groupStartTime = clippedStart + groupIndex * timePerGroup;
          const groupEndTime = Math.min(
            groupStartTime + timePerGroup,
            clippedEnd
          );

          const subtitle = {
            text: wordGroup.join(" "),
            start_time: groupStartTime - customStartTime,
            end_time: groupEndTime - customStartTime,
            font_family: subtitleFontFamily,
            font_size: subtitleFontSize,
            font_weight: subtitleFontWeight,
            color: subtitleTextColor,
            bg_color: subtitleBgHexColor,
            bg_opacity: subtitleBgOpacity,
            position: subtitlePosition as "top" | "center" | "bottom",
            bold: subtitleFontWeight >= 600,
            italic: false,
            border_radius: subtitleBorderRadius,
            shadow_blur: subtitleShadowBlur,
            transition: subtitleTransition,
            active_text_color: subtitleActiveTextColor,
            segmentId: transcript.segments.indexOf(seg),
            originalStart: seg.start,
            originalEnd: seg.end,
            groupIndex: groupIndex,
            totalGroups: groupedWords.length,
          };

          if (
            subtitle.end_time > subtitle.start_time &&
            subtitle.start_time >= 0
          ) {
            wordSubtitles.push(subtitle);
          }
        });
      }
    });

    subtitles = wordSubtitles;
    console.log("✅ [ClipEditor] Subtítulos generados:", {
      total: subtitles.length,
      firstSubtitle: subtitles[0]?.text,
      lastSubtitle: subtitles[subtitles.length - 1]?.text,
      sampleSubtitle: subtitles[0],
      timeRange:
        subtitles.length > 0
          ? `${subtitles[0]?.start_time.toFixed(2)}s - ${subtitles[subtitles.length - 1]?.end_time.toFixed(2)}s`
          : "N/A",
      segmentsUsed: clipSegments.length,
      clipRange: `${customStartTime}s - ${customEndTime}s`,
    });
  }

  function scrollToClipInTranscript(startTime: number) {
    if (!transcript || !transcriptContainerElement) return;

    // Find the first segment that starts at or after the clip start time
    const segmentIndex = transcript.segments.findIndex(
      (seg) => seg.start >= startTime
    );

    if (segmentIndex === -1) return;

    // Wait for DOM to update, then scroll
    setTimeout(() => {
      if (!transcriptContainerElement) return;

      const segmentElement = transcriptContainerElement.querySelector(
        `[data-segment-index="${segmentIndex}"]`
      );

      if (segmentElement) {
        segmentElement.scrollIntoView({
          behavior: "smooth",
          block: "center",
        });
        console.log("📜 Auto-scrolled to segment:", segmentIndex);
      }
    }, 100);
  }

  function handleSubtitlePositionChange(position: "top" | "center" | "bottom") {
    if (currentSubtitle) {
      currentSubtitle.position = position;
      const currentStartTime = currentSubtitle.start_time;
      subtitles.forEach((sub) => {
        if (sub.start_time === currentStartTime) {
          sub.position = position;
        }
      });
    }
  }

  function handleSettingsChange() {
    generateSubtitles();
  }

  async function exportClip() {
    if (!selectedClip) {
      console.error("❌ No hay clip seleccionado");
      alert("Por favor selecciona un clip primero");
      return;
    }

    if (
      !canvasRenderer ||
      typeof canvasRenderer.getCanvas !== "function" ||
      typeof canvasRenderer.getVideoElement !== "function"
    ) {
      console.error(
        "❌ Canvas renderer no está disponible o no está completamente inicializado",
        {
          canvasRenderer,
          hasGetCanvas:
            canvasRenderer && typeof canvasRenderer.getCanvas === "function",
          hasGetVideoElement:
            canvasRenderer &&
            typeof canvasRenderer.getVideoElement === "function",
        }
      );
      alert(
        "El renderizador de video no está listo. Por favor espera unos segundos o recarga la página."
      );
      return;
    }

    console.log(
      `🎬 Iniciando exportación en formato ${exportFormat.toUpperCase()}...`
    );
    console.log("📊 Estado:", {
      customStartTime,
      customEndTime,
      subtitles: subtitles.length,
      clipTitle,
      exportFormat,
    });

    isExporting = true;
    exportCancelled = false; // Resetear flag de cancelación
    exportProgress = 0;
    exportStage = "preparing";

    const activePlayer = getActivePlayer();
    if (activePlayer && typeof activePlayer.pause === "function") {
      activePlayer.pause();
    }

    try {
      // Obtener canvas y video element del renderer
      const canvas = canvasRenderer.getCanvas();
      const videoElement = canvasRenderer.getVideoElement();

      if (!canvas || !videoElement) {
        throw new Error("No se pudo obtener canvas o video element");
      }

      console.log("🎨 Canvas y video obtenidos:", {
        canvasWidth: canvas.width,
        canvasHeight: canvas.height,
        videoSrc: videoElement.src,
      });

      exportStage = "extracting";
      exportProgress = 5;

      // Intentar exportar usando MediaRecorder (Frontend-Only)
      let webmBlob: Blob;

      try {
        console.log("🎥 Iniciando grabación con MediaRecorder...");
        webmBlob = await exportClipWithMediaRecorder(
          canvas,
          videoElement,
          customStartTime,
          customEndTime,
          (progress, stage) => {
            if (exportFormat === "webm") {
              // Si solo queremos WebM, usar todo el rango de progreso
              exportProgress = Math.round(progress);
            } else {
              // Si queremos MP4, dejar espacio para conversión (0-60%)
              exportProgress = Math.round(progress * 0.6);
            }
            exportStage = stage as any;
            console.log(`📊 Progreso grabación: ${exportProgress}% - ${stage}`);
          },
          () => exportCancelled // Función para verificar cancelación
        );
      } catch (mediaRecorderError) {
        console.log(
          "⚠️ MediaRecorder falló, usando fallback con FFmpeg:",
          mediaRecorderError
        );

        // Fallback para Safari y otros navegadores sin soporte para captureStream
        exportStage = "safari-fallback";
        exportProgress = 5;

        console.log(
          "🍎 Detectado navegador sin soporte MediaRecorder, usando FFmpeg..."
        );

        webmBlob = await exportClipWithFFmpeg(
          canvas,
          videoElement,
          customStartTime,
          customEndTime,
          (progress, stage) => {
            if (exportFormat === "webm") {
              exportProgress = Math.round(progress);
            } else {
              exportProgress = Math.round(progress * 0.6);
            }
            exportStage = stage as any;
            console.log(`📊 Progreso FFmpeg: ${exportProgress}% - ${stage}`);
          },
          () => exportCancelled
        );
      }

      // Verificar si fue cancelado
      if (exportCancelled) {
        console.log("❌ Exportación cancelada por usuario");
        return;
      }

      console.log("✅ WebM generado:", {
        size: webmBlob.size,
        type: webmBlob.type,
      });

      let finalBlob = webmBlob;
      let extension = "webm";

      // Si se eligió MP4, convertir usando el backend
      if (exportFormat === "mp4") {
        exportStage = "converting";
        exportProgress = 60;
        console.log("� Convirtiendo a MP4 usando backend (FFmpeg nativo)...");

        finalBlob = await convertWebMToMP4OnBackend(
          webmBlob,
          (progress, stage) => {
            // Mapear progreso de conversión (60-95%)
            exportProgress = 60 + Math.round(progress * 0.35);
            exportStage = stage as any;
            console.log(
              `📊 Progreso conversión: ${exportProgress}% - ${stage}`
            );
          }
        );

        // Verificar si fue cancelado
        if (exportCancelled) {
          console.log("❌ Exportación cancelada por usuario");
          return;
        }

        extension = "mp4";
        console.log("✅ MP4 convertido:", {
          size: finalBlob.size,
          type: finalBlob.type,
        });
      }

      exportStage = "finalizing";
      exportProgress = 95;

      // Descargar automáticamente
      const url = URL.createObjectURL(finalBlob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${clipTitle || "clip"}_${Date.now()}.${extension}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);

      exportProgress = 100;
      exportStage = "complete";
      console.log("✅ Descarga completada");

      setTimeout(() => {
        isExporting = false;
        exportProgress = 0;
        exportStage = "preparing";
      }, 2000);
    } catch (error) {
      console.error("❌ Export failed:", error);
      if (error instanceof Error && error.message === "Exportación cancelada") {
        console.log("ℹ️ Exportación cancelada correctamente");
        exportStage = "error";
      } else {
        alert(`Error al exportar: ${error}`);
        exportStage = "error";
      }
      isExporting = false;
      exportProgress = 0;
      exportCancelled = false; // Resetear flag de cancelación para próximas exportaciones
    }
  }

  function cancelExport() {
    console.log("🛑 Solicitando cancelación de exportación...");
    exportCancelled = true; // Marcar como cancelado
    isExporting = false;
    exportProgress = 0;
    exportStage = "preparing";
  }

  function formatTime(seconds: number): string {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }

  function getClipColorClasses(score: number, isSelected: boolean): string {
    if (isSelected) {
      return "ring-2 ring-blue-400 bg-blue-600/20 shadow-lg shadow-blue-500/30";
    }

    if (score >= 90) {
      // VIRAL - Dorado brillante con efectos premium
      return "bg-gradient-to-br from-amber-400 via-yellow-500 to-orange-500 hover:from-amber-300 hover:via-yellow-400 hover:to-orange-400 border border-amber-400 shadow-lg shadow-amber-500/40 hover:shadow-2xl hover:shadow-amber-500/60";
    } else if (score >= 80) {
      // EXCELENTE - Verde esmeralda premium
      return "bg-gradient-to-br from-emerald-500 via-green-500 to-teal-600 hover:from-emerald-400 hover:via-green-400 hover:to-teal-500 border border-emerald-400 shadow-lg shadow-emerald-500/40 hover:shadow-2xl hover:shadow-emerald-500/60";
    } else if (score >= 70) {
      // MUY BUENO - Azul cielo brillante
      return "bg-gradient-to-br from-sky-500 via-blue-500 to-indigo-600 hover:from-sky-400 hover:via-blue-400 hover:to-indigo-500 border border-sky-400 shadow-lg shadow-blue-500/40 hover:shadow-2xl hover:shadow-blue-500/60";
    } else if (score >= 60) {
      // BUENO - Púrpura elegante
      return "bg-gradient-to-br from-violet-500 via-purple-500 to-fuchsia-600 hover:from-violet-400 hover:via-purple-400 hover:to-fuchsia-500 border border-purple-400 shadow-lg shadow-purple-500/40 hover:shadow-2xl hover:shadow-purple-500/60";
    } else if (score >= 50) {
      // PROMEDIO - Naranja cálido
      return "bg-gradient-to-br from-orange-500 via-amber-600 to-red-600 hover:from-orange-400 hover:via-amber-500 hover:to-red-500 border border-orange-400 shadow-lg shadow-orange-500/40 hover:shadow-2xl hover:shadow-orange-500/60";
    } else {
      // BÁSICO - Gris neutro
      return "bg-gradient-to-br from-gray-600 via-gray-700 to-slate-700 hover:from-gray-500 hover:via-gray-600 hover:to-slate-600 border border-gray-500 shadow-lg shadow-gray-500/30 hover:shadow-2xl hover:shadow-gray-500/50";
    }
  }

  function getScoreLabel(score: number): string {
    if (score >= 90) return "VIRAL";
    if (score >= 80) return "Premium";
    if (score >= 70) return "Destacado";
    if (score >= 60) return "Bueno";
    if (score >= 50) return "Promedio";
    return "Básico";
  }

  function getScoreEmoji(score: number): string {
    if (score >= 90) return "🔥";
    if (score >= 80) return "💎";
    if (score >= 70) return "⭐";
    if (score >= 60) return "👍";
    if (score >= 50) return "📈";
    return "📱";
  }

  function getScoreIcon(score: number): string {
    if (score >= 90) return "M13 10V3L4 14h7v7l9-11h-7z"; // Lightning/Viral
    if (score >= 80)
      return "M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z"; // Fire
    if (score >= 70)
      return "M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"; // Star
    if (score >= 60)
      return "M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.60L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5"; // Thumbs up
    if (score >= 50) return "M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"; // Trending up
    return "M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"; // Question mark
  }

  function getScoreFilterColor(score: number): string {
    if (score >= 90) return "rgba(245, 158, 11, 0.3)"; // Amber/Gold for Viral
    if (score >= 80) return "rgba(16, 185, 129, 0.3)"; // Emerald for Premium
    if (score >= 70) return "rgba(59, 130, 246, 0.3)"; // Blue for Destacado
    if (score >= 60) return "rgba(139, 92, 246, 0.3)"; // Purple for Bueno
    if (score >= 50) return "rgba(249, 115, 22, 0.3)"; // Orange for Promedio
    return "rgba(107, 114, 128, 0.3)"; // Gray for Básico
  }

  async function generateYoutubeSuggestions() {
    if (!selectedClip) return;

    generatingYoutubeSuggestions = true;

    try {
      console.log("🔥 Generando SEO profesional con DeepSeek...");

      const response = await fetch(
        `http://localhost:8080/api/videos/${video.id}/generate-seo`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            video_id: video.id,
            clip_start_time: customStartTime,
            clip_end_time: customEndTime,
            clip_title: clipTitle,
          }),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to generate SEO: ${response.statusText}`);
      }

      const seoContent = await response.json();

      console.log("✅ SEO generado:", seoContent);

      // Aplicar el contenido SEO generado
      youtubeTitle = seoContent.title || "";
      youtubeDescription = seoContent.description || "";
      youtubeTags = seoContent.tags || [];
    } catch (error) {
      console.error("❌ Error generando SEO:", error);

      // Fallback a contenido básico en caso de error
      youtubeTitle = `${selectedClip.title}`;
      youtubeDescription = `${selectedClip.description}\n\n#shorts #viral #trending`;
      youtubeTags = [
        "shorts",
        "viral",
        "trending",
        "momento",
        "destacado",
        "clip",
      ];

      alert(
        "No se pudo generar SEO profesional. Verifica que la API de DeepSeek esté configurada."
      );
    } finally {
      generatingYoutubeSuggestions = false;
    }
  }

  function startEditingSegment(index: number, currentText: string) {
    editingSegmentId = index;
    editingText = currentText;
  }

  function saveSegmentEdit(index: number) {
    if (transcript && transcript.segments[index]) {
      // Actualizar el segmento en la transcripción
      transcript.segments[index].text = editingText;

      // Buscar y actualizar el subtítulo correspondiente usando segmentId
      const subtitle = subtitles.find((sub) => sub.segmentId === index);
      if (subtitle) {
        subtitle.text = editingText;
        console.log("✅ [ClipEditor] Subtítulo actualizado:", {
          segmentId: index,
          newText: editingText,
        });
      }

      // También actualizar el currentSubtitle si es el que está siendo mostrado
      if (currentSubtitle && currentSubtitle.segmentId === index) {
        currentSubtitle.text = editingText;
      }
    }
    editingSegmentId = null;
  }

  function cancelEdit() {
    editingSegmentId = null;
  }

  function findAndSelectClipForTime(time: number) {
    const matchingClip = suggestedClips.find(
      (clip) => time >= clip.start_time && time <= clip.end_time
    );

    if (matchingClip && matchingClip.id !== selectedClip?.id) {
      selectClip(matchingClip);
    } else {
      const player = getActivePlayer();
      if (player) {
        player.seekTo(time);
        if (!isPlaying) {
          player.play();
        }
      }
    }
  }

  function handleEndTimeChange() {
    if (customEndTime <= customStartTime) {
      customEndTime = customStartTime + 1;
    }
  }

  async function createClip() {
    if (!clipTitle.trim()) {
      alert("Por favor ingresa un título para el clip");
      return;
    }

    try {
      creating = true;

      const response = await fetch("/api/clips", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          video_id: video.id,
          title: clipTitle,
          start_time: customStartTime,
          end_time: customEndTime,
          subtitles,
        }),
      });

      if (!response.ok) throw new Error("Failed to create clip");

      const clip = await response.json();
      alert("¡Clip creado exitosamente! Procesando...");
    } catch (error) {
      alert("Error al crear el clip");
    } finally {
      creating = false;
    }
  }
</script>

<div class="fixed inset-0 bg-[#0a0a0a] flex flex-col overflow-hidden">
  <!-- Header con efecto glassmorphism - Ancho completo -->
  <div
    class="bg-gradient-to-r from-gray-900/80 via-gray-900/90 to-gray-900/80 backdrop-blur-xl text-white px-6 py-4 flex justify-between items-center border-b border-white/5 shadow-2xl shadow-purple-500/10 relative z-20 w-full"
  >
    <div class="w-full flex justify-center items-center">
      <!-- 1 -->
      <div
        class="w-full max-w-screen-2xl mx-auto flex justify-between items-center"
      >
        <!-- Línea de acento superior -->
        <div
          class="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-purple-500/50 to-transparent"
        ></div>

        <div class="flex items-center gap-4">
          <a
            href="/"
            class="flex items-center gap-2 text-gray-400 hover:text-white transition-all duration-300 group px-3 py-1.5 rounded-lg hover:bg-white/5"
          >
            <svg
              class="w-4 h-4 transform group-hover:-translate-x-1 transition-transform"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 19l-7-7 7-7"
              />
            </svg>
            Atrás
          </a>
          <div
            class="h-6 w-px bg-gradient-to-b from-transparent via-gray-600 to-transparent"
          ></div>
          <h1
            class="text-xl font-bold truncate max-w-2xl bg-clip-text text-transparent bg-gradient-to-r from-purple-400 via-pink-400 to-blue-400"
          >
            {video.title}
          </h1>
        </div>

        <div class="flex items-center gap-3">
          <!-- Selector de formato de exportación mejorado -->
          <div class="flex items-center gap-3 text-sm">
            <label
              for="export-format"
              class="text-gray-300 font-medium flex items-center gap-2"
            >
              <svg
                class="w-4 h-4 text-emerald-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 011-1h1m-1 1v1m0-1h1m0 0h1"
                />
              </svg>
              Formato:
            </label>
            <div class="relative">
              <select
                id="export-format"
                bind:value={exportFormat}
                disabled={isExporting}
                class="appearance-none bg-gray-800/50 backdrop-blur-sm text-white px-4 py-2.5 pr-10 rounded-lg border border-purple-500/20 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-700/50 hover:border-purple-500/40 transition-all duration-200 cursor-pointer"
                style="color-scheme: dark;"
              >
                <option value="webm">🚀 WebM (Rápido)</option>
                <option value="mp4">💎 MP4 (Calidad)</option>
              </select>
              <div
                class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none"
              >
                <svg
                  class="w-4 h-4 text-gray-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </div>
            </div>
          </div>

          <button
            onclick={exportClip}
            disabled={isExporting || !selectedClip || !isRendererReady}
            class="relative group bg-gradient-to-r from-emerald-600 via-green-600 to-teal-600 text-white font-semibold py-3 px-8 rounded-xl overflow-hidden shadow-lg hover:shadow-emerald-500/25 transform hover:scale-[1.02] transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 disabled:hover:shadow-lg"
          >
            <!-- Animated background overlay -->
            <div
              class="absolute inset-0 bg-gradient-to-r from-emerald-400 via-green-400 to-teal-400 opacity-0 group-hover:opacity-100 transition-opacity duration-300"
            ></div>

            <!-- Shimmer effect -->
            <div
              class="absolute inset-0 -top-2 -bottom-2 bg-gradient-to-r from-transparent via-white/20 to-transparent skew-x-12 -translate-x-full group-hover:translate-x-full transition-transform duration-700 ease-out"
            ></div>

            <!-- Particle effects -->
            <div
              class="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300"
            >
              <div
                class="absolute top-1/2 left-1/4 w-1 h-1 bg-white/60 rounded-full animate-pulse"
              ></div>
              <div
                class="absolute top-1/3 right-1/4 w-1 h-1 bg-white/60 rounded-full animate-pulse"
                style="animation-delay: 0.2s;"
              ></div>
              <div
                class="absolute bottom-1/3 left-1/3 w-1 h-1 bg-white/60 rounded-full animate-pulse"
                style="animation-delay: 0.4s;"
              ></div>
            </div>

            <!-- Content -->
            <span class="relative z-10 flex items-center gap-2">
              {#if isExporting}
                <svg
                  class="w-4 h-4 animate-spin"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    class="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    stroke-width="4"
                  ></circle>
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path>
                </svg>
                Exportando...
              {:else if !isRendererReady}
                <svg
                  class="w-4 h-4 animate-spin"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    class="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    stroke-width="4"
                  ></circle>
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path>
                </svg>
                Cargando renderizador...
              {:else}
                <svg
                  class="w-5 h-5 group-hover:scale-110 transition-transform duration-200"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                Exportar Clip
              {/if}
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Main Content Area - Ancho completo -->
  <div
    class="flex-1 flex flex-col overflow-hidden bg-gradient-to-br from-gray-900 via-[#0a0a0a] to-gray-900 relative w-full"
  >
    <!-- Background effects -->
    <div
      class="absolute inset-0 overflow-hidden pointer-events-none opacity-30"
    >
      <div
        class="absolute inset-0 bg-grid-pattern animate-grid-move opacity-20"
      ></div>
    </div>

    <!-- Content Section -->
    <!-- max width desktop -->
    <div
      class="flex-1 overflow-hidden relative z-10 w-full max-w-screen-2xl mx-auto"
    >
      {#if !selectedClip}
        <div class="flex items-center justify-center h-full">
          <div class="text-center text-gray-400">
            <svg
              class="mx-auto h-24 w-24 mb-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
              />
            </svg>
            <p class="text-xl mb-2">No hay clip seleccionado</p>
            <p class="text-sm">
              Selecciona un clip de los clips sugeridos para comenzar a editar
            </p>
          </div>
        </div>
      {:else}
        <div class="grid grid-cols-3 gap-6 h-full p-6 overflow-hidden">
          <!-- YouTube Section -->
          <div class="relative group flex flex-col gap-4">
            <div class="relative glass-effect p-6">
              <div class="flex items-center justify-between mb-5">
                <h3
                  class="text-lg font-semibold text-white flex items-center gap-2"
                >
                  <svg
                    class="w-6 h-6 text-red-500"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"
                    />
                  </svg>
                  <span>YouTube</span>
                </h3>
                <button
                  onclick={generateYoutubeSuggestions}
                  disabled={!selectedClip || generatingYoutubeSuggestions}
                  class="px-3 py-1.5 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 transition-colors"
                >
                  {#if generatingYoutubeSuggestions}
                    <svg
                      class="w-4 h-4 animate-spin"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                      />
                    </svg>
                    Generando...
                  {:else}
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                      />
                    </svg>
                    Generar
                  {/if}
                </button>
              </div>

              <div class="space-y-4">
                <!-- Título de YouTube -->
                <div>
                  <label
                    for="youtube-title"
                    class="block text-sm font-medium text-gray-300 mb-2"
                  >
                    Título sugerido (máx. 100 caracteres)
                  </label>
                  <div class="relative">
                    <input
                      id="youtube-title"
                      bind:value={youtubeTitle}
                      placeholder="El título se generará automáticamente..."
                      maxlength="100"
                      class="w-full p-3 pr-20 bg-gray-800 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-red-500 focus:border-transparent transition-colors"
                    />
                    <div
                      class="absolute right-3 top-1/2 transform -translate-y-1/2 flex items-center gap-2"
                    >
                      <span class="text-xs text-gray-400">
                        {youtubeTitle.length}/100
                      </span>
                      {#if youtubeTitle}
                        <button
                          onclick={() =>
                            navigator.clipboard.writeText(youtubeTitle)}
                          class="p-1 text-gray-400 hover:text-white transition-colors"
                          title="Copiar título"
                        >
                          <svg
                            class="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              stroke-width="2"
                              d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                            />
                          </svg>
                        </button>
                      {/if}
                    </div>
                  </div>
                </div>

                <!-- Descripción de YouTube -->
                <div>
                  <label
                    for="youtube-description"
                    class="block text-sm font-medium text-gray-300 mb-2"
                  >
                    Descripción sugerida (máx. 5000 caracteres)
                  </label>
                  <div class="relative">
                    <textarea
                      id="youtube-description"
                      bind:value={youtubeDescription}
                      placeholder="La descripción se generará automáticamente..."
                      maxlength="5000"
                      rows="3"
                      class="w-full p-3 pr-20 bg-gray-800 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-red-500 focus:border-transparent transition-colors resize-none"
                    ></textarea>
                    <div
                      class="absolute right-3 bottom-3 flex items-center gap-2"
                    >
                      <span class="text-xs text-gray-400">
                        {youtubeDescription.length}/5000
                      </span>
                      {#if youtubeDescription}
                        <button
                          onclick={() =>
                            navigator.clipboard.writeText(youtubeDescription)}
                          class="p-1 text-gray-400 hover:text-white transition-colors"
                          title="Copiar descripción"
                        >
                          <svg
                            class="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              stroke-width="2"
                              d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                            />
                          </svg>
                        </button>
                      {/if}
                    </div>
                  </div>
                </div>

                <!-- Tags -->
                {#if youtubeTitle || youtubeDescription}
                  <div
                    class="bg-red-950/30 border border-red-500/20 rounded-lg p-3"
                  >
                    <h4
                      class="text-sm font-medium text-red-300 mb-2 flex items-center gap-2"
                    >
                      <svg
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
                        />
                      </svg>
                      Tags
                    </h4>
                    <div class="relative">
                      <input
                        type="text"
                        value={youtubeTags.length > 0
                          ? youtubeTags.join(", ")
                          : "Generando tags..."}
                        disabled
                        class="w-full px-3 py-2 pr-10 bg-gray-800/50 border border-red-500/30 rounded-lg text-gray-300 text-xs font-mono"
                      />
                      <button
                        onclick={(e) => {
                          const tags = youtubeTags.join(", ");
                          navigator.clipboard.writeText(tags);
                          // Opcional: mostrar feedback visual
                          const btn = e?.currentTarget as HTMLButtonElement;
                          if (btn) {
                            const originalHTML = btn.innerHTML;
                            btn.innerHTML =
                              '<svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>';
                            setTimeout(() => {
                              btn.innerHTML = originalHTML;
                            }, 1500);
                          }
                        }}
                        class="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 hover:bg-red-500/20 rounded transition-colors"
                        title="Copiar tags"
                      >
                        <svg
                          class="w-4 h-4 text-red-300"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                          />
                        </svg>
                      </button>
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <!-- Transcripción Section -->
            <div class="glass-effect p-6">
              <h3
                class="text-lg font-semibold text-white mb-5 flex items-center gap-2"
              >
                <svg
                  class="w-6 h-6 text-purple-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <span>Transcripción</span>
                {#if selectedClip && clipTranscriptSegments.length > 0}
                  <span class="text-sm text-purple-300"
                    >({clipTranscriptSegments.length} segmentos)</span
                  >
                {/if}
              </h3>

              <!-- Editable Transcript Display -->
              {#if transcript && clipTranscriptSegments.length > 0}
                <div
                  bind:this={transcriptContainerElement}
                  class="bg-gray-900/50 rounded-xl p-4 max-h-96 overflow-y-auto ring-1 ring-purple-500/20 scrollbar-thin scrollbar-thumb-gray-600 scrollbar-track-gray-800"
                >
                  <div class="flex items-center justify-between mb-4">
                    <p class="text-xs text-purple-300">
                      Clic para reproducir • Doble clic para editar
                    </p>
                    <div class="text-xs text-purple-400">
                      {clipTranscriptSegments.length} segmentos
                    </div>
                  </div>
                  <div class="space-y-2 text-sm">
                    {#each clipTranscriptSegments as segment, clipIndex (`${segment.start}-${segment.end}`)}
                      {@const originalIndex = transcript.segments.findIndex(
                        (s) =>
                          s.start === segment.start && s.text === segment.text
                      )}
                      <div
                        data-segment-index={originalIndex}
                        class="group relative p-3 rounded-lg transition-all duration-300 ease-out {currentTime >=
                          segment.start && currentTime < segment.end
                          ? 'bg-gradient-to-r from-purple-900/40 to-purple-800/30 ring-2 ring-purple-500 shadow-lg shadow-purple-500/20'
                          : 'hover:bg-gray-700/50 hover:border-purple-500/30'} border border-gray-700/50"
                      >
                        {#if editingSegmentId === originalIndex}
                          <!-- Editing Mode -->
                          <div class="space-y-2">
                            <div class="flex items-start gap-2">
                              <span
                                class="text-xs text-purple-300 mt-2 font-medium"
                              >
                                {formatTime(segment.start)}
                              </span>
                              <textarea
                                bind:value={editingText}
                                class="flex-1 px-3 py-2 bg-gray-800 border border-purple-500 rounded text-white text-sm resize-none focus:outline-none focus:ring-2 focus:ring-purple-400"
                                rows="2"
                                onkeydown={(e) => {
                                  if (e.key === "Enter" && !e.shiftKey) {
                                    e.preventDefault();
                                    saveSegmentEdit(originalIndex);
                                  } else if (e.key === "Escape") {
                                    cancelEdit();
                                  }
                                }}
                              ></textarea>
                            </div>
                            <div class="flex justify-end gap-2">
                              <button
                                onclick={cancelEdit}
                                class="px-3 py-1 bg-gray-600 hover:bg-gray-500 rounded text-xs text-white transition-colors"
                              >
                                Cancelar
                              </button>
                              <button
                                onclick={() => saveSegmentEdit(originalIndex)}
                                class="px-3 py-1 bg-purple-600 hover:bg-purple-700 rounded text-xs text-white transition-colors"
                              >
                                Guardar
                              </button>
                            </div>
                          </div>
                        {:else}
                          <!-- Display Mode -->
                          <button
                            onclick={() => {
                              findAndSelectClipForTime(segment.start);
                            }}
                            ondblclick={() =>
                              startEditingSegment(originalIndex, segment.text)}
                            class="block w-full text-left"
                          >
                            <div class="flex items-start gap-3">
                              <span
                                class="text-xs font-medium min-w-[60px] text-purple-300"
                              >
                                {formatTime(segment.start)}
                              </span>
                              <span class="text-gray-200 leading-relaxed">
                                {segment.text}
                              </span>
                            </div>
                          </button>

                          <!-- Edit button on hover -->
                          <button
                            onclick={() =>
                              startEditingSegment(originalIndex, segment.text)}
                            class="absolute right-3 top-3 opacity-0 group-hover:opacity-100 p-1 bg-purple-600 hover:bg-purple-700 rounded text-xs text-white transition-all"
                            title="Editar segmento"
                          >
                            <svg
                              class="w-4 h-4"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                              />
                            </svg>
                          </button>
                        {/if}
                      </div>
                    {/each}
                  </div>
                </div>
              {:else if selectedClip}
                <div class="text-center py-8">
                  <svg
                    class="w-12 h-12 text-gray-600 mx-auto mb-3"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                  <p class="text-gray-400 text-sm">
                    No hay transcripción disponible para este clip
                  </p>
                </div>
              {:else}
                <div class="text-center py-8">
                  <svg
                    class="w-12 h-12 text-gray-600 mx-auto mb-3"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.121 2.122"
                    />
                  </svg>
                  <p class="text-gray-400 text-sm">
                    Selecciona un clip para ver su transcripción
                  </p>
                </div>
              {/if}
            </div>
          </div>

          <!-- Columna Central: Device Preview con controles -->
          <div class="overflow-y-auto space-y-4">
            <!-- iPhone 14 Pro Mockup -->
            <div class="relative mx-auto w-full flex justify-center">
              <!-- iPhone Frame -->
              <div
                class="relative bg-slate-900 rounded-[3.5rem] p-2 shadow-2xl shadow-black/40 border-2 border-slate-700"
              >
                <!-- iPhone Screen Container -->
                <div
                  class="relative bg-black rounded-[3rem] overflow-hidden w-[320px] h-[640px]"
                >
                  <!-- Dynamic Island -->
                  <div
                    class="absolute top-2 left-1/2 transform -translate-x-1/2 bg-black w-[120px] h-[30px] rounded-full z-20 border border-gray-700"
                  ></div>

                  <!-- Screen Content Area -->
                  <div
                    class="absolute inset-0 bg-black rounded-[3rem] overflow-hidden"
                  >
                    <!-- Status Bar -->
                    <div
                      class="absolute top-0 left-0 right-0 h-12 bg-gradient-to-b from-black/80 to-transparent z-10 flex items-center justify-between px-6 pt-2"
                    >
                      <div class="text-white text-sm font-medium">9:41</div>
                      <div class="flex items-center gap-1">
                        <div class="w-4 h-2 border border-white rounded-sm">
                          <div class="w-3 h-1 bg-white rounded-sm m-0.5"></div>
                        </div>
                      </div>
                    </div>

                    <!-- Video Container - Clickeable -->
                    <div
                      class="w-full h-full bg-black flex items-center justify-center relative group cursor-pointer"
                      onclick={() => {
                        const player = getActivePlayer();
                        if (player) player.togglePlayPause();
                      }}
                      onkeydown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          const player = getActivePlayer();
                          if (player) player.togglePlayPause();
                        }
                      }}
                      role="button"
                      tabindex="0"
                      title={isPlaying ? "Pausar video" : "Reproducir video"}
                    >
                      <!-- Play/Pause Overlay Icon -->
                      <div
                        class="absolute inset-0 flex items-center justify-center z-30 opacity-0 group-hover:opacity-100 transition-all duration-300 pointer-events-none"
                      >
                        <div
                          class="device-overlay rounded-full p-5 border border-white/30 shadow-2xl shadow-black/50 transform group-hover:scale-110 transition-all duration-300"
                          style="animation: gentle-pulse 2s ease-in-out infinite;"
                        >
                          {#if isPlaying}
                            <!-- Pause Icon -->
                            <svg
                              class="play-pause-icon w-10 h-10 text-white"
                              fill="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
                            </svg>
                          {:else}
                            <!-- Play Icon -->
                            <svg
                              class="play-pause-icon w-10 h-10 text-white"
                              fill="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path d="M8 5v14l11-7z" />
                            </svg>
                          {/if}
                        </div>
                      </div>

                      <!-- Video Content -->
                      {#if useCanvasRenderer}
                        <!-- VideoCanvasRenderer -->
                        <VideoCanvasRenderer
                          bind:this={canvasRenderer}
                          videoSrc="http://localhost:8080/api/videos/{video.id}/stream"
                          {currentTime}
                          {subtitles}
                          width={960}
                          height={1920}
                          onTimeUpdate={(time) => {
                            currentTime = time;
                          }}
                          onPlayingChange={(playing) => {
                            isPlaying = playing;
                          }}
                        />
                      {:else}
                        <!-- VideoPlayer + SubtitleCanvas -->
                        <VideoPlayer
                          bind:this={videoPlayerRef}
                          videoSrc="http://localhost:8080/api/videos/{video.id}/stream"
                          bind:currentTime
                          bind:isPlaying
                          {customStartTime}
                          {customEndTime}
                        />

                        <SubtitleCanvas
                          {currentSubtitle}
                          width={320}
                          height={640}
                          transition={subtitleTransition}
                          onPositionChange={handleSubtitlePositionChange}
                          {currentTime}
                          bgHexColor={subtitleBgHexColor}
                          bgOpacity={subtitleBgOpacity}
                          textColor={subtitleTextColor}
                          activeTextColor={subtitleActiveTextColor}
                          borderRadius={subtitleBorderRadius}
                          shadowBlur={subtitleShadowBlur}
                          position={subtitlePosition}
                          fontSize={subtitleFontSize}
                          fontFamily={subtitleFontFamily}
                          fontWeight={subtitleFontWeight}
                        />
                      {/if}
                    </div>

                    <!-- Home Indicator -->
                    <div
                      class="absolute bottom-2 left-1/2 transform -translate-x-1/2 w-32 h-1 bg-white rounded-full opacity-60"
                    ></div>
                  </div>
                </div>

                <!-- iPhone Side Buttons -->
                <div
                  class="absolute left-[-4px] top-20 w-1 h-8 bg-slate-700 rounded-l-full"
                ></div>
                <div
                  class="absolute left-[-4px] top-32 w-1 h-12 bg-slate-700 rounded-l-full"
                ></div>
                <div
                  class="absolute left-[-4px] top-48 w-1 h-12 bg-slate-700 rounded-l-full"
                ></div>
                <div
                  class="absolute right-[-4px] top-36 w-1 h-16 bg-slate-700 rounded-r-full"
                ></div>
              </div>
            </div>

            <!-- Video Controls -->
            <div class="glass-effect p-5">
              <div class="flex items-center justify-center gap-6 mb-4">
                <button
                  onclick={() => {
                    const player = getActivePlayer();
                    if (player) {
                      player.seekTo(Math.max(customStartTime, currentTime - 5));
                    }
                  }}
                  class="p-3 hover:bg-gray-700/50 rounded-xl transition-all hover:scale-110"
                  title="Retroceder 5s"
                >
                  <svg
                    class="w-5 h-5 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M12.066 11.2a1 1 0 000 1.6l5.334 4A1 1 0 0019 16V8a1 1 0 00-1.6-.8l-5.333 4zM4.066 11.2a1 1 0 000 1.6l5.334 4A1 1 0 0011 16V8a1 1 0 00-1.6-.8l-5.334 4z"
                    />
                  </svg>
                </button>

                <button
                  onclick={() => {
                    const player = getActivePlayer();
                    if (player) {
                      player.togglePlayPause();
                    }
                  }}
                  class="p-4 bg-blue-600 hover:bg-blue-700 rounded-full transition-all shadow-lg hover:scale-105 active:scale-95"
                >
                  {#if isPlaying}
                    <svg
                      class="w-6 h-6 text-white"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
                    </svg>
                  {:else}
                    <svg
                      class="w-6 h-6 text-white"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path d="M8 5v14l11-7z" />
                    </svg>
                  {/if}
                </button>

                <button
                  onclick={() => {
                    const player = getActivePlayer();
                    if (player) {
                      player.seekTo(Math.min(customEndTime, currentTime + 5));
                    }
                  }}
                  class="p-3 hover:bg-gray-700/50 rounded-xl transition-all hover:scale-110"
                  title="Avanzar 5s"
                >
                  <svg
                    class="w-5 h-5 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M11.933 12.8a1 1 0 000-1.6L6.6 7.2A1 1 0 005 8v8a1 1 0 001.6.8l5.333-4zM19.933 12.8a1 1 0 000-1.6l-5.333-4A1 1 0 0013 8v8a1 1 0 001.6.8l5.333-4z"
                    />
                  </svg>
                </button>
              </div>

              <!-- Progress Bar -->
              <div class="space-y-2">
                <input
                  type="range"
                  min={customStartTime}
                  max={customEndTime}
                  step="0.1"
                  value={currentTime}
                  oninput={(e) => {
                    const player = getActivePlayer();
                    const target = e.target as HTMLInputElement;
                    currentTime = parseFloat(target.value);
                    if (player && isPlaying) {
                      player.pause();
                    }
                  }}
                  onchange={(e) => {
                    const player = getActivePlayer();
                    const target = e.target as HTMLInputElement;
                    currentTime = parseFloat(target.value);
                    if (player) player.seekTo(currentTime);
                  }}
                  class="w-full h-1.5 bg-gray-700 rounded-full appearance-none cursor-pointer"
                  style="background: linear-gradient(to right, rgb(59 130 246) 0%, rgb(59 130 246) {((currentTime -
                    customStartTime) /
                    (customEndTime - customStartTime)) *
                    100}%, rgb(55 65 81) {((currentTime - customStartTime) /
                    (customEndTime - customStartTime)) *
                    100}%, rgb(55 65 81) 100%)"
                />

                <!-- Time Labels -->
                <div
                  class="flex justify-between text-xs text-gray-400 font-mono"
                >
                  <span>{formatTime(currentTime)}</span>
                  <span>{formatTime(customEndTime)}</span>
                </div>
              </div>
            </div>

            <!-- Sync Offset Control -->
            <div class="glass-effect p-6">
              <div class="flex items-center justify-between mb-3">
                <h4
                  class="text-sm font-semibold text-white flex items-center gap-2"
                >
                  <svg
                    class="w-4 h-4 text-blue-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  Ajuste de Sincronización
                </h4>
              </div>
              <div class="flex items-center space-x-3">
                <input
                  type="range"
                  bind:value={syncOffset}
                  min="-5"
                  max="5"
                  step="0.1"
                  class="flex-1 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                />
                <span
                  class="text-xs text-blue-300 min-w-[50px] text-center font-medium"
                >
                  {syncOffset > 0 ? "+" : ""}{syncOffset}s
                </span>
                {#if syncOffset !== 0}
                  <button
                    onclick={() => {
                      syncOffset = 0;
                    }}
                    class="p-1.5 text-gray-400 hover:text-white bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors"
                    title="Reset"
                  >
                    <svg
                      class="w-3 h-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                      />
                    </svg>
                  </button>
                {/if}
              </div>
              <div class="flex justify-between text-xs text-gray-500 mt-1">
                <span>Adelantar</span>
                <span>Retrasar</span>
              </div>
            </div>
          </div>

          <!-- Columna Derecha: Text & Estilo + Colores & Efectos -->
          <div class="overflow-y-auto space-y-4">
            <!-- Upper Settings Box: Basic & Style -->
            <div class="glass-effect p-6">
              {#if selectedClip}
                <div class="flex items-center justify-between mb-4">
                  <h4
                    class="text-lg font-semibold text-white flex items-center gap-2"
                  >
                    <svg
                      class="w-5 h-5 text-yellow-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2h4a1 1 0 011 1v1a1 1 0 01-1 1v9a3 3 0 01-3 3H6a3 3 0 01-3-3V7a1 1 0 01-1-1V5a1 1 0 011-1h4zM9 3v1h6V3H9zm-3 4v9a1 1 0 001 1h10a1 1 0 001-1V7H6z"
                      />
                    </svg>
                    Texto & Estilo
                  </h4>

                  <button
                    onclick={() => (showTemplateGallery = true)}
                    class="flex items-center gap-2 px-3 py-1.5 bg-yellow-600 hover:bg-yellow-700 text-white text-sm rounded-lg transition-colors"
                    title="Abrir galería de plantillas"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                      />
                    </svg>
                    Plantillas
                  </button>
                </div>

                <!-- Basic Controls -->
                <div class="space-y-4">
                  <!-- Words Per Line -->
                  <div class="space-y-2">
                    <label class="block text-sm font-medium text-yellow-200">
                      Palabras por Línea: {wordsPerLine}
                    </label>
                    <input
                      type="range"
                      bind:value={wordsPerLine}
                      min="1"
                      max="10"
                      step="1"
                      class="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                      onchange={handleSettingsChange}
                    />
                    <div class="flex justify-between text-xs text-gray-400">
                      <span>1 palabra</span>
                      <span>10 palabras</span>
                    </div>
                  </div>

                  <!-- Font Family -->
                  <div class="space-y-2">
                    <label class="block text-sm font-medium text-yellow-200"
                      >Fuente</label
                    >
                    <select
                      bind:value={subtitleFontFamily}
                      onchange={handleSettingsChange}
                      class="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-yellow-500"
                    >
                      <option value="Inter">Inter</option>
                      <option value="Poppins">Poppins</option>
                      <option value="Montserrat">Montserrat</option>
                      <option value="Roboto">Roboto</option>
                      <option value="Arial">Arial</option>
                    </select>
                  </div>

                  <!-- Font Size -->
                  <div class="space-y-2">
                    <label class="block text-sm font-medium text-yellow-200">
                      Tamaño: {subtitleFontSize}px
                    </label>
                    <input
                      type="range"
                      bind:value={subtitleFontSize}
                      min="12"
                      max="48"
                      step="1"
                      class="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                      onchange={handleSettingsChange}
                    />
                  </div>

                  <!-- Transition -->
                  <div class="space-y-2">
                    <label class="block text-sm font-medium text-yellow-200"
                      >Animación</label
                    >
                    <select
                      bind:value={subtitleTransition}
                      onchange={handleSettingsChange}
                      class="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-yellow-500"
                    >
                      <option value="none">Sin animación</option>
                      <option value="fade">Fade</option>
                      <option value="pop">Pop</option>
                      <option value="slide">Slide</option>
                    </select>
                  </div>
                </div>
              {:else}
                <div class="text-center py-8">
                  <svg
                    class="w-12 h-12 text-gray-600 mx-auto mb-3"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2h4a1 1 0 011 1v1a1 1 0 01-1 1v9a3 3 0 01-3 3H6a3 3 0 01-3-3V7a1 1 0 01-1-1V5a1 1 0 011-1h4zM9 3v1h6V3H9zm-3 4v9a1 1 0 001 1h10a1 1 0 001-1V7H6z"
                    />
                  </svg>
                  <p class="text-gray-400 text-sm">
                    Selecciona un clip para configurar
                  </p>
                </div>
              {/if}
            </div>

            <!-- Lower Settings Box: Colors & Effects -->
            <div class="glass-effect p-6">
              {#if selectedClip}
                <h4
                  class="text-lg font-semibold text-white mb-4 flex items-center gap-2"
                >
                  <svg
                    class="w-5 h-5 text-purple-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zM21 5a2 2 0 00-2-2h-4a2 2 0 00-2 2v12a4 4 0 004 4h4a2 2 0 002-2V5z"
                    />
                  </svg>
                  Colores & Efectos
                </h4>

                <div class="space-y-4">
                  <!-- Text Colors -->
                  <div class="grid grid-cols-2 gap-4">
                    <div class="space-y-2">
                      <label class="block text-sm font-medium text-purple-200"
                        >Color del Texto</label
                      >
                      <input
                        type="color"
                        bind:value={subtitleTextColor}
                        onchange={handleSettingsChange}
                        class="w-full h-10 bg-gray-800 border border-gray-600 rounded-lg cursor-pointer"
                      />
                    </div>
                    <div class="space-y-2">
                      <label class="block text-sm font-medium text-purple-200"
                        >Color Activo</label
                      >
                      <input
                        type="color"
                        bind:value={subtitleActiveTextColor}
                        onchange={handleSettingsChange}
                        class="w-full h-10 bg-gray-800 border border-gray-600 rounded-lg cursor-pointer"
                      />
                    </div>
                  </div>

                  <!-- Background -->
                  <div class="space-y-2">
                    <label class="block text-sm font-medium text-purple-200"
                      >Color de Fondo</label
                    >
                    <div class="flex gap-2">
                      <input
                        type="color"
                        bind:value={subtitleBgHexColor}
                        onchange={handleSettingsChange}
                        class="w-16 h-10 bg-gray-800 border border-gray-600 rounded-lg cursor-pointer"
                      />
                      <div class="flex-1">
                        <input
                          type="range"
                          bind:value={subtitleBgOpacity}
                          min="0"
                          max="1"
                          step="0.1"
                          class="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer mt-3"
                          onchange={handleSettingsChange}
                        />
                        <div class="text-xs text-gray-400 text-center mt-1">
                          Opacidad: {Math.round(subtitleBgOpacity * 100)}%
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Position -->
                  <div class="space-y-2">
                    <label class="block text-sm font-medium text-purple-200"
                      >Posición</label
                    >
                    <select
                      bind:value={subtitlePosition}
                      onchange={handleSettingsChange}
                      class="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-purple-500"
                    >
                      <option value="bottom">Abajo</option>
                      <option value="center">Centro</option>
                      <option value="top">Arriba</option>
                    </select>
                  </div>
                </div>
              {:else}
                <div class="text-center py-8">
                  <svg
                    class="w-12 h-12 text-gray-600 mx-auto mb-3"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zM21 5a2 2 0 00-2-2h-4a2 2 0 00-2 2v12a4 4 0 004 4h4a2 2 0 002-2V5z"
                    />
                  </svg>
                  <p class="text-gray-400 text-sm">
                    Selecciona un clip para personalizar
                  </p>
                </div>
              {/if}
            </div>
          </div>
        </div>
      {/if}
    </div>

    <!-- Bottom Section: Suggested Clips -->
    <div class="flex-shrink-0 max-h-96 max-w-screen-2xl mx-auto">
      {#if showClips}
        <!-- Full Section: Header + Content -->
        <div
          class="p-4 lg:p-6 rounded-t-3xl border-t border-gray-700 bg-gray-800/50 max-w-screen-2xl"
          style="position: absolute;
          bottom: 0px;
          left: 50%;
          transform: translateX(-50%);
          z-index: 999;"
        >
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-bold flex items-center gap-2 text-white">
              <svg
                class="w-5 h-5 text-blue-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              Clips Sugeridos
              {#if suggestedClips.length > 0}
                <span class="text-sm text-blue-300"
                  >({suggestedClips.length})</span
                >
              {/if}
            </h2>

            <button
              onclick={() => (showClips = false)}
              class="flex items-center gap-2 px-3 py-1.5 bg-gray-700 hover:bg-gray-600 text-white text-sm rounded-lg transition-colors"
              title="Ocultar clips"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 9l-7 7-7-7"
                />
              </svg>
              <span>Ocultar</span>
            </button>
          </div>

          {#if suggestedClips.length === 0}
            <p class="text-gray-400 text-sm">Aún no hay clips sugeridos...</p>
          {:else}
            <div
              class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 gap-3 lg:gap-4"
            >
              {#each suggestedClips as clip}
                <button
                  onclick={() => selectClip(clip)}
                  class="clip-card group relative w-full text-left rounded-xl transition-all duration-700 ease-out overflow-hidden h-[140px] hover:scale-[1.02] {getClipColorClasses(
                    clip.score,
                    selectedClip?.id === clip.id
                  )} {clip.score >= 90 ? 'viral-pulse' : ''}"
                >
                  <!-- Static Thumbnail Background (Always Visible) -->
                  <div class="absolute inset-0">
                    <video
                      class="w-full h-full object-cover blur-[1px] scale-105 opacity-50"
                      muted
                      preload="metadata"
                      poster="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='300'%3E%3Crect width='100%25' height='100%25' fill='%23374151'/%3E%3C/svg%3E"
                      onloadedmetadata={(e) => {
                        const videoEl = e.target as HTMLVideoElement;
                        videoEl.currentTime = clip.start_time;
                      }}
                    >
                      <source
                        src="http://localhost:8080/api/videos/{video.id}/stream#t={clip.start_time}"
                        type="video/mp4"
                      />
                    </video>
                    <!-- Base gradient for readability -->
                    <div
                      class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/40 to-black/20"
                    ></div>
                    <!-- Color overlay to maintain theme -->
                    <div
                      class="absolute inset-0 bg-current opacity-30 mix-blend-overlay"
                    ></div>
                  </div>

                  <!-- Video Background on Hover (Minimalist) -->
                  <div
                    role="presentation"
                    class="absolute inset-0 opacity-0 group-hover:opacity-100 transition-all duration-500 ease-out"
                    onmouseenter={(e) => {
                      const videoEl = e.currentTarget.querySelector(
                        "video"
                      ) as HTMLVideoElement;
                      if (videoEl) {
                        videoEl.currentTime = clip.start_time;
                        videoEl.play().catch(() => {});
                      }
                    }}
                    onmouseleave={(e) => {
                      const videoEl = e.currentTarget.querySelector(
                        "video"
                      ) as HTMLVideoElement;
                      if (videoEl) {
                        videoEl.pause();
                      }
                    }}
                  >
                    <!-- Background video with heavy blur -->
                    <video
                      class="w-full h-full object-cover blur-[8px] scale-110 opacity-40"
                      muted
                      loop
                      preload="metadata"
                      playsinline
                    >
                      <source
                        src="http://localhost:8080/api/videos/{video.id}/stream"
                        type="video/mp4"
                      />
                    </video>

                    <!-- Score-based color filter overlay -->
                    <div
                      class="absolute inset-0 mix-blend-multiply opacity-60"
                      style="background-color: {getScoreFilterColor(
                        clip.score
                      )}"
                    ></div>

                    <!-- Dark gradient for text readability -->
                    <div
                      class="absolute inset-0 bg-gradient-to-t from-black/90 via-transparent to-black/60"
                    ></div>
                  </div>

                  <!-- Content Container - Normal State -->
                  <div
                    class="relative z-10 p-3 bg-black bg-opacity-10 group-hover:opacity-0 transition-all duration-700 ease-out"
                  >
                    <!-- Header -->
                    <div class="flex justify-between items-start mb-1">
                      <h3 class="font-semibold text-white text-sm line-clamp-2">
                        {clip.title}
                      </h3>
                      <div class="flex items-center gap-1 ml-2 flex-shrink-0">
                        <span
                          class="text-xs font-bold text-white bg-gradient-to-r from-black/80 to-black/60 backdrop-blur-sm px-3 py-1.5 rounded-full border border-white/20 shadow-lg flex items-center gap-1"
                        >
                          <svg
                            class="w-3 h-3"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path d={getScoreIcon(clip.score)} />
                          </svg>
                          {Math.round(clip.score)}
                        </span>
                      </div>
                    </div>

                    <!-- Score Badge -->
                    <div class="flex items-center gap-1 mb-2">
                      <span
                        class="text-xs font-bold text-white bg-gradient-to-r from-black/80 to-black/60 backdrop-blur-sm px-3 py-1 rounded-full border border-white/30 shadow-lg flex items-center gap-2"
                      >
                        <svg
                          class="w-3 h-3"
                          fill="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path d={getScoreIcon(clip.score)} />
                        </svg>
                        {getScoreLabel(clip.score)}
                      </span>
                    </div>

                    <!-- Time Info -->
                    <div
                      class="flex justify-between items-center text-xs text-gray-300"
                    >
                      <span class="bg-black bg-opacity-50 px-2 py-1 rounded">
                        {formatTime(clip.start_time)} - {formatTime(
                          clip.end_time
                        )}
                      </span>
                      <div
                        class="flex items-center gap-1 bg-black bg-opacity-50 px-2 py-1 rounded"
                      >
                        <svg
                          class="w-3 h-3"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                          />
                        </svg>
                        <span class="font-medium"
                          >{Math.round(clip.end_time - clip.start_time)}s</span
                        >
                      </div>
                    </div>
                  </div>

                  <!-- Content Container - Hover State (Description Only) -->
                  <div
                    class="absolute inset-0 z-20 px-4 opacity-0 group-hover:opacity-100 transition-all duration-500 ease-out flex flex-col justify-center"
                  >
                    <!-- Clean centered content -->
                    <div class="text-center">
                      <!-- Main description (larger and more prominent) -->
                      <div class="">
                        <p
                          class="text-white text-xs leading-relaxed font-medium"
                        >
                          {clip.description}
                        </p>
                      </div>
                    </div>
                  </div>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {:else}
        <!-- Minimized Section: Only Arrow -->
        <div class="flex items-center justify-center py-3">
          <button
            onclick={() => (showClips = true)}
            class="flex items-center justify-center w-12 h-8 bg-gray-700 hover:bg-gray-600 text-gray-300 hover:text-white rounded-lg transition-all duration-300 hover:scale-110"
            title="Mostrar clips sugeridos"
            style="border-radius: 100%;
    height: 5rem;
    width: 5rem;"
          >
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 15l7-7 7 7"
              />
            </svg>
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>

<!-- Export Progress Modal -->
<ExportProgress
  isVisible={isExporting}
  progress={exportProgress}
  stage={exportStage}
  onCancel={cancelExport}
/>

<!-- Template Gallery Modal -->
<TemplateGallery
  bind:isOpen={showTemplateGallery}
  templates={subtitleTemplates}
  onSelect={(template) => {
    // Aplicar configuración de la plantilla seleccionada
    subtitleFontFamily = template.settings.fontFamily;
    subtitleFontWeight = template.settings.fontWeight;
    subtitleTransition = template.settings.transition as
      | "none"
      | "pop"
      | "fade"
      | "slide"
      | "bounce"
      | "elastic"
      | "rotate";
    subtitleBgHexColor = template.settings.bgHexColor;
    subtitleBgOpacity = template.settings.bgOpacity;
    subtitleTextColor = template.settings.textColor;
    subtitleActiveTextColor = template.settings.activeTextColor;
    subtitleBorderRadius = template.settings.borderRadius;
    subtitleShadowBlur = template.settings.shadowBlur;

    // Notificar cambios
    handleSettingsChange();

    console.log("🎨 Plantilla aplicada:", template.name);
  }}
/>

<style>
  /* Animaciones premium para botón de exportar */
  @keyframes shimmer {
    0% {
      transform: translateX(-100%) skewX(-12deg);
    }
    100% {
      transform: translateX(300%) skewX(-12deg);
    }
  }

  @keyframes pulse-glow {
    0%,
    100% {
      box-shadow:
        0 0 20px rgba(16, 185, 129, 0.3),
        0 0 60px rgba(16, 185, 129, 0.1);
    }
    50% {
      box-shadow:
        0 0 30px rgba(16, 185, 129, 0.5),
        0 0 80px rgba(16, 185, 129, 0.2);
    }
  }

  @keyframes float-particles {
    0%,
    100% {
      opacity: 0.6;
      transform: translateY(0px) scale(1);
    }
    50% {
      opacity: 1;
      transform: translateY(-3px) scale(1.2);
    }
  }

  /* Aplicar animaciones solo al botón de exportar */
  button.group:hover {
    animation: pulse-glow 2s ease-in-out infinite;
  }

  button.group:hover .animate-pulse {
    animation: float-particles 1.5s ease-in-out infinite;
  }

  /* Mejorar select personalizado */
  #export-format {
    background-color: #1f2937 !important;
    color: white !important;
  }

  #export-format option {
    background-color: #1f2937 !important;
    color: white !important;
    padding: 8px 12px;
  }

  #export-format option:checked,
  #export-format option:hover {
    background-color: #10b981 !important;
    color: white !important;
  }

  /* Asegurar que el select mantenga el color oscuro */
  #export-format:focus,
  #export-format:active {
    background-color: #374151 !important;
    color: white !important;
  }

  /* Efectos especiales para clips según score - solo botones de clips */
  .clip-card:hover {
    animation: subtle-glow 3s ease-in-out infinite;
  }

  @keyframes subtle-glow {
    0%,
    100% {
      filter: brightness(1) saturate(1);
    }
    50% {
      filter: brightness(1.1) saturate(1.2);
    }
  }

  /* Mejoras visuales para badges */
  .score-badge {
    backdrop-filter: blur(8px);
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
  }

  /* Efectos de pulso para clips virales (score >= 90) */
  .viral-pulse {
    animation: viral-glow 2s ease-in-out infinite;
  }

  @keyframes viral-glow {
    0%,
    100% {
      box-shadow:
        0 0 20px rgba(245, 158, 11, 0.4),
        0 0 40px rgba(245, 158, 11, 0.2);
    }
    50% {
      box-shadow:
        0 0 30px rgba(245, 158, 11, 0.6),
        0 0 60px rgba(245, 158, 11, 0.3);
    }
  }

  /* Efectos hover específicos por score para mantener colores correctos */
  .clip-card:hover {
    border-color: inherit !important;
  }

  /* Asegurar que los gradientes de fondo mantengan precedencia */
  .clip-card {
    background: inherit !important;
  }

  /* Efectos para el overlay de play/pause en el device */
  .device-overlay {
    background: radial-gradient(
      circle,
      rgba(0, 0, 0, 0.8) 0%,
      rgba(0, 0, 0, 0.6) 100%
    );
  }

  /* Animación suave para el ícono de play/pause */
  .play-pause-icon {
    filter: drop-shadow(0 0 10px rgba(255, 255, 255, 0.3));
  }

  .play-pause-icon:hover {
    filter: drop-shadow(0 0 15px rgba(255, 255, 255, 0.5));
  }

  /* Pulso suave para el overlay */
  @keyframes gentle-pulse {
    0%,
    100% {
      opacity: 0.9;
      transform: scale(1);
    }
    50% {
      opacity: 1;
      transform: scale(1.05);
    }
  }

  /* Grid pattern de fondo */
  .bg-grid-pattern {
    background-image: linear-gradient(
        to right,
        rgb(255, 255, 255) 1px,
        transparent 1px
      ),
      linear-gradient(to bottom, rgb(255, 255, 255) 1px, transparent 1px);
    background-size: 100px 100px;
    filter: blur(2px);
  }
  @keyframes grid-move {
    0% {
      transform: translate(0, 0);
    }
    100% {
      transform: translate(-100px, 100px);
    }
  }

  .animate-grid-move {
    animation: grid-move 20s linear infinite;
  }

  /* Glass morphism effect - From https://css.glass */
  .glass-effect {
    /* From https://css.glass */
    background: rgba(255, 255, 255, 0);
    border-radius: 16px;
    box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
    backdrop-filter: blur(4.7px);
    -webkit-backdrop-filter: blur(4.7px);
    border: 1px solid rgba(255, 255, 255, 0.14);
  }
</style>
