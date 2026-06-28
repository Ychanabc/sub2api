import { ref } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'system'

const media = window.matchMedia?.('(prefers-color-scheme: dark)')
const savedTheme = localStorage.getItem('theme') as ThemeMode | null

export const themeMode = ref<ThemeMode>(
  savedTheme === 'light' || savedTheme === 'dark' ? savedTheme : 'system'
)

export function applyThemeMode() {
  const shouldUseDark = themeMode.value === 'dark' || (themeMode.value === 'system' && media?.matches)
  document.documentElement.classList.toggle('dark', shouldUseDark)
}

export function setThemeMode(mode: ThemeMode) {
  themeMode.value = mode
  if (mode === 'system') {
    localStorage.removeItem('theme')
  } else {
    localStorage.setItem('theme', mode)
  }
  applyThemeMode()
}

media?.addEventListener('change', () => {
  if (themeMode.value === 'system') {
    applyThemeMode()
  }
})

applyThemeMode()
