import { ref, watch, onMounted } from 'vue'

export type Theme = 'clean' | 'lime'

const current = ref<Theme>('clean')

export function useTheme() {
  function toggle() {
    current.value = current.value === 'clean' ? 'lime' : 'clean'
  }

  function set(theme: Theme) {
    current.value = theme
  }

  function applyTheme(theme: Theme) {
    if (typeof document === 'undefined') return
    const html = document.documentElement
    html.classList.remove('theme-clean', 'theme-lime')
    html.classList.add(`theme-${theme}`)
  }

  onMounted(() => {
    applyTheme(current.value)
  })

  watch(current, (theme) => {
    applyTheme(theme)
  })

  return { current, toggle, set }
}
