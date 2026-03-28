// Fiber — это веб-фреймворк, вдохновлённый Express.js, построенный на fasthttp —
// самом быстром HTTP-движке для Go. Fiber разработан для упрощения разработки
// для тех, кто пришёл из Node.js/Express мира.
//
// Официальный сайт: https://gofiber.io/
// GitHub: https://github.com/gofiber/fiber
//
// Ключевые особенности Fiber:
// - Построен на fasthttp (в 10 раз быстрее net/http)
// - Express-подобный синтаксис
// - Низкое потребление памяти
// - Встроенные middleware
// - Статические файлы, WebSocket, Rate Limiting
// - Template engines
// - Простой и понятный API
//
// ВАЖНО: Fiber НЕ использует net/http, поэтому не совместим с
// стандартными http.Handler и http.HandlerFunc!
package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Order представляет модель заказа.
type Order struct {
	ID          int         `json:"id"`
	CustomerID  int         `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"` // pending, processing, shipped, delivered, cancelled
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderItem представляет позицию в заказе.
type OrderItem struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// CreateOrderRequest — запрос на создание заказа.
type CreateOrderRequest struct {
	CustomerID int         `json:"customer_id"`
	Items      []OrderItem `json:"items"`
}

// UpdateStatusRequest — запрос на обновление статуса.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// OrderStore — хранилище заказов.
type OrderStore struct {
	mu     sync.RWMutex
	orders map[int]*Order
	nextID int
}

// NewOrderStore создаёт хранилище с тестовыми данными.
func NewOrderStore() *OrderStore {
	s := &OrderStore{
		orders: make(map[int]*Order),
		nextID: 1,
	}

	// Тестовые данные
	now := time.Now()
	s.orders[1] = &Order{
		ID:         1,
		CustomerID: 101,
		Items: []OrderItem{
			{ProductID: 1, ProductName: "Laptop", Quantity: 1, Price: 999.99},
			{ProductID: 2, ProductName: "Mouse", Quantity: 2, Price: 29.99},
		},
		TotalAmount: 1059.97,
		Status:      "processing",
		CreatedAt:   now.AddDate(0, 0, -2),
		UpdatedAt:   now,
	}
	s.orders[2] = &Order{
		ID:         2,
		CustomerID: 102,
		Items: []OrderItem{
			{ProductID: 3, ProductName: "Keyboard", Quantity: 1, Price: 149.99},
		},
		TotalAmount: 149.99,
		Status:      "pending",
		CreatedAt:   now.AddDate(0, 0, -1),
		UpdatedAt:   now.AddDate(0, 0, -1),
	}
	s.nextID = 3

	return s
}

var store = NewOrderStore()

// AuthMiddleware проверяет токен авторизации.
// В Fiber middleware возвращает fiber.Handler.
func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token != "Bearer admin-token" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Valid authorization token required",
		})
	}
	// Сохраняем информацию о пользователе в Locals
	c.Locals("user", "admin")
	return c.Next()
}

// TimingMiddleware измеряет время выполнения.
func TimingMiddleware(c *fiber.Ctx) error {
	start := time.Now()

	// Обрабатываем запрос
	err := c.Next()

	// Добавляем заголовок с временем выполнения
	c.Set("X-Response-Time", time.Since(start).String())

	return err
}

// ListOrders возвращает список заказов.
// GET /api/orders
// GET /api/orders?status=pending&customer_id=101
func ListOrders(c *fiber.Ctx) error {
	// Получаем query параметры
	status := c.Query("status")
	customerIDStr := c.Query("customer_id")

	store.mu.RLock()
	defer store.mu.RUnlock()

	orders := make([]*Order, 0)
	for _, o := range store.orders {
		// Фильтр по статусу
		if status != "" && o.Status != status {
			continue
		}

		// Фильтр по customer_id
		if customerIDStr != "" {
			customerID, _ := strconv.Atoi(customerIDStr)
			if o.CustomerID != customerID {
				continue
			}
		}

		orders = append(orders, o)
	}

	return c.JSON(fiber.Map{
		"data":  orders,
		"count": len(orders),
	})
}

// GetOrder возвращает заказ по ID.
// GET /api/orders/:id
func GetOrder(c *fiber.Ctx) error {
	// c.Params() извлекает параметр из URL
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	store.mu.RLock()
	order, exists := store.orders[id]
	store.mu.RUnlock()

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": order,
	})
}

// CreateOrder создаёт новый заказ.
// POST /api/orders
func CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest

	// c.BodyParser() парсит тело запроса
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Валидация
	if req.CustomerID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer ID is required",
		})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one item is required",
		})
	}

	// Вычисляем общую сумму
	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	now := time.Now()
	store.mu.Lock()
	order := &Order{
		ID:          store.nextID,
		CustomerID:  req.CustomerID,
		Items:       req.Items,
		TotalAmount: totalAmount,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	store.orders[order.ID] = order
	store.nextID++
	store.mu.Unlock()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Order created successfully",
		"data":    order,
	})
}

// UpdateOrderStatus обновляет статус заказа.
// PATCH /api/orders/:id/status
func UpdateOrderStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Валидация статуса
	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"shipped":    true,
		"delivered":  true,
		"cancelled":  true,
	}
	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":          "Invalid status",
			"valid_statuses": []string{"pending", "processing", "shipped", "delivered", "cancelled"},
		})
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	order, exists := store.orders[id]
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	order.Status = req.Status
	order.UpdatedAt = time.Now()

	return c.JSON(fiber.Map{
		"message": "Order status updated",
		"data":    order,
	})
}

// CancelOrder отменяет заказ.
// POST /api/orders/:id/cancel
func CancelOrder(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	order, exists := store.orders[id]
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// Проверяем, можно ли отменить заказ
	if order.Status == "shipped" || order.Status == "delivered" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot cancel shipped or delivered order",
		})
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()

	return c.JSON(fiber.Map{
		"message": "Order cancelled",
		"data":    order,
	})
}

// DeleteOrder удаляет заказ.
// DELETE /api/orders/:id
func DeleteOrder(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.orders[id]; !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	delete(store.orders, id)
	return c.SendStatus(fiber.StatusNoContent)
}

// OrderStats возвращает статистику по заказам.
// GET /api/orders/stats
func OrderStats(c *fiber.Ctx) error {
	store.mu.RLock()
	defer store.mu.RUnlock()

	stats := fiber.Map{
		"total":         len(store.orders),
		"by_status":     make(map[string]int),
		"total_revenue": 0.0,
		"average_order": 0.0,
	}

	var totalRevenue float64
	statusCount := make(map[string]int)

	for _, o := range store.orders {
		statusCount[o.Status]++
		if o.Status != "cancelled" {
			totalRevenue += o.TotalAmount
		}
	}

	stats["by_status"] = statusCount
	stats["total_revenue"] = totalRevenue
	if len(store.orders) > 0 {
		stats["average_order"] = totalRevenue / float64(len(store.orders))
	}

	return c.JSON(fiber.Map{
		"data": stats,
	})
}

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "Orders API v1.0.0",
		ServerHeader: "Fiber",

		// Обработка ошибок
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Получаем код ошибки
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"

			// Проверяем тип ошибки
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			return c.Status(code).JSON(fiber.Map{
				"error": message,
			})
		},

		// Производительность
		Prefork:       false, // Включить для multi-process mode
		CaseSensitive: true,
		StrictRouting: false,
	})

	// Recovery — восстановление после паники
	app.Use(recover.New())

	app.Use(requestid.New())

	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	app.Use(TimingMiddleware)


	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	api := app.Group("/api")

	api.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests",
			})
		},
	}))

	orders := api.Group("/orders")

	// Публичные маршруты
	orders.Get("/", ListOrders)
	orders.Get("/stats", OrderStats)
	orders.Get("/:id", GetOrder)

	// Защищённые маршруты
	protected := orders.Group("")
	protected.Use(AuthMiddleware)
	{
		protected.Post("/", CreateOrder)
		protected.Patch("/:id/status", UpdateOrderStatus)
		protected.Post("/:id/cancel", CancelOrder)
		protected.Delete("/:id", DeleteOrder)
	}


	// Redirect
	app.Get("/old-orders", func(c *fiber.Ctx) error {
		return c.Redirect("/api/orders")
	})

	// Разные форматы ответа
	app.Get("/format", func(c *fiber.Ctx) error {
		data := fiber.Map{
			"message": "Hello from Fiber!",
			"format":  c.Accepts("application/json", "text/xml"),
		}

		// Fiber автоматически определяет формат по Accept
		return c.JSON(data)
	})

	app.Get("/download", func(c *fiber.Ctx) error {
	    return c.Download("./file.pdf", "report.pdf")
	})

	// WebSocket (простой пример)
	// app.Get("/ws", websocket.New(func(c *websocket.Conn) {
	//     // WebSocket handler
	// }))

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "Route not found",
			"path":   c.Path(),
			"method": c.Method(),
		})
	})

	if err := app.Listen(":8080"); err != nil {
		panic(err)
	}
}
