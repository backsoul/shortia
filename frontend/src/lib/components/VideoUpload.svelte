<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Video } from '$lib/types';

	const dispatch = createEventDispatcher<{ videoCreated: Video }>();

	let url = '';
	let loading = false;
	let error = '';

	async function handleSubmit() {
		if (!url.trim()) {
			error = 'Please enter a YouTube URL';
			return;
		}

		try {
			loading = true;
			error = '';

			const response = await fetch('/api/videos', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ url }),
			});

			if (!response.ok) {
				throw new Error('Failed to process video');
			}

			const video: Video = await response.json();
			dispatch('videoCreated', video);
			
			url = '';
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	}
</script>

<div class="bg-white rounded-2xl shadow-lg p-8">
	<h2 class="text-2xl font-bold text-gray-900 mb-6">
		📎 Paste YouTube URL
	</h2>

	<form on:submit|preventDefault={handleSubmit} class="space-y-4">
		<div>
			<input
				type="text"
				bind:value={url}
				placeholder="https://youtube.com/watch?v=..."
				disabled={loading}
				class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-blue-500 focus:border-transparent transition disabled:bg-gray-100 disabled:cursor-not-allowed"
			/>
		</div>

		{#if error}
			<div class="bg-red-50 border border-red-200 rounded-lg p-3 text-red-800 text-sm">
				{error}
			</div>
		{/if}

		<button
			type="submit"
			disabled={loading || !url.trim()}
			class="w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white font-semibold py-3 px-6 rounded-lg hover:from-blue-700 hover:to-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
		>
			{#if loading}
				<div class="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
				Processing...
			{:else}
				<span>🚀</span>
				Generate Clips
			{/if}
		</button>
	</form>

	<div class="mt-6 text-sm text-gray-600">
		<p class="font-semibold mb-2">How it works:</p>
		<ol class="list-decimal list-inside space-y-1">
			<li>Paste a YouTube video URL</li>
			<li>AI analyzes the content and finds interesting moments</li>
			<li>Customize subtitles and export your clips</li>
		</ol>
	</div>
</div>
