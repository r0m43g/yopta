<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .unprintable.bg-base-100.wide(class="min-h-11/12")
    .qp-4.bg-base-200.border-b.border-base-300
      .flex.flex-col.flex-row.justify-between.items-center.gap-4
        h1.text-2xl.font-bold Riedmenų atvykimai/išvykimai

        .flex.flex-wrap.items-center.gap-2
          select.select.select-bordered.select-sm(
            v-model="selectedDepot"
            :disabled="isLoading || !availableDepots.length"
          )
            option(disabled, value="") Pasirinkite depą
            option(
              v-for="depot in availableDepots"
              :key="depot"
              :value="depot"
            ) {{ depot }}

          select.select.select-bordered.select-sm(
            v-model="selectedDate"
            :disabled="isLoading || !availableDates.length"
          )
            option(disabled, value="") Pasirinkite datą
            option(
              v-for="date in availableDates"
              :key="date"
              :value="date"
            ) {{ date }}

          select.select.select-bordered.select-sm(
            v-model="timeRange"
            :disabled="isLoading"
          )
            option(value="all") Para (00:00-23:59)
            option(value="day") Dieninė pamaina (06:00-20:00)
            option(value="night") Naktinė pamaina (18:00-08:00+1)

          button.btn.btn-primary.btn-sm(
            @click="applyFilters"
            :disabled="isLoading"
          )
            span.loading.loading-spinner.loading-xs(v-if="isLoading")
            | Atnaujinti

    .grid.grid-cols-2.gap-4.p-4
      .card.bg-base-100.shadow-xl
        .card-body.p-4(ref="arrivalsCard")
          h2.card-title.text-xl.copible(@click="copyCardToClipboard(true)")
            strong Atvykimai
            span &nbsp;
              sup.text-xs {{ selectedDate }} {{ selectedDepot }} {{ timeRangeLabel }}

          .overflow-x-auto
            table.table.table-zebra.w-full(v-if="filteredArrivals.length > 0")
              thead
                tr
                  th.th Tr.Nr.
                  th.th Atvykimas
                  th.th Riedmuo
                  th.th Mašinistas
                  th.th Kelias
                  th.th Pastabos
              tbody
                tr(v-for="(item, index) in filteredArrivals" :key="index")
                  td.td {{ item.arrivalTrainNumber }}
                  td.td {{ item.arrivalPlanned}}
                  td.td {{ item.vehicle }}
                  td.td {{ formatName(item.arrivalEmployee1) }}
                  td.td {{ item.targetTrack }}
                  td.td {{ item.notes }}

            .text-center.py-8(v-else)
              .text-4xl.mb-2 🚂
              p.text-lg Nėra atvykimų duomenų pasirinktai datai ir depui

      .card.bg-base-100.shadow-xl
        .card-body.p-4(ref="departuresCard")
          h2.card-title.text-xl.copible(@click="copyCardToClipboard(false)")
            strong Išvykimai
            span &nbsp;
              sup.text-xs {{ selectedDate }} {{ selectedDepot }} {{ timeRangeLabel }}

          .overflow-x-auto
            table.table.table-zebra.w-full(v-if="filteredDepartures.length > 0")
              thead
                tr
                  th.th Tr.Nr.
                  th.th Išvykimas
                  th.th Riedmuo
                  th.th Mašinistas
                  th.th Kelias
                  th.th Pastabos
              tbody
                tr(v-for="(item, index) in filteredDepartures" :key="index")
                  td.td {{ item.departureTrainNumber }}
                  td.td {{ item.departurePlanned }}
                  td.td {{ item.vehicle }}
                  td.td {{ formatName(item.departureEmployee1) }}
                  td.td {{ item.startingTrack }}
                  td.td {{ item.notes }}

            .text-center.py-8(v-else)
              .text-4xl.mb-2 🚂
              p.text-lg Nėra išvykimų duomenų pasirinktai datai ir depui

      .fixed.bottom-4.right-4.flex.flex-col.gap-2.z-50
        .dropdown.dropdown-top.dropdown-end
          button.btn.btn-circle.btn-primary(tabindex="0")
            svg.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
              path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6")
          ul.dropdown-content.menu.p-2.shadow.bg-base-100.rounded-box.w-52(tabindex="0")
            li
              button(@click="handleClipboardImport") Importuoti iš iškarpinės
            li
              button(@click="exportToPDF") Eksportuoti į PDF (spausdinti)

      .timeline-drawer-wrapper(v-if="tracksLoaded && tracks.length > 0"
          :class="{ 'timeline-open': isDrawerOpen }")

        .timeline-drawer
          .timeline-tab(
            @click="toggleDrawer"
              :class="{ 'timeline-tab-active': isDrawerOpen }"
            )
            .timeline-tab-icon
              svg(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7")
          .timeline-drawer-content
            h2.text-xl.font-bold.mb-4 Kelių Grafikas
            TracksTimeLine(
              :selectedDate="selectedDate"
              :selectedDepot="selectedDepot"
              :tracks="tracks"
              :arrivals="allArrivals"
              :departures="allDepartures"
              @track-assigned="handleTrackAssigned"
              @movement-created="handleMovementCreated"
            )

  #print-section
    .print-header.p-4.mb-4
      h1.text-xl.font-bold.mb-1 {{ selectedDepot }} atvykimai/išvykimai
      p {{ selectedDate }} - {{ timeRangeLabel }}

    .print-content
      .mb-6
        h2.text-lg.font-bold.mb-2 Atvykimai ({{ filteredArrivals.length }})
        table.print-table
          thead
            tr
              th Tr.Nr.
              th Atvykimas
              th Riedmuo
              th Mašinistas
              th Kelias
              th Pastabos
          tbody
            tr(v-for="item in filteredArrivals" :key="item.id")
              td {{ item.arrivalTrainNumber }}
              td {{ item.arrivalPlanned}}
              td {{ item.vehicle }}
              td {{ formatName(item.arrivalEmployee1) }}
              td {{ item.targetTrack }}
              td {{ item.notes }}

      div
        h2.text-lg.font-bold.mb-2 Išvykimai ({{ filteredDepartures.length }})
        table.print-table
          thead
            tr
              th Tr.Nr.
              th Išvykimas
              th Riedmuo
              th Mašinistas
              th Kelias
              th Pastabos
          tbody
            tr(v-for="item in filteredDepartures" :key="item.id")
              td {{ item.departureTrainNumber }}
              td {{ item.departurePlanned }}
              td {{ item.vehicle }}
              td {{ formatName(item.departureEmployee1) }}
              td {{ item.startingTrack }}
              td {{ item.notes }}

    .print-footer.mt-4.text-sm
      p Spausdinta: {{ new Date().toLocaleString('lt-LT') }}
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue';
import { useLoggingStore } from '@/stores/logging';
import { useSlogStore } from '@/stores/slog';
import { useClipStore } from '@/stores/clip';
import clip from '@/assets/clip02.js';
import { timeToDecimal } from '@/utils/helpers'
import api from '@/services/api'
import TracksTimeLine from '@/components/parking/TracksTimeLine.vue'

// Store instances initialization
const clipStore = useClipStore();
const loggingStore = useLoggingStore();
const slogStore = useSlogStore();

// UI state management
const isLoading = ref(false);
const selectedDepot = ref('');
const selectedDate = ref('');
const timeRange = ref('all');
const tracks = ref([]);
const tracksLoaded = ref(false);

// Timeline drawer state
const isDrawerOpen = ref(false);
const arrivalsCard = ref(null);
const departuresCard = ref(null);
// Computed properties for data presentation
const availableDepots = computed(() => clipStore.availableDepots);
const availableDates = computed(() => clipStore.availableDates);
const records = computed(() => clipStore.filteredRecords);
const filteredArrivals = computed(() => records.value ? records.value.intimeArrivals : []);
const filteredDepartures = computed(() => records.value ? records.value.intimeDepartures : []);
const allArrivals = computed(() => records.value ? records.value.arrivals : []);
const allDepartures = computed(() => records.value ? records.value.departures : []);

const timeRangeLabel = computed(() => {
  switch(timeRange.value) {
    case 'day': return 'Dieninė pamaina (06:00-20:00)';
    case 'night': return 'Naktinė pamaina (18:00-08:00)';
    default: return 'Para (00:00-23:59)';
  }
});

const selectedDateDecimal = computed(() => {
  if (!selectedDate.value) return 0
  
  const dateObj = new Date(`${selectedDate.value}T00:00:00Z`)
  return timeToDecimal(dateObj)
})

const formatName = (name) => {
  if (!name) return '...'
  const parts = name.split(',')
  if (parts.length > 1) {
    return `${parts[1]} ${parts[0]}`
  }
  return name
}

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
  } else {
    slogStore.addToast({
      message: 'Nepavyko kopijuoti',
      type: 'alert-error'
    })
  }
}

function toggleDrawer() {
  isDrawerOpen.value = !isDrawerOpen.value;
  
  loggingStore.uiEvent('Klasika', `timeline_drawer_${isDrawerOpen.value ? 'opened' : 'closed'}`, {
    component: 'TimelineDrawer'
  });
}

/**
 * Обрабатывает назначение пути для локомотива
 * @param {Object} data - Данные о назначении пути
 */
const handleTrackAssigned = (data) => {
  if (!data || !data.id) return
  let movementId = null

  try {
    if (data.type === 'arrival') {
      // Найти запись прибытия по ID
      const arrival = allArrivals.value.find(arr => arr.id === data.id)
      if (arrival) {
        // Обновить запись в хранилище
        clipStore.updateRecord(data.id, {
          targetTrack: data.trackAssignment // Будет null при удалении
        })
        
        if (data.trackAssignment) {
          loggingStore.info('Kelias priskirtas atvykimui', {
            component: 'Klasika',
            train: arrival.arrivalTrainNumber,
            vehicle: arrival.vehicle,
            track: data.trackAssignment
          })
        } else {
          if (arrival.arrivalTrainNumber.indexOf('M') > -1) {
            movementId = arrival.id
          }
          loggingStore.info('Kelio priskyrimas pašalintas atvykimui', {
            component: 'Klasika',
            train: arrival.arrivalTrainNumber,
            vehicle: arrival.vehicle
          })
          
          slogStore.addToast({
            message: `Kelio priskyrimas pašalintas ${arrival.vehicle} atvykimui`,
            type: 'alert-info'
          })
        }
      }
    } else if (data.type === 'departure') {
      // Найти запись отправления по ID
      const departure = allDepartures.value.find(dep => dep.id === data.id)
      if (departure) {
        // Обновить запись в хранилище
        clipStore.updateRecord(data.id, {
          startingTrack: data.trackAssignment // Будет '' при удалении
        })
        
        if (data.trackAssignment) {
          loggingStore.info('Kelias priskirtas išvykimui', {
            component: 'Klasika',
            train: departure.departureTrainNumber,
            vehicle: departure.vehicle,
            track: data.trackAssignment
          })
        } else {
          if (departure.departureTrainNumber.indexOf('M') > -1) {
            movementId = departure.id
          }
          loggingStore.info('Kelio priskyrimas pašalintas išvykimui', {
            component: 'Klasika',
            train: departure.departureTrainNumber,
            vehicle: departure.vehicle
          })
          
          slogStore.addToast({
            message: `Kelio priskyrimas pašalintas ${departure.vehicle} išvykimui`,
            type: 'alert-info'
          })
        }
      }
    }
    if (movementId) clipStore.deleteRecord(movementId)
  } catch (error) {
    loggingStore.error('Klaida keičiant kelio priskyrimą', {
      component: 'Klasika',
      error: error.message
    })
    
    slogStore.addToast({
      message: 'Klaida keičiant kelio priskyrimą',
      type: 'alert-error'
    })
  }
}
/**
 * Обрабатывает создание маневра
 * @param {Object} data - Данные о маневре
 */
const handleMovementCreated = (data) => {
  if (!data || !data.vehicle || !data.fromTrack || !data.toTrack) return
  
  try {
    // Генерировать уникальный номер маневра
    const movementNumber = `M${Math.floor(1000 + Math.random() * 9000)}`
    
    // Создать запись маневра
    const movementRecord = {
      vehicle: data.vehicle,
      startingLocation: selectedDepot.value,
      endLocation: selectedDepot.value,
      arrivalTrainNumber: movementNumber,
      departureTrainNumber: movementNumber,
      startingTrack: data.fromTrack,
      targetTrack: data.toTrack,
      arrivalPlanned: minutesToTime(data.endTime),
      departurePlanned: minutesToTime(data.startTime),
      arrivalDate: selectedDate.value,
      departureDate: selectedDate.value,
      arrivalDecimal: data.endTime,
      departureDecimal: data.startTime,
      arrivalEmployee1: 'Paranga',
      departureEmployee1: 'Paranga',
      notes: `Manevras: iš ${data.fromTrack} į ${data.toTrack}`,
      movement: true,
    }
    
    // Добавить запись в хранилище
    clipStore.addRecord(movementRecord)
    
    loggingStore.info('Создан маневр', {
      component: 'Klasika',
      movement: movementNumber,
      vehicle: data.vehicle,
      fromTrack: data.fromTrack,
      toTrack: data.toTrack
    })
    
    slogStore.addToast({
      message: `Manevras sukurtas: ${data.vehicle}`,
      type: 'alert-success'
    })
  } catch (error) {
    loggingStore.error('Ошибка при создании маневра', {
      component: 'Klasika',
      error: error.message
    })
    
    slogStore.addToast({
      message: 'Klaida kuriant manevrą',
      type: 'alert-error'
    })
  }
}



/**
 * Преобразует минуты в строковое представление времени (HH:MM)
 * @param {number} minutes - Количество минут с 2025-01-01 00:00
 * @returns {string} Время в формате HH:MM
 */
const minutesToTime = (minutes) => {
  const date = new Date('2025-01-01T00:00:00Z')
  date.setUTCMinutes(minutes)
  return date.toISOString().substring(11, 16)
}

async function loadStations() {
  console.log('Loading tracks for stations')
  try {
    const response = await api.get('/stations')
    const stations = response.data
    console.log('Loaded stations:', stations)
    clipStore.setStations(stations)
  } catch (error) {
    loggingStore.error(`Klaida įkeliant stočių duomenis: ${error.message}`, {
      component: 'Klasika',
      error: error.message
    })
    slogStore.addToast({
      message: `Klaida įkeliant stočių duomenys: ${error.message}`,
      type: 'alert-error'
    })
  }
}

function applyFilters() {
  if (!selectedDepot.value || !selectedDate.value) {
    slogStore.addToast({
      message: 'Pasirinkite depą ir datą',
      type: 'alert-warning'
    });
    return;
  }
  clipStore.setFilters({
    depot: selectedDepot.value,
    date: selectedDate.value,
    timeRange: timeRange.value
  });

  loggingStore.info('Taikyti filtrai', {
    component: 'Klasika',
    depot: selectedDepot.value,
    date: selectedDate.value,
    timeRange: timeRange.value
  });
}

async function handleClipboardImport() {
  try {
    // For testing purposes we use dummy data
    //const text = clip;
    const text = await navigator.clipboard.readText();

    const count = clipStore.importFromText(text);

    if (count > 0) {
      slogStore.addToast({
        message: `Sėkmingai importuota ${count} įrašų`,
        type: 'alert-success'
      });

      if (!selectedDepot.value && availableDepots.value.length > 0) {
        selectedDepot.value = availableDepots.value[0];
      }

      if (!selectedDate.value && availableDates.value.length > 0) {
        selectedDate.value = availableDates.value[0];
      }

      applyFilters();
    } else {
      slogStore.addToast({
        message: 'Nepavyko importuoti: netinkamas duomenų formatas',
        type: 'alert-warning'
      });
    }
  } catch (error) {
    loggingStore.error('Klaida importuojant duomenis', {
      component: 'Klasika',
      action: 'clipboard_import_error',
      error: error.message || 'Unknown error'
    });

    slogStore.addToast({
      message: `Klaida importuojant duomenis: ${error.message}`,
      type: 'alert-error'
    });
  }
}

function exportToPDF() {
  if (filteredArrivals.value.length === 0 && filteredDepartures.value.length === 0) {
    slogStore.addToast({
      message: 'Nėra duomenų eksportavimui į PDF',
      type: 'alert-warning'
    });
    return;
  }

  printTables();

  loggingStore.info('Duomenys eksportuoti į PDF', {
    component: 'Klasika',
    action: 'export_to_pdf',
    depot: selectedDepot.value,
    date: selectedDate.value,
    arrivals: filteredArrivals.value.length,
    departures: filteredDepartures.value.length
  });
}

function printTables() {
  if (filteredArrivals.value.length === 0 && filteredDepartures.value.length === 0) {
    slogStore.addToast({
      message: 'Nėra duomenų spausdinimui',
      type: 'alert-warning'
    });
    return;
  }

  try {
    window.print();

    loggingStore.info('Duomenys išsiųsti spausdintuvui', {
      component: 'Klasika',
      action: 'print_tables',
      depot: selectedDepot.value,
      date: selectedDate.value,
      timeRange: timeRange.value
    });
  } catch (error) {
    loggingStore.error('Klaida spausdinant', {
      component: 'Klasika',
      action: 'print_error',
      error: error.message || 'Unknown error'
    });

    slogStore.addToast({
      message: 'Klaida spausdinant duomenis',
      type: 'alert-error'
    });
  }
}

watch([selectedDepot, selectedDate, timeRange], () => {
  if (selectedDepot.value && selectedDate.value) {
    clipStore.setFilters({
      depot: selectedDepot.value,
      date: selectedDate.value,
      timeRange: timeRange.value
    });
    loggingStore.info('Filtrai atnaujinti', {
      component: 'Klasika',
      depot: selectedDepot.value,
      date: selectedDate.value,
      timeRange: timeRange.value
    });
    const trx = clipStore.availableTracks;
    if (trx && trx.length > 0) {
      tracksLoaded.value = true;
      tracks.value = trx;
      console.log('Available tracks:', tracks.value, tracksLoaded.value);
    } else {
      tracksLoaded.value = false;
      tracks.value = [];
    }
  }
});

// Initialize component on mount
onMounted(async () => {
  isLoading.value = true;

  try {
    if (!clipStore.fieldMappingsLoaded) {
      const mappingsLoaded = await clipStore.loadFieldMappings();
      if (!mappingsLoaded) {
        slogStore.addToast({
          message: 'Nepavyko įkelti laukų konfigūracijos',
          type: 'alert-error'
        });
        
        loggingStore.error('Klaida įkeliant laukų konfigūraciją', {
          component: 'Klasika',
          action: 'load_field_mappings_failed'
        });
      }
    }

    if (!clipStore.stationsAvailable) {
      await loadStations();
    }

    if ( clipStore.selectedDate && clipStore.selectedDepot ) {
      selectedDepot.value = clipStore.selectedDepot;
      selectedDate.value = clipStore.selectedDate;
    } else {

      if (availableDepots.value.length > 0 && !selectedDepot.value) {
        selectedDepot.value = availableDepots.value[0];
      }

      if (availableDates.value.length > 0 && !selectedDate.value) {
        selectedDate.value = availableDates.value[0];
      }

      clipStore.setFilters({
        depot: selectedDepot.value,
        date: selectedDate.value,
        timeRange: timeRange.value
      });
    }

    loggingStore.info('Klasika puslapis užkrautas', {
      component: 'Klasika',
      timestamp: new Date().toISOString()
    });
  } catch (error) {
    loggingStore.error('Klaida kraunant puslapį', {
      component: 'Klasika',
      action: 'page_load_error',
      error: error.message || 'Unknown error'
    });

    slogStore.addToast({
      message: 'Klaida kraunant duomenis',
      type: 'alert-error'
    });
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.klasika-page {
  min-height: 100vh;
}

.wide {
  width: 98%;
}

/* Timeline drawer styles */
.timeline-drawer-wrapper {
  position: fixed;
  top: 0;
  right: -80vw;
  height: 100vh;
  z-index: 50;
  background-color: white;
  transition: right 0.3s ease-in-out;
}
.timeline-open {
  right: 0;
  transition: right 0.3s ease-in-out;
}

.timeline-drawer {
  width: 80vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--base-100);
  box-shadow: -2px 0 5px rgba(0, 0, 0, 0.1);
  position: relative;
}

.timeline-tab {
  position: absolute;
  top: 20%;
  left: -24px;
  background-color: #ff5722;
  color: white;
  padding: 1.2rem 0;
  border-radius: 0.5rem 0 0 0.5rem;
  cursor: pointer;
  width: 24px;
  height: 60px;
  z-index: 100;
}

.timeline-tab-icon {
  width: 24px;
  height: 24px;
  transform: rotate(180deg);
  transition: transform 0.3s ease-in-out;
}
.timeline-tab-active .timeline-tab-icon {
  transform: rotate(0);
}

.timeline-drawer-content {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}
.timeline-drawer-content h2 {
  margin-bottom: 0.5rem;
}

/* Hide print section in normal view */
#print-section {
  display: none;
}

.td {
  text-align: center;
  vertical-align: middle;
}

.th {
  text-align: center;
  vertical-align: middle;
  font-weight: bold;
  font-size: smaller;
  font-style: italic;
}

.copible {
  cursor: pointer;
}

/* Print styles */
@media print {
  #print-section {
    display: flex;
    flex-direction: column;
    width: 100%;
    margin: 0 auto;
    padding: 1rem;
    background-color: #fff;
    color: #000;
  }

  .print-content {
    display: flex;
    flex-direction: row;
    margin: 0 auto;
    width: 100%;
  }
  .unprintable, .timeline-drawer-wrapper {
    display: none;
  }

  .print-table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 1rem;
  }

  .print-table th,
  .print-table td {
    padding: 0.5rem;
    text-align: left;
    border: 1px solid #ddd;
  }

  .print-table th {
    font-weight: bold;
    background-color: #f2f2f2;
  }

  .print-header {
    margin-bottom: 1rem;
  }

  .print-footer {
    margin-top: 1rem;
    font-size: 0.8rem;
    color: #666;
    text-align: right;
  }

  @page {
    size: A4;
    margin: 1cm;
    scale: .5;
  }

  .timeline-container {
    display: none;
  }
}
</style>
