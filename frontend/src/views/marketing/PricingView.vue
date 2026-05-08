<template>
  <div class="pricing-shell">
    <MarketingHeader />

    <main class="pricing-main">
      <header class="page-hero">
        <span class="page-badge">{{ t('marketing.pricing.badge') }}</span>
        <h1 class="page-title">{{ t('marketing.pricing.title') }}</h1>
        <p class="page-desc">{{ t('marketing.pricing.desc') }}</p>
      </header>

      <div class="toolbar">
        <div class="search-wrap">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="7" />
            <path stroke-linecap="round" d="M21 21l-4.3-4.3" />
          </svg>
          <input
            v-model="search"
            type="text"
            class="search-input"
            :placeholder="t('marketing.pricing.searchPlaceholder')"
          />
        </div>
        <router-link :to="ctaTarget" class="hero-cta">
          {{ isAuthenticated ? t('marketing.pricing.ctaAuthed') : t('marketing.pricing.cta') }}
        </router-link>
      </div>

      <section v-if="loading" class="state state--loading">
        <div class="spinner"></div>
        <p>{{ t('marketing.pricing.loading') }}</p>
      </section>

      <section v-else-if="error" class="state state--error">
        <p>{{ error }}</p>
        <button class="reload-btn" @click="loadPricing">{{ t('marketing.pricing.retry') }}</button>
      </section>

      <section v-else-if="filteredRows.length === 0" class="state state--empty">
        <p>{{ t('marketing.pricing.empty') }}</p>
      </section>

      <section v-else class="pricing-card">
        <div class="table-scroll">
          <table class="pricing-table">
            <thead>
              <tr>
                <th class="col-model">{{ t('marketing.pricing.columns.model') }}</th>
                <th class="col-platform">{{ t('marketing.pricing.columns.platform') }}</th>
                <th class="col-mode">{{ t('marketing.pricing.columns.billingMode') }}</th>
                <th class="col-num">{{ t('marketing.pricing.columns.input') }}</th>
                <th class="col-num">{{ t('marketing.pricing.columns.cacheWrite') }}</th>
                <th class="col-num">{{ t('marketing.pricing.columns.cacheRead') }}</th>
                <th class="col-num">{{ t('marketing.pricing.columns.output') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, i) in filteredRows" :key="i" :class="`platform-${platformKey(row.platform)}`">
                <td class="col-model">
                  <span class="model-name">{{ row.model }}</span>
                  <span v-if="row.channel" class="channel-tag">{{ row.channel }}</span>
                </td>
                <td class="col-platform">
                  <span class="platform-pill">{{ row.platform || '—' }}</span>
                </td>
                <td class="col-mode">
                  <span class="mode-tag">{{ billingLabel(row.billingMode) }}</span>
                </td>
                <td class="col-num">{{ priceCell(row, 'input') }}</td>
                <td class="col-num">{{ priceCell(row, 'cacheWrite') }}</td>
                <td class="col-num">{{ priceCell(row, 'cacheRead') }}</td>
                <td class="col-num">{{ priceCell(row, 'output') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="table-note">
          <span>{{ t('marketing.pricing.unitPerMillion') }}</span>
          <span class="dot">·</span>
          <span>{{ t('marketing.pricing.unitPerRequest') }}</span>
        </p>
      </section>
    </main>

    <MarketingFooter />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores'
import { userChannelsAPI } from '@/api/channels'
import type { UserAvailableChannel, UserSupportedModelPricing } from '@/api/channels'
import MarketingHeader from '@/components/marketing/MarketingHeader.vue'
import MarketingFooter from '@/components/marketing/MarketingFooter.vue'

interface PricingRow {
  model: string
  platform: string
  channel: string
  billingMode: string
  pricing: UserSupportedModelPricing | null
}

const { t } = useI18n()
const authStore = useAuthStore()

const channels = ref<UserAvailableChannel[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const search = ref('')

const isAuthenticated = computed(() => authStore.isAuthenticated)
const ctaTarget = computed(() => (isAuthenticated.value ? '/dashboard' : '/register'))

const rows = computed<PricingRow[]>(() => {
  const seen = new Set<string>()
  const out: PricingRow[] = []
  for (const ch of channels.value) {
    for (const section of ch.platforms || []) {
      for (const model of section.supported_models || []) {
        const key = `${section.platform}::${model.name}`
        if (seen.has(key)) continue
        seen.add(key)
        out.push({
          model: model.name,
          platform: section.platform,
          channel: ch.name,
          billingMode: model.pricing?.billing_mode ?? '',
          pricing: model.pricing ?? null
        })
      }
    }
  }
  return out.sort((a, b) => {
    const p = (a.platform || '').localeCompare(b.platform || '')
    return p !== 0 ? p : a.model.localeCompare(b.model)
  })
})

const filteredRows = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter(
    (r) =>
      r.model.toLowerCase().includes(q) ||
      (r.platform || '').toLowerCase().includes(q) ||
      (r.channel || '').toLowerCase().includes(q)
  )
})

async function loadPricing() {
  loading.value = true
  error.value = null
  try {
    channels.value = await userChannelsAPI.getPublic()
  } catch (e: any) {
    error.value = e?.message || t('marketing.pricing.loadFailed')
  } finally {
    loading.value = false
  }
}

function platformKey(platform: string): string {
  const p = (platform || '').toLowerCase()
  if (p.includes('anthropic') || p.includes('claude')) return 'anthropic'
  if (p.includes('openai') || p.includes('gpt')) return 'openai'
  if (p.includes('gemini') || p.includes('google')) return 'gemini'
  return 'default'
}

function billingLabel(mode: string): string {
  if (mode === 'token') return t('marketing.pricing.billingMode.token')
  if (mode === 'per_request') return t('marketing.pricing.billingMode.perRequest')
  if (mode === 'image') return t('marketing.pricing.billingMode.image')
  return '—'
}

function format(value: number | null | undefined, scale: number): string {
  if (value == null) return t('marketing.pricing.noPricing')
  const scaled = value * scale
  return `$${scaled.toPrecision(10).replace(/\.?0+$/, '')}`
}

type PriceField = 'input' | 'output' | 'cacheWrite' | 'cacheRead'

function priceCell(row: PricingRow, field: PriceField): string {
  const p = row.pricing
  if (!p) return t('marketing.pricing.noPricing')

  // Per-request / image billing: input cell carries the per-request price,
  // other cells dash out so the table stays scannable.
  if (p.billing_mode === 'per_request') {
    if (field === 'input') return format(p.per_request_price, 1) + ' / req'
    return '—'
  }
  if (p.billing_mode === 'image') {
    if (field === 'output') return format(p.image_output_price, 1) + ' / req'
    return '—'
  }

  switch (field) {
    case 'input':
      return format(p.input_price, 1_000_000)
    case 'output':
      return format(p.output_price, 1_000_000)
    case 'cacheWrite':
      return format(p.cache_write_price, 1_000_000)
    case 'cacheRead':
      return format(p.cache_read_price, 1_000_000)
  }
}

onMounted(loadPricing)
</script>

<style scoped>
.pricing-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #eef2ff;
  color: #0f172a;
}

.pricing-main {
  flex: 1;
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  padding: 40px 32px 80px;
}

.page-hero {
  text-align: center;
  margin-bottom: 32px;
}

.page-badge {
  display: inline-flex;
  align-items: center;
  padding: 6px 16px;
  border-radius: 999px;
  background: rgba(219, 234, 254, 0.7);
  color: #1d4ed8;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid rgba(147, 197, 253, 0.5);
  margin-bottom: 18px;
}

.page-title {
  font-size: clamp(32px, 4vw, 44px);
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.02em;
  margin: 0 0 14px;
  color: #0f172a;
}

.page-desc {
  font-size: 15px;
  line-height: 1.7;
  color: #475569;
  max-width: 720px;
  margin: 0 auto;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin: 0 0 22px;
}

.search-wrap {
  position: relative;
  flex: 1;
  min-width: 240px;
  max-width: 480px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #94a3b8;
}

.search-input {
  width: 100%;
  padding: 9px 14px 9px 36px;
  border-radius: 10px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(6px);
  font-size: 13px;
  color: #0f172a;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.search-input:focus {
  outline: none;
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.15);
}

.hero-cta {
  display: inline-flex;
  align-items: center;
  padding: 9px 18px;
  border-radius: 10px;
  background: #2563eb;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
  transition: background 0.18s ease, transform 0.18s ease;
}
.hero-cta:hover {
  background: #1d4ed8;
  transform: translateY(-1px);
}

.state {
  text-align: center;
  padding: 80px 24px;
  color: #64748b;
}

.state .reload-btn {
  margin-top: 16px;
  padding: 8px 20px;
  border-radius: 8px;
  background: #2563eb;
  color: #fff;
  font-weight: 500;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.18s ease;
}
.state .reload-btn:hover {
  background: #1d4ed8;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(37, 99, 235, 0.2);
  border-top-color: #2563eb;
  border-radius: 50%;
  margin: 0 auto 16px;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.pricing-card {
  background: rgba(255, 255, 255, 0.78);
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 14px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.08),
    0 1px 2px rgba(15, 23, 42, 0.04);
  overflow: hidden;
}

.table-scroll {
  overflow-x: auto;
}

.pricing-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.pricing-table thead th {
  position: sticky;
  top: 0;
  background: rgba(245, 243, 255, 0.92);
  backdrop-filter: blur(8px);
  text-align: left;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #64748b;
  padding: 12px 14px;
  border-bottom: 1px solid rgba(203, 213, 225, 0.6);
}

.pricing-table tbody td {
  padding: 11px 14px;
  border-bottom: 1px solid rgba(226, 232, 240, 0.6);
  color: #1f2937;
  vertical-align: middle;
}

.pricing-table tbody tr:last-child td {
  border-bottom: none;
}

.pricing-table tbody tr {
  transition: background 0.15s ease;
}
.pricing-table tbody tr:hover {
  background: rgba(245, 243, 255, 0.45);
}

.col-model { min-width: 220px; }
.col-platform { width: 130px; }
.col-mode { width: 90px; }
.col-num {
  text-align: right;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  white-space: nowrap;
}

.model-name {
  display: block;
  font-weight: 600;
  color: #0f172a;
}

.channel-tag {
  display: inline-block;
  margin-top: 2px;
  font-size: 11px;
  color: #94a3b8;
}

.platform-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 9px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: rgba(241, 245, 249, 0.9);
  color: #475569;
  border: 1px solid rgba(226, 232, 240, 0.8);
}

.platform-anthropic .platform-pill {
  background: rgba(255, 237, 213, 0.85);
  color: #b45309;
  border-color: rgba(253, 186, 116, 0.55);
}
.platform-openai .platform-pill {
  background: rgba(220, 252, 231, 0.85);
  color: #047857;
  border-color: rgba(110, 231, 183, 0.55);
}
.platform-gemini .platform-pill {
  background: rgba(219, 234, 254, 0.85);
  color: #1d4ed8;
  border-color: rgba(147, 197, 253, 0.55);
}

.mode-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(226, 232, 240, 0.6);
  color: #475569;
  font-size: 11px;
  font-weight: 500;
}

.table-note {
  margin: 0;
  padding: 10px 14px;
  background: rgba(248, 250, 252, 0.6);
  font-size: 11px;
  color: #94a3b8;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.table-note .dot {
  color: #cbd5e1;
}
</style>

<style>
html.dark .pricing-shell {
  background: #0a0f1a;
  color: #e2e8f0;
}
html.dark .page-title { color: #f8fafc; }
html.dark .page-desc { color: #94a3b8; }
html.dark .page-badge {
  background: rgba(30, 64, 175, 0.25);
  color: #93c5fd;
  border-color: rgba(59, 130, 246, 0.3);
}
html.dark .search-input {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.6);
  color: #e2e8f0;
}
html.dark .pricing-card {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.5);
}
html.dark .pricing-table thead th {
  background: rgba(30, 41, 59, 0.85);
  color: #94a3b8;
  border-bottom-color: rgba(71, 85, 105, 0.6);
}
html.dark .pricing-table tbody td {
  color: #cbd5e1;
  border-bottom-color: rgba(51, 65, 85, 0.5);
}
html.dark .pricing-table tbody tr:hover {
  background: rgba(51, 65, 85, 0.35);
}
html.dark .model-name { color: #f8fafc; }
html.dark .channel-tag { color: #64748b; }
html.dark .platform-pill {
  background: rgba(51, 65, 85, 0.6);
  color: #cbd5e1;
  border-color: rgba(71, 85, 105, 0.5);
}
html.dark .mode-tag {
  background: rgba(51, 65, 85, 0.5);
  color: #cbd5e1;
}
html.dark .table-note {
  background: rgba(15, 23, 42, 0.5);
  color: #64748b;
}
</style>