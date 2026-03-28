# Лекция: Контексты и Graceful Shutdown

## Введение в мир контекстов

Представьте, что вы дирижер в большом оркестре. У вас есть множество музыкантов (горутин), которые играют свои партии. Но что если внезапно нужно остановить выступление? Вы же не будете кричать на каждого музыканта отдельно? У вас есть система жестов и сигналов, которая позволяет координированно управлять всем оркестром.

В Go контексты — это именно такая система дирижирования для наших программ. Они позволяют управлять жизненным циклом горутин, передавать значения и обрабатывать отмены операций элегантным и предсказуемым способом.

### Что такое контекст?

Контекст в Go — это специальный тип данных, который несёт в себе три важные вещи:

1. **Сигналы отмены** — позволяет сообщить горутинам, что нужно остановиться
2. **Дедлайны** — указывает, когда операция должна быть завершена
3. **Значения** — позволяет передавать данные через цепочку вызовов

Контекст как рюкзак путешественника: в нём можно хранить разные вещи (значения), есть карта со сроками (дедлайны), и есть способ срочно позвать всех обратно (отмена).

### Зачем нужны контексты?

Без контекстов наши программы были бы похожи на машину без руля и тормозов — они ехали бы вперёд до тех пор, пока не упрутся в стену или не закончится бензин.

**Основные сценарии использования контекстов:**

- **HTTP-запросы** — пользователь закрыл вкладку, а мы всё ещё выполняем запрос к базе данных
- **База данных** — запрос выполняется слишком долго, нужно его отменить
- **Микросервисы** — один сервис упал, остальные должны перестать отправлять ему запросы
- **Фоновые задачи** — приложение завершается, а фоновые процессы всё ещё работают

---

## Основы работы с контекстами

> [!INFO]
> Интересно, что контексты не всегда были частью стандартной библиотеки Go. Изначально они появились в Google как внутренний пакет `golang.org/x/net/context`, и только в Go 1.7 стали частью стандартной библиотеки.

Давайте начнём с простого и постепенно будем усложнять. Контекст в Go — это интерфейс с небольшим набором методов, но за этим скрывается мощная система управления жизненным циклом программ.

### Базовый интерфейс context.Context

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```

Давайте разберём каждый метод:

- **`Deadline()`** — возвращает время, когда контекст должен быть отменён, и флаг, установлен ли дедлайн
- **`Done()`** — возвращает канал, который закрывается при отмене контекста
- **`Err()`** — возвращает причину отмены контекста
- **`Value()`** — позволяет получить значение из контекста по ключу

### Создание контекстов

Существует несколько способов создания контекстов:

```go
// 1. Пустой контекст (начальная точка)
ctx := context.Background()

// 2. Пустой контекст для неизвестных операций
ctx := context.TODO()

// 3. Контекст с отменой
ctx, cancel := context.WithCancel(context.Background())

// 4. Контекст с таймаутом
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

// 5. Контекст с дедлайном
ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Minute))
```

> [!IMPORTANT]
> Для контекстов, созданных с `WithCancel`, `WithTimeout`, и `WithDeadline`, всегда нужно вызывать функцию `cancel()`! Это похоже на `defer close()` для файлов — освобождаем ресурсы.

### Наш первый пример с контекстами

Давайте создадим простую программу, которая демонстрирует базовую работу с контекстами:

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Worker %d: остановка, причина: %v\n", id, ctx.Err())
            return
        default:
            fmt.Printf("Worker %d: работает...\n", id)
            time.Sleep(500 * time.Millisecond)
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())

    // Запускаем несколько горутин
    for i := 1; i <= 3; i++ {
        go worker(ctx, i)
    }

    // Даём им поработать 2 секунды
    time.Sleep(2 * time.Second)

    // Отменяем контекст
    fmt.Println("main: отправляем сигнал отмены...")
    cancel()

    // Даём время на завершение
    time.Sleep(1 * time.Second)
    fmt.Println("main: программа завершена")
}
```

Запустим эту программу:

```bash
go run main.go
```

Вы увидите, как все воркеры получают сигнал отмены и корректно завершаются. Это как дирижёр, взмахнувший рукой — все музыканты понимают сигнал и прекращают играть.

---

## Хранение данных в контексте

Контексты позволяют передавать данные через цепочку вызовов функций. Это удобно для передачи метаданных запроса, информации об аутентификации и других сценариев.

### Как сохранять и извлекать данные

Для работы с данными в контексте используются две функции:
- `context.WithValue(parent, key, value)` — создает новый контекст с добавленным значением
- `ctx.Value(key)` — извлекает значение из контекста

> [!IMPORTANT]
> Контексты **неизменяемы**! `WithValue` создает новый контекст, а не изменяет существующий.

```go
// Создаем контекст со значением
ctx := context.Background()
ctx = context.WithValue(ctx, "userID", 12345)

// Извлекаем значение
if userID := ctx.Value("userID"); userID != nil {
    fmt.Printf("User ID: %v\n", userID)
}
```

### Какие данные стоит хранить в контексте

В контексте, как бы странно это не звучало, можно сохранять только контекстозависимые данные. Например:
- Метаданные запроса: идентификатор трейса, запроса или пользователя, который выполняет этот запрос
- Информация об аутентификации: идентификатор пользователя, его имя, роли и тд.
- Конфигурация для конкретного запроса: установленный таймаут, количество возможных повторных запросов и пр.
- В отдельных случаях так же удобно пробрасывать объект логгера, который уже обогащен некоторыми данными

Не рекомендуется сохранять в контекст:
- Большие объемы данных
- Части данных бизнес-логики
- Изменяемые данные: счетчики, слайсы и структуры, которые могут изменять во время выполнения запроса
- Любая чувствительная информация, включая конфигурацию запуска вашего приложения

### Важные ограничения

1. **Производительность:** `WithValue` создает копии контекста, что может быть затратно при частых вызовах
2. **Типобезопасность:** `Value()` возвращает `interface{}`, требуется приведение типов
3. **Отладка:** Сложно отследить, откуда взялось значение в контексте
4. **Тестирование:** Мокировать контекст сложнее, чем обычные параметры

### Советы по использованию

1. **Создавайте функции-геттеры** для извлечения значений:
   ```go
   func GetRequestID(ctx context.Context) string {
       if id, ok := ctx.Value(requestIDKey).(string); ok {
           return id
       }
       return ""
   }
   ```

2. **Используйте пустые структуры как ключи** для уникальности
3. **Не передавайте опциональные параметры** через контекст — лучше использовать `*Options`
4. **Документируйте**, какие значения ожидаются в контексте
5. **Проверяйте наличие значений** перед использованием

### Пример трассировки запросов через контекст

В микросервисной архитектуре важно отслеживать, как запрос проходит через разные сервисы. Контексты идеально подходят для передачи Request ID:

```go
// Определяем ключ для Request ID
type contextKey struct{}

const RequestIDKey contextKey = "requestID"

// Middleware для добавления Request ID
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Проверяем, есть ли уже Request ID
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            // Генерируем новый
            requestID = generateUUID()
        }

        // Добавляем в контекст
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

        // Добавляем в заголовок ответа
        w.Header().Set("X-Request-ID", requestID)

        // Передаём управление дальше
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Функция для получения Request ID из контекста
func GetRequestID(ctx context.Context) string {
    if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
        return requestID
    }
    return ""
}

// Использование в обработчике
func handler(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    log.Printf("[%s] Обработка запроса", requestID)

    // Передаём контекст в другие сервисы
    processData(r.Context())
}

func processData(ctx context.Context) {
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Обработка данных", requestID)

    // Здесь может быть вызов других микросервисов с передачей Request ID
    callExternalService(ctx)
}

func callExternalService(ctx context.Context) {
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Вызов внешнего сервиса", requestID)

    // Создаём HTTP-запрос с передачей Request ID
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://external-service/api", nil)
    req.Header.Set("X-Request-ID", requestID)

    // Отправляем запрос...
}
```

---

## Таймауты и дедлайны: когда время имеет значение

В реальном мире у нас всегда есть ограничения по времени. Запрос к базе данных не может выполняться вечно, API-вызов должен завершиться за разумное время, а фоновая задача не должна висеть часами.

### Timeout vs Deadline: в чём разница?

Хотя оба механизма ограничивают время выполнения, между ними есть важное различие:

- **Timeout** — относительное время ("выполнить в течение 5 секунд")
- **Deadline** — абсолютное время ("выполнить до 15:30")

```go
// Timeout: выполниться в течение 3 секунд
ctx, cancel := context.WithTimeout(parent, 3*time.Second)

// Deadline: выполниться до конкретного момента времени
deadline := time.Now().Add(5 * time.Minute)
ctx, cancel := context.WithDeadline(parent, deadline)
```

### Практический пример: HTTP-сервер с таймаутами

Давайте создадим простой HTTP-сервер, который использует контексты для управления запросами:

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "net/http"
    "time"
)

func slowHandler(w http.ResponseWriter, r *http.Request) {
    // Получаем контекст из запроса
    ctx := r.Context()

    fmt.Println("Handler: начал обработку запроса")
    defer fmt.Println("Handler: завершил обработку запроса")

    // Имитируем долгую операцию
    select {
    case <-time.After(time.Duration(rand.Intn(5)) * time.Second):
        fmt.Fprintln(w, "Операция завершена успешно!")
    case <-ctx.Done():
        // Клиент отменил запрос
        fmt.Printf("Handler: запрос отменён, причина: %v\n", ctx.Err())
        http.Error(w, "Request timeout", http.StatusRequestTimeout)
    }
}

func main() {
    http.HandleFunc("/", slowHandler)

    server := &http.Server{
        Addr:         ":8080",
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    fmt.Println("Сервер запущен на :8080")
    if err := server.ListenAndServe(); err != nil {
        panic(err)
    }
}
```

Теперь давайте протестируем этот сервер:

```bash
# В одном терминале запускаем сервер
go run server.go

# В другом терминале отправляем запрос
curl http://localhost:8080
```

Если вы прервёте `curl` (Ctrl+C), вы увидите в логах сервера, что обработчик получил сигнал отмены и корректно завершился.

### Цепочки контекстов: наследование и композиция

Контексты можно комбинировать, создавая сложные иерархии. Это как семейное дерево — у каждого контекста может быть родитель, и наследники получают свойства от предков.

```go
func contextHierarchy() {
    // Корневой контекст
    root := context.Background()

    // Добавляем таймаут
    timeoutCtx, cancel1 := context.WithTimeout(root, 5*time.Second)
    defer cancel1()

    // Добавляем значение
    valueCtx := context.WithValue(timeoutCtx, "userID", 12345)

    // Добавляем ещё один уровень отмены
    cancelCtx, cancel2 := context.WithCancel(valueCtx)
    defer cancel2()

    // Теперь cancelCtx имеет все свойства:
    // - Таймаут от timeoutCtx
    // - Значение userID от valueCtx
    // - Собственную возможность отмены

    workWithContext(cancelCtx)
}
```

### Дочерний контекст без отмены: полезный трюк

Иногда возникает ситуация: у нас есть контекст с таймаутом или отменой, но для какой-то операции нам нужно создать дочерний контекст, который не будет отменяться автоматически. Как это сделать?

**Проблема:** Все стандартные функции (`WithCancel`, `WithTimeout`, `WithDeadline`) создают контексты, которые автоматически отменяются при отмене родителя.

**Решение:** Начиная с Go 1.21, в стандартной библиотеке есть функция `context.WithoutCancel()`!

```go
// Начиная с Go 1.21
import "context"

// Создаем дочерний контекст без отмены
parent, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Дочерний контекст НЕ будет отменён при отмене parent
// Но сохранит все значения из parent
child := context.WithoutCancel(parent)
```

**Как это работает:** `context.WithoutCancel()` создает новый контекст, который:
- Сохраняет все значения из родительского контекста
- Не отменяется автоматически при отмене родителя
- Не имеет дедлайна или таймаута (если только они не были установлены через отдельный вызов)

**Практический пример использования:**

```go
func exampleWithoutCancel() {
    // Родительский контекст с таймаутом
    parent, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Добавляем в родительский контекст Request ID
    parent = context.WithValue(parent, "requestID", "req-12345")

    // Создаем дочерний контекст БЕЗ отмены (стандартная функция)
    // Он сохранит requestID, но не будет отменяться при отмене parent
    childWithoutCancel := context.WithoutCancel(parent)

    // Запускаем две горутины
    go func() {
        // Эта горутина будет отменена при таймауте
        select {
        case <-time.After(5 * time.Second):
            fmt.Println("Долгая операция с отменой завершена")
        case <-parent.Done():
            fmt.Printf("Долгая операция отменена: %v\n", parent.Err())
        }
    }()

    go func() {
        // Эта операция продолжится даже после отмены parent
        // Но сохранит request ID из родительского контекста
        select {
        case <-time.After(5 * time.Second):
            requestID := childWithoutCancel.Value("requestID")
            fmt.Printf("Операция без отмены завершена. Request ID: %v\n", requestID)
        }
    }()

    // Ждём таймаут родительского контекста
    <-parent.Done()
    fmt.Printf("Родительский контекст отменён: %v\n", parent.Err())

    // Даём второй операции время завершиться
    time.Sleep(6 * time.Second)
}
```

> [!IMPORTANT]
> Используйте контексты без отмены осторожно! Они могут привести к утечке горутин, если забыть об их существовании. Основные сценарии использования:
> 1. **Логирование и аудит** — операции должны завершиться независимо от отмены запроса
> 2. **Очистка ресурсов** — фоновая очистка не должна прерываться
> 3. **Метрики и мониторинг** — сбор метрик продолжается даже при отмене основной операции

---

## Graceful Shutdown: элегантное завершение работы

Что произойдет, если ваше приложение получает сигнал SIGTERM (например, от Kubernetes или systemd)? Будет ли просто убито на полуслове, или корректно завершит все операции?

Graceful Shutdown — это способность приложения корректно завершать работу, сохраняя целостность данных и не прерывая критические операции.

### Почему обычный shutdown — это плохо?

Представьте интернет-магазин в момент распродажи, но в этот же момент у нас невыдерживает инфраструктура и kubernetes начинает "убивать" приложение. В этом сценарии:
- Пользователи, которые оформляют заказ, получат ошибку
- Транзакции в базе данных могут остаться незавершёнными
- Кэш не будет сохранён
- Логи могут потеряться

### Как система отправляет сигналы приложению

Операционная система и оркестраторы (Kubernetes, Docker) общаются с нашими приложениями через **сигналы**. Сигналы — это способ уведомить процесс о каком-то событии.

В Go мы можем "слушать" эти сигналы через канал с помощью пакета `os/signal`:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // Создаем канал для приема сигналов
    sigChan := make(chan os.Signal, 1)

    // Указываем, какие сигналы мы хотим получать
    signal.Notify(sigChan,
        syscall.SIGINT,  // Ctrl+C
        syscall.SIGTERM, // Сигнал завершения от системы/Kubernetes
        syscall.SIGHUP,  // Сигнал перезапуска конфигурации
    )

    fmt.Println("Приложение запущено. Ожидаю сигналы...")

    // Запускаем фоновую работу
    go func() {
        for i := 1; i <= 10; i++ {
            fmt.Printf("Выполняю работу %d/10...\n", i)
            time.Sleep(2 * time.Second)
        }
    }()

    // Блокируемся до получения сигнала
    sig := <-sigChan
    log.Printf("Получен сигнал: %v (%v)\n", sig, sig)
}
```

#### Основные сигналы в Linux/macOS:

- **SIGINT (2)** — Сигнал прерывания, обычно посылается при нажатии Ctrl+C
- **SIGTERM (15)** — Сигнал завершения. Это "вежливый" запрос на завершение, который дает приложению время на "уборку"
- **SIGKILL (9)** — Сигнал принудительного завершения. Не может быть перехвачен или проигнорирован. Используется как крайняя мера
- **SIGHUP (1)** — Сигнал "повесить трубку". Исторически использовался для уведомления об обрыве соединения с терминалом. Сегодня часто используется для перезагрузки конфигурации

#### Как это работает в Kubernetes:

В Kubernetes жизненный цикл Pod'а выглядит так:

1. **Pod должен быть удалён** (kubectl delete, deployment update, etc.)
2. **Kubernetes отправляет SIGTERM** всем процессам в контейнере
3. **Ожидание `terminationGracePeriodSeconds`** (по умолчанию 30 секунд)
4. **Если процесс не завершился** → Kubernetes отправляет SIGKILL
5. **Pod удаляется**

### Правильный Graceful Shutdown

```go
// Хороший пример: graceful shutdown
func gracefulServer() {
    server := &http.Server{
        Addr:    ":8080",
        Handler: createRouter(),
    }

    // Запускаем сервер в горутине
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("Ошибка сервера: %v", err)
        }
    }()

    // Ждём сигнала
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig

    log.Println("Получен сигнал завершения, начинаем graceful shutdown...")

    // Создаём контекст с таймаутом для shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Корректно завершаем работу
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Ошибка при shutdown: %v", err)
        server.Close() // вынужденное завершение
    }

    log.Println("Сервер корректно завершён")
}
```

### Тестирование graceful shutdown

Давайте протестируем наше приложение:

```bash
# Запускаем приложение
go run main.go

# В другом терминале отправляем запросы
curl http://localhost:8080/ &
curl http://localhost:8080/ &
curl http://localhost:8080/ &

# Теперь отправляем сигнал завершения
kill -TERM <PID>

# Или используем Ctrl+C для SIGINT
```

Вы увидите, как приложение:
1. Перестаёт принимать новые запросы
2. Дождётся завершения текущих запросов
3. Остановит выполнение фоновых воркеров (при наличии)
4. Корректно закроет соединения с базой данных (при наличии)
5. Завершит работу без потери данных

---

## Лучшие практики и антипаттерны

Давайте обобщим опыт использования контекстов и graceful shutdown в реальных проектах.

### Лучшие практики

#### 1. Всегда передавайте контекст первым параметром

```go
// Хорошо
func GetUser(ctx context.Context, id int) (*User, error) {
    // ...
}

// Плохо
func GetUser(id int, ctx context.Context) (*User, error) {
    // ...
}
```

#### 2. Никогда не храните контекст в структурах

```go
// Плохо
type Service struct {
    ctx context.Context // НЕ ДЕЛАЙТЕ ТАК
}

// Хорошо
type Service struct{}

func (s *Service) GetUser(ctx context.Context, id int) (*User, error) {
    // ...
}
```

#### 3. Используйте типизированные ключи для значений

```go
// Хорошо
type contextKey struct{}

const UserIDKey contextKey = "userID"

func GetUserID(ctx context.Context) int {
    if userID, ok := ctx.Value(UserIDKey).(int); ok {
        return userID
    }
    return 0
}

// Плохо
func GetUserID(ctx context.Context) int {
    if userID, ok := ctx.Value("userID").(int); ok {
        return userID
    }
    return 0
}
```

#### 4. Всегда проверяйте контекст в долгих операциях

```go
func longOperation(ctx context.Context) error {
    // Проверяем перед началом
    if err := ctx.Err(); err != nil {
        return err
    }

    // Проверяем во время операции
    for i := 0; i < 100; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Выполняем шаг операции
            time.Sleep(100 * time.Millisecond)
        }
    }

    return nil
}
```

#### 5. Устанавливайте разумные таймауты

```go
// Хорошо — разные таймауты для разных операций
func processRequest(ctx context.Context) error {
    // Быстрая операция
    quickCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := quickOperation(quickCtx); err != nil {
        return err
    }

    // Долгая операция
    slowCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    return slowOperation(slowCtx)
}
```

### Антипаттерны

#### 1. Использование context.TODO() в production коде

```go
// Плохо
func processData() error {
    ctx := context.TODO() // Почему TODO?
    return database.Query(ctx, query)
}

// Хорошо
func processData(ctx context.Context) error {
    return database.Query(ctx, query)
}
```

#### 2. Игнорирование контекста

```go
// Плохо — игнорируем контекст
func fetchData(ctx context.Context) error {
    // Запускаем операцию без учёта контекста
    go longRunningOperation() // Эта операция не остановится при отмене контекста
    return nil
}

// Хорошо — передаём контекст
func fetchData(ctx context.Context) error {
    go longRunningOperation(ctx) // Операция будет отменена при отмене контекста
    return nil
}
```

#### 3. Слишком длинные цепочки контекстов

```go
// Плохо — слишком много уровней
func process(ctx context.Context) error {
    ctx1, cancel1 := context.WithTimeout(ctx, time.Second)
    defer cancel1()

    ctx2, cancel2 := context.WithTimeout(ctx1, 500*time.Millisecond)
    defer cancel2()

    ctx3, cancel3 := context.WithTimeout(ctx2, 200*time.Millisecond)
    defer cancel3()

    // Зачем так много уровней?
    return operation(ctx3)
}

// Хорошо — один уровень с нужным таймаутом
func process(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
    defer cancel()

    return operation(ctx)
}
```

#### 4. Бесконечные циклы без проверки контекста

```go
// Плохо — невозможно остановить
func monitor(ctx context.Context) {
    for {
        fmt.Println("Monitoring...")
        time.Sleep(time.Second) // Этот цикл никогда не остановится
    }
}

// Хорошо — проверяем контекст
func monitor(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            fmt.Println("Monitoring...")
        case <-ctx.Done():
            fmt.Println("Monitoring stopped")
            return
        }
    }
}
```

---

## Заключение: что мы сегодня узнали

Сегодня мы погрузились в мир контекстов и graceful shutdown — одних из самых важных концепций современной разработки на Go. Давайте подведём итоги:

### Основные концепции

1. **Контексты** — это стандартный способ управления жизненным циклом операций в Go
2. **Graceful Shutdown** — позволяет корректно завершать приложения без потери данных

### Лучшие практики

- Всегда передавайте контекст первым параметром
- Не храните контексты в структурах
- Проверяйте контекст в долгих операциях
- Используйте разумные таймауты
- Реализуйте proper graceful shutdown

### Производительность и надёжность

Правильное использование контекстов позволяет создавать системы, которые:
- Быстро реагируют на отмены и таймауты
- Корректно работают в контейнерованных средах
- Эффективно используют ресурсы
- Обеспечивают целостность данных при перезапусках

Контексты и graceful shutdown — это не просто технические детали, а фундаментальные принципы построения современных распределённых систем. Освоив эти концепции, вы сможете создавать надёжные, масштабируемые и отзывчивые приложения, которые работают надёжно даже в самых сложных условиях production-сред.
