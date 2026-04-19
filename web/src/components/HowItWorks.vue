<script setup lang="ts">
import { ref } from 'vue'
import { Lock, Link, Flame } from 'lucide-vue-next'
import { useTheme } from '../composables/useTheme'

const { current } = useTheme()
const activeStep = ref(0)

const steps = [
  {
    icon: Lock,
    title: 'Push your .env',
    desc: 'EnvShare encrypts your file locally with a unique random key. That key never touches our server.',
    code: '<span class="text-hi">$</span> envshare push .env\n<span class="text-t3">  --expires 24h --reads 1</span>',
    terminal: `<div class="text-t3"># Encrypt &amp; upload</div>
<div><span class="text-hi">$</span> <span class="text-t1">envshare push .env --expires 24h</span></div>
<br>
<div class="text-t3">  Generating AES-256-GCM key...  <span class="text-emerald">done</span></div>
<div class="text-t3">  Encrypting file locally...     <span class="text-emerald">done</span></div>
<div class="text-t3">  Uploading ciphertext...        <span class="text-emerald">done</span></div>
<br>
<div class="text-emerald">  ✓ Ready to share</div>
<br>
<div>  <span class="text-sky">Link:</span>  <span class="text-amber">https://envshare.dev/s/k9xm2p</span><span class="text-hi">#AZ7kQ...</span></div>
<div>  <span class="text-t3">TTL:</span>   <span class="text-amber">24 hours</span></div>
<div>  <span class="text-t3">Reads:</span> <span class="text-t2">1 (self-destructs on open)</span></div>`,
  },
  {
    icon: Link,
    title: 'Share the link',
    desc: 'The decryption key lives in the URL #fragment — browsers never send that to servers. It\'s an HTTP protocol guarantee.',
    code: '<span class="text-amber">https://envshare.dev/s/</span><span class="text-hi">k9xm2p<span class="text-amber">#</span>AZ7k...</span>',
    terminal: `<div class="text-t3"># URL anatomy</div>
<br>
<div class="text-t1">  https://envshare.dev/s/<span class="text-amber">k9xm2p</span><span class="text-hi">#AZ7kQR...</span></div>
<br>
<div class="text-t3">  ├─ server sees: <span class="text-t2">/s/k9xm2p</span></div>
<div class="text-t3">  │  (encrypted blob ID)</div>
<div class="text-t3">  │</div>
<div class="text-t3">  └─ browser keeps: <span class="text-hi">#AZ7kQR...</span></div>
<div class="text-t3">     (decryption key — never sent over HTTP)</div>
<br>
<div class="text-t3">  This is an HTTP protocol guarantee,</div>
<div class="text-t3">  not a policy promise.</div>`,
  },
  {
    icon: Flame,
    title: 'It self-destructs',
    desc: 'After the configured reads, the encrypted blob is permanently deleted. TTL ensures deletion even if nobody reads it.',
    code: '<span class="text-hi">$</span> envshare pull <span class="text-amber">https://...</span>\n<span class="text-t3">  ✓ Saved → .env · Link destroyed</span>',
    terminal: `<div class="text-t3"># Receiver pulls the file</div>
<div><span class="text-hi">$</span> <span class="text-t1">envshare pull https://envshare.dev/s/k9xm2p#AZ7k...</span></div>
<br>
<div class="text-t3">  Fetching encrypted payload...   <span class="text-emerald">done</span></div>
<div class="text-t3">  Extracting key from URL...      <span class="text-emerald">done</span></div>
<div class="text-t3">  Decrypting with AES-256-GCM...  <span class="text-emerald">done</span></div>
<br>
<div class="text-emerald">  ✓ Saved to .env</div>
<div class="text-emerald">  ✓ Link destroyed</div>
<br>
<div class="text-t3">  No traces remain on the server.</div>`,
  },
]
</script>

<template>
  <section id="how" class="py-24 px-6 lg:px-12 max-w-[1240px] mx-auto">
    <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">How it works</span>
    <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1] mb-5">
      Three commands.<br>Zero friction.
    </h2>
    <p class="text-[17px] text-t2 max-w-[540px] leading-[1.7] font-light mb-14">
      The receiver doesn't even need the CLI — they can open the link in a browser. That's the point.
    </p>

    <div class="grid lg:grid-cols-2 gap-12 items-start">
      <!-- Steps -->
      <div class="space-y-0">
        <div
          v-for="(step, i) in steps"
          :key="i"
          @click="activeStep = i"
          class="flex gap-5 py-5 cursor-pointer border-b border-b1 last:border-b-0 transition-all"
          :class="activeStep === i ? '' : 'opacity-50 hover:opacity-75'"
        >
          <div
            class="w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center border transition-all"
            :class="activeStep === i
              ? 'bg-hi/10 border-hi/30 text-hi'
              : 'bg-s3 border-b1 text-t3'"
          >
            <component :is="step.icon" :size="18" />
          </div>
          <div>
            <h3
              class="font-display font-bold text-lg mb-2 tracking-[-0.3px] transition-colors"
              :class="activeStep === i ? 'text-hi' : 'text-t1'"
            >
              {{ step.title }}
            </h3>
            <p class="text-sm text-t2 leading-[1.65]">{{ step.desc }}</p>
            <div class="mt-3 bg-s1 border border-b1 rounded-lg px-3.5 py-2.5 font-mono text-xs text-t2 leading-[1.6]" v-html="step.code" />
          </div>
        </div>
      </div>

      <!-- Terminal panel -->
      <div class="rounded-2xl overflow-hidden border border-b2 sticky top-24" :style="{ background: `var(--t-term)`, boxShadow: '0 32px 80px rgba(0,0,0,0.5)' }">
        <div class="px-4 py-3 flex items-center gap-2 border-b border-b1" :style="{ background: `var(--t-s3)` }">
          <div class="w-[10px] h-[10px] rounded-full bg-[#FF5F56]" />
          <div class="w-[10px] h-[10px] rounded-full bg-[#FFBD2E]" />
          <div class="w-[10px] h-[10px] rounded-full bg-[#28C840]" />
          <span class="font-mono text-[11px] text-t3 ml-2">terminal</span>
        </div>
        <div class="p-5 font-mono text-[12.5px] leading-[1.9] min-h-[300px]">
          <Transition
            mode="out-in"
            enter-active-class="transition-all duration-300"
            enter-from-class="opacity-0 translate-y-2"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition-all duration-150"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
          >
            <div :key="activeStep" v-html="steps[activeStep].terminal" />
          </Transition>
        </div>
      </div>
    </div>
  </section>
</template>
