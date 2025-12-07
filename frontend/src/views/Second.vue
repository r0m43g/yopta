<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  h1.text-3xl.font-bold.mb-8 Antras puslapis
  .card.bg-base-100.shadow-xl
    .card-body
      p.text-lg Tai yra antrojo puslapio turinio pavyzdys.
      p.mt-4 Čia gali būti bet kokia informacija arba funkcionalumas.

      .mt-6.p-4.bg-base-200.rounded-lg
        h3.text-lg.font-semibold.mb-2 Interaktyvus elementas
        p.mb-2 Paspaudimų skaičius: {{ clickCount }}
        button.btn.btn-primary(@click="incrementCount")
          | Paspauskite mane
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useLoggingStore } from '@/stores/logging'

/**
 * A simple secondary page component with an interactive click counter
 */

const clickCount = ref(0)
const loggingStore = useLoggingStore()

/**
 * Increments the click counter and logs the event
 * Adds special logging for milestone (10 clicks)
 */
const incrementCount = () => {
  clickCount.value++

  loggingStore.info('Antro puslapio mygtukas paspaustas', {
    component: 'Second',
    clickCount: clickCount.value,
    timestamp: new Date().toISOString()
  })

  if (clickCount.value === 10) {
    loggingStore.info('Vartotojas pasiekė 10 paspaudimų!', {
      component: 'Second',
      achievement: 'dedicated_clicker',
      timestamp: new Date().toISOString()
    })
  }
}

/**
 * Logs component initialization when mounted
 */
onMounted(() => {
  loggingStore.info('Antras puslapis užkrautas', {
    component: 'Second',
    timestamp: new Date().toISOString()
  })
})

/**
 * Logs component cleanup with final click count when unmounted
 */
onUnmounted(() => {
  loggingStore.info('Antras puslapis uždarytas', {
    component: 'Second',
    finalClickCount: clickCount.value,
    timestamp: new Date().toISOString()
  })
})

</script>
