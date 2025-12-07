<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .card.bg-base-200(class="w-11/12 h-11/12 shadow-xl overflow-auto scrollbar-thin")
    .card-body
      h1.card-title.text-3xl.mb-6 Vartotojų valdymas

      .grid.grid-cols-3.gap-6
        // Users list
        .col-span-2
          .card.bg-base-100.shadow-md
            .card-body.p-4
              h2.card-title.text-xl.mb-4
                svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z")
                | Vartotojų sąrašas

              .overflow-x-auto
                table.table.table-zebra
                  thead
                    tr
                      th ID
                      th Vartotojas
                      th.table-cell El. paštas
                      th.table-cell Rolė
                      th.table-cell Tema
                      th.table-cell Būsena
                      th Patvirtintas
                      th Veiksmai
                  tbody
                    tr(v-for="user in users" :key="user.id" class="hover")
                      td {{ user.id }}
                      td
                        .flex.items-center.space-x-3
                          .avatar
                            .mask.mask-squircle.w-10.h-10
                              img(v-if="user.avatar" :src="user.avatar" alt="Vartotojo nuotrauka")
                          div
                            .font-bold {{ user.username }}
                            .text-xs {{ user.email }}
                      td.table-cell {{ user.email }}
                      td.table-cell
                        .badge.cursor-pointer(:class="getRoleBadgeClass(user.role)" @click="openRoleModal(user)") {{ user.role }}
                      td.table-cell {{ user.theme }}
                      td.table-cell
                        .badge.cursor-pointer(:class="getStatusBadgeClass(user.user_status)" @click="openStatusModal(user)") {{ user.user_status || 'active' }}
                      td
                        .badge.cursor-pointer(:class="getVerifiedBadgeClass(user.verified)" @click="toggleUserVerification(user)") {{ user.verified ? 'Taip' : 'Ne' }}
                      td
                        .flex.space-x-2
                          button.btn.btn-square.btn-sm.btn-ghost(@click="editUser(user)")
                            svg.h-5.w-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                              path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z")
                          button.btn.btn-square.btn-sm.btn-ghost.text-error(@click="confirmDelete(user)")
                            svg.h-5.w-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                              path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")

        .col-span-1
          .card.bg-base-100.shadow-md.sticky(class="top-6")
            .card-body
              h2.card-title.text-xl.mb-4
                svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z")
                | Pridėti vartotoją

              form.space-y-4(@submit.prevent="addUser")
                .form-control
                  label.label
                    span.label-text Vartotojo vardas
                  input.input.input-bordered.w-full(
                    type="text"
                    v-model="username"
                    placeholder="Įveskite vartotojo vardą"
                    required
                  )

                .form-control
                  label.label
                    span.label-text El. paštas
                  input.input.input-bordered.w-full(
                    type="email"
                    v-model="email"
                    placeholder="vartotojas@pavyzdys.lt"
                    required
                  )

                .form-control
                  label.label
                    span.label-text Slaptažodis
                  input.input.input-bordered.w-full(
                    type="password"
                    v-model="password"
                    placeholder="Mažiausiai 8 simboliai"
                    required
                  )

                .form-control
                  label.label
                    span.label-text Rolė
                  select.select.select-bordered.w-full(v-model="role")
                    option(value="viewer") Peržiūrėtojas
                    option(value="user") Vartotojas
                    option(value="admin") Administratorius

                .form-control
                  label.label
                    span.label-text Tema
                  select.select.select-bordered.w-full(v-model="theme")
                    option(value="light") Šviesi
                    option(value="dark") Tamsi
                    option(value="dim") Pritemdyta
                    option(value="synthwave") Synthwave
                    option(value="dracula") Dracula

                .form-control
                  label.label.cursor-pointer.justify-start.gap-2
                    input.checkbox.checkbox-primary(type="checkbox" v-model="isVerified")
                    span.label-text Patvirtintas vartotojas

                .flex.justify-between.mt-6
                  button.btn.btn-ghost(type="reset" @click="resetForm") Atstatyti
                  button.btn.btn-primary(type="submit") Pridėti

  .modal(:class="{'modal-open': userToDelete}")
    .modal-box
      h3.font-bold.text-lg Patvirtinkite ištrynimą
      p.py-4
        | Ar tikrai norite ištrinti vartotoją
        span.font-bold {{ userToDelete?.username }}?
        | Šio veiksmo negalima atšaukti.
      .modal-action
        button.btn(@click="userToDelete = null") Atšaukti
        button.btn.btn-error(@click="deleteUser") Ištrinti

  // Modal for changing user role
  .modal(:class="{'modal-open': roleModalUser}")
    .modal-box
      h3.font-bold.text-lg Vartotojo rolės keitimas
      p.py-2
        | Pasirinkite naują rolę vartotojui
        span.font-bold {{ roleModalUser?.username }}:
      .form-control.w-full.py-4
        select.select.select-bordered.w-full(v-model="selectedRole")
          option(value="viewer") Peržiūrėtojas
          option(value="user") Vartotojas
          option(value="admin") Administratorius
      .modal-action
        button.btn(@click="roleModalUser = null") Atšaukti
        button.btn.btn-primary(@click="updateUserRole") Išsaugoti

  // Modal for changing user status
  .modal(:class="{'modal-open': statusModalUser}")
    .modal-box
      h3.font-bold.text-lg Vartotojo būsenos keitimas
      p.py-2
        | Pasirinkite naują būseną vartotojui
        span.font-bold {{ statusModalUser?.username }}:
      .form-control.w-full.py-4
        select.select.select-bordered.w-full(v-model="selectedStatus")
          option(value="active") Aktyvus
          option(value="inactive") Neaktyvus
          option(value="suspended") Suspenduotas
      .modal-action
        button.btn(@click="statusModalUser = null") Atšaukti
        button.btn.btn-primary(@click="updateUserStatus") Išsaugoti
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '@/services/api'
import { useSlogStore } from '@/stores/slog'
import { useLoggingStore } from '@/stores/logging'
import CachedAvatar from '@/components/common/CachedAvatar.vue'

// State for users list
const users = ref([])

// State for user form
const username = ref('')
const email = ref('')
const password = ref('')
const role = ref('viewer')
const theme = ref('dim')
const isVerified = ref(false)

// State for modals
const userToDelete = ref(null)
const roleModalUser = ref(null)
const statusModalUser = ref(null)
const selectedRole = ref('')
const selectedStatus = ref('')

// Store instances
const slogStore = useSlogStore()
const loggingStore = useLoggingStore()

/**
 * Adds a new user to the system
 * Collects form data, sends API request, and updates UI on success
 */
const addUser = async () => {
  try {
    const userData = {
      username: username.value,
      email: email.value,
      password: password.value,
      role: role.value,
      theme: theme.value,
      verified: isVerified.value
    }

    let response = await api.post('/add-user', userData)
    response = await api.get('/users')
    users.value = response.data

    resetForm()

    slogStore.addToast({
      message: 'Vartotojas sėkmingai pridėtas',
      type: 'alert-success'
    })

    loggingStore.info('Vartotojas sukurtas', {
      component: 'Users',
      action: 'add_user',
      username: userData.username,
      email: userData.email,
      role: userData.role
    })
  } catch (error) {
    slogStore.addToast({
      message: error.response?.data || 'Klaida pridedant vartotoją',
      type: 'alert-error'
    })

    loggingStore.error('Klaida pridedant vartotoją', {
      component: 'Users',
      action: 'add_user_failed',
      error: error.response?.data || error.message,
      username: username.value,
      email: email.value
    })
  }
}

/**
 * Resets the user form fields to default values
 */
const resetForm = () => {
  username.value = ''
  email.value = ''
  password.value = ''
  role.value = 'viewer'
  theme.value = 'dim'
  isVerified.value = false
}

/**
 * Prepares for user deletion by setting state and opening confirmation modal
 * @param {Object} user - The user to be deleted
 */
const confirmDelete = (user) => {
  userToDelete.value = user
}

/**
 * Deletes a user from the system after confirmation
 */
const deleteUser = async () => {
  if (!userToDelete.value) return

  try {
    await api.delete(`/delete-user/${userToDelete.value.id}`)
    users.value = users.value.filter(user => user.id !== userToDelete.value.id)

    slogStore.addToast({
      message: `Vartotojas ${userToDelete.value.username} sėkmingai ištrintas`,
      type: 'alert-success'
    })

    loggingStore.info('Vartotojas ištrintas', {
      component: 'Users',
      action: 'delete_user',
      userId: userToDelete.value.id,
      username: userToDelete.value.username
    })

    userToDelete.value = null
  } catch (error) {
    slogStore.addToast({
      message: error.response?.data || 'Klaida trinant vartotoją',
      type: 'alert-error'
    })

    loggingStore.error('Klaida ištrinant vartotoją', {
      component: 'Users',
      action: 'delete_user_failed',
      error: error.response?.data || error.message,
      userId: userToDelete.value?.id,
      username: userToDelete.value?.username
    })
  }
}

/**
 * Placeholder for future user editing functionality
 * @param {Object} user - The user to edit
 */
const editUser = (user) => {
  slogStore.addToast({
    type: 'alert-info',
    message: `Vartotojo ${user.username} redagavimas (funkcija kuriama)`
  })

  loggingStore.info('Vartotojo redagavimo bandymas', {
    component: 'Users',
    action: 'edit_user_attempt',
    userId: user.id,
    username: user.username
  })
}

/**
 * Opens the role change modal for a user
 * @param {Object} user - The user whose role will be changed
 */
const openRoleModal = (user) => {
  roleModalUser.value = user
  selectedRole.value = user.role
}

/**
 * Updates a user's role in the system
 */
const updateUserRole = async () => {
  if (!roleModalUser.value) return

  try {
    await api.put('/update-user-role', {
      user_id: roleModalUser.value.id,
      role: selectedRole.value
    })

    const userIndex = users.value.findIndex(u => u.id === roleModalUser.value.id)
    if (userIndex !== -1) {
      users.value[userIndex].role = selectedRole.value
    }

    slogStore.addToast({
      message: `Vartotojo ${roleModalUser.value.username} rolė pakeista į ${selectedRole.value}`,
      type: 'alert-success'
    })

    loggingStore.info('Vartotojo rolė pakeista', {
      component: 'Users',
      action: 'update_user_role',
      userId: roleModalUser.value.id,
      username: roleModalUser.value.username,
      oldRole: roleModalUser.value.role,
      newRole: selectedRole.value
    })

    roleModalUser.value = null
  } catch (error) {
    slogStore.addToast({
      message: error.response?.data || 'Klaida keičiant vartotojo rolę',
      type: 'alert-error'
    })

    loggingStore.error('Klaida keičiant vartotojo rolę', {
      component: 'Users',
      action: 'update_user_role_failed',
      error: error.response?.data || error.message,
      userId: roleModalUser.value?.id,
      username: roleModalUser.value?.username,
      targetRole: selectedRole.value
    })
  }
}

/**
 * Opens the status change modal for a user
 * @param {Object} user - The user whose status will be changed
 */
const openStatusModal = (user) => {
  statusModalUser.value = user
  selectedStatus.value = user.user_status || 'active'
}

/**
 * Updates a user's status in the system
 */
const updateUserStatus = async () => {
  if (!statusModalUser.value) return

  try {
    await api.put('/update-user-status', {
      user_id: statusModalUser.value.id,
      status: selectedStatus.value
    })

    const userIndex = users.value.findIndex(u => u.id === statusModalUser.value.id)
    if (userIndex !== -1) {
      users.value[userIndex].user_status = selectedStatus.value
    }

    slogStore.addToast({
      message: `Vartotojo ${statusModalUser.value.username} būsena pakeista į ${selectedStatus.value}`,
      type: 'alert-success'
    })

    loggingStore.info('Vartotojo būsena pakeista', {
      component: 'Users',
      action: 'update_user_status',
      userId: statusModalUser.value.id,
      username: statusModalUser.value.username,
      oldStatus: statusModalUser.value.user_status || 'active',
      newStatus: selectedStatus.value
    })

    statusModalUser.value = null
  } catch (error) {
    slogStore.addToast({
      message: error.response?.data || 'Klaida keičiant vartotojo būseną',
      type: 'alert-error'
    })

    loggingStore.error('Klaida keičiant vartotojo būseną', {
      component: 'Users',
      action: 'update_user_status_failed',
      error: error.response?.data || error.message,
      userId: statusModalUser.value?.id,
      username: statusModalUser.value?.username,
      targetStatus: selectedStatus.value
    })
  }
}

/**
 * Toggles a user's email verification status
 * @param {Object} user - The user whose verification status will be toggled
 */
const toggleUserVerification = async (user) => {
  try {
    const newVerifiedStatus = !user.verified

    await api.put('/update-user-verification', {
      user_id: user.id,
      verified: newVerifiedStatus
    })

    const userIndex = users.value.findIndex(u => u.id === user.id)
    if (userIndex !== -1) {
      users.value[userIndex].verified = newVerifiedStatus
    }

    slogStore.addToast({
      message: newVerifiedStatus
        ? `Vartotojas ${user.username} sėkmingai patvirtintas`
        : `Vartotojo ${user.username} patvirtinimas atšauktas`,
      type: 'alert-success'
    })

    loggingStore.info('Vartotojo patvirtinimo būsena pakeista', {
      component: 'Users',
      action: 'toggle_user_verification',
      userId: user.id,
      username: user.username,
      oldVerified: user.verified,
      newVerified: newVerifiedStatus
    })
  } catch (error) {
    slogStore.addToast({
      message: error.response?.data || 'Klaida keičiant vartotojo patvirtinimo būseną',
      type: 'alert-error'
    })

    loggingStore.error('Klaida keičiant vartotojo patvirtinimo būseną', {
      component: 'Users',
      action: 'toggle_user_verification_failed',
      error: error.response?.data || error.message,
      userId: user.id,
      username: user.username,
      currentVerified: user.verified
    })
  }
}

/**
 * Returns the appropriate CSS class for a role badge
 * @param {string} role - The user role
 * @returns {string} CSS class name for the badge
 */
const getRoleBadgeClass = (role) => {
  switch (role) {
    case 'admin': return 'badge-primary'
    case 'user': return 'badge-accent'
    default: return 'badge-secondary'
  }
}

/**
 * Returns the appropriate CSS class for a status badge
 * @param {string} status - The user status
 * @returns {string} CSS class name for the badge
 */
const getStatusBadgeClass = (status) => {
  switch (status) {
    case 'active': return 'badge-success'
    case 'suspended': return 'badge-warning'
    case 'inactive': return 'badge-error'
    default: return 'badge-ghost'
  }
}

/**
 * Returns the appropriate CSS class for a verification badge
 * @param {boolean} verified - Whether the user is verified
 * @returns {string} CSS class name for the badge
 */
const getVerifiedBadgeClass = (verified) => {
  return verified ? 'badge-success' : 'badge-error'
}

/**
 * Fetches the users list when component is mounted
 */
onMounted(async () => {
  try {
    const response = await api.get('/users')
    users.value = response.data

    loggingStore.info('Vartotojų sąrašas užkrautas', {
      component: 'Users',
      action: 'load_users',
      count: response.data.length
    })
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida įkeliant vartotojų sąrašą',
      type: 'alert-error'
    })

    loggingStore.error('Klaida įkeliant vartotojų sąrašą', {
      component: 'Users',
      action: 'load_users_failed',
      error: error.response?.data || error.message
    })
  }
})
</script>

<style scoped>
.badge {
  transition: all 0.2s ease;
}

.badge:hover {
  transform: scale(1.05);
  opacity: 0.9;
  cursor: pointer;
}
</style>
