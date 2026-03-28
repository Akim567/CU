// gorilla/mux — один из старейших и самых популярных роутеров для Go.
// Он расширяет стандартный http.ServeMux множеством полезных функций.
//
// Официальный репозиторий: https://github.com/gorilla/mux
//
// Ключевые возможности gorilla/mux:
// - Параметры в URL (/users/{id})
// - Регулярные выражения для параметров
// - Поддержка хостов и схем
// - Middleware через Use()
// - Подроутеры (subrouters)
// - Обратный маршрутизация (named routes)
// - Полная совместимость с net/http
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// ===================================================================
// РАЗДЕЛ 1: Модели данных
// ===================================================================

// User представляет модель пользователя.
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Product представляет модель товара.
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	CategoryID  int     `json:"category_id"`
	Description string  `json:"description"`
}


// Store — простое хранилище данных.
type Store struct {
	mu       sync.RWMutex
	users    map[int]*User
	products map[int]*Product
	nextUID  int
	nextPID  int
}

// NewStore создаёт хранилище с тестовыми данными.
func NewStore() *Store {
	s := &Store{
		users:    make(map[int]*User),
		products: make(map[int]*Product),
		nextUID:  1,
		nextPID:  1,
	}

	// Тестовые пользователи
	s.users[1] = &User{ID: 1, Name: "Алиса", Email: "alice@example.com", CreatedAt: time.Now()}
	s.users[2] = &User{ID: 2, Name: "Борис", Email: "boris@example.com", CreatedAt: time.Now()}
	s.nextUID = 3

	// Тестовые товары
	s.products[1] = &Product{ID: 1, Name: "Ноутбук", Price: 75000, CategoryID: 1, Description: "Мощный ноутбук"}
	s.products[2] = &Product{ID: 2, Name: "Смартфон", Price: 45000, CategoryID: 1, Description: "Флагманский смартфон"}
	s.nextPID = 3

	return s
}

var store = NewStore()


// LoggingMiddleware логирует все запросы.
// В gorilla/mux middleware имеет тот же тип, что и в стандартной библиотеке.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)

		// Логируем запрос
		log.Printf("[%s] %s %s %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

// ContentTypeMiddleware устанавливает Content-Type для JSON.
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware проверяет токен аутентификации.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}


// GetUsers возвращает список всех пользователей.
// GET /api/users
func GetUsers(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	users := make([]*User, 0, len(store.users))
	for _, u := range store.users {
		users = append(users, u)
	}

	json.NewEncoder(w).Encode(users)
}

// GetUser возвращает пользователя по ID.
// GET /api/users/{id}
func GetUser(w http.ResponseWriter, r *http.Request) {
	// Извлекаем параметры из URL с помощью mux.Vars()
	// Это ключевая функция gorilla/mux!
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ID"})
		return
	}

	store.mu.RLock()
	user, exists := store.users[id]
	store.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	json.NewEncoder(w).Encode(user)
}

// CreateUser создаёт нового пользователя.
// POST /api/users
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	store.mu.Lock()
	user.ID = store.nextUID
	user.CreatedAt = time.Now()
	store.users[user.ID] = &user
	store.nextUID++
	store.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// UpdateUser обновляет пользователя.
// PUT /api/users/{id}
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var update User
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	user, exists := store.users[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	user.Name = update.Name
	user.Email = update.Email

	json.NewEncoder(w).Encode(user)
}

// DeleteUser удаляет пользователя.
// DELETE /api/users/{id}
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.users[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	delete(store.users, id)
	w.WriteHeader(http.StatusNoContent)
}


// GetProducts возвращает список товаров с фильтрацией.
// GET /api/products
// GET /api/products?category=1&min_price=100&max_price=1000
func GetProducts(w http.ResponseWriter, r *http.Request) {
	// Получаем query параметры
	query := r.URL.Query()
	categoryStr := query.Get("category")
	minPriceStr := query.Get("min_price")
	maxPriceStr := query.Get("max_price")

	store.mu.RLock()
	defer store.mu.RUnlock()

	products := make([]*Product, 0)
	for _, p := range store.products {
		// Фильтр по категории
		if categoryStr != "" {
			categoryID, _ := strconv.Atoi(categoryStr)
			if p.CategoryID != categoryID {
				continue
			}
		}

		// Фильтр по минимальной цене
		if minPriceStr != "" {
			minPrice, _ := strconv.ParseFloat(minPriceStr, 64)
			if p.Price < minPrice {
				continue
			}
		}

		// Фильтр по максимальной цене
		if maxPriceStr != "" {
			maxPrice, _ := strconv.ParseFloat(maxPriceStr, 64)
			if p.Price > maxPrice {
				continue
			}
		}

		products = append(products, p)
	}

	json.NewEncoder(w).Encode(products)
}

// GetProduct возвращает товар по ID.
// GET /api/products/{id}
func GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	store.mu.RLock()
	product, exists := store.products[id]
	store.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}

	json.NewEncoder(w).Encode(product)
}

// GetProductsByCategory возвращает товары по категории.
// GET /api/categories/{categoryId}/products
func GetProductsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, _ := strconv.Atoi(vars["categoryId"])

	store.mu.RLock()
	defer store.mu.RUnlock()

	products := make([]*Product, 0)
	for _, p := range store.products {
		if p.CategoryID == categoryID {
			products = append(products, p)
		}
	}

	json.NewEncoder(w).Encode(products)
}


// PathPrefixHandler демонстрирует обработку путей с префиксом.
func PathPrefixHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Path prefix handler",
		"full_path": r.URL.Path,
	})
}

// RegexHandler демонстрирует использование регулярных выражений в путях.
// GET /api/files/{filename:[a-zA-Z0-9]+\\.pdf}
func RegexHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Regex pattern matched!",
		"filename": vars["filename"],
	})
}

func main() {
	// Создаём новый роутер gorilla/mux
	r := mux.NewRouter()

	r.Use(LoggingMiddleware)
	r.Use(ContentTypeMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/users", GetUsers).Methods("GET")
	api.HandleFunc("/users", CreateUser).Methods("POST")

	api.HandleFunc("/users/{id:[0-9]+}", GetUser).Methods("GET")
	api.HandleFunc("/users/{id:[0-9]+}", UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id:[0-9]+}", DeleteUser).Methods("DELETE")

	api.HandleFunc("/products", GetProducts).Methods("GET")
	api.HandleFunc("/products/{id:[0-9]+}", GetProduct).Methods("GET")

	api.HandleFunc("/categories/{categoryId:[0-9]+}/products", GetProductsByCategory).Methods("GET")

	api.HandleFunc("/users/{id:[0-9]+}/profile", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Profile of user %s", vars["id"]),
		})
	}).Methods("GET").Name("userProfile")

	api.HandleFunc("/files/{filename:[a-zA-Z0-9]+\\.pdf}", RegexHandler).Methods("GET")

	protected := api.PathPrefix("/admin").Subrouter()
	protected.Use(AuthMiddleware)

	protected.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Welcome to admin dashboard!",
		})
	}).Methods("GET")

	protected.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		store.mu.RLock()
		defer store.mu.RUnlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_users":    len(store.users),
			"total_products": len(store.products),
		})
	}).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	}).Methods("GET")

	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))),
	)

	r.HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
		url, err := r.URL.Parse("/")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		profileURL, _ := api.Get("userProfile").URL("id", "42")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"base_url":         url.String(),
			"user_profile_url": profileURL.String(),
			"message":          "Демонстрация обратной маршрутизации",
		})
	}).Methods("GET")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
