<template>
  <header class="marketing-header">
    <nav class="marketing-nav">
      <!-- Brand -->
      <router-link to="/home" class="brand">
        <span class="brand-mark">
          <img
            v-if="siteLogo"
            :src="siteLogo"
            alt="Logo"
            class="h-[46px] w-[46px] rounded-lg object-contain"
          />
          <svg v-else viewBox="0 0 32 32" class="h-[46px] w-[46px]">
            <defs>
              <linearGradient id="brandGradHeader" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stop-color="#3b82f6" />
                <stop offset="100%" stop-color="#6366f1" />
              </linearGradient>
            </defs>
            <path
              d="M16 3 L28 10 L28 22 L16 29 L4 22 L4 10 Z"
              fill="url(#brandGradHeader)"
            />
            <path
              d="M16 9 L22 12.5 L22 19.5 L16 23 L10 19.5 L10 12.5 Z"
              fill="rgba(255,255,255,0.85)"
            />
          </svg>
        </span>
        <span class="brand-name">{{ siteName }}</span>
      </router-link>

      <!-- Center nav -->
      <ul class="center-nav">
        <li><router-link to="/products" :class="{ active: $route.path === '/products' }">{{ t('home.nav.products') }}</router-link></li>
        <li><router-link to="/pricing" :class="{ active: $route.path === '/pricing' }">{{ t('home.nav.pricing') }}</router-link></li>
        <li><router-link to="/blog" :class="{ active: $route.path.startsWith('/blog') }">{{ t('home.nav.blog') }}</router-link></li>
        <li>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.nav.docs') }}</a>
          <router-link v-else to="/docs" :class="{ active: $route.path.startsWith('/docs') }">{{ t('home.nav.docs') }}</router-link>
        </li>
        <li><a :href="githubUrl" target="_blank" rel="noopener noreferrer">{{ t('home.nav.community') }}</a></li>
      </ul>

      <!-- Right actions -->
      <div class="header-actions">
        <LocaleSwitcher />
        <button
          class="icon-btn"
          @click="toggleTheme"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>

        <router-link v-if="isAuthenticated" :to="dashboardPath" class="ghost-link">
          {{ t('home.dashboard') }}
        </router-link>
        <router-link v-else to="/login" class="ghost-link">
          {{ t('home.login') }}
        </router-link>

        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="primary-pill"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
        </router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')

const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

const isDark = ref(document.documentElement.classList.contains('dark'))

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  isDark.value = document.documentElement.classList.contains('dark')
})
</script>

<style scoped>
.marketing-header {
  position: relative;
  z-index: 30;
  padding: 0 32px;
  background: transparent;
}

.marketing-nav {
  max-width: 1280px;
  margin: 0 auto;
  height: 88px;
  display: flex;
  align-items: center;
  gap: 36px;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  color: inherit;
  flex-shrink: 0;
}

.brand-mark img,
.brand-mark svg {
  width: 46px;
  height: 46px;
}

.brand-name {
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.018em;
}

.center-nav {
  display: none;
  align-items: center;
  gap: 44px;
  margin: 0 auto;
  list-style: none;
  padding: 0;
}

@media (min-width: 1024px) {
  .center-nav {
    display: flex;
  }
}

.center-nav a {
  font-size: 15px;
  font-weight: 500;
  color: #475569;
  text-decoration: none;
  transition: color 0.18s ease;
}

.center-nav a:hover,
.center-nav a.active {
  color: #2563eb;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  color: #64748b;
  background: transparent;
  transition: all 0.18s ease;
}

.icon-btn:hover {
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

.ghost-link {
  font-size: 15px;
  font-weight: 500;
  color: #475569;
  padding: 8px 14px;
  border-radius: 10px;
  text-decoration: none;
  transition: all 0.18s ease;
}

.ghost-link:hover {
  color: #0f172a;
  background: rgba(15, 23, 42, 0.05);
}

.primary-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 11px 22px;
  border-radius: 999px;
  background: #2563eb;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  box-shadow: 0 6px 18px -4px rgba(37, 99, 235, 0.4);
  transition: background 0.18s ease, transform 0.18s ease, box-shadow 0.18s ease;
}

.primary-pill:hover {
  background: #1d4ed8;
  transform: translateY(-1px);
  box-shadow: 0 10px 24px -4px rgba(37, 99, 235, 0.5);
}
</style>

<style>
html.dark .marketing-nav .center-nav a {
  color: #94a3b8;
}
html.dark .marketing-nav .center-nav a:hover,
html.dark .marketing-nav .center-nav a.active {
  color: #60a5fa;
}
html.dark .marketing-nav .icon-btn {
  color: #94a3b8;
}
html.dark .marketing-nav .icon-btn:hover {
  background: rgba(248, 250, 252, 0.08);
  color: #f8fafc;
}
html.dark .marketing-nav .ghost-link {
  color: #94a3b8;
}
html.dark .marketing-nav .ghost-link:hover {
  color: #f8fafc;
  background: rgba(248, 250, 252, 0.06);
}
</style>
