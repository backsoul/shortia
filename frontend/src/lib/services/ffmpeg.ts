import { FFmpeg } from "@ffmpeg/ffmpeg";
import { fetchFile, toBlobURL } from "@ffmpeg/util";

let ffmpegInstance: FFmpeg | null = null;
let isLoading = false;

export async function getFFmpeg(): Promise<FFmpeg> {
  if (ffmpegInstance) {
    console.log("✅ Reutilizando instancia de FFmpeg existente");
    return ffmpegInstance;
  }

  if (isLoading) {
    console.log("⏳ FFmpeg ya se está cargando, esperando...");
    while (isLoading) {
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
    return ffmpegInstance!;
  }

  isLoading = true;
  console.log("📦 Creando nueva instancia de FFmpeg...");
  ffmpegInstance = new FFmpeg();

  // Usar archivos locales servidos desde /static/ffmpeg/
  const baseURL = "/ffmpeg";

  ffmpegInstance.on("log", ({ message }) => {
    console.log("[FFmpeg Log]", message);
  });

  ffmpegInstance.on("progress", ({ progress, time }) => {
    console.log(
      `[FFmpeg Progress] ${Math.round(progress * 100)}% - Time: ${time}µs`
    );
  });

  try {
    console.log("🌐 Cargando FFmpeg core desde archivos locales:", baseURL);

    const coreURL = await toBlobURL(
      `${baseURL}/ffmpeg-core.js`,
      "text/javascript"
    );
    console.log("✅ ffmpeg-core.js cargado");

    const wasmURL = await toBlobURL(
      `${baseURL}/ffmpeg-core.wasm`,
      "application/wasm"
    );
    console.log("✅ ffmpeg-core.wasm cargado");

    console.log("⚙️ Inicializando FFmpeg...");
    await ffmpegInstance.load({
      coreURL,
      wasmURL,
    });
    console.log("✅ FFmpeg loaded successfully");
  } catch (error) {
    console.error("❌ Failed to load FFmpeg:", error);
    isLoading = false;
    ffmpegInstance = null;
    throw error;
  }

  isLoading = false;
  return ffmpegInstance;
}

/**
 * Exporta clip usando MediaRecorder API + Canvas Rendering
 * Esta es la solución Frontend-Only: los subtítulos se renderizan en canvas
 * y se graban directamente como video, sin procesamiento backend.
 *
 * VENTAJAS:
 * - Una sola fuente de verdad para estilos
 * - Animaciones ilimitadas (CSS/Canvas)
 * - Sin duplicación de código
 * - WYSIWYG real
 *
 * @param canvasElement - Canvas con video + subtítulos renderizados
 * @param videoElement - Elemento video para extraer audio
 * @param startTime - Tiempo de inicio
 * @param endTime - Tiempo de fin
 * @param onProgress - Callback de progreso (progress: 0-100, stage: string)
 * @param onCancel - Callback para detectar cancelación (debe retornar true si fue cancelado)
 */
// Función para detectar Safari
function isSafari(): boolean {
  return /^((?!chrome|android).)*safari/i.test(navigator.userAgent);
}

// Función para verificar si captureStream está disponible
function isCaptureStreamSupported(): boolean {
  return "captureStream" in HTMLVideoElement.prototype;
}

export async function exportClipWithMediaRecorder(
  canvasElement: HTMLCanvasElement,
  videoElement: HTMLVideoElement,
  startTime: number,
  endTime: number,
  onProgress?: (progress: number, stage: string) => void,
  onCancel?: () => boolean
): Promise<Blob> {
  console.log("🎬 Exportando con MediaRecorder API (Frontend-Only)");

  // Verificar compatibilidad con Safari
  if (isSafari() || !isCaptureStreamSupported()) {
    console.warn(
      "⚠️ Safari o captureStream no soportado, usando fallback con FFmpeg"
    );
    throw new Error("SAFARI_NOT_SUPPORTED");
  }

  return new Promise((resolve, reject) => {
    let recorder: MediaRecorder | null = null;
    let progressTimer: any = null;
    let recordingTimeout: any = null;
    let timeUpdateListener:
      | ((this: HTMLVideoElement, ev: Event) => any)
      | null = null;

    let restoreAudioState = () => {};
    let combinedStream: MediaStream | null = null;

    const cleanup = () => {
      if (progressTimer) clearInterval(progressTimer);
      if (recordingTimeout) clearTimeout(recordingTimeout);
      if (timeUpdateListener) {
        videoElement.removeEventListener("timeupdate", timeUpdateListener);
      }
      videoElement.onseeked = null;
      videoElement.pause();
      restoreAudioState();
      combinedStream
        ?.getTracks()
        .forEach((track: MediaStreamTrack) => track.stop());
    };

    try {
      if (onProgress) onProgress(5, "preparing");

      // Capturar stream del canvas con FPS optimizado para velocidad
      // 30 FPS es un buen balance entre calidad y velocidad de procesamiento
      const canvasStream = canvasElement.captureStream(30);
      console.log("✅ Canvas stream capturado a 30fps");

      if (onProgress) onProgress(10, "preparing");

      // Capturar audio del video original
      let audioTracks: MediaStreamTrack[] = [];
      try {
        // @ts-ignore - captureStream está disponible en HTMLVideoElement cuando es soportado
        const videoStream = videoElement.captureStream();
        audioTracks = videoStream.getAudioTracks();
        console.log("✅ Audio tracks extraídos:", audioTracks.length);
      } catch (error) {
        console.warn("⚠️ No se pudo capturar audio del video:", error);
        // Continuar sin audio si hay problemas
      }

      // Combinar video del canvas + audio del video
      combinedStream = new MediaStream([
        ...canvasStream.getVideoTracks(),
        ...audioTracks,
      ]);

      const previousMuted = videoElement.muted;
      const previousVolume = videoElement.volume ?? 1;
      restoreAudioState = () => {
        videoElement.muted = previousMuted;
        videoElement.volume = previousVolume;
      };

      videoElement.muted = true;
      videoElement.volume = 0;

      const chunks: Blob[] = [];

      // Usar VP8 que es más rápido de codificar que VP9
      let mimeType = "video/webm;codecs=vp8,opus";
      if (!MediaRecorder.isTypeSupported(mimeType)) {
        mimeType = "video/webm;codecs=vp9,opus";
      }
      if (!MediaRecorder.isTypeSupported(mimeType)) {
        mimeType = "video/webm";
      }
      console.log("📹 Usando codec:", mimeType);

      recorder = new MediaRecorder(combinedStream!, {
        mimeType,
        videoBitsPerSecond: 5000000, // 5 Mbps - reducido para velocidad sin perder mucha calidad
      });

      recorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          chunks.push(event.data);
          console.log("📦 Chunk recibido:", event.data.size, "bytes");
        }
      };

      recorder.onstop = () => {
        cleanup();

        // Verificar cancelación
        if (onCancel && onCancel()) {
          console.log("❌ Exportación cancelada por usuario");
          reject(new Error("Exportación cancelada"));
          return;
        }

        const blob = new Blob(chunks, { type: mimeType });
        console.log("✅ Grabación completada:", blob.size, "bytes");
        if (onProgress) onProgress(90, "finalizing");

        // Pequeño delay para mostrar el estado de finalización
        setTimeout(() => {
          if (onProgress) onProgress(100, "complete");
          resolve(blob);
        }, 500);
      };

      recorder.onerror = (error) => {
        console.error("❌ Error en grabación:", error);
        cleanup();
        reject(error);
      };

      if (onProgress) onProgress(15, "extracting");

      // Posicionar video en inicio del clip
      videoElement.currentTime = startTime;

      let isRecordingStarted = false; // Flag para evitar múltiples inicios

      // Listener para detener la grabación cuando el video llegue al endTime
      timeUpdateListener = () => {
        if (
          videoElement.currentTime >= endTime &&
          recorder &&
          recorder.state !== "inactive"
        ) {
          console.log("⏹️ Video llegó al endTime, deteniendo grabación");
          if (onProgress) onProgress(85, "encoding");
          recorder.stop();
          videoElement.pause();
          if (timeUpdateListener) {
            videoElement.removeEventListener("timeupdate", timeUpdateListener);
          }
        }
      };

      videoElement.onseeked = () => {
        // Evitar múltiples inicios
        if (isRecordingStarted) {
          console.log("⚠️ Ignorando evento onseeked duplicado");
          return;
        }

        // Verificar cancelación antes de iniciar
        if (onCancel && onCancel()) {
          console.log("❌ Exportación cancelada antes de iniciar grabación");
          restoreAudioState();
          videoElement.onseeked = null;
          videoElement.pause();
          combinedStream
            ?.getTracks()
            .forEach((track: MediaStreamTrack) => track.stop());
          reject(new Error("Exportación cancelada"));
          return;
        }

        isRecordingStarted = true;
        console.log(
          "▶️ Iniciando grabación en",
          startTime,
          "s hasta",
          endTime,
          "s"
        );
        if (onProgress) onProgress(20, "rendering");

        // Chunks más grandes = menos overhead = más rápido
        // 500ms chunks en vez de 100ms
        recorder!.start(500);

        // Agregar listener para detener cuando llegue al endTime
        if (timeUpdateListener) {
          videoElement.addEventListener("timeupdate", timeUpdateListener);
        }

        videoElement.play();

        // Duración del clip
        const duration = (endTime - startTime) * 1000; // ms
        console.log("⏱️ Duración del clip:", duration / 1000, "segundos");

        // Detener al final (fallback por si timeupdate falla)
        recordingTimeout = setTimeout(() => {
          if (recorder && recorder.state !== "inactive") {
            console.log("⏹️ Deteniendo grabación por timeout (fallback)");
            if (onProgress) onProgress(85, "encoding");
            recorder.stop();
            videoElement.pause();
            if (timeUpdateListener) {
              videoElement.removeEventListener(
                "timeupdate",
                timeUpdateListener
              );
            }
          }
        }, duration + 500); // Agregar 500ms extra como margen

        // Actualizar progreso durante grabación
        if (onProgress) {
          let currentProgress = 20;
          const progressIncrement = 65 / (duration / 1000); // De 20% a 85% en la duración

          progressTimer = setInterval(() => {
            // Verificar cancelación
            if (onCancel && onCancel()) {
              console.log("❌ Cancelando exportación...");
              if (recorder && recorder.state !== "inactive") {
                recorder.stop();
              }
              videoElement.pause();
              if (!recorder || recorder.state === "inactive") {
                restoreAudioState();
                combinedStream
                  ?.getTracks()
                  .forEach((track: MediaStreamTrack) => track.stop());
              }
              clearInterval(progressTimer);
              clearTimeout(recordingTimeout);
              return;
            }

            currentProgress += progressIncrement;
            if (currentProgress <= 85) {
              onProgress(
                Math.min(85, Math.round(currentProgress)),
                "rendering"
              );
            } else {
              clearInterval(progressTimer);
            }
          }, 1000); // Actualizar cada segundo
        }
      };
    } catch (error) {
      console.error("❌ Error configurando MediaRecorder:", error);
      cleanup();
      reject(error);
    }
  });
}

/**
 * Convierte un blob WebM a MP4 usando FFmpeg
 */
export async function convertWebMToMP4(
  webmBlob: Blob,
  onProgress?: (progress: number, stage: string) => void
): Promise<Blob> {
  console.log("🔄 Convirtiendo WebM a MP4...");

  try {
    if (onProgress) onProgress(0, "loading-ffmpeg");
    if (onProgress) onProgress(40, "loading-ffmpeg");

    const ffmpeg = await getFFmpeg();

    if (onProgress) onProgress(60, "encoding");

    // Escribir archivo WebM temporal
    const webmData = await fetchFile(webmBlob);
    await ffmpeg.writeFile("input.webm", webmData);

    if (onProgress) onProgress(70, "encoding");

    console.log("⚙️ Ejecutando conversión WebM -> MP4");

    // Convertir a MP4 con configuración optimizada para VELOCIDAD
    await ffmpeg.exec([
      "-i",
      "input.webm",
      "-c:v",
      "libx264", // Codec H.264
      "-preset",
      "veryfast", // Preset más rápido (ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow)
      "-crf",
      "23", // CRF 23 = buena calidad, más rápido que CRF 18
      "-c:a",
      "aac", // Codec de audio AAC
      "-b:a",
      "128k", // Bitrate de audio reducido (128k en vez de 192k)
      "-movflags",
      "+faststart", // Optimizar para streaming
      "-pix_fmt",
      "yuv420p", // Compatibilidad máxima
      "-threads",
      "0", // Usar todos los threads disponibles
      "output.mp4",
    ]);

    if (onProgress) onProgress(90, "finalizing");

    console.log("✅ Conversión completada, leyendo archivo");
    const data = await ffmpeg.readFile("output.mp4");
    const mp4Blob = new Blob([data as any], { type: "video/mp4" });

    // Limpieza
    await ffmpeg.deleteFile("input.webm");
    await ffmpeg.deleteFile("output.mp4");

    if (onProgress) onProgress(100, "complete");

    console.log("✅ MP4 generado:", mp4Blob.size, "bytes");
    return mp4Blob;
  } catch (error) {
    console.error("❌ Error convirtiendo a MP4:", error);
    throw error;
  }
}

/**
 * Convierte un video WebM a MP4 usando el backend (FFmpeg nativo)
 * Mucho más rápido que la conversión en el navegador (10x más rápido)
 *
 * @param webmBlob - Blob del video WebM
 * @param onProgress - Callback de progreso (progress: 0-100, stage: string)
 * @returns Blob del video MP4
 */
export async function convertWebMToMP4OnBackend(
  webmBlob: Blob,
  onProgress?: (progress: number, stage: string) => void
): Promise<Blob> {
  console.log("🚀 Convirtiendo WebM a MP4 en backend (FFmpeg nativo)");

  try {
    if (onProgress) onProgress(10, "uploading");

    // Crear FormData con el archivo WebM
    const formData = new FormData();
    formData.append("video", webmBlob, "clip.webm");

    if (onProgress) onProgress(30, "converting");

    // Enviar al backend
    const response = await fetch(
      "http://localhost:8080/api/convert-webm-to-mp4",
      {
        method: "POST",
        body: formData,
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Conversion failed");
    }

    if (onProgress) onProgress(80, "downloading");

    // Recibir el MP4 convertido
    const mp4Blob = await response.blob();

    if (onProgress) onProgress(100, "complete");

    console.log("✅ Conversión MP4 completa:", {
      originalSize: `${(webmBlob.size / 1024 / 1024).toFixed(2)} MB`,
      convertedSize: `${(mp4Blob.size / 1024 / 1024).toFixed(2)} MB`,
      type: mp4Blob.type,
    });

    return mp4Blob;
  } catch (error) {
    console.error("❌ Error convirtiendo WebM a MP4:", error);
    throw error;
  }
}

/**
 * Exporta clip con subtítulos usando FFmpeg (fallback para Safari)
 * Usa una estrategia simplificada: extrae el clip original y superpone subtítulos
 */
export async function exportClipWithFFmpeg(
  canvasElement: HTMLCanvasElement,
  videoElement: HTMLVideoElement,
  startTime: number,
  endTime: number,
  onProgress?: (progress: number, stage: string) => void,
  onCancel?: () => boolean
): Promise<Blob> {
  console.log("🎬 Exportando con FFmpeg (Safari fallback)");

  try {
    if (onProgress) onProgress(10, "loading-ffmpeg");

    const ffmpeg = await getFFmpeg();
    const duration = endTime - startTime;

    // Obtener el video original
    const originalVideoUrl = videoElement.src;
    console.log("📥 Descargando video original...");

    if (onProgress) onProgress(20, "downloading");
    const videoData = await fetchFile(originalVideoUrl);
    await ffmpeg.writeFile("input.mp4", videoData);

    // Capturar una imagen del canvas con los subtítulos actuales
    if (onProgress) onProgress(30, "capturing-subtitles");
    console.log("📸 Capturando subtítulos del canvas...");

    // Posicionar video en el tiempo de inicio para capturar subtítulos correctos
    videoElement.currentTime = startTime + duration / 2; // Usar el medio del clip
    await new Promise((resolve) => {
      videoElement.onseeked = () => {
        // Pequeño delay para que el canvas se actualice
        setTimeout(resolve, 100);
      };
    });

    // Capturar el canvas como imagen de subtítulos
    const subtitleBlob = await new Promise<Blob>((resolve) => {
      canvasElement.toBlob((blob) => {
        resolve(blob!);
      }, "image/png");
    });

    const subtitleData = new Uint8Array(await subtitleBlob.arrayBuffer());
    await ffmpeg.writeFile("subtitles.png", subtitleData);

    if (onProgress) onProgress(50, "processing");

    // Extraer el clip y superponer subtítulos
    console.log("✂️ Procesando clip con subtítulos...");
    await ffmpeg.exec([
      "-ss",
      startTime.toString(),
      "-i",
      "input.mp4",
      "-i",
      "subtitles.png",
      "-t",
      duration.toString(),
      "-filter_complex",
      "[0:v]scale=1080:1920[bg]; [bg][1:v]overlay=(W-w)/2:(H-h)/2[v]",
      "-map",
      "[v]",
      "-map",
      "0:a?", // Audio opcional
      "-c:v",
      "libx264",
      "-preset",
      "medium",
      "-crf",
      "23",
      "-c:a",
      "aac",
      "-b:a",
      "128k",
      "-movflags",
      "+faststart",
      "output.mp4",
    ]);

    if (onProgress) onProgress(90, "finalizing");

    // Leer resultado
    const data = await ffmpeg.readFile("output.mp4");
    const resultBlob = new Blob([data as any], { type: "video/mp4" });

    // Limpiar archivos temporales
    try {
      await ffmpeg.deleteFile("input.mp4");
      await ffmpeg.deleteFile("subtitles.png");
      await ffmpeg.deleteFile("output.mp4");
    } catch (cleanupError) {
      console.warn("⚠️ Error limpiando archivos temporales:", cleanupError);
    }

    if (onProgress) onProgress(100, "complete");

    console.log("✅ Exportación FFmpeg completa:", {
      size: `${(resultBlob.size / 1024 / 1024).toFixed(2)} MB`,
      type: resultBlob.type,
      duration: `${duration}s`,
    });

    return resultBlob;
  } catch (error) {
    console.error("❌ Error en exportación FFmpeg:", error);
    throw error;
  }
}

/**
 * Extrae solo el video clip sin subtítulos (para procesamiento posterior)
 * Útil cuando necesitas el clip crudo del backend
 */
export async function extractClipOnly(
  videoUrl: string,
  startTime: number,
  endTime: number,
  onProgress?: (progress: number, stage: string) => void
): Promise<Blob> {
  console.log("✂️ Extrayendo clip sin subtítulos");

  try {
    if (onProgress) onProgress(10, "loading-ffmpeg");

    const ffmpeg = await getFFmpeg();

    if (onProgress) onProgress(20, "downloading-video");

    const videoData = await fetchFile(videoUrl);
    await ffmpeg.writeFile("input.mp4", videoData);

    if (onProgress) onProgress(40, "processing");

    const duration = endTime - startTime;

    // Extraer clip con formato vertical optimizado
    await ffmpeg.exec([
      "-ss",
      startTime.toString(),
      "-i",
      "input.mp4",
      "-t",
      duration.toString(),
      "-vf",
      "scale=-1:1920,crop=min(iw\\,1080):1920", // Vertical 9:16
      "-c:v",
      "libx264",
      "-preset",
      "medium",
      "-crf",
      "18",
      "-c:a",
      "aac",
      "-b:a",
      "192k",
      "-movflags",
      "+faststart",
      "output.mp4",
    ]);

    if (onProgress) onProgress(80, "finalizing");

    const data = await ffmpeg.readFile("output.mp4");
    const blob = new Blob([data as any], { type: "video/mp4" });

    // Limpieza
    await ffmpeg.deleteFile("input.mp4");
    await ffmpeg.deleteFile("output.mp4");

    if (onProgress) onProgress(100, "complete");

    return blob;
  } catch (error) {
    console.error("❌ Error extrayendo clip:", error);
    throw error;
  }
}
