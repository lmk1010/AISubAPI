<template>
  <div
    class="flex flex-col rounded-xl border border-gray-200 bg-white transition-all hover:shadow-lg dark:border-dark-600 dark:bg-dark-800"
  >
    <div class="flex flex-1 flex-col p-5">
      <!-- Badge -->
      <div class="mb-3">
        <span :class="['inline-block rounded-md px-2.5 py-1 text-[11px] font-semibold', badgeLightClass]">
          {{ pLabel }}
        </span>
      </div>

      <!-- Name + Description -->
      <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ plan.name }}</h3>
      <p v-if="plan.description" class="mt-0.5 text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
        {{ plan.description }}
      </p>

      <!-- Price -->
      <div class="mt-4 mb-1">
        <div class="flex items-baseline gap-1">
          <span :class="['text-sm', textClass]">¥</span>
          <span :class="['text-3xl font-extrabold tracking-tight', textClass]">{{ plan.price }}</span>
          <span class="text-sm text-gray-400 dark:text-gray-500">/ {{ validitySuffix }}</span>
        </div>
        <div v-if="plan.original_price" class="mt-1 flex items-center gap-2">
          <span class="text-sm text-gray-400 line-through dark:text-gray-500">¥{{ plan.original_price }}</span>
          <span :class="['rounded-full px-2 py-0.5 text-[10px] font-bold', discountClass]">{{ discountText }}</span>
        </div>
      </div>

      <!-- Quota rows -->
      <div class="mt-3 mb-3 space-y-1">
        <div class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.rate') }}</span>
          <span class="font-semibold text-gray-700 dark:text-gray-200">{{ rateDisplay }}</span>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="font-semibold text-gray-700 dark:text-gray-200">${{ plan.daily_limit_usd }}</span>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="font-semibold text-gray-700 dark:text-gray-200">${{ plan.weekly_limit_usd }}</span>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="flex items-center justify-between text-xs">
          <span class="text-gray-400 dark:text-gray-500">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="font-semibold text-gray-700 dark:text-gray-200">${{ plan.monthly_limit_usd }}</span>
        </div>
      </div>

      <!-- Features list -->
      <div v-if="plan.features.length > 0 || modelScopeLabels.length > 0" class="space-y-1.5">
        <div v-for="feature in plan.features" :key="feature" class="flex items-center gap-2">
          <svg :class="['h-3.5 w-3.5 flex-shrink-0', iconClass]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
        <div v-if="modelScopeLabels.length > 0" class="flex items-center gap-2">
          <svg :class="['h-3.5 w-3.5 flex-shrink-0', iconClass]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ modelScopeLabels.join(', ') }}</span>
        </div>
      </div>
    </div>

    <!-- Subscribe Button -->
    <div class="px-5 pb-5">
      <button
        type="button"
        :class="['w-full rounded-xl py-2.5 text-sm font-bold tracking-wide transition-all active:scale-[0.98]', btnClass]"
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
import {
  platformBadgeLightClass,
  platformTextClass,
  platformIconClass,
  platformButtonClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

// Derived color classes from central config
const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const btnClass = computed(() => platformButtonClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
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
