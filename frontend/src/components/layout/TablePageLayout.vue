<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <div v-if="$slots.actions || $slots.filters" class="table-page-toolbar">
      <div v-if="$slots.filters" class="table-page-filters">
        <slot name="filters" />
      </div>

      <div v-if="$slots.actions" class="table-page-actions">
        <slot name="actions" />
      </div>
    </div>

    <div class="layout-section-scrollable">
      <div class="card table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <div v-if="$slots.pagination" class="layout-section-fixed table-page-pagination">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.table-page-layout {
  @apply flex flex-col gap-4;
  height: calc(100vh - 64px - 44px - 3rem);
}

.layout-section-fixed {
  @apply flex-shrink-0;
}

.table-page-toolbar {
  @apply flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between;
}

.table-page-actions {
  @apply flex min-w-0 items-center justify-end;
}

.table-page-actions :deep(> *) {
  @apply w-auto;
}

.table-page-filters {
  @apply min-w-0;
}

.layout-section-scrollable {
  @apply flex min-h-0 flex-1 flex-col;
}

.table-scroll-container {
  @apply flex h-full flex-col overflow-hidden border dark:border-dark-700;
  background: var(--ios-surface);
  border-color: var(--ios-border);
  border-radius: var(--ios-radius);
  box-shadow: var(--ios-shadow-soft);
  backdrop-filter: blur(22px) saturate(1.35);
  -webkit-backdrop-filter: blur(22px) saturate(1.35);
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content;
  display: table;
}

.table-scroll-container :deep(thead) {
  @apply backdrop-blur-sm dark:bg-dark-800/80;
  background: rgba(118, 118, 128, 0.08);
}

.table-scroll-container :deep(th) {
  @apply border-b px-5 py-4 text-left text-sm font-semibold text-gray-500 dark:text-dark-300;
  border-color: var(--ios-border);
}

.table-scroll-container :deep(td) {
  @apply border-b px-5 py-4 text-sm text-gray-700 dark:text-gray-300;
  border-color: rgba(15, 23, 42, 0.06);
}

.table-scroll-container :deep(tbody tr) {
  transition: background-color 180ms var(--ios-ease);
}

.table-scroll-container :deep(tbody tr:hover) {
  background: rgba(0, 122, 255, 0.045);
}

.table-page-pagination {
  @apply overflow-hidden border dark:border-dark-700;
  background: var(--ios-surface);
  border-color: var(--ios-border);
  border-radius: var(--ios-radius);
  box-shadow: var(--ios-shadow-soft);
  backdrop-filter: blur(22px) saturate(1.35);
  -webkit-backdrop-filter: blur(22px) saturate(1.35);
}

.table-page-layout.mobile-mode {
  height: auto;
}

.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto overflow-visible border-none bg-transparent shadow-none;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply min-h-fit flex-none;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(.table-wrapper) {
  @apply overflow-visible;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}

.table-page-layout.mobile-mode .table-page-actions {
  @apply justify-start;
}

.table-page-layout.mobile-mode .table-page-actions :deep(> *) {
  @apply w-full;
}
</style>
