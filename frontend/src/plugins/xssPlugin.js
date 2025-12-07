// frontend/src/plugins/xssPlugin.js

import { sanitizeData, sanitizeText, sanitizeUrl } from '../utils/xssSanitizer';

export default {
  install(app) {
    // Добавляем глобальные методы
    app.config.globalProperties.$sanitize = sanitizeText;
    app.config.globalProperties.$sanitizeUrl = sanitizeUrl;
    app.config.globalProperties.$sanitizeData = sanitizeData;
    
    // Перехватываем все данные, получаемые из API
    app.mixin({
      methods: {
        // Безопасное извлечение данных из API
        safeApiData(data) {
          return sanitizeData(data);
        }
      }
    });
    
    // Регистрируем обработчик ошибок для отлавливания XSS-атак
    app.config.errorHandler = (err, vm, info) => {
      console.error('Vue Error:', err);
      
      // Проверяем на возможные XSS-атаки
      if (err && err.message && (
        err.message.includes('<script>') || 
        err.message.includes('javascript:') ||
        err.message.includes('onerror=') ||
        err.message.includes('onload=')
      )) {
        console.error('Возможная XSS-атака обнаружена:', err.message);
        // Здесь можно добавить логирование инцидента на сервер
      }
    };
  }
};
