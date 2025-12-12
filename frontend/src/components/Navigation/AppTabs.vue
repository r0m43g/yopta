<!-- frontend/src/components/Navigation/AppTabs.vue -->
<template>
  <div class="navbar bg-base-100 unprintable">
    <div class="flex-1">
      <a class="btn btn-ghost text-2xl" href="/">Yopta<sup class="text-xs">TOP</sup></a>
    </div>
    <div class="flex-none gap-2">
      <ul class="menu menu-horizontal px-1">
        <li v-if="!links['login'].auth && !loggedIn">
          <router-link :to="links['login'].route"> {{ links['login'].name }}</router-link>
        </li>
        <li v-if="!links['register'].auth && !loggedIn && registrationEnabled">
        <router-link :to="links['register'].route"> {{ links['register'].name }}</router-link>
      </li>
        <li v-if="links['klasika'].auth && loggedIn">
          <router-link :to="links['klasika'].route"> {{ links['klasika'].name }}</router-link>
        </li>
        <li v-if="links['second'].auth && loggedIn">
          <router-link :to="links['second'].route"> {{ links['second'].name }}</router-link>
        </li>
      </ul>
      <div class="dropdown dropdown-end" v-if="loggedIn">
        <div tabindex="0" role="button" class="btn btn-ghost btn-circle avatar">
          <div class="w-10 rounded-full">
            <CachedAvatar
              :src="avatar"
              alt="Naudotojo avataras"
              defaultSrc="/avatars/yopta.webp"
              size="40"
              :rounded="true"
            />
          </div>
        </div>
        <ul
          tabindex="0"
          class="menu menu-sm dropdown-content bg-base-100 rounded-box z-[1] mt-3 w-52 p-2 shadow">
          <li v-if="isAdmin">
            <router-link class="justify-center" :to="links['stations'].route">
              {{ links['stations'].name }}
            </router-link>
          </li>
          <li v-if="isAdmin">
            <router-link class="justify-center" :to="links['fieldMappings'].route">
              {{ links['fieldMappings'].name }}
            </router-link>
          </li>
          <li v-if="isAdmin">
            <router-link class="justify-center" :to="links['antrasFieldMappings'].route">
              {{ links['antrasFieldMappings'].name }}
            </router-link>
          </li>
          <li>
            <router-link class="justify-center" :to="links['profile'].route">
              {{ links['profile'].name }}
            </router-link>
          </li>
          <li v-if="isAdmin">
            <router-link class="justify-center" :to="links['users'].route">
              {{ links['users'].name }}
            </router-link>
            <router-link class="justify-center" :to="links['logs'].route">
              {{ links['logs'].name }}
            </router-link>
          </li>
          <li v-if="isAdmin">
            <router-link class="justify-center" :to="links['clientLogs'].route">
              {{ links['clientLogs'].name }}
            </router-link>
          </li>
          <li v-if="isAdmin">
            <router-link class="justify-center" :to="links['systemSettings'].route">
              {{ links['systemSettings'].name }}
            </router-link>
          </li>
          <li v-if="isAdmin">
            <a class="justify-center" href="#" @click.prevent="clearCaches">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              {{ links['clearCache'].name }}
            </a>
          </li>
          <li>
            <a class="justify-center" href="#" @click="logout">
              Atsijungti
            </a>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLoggingStore } from '@/stores/logging'
import { useSlogStore } from '@/stores/slog'
import api from '@/services/api'
import CachedAvatar from '@/components/common/CachedAvatar.vue'

/**
 * The Digital Signpost Collection - your guide to navigating this digital landscape!
 *
 * This object defines all navigation destinations in our application, organized like
 * a meticulously labeled map at a tourist information center. Each entry contains the
 * Lithuanian name of the destination, its route path, and whether it requires authentication.
 * Like a medieval town where certain districts are only accessible to guild members,
 * some routes are protected behind the authentication gates.
 */
const links = {
  login: {
    name: 'Prisijungimas',
    route: '/login',
    auth: false,
  },
  register: {
    name: 'Registracija',
    route: '/register',
    auth: false,
  },
  profile: {
    name: 'Profilis',
    route: '/profile',
    auth: true,
  },
  klasika: {
    name: 'Klasika',
    route: '/klasika',
    auth: true,
  },
  second: {
    name: 'Antras',
    route: '/second',
    auth: true,
  },
  users: {
    name: 'Naudotojai',
    route: '/users',
    auth: true,
  },
  logs: {
    name: 'Serverio žurnalas',
    route: '/logs',
    auth: true,
  },
  clientLogs: {
    name: 'Kliento žurnalas',
    route: '/client-logs',
    auth: true,
  },
  clearCache: {
    name: 'Išvalyti podėlį',
    route: '#',
    auth: true,
  },
  stations: {
    name: 'Stotys/Depai',
    route: '/stations',
    auth: true,
  },
  fieldMappings: {
    name: 'Laukų atvaizdavimai',
    route: '/field-mappings',
    auth: true,
  },

  antrasFieldMappings: {
    name: 'Excel laukų atvaizdavimai',
    route: '/antras-field-mappings',
    auth: true,
  },
  systemSettings: {
    name: 'Sistemos nustatymai',
    route: '/system-settings',
    auth: true,
  },
}

/**
 * The Authentication Oracle - reveals whether you're among the chosen ones!
 *
 * This computed property connects to the authentication store to determine
 * if the current user has been blessed with login credentials. Like a magical
 * amulet that glows when worn by the worthy, it illuminates (returns true)
 * when a user has a valid ticket to access the protected parts of our realm.
 */
const loggedIn = computed(() => useAuthStore().isAuthenticated)

/**
 * The Authority Detector - checks if you wield administrative powers!
 *
 * Determines if the current user possesses administrative privileges.
 * Like a medieval official checking if you have the royal seal that grants
 * access to the king's private chambers, this computed property reveals
 * whether you're among the digital elite who can make changes to the kingdom.
 */
const isAdmin = computed(() => useAuthStore().isAdmin)

/**
 * The Digital Face Retriever - fetches your personalized visual identity!
 *
 * Obtains the current user's avatar image path from the authentication store.
 * Like a royal portrait artist who has your likeness on file, this computed
 * property provides the image that represents you in this digital realm.
 */
const avatar = computed(() => useAuthStore().getAvatar)

/**
 * The Current Location Spotter - identifies where you currently stand!
 *
 * Gets the current route using Vue Router, which helps with highlighting
 * the currently active navigation tab. Like a "You Are Here" dot on a mall
 * directory, it helps users understand their current position in the application.
 */
const route = useRoute()
const router = useRouter()
const loggingStore = useLoggingStore()
/**
 * The Notification Messenger - your connection to the toast notification system!
 *
 * Provides access to the notification system for displaying alerts and messages.
 * Like a town crier with a bell who delivers proclamations throughout the kingdom,
 * this store enables components to broadcast messages to the user.
 */
const slogStore = useSlogStore()

/**
 * The Active Tab Detector - highlights your current location!
 *
 * Checks if a tab should be visually marked as active based on the current route.
 * Like a glowing marker on a map that shows your current location, this function
 * helps users understand which section of the application they're currently exploring.
 *
 * @param {String} tabRoute - The route path to check against the current location
 * @returns {Boolean} - True if this tab matches your current location
 */
const isActive = (tabRoute) => {
  return route.path.startsWith(tabRoute)
}
const registrationEnabled = ref(true);

// В onMounted добавить:
const checkRegistrationSettings = async () => {
  try {
    const response = await api.get('/system-settings');
    const registrationSetting = response.data.find(s => s.setting_key === 'registration_enabled');
    registrationEnabled.value = registrationSetting?.setting_value === 'true';
  } catch (error) {
    console.error('Не удалось проверить статус регистрации', error);
    // По умолчанию показываем ссылку при ошибке
    registrationEnabled.value = true;
  }
};

onMounted(() => {
  checkRegistrationSettings();
});

const clearCaches = async () => {
  try {
    slogStore.addToast({
      message: 'Pradedamas podėlio valymas...',  // "Начинаем очистку кеша..."
      type: 'alert-info'
    });

    if (api.clearCache) {
      api.clearCache();
    }

    if (loggingStore.clearLogs) {
      loggingStore.clearLogs();
    }

    const imgCache = localStorage.getItem('image_cache');
    if (imgCache) {
      localStorage.removeItem('image_cache');
    }

    const keysToKeep = ['accessToken', 'refreshToken'];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && !keysToKeep.includes(key) && (key.includes('cache') || key.includes('podėlis'))) {
        localStorage.removeItem(key);
      }
    }

    await api.post('/cache/clear');

    slogStore.addToast({
      message: 'Podėlis sėkmingai išvalytas!',
      type: 'alert-success'
    });

    loggingStore.info('Podėlis išvalytas administratoriaus', {
      component: 'AppTabs',
      action: 'clear_cache',
      route: router.currentRoute.value.path
    });

    const currentPath = router.currentRoute.value.path;

    if (currentPath !== '/') {
      await router.push('/');
    }

    if (currentPath !== '/') {
      setTimeout(() => {
        router.push(currentPath);
      }, 100);
    } else {
      window.location.reload();
    }
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida valant podėlį: ' + (error.response?.data || error.message),
      type: 'alert-error'
    });

    loggingStore.error('Klaida valant podėlį', {
      component: 'AppTabs',
      action: 'clear_cache_failed',
      error: error.response?.data || error.message
    });
  }
};

/**
 * The Royal Escort - politely shows you to the exit!
 *
 * Handles user logout by clearing authentication tokens and redirecting to login.
 * Like a palace guard who courteously but firmly guides visitors to the exit
 * when their audience with the king has concluded, this function ensures
 * users properly end their session when they're done using the application.
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
 * The Portrait Error Handler - gracefully deals with missing avatars!
 *
 * Replaces broken avatar images with a default placeholder.
 * Like a royal painter who can sketch a generic silhouette when the subject
 * is unavailable for a sitting, this function ensures users always have
 * some image displayed, even when their custom avatar fails to load.
 *
 * @param {Event} e - The error event triggered by a failed image load
 */
const handleAvatarError = (e) => {
  // Set default image
  e.target.src = '/assets/yopta.webp'
}
</script>

<style scoped>
/* Custom styles would go here if needed */
@media print {
  .unprintable {
    display: none;
  }
}
</style>
