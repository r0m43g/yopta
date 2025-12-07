<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .card.bg-base-100.shadow-xl(class="w-11/12 h-11/12")
    .card-header.mx-4.my-4
      h1.card-title.text-4xl Profilis
    .card-body
      .flex.flex-row.items-center.justify-center.gap-6
        .flex.flex-col.items-center
          .avatar.mb-4
            .w-48.h-48.rounded-full.ring.ring-primary.ring-offset-base-100.ring-offset-2.overflow-hidden.bg-base-300
              CachedAvatar(
                :src="userAvatar"
                alt="Vartotojo nuotrauka"
                :defaultSrc="defaultAvatar"
                size="100%"
                :rounded="true"
              )
          .flex.flex-col.items-center.mb-6
            label.btn.btn-outline.btn-primary.mb-2
              input.hidden(type="file" accept="image/*" @change="handleImageUpload")
              svg.h-5.w-5.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z")
              | Įkelti nuotrauką
            button.btn.btn-sm.btn-ghost.text-error(
              v-if="isImageLoaded"
              @click="removeImage"
            ) Pašalinti nuotrauką
          .stats.shadow
            .stat
              .stat-title Vartotojo vardas
              .stat-value.text-2xl.truncate.max-w-64 {{username}}
            .stat
              .stat-title El. paštas
              .stat-value.text-xl.truncate.max-w-64 {{email}}
          .mt-4(v-if="imagePreview && !isImageSaved")
            p.text-sm.mb-2 Nuotraukos peržiūra:
            .flex.justify-center.mb-2
              img.h-36.w-36.object-cover.rounded-lg(:src="imagePreview" alt="Nuotraukos peržiūra")
            .flex.gap-2
              button.btn.btn-sm.btn-error(@click="cancelImageUpload") Atšaukti
              button.btn.btn-sm.btn-success(@click="saveAvatar") Išsaugoti
        .divider-horizontal
        .flex.flex-row.gap-6.w-auto
          form.flex.flex-col.align-center.justify-center.mx-auto(@submit.prevent="changePassword")
            fieldset.fieldset.border.border-base-300.rounded-box.p-4
              legend.fieldset-legend.text-xl Keisti slaptažodį
              label.input.input-bordered.flex.items-center.my-2
                svg.h-4.w-4.opacity-70(
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 16 16"
                  fill="currentColor"
                )
                  path(
                    fill-rule="evenodd"
                    d="M14 6a4 4 0 0 1-4.899 3.899l-1.955 1.955a.5.5 0 0 1-.353.146H5v1.5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1-.5-.5v-2.293a.5.5 0 0 1 .146-.353l3.955-3.955A4 4 0 1 1 14 6Zm-4-2a.75.75 0 0 0 0 1.5.5.5 0 0 1 .5.5.75.75 0 0 0 1.5 0 2 2 0 0 0-2-2Z"
                    clip-rule="evenodd"
                  )
                input.grow(v-model="oldPassword" type="password" placeholder="Dabartinis slaptažodis" required)
              label.input.input-bordered.flex.items-center.my-2
                svg.h-4.w-4.opacity-70(
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 16 16"
                  fill="currentColor"
                )
                  path(
                    fill-rule="evenodd"
                    d="M14 6a4 4 0 0 1-4.899 3.899l-1.955 1.955a.5.5 0 0 1-.353.146H5v1.5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1-.5-.5v-2.293a.5.5 0 0 1 .146-.353l3.955-3.955A4 4 0 1 1 14 6Zm-4-2a.75.75 0 0 0 0 1.5.5.5 0 0 1 .5.5.75.75 0 0 0 1.5 0 2 2 0 0 0-2-2Z"
                    clip-rule="evenodd"
                  )
                input.grow(v-model="password" type="password" placeholder="Naujas slaptažodis" required)
              password-strength(:password="password")
              label.input.input-bordered.flex.items-center.my-2
                svg.h-4.w-4.opacity-70(
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 16 16"
                  fill="currentColor"
                )
                  path(
                    fill-rule="evenodd"
                    d="M14 6a4 4 0 0 1-4.899 3.899l-1.955 1.955a.5.5 0 0 1-.353.146H5v1.5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1-.5-.5v-2.293a.5.5 0 0 1 .146-.353l3.955-3.955A4 4 0 1 1 14 6Zm-4-2a.75.75 0 0 0 0 1.5.5.5 0 0 1 .5.5.75.75 0 0 0 1.5 0 2 2 0 0 0-2-2Z"
                    clip-rule="evenodd"
                  )
                input.grow(v-model="confirmPassword" type="password" placeholder="Pakartokite slaptažodį" required)
              button.btn.btn-primary.grow.my-2(type="submit") Keisti slaptažodį
          .flex.flex-col.items-center.justify-center.mx-4
            fieldset.fieldset.border.border-base-300.rounded-box.p-4
              legend.fieldset-legend.text-xl Tema
              .grid.grid-cols-3.gap-2
                label.flex.gap-2.cursor-pointer.items-center(v-for="theme in themes" :key="theme.value")
                  input.radio.radio-sm.theme-controller(
                    type="radio"
                    name="theme-radios"
                    :value="theme.value"
                    :checked="selectedTheme === theme.value"
                    @change="changeTheme(theme.value)"
                  )
                  span.text-lg {{theme.label}}
      .card-actions.justify-end.mt-6
        button.btn.btn-primary(@click="logout") Atsijungti
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSlogStore } from '@/stores/slog'
import api from '@/services/api'
import PasswordStrength from '@/components/Auth/PasswordStrength.vue'
import { optimizeImage, validateImage } from '@/services/imageUtils'
import CachedAvatar from '@/components/common/CachedAvatar.vue'

// Router instance for navigation
const router = useRouter()

// Password change state variables
const oldPassword = ref('')
const password = ref('')
const confirmPassword = ref('')

// Theme selection state variables
const selectedTheme = ref('')

// Avatar management state variables
const userAvatar = ref(null)
const imagePreview = ref(null)
const avatarBase64 = ref(null)
const isImageLoaded = ref(false)
const isImageSaved = ref(false)
const defaultAvatar = ref('/avatars/yopta.webp')

// Store instances
const authStore = useAuthStore()
const slogStore = useSlogStore()

// Available themes
const themes = ref([
  { value: 'light', label: 'Šviesi' },
  { value: 'dark', label: 'Tamsi' },
  { value: 'cupcake', label: 'Keksiukas' },
  { value: 'bumblebee', label: 'Kamanė' },
  { value: 'emerald', label: 'Smaragdas' },
  { value: 'corporate', label: 'Korporatyvinė' },
  { value: 'synthwave', label: 'Sintezatorius' },
  { value: 'retro', label: 'Retro' },
  { value: 'cyberpunk', label: 'Kibernetinis' },
  { value: 'valentine', label: 'Valentinas' },
  { value: 'halloween', label: 'Helovynas' },
  { value: 'garden', label: 'Sodas' },
  { value: 'forest', label: 'Miškas' },
  { value: 'aqua', label: 'Vanduo' },
  { value: 'lofi', label: 'Lo-Fi' },
  { value: 'pastel', label: 'Pastelinė' },
  { value: 'fantasy', label: 'Fantazija' },
  { value: 'wireframe', label: 'Vielos rėmas' },
  { value: 'black', label: 'Juoda' },
  { value: 'luxury', label: 'Prabanga' },
  { value: 'dracula', label: 'Drakula' },
  { value: 'cmyk', label: 'CMYK' },
  { value: 'autumn', label: 'Ruduo' },
  { value: 'business', label: 'Verslas' },
  { value: 'acid', label: 'Rūgštis' },
  { value: 'lemonade', label: 'Limonadas' },
  { value: 'night', label: 'Naktis' },
  { value: 'coffee', label: 'Kava' },
  { value: 'winter', label: 'Žiema' },
  { value: 'dim', label: 'Priteminta' },
  { value: 'nord', label: 'Šiaurė' },
  { value: 'sunset', label: 'Saulėlydis' },
  { value: 'caramellatte', label: 'Karamelinis' },
  { value: 'abyss', label: 'Bedugnė' },
  { value: 'silk', label: 'Šilkas' }
])

// Computed properties
const username = computed(() => authStore.username)
const email = computed(() => authStore.email)

/**
 * Handles user logout by clearing authentication tokens and redirecting.
 * Cleans up various caches and state before signing out.
 */
const logout = () => {
  const authStore = useAuthStore()

  if (api.clearCache) {
    api.clearCache();
  }

  if (loggingStore.clearLogs) {
    loggingStore.clearLogs();
  }

  authStore.clearToken()
}

/**
 * Processes password change request if passwords match.
 * Sends update request to server and handles success/failure states.
 */
const changePassword = () => {
  if (password.value !== confirmPassword.value) {
    slogStore.addToast({
      message: 'Slaptažodžiai nesutampa',
      type: 'alert-error'
    })
    return
  }

  api.post('/change-password', {
    old_password: oldPassword.value,
    new_password: password.value
  })
  .then(response => {
    slogStore.addToast({
      message: 'Slaptažodis sėkmingai pakeistas',
      type: 'alert-success'
    })

    oldPassword.value = ''
    password.value = ''
    confirmPassword.value = ''
  })
  .catch(error => {
    slogStore.addToast({
      message: error.response?.data || 'Klaida keičiant slaptažodį',
      type: 'alert-error'
    })
  })
}

/**
 * Updates theme preference for user interface.
 * Applies theme immediately and saves preference to server.
 *
 * @param {string} theme - The theme identifier to apply
 */
const changeTheme = (theme) => {
  selectedTheme.value = theme
  document.documentElement.setAttribute('data-theme', theme)

  api.put('/profile-theme', {
    id: authStore.getID,
    theme
  })
  .then(response => {
    slogStore.addToast({
      message: 'Tema sėkmingai pakeista',
      type: 'alert-success'
    })
  })
  .catch(error => {
    console.error('Klaida keičiant temą', error)
  })
}

/**
 * Processes image upload, validates, optimizes, and prepares for preview.
 * Validates file type and size before processing.
 *
 * @param {Event} e - The file input change event
 */
const handleImageUpload = async (e) => {
  const file = e.target.files[0]
  if (!file) return

  const validation = validateImage(file, {
    maxSizeKB: 5120, // 5MB
    allowedTypes: ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
  })

  if (!validation.valid) {
    slogStore.addToast({
      message: validation.message,
      type: 'alert-error'
    })
    return
  }

  try {
    const result = await optimizeImage(file, {
      maxWidth: 1000,
      maxHeight: 1000,
      quality: 0.85,
      format: 'jpeg',
      cropToSquare: true
    })

    imagePreview.value = result.base64
    avatarBase64.value = result.base64
    isImageLoaded.value = true
    isImageSaved.value = false
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida apdorojant nuotrauką: ' + error.message,
      type: 'alert-error'
    })
  }
}

/**
 * Uploads optimized avatar image to server and updates UI.
 * Updates local model and store on successful upload.
 */
const saveAvatar = async () => {
  try {
    const response = await api.put('/profile-avatar', {
      avatar: avatarBase64.value
    })

    userAvatar.value = response.data.avatar
    authStore.avatar = response.data.avatar

    isImageSaved.value = true
    imagePreview.value = null

    slogStore.addToast({
      message: 'Nuotrauka sėkmingai atnaujinta',
      type: 'alert-success'
    })
  } catch (error) {
    slogStore.addToast({
      message: error.response?.data || 'Klaida išsaugant nuotrauką',
      type: 'alert-error'
    })
  }
}

/**
 * Cancels the current image upload and clears preview.
 * Reverts to previous avatar state without saving changes.
 */
const cancelImageUpload = () => {
  imagePreview.value = null
  avatarBase64.value = null
  isImageLoaded.value = userAvatar.value && userAvatar.value !== 'none'
}

/**
 * Removes user avatar from profile and updates UI.
 * Sends deletion request to server and updates local state.
 */
const removeImage = async () => {
  try {
    const response = await api.delete('/profile-avatar')

    userAvatar.value = 'none'
    imagePreview.value = null
    avatarBase64.value = null
    isImageLoaded.value = false
    isImageSaved.value = false

    authStore.avatar = null

    slogStore.addToast({
      message: 'Nuotrauka sėkmingai pašalinta',
      type: 'alert-success'
    })
  } catch (error) {
    slogStore.addToast({
      message: error.response?.data || 'Klaida šalinant nuotrauką',
      type: 'alert-error'
    })
  }
}

/**
 * Initializes the component by loading user profile data.
 * Sets up avatar, theme, and other user preferences on mount.
 */
onMounted(async () => {
  try {
    const response = await api.get('/profile')

    authStore.setUser(response.data)
    selectedTheme.value = authStore.getTheme

    userAvatar.value = response.data.avatar

    document.documentElement.setAttribute('data-theme', response.data.theme)

    if (userAvatar.value && userAvatar.value !== 'none') {
      isImageLoaded.value = true
      isImageSaved.value = true
    }
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida įkeliant profilį',
      type: 'alert-error'
    })
    console.error('Klaida įkeliant profilį', error)
  }
})
</script>

<style scoped>
.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
