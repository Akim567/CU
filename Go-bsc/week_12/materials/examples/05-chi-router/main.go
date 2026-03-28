// Chi — это компактный, идиоматичный и композируемый роутер для создания HTTP сервисов.
// Он построен на стандартной библиотеке net/http и не использует внешних зависимостей.
//
// Официальный репозиторий: https://github.com/go-chi/chi
//
// Ключевые особенности chi:
// - 100% совместим с net/http
// - Легковесный и быстрый
// - Middleware стек (встроенные и кастомные)
// - URL параметры и паттерны
// - Группировка маршрутов (Route groups)
// - Контекст запроса для передачи данных
// - Встроенные middleware (Logger, Recoverer, и др.)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Article представляет модель статьи.
type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  int       `json:"author_id"`
	Tags      []string  `json:"tags"`
	Published bool      `json:"published"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Comment представляет модель комментария.
type Comment struct {
	ID        int       `json:"id"`
	ArticleID int       `json:"article_id"`
	AuthorID  int       `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ArticleStore — хранилище статей.
type ArticleStore struct {
	mu       sync.RWMutex
	articles map[int]*Article
	comments map[int][]*Comment // articleID -> comments
	nextAID  int
	nextCID  int
}

// NewArticleStore создаёт хранилище с тестовыми данными.
func NewArticleStore() *ArticleStore {
	s := &ArticleStore{
		articles: make(map[int]*Article),
		comments: make(map[int][]*Comment),
		nextAID:  1,
		nextCID:  1,
	}

	// Тестовые данные
	now := time.Now()
	s.articles[1] = &Article{
		ID:        1,
		Title:     "Введение в Go",
		Content:   "Go — это компилируемый многопоточный язык программирования...",
		AuthorID:  1,
		Tags:      []string{"go", "programming", "tutorial"},
		Published: true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.articles[2] = &Article{
		ID:        2,
		Title:     "Chi Router: Полное руководство",
		Content:   "Chi — это легковесный и идиоматичный роутер для Go...",
		AuthorID:  1,
		Tags:      []string{"go", "chi", "web"},
		Published: true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.nextAID = 3

	// Тестовые комментарии
	s.comments[1] = []*Comment{
		{ID: 1, ArticleID: 1, AuthorID: 2, Content: "Отличная статья!", CreatedAt: now},
		{ID: 2, ArticleID: 1, AuthorID: 3, Content: "Очень полезно, спасибо!", CreatedAt: now},
	}
	s.nextCID = 3

	return s
}

var store = NewArticleStore()

// RequestIDMiddleware добавляет Request-ID к каждому запросу.
// Chi имеет встроенный middleware.RequestID, но мы создадим свой для демонстрации.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ContentTypeJSON устанавливает Content-Type: application/json.
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// AdminOnly проверяет права администратора.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("X-User-Role")
		if role != "admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Admin access required",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ArticleCtx загружает статью в контекст по ID из URL.
// Это мощный паттерн chi для работы с ресурсами.
func ArticleCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// chi.URLParam извлекает параметр из URL
		articleIDStr := chi.URLParam(r, "articleID")
		articleID, err := strconv.Atoi(articleIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid article ID"})
			return
		}

		store.mu.RLock()
		article, exists := store.articles[articleID]
		store.mu.RUnlock()

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Article not found"})
			return
		}

		// Добавляем статью в контекст запроса
		ctx := context.WithValue(r.Context(), "article", article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ListArticles возвращает список всех статей.
// GET /articles
func ListArticles(w http.ResponseWriter, r *http.Request) {
	// Получаем query параметры для фильтрации
	publishedOnly := r.URL.Query().Get("published") == "true"
	tag := r.URL.Query().Get("tag")

	store.mu.RLock()
	defer store.mu.RUnlock()

	articles := make([]*Article, 0)
	for _, a := range store.articles {
		// Фильтр по published
		if publishedOnly && !a.Published {
			continue
		}

		// Фильтр по тегу
		if tag != "" {
			hasTag := false
			for _, t := range a.Tags {
				if t == tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		articles = append(articles, a)
	}

	json.NewEncoder(w).Encode(articles)
}

// CreateArticle создаёт новую статью.
// POST /articles
func CreateArticle(w http.ResponseWriter, r *http.Request) {
	var article Article
	if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	now := time.Now()
	store.mu.Lock()
	article.ID = store.nextAID
	article.CreatedAt = now
	article.UpdatedAt = now
	store.articles[article.ID] = &article
	store.nextAID++
	store.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(article)
}

// GetArticle возвращает статью из контекста.
// GET /articles/{articleID}
func GetArticle(w http.ResponseWriter, r *http.Request) {
	// Статья уже загружена в контекст через ArticleCtx middleware
	article := r.Context().Value("article").(*Article)
	json.NewEncoder(w).Encode(article)
}

// UpdateArticle обновляет статью.
// PUT /articles/{articleID}
func UpdateArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*Article)

	var update Article
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	store.mu.Lock()
	article.Title = update.Title
	article.Content = update.Content
	article.Tags = update.Tags
	article.Published = update.Published
	article.UpdatedAt = time.Now()
	store.mu.Unlock()

	json.NewEncoder(w).Encode(article)
}

// DeleteArticle удаляет статью.
// DELETE /articles/{articleID}
func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*Article)

	store.mu.Lock()
	delete(store.articles, article.ID)
	delete(store.comments, article.ID)
	store.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// ListComments возвращает комментарии к статье.
// GET /articles/{articleID}/comments
func ListComments(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*Article)

	store.mu.RLock()
	comments := store.comments[article.ID]
	store.mu.RUnlock()

	if comments == nil {
		comments = []*Comment{}
	}

	json.NewEncoder(w).Encode(comments)
}

// CreateComment создаёт новый комментарий.
// POST /articles/{articleID}/comments
func CreateComment(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*Article)

	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	store.mu.Lock()
	comment.ID = store.nextCID
	comment.ArticleID = article.ID
	comment.CreatedAt = time.Now()
	store.comments[article.ID] = append(store.comments[article.ID], &comment)
	store.nextCID++
	store.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// SearchArticles поиск статей.
// GET /search?q=...
func SearchArticles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Query parameter 'q' is required"})
		return
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	results := make([]*Article, 0)
	for _, a := range store.articles {
		// Простой поиск в заголовке и контенте
		if containsIgnoreCase(a.Title, query) || containsIgnoreCase(a.Content, query) {
			results = append(results, a)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	})
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && containsIgnoreCase(s[1:], substr) ||
		len(s) >= len(substr) && s[:len(substr)] == substr)
}

// AdminDashboard — защищённая админ-панель.
// GET /admin/dashboard
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	totalComments := 0
	for _, comments := range store.comments {
		totalComments += len(comments)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_articles": len(store.articles),
		"total_comments": totalComments,
		"server_time":    time.Now().Format(time.RFC3339),
	})
}

func main() {
	// Создаём новый роутер chi
	r := chi.NewRouter()

	// Восстановление после паники
	r.Use(middleware.Recoverer)

	// Логирование запросов (встроенный middleware chi)
	r.Use(middleware.Logger)

	// Request ID
	r.Use(middleware.RequestID)

	// Таймаут для запросов (60 секунд)
	r.Use(middleware.Timeout(60 * time.Second))

	// Наш кастомный middleware для JSON
	r.Use(ContentTypeJSON)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	r.Route("/api", func(r chi.Router) {
		// Информация об API
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{
				"name":    "Articles API",
				"version": "1.0.0",
			})
		})

		// Поиск
		r.Get("/search", SearchArticles)

		r.Route("/articles", func(r chi.Router) {
			// GET /api/articles — список статей
			r.Get("/", ListArticles)

			// POST /api/articles — создать статью
			r.Post("/", CreateArticle)

			// Маршруты для конкретной статьи
			// {articleID} — параметр в URL
			r.Route("/{articleID}", func(r chi.Router) {
				// Применяем ArticleCtx middleware только к этой группе
				// Он загрузит статью в контекст
				r.Use(ArticleCtx)

				// GET /api/articles/{articleID}
				r.Get("/", GetArticle)

				// PUT /api/articles/{articleID}
				r.Put("/", UpdateArticle)

				// DELETE /api/articles/{articleID}
				r.Delete("/", DeleteArticle)

				r.Route("/comments", func(r chi.Router) {
					// GET /api/articles/{articleID}/comments
					r.Get("/", ListComments)

					// POST /api/articles/{articleID}/comments
					r.Post("/", CreateComment)
				})
			})
		})

		r.Route("/admin", func(r chi.Router) {
			// Применяем AdminOnly middleware ко всем админ-маршрутам
			r.Use(AdminOnly)

			r.Get("/dashboard", AdminDashboard)
		})
	})

	// Можно создать отдельный роутер и подключить его
	apiV2 := chi.NewRouter()
	apiV2.Get("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "API v2 coming soon!",
		})
	})
	r.Mount("/api/v2", apiV2)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Route not found",
			"path":  r.URL.Path,
		})
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "Method not allowed",
			"method": r.Method,
		})
	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
