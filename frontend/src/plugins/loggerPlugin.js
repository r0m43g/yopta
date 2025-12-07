// frontend/src/plugins/loggerPlugin.js
import logger from '../services/logger';
import { useLoggingStore } from '../stores/logging';

/**
 * Плагин для интеграции логирования в Vue приложение
 * - Перехватывает ошибки компонентов
 * - Добавляет возможность логирования через $log
 * - Интегрируется с Vue Router для логирования навигации
 * - Добавляет директивы для логирования событий UI
 */
export default {
  install(app, options = {}) {
    // Настраиваем обработчик ошибок
    app.config.errorHandler = (err, vm, info) => {
      // Логируем ошибку
      logger.error(err, {
        info,
        component: vm?.$options?.name || 'Unknown',
        props: vm?.$props,
        route: vm?.$route?.path
      });
      
      // Вызываем предыдущий обработчик ошибок, если он был
      if (options.prevErrorHandler) {
        options.prevErrorHandler(err, vm, info);
      } else {
        console.error(err);
      }
    };
    
    // Добавляем глобальные методы логирования
    app.config.globalProperties.$log = {
      info: (message, data) => logger.info(message, data),
      debug: (message, data) => logger.debug(message, data),
      warn: (message, data) => logger.warn(message, data),
      error: (message, data) => logger.error(message, data),
      uiEvent: (component, action, data) => logger.uiEvent(component, action, data)
    };
    
    // Добавляем глобальную директиву для логирования событий UI
    app.directive('log-click', {
      mounted(el, binding) {
        el.addEventListener('click', () => {
          const component = binding.instance?.$options?.name || 'Unknown';
          const value = binding.value || 'clicked';
          
          // Получаем хранилище логов
          const loggingStore = useLoggingStore();
          
          // Логируем событие клика
          loggingStore.uiEvent(component, value, {
            element: el.tagName,
            class: el.className,
            id: el.id
          });
        });
      }
    });
    
    // Интеграция с Vue Router, если он доступен
    if (options.router) {
      // Сохраняем ссылку на роутер в глобальном объекте для доступа из логгера
      window.router = options.router;
      
      // Логируем навигационные события
      options.router.beforeEach((to, from, next) => {

        logger.info(`Navigation: ${from.path} -> ${to.path}`, {
          type: 'navigation',
          from: {
            path: from.path,
            name: from.name,
            params: from.params
          },
          to: {
            path: to.path,
            name: to.name,
            params: to.params
          }
        });
        
        // Продолжаем навигацию
        next();
        
      });
      
      // Логируем ошибки навигации
      options.router.onError((error) => {
        logger.error(error, {
          type: 'routerError'
        });
      });
    }
    
    // Добавляем инструменты для разработки, если в режиме разработки
    if (process.env.NODE_ENV === 'development') {
      window.$logger = logger;
    }
  }
};
