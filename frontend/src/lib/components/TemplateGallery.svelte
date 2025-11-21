<script lang="ts">
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

  let {
    isOpen = $bindable(false),
    templates,
    onSelect,
  }: {
    isOpen: boolean;
    templates: Template[];
    onSelect: (template: Template) => void;
  } = $props();

  function selectTemplate(template: Template) {
    onSelect(template);
    isOpen = false;
  }
</script>

{#if isOpen}
  <!-- Modal Backdrop -->
  <div
    class="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    onclick={() => (isOpen = false)}
  >
    <!-- Modal Content -->
    <div
      class="bg-gray-900 rounded-3xl max-w-7xl w-full max-h-[90vh] overflow-hidden shadow-2xl border border-gray-700"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div
        class="p-6 border-b border-gray-700 flex items-center justify-between sticky top-0 bg-gray-900 z-10"
      >
        <div>
          <h2 class="text-2xl font-bold text-white">Galería de Plantillas</h2>
          <p class="text-sm text-gray-400 mt-1">
            {templates.length} estilos disponibles
          </p>
        </div>
        <button
          onclick={() => (isOpen = false)}
          class="w-10 h-10 rounded-full bg-gray-800 hover:bg-gray-700 flex items-center justify-center transition-colors text-gray-400 hover:text-white"
        >
          ✕
        </button>
      </div>

      <!-- Templates Grid -->
      <div class="p-6 overflow-y-auto max-h-[calc(90vh-100px)]">
        <div
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
        >
          {#each templates as template}
            <button
              onclick={() => selectTemplate(template)}
              class="group bg-gradient-to-br from-gray-800 to-gray-900 rounded-2xl p-5 hover:from-gray-700 hover:to-gray-800 transition-all duration-300 hover:scale-[1.02] border border-gray-700 hover:border-blue-500 hover:shadow-xl hover:shadow-blue-500/20"
            >
              <!-- Emoji & Name -->
              <div class="flex items-center gap-3 mb-4">
                <div
                  class="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-600/20 to-purple-600/20 flex items-center justify-center text-3xl group-hover:scale-110 transition-transform"
                >
                  {template.emoji}
                </div>
                <div class="flex-1 text-left">
                  <h3
                    class="font-bold text-white text-base group-hover:text-blue-400 transition-colors"
                  >
                    {template.name}
                  </h3>
                  <p class="text-xs text-gray-400 mt-0.5">
                    {template.description}
                  </p>
                </div>
              </div>

              <!-- Font Family Badge -->
              <div class="mb-3 flex items-center gap-2">
                <span
                  class="text-xs font-medium text-gray-500 bg-gray-800 px-2 py-1 rounded-md border border-gray-700"
                >
                  {template.settings.fontFamily
                    .split(",")[0]
                    .replace(/['"]/g, "")
                    .trim()}
                </span>
                <span class="text-xs text-gray-600">
                  {template.settings.fontWeight}
                </span>
              </div>

              <!-- Preview - Styled Text -->
              <div
                class="h-28 rounded-xl bg-gradient-to-br from-gray-950 to-black flex items-center justify-center overflow-hidden relative border border-gray-800"
              >
                <!-- Background pattern -->
                <div
                  class="absolute inset-0 opacity-5"
                  style="background-image: radial-gradient(circle, white 1px, transparent 1px); background-size: 20px 20px;"
                ></div>

                <!-- Text Preview -->
                <div class="relative z-10 px-4 text-center">
                  <div
                    class="inline-block px-5 py-2.5 font-semibold text-base transition-all duration-300 group-hover:scale-105"
                    style="
                      font-family: {template.settings.fontFamily};
                      font-weight: {template.settings.fontWeight};
                      background: {template.settings.bgHexColor}{Math.round(
                      template.settings.bgOpacity * 255
                    )
                      .toString(16)
                      .padStart(2, '0')};
                      color: {template.settings.textColor};
                      border-radius: {template.settings.borderRadius}px;
                      box-shadow: 0 4px {template.settings
                      .shadowBlur}px rgba(0,0,0,0.5);
                    "
                  >
                    Sample Text
                  </div>
                  <div
                    class="inline-block px-5 py-2.5 font-semibold text-base mt-2 transition-all duration-300 group-hover:scale-105"
                    style="
                      font-family: {template.settings.fontFamily};
                      font-weight: {template.settings.fontWeight};
                      background: {template.settings.bgHexColor}{Math.round(
                      template.settings.bgOpacity * 255
                    )
                      .toString(16)
                      .padStart(2, '0')};
                      color: {template.settings.activeTextColor};
                      border-radius: {template.settings.borderRadius}px;
                      box-shadow: 0 4px {template.settings
                      .shadowBlur}px rgba(0,0,0,0.5);
                    "
                  >
                    Active
                  </div>
                </div>
              </div>

              <!-- Style Info -->
              <div class="mt-4 flex items-center justify-between text-xs">
                <div class="flex items-center gap-2">
                  <span class="flex items-center gap-1.5 text-gray-500">
                    <div
                      class="w-4 h-4 rounded-full border-2 border-gray-600"
                      style="background: {template.settings.textColor}"
                    ></div>
                    Text
                  </span>
                  <span class="flex items-center gap-1.5 text-gray-500">
                    <div
                      class="w-4 h-4 rounded-full border-2 border-gray-600"
                      style="background: {template.settings.activeTextColor}"
                    ></div>
                    Karaoke
                  </span>
                </div>
                <span class="text-gray-600">
                  {template.settings.transition}
                </span>
              </div>

              <!-- Apply Button Hint -->
              <div class="mt-3 pt-3 border-t border-gray-800">
                <div
                  class="text-xs text-center text-gray-500 group-hover:text-blue-400 transition-colors font-medium"
                >
                  Click para aplicar ✨
                </div>
              </div>
            </button>
          {/each}
        </div>
      </div>
    </div>
  </div>
{/if}
