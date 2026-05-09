<template>
  <div class="flex flex-col rounded-xl border border-gray-200 bg-white transition-all hover:shadow-md dark:border-dark-600 dark:bg-dark-800">
    <div class="flex flex-1 flex-col px-4 pt-4 pb-3">
      <!-- Platform badge -->
      <span class="inline-block w-fit rounded px-2 py-0.5 text-[11px] font-medium bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400 mb-2">
        {{ pLabel }}
      </span>

      <!-- Name + Description -->
      <h3 class="text-base font-bold text-gray-900 dark:text-white leading-tight">{{ plan.name }}</h3>
      <p v-if="plan.description" class="mt-0.5 text-[11px] text-gray-500 dark:text-gray-400 line-clamp-1">
        {{ plan.description }}
      </p>

      <!-- Price -->
      <div class="mt-3 mb-2">
        <div class="flex items-baseline gap-1">
          <span class="text-sm text-primary-500 dark:text-primary-400">¥</span>
          <span class="text-3xl font-extrabold tracking-tight text-primary-500 dark:text-primary-400">{{ plan.price }}</span>
          <span class="text-xs text-gray-400 dark:text-gray-500">/ {{ validitySuffix }}</span>
        </div>
        <div v-if="plan.original_price" class="mt-0.5 flex items-center gap-2">
          <span class="text-xs text-gray-400 line-through dark:text-gray-500">¥{{ plan.original_price }}</span>
          <span class="rounded px-1.5 py-0.5 text-[10px] font-semibold bg-red-50 text-red-500 dark:bg-red-900/20 dark:text-red-400">{{ discountText }}</span>
        </div>
      </div>

      <!-- Quota rows -->
      <div class="mb-2.5 space-y-1 border-t border-gray-100 pt-2 dark:border-dark-700">
        <div class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.rate') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ rateDisplay }}</span>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.daily_limit_usd }}</span>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.weekly_limit_usd }}</span>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.monthly_limit_usd }}</span>
        </div>
      </div>

      <!-- Features list -->
      <div v-if="plan.features.length > 0 || modelScopeLabels.length > 0" class="space-y-1">
        <div v-for="feature in plan.features" :key="feature" class="flex items-center gap-1.5">
          <svg class="h-3 w-3 flex-shrink-0 text-primary-500 dark:text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
        <div v-if="modelScopeLabels.length > 0" class="flex items-center gap-1.5">
          <svg class="h-3 w-3 flex-shrink-0 text-primary-500 dark:text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ modelScopeLabels.join(', ') }}</span>
        </div>
      </div>
    </div>

    <!-- Subscribe Button -->
    <div class="px-4 pb-4 pt-2">
      <button
        type="button"
        class="w-full rounded-lg py-2 text-sm font-semibold transition-all active:scale-[0.98] bg-primary-500 text-white hover:bg-primary-600 dark:bg-primary-600 dark:hover:bg-primary-500"
        @click="emit('select', plan)"
      >
        {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import { platformLabel } from '@/utils/platformColors'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

const pLabel = computed(() => platformLabel(platform.value))

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const validitySuffix = computed(() => {
  const u = props.plan.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${props.plan.validity_days}${t('payment.days')}`
})
</script>
