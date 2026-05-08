<template>
  <AppLayout>
    <div class="admin-dashboard space-y-4">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <!-- Hero panel: balance hero on the left + 7 stat tiles (3+4) on the right -->
        <section class="hero-panel">
          <!-- Left: balance hero with the 3D coin-card illustration asset -->
          <div class="hero-balance">
            <div class="hero-balance__text">
              <p class="hero-balance__label">{{ t('dashboard.balance') }}</p>
              <p class="hero-balance__value">${{ formatBalance(authStore.user?.balance || 0) }}</p>
              <p class="hero-balance__hint">{{ t('common.available') }}</p>
            </div>
            <img class="hero-balance__art" src="/balance-illustration.png?v=2" alt="" aria-hidden="true" />
          </div>

          <!-- Right: 3+4 stats grid with internal hairline dividers (per design) -->
          <div class="hero-stats">
            <div class="hero-stats__row hero-stats__row--three">
            <!-- Users -->
            <div class="stat-tile">
              <span class="stat-tile__icon stat-tile__icon--violet">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
                </svg>
              </span>
              <div class="stat-tile__body">
                <p class="stat-tile__label">{{ t('admin.dashboard.users') }}</p>
                <p class="stat-tile__value">+{{ stats.today_new_users }}</p>
                <p class="stat-tile__hint">{{ t('common.total') }}: {{ formatNumber(stats.total_users) }}</p>
              </div>
            </div>

            <!-- API Keys -->
            <div class="stat-tile">
              <span class="stat-tile__icon stat-tile__icon--blue">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z" />
                </svg>
              </span>
              <div class="stat-tile__body">
                <p class="stat-tile__label">{{ t('admin.dashboard.apiKeys') }}</p>
                <p class="stat-tile__value">{{ stats.total_api_keys }}</p>
                <p class="stat-tile__hint hero-positive">{{ stats.active_api_keys }} {{ t('common.active') }}</p>
              </div>
            </div>

            <!-- Today Requests -->
            <div class="stat-tile">
              <span class="stat-tile__icon stat-tile__icon--green">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
                </svg>
              </span>
              <div class="stat-tile__body">
                <p class="stat-tile__label">{{ t('admin.dashboard.todayRequests') }}</p>
                <p class="stat-tile__value">{{ stats.today_requests }}</p>
                <p class="stat-tile__hint">{{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}</p>
              </div>
            </div>
            </div><!-- /row--three -->

            <div class="hero-stats__row hero-stats__row--four">
            <!-- Today Tokens -->
            <div class="stat-tile">
              <span class="stat-tile__icon stat-tile__icon--amber">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
                </svg>
              </span>
              <div class="stat-tile__body">
                <p class="stat-tile__label">{{ t('admin.dashboard.todayTokens') }}</p>
                <p class="stat-tile__value">{{ formatTokens(stats.today_tokens) }}</p>
                <p class="stat-tile__hint">
                  <span class="stat-tile__hint-strong">${{ formatCost(stats.today_actual_cost) }}</span>
                  <span class="stat-tile__hint-muted"> / ${{ formatCost(stats.today_cost) }}</span>
                </p>
              </div>
            </div>

            <!-- Total Tokens -->
            <div class="stat-tile">
              <span class="stat-tile__icon stat-tile__icon--indigo">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75" />
                </svg>
              </span>
              <div class="stat-tile__body">
                <p class="stat-tile__label">{{ t('admin.dashboard.totalTokens') }}</p>
                <p class="stat-tile__value">{{ formatTokens(stats.total_tokens) }}</p>
                <p class="stat-tile__hint">
                  <span class="stat-tile__hint-strong">${{ formatCost(stats.total_actual_cost) }}</span>
                  <span class="stat-tile__hint-muted"> / ${{ formatCost(stats.total_cost) }}</span>
                </p>
              </div>
            </div>

            <!-- Performance (RPM) -->
            <div class="stat-tile">
              <span class="stat-tile__icon stat-tile__icon--purple">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
                </svg>
              </span>
              <div class="stat-tile__body">
                <p class="stat-tile__label">{{ t('admin.dashboard.performance') }}</p>
                <p class="stat-tile__value">
                  {{ formatTokens(stats.rpm) }}<span class="stat-tile__value-unit"> RPM</span>
                </p>
                <p class="stat-tile__hint">
                  <span class="stat-tile__hint-strong">{{ formatTokens(stats.tpm) }}</span> TPM
                </p>
              </div>
            </div>

            <!-- Avg Response -->
            <div class="stat-tile">
              <span class="stat-tile__icon stat-tile__icon--rose">
                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </span>
              <div class="stat-tile__body">
                <p class="stat-tile__label">{{ t('admin.dashboard.avgResponse') }}</p>
                <p class="stat-tile__value">{{ formatDurationMs(stats.average_duration_ms) }}</p>
                <p class="stat-tile__hint">{{ t('admin.dashboard.averageTime') }}</p>
              </div>
            </div>
            </div><!-- /row--four -->
          </div>
        </section>

        <!-- Charts Section -->
        <div class="space-y-4">
          <!-- Date Range Filter: clean white pill row -->
          <div class="filter-bar">
            <div class="filter-bar__group">
              <span class="filter-bar__label">{{ t('admin.dashboard.timeRange') }}</span>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="onDateRangeChange"
              />
              <button
                @click="loadDashboardStats"
                :disabled="chartsLoading"
                class="filter-bar__refresh"
                :title="t('common.refresh')"
                aria-label="refresh"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </button>
            </div>
            <div class="filter-bar__group filter-bar__group--end">
              <span class="filter-bar__label">{{ t('admin.dashboard.granularity') }}</span>
              <div class="w-28">
                <Select
                  v-model="granularity"
                  :options="granularityOptions"
                  @change="loadChartData"
                />
              </div>
            </div>
          </div>

          <!-- Charts Grid -->
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <ModelDistributionChart
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-actual-cost="rankingTotalActualCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToUserUsage"
            />
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>

          <!-- User Usage Trend (Full Width) -->
          <div class="flat-card">
            <h3 class="flat-card__title">
              {{ t('admin.dashboard.recentUsage') }} (Top 12)
            </h3>
            <div class="h-64">
              <div v-if="userTrendLoading" class="flex h-full items-center justify-center">
                <LoadingSpinner size="md" />
              </div>
              <Line v-else-if="userTrendChartData" :data="userTrendChartData" :options="lineOptions" />
              <div
                v-else
                class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400"
              >
                {{ t('admin.dashboard.noDataAvailable') }}
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  TrendDataPoint,
  ModelStat,
  UserUsageTrendPoint,
  UserSpendingRankingItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

const appStore = useAppStore()
const authStore = useAuthStore()
const router = useRouter()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const userTrendLoading = ref(false)
const rankingLoading = ref(false)
const rankingError = ref(false)

// Chart data
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const userTrend = ref<UserUsageTrendPoint[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const rankingTotalActualCost = ref(0)
const rankingTotalRequests = ref(0)
const rankingTotalTokens = ref(0)
let chartLoadSeq = 0
let usersTrendLoadSeq = 0
let rankingLoadSeq = 0
const rankingLimit = 12

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const getLast24HoursRangeDates = (): { start: string; end: string } => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return {
    start: formatLocalDate(start),
    end: formatLocalDate(end)
  }
}

// Date range
const granularity = ref<'day' | 'hour'>('hour')
const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

// Granularity options for Select component
const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])

// Dark mode detection
const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

// Chart colors
const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb'
}))

// Line chart options (for user trend chart)
const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      itemSort: (a: any, b: any) => {
        const aValue = typeof a?.raw === 'number' ? a.raw : Number(a?.parsed?.y ?? 0)
        const bValue = typeof b?.raw === 'number' ? b.raw : Number(b?.parsed?.y ?? 0)
        return bValue - aValue
      },
      callbacks: {
        label: (context: any) => {
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        }
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    }
  }
}))

// User trend chart data
const userTrendChartData = computed(() => {
  if (!userTrend.value?.length) return null

  const getDisplayName = (point: UserUsageTrendPoint): string => {
    const username = point.username?.trim()
    if (username) {
      return username
    }

    const email = point.email?.trim()
    if (email) {
      return email
    }

    return t('admin.redeem.userPrefix', { id: point.user_id })
  }

  // Group by user_id to avoid merging different users with the same display name
  const userGroups = new Map<number, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  userTrend.value.forEach((point) => {
    allDates.add(point.date)
    const key = point.user_id
    if (!userGroups.has(key)) {
      userGroups.set(key, { name: getDisplayName(point), data: new Map() })
    }
    userGroups.get(key)!.data.set(point.date, point.tokens)
  })

  const sortedDates = Array.from(allDates).sort()
  const colors = [
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ef4444',
    '#8b5cf6',
    '#ec4899',
    '#14b8a6',
    '#f97316',
    '#6366f1',
    '#84cc16',
    '#06b6d4',
    '#a855f7'
  ]

  const datasets = Array.from(userGroups.values()).map((group, idx) => ({
    label: group.name,
    data: sortedDates.map((date) => group.data.get(date) || 0),
    borderColor: colors[idx % colors.length],
    backgroundColor: `${colors[idx % colors.length]}20`,
    fill: false,
    tension: 0.3
  }))

  return {
    labels: sortedDates,
    datasets
  }
})

// Format helpers
const formatTokens = (value: number | undefined): string => {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatBalance = (b: number): string =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatDurationMs = (ms: number): string =>
  ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${Math.round(ms)}ms`

const formatNumber = (value: number): string => {
  return value.toLocaleString()
}

const formatCost = (value: number): string => {
  if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  } else if (value >= 1) {
    return value.toFixed(2)
  } else if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}

const goToUserUsage = (item: UserSpendingRankingItem) => {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(item.user_id),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

// Date range change handler
const onDateRangeChange = (range: {
  startDate: string
  endDate: string
  preset: string | null
}) => {
  // Auto-select granularity based on date range
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))

  // If range is 1 day, use hourly granularity
  if (daysDiff <= 1) {
    granularity.value = 'hour'
  } else {
    granularity.value = 'day'
  }

  loadChartData()
}

// Load data
const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) {
    loading.value = true
  }
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUsersTrend = async () => {
  const currentSeq = ++usersTrendLoadSeq
  userTrendLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: 12
    })
    if (currentSeq !== usersTrendLoadSeq) return
    userTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== usersTrendLoadSeq) return
    console.error('Error loading users trend:', error)
    userTrend.value = []
  } finally {
    if (currentSeq === usersTrendLoadSeq) {
      userTrendLoading.value = false
    }
  }
}

const loadUserSpendingRanking = async () => {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  rankingError.value = false
  try {
    const response = await adminAPI.dashboard.getUserSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
    rankingTotalActualCost.value = response.total_actual_cost || 0
    rankingTotalRequests.value = response.total_requests || 0
    rankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading user spending ranking:', error)
    rankingItems.value = []
    rankingTotalActualCost.value = 0
    rankingTotalRequests.value = 0
    rankingTotalTokens.value = 0
    rankingError.value = true
  } finally {
    if (currentSeq === rankingLoadSeq) {
      rankingLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

onMounted(() => {
  loadDashboardStats()
})
</script>

<style scoped>
/* ============ Hero panel: balance hero on the left + 7 stat tiles on the right ============ */
.hero-panel {
  display: grid;
  grid-template-columns: minmax(240px, 280px) 1fr;
  gap: 0;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 12px;
  overflow: hidden;
  min-height: 240px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.08),
    0 1px 2px rgba(15, 23, 42, 0.04);
}

@media (max-width: 1023px) {
  .hero-panel {
    grid-template-columns: 1fr;
  }
}

/* ---------- Left: balance hero (text fills the left column with even
   vertical distribution; the 3D illustration sits on the right and
   overlaps with the lower text per the design mock) ---------- */
.hero-balance {
  position: relative;
  padding: 22px 24px;
  overflow: hidden;
  isolation: isolate;
  min-height: inherit;
  display: flex;
  border-right: 1px solid rgba(167, 139, 250, 0.18);
}
html.dark .hero-balance {
  border-right-color: rgba(124, 58, 237, 0.28);
}

.hero-balance__text {
  position: relative;
  z-index: 1;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-width: 0;
}

.hero-balance__label {
  font-size: 13px;
  color: rgb(71 85 105);
  margin: 0;
  font-weight: 500;
}

.hero-balance__value {
  font-size: 36px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: rgb(124 58 237);
  line-height: 1;
  margin: 0;
}

.hero-balance__hint {
  font-size: 12px;
  color: rgb(148 163 184);
  margin: 0;
}

.hero-balance__art {
  position: absolute;
  right: -10px;
  bottom: -8px;
  width: 180px;
  height: 180px;
  z-index: 0;
  object-fit: contain;
  user-select: none;
  pointer-events: none;
}

@media (max-width: 1279px) {
  .hero-balance__art {
    width: 150px;
    height: 150px;
  }
  .hero-balance__value {
    font-size: 32px;
  }
}

/* ---------- Right: 3+4 stats grid with internal hairline dividers ---------- */
.hero-stats {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.hero-stats__row {
  display: grid;
  align-items: center;
}
.hero-stats__row--three {
  grid-template-columns: repeat(3, 1fr);
  border-bottom: 1px solid rgba(167, 139, 250, 0.18);
}
.hero-stats__row--four {
  grid-template-columns: repeat(4, 1fr);
}

.hero-stats__row .stat-tile {
  padding: 16px 20px;
  border-right: 1px solid rgba(167, 139, 250, 0.18);
}
.hero-stats__row--three .stat-tile:nth-child(3),
.hero-stats__row--four .stat-tile:nth-child(4) {
  border-right: none;
}

@media (max-width: 1023px) {
  .hero-stats__row .stat-tile {
    padding: 14px 16px;
  }
}

@media (max-width: 767px) {
  .hero-stats__row--three,
  .hero-stats__row--four {
    grid-template-columns: repeat(2, 1fr);
  }
  .hero-stats__row .stat-tile {
    border-right: none;
  }
  .hero-stats__row .stat-tile:nth-child(odd) {
    border-right: 1px solid rgba(167, 139, 250, 0.18);
  }
}

@media (max-width: 479px) {
  .hero-stats__row--three,
  .hero-stats__row--four {
    grid-template-columns: 1fr;
  }
  .hero-stats__row .stat-tile {
    border-right: none !important;
  }
}

html.dark .hero-stats__row--three {
  border-bottom-color: rgba(124, 58, 237, 0.28);
}
html.dark .hero-stats__row .stat-tile {
  border-right-color: rgba(124, 58, 237, 0.28);
}

/* ---------- Individual stat tile (rounded-square icon + label/value/hint) ---------- */
.stat-tile {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.stat-tile__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  flex-shrink: 0;
}
.stat-tile__icon svg {
  width: 16px;
  height: 16px;
}

.stat-tile__icon--blue {
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}
.stat-tile__icon--green {
  background: rgb(220 252 231);
  color: rgb(22 163 74);
}
.stat-tile__icon--purple {
  background: rgb(237 233 254);
  color: rgb(124 58 237);
}
.stat-tile__icon--violet {
  background: rgb(237 233 254);
  color: rgb(139 92 246);
}
.stat-tile__icon--amber {
  background: rgb(254 243 199);
  color: rgb(217 119 6);
}
.stat-tile__icon--indigo {
  background: rgb(224 231 255);
  color: rgb(99 102 241);
}
.stat-tile__icon--rose {
  background: rgb(255 228 230);
  color: rgb(244 63 94);
}

.stat-tile__body {
  min-width: 0;
  flex: 1;
}

.stat-tile__label {
  font-size: 11px;
  color: rgb(100 116 139);
  margin: 0 0 2px;
}

.stat-tile__value {
  font-size: 16px;
  font-weight: 700;
  color: rgb(15 23 42);
  margin: 0 0 1px;
  line-height: 1.15;
  letter-spacing: -0.01em;
}

.stat-tile__value-unit {
  font-size: 12px;
  font-weight: 500;
  color: rgb(148 163 184);
}

.stat-tile__hint {
  font-size: 11px;
  color: rgb(100 116 139);
  margin: 0;
}

.stat-tile__hint-strong {
  color: rgb(15 23 42);
  font-weight: 600;
}

.stat-tile__hint-muted {
  color: rgb(148 163 184);
}

.hero-positive {
  color: rgb(22 163 74);
}

/* Override the legacy `.card` background used by ModelDistributionChart so
   it picks up the same translucent panel treatment as the rest of the
   dashboard cards. Scoped via :deep() to only affect the dashboard view. */
.admin-dashboard :deep(.card) {
  background: rgba(255, 255, 255, 0.72) !important;
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.85) !important;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.08),
    0 1px 2px rgba(15, 23, 42, 0.04) !important;
}
html.dark .admin-dashboard :deep(.card) {
  background: rgba(15, 23, 42, 0.7) !important;
  border-color: rgba(71, 85, 105, 0.5) !important;
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.35),
    0 1px 2px rgba(0, 0, 0, 0.25) !important;
}

/* ============ Generic flat card (User Usage Trend) ============ */
.flat-card {
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 12px;
  padding: 18px 20px 16px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.08),
    0 1px 2px rgba(15, 23, 42, 0.04);
}
.flat-card__title {
  font-size: 15px;
  font-weight: 600;
  color: rgb(15 23 42);
  margin: 0 0 16px;
}

/* ============ Filter bar (truly flat) ============ */
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 8px 14px;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 10px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.08),
    0 1px 2px rgba(15, 23, 42, 0.04);
}

/* Compact inner pickers/selects so the filter row stays slim. */
.filter-bar :deep(.date-picker-trigger),
.filter-bar :deep(.select-trigger) {
  padding: 5px 10px;
  font-size: 12px;
  border-radius: 8px;
}

.filter-bar__group {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex-wrap: wrap;
}

.filter-bar__group--end {
  margin-left: auto;
}

.filter-bar__label {
  font-size: 13px;
  font-weight: 500;
  color: rgb(100 116 139);
  white-space: nowrap;
}

.filter-bar__refresh {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  color: rgb(100 116 139);
  background: transparent;
  transition: background 0.18s ease, color 0.18s ease;
}
.filter-bar__refresh:hover:not(:disabled) {
  background: rgb(245 243 255);
  color: rgb(124 58 237);
}
.filter-bar__refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ============ Dark mode ============ */
html.dark .hero-panel {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(18px) saturate(180%);
  border-color: rgba(71, 85, 105, 0.5);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.35),
    0 1px 2px rgba(0, 0, 0, 0.25);
}
html.dark .hero-balance__label {
  color: rgb(203 213 225);
}
html.dark .hero-balance__value {
  color: rgb(196 181 253);
}
html.dark .hero-balance__hint,
html.dark .stat-tile__label,
html.dark .stat-tile__hint,
html.dark .filter-bar__label {
  color: rgb(148 163 184);
}
html.dark .stat-tile__value {
  color: rgb(248 250 252);
}
html.dark .stat-tile__hint-strong {
  color: rgb(248 250 252);
}

html.dark .stat-tile__icon--blue {
  background: rgba(37, 99, 235, 0.22);
  color: rgb(147 197 253);
}
html.dark .stat-tile__icon--green {
  background: rgba(22, 163, 74, 0.22);
  color: rgb(134 239 172);
}
html.dark .stat-tile__icon--purple {
  background: rgba(124, 58, 237, 0.22);
  color: rgb(196 181 253);
}
html.dark .stat-tile__icon--violet {
  background: rgba(139, 92, 246, 0.22);
  color: rgb(196 181 253);
}
html.dark .stat-tile__icon--amber {
  background: rgba(217, 119, 6, 0.22);
  color: rgb(253 186 116);
}
html.dark .stat-tile__icon--indigo {
  background: rgba(99, 102, 241, 0.22);
  color: rgb(165 180 252);
}
html.dark .stat-tile__icon--rose {
  background: rgba(244, 63, 94, 0.22);
  color: rgb(252 165 165);
}

html.dark .flat-card {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.5);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.35),
    0 1px 2px rgba(0, 0, 0, 0.25);
}
html.dark .flat-card__title {
  color: rgb(248 250 252);
}

html.dark .filter-bar {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.5);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.35),
    0 1px 2px rgba(0, 0, 0, 0.25);
}
html.dark .filter-bar__refresh:hover:not(:disabled) {
  background: rgba(124, 58, 237, 0.15);
  color: rgb(196 181 253);
}
</style>
