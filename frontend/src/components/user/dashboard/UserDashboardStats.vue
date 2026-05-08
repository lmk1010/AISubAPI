<template>
  <section class="hero-panel">
    <!-- Left: balance hero with the 3D coin-card illustration asset -->
    <div class="hero-balance">
      <div class="hero-balance__text">
        <p class="hero-balance__label">{{ t('dashboard.balance') }}</p>
        <p class="hero-balance__value">${{ formatBalance(balance) }}</p>
        <p class="hero-balance__hint">{{ t('common.available') }}</p>
      </div>
      <img class="hero-balance__art" src="/balance-illustration.png?v=2" alt="" aria-hidden="true" />
    </div>

    <!-- Right: 7 stat tiles in a 3+4 layout with internal dividers (per design) -->
    <div class="hero-stats">
      <div class="hero-stats__row hero-stats__row--three">
        <!-- API Keys -->
        <div class="stat-tile">
        <span class="stat-tile__icon stat-tile__icon--blue">
          <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z" />
          </svg>
        </span>
        <div class="stat-tile__body">
          <p class="stat-tile__label">{{ t('dashboard.apiKeys') }}</p>
          <p class="stat-tile__value">{{ stats?.total_api_keys || 0 }}</p>
          <p class="stat-tile__hint hero-positive">{{ stats?.active_api_keys || 0 }} {{ t('common.active') }}</p>
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
          <p class="stat-tile__label">{{ t('dashboard.todayRequests') }}</p>
          <p class="stat-tile__value">{{ stats?.today_requests || 0 }}</p>
          <p class="stat-tile__hint">{{ t('common.total') }}: {{ formatNumber(stats?.total_requests || 0) }}</p>
        </div>
      </div>

      <!-- Today Cost -->
      <div class="stat-tile">
        <span class="stat-tile__icon stat-tile__icon--purple">
          <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-2.21 0-4-1.79-4-4s1.79-4 4-4 4 1.79 4 4" />
          </svg>
        </span>
        <div class="stat-tile__body">
          <p class="stat-tile__label">{{ t('dashboard.todayCost') }}</p>
          <p class="stat-tile__value stat-tile__value--violet">${{ formatCost(stats?.today_actual_cost || 0) }}</p>
          <p class="stat-tile__hint">
            {{ t('common.total') }}: <span class="stat-tile__hint-strong">${{ formatCost(stats?.total_actual_cost || 0) }}</span>
          </p>
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
          <p class="stat-tile__label">{{ t('dashboard.todayTokens') }}</p>
          <p class="stat-tile__value">{{ formatTokens(stats?.today_tokens || 0) }}</p>
          <p class="stat-tile__hint">{{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }}</p>
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
          <p class="stat-tile__label">{{ t('dashboard.totalTokens') }}</p>
          <p class="stat-tile__value">{{ formatTokens(stats?.total_tokens || 0) }}</p>
          <p class="stat-tile__hint">{{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}</p>
        </div>
      </div>

      <!-- Performance (RPM) -->
      <div class="stat-tile">
        <span class="stat-tile__icon stat-tile__icon--violet">
          <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
          </svg>
        </span>
        <div class="stat-tile__body">
          <p class="stat-tile__label">{{ t('dashboard.performance') }}</p>
          <p class="stat-tile__value">
            {{ formatTokens(stats?.rpm || 0) }}<span class="stat-tile__value-unit"> RPM</span>
          </p>
          <p class="stat-tile__hint">
            <span class="stat-tile__hint-strong">{{ formatTokens(stats?.tpm || 0) }}</span> TPM
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
          <p class="stat-tile__label">{{ t('dashboard.avgResponse') }}</p>
          <p class="stat-tile__value">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
          <p class="stat-tile__hint">{{ t('dashboard.averageTime') }}</p>
        </div>
      </div>
      </div><!-- /row--four -->
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
}>()
const { t } = useI18n()

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>

<style scoped>
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
  /* Hairline divider between row 1 and row 2 */
  border-bottom: 1px solid rgba(167, 139, 250, 0.18);
}
.hero-stats__row--four {
  grid-template-columns: repeat(4, 1fr);
}

/* Each tile: vertical hairline divider on its right edge, except the last
   tile in the row. Padding gives the cell some breathing room. */
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

/* ---------- Individual stat tile (circular icon, label/value/hint) ---------- */
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
.stat-tile__icon--amber {
  background: rgb(254 243 199);
  color: rgb(217 119 6);
}
.stat-tile__icon--indigo {
  background: rgb(224 231 255);
  color: rgb(99 102 241);
}
.stat-tile__icon--violet {
  background: rgb(237 233 254);
  color: rgb(139 92 246);
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

.stat-tile__value--violet {
  color: rgb(124 58 237);
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

/* Dark mode */
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
html.dark .stat-tile__hint {
  color: rgb(148 163 184);
}
html.dark .stat-tile__value {
  color: rgb(248 250 252);
}
html.dark .stat-tile__value--violet {
  color: rgb(196 181 253);
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
html.dark .stat-tile__icon--amber {
  background: rgba(217, 119, 6, 0.22);
  color: rgb(253 186 116);
}
html.dark .stat-tile__icon--indigo {
  background: rgba(99, 102, 241, 0.22);
  color: rgb(165 180 252);
}
html.dark .stat-tile__icon--violet {
  background: rgba(139, 92, 246, 0.22);
  color: rgb(196 181 253);
}
html.dark .stat-tile__icon--rose {
  background: rgba(244, 63, 94, 0.22);
  color: rgb(252 165 165);
}
</style>
