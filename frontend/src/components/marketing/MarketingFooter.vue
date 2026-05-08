<template>
  <footer class="marketing-footer">
    <div class="footer-inner">
      <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
      <div class="footer-links">
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.docs') }}</a>
        <a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

const { t } = useI18n()

const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const currentYear = computed(() => new Date().getFullYear())
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
</script>

<style scoped>
.marketing-footer {
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

<style>
html.dark .marketing-footer .footer-inner {
  color: #94a3b8;
}
html.dark .marketing-footer .footer-links a:hover {
  color: #f8fafc;
}
</style>
