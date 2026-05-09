<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <select v-model="dateRange" @change="onDateRangeChange" class="input text-sm w-36">
              <option value="7d">{{ t('admin.upstreamCost.last7d') }}</option>
              <option value="30d">{{ t('admin.upstreamCost.last30d') }}</option>
              <option value="thisMonth">{{ t('admin.upstreamCost.thisMonth') }}</option>
              <option value="lastMonth">{{ t('admin.upstreamCost.lastMonth') }}</option>
              <option value="90d">{{ t('admin.upstreamCost.last90d') }}</option>
            </select>
            <!-- Global P/L badge -->
            <span v-if="globalSummary.hasData" class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold" :class="globalSummary.profitLoss >= 0 ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'">
              {{ t('admin.upstreamCost.profitLoss') }}: {{ globalSummary.profitLoss >= 0 ? '+' : '' }}${{ globalSummary.profitLoss.toFixed(2) }}
              <span class="opacity-70">({{ globalSummary.marginPct >= 0 ? '+' : '' }}{{ globalSummary.marginPct.toFixed(1) }}%)</span>
            </span>
          </div>
          <button @click="fetchAllChannelCosts" :disabled="batchFetching" class="btn btn-primary">
            <Icon name="refresh" size="md" :class="batchFetching ? 'animate-spin' : ''" />
            <span class="ml-1.5">{{ t('admin.upstreamCost.fetchAll') }}</span>
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="tableData" :loading="loadingChannels">
          <template #cell-name="{ row }">
            <div class="flex items-center gap-2">
              <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
              <span :class="row.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-dark-600 dark:text-gray-400'" class="rounded px-1.5 py-0.5 text-[10px] font-medium">{{ row.status }}</span>
            </div>
          </template>

          <template #cell-provider="{ row }">
            <span v-if="row._cfg" class="inline-flex rounded bg-blue-50 px-1.5 py-0.5 text-[10px] font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{{ row._cfg.type === 'newapi' ? 'New-API' : 'Sub2API' }}</span>
            <button v-else @click="openConfigDialog(row._channel)" class="text-[10px] font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">{{ t('admin.upstreamCost.configure') }}</button>
          </template>

          <template #cell-upstream_cost="{ row }">
            <span v-if="row._fetching" class="text-gray-400">
              <Icon name="refresh" size="sm" class="animate-spin" />
            </span>
            <span v-else-if="row._error" class="text-red-500 text-[10px]" :title="row._error">Error</span>
            <span v-else-if="row._result" class="font-medium text-red-600 dark:text-red-400">${{ row._result.upstreamCost.toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-revenue="{ row }">
            <span v-if="row._result" class="font-medium text-blue-600 dark:text-blue-400">${{ row._result.internalRevenue.toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-profit_loss="{ row }">
            <span v-if="row._result" class="font-semibold" :class="row._result.profitLoss >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
              {{ row._result.profitLoss >= 0 ? '+' : '' }}${{ row._result.profitLoss.toFixed(2) }}
            </span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-margin="{ row }">
            <span v-if="row._result" class="inline-flex rounded-full px-1.5 py-0.5 text-[10px] font-semibold" :class="row._result.profitLoss >= 0 ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'">
              {{ row._result.marginPct >= 0 ? '+' : '' }}{{ row._result.marginPct.toFixed(1) }}%
            </span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-balance="{ row }">
            <span v-if="row._result" class="font-medium text-emerald-600 dark:text-emerald-400">${{ (row._result.userInfo.quota / 500000).toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button @click="openConfigDialog(row._channel)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400">
                <Icon name="settings" size="sm" />
                <span class="text-[10px]">{{ t('admin.upstreamCost.configure') }}</span>
              </button>
              <button @click="fetchSingleChannelCost(row.id)" :disabled="row._fetching || !row._cfg" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 disabled:opacity-30 dark:hover:bg-dark-700 dark:hover:text-primary-400">
                <Icon name="refresh" size="sm" :class="row._fetching ? 'animate-spin' : ''" />
                <span class="text-[10px]">{{ t('admin.upstreamCost.fetchData') }}</span>
              </button>
              <button v-if="row._result" @click="openDetailDialog(row.id)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400">
                <Icon name="chart" size="sm" />
                <span class="text-[10px]">{{ t('admin.upstreamCost.viewDetail') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState :title="t('admin.upstreamCost.noChannels')" />
          </template>
        </DataTable>
      </template>
    </TablePageLayout>

    <!-- ====== Config Dialog ====== -->
    <BaseDialog :show="showConfigDialog" :title="t('admin.upstreamCost.configure') + ' — ' + (configChannel?.name || '')" @close="showConfigDialog = false">
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.providerType') }}</label>
            <select v-model="configForm.type" class="input w-full text-sm">
              <option value="newapi">New-API</option>
              <option value="sub2api">Sub2API</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.userId') }}</label>
            <input v-model.number="configForm.user_id" type="number" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.userIdPlaceholder')" />
          </div>
        </div>
        <div>
          <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.baseUrl') }}</label>
          <input v-model="configForm.base_url" type="text" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.baseUrlPlaceholder')" />
        </div>
        <div>
          <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.accessToken') }}</label>
          <input v-model="configForm.access_token" type="password" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.accessTokenPlaceholder')" />
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button @click="showConfigDialog = false" class="btn btn-secondary">{{ t('common.cancel', 'Cancel') }}</button>
          <button @click="saveConfig" :disabled="savingConfig" class="btn btn-primary">
            <Icon v-if="savingConfig" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ t('admin.upstreamCost.saveConfig') }}
          </button>
        </div>
      </div>
    </BaseDialog>

    <!-- ====== Detail Dialog ====== -->
    <BaseDialog :show="showDetailDialog" :title="detailChannel?.name + ' — ' + t('admin.upstreamCost.costDetail')" width="extra-wide" @close="showDetailDialog = false">
      <div v-if="detailResult" class="space-y-5">
        <!-- P&L Summary -->
        <div class="grid grid-cols-2 gap-3 md:grid-cols-5">
          <div class="rounded-lg bg-red-50 p-3 dark:bg-red-900/20">
            <p class="text-[10px] text-red-500">{{ t('admin.upstreamCost.upstreamCostLabel') }}</p>
            <p class="text-base font-bold text-red-600 dark:text-red-400">${{ detailResult.upstreamCost.toFixed(4) }}</p>
          </div>
          <div class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/20">
            <p class="text-[10px] text-blue-500">{{ t('admin.upstreamCost.internalRevenueLabel') }}</p>
            <p class="text-base font-bold text-blue-600 dark:text-blue-400">${{ detailResult.internalRevenue.toFixed(4) }}</p>
          </div>
          <div class="rounded-lg p-3" :class="detailResult.profitLoss >= 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'">
            <p class="text-[10px]" :class="detailResult.profitLoss >= 0 ? 'text-emerald-500' : 'text-red-500'">{{ t('admin.upstreamCost.profitLoss') }}</p>
            <p class="text-base font-bold" :class="detailResult.profitLoss >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
              {{ detailResult.profitLoss >= 0 ? '+' : '' }}${{ detailResult.profitLoss.toFixed(4) }}
            </p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
            <p class="text-[10px] text-gray-400">{{ t('admin.upstreamCost.availableQuota') }}</p>
            <p class="text-base font-bold text-emerald-600 dark:text-emerald-400">${{ (detailResult.userInfo.quota / 500000).toFixed(4) }}</p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
            <p class="text-[10px] text-gray-400">{{ t('admin.upstreamCost.requestCount') }}</p>
            <p class="text-base font-bold text-gray-900 dark:text-white">{{ detailResult.userInfo.request_count.toLocaleString() }}</p>
          </div>
        </div>

        <!-- Daily Trend -->
        <div v-if="detailTrend.length > 0" class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
          <h3 class="mb-3 text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('admin.upstreamCost.dailyTrend') }}</h3>
          <div class="flex items-end gap-0.5" style="height: 100px">
            <div v-for="(pt, idx) in detailTrend" :key="idx" class="relative flex flex-1 flex-col items-center justify-end" style="min-width: 0">
              <div class="w-full rounded-t bg-blue-400 dark:bg-blue-500 transition-all" :style="{ height: pt.pct + '%', minHeight: pt.val > 0 ? '2px' : '0' }" :title="pt.label + ': $' + pt.val.toFixed(4)" />
              <span v-if="idx % Math.max(1, Math.floor(detailTrend.length / 7)) === 0" class="mt-1 text-[7px] text-gray-400">{{ pt.label }}</span>
            </div>
          </div>
        </div>

        <!-- Model Breakdown -->
        <div v-if="modelBreakdown.length > 0" class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
          <h3 class="mb-3 text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('admin.upstreamCost.modelBreakdown') }}</h3>
          <div class="space-y-1.5">
            <div v-for="(item, idx) in modelBreakdown" :key="idx" class="flex items-center gap-2">
              <span class="w-32 truncate text-[11px] text-gray-600 dark:text-gray-400" :title="item.model">{{ item.model }}</span>
              <div class="flex-1 h-4 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-600">
                <div class="h-full rounded-full bg-gradient-to-r from-primary-400 to-primary-600 transition-all" :style="{ width: item.pct + '%', minWidth: item.pct > 0 ? '1rem' : '0' }" />
              </div>
              <span class="w-16 text-right text-[11px] font-medium text-amber-600 dark:text-amber-400">${{ item.cost.toFixed(4) }}</span>
            </div>
          </div>
        </div>

        <!-- Usage Logs -->
        <div class="rounded-lg border border-gray-200 dark:border-dark-600">
          <div class="border-b border-gray-200 px-4 py-2 dark:border-dark-600">
            <h3 class="text-xs font-semibold text-gray-700 dark:text-gray-300">
              {{ t('admin.upstreamCost.usageLogs') }}
              <span v-if="detailLogs.total > 0" class="ml-1 font-normal text-gray-400">({{ detailLogs.total }})</span>
            </h3>
          </div>
          <div class="max-h-64 overflow-auto">
            <table class="w-full text-left text-[11px]">
              <thead class="sticky top-0 border-b border-gray-100 bg-gray-50 dark:border-dark-600 dark:bg-dark-700">
                <tr>
                  <th class="px-3 py-1.5 font-medium text-gray-500">{{ t('admin.upstreamCost.time') }}</th>
                  <th class="px-3 py-1.5 font-medium text-gray-500">{{ t('admin.upstreamCost.model') }}</th>
                  <th class="px-3 py-1.5 font-medium text-gray-500">Tokens</th>
                  <th class="px-3 py-1.5 font-medium text-gray-500">{{ t('admin.upstreamCost.quotaUSD') }}</th>
                  <th class="px-3 py-1.5 font-medium text-gray-500">{{ t('admin.upstreamCost.tokenName') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-600">
                <tr v-for="log in detailLogs.items" :key="log.id">
                  <td class="whitespace-nowrap px-3 py-1 text-gray-500">{{ formatTime(log.created_at) }}</td>
                  <td class="px-3 py-1"><span class="rounded bg-blue-50 px-1 py-0.5 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{{ log.model_name }}</span></td>
                  <td class="px-3 py-1 text-gray-500">{{ (log.prompt_tokens + log.completion_tokens).toLocaleString() }}</td>
                  <td class="px-3 py-1 font-medium text-amber-600 dark:text-amber-400">${{ log.quota_usd.toFixed(6) }}</td>
                  <td class="px-3 py-1 text-gray-400">{{ log.token_name }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="detailLogs.items.length < detailLogs.total" class="border-t border-gray-200 px-4 py-2 text-center dark:border-dark-600">
            <button @click="loadMoreDetailLogs" :disabled="detailLogsLoading" class="text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
              {{ detailLogsLoading ? t('admin.upstreamCost.fetching') : t('admin.upstreamCost.loadMore') }}
            </button>
          </div>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/common/Icon.vue'
import { adminAPI } from '@/api/admin'
import { upstreamCostAPI } from '@/api/upstream-cost'
import type { UpstreamProvider, UpstreamUserInfo, UpstreamStat, UpstreamLogItem } from '@/api/upstream-cost'
import type { Channel } from '@/api/admin/channels'
import type { TrendDataPoint } from '@/types'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

// --- Date Range ---
const dateRange = ref('30d')

function getDateRange(): { start_date: string; end_date: string } {
  const now = new Date()
  const end = now.toISOString().slice(0, 10)
  let start: string
  switch (dateRange.value) {
    case '7d': { const d = new Date(now); d.setDate(d.getDate() - 7); start = d.toISOString().slice(0, 10); break }
    case '30d': { const d = new Date(now); d.setDate(d.getDate() - 30); start = d.toISOString().slice(0, 10); break }
    case '90d': { const d = new Date(now); d.setDate(d.getDate() - 90); start = d.toISOString().slice(0, 10); break }
    case 'thisMonth': { start = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`; break }
    case 'lastMonth': {
      const lm = new Date(now.getFullYear(), now.getMonth() - 1, 1)
      start = lm.toISOString().slice(0, 10)
      const lme = new Date(now.getFullYear(), now.getMonth(), 0)
      return { start_date: start, end_date: lme.toISOString().slice(0, 10) }
    }
    default: { const d = new Date(now); d.setDate(d.getDate() - 30); start = d.toISOString().slice(0, 10) }
  }
  return { start_date: start, end_date: end }
}

function onDateRangeChange() {
  if (Object.keys(costResults).length > 0) fetchAllChannelCosts()
}

// --- Channels ---
const channels = ref<Channel[]>([])
const loadingChannels = ref(false)

async function loadChannels() {
  loadingChannels.value = true
  try {
    const resp = await adminAPI.channels.list(1, 100)
    channels.value = resp.items || []
  } catch (err: any) {
    console.error('Failed to load channels:', err)
  } finally {
    loadingChannels.value = false
  }
}

// --- Upstream config from channel.features_config ---
interface UpstreamCfg {
  type: 'newapi' | 'sub2api'
  base_url: string
  access_token: string
  user_id: number
}

function getUpstreamCfg(ch: Channel): UpstreamCfg | null {
  const cfg = (ch.features_config as Record<string, unknown> | undefined)?.upstream_provider as UpstreamCfg | undefined
  if (!cfg || !cfg.base_url || !cfg.access_token || !cfg.user_id) return null
  return cfg
}

function cfgToProvider(cfg: UpstreamCfg): UpstreamProvider {
  return { provider_type: cfg.type || 'newapi', base_url: cfg.base_url, access_token: cfg.access_token, user_id: cfg.user_id }
}

// --- Config Dialog ---
const showConfigDialog = ref(false)
const configChannel = ref<Channel | null>(null)
const configForm = reactive<UpstreamCfg>({ type: 'newapi', base_url: '', access_token: '', user_id: 0 })
const savingConfig = ref(false)

function openConfigDialog(ch: Channel) {
  configChannel.value = ch
  const existing = getUpstreamCfg(ch)
  configForm.type = existing?.type || 'newapi'
  configForm.base_url = existing?.base_url || ''
  configForm.access_token = existing?.access_token || ''
  configForm.user_id = existing?.user_id || 0
  showConfigDialog.value = true
}

async function saveConfig() {
  if (!configChannel.value) return
  savingConfig.value = true
  try {
    const ch = configChannel.value
    const existingFeatures = (ch.features_config || {}) as Record<string, unknown>
    await adminAPI.channels.update(ch.id, {
      features_config: { ...existingFeatures, upstream_provider: { ...configForm } }
    })
    appStore.showSuccess(t('admin.upstreamCost.configSaved'))
    showConfigDialog.value = false
    await loadChannels()
  } catch (err: any) {
    appStore.showError(err?.message || 'Failed to save config')
  } finally {
    savingConfig.value = false
  }
}

// --- Cost results ---
interface ChannelCostResult {
  userInfo: UpstreamUserInfo
  stat: UpstreamStat
  upstreamCost: number
  internalRevenue: number
  profitLoss: number
  marginPct: number
  trend: TrendDataPoint[]
}

const costResults = reactive<Record<number, ChannelCostResult>>({})
const fetchingChannels = reactive<Record<number, boolean>>({})
const fetchErrors = reactive<Record<number, string>>({})
const batchFetching = ref(false)

const globalSummary = computed(() => {
  const results = Object.values(costResults)
  if (results.length === 0) return { hasData: false, upstreamCost: 0, internalRevenue: 0, profitLoss: 0, marginPct: 0 }
  const upstreamCost = results.reduce((s, r) => s + r.upstreamCost, 0)
  const internalRevenue = results.reduce((s, r) => s + r.internalRevenue, 0)
  const profitLoss = internalRevenue - upstreamCost
  const marginPct = upstreamCost > 0 ? (profitLoss / upstreamCost) * 100 : 0
  return { hasData: true, upstreamCost, internalRevenue, profitLoss, marginPct }
})

async function fetchSingleChannelCost(channelId: number) {
  const ch = channels.value.find(c => c.id === channelId)
  if (!ch) return
  const cfg = getUpstreamCfg(ch)
  if (!cfg) return
  fetchingChannels[channelId] = true
  delete fetchErrors[channelId]
  try {
    const prov = cfgToProvider(cfg)
    const dr = getDateRange()
    const groupIds = ch.group_ids || []
    const [ui, st, ...trendResults] = await Promise.all([
      upstreamCostAPI.getUserInfo(prov),
      upstreamCostAPI.getStats(prov),
      ...groupIds.map(gid =>
        adminAPI.dashboard.getUsageTrend({ ...dr, granularity: 'day', group_id: gid }).catch(() => ({ trend: [] as TrendDataPoint[] }))
      ),
    ])
    const trendMap = new Map<string, TrendDataPoint>()
    for (const tr of trendResults) {
      const resp = tr as { trend: TrendDataPoint[] }
      for (const pt of resp.trend || []) {
        const existing = trendMap.get(pt.date)
        if (existing) {
          existing.requests += pt.requests; existing.input_tokens += pt.input_tokens
          existing.output_tokens += pt.output_tokens; existing.total_tokens += pt.total_tokens
          existing.cost += pt.cost; existing.actual_cost += pt.actual_cost
        } else { trendMap.set(pt.date, { ...pt }) }
      }
    }
    const mergedTrend = [...trendMap.values()].sort((a, b) => a.date.localeCompare(b.date))
    const upstreamCost = st.quota_usd
    const internalRevenue = mergedTrend.reduce((s, pt) => s + pt.actual_cost, 0)
    const profitLoss = internalRevenue - upstreamCost
    const marginPct = upstreamCost > 0 ? (profitLoss / upstreamCost) * 100 : 0
    costResults[channelId] = { userInfo: ui, stat: st, upstreamCost, internalRevenue, profitLoss, marginPct, trend: mergedTrend }
  } catch (err: any) {
    fetchErrors[channelId] = err?.message || 'Failed'
  } finally {
    fetchingChannels[channelId] = false
  }
}

async function fetchAllChannelCosts() {
  batchFetching.value = true
  const configured = channels.value.filter(ch => !!getUpstreamCfg(ch))
  await Promise.allSettled(configured.map(ch => fetchSingleChannelCost(ch.id)))
  batchFetching.value = false
  if (configured.length === 0) appStore.showError(t('admin.upstreamCost.noConfiguredChannels'))
}

// --- Table ---
const columns = computed(() => [
  { key: 'name', label: t('admin.upstreamCost.channelName') },
  { key: 'provider', label: t('admin.upstreamCost.providerType') },
  { key: 'upstream_cost', label: t('admin.upstreamCost.upstreamCostLabel') },
  { key: 'revenue', label: t('admin.upstreamCost.internalRevenueLabel') },
  { key: 'profit_loss', label: t('admin.upstreamCost.profitLoss') },
  { key: 'margin', label: t('admin.upstreamCost.margin') },
  { key: 'balance', label: t('admin.upstreamCost.balance') },
  { key: 'actions', label: '' },
])

const tableData = computed(() =>
  channels.value.map(ch => ({
    id: ch.id,
    name: ch.name,
    status: ch.status,
    _channel: ch,
    _cfg: getUpstreamCfg(ch),
    _result: costResults[ch.id] || null,
    _fetching: !!fetchingChannels[ch.id],
    _error: fetchErrors[ch.id] || null,
  }))
)

// --- Detail Dialog ---
const showDetailDialog = ref(false)
const detailChannelId = ref<number | null>(null)
const detailChannel = computed(() => channels.value.find(ch => ch.id === detailChannelId.value))
const detailResult = computed(() => detailChannelId.value !== null ? costResults[detailChannelId.value] : null)
const detailLogs = reactive<{ items: UpstreamLogItem[]; total: number; page: number }>({ items: [], total: 0, page: 0 })
const detailLogsLoading = ref(false)

const detailTrend = computed(() => {
  if (!detailResult.value?.trend.length) return []
  const trend = detailResult.value.trend
  const maxVal = Math.max(...trend.map(pt => Math.max(pt.actual_cost, 0.001)))
  return trend.map(pt => ({ label: pt.date.slice(5), val: pt.actual_cost, pct: maxVal > 0 ? (pt.actual_cost / maxVal) * 100 : 0 }))
})

const modelBreakdown = computed(() => {
  if (!detailLogs.items.length) return []
  const map = new Map<string, number>()
  for (const log of detailLogs.items) map.set(log.model_name, (map.get(log.model_name) || 0) + log.quota_usd)
  const sorted = [...map.entries()].sort((a, b) => b[1] - a[1])
  const total = sorted.reduce((s, [, v]) => s + v, 0)
  return sorted.map(([model, cost]) => ({ model, cost, pct: total > 0 ? (cost / total) * 100 : 0 }))
})

async function openDetailDialog(channelId: number) {
  detailChannelId.value = channelId
  detailLogs.items = []; detailLogs.total = 0; detailLogs.page = 0
  showDetailDialog.value = true
  detailLogsLoading.value = true
  try {
    const ch = channels.value.find(c => c.id === channelId)
    const cfg = ch ? getUpstreamCfg(ch) : null
    if (!cfg) return
    const lg = await upstreamCostAPI.getLogs(cfgToProvider(cfg), { page: 0, page_size: 50 })
    detailLogs.items = lg.items || []; detailLogs.total = lg.total
  } catch (err: any) { appStore.showError(err?.message || 'Failed to load logs') }
  finally { detailLogsLoading.value = false }
}

async function loadMoreDetailLogs() {
  if (detailChannelId.value === null) return
  const ch = channels.value.find(c => c.id === detailChannelId.value)
  const cfg = ch ? getUpstreamCfg(ch) : null
  if (!cfg) return
  detailLogsLoading.value = true; detailLogs.page++
  try {
    const lg = await upstreamCostAPI.getLogs(cfgToProvider(cfg), { page: detailLogs.page, page_size: 50 })
    detailLogs.items.push(...(lg.items || [])); detailLogs.total = lg.total
  } catch (err: any) { appStore.showError(err?.message || 'Failed'); detailLogs.page-- }
  finally { detailLogsLoading.value = false }
}

function formatTime(ts: number): string {
  const d = new Date(ts * 1000)
  return `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
}

onMounted(() => loadChannels())
</script>
