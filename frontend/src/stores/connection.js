// stores/connection.js
import { defineStore } from 'pinia'
import api from '../services/api'

export const useConnectionStore = defineStore('connection', {
  state: () => ({
    isOnline: navigator.onLine,
    serverStatus: 'unknown', // 'online', 'offline', 'unknown'
    lastCheck: null,
    checkInterval: null,
    successiveResponses: 0, // Новый счетчик успешных ответов
    isActivelyPolling: false // Флаг активного опроса
  }),

  actions: {
    initConnectionListeners() {
      window.addEventListener('online', this.handleOnline)
      window.addEventListener('offline', this.handleOffline)

      this.checkServerStatus()
    },

    handleOnline() {
      this.isOnline = true
      this.checkServerStatus()
    },

    handleOffline() {
      this.isOnline = false
      this.serverStatus = 'offline'
      this.stopServerCheck() // Отключаем опрос если нет интернета
    },

    handleRequestFailure(error) {
      const is502 = error.response && error.response.status === 502;
      const isNetworkError = !error.response || error.code === 'ECONNABORTED';

      if (is502 || isNetworkError) {
        this.serverStatus = 'offline';
        this.successiveResponses = 0;

        if (!this.isActivelyPolling) {
          this.startServerCheck();
          this.isActivelyPolling = true;
        }
      }
    },

    handleRequestSuccess() {
      const previousStatus = this.serverStatus;

      if (previousStatus === 'offline') {
        this.successiveResponses++;
        this.serverStatus = 'online';

        if (this.successiveResponses >= 3) {
          this.stopServerCheck();
          this.isActivelyPolling = false;
        }
      } else {
        this.serverStatus = 'online';
      }
    },

    startServerCheck() {
      this.stopServerCheck();

      this.checkInterval = setInterval(() => {
        this.checkServerStatus();
      }, 30000); // 30 секунд
    },

    stopServerCheck() {
      if (this.checkInterval) {
        clearInterval(this.checkInterval);
        this.checkInterval = null;
        this.isActivelyPolling = false;
      }
    },

    async checkServerStatus() {
      if (!this.isOnline) {
        this.serverStatus = 'offline';
        return;
      }

      try {
        await api.get('/test', {
          timeout: 15000,
          cache: false,
          _retry: false,
          skipLogging: true // Чтобы избежать зацикливания логов
        });

        this.serverStatus = 'online';

        this.successiveResponses++;

        if (this.successiveResponses >= 3 && this.isActivelyPolling) {
          this.stopServerCheck();
          this.isActivelyPolling = false;
        }
      } catch (error) {
        this.serverStatus = 'offline';
        this.successiveResponses = 0; // Сбрасываем счетчик
      } finally {
        this.lastCheck = new Date();
      }
    },

    cleanup() {
      window.removeEventListener('online', this.handleOnline);
      window.removeEventListener('offline', this.handleOffline);
      this.stopServerCheck();
    }
  },

  getters: {
    isConnected() {
      return this.isOnline && this.serverStatus === 'online';
    },

    statusText() {
      if (!this.isOnline) return 'Nėra interneto ryšio';
      if (this.serverStatus === 'online') return 'Prisijungta';
      if (this.serverStatus === 'offline') return 'Serveris nepasiekiamas';
      return 'Tikrinamas ryšys...';
    }
  }
})
