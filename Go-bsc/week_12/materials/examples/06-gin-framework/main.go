// Gin — это высокопроизводительный HTTP веб-фреймворк, написанный на Go.
// Он предоставляет маршрутизацию с использованием httprouter, middleware,
// рендеринг JSON, валидацию и многое другое.
//
// Официальный сайт: https://gin-gonic.com/
// GitHub: https://github.com/gin-gonic/gin
//
// Ключевые особенности Gin:
// - Очень быстрый (использует httprouter)
// - Middleware поддержка
// - Crash-free (Recovery middleware)
// - JSON валидация
// - Route groups
// - Error management
// - Рендеринг (JSON, XML, HTML)
// - Extendable
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Book представляет модель книги.
// Gin использует теги для валидации и биндинга.
type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title" binding:"required,min=1,max=200"`
	Author      string    `json:"author" binding:"required"`
	ISBN        string    `json:"isbn" binding:"required,len=13"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	PublishedAt time.Time `json:"published_at"`
	InStock     bool      `json:"in_stock"`
	Quantity    int       `json:"quantity" binding:"gte=0"`
}

// BookUpdate — структура для частичного обновления.
type BookUpdate struct {
	Title    *string  `json:"title" binding:"omitempty,min=1,max=200"`
	Author   *string  `json:"author"`
	Price    *float64 `json:"price" binding:"omitempty,gt=0"`
	InStock  *bool    `json:"in_stock"`
	Quantity *int     `json:"quantity" binding:"omitempty,gte=0"`
}

// LoginRequest — структура для логина.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// BookStore — хранилище книг.
type BookStore struct {
	mu     sync.RWMutex
	books  map[int]*Book
	nextID int
}

// NewBookStore создаёт хранилище с тестовыми данными.
func NewBookStore() *BookStore {
	s := &BookStore{
		books:  make(map[int]*Book),
		nextID: 1,
	}

	// Тестовые данные
	testBooks := []Book{
		{Title: "The Go Programming Language", Author: "Alan Donovan", ISBN: "9780134190440", Price: 45.99, InStock: true, Quantity: 10},
		{Title: "Clean Code", Author: "Robert Martin", ISBN: "9780132350884", Price: 39.99, InStock: true, Quantity: 5},
		{Title: "Design Patterns", Author: "Gang of Four", ISBN: "9780201633610", Price: 54.99, InStock: false, Quantity: 0},
	}

	for _, b := range testBooks {
		book := b
		book.ID = s.nextID
		book.PublishedAt = time.Now().AddDate(-1, 0, 0)
		s.books[book.ID] = &book
		s.nextID++
	}

	return s
}

var store = NewBookStore()

// RequestLogger логирует информацию о запросе.
// В Gin middleware — это функция, возвращающая gin.HandlerFunc.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Обрабатываем запрос
		c.Next()

		// После обработки
		latency := time.Since(start)
		status := c.Writer.Status()

		fmt.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s%s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			status,
			latency,
			c.ClientIP(),
			c.Request.Method,
			path,
			func() string {
				if query != "" {
					return "?" + query
				}
				return ""
			}(),
		)
	}
}

// AuthRequired проверяет авторизацию.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "Bearer valid-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Valid authorization token required",
			})
			return
		}
		// Сохраняем информацию о пользователе в контексте
		c.Set("user", "authenticated_user")
		c.Next()
	}
}

// RateLimiter ограничивает количество запросов.
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	type client struct {
		count    int
		lastSeen time.Time
	}
	clients := make(map[string]*client)
	mu := sync.Mutex{}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		cl, exists := clients[ip]
		if !exists || time.Since(cl.lastSeen) > window {
			clients[ip] = &client{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}

		cl.count++
		if cl.count > maxRequests {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": window.Seconds(),
			})
			return
		}
		mu.Unlock()
		c.Next()
	}
}

// ListBooks возвращает список книг с фильтрацией и пагинацией.
// GET /api/v1/books
// GET /api/v1/books?in_stock=true&page=1&limit=10
func ListBooks(c *gin.Context) {
	// Получаем query параметры с значениями по умолчанию
	inStockStr := c.DefaultQuery("in_stock", "")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	var books []*Book
	for _, b := range store.books {
		// Фильтр по наличию
		if inStockStr != "" {
			inStock := inStockStr == "true"
			if b.InStock != inStock {
				continue
			}
		}
		books = append(books, b)
	}

	// Пагинация
	total := len(books)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"data": books[start:end],
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetBook возвращает книгу по ID.
// GET /api/v1/books/:id
func GetBook(c *gin.Context) {
	// c.Param() извлекает параметр из URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	store.mu.RLock()
	book, exists := store.books[id]
	store.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": book})
}

// CreateBook создаёт новую книгу.
// POST /api/v1/books
func CreateBook(c *gin.Context) {
	var book Book

	// ShouldBindJSON связывает JSON и валидирует по тегам binding
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	store.mu.Lock()
	book.ID = store.nextID
	book.PublishedAt = time.Now()
	store.books[book.ID] = &book
	store.nextID++
	store.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Book created successfully",
		"data":    book,
	})
}

// UpdateBook полностью обновляет книгу.
// PUT /api/v1/books/:id
func UpdateBook(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var book Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	existing, exists := store.books[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	book.ID = id
	book.PublishedAt = existing.PublishedAt
	store.books[id] = &book

	c.JSON(http.StatusOK, gin.H{
		"message": "Book updated successfully",
		"data":    book,
	})
}

// PatchBook частично обновляет книгу.
// PATCH /api/v1/books/:id
func PatchBook(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var update BookUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	book, exists := store.books[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Обновляем только переданные поля
	if update.Title != nil {
		book.Title = *update.Title
	}
	if update.Author != nil {
		book.Author = *update.Author
	}
	if update.Price != nil {
		book.Price = *update.Price
	}
	if update.InStock != nil {
		book.InStock = *update.InStock
	}
	if update.Quantity != nil {
		book.Quantity = *update.Quantity
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Book patched successfully",
		"data":    book,
	})
}

// DeleteBook удаляет книгу.
// DELETE /api/v1/books/:id
func DeleteBook(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.books[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	delete(store.books, id)
	c.Status(http.StatusNoContent)
}

// Login обрабатывает вход пользователя.
// POST /api/v1/auth/login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Простая проверка (в реальном приложении — проверка в БД)
	if req.Username == "admin" && req.Password == "password123" {
		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   "valid-token",
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Invalid credentials",
	})
}

// Profile возвращает профиль текущего пользователя.
// GET /api/v1/auth/profile
func Profile(c *gin.Context) {
	// Получаем пользователя из контекста (установлен в AuthRequired)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "This is your profile",
	})
}

func main() {
	// Устанавливаем режим Gin (debug, release, test)
	// gin.SetMode(gin.ReleaseMode) // для production

	// Создаём роутер Gin
	// gin.Default() включает Logger и Recovery middleware
	r := gin.Default()

	// Или создаём без middleware: gin.New()
	// r := gin.New()
	// r.Use(gin.Logger())
	// r.Use(gin.Recovery())

	r.Use(RateLimiter(100, time.Minute))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", Login)
		}

		books := v1.Group("/books")
		{
			books.GET("", ListBooks)
			books.GET("/:id", GetBook)

			protected := books.Group("")
			protected.Use(AuthRequired())
			{
				protected.POST("", CreateBook)
				protected.PUT("/:id", UpdateBook)
				protected.PATCH("/:id", PatchBook)
				protected.DELETE("/:id", DeleteBook)
			}
		}

		profile := v1.Group("/auth")
		profile.Use(AuthRequired())
		{
			profile.GET("/profile", Profile)
		}
	}

	// =====================================================
	// Статические файлы
	// =====================================================
	// r.Static("/static", "./static")
	// r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// =====================================================
	// Custom 404
	// =====================================================
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Route not found",
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		})
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
