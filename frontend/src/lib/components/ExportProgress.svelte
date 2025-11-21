<script lang="ts">
  let {
    isVisible = false,
    progress = 0,
    stage = "preparing",
    onCancel,
  }: {
    isVisible?: boolean;
    progress?: number;
    stage?:
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
      | "processing";
    onCancel?: () => void;
  } = $props();

  function handleCancel() {
    if (onCancel && stage !== "complete" && stage !== "error") {
      if (confirm("¿Estás seguro de que quieres cancelar la exportación?")) {
        onCancel();
      }
    }
  }

  function handleClose() {
    if (onCancel) {
      onCancel();
    }
  }
</script>

{#if isVisible}
  <div
    class="fixed inset-0 bg-black bg-opacity-60 flex items-center justify-center z-50 backdrop-blur-sm"
  >
    <div
      class="bg-gray-900 rounded-2xl shadow-2xl p-8 max-w-md w-full mx-4 border border-gray-800"
    >
      <!-- Header -->
      <div class="text-center mb-8">
        {#if stage === "complete"}
          <div
            class="inline-flex items-center justify-center w-16 h-16 bg-green-500 rounded-full mb-4"
          >
            <svg
              class="w-8 h-8 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <h2 class="text-2xl font-bold text-white mb-2">¡Listo!</h2>
          <p class="text-gray-400 text-sm">
            Tu video se ha descargado correctamente
          </p>
        {:else if stage === "error"}
          <div
            class="inline-flex items-center justify-center w-16 h-16 bg-red-500 rounded-full mb-4"
          >
            <svg
              class="w-8 h-8 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <h2 class="text-2xl font-bold text-white mb-2">Error</h2>
          <p class="text-gray-400 text-sm">
            Ocurrió un error durante la exportación
          </p>
        {:else}
          <div
            class="inline-flex items-center justify-center w-16 h-16 bg-blue-500 rounded-full mb-4"
          >
            <svg
              class="w-8 h-8 text-white animate-spin"
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
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          </div>
          <h2 class="text-2xl font-bold text-white mb-2">Exportando video</h2>
          <p class="text-gray-400 text-sm">Por favor espera...</p>
          {#if stage === "loading-ffmpeg"}
            <p class="text-xs text-blue-300 mt-2">
              Descargando FFmpeg (solo la primera vez puede tardar unos
              segundos)
            </p>
          {:else if stage === "uploading"}
            <p class="text-xs text-blue-300 mt-2">
              Subiendo video al servidor...
            </p>
          {:else if stage === "converting"}
            <p class="text-xs text-blue-300 mt-2">
              Convirtiendo a MP4 (FFmpeg nativo 10x más rápido)
            </p>
          {:else if stage === "downloading"}
            <p class="text-xs text-blue-300 mt-2">
              Descargando video convertido...
            </p>
          {:else if stage === "safari-fallback"}
            <p class="text-xs text-orange-300 mt-2 flex items-center gap-1">
              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10 2C5.588 2 2 5.588 2 10s3.588 8 8 8 8-3.588 8-8-3.588-8-8-8zm0 14c-3.31 0-6-2.69-6-6s2.69-6 6-6 6 2.69 6 6-2.69 6-6 6z"/>
                <path d="M6 10c0 2.21 1.79 4 4 4s4-1.79 4-4-1.79-4-4-4-4 1.79-4 4z"/>
              </svg>
              Usando método compatible con Safari...
            </p>
          {:else if stage === "capturing-subtitles"}
            <p class="text-xs text-blue-300 mt-2 flex items-center gap-1">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"/>
              </svg>
              Capturando subtítulos...
            </p>
          {:else if stage === "processing"}
            <p class="text-xs text-blue-300 mt-2 flex items-center gap-1">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
              </svg>
              Procesando video con subtítulos...
            </p>
          {/if}
        {/if}
      </div>

      <!-- Progress Bar -->
      {#if stage !== "complete" && stage !== "error"}
        <div class="mb-8">
          <div class="flex justify-between text-sm text-gray-400 mb-3">
            <span>Progreso</span>
            <span class="font-semibold text-white">{progress}%</span>
          </div>
          <div class="w-full bg-gray-800 rounded-full h-2.5 overflow-hidden">
            <div
              class="bg-gradient-to-r from-blue-500 to-purple-500 h-2.5 rounded-full transition-all duration-300 ease-out relative"
              style="width: {progress}%"
            >
              <div
                class="absolute inset-0 bg-gradient-to-r from-transparent via-white to-transparent opacity-20 animate-shimmer"
              ></div>
            </div>
          </div>
        </div>
      {/if}

      <!-- Actions -->
      {#if stage === "complete"}
        <button
          onclick={handleClose}
          class="w-full bg-green-600 hover:bg-green-700 text-white font-semibold py-3.5 px-6 rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl hover:scale-[1.02] active:scale-95"
        >
          Cerrar
        </button>
      {:else if stage === "error"}
        <button
          onclick={handleClose}
          class="w-full bg-red-600 hover:bg-red-700 text-white font-semibold py-3.5 px-6 rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl hover:scale-[1.02] active:scale-95"
        >
          Cerrar
        </button>
      {:else if onCancel}
        <button
          onclick={handleCancel}
          class="w-full bg-gray-800 hover:bg-gray-700 text-white font-semibold py-3.5 px-6 rounded-xl transition-all duration-200 border border-gray-700 hover:border-gray-600 flex items-center justify-center gap-2"
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
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
          Cancelar
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  @keyframes shimmer {
    0% {
      transform: translateX(-100%);
    }
    100% {
      transform: translateX(100%);
    }
  }

  .animate-shimmer {
    animation: shimmer 2s infinite;
  }
</style>
