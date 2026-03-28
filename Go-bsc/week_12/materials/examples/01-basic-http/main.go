package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// helloHandler - простейший HTTP обработчик.
// Принимает два параметра:
//   - w http.ResponseWriter — интерфейс для записи ответа клиенту
//   - r *http.Request — структура с информацией о входящем запросе
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Привет, мир! Время на сервере: %s", time.Now().Format("15:04:05"))
}

// methodsHandler демонстрирует обработку разных HTTP методов.
// В реальных REST API каждый метод имеет своё семантическое значение:
//   - GET — получение данных (безопасный, идемпотентный)
//   - POST — создание нового ресурса
//   - PUT — полное обновление существующего ресурса (идемпотентный)
//   - PATCH — частичное обновление ресурса
//   - DELETE — удаление ресурса (идемпотентный)
func methodsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET запрос — возвращаем информацию
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		
		fmt.Fprintln(w, "GET запрос: Получение данных")
		fmt.Fprintf(w, "Путь запроса: %s\n", r.URL.Path)
		fmt.Fprintf(w, "Полный URL: %s\n", r.URL.String())

	case http.MethodPost:
		// POST запрос — создание нового ресурса
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated) // 201 Created
		fmt.Fprintln(w, "POST запрос: Создание ресурса")

	case http.MethodPut:
		// PUT запрос — полное обновление ресурса
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "PUT запрос: Полное обновление ресурса")

	case http.MethodPatch:
		// PATCH запрос — частичное обновление ресурса
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "PATCH запрос: Частичное обновление ресурса")

	case http.MethodDelete:
		// DELETE запрос — удаление ресурса
		w.WriteHeader(http.StatusNoContent) // 204 No Content — успешное удаление без тела ответа

	case http.MethodOptions:
		// OPTIONS запрос — используется для CORS preflight запросов
		w.Header().Set("Allow", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.WriteHeader(http.StatusNoContent)

	default:
		// Неподдерживаемый метод
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func headersHandler(w http.ResponseWriter, r *http.Request) {
	// --- Чтение заголовков запроса ---
	// r.Header — это map[string][]string с заголовками запроса

	// Получить конкретный заголовок (регистронезависимо)
	userAgent := r.Header.Get("User-Agent")
	acceptLanguage := r.Header.Get("Accept-Language")
	contentType := r.Header.Get("Content-Type")
	authorization := r.Header.Get("Authorization")

	// --- Установка заголовков ответа ---
	// w.Header() возвращает заголовки ответа для модификации

	// Устанавливаем Content-Type для JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Добавляем кастомные заголовки
	w.Header().Set("X-Request-ID", fmt.Sprintf("req-%d", time.Now().UnixNano()))
	w.Header().Set("X-Server-Version", "1.0.0")

	// Заголовки для кэширования
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Формируем ответ с информацией о заголовках
	response := map[string]interface{}{
		"message": "Информация о заголовках",
		"request_headers": map[string]string{
			"User-Agent":      userAgent,
			"Accept-Language": acceptLanguage,
			"Content-Type":    contentType,
			"Authorization":   authorization,
		},
		"all_headers_count": len(r.Header),
	}

	// Кодируем ответ в JSON
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	// Парсим query параметры из URL
	// r.URL.Query() возвращает url.Values (map[string][]string)
	queryParams := r.URL.Query()

	query := queryParams.Get("query")
	page := queryParams.Get("page")
	limit := queryParams.Get("limit")
	sort := queryParams.Get("sort")

	// Для параметров с несколькими значениями используем []string
	// Пример: /search?tag=go&tag=web&tag=api
	tags := queryParams["tag"]

	_, hasDebug := queryParams["debug"]

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := map[string]interface{}{
		"message": "Разбор Query параметров",
		"parsed": map[string]interface{}{
			"query": query,
			"page":  page,
			"limit": limit,
			"sort":  sort,
			"tags":  tags,
		},
		"debug_mode":         hasDebug,
		"raw_query":          r.URL.RawQuery,
		"total_params_count": len(queryParams),
	}

	json.NewEncoder(w).Encode(response)
}

func pathHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := map[string]interface{}{
		"message": "Информация о пути URL",
		"path_info": map[string]interface{}{
			"full_path":   path,
			"raw_query":   r.URL.RawQuery,
			"host":        r.Host,
			"remote_addr": r.RemoteAddr,
			"request_uri": r.RequestURI,
		},
	}

	json.NewEncoder(w).Encode(response)
}

type CustomHandler struct {
	Name    string
	Counter int
}

// ServeHTTP реализует интерфейс http.Handler.
// Это позволяет использовать CustomHandler как обработчик HTTP.
func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Counter++

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := map[string]interface{}{
		"handler_name":  h.Name,
		"request_count": h.Counter,
		"message":       fmt.Sprintf("Обработчик '%s' вызван %d раз", h.Name, h.Counter),
	}

	json.NewEncoder(w).Encode(response)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	statusParam := r.URL.Query().Get("status")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch statusParam {
	case "created":
		// 201 Created — ресурс успешно создан
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "201 Created",
			"message": "Ресурс успешно создан",
		})

	case "bad-request":
		// 400 Bad Request — неверный запрос
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "400 Bad Request",
			"message": "Неверный формат запроса",
		})
	case "not-found":
		// 404 Not Found — ресурс не найден
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "404 Not Found",
			"message": "Ресурс не найден",
		})

	case "internal-error":
		// 500 Internal Server Error — внутренняя ошибка сервера
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "500 Internal Server Error",
			"message": "Внутренняя ошибка сервера",
		})

	default:
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "200 OK",
			"message": "Используйте ?status=created|bad-request|not-found|internal-error",
		})
	}
}

func requestInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Собираем все заголовки в map
	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[name] = values[0]
	}

	response := map[string]interface{}{
		"method":         r.Method,
		"url":            r.URL.String(),
		"path":           r.URL.Path,
		"raw_query":      r.URL.RawQuery,
		"protocol":       r.Proto,
		"host":           r.Host,
		"remote_addr":    r.RemoteAddr,
		"content_length": r.ContentLength,
		"headers":        headers,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)
}

// slowHandler демонстрирует долгий запрос для тестирования Graceful Shutdown.
// Запрос выполняется указанное количество секунд (по умолчанию 10).
// Использование: GET /slow?seconds=15
//
// Этот обработчик важен для понимания Graceful Shutdown:
// 1. Запустите сервер
// 2. Отправьте запрос: curl http://localhost:8080/slow?seconds=30
// 3. Пока запрос выполняется, нажмите Ctrl+C
// 4. Сервер дождётся завершения запроса (или таймаута) перед остановкой
func slowHandler(w http.ResponseWriter, r *http.Request) {
	secondsStr := r.URL.Query().Get("seconds")
	seconds := 10
	if secondsStr != "" {
		if s, err := time.ParseDuration(secondsStr + "s"); err == nil {
			seconds = int(s.Seconds())
		}
	}

	log.Printf("[SLOW] Начало долгого запроса на %d секунд...", seconds)
	startTime := time.Now()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	for range seconds {
		select {
		case <-r.Context().Done():
			// Клиент отключился или сервер завершается
			log.Printf("[SLOW] Request cancelled")
			return
		case <-time.After(1 * time.Second):
			// Продолжаем работу
			log.Printf("[SLOW] Working...")
		}
	}

	duration := time.Since(startTime)
	log.Printf("[SLOW] Запрос завершён за %v", duration)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Долгий запрос успешно завершён!",
		"duration": duration.String(),
		"seconds":  seconds,
	})
}

func main() {
	mux := http.NewServeMux()

	// Регистрируем обработчики
	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/methods", methodsHandler)
	mux.HandleFunc("/headers", headersHandler)
	mux.HandleFunc("/query", queryHandler)
	mux.HandleFunc("/path/", pathHandler)
	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/info", requestInfoHandler)
	mux.HandleFunc("/slow", slowHandler) // Долгий запрос для демонстрации Graceful Shutdown

	customHandler := &CustomHandler{Name: "CustomCounter"}
	mux.Handle("/custom", customHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second, // Увеличено для долгих запросов
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown — это паттерн корректного завершения сервера:
	// 1. Сервер перестаёт принимать НОВЫЕ соединения
	// 2. Ждёт завершения ТЕКУЩИХ запросов (до таймаута)
	// 3. Только после этого завершает работу
	//
	// Это важно для:
	// - Не потерять данные при обработке запросов
	// - Корректно закрыть соединения с БД
	// - Завершить фоновые задачи
	// - Деплой без потери запросов (rolling update)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Блокируемся и ждём сигнал завершения
	<-quit

	// Даём 30 секунд на завершение текущих запросов
	shutdownTimeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()


	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful Shutdown failed: %v", err)
	}
}
