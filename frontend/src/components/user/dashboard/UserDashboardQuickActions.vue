<template>
  <section class="quick-actions">
    <h2 class="quick-actions__title">{{ t('dashboard.quickActions') }}</h2>
    <ul class="quick-actions__list">
      <li>
        <button @click="router.push('/keys')" class="quick-action group">
          <span class="quick-action__icon quick-action__icon--primary">
            <Icon name="key" size="lg" />
          </span>
          <span class="quick-action__text">
            <span class="quick-action__label">{{ t('dashboard.createApiKey') }}</span>
            <span class="quick-action__hint">{{ t('dashboard.generateNewKey') }}</span>
          </span>
          <Icon name="chevronRight" size="md" class="quick-action__chevron" />
        </button>
      </li>
      <li>
        <button @click="router.push('/usage')" class="quick-action group">
          <span class="quick-action__icon quick-action__icon--emerald">
            <Icon name="chart" size="lg" />
          </span>
          <span class="quick-action__text">
            <span class="quick-action__label">{{ t('dashboard.viewUsage') }}</span>
            <span class="quick-action__hint">{{ t('dashboard.checkDetailedLogs') }}</span>
          </span>
          <Icon name="chevronRight" size="md" class="quick-action__chevron" />
        </button>
      </li>
      <li>
        <button @click="router.push('/redeem')" class="quick-action group">
          <span class="quick-action__icon quick-action__icon--amber">
            <Icon name="gift" size="lg" />
          </span>
          <span class="quick-action__text">
            <span class="quick-action__label">{{ t('dashboard.redeemCode') }}</span>
            <span class="quick-action__hint">{{ t('dashboard.addBalanceWithCode') }}</span>
          </span>
          <Icon name="chevronRight" size="md" class="quick-action__chevron" />
        </button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
const router = useRouter()
const { t } = useI18n()
</script>

<style scoped>
.quick-actions {
  padding: 4px 2px;
}

.quick-actions__title {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: rgb(100 116 139);
  padding: 0 4px 12px;
}

.quick-actions__list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  list-style: none;
  padding: 0;
  margin: 0;
}

.quick-action {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 14px;
  padding: 12px 12px;
  border-radius: 14px;
  text-align: left;
  background: transparent;
  transition:
    background 0.18s ease,
    transform 0.18s ease;
}

.quick-action:hover {
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(6px);
}

.quick-action__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  flex-shrink: 0;
  transition: transform 0.18s ease;
}
.quick-action:hover .quick-action__icon {
  transform: scale(1.05);
}

.quick-action__icon--primary {
  background: rgb(204 251 241);
  color: rgb(13 148 136);
}
.quick-action__icon--emerald {
  background: rgb(209 250 229);
  color: rgb(5 150 105);
}
.quick-action__icon--amber {
  background: rgb(254 243 199);
  color: rgb(217 119 6);
}

.quick-action__text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.quick-action__label {
  font-size: 14px;
  font-weight: 500;
  color: rgb(15 23 42);
}

.quick-action__hint {
  font-size: 12px;
  color: rgb(100 116 139);
}

.quick-action__chevron {
  color: rgb(148 163 184);
  transition: color 0.18s ease, transform 0.18s ease;
}
.quick-action:hover .quick-action__chevron {
  color: rgb(71 85 105);
  transform: translateX(2px);
}

html.dark .quick-actions__title {
  color: rgb(148 163 184);
}
html.dark .quick-action:hover {
  background: rgba(15, 23, 42, 0.45);
}
html.dark .quick-action__label {
  color: rgb(248 250 252);
}
html.dark .quick-action__hint {
  color: rgb(148 163 184);
}
html.dark .quick-action__icon--primary {
  background: rgba(13, 148, 136, 0.18);
  color: rgb(94 234 212);
}
html.dark .quick-action__icon--emerald {
  background: rgba(5, 150, 105, 0.18);
  color: rgb(110 231 183);
}
html.dark .quick-action__icon--amber {
  background: rgba(217, 119, 6, 0.18);
  color: rgb(252 211 77);
}
html.dark .quick-action__chevron {
  color: rgb(100 116 139);
}
</style>
