# Лекция: Protocol Buffers — основы форматирования и сериализации данных

## Введение в мир сериализации данных

Когда мы пишем программы, нам часто нужно сохранять данные или передавать их между разными частями системы. Представьте, что у вас есть сложный объект в памяти программы, и вам нужно отправить его по сети или сохранить в файл. Как это сделать? В raw-виде это просто набор байтов, который другой программе будет сложно прочитать.

Здесь на помощь приходит **сериализация** — процесс преобразования структур данных в формат, который можно легко передавать и хранить. А **десериализация** — обратный процесс.

### Почему не просто JSON?

Вы наверняка работали с JSON. Он удобен, человекочитаем и поддерживается везде. Но у него есть недостатки:

- **Избыточность** — каждое поле повторяется в каждом объекте
- **Отсутствие строгой типизации** — легко сделать опечатку в названии поля
- **Большой размер** — текстовый формат занимает больше места
- **Медленная обработка** — нужно парсить текст

Представьте, что вы отправляете 1000 объектов с полями `{"user_id": 123, "username": "alice"}`. В JSON это будет выглядеть так:

```json
[
  {"user_id": 123, "username": "alice"},
  {"user_id": 124, "username": "bob"},
  // ... ещё 998 объектов
]
```

Каждый объект повторяет названия полей! Это как если бы в анкете для каждого человека вы каждый раз писали бы "Имя: " перед именем.

### Protocol Buffers — эффективное решение

Protocol Buffers (ProtoBuf) решают эти проблемы. Это язык описания данных и их бинарный формат сериализации, созданный в Google.

Представьте ProtoBuf как компактную телеграмму вместо длинного письма:

```
// Вместо JSON:
{"user_id": 123, "username": "alice", "email": "alice@example.com"}

// ProtoBuf будет примерно таким (в бинарном виде):
[0x08][0x7B][0x12][0x05][alice][0x1A][0x11][alice@example.com]
```

ProtoBuf работает в два этапа:
1. **Описание схемы** в специальном файле `.proto`
2. **Генерация кода** для разных языков программирования

---

## Protocol Buffers: основы синтаксиса

Давайте подробно разберём синтаксис Protocol Buffers. Это фундамент, на котором строится всё остальное в gRPC.

### Первая структура данных

Начнём с простого примера. Представьте, что мы создаём систему управления пользователями:

```protobuf
// user.proto
syntax = "proto3";

package user;

message UserProfile {
    int64 user_id = 1;
    string username = 2;
    string email = 3;
    bool is_active = 4;
}
```

Разберём по частям:

- `syntax = "proto3";` — указывает версию синтаксиса. Всегда используйте proto3
- `package user;` — пространство имён для предотвращения конфликтов
- `message UserProfile` — определяет структуру данных, похожую на struct в Go
- `int64 user_id = 1;` — поле с номером 1 типа int64

### Что означают номера полей?

Номера полей (1, 2, 3, 4) — это самое важное в ProtoBuf. Именно они обеспечивают совместимость версий:

```protobuf
message UserProfile {
    int64 user_id = 1;        // ← Важно! Нельзя изменять
    string username = 2;      // ← Важно! Нельзя изменять
    string email = 3;         // ← Важно! Нельзя изменять
    bool is_active = 4;       // ← Важно! Нельзя изменять
}
```

Эти номера используются в бинарном формате, а не имена полей! Если вы удалите поле с номером 2 и создадите новое с тем же номером, старые клиенты будут читать новые данные как старое поле — каша!

### Правила эволюции схем

ProtoBuf разработан с учётом того, что данные меняются со временем:

```protobuf
// Версия 1
message UserProfile {
    int64 user_id = 1;
    string username = 2;
    string email = 3;
}

// Версия 2 (совместима с версией 1)
message UserProfile {
    int64 user_id = 1;
    string username = 2;
    string email = 3;
    string phone = 4;        // ← Новое поле, старые клиенты его проигнорируют
    bool is_active = 5;      // ← Новое поле
    int32 deprecated_field = 6 [deprecated = true]; // ← Устаревшее поле
}
```

**Правила совместимости:**
- ✅ Можно добавлять новые поля
- ✅ Можно удалять поля (их номера нельзя использовать повторно)
- ✅ Можно помечать поля устаревшими
- ❌ Нельзя изменять номера существующих полей
- ❌ Нельзя изменять типы полей

### Типы данных в ProtoBuf

ProtoBuf поддерживает множество типов данных:

```protobuf
message DataTypes {
    // Числовые типы
    int32    integer_32 = 1;    // 32-битное целое
    int64    integer_64 = 2;    // 64-битное целое
    uint32   unsigned_32 = 3;   // Беззнаковое 32-битное
    uint64   unsigned_64 = 4;   // Беззнаковое 64-битное
    sint32   signed_32 = 5;     // Знаковое с эффективным кодированием отрицательных
    sint64   signed_64 = 6;     // Знаковое с эффективным кодированием отрицательных
    fixed32  fixed_32 = 7;      // Всегда 4 байта, эффективно для больших значений
    fixed64  fixed_64 = 8;      // Всегда 8 байт, эффективно для больших значений
    sfixed32 signed_fixed_32 = 9;  // Всегда 4 байта, знаковое
    sfixed64 signed_fixed_64 = 10; // Всегда 8 байт, знаковое

    // Вещественные числа
    float    single_float = 11;   // 32-битное с плавающей точкой
    double   double_float = 12;   // 64-битное с плавающей точкой

    // Текстовые типы
    string   text = 13;           // UTF-8 строка
    bytes    binary_data = 14;    // Произвольные байты

    // Булев тип
    bool     flag = 15;           // true/false
}
```

**Когда использовать разные числовые типы:**

- `int32/int64` — для целых чисел, когда важна эффективность кодирования (маленькие числа занимают меньше места)
- `uint32/uint64` — для беззнаковых чисел (не могут быть отрицательными)
- `sint32/sint64` — когда часто встречаются отрицательные числа (эффективнее кодирует их)
- `fixed32/fixed64` — когда числа обычно большие (всегда занимают фиксированное количество байт)
- `sfixed32/sfixed64` — знаковые аналоги с фиксированным размером

### Сложные типы: вложенные сообщения и повторяющиеся поля

```protobuf
message Address {
    string street = 1;
    string city = 2;
    string country = 3;
    string postal_code = 4;
}

message User {
    int64 id = 1;
    string name = 2;
    string email = 3;

    // Вложенное сообщение
    Address address = 4;

    // Массив адресов
    repeated Address addresses = 5;

    // Массив телефонов
    repeated string phone_numbers = 6;

    // Опциональное поле (может отсутствовать)
    optional string nickname = 7;
}
```

Ключевое слово `repeated` создаёт массив (слайс в Go) элементов указанного типа.

### Перечисления (Enums)

Enums позволяют ограничить набор возможных значений:

```protobuf
enum UserStatus {
    USER_STATUS_UNSPECIFIED = 0;    // ← Значение по умолчанию (обязательно!)
    USER_STATUS_ACTIVE = 1;
    USER_STATUS_INACTIVE = 2;
    USER_STATUS_SUSPENDED = 3;
    USER_STATUS_DELETED = 4;
}

message User {
    int64 id = 1;
    string name = 2;
    UserStatus status = 3;          // Используем enum
}
```

> [!IMPORTANT]
> Первое значение enum всегда должно быть 0. Это значение по умолчанию.

### Oneof — выбор одного из нескольких полей

`oneof` позволяет хранить только одно значение из нескольких возможных:

```protobuf
message ContactInfo {
    oneof contact {
        string email = 1;           // Либо email
        string phone = 2;           // Либо phone
        string social_media = 3;    // Либо social media
    }
}
```

`oneof` гарантирует, что только одно поле из группы будет установлено. Экономит память и делает API более понятным.

### Maps — словари

Maps удобны для хранения пар ключ-значение:

```protobuf
message UserPreferences {
    map<string, string> settings = 1;    // Ключ: string, Значение: string
    map<string, int32> scores = 2;       // Ключ: string, Значение: int32
    map<int32, bool> permissions = 3;    // Ключ: int32, Значение: bool
}
```

Maps эквивалентны `repeated` полям с парами ключ-значение, но гораздо удобнее в использовании.

### Коментарии и документация

В .proto файлах можно добавлять комментарии:

```protobuf
// UserProfile представляет основную информацию о пользователе
message UserProfile {
    // Уникальный идентификатор пользователя
    int64 user_id = 1;

    // Имя пользователя для входа в систему
    string username = 2;

    // Email адрес пользователя, используется для восстановления пароля
    string email = 3;

    // Флаг активности пользователя
    // true - пользователь может входить в систему
    // false - аккаунт заблокирован
    bool is_active = 4;
}
```

---

## Продвинутые возможности Protocol Buffers

### Опции и их использование

ProtoBuf поддерживает различные опции для изменения поведения генерации кода:

```protobuf
syntax = "proto3";

package user;

// Опция для указания Go пакета
option go_package = "github.com/yourusername/project/proto";

// Опция для Java
option java_package = "com.example.user";
option java_multiple_files = true;

// Опция для оптимизации
option optimize_for = SPEED; // или CODE_SIZE, LITE_RUNTIME

message User {
    int64 id = 1;
    string name = 2;

    // Поле с опцией устаревания
    string old_field = 3 [deprecated = true];

    // Поле с custom опцией
    string sensitive_data = 4 [(my.custom_option) = "encrypted"];
}
```

### Импорты и переиспользование

ProtoBuf поддерживает импорты из других .proto файлов:

```protobuf
// common.proto
syntax = "proto3";

package common;

message Address {
    string street = 1;
    string city = 2;
    string country = 3;
    string postal_code = 4;
}

message ContactInfo {
    string email = 1;
    string phone = 2;
}
```

```protobuf
// user.proto
syntax = "proto3";

package user;

import "common.proto";
import "google/protobuf/timestamp.proto";

message User {
    int64 id = 1;
    string name = 2;

    // Используем импортированные сообщения
    common.Address address = 3;
    common.ContactInfo contact = 4;

    // Используем стандартные типы
    google.protobuf.Timestamp created_at = 5;
    google.protobuf.Timestamp updated_at = 6;
}
```

### Вложенные сообщения

Сообщения могут быть вложенными друг в друга:

```protobuf
message Order {
    int64 order_id = 1;

    message Item {
        string product_id = 1;
        string name = 2;
        int32 quantity = 3;
        double price = 4;
    }

    repeated Item items = 2;

    message ShippingInfo {
        string address = 1;
        string method = 2;
        double cost = 3;
    }

    ShippingInfo shipping = 3;
}
```

В Go это сгенерируется как вложенные структуры:
```go
type Order struct {
    OrderId int64 `protobuf:"varint,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
    Items   []*Order_Item `protobuf:"bytes,2,rep,name=items,proto3" json:"items,omitempty"`
    Shipping *Order_ShippingInfo `protobuf:"bytes,3,opt,name=shipping,proto3" json:"shipping,omitempty"`
}
```

### Well-known types

Google предоставляет набор стандартных типов:

```protobuf
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";
import "google/protobuf/struct.proto";
import "google/protobuf/any.proto";

message Event {
    string event_id = 1;
    google.protobuf.Timestamp timestamp = 2;
    google.protobuf.Duration duration = 3;
    google.protobuf.Struct metadata = 4;
    google.protobuf.Any payload = 5;
}
```

Основные well-known types:
- `Timestamp` — для времени
- `Duration` — для продолжительности
- `Struct` — для динамических JSON-like структур
- `Any` — для хранения любого типа сообщения
- `Empty` — пустое сообщение

---

## Генерация Go кода из .proto файлов

Теперь самое интересное — как превратить наши .proto определения в работающий Go код.

### Установка необходимых инструментов

Нам понадобится компилятор Protocol Buffers и плагины для Go:

```bash
# 1. Устанавливаем компилятор Protocol Buffers
# Для macOS:
brew install protobuf

# Для Ubuntu/Debian:
sudo apt-get install protobuf-compiler

# Для Windows (через Chocolatey):
choco install protoc

# 2. Проверяем версию
protoc --version
# Должно быть что-то вроде: libprotoc 3.19.4 или новее
```

### Установка Go плагинов

```bash
# Плагины для генерации Go кода
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

# Убедитесь, что $GOPATH/bin добавлен в PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Проверяем установку
protoc-gen-go --version
protoc-gen-go-grpc --version
```

### Создание проекта

Давайте создадим простой проект:

```bash
# Создаём структуру проекта
mkdir proto-example
cd proto-example

# Инициализируем Go модуль
go mod init github.com/yourusername/proto-example

# Создаём директорию для proto файлов
mkdir -p proto
```

### Пример .proto файла

Создайте файл `proto/user.proto`:

```protobuf
syntax = "proto3";

package proto;
option go_package = "github.com/yourusername/proto-example/proto";

// Импортируем для использования timestamp
import "google/protobuf/timestamp.proto";

message User {
    int64 id = 1;
    string username = 2;
    string email = 3;
    string full_name = 4;
    bool is_active = 5;
    google.protobuf.Timestamp created_at = 6;
    google.protobuf.Timestamp updated_at = 7;
}

message CreateUserRequest {
    string username = 1;
    string email = 2;
    string full_name = 3;
}

message CreateUserResponse {
    User user = 1;
    string message = 2;
}

message GetUserRequest {
    int64 user_id = 1;
}

message GetUserResponse {
    User user = 1;
}
```

### Генерация Go кода

Теперь сгенерируем Go код из нашего .proto файла:

```bash
# Запускаем генерацию из корня проекта
protoc --go_out=. --go_opt=paths=source_relative \
    proto/user.proto
```

После выполнения этой команды у вас появится файл `proto/user.pb.go`.

### Понимание опций компилятора

Давайте разберём, что означают параметры команды:

- `--go_out=.` — указывает, что нужно сгенерировать Go код в текущую директорию
- `--go_opt=paths=source_relative` — сохраняет структуру директорий как в исходных файлах
- `proto/user.proto` — путь к .proto файлу

Другие полезные опции:

```bash
# Генерация в отдельную директорию
protoc --go_out=gen/go --go_opt=paths=source_relative proto/user.proto

# Генерация с указанием модуля
protoc --go_out=. --go_opt=paths=source_relative \
    --go_opt=Mproto/user.proto=github.com/yourusername/proto-example/proto \
    proto/user.proto
```

### Структура сгенерированного кода

Давайте посмотрим, что сгенерировалось в `proto/user.pb.go`:

```go
// Код сгенерирован автоматически, не редактируйте!

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        int64                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Username  string                `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Email     string                `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	FullName  string                `protobuf:"bytes,4,opt,name=full_name,json=fullName,proto3" json:"full_name,omitempty"`
	IsActive  bool                  `protobuf:"varint,5,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

// Конструкторы и методы
func (x *User) Reset() {
	*x = User{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_user_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}
func (x *User) ProtoReflect() protoreflect.Message {
	// ...
}

// Getters и Setters
func (x *User) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *User) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

// ... остальные getters
```

### Основные компоненты сгенерированного кода

1. **Структуры данных** — для каждого сообщения генерируется Go struct
2. **Getters/Setters** — безопасный доступ к полям
3. **Методы ProtoMessage/ProtoReflect** — для рефлексии и сериализации
4. **Константы** — для enum значений
5. **Вспомогательные функции** — для работы с сообщениями

---

## Сравнение Protocol Buffers с другими форматами

Давайте сравним Protocol Buffers с другими популярными форматами сериализации.

### Сравнительная таблица

| Характеристика | Protocol Buffers | JSON | XML |
|---------------|------------------|------|-----|
| **Формат** | Бинарный | Текстовый | Текстовый |
| **Человекочитаемость** | Нет | Да | Да |
| **Размер данных** | Компактный | Большой | Очень большой |
| **Скорость** | Очень высокая | Низкая | Очень низкая |
| **Строгая типизация** | Да | Нет | Частично |
| **Схема** | Обязательная | Нет | Необязательная |
| **Эволюция схемы** | Отличная | Нужна ручная поддержка | Сложно |
| **Поддержка языков** | Множество | Все | Многие |

### Производительность на практике

Давайте посмотрим на реальный пример производительности:

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/yourusername/proto-example/proto"
)

// JSON структуры для сравнения
type UserJSON struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsActive bool   `json:"is_active"`
}

func benchmarkProtobuf(user *pb.User) {
	start := time.Now()

	// Сериализация
	data, _ := proto.Marshal(user)

	// Десериализация
	var restored pb.User
	proto.Unmarshal(data, &restored)

	duration := time.Since(start)
	fmt.Printf("ProtoBuf: %v, размер: %d байт\n", duration, len(data))
}

func benchmarkJSON(userJSON UserJSON) {
	start := time.Now()

	// Сериализация
	data, _ := json.Marshal(userJSON)

	// Десериализация
	var restored UserJSON
	json.Unmarshal(data, &restored)

	duration := time.Since(start)
	fmt.Printf("JSON: %v, размер: %d байт\n", duration, len(data))
}

func main() {
	user := &pb.User{
		Id:       12345,
		Username: "alice",
		Email:    "alice@example.com",
		FullName: "Алиса Смирнова",
		IsActive: true,
	}

	userJSON := UserJSON{
		ID:       12345,
		Username: "alice",
		Email:    "alice@example.com",
		FullName: "Алиса Смирнова",
		IsActive: true,
	}

	benchmarkProtobuf(user)
	benchmarkJSON(userJSON)
}
```

Результаты примерно такие:
```
ProtoBuf: 2.5µs, размер: 45 байт
JSON: 15.2µs, размер: 112 байт
```

Protocol Buffers в ~6 раз быстрее и занимает в ~2.5 раза меньше места!

### Когда использовать разные форматы

**Protocol Buffers идеально подходит для:**
- Внутренней коммуникации между микросервисами
- Мобильных приложений (экономия трафика и батареи)
- Высоконагруженных систем
- IoT устройств
- Систем, где важна строгая типизация

**JSON идеально подходит для:**
- Публичных API
- Веб-приложений
- Конфигурационных файлов
- Логирования
- Прототипирования

**XML подходит для:**
- Систем, требующих валидации документов
- Интеграции с legacy системами
- Когда важна человекочитаемость и структура

### Гибридный подход

Часто лучший результат даёт комбинация форматов:

```
┌─────────────────┐    ProtoBuf     ┌─────────────────┐    JSON     ┌─────────────────┐
│   Mobile App    │◄─────────–─────►│   Backend API   │◄────–──────►│   Web Client    │
│                 │                 │                 │             │                 │
│ - Компактный    │                 │ - Быстро        │             │ - Простой       │
│ - Быстрый       │                 │ - Типизирован   │             │ - Совместимый   │
└─────────────────┘                 └─────────────────┘             └─────────────────┘
```

---

## Заключение: что мы сегодня узнали

Сегодня мы глубоко изучили Protocol Buffers — мощную систему сериализации данных, которая лежит в основе современных высокопроизводительных систем.

### Основные концепции

**Protocol Buffers** — это одновременно и язык описания данных (.proto файлы), и бинарный формат сериализации. Мы узнали, почему он эффективнее JSON за счёт компактного бинарного представления и строгой типизации.

**Синтаксис .proto файлов** — освоили все основные конструкции:
   - Базовые типы данных и их правильный выбор
   - Нумерация полей и её важность для совместимости версий
   - Сложные структуры: nested messages, repeated поля, enums, oneof, maps
   - Опции и импорты для переиспользования кода

**Правила эволюции схем** — поняли, как безопасно изменять структуры данных со временем, сохраняя обратную совместимость.
