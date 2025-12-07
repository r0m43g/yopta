// stores/logging.js

import { defineStore } from 'pinia';
import logger from '../services/logger';

// Helper function to identify logging-related URLs to prevent circular logging
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

export const useLoggingStore = defineStore('logging', {
  state: () => ({
    // Our trusty log memory - the digital scroll that records recent events
    recentLogs: [],
    // The memory limit - even digital brains can only remember so much
    maxLogsCount: 100,
    // Feature toggles - like choosing which superpowers to activate
    settings: {
      logNetworkRequests: true,
      logUIEvents: true,
      logPerformance: false,
      logErrors: true,
      logLevel: 'debug' // 'debug', 'info', 'warn', 'error'
    },
    // Throttling controls - because nobody likes a chatty application
    lastLogTime: 0,
    minLogInterval: 200, // Minimum time between logs in ms
    // Rate limiting - even the most exciting events get boring when repeated
    logTypeCounts: {},
    logTypeLimit: 5, // Maximum number of similar logs within 5 seconds
    logTypeResetInterval: 5000, // Reset counters interval
    // Initialization flag - ensures we don't send "Hello World" twice
    initialized: false
  }),

  getters: {
    /**
     * The Disaster Historian - collects tales of digital misfortune!
     * Returns an array of error logs for analysis or simply wallowing in your application's misery.
     * Like a collection of dramatic tragedy plays, but starring your code as the protagonist.
     *
     * @returns {Array} - A catalog of calamities that befell your application
     */
    errors() {
      return this.recentLogs.filter(log => log.level === 'error');
    },

    /**
     * The Noise Filter - decides which messages deserve attention!
     * Determines if a log of the specified level should be recorded based on current settings.
     * Like a bouncer at an exclusive club who only lets in VIPs (Very Important Problems),
     * this function maintains the quality of your log collection.
     *
     * @returns {Function} - A function that evaluates if a log level meets the threshold
     */
    shouldLog() {
      return (level) => {
        const levels = {
          debug: 0,
          info: 1,
          warn: 2,
          error: 3
        };

        // Get numerical values for current level and level being checked
        const currentLevelValue = levels[this.settings.logLevel] || 1;
        const checkLevelValue = levels[level] || 1;

        // Only log if level is higher or equal to current level
        return checkLevelValue >= currentLevelValue;
      };
    },

    /**
     * The Anti-Spam Police - prevents your logs from becoming a flood!
     * Checks if we should throttle logging based on recent activity.
     * Like a dam controlling a river of information, this ensures your logs
     * flow at a reasonable rate instead of causing a digital flash flood.
     *
     * @returns {Function} - A function that returns true if logs should be throttled
     */
    shouldThrottle() {
      return () => {
        const now = Date.now();
        if (now - this.lastLogTime < this.minLogInterval) {
          return true;
        }
        return false;
      };
    }
  },

  actions: {
    /**
     * The Grand Initializer - kickstarts the logging machine!
     * Sets up the logging system and creates initial logs to mark the beginning of time.
     * Like the Big Bang for your application's universe of events, this function
     * creates the first records in the great cosmic log book.
     */
    initialize() {
      if (this.initialized) return;

      this.initialized = true;
      this.loadSettings();

      // Create initialization log
      this.info('Registravimo saugykla inicijuota', {
        timestamp: new Date().toISOString(),
        isDevelopment: process.env.NODE_ENV === 'development'
      });

      // In development mode, create test logs for the debug panel
      if (process.env.NODE_ENV === 'development') {
        this.debug('Derinimo režimas įjungtas', {
          component: 'LoggingStore',
          timestamp: new Date().toISOString()
        });
      }
    },

    /**
     * The Repetition Detective - spots when logs are getting too samey!
     * Checks if we've seen too many similar logs recently and should hold back.
     * Like a teacher who says "I've heard that excuse five times today already!",
     * this function prevents your logs from becoming an echo chamber.
     *
     * @param {String} type - The category of log we're checking
     * @returns {Boolean} - True if we've heard enough of this particular type of log
     */
    isLogTypeLimited(type) {
      if (!type) return false;

      const key = `${type}`;
      const now = Date.now();

      // Clean old counters
      Object.keys(this.logTypeCounts).forEach(k => {
        if (now - this.logTypeCounts[k].timestamp > this.logTypeResetInterval) {
          delete this.logTypeCounts[k];
        }
      });

      // Check and update counter for this type
      if (!this.logTypeCounts[key]) {
        this.logTypeCounts[key] = {
          count: 1,
          timestamp: now
        };
        return false;
      }

      // Reset counter if too old
      if (now - this.logTypeCounts[key].timestamp > this.logTypeResetInterval) {
        this.logTypeCounts[key] = {
          count: 1,
          timestamp: now
        };
        return false;
      }

      // Increment counter and check limit
      this.logTypeCounts[key].count++;
      return this.logTypeCounts[key].count > this.logTypeLimit;
    },

    /**
     * The Universal Scribe - records events in the great book of application history!
     * Adds a log to the store and sends it through the logging service.
     * Like a medieval monk painstakingly copying manuscripts, this function
     * ensures that your application's story is properly documented for posterity.
     *
     * @param {String} level - The severity level (debug, info, warn, error)
     * @param {String} message - The tale that needs telling
     * @param {Object} data - Additional context for future archaeologists
     */
    log(level, message, data = {}) {
      // If message is undefined, use empty string
      const safeMessage = message || 'Žinutės nėra';

      // Check if this level should be logged
      if (!this.shouldLog(level)) {
        return;
      }

      // Check throttling
      if (this.shouldThrottle()) {
        return;
      }

      // Update last log time
      this.lastLogTime = Date.now();

      // Skip network requests to logging endpoints
      if (data.type === 'network' && data.request && shouldSkipLogging(data.request.url)) {
        return;
      }

      // Get log type for rate limiting
      let logType = level;
      if (data.type) {
        logType = `${level}_${data.type}`;
        if (data.type === 'network' && data.request) {
          logType = `${logType}_${data.request.method}_${data.request.url.split('?')[0]}`;
        }
      }

      // Check if this type of log is being spammed
      if (this.isLogTypeLimited(logType)) {
        return;
      }

      // Create log entry with timestamp
      const timestamp = new Date().toISOString();
      const logEntry = {
        level,
        message: safeMessage,
        data,
        timestamp,
        logType
      };

      // Add to recent logs array
      this.recentLogs.unshift(logEntry);

      // Limit array size
      if (this.recentLogs.length > this.maxLogsCount) {
        this.recentLogs = this.recentLogs.slice(0, this.maxLogsCount);
      }

      // Send log through logging service
      this.sendLog(level, safeMessage, data);
    },

    /**
     * The Messenger - dispatches your log to its final destination!
     * Sends the log through the dedicated logging service for processing.
     * Like a postal worker ensuring your letter reaches its recipient,
     * this function makes sure your log gets to where it needs to go.
     *
     * @param {String} level - The importance level of the message
     * @param {String} message - The content needing delivery
     * @param {Object} data - The package of extra details
     */
    sendLog(level, message, data) {
      try {
        switch (level) {
          case 'debug':
            logger.debug(message, data);
            break;
          case 'info':
            logger.info(message, data);
            break;
          case 'warn':
            logger.warn(message, data);
            break;
          case 'error':
            logger.error(message, data);
            break;
          default:
            logger.info(message, data);
        }
      } catch (error) {
        console.error('Klaida siunčiant žurnalą:', error);
      }
    },

    /**
     * The News Reporter - broadcasts informational updates!
     * Logs a message at the 'info' level for general announcements.
     * Like a news anchor calmly reporting the day's events, this function
     * documents the normal happenings without causing any alarm.
     *
     * @param {String} message - The informational bulletin
     * @param {Object} data - Supporting facts and figures
     */
    info(message, data = {}) {
      this.log('info', message, data);
    },

    /**
     * The Curious Detective - investigates the minutiae of application behavior!
     * Logs detailed debug information for development and troubleshooting.
     * Like Sherlock Holmes examining a crime scene with a magnifying glass,
     * this function captures the subtle clues that help solve mysteries.
     *
     * @param {String} message - The curious observation
     * @param {Object} data - The evidence collected at the scene
     */
    debug(message, data = {}) {
      this.log('debug', message, data);
    },

    /**
     * The Worried Parent - expresses concern without full panic!
     * Logs warnings about potentially problematic situations.
     * Like a parent saying "I'm not angry, just disappointed," this function
     * documents situations that aren't failures yet but are heading that way.
     *
     * @param {String} message - The gentle admonishment
     * @param {Object} data - Details about what's concerning you
     */
    warn(message, data = {}) {
      this.log('warn', message, data);
    },

    /**
     * The Disaster Announcer - proclaims when things have gone terribly wrong!
     * Logs errors that represent actual failures in the application.
     * Like a town crier announcing that the castle is on fire, this function
     * makes it very clear when something requires immediate attention.
     *
     * @param {Error|String} error - The calamity that has befallen us
     * @param {Object} data - Details about the catastrophe
     */
    error(error, data = {}) {
      // Check if error logging is enabled
      if (!this.settings.logErrors) {
        return;
      }

      let errorMessage = error;
      let errorData = { ...data };

      if (error instanceof Error) {
        errorMessage = error.message;
        errorData = {
          ...errorData,
          stack: error.stack,
          name: error.name
        };

        // Skip logging errors from logging endpoints
        if (error.config && shouldSkipLogging(error.config.url)) {
          return;
        }
      }

      this.log('error', errorMessage, errorData);
    },

    /**
     * The User Behavior Anthropologist - studies how humans interact with the UI!
     * Logs events that occur in the user interface for analysis.
     * Like a scientist observing animals in the wild, this function
     * documents how users navigate and interact with your application.
     *
     * @param {String} component - The habitat where the interaction occurred
     * @param {String} action - The behavior being observed
     * @param {Object} data - Field notes about the interaction
     */
    uiEvent(component, action, data = {}) {
      // Check if UI event logging is enabled
      if (!this.settings.logUIEvents) {
        return;
      }

      this.log('info', `Vartotojo sąsajos įvykis: ${action}`, {
        component,
        action,
        eventType: 'ui',
        ...data
      });
    },

    /**
     * The Efficiency Expert - measures how well things are performing!
     * Logs performance metrics to track application speed and resource usage.
     * Like a sports coach with a stopwatch timing athletes, this function
     * keeps track of how quickly your application is completing tasks.
     *
     * @param {String} name - The performance metric being measured
     * @param {Number} value - The measurement result
     * @param {Object} data - Additional context about the measurement
     */
    performance(name, value, data = {}) {
      // Check if performance logging is enabled
      if (!this.settings.logPerformance) {
        return;
      }

      this.log('info', `Našumas: ${name}`, {
        metricName: name,
        metricValue: value,
        metricType: 'performance',
        ...data
      });
    },

    /**
     * The Network Traffic Reporter - monitors the digital highways!
     * Logs details about HTTP requests and responses for monitoring.
     * Like a traffic helicopter reporting on road conditions, this function
     * keeps track of data flowing in and out of your application.
     *
     * @param {Object} request - The outbound journey details
     * @param {Object} response - The return trip information
     * @param {Number} duration - How long the round trip took
     */
    networkRequest(request, response, duration) {
      // Check if network request logging is enabled
      if (!this.settings.logNetworkRequests) {
        return;
      }

      // Skip logging requests to logging endpoints
      if (shouldSkipLogging(request.url)) {
        return;
      }

      const level = (response.status >= 400) ? 'error' : 'info';
      const message = `Tinklas: ${request.method} ${request.url} (${response.status})`;

      const logData = {
        type: 'network',
        request: {
          url: request.url,
          method: request.method,
          headers: request.headers
        },
        response: {
          status: response.status,
          statusText: response.statusText
        },
        duration
      };

      this.log(level, message, logData);
    },

    /**
     * The Preference Updater - adjusts your logging experience!
     * Updates logging settings based on user preferences.
     * Like a personal assistant who remembers how you like your coffee,
     * this function ensures logs are collected according to your specifications.
     *
     * @param {Object} newSettings - Your updated logging preferences
     */
    updateSettings(newSettings) {
      this.settings = {
        ...this.settings,
        ...newSettings
      };

      // Save settings to localStorage for persistence
      try {
        localStorage.setItem('yopta_log_settings', JSON.stringify(this.settings));
      } catch (e) {
        console.error('Klaida išsaugant žurnalo nustatymus:', e);
      }
    },

    /**
     * The Memory Restorer - recalls your previous preferences!
     * Loads logging settings from localStorage to maintain consistency.
     * Like waking up with all your memories intact, this function
     * ensures your application remembers how you like your logs served.
     */
    loadSettings() {
      try {
        const savedSettings = localStorage.getItem('yopta_log_settings');
        if (savedSettings) {
          this.settings = {
            ...this.settings,
            ...JSON.parse(savedSettings)
          };
        }

        // Determine if we're in development or production
        const isDevelopment = process.env.NODE_ENV === 'development';

        // In development mode, disable some logging by default
        if (isDevelopment && !savedSettings) {
          this.settings.logNetworkRequests = true;
          this.settings.logUIEvents = true;
          this.settings.logLevel = 'debug'; // Show all logs in development mode
        }
      } catch (e) {
        console.error('Klaida įkeliant žurnalo nustatymus:', e);
      }

      // Also load max logs count from localStorage
      try {
        const savedMaxLogs = localStorage.getItem('yopta_log_max_count');
        if (savedMaxLogs) {
          this.maxLogsCount = parseInt(savedMaxLogs, 10);

          // Check for valid values
          if (isNaN(this.maxLogsCount) || this.maxLogsCount < 10) {
            this.maxLogsCount = 100;
          }
          if (this.maxLogsCount > 1000) {
            this.maxLogsCount = 1000;
          }
        }
      } catch (e) {
        console.error('Klaida įkeliant maksimalų žurnalų skaičiaus nustatymą:', e);
      }
    },

    /**
     * The Great Eraser - wipes the slate clean!
     * Clears all stored logs from memory.
     * Like hitting the reset button or using a digital eraser,
     * this function gives you a fresh start when log history is no longer needed.
     */
    clearLogs() {
      this.recentLogs = [];
      // Also reset log type counters
      this.logTypeCounts = {};
    }
  }
});
