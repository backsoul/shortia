<script lang="ts">
  let {
    videoSrc,
    currentTime = $bindable(0),
    isPlaying = $bindable(false),
    customStartTime,
    customEndTime,
  }: {
    videoSrc: string;
    currentTime: number;
    isPlaying: boolean;
    customStartTime: number;
    customEndTime: number;
  } = $props();

  let videoElement: HTMLVideoElement;
  let duration = $state(0);

  function handleTimeUpdate() {
    if (!videoElement) return;
    currentTime = videoElement.currentTime;

    // Auto-pause at clip end
    if (currentTime >= customEndTime) {
      videoElement.pause();
      videoElement.currentTime = customStartTime;
    }
  }

  export function seekTo(time: number) {
    if (videoElement) {
      videoElement.currentTime = time;
      currentTime = time;
    }
  }

  export function play() {
    if (videoElement) {
      if (currentTime >= customEndTime) {
        videoElement.currentTime = customStartTime;
      }
      videoElement.play();
    }
  }

  export function pause() {
    videoElement?.pause();
  }

  export function togglePlayPause() {
    if (!videoElement) return;
    if (isPlaying) {
      videoElement.pause();
    } else {
      play();
    }
  }
</script>

<!-- Video Element -->
<video
  bind:this={videoElement}
  bind:duration
  ontimeupdate={handleTimeUpdate}
  onplay={() => (isPlaying = true)}
  onpause={() => (isPlaying = false)}
  onloadedmetadata={() => {
    if (videoElement) {
      duration = videoElement.duration;
      videoElement.currentTime = customStartTime;
    }
  }}
  src={videoSrc}
  class="absolute inset-0 w-full h-full object-cover"
  style="object-fit: cover; z-index: 1;"
>
  <track kind="captions" label="Español" />
</video>
