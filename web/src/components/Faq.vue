<script setup lang="ts">
import { ref } from 'vue'

const faqs = [
  { q: 'What happens if the server is compromised?', a: 'Attackers get encrypted blobs and nonces. Without the key (which lives only in your URL fragment, never on our server), they decrypt to garbage. This isn\'t a policy — it\'s a cryptographic guarantee.' },
  { q: 'Does the receiver need to install anything?', a: 'No. They can open the link in a browser. Decryption happens client-side with the Web Crypto API. The key in the URL fragment is never sent to our server. They can optionally install the CLI for a faster workflow.' },
  { q: 'What\'s the difference between Free and Pro?', a: 'Free gives you unlimited links with 24h max TTL and 3 reads per link. Pro unlocks 30-day TTL, configurable reads, password-protected links, teams, and included AI scan. No link quantity limits on either tier.' },
  { q: 'Can I self-host for my company?', a: 'Yes, and it\'s free forever. MIT License. The server + Redis + SQLite runs with one Docker Compose file. You own the data, the server, and the crypto keys. No seat-based pricing, ever.' },
  { q: 'How do teams work?', a: 'Teams are like game lobbies. You create one, share an invite code, and members join. No admin panels, no complex roles. You can push to the whole team or a specific member, pin an "official" .env, and see the inbox of pending shares.' },
  { q: 'What does the AI scan actually do?', a: 'Before sharing, it scans your .env for live production keys (AWS, Stripe, GitHub tokens), weak secrets, and URLs pointing to prod systems. Regex patterns run locally, always free. Connect your own AI for deeper analysis. Pro tier includes cloud AI.' },
]

const openIndex = ref<number | null>(null)
function toggle(i: number) { openIndex.value = openIndex.value === i ? null : i }
</script>

<template>
  <section class="py-24 px-6 lg:px-12 max-w-[1000px] mx-auto">
    <div class="text-center mb-14">
      <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">FAQ</span>
      <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1]">
        Straight answers.<br>No marketing copy.
      </h2>
    </div>

    <div class="grid md:grid-cols-2 gap-x-10 gap-y-0">
      <div
        v-for="(faq, i) in faqs"
        :key="i"
        class="border-b border-b1 py-6 cursor-pointer"
        @click="toggle(i)"
      >
        <h3 class="font-display font-bold text-base flex justify-between items-start gap-3">
          {{ faq.q }}
          <span class="text-t3 text-lg flex-shrink-0 transition-transform" :class="openIndex === i ? 'rotate-45' : ''">+</span>
        </h3>
        <Transition
          enter-active-class="transition-all duration-300 ease-out"
          enter-from-class="max-h-0 opacity-0"
          enter-to-class="max-h-96 opacity-100"
          leave-active-class="transition-all duration-200 ease-in"
          leave-from-class="max-h-96 opacity-100"
          leave-to-class="max-h-0 opacity-0"
        >
          <p v-if="openIndex === i" class="text-sm text-t2 leading-[1.7] mt-3 overflow-hidden">{{ faq.a }}</p>
        </Transition>
      </div>
    </div>
  </section>
</template>
