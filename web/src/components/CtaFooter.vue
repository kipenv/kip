<script setup lang="ts">
import { ref } from 'vue'

const copied = ref(false)
function copyInstall() {
  navigator.clipboard.writeText('go install github.com/kipenv/kip/cmd/kip@latest').then(() => {
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  })
}
</script>

<template>
  <section class="bg-s2 border-t border-b1 py-24 px-6 lg:px-12 text-center">
    <h2 class="font-display font-extrabold text-[clamp(36px,5vw,64px)] tracking-[-2px] leading-[1.05] mb-4">
      Ship it.<br>Don't <em class="not-italic text-hi">Slack</em> it.
    </h2>
    <p class="text-lg text-t2 font-light mb-10">
      Install takes 5 seconds. First share takes 3 more.
    </p>

    <div
      @click="copyInstall"
      class="inline-flex items-center gap-4 bg-s1 border border-b2 rounded-xl px-6 py-4 font-mono text-sm cursor-pointer hover:border-hi/40 transition-all mb-4"
    >
      <span class="text-hi">$</span>
      <span class="text-t1">go install github.com/kipenv/kip/cmd/kip@latest</span>
      <span class="text-[11px] text-t3 bg-s3 border border-b1 px-2.5 py-1 rounded-md">
        {{ copied ? 'Copied!' : 'Click to copy' }}
      </span>
    </div>

    <p class="text-[13px] text-t3 font-mono">
      Also: <code class="text-t2">go install github.com/kipenv/kip/cmd/kip@latest</code> ·
      <a href="#selfhost" class="text-t2 hover:text-t1 transition-colors">self-host guide</a>
    </p>
  </section>
</template>
