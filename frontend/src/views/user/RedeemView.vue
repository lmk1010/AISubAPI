<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Stat Cards Row -->
      <div class="grid grid-cols-3 gap-4">
        <!-- Balance -->
        <div class="card flex items-center gap-3 p-4">
          <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30">
            <Icon name="creditCard" size="md" class="text-primary-600 dark:text-primary-400" />
          </div>
          <div class="min-w-0">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('redeem.currentBalance') }}</p>
            <p class="text-lg font-bold text-gray-900 dark:text-white">${{ user?.balance?.toFixed(2) || '0.00' }}</p>
          </div>
        </div>
        <!-- Concurrency -->
        <div class="card flex items-center gap-3 p-4">
          <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-blue-100 dark:bg-blue-900/30">
            <Icon name="bolt" size="md" class="text-blue-600 dark:text-blue-400" />
          </div>
          <div class="min-w-0">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('redeem.concurrency') }}</p>
            <p class="text-lg font-bold text-gray-900 dark:text-white">{{ user?.concurrency || 0 }} <span class="text-sm font-normal text-gray-400 dark:text-gray-500">{{ t('redeem.requests') }}</span></p>
          </div>
        </div>
        <!-- Recent Redeem -->
        <div class="card flex items-center gap-3 p-4">
          <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30">
            <Icon name="clock" size="md" class="text-emerald-600 dark:text-emerald-400" />
          </div>
          <div class="min-w-0">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('redeem.recentActivity') }}</p>
            <p v-if="history.length > 0" class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ formatHistoryValue(history[0]) }}</p>
            <p v-else class="text-sm text-gray-400 dark:text-gray-500">—</p>
          </div>
        </div>
      </div>

      <!-- Two-column: Form + Info -->
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-5">
        <!-- Redeem Form (left, wider) -->
        <div class="card p-6 lg:col-span-3">
          <h3 class="mb-4 text-base font-bold text-gray-900 dark:text-white">{{ t('redeem.redeemCodeLabel') }}</h3>
          <form @submit.prevent="handleRedeem" class="space-y-4">
            <div class="relative">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                <Icon name="gift" size="md" class="text-gray-400 dark:text-dark-500" />
              </div>
              <input
                id="code"
                v-model="redeemCode"
                type="text"
                required
                :placeholder="t('redeem.redeemCodePlaceholder')"
                :disabled="submitting"
                class="input py-3.5 pl-12 text-base"
              />
            </div>
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('redeem.redeemCodeHint') }}</p>

            <button
              type="submit"
              :disabled="!redeemCode || submitting"
              class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-violet-500 to-primary-500 py-3.5 text-sm font-bold text-white shadow-lg shadow-primary-500/25 transition-all hover:shadow-xl hover:shadow-primary-500/30 active:scale-[0.98] disabled:opacity-50 disabled:shadow-none"
            >
              <svg
                v-if="submitting"
                class="h-5 w-5 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <Icon v-else name="checkCircle" size="md" />
              {{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}
            </button>
          </form>
        </div>

        <!-- Information Card (right, narrower) -->
        <div class="card p-6 lg:col-span-2">
          <h3 class="mb-4 text-base font-bold text-gray-900 dark:text-white">{{ t('redeem.aboutCodes') }}</h3>
          <div class="space-y-4">
            <div class="flex items-start gap-3">
              <div class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 dark:bg-primary-900/30">
                <Icon name="checkCircle" size="sm" class="text-primary-500 dark:text-primary-400" />
              </div>
              <span class="text-sm text-gray-600 dark:text-gray-400">{{ t('redeem.codeRule1') }}</span>
            </div>
            <div class="flex items-start gap-3">
              <div class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 dark:bg-primary-900/30">
                <Icon name="gift" size="sm" class="text-primary-500 dark:text-primary-400" />
              </div>
              <span class="text-sm text-gray-600 dark:text-gray-400">{{ t('redeem.codeRule2') }}</span>
            </div>
            <div class="flex items-start gap-3">
              <div class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 dark:bg-primary-900/30">
                <Icon name="infoCircle" size="sm" class="text-primary-500 dark:text-primary-400" />
              </div>
              <span class="text-sm text-gray-600 dark:text-gray-400">
                {{ t('redeem.codeRule3') }}
                <span
                  v-if="contactInfo"
                  class="ml-1 inline-flex items-center rounded-md bg-primary-100 px-1.5 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                >
                  {{ contactInfo }}
                </span>
              </span>
            </div>
            <div class="flex items-start gap-3">
              <div class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 dark:bg-primary-900/30">
                <Icon name="bolt" size="sm" class="text-primary-500 dark:text-primary-400" />
              </div>
              <span class="text-sm text-gray-600 dark:text-gray-400">{{ t('redeem.codeRule4') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Success Message -->
      <transition name="fade">
        <div
          v-if="redeemResult"
          class="card border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20"
        >
          <div class="p-5">
            <div class="flex items-start gap-3">
              <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-emerald-100 dark:bg-emerald-900/30">
                <Icon name="checkCircle" size="md" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-300">
                  {{ t('redeem.redeemSuccess') }}
                </h3>
                <div class="mt-1.5 text-sm text-emerald-700 dark:text-emerald-400">
                  <p>{{ redeemResult.message }}</p>
                  <div class="mt-2 space-y-0.5">
                    <p v-if="redeemResult.type === 'balance'" class="font-medium">
                      {{ t('redeem.added') }}: ${{ redeemResult.value.toFixed(2) }}
                    </p>
                    <p v-else-if="redeemResult.type === 'concurrency'" class="font-medium">
                      {{ t('redeem.added') }}: {{ redeemResult.value }}
                      {{ t('redeem.concurrentRequests') }}
                    </p>
                    <p v-else-if="redeemResult.type === 'subscription'" class="font-medium">
                      {{ t('redeem.subscriptionAssigned') }}
                      <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
                      <span v-if="redeemResult.validity_days">
                        ({{ t('redeem.subscriptionDays', { days: redeemResult.validity_days }) }})</span>
                    </p>
                    <p v-if="redeemResult.new_balance !== undefined">
                      {{ t('redeem.newBalance') }}:
                      <span class="font-semibold">${{ redeemResult.new_balance.toFixed(2) }}</span>
                    </p>
                    <p v-if="redeemResult.new_concurrency !== undefined">
                      {{ t('redeem.newConcurrency') }}:
                      <span class="font-semibold">{{ redeemResult.new_concurrency }} {{ t('redeem.requests') }}</span>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Error Message -->
      <transition name="fade">
        <div
          v-if="errorMessage"
          class="card border-red-200 bg-red-50 dark:border-red-800/50 dark:bg-red-900/20"
        >
          <div class="p-5">
            <div class="flex items-start gap-3">
              <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-red-100 dark:bg-red-900/30">
                <Icon name="exclamationCircle" size="md" class="text-red-600 dark:text-red-400" />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-red-800 dark:text-red-300">
                  {{ t('redeem.redeemFailed') }}
                </h3>
                <p class="mt-1 text-sm text-red-700 dark:text-red-400">
                  {{ errorMessage }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Recent Activity -->
      <div v-if="history.length > 0 || loadingHistory" class="card">
        <div class="border-b border-gray-100 px-5 py-3.5 dark:border-dark-700">
          <h2 class="text-sm font-bold text-gray-900 dark:text-white">
            {{ t('redeem.recentActivity') }}
          </h2>
        </div>
        <div class="p-5">
          <!-- Loading State -->
          <div v-if="loadingHistory" class="flex items-center justify-center py-6">
            <svg class="h-5 w-5 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <!-- History List -->
          <div v-else class="space-y-2.5">
            <div
              v-for="item in history"
              :key="item.id"
              class="flex items-center justify-between rounded-xl bg-gray-50 px-4 py-3 dark:bg-dark-800"
            >
              <div class="flex items-center gap-3">
                <div
                  :class="[
                    'flex h-9 w-9 items-center justify-center rounded-lg',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'bg-emerald-100 dark:bg-emerald-900/30'
                        : 'bg-red-100 dark:bg-red-900/30'
                      : isSubscriptionType(item.type)
                        ? 'bg-purple-100 dark:bg-purple-900/30'
                        : item.value >= 0
                          ? 'bg-blue-100 dark:bg-blue-900/30'
                          : 'bg-orange-100 dark:bg-orange-900/30'
                  ]"
                >
                  <Icon
                    v-if="isBalanceType(item.type)"
                    name="dollar"
                    size="sm"
                    :class="item.value >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'"
                  />
                  <Icon
                    v-else-if="isSubscriptionType(item.type)"
                    name="badge"
                    size="sm"
                    class="text-purple-600 dark:text-purple-400"
                  />
                  <Icon
                    v-else
                    name="bolt"
                    size="sm"
                    :class="item.value >= 0 ? 'text-blue-600 dark:text-blue-400' : 'text-orange-600 dark:text-orange-400'"
                  />
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ getHistoryItemTitle(item) }}
                  </p>
                  <p class="text-xs text-gray-400 dark:text-dark-500">
                    {{ formatDateTime(item.used_at) }}
                  </p>
                </div>
              </div>
              <div class="text-right">
                <p
                  :class="[
                    'text-sm font-semibold',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                      : isSubscriptionType(item.type)
                        ? 'text-purple-600 dark:text-purple-400'
                        : item.value >= 0
                          ? 'text-blue-600 dark:text-blue-400'
                          : 'text-orange-600 dark:text-orange-400'
                  ]"
                >
                  {{ formatHistoryValue(item) }}
                </p>
                <p
                  v-if="!isAdminAdjustment(item.type)"
                  class="font-mono text-xs text-gray-400 dark:text-dark-500"
                >
                  {{ item.code.slice(0, 8) }}...
                </p>
                <p v-else class="text-xs text-gray-400 dark:text-dark-500">
                  {{ t('redeem.adminAdjustment') }}
                </p>
                <p
                  v-if="item.notes"
                  class="mt-0.5 text-xs text-gray-500 dark:text-dark-400 italic max-w-[200px] truncate"
                  :title="item.notes"
                >
                  {{ item.notes }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { redeemAPI, authAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const user = computed(() => authStore.user)

const redeemCode = ref('')
const submitting = ref(false)
const redeemResult = ref<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
} | null>(null)
const errorMessage = ref('')

// History data
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const contactInfo = ref('')

// Helper functions for history display
const isBalanceType = (type: string) => {
  return type === 'balance' || type === 'admin_balance'
}

const isSubscriptionType = (type: string) => {
  return type === 'subscription'
}

const isAdminAdjustment = (type: string) => {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

const getHistoryItemTitle = (item: RedeemHistoryItem) => {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  } else if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  } else if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  } else if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
  } else if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }
  return t('common.unknown')
}

const formatHistoryValue = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  } else if (isSubscriptionType(item.type)) {
    // 订阅类型显示有效天数和分组名称
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${t('redeem.requests')}`
  }
}

const fetchHistory = async () => {
  loadingHistory.value = true
  try {
    history.value = await redeemAPI.getHistory()
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const handleRedeem = async () => {
  if (!redeemCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(redeemCode.value.trim())

    redeemResult.value = result

    // Refresh user data to get updated balance/concurrency
    await authStore.refreshUser()

    // If subscription type, immediately refresh subscription status
    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true) // force refresh
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    // Clear the input
    redeemCode.value = ''

    // Refresh history
    await fetchHistory()

    // Show success toast
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: any) {
    errorMessage.value = error.response?.data?.detail || t('redeem.failedToRedeem')

    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  fetchHistory()
  try {
    const settings = await authAPI.getPublicSettings()
    contactInfo.value = settings.contact_info || ''
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
