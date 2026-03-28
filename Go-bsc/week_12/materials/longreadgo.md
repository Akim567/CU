# Лекция: Протокол HTTP и реализация REST-сервисов

## Введение в мир веб-протоколов

Представьте, что вы отправляете письмо по почте. Вы пишете адрес получателя, указываете обратный адрес, кладёте письмо в конверт и опускаете в почтовый ящик. Почтовая служба доставляет ваше письмо, получатель его читает и может отправить вам ответ.

Протокол HTTP работает похоже — это набор правил для общения между компьютерами в интернете. Каждый день вы используете HTTP сотни раз: когда открываете сайты, смотрите видео, общаетесь в мессенджерах. За всем этим стоит простой, но мощный протокол, который делает современный интернет возможным.

### Что такое HTTP и зачем он нужен?

HTTP (HyperText Transfer Protocol) — это протокол прикладного уровня для передачи данных в интернете.

**Основные задачи HTTP:**
- Передача веб-страниц, изображений, видео и других файлов
- Обработка форм и пользовательского ввода
- Управление состоянием соединения (cookies, сессии)
- Поддержка разных типов данных (JSON, XML, текст, бинарные файлы)

### Немного истории: как HTTP изменил мир

HTTP появился в 1991 году вместе с Всемирной паутиной. Создатель Тим Бернерс-Ли просто хотел способ делиться научными документами между университетами.

**Эволюция HTTP:**
- **HTTP/0.9 (1991)** — Очень простой: только GET-запросы, только HTML
- **HTTP/1.0 (1996)** — Добавил POST, HEAD, статусы ответов, заголовки
- **HTTP/1.1 (1997)** — Стабильная версия с keep-alive, pipelining, хостами
- **HTTP/2 (2015)** — Мультиплексирование, сжатие заголовков, приоритеты
- **HTTP/3 (2022)** — Работает поверх QUIC (UDP), быстрее и надёжнее

Интересный факт: HTTP/1.1 служил человечеству почти 20 лет без серьёзных изменений! Это говорит о том, насколько гениальным и универсальным был его дизайн.

### Как устроен HTTP: запрос-ответ

В основе HTTP лежит простая модель **клиент-сервер**:

1. **Клиент** (браузер, мобильное приложение) отправляет **запрос**
2. **Сервер** обрабатывает запрос и отправляет **ответ**
3. **Соединение** может быть как кратковременным, так и долгоживущим

> [!IMPORTANT]
> HTTP — **протокол без сохранения состояния** (stateless). Каждый запрос независим от других. Сервер не помнит предыдущие запросы от того же клиента. Для сохранения состояния используются cookies и сессии.

---

## Анатомия HTTP-запроса

Давайте заглянем "под капот" HTTP и посмотрим, из чего состоит запрос. Представьте, что вы отправляете письмо — у него есть конверт (метаданные) и содержимое (тело письма). HTTP-запрос устроен так же.

### Структура HTTP-запроса

HTTP-запрос состоит из трёх основных частей:

```
GET /api/users/123 HTTP/1.1          ← Стартовая строка
Host: example.com                    ← Заголовки
User-Agent: Mozilla/5.0
Accept: application/json
Authorization: Bearer abc123

                                       ← Пустая строка
                                       ← Тело запроса (для POST/PUT)
```

#### 1. Стартовая строка

Формат: `Метод URI Версия`

**Методы HTTP:**
- **GET** — Получить ресурс (прочитать веб-страницу, скачать файл)
- **POST** — Создать новый ресурс (отправить форму, создать пользователя)
- **PUT** — Полностью обновить ресурс (заменить данные пользователя)
- **PATCH** — Частично обновить ресурс (изменить только email пользователя)
- **DELETE** — Удалить ресурс
- **HEAD** — Получить только заголовки (без тела ответа)
- **OPTIONS** — Узнать допустимые методы для ресурса

**URI (Uniform Resource Identifier):**
- **Путь** — `/api/users/123`
- **Параметры запроса** — `/api/users?sort=name&limit=10`
- **Фрагмент** — `/page.html#section1` (не отправляется на сервер)

#### 2. Заголовки (Headers)

Заголовки передают метаданные о запросе:

```http
Host: api.example.com              ← Обязательный заголовок
User-Agent: MyApp/1.0              ← Идентификатор клиента
Accept: application/json           ← Какие форматы ответа принимает клиент
Content-Type: application/json     ← Тип содержимого в теле запроса
Content-Length: 156                ← Размер тела в байтах
Authorization: Bearer xyz123       ← Аутентификация
Cache-Control: no-cache            ← Управление кэшированием
```

#### 3. Тело запроса

Тело запроса содержит данные, которые отправляются на сервер. Используется в POST, PUT, PATCH запросах:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 30
}
```

### Практический пример: разбор реального запроса

Давайте посмотрим на реальный запрос, который отправляет браузер при посещении Google:

```http
GET / HTTP/1.1
Host: www.google.com
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Connection: keep-alive
Upgrade-Insecure-Requests: 1
```

**Что здесь происходит:**
- `GET /` — просим корневую страницу сайта
- `Host: www.google.com` — указываем, какой именно сайт хотим посетить
- `User-Agent` — представляемся браузером Chrome
- `Accept` — говорим, что принимаем HTML и XML
- `Connection: keep-alive` — просим держать соединение открытым

---

## Анатомия HTTP-ответа

Когда сервер получает запрос, он обрабатывает его и отправляет ответ. Ответ HTTP имеет похожую структуру на запрос, но с другими компонентами.

### Структура HTTP-ответа

```
HTTP/1.1 200 OK                     ← Стартовая строка ответа
Content-Type: application/json     ← Заголовки ответа
Content-Length: 1234
Date: Mon, 23 Oct 2023 12:00:00 GMT
Server: nginx/1.18.0

                                       ← Пустая строка
{                                      ← Тело ответа
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com"
}
```

#### 1. Стартовая строка ответа

Формат: `Версия Код_статуса Текст_статуса`

**Коды состояния HTTP:**

**2xx (Успешно):**
- `200 OK` — Запрос успешно выполнен
- `201 Created` — Ресурс успешно создан
- `204 No Content` — Успешно, но без содержимого в ответе

**3xx (Перенаправление):**
- `301 Moved Permanently` — Ресурс навсегда перемещён
- `302 Found` — Временное перенаправление
- `304 Not Modified` — Ресурс не изменён (используйте кэш)

**4xx (Ошибка клиента):**
- `400 Bad Request` — Неверный формат запроса
- `401 Unauthorized` — Нужна аутентификация
- `403 Forbidden` — Доступ запрещён
- `404 Not Found` — Ресурс не найден
- `429 Too Many Requests` — Слишком много запросов

**5xx (Ошибка сервера):**
- `500 Internal Server Error` — Внутренняя ошибка сервера
- `502 Bad Gateway` — Неверный ответ от upstream-сервера
- `503 Service Unavailable` — Сервис временно недоступен

#### 2. Заголовки ответа

Сервер отправляет важные метаданные в заголовках:

```http
Content-Type: application/json     ← Тип содержимого
Content-Length: 1024               ← Размер тела
Cache-Control: max-age=3600         ← Правила кэширования
Set-Cookie: session=abc123; HttpOnly ← Установка cookies
Server: Apache/2.4.41              ← Информация о сервере
Date: Mon, 23 Oct 2023 12:00:00 GMT ← Время ответа
```

#### 3. Тело ответа

Содержимое, которое получает клиент:

```html
<!DOCTYPE html>
<html>
<head><title>Welcome</title></head>
<body><h1>Hello, World!</h1></body>
</html>
```

или JSON:
```json
{
  "status": "success",
  "data": {
    "users": [
      {"id": 1, "name": "Alice"},
      {"id": 2, "name": "Bob"}
    ]
  }
}
```

---

## REST: Архитектурный стиль веб-сервисов

Теперь, когда мы понимаем, как работает HTTP, давайте поговорим о REST — самом популярном подходе к созданию веб-сервисов.

### Что такое REST?

REST (Representational State Transfer) — это архитектурный стиль, а не протокол. Это набор принципов и ограничений для создания распределённых систем.

**Ключевые принципы REST:**

1. **Клиент-серверная архитектура** — Чёткое разделение ответственности
2. **Отсутствие состояния** (Stateless) — Каждый запрос содержит всю необходимую информацию
3. **Кэшируемость** — Ответы должны отмечаться как кэшируемые или нет
4. **Единый интерфейс** — Стандартизированный способ взаимодействия
5. **Слоистая система** — Промежуточные слои могут быть между клиентом и сервером
6. **Код по требованию** (необязательно) — Сервер может отправлять исполняемый код

> [!NOTE]
> REST — это не строгий стандарт, а философия. Однако у философии также бывают обязательные и не обязательные принципы. У REST все принципы являются обязательными за исключением последнего, поэтому RESTful сервисом считается тот, что реализовывает все обязательные принципы.

### RESTful ресурсы и URI

В REST всё вокруг **ресурсов**. Ресурс — это любая сущность, с которой мы хотим работать: пользователь, товар, заказ, изображение.

**Правила хорошего URI дизайна:**

```http
# ✅ ХОРОШО: Ресурсно-ориентированные URI
GET    /api/users          — Получить всех пользователей
GET    /api/users/123      — Получить пользователя с ID=123
POST   /api/users          — Создать нового пользователя
PUT    /api/users/123      — Обновить пользователя 123
DELETE /api/users/123      — Удалить пользователя 123

GET    /api/users/123/orders     — Заказы пользователя 123
GET    /api/orders/123/items     — Позиции в заказе 123

# ❌ ПЛОХО: Глаголы в URI
GET    /api/getUsers        — Неправильно!
POST   /api/createUser     — Неправильно!
PUT    /api/updateUser/123 — Неправильно!
```

### HTTP-методы в контексте REST

В REST HTTP-методы соответствуют CRUD-операциям:

| HTTP метод | CRUD операция | Пример использования |
|------------|---------------|---------------------|
| GET        | Read          | `GET /api/products` |
| POST       | Create        | `POST /api/products` |
| PUT/PATCH  | Update        | `PUT /api/products/123` |
| DELETE     | Delete        | `DELETE /api/products/123` |

### Форматы данных в REST

Хотя REST может работать с любым форматом данных, **JSON** стал де-факто стандартом:

```json
{
  "id": 123,
  "name": "iPhone 15 Pro",
  "price": 999.99,
  "inStock": true,
  "category": {
    "id": 5,
    "name": "Smartphones"
  },
  "tags": ["apple", "smartphone", "premium"]
}
```

**Почему JSON так популярен:**
- Легко читается человеком
- Легко парсится в большинстве языков
- Поддерживает вложенные структуры
- Меньше избыточности, чем XML

---

## HTTP-серверы в Go: основы

Go имеет прекрасную стандартную библиотеку `net/http` для создания веб-серверов. Давайте начнём с простого и постепенно будем усложнять.

### Наш первый HTTP-сервер

```go
package main

import (
    "fmt"
    "log"
    "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World! Ваш запрос: %s %s", r.Method, r.URL.Path)
}

func main() {
    // Регистрируем обработчик для корневого пути
    http.HandleFunc("/", helloHandler)

    fmt.Println("Сервер запущен на :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Что здесь происходит:**
- `http.HandleFunc` регистрирует функцию-обработчик для пути
- `helloHandler` получает два параметра:
  - `http.ResponseWriter` — для записи ответа
  - `*http.Request` — информация о запросе
- `http.ListenAndServe` запускает сервер на указанном порту

Запустим и проверим:
```bash
go run server.go
curl http://localhost:8080
# Вывод: Hello, World! Ваш запрос: GET /
```

### Анатомия HTTP-обработчика в Go

Давайте разберём подробнее, что такое обработчик (handler) в Go:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // w — интерфейс для записи ответа
    // r — структура с информацией о запросе
}
```

**http.ResponseWriter** позволяет:
- Записывать заголовки: `w.Header().Set("Content-Type", "application/json")`
- Устанавливать статус: `w.WriteHeader(http.StatusNotFound)`
- Записывать тело ответа: `w.Write([]byte("response"))`

**http.Request** содержит:
- `r.Method` — HTTP метод (GET, POST, и т.д.)
- `r.URL.Path` — путь запроса (/api/users)
- `r.Header` — заголовки запроса
- `r.Body` — тело запроса
- `r.Form` — данные формы
- `r.Cookies` — cookies

### Более сложный пример: REST API для управления пользователями

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "sync"
    "time"
)

// User представляет модель пользователя
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Age       int       `json:"age"`
    CreatedAt time.Time `json:"created_at"`
}

// Обработчики
func usersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        getUsersHandler(w, r)
    case http.MethodPost:
        createUserHandler(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    // Извлекаем ID из URL
    idStr := r.URL.Path[len("/api/users/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    switch r.Method {
    case http.MethodGet:
        getUserHandler(w, r, id)
    case http.MethodPut:
        updateUserHandler(w, r, id)
    case http.MethodDelete:
        deleteUserHandler(w, r, id)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
		// Get users
    // ...

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func getUserHandler(w http.ResponseWriter, r *http.Request, id int) {
		// Get user
    // ...
    notFound := false

    if notFound {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Валидация
    if user.Name == "" || user.Email == "" {
        http.Error(w, "Name and email are required", http.StatusBadRequest)
        return
    }

    // Create user
    // ...

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request, id int) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    updated := true

    // Update user
    // ...

    if !updated {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request, id int) {
		deleted := true
	  // Delete user
		// ...

    if deleted {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func main() {
    // Регистрируем обработчики
    http.HandleFunc("/api/users", usersHandler)
    http.HandleFunc("/api/users/", userHandler)

    fmt.Println("REST API сервер запущен на :8080")
    fmt.Println("Доступные эндпоинты:")
    fmt.Println("  GET    /api/users       — получить всех пользователей")
    fmt.Println("  GET    /api/users/1     — получить пользователя по ID")
    fmt.Println("  POST   /api/users       — создать пользователя")
    fmt.Println("  PUT    /api/users/1     — обновить пользователя")
    fmt.Println("  DELETE /api/users/1     — удалить пользователя")

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Тестирование нашего REST API

Давайте протестируем наш сервер с помощью `curl`:

```bash
# Получить всех пользователей
curl http://localhost:8080/api/users

# Создать нового пользователя
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com","age":35}'

# Получить пользователя по ID
curl http://localhost:8080/api/users/1

# Обновить пользователя
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice@newmail.com","age":26}'

# Удалить пользователя
curl -X DELETE http://localhost:8080/api/users/2
```

---

## Продвинутые техники HTTP-серверов

Теперь, когда у нас есть базовый REST API, давайте добавим продвинутые функции, которые делают сервисы надёжными и производительными.

### Middleware: цепочки обработки запросов

Middleware (промежуточное ПО) — это функции, которые выполняются до или после основного обработчика. Они позволяют добавить общую логику: логирование, аутентификацию, CORS, rate limiting.

```go
// Тип Middleware
type Middleware func(http.Handler) http.Handler

// Middleware для логирования запросов
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Создаём writer для перехвата статуса
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Выполняем следующий обработчик
        next.ServeHTTP(wrapped, r)

        // Логируем запрос
        duration := time.Since(start)
        log.Printf("%s %s %d %v",
            r.Method, r.URL.Path, wrapped.statusCode, duration)
    })
}

// responseWriter оборачивает http.ResponseWriter для перехвата статуса
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// Middleware для CORS
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Устанавливаем CORS заголовки
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Обрабатываем preflight запросы
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Middleware для аутентификации
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Authorization required", http.StatusUnauthorized)
            return
        }

        // В реальном приложении здесь была бы валидация JWT токена
        if !isValidToken(token) {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Добавляем информацию о пользователе в контекст запроса
        ctx := context.WithValue(r.Context(), "user", getUserFromToken(token))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func isValidToken(token string) bool {
    // Упрощённая проверка токена
    return token == "Bearer valid-token-123"
}

func getUserFromToken(token string) string {
    if token == "Bearer valid-token-123" {
        return "user123"
    }
    return ""
}

// Функция для применения middleware
func applyMiddleware(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Использование middleware
func main() {
    // Создаём маршрутизатор
    router := http.NewServeMux()

    // Регистрируем обработчики
    router.HandleFunc("/api/users", usersHandler)
    router.HandleFunc("/api/users/", userHandler)
    router.HandleFunc("/api/protected", protectedHandler)

    // Применяем middleware ко всем маршрутам
    handler := applyMiddleware(router,
        loggingMiddleware,
        corsMiddleware,
    )

    // Применяем auth middleware только к защищённым маршрутам
    protectedHandler := applyMiddleware(http.HandlerFunc(protectedHandlerFunc), authMiddleware)
    router.Handle("/api/protected", protectedHandler)

    fmt.Println("Сервер с middleware запущен на :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}

func protectedHandlerFunc(w http.ResponseWriter, r *http.Request) {
    user := r.Context().Value("user").(string)
    fmt.Fprintf(w, "Hello, %s! This is protected content.", user)
}
```

### Обработка статических файлов

Go может легко обслуживать статические файлы:

```go
func setupStaticFiles() {
    // Обслуживание файлов из директории
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // Индексная страница
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.ServeFile(w, r, "./static/index.html")
            return
        }
        http.NotFound(w, r)
    })
}

// Структура директории:
// project/
// ├── main.go
// └── static/
//     ├── index.html
//     ├── css/
//     │   └── style.css
//     └── js/
//         └── app.js
```

---

## Работа с данными: JSON, Forms, Files

Веб-сервисы работают с разными типами данных. Давайте научимся правильно обрабатывать JSON, формы и загрузку файлов.

### Продвинутая работа с JSON

Go предоставляет мощные возможности для работы с JSON, включая кастомную сериализацию и валидацию.

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// Продвинутая модель пользователя с валидацией
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Age       int       `json:"age"`
    Password  string    `json:"-"` // Не включаем в JSON
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`

    // Кастомные поля
    IsActive  bool      `json:"is_active"`
    Tags      []string  `json:"tags,omitempty"` // omitempty если пустое
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Обработчик с продвинутой обработкой JSON
func createUserAdvancedHandler(w http.ResponseWriter, r *http.Request) {
    var user User

    // Используем json.Decoder для потоковой обработки больших JSON
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields() // Запрещаем неизвестные поля

    if err := decoder.Decode(&user); err != nil {
        // Обрабатываем разные типы ошибок JSON
        var syntaxErr *json.SyntaxError
        var unmarshalErr *json.UnmarshalTypeError

        switch {
        case err == io.EOF:
            http.Error(w, "Request body must not be empty", http.StatusBadRequest)
        case err.Error() == "json: unknown field \"...\"":
            http.Error(w, "Unknown field in request", http.StatusBadRequest)
        case err.Error() == "http: request body too large":
            http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
        case errors.As(err, &syntaxErr):
            http.Error(w, fmt.Sprintf("Malformed JSON at position %d", syntaxErr.Offset), http.StatusBadRequest)
        case errors.As(err, &unmarshalErr):
            http.Error(w, fmt.Sprintf("Invalid value for field %q at position %d", unmarshalErr.Field, unmarshalErr.Offset), http.StatusBadRequest)
        default:
            http.Error(w, "Error parsing JSON", http.StatusBadRequest)
        }
        return
    }

    // Валидация данных
    if err := validateUser(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Создаём пользователя
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()

    // Отправляем ответ
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func validateUser(user *User) error {
    if user.Name == "" {
        return fmt.Errorf("name is required")
    }
    if len(user.Name) < 2 {
        return fmt.Errorf("name must be at least 2 characters")
    }
    if user.Email == "" {
        return fmt.Errorf("email is required")
    }
    // Здесь может быть более сложная валидация email
    if user.Age < 0 || user.Age > 150 {
        return fmt.Errorf("age must be between 0 and 150")
    }
    return nil
}
```

### Работа с HTML формами

```go
func handleForm(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // Отправляем HTML форму
        formHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>User Registration</title>
</head>
<body>
    <h1>Register New User</h1>
    <form method="POST" action="/register">
        <div>
            <label for="name">Name:</label>
            <input type="text" id="name" name="name" required>
        </div>
        <div>
            <label for="email">Email:</label>
            <input type="email" id="email" name="email" required>
        </div>
        <div>
            <label for="age">Age:</label>
            <input type="number" id="age" name="age" min="1" max="150">
        </div>
        <div>
            <label for="bio">Bio:</label>
            <textarea id="bio" name="bio" rows="4"></textarea>
        </div>
        <div>
            <label>
                <input type="checkbox" name="newsletter" value="yes">
                Subscribe to newsletter
            </label>
        </div>
        <button type="submit">Register</button>
    </form>
</body>
</html>`
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprint(w, formHTML)

    case http.MethodPost:
        // Обрабатываем отправленную форму
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing form", http.StatusBadRequest)
            return
        }

        // Извлекаем данные формы
        name := r.FormValue("name")
        email := r.FormValue("email")
        ageStr := r.FormValue("age")
        bio := r.FormValue("bio")
        newsletter := r.FormValue("newsletter")

        // Конвертируем возраст
        var age int
        if ageStr != "" {
            var err error
            age, err = strconv.Atoi(ageStr)
            if err != nil {
                http.Error(w, "Invalid age", http.StatusBadRequest)
                return
            }
        }

        // Создаём пользователя
        user := User{
            Name:      name,
            Email:     email,
            Age:       age,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
            IsActive:  newsletter == "yes",
        }

        // В реальном приложении здесь было бы сохранение в базу данных

        // Отправляем ответ
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Registration Successful</title>
</head>
<body>
    <h1>Registration Successful!</h1>
    <p>User <strong>%s</strong> has been registered.</p>
    <p><a href="/register">Register another user</a></p>
</body>
</html>`, user.Name)
    }
}
```

### Загрузка файлов

```go
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // Отправляем форму для загрузки файлов
        uploadForm := `
<!DOCTYPE html>
<html>
<head>
    <title>File Upload</title>
</head>
<body>
    <h1>Upload File</h1>
    <form method="POST" action="/upload" enctype="multipart/form-data">
        <div>
            <label for="file">Choose file:</label>
            <input type="file" id="file" name="file" required>
        </div>
        <div>
            <label for="description">Description:</label>
            <input type="text" id="description" name="description">
        </div>
        <button type="submit">Upload</button>
    </form>
</body>
</html>`
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprint(w, uploadForm)

    case http.MethodPost:
        // Обрабатываем загрузку файла
        // Ограничиваем размер файла (10MB)
        r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

        if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
            if err.Error() == "http: request body too large" {
                http.Error(w, "File too large (max 10MB)", http.StatusRequestEntityTooLarge)
            } else {
                http.Error(w, "Error parsing form", http.StatusBadRequest)
            }
            return
        }

        // Получаем файл
        file, handler, err := r.FormFile("file")
        if err != nil {
            http.Error(w, "Error retrieving file", http.StatusBadRequest)
            return
        }
        defer file.Close()

        // Получаем описание
        description := r.FormValue("description")

        // Создаём директорию для загрузок, если её нет
        uploadDir := "./uploads"
        if err := os.MkdirAll(uploadDir, 0755); err != nil {
            http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
            return
        }

        // Создаём уникальное имя файла
        ext := filepath.Ext(handler.Filename)
        filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
        filepath := filepath.Join(uploadDir, filename)

        // Сохраняем файл
        dst, err := os.Create(filepath)
        if err != nil {
            http.Error(w, "Error saving file", http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        if _, err := io.Copy(dst, file); err != nil {
            http.Error(w, "Error saving file", http.StatusInternalServerError)
            return
        }

        // Сохраняем метаданные о файле
        fileInfo := map[string]interface{}{
            "original_name": handler.Filename,
            "saved_name":    filename,
            "size":          handler.Size,
            "content_type":  handler.Header.Get("Content-Type"),
            "description":   description,
            "uploaded_at":   time.Now(),
        }

        // В реальном приложении здесь было бы сохранение в базу данных

        // Отправляем ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "File uploaded successfully",
            "file":    fileInfo,
        })
    }
}
```

---

## Заключение: что мы сегодня узнали

Сегодня мы сделали глубокое погружение в мир HTTP и REST-сервисов на Go. Давайте подведём итоги нашего путешествия.

### Основные концепции

**HTTP протокол:**
- Поняли структуру HTTP запросов и ответов
- Изучили методы HTTP, статусы, заголовки
- Узнали об эволюции протокола от HTTP/1.1 до HTTP/3

**REST архитектура:**
- Освоили принципы REST и ресурсно-ориентированный дизайн
- Научились создавать правильные URI и использовать HTTP методы
