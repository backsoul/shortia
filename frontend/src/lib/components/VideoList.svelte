<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Video } from '$lib/types';

	export let videos: Video[];

	const dispatch = createEventDispatcher<{ refresh: void }>();

	function getStatusBadge(status: Video['status']) {
		const badges = {
			pending: { class: 'bg-gray-100 text-gray-800', text: '⏳ Pending' },
			downloading: { class: 'bg-blue-100 text-blue-800', text: '⬇️ Downloading' },
			transcribing: { class: 'bg-yellow-100 text-yellow-800', text: '🎤 Transcribing' },
			analyzing: { class: 'bg-purple-100 text-purple-800', text: '🤖 Analyzing' },
			completed: { class: 'bg-green-100 text-green-800', text: '✅ Completed' },
			error: { class: 'bg-red-100 text-red-800', text: '❌ Error' },
		};
		return badges[status] || badges.pending;
	}

	function formatDuration(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
	}
</script>

<div class="bg-white rounded-2xl shadow-lg p-8">
	<div class="flex justify-between items-center mb-6">
		<h2 class="text-2xl font-bold text-gray-900">
			📹 Your Videos
		</h2>
		<button
			on:click={() => dispatch('refresh')}
			class="text-blue-600 hover:text-blue-700 font-medium"
		>
			🔄 Refresh
		</button>
	</div>

	{#if videos.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No videos yet</p>
			<p class="text-sm">Upload your first YouTube URL to get started</p>
		</div>
	{:else}
		<div class="grid gap-4">
			{#each videos as video}
				<a
					href={video.status === 'completed' ? `/video/${video.id}` : '#'}
					class="block border border-gray-200 rounded-lg p-4 hover:border-blue-300 hover:shadow-md transition {video.status !== 'completed' ? 'cursor-default' : ''}"
				>
					<div class="flex gap-4">
						<!-- Thumbnail -->
						<div class="flex-shrink-0">
							{#if video.thumbnail_url}
								<img
									src={video.thumbnail_url}
									alt={video.title}
									class="w-32 h-20 object-cover rounded"
								/>
							{:else}
								<div class="w-32 h-20 bg-gray-200 rounded flex items-center justify-center">
									<span class="text-3xl">🎬</span>
								</div>
							{/if}
						</div>

						<!-- Info -->
						<div class="flex-1 min-w-0">
							<h3 class="font-semibold text-gray-900 truncate mb-1">
								{video.title || 'Processing...'}
							</h3>
							
							<div class="flex flex-wrap gap-2 text-sm text-gray-600 mb-2">
								{#if video.duration}
									<span>⏱️ {formatDuration(video.duration)}</span>
								{/if}
								<span>📅 {formatDate(video.created_at)}</span>
							</div>

							<div class="flex items-center gap-2">
								<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadge(video.status).class}">
									{getStatusBadge(video.status).text}
								</span>

								{#if video.status === 'completed'}
									<span class="text-xs text-blue-600 font-medium">
										Click to edit →
									</span>
								{/if}
							</div>
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
