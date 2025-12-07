// backend/internal/middleware/logging.go
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"yopta-template/internal/models"
)

// BufferedResponseWriter - структура-обертка для http.ResponseWriter
// с возможностью сохранения ответа для логирования
type BufferedResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Buffer     *bytes.Buffer
}

// shouldSkipLogging проверяет, нужно ли пропустить логирование для данного пути
func shouldSkipLogging(path string) bool {
	// Пути, которые не должны логироваться
	excludedPaths := []string{
		"/api/v1/client-logs",
		"/api/v1/test",
	}

	for _, excludedPath := range excludedPaths {
		if strings.HasPrefix(path, excludedPath) {
			return true
		}
	}

	return false
}

// NewBufferedResponseWriter создает новый BufferedResponseWriter
func NewBufferedResponseWriter(w http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Buffer:         bytes.NewBuffer(nil),
	}
}

// Write перехватывает ответ и сохраняет его в буфер
func (bw *BufferedResponseWriter) Write(b []byte) (int, error) {
	bw.Buffer.Write(b)
	return bw.ResponseWriter.Write(b)
}

// WriteHeader перехватывает код статуса HTTP
func (bw *BufferedResponseWriter) WriteHeader(statusCode int) {
	bw.StatusCode = statusCode
	bw.ResponseWriter.WriteHeader(statusCode)
}

// LoggingMiddleware создает middleware для логирования запросов и ответов HTTP
func LoggingMiddleware(loggerType string) func(http.Handler) http.Handler {
	var logFile *os.File
	var err error

	// Настройка логгера
	if loggerType == "file" {
		// Создаем директорию для логов, если она не существует
		err = os.MkdirAll("logs", 0o755)
		if err != nil {
			log.Printf("Ошибка создания директории для логов: %v", err)
		}

		// Создаем файл лога с текущей датой
		currentTime := time.Now().Format("2006-01-02")
		logFile, err = os.OpenFile(
			fmt.Sprintf("logs/app-%s.log", currentTime),
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0o644,
		)
		if err != nil {
			log.Printf("Ошибка открытия файла лога: %v", err)
		} else {
			log.SetOutput(logFile)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			// Импортируем функцию shouldSkipLogging из package middleware
			if shouldSkipLogging(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			// Создаем копию тела запроса для логирования
			var requestBody []byte
			if r.Body != nil && r.Method != "GET" {
				requestBody, _ = io.ReadAll(r.Body)
				// Восстанавливаем тело запроса для дальнейшей обработки
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// Получаем IP-адрес клиента
			clientIP := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				clientIP = strings.Split(forwardedFor, ",")[0]
			}

			// Создаем буферизованный ResponseWriter для перехвата ответа
			bufferedWriter := NewBufferedResponseWriter(w)

			// Создаем объект лога с базовой информацией
			logEntry := models.LogEntry{
				Timestamp:  time.Now().Format(time.RFC3339),
				Method:     r.Method,
				Path:       r.URL.Path,
				Query:      r.URL.RawQuery,
				ClientIP:   clientIP,
				UserAgent:  r.UserAgent(),
				RequestID:  r.Header.Get("X-Request-ID"),
				StatusCode: bufferedWriter.StatusCode,
			}

			// Маскируем конфиденциальные данные в теле запроса
			if len(requestBody) > 0 {
				// Определяем пути, которые содержат конфиденциальные данные
				sensitiveRoutes := []string{
					"/api/v1/login",
					"/api/v1/register",
					"/api/v1/change-password",
				}

				// Проверяем, является ли текущий путь чувствительным
				isSensitive := false
				for _, route := range sensitiveRoutes {
					if strings.Contains(r.URL.Path, route) {
						isSensitive = true
						break
					}
				}

				if isSensitive {
					// Для чувствительных данных маскируем пароль
					var requestMap map[string]interface{}
					if err := json.Unmarshal(requestBody, &requestMap); err == nil {
						if password, exists := requestMap["password"]; exists && password != nil {
							requestMap["password"] = "[MASKED]"
						}
						if oldPassword, exists := requestMap["old_password"]; exists &&
							oldPassword != nil {
							requestMap["old_password"] = "[MASKED]"
						}
						if newPassword, exists := requestMap["new_password"]; exists &&
							newPassword != nil {
							requestMap["new_password"] = "[MASKED]"
						}
						maskedBody, _ := json.Marshal(requestMap)
						logEntry.RequestBody = string(maskedBody)
					} else {
						logEntry.RequestBody = "[ERROR PARSING REQUEST]"
					}
				} else {
					// Для нечувствительных данных логируем как есть
					logEntry.RequestBody = string(requestBody)
				}
			}

			// Выполняем запрос через следующий обработчик
			next.ServeHTTP(bufferedWriter, r)

			// Завершаем логирование после обработки запроса
			duration := time.Since(startTime)
			logEntry.Duration = duration.Milliseconds()
			logEntry.StatusCode = bufferedWriter.StatusCode

			// ВАЖНОЕ ИЗМЕНЕНИЕ: Получаем информацию о пользователе из контекста
			// ПОСЛЕ выполнения обработчика, когда JWT middleware мог добавить эту информацию
			if userID := r.Context().Value("user_id"); userID != nil {
				// Убеждаемся, что userID - это число
				if uid, ok := userID.(int); ok {
					logEntry.UserID = uid // Сохраняем как число, а не интерфейс
				} else {
					// Если это не int, преобразуем в строку для логирования
					logEntry.UserID = fmt.Sprintf("%v", userID)
				}
			}

			if role := r.Context().Value("role"); role != nil {
				logEntry.UserRole = role
			}

			// Логируем ответ, если это не файл
			contentType := bufferedWriter.Header().Get("Content-Type")
			if !strings.Contains(contentType, "image") &&
				!strings.Contains(contentType, "font") &&
				!strings.Contains(contentType, "video") {
				// Ограничиваем размер логируемого ответа
				responseBody := bufferedWriter.Buffer.String()
				if len(responseBody) > 1024 {
					responseBody = responseBody[:1024] + "... [truncated]"
				}
				logEntry.ResponseBody = responseBody
			}

			// Логирование в БД (реализация ниже)
			go func(entry models.LogEntry) {
				// Получаем соединение из пула
				db := models.GetDBConnection()
				if db == nil {
					log.Printf("Ошибка: невозможно получить соединение с БД для логирования")
					return
				}

				// Сохраняем лог в БД
				err := models.SaveLogEntry(db, entry)
				if err != nil {
					log.Printf("Ошибка сохранения лога: %v", err)
				}
			}(logEntry)

			// Также выводим в консоль или файл
			logJSON, _ := json.Marshal(logEntry)
			log.Printf("%s", logJSON)
		})
	}
}

// ContextWithRequestID добавляет уникальный идентификатор запроса в контекст
func ContextWithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			r.Header.Set("X-Request-ID", requestID)
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

