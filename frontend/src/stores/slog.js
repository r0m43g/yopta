// stores/slog.js
import { defineStore } from 'pinia'
import { useLoggingStore } from './logging'

export const useSlogStore = defineStore('slog', {
  state: () => ({
    toasts: [],
    maxToasts: 10, // Maximum number of toasts displayed simultaneously
    toastDuration: 6000, // Duration in ms (0 means don't auto-hide)
  }),
  actions: {

    addToast(toast) {
      // Add timestamp for tracking and animation
      const toastWithTimestamp = {
        ...toast,
        timestamp: new Date().toISOString()
      }

      // Add to the start of the array
      this.toasts.unshift(toastWithTimestamp)

      // Limit the number of displayed notifications
      if (this.toasts.length > this.maxToasts) {
        this.toasts.pop()
      }

      // Auto-close notification if configured
      const autoClose = toast.autoClose !== false
      const duration = toast.duration || this.toastDuration

      if (autoClose && duration > 0) {
        setTimeout(() => {
          // Find toast index by timestamp
          const index = this.toasts.findIndex(t => t.timestamp === toastWithTimestamp.timestamp)
          if (index !== -1) {
            this.removeToast(index)
          }
        }, duration)
      }
    },

    clearToasts() {
      this.toasts = []
    },

    removeToast(index) {
      if (index >= 0 && index < this.toasts.length) {
        this.toasts.splice(index, 1)
      }
    }
  },
})
