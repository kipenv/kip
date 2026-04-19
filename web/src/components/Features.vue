<script setup lang="ts">
import { Lock, Timer, Users, Globe, Bot, Container } from 'lucide-vue-next'
import { useTheme } from '../composables/useTheme'

const { current } = useTheme()

const features = [
  { icon: Lock, title: 'AES-256-GCM encryption', desc: 'Authenticated encryption. Tampered ciphertext is rejected, not silently corrupted.' },
  { icon: Timer, title: 'Self-destructing links', desc: 'Configure TTL and max reads independently. Links die after use — Redis TTL guarantees it.' },
  { icon: Users, title: 'Lightweight teams', desc: 'Teams work like game lobbies. Create, share an invite code, done. No admin panels, no RBAC.' },
  { icon: Globe, title: 'Web decrypt page', desc: 'Receivers don\'t need the CLI. Open the link in a browser — decryption happens client-side.' },
  { icon: Bot, title: 'AI secret scan', desc: 'Regex patterns detect AWS keys, Stripe keys, GitHub tokens. BYO AI for deeper analysis.' },
  { icon: Container, title: 'Self-host in 30 seconds', desc: 'Docker Compose. Server + Redis + SQLite in one file. No cloud lock-in. Free forever.' },
]
</script>

<template>
  <section id="features" class="py-24 px-6 lg:px-12 max-w-[1240px] mx-auto">
    <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">Features</span>
    <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1] mb-5">
      Everything you need.<br>Nothing you don't.
    </h2>
    <p class="text-[17px] text-t2 max-w-[540px] leading-[1.7] font-light mb-14">
      We do one thing — ephemeral secret sharing — and we do it correctly.
    </p>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-px bg-b1 border border-b1 overflow-hidden"
      :class="current === 'clean' ? 'rounded-2xl' : 'rounded-xl'"
    >
      <div
        v-for="(f, i) in features"
        :key="i"
        class="bg-s2 p-8 hover:bg-s3 transition-all"
        :class="current === 'clean' ? 'hover:-translate-y-0.5' : ''"
        :style="{ transitionDelay: `${i * 50}ms` }"
      >
        <div class="w-9 h-9 bg-hi/8 border border-hi/15 rounded-lg flex items-center justify-center mb-4 text-hi">
          <component :is="f.icon" :size="18" />
        </div>
        <h3 class="font-display font-bold text-base mb-2">{{ f.title }}</h3>
        <p class="text-[13px] text-t2 leading-[1.6]">{{ f.desc }}</p>
      </div>
    </div>
  </section>
</template>
