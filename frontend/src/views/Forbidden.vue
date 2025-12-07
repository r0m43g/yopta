<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .card.w-96.bg-base-100.shadow-xl
    .card-body.items-center.text-center
      h2.card-title.text-3xl.text-error.mb-4 Prieiga uždrausta
      .text-6xl.mb-4 🚫
      p.mb-6 Jūs neturite pakankamai teisių pasiekti šį puslapį.
      .card-actions
        router-link.btn.btn-primary(to="/profile") Grįžti į profilį
</template>

<script setup>
import { onMounted } from 'vue'
import { useLoggingStore } from '@/stores/logging'

/**
 * Forbidden page component displayed when a user attempts to access
 * a restricted area without proper permissions. Serves as an access
 * control boundary with a user-friendly error message and navigation
 * option to return to an authorized section of the application.
 */

/**
 * Logs an access attempt to a forbidden page when the component mounts.
 * Captures component name, timestamp, and requested URL for security
 * audit and debugging purposes.
 */
onMounted(() => {
  const loggingStore = useLoggingStore()
  loggingStore.warn('Vartotojas bandė pasiekti draudžiamą puslapį', {
    component: 'Forbidden',
    timestamp: new Date().toISOString(),
    path: window.location.href
  })
})
</script>
