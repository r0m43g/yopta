// frontend/src/main.js
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import './index.css' // Tailwind & DaisyUI styles - because naked apps are just embarrassing
import { safeHtml } from './utils/xssSanitizer'
import XssPlugin from './plugins/xssPlugin'
import LoggerPlugin from './plugins/loggerPlugin'
import logger from './services/logger'

// Create the app - it's alive! ALIVE!!! *mad scientist laughter*
const app = createApp(App)
app.use(XssPlugin) // Because letting users inject scripts is like letting toddlers play with matches
app.directive('safeHtml', safeHtml)

// Initialize Pinia - our state management superhero, no cape required
const pinia = createPinia()
app.use(pinia)

// Import our state stores - where all the app's secrets are kept
import { useConnectionStore } from './stores/connection'
import { useLoggingStore } from './stores/logging'

// Initialize stores - like waking up grumpy roommates
const connectionStore = useConnectionStore()
const loggingStore = useLoggingStore()

// Load logging settings and initialize - preparing our detective to track down bugs
loggingStore.initialize()

// Development mode special treatment - because developers need extra love
if (process.env.NODE_ENV === 'development') {
  try {
    // Expand debug panel by default in dev mode - we're not hiding anything here!
    localStorage.setItem('debug_panel_expanded', '0')
    localStorage.setItem('debug_panel_visible', '1')
  } catch (e) {
    console.error('Error setting debug panel state', e)
  }
}

// Determine if we're in development mode - different rules apply when you're experimenting
const isDevelopment = process.env.NODE_ENV === 'development'

// Initialize logging plugin - our trusty stenographer recording the app's life story
app.use(LoggerPlugin, {
  router,
  prevErrorHandler: app.config.errorHandler,
  isDevelopment
})

// Add global v-lazy directive for lazy loading - because efficiency matters and laziness is a virtue
app.directive('lazy', {
  mounted(el, binding) {
    // This function loads the resource when needed - procrastination at its finest
    function loadResource() {
      if (binding.value) {
        if (el.tagName === 'IMG') {
          el.src = binding.value
        } else {
          el.style.backgroundImage = `url(${binding.value})`
        }
      }
    }

    // Handle intersection observer callback - "Hey, is this thing visible yet?"
    function handleIntersect(entries, observer) {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          loadResource()
          observer.unobserve(el) // Our work here is done, moving on!
        }
      })
    }

    // Use IntersectionObserver if available - like having a lookout for your images
    if ('IntersectionObserver' in window) {
      const observer = new IntersectionObserver(handleIntersect, {
        rootMargin: '50px 0px', // A bit of buffer, because we're generous like that
        threshold: 0.1 // Just a peek is all we need
      })
      observer.observe(el)
    } else {
      // Fallback for browsers stuck in the stone age - no lazy loading for you!
      loadResource()
    }
  }
})

// Directive for logging UI clicks - because we're nosy and want to know what users are clicking
app.directive('log-click', {
  mounted(el, binding) {
    el.addEventListener('click', () => {
      // Check if UI event logging is enabled - respect privacy settings, we're not animals
      if (!loggingStore.settings.logUIEvents) {
        return;
      }

      // Get component from context or fallback to "Unknown" - name tags are important at parties
      const component = binding.instance?.$options?.name || 'Unknown'
      // Get directive value or default to 'clicked' - because "clicked" is better than "interacted with"
      const action = binding.value || 'clicked'

      // Log click through store - documenting user's adventures
      loggingStore.uiEvent(component, action, {
        element: el.tagName,
        class: el.className,
        id: el.id,
        text: el.textContent?.trim()
      })
    })
  }
})

// Mount the app to our router - let's get this party started!
app.use(router)

// Create a variable for throttling timer - preventing log flooding, because server admins have feelings too
let logThrottleTimer = null;

// Function for throttled log processing - like spacing out your texts to not seem desperate
const processLogs = () => {
  if (logThrottleTimer) {
    clearTimeout(logThrottleTimer);
  }

  logThrottleTimer = setTimeout(() => {
    // In development mode, we don't send logs to server - what happens in dev stays in dev
    if (!isDevelopment) {
      logger.processPendingLogs();
    }
  }, 10000); // Send logs no more than once every 10 seconds - patience is a virtue
};

// Log application start - Hello World! The app has entered the building!
logger.info('Application started', {
  timestamp: new Date().toISOString(),
  environment: process.env.NODE_ENV,
  version: import.meta.env.VITE_APP_VERSION || 'development'
})

// Track server connection for processing pending logs - waiting for the mothership to come online
connectionStore.$subscribe((mutation, state) => {
  if (state.isConnected) {
    // When connection is established, process pending logs with throttling
    // Like waiting for the perfect moment to send all those messages you drafted offline
    processLogs();
  }
})

// Handle all unhandled errors through logging system - catching bugs that tried to escape
window.addEventListener('error', event => {
  // Check if error is related to logging - avoiding infinite loops like a sensible time traveler
  const errorMessage = event.message || '';

  // Ignore errors related to logging to prevent cycles - don't chase your own tail
  if (errorMessage.includes('/client-logs') ||
      errorMessage.includes('Failed to fetch') ||
      errorMessage.includes('Network Error')) {
    return;
  }

  // Log the error - documenting the crime scene
  loggingStore.error(event.error || new Error(event.message), {
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    type: 'unhandled_error'
  })
})

// Global error handling - because errors should be embraced, not ignored
app.config.errorHandler = (err, vm, info) => {
  // Ignore logging-related errors to prevent cycles - avoid the snake eating its own tail
  const errorMessage = err?.message || '';

  if (errorMessage.includes('/client-logs') ||
      errorMessage.includes('Failed to fetch') ||
      errorMessage.includes('Network Error')) {
    // Just log to console - the bare minimum for our sanity
    console.error('Registravimo klaida:', err);
    return;
  }

  console.error('Globali programos klaida:', err)
  console.error('Komponento informacija:', vm)
  console.error('Klaidos informacija:', info)

  // Log error through store - immortalizing the bug for future archaeologists
  loggingStore.error(err, {
    component: vm?.$options?.name || 'Unknown',
    props: vm?.$props,
    info: info,
    route: router.currentRoute.value?.path
  })
}

// Log promise errors with throttling - even asynchronous bugs can't hide
let promiseErrorCount = 0;
let lastPromiseErrorTime = 0;

window.addEventListener('unhandledrejection', event => {
  // Check if error is related to logging - no logging loops allowed in this neighborhood
  const errorMessage = event.reason?.message || String(event.reason) || '';

  // Ignore logging-related errors to prevent cycles - don't feed the infinity monster
  if (errorMessage.includes('/client-logs') ||
      errorMessage.includes('Failed to fetch') ||
      errorMessage.includes('Network Error')) {
    return;
  }

  // Throttle identical errors - we heard you the first five times, thanks
  const now = Date.now();
  if (now - lastPromiseErrorTime < 1000) { // 1 second
    promiseErrorCount++;

    // If more than 5 identical errors per second, ignore - enough is enough
    if (promiseErrorCount > 5) {
      return;
    }
  } else {
    // Reset counter - new second, new opportunities to fail!
    promiseErrorCount = 1;
    lastPromiseErrorTime = now;
  }

  let error = event.reason;
  if (!(error instanceof Error)) {
    error = new Error(String(error));
  }

  loggingStore.error(error, {
    type: 'unhandled_rejection',
    promise: 'Neapdorota pažado klaida'
  })
})

// In development mode, we don't log window resizes - developers resize windows constantly, like fidgeting
if (isDevelopment) {
  // Disable window resize logging - saving trees one log entry at a time
} else {
  // Log window size changes for adaptive analysis - stalking your window for science
  // with throttling applied
  let resizeTimer = null;
  window.addEventListener('resize', () => {
    if (resizeTimer) {
      clearTimeout(resizeTimer);
    }

    resizeTimer = setTimeout(() => {
      loggingStore.info('Lango dydis pakeistas', {
        width: window.innerWidth,
        height: window.innerHeight,
        type: 'window_resize'
      });
    }, 1000); // Throttle to one event per second - windows need privacy too
  });
}

// Mount the app - ignition sequence start, we have liftoff!
app.mount('#app')
