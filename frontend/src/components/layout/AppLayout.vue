<template>
  <div class="app-shell min-h-screen">
    <!-- Background image: pixel-wave lavender backdrop, covering the entire viewport -->
    <div class="app-shell__bg" aria-hidden="true"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="app-shell__main relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header (suppressed when the page renders its own inline header) -->
      <AppHeader v-if="!hideHeader" />

      <!-- Main Content -->
      <main :class="hideHeader ? 'px-4 md:px-6 lg:px-8 py-4 md:py-5 lg:py-6' : 'px-4 md:px-6 lg:px-8 pt-2 md:pt-3 lg:pt-4 pb-4 md:pb-5 lg:pb-6'">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

defineProps<{ hideHeader?: boolean }>()

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>

<style scoped>
.app-shell {
  position: relative;
  background-color: #f1f3fb;
}
.app-shell__bg {
  position: fixed;
  inset: 0;
  pointer-events: none;
  background-image: url('/dashboard-bg.png?v=3');
  background-repeat: no-repeat;
  background-position: center top;
  background-size: cover;
  background-attachment: fixed;
  z-index: 0;
}
.app-shell__main {
  position: relative;
  z-index: 1;
}
</style>

<style>
html.dark .app-shell {
  background-color: #0a0f1a;
}
html.dark .app-shell__bg {
  opacity: 0.28;
  filter: hue-rotate(10deg) saturate(0.85);
  mix-blend-mode: screen;
}
</style>
