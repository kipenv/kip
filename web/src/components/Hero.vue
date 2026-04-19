<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useTheme } from '../composables/useTheme'

const { current } = useTheme()

const installMethods = [
  { label: 'Homebrew', command: 'brew install envshare' },
  { label: 'Go', command: 'go install github.com/antoniojosev/envshare@latest' },
  { label: 'curl', command: 'curl -fsSL https://envshare.dev/install.sh | sh' },
]

const activeTab = ref(0)
const copied = ref(false)

function copyCommand() {
  navigator.clipboard.writeText(installMethods[activeTab.value].command).then(() => {
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  })
}

// ── Terminal typewriter ──
interface TermLine {
  text: string
  class: string
  indent?: boolean
  separator?: boolean
}

const lines: TermLine[] = [
  { text: '# sender → one command', class: 'text-t3' },
  { text: '$ envshare push .env', class: 'prompt' },
  { text: '', class: '', separator: false }, // small pause
  { text: 'Scanning for exposed keys...', class: 'text-t3', indent: true },
  { text: '✓ No secrets detected in plaintext', class: 'text-emerald', indent: true },
  { text: 'Encrypting with AES-256-GCM...', class: 'text-t3', indent: true },
  { text: 'Uploading ciphertext...', class: 'text-t3', indent: true },
  { text: '---', class: '', separator: true },
  { text: '✓ Ready in 1.3s', class: 'text-emerald', indent: true },
  { text: '', class: '', separator: false },
  { text: 'Link: https://envshare.dev/s/k9xm2p#AZ7kQ...', class: 'link', indent: true },
  { text: 'Expires: 24h · Reads: 1', class: 'text-t3', indent: true },
  { text: '---', class: '', separator: true },
  { text: '# receiver → paste the link', class: 'text-t3' },
  { text: '$ envshare pull https://envshare.dev/s/k9xm2p#AZ7kQ...', class: 'prompt' },
  { text: '', class: '', separator: false },
  { text: '✓ Decrypted and saved → .env', class: 'text-emerald', indent: true },
  { text: '✓ Link destroyed', class: 'text-emerald', indent: true },
]

const visibleCount = ref(0)
const showCursor = ref(true)
let timer: ReturnType<typeof setTimeout> | null = null

function typeNext() {
  if (visibleCount.value < lines.length) {
    visibleCount.value++
    // Vary timing for realism
    const current = lines[visibleCount.value - 1]
    let delay = 90
    if (current?.text === '') delay = 300 // pause on empty
    else if (current?.class === 'prompt') delay = 400 // pause after commands
    else if (current?.separator) delay = 200
    else if (current?.text.startsWith('✓')) delay = 150
    else delay = 80 + Math.random() * 60
    timer = setTimeout(typeNext, delay)
  }
}

onMounted(() => {
  // Start typing after a brief delay
  timer = setTimeout(typeNext, 1000)
})

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})
</script>

<template>
  <section class="relative min-h-screen flex items-center overflow-hidden">
    <!-- BG: Clean theme = mesh gradient -->
    <div v-if="current === 'clean'" class="absolute inset-0 pointer-events-none hero-bg-clean" />
    <!-- BG: Lime theme = grid + scan -->
    <template v-else>
      <div class="absolute inset-0 pointer-events-none hero-bg-lime" />
      <div class="absolute top-0 left-0 right-0 h-px pointer-events-none opacity-30" style="background: linear-gradient(90deg, transparent, var(--t-hi), transparent); animation: scan 4s ease-in-out 3;" />
    </template>

    <!-- Grid noise (clean theme) -->
    <div
      v-if="current === 'clean'"
      class="absolute inset-0 pointer-events-none opacity-[0.025]"
      style="background-image: linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px); background-size: 60px 60px;"
    />

    <div class="relative z-10 mx-auto w-full px-6 lg:px-12 pt-28 pb-20" :class="current === 'clean' ? 'max-w-7xl' : 'max-w-[1240px]'">
      <div class="grid lg:grid-cols-2 gap-16 items-center">
        <!-- Left -->
        <div :class="current === 'clean' ? 'lg:col-span-1 space-y-8' : ''">
          <!-- Badge -->
          <div class="inline-flex items-center gap-2 rounded-full px-4 py-1.5 text-[11px] font-mono mb-7 border"
            :class="current === 'clean'
              ? 'border-b2 bg-s2/60 text-t2'
              : 'bg-[color-mix(in_srgb,var(--t-hi)_5%,transparent)] border-[color-mix(in_srgb,var(--t-hi)_15%,transparent)] text-[color-mix(in_srgb,var(--t-hi)_70%,transparent)]'"
          >
            <span class="w-[5px] h-[5px] rounded-full bg-hi pulse-glow" />
            Zero-knowledge encryption · Open Source · MIT License
          </div>

          <!-- Headline -->
          <h1 class="font-display font-extrabold leading-[1.02] tracking-[-2px] mb-5"
            :class="current === 'clean' ? 'text-5xl sm:text-6xl lg:text-7xl' : 'text-[clamp(40px,4.5vw,68px)]'"
          >
            <template v-if="current === 'clean'">
              Stop pasting<br>
              <span class="italic text-hi">.env files</span><br>
              in Slack
            </template>
            <template v-else>
              Stop sending<br>
              <span class="text-hi">.env files</span><br>
              <span class="relative text-t3">
                over Slack
                <span class="absolute left-0 right-0 top-1/2 h-[2px] bg-red rounded-sm" />
              </span>
            </template>
          </h1>

          <!-- Sub -->
          <p class="text-t2 max-w-[460px] mb-9 font-light"
            :class="current === 'clean' ? 'text-lg leading-relaxed' : 'text-[17px] leading-[1.75]'"
          >
            A CLI that encrypts your secrets, generates a <strong class="text-t1 font-medium">self-destructing link</strong>, and shares it in 3 seconds. The server never sees your keys.
          </p>

          <!-- Install tabs -->
          <div class="mb-2 flex gap-1">
            <button
              v-for="(method, i) in installMethods"
              :key="method.label"
              @click="activeTab = i; copied = false"
              class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer"
              :class="activeTab === i ? 'bg-s3 text-t1' : 'text-t3 hover:text-t2'"
            >
              {{ method.label }}
            </button>
          </div>

          <!-- Command box -->
          <div
            @click="copyCommand"
            class="group flex items-center gap-3 rounded-xl px-5 py-3.5 font-mono text-sm cursor-pointer transition-all border"
            :class="current === 'clean'
              ? 'bg-s2/80 border-b2 hover:border-hi/30 hover:shadow-[0_0_30px_rgba(167,139,250,0.1)]'
              : 'bg-s2/80 border-b1 hover:border-hi/40 hover:shadow-[0_0_30px_color-mix(in_srgb,var(--t-hi)_8%,transparent)]'"
          >
            <span class="text-hi">$</span>
            <span class="text-t1 break-all">{{ installMethods[activeTab].command }}</span>
            <span class="ml-auto text-t3 group-hover:text-t2 text-xs whitespace-nowrap transition-colors">
              {{ copied ? 'Copied!' : 'click to copy' }}
            </span>
          </div>

          <p class="mt-3 text-xs text-t3">
            No signup. No credit card. Works in 5 seconds.
          </p>

          <!-- Proof strip -->
          <div class="flex flex-wrap items-center gap-x-5 gap-y-2 mt-8 text-xs text-t3 font-mono">
            <span class="flex items-center gap-1.5" v-for="item in ['AES-256-GCM', 'Zero-knowledge', 'MIT License']" :key="item">
              <span class="w-[5px] h-[5px] rounded-full bg-hi opacity-60" />
              {{ item }}
            </span>
          </div>
        </div>

        <!-- Right — Terminal with typewriter -->
        <div class="rounded-2xl overflow-hidden shadow-2xl border border-b2"
          :style="{
            background: `var(--t-term)`,
            boxShadow: current === 'clean'
              ? '0 25px 80px rgba(0,0,0,0.5), 0 0 60px color-mix(in srgb, var(--t-hi) 6%, transparent)'
              : '0 32px 80px rgba(0,0,0,0.6)'
          }"
        >
          <!-- Title bar -->
          <div class="flex items-center gap-2 px-4 py-3 border-b border-b1" :style="{ background: `var(--t-s3)` }">
            <div class="w-[10px] h-[10px] rounded-full bg-[#FF5F56]" />
            <div class="w-[10px] h-[10px] rounded-full bg-[#FFBD2E]" />
            <div class="w-[10px] h-[10px] rounded-full bg-[#28C840]" />
            <span class="font-mono text-[11px] text-t3 ml-2">terminal</span>
          </div>

          <!-- Body -->
          <div class="p-5 font-mono text-[12.5px] leading-[1.9] min-h-[340px]">
            <template v-for="(line, i) in lines" :key="i">
              <template v-if="i < visibleCount">
                <!-- Separator -->
                <div v-if="line.separator" class="border-t border-b1 my-2" />

                <!-- Empty line (pause) -->
                <div v-else-if="line.text === ''" class="h-1" />

                <!-- Prompt line ($ command) -->
                <div v-else-if="line.class === 'prompt'" class="term-line-enter">
                  <span class="text-hi">$</span>
                  <span class="text-t1 ml-1">{{ line.text.slice(2) }}</span>
                </div>

                <!-- Link line -->
                <div v-else-if="line.class === 'link'" class="term-line-enter" :class="line.indent ? 'pl-4' : ''">
                  <span class="text-sky">Link: </span>
                  <span class="text-amber">https://envshare.dev/s/k9xm2p<span class="text-hi">#AZ7kQ...</span></span>
                </div>

                <!-- Normal line -->
                <div v-else class="term-line-enter" :class="[line.class, line.indent ? 'pl-4' : '']">
                  {{ line.text }}
                </div>
              </template>
            </template>

            <!-- Blinking cursor -->
            <span v-if="showCursor" class="inline-block w-[7px] h-[13px] bg-hi align-middle cursor-blink" :class="visibleCount > 0 ? 'ml-0.5' : 'ml-1'" />
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.term-line-enter {
  animation: termFadeIn 0.15s ease-out both;
}

@keyframes termFadeIn {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}
</style>
