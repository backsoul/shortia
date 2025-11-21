<script lang="ts">
  import { onMount } from "svelte";
  import type { SubtitleConfig } from "$lib/types";

  let {
    currentSubtitle,
    width = 405,
    height = 720,
    transition = "pop",
    onPositionChange,
    currentTime = 0, // Tiempo actual del video para efecto karaoke
    bgHexColor = "#000000",
    bgOpacity = 0.8,
    textColor = "#FFFFFF",
    activeTextColor = "#22c55e",
    borderRadius = 8,
    shadowBlur = 12,
    position = "bottom",
    fontSize = 20,
    fontFamily = "Inter",
    fontWeight = 600,
  }: {
    currentSubtitle: SubtitleConfig | null;
    width?: number;
    height?: number;
    transition?:
      | "none"
      | "pop"
      | "fade"
      | "slide"
      | "bounce"
      | "zoom"
      | "blur"
      | "scale"
      | "rotate"
      | "flip"
      | "elastic"
      | "spring";
    onPositionChange: (position: "top" | "center" | "bottom") => void;
    currentTime?: number;
    bgHexColor?: string;
    bgOpacity?: number;
    textColor?: string;
    activeTextColor?: string;
    borderRadius?: number;
    shadowBlur?: number;
    position?: string;
    fontSize?: number;
    fontFamily?: string;
    fontWeight?: number;
  } = $props();

  let canvasElement: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D | null = null;
  let customTopPosition: number | null = null; // Posición Y personalizada
  let animationFrameId: number | null = null;

  // Estado para drag & drop
  let isDragging = $state(false);
  let dragStartY = 0;
  let dragStartTop = 0;

  // Estado de la animación actual
  let currentAnimation: {
    startTime: number;
    duration: number;
    fromOpacity: number;
    toOpacity: number;
    fromScale: number;
    toScale: number;
    fromTop: number;
    toTop: number;
    easing: (t: number) => number;
  } | null = null;

  // Estado del subtítulo actual
  let subtitleState = {
    text: "",
    words: [] as string[], // Array de palabras
    opacity: 1,
    scale: 1,
    top: 0,
    blur: 0, // NUEVO: para efecto blur
    rotation: 0, // NUEVO: para efecto rotate
    fontSize: 56,
    fontFamily: "Arial",
    color: "#FFFFFF",
    bgColor: "rgba(0,0,0,0.8)",
    bold: false,
    startTime: 0,
    endTime: 0,
  };

  // Funciones de easing mejoradas estilo Apple - transiciones ultra suaves
  const easingFunctions = {
    // Easing principal de Apple - smooth y natural
    easeOutApple: (t: number) => {
      // Bezier curve similar a ease-out de Apple (0.25, 0.1, 0.25, 1)
      return 1 - Math.pow(1 - t, 3);
    },
    easeInOutApple: (t: number) => {
      // Bezier curve similar a ease-in-out de Apple (0.42, 0, 0.58, 1)
      return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
    },
    easeOutBack: (t: number) => {
      // Mucho más suave - casi imperceptible bounce
      const c1 = 0.8; // Muy reducido para efecto super sutil
      const c3 = c1 + 1;
      return 1 + c3 * Math.pow(t - 1, 3) + c1 * Math.pow(t - 1, 2);
    },
    easeInOutQuad: (t: number) => {
      // Más suave tipo Apple
      return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
    },
    easeOutCubic: (t: number) => {
      // Suavizado máximo
      return 1 - Math.pow(1 - t, 3);
    },
    easeOutBounce: (t: number) => {
      // Bounce muy sutil estilo Apple
      const n1 = 4.0; // Muy reducido
      const d1 = 2.75;
      if (t < 1 / d1) {
        return n1 * t * t;
      } else if (t < 2 / d1) {
        return n1 * (t -= 1.5 / d1) * t + 0.75;
      } else {
        return n1 * (t -= 2.25 / d1) * t + 0.9375;
      }
    },
    // NUEVAS FUNCIONES DE EASING ULTRA SMOOTH
    easeOutExpo: (t: number) => {
      // Exponencial suave - Screen.studio style
      return t === 1 ? 1 : 1 - Math.pow(2, -10 * t);
    },
    easeOutElastic: (t: number) => {
      // Elastic super sutil y premium
      const c4 = (2 * Math.PI) / 3;
      return t === 0
        ? 0
        : t === 1
          ? 1
          : Math.pow(2, -8 * t) * Math.sin((t * 10 - 0.75) * c4) + 1;
    },
    easeInOutElastic: (t: number) => {
      // Elastic bilateral premium
      const c5 = (2 * Math.PI) / 4.5;
      return t === 0
        ? 0
        : t === 1
          ? 1
          : t < 0.5
            ? -(Math.pow(2, 20 * t - 10) * Math.sin((20 * t - 11.125) * c5)) / 2
            : (Math.pow(2, -20 * t + 10) * Math.sin((20 * t - 11.125) * c5)) /
                2 +
              1;
    },
    easeOutQuart: (t: number) => {
      // Ultra smooth - para blur y scale
      return 1 - Math.pow(1 - t, 4);
    },
    easeInOutQuart: (t: number) => {
      // Extra smooth bilateral
      return t < 0.5 ? 8 * t * t * t * t : 1 - Math.pow(-2 * t + 2, 4) / 2;
    },
    easeOutCirc: (t: number) => {
      // Circular suave - muy natural
      return Math.sqrt(1 - Math.pow(t - 1, 2));
    },
    easeInOutBack: (t: number) => {
      // Back bilateral premium
      const c1 = 1.70158;
      const c2 = c1 * 1.525;
      return t < 0.5
        ? (Math.pow(2 * t, 2) * ((c2 + 1) * 2 * t - c2)) / 2
        : (Math.pow(2 * t - 2, 2) * ((c2 + 1) * (t * 2 - 2) + c2) + 2) / 2;
    },
    linear: (t: number) => t,
  };

  onMount(() => {
    console.log("🎨 [SubtitleCanvas] Inicializando canvas nativo...");
    ctx = canvasElement.getContext("2d");
    if (ctx) {
      console.log("✅ [SubtitleCanvas] Canvas 2D inicializado:", {
        width,
        height,
      });
    } else {
      console.error("❌ [SubtitleCanvas] Error obteniendo contexto 2D");
    }
  });

  // Calcular posición basada en preset
  function calculateTopPosition(pos: string): number {
    const canvasHeight = height || 720;
    const MARGIN = 40; // Fixed margin for positioning

    if (pos === "top") {
      return MARGIN;
    } else if (pos === "center") {
      return canvasHeight / 2;
    } else {
      // bottom - usar el margin configurado
      return canvasHeight - MARGIN;
    }
  }

  // Función de renderizado con efecto Screen.studio + Karaoke
  function render() {
    if (!ctx) return;

    // Limpiar canvas
    ctx.clearRect(0, 0, width, height);

    if (!subtitleState.text || subtitleState.words.length === 0) return;

    // Aplicar transformaciones
    ctx.save();

    // Efecto blur dinámico (puede venir de transición o de estado)
    const transitionBlur = (1 - subtitleState.opacity) * 8;
    const totalBlur = Math.max(subtitleState.blur, transitionBlur);
    if (totalBlur > 0.5) {
      ctx.filter = `blur(${totalBlur}px)`;
    }

    ctx.globalAlpha = subtitleState.opacity;
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";

    // Aplicar escala con efecto de "bounce in"
    const centerX = width / 2;
    const centerY = subtitleState.top;
    ctx.translate(centerX, centerY);

    // Rotación si existe
    if (subtitleState.rotation !== 0) {
      ctx.rotate((subtitleState.rotation * Math.PI) / 180);
    }

    // Efecto Screen.studio: escala ligeramente mayor al inicio
    const screenStudioScale =
      subtitleState.scale * (1 + (1 - subtitleState.opacity) * 0.15);
    ctx.scale(screenStudioScale, screenStudioScale);

    ctx.translate(-centerX, -centerY);

    // Configurar fuente
    const fontSize = subtitleState.fontSize;
    const font = `${fontWeight} ${fontSize}px ${fontFamily}`;
    ctx.font = font;

    // Calcular progreso karaoke (qué palabra está activa)
    const duration = subtitleState.endTime - subtitleState.startTime;
    const elapsed = currentTime - subtitleState.startTime;
    const progress =
      duration > 0 ? Math.min(Math.max(elapsed / duration, 0), 1) : 0;
    const totalWords = subtitleState.words.length;
    const currentWordIndex = Math.floor(progress * totalWords);

    // Medir el texto completo para el fondo
    const textMetrics = ctx.measureText(subtitleState.text);
    const textWidth = textMetrics.width;
    const padding = 12;

    // Fondo con bordes redondeados estilo Screen.studio
    const bgX = centerX - textWidth / 2 - padding;
    const bgY = subtitleState.top - fontSize / 2 - padding;
    const bgWidth = textWidth + padding * 2;
    const bgHeight = fontSize + padding * 2;

    // Dibujar fondo con sombra personalizable
    ctx.shadowColor = "rgba(0, 0, 0, 0.3)";
    ctx.shadowBlur = shadowBlur;
    ctx.shadowOffsetY = 4;

    // Convertir hex + opacity a rgba
    const rgb = parseInt(bgHexColor.slice(1), 16);
    const r = (rgb >> 16) & 255;
    const g = (rgb >> 8) & 255;
    const b = rgb & 255;
    ctx.fillStyle = `rgba(${r}, ${g}, ${b}, ${bgOpacity})`;

    ctx.beginPath();
    ctx.roundRect(bgX, bgY, bgWidth, bgHeight, borderRadius);
    ctx.fill();

    // Reset shadow para el texto
    ctx.shadowColor = "transparent";
    ctx.shadowBlur = 0;
    ctx.shadowOffsetY = 0;

    // Dibujar texto palabra por palabra con efecto karaoke
    ctx.shadowColor = "rgba(0, 0, 0, 0.5)";
    ctx.shadowBlur = 4;
    ctx.shadowOffsetY = 1;
    ctx.textAlign = "left";

    // Calcular posición inicial del texto
    let xOffset = centerX - textWidth / 2;
    const yPosition = subtitleState.top;

    subtitleState.words.forEach((word, index) => {
      if (!ctx) return; // Guard clause

      // Determinar color: verde si es la palabra actual o anterior, blanco si no
      if (index <= currentWordIndex) {
        ctx.fillStyle = activeTextColor; // Color activo (karaoke)
      } else {
        ctx.fillStyle = textColor; // Color normal
      }

      // Dibujar la palabra
      ctx.fillText(word, xOffset, yPosition);

      // Medir el ancho de la palabra para la siguiente
      const wordMetrics = ctx.measureText(word + " ");
      xOffset += wordMetrics.width;
    });

    ctx.restore();
  }

  // Animar propiedades
  function animate() {
    if (!currentAnimation) {
      render();
      return;
    }

    const now = Date.now();
    const elapsed = now - currentAnimation.startTime;
    const progress = Math.min(elapsed / currentAnimation.duration, 1);
    const easedProgress = currentAnimation.easing(progress);

    // Interpolar valores
    subtitleState.opacity =
      currentAnimation.fromOpacity +
      (currentAnimation.toOpacity - currentAnimation.fromOpacity) *
        easedProgress;
    subtitleState.scale =
      currentAnimation.fromScale +
      (currentAnimation.toScale - currentAnimation.fromScale) * easedProgress;
    subtitleState.top =
      currentAnimation.fromTop +
      (currentAnimation.toTop - currentAnimation.fromTop) * easedProgress;

    // Interpolar blur (si está activo)
    if (subtitleState.blur > 0) {
      subtitleState.blur = subtitleState.blur * (1 - easedProgress);
      if (subtitleState.blur < 0.5) subtitleState.blur = 0;
    }

    // Interpolar rotation (si está activo)
    if (subtitleState.rotation !== 0) {
      subtitleState.rotation = subtitleState.rotation * (1 - easedProgress);
      if (Math.abs(subtitleState.rotation) < 0.5) subtitleState.rotation = 0;
    }

    render();

    if (progress < 1) {
      animationFrameId = requestAnimationFrame(animate);
    } else {
      // Animación completada - limpiar efectos
      currentAnimation = null;
      subtitleState.blur = 0;
      subtitleState.rotation = 0;
      console.log("✅ [SubtitleCanvas] Animación completada");
    }
  }

  // Aplicar transición
  function applyTransition(transitionType: string, finalTop: number) {
    // Cancelar animación anterior
    if (animationFrameId !== null) {
      cancelAnimationFrame(animationFrameId);
      animationFrameId = null;
    }

    console.log("🎭 [SubtitleCanvas] Aplicando transición:", transitionType);

    let fromOpacity = subtitleState.opacity;
    let fromScale = subtitleState.scale;
    let fromTop = subtitleState.top;
    let fromBlur = 0;
    let fromRotation = 0;

    switch (transitionType) {
      case "pop":
        // Estilo Apple: fade + scale ultra suave (MEJORADO)
        fromOpacity = 0;
        fromScale = 0.94; // Más dramático pero smooth
        fromTop = finalTop;
        currentAnimation = {
          startTime: Date.now(),
          duration: 650, // Más rápido y snappy
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutQuart, // Más suave
        };
        break;

      case "fade":
        // Estilo Apple: fade puro super suave (MEJORADO)
        fromOpacity = 0;
        fromScale = 0.99; // Casi imperceptible
        fromTop = finalTop;
        currentAnimation = {
          startTime: Date.now(),
          duration: 700, // Más suave
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutCubic,
        };
        break;

      case "slide":
        // Estilo Apple: slide sutil y elegante (MEJORADO)
        fromOpacity = 0;
        fromScale = 0.98;
        fromTop = finalTop + 30; // Más movimiento
        currentAnimation = {
          startTime: Date.now(),
          duration: 700,
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutExpo, // NUEVO easing smooth
        };
        break;

      case "bounce":
        // Estilo Apple: bounce casi imperceptible (MEJORADO)
        fromOpacity = 0;
        fromScale = 0.95;
        fromTop = finalTop - 15; // Bounce sutil
        currentAnimation = {
          startTime: Date.now(),
          duration: 800,
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutBack, // Back suave
        };
        break;

      case "zoom":
        // Estilo Apple: zoom suave y elegante (MEJORADO)
        fromOpacity = 0;
        fromScale = 0.85; // Más dramático
        fromTop = finalTop;
        currentAnimation = {
          startTime: Date.now(),
          duration: 700,
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutQuart, // Extra smooth
        };
        break;

      case "blur":
        // NUEVO: Efecto blur Screen.studio style
        fromOpacity = 0;
        fromScale = 1;
        fromTop = finalTop;
        fromBlur = 20; // Blur inicial
        subtitleState.blur = fromBlur;
        currentAnimation = {
          startTime: Date.now(),
          duration: 800, // Lento para efecto premium
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutExpo, // Exponencial suave
        };
        break;

      case "scale":
        // NUEVO: Scale puro ultra smooth
        fromOpacity = 0;
        fromScale = 0.7; // Muy dramático
        fromTop = finalTop;
        currentAnimation = {
          startTime: Date.now(),
          duration: 750,
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutCirc, // Circular smooth
        };
        break;

      case "rotate":
        // NUEVO: Rotación elegante
        fromOpacity = 0;
        fromScale = 0.9;
        fromTop = finalTop;
        fromRotation = -10; // Rotación inicial
        subtitleState.rotation = fromRotation;
        currentAnimation = {
          startTime: Date.now(),
          duration: 800,
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutQuart,
        };
        break;

      case "flip":
        // NUEVO: Flip 3D style
        fromOpacity = 0;
        fromScale = 0.8;
        fromTop = finalTop;
        fromRotation = 90; // Rotación inicial 90 grados
        subtitleState.rotation = fromRotation;
        currentAnimation = {
          startTime: Date.now(),
          duration: 850,
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutBack, // Back para efecto flip
        };
        break;

      case "elastic":
        // NUEVO: Elastic premium ultra smooth
        fromOpacity = 0;
        fromScale = 0.88;
        fromTop = finalTop - 20;
        currentAnimation = {
          startTime: Date.now(),
          duration: 900, // Más largo para elastic
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeOutElastic, // Elastic suave
        };
        break;

      case "spring":
        // NUEVO: Spring bounce natural
        fromOpacity = 0;
        fromScale = 0.92;
        fromTop = finalTop + 25; // Desde abajo
        currentAnimation = {
          startTime: Date.now(),
          duration: 850,
          fromOpacity,
          toOpacity: 1,
          fromScale,
          toScale: 1,
          fromTop,
          toTop: finalTop,
          easing: easingFunctions.easeInOutBack, // Back bilateral
        };
        break;

      default: // "none"
        subtitleState.opacity = 1;
        subtitleState.scale = 1;
        subtitleState.top = finalTop;
        subtitleState.blur = 0;
        subtitleState.rotation = 0;
        render();
        return;
    }

    // Establecer estado inicial
    subtitleState.opacity = fromOpacity;
    subtitleState.scale = fromScale;
    subtitleState.top = fromTop;

    // Iniciar animación
    animationFrameId = requestAnimationFrame(animate);
  }

  $effect(() => {
    console.log("🔄 [SubtitleCanvas] $effect ejecutado:", {
      hasCtx: !!ctx,
      currentSubtitle: currentSubtitle?.text || "null",
      transition,
    });

    if (!ctx) {
      console.log("⏳ [SubtitleCanvas] Canvas no listo aún");
      return;
    }

    if (!currentSubtitle) {
      console.log("❌ [SubtitleCanvas] No hay subtítulo actual");
      subtitleState.text = "";
      render();
      return;
    }

    console.log("🎬 [SubtitleCanvas] Renderizando subtítulo:", {
      text: currentSubtitle.text,
      position: currentSubtitle.position,
      fontSize: currentSubtitle.font_size,
      fontFamily: currentSubtitle.font_family,
      customTopPosition,
    });

    // Actualizar estado del subtítulo
    subtitleState.text = currentSubtitle.text;
    subtitleState.words = currentSubtitle.text.split(" "); // Separar en palabras
    subtitleState.fontSize = currentSubtitle.font_size || 56;
    subtitleState.fontFamily = currentSubtitle.font_family || "Arial";
    subtitleState.color = currentSubtitle.color || "#FFFFFF";
    subtitleState.bgColor = currentSubtitle.bg_color || "rgba(0,0,0,0.8)";
    subtitleState.bold = currentSubtitle.bold || false;
    subtitleState.startTime = currentSubtitle.start_time;
    subtitleState.endTime = currentSubtitle.end_time;

    // Calcular posición final
    const finalTop =
      customTopPosition !== null
        ? customTopPosition
        : calculateTopPosition(currentSubtitle.position);

    console.log("📍 [SubtitleCanvas] Posición calculada:", {
      topPosition: finalTop,
      canvasHeight: height,
      canvasWidth: width,
      position: currentSubtitle.position,
    });

    // Aplicar transición
    applyTransition(transition, finalTop);

    console.log("✅ [SubtitleCanvas] Subtítulo configurado exitosamente");
  });

  // Efecto para re-renderizar continuamente y actualizar el karaoke
  $effect(() => {
    // Dependemos de currentTime para re-renderizar cuando cambia
    if (ctx && subtitleState.text && currentTime) {
      render();
    }
  });

  // Funciones para drag & drop del texto
  function handleMouseDown(e: MouseEvent) {
    const rect = canvasElement.getBoundingClientRect();
    const mouseY = e.clientY - rect.top;

    // Verificar si el click está cerca del texto
    const textY =
      customTopPosition !== null ? customTopPosition : subtitleState.top;
    const distance = Math.abs(mouseY - textY);

    // Si está dentro de un rango razonable (50px), iniciar drag
    if (distance < 50) {
      isDragging = true;
      dragStartY = mouseY;
      dragStartTop = textY;
      canvasElement.style.cursor = "grabbing";
    }
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isDragging) {
      // Cambiar cursor cuando está sobre el texto
      const rect = canvasElement.getBoundingClientRect();
      const mouseY = e.clientY - rect.top;
      const textY =
        customTopPosition !== null ? customTopPosition : subtitleState.top;
      const distance = Math.abs(mouseY - textY);

      canvasElement.style.cursor = distance < 50 ? "grab" : "default";
      return;
    }

    const rect = canvasElement.getBoundingClientRect();
    const mouseY = e.clientY - rect.top;
    const deltaY = mouseY - dragStartY;

    // Actualizar posición personalizada
    const newTop = Math.max(20, Math.min(height - 20, dragStartTop + deltaY));
    customTopPosition = newTop;
    subtitleState.top = newTop;
    render();
  }

  function handleMouseUp() {
    if (isDragging) {
      isDragging = false;
      canvasElement.style.cursor = "grab";
      console.log(
        "📍 [SubtitleCanvas] Nueva posición personalizada:",
        customTopPosition
      );
    }
  }

  function handleMouseLeave() {
    if (isDragging) {
      isDragging = false;
      canvasElement.style.cursor = "default";
    }
  }
</script>

<canvas
  bind:this={canvasElement}
  {width}
  {height}
  class="absolute left-0 top-0"
  style="z-index: 10; pointer-events: auto; user-select: none;"
  onmousedown={handleMouseDown}
  onmousemove={handleMouseMove}
  onmouseup={handleMouseUp}
  onmouseleave={handleMouseLeave}
></canvas>
