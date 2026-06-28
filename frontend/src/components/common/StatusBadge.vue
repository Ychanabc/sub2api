<template>
  <span :class="['status-tag', variantClass]">
    <span
      :class="[
        'inline-block h-1.5 w-1.5 rounded-full',
        dotClass
      ]"
    ></span>
    <span>{{ label }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
  label: string
}>()

const variantClass = computed(() => {
  switch (props.status) {
    case 'active':
    case 'success':
      return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300'
    case 'disabled':
    case 'inactive':
    case 'warning':
      return 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300'
    case 'error':
    case 'danger':
      return 'border-red-200 bg-red-50 text-red-700 dark:border-red-800 dark:bg-red-950/40 dark:text-red-300'
    default:
      return 'border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300'
  }
})

const dotClass = computed(() => {
  switch (props.status) {
    case 'active':
    case 'success':
      return 'bg-emerald-500'
    case 'disabled':
    case 'inactive':
    case 'warning':
      return 'bg-amber-500'
    case 'error':
    case 'danger':
      return 'bg-red-500'
    default:
      return 'bg-gray-400'
  }
})
</script>

<style scoped>
.status-tag {
  @apply inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-xs font-medium;
}
</style>
