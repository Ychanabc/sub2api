<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="handleCancel">
    <div class="flex gap-4">
      <div
        :class="[
          'flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg',
          danger
            ? 'bg-red-50 text-red-600 dark:bg-red-950/40 dark:text-red-300'
            : 'bg-primary-50 text-primary-700 dark:bg-primary-950/40 dark:text-primary-300'
        ]"
      >
        <Icon :name="danger ? 'exclamationTriangle' : 'infoCircle'" size="md" />
      </div>
      <div class="min-w-0 space-y-3">
        <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">{{ message }}</p>
        <slot></slot>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button
          @click="handleCancel"
          type="button"
          class="btn btn-secondary btn-md"
        >
          {{ cancelText }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :class="[
            'btn btn-md',
            danger ? 'btn-danger' : 'btn-primary'
          ]"
        >
          {{ confirmText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  danger: false
})

const confirmText = computed(() => props.confirmText || t('common.confirm'))
const cancelText = computed(() => props.cancelText || t('common.cancel'))

const emit = defineEmits<Emits>()

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
}
</script>
