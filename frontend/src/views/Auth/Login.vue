<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .card.card-side.glass.shadow-xl
    figure.max-w-80
      img(src="../../assets/yopta.webp" alt="Yopta")
    .card-body.p-4
      form(@submit.prevent="login")
        fieldset.fieldset.w-xs.border.border-base-300.rounded-box.p-4
          legend.fieldset-legend.text-xl Prisijungimas
          label.input
            svg(
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 16 16"
              fill="currentColor"
              class="h-4 w-4 opacity-70"
            )
              path(d="M8 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM12.735 14c.618 0 1.093-.561.872-1.139a6.002 6.002 0 0 0-11.215 0c-.22.578.254 1.139.872 1.139h9.47Z")
            input(v-model="email" type="email" class="grow" placeholder="El. paštas" required)
          label.input.mt-2
            svg(
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 16 16"
              fill="currentColor"
              class="h-4 w-4 opacity-70"
            )
              path(
                fill-rule="evenodd"
                d="M14 6a4 4 0 0 1-4.899 3.899l-1.955 1.955a.5.5 0 0 1-.353.146H5v1.5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1-.5-.5v-2.293a.5.5 0 0 1 .146-.353l3.955-3.955A4 4 0 1 1 14 6Zm-4-2a.75.75 0 0 0 0 1.5.5.5 0 0 1 .5.5.75.75 0 0 0 1.5 0 2 2 0 0 0-2-2Z"
                clip-rule="evenodd"
              )
            input(v-model="password" type="password" placeholder="Slaptažodis" class="grow" required)
          .card-actions.justify-center.mt-12
            button.btn.btn-outline.flex-1.mx-2(type="reset" @click="resetForm") Išvalyti
            button.btn.btn-outline.btn-primary.flex-1.mx-2(type="submit" :disabled="isLoggingIn")
              span.loading.loading-spinner.loading-sm(v-if="isLoggingIn")
              | Prisijungti
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { useSlogStore } from '@/stores/slog'

const email = ref('')
const password = ref('')
const isLoggingIn = ref(false)
const router = useRouter()
const authStore = useAuthStore()
const slogStore = useSlogStore()

/**
 * Resets form fields to their initial empty state
 */
const resetForm = () => {
  email.value = ''
  password.value = ''
}

/**
 * Handles the authentication process
 * Sends credentials to server, manages tokens and redirects on success
 * Displays appropriate error messages on failure
 */
const login = async () => {
  if (isLoggingIn.value) return

  isLoggingIn.value = true

  try {
    const response = await api.post('/login', {
      email: email.value,
      password: password.value
    })

    const { access_token, refresh_token } = response.data
    authStore.setAccessToken(access_token)
    authStore.setRefreshToken(refresh_token)

    slogStore.addToast({
      message: 'Sėkmingai prisijungėte prie sistemos.',
      type: 'alert-success',
      log: `method: login, email: ${email.value}, result: success`
    })

    router.push({ name: 'profile' })
  } catch (error) {
    if (error.response) {
      if (error.response.status === 403 && error.response.data === "Email not verified") {
        slogStore.addToast({
          message: 'Prašome patvirtinti el. paštą prieš prisijungiant',
          type: 'alert-warning',
          log: `method: login, email: ${email.value}, result: email_not_verified`
        })
      }
      else {
        slogStore.addToast({
          message: error.response.data,
          type: 'alert-error',
          log: `method: login, email: ${email.value}, result: ${error.response.data}`
        })
      }
    }
    else {
      slogStore.addToast({
        message: 'Kažkas nepavyko. Bandykite dar kartą.',
        type: 'alert-error',
        log: `method: login, email: ${email.value}, result: ${error.message || 'unknown_error'}`
      })
    }
  } finally {
    isLoggingIn.value = false
  }
}

/**
 * Checks if user is already authenticated on component mount
 * Redirects to profile page if authentication is valid
 */
onMounted(() => {
  if (api.clearCache) {
    api.clearCache();
  }

  if (authStore.isAuthenticated) {
    router.push({ name: 'profile' })
  }
})
</script>

<style scoped>
.form-control {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  margin-top: 1rem;
}
</style>
