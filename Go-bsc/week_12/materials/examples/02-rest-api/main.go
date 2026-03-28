package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)


// User представляет модель пользователя в системе.
// JSON теги определяют имена полей при сериализации.
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age,omitempty"` // omitempty — не включать если 0
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest — структура для создания пользователя.
// Отдельная структура позволяет контролировать какие поля
// клиент может передать при создании.
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// UpdateUserRequest — структура для обновления пользователя.
// Указатели позволяют различать "не передано" и "передано пустое значение".
type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Age      *int    `json:"age"`
	IsActive *bool   `json:"is_active"`
}

// APIResponse — стандартный формат ответа API.
// Единый формат упрощает работу клиентов с API.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError — структура для описания ошибки.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta — метаданные для пагинации и дополнительной информации.
type Meta struct {
	Total      int `json:"total,omitempty"`
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}


// UserStore — потокобезопасное хранилище пользователей.
// В реальном приложении здесь была бы база данных.
type UserStore struct {
	mu       sync.RWMutex   // Мьютекс для потокобезопасности
	users    map[int]*User  // Хранилище пользователей
	nextID   int            // Счётчик для генерации ID
	emailIdx map[string]int // Индекс email -> ID для проверки уникальности
}

// NewUserStore создаёт новое хранилище с тестовыми данными.
func NewUserStore() *UserStore {
	store := &UserStore{
		users:    make(map[int]*User),
		nextID:   1,
		emailIdx: make(map[string]int),
	}

	// Добавляем тестовые данные
	testUsers := []CreateUserRequest{
		{Name: "Алиса Иванова", Email: "alice@example.com", Age: 28},
		{Name: "Борис Петров", Email: "boris@example.com", Age: 35},
		{Name: "Виктория Сидорова", Email: "victoria@example.com", Age: 22},
		{Name: "Дмитрий Козлов", Email: "dmitry@example.com", Age: 41},
		{Name: "Елена Новикова", Email: "elena@example.com", Age: 30},
	}

	for _, req := range testUsers {
		store.Create(req)
	}

	return store
}

// Create создаёт нового пользователя.
func (s *UserStore) Create(req CreateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем уникальность email
	if _, exists := s.emailIdx[req.Email]; exists {
		return nil, fmt.Errorf("email already exists")
	}

	now := time.Now()
	user := &User{
		ID:        s.nextID,
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.users[user.ID] = user
	s.emailIdx[req.Email] = user.ID
	s.nextID++

	return user, nil
}

// GetByID возвращает пользователя по ID.
func (s *UserStore) GetByID(id int) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	return user, exists
}

// GetAll возвращает всех пользователей с пагинацией.
func (s *UserStore) GetAll(page, perPage int, nameFilter string) ([]*User, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Собираем пользователей в слайс с фильтрацией
	var filtered []*User
	for _, user := range s.users {
		// Применяем фильтр по имени (если указан)
		if nameFilter != "" && !strings.Contains(
			strings.ToLower(user.Name),
			strings.ToLower(nameFilter),
		) {
			continue
		}
		filtered = append(filtered, user)
	}

	total := len(filtered)

	// Применяем пагинацию
	start := (page - 1) * perPage
	if start >= total {
		return []*User{}, total
	}

	end := start + perPage
	if end > total {
		end = total
	}

	return filtered[start:end], total
}

// Update обновляет существующего пользователя.
func (s *UserStore) Update(id int, req UpdateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Обновляем только переданные поля
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		// Проверяем уникальность нового email
		if existingID, exists := s.emailIdx[*req.Email]; exists && existingID != id {
			return nil, fmt.Errorf("email already exists")
		}
		// Удаляем старый индекс и добавляем новый
		delete(s.emailIdx, user.Email)
		user.Email = *req.Email
		s.emailIdx[*req.Email] = id
	}
	if req.Age != nil {
		user.Age = *req.Age
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	return user, nil
}

// Delete удаляет пользователя по ID.
func (s *UserStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return false
	}

	delete(s.emailIdx, user.Email)
	delete(s.users, id)
	return true
}

// UserHandler содержит обработчики для работы с пользователями.
type UserHandler struct {
	store *UserStore
}

// NewUserHandler создаёт новый обработчик.
func NewUserHandler(store *UserStore) *UserHandler {
	return &UserHandler{store: store}
}

// ServeHTTP реализует интерфейс http.Handler.
// Маршрутизирует запросы к соответствующим методам.
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	path := strings.TrimPrefix(r.URL.Path, "/api/users")
	path = strings.TrimPrefix(path, "/")

	// Маршрутизация
	switch {
	case path == "" || path == "/":
		// /api/users — коллекция
		switch r.Method {
		case http.MethodGet:
			h.List(w, r)
		case http.MethodPost:
			h.Create(w, r)
		default:
			h.methodNotAllowed(w, []string{"GET", "POST"})
		}

	default:
		// /api/users/{id} — конкретный ресурс
		id, err := strconv.Atoi(path)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID", "")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.Get(w, r, id)
		case http.MethodPut:
			h.Update(w, r, id)
		case http.MethodPatch:
			h.PartialUpdate(w, r, id)
		case http.MethodDelete:
			h.Delete(w, r, id)
		default:
			h.methodNotAllowed(w, []string{"GET", "PUT", "PATCH", "DELETE"})
		}
	}
}

// List возвращает список пользователей с пагинацией.
// GET /api/users?page=1&per_page=10&name=Алиса
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(query.Get("per_page"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 10
	}

	nameFilter := query.Get("name")

	users, total := h.store.GetAll(page, perPage, nameFilter)

	totalPages := (total + perPage - 1) / perPage

	h.respondSuccess(w, http.StatusOK, users, &Meta{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// Get возвращает пользователя по ID.
// GET /api/users/{id}
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request, id int) {
	user, exists := h.store.GetByID(id)
	if !exists {
		h.respondError(w, http.StatusNotFound, "USER_NOT_FOUND",
			"User not found", fmt.Sprintf("User with ID %d does not exist", id))
		return
	}

	h.respondSuccess(w, http.StatusOK, user, nil)
}

// Create создаёт нового пользователя.
// POST /api/users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	// Декодируем JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_JSON",
			"Invalid JSON in request body", err.Error())
		return
	}

	// Валидация
	if err := h.validateCreateRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Validation failed", err.Error())
		return
	}

	// Создаём пользователя
	user, err := h.store.Create(req)
	if err != nil {
		if err.Error() == "email already exists" {
			h.respondError(w, http.StatusConflict, "EMAIL_EXISTS",
				"Email already registered", "")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to create user", err.Error())
		return
	}

	// Возвращаем созданного пользователя со статусом 201 Created
	h.respondSuccess(w, http.StatusCreated, user, nil)
}

// Update полностью обновляет пользователя (PUT).
// PUT /api/users/{id}
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request, id int) {
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_JSON",
			"Invalid JSON in request body", err.Error())
		return
	}

	// При PUT все поля обязательны
	if err := h.validateCreateRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Validation failed", err.Error())
		return
	}

	// Конвертируем в UpdateUserRequest
	updateReq := UpdateUserRequest{
		Name:  &req.Name,
		Email: &req.Email,
		Age:   &req.Age,
	}

	user, err := h.store.Update(id, updateReq)
	if err != nil {
		if err.Error() == "user not found" {
			h.respondError(w, http.StatusNotFound, "USER_NOT_FOUND",
				"User not found", "")
			return
		}
		if err.Error() == "email already exists" {
			h.respondError(w, http.StatusConflict, "EMAIL_EXISTS",
				"Email already registered", "")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to update user", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, user, nil)
}

// PartialUpdate частично обновляет пользователя (PATCH).
// PATCH /api/users/{id}
func (h *UserHandler) PartialUpdate(w http.ResponseWriter, r *http.Request, id int) {
	var req UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_JSON",
			"Invalid JSON in request body", err.Error())
		return
	}

	// Валидация частичного обновления
	if err := h.validateUpdateRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Validation failed", err.Error())
		return
	}

	user, err := h.store.Update(id, req)
	if err != nil {
		if err.Error() == "user not found" {
			h.respondError(w, http.StatusNotFound, "USER_NOT_FOUND",
				"User not found", "")
			return
		}
		if err.Error() == "email already exists" {
			h.respondError(w, http.StatusConflict, "EMAIL_EXISTS",
				"Email already registered", "")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to update user", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, user, nil)
}

// Delete удаляет пользователя.
// DELETE /api/users/{id}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request, id int) {
	if !h.store.Delete(id) {
		h.respondError(w, http.StatusNotFound, "USER_NOT_FOUND",
			"User not found", "")
		return
	}

	// 204 No Content — успешное удаление без тела ответа
	w.WriteHeader(http.StatusNoContent)
}

// validateCreateRequest валидирует запрос на создание пользователя.
func (h *UserHandler) validateCreateRequest(req *CreateUserRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if len(req.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("email must be valid")
	}
	if req.Age < 0 || req.Age > 150 {
		return fmt.Errorf("age must be between 0 and 150")
	}
	return nil
}

// validateUpdateRequest валидирует запрос на обновление пользователя.
func (h *UserHandler) validateUpdateRequest(req *UpdateUserRequest) error {
	if req.Name != nil {
		if len(*req.Name) < 2 {
			return fmt.Errorf("name must be at least 2 characters")
		}
		if len(*req.Name) > 100 {
			return fmt.Errorf("name must not exceed 100 characters")
		}
	}
	if req.Email != nil && !strings.Contains(*req.Email, "@") {
		return fmt.Errorf("email must be valid")
	}
	if req.Age != nil && (*req.Age < 0 || *req.Age > 150) {
		return fmt.Errorf("age must be between 0 and 150")
	}
	return nil
}

// respondSuccess отправляет успешный ответ в стандартном формате.
func (h *UserHandler) respondSuccess(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// respondError отправляет ответ с ошибкой в стандартном формате.
func (h *UserHandler) respondError(w http.ResponseWriter, status int, code, message, details string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// methodNotAllowed отправляет ответ 405 Method Not Allowed.
func (h *UserHandler) methodNotAllowed(w http.ResponseWriter, allowed []string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	h.respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
		"Method not allowed", fmt.Sprintf("Allowed methods: %s", strings.Join(allowed, ", ")))
}

// HealthHandler обрабатывает запросы к /health.
// Используется для проверки работоспособности сервиса.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// InfoHandler возвращает информацию об API.
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":        "Users REST API",
		"version":     "1.0.0",
		"description": "RESTful API для управления пользователями",
		"endpoints": []string{
			"GET    /health             - Health check",
			"GET    /api/info           - API информация",
			"GET    /api/users          - Список пользователей",
			"POST   /api/users          - Создать пользователя",
			"GET    /api/users/{id}     - Получить пользователя",
			"PUT    /api/users/{id}     - Обновить пользователя (полностью)",
			"PATCH  /api/users/{id}     - Обновить пользователя (частично)",
			"DELETE /api/users/{id}     - Удалить пользователя",
		},
	})
}

func main() {
	store := NewUserStore()

	userHandler := NewUserHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/api/info", InfoHandler)
	mux.Handle("/api/users", userHandler)
	mux.Handle("/api/users/", userHandler)

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
