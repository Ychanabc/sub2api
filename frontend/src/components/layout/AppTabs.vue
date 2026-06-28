<template>
  <div class="app-tabs-bar">
    <div class="app-tabs-scroll scrollbar-hide">
      <div
        v-for="tab in tabs"
        :key="tab.path"
        role="button"
        tabindex="0"
        class="app-tab"
        :class="{ 'app-tab-active': tab.path === route.fullPath }"
        @click="go(tab.path)"
        @keydown.enter.prevent="go(tab.path)"
        @keydown.space.prevent="go(tab.path)"
      >
        <span class="app-tab-dot"></span>
        <span class="app-tab-label">{{ tab.title }}</span>
        <button
          v-if="tabs.length > 1"
          type="button"
          class="app-tab-close"
          :aria-label="`Close ${tab.title}`"
          @click.stop="close(tab.path)"
        >
          <Icon name="x" size="xs" />
        </button>
      </div>
    </div>
    <div class="app-tabs-actions">
      <button
        v-if="tabs.length > 1"
        type="button"
        class="app-tabs-action"
        :title="t('common.closeAllTabs')"
        :aria-label="t('common.closeAllTabs')"
        @click="closeAll"
      >
        <Icon name="x" size="sm" />
      </button>
      <button
        type="button"
        class="app-tabs-action"
        :title="t('common.refreshCurrentPage')"
        :aria-label="t('common.refreshCurrentPage')"
        @click="refreshCurrent"
      >
        <Icon name="refresh" size="sm" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

interface VisitedTab {
  path: string
  title: string
}

const STORAGE_KEY = 'sub2api_visited_tabs'
const MAX_TABS = 12

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const tabs = ref<VisitedTab[]>(readTabs())

function readTabs(): VisitedTab[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as VisitedTab[]
    return parsed.filter((item) => item.path && item.title)
  } catch {
    return []
  }
}

function persistTabs() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(tabs.value))
}

function resolveTitle(): string {
  const titleKey = route.meta.titleKey as string | undefined
  if (titleKey) return t(titleKey)
  return (route.meta.title as string | undefined) || route.name?.toString() || route.path
}

function syncCurrentTab() {
  if (route.meta.requiresAuth === false) return
  const current: VisitedTab = {
    path: route.fullPath,
    title: resolveTitle()
  }
  const existingIndex = tabs.value.findIndex((item) => item.path === current.path)
  if (existingIndex >= 0) {
    tabs.value[existingIndex] = current
  } else {
    tabs.value.push(current)
  }
  if (tabs.value.length > MAX_TABS) {
    tabs.value = tabs.value.slice(tabs.value.length - MAX_TABS)
  }
  persistTabs()
}

function go(path: string) {
  if (path !== route.fullPath) {
    router.push(path)
  }
}

function close(path: string) {
  const index = tabs.value.findIndex((item) => item.path === path)
  if (index < 0 || tabs.value.length <= 1) return

  tabs.value.splice(index, 1)
  persistTabs()

  if (path === route.fullPath) {
    const fallback = tabs.value[Math.max(0, index - 1)] || tabs.value[0]
    if (fallback) {
      router.push(fallback.path)
    }
  }
}

function closeAll() {
  const current = tabs.value.find((item) => item.path === route.fullPath)
  tabs.value = current ? [current] : []
  persistTabs()
}

function refreshCurrent() {
  router.go(0)
}

watch(
  () => route.fullPath,
  () => syncCurrentTab(),
  { immediate: true }
)
</script>
