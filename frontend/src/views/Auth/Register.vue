<!-- frontend/src/views/Auth/Register.vue -->
<template>
  <div class="flex flex-col items-center justify-center h-screen w-screen overflow-auto scrollbar-thin">
    <div role="tablist" class="tabs tabs-border mb-8 tabs-lg">
      <a href="#" class="tab"
        :class="{ 'tab-active': regActive }"
        @click="switchTab">Registracija</a>
      <a href="#" class="tab"
        :class="{ 'tab-active': !regActive }"
        @click="switchTab">Patvirtinimas</a>
    </div>
    <div class="flex flex-row items-center justify-center container">
      <div class="card lg:card-side glass shadow-xl" v-show="regActive" key="reg">
        <figure class="max-w-80">
          <img src="../../assets/yopta-tablet.webp" alt="Vaizdo užsklanda" class="shadow-lg" />
        </figure>
        <div class="card-body p-4">
          <form @submit.prevent="register">
            <fieldset class="fieldset w-xs border border-base-300 rounded-box p-4">
              <legend class="fieldset-legend text-xl">Registracija</legend>
              <SafeInput
                v-model="username"
                label="Vartotojo vardas"
                placeholder="Įveskite vartotojo vardą"
                pattern="^[a-zA-Z0-9_-]{3,20}$"
                errorMessage="Vardas turi būti 3-20 simbolių (raidės, skaičiai, - ir _)"
              />
              <SafeInput
                v-model="email"
                type="email"
                label="El. paštas"
                placeholder="pavyzdys@domenas.lt"
                pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
                errorMessage="Įveskite teisingą el. paštą"
              />
              <label class="input input-bordered flex items-center gap-2 my-2">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 16 16"
                  fill="currentColor"
                  class="h-4 w-4 opacity-70">
                  <path
                    fill-rule="evenodd"
                    d="M14 6a4 4 0 0 1-4.899 3.899l-1.955 1.955a.5.5 0 0 1-.353.146H5v1.5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1-.5-.5v-2.293a.5.5 0 0 1 .146-.353l3.955-3.955A4 4 0 1 1 14 6Zm-4-2a.75.75 0 0 0 0 1.5.5.5 0 0 1 .5.5.75.75 0 0 0 1.5 0 2 2 0 0 0-2-2Z"
                    clip-rule="evenodd" />
                </svg>
                <input v-model="password" type="password" placeholder="Slaptažodis" class="grow" required />
              </label>
              <password-strength :password="password" />
              <label class="input input-bordered flex items-center gap-2 my-2">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 16 16"
                  fill="currentColor"
                  class="h-4 w-4 opacity-70">
                  <path
                    fill-rule="evenodd"
                    d="M14 6a4 4 0 0 1-4.899 3.899l-1.955 1.955a.5.5 0 0 1-.353.146H5v1.5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1-.5-.5v-2.293a.5.5 0 0 1 .146-.353l3.955-3.955A4 4 0 1 1 14 6Zm-4-2a.75.75 0 0 0 0 1.5.5.5 0 0 1 .5.5.75.75 0 0 0 1.5 0 2 2 0 0 0-2-2Z"
                    clip-rule="evenodd" />
                </svg>
                <input v-model="confirmPassword" type="password" placeholder="Pakartokite slaptažodį" class="grow" required />
              </label>
              <div class="card-actions mt-4">
                <button class="btn btn-outline flex-1 mx-2" type="reset">Atstatyti</button>
                <button class="btn btn-outline btn-primary flex-1 mx-2" type="submit" active="!isWaitingForRegistration">
                  <span v-if="!isWaitingForRegistration" class="">Registruotis</span>
                  <span v-if="isWaitingForRegistration" class="loading loading-dots loading-sm"></span>
                </button>
              </div>
            </fieldset>
          </form>
        </div>
      </div>
      <div class="card lg:card-side glass shadow-xl" v-show="!regActive" key="ver">
        <figure class="max-w-80">
          <img src="../../assets/yopta-letter.webp" alt="Laiško vaizdas" class="shadow-lg" />
        </figure>
        <div class="card-body p-4">
          <form @submit.prevent="verify">
            <fieldset class="fieldset w-xs border border-base-300 rounded-box p-4">
              <legend class="fieldset-legend text-xl">Patvirtinimas</legend>
              <label class="input input-bordered flex items-center gap-2 my-2">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 16 16"
                  fill="currentColor"
                  class="h-4 w-4 opacity-70">
                  <path
                    d="M8 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM12.735 14c.618 0 1.093-.561.872-1.139a6.002 6.002 0 0 0-11.215 0c-.22.578.254 1.139.872 1.139h9.47Z" />
                </svg>
                <input v-model="verificationCode" type="text" class="grow" placeholder="Patvirtinimo kodas" required />
              </label>
              <div class="card-actions mt-24">
                <button class="btn btn-outline flex-1 mx-2" type="reset">Atstatyti</button>
                <button class="btn btn-outline btn-primary flex-1 mx-2" type="submit">Patvirtinti</button>
              </div>
            </fieldset>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import api from '@/services/api'
import PasswordStrength from '@/components/Auth/PasswordStrength.vue'
import SafeInput from '@/components/common/SafeInput.vue'
import { sanitizeData } from '@/utils/xssSanitizer'
import { useSlogStore } from '@/stores/slog'

// State variables - because our app needs a memory better than a goldfish
const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const regActive = ref(true)  // Controls which tab is active - registration or verification
const verificationCode = ref('')
const isWaitingForRegistration = ref(false) // Indicates if registration is in progress
const slogStore = useSlogStore()

/**
 * Toggles between registration and verification tabs
 * Like a digital light switch that controls which room is illuminated
 * "Honey, I'm in the verification room now!"
 */
const switchTab = () => {
  regActive.value = !regActive.value
}

/**
 * Handles the registration process
 * Validates inputs, sanitizes data, sends registration request,
 * and switches to verification tab when successful
 * It's like a bouncer who checks your ID before letting you into the club
 */
const register = async () => {
  // Check if passwords match - because consistency is the hobgoblin of secure systems
  if (password.value !== confirmPassword.value) {
    slogStore.addToast({
      message: 'Slaptažodžiai nesutampa',
      type: 'alert-error',
      log: `method: register, message: Slaptažodžiai nesutampa`
    })
    return
  }

  try {
    // Sanitize data - washing our inputs with digital soap to remove XSS germs
    const safeData = sanitizeData({
      username: username.value,
      email: email.value,
      password: password.value
    })

    isWaitingForRegistration.value = true
    // Send registration request - launching our data into the server void
    let resp = await api.post('/register', safeData)

    // Clear inputs and switch to verification tab - clean slate policy
    clearInputs()
    regActive.value = false
    isWaitingForRegistration.value = false
    // Show success toast - digital confetti for your achievement
    slogStore.addToast({
      message: resp.data,
      type: 'alert-success',
      log: `method: register, username: ${username.value}, email: ${email.value}, result: success`
    })
  } catch (error) {
    // Show error toast - digital face palm moment
    slogStore.addToast({
      message: error.response.data,
      type: 'alert-error',
      log: `method: register, username: ${username.value}, email: ${email.value}, result: ${error.response.data}`
    })
  }
}

/**
 * Handles the email verification process
 * Sends verification code to server and displays result
 * It's like the postal service confirming you actually live at your address
 */
const verify = async () => {
  try {
    // Send verification request - "Yes, I am who I claim to be!"
    let resp = await api.post('/verify-email', { verify_code: verificationCode.value })

    // Clear inputs - wiping the digital slate clean
    clearInputs()

    // Show success toast - virtual high-five for completing verification
    slogStore.addToast({
      message: resp.data,
      type: 'alert-success',
      log: `method: verify, verify_code: ${verificationCode.value}, result: success`
    })
  } catch (error) {
    // Show error toast - the digital equivalent of tripping at the finish line
    slogStore.addToast({
      message: error.response.data,
      type: 'alert-error',
      log: `method: verify, verify_code: ${verificationCode.value}, result: ${error.response.data}`
    })
  }
}

/**
 * Clears all input fields
 * Like using a giant digital eraser on all your form data
 * "Was I ever here? No evidence remains!"
 */
const clearInputs = () => {
  username.value = ''
  email.value = ''
  password.value = ''
  confirmPassword.value = ''
  verificationCode.value = ''
}
</script>

<style scoped>
.container {
  height: 430px;
}
</style>
