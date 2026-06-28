<template>
  <div class="min-h-screen text-gray-900 dark:text-gray-100">
    <div class="grid min-h-screen lg:grid-cols-[minmax(0,1fr)_480px]">
      <section
        class="hidden border-r px-10 py-8 dark:border-dark-800 lg:flex lg:flex-col"
      >
        <div v-if="settingsLoaded" class="flex items-center gap-3">
          <div class="auth-logo-box">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div class="min-w-0">
            <h1 class="truncate text-base font-semibold text-gray-900 dark:text-white">
              {{ siteName }}
            </h1>
            <p class="truncate text-xs text-gray-500 dark:text-dark-400">
              {{ siteSubtitle }}
            </p>
          </div>
        </div>

        <div class="flex flex-1 items-center">
          <div class="max-w-xl">
            <p class="mb-3 text-sm font-medium text-primary-600 dark:text-primary-300">
              Sub2API Console
            </p>
            <h2 class="text-4xl font-semibold leading-tight text-gray-950 dark:text-white">
              {{ siteName }}
            </h2>
            <p class="mt-4 max-w-md text-sm leading-6 text-gray-500 dark:text-dark-300">
              {{ siteSubtitle }}
            </p>

            <div class="mt-10 grid max-w-md gap-3" aria-hidden="true">
              <div class="auth-preview-card">
                <div class="h-2 w-20 rounded bg-primary-200 dark:bg-primary-900/70"></div>
                <div class="mt-4 grid gap-2">
                  <div class="h-2 rounded bg-gray-200 dark:bg-dark-700"></div>
                  <div class="h-2 w-3/4 rounded bg-gray-200 dark:bg-dark-700"></div>
                </div>
              </div>
              <div class="auth-preview-card">
                <div class="grid grid-cols-3 gap-3">
                  <div class="h-14 rounded-2xl bg-white/80 dark:bg-white/10"></div>
                  <div class="h-14 rounded-2xl bg-white/80 dark:bg-white/10"></div>
                  <div class="h-14 rounded-2xl bg-white/80 dark:bg-white/10"></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="text-xs text-gray-400 dark:text-dark-500">
          &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
        </div>
      </section>

      <section class="flex min-h-screen items-center justify-center px-4 py-8 sm:px-6">
        <div class="w-full max-w-md">
          <div v-if="settingsLoaded" class="mb-6 text-center lg:hidden">
            <div class="auth-logo-box mb-3 inline-flex h-12 w-12">
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
              {{ siteName }}
            </h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ siteSubtitle }}
            </p>
          </div>

          <div class="auth-panel p-6 sm:p-8">
            <slot />
          </div>

          <div class="mt-6 text-center text-sm">
            <slot name="footer" />
          </div>

          <div class="mt-8 text-center text-xs text-gray-400 dark:text-dark-500 lg:hidden">
            &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
section {
  border-color: var(--ios-border);
}

.auth-logo-box {
  @apply flex items-center justify-center overflow-hidden border;
  background: var(--ios-surface);
  border-color: var(--ios-border);
  border-radius: 16px;
  box-shadow: var(--ios-shadow-soft);
}

.auth-preview-card,
.auth-panel {
  @apply border;
  background: var(--ios-surface);
  border-color: var(--ios-border);
  border-radius: var(--ios-radius);
  box-shadow: var(--ios-shadow-soft);
  backdrop-filter: blur(22px) saturate(1.35);
  -webkit-backdrop-filter: blur(22px) saturate(1.35);
}
</style>
