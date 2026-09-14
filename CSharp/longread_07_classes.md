# Конспект лекции 7 — Классы и объекты

## 1. От разрозненных данных к модели

```csharp
string[] names = ["Ada", "Grace"];
int[] ages = [36, 40];
```

Параллельные массивы связывают данные индексом и легко расходятся. Класс объединяет состояние и поведение:

```csharp
public sealed class User
{
    private string _name;

    public User(Guid id, string name)
    {
        if (id == Guid.Empty)
            throw new ArgumentException("Id must not be empty", nameof(id));

        Id = id;
        _name = NormalizeName(name);
    }

    public Guid Id { get; }
    public string Name => _name;

    public void Rename(string name)
    {
        _name = NormalizeName(name);
    }

    private static string NormalizeName(string value)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(value);
        return value.Trim();
    }
}
```

Объект хранит допустимое состояние и не позволяет обойти правила переименования.

## 2. Объявление и создание

```csharp
User user = new User(Guid.NewGuid(), "Ada");
var second = new User(Guid.NewGuid(), "Grace");
User third = new(Guid.NewGuid(), "Linus");
```

`new` выделяет и инициализирует объект, затем выполняется constructor body. Target-typed `new` использует ожидаемый тип слева.

## 3. Fields

Field — непосредственная часть состояния типа:

```csharp
private int _attempts;
private readonly Guid _id;
private static int _createdCount;
```

- instance field существует у каждого объекта;
- static field принадлежит типу;
- `readonly` можно присвоить в declaration или instance constructor;
- `const` является compile-time constant и не instance state.

Публичные mutable fields почти всегда нарушают инкапсуляцию. Для API используются properties и methods.

## 4. Конструкторы

```csharp
public sealed class ConnectionOptions
{
    public ConnectionOptions(string host)
        : this(host, 5432)
    {
    }

    public ConnectionOptions(string host, int port)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(host);
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(port);

        Host = host;
        Port = port;
    }

    public string Host { get; }
    public int Port { get; }
}
```

Constructor initializer `: this(...)` делегирует другому constructor того же типа. `: base(...)` вызывает constructor базового класса.

Если не объявлен ни один instance constructor, class получает public parameterless constructor только при соответствующей доступности типа. После объявления собственного constructor он автоматически не добавляется.

## 5. Primary constructors

Современный C# поддерживает primary constructor parameters у classes и structs:

```csharp
public sealed class TemperatureSensor(string id)
{
    public string Id { get; } =
        string.IsNullOrWhiteSpace(id)
            ? throw new ArgumentException("Invalid id", nameof(id))
            : id;
}
```

Параметр `id` находится в scope тела типа, но сам по себе не становится property. Не путайте class primary constructor с positional record, который синтезирует members.

## 6. Access modifiers

| Модификатор | Доступ |
|---|---|
| `public` | отовсюду при доступности containing type |
| `private` | только внутри containing type |
| `protected` | внутри типа и наследников |
| `internal` | внутри assembly |
| `protected internal` | либо та же assembly, либо наследник |
| `private protected` | наследник внутри той же assembly |
| `file` | только текущий source file |

Выбирайте минимальную доступность. `internal` полезен для сокрытия implementation details внутри package/assembly.

## 7. Instance и static members

```csharp
public sealed class OrderNumberGenerator
{
    private static long _current;

    public static long Next() =>
        Interlocked.Increment(ref _current);
}
```

Static state разделяется всеми consumers и требует thread-safety. Глобальный mutable static state усложняет тестирование и жизненный цикл приложения.

Static constructor выполняется runtime перед первым активным использованием типа:

```csharp
public static class Registry
{
    public static readonly IReadOnlyDictionary<string, int> Values;

    static Registry()
    {
        Values = new Dictionary<string, int> { ["one"] = 1 };
    }
}
```

Исключение static constructor обычно делает type unusable в текущем process.

## 8. `this`

`this` обозначает текущий instance:

```csharp
public void Rename(string name)
{
    this._name = name;
}
```

Явный `this` нужен при shadowing или передаче текущего объекта. Во многих style guides его опускают, а fields обозначают `_`.

## 9. Object initializers

```csharp
var request = new CreateUserRequest
{
    Name = "Ada",
    Email = "ada@example.test"
};
```

Initializer выполняется после constructor. Он не заменяет validation invariant: object может временно пройти через неполное состояние, если properties mutable. `required` и `init` улучшают contract, но runtime boundaries всё равно требуют проверки.

## 10. Partial types

```csharp
public partial class ApiClient
{
    public void Send() { }
}

public partial class ApiClient
{
    public void Receive() { }
}
```

Части объединяются compiler в один type. Механизм важен для source generators и designer-generated code. Не используйте `partial`, чтобы скрыть чрезмерно большой ручной класс.

C# 14 расширяет partial members, включая constructors и events, что помогает generated/manual parts сотрудничать без reflection.

## 11. Nested types

```csharp
public sealed class Cache
{
    private sealed class Entry
    {
        public required object Value { get; init; }
        public DateTimeOffset ExpiresAt { get; init; }
    }
}
```

Nested type не хранит неявную ссылку на outer object. Это отличается от non-static inner class в Java.

## 12. Object lifetime

Переменная выходит из scope не равнозначно немедленному уничтожению объекта. GC освобождает managed memory, когда объект становится недостижимым и collector решает провести сборку. Внешние ресурсы освобождаются через `IDisposable`, а не ожидание GC.

## 13. Сравнение с Java

- C# file name не обязан совпадать с public class.
- Namespace не обязан отражать каталог.
- В одном файле может быть несколько public types, хотя это редко улучшает навигацию.
- `internal` соответствует границе assembly; точного аналога package-private нет.
- C# nested class не захватывает outer instance автоматически.
- Properties — настоящие language members, а не naming convention getters/setters.
- Primary constructors у обычных classes не синтезируют properties.

## 14. Практические рекомендации

- Создавайте объект сразу валидным.
- Скрывайте mutable fields.
- Не выполняйте I/O в constructor: это усложняет ошибки, async и тесты.
- Используйте factories, если создание имеет несколько исходов или требует async.
- Избегайте service locator и глобального mutable static state.
- Делайте type `sealed`, если наследование не является частью контракта.

## Самопроверка

1. Когда compiler создаёт parameterless constructor?
2. Чем `readonly` field отличается от `const`?
3. Становится ли primary constructor parameter property?
4. Почему static mutable field требует синхронизации?
5. Имеет ли nested class неявный outer reference?

