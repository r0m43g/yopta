<!-- frontend/src/App.vue -->
<template>
  <div class="flex flex-col items-center justify-center h-screen w-screen">
    <!-- Connection status indicator - our digital pulse check -->
    <ConnectionStatus class="mb-4" />
    <AppTabs />

    <!-- Wrap router-view in transition for page transitions -->
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>

    <!-- Toast notifications with animation - our fancy message delivery system -->
    <div class="toast ztop">
      <transition-group name="toast-fade" tag="div">
        <div v-for="(toast, index) in toasts"
          class="alert cursor-pointer toast-item"
          :class="toast.type"
          :key="toast.timestamp || index"
          @click="removeToast(index)"
          v-log-click="`toast_${toast.type}`">
          <span>{{ toast.message }}</span>
        </div>
      </transition-group>
    </div>

    <!-- Debug panel (development mode only) - our digital surgery room -->
    <DebugPanel v-if="isDevelopment" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import AppTabs from './components/Navigation/AppTabs.vue'
import { useSlogStore } from './stores/slog'
import { useLoggingStore } from './stores/logging'
import ConnectionStatus from './components/common/ConnectionStatus.vue'
import DebugPanel from './components/development/DebugPanel.vue'

// Determine if we're in development mode - like checking if we're wearing lab coats
const isDevelopment = ref(process.env.NODE_ENV === 'development')

// Get toast notifications from our toast store - our little message carriers
const toasts = computed(() => {
  return useSlogStore().toasts
})

const removeToast = (index) => {
  useSlogStore().removeToast(index)
}
</script>

<style>
.ztop{
  z-index: 9999;
}
/* Toast animation - giving our notifications a dramatic entrance and exit */
.toast-fade-enter-active {
  transition: all 0.3s ease;
}

.toast-fade-leave-active {
  transition: all 0.8s cubic-bezier(1.0, 0.5, 0.8, 1.0);
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  transform: translateX(20px);
  opacity: 0;
}

/* Base toast styles - dressing our notifications in proper attire */
.toast-item {
  margin-bottom: 0.5rem;
  transition: all 0.3s ease;
}

/* Page transition animations - because abrupt changes are so last decade */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
