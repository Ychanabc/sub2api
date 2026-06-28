<template>
  <AppLayout>
    <div class="audit-shell">
      <section v-if="!unlocked" class="audit-gate">
        <div class="audit-gate-panel">
          <div>
            <p class="audit-kicker">{{ t('admin.conversationAudit.secureArea') }}</p>
            <h1 class="audit-title">{{ t('admin.conversationAudit.title') }}</h1>
            <p class="audit-copy">{{ t('admin.conversationAudit.unlockHint') }}</p>
          </div>

          <div v-if="status && !status.secondary_password_configured" class="audit-warning">
            {{ t('admin.conversationAudit.passwordNotConfigured') }}
          </div>
          <div v-else-if="status && !status.totp_enabled" class="audit-warning">
            {{ t('admin.conversationAudit.totpNotEnabled') }}
          </div>

          <div class="audit-gate-form">
            <label class="field-label">{{ t('admin.conversationAudit.secondaryPassword') }}</label>
            <input
              v-model="unlockForm.secondary_password"
              class="input"
              type="password"
              autocomplete="current-password"
              @keyup.enter="unlockAudit"
            />
            <label class="field-label">{{ t('admin.conversationAudit.totpCode') }}</label>
            <input
              v-model="unlockForm.totp_code"
              class="input tracking-[0.3em]"
              inputmode="numeric"
              maxlength="6"
              autocomplete="one-time-code"
              @keyup.enter="unlockAudit"
            />
          </div>

          <button class="btn btn-primary h-11 w-full" :disabled="unlocking" @click="unlockAudit">
            <Icon name="shield" size="md" :class="unlocking ? 'animate-pulse' : ''" />
            <span>{{ t('admin.conversationAudit.unlock') }}</span>
          </button>
        </div>
      </section>

      <TablePageLayout v-else>
        <template #filters>
          <div class="flex flex-wrap items-center gap-3">
            <div class="segmented">
              <button :class="{ active: activeTab === 'records' }" @click="activeTab = 'records'">
                {{ t('admin.conversationAudit.records') }}
              </button>
              <button :class="{ active: activeTab === 'accessLogs' }" @click="activeTab = 'accessLogs'">
                {{ t('admin.conversationAudit.accessLogs') }}
              </button>
            </div>
            <template v-if="activeTab === 'records'">
              <SearchInput
                v-model="filters.q"
                :placeholder="t('admin.conversationAudit.searchPlaceholder')"
                class="w-full sm:w-72"
                @search="loadLogs"
              />
              <input
                v-model="filters.model"
                class="input h-10 w-full sm:w-52"
                :placeholder="t('admin.conversationAudit.modelPlaceholder')"
                @keyup.enter="loadLogs"
              />
            </template>
          </div>
        </template>

        <template #actions>
          <button
            v-if="activeTab === 'records'"
            class="btn btn-secondary"
            :disabled="loading || exporting || logs.length === 0"
            :title="t('admin.conversationAudit.exportJson')"
            @click="exportJson"
          >
            <Icon name="download" size="md" :class="exporting ? 'animate-spin' : ''" />
          </button>
          <button
            v-if="activeTab === 'records'"
            class="btn btn-danger"
            :disabled="loading || clearing"
            :title="t('admin.conversationAudit.clearAll')"
            @click="clearAll"
          >
            <Icon name="trash" size="md" :class="clearing ? 'animate-pulse' : ''" />
          </button>
          <button class="btn btn-secondary" :disabled="loading || accessLoading" @click="refreshActive">
            <Icon name="refresh" size="md" :class="loading || accessLoading ? 'animate-spin' : ''" />
          </button>
        </template>

        <template #table>
          <DataTable v-if="activeTab === 'records'" :columns="columns" :data="logs" :loading="loading">
            <template #cell-request_id="{ value }">
              <code class="code text-xs">{{ value || '-' }}</code>
            </template>
            <template #cell-model="{ value }">
              <span class="font-medium text-gray-900 dark:text-white">{{ value || '-' }}</span>
            </template>
            <template #cell-request_excerpt="{ row }">
              <button class="max-w-[420px] truncate text-left text-sm text-gray-600 hover:text-primary-600 dark:text-dark-300" @click="openDetail(row)">
                {{ row.request_excerpt || '-' }}
              </button>
            </template>
            <template #cell-response_excerpt="{ row }">
              <button class="max-w-[420px] truncate text-left text-sm text-gray-600 hover:text-primary-600 dark:text-dark-300" @click="openDetail(row)">
                {{ row.response_excerpt || '-' }}
              </button>
            </template>
            <template #cell-created_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
            </template>
          </DataTable>

          <DataTable v-else :columns="accessColumns" :data="accessLogs" :loading="accessLoading">
            <template #cell-success="{ value }">
              <span :class="value ? 'status-ok' : 'status-fail'">
                {{ value ? t('common.success') : t('common.failed') }}
              </span>
            </template>
            <template #cell-fingerprint="{ value }">
              <code class="code text-xs">{{ value || '-' }}</code>
            </template>
            <template #cell-user_agent="{ value }">
              <span class="block max-w-[360px] truncate text-sm text-gray-600 dark:text-dark-300">{{ value || '-' }}</span>
            </template>
            <template #cell-created_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
            </template>
          </DataTable>
        </template>

        <template #pagination>
          <Pagination
            v-if="activePagination.total > 0"
            :page="activePagination.page"
            :total="activePagination.total"
            :page-size="activePagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </template>
      </TablePageLayout>
    </div>

    <BaseDialog :show="detailOpen" :title="t('admin.conversationAudit.detailTitle')" width="wide" @close="detailOpen = false">
      <div class="audit-detail">
        <div class="audit-detail-section">
          <div class="audit-detail-label">{{ t('admin.conversationAudit.requestExcerpt') }}</div>
          <pre class="audit-pre">{{ selectedRequest || '-' }}</pre>
        </div>
        <div class="audit-detail-section">
          <div class="audit-detail-label">{{ t('admin.conversationAudit.responseExcerpt') }}</div>
          <pre class="audit-pre">{{ selectedResponse || '-' }}</pre>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  conversationAuditAPI,
  type ConversationAuditAccessLog,
  type ConversationAuditLog,
  type ConversationAuditStatus
} from '@/api/admin/conversationAudit'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const accessLoading = ref(false)
const unlocking = ref(false)
const exporting = ref(false)
const clearing = ref(false)
const unlocked = ref(false)
const unlockToken = ref(sessionStorage.getItem('conversation_audit_unlock_token') || '')
const status = ref<ConversationAuditStatus | null>(null)
const activeTab = ref<'records' | 'accessLogs'>('records')
const logs = ref<ConversationAuditLog[]>([])
const accessLogs = ref<ConversationAuditAccessLog[]>([])
const detailOpen = ref(false)
const selectedRequest = ref('')
const selectedResponse = ref('')
const filters = reactive({ q: '', model: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0, pages: 0 })
const accessPagination = reactive({ page: 1, page_size: 20, total: 0, pages: 0 })
const unlockForm = reactive({ secondary_password: '', totp_code: '' })

const activePagination = computed(() => (activeTab.value === 'records' ? pagination : accessPagination))

const columns = computed(() => [
  { key: 'created_at', label: t('common.createdAt') },
  { key: 'request_id', label: t('admin.conversationAudit.requestId') },
  { key: 'user_id', label: t('admin.conversationAudit.userId') },
  { key: 'api_key_id', label: t('admin.conversationAudit.apiKeyId') },
  { key: 'account_id', label: t('admin.conversationAudit.accountId') },
  { key: 'model', label: t('usage.model') },
  { key: 'endpoint', label: t('admin.conversationAudit.endpoint') },
  { key: 'request_excerpt', label: t('admin.conversationAudit.requestExcerpt') },
  { key: 'response_excerpt', label: t('admin.conversationAudit.responseExcerpt') },
])

const accessColumns = computed(() => [
  { key: 'created_at', label: t('common.createdAt') },
  { key: 'email', label: t('admin.conversationAudit.adminAccount') },
  { key: 'user_id', label: t('admin.conversationAudit.userId') },
  { key: 'ip_address', label: 'IP' },
  { key: 'fingerprint', label: t('admin.conversationAudit.fingerprint') },
  { key: 'user_agent', label: t('admin.conversationAudit.browser') },
  { key: 'success', label: t('admin.conversationAudit.result') },
  { key: 'failure_reason', label: t('admin.conversationAudit.failureReason') },
])

async function loadStatus() {
  status.value = await conversationAuditAPI.status()
}

async function unlockAudit() {
  if (unlocking.value) return
  unlocking.value = true
  try {
    const fingerprint = await getBrowserFingerprint()
    const res = await conversationAuditAPI.unlock({
      secondary_password: unlockForm.secondary_password,
      totp_code: unlockForm.totp_code,
      fingerprint,
    })
    unlockToken.value = res.unlock_token
    sessionStorage.setItem('conversation_audit_unlock_token', res.unlock_token)
    unlocked.value = true
    unlockForm.secondary_password = ''
    unlockForm.totp_code = ''
    await loadLogs()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.conversationAudit.unlockFailed'))
  } finally {
    unlocking.value = false
  }
}

async function loadLogs() {
  if (!unlockToken.value) return
  loading.value = true
  try {
    const res = await conversationAuditAPI.list({
      page: pagination.page,
      page_size: pagination.page_size,
      q: filters.q || undefined,
      model: filters.model || undefined,
    }, unlockToken.value)
    logs.value = res.items
    Object.assign(pagination, res.pagination)
  } catch (error: any) {
    handleLockedError(error)
  } finally {
    loading.value = false
  }
}

async function loadAccessLogs() {
  if (!unlockToken.value) return
  accessLoading.value = true
  try {
    const res = await conversationAuditAPI.accessLogs({
      page: accessPagination.page,
      page_size: accessPagination.page_size,
    }, unlockToken.value)
    accessLogs.value = res.items
    Object.assign(accessPagination, res.pagination)
  } catch (error: any) {
    handleLockedError(error)
  } finally {
    accessLoading.value = false
  }
}

function handleLockedError(error: any) {
  if (error?.status === 401 || String(error?.reason || '').includes('CONVERSATION_AUDIT_TOKEN')) {
    unlocked.value = false
    unlockToken.value = ''
    sessionStorage.removeItem('conversation_audit_unlock_token')
  }
  appStore.showError(error?.message || t('admin.conversationAudit.loadFailed'))
}

function refreshActive() {
  if (activeTab.value === 'records') {
    loadLogs()
  } else {
    loadAccessLogs()
  }
}

function handlePageChange(page: number) {
  activePagination.value.page = page
  refreshActive()
}

function handlePageSizeChange(size: number) {
  activePagination.value.page_size = size
  activePagination.value.page = 1
  refreshActive()
}

function openDetail(row: ConversationAuditLog) {
  selectedRequest.value = row.request_excerpt || ''
  selectedResponse.value = row.response_excerpt || ''
  detailOpen.value = true
}

async function clearAll() {
  if (!unlockToken.value || clearing.value) return
  if (!window.confirm(t('admin.conversationAudit.clearConfirm'))) return
  clearing.value = true
  try {
    const res = await conversationAuditAPI.clearAll(unlockToken.value)
    logs.value = []
    pagination.page = 1
    pagination.total = 0
    pagination.pages = 0
    appStore.showSuccess(t('admin.conversationAudit.clearSuccess', { count: res.deleted }))
  } catch (error: any) {
    handleLockedError(error)
  } finally {
    clearing.value = false
  }
}

async function exportJson() {
  if (exporting.value || !unlockToken.value) return
  exporting.value = true
  try {
    const all: ConversationAuditLog[] = []
    const pageSize = 100
    let page = 1
    const maxPages = 50
    while (page <= maxPages) {
      const res = await conversationAuditAPI.list({
        page,
        page_size: pageSize,
        q: filters.q || undefined,
        model: filters.model || undefined,
      }, unlockToken.value)
      all.push(...res.items)
      const total = res.pagination?.total ?? all.length
      if (all.length >= total || res.items.length === 0) break
      page += 1
    }
    const blob = new Blob([JSON.stringify(all, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const stamp = new Date().toISOString().replace(/[:.]/g, '-')
    a.download = `conversation-audit-${stamp}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (error: any) {
    handleLockedError(error)
  } finally {
    exporting.value = false
  }
}

async function getBrowserFingerprint() {
  const parts = [
    navigator.userAgent,
    navigator.language,
    `${screen.width}x${screen.height}x${screen.colorDepth}`,
    Intl.DateTimeFormat().resolvedOptions().timeZone,
  ]
  const raw = parts.join('|')
  const subtle = globalThis.crypto?.subtle
  if (subtle) {
    const hash = await subtle.digest('SHA-256', new TextEncoder().encode(raw))
    return Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('')
  }
  return btoa(raw).slice(0, 64)
}

watch(activeTab, tab => {
  if (!unlocked.value) return
  if (tab === 'records') loadLogs()
  else loadAccessLogs()
})

onMounted(async () => {
  await loadStatus()
  if (unlockToken.value) {
    unlocked.value = true
    await loadLogs()
  }
})
</script>

<style scoped>
.audit-shell {
  min-height: 100%;
}

.audit-gate {
  display: grid;
  min-height: calc(100vh - 120px);
  place-items: center;
  padding: 24px;
}

.audit-gate-panel {
  width: min(100%, 460px);
  border: 1px solid var(--ios-border);
  border-radius: 28px;
  background: color-mix(in srgb, var(--ios-surface) 88%, transparent);
  box-shadow: var(--ios-shadow-lg);
  backdrop-filter: blur(22px);
  padding: 28px;
}

.audit-kicker {
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0;
  color: rgb(59 130 246);
}

.audit-title {
  font-size: 26px;
  font-weight: 800;
  color: rgb(15 23 42);
}

.dark .audit-title {
  color: white;
}

.audit-copy {
  margin-top: 10px;
  color: rgb(100 116 139);
  line-height: 1.7;
}

.dark .audit-copy {
  color: rgb(148 163 184);
}

.audit-warning {
  margin-top: 18px;
  border: 1px solid rgba(245, 158, 11, 0.28);
  border-radius: 18px;
  background: rgba(245, 158, 11, 0.1);
  padding: 12px 14px;
  color: rgb(180 83 9);
  font-size: 13px;
}

.audit-gate-form {
  margin: 22px 0;
  display: grid;
  gap: 10px;
}

.field-label {
  font-size: 13px;
  font-weight: 700;
  color: rgb(71 85 105);
}

.dark .field-label {
  color: rgb(203 213 225);
}

.segmented {
  display: inline-flex;
  border: 1px solid var(--ios-border);
  border-radius: 999px;
  background: var(--ios-surface);
  padding: 4px;
  box-shadow: var(--ios-shadow-sm);
}

.segmented button {
  min-width: 96px;
  border-radius: 999px;
  padding: 7px 14px;
  font-size: 13px;
  font-weight: 700;
  color: rgb(100 116 139);
}

.segmented button.active {
  background: rgb(37 99 235);
  color: white;
  box-shadow: 0 8px 22px rgba(37, 99, 235, 0.26);
}

.status-ok,
.status-fail {
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
}

.status-ok {
  background: rgba(34, 197, 94, 0.12);
  color: rgb(22 163 74);
}

.status-fail {
  background: rgba(239, 68, 68, 0.12);
  color: rgb(220 38 38);
}

.audit-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.audit-detail-label {
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 600;
  color: rgb(100 116 139);
}

.dark .audit-detail-label {
  color: rgb(148 163 184);
}

.audit-pre {
  max-height: min(68vh, 720px);
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  border: 1px solid var(--ios-border);
  border-radius: 18px;
  background: var(--ios-surface);
  padding: 16px;
  color: rgb(51 65 85);
}

.dark .audit-pre {
  color: rgb(226 232 240);
}
</style>
