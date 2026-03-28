// Package main демонстрирует использование Echo — высокопроизводительного веб-фреймворка для Go.
//
// Echo — это минималистичный и высокопроизводительный веб-фреймворк для Go,
// который предоставляет оптимизированную маршрутизацию, middleware, и многое другое.
//
// Официальный сайт: https://echo.labstack.com/
// GitHub: https://github.com/labstack/echo
//
// Ключевые особенности Echo:
// - Оптимизированный роутер (Radix tree)
// - Высокая производительность
// - Минимальное выделение памяти
// - Масштабируемый middleware фреймворк
// - Автоматический TLS через Let's Encrypt
// - HTTP/2 поддержка
// - Группировка маршрутов
// - Биндинг данных и валидация
// - Рендеринг шаблонов
// - Централизованная обработка ошибок
package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Task представляет модель задачи.
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title" validate:"required,min=1,max=200"`
	Description string    `json:"description"`
	Status      string    `json:"status" validate:"oneof=pending in_progress completed"`
	Priority    int       `json:"priority" validate:"min=1,max=5"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTaskRequest — запрос на создание задачи.
type CreateTaskRequest struct {
	Title       string    `json:"title" validate:"required,min=1,max=200"`
	Description string    `json:"description"`
	Priority    int       `json:"priority" validate:"min=1,max=5"`
	DueDate     time.Time `json:"due_date"`
}

// UpdateTaskRequest — запрос на обновление задачи.
type UpdateTaskRequest struct {
	Title       *string    `json:"title" validate:"omitempty,min=1,max=200"`
	Description *string    `json:"description"`
	Status      *string    `json:"status" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    *int       `json:"priority" validate:"omitempty,min=1,max=5"`
	DueDate     *time.Time `json:"due_date"`
}

// TaskStore — хранилище задач.
type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[int]*Task
	nextID int
}

// NewTaskStore создаёт хранилище с тестовыми данными.
func NewTaskStore() *TaskStore {
	s := &TaskStore{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}

	// Тестовые данные
	testTasks := []CreateTaskRequest{
		{Title: "Изучить Echo", Description: "Пройти туториал по Echo фреймворку", Priority: 5},
		{Title: "Написать REST API", Description: "Создать CRUD для задач", Priority: 4},
		{Title: "Добавить тесты", Description: "Покрыть API тестами", Priority: 3},
	}

	for _, t := range testTasks {
		now := time.Now()
		task := &Task{
			ID:          s.nextID,
			Title:       t.Title,
			Description: t.Description,
			Status:      "pending",
			Priority:    t.Priority,
			DueDate:     now.AddDate(0, 0, 7),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		s.tasks[task.ID] = task
		s.nextID++
	}

	return s
}

var store = NewTaskStore()

// CustomHTTPErrorHandler обрабатывает ошибки в едином формате.
// Echo позволяет настроить глобальный обработчик ошибок.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	// Проверяем тип ошибки
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			message = m
		}
	}

	// Не отправляем ответ если он уже отправлен
	if !c.Response().Committed {
		c.JSON(code, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    code,
				"message": message,
			},
		})
	}
}

// ServerHeader добавляет кастомный заголовок сервера.
func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("X-Server", "Echo-Demo/1.0")
		return next(c)
	}
}

// RequestTimer измеряет время выполнения запроса.
func RequestTimer(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		duration := time.Since(start)
		c.Response().Header().Set("X-Response-Time", duration.String())

		return err
	}
}

// ApiKeyAuth проверяет API ключ.
func ApiKeyAuth(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Header.Get("X-API-Key")
			if key != apiKey {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}
			return next(c)
		}
	}
}

// TaskHandler содержит обработчики для задач.
type TaskHandler struct {
	store *TaskStore
}

// NewTaskHandler создаёт новый обработчик.
func NewTaskHandler(store *TaskStore) *TaskHandler {
	return &TaskHandler{store: store}
}

// List возвращает список задач с фильтрацией.
// GET /api/tasks
// GET /api/tasks?status=pending&priority=5
func (h *TaskHandler) List(c echo.Context) error {
	// Получаем query параметры
	status := c.QueryParam("status")
	priorityStr := c.QueryParam("priority")

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, t := range h.store.tasks {
		// Фильтр по статусу
		if status != "" && t.Status != status {
			continue
		}

		// Фильтр по приоритету
		if priorityStr != "" {
			priority, _ := strconv.Atoi(priorityStr)
			if t.Priority != priority {
				continue
			}
		}

		tasks = append(tasks, t)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  tasks,
		"count": len(tasks),
	})
}

// Get возвращает задачу по ID.
// GET /api/tasks/:id
func (h *TaskHandler) Get(c echo.Context) error {
	// c.Param() извлекает параметр из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	h.store.mu.RLock()
	task, exists := h.store.tasks[id]
	h.store.mu.RUnlock()

	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": task,
	})
}

// Create создаёт новую задачу.
// POST /api/tasks
func (h *TaskHandler) Create(c echo.Context) error {
	var req CreateTaskRequest

	// c.Bind() связывает данные из запроса со структурой
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Простая валидация
	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Title is required")
	}

	now := time.Now()
	h.store.mu.Lock()
	task := &Task{
		ID:          h.store.nextID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if task.Priority == 0 {
		task.Priority = 3 // По умолчанию средний приоритет
	}
	if task.DueDate.IsZero() {
		task.DueDate = now.AddDate(0, 0, 7) // По умолчанию через неделю
	}

	h.store.tasks[task.ID] = task
	h.store.nextID++
	h.store.mu.Unlock()

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Task created successfully",
		"data":    task,
	})
}

// Update обновляет задачу.
// PUT /api/tasks/:id
func (h *TaskHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	task, exists := h.store.tasks[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	// Обновляем только переданные поля
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}
	task.UpdatedAt = time.Now()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Task updated successfully",
		"data":    task,
	})
}

// Delete удаляет задачу.
// DELETE /api/tasks/:id
func (h *TaskHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	if _, exists := h.store.tasks[id]; !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	delete(h.store.tasks, id)
	return c.NoContent(http.StatusNoContent)
}

// Complete отмечает задачу выполненной.
// POST /api/tasks/:id/complete
func (h *TaskHandler) Complete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	task, exists := h.store.tasks[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	task.Status = "completed"
	task.UpdatedAt = time.Now()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Task completed",
		"data":    task,
	})
}

// Stats возвращает статистику по задачам.
// GET /api/stats
func (h *TaskHandler) Stats(c echo.Context) error {
	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	stats := map[string]int{
		"total":       len(h.store.tasks),
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
	}

	for _, t := range h.store.tasks {
		stats[t.Status]++
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": stats,
	})
}

func main() {
	// Создаём экземпляр Echo
	e := echo.New()

	// Отключаем баннер и информацию о порте
	e.HideBanner = true

	// Устанавливаем кастомный обработчик ошибок
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	// Recovery middleware — восстановление после паники
	e.Use(middleware.Recover())

	// Logger middleware — логирование запросов
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | ${status} | ${latency_human} | ${remote_ip} | ${method} ${uri}\n",
	}))

	// Request ID middleware
	e.Use(middleware.RequestID())

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-API-Key"},
	}))

	// Наши кастомные middleware
	e.Use(ServerHeader)
	e.Use(RequestTimer)

	handler := NewTaskHandler(store)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	api := e.Group("/api")

	api.GET("/stats", handler.Stats)

	tasks := api.Group("/tasks")
	{
		tasks.GET("", handler.List)
		tasks.GET("/:id", handler.Get)

		protected := tasks.Group("")
		protected.Use(ApiKeyAuth("secret-api-key"))
		{
			protected.POST("", handler.Create)
			protected.PUT("/:id", handler.Update)
			protected.DELETE("/:id", handler.Delete)
			protected.POST("/:id/complete", handler.Complete)
		}
	}

	e.GET("/old-path", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/api/tasks")
	})

	e.GET("/format", func(c echo.Context) error {
		data := map[string]string{
			"message": "Hello, World!",
			"format":  "Based on Accept header",
		}

		accept := c.Request().Header.Get("Accept")
		switch accept {
		case "application/xml":
			return c.XML(http.StatusOK, data)
		default:
			return c.JSON(http.StatusOK, data)
		}
	})

	// e.GET("/download", func(c echo.Context) error {
	//     return c.File("./file.pdf")
	// })

	e.Logger.Fatal(e.Start(":8080"))
}
