<template>
  <div
    v-if="isVisible"
    class="debug-panel bg-base-200 text-base-content border border-base-300 shadow-lg"
    :class="{ 'expanded': isExpanded }"
  >
    <div class="debug-panel-header flex justify-between items-center p-2 border-b border-base-300">
      <div class="flex items-center">
        <button @click="toggleExpand" class="btn btn-sm btn-ghost mr-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke-width="1.5"
            stroke="currentColor"
            class="w-4 h-4"
            :class="{ 'rotate-180': isExpanded }"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
          </svg>
          <span class="font-semibold">Derinimo skydelis</span>
        </button>
      </div>
      <div class="flex items-center">
        <span v-if="connectionStatus !== undefined" class="text-xs mr-2" :class="connectionStatusClass">
          {{ connectionStatusText }}
        </span>
        <button @click="clearLogs" class="btn btn-xs btn-ghost mr-1">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
        <button @click="closePanel" class="btn btn-xs btn-ghost">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <div v-if="isExpanded" class="debug-panel-content">
      <div class="p-2 border-b border-base-300 flex justify-between">
        <div class="tabs tabs-sm">
          <a
            class="tab"
            :class="{ 'tab-active': activeTab === 'logs' }"
            @click="activeTab = 'logs'"
          >
            Žurnalai
          </a>
          <a
            class="tab"
            :class="{ 'tab-active': activeTab === 'network' }"
            @click="activeTab = 'network'"
          >
            Tinklas
          </a>
          <a
            class="tab"
            :class="{ 'tab-active': activeTab === 'state' }"
            @click="activeTab = 'state'"
          >
            Būsena
          </a>
          <a
            class="tab"
            :class="{ 'tab-active': activeTab === 'settings' }"
            @click="activeTab = 'settings'"
          >
            Nustatymai
          </a>
        </div>

        <div class="flex items-center">
          <div class="form-control">
            <input
              type="text"
              v-model="filterText"
              placeholder="Filtras..."
              class="input input-xs input-bordered w-32"
            />
          </div>
        </div>
      </div>

      <div class="tab-contents overflow-y-auto">
        <!-- Logs tab -->
        <div v-if="activeTab === 'logs'" class="log-table">
          <table class="table table-compact w-full">
            <thead>
              <tr>
                <th class="w-24">Laikas</th>
                <th class="w-16">Lygis</th>
                <th>Žinutė</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(log, index) in filteredLogs" :key="index" @click="selectLog(log)" class="hover">
                <td class="text-xs whitespace-nowrap">{{ formatTime(log.timestamp) }}</td>
                <td>
                  <span class="badge badge-xs" :class="getLevelClass(log.level)">
                    {{ log.level }}
                  </span>
                </td>
                <td class="truncate max-w-xs">{{ log.message }}</td>
              </tr>
              <tr v-if="filteredLogs.length === 0">
                <td colspan="3" class="text-center py-4">
                  <div class="flex flex-col items-center justify-center py-4">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-base-300 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <span>Nėra žurnalų įrašų</span>
                    <button @click="createTestLog" class="btn btn-xs btn-outline mt-2">Sukurti testinį įrašą</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Network tab -->
        <div v-if="activeTab === 'network'" class="log-table">
          <table class="table table-compact w-full">
            <thead>
              <tr>
                <th class="w-24">Laikas</th>
                <th class="w-16">Metodas</th>
                <th>URL</th>
                <th class="w-16">Statusas</th>
                <th class="w-24">Trukmė</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(log, index) in filteredNetworkLogs" :key="index" @click="selectLog(log)" class="hover">
                <td class="text-xs whitespace-nowrap">{{ formatTime(log.timestamp) }}</td>
                <td>
                  <span class="badge badge-xs" :class="getMethodClass(log.data?.request?.method)">
                    {{ log.data?.request?.method || 'GET' }}
                  </span>
                </td>
                <td class="truncate max-w-xs">{{ log.data?.request?.url || '-' }}</td>
                <td>
                  <span class="badge badge-xs" :class="getStatusClass(log.data?.response?.status)">
                    {{ log.data?.response?.status || '-' }}
                  </span>
                </td>
                <td class="text-right">{{ log.data?.duration || '-' }} ms</td>
              </tr>
              <tr v-if="filteredNetworkLogs.length === 0">
                <td colspan="5" class="text-center py-4">
                  <div class="flex flex-col items-center justify-center py-4">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-base-300 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span>Nėra tinklo užklausų</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- State tab -->
        <div v-if="activeTab === 'state'" class="state-panel">
          <div class="flex flex-col p-4 gap-2">
            <div class="collapse collapse-arrow border border-base-300 bg-base-100 rounded-box">
              <input type="checkbox" />
              <div class="collapse-title text-sm font-medium">
                Autentifikacija
              </div>
              <div class="collapse-content">
                <pre class="bg-base-300 p-2 rounded text-xs overflow-x-auto">{{ authState }}</pre>
              </div>
            </div>

            <div class="collapse collapse-arrow border border-base-300 bg-base-100 rounded-box">
              <input type="checkbox" />
              <div class="collapse-title text-sm font-medium">
                Ryšys
              </div>
              <div class="collapse-content">
                <pre class="bg-base-300 p-2 rounded text-xs overflow-x-auto">{{ connectionState }}</pre>
              </div>
            </div>

            <div class="collapse collapse-arrow border border-base-300 bg-base-100 rounded-box">
              <input type="checkbox" />
              <div class="collapse-title text-sm font-medium">
                Aplinkos kintamieji
              </div>
              <div class="collapse-content">
                <pre class="bg-base-300 p-2 rounded text-xs overflow-x-auto">{{ environmentVars }}</pre>
              </div>
            </div>

            <div class="collapse collapse-arrow border border-base-300 bg-base-100 rounded-box">
              <input type="checkbox" />
              <div class="collapse-title text-sm font-medium">
                Naršyklės informacija
              </div>
              <div class="collapse-content">
                <pre class="bg-base-300 p-2 rounded text-xs overflow-x-auto">{{ browserInfo }}</pre>
              </div>
            </div>
          </div>
        </div>

        <!-- Settings tab -->
        <div v-if="activeTab === 'settings'" class="settings-panel p-4">
          <h3 class="text-sm font-semibold mb-2">Žurnalų nustatymai</h3>

          <div class="form-control mb-2">
            <label class="label cursor-pointer justify-start">
              <input type="checkbox" v-model="settings.logNetworkRequests" class="checkbox checkbox-xs mr-2" />
              <span class="label-text">Registruoti tinklo užklausas</span>
            </label>
          </div>

          <div class="form-control mb-2">
            <label class="label cursor-pointer justify-start">
              <input type="checkbox" v-model="settings.logUIEvents" class="checkbox checkbox-xs mr-2" />
              <span class="label-text">Registruoti vartotojo sąsajos įvykius</span>
            </label>
          </div>

          <div class="form-control mb-2">
            <label class="label cursor-pointer justify-start">
              <input type="checkbox" v-model="settings.logPerformance" class="checkbox checkbox-xs mr-2" />
              <span class="label-text">Registruoti našumo metrikas</span>
            </label>
          </div>

          <div class="form-control mb-2">
            <label class="label cursor-pointer justify-start">
              <input type="checkbox" v-model="settings.logErrors" class="checkbox checkbox-xs mr-2" />
              <span class="label-text">Registruoti klaidas</span>
            </label>
          </div>

          <div class="form-control mb-4">
            <label class="label">
              <span class="label-text">Minimalus žurnalo lygis</span>
            </label>
            <select v-model="settings.logLevel" class="select select-bordered select-sm w-full">
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
            </select>
          </div>

          <div class="form-control mb-4">
            <label class="label">
              <span class="label-text">Maksimalus žurnalų skaičius atmintyje</span>
            </label>
            <input
              type="number"
              v-model.number="maxLogsCount"
              min="10"
              max="1000"
              step="10"
              class="input input-bordered input-sm w-full"
            />
          </div>

          <button @click="saveSettings" class="btn btn-sm btn-primary w-full">
            Išsaugoti nustatymus
          </button>

          <div class="divider"></div>

          <button @click="downloadLogs" class="btn btn-sm btn-outline w-full mb-2">
            Atsisiųsti žurnalus
          </button>

          <button @click="clearAllData" class="btn btn-sm btn-outline btn-error w-full">
            Išvalyti visus duomenis
          </button>
        </div>
      </div>

      <!-- Log details section -->
      <div v-if="selectedLog" class="log-details border-t border-base-300 p-4">
        <div class="flex justify-between items-center mb-2">
          <h3 class="text-sm font-semibold">Detalesnė informacija</h3>
          <button @click="selectedLog = null" class="btn btn-xs btn-ghost">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="grid grid-cols-2 gap-2 mb-2">
          <div>
            <div class="text-xs font-semibold">Lygis:</div>
            <div class="text-sm">
              <span class="badge" :class="getLevelClass(selectedLog.level)">
                {{ selectedLog.level }}
              </span>
            </div>
          </div>

          <div>
            <div class="text-xs font-semibold">Laikas:</div>
            <div class="text-sm">{{ formatTime(selectedLog.timestamp, true) }}</div>
          </div>
        </div>

        <div class="mb-2">
          <div class="text-xs font-semibold">Žinutė:</div>
          <div class="text-sm bg-base-300 p-2 rounded">{{ selectedLog.message }}</div>
        </div>

        <div class="mb-2">
          <div class="text-xs font-semibold">Duomenys:</div>
          <pre class="text-xs bg-base-300 p-2 rounded overflow-x-auto whitespace-pre-wrap break-words max-h-32">{{ JSON.stringify(selectedLog.data, null, 2) }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useLoggingStore } from '@/stores/logging';
import { useConnectionStore } from '@/stores/connection';
import { useAuthStore } from '@/stores/auth';

// State for the debug panel UI
// Like the control panel of a spaceship, but for developers
const isVisible = ref(true);
const isExpanded = ref(true);
const activeTab = ref('logs');
const filterText = ref('');
const selectedLog = ref(null);
const maxLogsCount = ref(100);

// Settings for logging behavior
// The dials and switches that control our digital surveillance system
const settings = ref({
  logNetworkRequests: true,
  logUIEvents: true,
  logPerformance: true,
  logErrors: true,
  logLevel: 'debug'
});

// Get store instances
// Our personal spies that report back about the state of the application
const loggingStore = useLoggingStore();
const connectionStore = useConnectionStore();

// Filtered logs based on search criteria
// Like using a magnifying glass to find that one mosquito in your room
const filteredLogs = computed(() => {
  // Check if logs exist and can be filtered
  if (!loggingStore.recentLogs) {
    console.warn('Debug Panel: recentLogs is not available in the loggingStore');
    return [];
  }

  return loggingStore.recentLogs
    .filter(log => {
      if (!filterText.value) return true;

      const searchText = filterText.value.toLowerCase();
      return (
        (log.message && log.message.toLowerCase().includes(searchText)) ||
        (log.level && log.level.toLowerCase().includes(searchText)) ||
        (log.data && JSON.stringify(log.data).toLowerCase().includes(searchText))
      );
    });
});

// Filtered network logs for the network tab
// Finding needles in a haystack of HTTP requests
const filteredNetworkLogs = computed(() => {
  if (!loggingStore.recentLogs) return [];

  return loggingStore.recentLogs
    .filter(log => log.data?.type === 'network')
    .filter(log => {
      if (!filterText.value) return true;

      const searchText = filterText.value.toLowerCase();
      return (
        (log.message && log.message.toLowerCase().includes(searchText)) ||
        (log.data?.request?.url && log.data.request.url.toLowerCase().includes(searchText)) ||
        (log.data?.request?.method && log.data.request.method.toLowerCase().includes(searchText))
      );
    });
});

// Connection status indicators
// The digital pulse of our application's heartbeat
const connectionStatus = computed(() => {
  return connectionStore.isConnected;
});

const connectionStatusText = computed(() => {
  return connectionStore.isConnected ? 'Prisijungta' : 'Atsijungta';
});

const connectionStatusClass = computed(() => {
  return connectionStore.isConnected ? 'text-success' : 'text-error';
});

// State information for display in the state tab
// Looking under the hood of our application
const authState = computed(() => {
  const authStore = useAuthStore();
  return JSON.stringify({
    isAuthenticated: authStore.isAuthenticated,
    username: authStore.username,
    email: authStore.email,
    role: authStore.role,
    id: authStore.getID
  }, null, 2);
});

const connectionState = computed(() => {
  return JSON.stringify({
    isOnline: connectionStore.isOnline,
    serverStatus: connectionStore.serverStatus,
    lastCheck: connectionStore.lastCheck,
    isConnected: connectionStore.isConnected
  }, null, 2);
});

const environmentVars = computed(() => {
  const env = {};

  // Collect environment variables
  if (import.meta.env) {
    Object.keys(import.meta.env).forEach(key => {
      // Exclude private keys
      if (!key.includes('PRIVATE') && !key.includes('SECRET') && !key.includes('PASSWORD')) {
        env[key] = import.meta.env[key];
      }
    });
  }

  return JSON.stringify(env, null, 2);
});

const browserInfo = computed(() => {
  return JSON.stringify({
    userAgent: navigator.userAgent,
    language: navigator.language,
    platform: navigator.platform,
    vendor: navigator.vendor,
    cookiesEnabled: navigator.cookieEnabled,
    screen: {
      width: window.screen.width,
      height: window.screen.height,
      pixelRatio: window.devicePixelRatio
    },
    viewport: {
      width: window.innerWidth,
      height: window.innerHeight
    }
  }, null, 2);
});

// Toggle panel expansion state
// Transforming our debug panel from Clark Kent to Superman with one click
const toggleExpand = () => {
  isExpanded.value = !isExpanded.value;

  // Save state to localStorage
  try {
    localStorage.setItem('debug_panel_expanded', isExpanded.value ? '1' : '0');
  } catch (e) {
    console.error('Error saving debug panel state', e);
  }
};

// Close the panel completely
// The digital equivalent of throwing the debug panel into Mount Doom
const closePanel = () => {
  isVisible.value = false;

  // Save state to localStorage
  try {
    localStorage.setItem('debug_panel_visible', '0');
  } catch (e) {
    console.error('Error saving debug panel state', e);
  }
};

// Clear all logs from the display
// Marie Kondo would be proud - these logs no longer spark joy
const clearLogs = () => {
  loggingStore.clearLogs();
  selectedLog.value = null;
};

// Select a log to view details
// Like using a microscope on a particularly interesting bug
const selectLog = (log) => {
  selectedLog.value = log;
};

// Format timestamp for display
// Turning machine time into human time
const formatTime = (timestamp, full = false) => {
  if (!timestamp) return '';

  try {
    const date = new Date(timestamp);

    if (full) {
      return date.toLocaleString('lt-LT');
    }

    return date.toLocaleTimeString('lt-LT');
  } catch (e) {
    return timestamp;
  }
};

// Get CSS class for log level badge
// Turning boring log levels into colorful fashion statements
const getLevelClass = (level) => {
  switch (level?.toLowerCase()) {
    case 'debug': return 'badge-primary';
    case 'info': return 'badge-info';
    case 'warn': return 'badge-warning';
    case 'error': return 'badge-error';
    default: return 'badge-ghost';
  }
};

// Get CSS class for HTTP method badge
// HTTP methods get to wear their own special uniforms
const getMethodClass = (method) => {
  switch (method?.toUpperCase()) {
    case 'GET': return 'badge-info';
    case 'POST': return 'badge-success';
    case 'PUT': return 'badge-warning';
    case 'DELETE': return 'badge-error';
    default: return 'badge-ghost';
  }
};

// Get CSS class for HTTP status badge
// Status codes dressed up for their own red carpet event
const getStatusClass = (status) => {
  if (!status) return 'badge-ghost';

  if (status < 300) return 'badge-success';
  if (status < 400) return 'badge-info';
  if (status < 500) return 'badge-warning';
  return 'badge-error';
};

// Save settings to store
// Writing our preferences into digital stone
const saveSettings = () => {
  // Update settings in store
  loggingStore.updateSettings(settings.value);

  // Update max logs count
  loggingStore.maxLogsCount = maxLogsCount.value;

  // Save to localStorage
  try {
    localStorage.setItem('yopta_log_max_count', maxLogsCount.value.toString());
  } catch (e) {
    console.error('Error saving log count', e);
  }
};

// Download logs as JSON file
// Taking our logs on a journey from RAM to hard drive
const downloadLogs = () => {
  // Prepare data for download
  const data = JSON.stringify(loggingStore.recentLogs, null, 2);
  const blob = new Blob([data], { type: 'application/json' });
  const url = URL.createObjectURL(blob);

  // Create download link
  const a = document.createElement('a');
  a.href = url;
  a.download = `client-logs-${new Date().toISOString().substring(0, 10)}.json`;
  document.body.appendChild(a);
  a.click();

  // Clean up resources
  setTimeout(() => {
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, 100);
};

// Clear all stored data
// The digital equivalent of "throw it all out and start over"
const clearAllData = () => {
  // Clear logs
  loggingStore.clearLogs();

  // Clear localStorage
  try {
    localStorage.removeItem('yopta_log_settings');
    localStorage.removeItem('yopta_pending_logs');
    localStorage.removeItem('yopta_log_max_count');
  } catch (e) {
    console.error('Error clearing localStorage', e);
  }

  // Reset settings
  settings.value = {
    logNetworkRequests: true,
    logUIEvents: true,
    logPerformance: true,
    logErrors: true,
    logLevel: 'info'
  };

  // Save default settings
  loggingStore.updateSettings(settings.value);
};

// Create a test log entry
// Like sending a test rocket into space, but for debugging
const createTestLog = () => {
  loggingStore.info('Tai yra testavimo žurnalas', {
    testKey: 'testValue',
    timestamp: new Date().toISOString()
  });
};

// Component initialization
// Setting up our debug mission control on component mount
onMounted(() => {
  // Load state from localStorage
  try {
    const visibleState = localStorage.getItem('debug_panel_visible');
    const expandedState = localStorage.getItem('debug_panel_expanded');

    if (visibleState !== null) {
      isVisible.value = visibleState === '1';
    }

    if (expandedState !== null) {
      isExpanded.value = expandedState === '1';
    }

    const savedMaxLogs = localStorage.getItem('yopta_log_max_count');
    if (savedMaxLogs !== null) {
      maxLogsCount.value = parseInt(savedMaxLogs, 10);
    }
  } catch (e) {
    console.error('Error loading debug panel state', e);
  }

  // Initialize settings from store
  settings.value = { ...loggingStore.settings };

  // Create test log if no logs exist
  if (loggingStore.recentLogs.length === 0) {
    loggingStore.info('Derinimo skydelis inicializuotas', {
      component: 'DebugPanel',
      timestamp: new Date().toISOString()
    });
  }
});
</script>

<style>
.debug-panel {
  position: fixed;
  bottom: 0;
  left: 0;
  width: 100%;
  max-width: 100%;
  max-height: 90vh;
  z-index: 1000;
  transition: all 0.3s ease;
  border-top-left-radius: 0.5rem;
  overflow: hidden;

  /* Initial collapsed size */
  height: 2.5rem;
}

.debug-panel.expanded {
  height: 90vh;
}

.debug-panel-content {
  display: flex;
  flex-direction: column;
  height: calc(100% - 2.5rem);
}

.tab-contents {
  flex: 1;
  overflow-y: auto;
  max-height: calc(100% - 3rem);
}

.log-table {
  height: 100%;
  overflow-y: auto;
}

.log-details {
  max-height: 60%;
  overflow-y: auto;
}

/* For larger screens */
@media (min-width: 768px) {
  .debug-panel {
    width: 640px;
  }
}

/* For very large screens */
@media (min-width: 1280px) {
  .debug-panel {
    width: 800px;
  }
}
</style>
