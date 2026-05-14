<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <span
              v-if="realSummary"
              class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold"
              :class="realSummary.totals.profit_rmb >= 0 ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'"
            >
              盈亏: {{ signedMoney(realSummary.totals.profit_rmb) }}
            </span>
            <span v-if="realSummary" class="inline-flex items-center gap-1.5 rounded-full bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
              上游累计消耗 {{ money(realSummary.totals.upstream_used_rmb) }} / 剩余 {{ money(realSummary.totals.upstream_remaining_rmb) }}
            </span>
            <span v-if="realSummary" class="inline-flex items-center gap-1.5 rounded-full bg-blue-50 px-3 py-1 text-xs font-semibold text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
              用户累计扣费 {{ money(realSummary.totals.downstream_revenue_rmb) }}
            </span>
            <span
              v-if="realSummary && realSummary.totals.unallocated_upstream_used_rmb > 0.000001"
              class="inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-3 py-1 text-xs font-semibold text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
            >
              未分摊 {{ money(realSummary.totals.unallocated_upstream_used_rmb) }}
            </span>
          </div>
          <button @click="fetchAllChannelCosts(true)" :disabled="batchFetching" class="btn btn-primary">
            <Icon name="refresh" size="md" :class="batchFetching ? 'animate-spin' : ''" />
            <span class="ml-1.5">刷新真实上游</span>
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="tableData" :loading="loading">
          <template #cell-name="{ row }">
            <div :class="row.rowType === 'account' ? 'pl-6' : row.rowType === 'group' ? 'pl-12' : ''">
              <div class="flex items-center gap-2">
                <span
                  v-if="row.rowType === 'pool'"
                  class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300"
                >
                  上游
                </span>
                <span
                  v-else-if="row.rowType === 'account'"
                  class="rounded bg-blue-50 px-1.5 py-0.5 text-[10px] font-semibold text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
                >
                  账号
                </span>
                <span
                  v-else
                  class="rounded bg-emerald-50 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
                >
                  分组
                </span>
                <span class="max-w-72 truncate font-medium text-gray-900 dark:text-white" :title="row.name">{{ row.name }}</span>
                <span
                  v-if="row.rowType === 'pool'"
                  class="rounded bg-gray-50 px-1.5 py-0.5 text-[10px] text-gray-500 dark:bg-dark-800 dark:text-gray-400"
                >
                  {{ row.accountCount }} 个账号
                </span>
              </div>
              <div class="mt-0.5 flex items-center gap-2 text-[10px] text-gray-400">
                <span class="max-w-80 truncate" :title="row.baseUrl">{{ row.baseUrl }}</span>
                <span v-if="row.rowType === 'account' && row.tokenHash">key {{ row.tokenHash }}</span>
                <span v-if="row.rowType === 'account' && row.tokenName" class="max-w-32 truncate" :title="row.tokenName">{{ row.tokenName }}</span>
                <span v-if="row.rowType === 'group'">ID {{ row.groupId || '—' }}</span>
              </div>
            </div>
          </template>

          <template #cell-provider="{ row }">
            <span class="inline-flex rounded bg-slate-100 px-1.5 py-0.5 text-[10px] font-medium uppercase text-slate-600 dark:bg-dark-700 dark:text-slate-300">
              {{ providerLabel(row.providerType) }}
            </span>
          </template>

          <template #cell-upstream_used_rmb="{ row }">
            <span class="font-semibold text-red-600 dark:text-red-400">{{ money(row.upstreamUsedRMB) }}</span>
            <span v-if="row.rowType === 'pool' && row.unallocatedUpstreamUsedRMB > 0.000001" class="ml-1 text-[10px] text-amber-600 dark:text-amber-300">
              未分摊 {{ money(row.unallocatedUpstreamUsedRMB) }}
            </span>
          </template>

          <template #cell-upstream_remaining_rmb="{ row }">
            <span v-if="row.rowType === 'group'" class="text-gray-300">—</span>
            <span v-else class="font-medium text-emerald-600 dark:text-emerald-400">{{ money(row.upstreamRemainingRMB) }}</span>
          </template>

          <template #cell-upstream_used_usd="{ row }">
            <span class="font-medium text-amber-600 dark:text-amber-400">${{ row.upstreamUsedUSD.toFixed(4) }}</span>
          </template>

          <template #cell-downstream_revenue_rmb="{ row }">
            <span class="font-medium text-blue-600 dark:text-blue-400">{{ money(row.downstreamRevenueRMB) }}</span>
          </template>

          <template #cell-profit_rmb="{ row }">
            <span class="font-semibold" :class="row.profitRMB >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
              {{ signedMoney(row.profitRMB) }}
            </span>
          </template>

          <template #cell-rates="{ row }">
            <div class="space-y-0.5 text-[10px]">
              <div v-if="row.rowType === 'group'">
                <span class="text-gray-400">当前分组</span>
                <span class="ml-1 font-semibold text-gray-700 dark:text-gray-200">{{ rate(row.currentGroupRate) }}</span>
              </div>
              <div v-else>
                <span class="text-gray-400">上游配置</span>
                <span class="ml-1 font-semibold text-gray-700 dark:text-gray-200">{{ rate(row.upstreamConfiguredRate) }}</span>
              </div>
              <div>
                <span class="text-gray-400">下游实扣</span>
                <span class="ml-1 font-semibold text-blue-600 dark:text-blue-400">{{ rate(row.downstreamEffectiveRate) }}</span>
              </div>
            </div>
          </template>

          <template #cell-requests="{ row }">
            <div class="text-xs text-gray-600 dark:text-gray-300">
              <div>{{ row.requests.toLocaleString() }} req</div>
              <div class="text-[10px] text-gray-400">{{ row.totalTokens.toLocaleString() }} tokens</div>
            </div>
          </template>

          <template #cell-status="{ row }">
            <div class="max-w-44">
              <span
                class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold"
                :class="statusClass(row.remoteStatus)"
                :title="row.error"
              >
                {{ statusLabel(row.remoteStatus) }}
              </span>
              <div v-if="row.rowType !== 'pool'" class="mt-1 text-[10px] text-gray-400">{{ sourceLabel(row.source) }}</div>
              <div v-if="row.error" class="mt-1 truncate text-[10px] text-red-500" :title="row.error">{{ row.error }}</div>
            </div>
          </template>

          <template #cell-actions="{ row }">
            <div v-if="row.rowType === 'account'" class="flex items-center gap-1">
              <button
                v-if="row.account"
                @click="openConfigDialog(row.account)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
              >
                <Icon name="cog" size="sm" />
                <span class="text-[10px]">配置</span>
              </button>
              <button
                @click="openDetailDialog(row.child)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
              >
                <Icon name="chart" size="sm" />
                <span class="text-[10px]">详情</span>
              </button>
            </div>
            <span v-else-if="row.rowType === 'group'" class="text-gray-300">—</span>
            <span v-else class="text-gray-300">—</span>
          </template>

          <template #empty>
            <EmptyState title="还没有配置上游成本账号" />
          </template>
        </DataTable>
      </template>
    </TablePageLayout>

    <BaseDialog :show="showConfigDialog" :title="'上游成本配置 — ' + (configAccount?.name || '')" @close="showConfigDialog = false">
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">类型</label>
            <select v-model="configForm.type" class="input w-full text-sm">
              <option value="newapi">New-API</option>
              <option value="sub2api">Sub2API</option>
            </select>
          </div>
          <div v-if="configForm.type === 'newapi'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">User ID</label>
            <input v-model.number="configForm.user_id" type="number" class="input w-full text-sm" />
          </div>
        </div>
        <div>
          <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">Base URL</label>
          <input v-model="configForm.base_url" type="text" class="input w-full text-sm" placeholder="https://example.com" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">Pool Key</label>
            <input v-model="configForm.pool_key" type="text" class="input w-full text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">Pool Name</label>
            <input v-model="configForm.pool_name" type="text" class="input w-full text-sm" />
          </div>
        </div>
        <div v-if="configForm.type === 'newapi'">
          <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">Access Token</label>
          <input v-model="configForm.access_token" type="password" class="input w-full text-sm" />
        </div>
        <template v-if="configForm.type === 'sub2api'">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">Email</label>
            <input v-model="configForm.email" type="email" class="input w-full text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">Password</label>
            <input v-model="configForm.password" type="password" class="input w-full text-sm" />
          </div>
        </template>
        <div class="flex justify-end gap-2 pt-2">
          <button @click="showConfigDialog = false" class="btn btn-secondary">取消</button>
          <button @click="saveConfig" :disabled="savingConfig" class="btn btn-primary">
            <Icon v-if="savingConfig" name="refresh" size="sm" class="mr-1 animate-spin" />
            保存
          </button>
        </div>
      </div>
    </BaseDialog>

    <BaseDialog :show="showDetailDialog" :title="(detailChild?.account_name || '') + ' — 成本详情'" width="extra-wide" @close="showDetailDialog = false">
      <div v-if="detailChild" class="space-y-5">
        <div class="grid grid-cols-2 gap-3 md:grid-cols-5">
          <div class="rounded-lg bg-red-50 p-3 dark:bg-red-900/20">
            <p class="text-[10px] text-red-500">上游真实消耗</p>
            <p class="text-base font-bold text-red-600 dark:text-red-400">{{ money(detailChild.upstream_used_rmb, 4) }}</p>
          </div>
          <div class="rounded-lg bg-emerald-50 p-3 dark:bg-emerald-900/20">
            <p class="text-[10px] text-emerald-500">上游剩余</p>
            <p class="text-base font-bold text-emerald-600 dark:text-emerald-400">{{ money(detailChild.upstream_remaining_rmb, 4) }}</p>
          </div>
          <div class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/20">
            <p class="text-[10px] text-blue-500">用户实际扣费</p>
            <p class="text-base font-bold text-blue-600 dark:text-blue-400">{{ money(detailChild.downstream_revenue_rmb, 4) }}</p>
          </div>
          <div class="rounded-lg p-3" :class="detailChild.profit_rmb >= 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'">
            <p class="text-[10px]" :class="detailChild.profit_rmb >= 0 ? 'text-emerald-500' : 'text-red-500'">盈亏</p>
            <p class="text-base font-bold" :class="detailChild.profit_rmb >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
              {{ signedMoney(detailChild.profit_rmb, 4) }}
            </p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
            <p class="text-[10px] text-gray-400">真实来源</p>
            <p class="text-base font-bold text-gray-900 dark:text-white">{{ sourceLabel(detailChild.source) }}</p>
          </div>
        </div>

        <div v-if="detailChild.error" class="rounded-lg border border-red-200 bg-red-50 p-3 text-xs text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">
          {{ detailChild.error }}
        </div>

        <div v-if="detailGroupBreakdown.length > 0" class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
          <table class="w-full text-left text-xs">
            <thead class="bg-gray-50 text-[10px] font-semibold text-gray-500 dark:bg-dark-700 dark:text-gray-400">
              <tr>
                <th class="px-3 py-2">分组</th>
                <th class="px-3 py-2">当前倍率</th>
                <th class="px-3 py-2">历史实扣</th>
                <th class="px-3 py-2">上游分摊</th>
                <th class="px-3 py-2">用户扣费</th>
                <th class="px-3 py-2">盈亏</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="group in detailGroupBreakdown" :key="group.group_id" class="text-gray-700 dark:text-gray-300">
                <td class="px-3 py-2">
                  <div class="font-medium text-gray-900 dark:text-white">{{ group.group_name }}</div>
                  <div class="text-[10px] text-gray-400">{{ group.requests.toLocaleString() }} req / {{ group.total_tokens.toLocaleString() }} tokens</div>
                </td>
                <td class="px-3 py-2">{{ rate(group.current_group_rate) }}</td>
                <td class="px-3 py-2 text-blue-600 dark:text-blue-400">{{ rate(group.downstream_effective_rate) }}</td>
                <td class="px-3 py-2 text-red-600 dark:text-red-400">{{ money(group.allocated_upstream_used_rmb, 4) }}</td>
                <td class="px-3 py-2 text-blue-600 dark:text-blue-400">{{ money(group.downstream_revenue_rmb, 4) }}</td>
                <td class="px-3 py-2 font-semibold" :class="group.profit_rmb >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                  {{ signedMoney(group.profit_rmb, 4) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="detailTrend.length > 0" class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
          <h3 class="mb-3 text-xs font-semibold text-gray-700 dark:text-gray-300">本地扣费趋势</h3>
          <div class="flex items-end gap-0.5" style="height: 100px">
            <div v-for="pt in detailTrend" :key="pt.label" class="relative flex flex-1 flex-col items-center justify-end" style="min-width: 0">
              <div class="w-full rounded-t bg-blue-400 transition-all dark:bg-blue-500" :style="{ height: pt.pct + '%', minHeight: pt.val > 0 ? '2px' : '0' }" :title="pt.label + ': ' + money(pt.val, 4)" />
              <span class="mt-1 text-[7px] text-gray-400">{{ pt.label }}</span>
            </div>
          </div>
        </div>

        <div v-if="modelBreakdown.length > 0" class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
          <h3 class="mb-3 text-xs font-semibold text-gray-700 dark:text-gray-300">本地模型扣费</h3>
          <div class="space-y-1.5">
            <div v-for="item in modelBreakdown" :key="item.model" class="flex items-center gap-2">
              <span class="w-40 truncate text-[11px] text-gray-600 dark:text-gray-400" :title="item.model">{{ item.model }}</span>
              <div class="h-4 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-600">
                <div class="h-full rounded-full bg-blue-500 transition-all" :style="{ width: item.pct + '%', minWidth: item.pct > 0 ? '1rem' : '0' }" />
              </div>
              <span class="w-20 text-right text-[11px] font-medium text-blue-600 dark:text-blue-400">{{ money(item.cost, 4) }}</span>
            </div>
          </div>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import { upstreamCostAPI } from '@/api/upstream-cost'
import type {
  UpstreamCostModelBreakdown,
  UpstreamRealAccountSummary,
  UpstreamRealGroupSummary,
  UpstreamRealPoolSummary,
  UpstreamRealSummary
} from '@/api/upstream-cost'
import type { Account } from '@/types'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const realSummaryCacheKey = 'sub2api:admin:upstream-real-summary:v1'
const accounts = ref<Account[]>([])
const realSummary = ref<UpstreamRealSummary | null>(null)
const loadingAccounts = ref(false)
const batchFetching = ref(false)

const loading = computed(() => loadingAccounts.value || batchFetching.value)

async function loadAccounts() {
  loadingAccounts.value = true
  try {
    const resp = await adminAPI.accounts.list(1, 1000, { sort_by: 'id', sort_order: 'asc' })
    accounts.value = resp.items || []
  } catch (err: any) {
    appStore.showError(err?.message || '账号列表加载失败')
  } finally {
    loadingAccounts.value = false
  }
}

function loadCachedRealSummary() {
  try {
    const raw = window.localStorage.getItem(realSummaryCacheKey)
    if (!raw) return
    const cached = JSON.parse(raw) as UpstreamRealSummary
    if (cached && Array.isArray(cached.pools)) {
      realSummary.value = cached
    }
  } catch {
    window.localStorage.removeItem(realSummaryCacheKey)
  }
}

function saveCachedRealSummary(summary: UpstreamRealSummary) {
  if (summary.scope === 'remote_upstream_cache_empty') return
  try {
    window.localStorage.setItem(realSummaryCacheKey, JSON.stringify(summary))
  } catch {
    // Ignore storage quota/private-mode failures; backend cache still works.
  }
}

async function fetchAllChannelCosts(refresh = false) {
  batchFetching.value = true
  try {
    const summary = await upstreamCostAPI.getRealSummary({
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      refresh,
    })
    if (summary.scope === 'remote_upstream_cache_empty' && realSummary.value) return
    realSummary.value = summary
    saveCachedRealSummary(summary)
  } catch (err: any) {
    appStore.showError(err?.message || '真实上游成本加载失败')
  } finally {
    batchFetching.value = false
  }
}

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
  const cfg = (acc.extra as Record<string, unknown> | undefined)?.upstream_provider as Partial<UpstreamCfg> | undefined
  if (!cfg || !cfg.base_url) return null
  return {
    type: cfg.type === 'sub2api' ? 'sub2api' : 'newapi',
    base_url: cfg.base_url || '',
    access_token: cfg.access_token || '',
    user_id: Number(cfg.user_id || 0),
    email: cfg.email || '',
    password: cfg.password || '',
    pool_key: cfg.pool_key || '',
    pool_name: cfg.pool_name || '',
  }
}

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
      extra: { ...existingExtra, upstream_provider: { ...configForm } },
    })
    appStore.showSuccess('配置已保存')
    showConfigDialog.value = false
    await Promise.all([loadAccounts(), fetchAllChannelCosts(true)])
  } catch (err: any) {
    appStore.showError(err?.message || '保存失败')
  } finally {
    savingConfig.value = false
  }
}

interface TableRow {
  id: string
  rowType: 'pool' | 'account' | 'group'
  name: string
  providerType: string
  baseUrl: string
  accountCount: number
  groupId: number
  currentGroupRate: number
  tokenName: string
  tokenHash: string
  source: string
  remoteStatus: string
  error: string
  requests: number
  totalTokens: number
  upstreamUsedRMB: number
  upstreamRemainingRMB: number
  upstreamUsedUSD: number
  downstreamRevenueRMB: number
  profitRMB: number
  upstreamConfiguredRate: number
  upstreamEffectiveRate: number
  downstreamEffectiveRate: number
  unallocatedUpstreamUsedRMB: number
  account: Account | null
  child: UpstreamRealAccountSummary | null
  groupBreakdown: UpstreamRealGroupSummary | null
  pool: UpstreamRealPoolSummary | null
}

const accountByID = computed(() => {
  const map = new Map<number, Account>()
  for (const account of accounts.value) map.set(account.id, account)
  return map
})

const tableData = computed<TableRow[]>(() => {
  const rows: TableRow[] = []
  for (const pool of realSummary.value?.pools || []) {
    rows.push({
      id: `pool-${pool.pool_key}`,
      rowType: 'pool',
      name: pool.pool_name,
      providerType: pool.provider_type,
      baseUrl: pool.base_url,
      accountCount: pool.account_count,
      groupId: 0,
      currentGroupRate: 0,
      tokenName: '',
      tokenHash: '',
      source: '',
      remoteStatus: pool.status,
      error: (pool.errors || []).join('; '),
      requests: pool.requests,
      totalTokens: pool.total_tokens,
      upstreamUsedRMB: pool.upstream_used_rmb,
      upstreamRemainingRMB: pool.upstream_remaining_rmb,
      upstreamUsedUSD: pool.upstream_used_usd,
      downstreamRevenueRMB: pool.downstream_revenue_rmb,
      profitRMB: pool.profit_rmb,
      upstreamConfiguredRate: pool.weighted_upstream_account_rate,
      upstreamEffectiveRate: pool.upstream_effective_rate,
      downstreamEffectiveRate: pool.downstream_effective_rate,
      unallocatedUpstreamUsedRMB: pool.unallocated_upstream_used_rmb,
      account: null,
      child: null,
      groupBreakdown: null,
      pool,
    })
    for (const child of pool.accounts || []) {
      rows.push({
        id: `account-${child.account_id}`,
        rowType: 'account',
        name: child.account_name,
        providerType: child.provider_type,
        baseUrl: child.base_url,
        accountCount: 1,
        groupId: 0,
        currentGroupRate: 0,
        tokenName: child.token_name,
        tokenHash: child.token_hash,
        source: child.source,
        remoteStatus: child.remote_status,
        error: child.error,
        requests: child.requests,
        totalTokens: child.total_tokens,
        upstreamUsedRMB: child.upstream_used_rmb,
        upstreamRemainingRMB: child.upstream_remaining_rmb,
        upstreamUsedUSD: child.upstream_used_usd,
        downstreamRevenueRMB: child.downstream_revenue_rmb,
        profitRMB: child.profit_rmb,
        upstreamConfiguredRate: child.upstream_configured_rate,
        upstreamEffectiveRate: child.upstream_effective_rate,
        downstreamEffectiveRate: child.downstream_effective_rate,
        unallocatedUpstreamUsedRMB: 0,
        account: accountByID.value.get(child.account_id) || null,
        child,
        groupBreakdown: null,
        pool,
      })
      for (const group of child.groups || []) {
        rows.push({
          id: `account-${child.account_id}-group-${group.group_id}`,
          rowType: 'group',
          name: group.group_name,
          providerType: child.provider_type,
          baseUrl: child.account_name,
          accountCount: 1,
          groupId: group.group_id,
          currentGroupRate: group.current_group_rate,
          tokenName: '',
          tokenHash: '',
          source: 'group_allocation',
          remoteStatus: 'allocated',
          error: '',
          requests: group.requests,
          totalTokens: group.total_tokens,
          upstreamUsedRMB: group.allocated_upstream_used_rmb,
          upstreamRemainingRMB: 0,
          upstreamUsedUSD: group.allocated_upstream_used_rmb,
          downstreamRevenueRMB: group.downstream_revenue_rmb,
          profitRMB: group.profit_rmb,
          upstreamConfiguredRate: child.upstream_configured_rate,
          upstreamEffectiveRate: child.upstream_effective_rate,
          downstreamEffectiveRate: group.downstream_effective_rate,
          unallocatedUpstreamUsedRMB: 0,
          account: accountByID.value.get(child.account_id) || null,
          child,
          groupBreakdown: group,
          pool,
        })
      }
    }
  }
  return rows
})

const columns = computed(() => [
  { key: 'name', label: '上游 / 账号' },
  { key: 'provider', label: '类型' },
  { key: 'upstream_used_rmb', label: '上游消耗 ¥' },
  { key: 'upstream_remaining_rmb', label: '上游剩余 ¥' },
  { key: 'upstream_used_usd', label: '消耗 $' },
  { key: 'downstream_revenue_rmb', label: '用户扣费 ¥' },
  { key: 'profit_rmb', label: '盈亏 ¥' },
  { key: 'rates', label: '倍率对比' },
  { key: 'requests', label: '请求' },
  { key: 'status', label: '状态' },
  { key: 'actions', label: '' },
])

const showDetailDialog = ref(false)
const detailChild = ref<UpstreamRealAccountSummary | null>(null)

function openDetailDialog(child: UpstreamRealAccountSummary | null) {
  if (!child) return
  detailChild.value = child
  showDetailDialog.value = true
}

const detailTrend = computed(() => {
  const trend = detailChild.value?.trend || []
  if (trend.length === 0) return []
  const maxVal = Math.max(...trend.map(pt => Math.max(pt.user_cost, 0.001)))
  return trend.map(pt => ({
    label: pt.date.slice(5),
    val: pt.user_cost,
    pct: maxVal > 0 ? (pt.user_cost / maxVal) * 100 : 0,
  }))
})

const detailGroupBreakdown = computed(() => detailChild.value?.groups || [])

const modelBreakdown = computed(() => {
  const models = detailChild.value?.models || []
  if (models.length === 0) return []
  const total = models.reduce((sum, item) => sum + item.user_cost, 0)
  return models.map((item: UpstreamCostModelBreakdown) => ({
    model: item.model,
    cost: item.user_cost,
    pct: total > 0 ? (item.user_cost / total) * 100 : 0,
  }))
})

function money(value: number, digits = 2): string {
  return `¥${Number(value || 0).toFixed(digits)}`
}

function signedMoney(value: number, digits = 2): string {
  const n = Number(value || 0)
  return `${n >= 0 ? '+' : ''}${money(n, digits)}`
}

function rate(value: number): string {
  const n = Number(value || 0)
  return `${n.toFixed(3)}x`
}

function providerLabel(provider: string): string {
  if (provider === 'newapi') return 'New-API'
  if (provider === 'sub2api') return 'Sub2API'
  return provider || '—'
}

function statusLabel(status: string): string {
  if (status === 'ok') return '真实'
  if (status === 'allocated') return '分摊'
  if (status === 'partial') return '部分'
  if (status === 'error') return '错误'
  return status || '—'
}

function statusClass(status: string): string {
  if (status === 'ok') return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  if (status === 'allocated') return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
  if (status === 'partial') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
}

function sourceLabel(source: string): string {
  if (source === 'token') return 'API key 真实'
  if (source === 'account_total') return '账号总量'
  if (source === 'unallocated') return '未拆分'
  if (source === 'token_error') return 'Key 拉取失败'
  if (source === 'group_allocation') return '按账号标准成本分摊'
  return source || '—'
}

onMounted(async () => {
  loadCachedRealSummary()
  await Promise.all([loadAccounts(), fetchAllChannelCosts(false)])
})
</script>
