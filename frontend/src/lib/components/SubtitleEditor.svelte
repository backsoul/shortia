<script lang="ts">
  import type { SubtitleConfig } from "$lib/types";

  export let subtitles: SubtitleConfig[];

  let selectedIndex = 0;

  function addSubtitle() {
    subtitles = [
      ...subtitles,
      {
        text: "New subtitle",
        start_time: 0,
        end_time: 2,
        font_family: "Arial",
        font_size: 48,
        color: "#FFFFFF",
        bg_color: "#000000AA",
        position: "bottom",
        bold: true,
        italic: false,
        transition: "none",
      },
    ];
    selectedIndex = subtitles.length - 1;
  }

  function removeSubtitle(index: number) {
    subtitles = subtitles.filter((_, i) => i !== index);
    if (selectedIndex >= subtitles.length) {
      selectedIndex = Math.max(0, subtitles.length - 1);
    }
  }

  $: currentSubtitle = subtitles[selectedIndex];
</script>

<div class="border-t border-gray-200 pt-6">
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900">📝 Subtitles</h3>
    <button
      on:click={addSubtitle}
      class="text-sm bg-blue-600 text-white px-3 py-1.5 rounded-lg hover:bg-blue-700 transition"
    >
      + Add Subtitle
    </button>
  </div>

  {#if subtitles.length === 0}
    <p class="text-gray-500 text-sm text-center py-4">
      No subtitles yet. Click "Add Subtitle" to create one.
    </p>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Subtitle List -->
      <div class="space-y-2 max-h-96 overflow-y-auto">
        {#each subtitles as subtitle, index}
          <div
            on:click={() => (selectedIndex = index)}
            on:keydown={(e) => e.key === "Enter" && (selectedIndex = index)}
            role="button"
            tabindex="0"
            class="w-full text-left p-3 border rounded-lg transition cursor-pointer {selectedIndex ===
            index
              ? 'border-blue-500 bg-blue-50'
              : 'border-gray-200 hover:border-gray-300'}"
          >
            <div class="flex justify-between items-start mb-1">
              <span class="text-sm font-medium text-gray-900 line-clamp-1">
                {subtitle.text}
              </span>
              <button
                on:click|stopPropagation={() => removeSubtitle(index)}
                class="text-red-500 hover:text-red-700 text-xs"
                type="button"
              >
                ✕
              </button>
            </div>
            <div class="text-xs text-gray-500">
              {subtitle.start_time.toFixed(1)}s - {subtitle.end_time.toFixed(
                1
              )}s
            </div>
          </div>
        {/each}
      </div>

      <!-- Subtitle Editor -->
      {#if currentSubtitle}
        <div class="space-y-3 border border-gray-200 rounded-lg p-4">
          <div>
            <label class="block text-xs font-medium text-gray-700 mb-1"
              >Text</label
            >
            <textarea
              bind:value={currentSubtitle.text}
              rows="2"
              class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
            ></textarea>
          </div>

          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1"
                >Start (s)</label
              >
              <input
                type="number"
                bind:value={currentSubtitle.start_time}
                step="0.1"
                class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1"
                >End (s)</label
              >
              <input
                type="number"
                bind:value={currentSubtitle.end_time}
                step="0.1"
                class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <div>
            <label class="block text-xs font-medium text-gray-700 mb-1"
              >Position</label
            >
            <select
              bind:value={currentSubtitle.position}
              class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
            >
              <option value="top">Top</option>
              <option value="center">Center</option>
              <option value="bottom">Bottom</option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1"
                >Font Size</label
              >
              <input
                type="number"
                bind:value={currentSubtitle.font_size}
                min="12"
                max="120"
                class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1"
                >Font</label
              >
              <select
                bind:value={currentSubtitle.font_family}
                class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
              >
                <option value="Arial">Arial</option>
                <option value="Impact">Impact</option>
                <option value="Comic Sans MS">Comic Sans</option>
                <option value="Times New Roman">Times New Roman</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1"
                >Text Color</label
              >
              <input
                type="color"
                bind:value={currentSubtitle.color}
                class="w-full h-8 rounded border border-gray-300"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1"
                >BG Color</label
              >
              <input
                type="color"
                bind:value={currentSubtitle.bg_color}
                class="w-full h-8 rounded border border-gray-300"
              />
            </div>
          </div>

          <div class="flex gap-3">
            <label class="flex items-center gap-2">
              <input
                type="checkbox"
                bind:checked={currentSubtitle.bold}
                class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <span class="text-xs text-gray-700">Bold</span>
            </label>
            <label class="flex items-center gap-2">
              <input
                type="checkbox"
                bind:checked={currentSubtitle.italic}
                class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <span class="text-xs text-gray-700">Italic</span>
            </label>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>
