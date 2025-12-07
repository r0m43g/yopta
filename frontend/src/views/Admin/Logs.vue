<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .card.bg-base-200(class="w-11/12 h-11/12 shadow-xl overflow-auto scrollbar-thin")
    .card-body
      h1.card-title.text-3xl.mb-6 Sistemos žurnalų peržiūra
      .flex.flex-col.space-y-6
        .card.bg-base-100.shadow-md(v-if="showStatistics")
          .card-body.p-4
            h2.card-title.text-xl.mb-4.flex.items-center
              svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z")
              | Statistika

            .stats.shadow
              .stat
                .stat-figure.text-primary
                  svg.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z")
                .stat-title Užklausos
                .stat-value.text-primary {{ statistics.total_requests || 0 }}
                .stat-desc Pasirinktam laikotarpiui

              .stat
                .stat-figure.text-error
                  svg.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z")
                .stat-title Klaidos
                .stat-value.text-error {{ statistics.total_errors || 0 }}
                .stat-desc Klaidų procentas: {{ (statistics.error_rate || 0).toFixed(2) }}%

              .stat
                .stat-figure.text-accent
                  svg.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z")
                .stat-title Vidutinis laikas
                .stat-value.text-accent {{ (statistics.avg_duration_ms || 0).toFixed(2) }}ms
                .stat-desc Užklausos apdorojimo trukmė

            .grid.grid-cols-2.gap-4.mt-4
              .card.bg-base-200.shadow-sm
                .card-body.p-4
                  h3.card-title.text-lg HTTP metodai
                  .overflow-x-auto
                    table.table.table-zebra.table-sm
                      thead
                        tr
                          th Metodas
                          th Kiekis
                      tbody
                        tr(v-for="(item, index) in statistics.method_stats" :key="'method-'+index")
                          td {{ item.method }}
                          td {{ item.count }}

              .card.bg-base-200.shadow-sm
                .card-body.p-4
                  h3.card-title.text-lg Būsenos kodai
                  .overflow-x-auto
                    table.table.table-zebra.table-sm
                      thead
                        tr
                          th Būsena
                          th Kiekis
                      tbody
                        tr(v-for="(item, index) in statistics.status_stats" :key="'status-'+index")
                          td
                            span(:class="getStatusClass(item.status)") {{ item.status }}
                          td {{ item.count }}

            .grid.grid-cols-1(class="md:grid-cols-2 gap-4 mt-4")
              .card.bg-base-200.shadow-sm
                .card-body.p-4
                  h3.card-title.text-lg Dažniausiai užklausiami keliai
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

              .card.bg-base-200.shadow-sm
                .card-body.p-4
                  h3.card-title.text-lg Aktyviausi vartotojai
                  .overflow-x-auto
                    table.table.table-zebra.table-sm
                      thead
                        tr
                          th Vartotojo ID
                          th Veiksmų skaičius
                      tbody
                        tr(v-for="(item, index) in statistics.user_stats" :key="'user-'+index")
                          td {{ item.user_id }}
                          td {{ item.count }}

        .card.bg-base-100.shadow-md
          .card-body.p-4
            .flex.flex-row.justify-between.items-center.mb-4
              h2.card-title.text-xl.mb-0.flex.items-center
                svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z")
                | Sistemos žurnalas

              .flex.flex-wrap.justify-end.gap-2
                button.btn.btn-sm.btn-primary(@click="toggleStatistics") {{ showStatistics ? 'Slėpti statistiką' : 'Rodyti statistiką' }}
                button.btn.btn-sm.btn-outline(@click="loadStatistics") Atnaujinti statistiką
                button.btn.btn-sm.btn-error(@click="openClearLogsModal") Išvalyti senus įrašus

            .collapse.collapse-arrow.border.border-base-300.bg-base-200.rounded-box.mb-4
              input(type="checkbox")
              .collapse-title.text-lg.font-medium Filtrai
              .collapse-content
                .grid.grid-cols-3.gap-4
                  .form-control
                    label.label
                      span.label-text Metodas
                    select.select.select-bordered.w-full(v-model="filters.method")
                      option(value="") Visi metodai
                      option(value="GET") GET
                      option(value="POST") POST
                      option(value="PUT") PUT
                      option(value="DELETE") DELETE

                  .form-control
                    label.label
                      span.label-text Kelias
                    input.input.input-bordered.w-full(type="text" v-model="filters.path" placeholder="Pvz.: /api/v1/users")

                  .form-control
                    label.label
                      span.label-text Būsenos kodas
                    select.select.select-bordered.w-full(v-model="filters.status_code")
                      option(value="") Visi statusai
                      option(value="200") 200 (OK)
                      option(value="201") 201 (Created)
                      option(value="400") 400 (Bad Request)
                      option(value="401") 401 (Unauthorized)
                      option(value="403") 403 (Forbidden)
                      option(value="404") 404 (Not Found)
                      option(value="500") 500 (Server Error)

                  .form-control
                    label.label
                      span.label-text Vartotojo ID
                    input.input.input-bordered.w-full(type="text" v-model="filters.user_id" placeholder="Vartotojo ID")

                  .form-control
                    label.label
                      span.label-text IP adresas
                    input.input.input-bordered.w-full(type="text" v-model="filters.client_ip" placeholder="Pvz.: 192.168.1.1")

                  .form-control
                    label.label
                      span.label-text Laikotarpis
                    .flex.gap-2
                      input.input.input-bordered.w-full(type="date" v-model="filters.from_date")
                      input.input.input-bordered.w-full(type="date" v-model="filters.to_date")

                .flex.justify-end.mt-4
                  button.btn.btn-sm.mr-2(@click="resetFilters") Atstatyti
                  button.btn.btn-sm.btn-primary(@click="applyFilters") Taikyti

            .overflow-x-auto
              table.table.table-zebra
                thead
                  tr
                    th.table-cell ID
                    th Laikas
                    th Metodas
                    th.table-cell Kelias
                    th.table-cell Vartotojas
                    th Būsena
                    th.table-cell Trukmė (ms)
                    th Veiksmai
                tbody
                  tr.hover(v-for="log in logs" :key="log.id")
                    td.table-cell {{ log.id }}
                    td {{ formatDate(log.timestamp) }}
                    td
                      span.badge(:class="getMethodClass(log.method)") {{ log.method }}
                    td.table-cell.truncate.max-w-xs {{ log.path }}
                    td.table-cell
                      span(v-if="log.user_id !== null && log.user_id !== undefined") {{ log.user_id }}
                      span(v-else) Svečias
                    td
                      span.badge(:class="getStatusClass(log.status_code)") {{ log.status_code }}
                    td.table-cell {{ log.duration_ms }}
                    td
                      button.btn.btn-sm.btn-ghost(@click="viewLogDetails(log)")
                        svg.h-4.w-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z")
                          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z")

            div.text-center.py-6(v-if="logs.length === 0")
              .text-2xl.mb-2 📋
              p Žurnale įrašų nerasta

            .flex.justify-between.items-center.mt-4
              .text-sm.text-gray-500 Viso: {{ totalLogs }} įrašų
              .join
                button.join-item.btn.btn-sm(:disabled="currentPage === 1" @click="changePage(currentPage - 1)") «
                button.join-item.btn.btn-sm Puslapis {{ currentPage }} iš {{ totalPages }}
                button.join-item.btn.btn-sm(:disabled="currentPage === totalPages" @click="changePage(currentPage + 1)") »
              .flex.items-center
                span.mr-2.text-sm Rodyti po:
                select.select.select-bordered.select-sm(v-model="pageSize" @change="changePageSize")
                  option(:value="10") 10
                  option(:value="20") 20
                  option(:value="50") 50
                  option(:value="100") 100

  .modal(:class="{'modal-open': selectedLog !== null}")
    .modal-box(class="w-11/12.max-w-5xl")
      h3.font-bold.text-lg Žurnalo įrašo informacija
      .py-4(v-if="selectedLog")
        .grid.grid-cols-2.gap-4.mb-4
          .form-control
            label.floating-label
              span.label-text.font-semibold Būsenos kodas
              input.input.input-bordered.w-full(type="text" :value="selectedLog.status_code" readonly)

          .form-control
            label.floating-label
              span.label-text.font-semibold Vykdymo laikas (ms)
              input.input.input-bordered.w-full(type="text" :value="selectedLog.duration_ms" readonly)

          .form-control
            label.floating-label
              span.label-text.font-semibold IP adresas
              input.input.input-bordered.w-full(type="text" :value="selectedLog.client_ip" readonly)

          .form-control
            label.floating-label
              span.label-text.font-semibold Vartotojo ID
              input.input.input-bordered.w-full(type="text" :value="selectedLog.user_id" readonly)

        .form-control.mb-4
          label.floating-label
            span.label-text.font-semibold Užklausos parametrai
            textarea.textarea.textarea-bordered.h-12.w-full(readonly) {{ selectedLog.query || 'Nėra' }}

        .form-control.mb-4
          label.floating-label
            span.label-text.font-semibold Užklausos turinys
            textarea.textarea.textarea-bordered.w-full.h-24.font-mono.text-sm(readonly) {{ formatJSON(selectedLog.request_body) || 'Nėra' }}

        .form-control.mb-4
          label.floating-label
            span.label-text.font-semibold Atsakymo turinys
            textarea.textarea.textarea-bordered.w-full.h-24.font-mono.text-sm(readonly) {{ formatJSON(selectedLog.response_body) || 'Nėra' }}

        .form-control(v-if="selectedLog.error")
          label.floating-label
            span.label-text.font-semibold Klaida
            textarea.textarea.textarea-bordered.w-full.h-16.text-error(readonly) {{ selectedLog.error }}

      .modal-action
        button.btn(@click="selectedLog = null") Uždaryti

  .modal(:class="{'modal-open': showClearLogsModal}")
    .modal-box
      h3.font-bold.text-lg Senų žurnalo įrašų valymas
      .py-4
        p.mb-4 Pasirinkite laikotarpį senų žurnalo įrašų ištrynimui:

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
              |  Šis veiksmas negrįžtamas. Ištrintų įrašų atkurti nebus įmanoma.

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
import { decodeHTMLEntities } from '@/utils/htmlUtils'

// Component state variables
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

// Filters state
const filters = ref({
  method: '',
  path: '',
  status_code: '',
  user_id: '',
  client_ip: '',
  from_date: '',
  to_date: ''
})

// Store instances
const slogStore = useSlogStore()
const loggingStore = useLoggingStore()

/**
 * Loads system logs with current pagination and filters
 */
const loadLogs = async () => {
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value
    }

    Object.keys(filters.value).forEach(key => {
      if (filters.value[key]) {
        params[key] = filters.value[key]
      }
    })

    const response = await api.get('/logs', { params })
    logs.value = response.data.data
    totalLogs.value = response.data.total

    loggingStore.info('Sistemos žurnalai užkrauti', {
      component: 'Logs',
      action: 'load_logs',
      count: response.data.data.length,
      totalCount: response.data.total,
      page: currentPage.value
    })
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida įkeliant žurnalus: ' + (error.response?.data || error.message),
      type: 'alert-error'
    })

    loggingStore.error('Klaida įkeliant sistemos žurnalus', {
      component: 'Logs',
      action: 'load_logs_failed',
      error: error.response?.data || error.message
    })
  }
}

/**
 * Loads statistics data from the server
 */
const loadStatistics = async () => {
  try {
    const params = {}

    if (filters.value.from_date) {
      params.from = filters.value.from_date
    }

    if (filters.value.to_date) {
      params.to = filters.value.to_date
    }

    const response = await api.get('/logs/statistics', { params })
    statistics.value = response.data

    if (!showStatistics.value) {
      showStatistics.value = true
    }

    loggingStore.info('Sistemos žurnalų statistika užkrauta', {
      component: 'Logs',
      action: 'load_statistics',
      from: params.from,
      to: params.to
    })
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida įkeliant statistiką: ' + (error.response?.data || error.message),
      type: 'alert-error'
    })

    loggingStore.error('Klaida įkeliant sistemos žurnalų statistiką', {
      component: 'Logs',
      action: 'load_statistics_failed',
      error: error.response?.data || error.message
    })
  }
}

/**
 * Toggles the visibility of the statistics section
 */
const toggleStatistics = () => {
  showStatistics.value = !showStatistics.value

  if (showStatistics.value && Object.keys(statistics.value).length === 0) {
    loadStatistics()
  }

  loggingStore.info(`Statistikos rodinys ${showStatistics.value ? 'įjungtas' : 'išjungtas'}`, {
    component: 'Logs',
    action: 'toggle_statistics',
    visible: showStatistics.value
  })
}

/**
 * Opens the log details modal with the selected log
 * @param {Object} log - The log entry to display details for
 */
const viewLogDetails = (log) => {
  selectedLog.value = log

  loggingStore.info('Peržiūrimas žurnalo įrašo detales', {
    component: 'Logs',
    action: 'view_log_details',
    logId: log.id,
    logMethod: log.method,
    logPath: log.path
  })
}

/**
 * Formats a date string for display
 * @param {string} dateString - The date string to format
 * @returns {string} - The formatted date string
 */
const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('lt-LT')
}

/**
 * Formats JSON data for human readability
 * @param {string} jsonString - The JSON string to format
 * @returns {string} - The formatted JSON string
 */
const formatJSON = (jsonString) => {
  if (!jsonString) return ''

  try {
    const decodedString = decodeHTMLEntities(jsonString)
    const parsed = JSON.parse(decodedString)
    return JSON.stringify(parsed, null, 2)
  } catch (e) {
    return decodeHTMLEntities(jsonString)
  }
}

/**
 * Returns CSS class based on HTTP method
 * @param {string} method - The HTTP method
 * @returns {string} - The CSS class for the method badge
 */
const getMethodClass = (method) => {
  switch (method) {
    case 'GET': return 'badge-info'
    case 'POST': return 'badge-success'
    case 'PUT': return 'badge-warning'
    case 'DELETE': return 'badge-error'
    default: return 'badge-ghost'
  }
}

/**
 * Returns CSS class based on HTTP status code
 * @param {number} status - The HTTP status code
 * @returns {string} - The CSS class for the status badge
 */
const getStatusClass = (status) => {
  if (status < 300) return 'badge-success'
  if (status < 400) return 'badge-info'
  if (status < 500) return 'badge-warning'
  return 'badge-error'
}

/**
 * Changes the current page of logs
 * @param {number} page - The page number to navigate to
 */
const changePage = (page) => {
  currentPage.value = page
  loadLogs()

  loggingStore.info('Pakeistas žurnalų puslapis', {
    component: 'Logs',
    action: 'change_page',
    page: page
  })
}

/**
 * Changes the number of logs displayed per page
 */
const changePageSize = () => {
  currentPage.value = 1
  loadLogs()

  loggingStore.info('Pakeistas žurnalų puslapio dydis', {
    component: 'Logs',
    action: 'change_page_size',
    pageSize: pageSize.value
  })
}

/**
 * Applies current filters to the logs listing
 */
const applyFilters = () => {
  currentPage.value = 1
  loadLogs()

  if (showStatistics.value) {
    loadStatistics()
  }

  loggingStore.info('Pritaikyti žurnalų filtrai', {
    component: 'Logs',
    action: 'apply_filters',
    filters: JSON.parse(JSON.stringify(filters.value))
  })
}

/**
 * Resets all filters to their default values
 */
const resetFilters = () => {
  Object.keys(filters.value).forEach(key => {
    filters.value[key] = ''
  })

  currentPage.value = 1
  loadLogs()

  if (showStatistics.value) {
    loadStatistics()
  }

  loggingStore.info('Žurnalų filtrai atstatyti', {
    component: 'Logs',
    action: 'reset_filters'
  })
}

/**
 * Opens the modal for clearing old logs
 */
const openClearLogsModal = () => {
  showClearLogsModal.value = true
  clearLogsOption.value = '30'
  clearLogsCustomDate.value = ''

  loggingStore.info('Atidarytas senų žurnalų valymo modalinis langas', {
    component: 'Logs',
    action: 'open_clear_logs_modal'
  })
}

/**
 * Clears old logs based on selected criteria
 */
const clearLogs = async () => {
  clearLogsLoading.value = true

  try {
    let beforeDate

    if (clearLogsOption.value === 'custom') {
      if (!clearLogsCustomDate.value) {
        slogStore.addToast({
          message: 'Nurodykite datą',
          type: 'alert-warning'
        })
        clearLogsLoading.value = false
        return
      }
      beforeDate = new Date(clearLogsCustomDate.value).toISOString()
    } else {
      const days = parseInt(clearLogsOption.value)
      beforeDate = new Date(Date.now() - days * 24 * 60 * 60 * 1000).toISOString()
    }

    const response = await api.delete('/logs', {
      params: { before: beforeDate }
    })

    showClearLogsModal.value = false

    slogStore.addToast({
      message: `Ištrinta ${response.data.deleted_logs} žurnalo įrašų`,
      type: 'alert-success'
    })

    loggingStore.info('Seni žurnalai ištrinti', {
      component: 'Logs',
      action: 'clear_logs',
      deletedCount: response.data.deleted_logs,
      beforeDate: beforeDate
    })

    loadLogs()
    if (showStatistics.value) {
      loadStatistics()
    }
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida trinant įrašus: ' + (error.response?.data || error.message),
      type: 'alert-error'
    })

    loggingStore.error('Klaida trinant senus žurnalus', {
      component: 'Logs',
      action: 'clear_logs_failed',
      error: error.response?.data || error.message,
      option: clearLogsOption.value,
      customDate: clearLogsCustomDate.value
    })
  } finally {
    clearLogsLoading.value = false
  }
}

onMounted(() => {
  loadLogs()

  loggingStore.info('Sistemos žurnalų puslapis užkrautas', {
    component: 'Logs',
    timestamp: new Date().toISOString()
  })
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
