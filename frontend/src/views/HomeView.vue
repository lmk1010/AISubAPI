<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div v-else class="home-shell">
    <!-- Decorative pixel-wave background (constrained band, not full-page wash) -->
    <div class="bg-layer" aria-hidden="true">
      <img class="bg-image" src="/home-hero-bg.png?v=2" alt="" aria-hidden="true" />
    </div>

    <!-- Header -->
    <header class="site-header">
      <nav class="site-nav">
        <!-- Brand -->
        <router-link to="/" class="brand">
          <span class="brand-mark">
            <img
              v-if="siteLogo"
              :src="siteLogo"
              alt="Logo"
              class="h-[46px] w-[46px] rounded-lg object-contain"
            />
            <svg v-else viewBox="0 0 32 32" class="h-[46px] w-[46px]">
              <defs>
                <linearGradient id="brandGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" stop-color="#3b82f6" />
                  <stop offset="100%" stop-color="#6366f1" />
                </linearGradient>
              </defs>
              <path
                d="M16 3 L28 10 L28 22 L16 29 L4 22 L4 10 Z"
                fill="url(#brandGrad)"
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
          <li><a href="#products">{{ t('home.nav.products') }}</a></li>
          <li><a href="#pricing">{{ t('home.nav.pricing') }}</a></li>
          <li><a href="#blog">{{ t('home.nav.blog') }}</a></li>
          <li>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">
              {{ t('home.nav.docs') }}
            </a>
            <a v-else href="#docs">{{ t('home.nav.docs') }}</a>
          </li>
          <li><a href="#changelog">{{ t('home.nav.changelog') }}</a></li>
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

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="ghost-link"
          >
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

    <!-- Hero -->
    <main class="hero">
      <!-- Floating brand pills -->
      <div class="brand-pill brand-pill--openai" aria-hidden="true">
        <span class="brand-pill__icon">
          <svg viewBox="0 0 24 24" fill="currentColor" class="h-4 w-4 text-gray-900">
            <path d="M22.282 9.821a5.985 5.985 0 0 0-.516-4.91 6.046 6.046 0 0 0-6.51-2.9A6.065 6.065 0 0 0 4.981 4.18a5.985 5.985 0 0 0-3.998 2.9 6.046 6.046 0 0 0 .743 7.097 5.98 5.98 0 0 0 .51 4.911 6.051 6.051 0 0 0 6.515 2.9A5.985 5.985 0 0 0 13.26 24a6.056 6.056 0 0 0 5.772-4.206 5.99 5.99 0 0 0 3.997-2.9 6.056 6.056 0 0 0-.747-7.073zM13.26 22.43a4.476 4.476 0 0 1-2.876-1.04l.141-.081 4.779-2.758a.795.795 0 0 0 .392-.681v-6.737l2.02 1.168a.071.071 0 0 1 .038.052v5.583a4.504 4.504 0 0 1-4.494 4.494zM3.6 18.304a4.47 4.47 0 0 1-.535-3.014l.142.085 4.783 2.759a.771.771 0 0 0 .78 0l5.843-3.369v2.332a.08.08 0 0 1-.033.062L9.74 19.95a4.5 4.5 0 0 1-6.14-1.646zM2.34 7.896a4.485 4.485 0 0 1 2.366-1.973l-.001.142v5.516a.768.768 0 0 0 .388.676l5.819 3.355-2.02 1.168a.076.076 0 0 1-.071 0l-4.83-2.787a4.504 4.504 0 0 1-1.65-6.097zm16.59 3.86l-5.843-3.387L15.092 7.2a.072.072 0 0 1 .071 0l4.83 2.791a4.495 4.495 0 0 1-.676 8.105v-5.678a.79.79 0 0 0-.392-.681zm2.01-3.013l-.142-.085-4.774-2.782a.776.776 0 0 0-.785 0L9.396 9.255V6.923a.07.07 0 0 1 .029-.058l4.83-2.787a4.5 4.5 0 0 1 6.685 4.66zM8.295 12.834l-2.02-1.164a.08.08 0 0 1-.038-.057V6.039a4.5 4.5 0 0 1 7.375-3.453l-.142.08L8.69 5.426a.795.795 0 0 0-.39.68z" />
          </svg>
        </span>
        <span>OpenAI</span>
      </div>

      <div class="brand-pill brand-pill--anthropic" aria-hidden="true">
        <span class="brand-pill__icon">
          <svg viewBox="0 0 24 24" class="h-4 w-4">
            <path d="M13.827 3.52h3.603L24 20h-3.603zM6.155 3.52h3.752L16.48 20h-3.685l-1.272-3.5H4.169l-1.27 3.5H0zm4.155 9.32-2.328-6.4-2.302 6.4z" fill="#1f1f1f"/>
          </svg>
        </span>
        <span>Anthropic</span>
      </div>

      <div class="brand-pill brand-pill--google" aria-hidden="true">
        <span class="brand-pill__icon">
          <svg viewBox="0 0 48 48" class="h-4 w-4">
            <path fill="#FFC107" d="M43.611 20.083H42V20H24v8h11.303a12.04 12.04 0 0 1-4.087 5.571l6.19 5.238C41.38 35.091 44 30 44 24c0-1.341-.138-2.65-.389-3.917z"/>
            <path fill="#FF3D00" d="m6.306 14.691 6.571 4.819A11.99 11.99 0 0 1 24 12c3.059 0 5.842 1.154 7.961 3.039l5.657-5.657C34.046 6.053 29.268 4 24 4 16.318 4 9.656 8.337 6.306 14.691z"/>
            <path fill="#4CAF50" d="M24 44c5.166 0 9.86-1.977 13.409-5.192l-6.19-5.238A11.96 11.96 0 0 1 24 36a11.99 11.99 0 0 1-11.286-7.946l-6.522 5.025C9.505 39.556 16.227 44 24 44z"/>
            <path fill="#1976D2" d="M43.611 20.083 43.595 20H24v8h11.303a12.04 12.04 0 0 1-4.087 5.571l6.19 5.238C36.971 39.205 44 34 44 24c0-1.341-.138-2.65-.389-3.917z"/>
          </svg>
        </span>
        <span>Google</span>
      </div>

      <div class="brand-pill brand-pill--deepseek" aria-hidden="true">
        <span class="brand-pill__icon">
          <svg viewBox="0 0 24 24" class="h-4 w-4">
            <path d="M21.5 9.5c-.6 0-1.2.4-1.6.9-.7-1.6-2.7-3.7-7.4-3.7-3 0-6.6 1.4-8.5 4-1.4 1.9-1.5 4.4-.4 6.4 1.6 2.7 4.7 4.4 8.4 4.4 4.7 0 8.5-2.6 9.6-6.5.4-1.4.4-2.7 0-3.9.4 0 .8-.1 1.1-.3.6-.4 1-1 1-1.7 0-1-.8-1.6-2.2-1.6zM12 18.4c-3.4 0-6.2-1.4-7.5-3.7-.7-1.3-.7-2.9.1-4.1 1.4-2 4.3-3.2 7.1-3.2 4.6 0 6.5 2.3 6.7 4.8.2 2.6-1.4 6.2-6.4 6.2zM10 11.5c-.6 0-1.1.5-1.1 1.1s.5 1.1 1.1 1.1 1.1-.5 1.1-1.1-.5-1.1-1.1-1.1z" fill="#1f6feb"/>
          </svg>
        </span>
        <span>DeepSeek</span>
      </div>

      <div class="hero-content">
        <span class="hero-badge">{{ t('home.heroBadge') }}</span>

        <h1 class="hero-title">
          <span>{{ t('home.heroTitleLead') }}</span>
          <span class="hero-title__ai">&nbsp;{{ t('home.heroTitleAi') }}&nbsp;</span>
          <span>{{ t('home.heroTitleTail') }}</span>
        </h1>

        <p class="hero-desc">{{ t('home.heroDescNew') }}</p>

        <div class="hero-cta">
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="cta-primary"
          >
            <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
            <Icon name="arrowRight" size="sm" :stroke-width="2.2" />
          </router-link>

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="cta-secondary"
          >
            <Icon name="document" size="sm" :stroke-width="1.8" />
            <span>{{ t('home.viewDocsBtn') }}</span>
          </a>
          <a v-else href="#docs" class="cta-secondary">
            <Icon name="document" size="sm" :stroke-width="1.8" />
            <span>{{ t('home.viewDocsBtn') }}</span>
          </a>
        </div>
      </div>

      <!-- Stats row -->
      <div class="stats-row">
        <div class="stat-item">
          <div class="stat-head">
            <span class="stat-dot stat-dot--green"></span>
            <span class="stat-value">{{ t('home.stats.uptimeValue') }}</span>
          </div>
          <p class="stat-label">{{ t('home.stats.uptimeLabel') }}</p>
        </div>
        <div class="stat-item">
          <div class="stat-head">
            <span class="stat-dot stat-dot--blue"></span>
            <span class="stat-value">{{ t('home.stats.latencyValue') }}</span>
          </div>
          <p class="stat-label">{{ t('home.stats.latencyLabel') }}</p>
        </div>
        <div class="stat-item">
          <div class="stat-head">
            <span class="stat-dot stat-dot--orange"></span>
            <span class="stat-value">{{ t('home.stats.costValue') }}</span>
          </div>
          <p class="stat-label">{{ t('home.stats.costLabel') }}</p>
        </div>
        <div class="stat-item">
          <div class="stat-head">
            <span class="stat-dot stat-dot--purple"></span>
            <span class="stat-value">{{ t('home.stats.compatValue') }}</span>
          </div>
          <p class="stat-label">{{ t('home.stats.compatLabel') }}</p>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="site-footer">
      <div class="footer-inner">
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div class="footer-links">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
          >{{ t('home.docs') }}</a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))

const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* ── Shell ───────────────────────────────────────── */
.home-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #eef2ff;
  overflow: hidden;
  color: #0f172a;
}

/* ── Background ──────────────────────────────────── */
.bg-layer {
  position: absolute;
  /* Offset the wave band so it sits below the header area, leaving the
     top of the page clean. Both top and bottom edges fade out so the
     wave reads as a floating illustration band, not a top-anchored wash. */
  top: 80px;
  left: 0;
  right: 0;
  height: 55vh;
  pointer-events: none;
  overflow: hidden;
  -webkit-mask-image: linear-gradient(
    180deg,
    rgba(0, 0, 0, 0) 0%,
    rgba(0, 0, 0, 1) 18%,
    rgba(0, 0, 0, 1) 65%,
    rgba(0, 0, 0, 0) 100%
  );
  mask-image: linear-gradient(
    180deg,
    rgba(0, 0, 0, 0) 0%,
    rgba(0, 0, 0, 1) 18%,
    rgba(0, 0, 0, 1) 65%,
    rgba(0, 0, 0, 0) 100%
  );
}

.bg-image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  user-select: none;
  filter: saturate(1.2) contrast(1.05);
}

/* ── Header ──────────────────────────────────────── */
.site-header {
  position: relative;
  z-index: 30;
  padding: 0 32px;
  background: transparent;
}

.site-nav {
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

.center-nav a:hover {
  color: #0f172a;
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

/* ── Hero ────────────────────────────────────────── */
.hero {
  position: relative;
  z-index: 10;
  flex: 1;
  padding: 40px 32px 120px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.hero-content {
  max-width: 820px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  padding: 6px 16px;
  border-radius: 999px;
  background: rgba(219, 234, 254, 0.7);
  color: #1d4ed8;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.01em;
  border: 1px solid rgba(147, 197, 253, 0.5);
  backdrop-filter: blur(8px);
  margin-bottom: 22px;
}

.hero-title {
  font-size: clamp(32px, 3.8vw, 48px);
  font-weight: 700;
  line-height: 1.22;
  letter-spacing: -0.025em;
  color: #0f172a;
  margin: 0 0 20px;
  white-space: nowrap;
}

@media (max-width: 820px) {
  .hero-title {
    white-space: normal;
    font-size: clamp(26px, 6vw, 36px);
  }
}

.hero-title__ai {
  color: #2563eb;
}

.hero-desc {
  font-size: 15px;
  line-height: 1.78;
  color: #475569;
  max-width: 600px;
  margin: 0 0 32px;
}

.hero-cta {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 14px;
}

.cta-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 26px;
  border-radius: 11px;
  background: #2563eb;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  box-shadow: 0 10px 24px -8px rgba(37, 99, 235, 0.5);
  transition: background 0.18s ease, transform 0.18s ease, box-shadow 0.18s ease;
}

.cta-primary:hover {
  background: #1d4ed8;
  transform: translateY(-1px);
  box-shadow: 0 14px 28px -8px rgba(37, 99, 235, 0.6);
}

.cta-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 26px;
  border-radius: 11px;
  background: rgba(255, 255, 255, 0.9);
  color: #0f172a;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  border: 1px solid rgba(203, 213, 225, 0.8);
  backdrop-filter: blur(8px);
  transition: all 0.18s ease;
}

.cta-secondary:hover {
  background: #fff;
  border-color: rgba(148, 163, 184, 0.8);
  transform: translateY(-1px);
}

/* ── Floating brand pills ───────────────────────── */
.brand-pill {
  position: absolute;
  display: inline-flex;
  align-items: center;
  gap: 9px;
  padding: 8px 16px 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(226, 232, 240, 0.9);
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  box-shadow:
    0 8px 24px -10px rgba(15, 23, 42, 0.18),
    0 1px 0 rgba(255, 255, 255, 0.6) inset;
  backdrop-filter: blur(10px);
  z-index: 5;
  white-space: nowrap;
}

.brand-pill::after {
  content: '';
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #94a3b8;
  opacity: 0.7;
}

.brand-pill__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
}

.brand-pill__icon svg {
  width: 100%;
  height: 100%;
}

/* Positions tuned to the reference design (anchored to .hero) */
.brand-pill--openai {
  top: 14%;
  left: 22%;
}

.brand-pill--openai::after {
  right: -26px;
  bottom: -10px;
}

.brand-pill--google {
  top: 22%;
  right: 14%;
}

.brand-pill--google::after {
  left: -26px;
  bottom: -10px;
}

.brand-pill--anthropic {
  top: 56%;
  left: 8%;
}

.brand-pill--anthropic::after {
  right: -26px;
  top: -10px;
}

.brand-pill--deepseek {
  top: 64%;
  right: 11%;
}

.brand-pill--deepseek::after {
  left: -26px;
  top: -10px;
}

@media (max-width: 1023px) {
  .brand-pill {
    display: none;
  }
}

/* ── Stats row ──────────────────────────────────── */
.stats-row {
  /* Anchored to the bottom edge of the hero so it sits just above the
     footer, leaving the hero content area (badge + title + CTA) free to
     vertically center against the wave illustration. */
  position: absolute;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 22px 48px;
  width: calc(100% - 64px);
  max-width: 920px;
}

@media (min-width: 768px) {
  .stats-row {
    grid-template-columns: repeat(4, 1fr);
  }
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.stat-head {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.stat-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.stat-dot--green {
  background: #22c55e;
}
.stat-dot--blue {
  background: #3b82f6;
}
.stat-dot--orange {
  background: #f97316;
}
.stat-dot--purple {
  background: #a855f7;
}

.stat-value {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.stat-label {
  margin: 6px 0 0;
  font-size: 12px;
  color: #64748b;
}

/* ── Footer ─────────────────────────────────────── */
.site-footer {
  position: relative;
  z-index: 10;
  padding: 20px 32px;
  background: transparent;
}

.footer-inner {
  max-width: 1280px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: #64748b;
}

@media (min-width: 640px) {
  .footer-inner {
    flex-direction: row;
    justify-content: space-between;
  }
}

.footer-links {
  display: inline-flex;
  gap: 16px;
}

.footer-links a {
  color: inherit;
  text-decoration: none;
  transition: color 0.18s ease;
}

.footer-links a:hover {
  color: #0f172a;
}
</style>

<!--
  Dark-mode overrides live in a NON-scoped block.
  Reason: Vue scoped CSS mis-compiles `:global(.dark) .x { ... }` into bare
  `.dark { ... }` (we've verified this against the dev-server output) — which
  applies the rules to <html class="dark"> itself (e.g. `opacity: 0.55` ends
  up on the entire HTML element, producing the "white veil" the user sees).
  Using plain `html.dark .x` here gives specificity (0,1,2), which beats the
  scoped (0,1,1) light rules and reliably overrides them.
-->
<style>
html.dark .home-shell {
  background: #0a0f1a;
  color: #e2e8f0;
}

html.dark .bg-image {
  opacity: 0.55;
  filter: hue-rotate(15deg) saturate(1.2);
  mix-blend-mode: screen;
}

html.dark .center-nav a {
  color: #94a3b8;
}

html.dark .center-nav a:hover {
  color: #f8fafc;
}

html.dark .icon-btn {
  color: #94a3b8;
}

html.dark .icon-btn:hover {
  background: rgba(248, 250, 252, 0.08);
  color: #f8fafc;
}

html.dark .ghost-link {
  color: #94a3b8;
}

html.dark .ghost-link:hover {
  color: #f8fafc;
  background: rgba(248, 250, 252, 0.06);
}

html.dark .hero-badge {
  background: rgba(30, 64, 175, 0.25);
  color: #93c5fd;
  border-color: rgba(59, 130, 246, 0.3);
}

html.dark .hero-title {
  color: #f8fafc;
}

html.dark .hero-desc {
  color: #94a3b8;
}

html.dark .cta-secondary {
  background: rgba(30, 41, 59, 0.7);
  color: #f8fafc;
  border-color: rgba(71, 85, 105, 0.6);
}

html.dark .cta-secondary:hover {
  background: rgba(30, 41, 59, 0.9);
}

html.dark .brand-pill {
  background: rgba(30, 41, 59, 0.85);
  border-color: rgba(71, 85, 105, 0.6);
  color: #e2e8f0;
}

html.dark .brand-pill--anthropic svg path {
  fill: #f1f5f9;
}

html.dark .stat-value {
  color: #f8fafc;
}

html.dark .stat-label {
  color: #94a3b8;
}

html.dark .footer-inner {
  color: #94a3b8;
}

html.dark .footer-links a:hover {
  color: #f8fafc;
}
</style>
