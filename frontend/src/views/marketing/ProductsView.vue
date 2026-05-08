<template>
  <div class="products-shell">
    <MarketingHeader />

    <main class="products-main">
      <header class="page-hero">
        <span class="page-badge">{{ t('home.nav.products') }}</span>
        <h1 class="page-title">{{ t('marketing.products.title') }}</h1>
        <p class="page-desc">{{ t('marketing.products.desc') }}</p>
      </header>

      <section v-if="loading" class="state state--loading">
        <div class="spinner"></div>
        <p>{{ t('marketing.products.loading') }}</p>
      </section>

      <section v-else-if="error" class="state state--error">
        <p>{{ error }}</p>
        <button class="reload-btn" @click="loadPlans">{{ t('marketing.products.retry') }}</button>
      </section>

      <section v-else-if="plans.length === 0" class="state state--empty">
        <p>{{ t('marketing.products.empty') }}</p>
      </section>

      <section v-else class="plan-grid">
        <article
          v-for="plan in plans"
          :key="plan.id"
          class="plan-card"
          :class="`plan-card--${platformVariant(plan.group_platform)}`"
        >
          <div class="plan-accent" />

          <div class="plan-body">
            <div class="plan-head">
              <div class="plan-head__left">
                <div class="plan-title-row">
                  <h3 class="plan-name">{{ plan.name }}</h3>
                  <span class="plan-platform-badge">{{ platformLabel(plan.group_platform) }}</span>
                </div>
                <p v-if="plan.description" class="plan-desc">{{ plan.description }}</p>
              </div>
              <div class="plan-price">
                <div class="plan-price__row">
                  <span class="plan-price__currency">¥</span>
                  <span class="plan-price__value">{{ plan.price }}</span>
                </div>
                <span class="plan-price__period">/ {{ validityLabel(plan) }}</span>
                <div v-if="plan.original_price" class="plan-price__original">
                  <span class="strike">¥{{ plan.original_price }}</span>
                  <span class="discount">{{ discountText(plan) }}</span>
                </div>
              </div>
            </div>

            <div class="plan-stats">
              <div class="stat">
                <span class="stat__label">{{ t('marketing.products.rate') }}</span>
                <span class="stat__value">×{{ plan.rate_multiplier ?? 1 }}</span>
              </div>
              <div v-if="plan.daily_limit_usd != null" class="stat">
                <span class="stat__label">{{ t('marketing.products.dailyLimit') }}</span>
                <span class="stat__value">${{ plan.daily_limit_usd }}</span>
              </div>
              <div v-if="plan.weekly_limit_usd != null" class="stat">
                <span class="stat__label">{{ t('marketing.products.weeklyLimit') }}</span>
                <span class="stat__value">${{ plan.weekly_limit_usd }}</span>
              </div>
              <div v-if="plan.monthly_limit_usd != null" class="stat">
                <span class="stat__label">{{ t('marketing.products.monthlyLimit') }}</span>
                <span class="stat__value">${{ plan.monthly_limit_usd }}</span>
              </div>
              <div
                v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null"
                class="stat"
              >
                <span class="stat__label">{{ t('marketing.products.quota') }}</span>
                <span class="stat__value">{{ t('marketing.products.unlimited') }}</span>
              </div>
            </div>

            <ul v-if="plan.features.length > 0" class="plan-features">
              <li v-for="feature in plan.features" :key="feature">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="check">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                </svg>
                <span>{{ feature }}</span>
              </li>
            </ul>

            <div
              v-if="plan.supported_model_scopes && plan.supported_model_scopes.length > 0"
              class="plan-scopes"
            >
              <span class="plan-scopes__label">{{ t('marketing.products.models') }}</span>
              <div class="plan-scopes__list">
                <span v-for="scope in plan.supported_model_scopes" :key="scope" class="plan-scopes__item">{{ scope }}</span>
              </div>
            </div>

            <div class="plan-cta-wrap">
              <router-link :to="ctaTarget(plan)" class="plan-cta">
                {{ isAuthenticated ? t('marketing.products.subscribe') : t('marketing.products.signUpToBuy') }}
              </router-link>
            </div>
          </div>
        </article>
      </section>
    </main>

    <MarketingFooter />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import type { SubscriptionPlan } from '@/types/payment'
import MarketingHeader from '@/components/marketing/MarketingHeader.vue'
import MarketingFooter from '@/components/marketing/MarketingFooter.vue'

const { t } = useI18n()
const authStore = useAuthStore()

const plans = ref<SubscriptionPlan[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const isAuthenticated = computed(() => authStore.isAuthenticated)

async function loadPlans() {
  loading.value = true
  error.value = null
  try {
    const response = await paymentAPI.getPublicPlans()
    const list: SubscriptionPlan[] = response.data || []
    plans.value = list
      .filter((p: SubscriptionPlan) => p.for_sale)
      .sort((a: SubscriptionPlan, b: SubscriptionPlan) => a.sort_order - b.sort_order)
  } catch (e: any) {
    error.value = e?.message || t('marketing.products.loadFailed')
  } finally {
    loading.value = false
  }
}

function ctaTarget(plan: SubscriptionPlan) {
  if (isAuthenticated.value) {
    return `/purchase?plan=${plan.id}`
  }
  return `/login?redirect=${encodeURIComponent(`/purchase?plan=${plan.id}`)}`
}

function platformVariant(platform?: string): string {
  const p = (platform || '').toLowerCase()
  if (p.includes('anthropic') || p.includes('claude')) return 'anthropic'
  if (p.includes('openai') || p.includes('gpt')) return 'openai'
  if (p.includes('gemini') || p.includes('google')) return 'gemini'
  return 'default'
}

function platformLabel(platform?: string): string {
  if (!platform) return '通用'
  return platform
}

function validityLabel(plan: SubscriptionPlan): string {
  const unit = plan.validity_unit || 'days'
  if (plan.validity_days === 30 && unit === 'days') return t('marketing.products.month')
  if (plan.validity_days === 365 && unit === 'days') return t('marketing.products.year')
  if (plan.validity_days === 7 && unit === 'days') return t('marketing.products.week')
  return `${plan.validity_days} ${t(`marketing.products.unit.${unit}`)}`
}

function discountText(plan: SubscriptionPlan): string {
  if (!plan.original_price || plan.original_price <= plan.price) return ''
  const off = Math.round(((plan.original_price - plan.price) / plan.original_price) * 10)
  return `${off}${t('marketing.products.discountSuffix')}`
}

onMounted(() => {
  loadPlans()
})
</script>

<style scoped>
.products-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #eef2ff;
  color: #0f172a;
}

.products-main {
  flex: 1;
  max-width: 1280px;
  margin: 0 auto;
  width: 100%;
  padding: 40px 32px 80px;
}

.page-hero {
  text-align: center;
  margin-bottom: 56px;
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
  max-width: 640px;
  margin: 0 auto;
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

.plan-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
}

.plan-card {
  position: relative;
  display: flex;
  flex-direction: column;
  border-radius: 18px;
  border: 1px solid rgba(226, 232, 240, 0.8);
  background: #fff;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.plan-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 16px 40px -16px rgba(15, 23, 42, 0.18);
}

.plan-accent {
  height: 4px;
  background: linear-gradient(90deg, #94a3b8, #64748b);
}
.plan-card--anthropic .plan-accent {
  background: linear-gradient(90deg, #fb923c, #f97316);
}
.plan-card--openai .plan-accent {
  background: linear-gradient(90deg, #34d399, #10b981);
}
.plan-card--gemini .plan-accent {
  background: linear-gradient(90deg, #60a5fa, #3b82f6);
}

.plan-body {
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  flex: 1;
}

.plan-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.plan-head__left {
  flex: 1;
  min-width: 0;
}

.plan-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.plan-name {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
  color: #0f172a;
}

.plan-platform-badge {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(241, 245, 249, 0.9);
  color: #475569;
}

.plan-desc {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: #64748b;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.plan-price {
  text-align: right;
  flex-shrink: 0;
}

.plan-price__row {
  display: flex;
  align-items: baseline;
  justify-content: flex-end;
  gap: 2px;
}

.plan-price__currency {
  font-size: 14px;
  color: #94a3b8;
}

.plan-price__value {
  font-size: 30px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: #0f172a;
}

.plan-price__period {
  display: block;
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}

.plan-price__original {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
  margin-top: 4px;
}

.plan-price__original .strike {
  font-size: 12px;
  color: #94a3b8;
  text-decoration: line-through;
}

.plan-price__original .discount {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  background: #fef2f2;
  color: #dc2626;
}

.plan-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 14px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f8fafc;
}

.stat {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.stat__label {
  color: #94a3b8;
}

.stat__value {
  font-weight: 600;
  color: #334155;
}

.plan-features {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.plan-features li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 13px;
  color: #334155;
  line-height: 1.5;
}

.plan-features .check {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: #16a34a;
  margin-top: 1px;
}

.plan-scopes {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.plan-scopes__label {
  font-size: 12px;
  color: #94a3b8;
  margin-right: 4px;
}

.plan-scopes__item {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(226, 232, 240, 0.6);
  color: #475569;
}

.plan-cta-wrap {
  margin-top: auto;
}

.plan-cta {
  display: block;
  text-align: center;
  padding: 11px 18px;
  border-radius: 10px;
  background: #2563eb;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  transition: background 0.18s ease, transform 0.18s ease;
}

.plan-cta:hover {
  background: #1d4ed8;
  transform: translateY(-1px);
}
</style>

<style>
html.dark .products-shell {
  background: #0a0f1a;
  color: #e2e8f0;
}
html.dark .page-title {
  color: #f8fafc;
}
html.dark .page-desc {
  color: #94a3b8;
}
html.dark .page-badge {
  background: rgba(30, 64, 175, 0.25);
  color: #93c5fd;
  border-color: rgba(59, 130, 246, 0.3);
}
html.dark .plan-card {
  background: #1e293b;
  border-color: rgba(71, 85, 105, 0.5);
}
html.dark .plan-name,
html.dark .plan-price__value {
  color: #f8fafc;
}
html.dark .plan-desc,
html.dark .plan-price__period,
html.dark .stat__label,
html.dark .plan-scopes__label,
html.dark .plan-price__currency,
html.dark .plan-price__original .strike {
  color: #94a3b8;
}
html.dark .plan-stats {
  background: rgba(15, 23, 42, 0.6);
}
html.dark .stat__value,
html.dark .plan-features li {
  color: #cbd5e1;
}
html.dark .plan-platform-badge,
html.dark .plan-scopes__item {
  background: rgba(51, 65, 85, 0.6);
  color: #cbd5e1;
}
</style>
