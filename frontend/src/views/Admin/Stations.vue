<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .bg-base-200.px-2.mt-26.pb-2.rounded-xl(class="w-11/12")
    h1.card-title.text-3xl.my-2 Stotys / Depai
    h2.flex.justify-between.items-center.mb-2
      .text-xl.font-bold Stočių sąrašas
      button.btn.btn-primary(@click="showAddStationModal = true")
        svg.h-5.w-5.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
          path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4")
        | Pridėti stotį

    div(v-if="loading").flex.justify-center.my-8
      span.loading.loading-spinner.loading-lg

    div(v-else-if="stations.length === 0").text-center.py-8
      .text-4xl.mb-2 🚉
      p.text-lg Nėra registruotų stočių. Sukurkite pirmąją stotį!

    .grid.grid-cols-3.gap-4(v-else)
      .card.bg-base-100.shadow-md(v-for="station in stations" :key="station.id")
        .card-body
          .flex.justify-between.items-start
            h3.card-title {{ station.name }}
            .badge.badge-primary {{ station.code }}
            .dropdown.dropdown-end
              div.btn.btn-sm.btn-ghost(tabindex="0" role="button")
                svg.h-5.w-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z")
              ul.dropdown-content.menu.p-2.shadow.bg-base-100.rounded-box.w-52(tabindex="0")
                li
                  button(@click="editStation(station)") Redaguoti
                li
                  button.text-error(@click="confirmDeleteStation(station)") Ištrinti


          .divider.my-2 Keliai

          .overflow-x-auto(v-if="station.tracks && station.tracks.length > 0")
            table.table.table-xs.table-zebra.w-full
              thead
                tr
                  th Nr.
                  th Pozicijos
                  th Tipas
                  th Taisyklė
                  th Išimtys
                  th
              tbody
                tr(v-for="track in station.tracks.sort((a, b) => a.id - b.id)" :key="track.id")
                  td
                    a.cursor-pointer(class="hover:underline" @click="editTrack(station, track)") {{ track.track_number }}
                  td {{ track.positions }}
                  td {{ track.type === 'through' ? 'Pravažiuojamas' : 'Aklakelis' }}
                  td {{ track.rule.toUpperCase() }}
                  td {{ track.exceptions ? 'Taip' : 'Ne' }}
                  td.text-right
                    button.btn.btn-xs.btn-error(@click="confirmDeleteTrack(station, track)")
                      svg.h-4.w-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                        path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")

          .text-sm.text-center.text-gray-500.my-2(v-else) Nėra užregistruotų kelių
          .mt-2.text-sm(v-if="station.notes") {{ station.notes }}

          .card-actions.justify-end.mt-2
            button.btn.btn-sm.btn-outline(@click="addTrack(station)")
              | Pridėti kelią

  .modal(:class="{'modal-open': showAddStationModal || editingStation !== null}")
    .modal-box
      h3.font-bold.text-lg {{ editingStation ? 'Redaguoti stotį' : 'Pridėti naują stotį' }}

      form(@submit.prevent="editingStation ? updateStation() : createStation()")
        fieldset.fieldset.w-full.border.border-base-300.rounded-box.p-4
          label.input
            span.label-text Pavadinimas
            input.grow(
              type="text"
              v-model="stationForm.name"
              required
              placeholder="Stoties pavadinimas"
            )
          label.input
            span.label-text Kodas
            input.grow(
              type="text"
              v-model="stationForm.code"
              required
              placeholder="Stoties kodas"
            )
          textarea.textarea.textarea-bordered.h-24.grow(
            v-model="stationForm.notes"
            placeholder="Papildoma informacija"
          )
        .modal-action
          button.btn(type="button" @click="cancelStationEdit") Atšaukti
          button.btn.btn-primary(type="submit") {{ editingStation ? 'Atnaujinti' : 'Pridėti' }}

  .modal(:class="{'modal-open': showAddTrackModal || editingTrack !== null}")
    .modal-box
      h3.font-bold.text-lg
        | {{ editingTrack ? 'Redaguoti kelią' : 'Pridėti naują kelią' }}
        span.text-sm.font-normal.ml-2(v-if="selectedStation") ({{ selectedStation.name }})

      form(@submit.prevent="editingTrack ? updateTrack() : createTrack()")
        fieldset.fieldset.w-full.border.border-base-300.rounded-box.p-4
          label.input
            span.label-text Numeris
            input.grow(
              type="text"
              v-model="trackForm.track_number"
              required
              placeholder="Kelio numeris"
            )
          label.input
            span.label-text Pozicijų skaičius
            input.grow(
              type="number"
              v-model.number="trackForm.positions"
              required
              min="1"
              placeholder="Kiek pozicijų turi kelias"
            )
          label.input
            span.label-text Ilgis (m)
            input.grow(
              type="number"
              v-model.number="trackForm.length"
              min="1"
              placeholder="Kelio ilgis metrais (neprivaloma)"
            )
          select.select.select-bordered.w-full(
            v-model="trackForm.type"
            @change="handleTypeChange"
          )
            option(disabled selected) -- Pasirinkite kelio tipą --
            option(value="dead_end") Aklakelis
            option(value="through") Pravažiuojamas
          select.select.select-bordered.w-full(
            v-model="trackForm.rule"
            :disabled="trackForm.type === 'dead_end'"
          )
            option(disabled selected) -- Pasirinkite keliavimo taisyklę --
            option(value="filo") FILO (First In Last Out)
            option(value="fifo" :disabled="trackForm.type === 'dead_end'") FIFO (First In First Out)
          label.label.cursor-pointer.justify-start.gap-4
            span.label-text Išimtys
            input.checkbox(type="checkbox" v-model="trackForm.exceptions" :disabled="trackForm.type === 'dead_end'")
          textarea.textarea.textarea-bordered.h-24(
            v-model="trackForm.notes"
            placeholder="Papildoma informacija"
          )
        .modal-action
          button.btn(type="button" @click="cancelTrackEdit") Atšaukti
          button.btn.btn-primary(type="submit") {{ editingTrack ? 'Atnaujinti' : 'Pridėti' }}

  .modal(:class="{'modal-open': stationToDelete !== null}")
    .modal-box
      h3.font-bold.text-lg Patvirtinkite ištrynimą
      p.py-4
        | Ar tikrai norite ištrinti stotį "{{ stationToDelete?.name }}" ({{ stationToDelete?.code }})?
        br
        span.text-error.font-semibold Taip pat bus ištrinti visi stoties keliai!
      .modal-action
        button.btn(@click="stationToDelete = null") Atšaukti
        button.btn.btn-error(@click="deleteStation") Ištrinti

  .modal(:class="{'modal-open': trackToDelete !== null}")
    .modal-box
      h3.font-bold.text-lg Patvirtinkite ištrynimą
      p.py-4
        | Ar tikrai norite ištrinti kelią "{{ trackToDelete?.track_number }}"
        | stotyje "{{ trackStationToDelete?.name }}"?
      .modal-action
        button.btn(@click="trackToDelete = null; trackStationToDelete = null;") Atšaukti
        button.btn.btn-error(@click="deleteTrack") Ištrinti
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';
import api from '@/services/api';
import { useLoggingStore } from '@/stores/logging';
import { useSlogStore } from '@/stores/slog';

/**
 * The Station Manager - central control hub for railway logistics
 *
 * This component provides comprehensive management of stations and tracks:
 * - View a list of all stations with their tracks
 * - Add, edit, and delete stations
 * - Add, edit, and delete tracks within each station
 */

// Store instances for logging and notifications
const loggingStore = useLoggingStore();
const slogStore = useSlogStore();

// State variables for station management
const stations = ref([]);
const loading = ref(true);
const showAddStationModal = ref(false);
const editingStation = ref(null);
const stationToDelete = ref(null);

// State variables for track management
const showAddTrackModal = ref(false);
const selectedStation = ref(null);
const editingTrack = ref(null);
const trackToDelete = ref(null);
const trackStationToDelete = ref(null);

// Form data for station editing
const stationForm = reactive({
  name: '',
  code: '',
  notes: ''
});

// Form data for track editing
const trackForm = reactive({
  track_number: '',
  positions: 1,
  length: null,
  type: 'through',
  rule: 'fifo',
  exceptions: false,
  notes: ''
});

/**
 * Enforces proper rules for each track type
 * When track type changes, automatically sets the appropriate rule
 */
function handleTypeChange() {
  // If type is dead_end, force rule to be filo
  if (trackForm.type === 'dead_end') {
    trackForm.rule = 'filo';
  }
}

/**
 * Fetches the station catalog from the server
 * Retrieves all stations and their tracks from the API
 */
const loadStations = async () => {
  loading.value = true;

  try {
    const response = await api.get('/stations');
    stations.value = response.data;

    loggingStore.info('Stočių sąrašas gautas', {
      component: 'Stations',
      action: 'load_stations',
      count: response.data.length
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida gaunant stočių sąrašą',
      type: 'alert-error'
    });

    loggingStore.error('Klaida gaunant stočių sąrašą', {
      component: 'Stations',
      action: 'load_stations_failed',
      error: error.response?.data || error.message
    });
  } finally {
    loading.value = false;
  }
};

/**
 * Submits the form data to create a new station
 * Adds a new station to the database and updates the UI
 */
const createStation = async () => {
  try {
    const response = await api.post('/stations', stationForm);

    // Add the new station to the list
    stations.value.push(response.data);

    // Reset form and close modal
    resetStationForm();
    showAddStationModal.value = false;

    // Show success message
    slogStore.addToast({
      message: `Stotis "${response.data.name}" sėkmingai sukurta`,
      type: 'alert-success'
    });

    loggingStore.info('Nauja stotis sukurta', {
      component: 'Stations',
      action: 'create_station',
      stationId: response.data.id,
      stationName: response.data.name
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida kuriant stotį: ' + (error.response?.data || error.message),
      type: 'alert-error'
    });

    loggingStore.error('Klaida kuriant stotį', {
      component: 'Stations',
      action: 'create_station_failed',
      error: error.response?.data || error.message,
      formData: { ...stationForm }
    });
  }
};

/**
 * Sets up the form with the selected station's data
 * Prepares for editing an existing station
 *
 * @param {Object} station - The station to be edited
 */
const editStation = (station) => {
  editingStation.value = station;
  stationForm.name = station.name;
  stationForm.code = station.code;
  stationForm.notes = station.notes || '';
};

/**
 * Submits the edited station data to the server
 * Updates an existing station with modified information
 */
const updateStation = async () => {
  if (!editingStation.value) return;

  try {
    const response = await api.put(`/stations/${editingStation.value.id}`, stationForm);

    // Update the station in the local list
    const index = stations.value.findIndex(s => s.id === editingStation.value.id);
    if (index !== -1) {
      stations.value[index] = response.data;
    }

    // Reset form and close modal
    resetStationForm();
    editingStation.value = null;

    // Show success message
    slogStore.addToast({
      message: `Stotis "${response.data.name}" sėkmingai atnaujinta`,
      type: 'alert-success'
    });

    loggingStore.info('Stotis atnaujinta', {
      component: 'Stations',
      action: 'update_station',
      stationId: response.data.id,
      stationName: response.data.name
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida atnaujinant stotį: ' + (error.response?.data || error.message),
      type: 'alert-error'
    });

    loggingStore.error('Klaida atnaujinant stotį', {
      component: 'Stations',
      action: 'update_station_failed',
      error: error.response?.data || error.message,
      stationId: editingStation.value.id
    });
  }
};

/**
 * Shows a confirmation dialog before deleting a station
 *
 * @param {Object} station - The station to be deleted
 */
const confirmDeleteStation = (station) => {
  stationToDelete.value = station;
};

/**
 * Deletes a station after confirmation
 * Removes a station and all its tracks from the database
 */
const deleteStation = async () => {
  if (!stationToDelete.value) return;

  try {
    await api.delete(`/stations/${stationToDelete.value.id}`);

    // Remove station from local list
    stations.value = stations.value.filter(s => s.id !== stationToDelete.value.id);

    // Show success message
    slogStore.addToast({
      message: `Stotis "${stationToDelete.value.name}" sėkmingai ištrinta`,
      type: 'alert-success'
    });

    loggingStore.info('Stotis ištrinta', {
      component: 'Stations',
      action: 'delete_station',
      stationId: stationToDelete.value.id,
      stationName: stationToDelete.value.name
    });

    // Close confirmation dialog
    stationToDelete.value = null;
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida ištrinant stotį: ' + (error.response?.data || error.message),
      type: 'alert-error'
    });

    loggingStore.error('Klaida ištrinant stotį', {
      component: 'Stations',
      action: 'delete_station_failed',
      error: error.response?.data || error.message,
      stationId: stationToDelete.value.id
    });
  }
};

/**
 * Resets the station form to its default empty state
 */
const resetStationForm = () => {
  stationForm.name = '';
  stationForm.code = '';
  stationForm.notes = '';
};

/**
 * Cancels station editing, resets form and closes modal
 */
const cancelStationEdit = () => {
  resetStationForm();
  editingStation.value = null;
  showAddStationModal.value = false;
};

/**
 * Initializes the track form for a specific station
 * Prepares to add a new track
 *
 * @param {Object} station - The station where the track will be added
 */
const addTrack = (station) => {
  selectedStation.value = station;
  showAddTrackModal.value = true;
  resetTrackForm();
};

/**
 * Submits form data to create a new track in a station
 * Adds a new railway track to an existing station
 */
const createTrack = async () => {
  if (!selectedStation.value) return;

  try {
    const response = await api.post(
      `/stations/${selectedStation.value.id}/tracks`,
      trackForm
    );

    // Update station in the local list with the new track
    const index = stations.value.findIndex(s => s.id === selectedStation.value.id);
    if (index !== -1) {
      stations.value[index] = response.data;
    }

    // Reset form and close modal
    resetTrackForm();
    showAddTrackModal.value = false;
    selectedStation.value = null;

    // Show success message
    slogStore.addToast({
      message: `Kelias sėkmingai pridėtas`,
      type: 'alert-success'
    });

    loggingStore.info('Naujas kelias sukurtas', {
      component: 'Stations',
      action: 'create_track',
      stationId: response.data.id,
      stationName: response.data.name,
      trackNumber: trackForm.track_number
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida kuriant kelią: ' + (error.response?.data || error.message),
      type: 'alert-error'
    });

    loggingStore.error('Klaida kuriant kelią', {
      component: 'Stations',
      action: 'create_track_failed',
      error: error.response?.data || error.message,
      stationId: selectedStation.value.id,
      formData: { ...trackForm }
    });
  }
};

/**
 * Sets up the form with the selected track's data
 * Prepares a track for modification
 *
 * @param {Object} station - The station containing the track
 * @param {Object} track - The track to be edited
 */
const editTrack = (station, track) => {
  selectedStation.value = station;
  editingTrack.value = track;

  // Fill form with track data
  trackForm.track_number = track.track_number;
  trackForm.positions = track.positions;
  trackForm.length = track.length || null;
  trackForm.type = track.type || 'through';
  trackForm.rule = track.rule || 'fifo';
  trackForm.exceptions = track.exceptions || false;
  trackForm.notes = track.notes || '';
};

/**
 * Submits the edited track data to the server
 * Updates an existing track with modified information
 */
const updateTrack = async () => {
  if (!editingTrack.value) return;

  try {
    const response = await api.put(`/tracks/${editingTrack.value.id}`, trackForm);

    // Update track in the station's track list
    const stationIndex = stations.value.findIndex(s => s.id === selectedStation.value.id);
    if (stationIndex !== -1) {
      stations.value[stationIndex] = response.data;
    }

    // Reset form and close modal
    resetTrackForm();

    // Show success message
    slogStore.addToast({
      message: `Kelias sėkmingai atnaujintas`,
      type: 'alert-success'
    });

    loggingStore.info('Kelias atnaujintas', {
      component: 'Stations',
      action: 'update_track',
      trackId: editingTrack.value.id,
      trackNumber: trackForm.track_number
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida atnaujinant kelią: ' + (error.response?.data || error.message),
      type: 'alert-error'
    });

    loggingStore.error('Klaida atnaujinant kelią', {
      component: 'Stations',
      action: 'update_track_failed',
      error: error.response?.data || error.message,
      trackId: editingTrack.value.id
    });
  } finally {
    editingTrack.value = null;
    selectedStation.value = null;
  }
};

/**
 * Shows a confirmation dialog before deleting a track
 *
 * @param {Object} station - The station containing the track
 * @param {Object} track - The track to be deleted
 */
const confirmDeleteTrack = (station, track) => {
  trackStationToDelete.value = station;
  trackToDelete.value = track;
};

/**
 * Deletes a track after confirmation
 * Removes a track from a station in the database
 */
const deleteTrack = async () => {
  if (!trackToDelete.value || !trackStationToDelete.value) return;

  try {
    const response = await api.delete(`/tracks/${trackToDelete.value.id}`);

    // Update station in the local list with the updated tracks
    const stationIndex = stations.value.findIndex(s => s.id === trackStationToDelete.value.id);
    if (stationIndex !== -1) {
      stations.value[stationIndex] = response.data;
    }

    // Show success message
    slogStore.addToast({
      message: `Kelias "${trackToDelete.value.track_number}" sėkmingai ištrintas`,
      type: 'alert-success'
    });

    loggingStore.info('Kelias ištrintas', {
      component: 'Stations',
      action: 'delete_track',
      trackId: trackToDelete.value.id,
      trackNumber: trackToDelete.value.track_number,
      stationId: trackStationToDelete.value.id
    });

    // Close confirmation dialog
    trackToDelete.value = null;
    trackStationToDelete.value = null;
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida ištrinant kelią: ' + (error.response?.data || error.message),
      type: 'alert-error'
    });

    loggingStore.error('Klaida ištrinant kelią', {
      component: 'Stations',
      action: 'delete_track_failed',
      error: error.response?.data || error.message,
      trackId: trackToDelete.value.id
    });
  }
};

/**
 * Resets the track form to its default empty state
 */
const resetTrackForm = () => {
  trackForm.track_number = '';
  trackForm.positions = 1;
  trackForm.length = null;
  trackForm.type = 'through';
  trackForm.rule = 'fifo';
  trackForm.exceptions = false;
  trackForm.notes = '';
};

/**
 * Cancels track editing, resets form and closes modal
 */
const cancelTrackEdit = () => {
  resetTrackForm();
  editingTrack.value = null;
  showAddTrackModal.value = false;
  selectedStation.value = null;
};

// Load stations when component mounts
onMounted(() => {
  loadStations();

  loggingStore.info('Stočių administravimo puslapis užkrautas', {
    component: 'Stations',
    timestamp: new Date().toISOString()
  });
});
</script>

<style scoped>
.card-body {
  padding: 1.5rem;
}

.modal-box {
  max-width: 32rem;
}
</style>
