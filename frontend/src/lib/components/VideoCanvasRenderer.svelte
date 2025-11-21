<script lang="ts">
  import { onMount } from "svelte";
  import type { SubtitleConfig } from "$lib/types";

  /**
   * VideoCanvasRenderer - Renderiza video + subtítulos con animaciones en Canvas
   * Este componente es la "fuente única de verdad" para la visualización y exportación
   */

  type Props = {
    videoSrc: string;
    currentTime?: number;
    subtitles?: SubtitleConfig[];
    width?: number;
    height?: number;
    isExporting?: boolean;
    onRenderFrame?: (
      canvas: HTMLCanvasElement,
      ctx: CanvasRenderingContext2D
    ) => void;
    onTimeUpdate?: (time: number) => void;
    onPlayingChange?: (isPlaying: boolean) => void;
  };

  let {
    videoSrc,
    currentTime = 0,
    subtitles = [],
    width = 1080,
    height = 1920,
    isExporting = false,
    onRenderFrame,
    onTimeUpdate,
    onPlayingChange,
  }: Props = $props();

  let canvasElement: HTMLCanvasElement;
  let videoElement: HTMLVideoElement;
  let ctx: CanvasRenderingContext2D | null = null;
  let animationFrameId: number | null = null;
  let isReady = $state(false);

  onMount(() => {
    if (!canvasElement || !videoElement) return;

    ctx = canvasElement.getContext("2d", {
      alpha: false,
      desynchronized: true, // Mejor performance
      willReadFrequently: false, // Optimización para escritura
    });

    if (!ctx) {
      console.error("❌ No se pudo obtener contexto 2D del canvas");
      return;
    }

    // Optimizar calidad de renderizado
    ctx.imageSmoothingEnabled = true;
    ctx.imageSmoothingQuality = isExporting ? "low" : "high"; // low = más rápido durante exportación

    // Guardar referencias a los event listeners para poder limpiarlos
    const handleLoadedData = () => {
      console.log("✅ Video cargado en VideoCanvasRenderer");
      isReady = true;
      renderFrame();
    };

    const handleSeeked = () => {
      renderFrame();
    };

    const handlePlay = () => {
      console.log("▶️ Video 'play' event fired");
      if (onPlayingChange) {
        console.log("▶️ Calling onPlayingChange(true)");
        onPlayingChange(true);
      }
    };

    const handlePause = () => {
      console.log("⏸️ Video 'pause' event fired");
      if (onPlayingChange) {
        console.log("⏸️ Calling onPlayingChange(false)");
        onPlayingChange(false);
      }
    };

    const handleTimeUpdate = () => {
      if (videoElement && onTimeUpdate && !isExporting) {
        onTimeUpdate(videoElement.currentTime);
      }
    };

    videoElement.addEventListener("loadeddata", handleLoadedData);
    videoElement.addEventListener("seeked", handleSeeked);
    videoElement.addEventListener("play", handlePlay);
    videoElement.addEventListener("pause", handlePause);
    videoElement.addEventListener("timeupdate", handleTimeUpdate);

    return () => {
      if (animationFrameId) {
        cancelAnimationFrame(animationFrameId);
      }

      // Limpiar event listeners
      if (videoElement) {
        videoElement.removeEventListener("loadeddata", handleLoadedData);
        videoElement.removeEventListener("seeked", handleSeeked);
        videoElement.removeEventListener("play", handlePlay);
        videoElement.removeEventListener("pause", handlePause);
        videoElement.removeEventListener("timeupdate", handleTimeUpdate);
      }
    };
  });

  // Actualizar tiempo del video cuando cambia currentTime desde el padre
  $effect(() => {
    if (
      videoElement &&
      isReady &&
      !isExporting &&
      Math.abs(videoElement.currentTime - currentTime) > 0.1
    ) {
      videoElement.currentTime = currentTime;
    }
  });

  // Pausar video y silenciar durante exportación
  $effect(() => {
    if (videoElement && isReady) {
      if (isExporting) {
        // Pausar y silenciar video al empezar exportación
        videoElement.pause();
        videoElement.muted = true;
      } else {
        // Restaurar audio cuando termina exportación
        videoElement.muted = false;
      }
    }
  });

  // Renderizar frame continuamente (siempre, incluso cuando está pausado para ver transiciones)
  $effect(() => {
    if (isReady) {
      const loop = () => {
        renderFrame();
        animationFrameId = requestAnimationFrame(loop);
      };
      loop();

      return () => {
        if (animationFrameId) {
          cancelAnimationFrame(animationFrameId);
          animationFrameId = null;
        }
      };
    }
  });

  // Cache para mediciones de texto (evita recalcular en cada frame)
  let textMeasureCache = new Map<string, { width: number; height: number }>();
  let lastSubtitleText = "";

  // Cache del índice del último subtítulo activo para búsqueda optimizada
  let lastActiveSubtitleIndex = 0;

  function renderFrame() {
    if (!ctx || !videoElement || !canvasElement) return;

    // Limpiar canvas
    ctx.clearRect(0, 0, width, height);

    // Dibujar video
    // Escalar video para llenar canvas manteniendo aspect ratio (crop center)
    const videoAspect = videoElement.videoWidth / videoElement.videoHeight;
    const canvasAspect = width / height;

    let sx = 0,
      sy = 0,
      sw = videoElement.videoWidth,
      sh = videoElement.videoHeight;

    if (videoAspect > canvasAspect) {
      // Video es más ancho, crop horizontalmente
      sw = videoElement.videoHeight * canvasAspect;
      sx = (videoElement.videoWidth - sw) / 2;
    } else {
      // Video es más alto, crop verticalmente
      sh = videoElement.videoWidth / canvasAspect;
      sy = (videoElement.videoHeight - sh) / 2;
    }

    ctx.drawImage(videoElement, sx, sy, sw, sh, 0, 0, width, height);

    // Renderizar subtítulos
    renderSubtitles(ctx, videoElement.currentTime);

    // Callback para exportación (si está activo)
    if (onRenderFrame) {
      onRenderFrame(canvasElement, ctx);
    }
  }

  function renderSubtitles(ctx: CanvasRenderingContext2D, time: number) {
    // Búsqueda optimizada: empezar desde el último índice conocido
    let activeSubtitle = null;

    // Verificar primero el último subtítulo activo (muy probable que siga siendo el mismo)
    if (lastActiveSubtitleIndex < subtitles.length) {
      const sub = subtitles[lastActiveSubtitleIndex];
      if (time >= sub.start_time && time <= sub.end_time) {
        activeSubtitle = sub;
      }
    }

    // Si no es el mismo, buscar hacia adelante (más común) o hacia atrás
    if (!activeSubtitle) {
      // Buscar hacia adelante
      for (let i = lastActiveSubtitleIndex + 1; i < subtitles.length; i++) {
        const sub = subtitles[i];
        if (time >= sub.start_time && time <= sub.end_time) {
          activeSubtitle = sub;
          lastActiveSubtitleIndex = i;
          break;
        }
        if (time < sub.start_time) break; // Ya pasamos el rango
      }

      // Si no lo encontramos hacia adelante, buscar hacia atrás
      if (!activeSubtitle && lastActiveSubtitleIndex > 0) {
        for (let i = lastActiveSubtitleIndex - 1; i >= 0; i--) {
          const sub = subtitles[i];
          if (time >= sub.start_time && time <= sub.end_time) {
            activeSubtitle = sub;
            lastActiveSubtitleIndex = i;
            break;
          }
          if (time > sub.end_time) break; // Ya pasamos el rango
        }
      }
    }

    if (!activeSubtitle) {
      // Limpiar cache si no hay subtítulo activo
      if (lastSubtitleText) {
        textMeasureCache.clear();
        lastSubtitleText = "";
      }
      return;
    }

    const sub = activeSubtitle;
    const duration = sub.end_time - sub.start_time;
    const progress = Math.max(
      0,
      Math.min(1, (time - sub.start_time) / duration)
    );

    // Aplicar transiciones/animaciones
    ctx.save();

    applyTransition(ctx, sub.transition || "none", progress);

    // Configurar texto
    const fontSize = sub.font_size || 48;
    const fontWeight = sub.font_weight || 600;
    const fontFamily = sub.font_family || "Inter";
    ctx.font = `${fontWeight} ${fontSize}px "${fontFamily}", sans-serif`;
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";

    // Cache de mediciones de texto - solo calcular si cambió el subtítulo
    const cacheKey = `${sub.text}_${fontSize}_${fontWeight}_${fontFamily}`;
    let textWidth: number;
    let textHeight: number;

    if (lastSubtitleText === cacheKey && textMeasureCache.has(cacheKey)) {
      const cached = textMeasureCache.get(cacheKey)!;
      textWidth = cached.width;
      textHeight = cached.height;
    } else {
      // Medir solo cuando cambia el subtítulo
      const metrics = ctx.measureText(sub.text);
      textWidth = metrics.width;
      textHeight = fontSize * 1.2;
      textMeasureCache.set(cacheKey, { width: textWidth, height: textHeight });
      lastSubtitleText = cacheKey;
    }

    // Calcular posición
    let yPos = height - 100; // bottom
    if (sub.position === "top") {
      yPos = 100;
    } else if (sub.position === "center") {
      yPos = height / 2;
    }

    const xPos = width / 2;
    const padding = sub.font_size * 0.6 || 24;
    const borderRadius = sub.border_radius || 8;

    // Dibujar sombra del background
    if (sub.shadow_blur > 0) {
      ctx.shadowColor = "rgba(0, 0, 0, 0.3)";
      ctx.shadowBlur = sub.shadow_blur;
      ctx.shadowOffsetX = 0;
      ctx.shadowOffsetY = 4;
    }

    // Dibujar background con border radius
    ctx.fillStyle = parseColorWithAlpha(sub.bg_color, sub.bg_opacity);
    drawRoundedRect(
      ctx,
      xPos - textWidth / 2 - padding,
      yPos - textHeight / 2 - padding / 2,
      textWidth + padding * 2,
      textHeight + padding,
      borderRadius
    );
    ctx.fill();

    // Reset shadow para texto
    ctx.shadowColor = "transparent";
    ctx.shadowBlur = 0;

    // Efecto karaoke (palabra por palabra) - optimizado con cache
    const words = sub.text.split(" ");
    const wordsProgress = Math.min(progress * 1.2, 1); // Ligeramente más rápido
    const activeWordIndex = Math.floor(wordsProgress * words.length);

    // Cache de anchos de palabras para evitar medir en cada frame
    const wordsCacheKey = `${cacheKey}_words`;
    let wordWidths: number[];

    if (textMeasureCache.has(wordsCacheKey)) {
      wordWidths = textMeasureCache.get(wordsCacheKey) as any;
    } else {
      // Calcular anchos solo cuando cambia el subtítulo
      wordWidths = words.map((word) => ctx.measureText(word + " ").width);
      textMeasureCache.set(wordsCacheKey, wordWidths as any);
    }

    let currentX = xPos - textWidth / 2;

    // Configurar sombra de texto una sola vez (fuera del loop)
    ctx.shadowColor = "rgba(0, 0, 0, 0.5)";
    ctx.shadowBlur = 2;
    ctx.shadowOffsetX = 0;
    ctx.shadowOffsetY = 2;

    words.forEach((word, index) => {
      const wordWidth = wordWidths[index];
      const isActive = index <= activeWordIndex;

      // Color del texto
      ctx.fillStyle = isActive ? sub.active_text_color || sub.color : sub.color;

      ctx.fillText(word, currentX + wordWidth / 2, yPos);
      currentX += wordWidth;
    });

    ctx.restore();
  }

  function applyTransition(
    ctx: CanvasRenderingContext2D,
    transition: string = "none",
    progress: number
  ) {
    // Easing functions premium - Apple/Screen.studio style
    const easeOutExpo = (t: number) => (t === 1 ? 1 : 1 - Math.pow(2, -10 * t));
    const easeOutQuint = (t: number) => 1 - Math.pow(1 - t, 5);
    const easeOutCubic = (t: number) => 1 - Math.pow(1 - t, 3);

    // Apple spring physics - más suave y natural
    const appleSpring = (t: number) => {
      const c1 = 1.70158 * 0.8; // Reducido para ser más sutil
      const c3 = c1 + 1;
      return 1 + c3 * Math.pow(t - 1, 3) + c1 * Math.pow(t - 1, 2);
    };

    // Screen.studio blur fade
    const screenStudioFade = (t: number) => {
      // Combinación de easeOutQuint con aceleración inicial
      if (t < 0.5) {
        return 2 * t * t;
      }
      return 1 - Math.pow(-2 * t + 2, 2) / 2;
    };

    // Entrada más rápida y suave (primeros 25% del tiempo)
    const entranceProgress = Math.min(progress / 0.25, 1);

    switch (transition) {
      case "pop":
        // Apple-style pop - spring sutil con fade exponencial
        const popScale = 0.8 + appleSpring(entranceProgress) * 0.2;
        const popAlpha = easeOutExpo(entranceProgress);
        ctx.globalAlpha = popAlpha;
        ctx.translate(width / 2, height / 2);
        ctx.scale(popScale, popScale);
        ctx.translate(-width / 2, -height / 2);
        break;

      case "bounce":
        // Bounce más controlado - dos rebotes suaves
        const bounceT = entranceProgress;
        let bounceScale = 1;
        if (bounceT < 0.4) {
          bounceScale = easeOutQuint(bounceT / 0.4);
        } else if (bounceT < 0.7) {
          bounceScale = 1 + Math.sin((bounceT - 0.4) * Math.PI * 3.33) * 0.08;
        } else {
          bounceScale = 1 + Math.sin((bounceT - 0.7) * Math.PI * 3.33) * 0.03;
        }
        ctx.globalAlpha = easeOutExpo(entranceProgress);
        ctx.translate(width / 2, height / 2);
        ctx.scale(bounceScale, bounceScale);
        ctx.translate(-width / 2, -height / 2);
        break;

      case "elastic":
        // Elastic suavizado - menos dramático
        const elasticT = entranceProgress;
        const elasticScale =
          1 -
          Math.pow(2, -8 * elasticT) *
            Math.sin(((elasticT * 8 - 0.5) * Math.PI) / 2) *
            0.3;
        ctx.globalAlpha = easeOutExpo(entranceProgress);
        ctx.translate(width / 2, height / 2);
        ctx.scale(elasticScale, elasticScale);
        ctx.translate(-width / 2, -height / 2);
        break;

      case "fade":
        // Screen.studio style fade - muy suave
        ctx.globalAlpha = screenStudioFade(entranceProgress);
        break;

      case "slide":
        // Slide suave desde abajo - movimiento fluido
        const slideY = (1 - easeOutQuint(entranceProgress)) * 60; // Reducido de 120 a 60
        const slideAlpha = easeOutExpo(entranceProgress);
        ctx.translate(0, slideY);
        ctx.globalAlpha = slideAlpha;
        break;

      case "rotate":
        // Rotación sutil - solo 10 grados
        const rotation = ((1 - easeOutQuint(entranceProgress)) * Math.PI) / 18; // Reducido de π/6 a π/18
        const rotateAlpha = easeOutExpo(entranceProgress);
        ctx.translate(width / 2, height / 2);
        ctx.rotate(rotation);
        ctx.translate(-width / 2, -height / 2);
        ctx.globalAlpha = rotateAlpha;
        break;

      case "none":
      default:
        // Sin animación
        break;
    }
  }

  function drawRoundedRect(
    ctx: CanvasRenderingContext2D,
    x: number,
    y: number,
    w: number,
    h: number,
    r: number
  ) {
    if (w < 2 * r) r = w / 2;
    if (h < 2 * r) r = h / 2;
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.arcTo(x + w, y, x + w, y + h, r);
    ctx.arcTo(x + w, y + h, x, y + h, r);
    ctx.arcTo(x, y + h, x, y, r);
    ctx.arcTo(x, y, x + w, y, r);
    ctx.closePath();
  }

  function parseColorWithAlpha(color: string, opacity: number): string {
    if (color.startsWith("#")) {
      const hex = color.replace("#", "");
      const r = parseInt(hex.substr(0, 2), 16);
      const g = parseInt(hex.substr(2, 2), 16);
      const b = parseInt(hex.substr(4, 2), 16);
      return `rgba(${r}, ${g}, ${b}, ${opacity})`;
    } else if (color.startsWith("rgba")) {
      // Reemplazar alpha existente
      return color.replace(/[\d.]+\)$/g, `${opacity})`);
    } else if (color.startsWith("rgb")) {
      return color.replace("rgb", "rgba").replace(")", `, ${opacity})`);
    }
    return `rgba(0, 0, 0, ${opacity})`;
  }

  // Exponer método para exportación frame-by-frame
  export function renderAtTime(time: number): Promise<void> {
    return new Promise((resolve) => {
      if (!videoElement) {
        resolve();
        return;
      }

      videoElement.currentTime = time;
      videoElement.onseeked = () => {
        renderFrame();
        resolve();
      };
    });
  }

  export function getCanvas(): HTMLCanvasElement {
    return canvasElement;
  }

  export function getVideoElement(): HTMLVideoElement {
    return videoElement;
  }

  export function getContext(): CanvasRenderingContext2D | null {
    return ctx;
  }

  // Métodos de control de video (compatibles con VideoPlayer)
  export function play() {
    console.log("🎮 VideoCanvasRenderer.play() called");
    if (videoElement) {
      videoElement.play();
    }
  }

  export function pause() {
    console.log("🎮 VideoCanvasRenderer.pause() called");
    if (videoElement) {
      videoElement.pause();
    }
  }

  export function togglePlayPause() {
    console.log(
      "🎮 VideoCanvasRenderer.togglePlayPause() called, current paused:",
      videoElement?.paused
    );
    if (videoElement) {
      if (videoElement.paused) {
        videoElement.play();
      } else {
        videoElement.pause();
      }
    }
  }

  export function seekTo(time: number) {
    if (videoElement) {
      videoElement.currentTime = time;
    }
  }

  export function isPlaying(): boolean {
    return videoElement ? !videoElement.paused : false;
  }
</script>

<div class="video-canvas-renderer">
  <canvas
    bind:this={canvasElement}
    {width}
    {height}
    class="render-canvas"
    class:exporting={isExporting}
  ></canvas>

  <!-- Video oculto usado como fuente -->
  <video
    bind:this={videoElement}
    src={videoSrc}
    class="hidden-video"
    preload="auto"
    crossorigin="anonymous"
  >
    <track kind="captions" label="Subtítulos" />
  </video>
</div>

<style>
  .video-canvas-renderer {
    position: relative;
    width: 100%;
    height: 100%;
  }

  .render-canvas {
    width: 100%;
    height: 100%;
    object-fit: contain;
    background: #000;
  }

  .render-canvas.exporting {
    position: absolute;
    visibility: hidden;
    pointer-events: none;
  }

  .hidden-video {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
  }
</style>
