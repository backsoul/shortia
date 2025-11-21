<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import type { Video } from '$lib/types';

	let youtubeUrl = $state('');
	let processing = $state(false);
	let currentVideoId = $state<string | null>(null);
	let status = $state<string>('');
	let progress = $state<number>(0);
	let ws: WebSocket | null = null;
	let inputFocused = $state(false);
	let wsConnected = $state(false);
	let mouseX = $state(0);
	let mouseY = $state(0);
	let trails = $state<Array<{ x: number; y: number; id: number; opacity: number }>>([]);
	let trailId = 0;

	function handleMouseMove(event: MouseEvent) {
		mouseX = event.clientX;
		mouseY = event.clientY;
		
		// Agregar nuevo punto al trail
		trails = [...trails, {
			x: mouseX,
			y: mouseY,
			id: trailId++,
			opacity: 1
		}].slice(-15); // Mantener solo los últimos 15 puntos
	}

	onDestroy(() => {
		if (ws) {
			ws.close();
		}
	});

	async function handleSubmit(event: Event) {
		event.preventDefault();
		
		if (!youtubeUrl.trim() || processing) return;

		processing = true;
		status = 'Iniciando...';
		progress = 0;

		try {
			// Crear el video
			const response = await fetch('http://localhost:8080/api/videos', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ url: youtubeUrl.trim() }),
			});

			if (!response.ok) {
				throw new Error('Error al crear el video');
			}

			const video: Video = await response.json();
			currentVideoId = video.id;

			// Conectar WebSocket para recibir actualizaciones
			connectWebSocket(video.id);
		} catch (error) {
			console.error('Error:', error);
			status = 'Error al procesar el video';
			processing = false;
		}
	}

	function connectWebSocket(videoId: string) {
		ws = new WebSocket(`ws://localhost:8080/api/videos/${videoId}/ws`);

		ws.onopen = () => {
			console.log('WebSocket conectado');
			wsConnected = true;
		};

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			
			if (data.status) {
				status = getStatusMessage(data.status);
				
				// Calcular progreso basado en el estado
				switch (data.status) {
					case 'pending':
						progress = 5;
						break;
					case 'downloading':
						progress = 25;
						break;
					case 'transcribing':
						progress = 50;
						break;
					case 'analyzing':
						progress = 75;
						break;
					case 'completed':
						progress = 100;
						setTimeout(() => {
							goto(`/video/${videoId}`);
						}, 500);
						break;
					case 'error':
						status = 'Error al procesar el video';
						processing = false;
						break;
				}
			}
		};

		ws.onerror = (error) => {
			console.error('WebSocket error:', error);
			status = 'Error de conexión';
			processing = false;
			wsConnected = false;
		};

		ws.onclose = () => {
			console.log('WebSocket cerrado');
			wsConnected = false;
		};
	}

	function getStatusMessage(status: string): string {
		const messages: Record<string, string> = {
			pending: 'Preparando...',
			downloading: 'Descargando video...',
			transcribing: 'Transcribiendo audio...',
			analyzing: 'Analizando contenido...',
			completed: '¡Listo!',
			error: 'Error',
		};
		return messages[status] || status;
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			handleSubmit(event);
		}
	}
</script>

<svelte:head>
	<title>Shortia - Crea clips virales con IA</title>
</svelte:head>

<svelte:window onmousemove={handleMouseMove} />

<div class="min-h-screen bg-[#0a0a0a] text-white flex items-center justify-center p-4">
	<!-- Mouse trail effect -->
	{#each trails as trail, i (trail.id)}
		{@const size = 8 - (i * 0.05)}
		{@const opacity = (i / trails.length) * 0.05}
		<div
			class="pointer-events-none fixed z-[100] rounded-full bg-gradient-to-r from-purple-500 via-pink-500 to-blue-500 blur-sm transition-all duration-300"
			style="
				left: {trail.x}px;
				top: {trail.y}px;
				width: {size}px;
				height: {size}px;
				opacity: {opacity};
				transform: translate(-50%, -50%) scale({1 - (i * 2)});
			"
		></div>
	{/each}

	<!-- Indicador de conexión WebSocket -->
	{#if processing && wsConnected}
		<div class="fixed top-6 right-6 z-50 flex items-center gap-3 px-4 py-2.5 bg-gradient-to-r from-green-900/30 to-emerald-900/30 backdrop-blur-md rounded-full border border-green-500/20 shadow-lg shadow-green-500/10">
			<div class="relative">
				<div class="w-2.5 h-2.5 bg-green-500 rounded-full animate-pulse"></div>
				<div class="absolute inset-0 w-2.5 h-2.5 bg-green-500 rounded-full animate-ping"></div>
			</div>
			<span class="text-sm font-medium text-green-300">Conectado</span>
		</div>
	{/if}

	<!-- Animated background with grid -->
	<div class="absolute inset-0 overflow-hidden pointer-events-none">
		<!-- Grid animado -->
		<div class="absolute inset-0 bg-grid-pattern animate-grid-move opacity-20"></div>
		
		<!-- Gradientes de fondo -->
		<div class="absolute w-[500px] h-[500px] -top-48 -left-48 bg-purple-500/20 rounded-full blur-3xl animate-pulse"></div>
		<div class="absolute w-[500px] h-[500px] -bottom-48 -right-48 bg-blue-500/20 rounded-full blur-3xl animate-pulse" style="animation-delay: 1s;"></div>
	</div>

	<div class="relative z-10 w-full max-w-3xl">
		{#if !processing}
			<!-- Estado inicial: Input -->
			<div class="text-center space-y-8 animate-fade-in">
				<!-- Logo y título -->
				<div class="space-y-4">
					<div class="inline-flex items-center justify-center w-20 h-20 rounded-3xl bg-gradient-to-br from-purple-600 via-violet-600 to-blue-600 mb-6 shadow-2xl shadow-purple-500/50">
						<svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<h1 class="text-6xl md:text-7xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-purple-400 via-pink-400 to-blue-400 tracking-tight">
						Shortia
					</h1>
					<p class="text-xl text-gray-400 max-w-2xl mx-auto leading-relaxed">
						Transforma videos de YouTube en clips virales con inteligencia artificial
					</p>
				</div>

				<!-- Input principal -->
				<form onsubmit={handleSubmit} class="mt-16">
					<div class="relative group">
						<!-- Sombra que se mueve alrededor del borde -->
						<div class="absolute -inset-[2px] rounded-[18px] opacity-0 group-hover:opacity-100 {inputFocused ? 'opacity-100' : ''} transition-opacity duration-300 blur-xl pointer-events-none">
							<div class="absolute inset-0 rounded-[18px] bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 group-hover:animate-border-glow {inputFocused ? 'animate-border-glow' : ''}"></div>
						</div>
						
						<!-- Input container -->
						<div class="relative flex items-center bg-gradient-to-br from-[#1a1a1a] to-[#0f0f0f] rounded-2xl p-1.5 shadow-2xl border-0">
							<input
								type="text"
								bind:value={youtubeUrl}
								onkeypress={handleKeyPress}
								onfocus={() => inputFocused = true}
								onblur={() => inputFocused = false}
								placeholder="Ingresa la URL de YouTube..."
								class="input-custom flex-1 bg-transparent px-6 py-4 text-lg outline-none placeholder:text-gray-500 text-white"
								disabled={processing}
							/>
							<button
								type="submit"
								disabled={!youtubeUrl.trim() || processing}
								class="relative overflow-hidden px-8 py-4 bg-gradient-to-r from-purple-600 via-violet-600 to-blue-600 rounded-xl font-bold text-white shadow-lg shadow-purple-500/50 hover:shadow-2xl hover:shadow-purple-500/60 hover:scale-[1.02] transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 disabled:hover:shadow-lg group/btn"
							>
								<span class="relative z-10 flex items-center gap-2">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
									</svg>
									Crear clips
								</span>
								<div class="absolute inset-0 bg-gradient-to-r from-white/0 via-white/20 to-white/0 translate-x-[-100%] group-hover/btn:translate-x-[100%] transition-transform duration-1000"></div>
							</button>
						</div>
					</div>
				</form>

				<!-- Información adicional -->
				<div class="flex items-center justify-center gap-8 text-sm text-gray-500 mt-12">
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
						</svg>
						<span>100% Gratis</span>
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
							<path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z" />
						</svg>
						<span>Powered by IA</span>
					</div>
					<div class="flex items-center gap-2">
						<svg class="w-4 h-4 text-purple-500" fill="currentColor" viewBox="0 0 20 20">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
						</svg>
						<span>Clips en segundos</span>
					</div>
				</div>
			</div>
		{:else}
			<!-- Estado de procesamiento -->
			<div class="text-center space-y-8 animate-fade-in">
				<div class="inline-flex items-center justify-center w-24 h-24 rounded-3xl bg-gradient-to-br from-purple-600 via-violet-600 to-blue-600 mb-6 animate-pulse-glow shadow-2xl shadow-purple-500/50">
					<svg class="w-12 h-12 text-white animate-spin-slow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
				</div>

				<div class="space-y-4">
					<h2 class="text-4xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-purple-400 to-blue-400">
						Procesando tu video
					</h2>
					<p class="text-gray-400 text-xl font-medium">{status}</p>
				</div>

				<!-- Barra de progreso mejorada -->
				<div class="max-w-xl mx-auto space-y-4">
					<div class="relative h-4 bg-gradient-to-r from-gray-900 to-gray-800 rounded-full overflow-hidden shadow-inner">
						<!-- Glow effect debajo de la barra -->
						<div 
							class="absolute -inset-1 bg-gradient-to-r from-purple-500 to-blue-500 blur-lg opacity-50 transition-all duration-500"
							style="width: {progress}%; left: 0;"
						></div>
						<!-- Barra principal -->
						<div 
							class="absolute inset-y-0 left-0 bg-gradient-to-r from-purple-600 via-violet-600 to-blue-600 transition-all duration-500 ease-out rounded-full shadow-lg shadow-purple-500/50"
							style="width: {progress}%"
						>
							<!-- Efecto shimmer -->
							<div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent animate-shimmer"></div>
							<!-- Brillo en el borde -->
							<div class="absolute inset-0 rounded-full shadow-[inset_0_1px_0_0_rgba(255,255,255,0.4)]"></div>
						</div>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-sm text-gray-500 font-medium">{Math.round(progress)}% completado</span>
						<span class="text-sm text-purple-400 font-mono">{Math.round(progress / 25)}/4 pasos</span>
					</div>
				</div>

				<!-- Fases del proceso mejoradas -->
				<div class="grid grid-cols-2 md:grid-cols-4 gap-6 max-w-4xl mx-auto mt-16">
					{#each [
						{ label: 'Descarga', step: 1, icon: 'M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10', colors: { bg: 'from-purple-900/40 to-violet-900/40', glow: 'from-purple-600 to-violet-600', shadow: 'shadow-purple-500/50', icon: 'from-purple-600 to-violet-600', dot: 'bg-purple-500' } },
						{ label: 'Transcripción', step: 2, icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z', colors: { bg: 'from-blue-900/40 to-cyan-900/40', glow: 'from-blue-600 to-cyan-600', shadow: 'shadow-blue-500/50', icon: 'from-blue-600 to-cyan-600', dot: 'bg-blue-500' } },
						{ label: 'Análisis IA', step: 3, icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z', colors: { bg: 'from-pink-900/40 to-rose-900/40', glow: 'from-pink-600 to-rose-600', shadow: 'shadow-pink-500/50', icon: 'from-pink-600 to-rose-600', dot: 'bg-pink-500' } },
						{ label: 'Finalización', step: 4, icon: 'M5 13l4 4L19 7', colors: { bg: 'from-emerald-900/40 to-green-900/40', glow: 'from-emerald-600 to-green-600', shadow: 'shadow-emerald-500/50', icon: 'from-emerald-600 to-green-600', dot: 'bg-emerald-500' } }
					] as phase}
						{@const isActive = progress >= phase.step * 25}
						{@const isCompleted = progress > phase.step * 25}
						<div 
							class="relative p-6 rounded-2xl transition-all duration-700 ease-out transform {isActive ? 'scale-105 bg-gradient-to-br ' + phase.colors.bg : 'scale-100 bg-gray-900/30'}"
						>
							<!-- Glow effect -->
							{#if isActive}
								<div class="absolute -inset-0.5 bg-gradient-to-r {phase.colors.glow} rounded-2xl blur-lg opacity-60 animate-pulse-glow transition-all duration-700"></div>
							{/if}
							
							<!-- Contenido -->
							<div class="relative">
								<!-- Icono -->
								<div class="flex items-center justify-center w-14 h-14 mx-auto mb-4 rounded-xl transition-all duration-700 ease-out {isActive ? 'bg-gradient-to-br ' + phase.colors.icon + ' shadow-lg ' + phase.colors.shadow : 'bg-gray-800/50'}">
									<svg class="w-7 h-7 transition-all duration-700 {isActive ? 'text-white' : 'text-gray-600'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={phase.icon} />
									</svg>
									{#if isCompleted}
										<div class="absolute -top-1 -right-1 w-5 h-5 bg-green-500 rounded-full flex items-center justify-center shadow-lg animate-scale-in">
											<svg class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
												<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
											</svg>
										</div>
									{/if}
								</div>
								
								<!-- Texto -->
								<div class="text-center">
									<div class="text-sm font-semibold transition-colors duration-700 {isActive ? 'text-white' : 'text-gray-500'}">
										{phase.label}
									</div>
									{#if isActive && !isCompleted}
										<div class="mt-2 flex items-center justify-center gap-1">
											<div class="w-1.5 h-1.5 {phase.colors.dot} rounded-full animate-bounce" style="animation-delay: 0ms;"></div>
											<div class="w-1.5 h-1.5 {phase.colors.dot} rounded-full animate-bounce" style="animation-delay: 150ms;"></div>
											<div class="w-1.5 h-1.5 {phase.colors.dot} rounded-full animate-bounce" style="animation-delay: 300ms;"></div>
										</div>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.animate-fade-in {
		animation: fade-in 0.6s ease-out;
	}

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

	/* Eliminar completamente el border azul del input */
	.input-custom {
		border: none;
		outline: none;
		-webkit-appearance: none;
		-moz-appearance: none;
		appearance: none;
	}

	.input-custom:focus {
		border: none;
		outline: none;
		box-shadow: none;
	}

	.input-custom:focus-visible {
		outline: none;
	}

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

	@keyframes border-glow {
		0% {
			box-shadow: 
				-12px -12px 40px rgba(168, 85, 247, 0.8),
				12px -12px 40px rgba(168, 85, 247, 0.0),
				12px 12px 40px rgba(59, 130, 246, 0.0),
				-12px 12px 40px rgba(59, 130, 246, 0.0);
		}
		25% {
			box-shadow: 
				-12px -12px 40px rgba(168, 85, 247, 0.0),
				12px -12px 40px rgba(236, 72, 153, 0.8),
				12px 12px 40px rgba(236, 72, 153, 0.0),
				-12px 12px 40px rgba(168, 85, 247, 0.0);
		}
		50% {
			box-shadow: 
				-12px -12px 40px rgba(168, 85, 247, 0.0),
				12px -12px 40px rgba(236, 72, 153, 0.0),
				12px 12px 40px rgba(59, 130, 246, 0.8),
				-12px 12px 40px rgba(59, 130, 246, 0.0);
		}
		75% {
			box-shadow: 
				-12px -12px 40px rgba(168, 85, 247, 0.0),
				12px -12px 40px rgba(236, 72, 153, 0.0),
				12px 12px 40px rgba(59, 130, 246, 0.0),
				-12px 12px 40px rgba(236, 72, 153, 0.8);
		}
		100% {
			box-shadow: 
				-12px -12px 40px rgba(168, 85, 247, 0.8),
				12px -12px 40px rgba(168, 85, 247, 0.0),
				12px 12px 40px rgba(59, 130, 246, 0.0),
				-12px 12px 40px rgba(59, 130, 246, 0.0);
		}
	}

	.animate-border-glow {
		animation: border-glow 4s ease-in-out infinite;
	}

	@keyframes scale-in {
		from {
			transform: scale(0);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}

	.animate-scale-in {
		animation: scale-in 0.3s ease-out;
	}
</style>
