<script setup lang="ts">
import { ref } from 'vue'
import { Check } from 'lucide-vue-next'

const copied = ref(false)

const composeYml = `services:
  envshare:
    image: ghcr.io/envshare/server:latest
    environment:
      REDIS_URL: redis://redis:6379
      PORT: "8080"
    ports: ["8080:8080"]
    volumes: ["./data:/data"]
  redis:
    image: redis:7-alpine`

function copyCompose() {
  navigator.clipboard.writeText(composeYml).then(() => {
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  })
}

const benefits = [
  'No per-seat pricing. Ever.',
  'MIT License — fork it, own it.',
  'SQLite = single file backup',
  'Health check endpoint + structured logging',
]
</script>

<template>
  <section id="selfhost" class="bg-s2 border-t border-b border-b1 py-24 px-6 lg:px-12">
    <div class="max-w-[1240px] mx-auto grid lg:grid-cols-2 gap-20 items-center">
      <div>
        <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">Self-hosting</span>
        <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1] mb-5">
          Your infra.<br>Your data.<br>Zero ops.
        </h2>
        <p class="text-base text-t2 leading-[1.7] font-light mb-6">
          One Docker Compose file brings up the Go server, Redis, and an SQLite volume. It runs anywhere — your VPS, Fly.io, Railway, whatever.
        </p>
        <div class="space-y-3">
          <div v-for="b in benefits" :key="b" class="flex items-center gap-3 text-sm text-t2">
            <Check :size="16" class="text-hi flex-shrink-0" />
            {{ b }}
          </div>
        </div>
      </div>

      <div class="bg-s1 border border-b2 rounded-xl overflow-hidden">
        <div class="bg-s3 px-5 py-3 flex justify-between items-center border-b border-b1">
          <span class="font-mono text-xs text-t3">docker-compose.yml</span>
          <button
            @click="copyCompose"
            class="font-mono text-[11px] text-t3 bg-s4 border border-b1 px-2.5 py-1 rounded-md hover:text-t1 hover:border-b2 transition-all cursor-pointer"
          >
            {{ copied ? 'Copied!' : 'Copy' }}
          </button>
        </div>
        <div class="p-6 font-mono text-[13px] leading-[2]">
          <div class="text-t3"># That's it. No config files.</div>
          <div class="text-t1">services:</div>
          <div class="pl-5 text-t1">envshare:</div>
          <div class="pl-10"><span class="text-amber">image: </span><span class="text-hi">ghcr.io/envshare/server:latest</span></div>
          <div class="pl-10"><span class="text-amber">environment:</span></div>
          <div class="pl-14"><span class="text-hi">REDIS_URL: redis://redis:6379</span></div>
          <div class="pl-14"><span class="text-hi">PORT: "8080"</span></div>
          <div class="pl-10"><span class="text-amber">ports: </span><span class="text-hi">["8080:8080"]</span></div>
          <div class="pl-10"><span class="text-amber">volumes: </span><span class="text-hi">["./data:/data"]</span></div>
          <div class="pl-5 text-t1">redis:</div>
          <div class="pl-10"><span class="text-amber">image: </span><span class="text-hi">redis:7-alpine</span></div>
          <br>
          <div class="text-t3"># Boot in 30 seconds:</div>
          <div><span class="text-hi">$ </span><span class="text-t1">docker compose up -d</span></div>
          <div class="text-t3"># Point the CLI to your server:</div>
          <div><span class="text-hi">$ </span><span class="text-t1">envshare config set server http://your-host:8080</span></div>
        </div>
      </div>
    </div>
  </section>
</template>
