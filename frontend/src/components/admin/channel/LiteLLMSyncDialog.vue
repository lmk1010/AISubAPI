<template>
  <BaseDialog :show="show" :title="t('admin.channels.litellm.title')" width="wide" @close="emit('close')">
    <div class="space-y-3">
      <!-- Subtitle: channel + platform -->
      <p class="text-xs text-gray-500 dark:text-dark-400">
        {{ t('admin.channels.litellm.subtitle', { channel: channelName, platform }) }}
      </p>

      <!-- Search + bulk-action bar -->
      <div class="flex flex-wrap items-center gap-2">
        <div class="relative flex-1 min-w-[200px]">
          <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            v-model="search"
            type="text"
            class="input pl-10 text-xs"
            :placeholder="t('admin.channels.litellm.searchPlaceholder')"
          />
        </div>
        <button type="button" @click="toggleSelectVisible" class="btn btn-secondary btn-sm">
          {{ allVisibleSelected ? t('admin.channels.litellm.unselectAll') : t('admin.channels.litellm.selectAllVisible') }}
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-8">
        <div class="spinner h-6 w-6 border-2 border-gray-300 border-t-blue-600"></div>
      </div>

      <!-- Empty -->
      <div v-else-if="filtered.length === 0" class="rounded-lg border border-dashed border-gray-300 p-6 text-center text-xs text-gray-500 dark:border-dark-600">
        {{ items.length === 0 ? t('admin.channels.litellm.empty', { platform }) : t('admin.channels.litellm.noMatch') }}
      </div>

      <!-- Table -->
      <div v-else class="max-h-[60vh] overflow-y-auto rounded-lg border border-gray-200 dark:border-dark-700">
        <table class="w-full text-[12px]">
          <thead class="sticky top-0 bg-gray-50 dark:bg-dark-800">
            <tr class="text-left text-[11px] uppercase tracking-wide text-gray-500 dark:text-dark-400">
              <th class="w-8 py-2 pl-3"></th>
              <th class="py-2 pr-3">{{ t('admin.channels.litellm.cols.model') }}</th>
              <th class="py-2 pr-3">{{ t('admin.channels.litellm.cols.mode') }}</th>
              <th class="py-2 pr-3 text-right">{{ t('admin.channels.litellm.cols.input') }}</th>
              <th class="py-2 pr-3 text-right">{{ t('admin.channels.litellm.cols.cache') }}</th>
              <th class="py-2 pr-3 text-right">{{ t('admin.channels.litellm.cols.output') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr
              v-for="row in filtered"
              :key="row.model"
              :class="row.already_configured ? 'opacity-50' : 'hover:bg-violet-50/40 dark:hover:bg-dark-800/40'"
            >
              <td class="py-1.5 pl-3">
                <input
                  type="checkbox"
                  :checked="selected.has(row.model)"
                  :disabled="row.already_configured"
                  @change="toggleOne(row.model)"
                  class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 disabled:cursor-not-allowed"
                />
              </td>
              <td class="py-1.5 pr-3 font-mono text-[12px] text-gray-900 dark:text-gray-100">
                {{ row.model }}
                <span v-if="row.already_configured" class="ml-1 rounded bg-gray-100 px-1.5 py-0.5 text-[10px] text-gray-500 dark:bg-dark-700 dark:text-dark-400">
                  {{ t('admin.channels.litellm.alreadyConfigured') }}
                </span>
              </td>
              <td class="py-1.5 pr-3 text-[11px] text-gray-500 dark:text-dark-400">{{ row.mode || '—' }}</td>
              <td class="py-1.5 pr-3 text-right font-mono text-[11px]">{{ formatPerMillion(row.input_price) }}</td>
              <td class="py-1.5 pr-3 text-right font-mono text-[11px]">{{ formatPerMillion(row.cache_read_price) }}</td>
              <td class="py-1.5 pr-3 text-right font-mono text-[11px]">{{ formatPerMillion(row.output_price) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="emit('close')">
        {{ t('common.cancel') }}
      </button>
      <button
        type="button"
        class="btn btn-primary"
        :disabled="selected.size === 0 || importing"
        @click="onImport"
      >
        <Icon v-if="importing" name="refresh" size="sm" class="animate-spin" />
        {{ t('admin.channels.litellm.importN', { n: selected.size }) }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  getLiteLLMSuggestions,
  importLiteLLMModels,
  type LiteLLMSuggestionItem
} from '@/api/admin/channels'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  show: boolean
  channelId: number
  channelName: string
  platform: string
}>()

const emit = defineEmits<{
  close: []
  imported: [{ imported: string[]; skipped: string[]; missing: string[] }]
}>()

const { t } = useI18n()
const appStore = useAppStore()

const items = ref<LiteLLMSuggestionItem[]>([])
const loading = ref(false)
const importing = ref(false)
const search = ref('')
const selected = ref(new Set<string>())

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter((it) => it.model.toLowerCase().includes(q))
})

const allVisibleSelected = computed(() => {
  const eligible = filtered.value.filter((it) => !it.already_configured)
  if (eligible.length === 0) return false
  return eligible.every((it) => selected.value.has(it.model))
})

function formatPerMillion(value: number): string {
  if (!value || value === 0) return '—'
  const scaled = value * 1_000_000
  return `$${scaled.toPrecision(10).replace(/\.?0+$/, '')}`
}

function toggleOne(model: string) {
  const next = new Set(selected.value)
  if (next.has(model)) next.delete(model)
  else next.add(model)
  selected.value = next
}

function toggleSelectVisible() {
  const eligible = filtered.value.filter((it) => !it.already_configured)
  const next = new Set(selected.value)
  if (allVisibleSelected.value) {
    eligible.forEach((it) => next.delete(it.model))
  } else {
    eligible.forEach((it) => next.add(it.model))
  }
  selected.value = next
}

async function load() {
  loading.value = true
  selected.value = new Set()
  try {
    const resp = await getLiteLLMSuggestions(props.channelId, props.platform)
    items.value = resp.items
  } catch (e: any) {
    appStore.showError(e?.message || t('admin.channels.litellm.loadFailed'))
    items.value = []
  } finally {
    loading.value = false
  }
}

async function onImport() {
  if (selected.value.size === 0) return
  importing.value = true
  try {
    const result = await importLiteLLMModels(
      props.channelId,
      props.platform,
      Array.from(selected.value)
    )
    const parts: string[] = []
    if (result.imported.length > 0) parts.push(t('admin.channels.litellm.toastImported', { n: result.imported.length }))
    if (result.skipped.length > 0) parts.push(t('admin.channels.litellm.toastSkipped', { n: result.skipped.length }))
    if (result.missing.length > 0) parts.push(t('admin.channels.litellm.toastMissing', { n: result.missing.length }))
    appStore.showSuccess(parts.join(' · '))
    emit('imported', result)
    emit('close')
  } catch (e: any) {
    appStore.showError(e?.message || t('admin.channels.litellm.importFailed'))
  } finally {
    importing.value = false
  }
}

watch(
  () => [props.show, props.channelId, props.platform] as const,
  ([open]) => {
    if (open && props.channelId && props.platform) {
      load()
    }
  },
  { immediate: true }
)
</script>
