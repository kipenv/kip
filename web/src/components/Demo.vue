<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Play, Pause, RotateCcw } from 'lucide-vue-next'

const videoEl = ref<HTMLVideoElement | null>(null)
const playing = ref(false)
const reduceMotion = ref(false)
let observer: IntersectionObserver | null = null

function playVideo() {
  const el = videoEl.value
  if (!el) return
  // Autoplay can still be refused (low power mode, browser policy). State is
  // driven by the `playing` event, not by this promise: `play()` resolving only
  // means playback was accepted, not that a frame ever reached the screen. If
  // the file 404s or the codec is unsupported, binding to `play` would hide the
  // poster and the button and leave a black box.
  el.play().catch(() => {})
}

function toggle() {
  const el = videoEl.value
  if (!el) return
  if (el.paused) playVideo()
  else el.pause()
}

function restart() {
  const el = videoEl.value
  if (!el) return
  el.currentTime = 0
  playVideo()
}

onMounted(() => {
  reduceMotion.value = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (reduceMotion.value || !videoEl.value) return

  // Only run while it is actually on screen — a 25s loop playing in a
  // background tab or far off-screen is wasted battery.
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) playVideo()
        else if (videoEl.value && !videoEl.value.paused) {
          videoEl.value.pause()
          playing.value = false
        }
      }
    },
    { threshold: 0.35 },
  )
  observer.observe(videoEl.value)
})

onBeforeUnmount(() => observer?.disconnect())
</script>

<template>
  <section id="demo" class="relative py-24 px-6 lg:px-12 overflow-hidden">
    <!-- Key light behind the frame, so the video reads as lit rather than pasted on -->
    <div
      class="pointer-events-none absolute left-1/2 top-[38%] -translate-x-1/2 w-[min(1000px,90vw)] h-[420px] rounded-full bg-hi opacity-[0.13] blur-[130px]"
      aria-hidden="true"
    />

    <div class="relative max-w-[1240px] mx-auto">
      <div class="max-w-[760px] mb-12">
        <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">See it run</span>
        <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1] mb-5">
          From install to self-destruct,<br class="hidden sm:block" />
          in half a minute.
        </h2>
        <p class="text-t2 text-[17px] leading-[1.7] max-w-[560px]">
          Two machines, one link. Every line of output was captured from the real
          CLI running against a real server — nothing here is a mockup.
        </p>
      </div>

      <figure class="relative">
        <div
          class="relative overflow-hidden rounded-2xl border border-b2 bg-s2 shadow-[0_60px_120px_-40px_rgba(0,0,0,0.9)]"
        >
          <video
            ref="videoEl"
            class="block w-full aspect-video"
            muted
            loop
            playsinline
            preload="metadata"
            poster="/video/kip-demo-poster.jpg"
            @playing="playing = true"
            @pause="playing = false"
          >
            <source src="/video/kip-demo.mp4" type="video/mp4" />
          </video>

          <!-- Shown until playback actually starts: autoplay refused, reduced
               motion, or the viewer paused it. -->
          <button
            v-show="!playing"
            type="button"
            class="absolute inset-0 grid place-items-center bg-s1/45 backdrop-blur-[2px] transition-opacity duration-300 cursor-pointer"
            aria-label="Play the demo"
            @click="toggle"
          >
            <span
              class="grid place-items-center w-20 h-20 rounded-full bg-t1 text-s1 shadow-[0_20px_50px_-12px_rgba(0,0,0,0.8)] transition-transform duration-200 hover:scale-105"
            >
              <Play class="w-7 h-7 translate-x-[2px]" fill="currentColor" />
            </span>
          </button>
        </div>

        <figcaption
          class="mt-5 flex flex-wrap items-center justify-between gap-4 font-mono text-[12px] text-t3"
        >
          <span>26s · silent · loops</span>

          <span class="flex items-center gap-2">
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-b1 hover:border-b2 hover:text-t2 transition-colors cursor-pointer"
              :aria-label="playing ? 'Pause the demo' : 'Play the demo'"
              @click="toggle"
            >
              <Pause v-if="playing" class="w-3.5 h-3.5" />
              <Play v-else class="w-3.5 h-3.5" />
              {{ playing ? 'Pause' : 'Play' }}
            </button>
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-b1 hover:border-b2 hover:text-t2 transition-colors cursor-pointer"
              aria-label="Restart the demo"
              @click="restart"
            >
              <RotateCcw class="w-3.5 h-3.5" />
              Restart
            </button>
          </span>
        </figcaption>
      </figure>
    </div>
  </section>
</template>
