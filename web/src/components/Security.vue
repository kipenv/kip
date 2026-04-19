<script setup lang="ts">
import { Shield } from 'lucide-vue-next'
</script>

<template>
  <section class="bg-s2 border-t border-b border-b1 py-24 px-6 lg:px-12">
    <div class="max-w-[1240px] mx-auto grid lg:grid-cols-2 gap-20 items-center">
      <!-- Left -->
      <div>
        <span class="font-mono text-[11px] uppercase tracking-[2px] text-hi block mb-4">Zero-knowledge architecture</span>
        <h2 class="font-display font-extrabold text-[clamp(28px,4vw,48px)] tracking-[-1px] leading-[1.1] mb-5">
          Even if we're hacked,<br>attackers get <em class="not-italic text-hi">garbage</em>
        </h2>
        <p class="text-base text-t2 leading-[1.7] font-light mb-6">
          The server stores only encrypted blobs. The decryption key lives exclusively in your URL fragment — which browsers never transmit over HTTP. It's not a policy. It's cryptographic fact.
        </p>
        <div class="space-y-2">
          <div class="inline-flex items-center gap-2 bg-hi/6 border border-hi/15 rounded-full px-3 py-1 text-[11px] text-hi font-mono">
            <Shield :size="12" />
            AES-256-GCM · Argon2id · base58 URL encoding
          </div>
          <div class="block">
            <span class="inline-flex items-center gap-2 bg-hi/6 border border-hi/15 rounded-full px-3 py-1 text-[11px] text-hi font-mono">
              <Shield :size="12" />
              Authenticated encryption — tampered ciphertext is rejected
            </span>
          </div>
        </div>
      </div>

      <!-- Right — Crypto flow -->
      <div class="bg-s1 border border-b1 rounded-xl p-8 font-mono text-xs">
        <div class="flex gap-3 mb-5">
          <div class="w-6 h-6 rounded-full bg-s3 border border-hi/30 flex items-center justify-center text-hi text-[10px] flex-shrink-0">1</div>
          <div>
            <div class="text-t2 mb-1">CLI generates key locally</div>
            <div class="text-hi">key = crypto/rand<span class="text-t3">(32 bytes)</span></div>
            <div class="text-hi">nonce = crypto/rand<span class="text-t3">(12 bytes)</span></div>
          </div>
        </div>
        <div class="border-l border-dashed border-b2 ml-3 h-3" />

        <div class="flex gap-3 mb-5">
          <div class="w-6 h-6 rounded-full bg-s3 border border-hi/30 flex items-center justify-center text-hi text-[10px] flex-shrink-0">2</div>
          <div>
            <div class="text-t2 mb-1">Encrypt locally, upload ciphertext only</div>
            <div class="text-hi">POST /api/v1/secret</div>
            <div class="text-t3">{ ciphertext, nonce } <span class="text-amber">← no key</span></div>
          </div>
        </div>
        <div class="border-l border-dashed border-b2 ml-3 h-3" />

        <div class="flex gap-3 mb-5">
          <div class="w-6 h-6 rounded-full bg-s3 border border-amber/30 flex items-center justify-center text-amber text-[10px] flex-shrink-0">3</div>
          <div>
            <div class="text-t2 mb-1">Key goes in URL fragment only</div>
            <div class="text-hi">/s/k9xm2p<span class="text-amber">#</span>AZ7kQR...</div>
            <div class="text-t3">browsers never send # to servers</div>
          </div>
        </div>
        <div class="border-l border-dashed border-b2 ml-3 h-3" />

        <div class="flex gap-3">
          <div class="w-6 h-6 rounded-full bg-s3 border border-hi/30 flex items-center justify-center text-hi text-[10px] flex-shrink-0">4</div>
          <div>
            <div class="text-t2 mb-1">Server deletes after reads/TTL</div>
            <div class="text-hi">Redis DEL secret:<span class="text-amber">k9xm2p</span></div>
            <div class="text-t3">reads_left = 0 → gone forever</div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
