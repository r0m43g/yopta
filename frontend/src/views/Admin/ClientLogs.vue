<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .card.bg-base-200(class="w-11/12 h-11/12 shadow-xl overflow-auto scrollbar-thin")
    .card-body
      h1.card-title.text-3xl.mb-6 Klientų veiksmų žurnalas

      .flex.flex-col.space-y-6
        .card.bg-base-100.shadow-md(v-if="showStatistics")
          .card-body.p-4
            h2.card-title.text-xl.mb-4.flex.items-center
              svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z")
              | Klientų žurnalų statistika

            .stats.shadow
              .stat
                .stat-figure.text-primary
                  svg.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z")
                .stat-title Visi žurnalai
                .stat-value.text-primary {{ statistics.total_logs || 0 }}
                .stat-desc Per pasirinktą laikotarpį

              .stat
                .stat-figure.text-error
                  svg.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z")
                .stat-title Klaidos
                .stat-value.text-error {{ statistics.error_count || 0 }}
                .stat-desc Klaidų procentas: {{ (statistics.error_rate || 0).toFixed(2) }}%

            .grid-cols-2.gap-4.mt-4
              .card.bg-base-200.shadow-sm
                .card-body.p-4
                  h3.card-title.text-lg Registravimo lygiai
                  .overflow-x-auto
                    table.table.table-zebra.table-sm
                      thead
                        tr
                          th Lygis
                          th Kiekis
                      tbody
                        tr(v-for="(item, index) in statistics.level_stats" :key="'level-'+index")
                          td
                            span(:class="getLevelClass(item.level)") {{ item.level }}
                          td {{ item.count }}

              .card.bg-base-200.shadow-sm
                .card-body.p-4
                  h3.card-title.text-lg Populiarūs keliai
                  .overflow-x-auto
                    table.table.table-zebra.table-sm
                      thead
                        tr
                          th Kelias
                          th Kiekis
                      tbody
                        tr(v-for="(item, index) in statistics.path_stats" :key="'path-'+index")
                          td.truncate.max-w-xs {{ item.path }}
                          td {{ item.count }}

            .mt-4
              .card.bg-base-200.shadow-sm
                .card-body.p-4
                  h3.card-title.text-lg Aktyvūs vartotojai
                  .overflow-x-auto
                    table.table.table-zebra.table-sm
                      thead
                        tr
                          th Vartotojo ID
                          th Vartotojo vardas
                          th Žurnalų kiekis
                      tbody
                        tr(v-for="(item, index) in statistics.user_stats" :key="'user-'+index")
                          td {{ item.user_id }}
                          td {{ item.username || 'Nežinomas' }}
                          td {{ item.count }}

        .card.bg-base-100.shadow-md
          .card-body.p-4
            .flex.flex-row.justify-between.items-center.mb-4
              h2.card-title.text-xl.mb-0.flex.items-center
                svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z")
                | Kliento žurnalai

              .flex.flex-wrap.justify-end.gap-2
                button.btn.btn-sm.btn-primary(@click="toggleStatistics")
                  | {{ showStatistics ? 'Slėpti statistiką' : 'Rodyti statistiką' }}
                button.btn.btn-sm.btn-outline(@click="loadStatistics") Atnaujinti statistiką
                button.btn.btn-sm.btn-error(@click="openClearLogsModal") Išvalyti senus žurnalus

            .collapse.collapse-arrow.border.border-base-300.bg-base-200.rounded-box.mb-4
              input(type="checkbox")
              .collapse-title.text-lg.font-medium Filtrai
              .collapse-content
                .grid.grid-cols-3.gap-4
                  .form-control
                    label.label
                      span.label-text Lygis
                    select.select.select-bordered.w-full(v-model="filters.level")
                      option(value="") Visi lygiai
                      option(value="debug") Debug
                      option(value="info") Info
                      option(value="warn") Warn
                      option(value="error") Error

                  .form-control
                    label.label
                      span.label-text Pranešimas
                    input.input.input-bordered.w-full(type="text" v-model="filters.message" placeholder="Ieškoti pranešime")

                  .form-control
                    label.label
                      span.label-text Kelias
                    input.input.input-bordered.w-full(type="text" v-model="filters.path" placeholder="Pvz.: /profile")

                  .form-control
                    label.label
                      span.label-text Vartotojo ID
                    input.input.input-bordered.w-full(type="text" v-model="filters.user_id" placeholder="Vartotojo ID")

                  .form-control
                    label.label
                      span.label-text IP adresas
                    input.input.input-bordered.w-full(type="text" v-model="filters.ip_address" placeholder="Pvz.: 192.168.1.1")

                  .form-control
                    label.label
                      span.label-text Laikotarpis
                    .flex.gap-2
                      input.input.input-bordered.w-full(type="date" v-model="filters.from")
                      input.input.input-bordered.w-full(type="date" v-model="filters.to")

                .flex.justify-end.mt-4
                  button.btn.btn-sm.mr-2(@click="resetFilters") Atstatyti
                  button.btn.btn-sm.btn-primary(@click="applyFilters") Taikyti

            .overflow-x-auto
              table.table.table-zebra
                thead
                  tr
                    th ID
                    th Laikas
                    th Lygis
                    th.table-cell Pranešimas
                    th.table-cell Vartotojas
                    th.table-cell Kelias
                    th Veiksmai
                tbody
                  tr.hover(v-for="log in logs" :key="log.id")
                    td {{ log.id }}
                    td.whitespace-nowrap {{ formatDate(log.client_timestamp) }}
                    td
                      span.badge(:class="getLevelClass(log.level)") {{ log.level }}
                    td.table-cell.truncate.max-w-xs {{ log.message }}
                    td.table-cell
                      span(v-if="log.user_id") {{ log.user_id }}
                      span(v-else) Svečias
                    td.table-cell.truncate.max-w-xs {{ log.path || log.route || '-' }}
                    td
                      button.btn.btn-sm.btn-ghost(@click="viewLogDetails(log)")
                        svg.h-4.w-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z")
                          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z")

            div.text-center.py-6(v-if="logs.length === 0")
              .text-2xl.mb-2 📋
              p Kliento žurnalų nerasta

            .flex.justify-between.items-center.mt-4
              .text-sm.text-gray-500 Viso: {{ totalLogs }} įrašų
              .join
                button.join-item.btn.btn-sm(:disabled="currentPage === 1" @click="changePage(currentPage - 1)") «
                button.join-item.btn.btn-sm Puslapis {{ currentPage }} iš {{ totalPages }}
                button.join-item.btn.btn-sm(:disabled="currentPage === totalPages" @click="changePage(currentPage + 1)") »
              .flex.items-center
                span.mr-2.text-sm Rodyti:
                select.select.select-bordered.select-sm(v-model="pageSize" @change="changePageSize")
                  option(:value="10") 10
                  option(:value="20") 20
                  option(:value="50") 50
                  option(:value="100") 100

  .modal(:class="{'modal-open': selectedLog !== null}")
    .modal-box.w-11/12.max-w-5xl
      h3.font-bold.text-lg Kliento žurnalo detalės

      .py-4(v-if="selectedLog")
        .grid.grid-cols-2.gap-4.mb-4
          .form-control
            label.floating-label
              span.label-text.font-semibold Lygis
              input.input.input-bordered.w-full(:value="selectedLog.level" readonly)

          .form-control
            label.floating-label
              span.label-text.font-semibold Kliento laiko žyma
              input.input.input-bordered.w-full(:value="selectedLog.client_timestamp" readonly)

          .form-control
            label.floating-label
              span.label-text.font-semibold Serverio laiko žyma
              input.input.input-bordered.w-full(:value="selectedLog.server_timestamp" readonly)

          .form-control
            label.floating-label
              span.label-text.font-semibold IP adresas
              input.input.input-bordered.w-full(:value="selectedLog.ip_address" readonly)

        .form-control.mb-4
          label.floating-label
            span.label-text.font-semibold Pranešimas
            textarea.textarea.textarea-bordered.h-12.w-full(readonly) {{ selectedLog.message }}

        .form-control.mb-4
          label.floating-label
            span.label-text.font-semibold URL
            input.input.input-bordered.w-full(:value="selectedLog.url" readonly)

        .form-control.mb-4
          label.floating-label
            span.label-text.font-semibold Kelias
            input.input.input-bordered.w-full(:value="selectedLog.path" readonly)

        .form-control.mb-4
          label.floating-label
            span.label-text.font-semibold Papildomi duomenys
            textarea.textarea.textarea-bordered.w-full.h-48.font-mono.text-sm(readonly) {{ formatJSON(selectedLog.data) }}

        .form-control(v-if="selectedLog.user_agent")
          label.floating-label
            span.label-text.font-semibold Vartotojo naršyklė
            textarea.textarea.textarea-bordered.w-full.h-16(readonly) {{ selectedLog.user_agent }}

      .modal-action
        button.btn(@click="selectedLog = null") Uždaryti

  .modal(:class="{'modal-open': showClearLogsModal}")
    .modal-box
      h3.font-bold.text-lg Senų kliento žurnalų išvalymas
      .py-4
        p.mb-4 Pasirinkite laikotarpį senų kliento žurnalų išvalymui:

        .form-control.mb-4
          label.label
            span.label-text Ištrinti įrašus senesnius nei
          select.select.select-bordered.w-full(v-model="clearLogsOption")
            option(value="7") 7 dienos
            option(value="30") 30 dienų
            option(value="90") 90 dienų
            option(value="180") 6 mėnesiai
            option(value="365") 1 metai
            option(value="custom") Pasirinkti datą...

        .form-control.mb-4(v-if="clearLogsOption === 'custom'")
          label.label
            span.label-text Data (ištrinti įrašus iki šios datos)
          input.input.input-bordered.w-full(type="date" v-model="clearLogsCustomDate")

        .bg-warning.text-warning-content.p-4.rounded-lg
          .flex.items-start
            svg.h-6.w-6.mr-2.mt-1.flex-shrink-0(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
              path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z")
            p
              span.font-bold Dėmesio:
              |  Šis veiksmas negrįžtamas. Ištrintų žurnalų įrašų nebus galima atkurti.

      .modal-action
        button.btn(@click="showClearLogsModal = false") Atšaukti
        button.btn.btn-error(@click="clearLogs" :disabled="clearLogsLoading")
          span.loading.loading-spinner.loading-sm(v-if="clearLogsLoading")
          | Ištrinti įrašus
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useSlogStore } from '@/stores/slog'
import { useLoggingStore } from '@/stores/logging'
import api from '@/services/api'

// Component state
const logs = ref([])
const totalLogs = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const totalPages = computed(() => Math.max(1, Math.ceil(totalLogs.value / pageSize.value)))

const selectedLog = ref(null)
const showStatistics = ref(false)
const statistics = ref({})

const showClearLogsModal = ref(false)
const clearLogsOption = ref('30')
const clearLogsCustomDate = ref('')
const clearLogsLoading = ref(false)

// Filter state
const filters = ref({
  level: '',
  message: '',
  path: '',
  user_id: '',
  ip_address: '',
  from: '',
  to: ''
})

const slogStore = useSlogStore()
const loggingStore = useLoggingStore()

/**
 * Loads client logs with pagination and filters
 * Fetches log data from API based on current page and filter settings
 */
const loadLogs = async () => {
  try {
    // Prepare query parameters
    const params = {
      page: currentPage.value,
      limit: pageSize.value
    }

    // Add active filters to request
    Object.keys(filters.value).forEach(key => {
      if (filters.value[key]) {
        params[key] = filters.value[key]
      }
    })

    const response = await api.get('/client-logs', { params })
    logs.value = response.data.data
    totalLogs.value = response.data.total
  } catch (error) {
    slogStore.addToast({
      type: 'alert-error',
      message: 'Klaida įkeliant kliento žurnalus: ' + (error.response?.data || error.message)
    })
  }
}

/**
 * Loads statistics data for client logs
 * Updates statistics panel with aggregate metrics
 */
const loadStatistics = async () => {
  try {
    // Prepare query parameters for statistics
    const params = {}

    if (filters.value.from) {
      params.from = filters.value.from
    }

    if (filters.value.to) {
      params.to = filters.value.to
    }

    const response = await api.get('/client-logs/statistics', { params })
    statistics.value = response.data

    if (!showStatistics.value) {
      showStatistics.value = true
    }
  } catch (error) {
    slogStore.addToast({
      type: 'alert-error',
      message: 'Klaida įkeliant statistiką: ' + (error.response?.data || error.message)
    })
  }
}

/**
 * Toggles visibility of statistics panel
 * Loads statistics data if panel is being shown
 */
const toggleStatistics = () => {
  showStatistics.value = !showStatistics.value
  if (showStatistics.value && Object.keys(statistics.value).length === 0) {
    loadStatistics()
  }
}

/**
 * Shows log details in modal
 * @param {Object} log - Log entry to display
 */
const viewLogDetails = (log) => {
  selectedLog.value = log
}

/**
 * Formats date string into localized format
 * @param {string} dateString - ISO date string
 * @returns {string} Formatted date
 */
const formatDate = (dateString) => {
  if (!dateString) return ''

  try {
    const date = new Date(dateString)

    // Check if date is valid
    if (isNaN(date.getTime())) {
      return dateString
    }

    return date.toLocaleString('lt-LT')
  } catch (e) {
    return dateString
  }
}

/**
 * Formats JSON data for display
 * @param {Object|string} jsonData - JSON data to format
 * @returns {string} Pretty-printed JSON string
 */
const formatJSON = (jsonData) => {
  if (!jsonData) return ''

  try {
    if (typeof jsonData === 'string') {
      return JSON.stringify(JSON.parse(jsonData), null, 2)
    }
    return JSON.stringify(jsonData, null, 2)
  } catch (e) {
    return typeof jsonData === 'string' ? jsonData : JSON.stringify(jsonData)
  }
}

/**
 * Returns CSS class for log level badge
 * @param {string} level - Log level (debug, info, warn, error)
 * @returns {string} CSS class name
 */
const getLevelClass = (level) => {
  switch (level?.toLowerCase()) {
    case 'debug': return 'badge-primary'
    case 'info': return 'badge-info'
    case 'warn': return 'badge-warning'
    case 'error': return 'badge-error'
    default: return 'badge-ghost'
  }
}

/**
 * Changes current page of logs
 * @param {number} page - Page number to navigate to
 */
const changePage = (page) => {
  currentPage.value = page
  loadLogs()
}

/**
 * Updates number of items shown per page
 * Resets to first page when changing page size
 */
const changePageSize = () => {
  currentPage.value = 1
  loadLogs()
}

/**
 * Applies current filters to logs
 * Resets to first page when applying filters
 */
const applyFilters = () => {
  currentPage.value = 1
  loadLogs()

  // Update statistics if visible
  if (showStatistics.value) {
    loadStatistics()
  }
}

/**
 * Resets all filters to empty state
 * Reloads data with cleared filters
 */
const resetFilters = () => {
  Object.keys(filters.value).forEach(key => {
    filters.value[key] = ''
  })

  currentPage.value = 1
  loadLogs()

  // Update statistics if visible
  if (showStatistics.value) {
    loadStatistics()
  }
}

/**
 * Opens the modal for clearing old logs
 * Resets modal form to default state
 */
const openClearLogsModal = () => {
  showClearLogsModal.value = true
  clearLogsOption.value = '30'
  clearLogsCustomDate.value = ''
}

/**
 * Clears old logs based on selected timeframe
 * Sends delete request with calculated before date
 */
const clearLogs = async () => {
  clearLogsLoading.value = true

  try {
    let beforeDate

    if (clearLogsOption.value === 'custom') {
      if (!clearLogsCustomDate.value) {
        slogStore.addToast({
          type: 'alert-error',
          message: 'Prašome pasirinkti datą'
        })
        clearLogsLoading.value = false
        return
      }
      beforeDate = new Date(clearLogsCustomDate.value).toISOString()
    } else {
      // Calculate date based on selected period (in days)
      const days = parseInt(clearLogsOption.value)
      beforeDate = new Date(Date.now() - days * 24 * 60 * 60 * 1000).toISOString()
    }

    const response = await api.delete('/client-logs', {
      params: { before: beforeDate }
    })

    showClearLogsModal.value = false

    slogStore.addToast({
      type: 'alert-success',
      message: `Ištrinta ${response.data.deleted} žurnalų įrašų`
    })

    // Update data
    loadLogs()
    if (showStatistics.value) {
      loadStatistics()
    }
  } catch (error) {
    slogStore.addToast({
      type: 'alert-error',
      message: 'Klaida trinant įrašus: ' + (error.response?.data || error.message)
    })
  } finally {
    clearLogsLoading.value = false
  }
}

// Load data when component is mounted
onMounted(() => {
  loadLogs()
})
</script>

<style>
.badge {
  text-transform: uppercase;
}

.modal-box {
  max-height: 85vh;
  overflow-y: auto;
}
</style>
