<!-- frontend/src/components/common/ConnectionStatus.vue -->
<template>
 <div class="connection-status">
   <div
     v-if="!isConnected"
     class="p-2 bg-warning text-warning-content text-center w-full"
   >
     <div class="flex justify-center items-center gap-2">
       <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
         <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
       </svg>
       <span>{{ statusText }}</span>
       <button
         v-if="!isOnline"
         class="btn btn-xs btn-outline ml-2"
         @click="retryConnection"
       >
         Bandyti dar kartą
       </button>
     </div>
   </div>
   <div v-else class="connection-indicator">
     <span class="indicator-dot" :class="{'connected': isConnected}"></span>
     <span class="indicator-text" v-if="showText">{{ statusText }}</span>
   </div>
 </div>
</template>

<script setup>
import { computed } from 'vue'
import { useConnectionStore } from '@/stores/connection'

// Props definition - showing text is optional and defaults to false
// "To show or not to show, that is the question" - Shakespeare, if he was a frontend dev
const props = defineProps({
 showText: {
   type: Boolean,
   default: false
 }
})

// Get connection store instance - our digital crystal ball for network status
const connectionStore = useConnectionStore()

// Computed property to check if we're connected to the server
// Like checking if your crush has read your message - anxiety-inducing but necessary
const isConnected = computed(() => {
 return connectionStore.isConnected
})

// Computed property to check if we're online at all
// Equivalent of checking if you're wearing pants before a Zoom call
const isOnline = computed(() => {
 return connectionStore.isOnline
})

// Computed property that returns user-friendly status message
// Translates boring network states into Lithuanian panic messages
const statusText = computed(() => {
 if (!isOnline.value) return "Nėra interneto ryšio"
 if (!isConnected.value) return "Serveris nepasiekiamas"
 return "Prisijungta"
})

// Method to manually check server status
// The digital equivalent of frantically pushing elevator buttons
const retryConnection = () => {
 connectionStore.checkServerStatus()
}
</script>

<style scoped>
.connection-status {
 position: fixed;
 top: 0;
 left: 0;
 right: 0;
 z-index: 1000;
 display: flex;
 align-items: center;
}

.connection-indicator {
 display: flex;
 align-items: center;
 gap: 0.5rem;
}

.indicator-dot {
 width: 8px;
 height: 8px;
 border-radius: 50%;
 background-color: #f87171; /* red for disconnection */
 transition: background-color 0.3s;
}

.indicator-dot.connected {
 background-color: #10b981; /* green for connection */
}

.indicator-text {
 font-size: 0.75rem;
 color: #6b7280;
}
</style>
