<template lang="pug">
.container.mx-auto.p-4.max-w-8xl.h-screen.w-screen.overflow-auto.scrollbar-thin
  //- Header with import controls
  .flex.flex-wrap.justify-between.items-center.mb-6.gap-4
    h1.text-3xl.font-bold Traukinių eismas
    
    .flex.items-center.gap-2
      input.hidden(
        type="file"
        ref="fileInput"
        accept=".xlsx,.xls"
        @change="handleFileUpload"
      )
      
      button.btn.btn-primary(
        @click="$refs.fileInput.click()"
        :disabled="!canImport || isLoading"
      )
        span.loading.loading-spinner.loading-sm(v-if="isLoading")
        svg.w-5.h-5.mr-2(v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12")
        | Importuoti Excel
      
      button.btn.btn-outline.btn-error(
        v-if="hasData"
        @click="showClearModal = true"
      )
        svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")

  .alert.alert-info.mb-4(v-if="fileName")
    svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
      path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z")
    span Failas: {{ fileName }} | Importuota: {{ formatDateTime(lastImported) }}

  .alert.alert-warning.mb-4(v-if="!fieldMappingsLoaded")
    svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
      path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z")
    span Laukų atvaizdavimai dar neįkelti. Palaukite arba atnaujinkite puslapį.

  .grid.grid-cols-2.gap-2.mb-6(v-if="hasData" class="md:grid-cols-4 lg:grid-cols-8")
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
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Traukiniai
      .stat-value.text-2xl {{ stats.trains }}
    .stat.bg-base-100.rounded-lg.shadow.p-3
      .stat-title.text-xs Datos
      .stat-value.text-2xl {{ stats.dates }}

  .bg-base-100.rounded-lg.shadow-lg(v-if="hasData && !isLoading")
    .tabs.tabs-boxed.p-2.bg-base-200.rounded-t-lg
      a.tab(:class="{ 'tab-active': activeTab === 'stations' }" @click="activeTab = 'stations'") Stotys
      a.tab(:class="{ 'tab-active': activeTab === 'vehicles' }" @click="activeTab = 'vehicles'") Riedmenys
      a.tab(:class="{ 'tab-active': activeTab === 'staff' }" @click="activeTab = 'staff'") Personalas
      a.tab(:class="{ 'tab-active': activeTab === 'duties' }" @click="activeTab = 'duties'") Pamainos
      a.tab(:class="{ 'tab-active': activeTab === 'trains' }" @click="activeTab = 'trains'") Traukiniai
      a.tab(:class="{ 'tab-active': activeTab === 'raw' }" @click="activeTab = 'raw'") Visi įrašai

    .p-4
      div(v-show="activeTab === 'stations'")
        //- Filters and controls
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
          
          .flex.gap-2.ml-auto
            .form-control
              label.label.cursor-pointer.gap-2
                span.label-text Pasirinkimas
                input.toggle.toggle-primary(type="checkbox" v-model="selectionEnabled")
            
            button.btn.btn-sm.btn-outline(
              :disabled="selectedRecords.length === 0"
              @click="copyToBasket"
            )
              svg.w-4.h-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3")
              | ({{ selectedRecords.length }})
            
            button.btn.btn-sm.btn-outline.btn-success(
              :disabled="basket.length === 0 || !canPaste"
              @click="pasteFromBasket"
            )
              svg.w-4.h-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2")
              | ({{ basket.length }})
            
            button.btn.btn-sm.btn-outline.btn-error(
              :disabled="basket.length === 0"
              @click="clearBasket"
            )
              svg.w-4.h-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")

        .space-y-4(v-if="selectedStation")
          .flex.justify-between.items-center.mb-4
            div
              h3.text-xl.font-bold {{ selectedStation.code }}
              p.text-sm.text-base-content.opacity-70 {{ selectedStation.networkPointName }}
            .badge.badge-lg.badge-outline(v-if="tracksLoaded")
              | {{ tracks.length }} keliai

          .grid.grid-cols-2.gap-4
            .card.bg-base-200.shadow
              .card-body.p-4(ref="arrivalsCard")
                h4.card-title.text-success.cursor-pointer(@click="copyCardToClipboard(true)")
                  | Atvykimai {{ selectedStation.code }}
                  svg.w-4.h-4.ml-2.opacity-50(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z")
                .overflow-x-auto
                  table.table.table-xs(v-if="stationArrivals.length > 0")
                    thead
                      tr.text-xs.bold.italic
                        th(v-if="selectionEnabled")
                          input.checkbox.checkbox-xs(type="checkbox" @change="toggleAllArrivals" :checked="allArrivalsSelected")
                        th Tr.Nr.
                        th Atvykimas
                        th Riedmuo
                        th Mašinistai
                        th Kelias
                        th Pastabos
                    tbody
                      tr(
                        v-for="arr in stationArrivals"
                        :key="arr.id"
                        :class="getRowClass(arr, 'arrival')"
                        @mouseenter="handleRowHover(arr, 'arrival')"
                        @mouseleave="handleRowLeave"
                      )
                        td(v-if="selectionEnabled")
                          input.checkbox.checkbox-xs(
                            type="checkbox"
                            :checked="isSelected(arr)"
                            @change="toggleSelection(arr, 'arrival')"
                          )
                        td {{ arr.trainNo || '-' }}
                        td {{ formatTime(arr.arrival) }}
                        td {{ arr.vehicle || '-' }}
                        td {{ getDriverNames(arr.staff) }}
                        td {{ arr.targetTrack || '-' }}
                        td {{ arr.departureTrainNo != arr.trainNo ? arr.departureTrainNo : '---'}}
                  .text-center.py-4.text-base-content.opacity-50(v-else) Nėra atvykimų

            .card.bg-base-200.shadow
              .card-body.p-4(ref="departuresCard")
                h4.card-title.text-error.cursor-pointer(@click="copyCardToClipboard(false)")
                  | Išvykimai {{ selectedStation.code }}
                  svg.w-4.h-4.ml-2.opacity-50(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z")
                .overflow-x-auto
                  table.table.table-xs(v-if="stationDepartures.length > 0")
                    thead
                      tr.text-xs.bold.italic
                        th(v-if="selectionEnabled")
                          input.checkbox.checkbox-xs(type="checkbox" @change="toggleAllDepartures" :checked="allDeparturesSelected")
                        th Tr.Nr.
                        th Išvykimas
                        th Riedmuo
                        th Mašinistai
                        th Kelias
                        th Pastabos
                    tbody
                      tr(
                        v-for="dep in stationDepartures"
                        :key="dep.id"
                        :class="getRowClass(dep, 'departure')"
                        @mouseenter="handleRowHover(dep, 'departure')"
                        @mouseleave="handleRowLeave"
                      )
                        td(v-if="selectionEnabled")
                          input.checkbox.checkbox-xs(
                            type="checkbox"
                            :checked="isSelected(dep)"
                            @change="toggleSelection(dep, 'departure')"
                          )
                        td {{ dep.trainNo || '-' }}
                        td {{ formatTime(dep.departure) }}
                        td {{ dep.vehicle || '-' }}
                        td {{ getDriverNames(dep.staff) }}
                        td {{ dep.startingTrack || '-' }}
                        td ---
                  .text-center.py-4.text-base-content.opacity-50(v-else) Nėra išvykimų

        .text-center.py-12(v-else)
          .text-6xl.mb-4 🚉
          p.text-lg.text-base-content.opacity-70 Pasirinkite stotį arba depą

      div(v-show="activeTab === 'vehicles'")
        .form-control.mb-4.w-full.max-w-xs
          label.label
            span.label-text Ieškoti riedmens
          input.input.input-bordered.w-full(type="text" v-model="vehicleSearch" placeholder="Pvz: 730ML-007, EJ575...")
        .overflow-x-auto
          table.table.table-zebra
            thead
              tr
                th Riedmuo
                th Vagonų Nr.
                th Registracijos Nr.
                th Maršrutai
            tbody
              tr(v-for="v in filteredVehicles" :key="v.vehicle")
                td.font-mono.font-bold {{ v.vehicle }}
                td.text-xs {{ v.vehicleNo.join(', ') || '-' }}
                td.text-xs {{ v.vehicleRegNo.join(', ') || '-' }}
                td
                  .flex.flex-wrap.gap-1
                    span.badge.badge-sm.badge-outline(v-for="w in v.vehicleWorkings" :key="w") {{ w }}
          .text-center.py-4.text-base-content.opacity-50(v-if="filteredVehicles.length === 0") Nerasta riedmenų

      div(v-show="activeTab === 'staff'")
        .form-control.mb-4.w-full.max-w-xs
          label.label
            span.label-text Ieškoti darbuotojo
          input.input.input-bordered.w-full(type="text" v-model="staffSearch" placeholder="Vardas, tel. nr., tabelio nr....")
        .overflow-x-auto
          table.table.table-zebra
            thead
              tr
                th Pareigos
                th Tabelio Nr.
                th Vardas
                th Telefonas
                th Pamainos
            tbody
              tr(v-for="s in filteredStaff" :key="s.id")
                td
                  span.badge(:class="occBadgeClass(s.occ)") {{ occLabel(s.occ) }}
                td.font-mono {{ s.id }}
                td {{ s.name }}
                td.text-xs {{ s.phone || '-' }}
                td
                  .flex.flex-wrap.gap-1
                    span.badge.badge-sm.badge-ghost(v-for="d in s.duties.slice(0, 5)" :key="d") {{ d.split(':')[1] || d }}
                    span.badge.badge-sm.badge-ghost(v-if="s.duties.length > 5") +{{ s.duties.length - 5 }}
          .text-center.py-4.text-base-content.opacity-50(v-if="filteredStaff.length === 0") Nerasta darbuotojų

      div(v-show="activeTab === 'duties'")
        .form-control.mb-4.w-full.max-w-xs
          label.label
            span.label-text Ieškoti pamainos
          input.input.input-bordered.w-full(type="text" v-model="dutySearch" placeholder="Pamainos kodas arba traukinio nr....")
        .overflow-x-auto
          table.table.table-zebra
            thead
              tr
                th Data
                th Pamaina
                th Pradžia
                th Pabaiga
                th Traukiniai
            tbody
              tr(v-for="duty in filteredDuties" :key="duty.id + duty.date")
                td {{ duty.date || '-' }}
                td.font-mono.font-bold {{ duty.id }}
                td {{ formatTime(duty.startingTime) }}
                td {{ formatTime(duty.endTime) }}
                td
                  .flex.flex-wrap.gap-1
                    span.badge.badge-sm.badge-primary(v-for="t in duty.trains" :key="t") {{ t }}
          .text-center.py-4.text-base-content.opacity-50(v-if="filteredDuties.length === 0") Nerasta pamainų

      div(v-show="activeTab === 'trains'")
        .flex.flex-wrap.gap-4.mb-4.items-end
          .form-control.w-48
            label.label
              span.label-text Data
            select.select.select-bordered.w-full(v-model="selectedTrainDate")
              option(value="") Visos datos
              option(v-for="date in availableDates" :key="date" :value="date") {{ date }}
          
          .form-control.w-32
            label.label
              span.label-text Traukinys
            input.input.input-bordered.w-full(type="text" v-model="trainSearch" placeholder="Nr....")

        .overflow-x-auto
          table.table.table-zebra
            thead
              tr
                th Tr.Nr.
                th Data
                th Pradžia
                th Pabaiga
                th Personalas
            tbody
              tr(v-for="train in filteredTrains" :key="train.no + train.date")
                td.font-mono.font-bold {{ train.no }}
                td {{ train.date }}
                td {{ train.startingLocation || '-' }}
                td {{ train.endLocation || '-' }}
                td
                  .flex.flex-wrap.gap-1
                    span.badge.badge-sm.badge-outline(v-for="s in train.staff.slice(0, 3)" :key="s.id") {{ s.name.split(',')[0] }}
                    span.badge.badge-sm.badge-outline(v-if="train.staff.length > 3") +{{ train.staff.length - 3 }}
          .text-center.py-4.text-base-content.opacity-50(v-if="filteredTrains.length === 0") Nerasta traukinių

      div(v-show="activeTab === 'raw'")
        //- Controls for raw records
        .flex.flex-wrap.gap-4.mb-4.items-end
          .form-control.w-48
            label.label
              span.label-text Lapas
            select.select.select-bordered.w-full.select-sm(v-model="rawSelectedSheet")
              option(value="") Visi lapai
              option(v-for="sheet in sheets" :key="sheet" :value="sheet") {{ sheet }}
          
          .flex.gap-2.ml-auto
            .form-control
              label.label.cursor-pointer.gap-2
                span.label-text Pasirinkimas
                input.toggle.toggle-primary.toggle-sm(type="checkbox" v-model="rawSelectionEnabled")
            
            button.btn.btn-sm.btn-outline(
              :disabled="selectedRawRecords.length === 0"
              @click="copyRawToBasket"
            )
              svg.w-4.h-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3")
              | ({{ selectedRawRecords.length }})

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
                th(v-if="rawSelectionEnabled")
                  input.checkbox.checkbox-xs(type="checkbox" @change="toggleAllRaw" :checked="allRawSelected")
                th Stotis
                th Data
                th Traurinys (atv/išv)
                th Atv./Išv.
                th Riedmuo (atv/išv)
                th Atv.Maršrutas
                th Išv.Maršrutas
            tbody
              tr(v-for="record in paginatedRecords" :key="record.id")
                td(v-if="rawSelectionEnabled")
                  input.checkbox.checkbox-xs(
                    type="checkbox"
                    :checked="isRawSelected(record)"
                    @change="toggleRawSelection(record)"
                  )
                td.font-mono.text-xs {{ record.sheetName }}
                td.text-xs {{ record.arrivalDate || record.departureDate || '-' }}
                td {{ record.trainNoIn || '-' }} / {{ record.trainNoOut || '-' }}
                td {{ formatTime(record.arrivalTime) }} / {{ formatTime(record.departureTime) }}

                td.font-mono.text-xs {{ record.vehicleIn || '-' }} / {{ record.vehicleOut || '-' }}
                td.text-xs {{ record.startingLocationIn || '-' }} - {{ record.endLocationIn || '-' }}
                td.text-xs {{ record.startingLocationOut|| '-' }} - {{ record.endLocationOut || '-' }}

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

  .toast.toast-end(v-if="basket.length > 0")
    .alert.alert-info
      svg.w-5.h-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
        path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z")
      span Krepšelyje: {{ basket.length }} įrašų ({{ basketSourceStation }})
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
const trainSearch = ref('')
const selectedTrainDate = ref('')

// Raw records filters
const rawSelectedSheet = ref('')

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

// Selection state
const selectionEnabled = ref(false)
const selectedRecords = ref([])  // { record, type: 'arrival'|'departure' }

// Raw selection state
const rawSelectionEnabled = ref(false)
const selectedRawRecords = ref([])

// Basket for copy/paste
const basket = ref([])
const basketSourceStation = ref('')

// Row hover state for linked highlighting
const hoveredRecord = ref(null)
const hoveredType = ref(null)

// Computed properties
const isLoading = computed(() => antrasStore.isLoading)
const fieldMappingsLoaded = computed(() => antrasStore.fieldMappingsLoaded)
const canImport = computed(() => antrasStore.fieldMappingsLoaded)
const hasData = computed(() => antrasStore.records.length > 0)
const sheets = computed(() => antrasStore.sheets)
const availableSheets = computed(() => antrasStore.availableSheets)
const availableDates = computed(() => antrasStore.availableDates)
const fileName = computed(() => antrasStore.fileName)
const lastImported = computed(() => antrasStore.lastImported)

// Statistics
const stats = computed(() => antrasStore.getStatistics())

const timelineDate = computed(() => {
  if (selectedDate.value) return selectedDate.value
  if (stationArrivals.value.length > 0) {
    const firstArrival = stationArrivals.value[0]
    if (firstArrival.arrivalDate) return firstArrival.arrivalDate
  }
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
    arrivals = arrivals.filter(arr => arr.arrivalDate === selectedDate.value)
  }
  
  return arrivals.sort((a, b) => (a.arrivalDecimal || 0) - (b.arrivalDecimal || 0))
})

// Station departures filtered by date
const stationDepartures = computed(() => {
  if (!selectedStation.value) return []
  let departures = selectedStation.value.departures
  
  if (selectedDate.value) {
    departures = departures.filter(dep => dep.departureDate === selectedDate.value)
  }
  
  return departures.sort((a, b) => (a.departureDecimal || 0) - (b.departureDecimal || 0))
})

// Check if can paste (not same station)
const canPaste = computed(() => {
  return basket.value.length > 0 && basketSourceStation.value !== selectedSheet.value
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

// Filtered staff by search - only drivers (M)
const filteredStaff = computed(() => {
  const search = staffSearch.value.toLowerCase().trim()
  let staff = antrasStore.staff.filter(s => s.occ === 'M')
  
  if (!search) return staff
  
  return staff.filter(p => 
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

// Filtered trains
const filteredTrains = computed(() => {
  let trains = []
  
  // Collect all trains with their dates
  Object.entries(antrasStore.trains).forEach(([date, dayTrains]) => {
    dayTrains.forEach(train => {
      trains.push({ ...train, date })
    })
  })
  
  // Filter by date
  if (selectedTrainDate.value) {
    trains = trains.filter(t => t.date === selectedTrainDate.value)
  }
  
  // Filter by train number
  if (trainSearch.value) {
    const search = trainSearch.value.toLowerCase()
    trains = trains.filter(t => String(t.no).includes(search))
  }
  
  return trains.sort((a, b) => {
    if (a.date !== b.date) return a.date.localeCompare(b.date)
    return a.no - b.no
  })
})

// Raw records filtered by selected sheet
const filteredRawRecords = computed(() => {
  let records = antrasStore.records
  
  if (rawSelectedSheet.value) {
    records = records.filter(r => r.sheetName === rawSelectedSheet.value)
  }
  
  if (selectedDate.value) {
    records = records.filter(r => {
      return r.arrivalDate === selectedDate.value || r.departureDate === selectedDate.value
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

// Selection helpers
const allArrivalsSelected = computed(() => {
  if (stationArrivals.value.length === 0) return false
  return stationArrivals.value.every(arr => isSelected(arr))
})

const allDeparturesSelected = computed(() => {
  if (stationDepartures.value.length === 0) return false
  return stationDepartures.value.every(dep => isSelected(dep))
})

const allRawSelected = computed(() => {
  if (paginatedRecords.value.length === 0) return false
  return paginatedRecords.value.every(rec => isRawSelected(rec))
})

/**
 * Extract globalRowId from record ID
 */
const extractRowId = (id) => {
  if (!id) return null
  return id.split('---')[1]
}

/**
 * Get row CSS class with highlighting
 */
const getRowClass = (record, type) => {
  const classes = []
  
  // Selection highlight
  if (isSelected(record)) {
    classes.push('bg-primary/20')
  }
  
  // Linked row highlight by globalRowId
  if (hoveredRecord.value && hoveredRecord.value.id) {
    const hoveredRowId = extractRowId(hoveredRecord.value.id)
    const currentRowId = extractRowId(record.id)
    
    // Highlight if this is the hovered row
    if (record.id === hoveredRecord.value.id) {
      classes.push('bg-warning/30')
    }
    // Highlight linked row in opposite table (same globalRowId, different type)
    else if (hoveredRowId && currentRowId && hoveredRowId === currentRowId && hoveredType.value !== type) {
      classes.push('bg-info/30')
    }
  }
  
  return classes.join(' ')
}
/**
 * Handle row hover for linked highlighting
 */
const handleRowHover = (record, type) => {
  hoveredRecord.value = record
  hoveredType.value = type
}

const handleRowLeave = () => {
  hoveredRecord.value = null
  hoveredType.value = null
}

/**
 * Check if record is selected
 */
const isSelected = (record) => {
  return selectedRecords.value.some(sr => sr.record.id === record.id)
}

const isRawSelected = (record) => {
  return selectedRawRecords.value.some(r => r.id === record.id)
}

/**
 * Toggle selection
 */
const toggleSelection = (record, type) => {
  const idx = selectedRecords.value.findIndex(sr => sr.record.id === record.id)
  if (idx >= 0) {
    selectedRecords.value.splice(idx, 1)
  } else {
    selectedRecords.value.push({ record, type })
  }
}

const toggleRawSelection = (record) => {
  const idx = selectedRawRecords.value.findIndex(r => r.id === record.id)
  if (idx >= 0) {
    selectedRawRecords.value.splice(idx, 1)
  } else {
    selectedRawRecords.value.push(record)
  }
}

const toggleAllArrivals = () => {
  if (allArrivalsSelected.value) {
    selectedRecords.value = selectedRecords.value.filter(sr => sr.type !== 'arrival')
  } else {
    stationArrivals.value.forEach(arr => {
      if (!isSelected(arr)) {
        selectedRecords.value.push({ record: arr, type: 'arrival' })
      }
    })
  }
}

const toggleAllDepartures = () => {
  if (allDeparturesSelected.value) {
    selectedRecords.value = selectedRecords.value.filter(sr => sr.type !== 'departure')
  } else {
    stationDepartures.value.forEach(dep => {
      if (!isSelected(dep)) {
        selectedRecords.value.push({ record: dep, type: 'departure' })
      }
    })
  }
}

const toggleAllRaw = () => {
  if (allRawSelected.value) {
    const pageIds = new Set(paginatedRecords.value.map(r => r.id))
    selectedRawRecords.value = selectedRawRecords.value.filter(r => !pageIds.has(r.id))
  } else {
    paginatedRecords.value.forEach(rec => {
      if (!isRawSelected(rec)) {
        selectedRawRecords.value.push(rec)
      }
    })
  }
}

/**
 * Copy selected records to basket
 */
const copyToBasket = () => {
  if (selectedRecords.value.length === 0) return
  
  basket.value = selectedRecords.value.map(sr => ({
    ...sr.record,
    _type: sr.type,
    _sourceStation: selectedSheet.value
  }))
  basketSourceStation.value = selectedSheet.value
  
  slogStore.addToast({
    message: `${basket.value.length} įrašų nukopijuota į krepšelį`,
    type: 'alert-success'
  })
  
  loggingStore.info('Įrašai nukopijuoti į krepšelį', {
    component: 'Second',
    count: basket.value.length,
    sourceStation: selectedSheet.value
  })
  
  // Clear selection after copy
  selectedRecords.value = []
}

const copyRawToBasket = () => {
  if (selectedRawRecords.value.length === 0) return
  
  basket.value = selectedRawRecords.value.map(rec => ({
    ...rec,
    _type: 'raw',
    _sourceStation: rec.sheetName
  }))
  basketSourceStation.value = selectedRawRecords.value[0]?.sheetName || ''
  
  slogStore.addToast({
    message: `${basket.value.length} įrašų nukopijuota į krepšelį`,
    type: 'alert-success'
  })
  
  selectedRawRecords.value = []
}

/**
 * Paste records from basket to current station
 */
const pasteFromBasket = () => {
  if (!canPaste.value) {
    slogStore.addToast({
      message: 'Negalima įklijuoti į tą pačią stotį',
      type: 'alert-warning'
    })
    return
  }
  
  const station = selectedStation.value
  if (!station) return
  
  let arrivals = 0
  let departures = 0
  
  basket.value.forEach(item => {
    const newId = `pasted-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    
    if (item._type === 'arrival') {
      station.arrivals.push({
        ...item,
        id: newId,
        targetTrack: item._sourceStation,  // Show source as track
        _pasted: true
      })
      arrivals++
    } else if (item._type === 'departure') {
      station.departures.push({
        ...item,
        id: newId,
        startingTrack: item._sourceStation,  // Show source as track
        _pasted: true
      })
      departures++
    }
  })
  
  clearBasket()

  slogStore.addToast({
    message: `Įklijuota: ${arrivals} atvykimų, ${departures} išvykimų`,
    type: 'alert-success'
  })
  
  loggingStore.info('Įrašai įklijuoti iš krepšelio', {
    component: 'Second',
    arrivals,
    departures,
    targetStation: selectedSheet.value,
    sourceStation: basketSourceStation.value
  })
}

/**
 * Clear basket
 */
const clearBasket = () => {
  basket.value = []
  basketSourceStation.value = ''
  
  slogStore.addToast({
    message: 'Krepšelis išvalytas',
    type: 'alert-info'
  })
}

/**
 * Get driver names (only occ='M') from staff array
 */
const getDriverNames = (staffList) => {
  if (!staffList || staffList.length === 0) return '-'
  
  const drivers = staffList.filter(s => s.occ === 'M')
  if (drivers.length === 0) return '-'
  
  return drivers.map(d => {
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
    selectedSheet.value = ''
    selectedDate.value = ''
    currentPage.value = 1
    activeTab.value = 'stations'
    selectedRecords.value = []
    selectedRawRecords.value = []
    basket.value = []
    
    if (availableSheets.value.length > 0) {
      selectedSheet.value = availableSheets.value[0]
    }
  }

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
    case 'R': return 'Rezervas'
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
    case 'R': return 'badge-accent'
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

  if (clipStore.stationsAvailable) {
    const station = clipStore.stations.find(s => s.code === selectedSheet.value)
    if (station && station.tracks && station.tracks.length > 0) {
      tracks.value = station.tracks.sort((a, b) => a.id - b.id)
      tracksLoaded.value = true
      
      loggingStore.info('Keliai rasti stotyje', {
        component: 'Second',
        station: selectedSheet.value,
        tracksCount: tracks.value.length
      })
    } else {
      tracks.value = []
      tracksLoaded.value = false
    }
  }
}

/**
 * Load stations from API
 */
const loadStations = async () => {
  try {
    await clipStore.loadStations()
    loggingStore.info('Stočių duomenys įkelti', {
      component: 'Second',
      stationsCount: clipStore.stations.length
    })
  } catch (error) {
    loggingStore.error('Klaida įkeliant stočių duomenis', {
      component: 'Second',
      error: error.message
    })
  }
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
  selectedRecords.value = []
  selectedRawRecords.value = []
  basket.value = []
  basketSourceStation.value = ''
  
  slogStore.addToast({
    message: 'Duomenys išvalyti',
    type: 'alert-info'
  })
  
  loggingStore.info('Antras duomenys išvalyti', {
    component: 'Second',
    action: 'clear_data'
  })
}

/**
 * Export data to JSON
 */
const exportToJson = () => {
  antrasStore.exportToJson()
  
  slogStore.addToast({
    message: 'Duomenys eksportuoti į JSON',
    type: 'alert-success'
  })
}

// Watch filters
watch([selectedSheet, selectedDate], () => {
  antrasStore.setFilters({
    sheet: selectedSheet.value,
    date: selectedDate.value
  })
  currentPage.value = 1
  selectedRecords.value = []
})

// Watch selected sheet to load tracks
watch(selectedSheet, () => {
  loadTracksForStation()
})

// Load field mappings on mount
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

/* Row hover transition */
tbody tr {
  transition: background-color 0.15s ease;
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
  background: #f55;
  border-radius: 8px 0 0 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.timeline-tab:hover {
  background: #f44;
}

.timeline-tab-active {
  background: #f77;
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
  .timeline-drawer-wrapper,
  .toast {
    display: none !important;
  }
}
</style>
