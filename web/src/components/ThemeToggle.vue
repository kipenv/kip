<script setup lang="ts">
import { useTheme } from '../composables/useTheme'
import { Paintbrush } from 'lucide-vue-next'

const { current, toggle } = useTheme()
</script>

<template>
  <button
    @click="toggle"
    class="fixed bottom-6 right-6 z-[200] flex items-center gap-2.5 px-4 py-2.5 rounded-full text-xs font-mono font-medium shadow-2xl transition-all hover:-translate-y-0.5 cursor-pointer"
    :class="current === 'clean'
      ? 'bg-[#18181B] border border-[#2A2A30] text-[#A1A1AA] hover:text-white hover:border-[#A78BFA]/40'
      : 'bg-[#182018] border border-[#263626] text-[#829882] hover:text-[#A8FF3E] hover:border-[#A8FF3E]/40'"
    style="box-shadow: 0 8px 32px rgba(0,0,0,0.5);"
  >
    <Paintbrush :size="14" />
    <span>{{ current === 'clean' ? 'Try Lime theme' : 'Try Clean theme' }}</span>
    <span
      class="w-3 h-3 rounded-full"
      :class="current === 'clean' ? 'bg-[#A78BFA]' : 'bg-[#A8FF3E]'"
    />
  </button>
</template>
