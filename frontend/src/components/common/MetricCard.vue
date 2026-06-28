<template>
  <div class="metric-card">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <p class="metric-label">{{ title }}</p>
        <p class="metric-value" :class="valueClass" :title="String(value)">
          {{ value }}
        </p>
      </div>
      <div :class="['metric-icon', iconClass]">
        <Icon :name="icon" size="md" :stroke-width="2" />
      </div>
    </div>

    <div v-if="$slots.description || description" class="metric-description">
      <slot name="description">{{ description }}</slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'

type IconName = InstanceType<typeof Icon>['$props']['name']
type MetricTone = 'primary' | 'blue' | 'emerald' | 'amber' | 'violet' | 'rose' | 'gray'

const props = withDefaults(defineProps<{
  title: string
  value: string | number
  icon: IconName
  tone?: MetricTone
  description?: string
  highlightValue?: boolean
}>(), {
  tone: 'primary',
  description: '',
  highlightValue: false
})

const iconClass = computed(() => {
  const classes: Record<MetricTone, string> = {
    primary: 'metric-icon-primary',
    blue: 'metric-icon-blue',
    emerald: 'metric-icon-emerald',
    amber: 'metric-icon-amber',
    violet: 'metric-icon-violet',
    rose: 'metric-icon-rose',
    gray: 'metric-icon-gray'
  }
  return classes[props.tone]
})

const valueClass = computed(() => (props.highlightValue ? 'text-emerald-600 dark:text-emerald-400' : ''))
</script>
