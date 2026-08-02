<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useTheme } from '../composables/useTheme'

const { current } = useTheme()
const scrolled = ref(false)
const mobileOpen = ref(false)
const copied = ref(false)

function onScroll() {
  scrolled.value = window.scrollY > 60
}

function copyInstall() {
  navigator.clipboard.writeText('go install github.com/kipenv/kip/cmd/kip@latest').then(() => {
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  })
}

onMounted(() => window.addEventListener('scroll', onScroll))
onUnmounted(() => window.removeEventListener('scroll', onScroll))
</script>

<template>
  <nav
    class="fixed top-0 left-0 right-0 z-50 transition-all duration-300"
    :class="current === 'clean'
      ? (scrolled ? 'glass border-b border-b1' : '')
      : (scrolled ? 'bg-s1/90 border-b border-b1 backdrop-blur-xl' : '')"
  >
    <div
      :class="current === 'clean'
        ? 'mx-auto max-w-7xl px-6 py-3'
        : 'px-6 lg:px-12 h-16 flex items-center justify-between'"
    >
      <div
        :class="current === 'clean'
          ? 'flex items-center justify-between rounded-2xl px-6 py-3 glass border border-b1'
          : 'contents'"
      >
        <!-- Logo -->
        <a href="/" class="flex items-center gap-2 font-mono font-bold text-sm tracking-wide text-t1">
          <span class="w-2 h-2 rounded-full bg-hi pulse-glow" />
          kip
        </a>

        <!-- Desktop links -->
        <div class="hidden md:flex items-center gap-8">
          <a href="#how" class="text-[13px] text-t3 hover:text-t2 transition-colors">How it works</a>
          <a href="#features" class="text-[13px] text-t3 hover:text-t2 transition-colors">Features</a>
          <a href="#selfhost" class="text-[13px] text-t3 hover:text-t2 transition-colors">Self-host</a>
          <a href="https://github.com/kipenv/kip" target="_blank" rel="noopener" class="text-[13px] text-t3 hover:text-t2 transition-colors">GitHub</a>
        </div>

        <!-- Desktop CTA -->
        <div class="hidden md:flex items-center gap-3">
          <button
            @click="copyInstall"
            class="font-mono text-[11px] bg-s3 border border-b2 px-3.5 py-1.5 rounded-lg text-t3 flex items-center gap-2 hover:border-hi/40 hover:text-hi transition-all cursor-pointer"
          >
            <span class="text-hi">$</span>
            {{ copied ? 'Copied!' : 'go install \u2026/kip@latest' }}
          </button>
          <a
            href="#selfhost"
            class="bg-accent text-accent-text font-semibold text-[13px] px-5 py-2 rounded-lg hover:bg-accent-hover hover:-translate-y-px transition-all"
          >
            Get started
          </a>
        </div>

        <!-- Mobile burger -->
        <button
          @click="mobileOpen = !mobileOpen"
          class="md:hidden w-10 h-10 flex items-center justify-center rounded-lg hover:bg-s3 transition-colors"
          aria-label="Menu"
        >
          <div class="space-y-1.5">
            <span class="block w-5 h-0.5 bg-t1 transition-transform" :class="mobileOpen ? 'rotate-45 translate-y-2' : ''" />
            <span class="block w-5 h-0.5 bg-t1 transition-opacity" :class="mobileOpen ? 'opacity-0' : ''" />
            <span class="block w-5 h-0.5 bg-t1 transition-transform" :class="mobileOpen ? '-rotate-45 -translate-y-2' : ''" />
          </div>
        </button>
      </div>

      <!-- Mobile menu -->
      <Transition
        enter-active-class="transition-all duration-200"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div
          v-if="mobileOpen"
          class="md:hidden absolute top-full left-4 right-4 mt-2 rounded-xl bg-s2/95 backdrop-blur-xl border border-b1 p-6 space-y-4"
        >
          <a href="#how" @click="mobileOpen = false" class="block text-t2 hover:text-t1">How it works</a>
          <a href="#features" @click="mobileOpen = false" class="block text-t2 hover:text-t1">Features</a>
          <a href="#selfhost" @click="mobileOpen = false" class="block text-t2 hover:text-t1">Self-host</a>
          <a
            href="#selfhost"
            @click="mobileOpen = false"
            class="block w-full text-center rounded-lg bg-accent px-5 py-3 text-sm font-semibold text-accent-text"
          >
            Get started
          </a>
        </div>
      </Transition>
    </div>
  </nav>
</template>
