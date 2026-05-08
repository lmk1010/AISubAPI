<template>
  <div class="trend-card">
    <h3 class="trend-card__title">{{ t('admin.dashboard.tokenUsageTrend') }}</h3>

    <div class="trend-card__body">
      <div v-if="loading" class="trend-card__loading">
        <LoadingSpinner />
      </div>
      <div v-else-if="trendData.length > 0 && chartData" class="trend-card__chart">
        <Line :data="chartData" :options="lineOptions" />
      </div>
      <div v-else class="trend-card__empty">
        <div class="trend-card__empty-art" aria-hidden="true">
          <svg viewBox="0 0 200 80" fill="none" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="trendGrad" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stop-color="#3b82f6" />
                <stop offset="50%" stop-color="#7c3aed" />
                <stop offset="100%" stop-color="#06b6d4" />
              </linearGradient>
            </defs>
            <path d="M 4 60 Q 30 40 50 50 T 100 30 T 150 40 T 196 20" stroke="url(#trendGrad)" stroke-width="2.5" fill="none" stroke-linecap="round" />
            <path d="M 4 60 Q 30 40 50 50 T 100 30 T 150 40 T 196 20 L 196 76 L 4 76 Z" fill="url(#trendGrad)" opacity="0.08" />
            <circle cx="196" cy="20" r="3.5" fill="#7c3aed" />
          </svg>
        </div>
        <p class="trend-card__empty-title">{{ t('admin.dashboard.noDataAvailable') }}</p>
        <p class="trend-card__empty-hint">{{ t('admin.dashboard.noTokenTrendInRange') }}</p>
      </div>
    </div>

    <div class="trend-card__legend">
      <span class="legend-pill"><i class="legend-dot" style="background:#3b82f6"></i>{{ t('admin.dashboard.input') }} Token</span>
      <span class="legend-pill"><i class="legend-dot" style="background:#10b981"></i>{{ t('admin.dashboard.output') }} Token</span>
      <span class="legend-pill"><i class="legend-dot" style="background:#7c3aed"></i>{{ t('common.total') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { TrendDataPoint } from '@/types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()

const props = defineProps<{
  trendData: TrendDataPoint[]
  loading?: boolean
}>()

const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb',
  input: '#3b82f6',
  output: '#10b981',
  cacheCreation: '#f59e0b',
  cacheRead: '#06b6d4',
  cacheHitRate: '#8b5cf6'
}))

const chartData = computed(() => {
  if (!props.trendData?.length) return null

  return {
    labels: props.trendData.map((d) => d.date),
    datasets: [
      {
        label: 'Input',
        data: props.trendData.map((d) => d.input_tokens),
        borderColor: chartColors.value.input,
        backgroundColor: `${chartColors.value.input}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Output',
        data: props.trendData.map((d) => d.output_tokens),
        borderColor: chartColors.value.output,
        backgroundColor: `${chartColors.value.output}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Cache Creation',
        data: props.trendData.map((d) => d.cache_creation_tokens),
        borderColor: chartColors.value.cacheCreation,
        backgroundColor: `${chartColors.value.cacheCreation}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Cache Read',
        data: props.trendData.map((d) => d.cache_read_tokens),
        borderColor: chartColors.value.cacheRead,
        backgroundColor: `${chartColors.value.cacheRead}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: 'Cache Hit Rate',
        data: props.trendData.map((d) => {
          const total = d.cache_read_tokens + d.cache_creation_tokens
          return total > 0 ? (d.cache_read_tokens / total) * 100 : 0
        }),
        borderColor: chartColors.value.cacheHitRate,
        backgroundColor: `${chartColors.value.cacheHitRate}20`,
        borderDash: [5, 5],
        fill: false,
        tension: 0.3,
        yAxisID: 'yPercent'
      }
    ]
  }
})

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
      callbacks: {
        label: (context: any) => {
          if (context.dataset.yAxisID === 'yPercent') {
            return `${context.dataset.label}: ${context.raw.toFixed(1)}%`
          }
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
        },
        footer: (tooltipItems: any) => {
          const dataIndex = tooltipItems[0]?.dataIndex
          if (dataIndex !== undefined && props.trendData[dataIndex]) {
            const data = props.trendData[dataIndex]
            return `Actual: $${formatCost(data.actual_cost)} | Standard: $${formatCost(data.cost)}`
          }
          return ''
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
    },
    yPercent: {
      position: 'right' as const,
      min: 0,
      max: 100,
      grid: {
        drawOnChartArea: false
      },
      ticks: {
        color: chartColors.value.cacheHitRate,
        font: {
          size: 10
        },
        callback: (value: string | number) => `${value}%`
      }
    }
  }
}))

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
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
</script>

<style scoped>
.trend-card {
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

.trend-card__title {
  font-size: 15px;
  font-weight: 600;
  color: rgb(15 23 42);
  margin: 0 0 16px;
}

.trend-card__body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.trend-card__loading,
.trend-card__chart {
  width: 100%;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.trend-card__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 8px;
}

.trend-card__empty-art {
  width: 220px;
  height: 90px;
}

.trend-card__empty-title {
  font-size: 14px;
  color: rgb(71 85 105);
  font-weight: 500;
  margin: 4px 0 0;
}

.trend-card__empty-hint {
  font-size: 12px;
  color: rgb(148 163 184);
  margin: 0;
}

.trend-card__legend {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 18px;
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

.legend-pill::after {
  content: '';
  display: inline-block;
  width: 18px;
  height: 1.5px;
  background: currentColor;
  margin-left: 4px;
  opacity: 0.5;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

html.dark .trend-card {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.5);
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.35),
    0 1px 2px rgba(0, 0, 0, 0.25);
}
html.dark .trend-card__title {
  color: rgb(248 250 252);
}
html.dark .trend-card__empty-title {
  color: rgb(203 213 225);
}
html.dark .trend-card__empty-hint,
html.dark .legend-pill {
  color: rgb(148 163 184);
}
</style>
