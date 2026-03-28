# Лекция: Безопасность в Go: лучшие практики и инструменты

## Введение в безопасность Go-приложений

### Почему Go считается безопасным языком

Go был разработан с учётом современных требований к безопасности. Язык включает несколько встроенных механизмов, которые делают его более безопасным по сравнению с некоторыми другими языками программирования.

**Управление памятью** — Go имеет автоматический сборщик мусора (garbage collector), который предотвращает целые классы уязвимостей, связанных с ручным управлением памятью. В отличие от C/C++, где ошибки с указателями могут приводить к buffer overflow, use-after-free и другим проблемам, Go управляет памятью автоматически.

**Строгая типизация** — Go — статически типизированный язык с сильной типизацией. Это предотвращает множество ошибок, которые возможны в динамических языках. Например, невозможно случайно выполнить строку как код или интерпретировать пользовательские данные как исполняемые команды.

**Конкурентность с безопасностью** - Вместо традиционных потоков и разделяемой памяти Go предлагает модель передачи сообщений через каналы. Это снижает вероятность race conditions и других проблем многопоточности, характерных для lock-based подходов.

### Принцип "Security by Default" в Go

Go следует принципу "безопасность по умолчанию" — язык и стандартная библиотека настроены так, чтобы поощрять безопасные практики.

Например, строки в Go неизменяемы (immutable), что предотвращает модификацию данных в непредвиденных местах. Массивы имеют фиксированный размер, а срезы (slices) проверяют границы при доступе, предотвращая выход за пределы массива.

### Области атак

Go используется в различных типах приложений, каждое из которых имеет свои специфические угрозы:

**Веб-приложения** — подвержены XSS, CSRF, SQL-инъекциям, атакам аутентификации и авторизации.

**Сетевые сервисы** — уязвимы для DoS-атак, атак на уровень транспортной безопасности, неправильной обработки сетевых ошибок.

**CLI утилиты** — могут содержать уязвимости при обработке файлов, аргументов командной строки, переменных окружения.

## Безопасная работа с данными

### Безопасная обработка строк и байтовых срезов

Работа со строками и байтовыми данными требует особого внимания, особенно при обработке пользовательского ввода.

```go
// Опасно: небезопасная конкатенация строк с пользовательскими данными
func unsafeConcat userInput string {
    return "SELECT * FROM users WHERE name = '" + userInput + "'"
}

// Безопасно: использование параметризованных запросов
func safeQuery(db *sql.DB, userInput string) {
    query := "SELECT * FROM users WHERE name = ?"
    db.Query(query, userInput)
}
```

При работе с байтовыми срезами важно проверять границы и использовать безопасные функции копирования:

```go
// Безопасное копирование с проверкой границ
func safeCopy(src []byte) []byte {
    dst := make([]byte, len(src))
    copy(dst, src) // copy проверяет размеры автоматически
    return dst
}
```

### Предотвращение race conditions

Race conditions возникают, когда несколько goroutines одновременно обращаются к общим данным без proper синхронизации.

```go
// Опасно: гонка данных
var counter int

func dangerousIncrement() {
    for i := 0; i < 1000; i++ {
        go func() {
            counter++ // Race condition!
        }()
    }
}

// Безопасно: использование sync.Mutex
var safeCounter int
var mutex sync.Mutex

func safeIncrement() {
    for i := 0; i < 1000; i++ {
        go func() {
            mutex.Lock()
            safeCounter++
            mutex.Unlock()
        }()
    }
}
```

Использование каналов для передачи данных между goroutines помогает избежать race conditions:

```go
func channelBasedCounter() {
    counter := make(chan int, 1000)

    // Запуск writer goroutines
    for i := 0; i < 10; i++ {
        go func() {
            for j := 0; j < 100; j++ {
                counter <- 1
            }
        }()
    }

    // Сбор результатов
    total := 0
    for i := 0; i < 1000; i++ {
        total += <-counter
    }
}
```

### Безопасное сравнение данных

При сравнении чувствительных данных (паролей, токенов) необходимо использовать timing-attack resistant функции из пакета `crypto/subtle`:

```go
import "crypto/subtle"

// Опасно: обычное сравнение уязвимо к timing атакам
func unsafeCompare(a, b []byte) bool {
    return string(a) == string(b) // Может возвращать false на разных стадиях
}

// Безопасно: постоянное время сравнения
func secureCompare(a, b []byte) bool {
    return subtle.ConstantTimeCompare(a, b) == 1
}
```

### Очистка чувствительных данных

Чувствительные данные (пароли, ключи шифрования) должны быть удалены из памяти после использования:

```go
import "runtime"

func clearPassword(password []byte) {
    for i := range password {
        password[i] = 0 // Обнуляем байты
    }
    // Дополнительная защита от оптимизаций компилятора
    runtime.KeepAlive(password)
}
```

**Зачем нужен `runtime.KeepAlive`?**

`runtime.KeepAlive` — это специальная функция в Go, которая предотвращает преждевременную сборку мусора для переменной. Без этой функции компилятор Go может оптимизировать код и удалить обнуление пароля из памяти, если определит, что переменная больше не используется.

**Проблема:** Go — язык с garbage collector. Компилятор может определить, что переменная `password` больше не используется после цикла обнуления, и переместить её в память, которая будет очищена сборщиком мусора. Но это произойдёт ПОСЛЕ цикла обнуления, а значит, обнуление может быть оптимизировано как "бесполезное".

**Решение:** `runtime.KeepAlive(password)` говорит компилятору: "Эта переменная всё ещё нужна до этой точки, не удаляй её и не оптимизируй операции с ней".

**Важные моменты при очистке памяти:**

1. **Немедленная очистка** — очищайте секреты сразу после использования
2. **Используйте `bytes` вместо `string`** — строки в Go иммутабельны, их нельзя безопасно очистить
3. **Избегайте логирования секретов** — даже в debug режимах
4. **Используйте специализированные типы** — для критически важных данных рассматривайте типы из `crypto/subtle`

## Криптография в Go

### Правильное использование пакета crypto

Стандартная библиотека Go предоставляет обширные криптографические функции. Важно использовать их правильно и избегать распространённых ошибок.

**Хеширование** — для проверки целостности данных:

```go
import "crypto/sha256"

func hashData(data []byte) []byte {
    h := sha256.New()
    h.Write(data)
    return h.Sum(nil)
}
```

**Генерация случайных данных** — для ключей и токенов:

```go
import "crypto/rand"

func generateRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }
    return b, nil
}
```

### Хеширование паролей

Пароли никогда не должны храниться в открытом виде или с простым хешированием. Используйте специализированные алгоритмы:

```go
import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

func checkPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
```

### Шифрование

Для симметричного шифрования в Go доступны AES и ChaCha20:

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
)

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    // GCM mode provides authenticated encryption
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err = rand.Read(nonce); err != nil {
        return nil, err
    }

    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
```

### Работа с TLS/SSL

Go предоставляет удобный API для работы с TLS:

```go
import (
    "crypto/tls"
    "net/http"
)

func createTLSClient() *http.Client {
    // Валидация сертификата (поведение по умолчанию)
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            // Require at least TLS 1.2
            MinVersion: tls.VersionTLS12,
        },
    }

    return &http.Client{Transport: transport}
}

// Для безопасного TLS сервера
func createTLSServer() *http.Server {
    return &http.Server{
        Addr: ":8443",
        TLSConfig: &tls.Config{
            MinVersion:               tls.VersionTLS12,
            CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
            PreferServerCipherSuites: true,
            CipherSuites: []uint16{
                tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
                tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_RSA_WITH_AES_256_CBC_SHA,
            },
        },
    }
}
```

### Управление ключами и секретами

Никогда не храните ключи и секреты в коде. Используйте специальные решения:

- Переменные окружения
- Секрет-менеджеры (HashiCorp Vault, AWS Secrets Manager, Azure Key Vault)
- Конфигурационные файлы с ограниченным доступом
- Kubernetes Secrets

```go
// Пример безопасной загрузки конфигурации
type Config struct {
    DBPassword string
    APIKey     string
}

func loadConfig() (*Config, error) {
    cfg := &Config{
        DBPassword: os.Getenv("DB_PASSWORD"),
        APIKey:     os.Getenv("API_KEY"),
    }

    if cfg.DBPassword == "" || cfg.APIKey == "" {
        return nil, errors.New("требуемые переменные окружения не установлены")
    }

    return cfg, nil
}
```

## Веб-безопасность

### Защита от SQL-инъекций

SQL-инъекции возникают при прямой подстановке пользовательских данных в SQL-запросы. Всегда используйте параметризованные запросы или prepared statements:

```go
// Опасно: прямая конкатенация
func dangerousQuery(userID string) {
    query := "SELECT * FROM users WHERE id = " + userID
    db.Query(query) // Уязвимо!
}

// Безопасно: параметризованный запрос
func safeQuery(userID string) {
    query := "SELECT * FROM users WHERE id = ?"
    db.Query(query, userID)
}

// Также безопасно: подготовленные выражения
func preparedQuery(userID string) {
    stmt, err := db.Prepare("SELECT * FROM users WHERE id = ?")
    if err != nil {
        return
    }
    defer stmt.Close()

    stmt.QueryRow(userID)
}
```

### XSS и CSRF Prevention

**XSS (Cross-Site Scripting)** — предотвращение через экранирование вывода:

```go
import "html/template"

// Безопасно: автоматическое экранирование
func safeHTML(userInput string) template.HTML {
    // template.HTML отключает экранирование только для доверенного контента
    return template.HTML(userInput) // Используйте с осторожностью!
}

// Правильно: всегда экранировать пользовательский ввод
func escapeHTML(userInput string) string {
    return template.HTMLEscapeString(userInput)
}
```

**CSRF (Cross-Site Request Forgery)** — защита через CSRF токены:

```go
func generateCSRFToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

// Middleware для проверки CSRF токена
func csrfMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" && r.Method != "HEAD" {
            token := r.FormValue("csrf_token")
            if !validateCSRFToken(token) {
                http.Error(w, "Invalid CSRF token", http.StatusForbidden)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

### Безопасная работа с cookies

При работе с cookies важно устанавливать правильные флаги безопасности:

```go
func setSecureCookie(w http.ResponseWriter, name, value string) {
    http.SetCookie(w, &http.Cookie{
        Name:     name,
        Value:    value,
        Path:     "/",
        HttpOnly: true,  // Защита от XSS
        Secure:   true,  // Только через HTTPS
        SameSite: http.SameSiteStrictMode, // Защита от CSRF
        MaxAge:   3600, // Срок действия 1 час
    })
}
```

### Валидация и санитизация входных данных

Всегда валидируйте и санитизируйте пользовательские данные:

```go
import (
    "regexp"
    "strconv"
)

// Валидация email
func validateEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

// Санитизация числовых данных
func sanitizeInt(input string) (int, error) {
    // Удаляем все символы кроме цифр
    numRegex := regexp.MustCompile(`[^0-9]`)
    clean := numRegex.ReplaceAllString(input, "")

    return strconv.Atoi(clean)
}

// Валидация длины данных
func validateLength(input string, min, max int) bool {
    length := len(input)
    return length >= min && length <= max
}
```

### Безопасная загрузка файлов

Загрузка файлов может быть опасна. Всегда проверяйте тип и размер файла:

```go
import (
    "mime/multipart"
    "net/http"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Ограничиваем размер файла (10MB)
    r.ParseMultipartForm(10 << 20)

    file, handler, err := r.FormFile("uploadfile")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Проверяем тип файла
    buffer := make([]byte, 512)
    file.Read(buffer)
    contentType := http.DetectContentType(buffer)

    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "application/pdf": true,
    }

    if !allowedTypes[contentType] {
        http.Error(w, "Invalid file type", http.StatusBadRequest)
        return
    }

    // Сохраняем файл в безопасную директорию
    // ...
}
```

## Аутентификация и авторизация

### JWT токены: правильная реализация

JSON Web Tokens (JWT) — популярный метод аутентификации. JWT представляет собой стандарт RFC 7519 для создания токенов доступа с подписью, которые содержат утверждения (claims) о пользователе.

**Структура JWT:**

JWT состоит из трёх частей, разделённых точками:
```
xxxxx.yyyyy.zzzzz
```
- **Header** (xxxxx) — метаданные о типе токена и алгоритме подписи
- **Payload** (yyyyy) — утверждения (claims) о пользователе
- **Signature** (zzzzz) — цифровая подпись для проверки целостности

**Преимущества JWT:**
- Stateless — серверу не нужно хранить состояние токенов
- Компактность — можно передавать в URL, HTTP header или POST параметрах
- Самодостаточность — содержит всю необходимую информацию о пользователе
- Стандартизация — RFC 7519 обеспечивает совместимость между системами

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "time"
)

type Claims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func generateToken(userID, username, role, secretKey string) (string, error) {
    // Устанавливаем разумное время жизни токена
    expirationTime := time.Now().Add(15 * time.Minute)

    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "your-app",
            Subject:   userID,
            ID:        generateUniqueID(), // Уникальный ID токена
        },
    }

    // Используем сильный алгоритм подписи
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secretKey))
}

func validateToken(tokenString, secretKey string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Проверяем алгоритм подписи
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
        }
        return []byte(secretKey), nil
    })

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, err
}
```

**Лучшие практики безопасности JWT:**

1. **Используйте короткое время жизни** — до 15 минут для access токенов
2. **Implement refresh tokens** — для долгосрочной сессии используйте отдельные refresh токены
3. **Strong secret keys** — минимум 32 символа для HMAC
4. **Validate all claims** — проверяйте issuer, audience, expiration
5. **HTTPS only** — никогда не передавайте JWT через незащищённые соединения
6. **Token blacklisting** — реализуйте механизм отзыва токенов при необходимости

### OAuth2 и OpenID Connect

Go предоставляет готовые решения для OAuth2:

```go
import (
    "golang.org/x/oauth2"
)

// Конфигурация OAuth2
var oauthConfig = &oauth2.Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Scopes:       []string{"openid", "profile", "email"},
    Endpoint: oauth2.Endpoint{
        AuthURL:  "https://provider.com/oauth2/auth",
        TokenURL: "https://provider.com/oauth2/token",
    },
    RedirectURL: "http://localhost:8080/auth/callback",
}

func handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
    // Генерируем случайный state для CSRF защиты
    state := generateRandomString(16)
    setSessionCookie(w, "oauth_state", state)

    url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
```

### RBAC (Role-Based Access Control)

Реализация разграничения доступа на основе ролей:

```go
type Role string

const (
    RoleGuest  Role = "guest"
    RoleUser   Role = "user"
    RoleAdmin  Role = "admin"
)

type Permission string

const (
    PermReadUsers   Permission = "read:users"
    PermWriteUsers  Permission = "write:users"
    PermDeleteUsers Permission = "delete:users"
)

// Роль -> Разрешения
var rolePermissions = map[Role][]Permission{
    RoleGuest: {PermReadUsers},
    RoleUser:  {PermReadUsers},
    RoleAdmin: {PermReadUsers, PermWriteUsers, PermDeleteUsers},
}

func hasPermission(userRole Role, permission Permission) bool {
    permissions, exists := rolePermissions[userRole]
    if !exists {
        return false
    }

    for _, p := range permissions {
        if p == permission {
            return true
        }
    }
    return false
}

// Middleware для проверки разрешений
func requirePermission(permission Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRole := getUserRole(r) // Получаем роль из JWT/сессии

            if !hasPermission(userRole, permission) {
                http.Error(w, "Access denied", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Инструменты безопасности

### Базовые проверки

Go предоставляет встроенные инструменты для базовых проверок кода:

- **go vet** — находит подозрительные конструкции в коде
- **go fmt** — форматирует код единообразно
- **go mod tidy** — удаляет неиспользуемые зависимости

```bash
# Запуск всех проверок
go vet ./...
go fmt ./...
go mod tidy
```

### Статические анализаторы

**gosec** — сканер безопасности для Go кода:

```bash
# Установка
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Запуск сканирования
gosec ./...
```

### Проверка уязвимостей

**govulncheck** — официальный инструмент от Go для проверки уязвимостей:

```bash
# Установка
go install golang.org/x/vuln/cmd/govulncheck@latest

# Проверка всего проекта
govulncheck ./...

# Проверка конкретного модуля
govulncheck github.com/example/project
```

Результаты govulncheck показывают:
- Обнаруженные уязвимости
- Версии, в которых они исправлены
- Путь к уязвимому коду в вашем проекте

## Заключение

### Культура безопасности в команде

Безопасность — это не разовая задача, а постоянный процесс. Создание культуры безопасности в команде включает:

1. **Регулярное обучение** — команда должна быть в курсе последних угроз и методов защиты
2. **Code Review** — каждый PR должен проходить проверку на безопасность
3. **Автоматизация** — инструменты безопасности должны быть встроены в CI/CD
4. **Инцидент-менеджмент** — план действий при обнаружении уязвимости

### Чек-лист безопасности для Go-проекта

- [ ] Включены `go vet` и `staticcheck` в CI/CD
- [ ] Регулярно запускается `govulncheck`
- [ ] Все пользовательские данные валидируются и санитизируются
- [ ] Пароли хешируются bcrypt/scrypt/Argon2
- [ ] Используется TLS 1.2+ для сетевых соединений
- [ ] Реализована защита от CSRF/XSS
- [ ] SQL-запросы используют параметризацию
- [ ] Файловая загрузка проверяет тип и размер
- [ ] Секреты хранятся в безопасном месте (не в коде)
- [ ] Логирование не раскрывает чувствительные данные
- [ ] Настроены rate limiting для API эндпоинтов
- [ ] Проведен пентест критически важных компонентов

Безопасность в Go требует постоянного внимания и обновления знаний. Используйте эти практики и инструменты как основу для создания надежных и безопасных приложений.
