<script setup lang="ts">
import { ref } from 'vue'
import { Check, X } from 'lucide-vue-next'
import { useTheme } from '../composables/useTheme'

const { current } = useTheme()
const yearly = ref(false)

const plans = [
  {
    name: 'Free', price: { monthly: 0, yearly: 0 },
    desc: 'For individuals. No signup. Works immediately after install.',
    cta: 'brew install envshare', ctaStyle: 'outline',
    features: [
      { text: 'Unlimited anonymous links', on: true },
      { text: '24h max TTL per link', on: true },
      { text: 'Up to 3 reads per link', on: true },
      { text: '1 file per push', on: true },
      { text: 'Regex secret scan', on: true },
      { text: 'Web decrypt page', on: true },
      { text: 'BYO AI scan (Ollama, OpenAI)', on: true },
      { text: 'Password-protected links', on: false },
      { text: 'Teams', on: false },
    ],
  },
  {
    name: 'Pro', price: { monthly: 6, yearly: 5 },
    desc: 'For power users and small teams who need longer TTLs and more control.',
    cta: 'Get Pro', ctaStyle: 'solid', featured: true,
    features: [
      { text: 'Everything in Free', on: true },
      { text: '30-day max TTL', on: true },
      { text: 'Configurable reads (1–100)', on: true },
      { text: 'Up to 5 files per push', on: true },
      { text: 'Password-protected links', on: true },
      { text: 'Teams (up to 20 members)', on: true },
      { text: 'AI scan included', on: true },
      { text: 'Pinned team envs', on: true },
      { text: 'Email support', on: true },
    ],
  },
  {
    name: 'Teams', price: { monthly: 25, yearly: 20 },
    desc: 'For companies. Centralized billing, unlimited teams, priority support.',
    cta: 'Talk to us', ctaStyle: 'outline',
    features: [
      { text: 'Everything in Pro', on: true },
      { text: 'Unlimited teams', on: true },
      { text: 'Unlimited members per team', on: true },
      { text: 'Centralized billing & invoices', on: true },
      { text: 'Audit log (share history)', on: true },
      { text: 'Priority support (24h SLA)', on: true },
      { text: 'Custom TTL policies', on: true },
      { text: 'BYOS (bring your own server)', on: true },
    ],
  },
]

function getPrice(plan: typeof plans[0]) {
  if (plan.price.monthly === 0) return '0'
  return yearly.value ? plan.price.yearly : plan.price.monthly
}
</script>

<template>
  <section id="pricing" class="py-24 px-6 lg:px-12">
    <div class="max-w-[1240px] mx-auto">
      <div class="text-center mb-14">
        <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">Pricing</span>
        <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1] mb-5">
          Honest pricing.<br>No link limits. Ever.
        </h2>
        <p class="text-[17px] text-t2 max-w-[540px] mx-auto leading-[1.7] font-light mb-8">
          You pay for capabilities, not for using the tool. Free tier has no daily link limits — that would defeat the point.
        </p>
        <!-- Toggle -->
        <div class="flex items-center justify-center gap-4">
          <span class="text-sm" :class="yearly ? 'text-t3' : 'text-t1'">Monthly</span>
          <button
            @click="yearly = !yearly"
            class="relative w-12 h-[26px] rounded-full transition-all cursor-pointer border"
            :class="yearly ? 'bg-hi border-hi' : 'bg-s3 border-b2'"
          >
            <div
              class="absolute top-[3px] w-[18px] h-[18px] rounded-full transition-transform"
              :class="yearly ? 'bg-s1 translate-x-[22px]' : 'bg-t2 translate-x-[3px]'"
            />
          </button>
          <span class="text-sm" :class="yearly ? 'text-t1' : 'text-t3'">
            Yearly
            <span class="ml-1.5 bg-hi/10 border border-hi/20 text-hi text-[11px] font-mono px-2 py-0.5 rounded-full">–20%</span>
          </span>
        </div>
      </div>

      <!-- Cards -->
      <div class="grid md:grid-cols-3 gap-6">
        <div
          v-for="plan in plans"
          :key="plan.name"
          class="relative p-8 transition-all hover:-translate-y-1"
          :class="[
            current === 'clean' ? 'rounded-2xl' : 'rounded-xl',
            plan.featured
              ? 'bg-s3 border border-hi/25 shadow-[0_0_40px_color-mix(in_srgb,var(--t-hi)_5%,transparent)] hover:shadow-[0_0_50px_color-mix(in_srgb,var(--t-hi)_10%,transparent)]'
              : 'bg-s2 border border-b1 hover:shadow-[0_8px_30px_rgba(0,0,0,0.3)]'
          ]"
        >
          <div
            v-if="plan.featured"
            class="absolute -top-3 left-1/2 -translate-x-1/2 bg-accent text-accent-text text-[11px] font-semibold px-4 py-1 rounded-full font-mono whitespace-nowrap"
          >
            Most popular
          </div>

          <div class="font-mono text-xs text-t3 uppercase tracking-wider mb-2">{{ plan.name }}</div>

          <div class="font-display font-extrabold text-5xl tracking-[-2px] leading-none mb-1 flex items-start gap-1">
            <span class="text-xl font-normal mt-2 text-t2">$</span>
            <span>{{ getPrice(plan) }}</span>
            <span class="text-sm font-normal text-t3 self-end mb-1">/mo</span>
          </div>

          <p class="text-[13px] text-t2 mb-7 mt-3 pb-7 border-b border-b1 min-h-[44px]">{{ plan.desc }}</p>

          <button
            class="w-full py-3 rounded-lg text-sm font-medium mb-7 transition-all cursor-pointer"
            :class="plan.ctaStyle === 'solid'
              ? 'bg-accent text-accent-text font-semibold hover:bg-accent-hover'
              : 'bg-transparent border border-b2 text-t2 hover:border-t3 hover:text-t1'"
          >
            {{ plan.cta }}
          </button>

          <ul class="space-y-3">
            <li
              v-for="f in plan.features"
              :key="f.text"
              class="flex items-start gap-2.5 text-[13px] leading-[1.5]"
              :class="f.on ? 'text-t1' : 'text-t3'"
            >
              <Check v-if="f.on" :size="14" class="mt-0.5 flex-shrink-0 text-hi" />
              <X v-else :size="14" class="mt-0.5 flex-shrink-0 text-b2" />
              {{ f.text }}
            </li>
          </ul>
        </div>
      </div>

      <p class="text-center text-[13px] text-t3 mt-8 font-mono">
        Self-hosting is free forever. MIT License.
        <a href="#selfhost" class="text-t2 border-b border-b2 hover:text-t1 transition-colors">See setup guide →</a>
      </p>
    </div>
  </section>
</template>
