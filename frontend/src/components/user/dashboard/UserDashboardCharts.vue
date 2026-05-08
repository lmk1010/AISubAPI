<template>
  <div class="space-y-4">
    <!-- Date Range Filter: clean white pill row matching the design -->
    <div class="filter-bar">
      <div class="filter-bar__group">
        <span class="filter-bar__label">{{ t('dashboard.timeRange') }}</span>
        <DateRangePicker :start-date="startDate" :end-date="endDate" @update:startDate="$emit('update:startDate', $event)" @update:endDate="$emit('update:endDate', $event)" @change="$emit('dateRangeChange', $event)" />
        <button
          @click="$emit('refresh')"
          :disabled="loading"
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
        <span class="filter-bar__label">{{ t('dashboard.granularity') }}</span>
        <div class="w-28">
          <Select :model-value="granularity" :options="[{value:'day', label:t('dashboard.day')}, {value:'hour', label:t('dashboard.hour')}]" @update:model-value="$emit('update:granularity', $event)" @change="$emit('granularityChange')" />
        </div>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Model Distribution Chart -->
      <div class="chart-card">
        <div v-if="loading" class="chart-card__loading">
          <LoadingSpinner size="md" />
        </div>
        <h3 class="chart-card__title">{{ t('dashboard.modelDistribution') }}</h3>

        <div class="chart-card__body">
          <div v-if="modelData" class="chart-card__doughnut">
            <Doughnut :data="modelData" :options="doughnutOptions" />
          </div>
          <div v-else class="chart-card__empty">
            <div class="chart-card__empty-art" aria-hidden="true">
              <svg viewBox="0 0 120 80" fill="none" xmlns="http://www.w3.org/2000/svg">
                <defs>
                  <linearGradient id="emptyDonutGrad" x1="0" y1="0" x2="1" y2="1">
                    <stop offset="0%" stop-color="#a78bfa" />
                    <stop offset="100%" stop-color="#7c3aed" />
                  </linearGradient>
                </defs>
                <circle cx="60" cy="44" r="22" fill="url(#emptyDonutGrad)" opacity="0.18" />
                <circle cx="60" cy="44" r="22" stroke="#a78bfa" stroke-width="6" fill="none" />
                <circle cx="60" cy="44" r="22" stroke="#06b6d4" stroke-width="6" fill="none" stroke-dasharray="35 200" stroke-dashoffset="-30" />
                <circle cx="60" cy="44" r="22" stroke="#7c3aed" stroke-width="6" fill="none" stroke-dasharray="22 200" stroke-dashoffset="-72" />
                <rect x="20" y="14" width="6" height="22" rx="2" fill="#c4b5fd" opacity="0.6" />
                <rect x="30" y="22" width="6" height="14" rx="2" fill="#a78bfa" opacity="0.6" />
                <rect x="92" y="18" width="6" height="18" rx="2" fill="#06b6d4" opacity="0.6" />
                <rect x="102" y="24" width="6" height="12" rx="2" fill="#7c3aed" opacity="0.5" />
              </svg>
            </div>
            <p class="chart-card__empty-title">{{ t('dashboard.noDataAvailable') }}</p>
            <p class="chart-card__empty-hint">{{ t('dashboard.noModelDataInRange') }}</p>
          </div>
        </div>

        <div class="chart-card__legend">
          <span class="legend-pill"><i class="legend-dot" style="background:#3b82f6"></i>{{ t('dashboard.model') }}</span>
          <span class="legend-pill"><i class="legend-dot" style="background:#7c3aed"></i>{{ t('dashboard.requests') }}</span>
          <span class="legend-pill"><i class="legend-dot" style="background:#14b8a6"></i>Token</span>
          <span class="legend-pill"><i class="legend-dot" style="background:#f59e0b"></i>{{ t('dashboard.actual') }}</span>
          <span class="legend-pill"><i class="legend-dot" style="background:#94a3b8"></i>{{ t('dashboard.standard') }}</span>
        </div>
      </div>

      <!-- Token Usage Trend Chart -->
      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatTokensK as formatTokens } from '@/utils/format'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler } from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ loading: boolean, startDate: string, endDate: string, granularity: string, trend: TrendDataPoint[], models: ModelStat[] }>()
defineEmits(['update:startDate', 'update:endDate', 'update:granularity', 'dateRangeChange', 'granularityChange', 'refresh'])
const { t } = useI18n()

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: ['#7c3aed', '#a78bfa', '#06b6d4', '#3b82f6', '#10b981', '#f59e0b', '#ec4899', '#94a3b8'],
    borderColor: '#ffffff',
    borderWidth: 3,
    hoverOffset: 6,
  }]
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '64%',
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens`
      }
    }
  }
}
</script>

<style scoped>
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

/* Compact the inner DateRangePicker / Select inside the filter bar so the
   row stays slim, matching the design's tight pill row. */
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
  transition:
    background 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}
.filter-bar__refresh:hover:not(:disabled) {
  background: rgb(245 243 255);
  color: rgb(124 58 237);
}
.filter-bar__refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

html.dark .filter-bar {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.5);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.35),
    0 1px 2px rgba(0, 0, 0, 0.25);
}
html.dark .filter-bar__label {
  color: rgb(148 163 184);
}
html.dark .filter-bar__refresh:hover:not(:disabled) {
  background: rgba(124, 58, 237, 0.15);
  color: rgb(196 181 253);
}

/* ============ Chart card (donut) — truly flat ============ */
.chart-card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 12px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.08),
    0 1px 2px rgba(15, 23, 42, 0.04);
  padding: 16px 18px 14px;
  min-height: 240px;
}

.chart-card__loading {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(4px);
  border-radius: inherit;
}

.chart-card__title {
  font-size: 15px;
  font-weight: 600;
  color: rgb(15 23 42);
  margin: 0 0 16px;
}

.chart-card__body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.chart-card__doughnut {
  width: 200px;
  height: 200px;
}

.chart-card__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 8px;
  padding: 8px 0;
}

.chart-card__empty-art {
  width: 140px;
  height: 100px;
}

.chart-card__empty-title {
  font-size: 14px;
  color: rgb(71 85 105);
  font-weight: 500;
  margin: 4px 0 0;
}

.chart-card__empty-hint {
  font-size: 12px;
  color: rgb(148 163 184);
  margin: 0;
}

.chart-card__legend {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding-top: 16px;
  margin-top: auto;
}

.legend-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgb(100 116 139);
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

html.dark .chart-card {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.5);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.35),
    0 1px 2px rgba(0, 0, 0, 0.25);
}
html.dark .chart-card__title {
  color: rgb(248 250 252);
}
html.dark .chart-card__empty-title {
  color: rgb(203 213 225);
}
html.dark .chart-card__empty-hint,
html.dark .legend-pill {
  color: rgb(148 163 184);
}
html.dark .chart-card__loading {
  background: rgba(15, 23, 42, 0.55);
}
</style>
