<script lang="ts">
  import TemplateGallery from "./TemplateGallery.svelte";

  let {
    fontSize = $bindable(20),
    fontFamily = $bindable("Inter"),
    fontWeight = $bindable(600),
    wordsPerLine = $bindable(4),
    transition = $bindable("pop"),
    bgHexColor = $bindable("#000000"),
    bgOpacity = $bindable(0.8),
    textColor = $bindable("#FFFFFF"),
    activeTextColor = $bindable("#22c55e"),
    borderRadius = $bindable(8),
    shadowBlur = $bindable(12),
    position = $bindable("bottom"),
    syncOffset = $bindable(0),
    onSettingsChange,
  }: {
    fontSize: number;
    fontFamily: string;
    fontWeight: number;
    wordsPerLine: number;
    transition: string;
    bgHexColor: string;
    bgOpacity: number;
    textColor: string;
    activeTextColor: string;
    borderRadius: number;
    shadowBlur: number;
    position: string;
    syncOffset: number;
    onSettingsChange: () => void;
  } = $props();

  // Plantillas modernas - NO modifican fontSize 
  type Template = {
    name: string;
    emoji: string;
    description: string;
    settings: {
      fontFamily: string;
      fontWeight: number;
      transition: string;
      bgHexColor: string;
      bgOpacity: number;
      textColor: string;
      activeTextColor: string;
      borderRadius: number;
      shadowBlur: number;
    };
  };

  const templates: Template[] = [
    {
      name: "Apple",
      emoji: "🍎",
      description: "Clean, minimal, ultra-smooth",
      settings: {
        fontFamily: "-apple-system, BlinkMacSystemFont, 'SF Pro Display'",
        fontWeight: 600,
        transition: "fade",
        bgHexColor: "#000000",
        bgOpacity: 0.25,
        textColor: "#FFFFFF",
        activeTextColor: "#007AFF",
        borderRadius: 16,
        shadowBlur: 8,
      },
    },
    {
      name: "Screen.studio",
      emoji: "✨",
      description: "Premium blur, micro-animations",
      settings: {
        fontFamily: "Inter",
        fontWeight: 500,
        transition: "slide",
        bgHexColor: "#1E1E1E",
        bgOpacity: 0.6,
        textColor: "#F5F5F7",
        activeTextColor: "#00D9FF",
        borderRadius: 20,
        shadowBlur: 16,
      },
    },
    {
      name: "Modern",
      emoji: "🎨",
      description: "Vibrant gradients, bold",
      settings: {
        fontFamily: "Poppins",
        fontWeight: 700,
        transition: "bounce",
        bgHexColor: "#6366F1",
        bgOpacity: 0.9,
        textColor: "#FFFFFF",
        activeTextColor: "#FBBF24",
        borderRadius: 12,
        shadowBlur: 20,
      },
    },
    {
      name: "Minimal",
      emoji: "⚪",
      description: "High contrast, no background",
      settings: {
        fontFamily: "Inter",
        fontWeight: 300,
        transition: "fade",
        bgHexColor: "#000000",
        bgOpacity: 0,
        textColor: "#FFFFFF",
        activeTextColor: "#FFFFFF",
        borderRadius: 0,
        shadowBlur: 24,
      },
    },
    {
      name: "Glassmorphism",
      emoji: "💎",
      description: "Frosted glass effect",
      settings: {
        fontFamily: "Poppins",
        fontWeight: 400,
        transition: "slide",
        bgHexColor: "#FFFFFF",
        bgOpacity: 0.15,
        textColor: "#F0F0F0",
        activeTextColor: "#A78BFA",
        borderRadius: 24,
        shadowBlur: 12,
      },
    },
    {
      name: "Neon",
      emoji: "🌈",
      description: "Bright borders, dark bg",
      settings: {
        fontFamily: "Space Grotesk",
        fontWeight: 800,
        transition: "pop",
        bgHexColor: "#000000",
        bgOpacity: 0.85,
        textColor: "#00FF88",
        activeTextColor: "#FF00FF",
        borderRadius: 8,
        shadowBlur: 28,
      },
    },
    {
      name: "Cinematic",
      emoji: "🎬",
      description: "Dark, subtle glow",
      settings: {
        fontFamily: "Playfair Display",
        fontWeight: 400,
        transition: "fade",
        bgHexColor: "#0A0A0A",
        bgOpacity: 0.7,
        textColor: "#E5E5E5",
        activeTextColor: "#FFD700",
        borderRadius: 6,
        shadowBlur: 18,
      },
    },
    {
      name: "Podcast",
      emoji: "🎙️",
      description: "Clean, highly readable",
      settings: {
        fontFamily: "Open Sans",
        fontWeight: 500,
        transition: "fade",
        bgHexColor: "#1F2937",
        bgOpacity: 0.85,
        textColor: "#F3F4F6",
        activeTextColor: "#10B981",
        borderRadius: 16,
        shadowBlur: 12,
      },
    },
    {
      name: "Retro",
      emoji: "👾",
      description: "8-bit style, sharp edges",
      settings: {
        fontFamily: "Courier New",
        fontWeight: 700,
        transition: "pop",
        bgHexColor: "#000000",
        bgOpacity: 1,
        textColor: "#00FF00",
        activeTextColor: "#FFFF00",
        borderRadius: 0,
        shadowBlur: 0,
      },
    },
    {
      name: "Soft",
      emoji: "🌸",
      description: "Pastel, gentle",
      settings: {
        fontFamily: "Raleway",
        fontWeight: 400,
        transition: "fade",
        bgHexColor: "#FDF2F8",
        bgOpacity: 0.85,
        textColor: "#831843",
        activeTextColor: "#EC4899",
        borderRadius: 24,
        shadowBlur: 6,
      },
    },
    // 30 Additional Modern Templates
    {
      name: "Bold Impact",
      emoji: "💥",
      description: "Ultra-bold, high contrast",
      settings: {
        fontFamily: "Bebas Neue",
        fontWeight: 900,
        transition: "zoom",
        bgHexColor: "#FF0000",
        bgOpacity: 0.9,
        textColor: "#FFFFFF",
        activeTextColor: "#FFFF00",
        borderRadius: 4,
        shadowBlur: 30,
      },
    },
    {
      name: "Elegant",
      emoji: "👑",
      description: "Sophisticated serif",
      settings: {
        fontFamily: "Playfair Display",
        fontWeight: 300,
        transition: "fade",
        bgHexColor: "#2C1810",
        bgOpacity: 0.75,
        textColor: "#F5E6D3",
        activeTextColor: "#D4AF37",
        borderRadius: 12,
        shadowBlur: 10,
      },
    },
    {
      name: "Tech Pro",
      emoji: "⚡",
      description: "Futuristic, angular",
      settings: {
        fontFamily: "Space Grotesk",
        fontWeight: 600,
        transition: "rotate",
        bgHexColor: "#0A1929",
        bgOpacity: 0.85,
        textColor: "#00D4FF",
        activeTextColor: "#7C3AED",
        borderRadius: 2,
        shadowBlur: 16,
      },
    },
    {
      name: "Gaming",
      emoji: "🎮",
      description: "Vibrant, energetic",
      settings: {
        fontFamily: "Oswald",
        fontWeight: 800,
        transition: "elastic",
        bgHexColor: "#1A0033",
        bgOpacity: 0.9,
        textColor: "#FF00FF",
        activeTextColor: "#00FFFF",
        borderRadius: 8,
        shadowBlur: 25,
      },
    },
    {
      name: "Corporate",
      emoji: "💼",
      description: "Professional, clean",
      settings: {
        fontFamily: "Roboto",
        fontWeight: 500,
        transition: "slide",
        bgHexColor: "#1E3A8A",
        bgOpacity: 0.8,
        textColor: "#FFFFFF",
        activeTextColor: "#60A5FA",
        borderRadius: 10,
        shadowBlur: 8,
      },
    },
    {
      name: "Luxury",
      emoji: "💎",
      description: "Premium gold accents",
      settings: {
        fontFamily: "Playfair Display",
        fontWeight: 400,
        transition: "fade",
        bgHexColor: "#000000",
        bgOpacity: 0.85,
        textColor: "#FFD700",
        activeTextColor: "#FFF8DC",
        borderRadius: 16,
        shadowBlur: 20,
      },
    },
    {
      name: "Sport",
      emoji: "⚽",
      description: "Dynamic, energetic",
      settings: {
        fontFamily: "Bebas Neue",
        fontWeight: 900,
        transition: "pop",
        bgHexColor: "#DC2626",
        bgOpacity: 0.85,
        textColor: "#FFFFFF",
        activeTextColor: "#FDE047",
        borderRadius: 6,
        shadowBlur: 18,
      },
    },
    {
      name: "Music",
      emoji: "🎵",
      description: "Rhythmic, colorful",
      settings: {
        fontFamily: "Poppins",
        fontWeight: 600,
        transition: "bounce",
        bgHexColor: "#7C3AED",
        bgOpacity: 0.8,
        textColor: "#FFFFFF",
        activeTextColor: "#F472B6",
        borderRadius: 20,
        shadowBlur: 22,
      },
    },
    {
      name: "Nature",
      emoji: "🌿",
      description: "Organic, earthy",
      settings: {
        fontFamily: "Open Sans",
        fontWeight: 400,
        transition: "fade",
        bgHexColor: "#166534",
        bgOpacity: 0.75,
        textColor: "#ECFCCB",
        activeTextColor: "#FDE047",
        borderRadius: 18,
        shadowBlur: 12,
      },
    },
    {
      name: "Ocean",
      emoji: "🌊",
      description: "Cool blues, fluid",
      settings: {
        fontFamily: "Raleway",
        fontWeight: 500,
        transition: "slide",
        bgHexColor: "#0C4A6E",
        bgOpacity: 0.7,
        textColor: "#E0F2FE",
        activeTextColor: "#38BDF8",
        borderRadius: 22,
        shadowBlur: 14,
      },
    },
    {
      name: "Fire",
      emoji: "🔥",
      description: "Hot reds, intense",
      settings: {
        fontFamily: "Oswald",
        fontWeight: 700,
        transition: "elastic",
        bgHexColor: "#7F1D1D",
        bgOpacity: 0.85,
        textColor: "#FEF3C7",
        activeTextColor: "#FB923C",
        borderRadius: 10,
        shadowBlur: 26,
      },
    },
    {
      name: "Midnight",
      emoji: "🌙",
      description: "Deep night theme",
      settings: {
        fontFamily: "Inter",
        fontWeight: 400,
        transition: "fade",
        bgHexColor: "#0F172A",
        bgOpacity: 0.9,
        textColor: "#CBD5E1",
        activeTextColor: "#818CF8",
        borderRadius: 14,
        shadowBlur: 16,
      },
    },
    {
      name: "Sunrise",
      emoji: "🌅",
      description: "Warm dawn colors",
      settings: {
        fontFamily: "Raleway",
        fontWeight: 300,
        transition: "fade",
        bgHexColor: "#FFF7ED",
        bgOpacity: 0.7,
        textColor: "#78350F",
        activeTextColor: "#F59E0B",
        borderRadius: 16,
        shadowBlur: 8,
      },
    },
    {
      name: "Cyber",
      emoji: "🤖",
      description: "Cyberpunk aesthetic",
      settings: {
        fontFamily: "Space Grotesk",
        fontWeight: 700,
        transition: "flip",
        bgHexColor: "#0A0A0A",
        bgOpacity: 0.9,
        textColor: "#39FF14",
        activeTextColor: "#FF10F0",
        borderRadius: 0,
        shadowBlur: 28,
      },
    },
    {
      name: "Vintage",
      emoji: "📻",
      description: "Classic retro feel",
      settings: {
        fontFamily: "Playfair Display",
        fontWeight: 400,
        transition: "fade",
        bgHexColor: "#78350F",
        bgOpacity: 0.85,
        textColor: "#FEF3C7",
        activeTextColor: "#FBBF24",
        borderRadius: 8,
        shadowBlur: 10,
      },
    },
    {
      name: "Business",
      emoji: "📊",
      description: "Professional charts",
      settings: {
        fontFamily: "Roboto",
        fontWeight: 600,
        transition: "slide",
        bgHexColor: "#0F172A",
        bgOpacity: 0.8,
        textColor: "#F1F5F9",
        activeTextColor: "#22D3EE",
        borderRadius: 12,
        shadowBlur: 10,
      },
    },
    {
      name: "Pastel Dream",
      emoji: "🦄",
      description: "Soft rainbow colors",
      settings: {
        fontFamily: "Poppins",
        fontWeight: 400,
        transition: "fade",
        bgHexColor: "#F3E8FF",
        bgOpacity: 0.8,
        textColor: "#6B21A8",
        activeTextColor: "#EC4899",
        borderRadius: 20,
        shadowBlur: 6,
      },
    },
    {
      name: "Monochrome",
      emoji: "⬛",
      description: "Pure B&W contrast",
      settings: {
        fontFamily: "Montserrat",
        fontWeight: 700,
        transition: "fade",
        bgHexColor: "#000000",
        bgOpacity: 0.95,
        textColor: "#FFFFFF",
        activeTextColor: "#D1D5DB",
        borderRadius: 0,
        shadowBlur: 0,
      },
    },
    {
      name: "Gradient Pop",
      emoji: "🎨",
      description: "Colorful transitions",
      settings: {
        fontFamily: "Poppins",
        fontWeight: 600,
        transition: "zoom",
        bgHexColor: "#EC4899",
        bgOpacity: 0.85,
        textColor: "#FFFFFF",
        activeTextColor: "#FDE047",
        borderRadius: 16,
        shadowBlur: 20,
      },
    },
    {
      name: "Comic",
      emoji: "💬",
      description: "Fun, playful bubbles",
      settings: {
        fontFamily: "Bebas Neue",
        fontWeight: 900,
        transition: "pop",
        bgHexColor: "#FFFFFF",
        bgOpacity: 0.95,
        textColor: "#000000",
        activeTextColor: "#EF4444",
        borderRadius: 24,
        shadowBlur: 12,
      },
    },
    {
      name: "Professional",
      emoji: "✅",
      description: "Clean corporate look",
      settings: {
        fontFamily: "Open Sans",
        fontWeight: 500,
        transition: "slide",
        bgHexColor: "#374151",
        bgOpacity: 0.85,
        textColor: "#F9FAFB",
        activeTextColor: "#34D399",
        borderRadius: 10,
        shadowBlur: 10,
      },
    },
    {
      name: "Festival",
      emoji: "🎪",
      description: "Vibrant carnival",
      settings: {
        fontFamily: "Bebas Neue",
        fontWeight: 800,
        transition: "bounce",
        bgHexColor: "#DC2626",
        bgOpacity: 0.9,
        textColor: "#FEF3C7",
        activeTextColor: "#FACC15",
        borderRadius: 14,
        shadowBlur: 24,
      },
    },
    {
      name: "Minimal Pro",
      emoji: "⚫",
      description: "Ultra-clean design",
      settings: {
        fontFamily: "Inter",
        fontWeight: 200,
        transition: "fade",
        bgHexColor: "#000000",
        bgOpacity: 0.1,
        textColor: "#FFFFFF",
        activeTextColor: "#E5E5E5",
        borderRadius: 0,
        shadowBlur: 30,
      },
    },
    {
      name: "Artistic",
      emoji: "🎭",
      description: "Creative expression",
      settings: {
        fontFamily: "Playfair Display",
        fontWeight: 400,
        transition: "rotate",
        bgHexColor: "#4C1D95",
        bgOpacity: 0.75,
        textColor: "#E9D5FF",
        activeTextColor: "#F472B6",
        borderRadius: 18,
        shadowBlur: 16,
      },
    },
    {
      name: "Street",
      emoji: "🏙️",
      description: "Urban graffiti style",
      settings: {
        fontFamily: "Bebas Neue",
        fontWeight: 900,
        transition: "pop",
        bgHexColor: "#1C1917",
        bgOpacity: 0.9,
        textColor: "#FBBF24",
        activeTextColor: "#EF4444",
        borderRadius: 4,
        shadowBlur: 22,
      },
    },
    {
      name: "Sci-Fi",
      emoji: "🚀",
      description: "Futuristic space",
      settings: {
        fontFamily: "Space Grotesk",
        fontWeight: 600,
        transition: "flip",
        bgHexColor: "#1E1B4B",
        bgOpacity: 0.85,
        textColor: "#A5F3FC",
        activeTextColor: "#C084FC",
        borderRadius: 8,
        shadowBlur: 24,
      },
    },
    {
      name: "Fashion",
      emoji: "👗",
      description: "Stylish, elegant",
      settings: {
        fontFamily: "Playfair Display",
        fontWeight: 300,
        transition: "fade",
        bgHexColor: "#1F2937",
        bgOpacity: 0.7,
        textColor: "#FDF2F8",
        activeTextColor: "#F9A8D4",
        borderRadius: 20,
        shadowBlur: 12,
      },
    },
    {
      name: "News",
      emoji: "📰",
      description: "Breaking news style",
      settings: {
        fontFamily: "Oswald",
        fontWeight: 700,
        transition: "slide",
        bgHexColor: "#DC2626",
        bgOpacity: 0.95,
        textColor: "#FFFFFF",
        activeTextColor: "#FDE047",
        borderRadius: 0,
        shadowBlur: 8,
      },
    },
    {
      name: "Zen",
      emoji: "🧘",
      description: "Peaceful, calming",
      settings: {
        fontFamily: "Raleway",
        fontWeight: 300,
        transition: "fade",
        bgHexColor: "#ECFCCB",
        bgOpacity: 0.6,
        textColor: "#365314",
        activeTextColor: "#84CC16",
        borderRadius: 24,
        shadowBlur: 6,
      },
    },
    {
      name: "Horror",
      emoji: "👻",
      description: "Spooky, dark",
      settings: {
        fontFamily: "Playfair Display",
        fontWeight: 700,
        transition: "blur",
        bgHexColor: "#000000",
        bgOpacity: 0.95,
        textColor: "#991B1B",
        activeTextColor: "#DC2626",
        borderRadius: 4,
        shadowBlur: 32,
      },
    },
  ];

  let isGalleryOpen = $state(false);
  let previewTemplate: Template | null = $state(null);

  function applyTemplate(template: Template) {
    // fontSize NO se modifica - el usuario lo controla
    fontFamily = template.settings.fontFamily;
    fontWeight = template.settings.fontWeight;
    transition = template.settings.transition;
    bgHexColor = template.settings.bgHexColor;
    bgOpacity = template.settings.bgOpacity;
    textColor = template.settings.textColor;
    activeTextColor = template.settings.activeTextColor;
    borderRadius = template.settings.borderRadius;
    shadowBlur = template.settings.shadowBlur;
    previewTemplate = null;
    onSettingsChange();
  }

  function previewTemplateSettings(template: Template) {
    previewTemplate = template;
  }

  function clearPreview() {
    previewTemplate = null;
  }
</script>

<!-- Template Gallery Modal -->
<TemplateGallery
  bind:isOpen={isGalleryOpen}
  {templates}
  onSelect={applyTemplate}
/>

<div class="space-y-8">
  <!-- Header -->
  <div>
    <h4 class="text-lg font-semibold text-white mb-1">
      Configuración de Subtítulos
    </h4>
    <p class="text-sm text-gray-400">
      Personaliza el estilo y animación de tus subtítulos
    </p>
  </div>

  <!-- Plantillas Section -->
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h5 class="text-sm font-medium text-gray-300">Plantillas</h5>
      <span class="text-xs text-gray-500">{templates.length} disponibles</span>
    </div>

    <button
      onclick={() => (isGalleryOpen = true)}
      class="w-full px-6 py-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 text-white rounded-xl font-semibold transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:shadow-blue-500/30 active:scale-95 flex items-center justify-center gap-3"
    >
      <span class="text-2xl">🎨</span>
      <span>Abrir Galería de Plantillas</span>
    </button>
  </div>

  <!-- Texto Section -->
  <div class="space-y-5">
    <h5 class="text-sm font-medium text-gray-300 border-b border-gray-700 pb-2">
      Texto
    </h5>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-2">
        <label for="fontSize" class="block text-xs font-medium text-gray-400"
          >Tamaño</label
        >
        <select
          id="fontSize"
          bind:value={fontSize}
          onchange={() => onSettingsChange()}
          class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
        >
          <option value={16}>Muy Pequeño (16px)</option>
          <option value={20}>Pequeño (20px)</option>
          <option value={24}>Mediano (24px)</option>
          <option value={32}>Grande (32px)</option>
          <option value={40}>Muy Grande (40px)</option>
          <option value={48}>Extra Grande (48px)</option>
          <option value={60}>Enorme (60px)</option>
        </select>
      </div>

      <div class="space-y-2">
        <label
          for="wordsPerLine"
          class="block text-xs font-medium text-gray-400">Sincronización</label
        >
        <select
          id="wordsPerLine"
          bind:value={wordsPerLine}
          onchange={() => onSettingsChange()}
          class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
        >
          <option value={1}>🎵 Karaoke (1 palabra) - Efecto sincronizado</option>
          <option value={2}>⚡ Muy Rápido (2 palabras)</option>
          <option value={3}>🔥 Rápido (3 palabras)</option>
          <option value={4}>⭐ Normal (4 palabras) - Recomendado</option>
          <option value={5}>📖 Cómodo (5 palabras)</option>
          <option value={6}>💬 Conversacional (6 palabras)</option>
          <option value={7}>📚 Párrafo (7 palabras)</option>
          <option value={8}>📄 Extenso (8 palabras)</option>
          <option value={9}>📜 Largo (9 palabras)</option>
          <option value={10}>📋 Máximo (10 palabras)</option>
        </select>
        
        {#if wordsPerLine === 1}
          <div class="mt-2 bg-blue-900/20 border border-blue-500/30 rounded-xl p-3">
            <div class="flex items-center gap-2 mb-1">
              <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
              </svg>
              <span class="text-sm font-medium text-blue-400">Modo Karaoke Activo</span>
            </div>
            <p class="text-xs text-blue-300">
              Cada palabra aparecerá sincronizada con el audio para un efecto karaoke perfecto.
            </p>
          </div>
{/if}
      </div>
    </div>

    <div class="space-y-2">
      <label for="fontFamily" class="block text-xs font-medium text-gray-400"
        >Fuente</label
      >
      <select
        id="fontFamily"
        bind:value={fontFamily}
        onchange={() => onSettingsChange()}
        class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
      >
        <optgroup label="Premium - Modernas">
          <option value="Inter">Inter (Screen.studio)</option>
          <option value="Poppins">Poppins (YouTube)</option>
          <option value="Montserrat">Montserrat (Clean)</option>
          <option value="Space Grotesk">Space Grotesk (Tech)</option>
          <option value="Raleway">Raleway (Elegant)</option>
        </optgroup>
        <optgroup label="Premium - Display">
          <option value="Bebas Neue">Bebas Neue (Bold)</option>
          <option value="Oswald">Oswald (Strong)</option>
          <option value="Playfair Display">Playfair Display (Luxury)</option>
        </optgroup>
        <optgroup label="Premium - Sans-Serif">
          <option value="Roboto">Roboto (Google)</option>
          <option value="Open Sans">Open Sans (Readable)</option>
        </optgroup>
        <optgroup label="Sistema - Fallback">
          <option value="-apple-system, BlinkMacSystemFont, 'SF Pro Display'"
            >SF Pro Display (Apple)</option
          >
          <option value="Arial">Arial</option>
          <option value="Helvetica">Helvetica</option>
          <option value="Impact">Impact</option>
          <option value="Georgia">Georgia</option>
          <option value="Verdana">Verdana</option>
        </optgroup>
      </select>
    </div>

    <div class="space-y-2">
      <label for="fontWeight" class="block text-xs font-medium text-gray-400">
        Grosor de Fuente
      </label>
      <select
        id="fontWeight"
        bind:value={fontWeight}
        onchange={() => onSettingsChange()}
        class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
      >
        <option value={100}>Ultra Delgada (Thin)</option>
        <option value={200}>Extra Delgada (Extra Light)</option>
        <option value={300}>Delgada (Light)</option>
        <option value={400}>Normal (Regular)</option>
        <option value={500}>Media (Medium)</option>
        <option value={600}>Semi Negrita (Semibold)</option>
        <option value={700}>Negrita (Bold)</option>
        <option value={800}>Extra Negrita (Extra Bold)</option>
        <option value={900}>Ultra Negrita (Black)</option>
      </select>
    </div>
  </div>

  <!-- Animación Section -->
  <div class="space-y-5">
    <h5 class="text-sm font-medium text-gray-300 border-b border-gray-700 pb-2">
      Animación
    </h5>

    <div class="space-y-2">
      <label for="transition" class="block text-xs font-medium text-gray-400"
        >Transición</label
      >
      <select
        id="transition"
        bind:value={transition}
        onchange={() => onSettingsChange()}
        class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
      >
        <option value="none">Sin Transición</option>
        <option value="pop">Pop - Estilo Apple (Spring sutil)</option>
        <option value="fade">Fade - Estilo Screen.studio (Súper suave)</option>
        <option value="slide">Slide - Deslizamiento fluido</option>
        <option value="bounce">Bounce - Dos rebotes controlados</option>
        <option value="elastic">Elastic - Oscilación suave</option>
        <option value="rotate">Rotate - Rotación sutil (10°)</option>
      </select>
    </div>
  </div>

  <!-- Colores Section -->
  <div class="space-y-5">
    <h5 class="text-sm font-medium text-gray-300 border-b border-gray-700 pb-2">
      Colores
    </h5>

    <div class="space-y-4">
      <!-- Background -->
      <div class="space-y-3">
        <label for="bgColor" class="block text-xs font-medium text-gray-400"
          >Fondo</label
        >
        <div class="flex gap-3">
          <div class="relative group">
            <input
              id="bgColor"
              type="color"
              bind:value={bgHexColor}
              onchange={() => onSettingsChange()}
              class="w-16 h-16 rounded-xl border-2 border-gray-700 cursor-pointer transition-all hover:scale-105"
            />
            <div
              class="absolute inset-0 rounded-xl border-2 border-blue-500 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"
            ></div>
          </div>
          <div class="flex-1 space-y-2">
            <div class="flex justify-between items-center">
              <span class="text-xs text-gray-400">Opacidad</span>
              <span class="text-xs text-blue-400 font-medium"
                >{Math.round(bgOpacity * 100)}%</span
              >
            </div>
            <div class="relative">
              <input
                type="range"
                bind:value={bgOpacity}
                onchange={() => onSettingsChange()}
                min="0"
                max="1"
                step="0.05"
                class="w-full h-2 rounded-full appearance-none cursor-pointer slider-opacity"
                style="background: linear-gradient(to right, 
                  rgb(59, 130, 246) 0%, 
                  rgb(59, 130, 246) {bgOpacity * 100}%, 
                  rgb(31, 41, 55) {bgOpacity * 100}%, 
                  rgb(31, 41, 55) 100%);"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Text Color -->
      <div class="space-y-3">
        <label for="textColor" class="block text-xs font-medium text-gray-400"
          >Color de Texto</label
        >
        <div class="relative group">
          <input
            id="textColor"
            type="color"
            bind:value={textColor}
            onchange={() => onSettingsChange()}
            class="w-full h-14 rounded-xl border-2 border-gray-700 cursor-pointer transition-all hover:scale-[1.02]"
          />
          <div
            class="absolute inset-0 rounded-xl border-2 border-blue-500 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"
          ></div>
        </div>
      </div>

      <!-- Active Text Color -->
      <div class="space-y-3">
        <label
          for="activeTextColor"
          class="block text-xs font-medium text-gray-400"
          >Color Activo (Karaoke)</label
        >
        <div class="relative group">
          <input
            id="activeTextColor"
            type="color"
            bind:value={activeTextColor}
            onchange={() => onSettingsChange()}
            class="w-full h-14 rounded-xl border-2 border-gray-700 cursor-pointer transition-all hover:scale-[1.02]"
          />
          <div
            class="absolute inset-0 rounded-xl border-2 border-blue-500 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"
          ></div>
        </div>
      </div>
    </div>
  </div>

  <!-- Estilo Section -->
  <div class="space-y-5">
    <h5 class="text-sm font-medium text-gray-300 border-b border-gray-700 pb-2">
      Estilo del Fondo
    </h5>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-3">
        <label
          for="borderRadius"
          class="block text-xs font-medium text-gray-400">Bordes</label
        >
        <select
          id="borderRadius"
          bind:value={borderRadius}
          onchange={() => onSettingsChange()}
          class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
        >
          <option value={0}>Sin Bordes (Cuadrado)</option>
          <option value={4}>Muy Suaves (4px)</option>
          <option value={8}>Suaves (8px)</option>
          <option value={12}>Medianos (12px)</option>
          <option value={16}>Redondeados (16px)</option>
          <option value={20}>Muy Redondeados (20px)</option>
          <option value={24}>Extra Redondeados (24px)</option>
          <option value={32}>Píldora Suave (32px)</option>
          <option value={50}>Píldora Completa (50px)</option>
        </select>
      </div>

      <div class="space-y-3">
        <label for="shadowBlur" class="block text-xs font-medium text-gray-400"
          >Sombra</label
        >
        <select
          id="shadowBlur"
          bind:value={shadowBlur}
          onchange={() => onSettingsChange()}
          class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
        >
          <option value={0}>Sin Sombra</option>
          <option value={4}>Muy Sutil (4px)</option>
          <option value={8}>Sutil (8px)</option>
          <option value={12}>Ligera (12px)</option>
          <option value={16}>Media (16px)</option>
          <option value={20}>Notable (20px)</option>
          <option value={24}>Intensa (24px)</option>
          <option value={28}>Muy Intensa (28px)</option>
          <option value={32}>Dramática (32px)</option>
        </select>
      </div>
    </div>
  </div>

  <!-- Posición Section -->
  <div class="space-y-5">
    <h5 class="text-sm font-medium text-gray-300 border-b border-gray-700 pb-2">
      Posición
    </h5>

    <!-- Posición del texto -->
    <div class="space-y-2">
      <label for="position" class="block text-xs font-medium text-gray-400">
        Posición de Texto
      </label>
      <select
        id="position"
        bind:value={position}
        onchange={() => onSettingsChange()}
        class="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-xl text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
      >
        <option value="top">Arriba</option>
        <option value="center">Centro</option>
        <option value="bottom">Abajo</option>
      </select>
      <p class="text-xs text-gray-500">
        Puedes arrastrar el texto en el canvas para posicionarlo libremente
      </p>
    </div>
  </div>
</div>

<style>
  /* Estilos para el slider de opacidad */
  .slider-opacity::-webkit-slider-thumb {
    appearance: none;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: rgb(59, 130, 246);
    cursor: pointer;
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.5);
    transition: all 0.2s ease;
  }

  .slider-opacity::-webkit-slider-thumb:hover {
    transform: scale(1.2);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.7);
  }

  .slider-opacity::-moz-range-thumb {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: rgb(59, 130, 246);
    cursor: pointer;
    border: none;
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.5);
    transition: all 0.2s ease;
  }

  .slider-opacity::-moz-range-thumb:hover {
    transform: scale(1.2);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.7);
  }

  /* Estilos para el slider de sincronización */
  .slider::-webkit-slider-thumb {
    appearance: none;
    height: 20px;
    width: 20px;
    border-radius: 50%;
    background: #3b82f6;
    cursor: pointer;
    box-shadow: 0 0 2px 0 #555;
    transition: background .15s ease-in-out;
  }

  .slider::-webkit-slider-thumb:hover {
    background: #2563eb;
  }

  .slider::-moz-range-thumb {
    height: 20px;
    width: 20px;
    border-radius: 50%;
    background: #3b82f6;
    cursor: pointer;
    border: none;
    box-shadow: 0 0 2px 0 #555;
  }
</style>
