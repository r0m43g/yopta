<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.bg-base-200.px-4
  .card.w-full.max-w-md.bg-base-100.shadow-xl
    .card-body.items-center.text-center
      h1.card-title.text-8xl.font-bold.text-error.mb-2 404
      .divider
      h2.text-2xl.font-semibold.mb-4 Puslapis nerastas

      .my-6.text-6xl
        .animate-bounce
          | 🧭
        .mt-2.text-base.opacity-75 Kažkur pasiklydome...

      p.mb-6 Deja, jūsų ieškomas puslapis neegzistuoja arba buvo perkeltas į kitą vietą.

      .card-actions.flex.flex-row.gap-2.justify-center
        button.btn.btn-outline(@click="goBack")
          svg.h-5.w-5.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18")
          | Grįžti atgal
        router-link.btn.btn-primary(to="/")
          svg.h-5.w-5.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6")
          | Grįžti į pagrindinį

      .mt-8.text-sm.opacity-75
        p.italic Galbūt norėjote aplankyti:
        .flex.flex-wrap.justify-center.gap-2.mt-2
          router-link.link.link-hover(to="/profile") Profilis
          router-link.link.link-hover(to="/first") Pirmas puslapis
          router-link.link.link-hover(to="/second") Antras puslapis
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useLoggingStore } from '@/stores/logging'

const router = useRouter()
const loggingStore = useLoggingStore()

/**
 * Navigates the user back one step in their browser history.
 * Provides a simple way for users to return to their previous location
 * after encountering a not-found page.
 */
const goBack = () => {
  router.go(-1)
}

/**
 * Logs 404 errors when the component mounts to track broken links
 * or navigation issues. Records component name, current path,
 * referrer information, and timestamp for debugging purposes.
 */
onMounted(() => {
  loggingStore.warn('Puslapis nerastas (404)', {
    component: 'NotFound',
    path: window.location.pathname,
    referrer: document.referrer,
    timestamp: new Date().toISOString()
  })
})
</script>

<style scoped>
/* Add a subtle pulsing animation to the 404 text */
h1 {
  animation: pulse 3s infinite;
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.8; }
  100% { opacity: 1; }
}
</style>
