package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"crypto/rand"
	"encoding/hex"
)

// Middleware — тип для middleware функции.
// Принимает http.Handler и возвращает новый http.Handler.
type Middleware func(http.Handler) http.Handler

// responseWriter оборачивает http.ResponseWriter для перехвата статуса и размера.
// Необходим для логирования информации об ответе.
type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	bytesWritten  int
	headerWritten bool
}

// WriteHeader перехватывает код статуса.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.headerWritten {
		rw.statusCode = code
		rw.headerWritten = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write перехватывает запись тела ответа.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.headerWritten {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// LoggingMiddleware логирует информацию о каждом запросе.
// Записывает: метод, путь, статус, время выполнения, размер ответа.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter для получения информации об ответе
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Получаем Request ID из контекста (если есть)
		requestID := r.Context().Value(contextKeyRequestID)
		if requestID == nil {
			requestID = "-"
		}

		// Выполняем следующий обработчик
		next.ServeHTTP(wrapped, r)

		// Вычисляем время выполнения
		duration := time.Since(start)

		// Логируем запрос
		log.Printf("[%s] %s %s %d %v %d bytes",
			requestID,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			wrapped.bytesWritten,
		)
	})
}

// getStatusColor возвращает ANSI цвет для статуса.
func getStatusColor(status int) string {
	switch {
	case status >= 500:
		return "\033[31m" // Red
	case status >= 400:
		return "\033[33m" // Yellow
	case status >= 300:
		return "\033[36m" // Cyan
	case status >= 200:
		return "\033[32m" // Green
	default:
		return "\033[0m" // Default
	}
}

// RecoveryMiddleware перехватывает паники и возвращает 500 ошибку.
// Критически важен для production — предотвращает падение сервера.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Получаем stack trace
				stack := debug.Stack()

				// Логируем панику с полным стеком
				log.Printf("\033[31m[PANIC RECOVERED]\033[0m %v\n%s", err, stack)

				// Отправляем ошибку клиенту
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Internal Server Error",
					"message": "Unexpected error occurred",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// contextKey — тип для ключей контекста (избегаем коллизий).
type contextKey string

const (
	contextKeyRequestID contextKey = "requestID"
	contextKeyUserID    contextKey = "userID"
)

// RequestIDMiddleware добавляет уникальный ID к каждому запросу.
// ID используется для отслеживания запросов в логах и между сервисами.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, есть ли уже Request ID в заголовке
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Генерируем новый ID
			requestID = generateRequestID()
		}

		// Добавляем ID в контекст запроса
		ctx := context.WithValue(r.Context(), contextKeyRequestID, requestID)

		// Добавляем ID в заголовок ответа
		w.Header().Set("X-Request-ID", requestID)

		// Передаём запрос с обновлённым контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID генерирует уникальный идентификатор запроса.
func generateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback к timestamp если rand не работает
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// CORSConfig содержит настройки CORS.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig возвращает конфигурацию CORS по умолчанию.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 часа
	}
}

// CORSMiddleware добавляет заголовки CORS.
// Необходим для работы API с браузерными приложениями на других доменах.
func CORSMiddleware(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Проверяем, разрешён ли origin
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				// Устанавливаем CORS заголовки
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}

				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

				if len(config.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
				}

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
				}
			}

			// Обрабатываем preflight запросы
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware проверяет наличие и валидность токена аутентификации.
// В реальном приложении здесь была бы проверка JWT токена.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			unauthorizedResponse(w, "Authorization header required")
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			unauthorizedResponse(w, "Invalid authorization format. Use: Bearer <token>")
			return
		}

		token := parts[1]

		// Валидируем токен (упрощённая проверка)
		userID, valid := validateToken(token)
		if !valid {
			unauthorizedResponse(w, "Invalid or expired token")
			return
		}

		// Добавляем информацию о пользователе в контекст
		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken проверяет токен и возвращает ID пользователя.
// В реальном приложении здесь была бы проверка JWT.
func validateToken(token string) (string, bool) {
	// Упрощённая проверка для демонстрации
	validTokens := map[string]string{
		"admin-token-123": "admin",
		"user-token-456":  "user123",
		"test-token-789":  "test",
	}

	userID, exists := validTokens[token]
	return userID, exists
}

// unauthorizedResponse отправляет 401 ошибку.
func unauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "Unauthorized",
		"message": message,
	})
}

// RateLimiter ограничивает количество запросов.
// Использует алгоритм Token Bucket.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   map[string]int       // Количество токенов для каждого клиента
	lastTime map[string]time.Time // Время последнего запроса
	rate     int                  // Максимум токенов
	perSec   int                  // Токенов в секунду
}

// NewRateLimiter создаёт новый rate limiter.
func NewRateLimiter(rate int, perSec int) *RateLimiter {
	return &RateLimiter{
		tokens:   make(map[string]int),
		lastTime: make(map[string]time.Time),
		rate:     rate,
		perSec:   perSec,
	}
}

// Allow проверяет, разрешён ли запрос для данного клиента.
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Инициализируем клиента если это первый запрос
	if _, exists := rl.tokens[clientID]; !exists {
		rl.tokens[clientID] = rl.rate
		rl.lastTime[clientID] = now
	}

	// Пополняем токены на основе прошедшего времени
	elapsed := now.Sub(rl.lastTime[clientID])
	refill := int(elapsed.Seconds()) * rl.perSec
	if refill > 0 {
		rl.tokens[clientID] += refill
		if rl.tokens[clientID] > rl.rate {
			rl.tokens[clientID] = rl.rate
		}
		rl.lastTime[clientID] = now
	}

	// Проверяем наличие токенов
	if rl.tokens[clientID] > 0 {
		rl.tokens[clientID]--
		return true
	}

	return false
}

// RateLimitMiddleware ограничивает количество запросов от клиента.
func RateLimitMiddleware(limiter *RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Используем IP адрес как идентификатор клиента
			clientID := r.RemoteAddr

			if !limiter.Allow(clientID) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "Too Many Requests",
					"message": "Rate limit exceeded. Please try again later.",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware ограничивает время выполнения запроса.
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Создаём контекст с таймаутом
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Канал для сигнализации о завершении обработки
			done := make(chan struct{})

			// Оборачиваем ResponseWriter для предотвращения записи после таймаута
			tw := &timeoutWriter{
				ResponseWriter: w,
				done:           done,
			}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Запрос обработан успешно
				return
			case <-ctx.Done():
				// Таймаут
				tw.mu.Lock()
				defer tw.mu.Unlock()

				if !tw.written {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusGatewayTimeout)
					json.NewEncoder(w).Encode(map[string]string{
						"error":   "Gateway Timeout",
						"message": "Request processing took too long",
					})
				}
			}
		})
	}
}

// timeoutWriter предотвращает запись после таймаута.
type timeoutWriter struct {
	http.ResponseWriter
	mu      sync.Mutex
	written bool
	done    chan struct{}
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	select {
	case <-tw.done:
		return 0, context.DeadlineExceeded
	default:
		tw.written = true
		return tw.ResponseWriter.Write(b)
	}
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	select {
	case <-tw.done:
		return
	default:
		tw.written = true
		tw.ResponseWriter.WriteHeader(code)
	}
}

// ContentTypeMiddleware проверяет и устанавливает Content-Type.
func ContentTypeMiddleware(contentType string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Для POST/PUT/PATCH проверяем Content-Type запроса
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				ct := r.Header.Get("Content-Type")
				if ct != "" && !strings.HasPrefix(ct, contentType) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnsupportedMediaType)
					json.NewEncoder(w).Encode(map[string]string{
						"error":   "Unsupported Media Type",
						"message": fmt.Sprintf("Content-Type must be %s", contentType),
					})
					return
				}
			}

			// Устанавливаем Content-Type для ответа
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

// Chain применяет middleware в порядке слева направо.
// Первый middleware в списке будет внешним (выполнится первым).
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Применяем middleware в обратном порядке
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// helloHandler — простой обработчик.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello, World!",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// protectedHandler — защищённый обработчик (требует аутентификации).
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID := r.Context().Value(contextKeyUserID).(string)
	requestID := r.Context().Value(contextKeyRequestID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Welcome to protected resource!",
		"user_id":    userID,
		"request_id": requestID,
	})
}

// slowHandler — медленный обработчик для демонстрации таймаута.
func slowHandler(w http.ResponseWriter, r *http.Request) {
	// Имитируем долгую операцию
	select {
	case <-time.After(10 * time.Second):
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Slow operation completed",
		})
	case <-r.Context().Done():
		// Контекст отменён (таймаут)
		return
	}
}

// panicHandler — обработчик, который вызывает панику.
func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("Something went terribly wrong!")
}

func main() {
	// Создаём rate limiter: 10 запросов, пополнение 2/сек
	rateLimiter := NewRateLimiter(10, 2)

	corsConfig := DefaultCORSConfig()

	mux := http.NewServeMux()

	baseMiddlewares := []Middleware{
		RecoveryMiddleware,                        // 1. Перехват паники (самый внешний)
		RequestIDMiddleware,                       // 2. Добавление Request ID
		LoggingMiddleware,                         // 3. Логирование
		CORSMiddleware(corsConfig),                // 4. CORS
		RateLimitMiddleware(rateLimiter),          // 5. Rate limiting
		ContentTypeMiddleware("application/json"), // 6. Content-Type
	}

	// Базовый обработчик с middleware
	mux.Handle("/", Chain(http.HandlerFunc(helloHandler), baseMiddlewares...))

	// --- Защищённый эндпоинт ---
	protectedMiddlewares := append(baseMiddlewares, AuthMiddleware)
	mux.Handle("/protected", Chain(http.HandlerFunc(protectedHandler), protectedMiddlewares...))

	// --- Медленный эндпоинт с таймаутом ---
	timeoutMiddlewares := append(baseMiddlewares, TimeoutMiddleware(3*time.Second))
	mux.Handle("/slow", Chain(http.HandlerFunc(slowHandler), timeoutMiddlewares...))

	// --- Эндпоинт с паникой ---
	mux.Handle("/panic", Chain(http.HandlerFunc(panicHandler), baseMiddlewares...))

	// --- Health check (минимум middleware) ---
	mux.Handle("/health", Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
		}),
		RequestIDMiddleware,
		LoggingMiddleware,
	))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
