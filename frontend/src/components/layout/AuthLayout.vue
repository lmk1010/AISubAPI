<template>
  <div class="auth-shell">
    <!-- Pixel-wave decorative background, same as the marketing pages. -->
    <div class="auth-bg" aria-hidden="true">
      <img class="auth-bg__image" src="/home-hero-bg.png?v=2" alt="" aria-hidden="true" />
    </div>

    <!-- Top-right minimal actions: locale + theme -->
    <div class="auth-topbar">
      <LocaleSwitcher />
      <button
        class="auth-topbar__icon"
        @click="toggleTheme"
        :title="isDark ? '切换到浅色' : '切换到深色'"
      >
        <Icon v-if="isDark" name="sun" size="md" />
        <Icon v-else name="moon" size="md" />
      </button>
    </div>

    <!-- Centered content -->
    <div class="auth-frame">
      <!-- Brand: logo + name horizontal, matches the marketing design -->
      <router-link to="/home" class="auth-brand">
        <span class="auth-brand__mark">
          <img
            v-if="siteLogo"
            :src="siteLogo"
            alt="Logo"
            class="h-12 w-12 rounded-lg object-contain"
          />
          <svg v-else viewBox="0 0 32 32" class="h-12 w-12">
            <defs>
              <linearGradient id="authBrandGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stop-color="#3b82f6" />
                <stop offset="100%" stop-color="#6366f1" />
              </linearGradient>
            </defs>
            <path d="M16 3 L28 10 L28 22 L16 29 L4 22 L4 10 Z" fill="url(#authBrandGrad)" />
            <path d="M16 9 L22 12.5 L22 19.5 L16 23 L10 19.5 L10 12.5 Z" fill="rgba(255,255,255,0.85)" />
          </svg>
        </span>
        <span class="auth-brand__name">{{ siteName }}</span>
      </router-link>

      <!-- Slot: form content -->
      <div class="auth-content">
        <slot />
      </div>

      <!-- Slot: footer (e.g. "still no account? sign up") -->
      <div v-if="$slots.footer" class="auth-footer-slot">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="auth-copyright">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const currentYear = computed(() => new Date().getFullYear())

const isDark = ref(document.documentElement.classList.contains('dark'))

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  appStore.fetchPublicSettings()
  isDark.value = document.documentElement.classList.contains('dark')
})
</script>

<style scoped>
.auth-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #eef2ff;
  overflow: hidden;
  padding: 24px 16px;
}

.auth-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
  -webkit-mask-image: linear-gradient(
    180deg,
    rgba(0, 0, 0, 0) 0%,
    rgba(0, 0, 0, 0) 8%,
    rgba(0, 0, 0, 1) 28%,
    rgba(0, 0, 0, 1) 92%,
    rgba(0, 0, 0, 0) 100%
  );
  mask-image: linear-gradient(
    180deg,
    rgba(0, 0, 0, 0) 0%,
    rgba(0, 0, 0, 0) 8%,
    rgba(0, 0, 0, 1) 28%,
    rgba(0, 0, 0, 1) 92%,
    rgba(0, 0, 0, 0) 100%
  );
}

.auth-bg__image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  user-select: none;
  filter: saturate(1.2) contrast(1.05);
}

.auth-topbar {
  position: absolute;
  top: 20px;
  right: 24px;
  z-index: 20;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.auth-topbar__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  color: #64748b;
  background: transparent;
  transition: all 0.18s ease;
}
.auth-topbar__icon:hover {
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

.auth-frame {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 440px;
  display: flex;
  flex-direction: column;
  align-items: stretch;
}

.auth-brand {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  text-decoration: none;
  color: inherit;
  margin-bottom: 28px;
}

.auth-brand__mark img,
.auth-brand__mark svg {
  width: 48px;
  height: 48px;
}

.auth-brand__name {
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #0f172a;
}

.auth-content {
  width: 100%;
}

.auth-footer-slot {
  margin-top: 22px;
  text-align: center;
  font-size: 14px;
  color: #64748b;
}

.auth-copyright {
  margin-top: 56px;
  text-align: center;
  font-size: 12px;
  color: #94a3b8;
}
</style>

<style>
html.dark .auth-shell {
  background: #0a0f1a;
}
html.dark .auth-bg__image {
  opacity: 0.5;
  filter: hue-rotate(15deg) saturate(1.2);
  mix-blend-mode: screen;
}
html.dark .auth-brand__name {
  color: #f8fafc;
}
html.dark .auth-topbar__icon {
  color: #94a3b8;
}
html.dark .auth-topbar__icon:hover {
  background: rgba(248, 250, 252, 0.08);
  color: #f8fafc;
}
html.dark .auth-footer-slot {
  color: #94a3b8;
}
html.dark .auth-copyright {
  color: #64748b;
}
</style>
