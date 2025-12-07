// services/logger.js
/**
 * The Application's Memory Keeper - recording the digital diary of your app's life!
 *
 * This service manages client-side logging, batching log messages, and sending them
 * to the server when appropriate. Think of it as the ship's log on the Starship Enterprise,
 * meticulously documenting all the adventures (and misadventures) of your application.
 */

import api from './api';
import { useAuthStore } from '../stores/auth';
import { useConnectionStore } from '../stores/connection';

/**
 * The Logger Class - the historian of your application's journey!
 *
 * Collects, organizes, and transmits logs while ensuring efficiency
 * and avoiding recursive logging doomsday scenarios.
 */
class Logger {
  constructor() {
    this.logQueue = [];
    this.maxQueueSize = 100; // Maximum logs in queue before flushing
    this.isProcessing = false;
    this.pendingLogs = this.loadPendingLogs();
    this.levels = {
      DEBUG: 'debug',
      INFO: 'info',
      WARN: 'warn',
      ERROR: 'error'
    };

    // Flags to prevent logging loops
    this.loggingEndpoints = ['/client-logs', '/refresh-token', '/auth-ping', '/test'];

    // Batch processing and throttling controls
    this.batchTimeout = null;
    this.batchInterval = 5000; // 5 seconds between batches in production
    this.lastProcessTime = 0;
    this.minProcessInterval = 3000; // Minimum 3 seconds between processing

    // Environment-specific settings
    this.isDevelopment = process.env.NODE_ENV === 'development';

    // In development mode: longer intervals, fewer logs
    if (this.isDevelopment) {
      this.batchInterval = 10000; // 10 seconds
      this.minProcessInterval = 5000; // 5 seconds
      this.maxQueueSize = 50; // Store fewer logs
    }

    // Recover log processing on initial initialization
    // But don't actively start in development mode
    if (!this.isDevelopment) {
      // Allow time for network connection to initialize
      setTimeout(() => this.processPendingLogs(), 3000);
    }
  }

  /**
   * The Spy Detector - identifies URLs connected to the logging system!
   *
   * Prevents recursive logging loops by detecting logging-related endpoints.
   * Like a security agent spotting another undercover agent at a party,
   * this function recognizes URLs that are part of the logging system itself.
   *
   * @param {String} url - The URL to investigate for logging connections
   * @returns {Boolean} - true if the URL is working for the logging agency
   */
  shouldSkipLogging(url) {
    if (!url) return false;

    // Check if URL contains any logging endpoints
    return this.loggingEndpoints.some(endpoint => url.includes(endpoint));
  }

  /**
   * The Amnesia Recovery Specialist - retrieves logs lost in the void!
   *
   * Loads logs that couldn't be sent previously due to connection issues.
   * Like a psychologist retrieving repressed memories, this function
   * brings back logs that were stored during offline episodes.
   *
   * @returns {Array} - Array of unsent logs
   */
  loadPendingLogs() {
    try {
      const savedLogs = localStorage.getItem('yopta_pending_logs');
      return savedLogs ? JSON.parse(savedLogs) : [];
    } catch (e) {
      console.error('Klaida įkeliant atidėtus žurnalus:', e);
      return [];
    }
  }

  /**
   * The Digital Preserver - saves unsent logs for future transmission!
   *
   * Stores logs in localStorage when they can't be sent immediately.
   * Like a time capsule buried for future generations, this function
   * safeguards logs until a connection is restored and they can be sent.
   */
  savePendingLogs() {
    try {
      if (this.pendingLogs.length > 0) {
        localStorage.setItem('yopta_pending_logs', JSON.stringify(this.pendingLogs));
      } else {
        localStorage.removeItem('yopta_pending_logs');
      }
    } catch (e) {
      console.error('Klaida išsaugant atidėtus žurnalus:', e);
    }
  }

  /**
   * The Queue Manager - adds logs to the waiting line!
   *
   * Places a log in the queue and schedules processing of the queue.
   * Like an efficient postal clerk organizing mail for delivery,
   * this function ensures logs are properly queued and scheduled for processing.
   *
   * @param {Object} logData - The log information to be queued
   */
  enqueueLog(logData) {
    // Check if log is related to logging requests
    if (logData.type === 'network' && logData.request &&
        this.shouldSkipLogging(logData.request.url)) {
      // Skip logging of logging requests
      return;
    }

    // Add timestamp
    logData.timestamp = new Date().toISOString();

    // Add user ID if authenticated
    const authStore = useAuthStore();
    if (authStore.isAuthenticated) {
      logData.userId = authStore.getID;
      logData.userRole = authStore.role;
    }

    // Add User-Agent and platform info
    logData.userAgent = navigator.userAgent;
    logData.platform = {
      vendor: navigator.vendor,
      language: navigator.language,
      platform: navigator.platform,
      screen: {
        width: window.screen.width,
        height: window.screen.height
      }
    };

    // Limit queue size
    if (this.logQueue.length >= this.maxQueueSize) {
      this.logQueue.shift(); // Remove oldest log
    }

    // Add new log to queue
    this.logQueue.push(logData);

    // Clear existing timeout
    if (this.batchTimeout) {
      clearTimeout(this.batchTimeout);
    }

    // Set new timeout for batch processing
    this.batchTimeout = setTimeout(() => {
      this.processLogQueue();
    }, this.batchInterval);
  }

  /**
   * The Batch Processor - sends logs to the server in efficient groups!
   *
   * Processes queued logs and sends them to the server.
   * Like a factory assembly line that packages products in bulk,
   * this function efficiently sends logs in groups rather than one by one.
   */
  async processLogQueue() {
    // Check if processing is already running or queue is empty
    if (this.isProcessing || this.logQueue.length === 0) {
      return;
    }

    // Check if we need to throttle
    const now = Date.now();
    if (now - this.lastProcessTime < this.minProcessInterval) {
      // Too soon for processing, reschedule
      if (!this.batchTimeout) {
        this.batchTimeout = setTimeout(() => {
          this.processLogQueue();
        }, this.minProcessInterval);
      }
      return;
    }

    // Update last process time and reset timer
    this.lastProcessTime = now;
    this.batchTimeout = null;

    // In development mode, skip server transmission
    if (this.isDevelopment) {
      console.log(`[Žurnalas] Siunčiama ${this.logQueue.length} žurnalų praleista kūrimo režime`);
      this.logQueue = [];
      return;
    }

    // Check connection to server
    const connectionStore = useConnectionStore();
    if (!connectionStore.isConnected) {
      // Save logs for later transmission
      this.pendingLogs = [...this.pendingLogs, ...this.logQueue];
      if (this.pendingLogs.length > this.maxQueueSize * 2) {
        // If too many logs accumulated, remove oldest
        this.pendingLogs = this.pendingLogs.slice(-this.maxQueueSize);
      }
      this.savePendingLogs();
      this.logQueue = [];
      return;
    }

    this.isProcessing = true;

    try {
      // Copy current queue and clear main queue
      const logsToProcess = [...this.logQueue];
      this.logQueue = [];

      // Prevent sending empty log arrays
      if (logsToProcess.length === 0) {
        this.isProcessing = false;
        return;
      }

      // Send logs to server
      await api.post('/client-logs', { logs: logsToProcess }, {
        // Set flag to avoid logging this request (prevent recursion)
        skipLogging: true,
        // Set timeout
        timeout: 10000
      });
    } catch (error) {
      // On error, return logs to queue or save as pending
      if (connectionStore.isConnected) {
        // If connection exists but request failed, return logs to queue
        this.logQueue = [...this.logQueue, ...this.pendingLogs];
      } else {
        // If no connection, add to pending logs
        this.pendingLogs = [...this.pendingLogs, ...this.logQueue];
        this.savePendingLogs();
        this.logQueue = [];
      }
    } finally {
      this.isProcessing = false;

      // Check if new logs appeared during processing
      if (this.logQueue.length > 0) {
        // Start processing again if new logs appeared, but with delay
        setTimeout(() => this.processLogQueue(), this.minProcessInterval);
      }
    }
  }

  /**
   * The Backlog Processor - handles logs that accumulated during offline periods!
   *
   * Processes stored logs when connection is restored.
   * Like a postal worker dealing with mail that accumulated during a snowstorm,
   * this function works through the backlog of logs once connectivity returns.
   */
  async processPendingLogs() {
    if (this.pendingLogs.length === 0) {
      return;
    }

    const connectionStore = useConnectionStore();
    if (!connectionStore.isConnected) {
      // If no connection, exit and wait for restoration
      return;
    }

    // In development mode, skip sending to server
    if (this.isDevelopment) {
      console.log(`[Žurnalas] Apdorojama ${this.pendingLogs.length} atidėtų žurnalų praleista kūrimo režime`);
      this.pendingLogs = [];
      this.savePendingLogs();
      return;
    }

    try {
      // Check if there are logs to send
      if (this.pendingLogs.length === 0) {
        return;
      }

      // Copy pending logs and clear array
      const logsToProcess = [...this.pendingLogs];
      this.pendingLogs = [];

      // Send pending logs to server in groups of no more than 50 at once
      const batchSize = 50;

      for (let i = 0; i < logsToProcess.length; i += batchSize) {
        const batch = logsToProcess.slice(i, i + batchSize);

        try {
          await api.post('/client-logs', { logs: batch }, {
            skipLogging: true,
            timeout: 15000
          });

          // Pause between batches
          if (i + batchSize < logsToProcess.length) {
            await new Promise(resolve => setTimeout(resolve, 1000));
          }
        } catch (batchError) {
          // Return unprocessed logs to pending
          this.pendingLogs = [...this.pendingLogs, ...logsToProcess.slice(i)];
          break;
        }
      }

      // Clear storage if all logs sent successfully
      this.savePendingLogs();
    } catch (error) {
      // On error, return logs to pending
      this.pendingLogs = [...logsToProcess];
      this.savePendingLogs();
    }
  }

  /**
   * The Town Crier - announces informational updates!
   *
   * Logs an informational message for general knowledge.
   * Like the person who walks around town announcing "All is well!",
   * this function records normal, everyday happenings in your application.
   *
   * @param {String} message - The informational announcement
   * @param {Object} data - Additional context about the announcement
   */
  info(message, data = {}) {
    this.log(this.levels.INFO, message, data);
  }

  /**
   * The Curious Explorer - documents discoveries for the scientifically minded!
   *
   * Logs detailed debug information for development and troubleshooting.
   * Like a scientist recording observations in a laboratory notebook,
   * this function captures technical details useful during development.
   *
   * @param {String} message - The scientific observation
   * @param {Object} data - The experimental data
   */
  debug(message, data = {}) {
    this.log(this.levels.DEBUG, message, data);
  }

  /**
   * The Worried Oracle - predicts potential problems!
   *
   * Logs warnings about concerning but non-fatal situations.
   * Like a fortune teller who sees storm clouds gathering in your future,
   * this function alerts you to potential issues before they become critical.
   *
   * @param {String} message - The ominous prediction
   * @param {Object} data - Details about the impending situation
   */
  warn(message, data = {}) {
    this.log(this.levels.WARN, message, data);
  }

  /**
   * The Digital Paramedic - documents emergencies needing attention!
   *
   * Logs errors representing actual application failures.
   * Like an EMT reporting a medical emergency with vital statistics,
   * this function records critical issues that require immediate attention.
   *
   * @param {String|Error} error - The emergency situation or error object
   * @param {Object} data - Additional diagnostic information
   */
  error(error, data = {}) {
    let errorMessage = error;
    let errorStack = null;
    let errorData = { ...data };

    if (error instanceof Error) {
      errorMessage = error.message;
      errorStack = error.stack;

      // Add additional error properties if they exist
      if (error.response) {
        errorData.response = {
          status: error.response.status,
          statusText: error.response.statusText,
          data: error.response.data
        };

        // If error is from logging URLs, skip
        if (this.shouldSkipLogging(error.response.config?.url)) {
          return;
        }
      }

      if (error.request) {
        errorData.request = {
          url: error.request.url,
          method: error.request.method
        };

        // If error is from logging URLs, skip
        if (this.shouldSkipLogging(error.request.url)) {
          return;
        }
      }

      if (error.config) {
        errorData.config = {
          url: error.config.url,
          method: error.config.method,
          headers: error.config.headers
        };

        // If error is from logging URLs, skip
        if (this.shouldSkipLogging(error.config.url)) {
          return;
        }
      }
    }

    this.log(this.levels.ERROR, errorMessage, {
      ...errorData,
      stack: errorStack
    });
  }

  /**
   * The Anthropologist of Digital Behavior - studies user interactions!
   *
   * Logs events occurring in the user interface for analysis.
   * Like a cultural anthropologist observing tribal rituals,
   * this function documents how users interact with your application.
   *
   * @param {String} component - The village where the interaction occurred
   * @param {String} action - The ritual being performed
   * @param {Object} data - Field notes about the cultural practice
   */
  uiEvent(component, action, data = {}) {
    this.log(this.levels.INFO, `Vartotojo sąsajos įvykis: ${action}`, {
      component,
      action,
      ...data,
      type: 'ui'
    });
  }

  /**
   * The Universal Scribe - records all manner of digital events!
   *
   * Core logging function that all other logging methods use.
   * Like the ancient Egyptian scribes who documented everything from
   * grain harvests to royal decrees, this function is the foundation
   * of the entire logging system.
   *
   * @param {String} level - The importance level of the scroll
   * @param {String} message - The hieroglyphic message to record
   * @param {Object} data - Additional pictograms providing context
   */
  log(level, message, data = {}) {
    // If message is empty, use placeholder
    const safeMessage = message || 'Žinutės nėra';

    // Create log entry
    const logEntry = {
      level,
      message: safeMessage,
      data,
      url: window.location.href,
      path: window.location.pathname,
      route: this.getCurrentRouteName()
    };

    // Add to queue for sending
    this.enqueueLog(logEntry);
  }

  /**
   * The Route Cartographer - identifies the current territory!
   *
   * Gets the name of the current route for contextual logging.
   * Like an explorer determining their location on a map,
   * this function identifies which part of the application the log came from.
   *
   * @returns {String} - The name of the current route or null
   */
  getCurrentRouteName() {
    try {
      // Check for Vue Router
      const router = window.router;
      if (router && router.currentRoute && router.currentRoute.value) {
        return router.currentRoute.value.name;
      }
      return null;
    } catch (e) {
      return null;
    }
  }

  /**
   * The Development Notifier - echoes logs to the console!
   *
   * Outputs log information to the console during development.
   * Like a stage whisper in a theater production, this function
   * provides behind-the-scenes information visible only to developers.
   *
   * @param {String} level - The severity of the message
   * @param {String} message - The content to be whispered
   * @param {Object} data - Supplementary details for context
   */
  consoleLog(level, message, data) {
    const isDev = process.env.NODE_ENV === 'development';

    // Output logs to console only in development mode
    if (isDev) {
      switch (level) {
        case this.levels.DEBUG:
          console.debug(`[YOPTA] ${message}`, data);
          break;
        case this.levels.INFO:
          console.info(`[YOPTA] ${message}`, data);
          break;
        case this.levels.WARN:
          console.warn(`[YOPTA] ${message}`, data);
          break;
        case this.levels.ERROR:
          console.error(`[YOPTA] ${message}`, data);
          break;
        default:
          console.log(`[YOPTA] ${message}`, data);
      }
    }
  }
}

// Create a singleton logger instance
const logger = new Logger();

// Set up global handlers for unhandled errors
window.addEventListener('error', (event) => {
  logger.error(event.error || new Error(event.message), {
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    type: 'unhandled'
  });
});

// Handler for unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
  let error = event.reason;
  if (!(error instanceof Error)) {
    error = new Error(String(error));
  }

  logger.error(error, {
    type: 'unhandledrejection',
    promise: event.promise
  });
});

export default logger;
