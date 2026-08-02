<script setup lang="ts">
const rows = [
  { feature: 'Native CLI', kip: 'YES', slack: 'NO', yopass: 'NO', doppler: 'YES' },
  { feature: 'Zero-knowledge encryption', kip: 'YES', slack: 'NO', yopass: 'PARTIAL', doppler: 'NO' },
  { feature: 'Self-destructing links', kip: 'YES', slack: 'NO', yopass: 'YES', doppler: 'NO' },
  { feature: 'Self-hostable (free)', kip: 'YES', slack: 'NO', yopass: 'YES', doppler: 'NO' },
  { feature: 'Receiver needs account', kip: 'NO', slack: 'NO', yopass: 'NO', doppler: 'YES' },
  { feature: 'Lightweight teams', kip: 'YES', slack: 'YES', yopass: 'NO', doppler: 'YES' },
  { feature: 'Secret scanning', kip: 'YES', slack: 'NO', yopass: 'NO', doppler: 'PARTIAL' },
  { feature: 'Starts at', kip: 'FREE', slack: 'FREE', yopass: 'FREE', doppler: '$21/user/mo' },
]

function cls(val: string, isKip: boolean) {
  if (isKip) return 'text-hi font-semibold'
  if (val === 'YES' || val === 'FREE') return 'text-emerald'
  if (val === 'PARTIAL') return 'text-amber'
  return 'text-t3'
}
</script>

<template>
  <section class="py-24 px-6 lg:px-12">
    <div class="max-w-[1100px] mx-auto">
      <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">vs. alternatives</span>
      <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1] mb-12">
        The only tool that does<br>all of this at once.
      </h2>

      <div class="overflow-x-auto -mx-6 px-6">
        <table class="w-full border-collapse border border-b1 rounded-xl overflow-hidden min-w-[640px]">
          <thead>
            <tr>
              <th class="bg-s3 px-5 py-4 text-left font-mono text-xs uppercase tracking-wider text-t2 border-b border-b1">Feature</th>
              <th class="bg-s3 px-5 py-4 text-left font-mono text-xs uppercase tracking-wider border-b border-b1">
                <span class="flex items-center gap-1.5 text-hi font-semibold"><span class="w-2 h-2 rounded-full bg-hi" />kip</span>
              </th>
              <th class="bg-s3 px-5 py-4 text-left font-mono text-xs uppercase tracking-wider text-t3 border-b border-b1">Slack DM</th>
              <th class="bg-s3 px-5 py-4 text-left font-mono text-xs uppercase tracking-wider text-t3 border-b border-b1">Yopass</th>
              <th class="bg-s3 px-5 py-4 text-left font-mono text-xs uppercase tracking-wider text-t3 border-b border-b1">Doppler</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, i) in rows"
              :key="row.feature"
              class="hover:bg-s2/50 transition-colors"
              :class="i % 2 === 0 ? 'bg-hi/[0.02]' : ''"
            >
              <td class="px-5 py-4 text-[13px] text-t2 border-b border-b1">{{ row.feature }}</td>
              <td class="px-5 py-4 text-[11px] font-mono border-b border-b1" :class="cls(row.kip, true)">{{ row.kip }}</td>
              <td class="px-5 py-4 text-[11px] font-mono border-b border-b1" :class="cls(row.slack, false)">{{ row.slack }}</td>
              <td class="px-5 py-4 text-[11px] font-mono border-b border-b1" :class="cls(row.yopass, false)">{{ row.yopass }}</td>
              <td class="px-5 py-4 text-[11px] font-mono border-b border-b1" :class="cls(row.doppler, false)">{{ row.doppler }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>
