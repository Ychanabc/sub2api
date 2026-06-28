<template>
  <button
    v-if="variant === 'toolbar'"
    type="button"
    class="theme-mode-toolbar-button"
    :title="currentOption.label"
    :aria-label="currentOption.label"
    @click="cycleThemeMode"
  >
    <Icon :name="currentOption.icon" size="md" />
  </button>

  <div
    v-else
    class="theme-mode-switch"
    :class="{ 'theme-mode-switch-sidebar': variant === 'sidebar' }"
    role="group"
    aria-label="Theme mode"
  >
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      class="theme-mode-option"
      :class="{ 'theme-mode-option-active': themeMode === option.value }"
      :title="option.label"
      :aria-pressed="themeMode === option.value"
      @click="setThemeMode(option.value)"
    >
      <Icon :name="option.icon" size="sm" />
      <span class="theme-mode-label">{{ option.label }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { setThemeMode, themeMode, type ThemeMode } from './themeMode'

type IconName = InstanceType<typeof Icon>['$props']['name']

withDefaults(defineProps<{
  variant?: 'default' | 'sidebar' | 'toolbar'
}>(), {
  variant: 'default'
})

const options: Array<{ value: ThemeMode; label: string; icon: IconName }> = [
  { value: 'light', label: 'Light', icon: 'sun' },
  { value: 'dark', label: 'Dark', icon: 'moon' },
  { value: 'system', label: 'System', icon: 'cpu' }
]

const currentOption = computed(() => options.find(option => option.value === themeMode.value) ?? options[2])

function cycleThemeMode() {
  const next: Record<ThemeMode, ThemeMode> = {
    light: 'dark',
    dark: 'system',
    system: 'light'
  }
  setThemeMode(next[themeMode.value])
}
</script>
