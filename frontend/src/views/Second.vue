<template lang="pug">
.antras-page.p-6.max-w-7xl.mx-auto
  //- Header section
  .flex.justify-between.items-center.mb-6
    div
      h1.text-3xl.font-bold.mb-2 Antras - Traukinių judėjimas
      p.text-sm.text-base-content.opacity-70(v-if="fileName") 
        | Failas: 
        span.font-semibold {{ fileName }}
        |  | Įkelta: {{ formatDateTime(lastImported) }}
    
    .flex.gap-3
      //- Import Excel button
      label.btn.btn-primary.gap-2
        svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12")
        | Importuoti Excel
        input.hidden(
          type="file"
          accept=".xlsx,.xls"
          @change="handleFileUpload"
          ref="fileInput"
        )
      
      //- Clear data button
      button.btn.btn-outline.btn-error(
        v-if="recordsCount > 0"
        @click="confirmClearData"
      )
        svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")
        | Išvalyti

  //- Info alert when no data
  .alert.alert-info.mb-6(v-if="recordsCount === 0 && !isLoading")
    svg.stroke-current.shrink-0.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24")
      path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z")
    div
      div.font-bold Pradėkite nuo duomenų importavimo
      div.text-sm Pasirinkite Excel failą su traukinių judėjimo duomenimis (182 lapai, 37 stulpeliai)

  //- Loading state
  .flex.justify-center.items-center(v-if="isLoading")
    div.text-center
      span.loading.loading-spinner.loading-lg.mb-4
      p.text-lg Importuojama...

  //- Data display section
  .bg-base-100.rounded-lg.shadow-lg.p-6(v-if="recordsCount > 0 && !isLoading")
    //- Filters section
    .flex.flex-wrap.gap-4.mb-6.items-end
      .form-control.w-64
        label.label
          span.label-text Tinklo punktas (lapas)
        select.select.select-bordered.w-full(v-model="selectedSheet")
          option(value="") Visi punktai ({{ availableSheets.length }})
          option(v-for="sheet in availableSheets" :key="sheet" :value="sheet") {{ sheet }}
      
      .form-control.w-48
        label.label
          span.label-text Data
        select.select.select-bordered.w-full(v-model="selectedDate")
          option(value="") Visos datos
          option(v-for="date in availableDates" :key="date" :value="date") {{ date }}
      
      .flex.gap-2.items-center
        button.btn.btn-outline(@click="resetFilters")
          svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15")
          | Atstatyti
        
        button.btn.btn-outline.btn-success(
          v-if="filteredRecords.length > 0"
          @click="exportToJson"
        )
          svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z")
          | Eksportuoti JSON

    //- Stats
    .stats.shadow.w-full.mb-6
      .stat
        .stat-figure.text-primary
          svg.w-8.h-8(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z")
        .stat-title Iš viso įrašų
        .stat-value.text-primary {{ recordsCount }}
        .stat-desc {{ availableSheets.length }} tinklo punktai

      .stat
        .stat-figure.text-secondary
          svg.w-8.h-8(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z")
        .stat-title Filtruota
        .stat-value.text-secondary {{ filteredRecords.length }}
        .stat-desc {{ selectedSheet || 'Visi punktai' }}

      .stat
        .stat-figure.text-accent
          svg.w-8.h-8(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z")
        .stat-title Datos
        .stat-value.text-accent {{ availableDates.length }}
        .stat-desc {{ selectedDate || 'Visos' }}

    //- Data table
    .overflow-x-auto
      table.table.table-zebra.table-sm.w-full
        thead
          tr.bg-base-200
            th.w-12 #
            th Tinklo punktas
            th Riedmens tipas
            th Riedmens nr.
            th Tr. nr.
            th Atvykimas
            th Išvykimas
            th Mašinistas
            th Telefonas
            th Pareigos
            th Galiojimas
            th Pradžios vieta
            th Pabaigos vieta
        tbody
          tr(v-for="(record, index) in paginatedRecords" :key="record.id")
            td {{ (currentPage - 1) * itemsPerPage + index + 1 }}
            td 
              .badge.badge-primary {{ record.networkPointName }}
            td {{ record.technicalVehicleTypeIn || record.technicalVehicleTypeOut }}
            td {{ record.vehicleNoIn || record.vehicleNoOut }}
            td {{ record.trainNoIn || record.trainNoOut }}
            td 
              .badge.badge-ghost(v-if="record.arrival") {{ record.arrival }}
            td 
              .badge.badge-ghost(v-if="record.departure") {{ record.departure }}
            td 
              .text-sm {{ formatName(record.driverIn || record.driverOut) }}
            td 
              .text-xs {{ record.phoneIn || record.phoneOut }}
            td 
              .text-xs {{ record.dutyIn || record.dutyOut }}
            td 
              .badge.badge-info.badge-sm(v-if="record.validityIn || record.validityOut") 
                | {{ record.validityIn || record.validityOut }}
            td 
              .text-sm {{ record.startingLocationIn || record.startingLocationOut }}
            td 
              .text-sm {{ record.endLocationIn || record.endLocationOut }}

    //- Pagination
    .flex.justify-center.items-center.gap-4.mt-6(v-if="totalPages > 1")
      button.btn.btn-sm(
        :disabled="currentPage === 1"
        @click="currentPage--"
      ) «
      
      .flex.gap-2
        button.btn.btn-sm(
          v-for="page in visiblePages"
          :key="page"
          :class="{ 'btn-active': page === currentPage }"
          @click="currentPage = page"
        ) {{ page }}
      
      button.btn.btn-sm(
        :disabled="currentPage === totalPages"
        @click="currentPage++"
      ) »

  //- Clear confirmation modal
  .modal(:class="{ 'modal-open': showClearModal }")
    .modal-box
      h3.font-bold.text-lg.mb-4 Patvirtinkite išvalymą
      p.mb-4 Ar tikrai norite išvalyti visus importuotus duomenis?
      p.text-sm.text-warning Šis veiksmas negalimas atšaukti.
      .stats.shadow.w-full.my-4
        .stat.py-2
          .stat-title.text-xs Iš viso įrašų
          .stat-value.text-2xl {{ recordsCount }}
        .stat.py-2
          .stat-title.text-xs Tinklo punktai
          .stat-value.text-2xl {{ availableSheets.length }}
      
      .modal-action
        button.btn(@click="showClearModal = false") Atšaukti
        button.btn.btn-error(@click="clearData") Išvalyti
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAntrasStore } from '@/stores/antras'
import { useLoggingStore } from '@/stores/logging'
import { useSlogStore } from '@/stores/slog'

const antrasStore = useAntrasStore()
const loggingStore = useLoggingStore()
const slogStore = useSlogStore()

// Refs
const fileInput = ref(null)
const selectedSheet = ref('')
const selectedDate = ref('')
const showClearModal = ref(false)

// Pagination
const currentPage = ref(1)
const itemsPerPage = ref(50)

// Computed properties
const isLoading = computed(() => antrasStore.isLoading)
const recordsCount = computed(() => antrasStore.recordsCount)
const availableSheets = computed(() => antrasStore.availableSheets)
const availableDates = computed(() => antrasStore.availableDates)
const filteredRecords = computed(() => antrasStore.filteredRecords)
const fileName = computed(() => antrasStore.fileName)
const lastImported = computed(() => antrasStore.lastImported)

const totalPages = computed(() => Math.ceil(filteredRecords.value.length / itemsPerPage.value))
const paginatedRecords = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredRecords.value.slice(start, end)
})

const visiblePages = computed(() => {
  const pages = []
  const maxVisible = 5
  const start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2))
  const end = Math.min(totalPages.value, start + maxVisible - 1)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  return pages
})

/**
 * Handle Excel file upload
 */
const handleFileUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  loggingStore.info('Excel failas pasirinktas', {
    component: 'Antras',
    fileName: file.name,
    fileSize: file.size
  })

  const count = await antrasStore.importFromExcel(file)

  if (count > 0) {
    // Reset filters and pagination
    selectedSheet.value = ''
    selectedDate.value = ''
    currentPage.value = 1
  }

  // Clear file input
  event.target.value = ''
}

/**
 * Format date/time for display
 */
const formatDateTime = (dateStr) => {
  if (!dateStr) return ''
  try {
    return new Date(dateStr).toLocaleString('lt-LT')
  } catch {
    return dateStr
  }
}

/**
 * Format driver name (take first name only)
 */
const formatName = (name) => {
  if (!name) return ''
  // Take first name from comma-separated list
  return name.split(',')[0].trim()
}

/**
 * Reset filters
 */
const resetFilters = () => {
  selectedSheet.value = ''
  selectedDate.value = ''
  currentPage.value = 1

  loggingStore.info('Antras filtrai atstatyti', {
    component: 'Antras',
    action: 'reset_filters'
  })
}

/**
 * Show clear confirmation modal
 */
const confirmClearData = () => {
  showClearModal.value = true
}

/**
 * Clear all data
 */
const clearData = () => {
  antrasStore.clearData()
  showClearModal.value = false
  selectedSheet.value = ''
  selectedDate.value = ''
  currentPage.value = 1

  slogStore.addToast({
    message: 'Visi duomenys išvalyti',
    type: 'alert-success'
  })
}

/**
 * Export to JSON
 */
const exportToJson = () => {
  antrasStore.exportToJson()

  slogStore.addToast({
    message: 'Duomenys eksportuoti į JSON',
    type: 'alert-success'
  })
}

// Watch filters
import { watch } from 'vue'
watch([selectedSheet, selectedDate], () => {
  antrasStore.setFilters({
    sheet: selectedSheet.value,
    date: selectedDate.value
  })
  currentPage.value = 1 // Reset to first page when filters change
})

// Load field mappings on mount
onMounted(async () => {
  await antrasStore.loadFieldMappings()

  loggingStore.info('Antras puslapis atidarytas', {
    component: 'Antras',
    action: 'page_opened'
  })
})
</script>

<style scoped>
/* Responsive table on mobile */
@media (max-width: 768px) {
  .table {
    font-size: 0.75rem;
  }
  
  .table th,
  .table td {
    padding: 0.5rem 0.25rem;
  }
}

/* Print styles */
@media print {
  .btn,
  .form-control,
  .modal,
  .pagination {
    display: none !important;
  }
}
</style>
