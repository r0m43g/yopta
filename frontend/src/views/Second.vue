<template lang="pug">
.antras-page.p-6.max-w-8xl.mx-auto.h-screen.w-screen.overflow-auto.scrollbar-thin
  //- Header section
  .flex.flex-wrap.justify-between.items-center.mb-6.gap-4
    div
      h1.text-3xl.font-bold.mb-2 Traukinių eismas
      p.text-sm.text-base-content.opacity-70(v-if="fileName") 
        | Failas: 
        span.font-semibold {{ fileName }}
        |  | Įkelta: {{ formatDateTime(lastImported) }}
    
    .flex.gap-3.flex-wrap
      //- Import Excel button (disabled if mappings not loaded)
      label.btn.btn-primary.gap-2(:class="{ 'btn-disabled': !canImport }")
        svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12")
        | Importuoti Excel
        input.hidden(
          type="file"
          accept=".xlsx,.xls"
          @change="handleFileUpload"
          ref="fileInput"
          :disabled="!canImport"
        )
      
      //- Export JSON button
      button.btn.btn-outline.btn-success(
        v-if="hasData"
        @click="exportToJson"
      )
        svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z")
        | Eksportuoti JSON
      
      //- Clear data button
      button.btn.btn-outline.btn-error(
        v-if="hasData"
        @click="confirmClearData"
      )
        svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")
        | Išvalyti

  //- Loading field mappings indicator
  .alert.alert-warning.mb-6(v-if="!fieldMappingsLoaded && !isLoading")
    svg.stroke-current.shrink-0.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24")
      path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z")
    div
      div.font-bold Laukų atvaizdavimai neįkelti
      div.text-sm Bandoma įkelti laukų atvaizdavimus iš serverio...

  //- Info alert when no data
  .alert.alert-info.mb-6(v-if="!hasData && fieldMappingsLoaded && !isLoading")
    svg.stroke-current.shrink-0.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24")
      path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z")
    div
      div.font-bold Pradėkite nuo duomenų importavimo
      div.text-sm Pasirinkite Excel failą su traukinių judėjimo duomenimis

  //- Loading state
  .flex.justify-center.items-center.py-12(v-if="isLoading")
    div.text-center
      span.loading.loading-spinner.loading-lg.mb-4
      p.text-lg Importuojama...

  //- Statistics cards
  .grid.grid-cols-6.gap-4.mb-6(v-if="hasData && !isLoading")
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Stotys
      .stat-value.text-2xl {{ stats.stations }}
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Riedmenys
      .stat-value.text-2xl {{ stats.vehicles }}
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Personalas
      .stat-value.text-2xl {{ stats.staff }}
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Pamainos
      .stat-value.text-2xl {{ stats.duties }}
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Atvykimai
      .stat-value.text-2xl {{ stats.arrivals }}
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Išvykimai
      .stat-value.text-2xl {{ stats.departures }}

  //- Data display section with tabs
  .bg-base-100.rounded-lg.shadow-lg(v-if="hasData && !isLoading")
    //- Tab navigation
    .tabs.tabs-boxed.p-2.bg-base-200.rounded-t-lg
      a.tab(:class="{ 'tab-active': activeTab === 'stations' }" @click="activeTab = 'stations'") Stotys
      a.tab(:class="{ 'tab-active': activeTab === 'vehicles' }" @click="activeTab = 'vehicles'") Riedmenys
      a.tab(:class="{ 'tab-active': activeTab === 'staff' }" @click="activeTab = 'staff'") Personalas
      a.tab(:class="{ 'tab-active': activeTab === 'duties' }" @click="activeTab = 'duties'") Pamainos
      a.tab(:class="{ 'tab-active': activeTab === 'raw' }" @click="activeTab = 'raw'") Visi įrašai

    .p-4
      //- Stations tab
      div(v-show="activeTab === 'stations'")
        //- Filters - NO "Visi" option
        .flex.flex-wrap.gap-4.mb-4.items-end
          .form-control.w-64
            label.label
              span.label-text Stotis / Depas
            select.select.select-bordered.w-full(v-model="selectedSheet")
              option(value="" disabled) Pasirinkite stotį...
              option(v-for="sheet in availableSheets" :key="sheet" :value="sheet") {{ sheet }}
          
          .form-control.w-48
            label.label
              span.label-text Data
            select.select.select-bordered.w-full(v-model="selectedDate")
              option(value="") Visos datos
              option(v-for="date in availableDates" :key="date" :value="date") {{ date }}

        //- Selected station data
        .space-y-4(v-if="selectedStation")
          //- Station header
          .flex.justify-between.items-center.mb-4
            div
              h3.text-xl.font-bold {{ selectedStation.code }}
              p.text-sm.text-base-content.opacity-70 {{ selectedStation.networkPointName }}
            .badge.badge-lg.badge-outline(v-if="tracksLoaded")
              | {{ tracks.length }} keliai

          //- Arrivals and Departures cards
          .grid.grid-cols-2.gap-4
            //- Arrivals card
            .card.bg-base-200.shadow
              .card-body.p-4(ref="arrivalsCard")
                h4.card-title.text-success.cursor-pointer(@click="copyCardToClipboard(true)")
                  | Atvykimai {{ selectedStation.code }}
                  svg.w-4.h-4.ml-2.opacity-50(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z")
                .overflow-x-auto
                  table.table.table-xs.table-zebra(v-if="stationArrivals.length > 0")
                    thead
                      tr
                        th Tr.Nr.
                        th Atvykimas
                        th Riedmuo
                        th Mašinistai
                        th Kelias
                        th Pastabos
                    tbody
                      tr(v-for="arr in stationArrivals" :key="arr.trainNo + '-' + arr.arrivalDecimal")
                        td {{ arr.trainNo || '-' }}
                        td {{ formatTime(arr.arrival) }}
                        td.font-mono.text-xs {{ arr.vehicle || '-' }}
                        td.text-xs {{ getDriverNames(arr.staff) }}
                        td {{ arr.targetTrack || '-' }}
                        td ...
                  .text-center.py-4.text-base-content.opacity-50(v-else) Nėra atvykimų

            //- Departures card
            .card.bg-base-200.shadow
              .card-body.p-4(ref="departuresCard")
                h4.card-title.text-error.cursor-pointer(@click="copyCardToClipboard(false)")
                  | Išvykimai {{ selectedStation.code }}
                  svg.w-4.h-4.ml-2.opacity-50(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z")
                .overflow-x-auto
                  table.table.table-xs.table-zebra(v-if="stationDepartures.length > 0")
                    thead
                      tr
                        th Tr.Nr.
                        th Išvykimas
                        th Riedmuo
                        th Mašinistai
                        th Kelias
                        th Pastabos
                    tbody
                      tr(v-for="dep in stationDepartures" :key="dep.trainNo + '-' + dep.departureDecimal")
                        td {{ dep.trainNo || '-' }}
                        td {{ formatTime(dep.departure) }}
                        td.font-mono.text-xs {{ dep.vehicle || '-' }}
                        td.text-xs {{ getDriverNames(dep.staff) }}
                        td {{ dep.startingTrack || '-' }}
                        td ...
                  .text-center.py-4.text-base-content.opacity-50(v-else) Nėra išvykimų

        //- No station selected prompt
        .text-center.py-12(v-else)
          .text-6xl.mb-4 🚉
          p.text-lg.text-base-content.opacity-70 Pasirinkite stotį arba depą

      //- Vehicles tab with search
      div(v-show="activeTab === 'vehicles'")
        .form-control.mb-4.w-full.max-w-xs
          label.label
            span.label-text Ieškoti riedmens
          input.input.input-bordered.w-full(
            type="text"
            v-model="vehicleSearch"
            placeholder="Pvz: 730ML-007, EJ575..."
          )
        .overflow-x-auto
          table.table.table-zebra
            thead
              tr
                th Riedmuo
                th Vagonų Nr.
                th Registracijos Nr.
                th Darbo grafikai
            tbody
              tr(v-for="vehicle in filteredVehicles" :key="vehicle.vehicle")
                td.font-semibold {{ vehicle.vehicle }}
                td.font-mono.text-xs {{ vehicle.vehicleNo.join(', ') }}
                td.font-mono.text-xs.max-w-xs.truncate {{ vehicle.vehicleRegNo.join(', ') }}
                td
                  .flex.flex-wrap.gap-1
                    span.badge.badge-sm.badge-outline(v-for="w in vehicle.vehicleWorkings" :key="w") {{ w }}
          .text-center.py-4.text-base-content.opacity-50(v-if="filteredVehicles.length === 0") Nerasta riedmenų

      //- Staff tab with search
      div(v-show="activeTab === 'staff'")
        .form-control.mb-4.w-full.max-w-md
          label.label
            span.label-text Ieškoti (Nr, vardas, telefonas)
          input.input.input-bordered.w-full(
            type="text"
            v-model="staffSearch"
            placeholder="Pvz: 36178, Andrej, 68288..."
          )
        .overflow-x-auto
          table.table.table-zebra
            thead
              tr
                th Personalo Nr.
                th Pareigos
                th Vardas
                th Telefonas
                th Pamainos
            tbody
              tr(v-for="person in filteredStaff" :key="person.id")
                td.font-mono {{ person.id }}
                td
                  span.badge(:class="occBadgeClass(person.occ)") {{ occLabel(person.occ) }}
                td {{ person.name }}
                td.font-mono.text-sm {{ person.phone || '-' }}
                td
                  .flex.flex-wrap.gap-1
                    span.badge.badge-sm.badge-ghost(v-for="d in person.duties" :key="d") {{ d }}
          .text-center.py-4.text-base-content.opacity-50(v-if="filteredStaff.length === 0") Nerasta personalo

      //- Duties tab with search
      div(v-show="activeTab === 'duties'")
        .form-control.mb-4.w-full.max-w-xs
          label.label
            span.label-text Ieškoti pamainos
          input.input.input-bordered.w-full(
            type="text"
            v-model="dutySearch"
            placeholder="Pvz: MLT202, KRA108..."
          )
        .overflow-x-auto
          table.table.table-zebra
            thead
              tr
                th Pamaina
                th Pradžia
                th Pabaiga
                th Traukiniai
            tbody
              tr(v-for="duty in filteredDuties" :key="duty.id")
                td.font-semibold {{ duty.id }}
                td {{ formatTime(duty.startingTime) }}
                td {{ formatTime(duty.endTime) }}
                td
                  .flex.flex-wrap.gap-1
                    span.badge.badge-sm.badge-primary(v-for="t in duty.trains" :key="t") {{ t }}
          .text-center.py-4.text-base-content.opacity-50(v-if="filteredDuties.length === 0") Nerasta pamainų

      //- Raw records tab
      div(v-show="activeTab === 'raw'")
        //- Pagination info
        .flex.justify-between.items-center.mb-4
          p.text-sm.text-base-content.opacity-70 Rodoma {{ paginatedRecords.length }} iš {{ filteredRawRecords.length }} įrašų
          .join
            button.join-item.btn.btn-sm(:disabled="currentPage <= 1" @click="currentPage--") «
            span.join-item.btn.btn-sm.btn-disabled {{ currentPage }} / {{ totalPages }}
            button.join-item.btn.btn-sm(:disabled="currentPage >= totalPages" @click="currentPage++") »

        .overflow-x-auto
          table.table.table-xs.table-zebra
            thead
              tr
                th Lapas
                th Traukinys (atv/išv)
                th Atvykimas
                th Išvykimas
                th Riedmuo (atv/išv)
                th Pradžia
                th Pabaiga
            tbody
              tr(v-for="record in paginatedRecords" :key="record.id")
                td.font-mono.text-xs {{ record.sheetName }}
                td {{ record.trainNoIn || '-' }} / {{ record.trainNoOut || '-' }}
                td {{ formatTime(record.arrivalTime) }}
                td {{ formatTime(record.departureTime) }}
                td.font-mono.text-xs {{ record.vehicleIn || '-' }} / {{ record.vehicleOut || '-' }}
                td.text-xs {{ record.startingLocationIn || '-' }}
                td.text-xs {{ record.endLocationIn || '-' }}

  //- Timeline drawer (only when tracks are loaded)
  .timeline-drawer-wrapper(
    v-if="tracksLoaded && tracks.length > 0 && selectedSheet"
    :class="{ 'timeline-open': isDrawerOpen }"
  )
    .timeline-drawer
      .timeline-tab(
        @click="toggleDrawer"
        :class="{ 'timeline-tab-active': isDrawerOpen }"
      )
        .timeline-tab-icon
          svg(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7")
      .timeline-drawer-content
        h2.text-xl.font-bold.mb-4 Kelių Grafikas - {{ selectedSheet }}
        TracksTimeLine(
          :selectedDate="timelineDate"
          :selectedDepot="selectedSheet"
          :tracks="tracks"
          :arrivals="stationArrivals"
          :departures="stationDepartures"
          @track-assigned="handleTrackAssigned"
        )

  //- Clear confirmation modal
  dialog.modal(:class="{ 'modal-open': showClearModal }")
    .modal-box
      h3.font-bold.text-lg Išvalyti duomenis?
      p.py-4 Ar tikrai norite išvalyti visus importuotus duomenis?
      p.text-sm.text-warning Šis veiksmas negalimas atšaukti.
      
      .stats.shadow.w-full.my-4(v-if="hasData")
        .stat.py-2
          .stat-title.text-xs Iš viso įrašų
          .stat-value.text-2xl {{ stats.records }}
        .stat.py-2
          .stat-title.text-xs Stotys / Depai
          .stat-value.text-2xl {{ stats.stations }}
      
      .modal-action
        button.btn(@click="showClearModal = false") Atšaukti
        button.btn.btn-error(@click="clearData") Išvalyti
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useAntrasStore } from '@/stores/antras'
import { useClipStore } from '@/stores/clip'
import { useLoggingStore } from '@/stores/logging'
import { useSlogStore } from '@/stores/slog'
import api from '@/services/api'
import TracksTimeLine from '@/components/parking/TracksTimeLine.vue'

const antrasStore = useAntrasStore()
const clipStore = useClipStore()
const loggingStore = useLoggingStore()
const slogStore = useSlogStore()

// Refs
const fileInput = ref(null)
const selectedSheet = ref('')
const selectedDate = ref('')
const showClearModal = ref(false)
const activeTab = ref('stations')

// Search filters
const vehicleSearch = ref('')
const staffSearch = ref('')
const dutySearch = ref('')

// Timeline drawer
const isDrawerOpen = ref(false)
const tracks = ref([])
const tracksLoaded = ref(false)

// Card refs for copy
const arrivalsCard = ref(null)
const departuresCard = ref(null)

// Pagination
const currentPage = ref(1)
const itemsPerPage = ref(50)

// Computed properties
const isLoading = computed(() => antrasStore.isLoading)
const fieldMappingsLoaded = computed(() => antrasStore.fieldMappingsLoaded)
const canImport = computed(() => antrasStore.fieldMappingsLoaded)
const hasData = computed(() => antrasStore.records.length > 0)

const availableSheets = computed(() => antrasStore.availableSheets)
const availableDates = computed(() => antrasStore.availableDates)
const fileName = computed(() => antrasStore.fileName)
const lastImported = computed(() => antrasStore.lastImported)

// Statistics
const stats = computed(() => antrasStore.getStatistics())
const timelineDate = computed(() => {
  // 1. Выбранная дата если есть
  if (selectedDate.value) return selectedDate.value
  
  // 2. Первая дата из arrivals
  if (stationArrivals.value.length > 0) {
    const firstArrival = stationArrivals.value[0]
    if (firstArrival.arrival instanceof Date) {
      return firstArrival.arrival.toISOString().split('T')[0]
    }
  }
  
  // 3. Первая дата из departures
  // ...
  
  // 4. Сегодня (fallback)
  return new Date().toISOString().split('T')[0]
})
// Selected station data
const selectedStation = computed(() => {
  if (!selectedSheet.value) return null
  return antrasStore.getStationByCode(selectedSheet.value)
})

// Station arrivals filtered by date
const stationArrivals = computed(() => {
  if (!selectedStation.value) return []
  let arrivals = selectedStation.value.arrivals
  
  if (selectedDate.value) {
    arrivals = arrivals.filter(arr => {
      if (arr.arrival instanceof Date) {
        return arr.arrival.toISOString().split('T')[0] === selectedDate.value
      }
      return false
    })
  }
  
  return arrivals.sort((a, b) => (a.arrivalDecimal || 0) - (b.arrivalDecimal || 0))
})

// Station departures filtered by date
const stationDepartures = computed(() => {
  if (!selectedStation.value) return []
  let departures = selectedStation.value.departures
  
  if (selectedDate.value) {
    departures = departures.filter(dep => {
      if (dep.departure instanceof Date) {
        return dep.departure.toISOString().split('T')[0] === selectedDate.value
      }
      return false
    })
  }
  
  return departures.sort((a, b) => (a.departureDecimal || 0) - (b.departureDecimal || 0))
})

// Filtered vehicles by search
const filteredVehicles = computed(() => {
  const search = vehicleSearch.value.toLowerCase().trim()
  if (!search) return antrasStore.vehicles
  
  return antrasStore.vehicles.filter(v => 
    v.vehicle.toLowerCase().includes(search) ||
    v.vehicleNo.some(n => n.toLowerCase().includes(search)) ||
    v.vehicleWorkings.some(w => w.toLowerCase().includes(search))
  )
})

// Filtered staff by search
const filteredStaff = computed(() => {
  const search = staffSearch.value.toLowerCase().trim()
  if (!search) return antrasStore.staff
  
  return antrasStore.staff.filter(p => 
    (p.id && p.id.toLowerCase().includes(search)) ||
    (p.name && p.name.toLowerCase().includes(search)) ||
    (p.phone && p.phone.toLowerCase().includes(search))
  )
})

// Filtered duties by search
const filteredDuties = computed(() => {
  const search = dutySearch.value.toLowerCase().trim()
  if (!search) return antrasStore.duties
  
  return antrasStore.duties.filter(d => 
    d.id.toLowerCase().includes(search) ||
    d.trains.some(t => String(t).toLowerCase().includes(search))
  )
})

// Raw records filtered by selected sheet
const filteredRawRecords = computed(() => {
  let records = antrasStore.records
  
  if (selectedSheet.value) {
    records = records.filter(r => r.sheetName === selectedSheet.value)
  }
  
  if (selectedDate.value) {
    records = records.filter(r => {
      const recordDate = r.arrivalTime || r.departureTime
      if (recordDate instanceof Date) {
        return recordDate.toISOString().split('T')[0] === selectedDate.value
      }
      return false
    })
  }
  
  return records
})

// Pagination
const totalPages = computed(() => Math.max(1, Math.ceil(filteredRawRecords.value.length / itemsPerPage.value)))

const paginatedRecords = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredRawRecords.value.slice(start, end)
})

/**
 * Get driver names (only occ='M') from staff array
 */
const getDriverNames = (staffList) => {
  if (!staffList || staffList.length === 0) return '-'
  
  // Filter only drivers (occ = 'M')
  const drivers = staffList.filter(s => s.occ === 'M')
  if (drivers.length === 0) return '-'
  
  // Return all driver names
  return drivers.map(d => {
    // Format: "Vardas P." from "Pavardė, Vardas" or just name
    const name = d.name || ''
    if (name.includes(',')) {
      const parts = name.split(',')
      return `${parts[1].trim()} ${parts[0].trim().charAt(0)}.`
    }
    return name
  }).join(', ')
}

/**
 * Handle Excel file upload
 */
const handleFileUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  loggingStore.info('Excel failas pasirinktas', {
    component: 'Second',
    fileName: file.name,
    fileSize: file.size
  })

  const count = await antrasStore.importFromExcel(file)

  if (count > 0) {
    // Reset filters and pagination
    selectedSheet.value = ''
    selectedDate.value = ''
    currentPage.value = 1
    activeTab.value = 'stations'
    
    // Auto-select first station if available
    if (availableSheets.value.length > 0) {
      selectedSheet.value = availableSheets.value[0]
    }
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
 * Format time only
 */
const formatTime = (date) => {
  if (!date) return '-'
  if (date instanceof Date) {
    return date.toLocaleTimeString('lt-LT', { hour: '2-digit', minute: '2-digit' })
  }
  return String(date)
}

/**
 * Get occupation label
 */
const occLabel = (occ) => {
  switch (occ) {
    case 'M': return 'Mašinistas'
    case 'K': return 'Konduktorius'
    default: return occ || '-'
  }
}

/**
 * Get occupation badge class
 */
const occBadgeClass = (occ) => {
  switch (occ) {
    case 'M': return 'badge-primary'
    case 'K': return 'badge-secondary'
    default: return 'badge-ghost'
  }
}

/**
 * Copy card content to clipboard
 */
const copyCardToClipboard = (arrivals = false) => {
  const card = arrivals ? arrivalsCard.value : departuresCard.value
  if (card) {
    const range = document.createRange()
    range.selectNode(card)
    window.getSelection().removeAllRanges()
    window.getSelection().addRange(range)
    document.execCommand('copy')
    window.getSelection().removeAllRanges()
    
    slogStore.addToast({
      message: 'Kopijuota į iškarpinę',
      type: 'alert-success'
    })
    
    loggingStore.info('Lentelė nukopijuota', {
      component: 'Second',
      type: arrivals ? 'arrivals' : 'departures',
      station: selectedSheet.value
    })
  } else {
    slogStore.addToast({
      message: 'Nepavyko kopijuoti',
      type: 'alert-error'
    })
  }
}

/**
 * Toggle timeline drawer
 */
const toggleDrawer = () => {
  isDrawerOpen.value = !isDrawerOpen.value
  
  loggingStore.uiEvent('Second', `timeline_drawer_${isDrawerOpen.value ? 'opened' : 'closed'}`, {
    component: 'TimelineDrawer',
    station: selectedSheet.value
  })
}

/**
 * Handle track assignment from TracksTimeLine
 */
const handleTrackAssigned = (data) => {
  if (!data || !data.id) return

  try {
    if (data.type === 'arrival') {
      const arrival = stationArrivals.value.find(arr => arr.id === data.id)
      if (arrival) {
        arrival.targetTrack = data.trackAssignment || null
        
        loggingStore.info('Kelias priskirtas atvykimui', {
          component: 'Second',
          train: arrival.trainNo,
          vehicle: arrival.vehicle,
          track: data.trackAssignment
        })
      }
    } else if (data.type === 'departure') {
      const departure = stationDepartures.value.find(dep => dep.id === data.id)
      if (departure) {
        departure.startingTrack = data.trackAssignment || null
        
        loggingStore.info('Kelias priskirtas išvykimui', {
          component: 'Second',
          train: departure.trainNo,
          vehicle: departure.vehicle,
          track: data.trackAssignment
        })
      }
    }
  } catch (error) {
    loggingStore.error('Klaida priskiriant kelią', {
      component: 'Second',
      error: error.message
    })
  }
}

/**
 * Load tracks for selected station from clipStore
 */
const loadTracksForStation = () => {
  if (!selectedSheet.value) {
    tracksLoaded.value = false
    tracks.value = []
    return
  }

  // Get tracks from clipStore (stations already loaded)
  if (clipStore.stationsAvailable) {
    const station = clipStore.stations.find(s => s.code === selectedSheet.value)
    if (station && station.tracks && station.tracks.length > 0) {
      tracks.value = station.tracks.sort((a, b) => a.id - b.id)
      tracksLoaded.value = true
      
      loggingStore.info('Keliai rasti stočiai', {
        component: 'Second',
        station: selectedSheet.value,
        tracksCount: tracks.value.length
      })
      return
    }
  }

  // No tracks found for this station
  tracksLoaded.value = false
  tracks.value = []
}

/**
 * Load stations data
 */
const loadStations = async () => {
  try {
    const response = await api.get('/stations')
    clipStore.setStations(response.data)
  } catch (error) {
    loggingStore.error('Klaida įkeliant stočių duomenis', {
      component: 'Second',
      error: error.message
    })
  }
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
  activeTab.value = 'stations'
  tracksLoaded.value = false
  tracks.value = []

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

// Watch selected sheet to load tracks
watch(selectedSheet, () => {
  antrasStore.setFilters({
    sheet: selectedSheet.value,
    date: selectedDate.value
  })
  currentPage.value = 1
  isDrawerOpen.value = false
  
  loadTracksForStation()
})

// Watch selected date
watch(selectedDate, () => {
  antrasStore.setFilters({
    sheet: selectedSheet.value,
    date: selectedDate.value
  })
  currentPage.value = 1
})

// Load field mappings and stations on mount
onMounted(async () => {
  const success = await antrasStore.loadFieldMappings()
  
  if (!success) {
    loggingStore.warn('Nepavyko įkelti laukų atvaizdavimų startuojant', {
      component: 'Second',
      action: 'mount_load_mappings_failed'
    })
  }

  // Load stations for tracks
  if (!clipStore.stationsAvailable) {
    await loadStations()
  }

  loggingStore.info('Antras puslapis atidarytas', {
    component: 'Second',
    action: 'page_opened',
    fieldMappingsLoaded: success
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

/* Timeline drawer styles */
.timeline-drawer-wrapper {
  position: fixed;
  top: 0;
  right: -80vw;
  width: 80vw;
  height: 100vh;
  z-index: 1000;
  transition: right 0.3s ease-in-out;
}

.timeline-drawer-wrapper.timeline-open {
  right: 0;
}

.timeline-drawer {
  position: relative;
  width: 100%;
  height: 100%;
  background: oklch(var(--b1));
  box-shadow: -4px 0 20px rgba(0, 0, 0, 0.3);
}

.timeline-tab {
  position: absolute;
  left: -40px;
  top: 50%;
  transform: translateY(-50%);
  width: 40px;
  height: 80px;
  background: oklch(var(--p));
  border-radius: 8px 0 0 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.timeline-tab:hover {
  background: oklch(var(--pf));
}

.timeline-tab-active {
  background: oklch(var(--s));
}

.timeline-tab-icon {
  color: oklch(var(--pc));
  transition: transform 0.3s;
}

.timeline-open .timeline-tab-icon {
  transform: rotate(180deg);
}

.timeline-tab-icon svg {
  width: 24px;
  height: 24px;
}

.timeline-drawer-content {
  padding: 1rem;
  height: 100%;
  overflow-y: auto;
}

/* Print styles */
@media print {
  .btn,
  .form-control,
  .modal,
  .tabs,
  .pagination,
  .timeline-drawer-wrapper {
    display: none !important;
  }
}
</style>
