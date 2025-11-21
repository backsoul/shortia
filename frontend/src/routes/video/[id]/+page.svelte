<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import ClipEditor from '$lib/components/ClipEditor.svelte';
	import type { Video, SuggestedClip, Transcript } from '$lib/types';

	let video: Video | null = null;
	let transcript: Transcript | null = null;
	let suggestedClips: SuggestedClip[] = [];
	let loading = true;
	let error = '';

	let videoId = '';

	onMount(async () => {
		// Get video ID from URL
		const id = window.location.pathname.split('/').pop();
		if (id) {
			videoId = id;
			await loadVideoData();
		}
	});

	async function loadVideoData() {
		try {
			loading = true;
			
			// Load video
			console.log(`Loading video: ${videoId}`);
			const videoRes = await fetch(`/api/videos/${videoId}`);
			if (!videoRes.ok) throw new Error('Video not found');
			video = await videoRes.json();
			console.log('Video loaded:', video);

			// Load transcript
			try {
				const transcriptRes = await fetch(`/api/videos/${videoId}/transcript`);
				if (transcriptRes.ok) {
					transcript = await transcriptRes.json();
					console.log('Transcript loaded:', transcript);
				}
			} catch (e) {
				console.log('Transcript not available yet:', e);
			}

			// Load suggested clips
			try {
				const clipsRes = await fetch(`/api/videos/${videoId}/clips`);
				console.log('Clips response status:', clipsRes.status);
				if (clipsRes.ok) {
					const clipsData = await clipsRes.json();
					suggestedClips = clipsData;
					console.log('Suggested clips loaded:', suggestedClips);
				}
			} catch (e) {
				console.log('Suggested clips error:', e);
			}

		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load video';
			console.error('Load error:', err);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>{video ? video.title : 'Cargando...'} - Shortia</title>
</svelte:head>

<div class="min-h-screen bg-[#0a0a0a] text-white relative overflow-hidden">
	<!-- Animated background with grid -->
	<div class="absolute inset-0 overflow-hidden pointer-events-none">
		<!-- Grid animado -->
		<div class="absolute inset-0 bg-grid-pattern animate-grid-move opacity-20"></div>
		
		<!-- Gradientes de fondo -->
		<div class="absolute w-[600px] h-[600px] -top-48 -left-48 bg-purple-500/20 rounded-full blur-3xl animate-pulse"></div>
		<div class="absolute w-[600px] h-[600px] top-1/2 -right-48 bg-blue-500/20 rounded-full blur-3xl animate-pulse" style="animation-delay: 1s;"></div>
		<div class="absolute w-[500px] h-[500px] -bottom-48 left-1/3 bg-pink-500/15 rounded-full blur-3xl animate-pulse" style="animation-delay: 2s;"></div>
	</div>

	<div class="relative z-10 container mx-auto px-4 py-8">
		{#if loading}
			<div class="flex flex-col justify-center items-center h-screen">
				<!-- Logo animado -->
				<div class="relative mb-8">
					<div class="w-24 h-24 rounded-3xl bg-gradient-to-br from-purple-600 via-violet-600 to-blue-600 flex items-center justify-center animate-pulse-glow shadow-2xl shadow-purple-500/50">
						<svg class="w-12 h-12 text-white animate-spin-slow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<!-- Anillos animados -->
					<div class="absolute inset-0 rounded-3xl border-2 border-purple-500/30 animate-ping"></div>
					<div class="absolute -inset-2 rounded-3xl border border-purple-500/20 animate-pulse"></div>
				</div>
				<h2 class="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-purple-400 to-blue-400 mb-2">
					Cargando editor...
				</h2>
				<p class="text-gray-500">Preparando tu espacio de trabajo</p>
			</div>
		{:else if error}
			<div class="flex justify-center items-center h-screen">
				<div class="max-w-md w-full bg-gradient-to-br from-red-900/40 to-orange-900/40 backdrop-blur-sm border border-red-500/20 rounded-2xl p-8 shadow-2xl">
					<div class="flex items-center gap-4 mb-4">
						<div class="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center">
							<svg class="w-6 h-6 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
							</svg>
						</div>
						<h2 class="text-xl font-bold text-red-400">Error al cargar</h2>
					</div>
					<p class="text-red-200">{error}</p>
				</div>
			</div>
		{:else if video}
			<ClipEditor {video} {transcript} {suggestedClips} />
		{/if}
	</div>
</div>

<style>
	/* Grid pattern de fondo */
	.bg-grid-pattern {
		background-image: 
			linear-gradient(to right, rgba(139, 92, 246, 0.1) 1px, transparent 5px),
			linear-gradient(to bottom, rgba(139, 92, 246, 0.1) 1px, transparent 5px);
		background-size: 200px 200px;
	}

	@keyframes grid-move {
		0% {
			transform: translate(0, 0);
		}
		100% {
			transform: translate(50px, 50px);
		}
	}

	.animate-grid-move {
		animation: grid-move 20s linear infinite;
	}

	@keyframes pulse-glow {
		0%, 100% {
			box-shadow: 0 0 20px rgba(168, 85, 247, 0.4),
						0 0 40px rgba(168, 85, 247, 0.2);
		}
		50% {
			box-shadow: 0 0 30px rgba(168, 85, 247, 0.6),
						0 0 60px rgba(168, 85, 247, 0.3);
		}
	}

	.animate-pulse-glow {
		animation: pulse-glow 2s ease-in-out infinite;
	}

	@keyframes spin-slow {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	.animate-spin-slow {
		animation: spin-slow 3s linear infinite;
	}
</style>
