<template>
  <div class="border-t border-b border-b1 bg-s2 py-5 overflow-hidden">
    <div class="flex gap-10 w-max" style="animation: scroll-logos 25s linear infinite;">
      <template v-for="n in 2" :key="n">
        <span class="font-mono text-xs text-t3 whitespace-nowrap px-5" v-for="item in items" :key="item + n">{{ item }}</span>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
const items = [
  'AES-256-GCM encryption', '·',
  'Self-destructing links', '·',
  'Zero-knowledge architecture', '·',
  'Self-hostable · Docker Compose', '·',
  'MIT License · Open source', '·',
  'Argon2id key derivation', '·',
  'Built in Go', '·',
  'No account needed to receive',
]
</script>
