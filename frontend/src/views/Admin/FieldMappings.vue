<template lang="pug">
.field-mappings-page.p-6
  .flex.justify-between.items-center.mb-6
    h1.text-3xl.font-bold Laukų atvaizdavimai
    button.btn.btn-primary(@click="showAddModal = true")
      svg.w-5.h-5.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
        path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4")
      | Pridėti naują lauką

  .bg-base-100.rounded-lg.shadow-lg.p-6(v-if="!isLoading")
    .alert.alert-info.mb-4
      svg.stroke-current.shrink-0.h-6.w-6(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24")
        path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z")
      span Čia galite redaguoti laukų pavadinimus iš išorinės sistemos. Pakeitus pavadinimus, importavimas veiks su naujais pavadinimais.

    .overflow-x-auto
      table.table.table-zebra.w-full
        thead
          tr
            th.w-12 #
            th Išorinis pavadinimas
            th Vidinis pavadinimas
            th Rodomas pavadinimas
            th Tipas
            th Privalomas
            th.w-32 Veiksmai
        tbody
          tr(v-for="(mapping, index) in mappings" :key="mapping.id")
            td {{ index + 1 }}
            td
              input.input.input-sm.input-bordered.w-full(
                v-model="mapping.external_name"
                @change="markAsModified(mapping.id)"
                :class="{ 'input-warning': modifiedIds.has(mapping.id) }"
              )
            td
              code.text-sm {{ mapping.internal_name }}
            td
              input.input.input-sm.input-bordered.w-full(
                v-model="mapping.display_name"
                @change="markAsModified(mapping.id)"
                :class="{ 'input-warning': modifiedIds.has(mapping.id) }"
              )
            td
              select.select.select-sm.select-bordered(
                v-model="mapping.field_type"
                @change="markAsModified(mapping.id)"
                :class="{ 'select-warning': modifiedIds.has(mapping.id) }"
              )
                option(value="string") String
                option(value="date") Date
                option(value="time") Time
                option(value="number") Number
                option(value="boolean") Boolean
            td
              input.checkbox.checkbox-sm(
                type="checkbox"
                v-model="mapping.is_required"
                @change="markAsModified(mapping.id)"
              )
            td
              .flex.gap-2
                button.btn.btn-sm.btn-success(
                  v-if="modifiedIds.has(mapping.id)"
                  @click="updateMapping(mapping)"
                  title="Išsaugoti pakeitimus"
                )
                  svg.w-4.h-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7")
                button.btn.btn-sm.btn-error(
                  @click="confirmDelete(mapping)"
                  title="Ištrinti"
                )
                  svg.w-4.h-4(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                    path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")

    .flex.justify-end.mt-4(v-if="modifiedIds.size > 0")
      button.btn.btn-success(@click="saveAllModified")
        | Išsaugoti visus pakeitimus ({{ modifiedIds.size }})

  .flex.justify-center.items-center.min-h-screen(v-else)
    span.loading.loading-spinner.loading-lg

  // Add Modal
  .modal(:class="{ 'modal-open': showAddModal }")
    .modal-box
      h3.font-bold.text-lg.mb-4 Pridėti naują lauką
      
      .form-control.mb-4
        label.label
          span.label-text Išorinis pavadinimas
        input.input.input-bordered(
          v-model="newMapping.external_name"
          placeholder="Pvz.: Vehicle working number"
        )
      
      .form-control.mb-4
        label.label
          span.label-text Vidinis pavadinimas
        input.input.input-bordered(
          v-model="newMapping.internal_name"
          placeholder="Pvz.: vehicleWorkingNumber"
        )
      
      .form-control.mb-4
        label.label
          span.label-text Rodomas pavadinimas
        input.input.input-bordered(
          v-model="newMapping.display_name"
          placeholder="Pvz.: Riedmens darbo numeris"
        )
      
      .form-control.mb-4
        label.label
          span.label-text Lauko tipas
        select.select.select-bordered(v-model="newMapping.field_type")
          option(value="string") String
          option(value="date") Date
          option(value="time") Time
          option(value="number") Number
          option(value="boolean") Boolean
      
      .form-control.mb-4
        label.label.cursor-pointer
          span.label-text Privalomas laukas
          input.checkbox(
            type="checkbox"
            v-model="newMapping.is_required"
          )
      
      .form-control.mb-4
        label.label
          span.label-text Aprašymas (neprivalomas)
        textarea.textarea.textarea-bordered(
          v-model="newMapping.description"
          rows="3"
        )
      
      .modal-action
        button.btn(@click="showAddModal = false") Atšaukti
        button.btn.btn-primary(@click="createMapping") Pridėti

  // Delete Confirmation Modal
  .modal(:class="{ 'modal-open': showDeleteModal }")
    .modal-box
      h3.font-bold.text-lg.mb-4 Patvirtinti ištrynimą
      p.mb-4 Ar tikrai norite ištrinti lauką "{{ mappingToDelete?.external_name }}"?
      .modal-action
        button.btn(@click="showDeleteModal = false") Atšaukti
        button.btn.btn-error(@click="deleteMapping") Ištrinti
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '@/services/api'
import { useLoggingStore } from '@/stores/logging'
import { useSlogStore } from '@/stores/slog'

const loggingStore = useLoggingStore()
const slogStore = useSlogStore()

const mappings = ref([])
const isLoading = ref(false)
const modifiedIds = ref(new Set())

const showAddModal = ref(false)
const showDeleteModal = ref(false)
const mappingToDelete = ref(null)

const newMapping = ref({
  external_name: '',
  internal_name: '',
  display_name: '',
  field_type: 'string',
  is_required: false,
  sort_order: 0,
  description: ''
})

/**
 * Load all field mappings from the server
 */
const loadMappings = async () => {
  isLoading.value = true
  try {
    const response = await api.get('/field-mappings')
    mappings.value = response.data

    loggingStore.info('Laukų atvaizdavimai užkrauti', {
      component: 'FieldMappings',
      action: 'load_mappings',
      count: response.data.length
    })
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida įkeliant laukų atvaizdavimus',
      type: 'alert-error'
    })

    loggingStore.error('Klaida įkeliant laukų atvaizdavimus', {
      component: 'FieldMappings',
      action: 'load_mappings_failed',
      error: error.response?.data || error.message
    })
  } finally {
    isLoading.value = false
  }
}

/**
 * Mark a mapping as modified (needs saving)
 */
const markAsModified = (id) => {
  modifiedIds.value.add(id)
}

/**
 * Update a single field mapping
 */
const updateMapping = async (mapping) => {
  try {
    await api.put(`/field-mappings/${mapping.id}`, mapping)

    modifiedIds.value.delete(mapping.id)

    slogStore.addToast({
      message: 'Laukas sėkmingai atnaujintas',
      type: 'alert-success'
    })

    loggingStore.info('Laukas atnaujintas', {
      component: 'FieldMappings',
      action: 'update_mapping',
      mappingId: mapping.id
    })

    // Reload to get fresh data
    await loadMappings()
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida atnaujinant lauką',
      type: 'alert-error'
    })

    loggingStore.error('Klaida atnaujinant lauką', {
      component: 'FieldMappings',
      action: 'update_mapping_failed',
      mappingId: mapping.id,
      error: error.response?.data || error.message
    })
  }
}

/**
 * Save all modified mappings
 */
const saveAllModified = async () => {
  const modified = mappings.value.filter(m => modifiedIds.value.has(m.id))

  for (const mapping of modified) {
    await updateMapping(mapping)
  }
}

/**
 * Create a new field mapping
 */
const createMapping = async () => {
  if (!newMapping.value.external_name || !newMapping.value.internal_name || !newMapping.value.display_name) {
    slogStore.addToast({
      message: 'Užpildykite visus privalomus laukus',
      type: 'alert-warning'
    })
    return
  }

  try {
    // Set sort_order to be last
    newMapping.value.sort_order = mappings.value.length

    await api.post('/field-mappings', newMapping.value)

    slogStore.addToast({
      message: 'Naujas laukas sėkmingai pridėtas',
      type: 'alert-success'
    })

    loggingStore.info('Naujas laukas pridėtas', {
      component: 'FieldMappings',
      action: 'create_mapping',
      externalName: newMapping.value.external_name
    })

    // Reset form
    newMapping.value = {
      external_name: '',
      internal_name: '',
      display_name: '',
      field_type: 'string',
      is_required: false,
      sort_order: 0,
      description: ''
    }

    showAddModal.value = false
    await loadMappings()
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida pridedant lauką',
      type: 'alert-error'
    })

    loggingStore.error('Klaida pridedant lauką', {
      component: 'FieldMappings',
      action: 'create_mapping_failed',
      error: error.response?.data || error.message
    })
  }
}

/**
 * Show delete confirmation modal
 */
const confirmDelete = (mapping) => {
  mappingToDelete.value = mapping
  showDeleteModal.value = true
}

/**
 * Delete a field mapping
 */
const deleteMapping = async () => {
  if (!mappingToDelete.value) return

  try {
    await api.delete(`/field-mappings/${mappingToDelete.value.id}`)

    slogStore.addToast({
      message: 'Laukas sėkmingai ištrintas',
      type: 'alert-success'
    })

    loggingStore.info('Laukas ištrintas', {
      component: 'FieldMappings',
      action: 'delete_mapping',
      mappingId: mappingToDelete.value.id
    })

    showDeleteModal.value = false
    mappingToDelete.value = null
    await loadMappings()
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida ištrinant lauką',
      type: 'alert-error'
    })

    loggingStore.error('Klaida ištrinant lauką', {
      component: 'FieldMappings',
      action: 'delete_mapping_failed',
      mappingId: mappingToDelete.value.id,
      error: error.response?.data || error.message
    })
  }
}

onMounted(() => {
  loadMappings()
})
</script>

<style scoped>
.field-mappings-page {
  min-height: 100vh;
}

.input-warning,
.select-warning {
  border-color: #fbbf24;
}
</style>
