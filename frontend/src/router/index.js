// frontend/src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useSlogStore } from '../stores/slog'
import { useConnectionStore } from '../stores/connection'
import api from '../services/api'

import Login from '../views/Auth/Login.vue'

/**
 * The Great Map of the Digital Kingdom - charting paths through the application!
 * These routes define the navigable territories in our SPA landscape.
 * Like a medieval cartographer plotting courses between villages and castles,
 * we establish the pathways users can travel along during their journey.
 */
const routes = [
  { path: '/', name: 'home', component: Login },

  // Proper lazy loading with direct import functions
  {
    path: '/klasika',
    name: 'klasika',
    component: () => import('../views/Klasika.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/second',
    name: 'second',
    component: () => import('../views/Second.vue'),
    meta: { requiresAuth: true }
  },
  { path: '/login', name: 'login', component: Login },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/Auth/Register.vue')
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('../views/Profile.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/forbidden',
    name: 'forbidden',
    component: () => import('../views/Forbidden.vue')
  },
  {
    path: '/users',
    name: 'users',
    component: () => import('../views/Admin/Users.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
    }
  },
  {
    path: '/logs',
    name: 'logs',
    component: () => import('../views/Admin/Logs.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
    }
  },
  {
    path: '/client-logs',
    name: 'clientLogs',
    component: () => import('../views/Admin/ClientLogs.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
    }
  },
  {
    path: '/system-settings',
    name: 'systemSettings',
    component: () => import('../views/Admin/SystemSettings.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
    }
  },
  {
    path: '/stations',
    name: 'stations',
    component: () => import('../views/Admin/Stations.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
    }
  },
  {
    path: '/:catchAll(.*)',
    component: () => import('../views/NotFound.vue'),
    name: 'notFound'
  },
]

/**
 * The Router Constructor - assembling our navigation apparatus!
 * Creates the router instance that powers the application's navigation system.
 * Like building a teleportation device that zaps users between different rooms,
 * this router materializes new views without actually changing physical pages.
 */
const router = createRouter({
  history: createWebHistory(),
  routes,
  // ScrollBehavior enhances user experience with sensible scrolling
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition // Return to previously scrolled position when using back/forward
    } else {
      return { top: 0 } // Start fresh pages at the top, like a proper gentleman
    }
  }
})

// Flag to track the token checking process and avoid infinite loops of misery
let isRefreshing = false;

/**
 * The Digital Bouncer - checks if you're allowed to enter each route!
 * Guards routes from unauthorized access and manages authentication state.
 * Like a nightclub bouncer who checks IDs before letting patrons in,
 * this function ensures users only access routes they're authorized to see.
 * It's suspiciously thorough, almost like it's seen too many fake IDs in its time.
 */
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  let token = authStore.accessToken

  // Check connection state
  const connectionStore = useConnectionStore()

  // Initialize connection monitoring on first navigation
  if (from.name === undefined) {
    connectionStore.initConnectionListeners()
  }

  if (to.meta.requiresAuth) {
    if (!token) {
      return next({ name: 'login' }) // No token? Straight to login with you!
    } else {
      if (!authStore.isAuthenticated) {
        try {
          // Check if we have a server connection
          if (!connectionStore.isConnected) {
            // If offline but we have a refresh token in localStorage,
            // allow navigation to cached pages like a merciful deity
            if (authStore.refreshToken) {
              return next()
            } else {
              return next({ name: 'login' })
            }
          }

          // If a token refresh process is already underway, wait patiently
          // like a British person queuing for tea
          if (isRefreshing) {
            // Brief pause before checking refresh results - deep breath now
            await new Promise(resolve => setTimeout(resolve, 500));
            // Check the results of our patience
            if (authStore.isAuthenticated) {
              // Token successfully refreshed, onwards brave adventurer!
              // But first, check if you're admin-material for admin routes
              if (to.meta.requiresAdmin && !authStore.isAdmin) {
                return next({ name: 'forbidden' });
              }
              return next();
            } else {
              // Token refresh failed, back to login purgatory
              return next({ name: 'login' });
            }
          }

          // Set the refreshing flag - no one else should try this right now
          isRefreshing = true;

          try {
            // Verify token with a gentle ping to the server
            await api.post('/auth-ping')
            // Success! Reset flag and proceed with dignity
            isRefreshing = false;
          } catch (refreshError) {
            // Reset flag after refresh attempt failure - like admitting defeat
            isRefreshing = false;
            console.error("Prieigos rakto tikrinimo klaida:", refreshError)
            authStore.clearToken()
            return next({ name: 'login' })
          }
        } catch (error) {
          // Reset flag after any other error - sometimes life is just hard
          isRefreshing = false;
          console.error("Autentifikavimo tikrinimo klaida:", error)
          authStore.clearToken()
          return next({ name: 'login' })
        }
      }

      // Admin routes require admin powers, unsurprisingly
      if (to.meta.requiresAdmin && !authStore.isAdmin) {
        const slogStore = useSlogStore()
        if (slogStore) {
          slogStore.addToast({
            message: 'Prieiga apribota. Tik administratoriai gali peržiūrėti šį puslapį.',
            type: 'alert-error',
          })
        }
        return next({ name: 'forbidden' })
      }
    }
  }

  if (to.name === 'register') {
    try {
      const response = await api.get('/system-settings');
      const registrationSetting = response.data.find(s => s.setting_key === 'registration_enabled');
      if (registrationSetting?.setting_value !== 'true') {
        const slogStore = useSlogStore();
        if (slogStore) {
          slogStore.addToast({
            message: 'Registracija šiuo metu išjungta',
            type: 'alert-error',
          });
        }
        return next({ name: 'login' });
      }
    } catch (error) {
      console.error("Не удалось проверить статус регистрации:", error);
      // При ошибке разрешаем доступ к странице
    }
  }
  next() // Proceed to your destination, esteemed user
})

export default router
