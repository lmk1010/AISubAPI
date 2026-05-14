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
            <!-- Exchange Rate -->
            <div class="flex items-center gap-1.5">
              <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.upstreamCost.exchangeRate') }}</label>
              <input v-model.number="exchangeRate" type="number" step="0.01" min="0.01" class="input w-20 text-xs text-center" />
            </div>
            <!-- Global P/L badge -->
            <span v-if="globalSummary.hasData" class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold" :class="globalSummary.profitRMB >= 0 ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'">
              {{ t('admin.upstreamCost.profitLoss') }}: {{ globalSummary.profitRMB >= 0 ? '+' : '' }}¥{{ globalSummary.profitRMB.toFixed(2) }}
              <span class="opacity-70">({{ globalSummary.marginPct >= 0 ? '+' : '' }}{{ globalSummary.marginPct.toFixed(1) }}%)</span>
            </span>
            <span v-if="globalSummary.hasData" class="inline-flex items-center gap-1.5 rounded-full bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
              {{ t('admin.upstreamCost.upstreamCostRMB') }}: ¥{{ globalSummary.upstreamCostRMB.toFixed(2) }} / {{ t('admin.upstreamCost.revenueRMB') }}: ¥{{ globalSummary.internalBillingRMB.toFixed(2) }}
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
            <div>
              <div class="flex items-center gap-2">
                <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
                <span :class="row.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-dark-600 dark:text-gray-400'" class="rounded px-1.5 py-0.5 text-[10px] font-medium">{{ row.status }}</span>
              </div>
              <div v-if="row._localAccount" class="mt-0.5 text-[10px] text-gray-400">
                {{ t('admin.upstreamCost.accountMultiplier') }} {{ row._localAccount.rate_multiplier.toFixed(4) }}x
              </div>
            </div>
          </template>

          <template #cell-pool="{ row }">
            <div v-if="row._pool" class="max-w-48">
              <div class="truncate text-xs font-medium text-gray-700 dark:text-gray-300" :title="row._pool.pool_key">{{ row._pool.pool_name }}</div>
              <div class="text-[10px] text-gray-400">
                {{ row._pool.account_count > 1 ? t('admin.upstreamCost.sharedPool', { count: row._pool.account_count }) : t('admin.upstreamCost.singlePool') }}
              </div>
            </div>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-provider="{ row }">
            <span v-if="row._cfg" class="inline-flex rounded bg-blue-50 px-1.5 py-0.5 text-[10px] font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{{ row._cfg.type === 'newapi' ? 'New-API' : 'Sub2API' }}</span>
            <button v-else @click="openConfigDialog(row._account)" class="text-[10px] font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">{{ t('admin.upstreamCost.configure') }}</button>
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

          <template #cell-upstream_cost_rmb="{ row }">
            <span v-if="row._fetching" class="text-gray-400">
              <Icon name="refresh" size="sm" class="animate-spin" />
            </span>
            <span v-else-if="row._error" class="text-red-500 text-[10px]" :title="row._error">Error</span>
            <span v-else-if="row._result" class="font-medium text-red-600 dark:text-red-400">¥{{ getUpstreamCostRMB(row._result).toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-revenue_rmb="{ row }">
            <span v-if="row._result" class="font-medium text-blue-600 dark:text-blue-400">¥{{ getRevenueRMB(row._result).toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-actual_payment_rmb="{ row }">
            <span v-if="row._subRevenue > 0" class="font-medium text-purple-600 dark:text-purple-400">¥{{ row._subRevenue.toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-profit_rmb="{ row }">
            <span v-if="row._result" class="font-semibold" :class="getProfitRMB(row._result) >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
              {{ getProfitRMB(row._result) >= 0 ? '+' : '' }}¥{{ getProfitRMB(row._result).toFixed(2) }}
            </span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-margin="{ row }">
            <span v-if="row._result" class="inline-flex rounded-full px-1.5 py-0.5 text-[10px] font-semibold" :class="getProfitRMB(row._result) >= 0 ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'">
              {{ getMarginPct(row._result) >= 0 ? '+' : '' }}{{ getMarginPct(row._result).toFixed(1) }}%
            </span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-balance="{ row }">
            <span v-if="row._pool" class="font-medium text-gray-600 dark:text-gray-300">{{ row._pool.requests.toLocaleString() }} req</span>
            <span v-else-if="row._result" class="font-medium text-emerald-600 dark:text-emerald-400">${{ (row._result.userInfo.quota / 500000).toFixed(2) }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button @click="openConfigDialog(row._account)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400">
                <Icon name="cog" size="sm" />
                <span class="text-[10px]">{{ t('admin.upstreamCost.configure') }}</span>
              </button>
              <button @click="fetchSingleChannelCost(row.id)" :disabled="row._fetching || !row._cfg || (!localSummary && !hasRemoteCredentials(row._cfg))" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 disabled:opacity-30 dark:hover:bg-dark-700 dark:hover:text-primary-400">
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
    <BaseDialog :show="showConfigDialog" :title="t('admin.upstreamCost.configure') + ' — ' + (configAccount?.name || '')" @close="showConfigDialog = false">
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.providerType') }}</label>
            <select v-model="configForm.type" class="input w-full text-sm">
              <option value="newapi">New-API</option>
              <option value="sub2api">Sub2API</option>
            </select>
          </div>
          <div v-if="configForm.type === 'newapi'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.userId') }}</label>
            <input v-model.number="configForm.user_id" type="number" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.userIdPlaceholder')" />
          </div>
        </div>
        <div>
          <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.baseUrl') }}</label>
          <input v-model="configForm.base_url" type="text" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.baseUrlPlaceholder')" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.poolKey') }}</label>
            <input v-model="configForm.pool_key" type="text" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.poolKeyPlaceholder')" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.poolName') }}</label>
            <input v-model="configForm.pool_name" type="text" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.poolNamePlaceholder')" />
          </div>
        </div>
        <!-- NewAPI: access token -->
        <div v-if="configForm.type === 'newapi'">
          <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.accessToken') }}</label>
          <input v-model="configForm.access_token" type="password" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.accessTokenPlaceholder')" />
        </div>
        <!-- Sub2API: email + password -->
        <template v-if="configForm.type === 'sub2api'">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.email', '邮箱') }}</label>
            <input v-model="configForm.email" type="email" class="input w-full text-sm" placeholder="user@example.com" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.upstreamCost.password', '密码') }}</label>
            <input v-model="configForm.password" type="password" class="input w-full text-sm" :placeholder="t('admin.upstreamCost.passwordPlaceholder', '输入登录密码')" />
          </div>
        </template>
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
            <p class="text-[10px] text-red-500">{{ t('admin.upstreamCost.upstreamCostRMB') }}</p>
            <p class="text-base font-bold text-red-600 dark:text-red-400">¥{{ getUpstreamCostRMB(detailResult).toFixed(4) }}</p>
          </div>
          <div class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/20">
            <p class="text-[10px] text-blue-500">{{ t('admin.upstreamCost.revenueRMB') }}</p>
            <p class="text-base font-bold text-blue-600 dark:text-blue-400">¥{{ getRevenueRMB(detailResult).toFixed(4) }}</p>
          </div>
          <div class="rounded-lg p-3" :class="getProfitRMB(detailResult) >= 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'">
            <p class="text-[10px]" :class="getProfitRMB(detailResult) >= 0 ? 'text-emerald-500' : 'text-red-500'">{{ t('admin.upstreamCost.profitLoss') }}</p>
            <p class="text-base font-bold" :class="getProfitRMB(detailResult) >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
              {{ getProfitRMB(detailResult) >= 0 ? '+' : '' }}¥{{ getProfitRMB(detailResult).toFixed(4) }}
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
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import adminPaymentAPI from '@/api/admin/payment'
import { upstreamCostAPI } from '@/api/upstream-cost'
import type {
  UpstreamProvider,
  UpstreamUserInfo,
  UpstreamStat,
  UpstreamLogItem,
  UpstreamCostLocalSummary,
  UpstreamCostPoolSummary,
  UpstreamCostAccountSummary,
  UpstreamCostModelBreakdown
} from '@/api/upstream-cost'
import type { Account, TrendDataPoint } from '@/types'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

// --- Exchange Rate (USD → CNY) ---
const exchangeRate = ref(7.2)

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
  if (Object.keys(costResults).length > 0 || localSummary.value) fetchAllChannelCosts()
}

// --- Accounts ---
const accounts = ref<Account[]>([])
const loadingChannels = ref(false)

async function loadAccounts() {
  loadingChannels.value = true
  try {
    const resp = await adminAPI.accounts.list(1, 200)
    accounts.value = resp.items || []
  } catch (err: any) {
    console.error('Failed to load accounts:', err)
  } finally {
    loadingChannels.value = false
  }
}

// --- Upstream config from account.extra.upstream_provider ---
interface UpstreamCfg {
  type: 'newapi' | 'sub2api'
  base_url: string
  access_token: string
  user_id: number
  email: string
  password: string
  pool_key?: string
  pool_name?: string
}

function getUpstreamCfg(acc: Account): UpstreamCfg | null {
  const cfg = (acc.extra as Record<string, unknown> | undefined)?.upstream_provider as UpstreamCfg | undefined
  if (!cfg || !cfg.base_url) return null
  return {
    type: cfg.type === 'sub2api' ? 'sub2api' : 'newapi',
    base_url: cfg.base_url,
    access_token: cfg.access_token || '',
    user_id: Number(cfg.user_id || 0),
    email: cfg.email || '',
    password: cfg.password || '',
    pool_key: cfg.pool_key || '',
    pool_name: cfg.pool_name || '',
  }
}

function hasRemoteCredentials(cfg: UpstreamCfg | null): boolean {
  if (!cfg) return false
  if (cfg.type === 'sub2api') return !!cfg.email && !!cfg.password
  return !!cfg.access_token && !!cfg.user_id
}

function cfgToProvider(cfg: UpstreamCfg): UpstreamProvider {
  return {
    provider_type: cfg.type || 'newapi',
    base_url: cfg.base_url,
    access_token: cfg.access_token,
    user_id: cfg.user_id,
    email: cfg.email,
    password: cfg.password,
  }
}

// --- Config Dialog ---
const showConfigDialog = ref(false)
const configAccount = ref<Account | null>(null)
const configForm = reactive<UpstreamCfg>({ type: 'newapi', base_url: '', access_token: '', user_id: 0, email: '', password: '', pool_key: '', pool_name: '' })
const savingConfig = ref(false)

function openConfigDialog(acc: Account) {
  configAccount.value = acc
  const existing = getUpstreamCfg(acc)
  configForm.type = existing?.type || 'newapi'
  configForm.base_url = existing?.base_url || ''
  configForm.access_token = existing?.access_token || ''
  configForm.user_id = existing?.user_id || 0
  configForm.email = existing?.email || ''
  configForm.password = existing?.password || ''
  configForm.pool_key = existing?.pool_key || ''
  configForm.pool_name = existing?.pool_name || ''
  showConfigDialog.value = true
}

async function saveConfig() {
  if (!configAccount.value) return
  savingConfig.value = true
  try {
    const acc = configAccount.value
    const existingExtra = (acc.extra || {}) as Record<string, unknown>
    await adminAPI.accounts.update(acc.id, {
      extra: { ...existingExtra, upstream_provider: { ...configForm } }
    })
    appStore.showSuccess(t('admin.upstreamCost.configSaved'))
    showConfigDialog.value = false
    await loadAccounts()
  } catch (err: any) {
    appStore.showError(err?.message || 'Failed to save config')
  } finally {
    savingConfig.value = false
  }
}

// --- Plan-based conversion: group_id → ¥ per $1 of usage ---
interface PlanConversion { dailyRateRMB: number; dailyLimitUSD: number; factor: number }
const planConversionMap = ref<Map<number, PlanConversion>>(new Map())

async function loadPlans() {
  try {
    const [plansResp, groups] = await Promise.all([
      adminPaymentAPI.getPlans(),
      adminAPI.groups.getAll(),
    ])
    const plans = (plansResp as any)?.data ?? plansResp ?? []
    if (!Array.isArray(plans)) return
    // Build group_id → daily_limit_usd map from groups
    const groupLimits = new Map<number, number>()
    for (const g of (groups || [])) {
      if (g.daily_limit_usd && g.daily_limit_usd > 0) groupLimits.set(g.id, g.daily_limit_usd)
    }
    const map = new Map<number, PlanConversion>()
    for (const p of plans) {
      const dailyLimitUSD = groupLimits.get(p.group_id) || 0
      const dailyRateRMB = p.validity_days > 0 ? p.price / p.validity_days : 0
      const factor = dailyLimitUSD > 0 ? dailyRateRMB / dailyLimitUSD : 0
      map.set(p.group_id, { dailyRateRMB, dailyLimitUSD, factor })
    }
    planConversionMap.value = map
  } catch (err: any) {
    console.error('Failed to load plans:', err)
  }
}

// --- Cost results ---
interface ChannelCostResult {
  userInfo: UpstreamUserInfo
  stat: UpstreamStat
  upstreamCost: number
  internalRevenue: number
  internalBillingRMB: number
  groupActualCosts: Record<number, number>
  profitLoss: number
  marginPct: number
  trend: TrendDataPoint[]
}

const CACHE_KEY = 'upstream-cost-cache-v3'

const costResults = reactive<Record<number, ChannelCostResult>>({})
const fetchingChannels = reactive<Record<number, boolean>>({})
const fetchErrors = reactive<Record<number, string>>({})
const batchFetching = ref(false)
const localSummary = ref<UpstreamCostLocalSummary | null>(null)

// --- Subscription Revenue (actual RMB payments by group) ---
const subscriptionRevenueByGroup = reactive<Record<number, number>>({})

function loadCache() {
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return
    const { data, range } = JSON.parse(raw)
    if (!data || range !== dateRange.value) return
    for (const [k, v] of Object.entries(data)) {
      const result = v as Partial<ChannelCostResult>
      if (typeof result.internalBillingRMB !== 'number' || !result.groupActualCosts) continue
      costResults[Number(k)] = result as ChannelCostResult
    }
  } catch { /* ignore */ }
}

function saveCache() {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({ ts: Date.now(), range: dateRange.value, data: { ...costResults } }))
  } catch { /* ignore */ }
}

const globalSummary = computed(() => {
  if (localSummary.value) {
    const totals = localSummary.value.totals
    const upstreamCost = totals.upstream_cost
    const upstreamCostRMB = upstreamCost * exchangeRate.value
    const internalRevenue = totals.user_cost
    const internalBillingRMB = internalRevenue * exchangeRate.value
    const profitRMB = internalBillingRMB - upstreamCostRMB
    const marginPct = upstreamCostRMB > 0 ? (profitRMB / upstreamCostRMB) * 100 : 0
    return { hasData: localSummary.value.pools.length > 0, upstreamCost, upstreamCostRMB, internalRevenue, internalBillingRMB, marginPct, profitRMB }
  }
  const results = Object.values(costResults)
  if (results.length === 0) return { hasData: false, upstreamCost: 0, upstreamCostRMB: 0, internalRevenue: 0, internalBillingRMB: 0, marginPct: 0, profitRMB: 0 }
  const upstreamCost = results.reduce((s, r) => s + r.upstreamCost, 0)
  const upstreamCostRMB = results.reduce((s, r) => s + getUpstreamCostRMB(r), 0)
  const internalRevenue = results.reduce((s, r) => s + r.internalRevenue, 0)
  const internalBillingRMB = results.reduce((s, r) => s + r.internalBillingRMB, 0)
  const profitRMB = internalBillingRMB - upstreamCostRMB
  const marginPct = upstreamCostRMB > 0 ? (profitRMB / upstreamCostRMB) * 100 : 0
  return { hasData: true, upstreamCost, upstreamCostRMB, internalRevenue, internalBillingRMB, marginPct, profitRMB }
})

function getUpstreamCostRMB(result: ChannelCostResult | null): number {
  if (!result) return 0
  return result.upstreamCost * exchangeRate.value
}

function getRevenueRMB(result: ChannelCostResult | null): number {
  if (!result) return 0
  return result.internalRevenue * exchangeRate.value
}

function getProfitRMB(result: ChannelCostResult | null): number {
  if (!result) return 0
  return getRevenueRMB(result) - getUpstreamCostRMB(result)
}

function getMarginPct(result: ChannelCostResult | null): number {
  if (!result) return 0
  const costRMB = getUpstreamCostRMB(result)
  return costRMB > 0 ? (getProfitRMB(result) / costRMB) * 100 : 0
}

async function fetchSingleChannelCost(accountId: number) {
  if (localSummary.value) {
    fetchingChannels[accountId] = true
    try {
      await fetchLocalSummary()
    } finally {
      fetchingChannels[accountId] = false
    }
    return
  }
  const acc = accounts.value.find(a => a.id === accountId)
  if (!acc) return
  const cfg = getUpstreamCfg(acc)
  if (!cfg) return
  if (!hasRemoteCredentials(cfg)) {
    fetchErrors[accountId] = t('admin.upstreamCost.missingRemoteCredentials')
    return
  }
  fetchingChannels[accountId] = true
  delete fetchErrors[accountId]
  try {
    const prov = cfgToProvider(cfg)
    const dr = getDateRange()
    const groupIds = acc.group_ids || []
    const trendRequests = groupIds.length > 0
      ? groupIds.map(gid =>
        adminAPI.dashboard.getUsageTrend({ ...dr, granularity: 'day', account_id: accountId, group_id: gid }).catch(() => ({ trend: [] as TrendDataPoint[] }))
      )
      : [
        adminAPI.dashboard.getUsageTrend({ ...dr, granularity: 'day', account_id: accountId }).catch(() => ({ trend: [] as TrendDataPoint[] }))
      ]

    const [ui, st, ...trendResults] = await Promise.all([
      upstreamCostAPI.getUserInfo(prov),
      upstreamCostAPI.getStats(prov),
      ...trendRequests,
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
    // Per-group proportional RMB billing based on plan pricing
    let internalBillingRMB = 0
    const groupActualCosts: Record<number, number> = {}
    for (let i = 0; i < groupIds.length; i++) {
      const gid = groupIds[i]
      const resp = trendResults[i] as { trend: TrendDataPoint[] }
      const groupCost = (resp.trend || []).reduce((s, pt) => s + pt.actual_cost, 0)
      groupActualCosts[gid] = groupCost
      const conv = planConversionMap.value.get(gid)
      if (conv && conv.factor > 0) internalBillingRMB += groupCost * conv.factor
    }
    internalBillingRMB = Math.round(internalBillingRMB * 100) / 100
    const profitLoss = internalRevenue - upstreamCost
    const marginPct = upstreamCost > 0 ? (profitLoss / upstreamCost) * 100 : 0
    costResults[accountId] = { userInfo: ui, stat: st, upstreamCost, internalRevenue, internalBillingRMB, groupActualCosts, profitLoss, marginPct, trend: mergedTrend }
    saveCache()
  } catch (err: any) {
    fetchErrors[accountId] = err?.message || 'Failed'
  } finally {
    fetchingChannels[accountId] = false
  }
}

async function fetchLocalSummary() {
  const dr = getDateRange()
  localSummary.value = await upstreamCostAPI.getLocalSummary({
    ...dr,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  })
}

async function fetchSubscriptionRevenue() {
  for (const key of Object.keys(subscriptionRevenueByGroup)) delete subscriptionRevenueByGroup[Number(key)]
  const allGroupIds = [...new Set(accounts.value.flatMap(a => a.group_ids || []))]
  if (allGroupIds.length === 0) return
  try {
    const dr = getDateRange()
    const resp = await adminPaymentAPI.getSubscriptionRevenue({ group_ids: allGroupIds, ...dr })
    const result = (resp as any)?.data ?? resp ?? []
    for (const item of (Array.isArray(result) ? result : [])) subscriptionRevenueByGroup[item.group_id] = item.pay_amount
  } catch (err: any) {
    console.error('Failed to fetch subscription revenue:', err)
  }
}

function getAllocatedSubscriptionRevenue(accountId: number): number {
  const result = costResults[accountId]
  if (!result?.groupActualCosts) return 0
  let total = 0
  for (const [gidRaw, accountGroupCost] of Object.entries(result.groupActualCosts)) {
    const gid = Number(gidRaw)
    const groupRevenue = subscriptionRevenueByGroup[gid] || 0
    if (groupRevenue <= 0 || accountGroupCost <= 0) continue
    const groupTotalCost = Object.values(costResults).reduce((sum, item) => sum + (item.groupActualCosts?.[gid] || 0), 0)
    if (groupTotalCost > 0) total += groupRevenue * (accountGroupCost / groupTotalCost)
  }
  return Math.round(total * 100) / 100
}

async function fetchAllChannelCosts() {
  batchFetching.value = true
  const configured = accounts.value.filter(acc => !!getUpstreamCfg(acc))
  const results = await Promise.allSettled([
    fetchLocalSummary(),
    fetchSubscriptionRevenue(),
  ])
  batchFetching.value = false
  const summaryResult = results[0]
  if (summaryResult.status === 'rejected') {
    appStore.showError(summaryResult.reason?.message || 'Failed to load local upstream cost summary')
  }
  if (configured.length === 0) appStore.showError(t('admin.upstreamCost.noConfiguredChannels'))
}

// --- Table ---
const columns = computed(() => [
  { key: 'name', label: t('admin.upstreamCost.channelName') },
  { key: 'pool', label: t('admin.upstreamCost.upstreamPool') },
  { key: 'provider', label: t('admin.upstreamCost.providerType') },
  { key: 'upstream_cost_rmb', label: t('admin.upstreamCost.upstreamCostRMB') },
  { key: 'revenue_rmb', label: t('admin.upstreamCost.revenueRMB') },
  { key: 'profit_rmb', label: t('admin.upstreamCost.profitRMB') },
  { key: 'margin', label: t('admin.upstreamCost.margin') },
  { key: 'upstream_cost', label: t('admin.upstreamCost.upstreamCostLabel') },
  { key: 'revenue', label: t('admin.upstreamCost.internalRevenueLabel') },
  { key: 'actual_payment_rmb', label: t('admin.upstreamCost.actualPaymentRMB') },
  { key: 'balance', label: t('admin.upstreamCost.balance') },
  { key: 'actions', label: '' },
])

const localAccountMap = computed(() => {
  const map = new Map<number, { account: UpstreamCostAccountSummary; pool: UpstreamCostPoolSummary }>()
  for (const pool of localSummary.value?.pools || []) {
    for (const account of pool.accounts || []) {
      map.set(account.account_id, { account, pool })
    }
  }
  return map
})

function localAccountToResult(account: UpstreamCostAccountSummary): ChannelCostResult {
  const trend = (account.trend || []).map(pt => ({
    date: pt.date,
    requests: pt.requests,
    input_tokens: pt.input_tokens,
    output_tokens: pt.output_tokens,
    cache_creation_tokens: 0,
    cache_read_tokens: pt.cache_tokens,
    total_tokens: pt.total_tokens,
    cost: pt.upstream_cost,
    actual_cost: pt.user_cost,
  } as TrendDataPoint))
  return {
    userInfo: {
      id: account.account_id,
      username: account.account_name,
      display_name: account.account_name,
      email: '',
      quota: 0,
      used_quota: 0,
      request_count: account.requests,
      group: '',
      provider_type: account.provider_type,
      base_url: account.base_url,
    },
    stat: { quota: Math.round(account.upstream_cost * 500000), quota_usd: account.upstream_cost, rpm: 0, tpm: 0 },
    upstreamCost: account.upstream_cost,
    internalRevenue: account.user_cost,
    internalBillingRMB: account.user_cost * exchangeRate.value,
    groupActualCosts: {},
    profitLoss: account.profit,
    marginPct: account.upstream_cost > 0 ? (account.profit / account.upstream_cost) * 100 : 0,
    trend,
  }
}

const tableData = computed(() =>
  accounts.value.map(acc => {
    const local = localAccountMap.value.get(acc.id)
    return {
      id: acc.id,
      name: acc.name,
      status: acc.status,
      _account: acc,
      _cfg: getUpstreamCfg(acc),
      _pool: local?.pool || null,
      _localAccount: local?.account || null,
      _result: local?.account ? localAccountToResult(local.account) : (costResults[acc.id] || null),
      _subRevenue: getAllocatedSubscriptionRevenue(acc.id),
      _fetching: !!fetchingChannels[acc.id] || batchFetching.value,
      _error: fetchErrors[acc.id] || null,
    }
  })
)

// --- Detail Dialog ---
const showDetailDialog = ref(false)
const detailChannelId = ref<number | null>(null)
const detailChannel = computed(() => accounts.value.find(a => a.id === detailChannelId.value))
const detailLocalAccount = computed(() => detailChannelId.value !== null ? localAccountMap.value.get(detailChannelId.value)?.account || null : null)
const detailResult = computed(() => {
  if (detailLocalAccount.value) return localAccountToResult(detailLocalAccount.value)
  return detailChannelId.value !== null ? costResults[detailChannelId.value] : null
})
const detailLogs = reactive<{ items: UpstreamLogItem[]; total: number; page: number }>({ items: [], total: 0, page: 0 })
const detailLogsLoading = ref(false)

const detailTrend = computed(() => {
  if (!detailResult.value?.trend.length) return []
  const trend = detailResult.value.trend
  const maxVal = Math.max(...trend.map(pt => Math.max(pt.actual_cost, 0.001)))
  return trend.map(pt => ({ label: pt.date.slice(5), val: pt.actual_cost, pct: maxVal > 0 ? (pt.actual_cost / maxVal) * 100 : 0 }))
})

const modelBreakdown = computed(() => {
  if (detailLocalAccount.value?.models?.length) {
    const total = detailLocalAccount.value.models.reduce((s, item) => s + item.upstream_cost, 0)
    return detailLocalAccount.value.models.map((item: UpstreamCostModelBreakdown) => ({
      model: item.model,
      cost: item.upstream_cost,
      pct: total > 0 ? (item.upstream_cost / total) * 100 : 0,
    }))
  }
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
  if (localAccountMap.value.has(channelId)) return
  detailLogsLoading.value = true
  try {
    const acc = accounts.value.find(a => a.id === channelId)
    const cfg = acc ? getUpstreamCfg(acc) : null
    if (!cfg || !hasRemoteCredentials(cfg)) return
    const lg = await upstreamCostAPI.getLogs(cfgToProvider(cfg), { page: 0, page_size: 50 })
    detailLogs.items = lg.items || []; detailLogs.total = lg.total
  } catch (err: any) { appStore.showError(err?.message || 'Failed to load logs') }
  finally { detailLogsLoading.value = false }
}

async function loadMoreDetailLogs() {
  if (detailChannelId.value === null) return
  const acc = accounts.value.find(a => a.id === detailChannelId.value)
  const cfg = acc ? getUpstreamCfg(acc) : null
  if (!cfg || !hasRemoteCredentials(cfg)) return
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

onMounted(async () => {
  loadCache()
  await Promise.all([loadAccounts(), loadPlans()])
  const configured = accounts.value.filter(acc => !!getUpstreamCfg(acc))
  if (configured.length > 0) {
    fetchAllChannelCosts()
  }
})
</script>
