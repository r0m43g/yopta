// services/api.js
import axios from 'axios'
import { useAuthStore } from '../stores/auth'
import { useConnectionStore } from '../stores/connection'
import { useSlogStore } from '../stores/slog'
import { useLoggingStore } from '../stores/logging'
import logger from './logger'

// Flag to track the token refresh process
let isRefreshingToken = false;
// Queue for requests waiting for token refresh
let refreshQueue = [];

/**
 * The Digital Bouncer - determines which URLs are VIPs in the logging club!
 * Checks if a URL should bypass the logging system to avoid recursive logging loops.
 * Like a nightclub bouncer who recognizes other security staff and lets them bypass the line,
 * this function prevents logging endpoints from creating infinite loops of log entries.
 *
 * @param {string} url - The URL trying to get into the exclusive logging club
 * @returns {boolean} - True if the URL has a VIP pass to skip logging
 */
function shouldSkipLogging(url) {
  if (!url) return false;

  const loggingEndpoints = [
    '/client-logs',
    '/refresh-token',
    '/auth-ping',
    '/test'
  ];

  return loggingEndpoints.some(endpoint => url.includes(endpoint));
}

/**
 * The Digital Vault - a sophisticated caching system for API responses!
 * A simple yet effective caching mechanism to prevent unnecessary network requests.
 * Think of it as a squirrel collecting nuts - store data when it's plentiful so
 * you don't have to forage when resources are scarce (or the server is down).
 */
const apiCache = {
  cache: new Map(),

  /**
   * The Digital Disinfectant - cleanses data of potential XSS nasties!
   * Sanitizes data to prevent cross-site scripting attacks.
   * Like a fastidious chef who thoroughly washes all ingredients before cooking,
   * this function ensures our data is clean and safe to consume.
   *
   * @param {*} data - The potentially contaminated data
   * @returns {*} - The same data, but now squeaky clean
   */
  sanitize(data) {
    // For simple strings
    if (typeof data === 'string') {
      // Create temporary div for safe escaping
      const tempDiv = document.createElement('div');
      tempDiv.textContent = data;
      return tempDiv.innerHTML;
    }

    // For objects and arrays - recursively process each field
    if (typeof data === 'object' && data !== null) {
      if (Array.isArray(data)) {
        return data.map(item => this.sanitize(item));
      }

      // For regular objects
      const sanitized = {};
      for (const [key, value] of Object.entries(data)) {
        sanitized[key] = this.sanitize(value);
      }
      return sanitized;
    }

    // For primitive types (number, boolean, null, undefined)
    return data;
  },

  /**
   * The Treasure Hunter - retrieves buried data from the cache!
   * Gets data from the cache if it exists and hasn't expired.
   * Like an archaeologist carefully retrieving an ancient artifact,
   * this function extracts previously stored responses from the cache tomb.
   *
   * @param {string} key - The treasure map to the cached data
   * @returns {*|null} - The treasure, if found, or null if the chest is empty
   */
  get(key) {
    const item = this.cache.get(key);
    if (!item) return null;

    // Check if the cache has expired
    if (item.expiry && item.expiry < Date.now()) {
      this.cache.delete(key);
      return null;
    }

    return item.data;
  },

  /**
   * The Archivist - carefully preserves data for future reference!
   * Stores data in cache with sanitization and expiration times.
   * Like a digital librarian meticulously cataloging and preserving knowledge,
   * this function ensures data is safely stored for quick retrieval later.
   *
   * @param {string} key - The catalog reference for this data
   * @param {*} data - The knowledge to preserve
   * @param {number} ttl - How long until this knowledge becomes obsolete (in ms)
   */
  set(key, data, ttl = 5 * 60 * 1000) {
    const expiry = ttl ? Date.now() + ttl : null;

    // Deep clone and sanitize before caching
    const safeData = this.sanitize(JSON.parse(JSON.stringify(data)));

    this.cache.set(key, { data: safeData, expiry });
  },

  /**
   * The Selective Eraser - removes specific entries from the archive!
   * Deletes a specific cache entry by key.
   * Like a targeted laser removing graffiti from a wall without damaging the paint,
   * this function precisely removes one item from the cache while leaving others intact.
   *
   * @param {string} key - The identifier of the entry to be wiped from existence
   */
  delete(key) {
    this.cache.delete(key);
  },

  /**
   * The Digital Apocalypse - wipes out all cached knowledge!
   * Completely empties the cache of all entries.
   * Like pressing a big red button that says "FACTORY RESET",
   * this function brings your cache back to its pristine, empty state.
   */
  clear() {
    this.cache.clear();
  },

  /**
   * The Unique Identifier Creator - generates distinctive keys for cache entries!
   * Creates a unique key based on request configuration to identify cache entries.
   * Like a fingerprint expert who can identify people from tiny ridge patterns,
   * this function creates a unique identifier for each distinct API request.
   *
   * @param {Object} config - The Axios request configuration
   * @returns {string} - A unique key that represents this specific request
   */
  generateKey(config) {
    const { url, method, params, data } = config;
    return `${method}:${url}:${JSON.stringify(params || {})}:${JSON.stringify(data || {})}`;
  }
};

// Create an axios instance with base URL
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 10000,
  withCredentials: true,
});

/**
 * The Queue Processor - handles the waiting line of requests during token refresh!
 * Processes queued requests once a token refresh completes (success or failure).
 * Like a nightclub doorman managing a line of people during a temporary closure,
 * this function deals with the backlog of requests when the token refresh finishes.
 *
 * @param {Error|null} error - Any error from the token refresh, or null if successful
 * @param {string|null} newToken - The shiny new token if refresh was successful
 */
const processQueue = (error, newToken) => {
  refreshQueue.forEach(promise => {
    if (error) {
      promise.reject(error);
    } else {
      promise.resolve(newToken);
    }
  });

  // Clear the queue after processing
  refreshQueue = [];
};

// Request interceptor for adding token to headers
api.interceptors.request.use(
  config => {
    const authStore = useAuthStore();
    if (authStore.accessToken) {
      config.headers['Authorization'] = `Bearer ${authStore.accessToken}`;
    }
    if (authStore.csrfToken) {
      config.headers['X-CSRF-Token'] = `${authStore.csrfToken}`;
    }

    // Skip logging for requests to logging endpoints
    if (shouldSkipLogging(config.url)) {
      config.skipLogging = true;
    }

    // Add metadata for request timing measurement
    config.metadata = { startTime: Date.now() };

    // Check if this request can be cached
    // Cache only GET requests
    if (config.method === 'get' && config.cache !== false) {
      const cacheKey = apiCache.generateKey(config);
      const cachedResponse = apiCache.get(cacheKey);

      if (cachedResponse) {
        // If data found in cache, create a mock response object
        const response = {
          data: cachedResponse,
          status: 200,
          statusText: 'OK',
          headers: { 'X-Cache': 'HIT' },
          config,
          cached: true
        };

        // Log cache hit if logging not explicitly disabled
        if (!config.skipLogging) {
          try {
            logger.info(`Podėlis HIT: ${config.method} ${config.url}`, {
              type: 'cache',
              cacheKey,
              url: config.url,
              method: config.method
            });
          } catch (e) {
            // Ignore logging errors
          }
        }

        // Cancel the real request
        return Promise.reject({
          config,
          response,
          isAxiosCache: true
        });
      }
    }

    return config;
  },
  error => Promise.reject(error)
);

// Response interceptor for handling responses and errors
api.interceptors.response.use(
  response => {
    const authStore = useAuthStore();
    const csrfToken = response.headers['X-CSRF-Token'];
    if (csrfToken) {
      authStore.setCsrfToken(csrfToken);
    }
    try {
      const connectionStore = useConnectionStore();
      connectionStore.handleRequestSuccess();
    } catch (e) {
      // Игнорируем ошибки в обработчике соединения
    }
    // If this is a GET request and caching is not explicitly disabled
    if (response.config.method === 'get' && response.config.cache !== false) {
      const cacheKey = apiCache.generateKey(response.config);
      const ttl = response.config.cacheTTL || 5 * 60 * 1000; // 5 minutes by default
      apiCache.set(cacheKey, response.data, ttl);
    }

    // Log successful request, if not a request to logging endpoints
    if (!response.config.skipLogging && !shouldSkipLogging(response.config.url)) {
      const startTime = response.config.metadata ? response.config.metadata.startTime : undefined;
      const duration = startTime ? (Date.now() - startTime) : undefined;

      try {
        const loggingStore = useLoggingStore();
        loggingStore.networkRequest(
          {
            url: response.config.url,
            method: response.config.method,
            headers: response.config.headers,
            data: response.config.data
          },
          {
            status: response.status,
            statusText: response.statusText,
            headers: response.headers,
            data: response.data
          },
          duration
        );
      } catch (e) {
        // Ignore logging errors to avoid breaking the main flow
        console.error('Užklausos registravimo klaida:', e);
      }
    }

    return response;
  },
  async error => {
    // If this is a "cache error", return the cached response
    if (error.isAxiosCache) {
      return Promise.resolve(error.response);
    }
    try {
      const connectionStore = useConnectionStore();
      connectionStore.handleRequestFailure(error);
    } catch (e) {
      // Игнорируем ошибки в обработчике соединения
    }

    const originalRequest = error.config;
    const authStore = useAuthStore();
    const slogStore = useSlogStore();

    // Log request error, if not related to logging endpoints
    if (originalRequest && !originalRequest.skipLogging && !shouldSkipLogging(originalRequest.url)) {
      const startTime = originalRequest.metadata ? originalRequest.metadata.startTime : undefined;
      const duration = startTime ? (Date.now() - startTime) : undefined;

      try {
        // Get the logging store
        const loggingStore = useLoggingStore();

        // Log network request error
        loggingStore.networkRequest(
          {
            url: originalRequest.url,
            method: originalRequest.method,
            headers: originalRequest.headers,
            data: originalRequest.data
          },
          {
            status: error.response ? error.response.status : 0,
            statusText: error.response ? error.response.statusText : error.message,
            headers: error.response ? error.response.headers : {},
            data: error.response ? error.response.data : error.message
          },
          duration
        );

        // Additionally log as an error for the logging system,
        // but only if not a logging-related request
        if (!shouldSkipLogging(originalRequest.url)) {
          logger.error(error, {
            type: 'network',
            request: {
              url: originalRequest.url,
              method: originalRequest.method
            },
            duration
          });
        }
      } catch (e) {
        // Ignore logging errors
        console.error('Užklausos klaidos registravimo klaida:', e);
      }
    }

    // Handle rate limiting errors (HTTP 429 Too Many Requests)
    if (error.response && error.response.status === 429) {
      // Get retry time from header if available
      const retryAfter = error.response.headers['retry-after'] || '60';
      const retrySeconds = parseInt(retryAfter, 10);

      // Show user-friendly message with retry information
      slogStore.addToast({
        message: `Per daug užklausų. Bandykite dar kartą po ${retrySeconds} sekundžių.`,
        type: 'alert-error',
      });
      return Promise.reject(error);
    }

    // Special handling for "Email not verified" error
    if (error.response &&
        error.response.status === 403 &&
        error.response.data === "Email not verified") {
      // Show message to user
      slogStore.addToast({
        message: 'Prašome patvirtinti savo el. paštą prieš prisijungimą',
        type: 'alert-warning',
      });

      // Don't try to refresh token in this case
      return Promise.reject(error);
    }

    // Handle authentication errors - only if 401, request isn't marked non-retriable and there's a refreshToken
    if (error.response &&
        error.response.status === 401 &&
        !originalRequest._retry &&
        authStore.refreshToken &&
        originalRequest.url !== '/refresh-token') {

      // Limit retry attempts
      originalRequest._retryCount = (originalRequest._retryCount || 0) + 1;

      // Stop after maximum retry attempts (e.g., 3)
      if (originalRequest._retryCount > 3) {
        authStore.clearToken();
        slogStore.addToast({
          message: 'Sesija baigėsi. Prašome prisijungti iš naujo.',
          type: 'alert-error'
        });
        return Promise.reject(error);
      }

      // If we're already in the process of refreshing the token, add request to queue
      if (isRefreshingToken) {
        return new Promise((resolve, reject) => {
          refreshQueue.push({
            resolve: (token) => {
              originalRequest.headers['Authorization'] = `Bearer ${token}`;
              resolve(api(originalRequest));
            },
            reject: (err) => {
              reject(err);
            }
          });
        });
      }

      // Mark request as a retry and set token refreshing flag
      originalRequest._retry = true;
      isRefreshingToken = true;

      try {
        const response = await api.post('/refresh-token',
          { token: authStore.refreshToken },
          {
            headers: { 'X-Refresh-Token': authStore.refreshToken },
            cache: false, // Disable caching for token refresh request
            skipLogging: true // Skip logging to avoid circular dependency
          });

        const newAccessToken = response.data.access_token;
        const newRefreshToken = response.data.refresh_token;

        // Update tokens in store
        authStore.setAccessToken(newAccessToken);
        authStore.setRefreshToken(newRefreshToken);

        // Update header in original request
        originalRequest.headers['Authorization'] = `Bearer ${newAccessToken}`;

        // Log successful token refresh
        logger.info('Prieigos raktas sėkmingai atnaujintas', { type: 'auth' });

        // Process queue of requests waiting for token refresh
        processQueue(null, newAccessToken);

        // Reset token refreshing flag
        isRefreshingToken = false;

        // Retry original request with new token
        return api(originalRequest);
      } catch (refreshError) {
        // If token refresh failed, clear auth data
        authStore.clearToken();

        // Clear cache on logout
        apiCache.clear();

        // Log token refresh error
        logger.error(refreshError, {
          type: 'auth',
          action: 'refresh_token_failed'
        });

        // Process queue with error
        processQueue(refreshError, null);

        // Reset token refreshing flag
        isRefreshingToken = false;

        // Show message to user
        slogStore.addToast({
          message: 'Jūsų sesija baigėsi. Prašome prisijungti iš naujo.',
          type: 'alert-warning',
        });

        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

// Extend API with additional methods for cache management
api.clearCache = apiCache.clear.bind(apiCache);
api.deleteCache = apiCache.delete.bind(apiCache);

// Extend get method to support caching options
const originalGet = api.get;
api.get = function(url, config = {}) {
  return originalGet(url, config);
};

export default api;
