# Лекция: gRPC: реализация сервисов на Go

## Введение в мир gRPC

В предыдущей лекции мы изучили Protocol Buffers — язык описания данных и формат сериализации. Теперь пришло время увидеть, как эти данные "оживают" в gRPC — современном фреймворке для создания высокопроизводительных RPC (Remote Procedure Call) сервисов.

### Что такое gRPC?

Представьте, что у вас есть два приложения. Одно работает на вашем ноутбуке, другое — где-то в облаке. Как им обмениваться информацией?

Традиционный подход с REST API:
```go
// REST подход
func getUserFromRemote(userID int64) (*User, error) {
    url := fmt.Sprintf("http://api.example.com/users/%d", userID)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user User
    err = json.NewDecoder(resp.Body).Decode(&user)
    return &user, err
}
```

gRPC подход:
```go
// gRPC подход
func getUserFromRemote(ctx context.Context, client UserServiceClient, userID int64) (*User, error) {
    req := &GetUserRequest{UserId: userID}
    resp, err := client.GetUser(ctx, req)
    return resp.User, err
}
```

Разница очевидна — gRPC выглядит как вызов обычной функции, а не как работа с HTTP!

### Почему gRPC так хорош?

1. **Производительность** — в 5-10 раз быстрее REST/JSON за счёт бинарного формата Protocol Buffers
2. **Строгая типизация** — компилятор проверяет всё на этапе сборки
3. **Автоматическая генерация кода** — меньше рутины, больше времени на логику
4. **Множество типов коммуникации** — не только запрос-ответ, но и стриминг
5. **Поддержка множества языков** — Go сервер может общаться с Python клиентом

---

## Архитектура gRPC

### Базовая архитектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   gRPC Client   │───▶│   gRPC Server    │───▶│   Business      │
│                 │    │                  │    │   Logic         │
│ - Generates     │    │ - Receives       │    │                 │
│   requests      │    │   requests       │    │ - Database      │
│ - Handles       │    │ - Validates      │    │ - External      │
│   responses     │    │ - Calls business │    │   services      │
│                 │    │   logic          │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Компоненты gRPC

1. **.proto файлы** — определяют контракты между сервисами
2. **Кодогенератор** — создаёт код для клиента и сервера
3. **gRPC Server** — принимает запросы и вызывает бизнес-логику
4. **gRPC Client** — отправляет запросы и обрабатывает ответы
5. **Transport Layer** — HTTP/2 для эффективной передачи данных

---

## Типы коммуникации в gRPC

gRPC поддерживает 4 типа коммуникации. Давайте рассмотрим каждый из них.

### 1. Unary Call (Простой вызов)

Самый распространённый тип — один запрос, один ответ.

**Пример в .proto файле:**
```protobuf
service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

message GetUserRequest {
    int64 user_id = 1;
}

message GetUserResponse {
    User user = 1;
}
```

На сервере unary вызов реализуется очень просто. Мы получаем запрос, обрабатываем его и возвращаем ответ:

**Реализация на сервере:**
```go
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Получаем пользователя из базы данных
    user, err := s.db.GetUser(ctx, req.UserId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "пользователь не найден")
    }

    return &pb.GetUserResponse{User: user}, nil
}
```

**Вызов на клиенте:**
```go
func (c *UserServiceClient) GetUser(userID int64) (*pb.User, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{UserId: userID})
    if err != nil {
        return nil, err
    }

    return resp.User, nil
}
```

### 2. Server Streaming (Стриминг с сервера)

Клиент отправляет один запрос, сервер отвечает потоком данных.

**Пример в .proto файле:**
```protobuf
service UserService {
    rpc StreamUsers(StreamUsersRequest) returns (stream User);
}

message StreamUsersRequest {
    int32 limit = 1;
}
```

Server streaming отличается тем, что мы отправляем данные порциями с помощью метода `stream.Send()`:

**Реализация на сервере:**
```go
func (s *UserServiceServer) StreamUsers(req *pb.StreamUsersRequest, stream pb.UserService_StreamUsersServer) error {
    // Получаем всех пользователей из базы данных
    users, err := s.db.GetAllUsers()
    if err != nil {
        return status.Error(codes.Internal, "ошибка получения пользователей")
    }

    // Отправляем пользователей по одному
    for i, user := range users {
        if i >= int(req.Limit) {
            break
        }

        if err := stream.Send(user); err != nil {
            return status.Error(codes.Internal, "ошибка отправки пользователя")
        }
    }

    return nil
}
```

**Использование на клиенте:**
```go
func (c *UserServiceClient) StreamUsers(limit int32) ([]*pb.User, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    stream, err := c.client.StreamUsers(ctx, &pb.StreamUsersRequest{Limit: limit})
    if err != nil {
        return nil, err
    }

    var users []*pb.User
    for {
        user, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, nil
}
```

### 3. Client Streaming (Стриминг с клиента)

Клиент отправляет поток данных, сервер отвечает одним сообщением.

**Пример в .proto файле:**
```protobuf
service UserService {
    rpc BulkCreateUsers(stream CreateUserRequest) returns (BulkCreateUsersResponse);
}

message BulkCreateUsersResponse {
    int32 created_count = 1;
    repeated string errors = 2;
}
```

Client streaming работает по-другому — мы получаем поток данных от клиента, обрабатываем каждый элемент и в конце отправляем один ответ:

**Реализация на сервере:**
```go
func (s *UserServiceServer) BulkCreateUsers(stream pb.UserService_BulkCreateUsersServer) error {
    var createdCount int32
    var errors []string

    for {
        req, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        // Создаём пользователя
        if err := s.db.CreateUser(req.User); err != nil {
            errors = append(errors, fmt.Sprintf("ошибка создания %s: %v", req.User.Username, err))
        } else {
            createdCount++
        }
    }

    return stream.SendAndClose(&pb.BulkCreateUsersResponse{
        CreatedCount: createdCount,
        Errors:       errors,
    })
}
```

**Использование на клиенте:**
```go
func (c *UserServiceClient) BulkCreateUsers(users []*pb.User) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    stream, err := c.client.BulkCreateUsers(ctx)
    if err != nil {
        return err
    }

    // Отправляем пользователей
    for _, user := range users {
        if err := stream.Send(&pb.CreateUserRequest{User: user}); err != nil {
            return err
        }
    }

    // Получаем результат
    resp, err := stream.CloseAndRecv()
    if err != nil {
        return err
    }

    fmt.Printf("Создано: %d, ошибок: %d\n", resp.CreatedCount, len(resp.Errors))
    return nil
}
```

### 4. Bidirectional Streaming (Двунаправленный стриминг)

И клиент, и сервер могут отправлять потоки данных одновременно.

**Пример в .proto файле:**
```protobuf
service ChatService {
    rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}

message ChatMessage {
    string user = 1;
    string text = 2;
    int64 timestamp = 3;
}
```

Bidirectional streaming — самый сложный тип. Мы обрабатываем входящие сообщения и одновременно можем отправлять ответы в любой момент:

**Реализация на сервере:**
```go
func (s *ChatServiceServer) Chat(stream pb.ChatService_ChatServer) error {
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }

        // Обрабатываем сообщение
        response := &pb.ChatMessage{
            User:      "server",
            Text:      fmt.Sprintf("Получил сообщение от %s: %s", req.User, req.Text),
            Timestamp: time.Now().Unix(),
        }

        // Отправляем ответ
        if err := stream.Send(response); err != nil {
            return err
        }
    }
}
```

**Использование на клиенте:**
```go
func (c *ChatServiceClient) StartChat() error {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    stream, err := c.client.Chat(ctx)
    if err != nil {
        return err
    }

    // Запускаем приём сообщений в отдельной goroutine
    go func() {
        for {
            msg, err := stream.Recv()
            if err == io.EOF {
                return
            }
            if err != nil {
                log.Printf("Ошибка получения сообщения: %v", err)
                return
            }
            fmt.Printf("%s: %s\n", msg.User, msg.Text)
        }
    }()

    // Отправляем сообщения
    messages := []string{"Привет!", "Как дела?", "Пока!"}
    for _, text := range messages {
        if err := stream.Send(&pb.ChatMessage{
            User:      "client",
            Text:      text,
            Timestamp: time.Now().Unix(),
        }); err != nil {
            return err
        }
        time.Sleep(time.Second)
    }

    return stream.CloseSend()
}
```

---

## Настройка gRPC Сервера

### Базовая настройка сервера

Давайте создадим простой gRPC сервер. Минимальная настройка включает создание TCP listener, настройку gRPC сервера и регистрацию нашего сервиса.

```go
package main

import (
    "log"
    "net"

    pb "github.com/yourusername/project/proto"
    "google.golang.org/grpc"
)

type UserServiceServer struct {
    pb.UnimplementedUserServiceServer
}

func main() {
    // Создаём TCP listener
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Не удалось начать прослушивание: %v", err)
    }

    // Создаём gRPC сервер
    grpcServer := grpc.NewServer()

    // Регистрируем наш сервис
    pb.RegisterUserServiceServer(grpcServer, &UserServiceServer{})

    log.Println("gRPC сервер запущен на порту 50051")

    // Запускаем сервер
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Не удалось запустить сервер: %v", err)
    }
}
```

### Опции сервера

gRPC предоставляет множество опций для настройки сервера. Давайте рассмотрим самые полезные из них:

```go
// Создание сервера с опциями
grpcServer := grpc.NewServer(
    // Максимальный размер сообщения (4MB)
    grpc.MaxRecvMsgSize(4*1024*1024),
    grpc.MaxSendMsgSize(4*1024*1024),

    // Количество соединений
    grpc.MaxConcurrentStreams(1000),

    // Таймауты
    grpc.ConnectionTimeout(30*time.Second),

    // Интерцепторы (рассмотрим ниже)
    grpc.UnaryInterceptor(loggingInterceptor),
)
```

### Graceful Shutdown

```go
func main() {
    // ... настройка сервера ...

    // Запускаем сервер в отдельной goroutine
    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            log.Printf("Сервер остановлен: %v", err)
        }
    }()

    // Ожидаем сигнала завершения
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Получен сигнал завершения")

    // Graceful shutdown
    done := make(chan struct{})
    go func() {
        grpcServer.GracefulStop()
        close(done)
    }()

    select {
    case <-done:
        log.Println("Сервер корректно остановлен")
    case <-time.After(30 * time.Second):
        log.Println("Таймаут, принудительная остановка")
        grpcServer.Stop()
    }
}
```

---

## Настройка gRPC Клиента

### Базовое подключение

Создание gRPC клиента также достаточно простое. Нам нужно установить соединение с сервером и создать клиент:

```go
package main

import (
    "time"

    pb "github.com/yourusername/project/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type UserServiceClient struct {
    conn   *grpc.ClientConn
    client pb.UserServiceClient
}

func NewUserServiceClient(serverAddr string) (*UserServiceClient, error) {
    // Создаём подключение
    conn, err := grpc.Dial(serverAddr,
        // Используем незащищённое соединение (для разработки)
        grpc.WithTransportCredentials(insecure.NewCredentials()),

        // Таймаут подключения
        grpc.WithBlock(),
        grpc.WithTimeout(5*time.Second),
    )
    if err != nil {
        return nil, err
    }

    return &UserServiceClient{
        conn:   conn,
        client: pb.NewUserServiceClient(conn),
    }, nil
}

func (c *UserServiceClient) Close() error {
    return c.conn.Close()
}
```

### Опции клиента

gRPC предоставляет множество опций для тонкой настройки клиентских соединений. Давайте рассмотрим самые важные из них:

```go
// Расширенные опции подключения
conn, err := grpc.Dial(serverAddr,
    // Тип подключения
    grpc.WithTransportCredentials(creds), // TLS или insecure

    // Балансировка нагрузки
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),

    // Keepalive
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second, // ping через каждые 10 сек
        Timeout:             3 * time.Second, // таймаут ответа
        PermitWithoutStream: true,           // разрешать keepalive без активных stream'ов
    }),

    // Интерцепторы
    grpc.WithUnaryInterceptor(clientInterceptor),

    // Размер сообщений
    grpc.WithDefaultCallOptions(
        grpc.MaxCallRecvMsgSize(4*1024*1024),
        grpc.MaxCallSendMsgSize(4*1024*1024),
    ),
)
```

### Работа с Metadata

Metadata — это аналог HTTP заголовков в gRPC. Они позволяют передавать дополнительную информацию между клиентом и сервером: токены авторизации, трассировку, версии API и другие метаданные.

#### Работа с metadata на сервере

На сервере мы можем получать metadata из входящего контекста и отправлять их клиенту:

```go
// Получение metadata из входящего запроса
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Получаем metadata из контекста
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        log.Println("Metadata не найдены в контексте")
    } else {
        // Читаем конкретные поля metadata
        if authTokens := md["authorization"]; len(authTokens) > 0 {
            token := authTokens[0]
            log.Printf("Получен токен авторизации: %s", maskToken(token))

            // Валидация токена
            if !validateToken(token) {
                return nil, status.Error(codes.Unauthenticated, "невалидный токен")
            }
        }

        // Читаем трейсинг информацию
        if traceIDs := md["x-trace-id"]; len(traceIDs) > 0 {
            log.Printf("Trace ID: %s", traceIDs[0])
        }

        // Читаем информацию о клиенте
        if userAgents := md["user-agent"]; len(userAgents) > 0 {
            log.Printf("User-Agent: %s", userAgents[0])
        }
    }

    // Добавляем metadata в ответ
    header := metadata.New(map[string]string{
        "server-version": "1.0.0",
        "request-id":     generateRequestID(),
        "processing-time": time.Now().Format(time.RFC3339),
    })

    // Устанавливаем заголовки перед отправкой ответа
    if err := grpc.SetHeader(ctx, header); err != nil {
        log.Printf("Ошибка установки заголовков: %v", err)
    }

    // Отправляем trailer после завершения обработки
    trailer := metadata.New(map[string]string{
        "status": "success",
        "debug":  "processing completed",
    })

    if err := grpc.SetTrailer(ctx, trailer); err != nil {
        log.Printf("Ошибка установки trailer: %v", err)
    }

    // ... основная логика обработки запроса ...
}

// Функция маскировки токена для логов
func maskToken(token string) string {
    if len(token) <= 8 {
        return "***"
    }
    return token[:4] + "***" + token[len(token)-4:]
}

// Функция генерации ID запроса
func generateRequestID() string {
    return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
```

#### Работа с metadata на клиенте

На клиенте мы добавляем metadata в исходящие запросы:

```go
func (c *UserServiceClient) GetUserWithMetadata(userID int64) (*pb.User, error) {
    // Создаём контекст с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Добавляем metadata в контекст
    ctx = metadata.AppendToOutgoingContext(ctx,
        "authorization", "Bearer "+c.token,
        "x-trace-id", generateTraceID(),
        "client-version", "2.1.0",
        "request-source", "mobile-app",
    )

    // Выполняем вызов с metadata
    resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{UserId: userID})
    if err != nil {
        return nil, err
    }

    return resp.User, nil
}

// Создание клиента с автоматическим добавлением metadata
func NewUserServiceClientWithMetadata(serverAddr, token string) (*UserServiceClient, error) {
    // Создаём подключение
    conn, err := grpc.Dial(serverAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithUnaryInterceptor(
            // Интерцептор для автоматического добавления metadata
            func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
                // Добавляем metadata в каждый запрос
                ctx = metadata.AppendToOutgoingContext(ctx,
                    "authorization", "Bearer "+token,
                    "client-id", getClientID(),
                    "timestamp", time.Now().Format(time.RFC3339),
                )

                return invoker(ctx, method, req, reply, cc, opts...)
            },
        ),
    )
    if err != nil {
        return nil, err
    }

    return &UserServiceClient{
        conn:   conn,
        client: pb.NewUserServiceClient(conn),
        token:  token,
    }, nil
}

// Получение metadata из ответа сервера
func (c *UserServiceClient) GetUserWithHeaders(userID int64) (*pb.User, metadata.MD, metadata.MD, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Переменные для получения header и trailer
    var header, trailer metadata.MD

    // Используем CallOptions для получения metadata
    resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{UserId: userID},
        grpc.Header(&header),    // Получаем header из ответа
        grpc.Trailer(&trailer),  // Получаем trailer из ответа
    )
    if err != nil {
        return nil, nil, nil, err
    }

    // Обрабатываем полученные header
    if serverVersion := header["server-version"]; len(serverVersion) > 0 {
        log.Printf("Версия сервера: %s", serverVersion[0])
    }

    if requestID := header["request-id"]; len(requestID) > 0 {
        log.Printf("ID запроса: %s", requestID[0])
    }

    // Обрабатываем trailer
    if status := trailer["status"]; len(status) > 0 {
        log.Printf("Статус обработки: %s", status[0])
    }

    return resp.User, header, trailer, nil
}
```

---

## Интерцепторы (Middleware)

Интерцепторы — это middleware для gRPC. Они позволяют выполнять код до/после вызова методов.

### Типы интерцепторов

1. **Unary Interceptors** — для обычных вызовов (запрос-ответ)
2. **Stream Interceptors** — для потоковых вызовов

### Unary Interceptor на сервере

Unary интерцепторы получают контекст, запрос, информацию о методе и обработчик. Давайте создадим несколько полезных интерцепторов:

```go
// Логирующий интерцептор
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    start := time.Now()

    log.Printf("Начало вызова: %s", info.FullMethod)

    // Вызываем оригинальный обработчик
    resp, err := handler(ctx, req)

    duration := time.Since(start)

    if err != nil {
        log.Printf("Ошибка вызова %s за %v: %v", info.FullMethod, duration, err)
    } else {
        log.Printf("Успешный вызов %s за %v", info.FullMethod, duration)
    }

    return resp, err
}

// Интерцептор аутентификации
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Пропускаем некоторые методы без аутентификации
    if isPublicMethod(info.FullMethod) {
        return handler(ctx, req)
    }

    // Проверяем токен
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "метаданные не найдены")
    }

    token := md["authorization"]
    if len(token) == 0 || !validateToken(token[0]) {
        return nil, status.Error(codes.Unauthenticated, "невалидный токен")
    }

    return handler(ctx, req)
}
```

### Stream Interceptor на сервере

```go
func StreamLoggingInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
    log.Printf("Начало stream вызова: %s", info.FullMethod)

    err := handler(srv, ss)

    if err != nil {
        log.Printf("Ошибка stream вызова %s: %v", info.FullMethod, err)
    } else {
        log.Printf("Успешный stream вызов %s", info.FullMethod)
    }

    return err
}
```

### Интерцепторы на клиенте

```go
// Клиентский интерцептор для добавления токена
func ClientAuthInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    // Добавляем токен в метаданные
    ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+getToken())

    return invoker(ctx, method, req, reply, cc, opts...)
}

// Интерцептор для retry
func RetryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    var lastErr error

    for attempt := 0; attempt < 3; attempt++ {
        if attempt > 0 {
            time.Sleep(time.Duration(attempt) * time.Second)
        }

        err := invoker(ctx, method, req, reply, cc, opts...)
        if err == nil {
            return nil
        }

        // Проверяем, можно ли повторить запрос
        if isRetryableError(err) {
            lastErr = err
            continue
        }

        return err
    }

    return lastErr
}
```

### Применение интерцепторов

```go
// На сервере
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(
        grpc.ChainUnaryInterceptor(
            LoggingInterceptor,
            AuthInterceptor,
            MetricsInterceptor,
        ),
    ),
    grpc.StreamInterceptor(
        grpc.ChainStreamInterceptor(
            StreamLoggingInterceptor,
            StreamAuthInterceptor,
        ),
    ),
)

// На клиенте
conn, err := grpc.Dial(serverAddr,
    grpc.WithUnaryInterceptor(
        grpc.ChainUnaryInterceptor(
            ClientAuthInterceptor,
            RetryInterceptor,
        ),
    ),
)
```

---

## Обработка ошибок в gRPC

gRPC использует стандартные status codes для обработки ошибок:

### Основные status codes

```go
// Часто используемые коды ошибок
codes.OK              // Успех
codes.Canceled        // Операция отменена
codes.Unknown         // Неизвестная ошибка
codes.InvalidArgument // Неверные аргументы
codes.DeadlineExceeded // Превышен таймаут
codes.NotFound        // Ресурс не найден
codes.AlreadyExists   // Ресурс уже существует
codes.PermissionDenied // Нет прав доступа
codes.Unauthenticated // Не аутентифицирован
codes.ResourceExhausted // Превышен лимит ресурсов
codes.FailedPrecondition // Предусловие не выполнено
codes.Aborted         // Операция прервана
codes.OutOfRange      // Выход за пределы диапазона
codes.Unimplemented   // Метод не реализован
codes.Internal        // Внутренняя ошибка
codes.Unavailable     // Сервис недоступен
codes.DataLoss        // Потеря данных
```

### Создание ошибок

```go
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    if req.UserId <= 0 {
        return nil, status.Error(codes.InvalidArgument, "ID должен быть положительным")
    }

    user, err := s.db.GetUser(ctx, req.UserId)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, status.Error(codes.NotFound, "пользователь не найден")
        }
        return nil, status.Error(codes.Internal, fmt.Sprintf("ошибка базы данных: %v", err))
    }

    return &pb.GetUserResponse{User: user}, nil
}

// Продвинутые ошибки с деталями
func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    if err := validateUser(req); err != nil {
        st, _ := status.New(codes.InvalidArgument, "ошибка валидации").WithDetails(
            &errdetails.BadRequest{
                FieldViolations: []*errdetails.BadRequest_FieldViolation{
                    {
                        Field:       "username",
                        Description: "имя пользователя должно содержать минимум 3 символа",
                    },
                },
            },
        )
        return nil, st.Err()
    }

    // ... создание пользователя ...
}
```

### Обработка ошибок на клиенте

```go
func (c *UserServiceClient) GetUser(userID int64) (*pb.User, error) {
    resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{UserId: userID})
    if err != nil {
        st, ok := status.FromError(err)
        if !ok {
            return nil, fmt.Errorf("неизвестная ошибка: %v", err)
        }

        switch st.Code() {
        case codes.NotFound:
            log.Printf("Пользователь не найден: %s", st.Message())
            return nil, st.Err()
        case codes.DeadlineExceeded:
            log.Printf("Таймаут запроса")
            return nil, st.Err()
        case codes.Unavailable:
            log.Printf("Сервис недоступен, можно повторить позже")
            return nil, st.Err()
        default:
            log.Printf("Ошибка gRPC: %s", st.Message())
            return nil, st.Err()
        }
    }

    return resp.User, nil
}
```

---

## Заключение: что мы сегодня узнали

Сегодня мы изучили основы реализации gRPC сервисов на Go. Давайте подведём итоги.

### Основные концепции

1. **Архитектура gRPC** — поняли, как работают клиент-серверные взаимодействия в gRPC
2. **Типы коммуникации** — изучили все 4 типа: unary, server streaming, client streaming, bidirectional streaming
3. **Настройка сервера и клиента** — научились настраивать gRPC соединения с различными опциями

### Практические навыки

4. **Реализация разных типов вызовов** — создали примеры для всех типов коммуникации
5. **Интерцепторы** — освоили middleware для логирования, аутентификации, метрик и rate limiting
6. **Обработка ошибок** — научились работать с gRPC status codes и детализацией ошибок
