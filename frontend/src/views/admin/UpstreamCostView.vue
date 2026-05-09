<template>
  <AppLayout>
    <div class="space-y-6 p-4 md:p-6">
      <!-- Page Header -->
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.upstreamCost.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.description') }}</p>
        </div>
        <div class="flex items-center gap-2">
          <select v-model="dateRange" @change="onDateRangeChange" class="input text-sm">
            <option value="7d">{{ t('admin.upstreamCost.last7d') }}</option>
            <option value="30d">{{ t('admin.upstreamCost.last30d') }}</option>
            <option value="thisMonth">{{ t('admin.upstreamCost.thisMonth') }}</option>
            <option value="lastMonth">{{ t('admin.upstreamCost.lastMonth') }}</option>
            <option value="90d">{{ t('admin.upstreamCost.last90d') }}</option>
          </select>
          <button @click="fetchAllChannelCosts" :disabled="batchFetching" class="btn btn-primary text-sm">
            <svg v-if="batchFetching" class="mr-1.5 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            {{ batchFetching ? t('admin.upstreamCost.fetching') : t('admin.upstreamCost.fetchAll') }}
          </button>
        </div>
      </div>

      <!-- Global Summary (shows when we have any results) -->
      <div v-if="globalSummary.hasData" class="grid grid-cols-2 gap-4 md:grid-cols-4">
        <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.upstreamTotal') }}</p>
          <p class="mt-1 text-xl font-bold text-red-600 dark:text-red-400">${{ globalSummary.upstreamCost.toFixed(2) }}</p>
          <p class="text-xs text-gray-400">{{ t('admin.upstreamCost.paidToUpstream') }}</p>
        </div>
        <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.internalRevenue') }}</p>
          <p class="mt-1 text-xl font-bold text-blue-600 dark:text-blue-400">${{ globalSummary.internalRevenue.toFixed(2) }}</p>
          <p class="text-xs text-gray-400">{{ t('admin.upstreamCost.chargedToUsers') }}</p>
        </div>
        <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.profitLoss') }}</p>
          <p class="mt-1 text-xl font-bold" :class="globalSummary.profitLoss >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
            {{ globalSummary.profitLoss >= 0 ? '+' : '' }}${{ globalSummary.profitLoss.toFixed(2) }}
          </p>
          <p class="text-xs" :class="globalSummary.profitLoss >= 0 ? 'text-emerald-500' : 'text-red-500'">
            {{ globalSummary.profitLoss >= 0 ? t('admin.upstreamCost.profitable') : t('admin.upstreamCost.losing') }}
          </p>
        </div>
        <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.margin') }}</p>
          <p class="mt-1 text-xl font-bold" :class="globalSummary.marginPct >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
            {{ globalSummary.marginPct >= 0 ? '+' : '' }}{{ globalSummary.marginPct.toFixed(1) }}%
          </p>
          <p class="text-xs text-gray-400">{{ t('admin.upstreamCost.marginDesc') }}</p>
        </div>
      </div>

      <!-- Channel Cards Grid -->
      <div v-if="loadingChannels" class="flex items-center justify-center py-16">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
      </div>

      <div v-else-if="channels.length === 0" class="flex flex-col items-center justify-center rounded-xl border border-dashed border-gray-300 py-16 dark:border-dark-600">
        <svg class="mb-3 h-12 w-12 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v6m3-3H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
        <p class="text-sm text-gray-400">{{ t('admin.upstreamCost.noChannels') }}</p>
      </div>

      <div v-else class="grid grid-cols-1 gap-4 lg:grid-cols-2 xl:grid-cols-3">
        <div
          v-for="ch in channels"
          :key="ch.id"
          class="group relative rounded-xl border border-gray-200 bg-white shadow-sm transition-shadow hover:shadow-md dark:border-dark-600 dark:bg-dark-800"
        >
          <!-- Channel Header -->
          <div class="flex items-center justify-between border-b border-gray-100 px-4 py-3 dark:border-dark-700">
            <div class="flex items-center gap-2">
              <span class="font-semibold text-gray-900 dark:text-white">{{ ch.name }}</span>
              <span :class="ch.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-dark-600 dark:text-gray-400'" class="rounded px-1.5 py-0.5 text-xs font-medium">{{ ch.status }}</span>
            </div>
            <button v-if="hasData(ch.id)" @click="openDetail(ch.id)" class="rounded-lg p-1 text-gray-400 opacity-0 transition-all hover:bg-gray-100 hover:text-primary-600 group-hover:opacity-100 dark:hover:bg-dark-700" :title="t('admin.upstreamCost.viewDetail')">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25"/></svg>
            </button>
          </div>

          <!-- Config Form -->
          <div class="space-y-3 px-4 py-3">
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="mb-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.providerType') }}</label>
                <select v-model="getConfig(ch.id).provider_type" @change="saveConfigs" class="input w-full text-xs">
                  <option value="newapi">New-API</option>
                  <option value="sub2api">Sub2API</option>
                </select>
              </div>
              <div>
                <label class="mb-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.userId') }}</label>
                <input v-model.number="getConfig(ch.id).user_id" @change="saveConfigs" type="number" class="input w-full text-xs" :placeholder="t('admin.upstreamCost.userIdPlaceholder')" />
              </div>
            </div>
            <div>
              <label class="mb-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.baseUrl') }}</label>
              <input v-model="getConfig(ch.id).base_url" @change="saveConfigs" type="text" class="input w-full text-xs" :placeholder="t('admin.upstreamCost.baseUrlPlaceholder')" />
            </div>
            <div>
              <label class="mb-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ t('admin.upstreamCost.accessToken') }}</label>
              <input v-model="getConfig(ch.id).access_token" @change="saveConfigs" type="password" class="input w-full text-xs" :placeholder="t('admin.upstreamCost.accessTokenPlaceholder')" />
            </div>

            <!-- Cost Comparison Result -->
            <div v-if="costResults[ch.id]" class="mt-2 space-y-2">
              <div class="grid grid-cols-3 gap-2 rounded-lg bg-gray-50 p-2.5 dark:bg-dark-700">
                <div class="text-center">
                  <p class="text-[10px] text-gray-400">{{ t('admin.upstreamCost.upstreamCostLabel') }}</p>
                  <p class="text-sm font-bold text-red-600 dark:text-red-400">${{ costResults[ch.id].upstreamCost.toFixed(2) }}</p>
                </div>
                <div class="text-center">
                  <p class="text-[10px] text-gray-400">{{ t('admin.upstreamCost.internalRevenueLabel') }}</p>
                  <p class="text-sm font-bold text-blue-600 dark:text-blue-400">${{ costResults[ch.id].internalRevenue.toFixed(2) }}</p>
                </div>
                <div class="text-center">
                  <p class="text-[10px] text-gray-400">{{ t('admin.upstreamCost.profitLoss') }}</p>
                  <p class="text-sm font-bold" :class="costResults[ch.id].profitLoss >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                    {{ costResults[ch.id].profitLoss >= 0 ? '+' : '' }}${{ costResults[ch.id].profitLoss.toFixed(2) }}
                  </p>
                </div>
              </div>
              <div class="flex items-center justify-center">
                <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="costResults[ch.id].profitLoss >= 0 ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'">
                  {{ costResults[ch.id].profitLoss >= 0 ? t('admin.upstreamCost.profitable') : t('admin.upstreamCost.losing') }}
                  {{ costResults[ch.id].marginPct >= 0 ? '+' : '' }}{{ costResults[ch.id].marginPct.toFixed(1) }}%
                </span>
              </div>
            </div>

            <!-- Fetching / Error / Not configured states -->
            <div v-else-if="fetchingChannels[ch.id]" class="flex items-center justify-center py-3">
              <svg class="h-5 w-5 animate-spin text-primary-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
              <span class="ml-2 text-xs text-gray-400">{{ t('admin.upstreamCost.fetching') }}</span>
            </div>
            <div v-else-if="fetchErrors[ch.id]" class="rounded-lg bg-red-50 p-2 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-400">{{ fetchErrors[ch.id] }}</div>
            <div v-else-if="!isConfigComplete(ch.id)" class="rounded-lg bg-amber-50 p-2 text-xs text-amber-600 dark:bg-amber-900/20 dark:text-amber-400">{{ t('admin.upstreamCost.configIncomplete') }}</div>
          </div>

          <!-- Actions -->
          <div class="flex items-center justify-between border-t border-gray-100 px-4 py-2 dark:border-dark-700">
            <button @click="fetchSingleChannelCost(ch.id)" :disabled="fetchingChannels[ch.id] || !isConfigComplete(ch.id)" class="text-xs font-medium text-primary-600 hover:text-primary-700 disabled:text-gray-300 dark:text-primary-400 dark:hover:text-primary-300 dark:disabled:text-gray-600">{{ t('admin.upstreamCost.fetchData') }}</button>
            <button v-if="costResults[ch.id]" @click="openDetail(ch.id)" class="text-xs font-medium text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">{{ t('admin.upstreamCost.viewDetail') }} →</button>
          </div>
        </div>
      </div>

      <!-- ====== Detail Slide-Over ====== -->
      <Teleport to="body">
        <Transition name="slide-over">
          <div v-if="detailChannelId !== null && costResults[detailChannelId]" class="fixed inset-0 z-50 flex justify-end">
            <div class="absolute inset-0 bg-black/30" @click="detailChannelId = null" />
            <div class="relative z-10 flex h-full w-full max-w-3xl flex-col overflow-y-auto bg-white shadow-2xl dark:bg-dark-800">
              <!-- Detail Header -->
              <div class="sticky top-0 z-10 flex items-center justify-between border-b border-gray-200 bg-white px-6 py-4 dark:border-dark-600 dark:bg-dark-800">
                <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ detailChannel?.name }} — {{ t('admin.upstreamCost.costDetail') }}</h2>
                <button @click="detailChannelId = null" class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
                </button>
              </div>

              <div class="flex-1 space-y-6 p-6" v-if="detailChannelId !== null && costResults[detailChannelId]">
                <!-- P&L Summary -->
                <div class="grid grid-cols-2 gap-3 md:grid-cols-4">
                  <div class="rounded-lg bg-red-50 p-3 dark:bg-red-900/20">
                    <p class="text-xs text-red-500">{{ t('admin.upstreamCost.upstreamCostLabel') }}</p>
                    <p class="text-lg font-bold text-red-600 dark:text-red-400">${{ costResults[detailChannelId].upstreamCost.toFixed(4) }}</p>
                  </div>
                  <div class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/20">
                    <p class="text-xs text-blue-500">{{ t('admin.upstreamCost.internalRevenueLabel') }}</p>
                    <p class="text-lg font-bold text-blue-600 dark:text-blue-400">${{ costResults[detailChannelId].internalRevenue.toFixed(4) }}</p>
                  </div>
                  <div class="rounded-lg p-3" :class="costResults[detailChannelId].profitLoss >= 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'">
                    <p class="text-xs" :class="costResults[detailChannelId].profitLoss >= 0 ? 'text-emerald-500' : 'text-red-500'">{{ t('admin.upstreamCost.profitLoss') }}</p>
                    <p class="text-lg font-bold" :class="costResults[detailChannelId].profitLoss >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                      {{ costResults[detailChannelId].profitLoss >= 0 ? '+' : '' }}${{ costResults[detailChannelId].profitLoss.toFixed(4) }}
                    </p>
                  </div>
                  <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
                    <p class="text-xs text-gray-400">{{ t('admin.upstreamCost.availableQuota') }}</p>
                    <p class="text-lg font-bold text-emerald-600 dark:text-emerald-400">${{ (costResults[detailChannelId].userInfo.quota / 500000).toFixed(4) }}</p>
                  </div>
                </div>

                <!-- Daily Trend Chart (CSS bars) -->
                <div v-if="detailTrend.length > 0" class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
                  <h3 class="mb-3 text-sm font-semibold text-gray-800 dark:text-gray-200">{{ t('admin.upstreamCost.dailyTrend') }}</h3>
                  <div class="flex items-end gap-1" style="height: 120px">
                    <div v-for="(pt, idx) in detailTrend" :key="idx" class="group/bar relative flex flex-1 flex-col items-center justify-end" style="min-width: 0">
                      <div class="relative flex w-full items-end justify-center gap-px" :style="{ height: '100%' }">
                        <div class="w-1/2 rounded-t bg-blue-400 dark:bg-blue-500 transition-all" :style="{ height: pt.revPct + '%', minHeight: pt.revenue > 0 ? '2px' : '0' }" :title="'Revenue: $' + pt.revenue.toFixed(4)" />
                        <div class="w-1/2 rounded-t bg-red-400 dark:bg-red-500 transition-all" :style="{ height: pt.costPct + '%', minHeight: pt.cost > 0 ? '2px' : '0' }" :title="'Cost: $' + pt.cost.toFixed(4)" />
                      </div>
                      <span class="mt-1 text-[8px] text-gray-400" v-if="idx % Math.max(1, Math.floor(detailTrend.length / 7)) === 0">{{ pt.label }}</span>
                    </div>
                  </div>
                  <div class="mt-2 flex items-center justify-center gap-4 text-xs text-gray-500">
                    <span class="flex items-center gap-1"><span class="h-2 w-2 rounded-full bg-blue-400" /> {{ t('admin.upstreamCost.internalRevenueLabel') }}</span>
                    <span class="flex items-center gap-1"><span class="h-2 w-2 rounded-full bg-red-400" /> {{ t('admin.upstreamCost.upstreamCostLabel') }}</span>
                  </div>
                </div>

                <!-- Model Cost Breakdown -->
                <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
                  <h3 class="mb-3 text-sm font-semibold text-gray-800 dark:text-gray-200">{{ t('admin.upstreamCost.modelBreakdown') }}</h3>
                  <div class="space-y-2">
                    <div v-for="(item, idx) in modelBreakdown" :key="idx" class="flex items-center gap-3">
                      <span class="w-36 truncate text-xs font-medium text-gray-700 dark:text-gray-300" :title="item.model">{{ item.model }}</span>
                      <div class="flex-1">
                        <div class="h-5 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-dark-600">
                          <div class="flex h-full items-center rounded-full bg-gradient-to-r from-primary-400 to-primary-600 px-2 text-xs font-medium text-white transition-all" :style="{ width: item.pct + '%', minWidth: item.pct > 0 ? '2rem' : '0' }">
                            {{ item.pct > 5 ? item.pct.toFixed(1) + '%' : '' }}
                          </div>
                        </div>
                      </div>
                      <span class="w-20 text-right text-xs font-medium text-amber-600 dark:text-amber-400">${{ item.cost.toFixed(4) }}</span>
                    </div>
                    <div v-if="modelBreakdown.length === 0" class="py-4 text-center text-xs text-gray-400">{{ t('admin.upstreamCost.noLogs') }}</div>
                  </div>
                </div>

                <!-- Usage Logs Table -->
                <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
                  <div class="flex items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-dark-600">
                    <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-200">
                      {{ t('admin.upstreamCost.usageLogs') }}
                      <span v-if="detailLogs.total > 0" class="ml-1 font-normal text-gray-400">({{ detailLogs.total }})</span>
                    </h3>
                  </div>
                  <div class="overflow-x-auto">
                    <table class="w-full text-left text-xs">
                      <thead class="border-b border-gray-100 bg-gray-50/50 dark:border-dark-600 dark:bg-dark-700/50">
                        <tr>
                          <th class="px-3 py-2 font-medium text-gray-500">{{ t('admin.upstreamCost.time') }}</th>
                          <th class="px-3 py-2 font-medium text-gray-500">{{ t('admin.upstreamCost.model') }}</th>
                          <th class="px-3 py-2 font-medium text-gray-500">{{ t('admin.upstreamCost.promptTokens') }}</th>
                          <th class="px-3 py-2 font-medium text-gray-500">{{ t('admin.upstreamCost.completionTokens') }}</th>
                          <th class="px-3 py-2 font-medium text-gray-500">{{ t('admin.upstreamCost.quotaUSD') }}</th>
                          <th class="px-3 py-2 font-medium text-gray-500">{{ t('admin.upstreamCost.tokenName') }}</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-gray-100 dark:divide-dark-600">
                        <tr v-for="log in detailLogs.items" :key="log.id" class="hover:bg-gray-50/50 dark:hover:bg-dark-700/30">
                          <td class="whitespace-nowrap px-3 py-1.5 text-gray-600 dark:text-gray-300">{{ formatTime(log.created_at) }}</td>
                          <td class="px-3 py-1.5"><span class="inline-flex rounded bg-blue-50 px-1.5 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{{ log.model_name }}</span></td>
                          <td class="px-3 py-1.5 text-gray-600 dark:text-gray-300">{{ log.prompt_tokens.toLocaleString() }}</td>
                          <td class="px-3 py-1.5 text-gray-600 dark:text-gray-300">{{ log.completion_tokens.toLocaleString() }}</td>
                          <td class="px-3 py-1.5 font-medium text-amber-600 dark:text-amber-400">${{ log.quota_usd.toFixed(6) }}</td>
                          <td class="px-3 py-1.5 text-gray-500 dark:text-gray-400">{{ log.token_name }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <div v-if="detailLogs.items.length < detailLogs.total" class="border-t border-gray-200 px-4 py-2 text-center dark:border-dark-600">
                    <button @click="loadMoreDetailLogs" :disabled="detailLogsLoading" class="text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
                      <span v-if="detailLogsLoading">{{ t('admin.upstreamCost.fetching') }}</span>
                      <span v-else>{{ t('admin.upstreamCost.loadMore') }}</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </Teleport>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { adminAPI } from '@/api/admin'
import { upstreamCostAPI } from '@/api/upstream-cost'
import type { UpstreamProvider, UpstreamUserInfo, UpstreamStat, UpstreamLogItem } from '@/api/upstream-cost'
import type { Channel } from '@/api/admin/channels'
import type { TrendDataPoint } from '@/types'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const STORAGE_KEY = 'upstream-cost-configs'

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

// --- Channel list ---
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

// --- Per-channel upstream configs (persisted in localStorage) ---
interface ChannelUpstreamConfig {
  provider_type: 'newapi' | 'sub2api'
  base_url: string
  access_token: string
  user_id: number
}

const channelConfigs = reactive<Record<number, ChannelUpstreamConfig>>({})

function loadConfigs() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw) as Record<string, ChannelUpstreamConfig>
      for (const [k, v] of Object.entries(parsed)) {
        channelConfigs[Number(k)] = v
      }
    }
  } catch { /* ignore */ }
}

function saveConfigs() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(channelConfigs))
  } catch { /* ignore */ }
}

function getConfig(channelId: number): ChannelUpstreamConfig {
  if (!channelConfigs[channelId]) {
    channelConfigs[channelId] = { provider_type: 'newapi', base_url: '', access_token: '', user_id: 0 }
  }
  return channelConfigs[channelId]
}

function hasData(channelId: number): boolean {
  return !!costResults[channelId]
}

function isConfigComplete(channelId: number): boolean {
  const c = channelConfigs[channelId]
  return !!c && c.base_url.trim() !== '' && c.access_token.trim() !== '' && c.user_id > 0
}

// --- Cost results per channel (with comparison) ---
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

// Global summary across all fetched channels
const globalSummary = computed(() => {
  const results = Object.values(costResults)
  if (results.length === 0) return { hasData: false, upstreamCost: 0, internalRevenue: 0, profitLoss: 0, marginPct: 0 }
  const upstreamCost = results.reduce((s, r) => s + r.upstreamCost, 0)
  const internalRevenue = results.reduce((s, r) => s + r.internalRevenue, 0)
  const profitLoss = internalRevenue - upstreamCost
  const marginPct = upstreamCost > 0 ? (profitLoss / upstreamCost) * 100 : 0
  return { hasData: true, upstreamCost, internalRevenue, profitLoss, marginPct }
})

function toProvider(channelId: number): UpstreamProvider {
  const c = getConfig(channelId)
  return { provider_type: c.provider_type, base_url: c.base_url, access_token: c.access_token, user_id: c.user_id }
}

async function fetchSingleChannelCost(channelId: number) {
  if (!isConfigComplete(channelId)) return
  fetchingChannels[channelId] = true
  delete fetchErrors[channelId]
  try {
    const prov = toProvider(channelId)
    const dr = getDateRange()

    // Fetch upstream data and internal trend in parallel
    // Find group_ids for this channel to query internal usage
    const ch = channels.value.find(c => c.id === channelId)
    const groupIds = ch?.group_ids || []

    const [ui, st, ...trendResults] = await Promise.all([
      upstreamCostAPI.getUserInfo(prov),
      upstreamCostAPI.getStats(prov),
      ...groupIds.map(gid =>
        adminAPI.dashboard.getUsageTrend({ ...dr, granularity: 'day', group_id: gid }).catch(() => ({ trend: [] as TrendDataPoint[] }))
      ),
    ])

    // Merge internal trends from all groups
    const trendMap = new Map<string, TrendDataPoint>()
    for (const tr of trendResults) {
      const resp = tr as { trend: TrendDataPoint[] }
      for (const pt of resp.trend || []) {
        const existing = trendMap.get(pt.date)
        if (existing) {
          existing.requests += pt.requests
          existing.input_tokens += pt.input_tokens
          existing.output_tokens += pt.output_tokens
          existing.total_tokens += pt.total_tokens
          existing.cost += pt.cost
          existing.actual_cost += pt.actual_cost
        } else {
          trendMap.set(pt.date, { ...pt })
        }
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
  const configured = channels.value.filter(ch => isConfigComplete(ch.id))
  await Promise.allSettled(configured.map(ch => fetchSingleChannelCost(ch.id)))
  batchFetching.value = false
  if (configured.length === 0) {
    appStore.showError(t('admin.upstreamCost.noConfiguredChannels'))
  }
}

// --- Detail slide-over ---
const detailChannelId = ref<number | null>(null)
const detailChannel = computed(() => channels.value.find(ch => ch.id === detailChannelId.value))
const detailLogs = reactive<{ items: UpstreamLogItem[]; total: number; page: number }>({ items: [], total: 0, page: 0 })
const detailLogsLoading = ref(false)

// Trend data for the detail chart
const detailTrend = computed(() => {
  if (detailChannelId.value === null || !costResults[detailChannelId.value]) return []
  const trend = costResults[detailChannelId.value].trend
  if (!trend.length) return []
  const maxVal = Math.max(...trend.map(pt => Math.max(pt.actual_cost, 0.001)))
  return trend.map(pt => ({
    label: pt.date.slice(5), // MM-DD
    revenue: pt.actual_cost,
    cost: 0, // We don't have daily upstream cost breakdown, show only revenue
    revPct: maxVal > 0 ? (pt.actual_cost / maxVal) * 100 : 0,
    costPct: 0,
  }))
})

const modelBreakdown = computed(() => {
  if (!detailLogs.items.length) return []
  const map = new Map<string, number>()
  for (const log of detailLogs.items) {
    map.set(log.model_name, (map.get(log.model_name) || 0) + log.quota_usd)
  }
  const sorted = [...map.entries()].sort((a, b) => b[1] - a[1])
  const total = sorted.reduce((s, [, v]) => s + v, 0)
  return sorted.map(([model, cost]) => ({ model, cost, pct: total > 0 ? (cost / total) * 100 : 0 }))
})

async function openDetail(channelId: number) {
  detailChannelId.value = channelId
  detailLogs.items = []
  detailLogs.total = 0
  detailLogs.page = 0
  detailLogsLoading.value = true
  try {
    const prov = toProvider(channelId)
    const lg = await upstreamCostAPI.getLogs(prov, { page: 0, page_size: 50 })
    detailLogs.items = lg.items || []
    detailLogs.total = lg.total
  } catch (err: any) {
    appStore.showError(err?.message || 'Failed to load logs')
  } finally {
    detailLogsLoading.value = false
  }
}

async function loadMoreDetailLogs() {
  if (detailChannelId.value === null) return
  detailLogsLoading.value = true
  detailLogs.page++
  try {
    const prov = toProvider(detailChannelId.value)
    const lg = await upstreamCostAPI.getLogs(prov, { page: detailLogs.page, page_size: 50 })
    detailLogs.items.push(...(lg.items || []))
    detailLogs.total = lg.total
  } catch (err: any) {
    appStore.showError(err?.message || 'Failed to load more logs')
    detailLogs.page--
  } finally {
    detailLogsLoading.value = false
  }
}

// --- Formatting helpers ---
function formatTime(ts: number): string {
  const d = new Date(ts * 1000)
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mi = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  return `${mm}-${dd} ${hh}:${mi}:${ss}`
}

// --- Init ---
onMounted(async () => {
  loadConfigs()
  await loadChannels()
})
</script>

<style scoped>
.slide-over-enter-active,
.slide-over-leave-active {
  transition: all 0.3s ease;
}
.slide-over-enter-active > div:last-child,
.slide-over-leave-active > div:last-child {
  transition: transform 0.3s ease;
}
.slide-over-enter-from,
.slide-over-leave-to {
  opacity: 0;
}
.slide-over-enter-from > div:last-child,
.slide-over-leave-to > div:last-child {
  transform: translateX(100%);
}
</style>
