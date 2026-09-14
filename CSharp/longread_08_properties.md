# Конспект лекции 8 — Properties, indexers и инкапсуляция

## 1. Зачем properties, если есть fields

Property выглядит как доступ к данным, но имеет accessor methods:

```csharp
public sealed class Account
{
    private decimal _balance;

    public decimal Balance => _balance;

    public void Deposit(decimal amount)
    {
        if (amount <= 0)
            throw new ArgumentOutOfRangeException(nameof(amount));

        _balance += amount;
    }
}
```

`Balance` открывает чтение, но mutation проходит через доменную операцию.

## 2. Auto-properties

```csharp
public string Name { get; set; } = "";
public Guid Id { get; init; }
public DateTimeOffset CreatedAt { get; } = DateTimeOffset.UtcNow;
```

Compiler создаёт скрытый backing field. Auto-property подходит, если custom logic не нужна.

## 3. Accessors с разной доступностью

```csharp
public OrderStatus Status { get; private set; }
public string ExternalId { get; internal set; } = "";
```

Accessor accessibility может быть строже property accessibility. `private set` позволяет типу контролировать transitions.

## 4. Computed property

```csharp
public decimal Total => Items.Sum(item => item.Price * item.Quantity);
```

Computed property не должна скрывать дорогой I/O или непредсказуемый side effect. Consumers ожидают, что property относительно дешева и похожа на данные. Долгую операцию оформляйте method и называйте явно.

## 5. Validation в setter

```csharp
private string _name = "";

public string Name
{
    get => _name;
    set
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(value);
        _name = value.Trim();
    }
}
```

В setter входное значение называется `value`.

C# 14 добавляет contextual keyword `field`, позволяющий обратиться к compiler-generated backing field:

```csharp
public string Name
{
    get;
    set
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(value);
        field = value.Trim();
    }
}
```

Если у типа уже есть member с именем `field`, используйте `this.field`, чтобы устранить неоднозначность.

## 6. `init`

```csharp
public sealed class ReportOptions
{
    public string Format { get; init; } = "json";
    public bool IncludeDetails { get; init; }
}

var options = new ReportOptions
{
    Format = "csv",
    IncludeDetails = true
};

// options.Format = "xml"; // после initialization запрещено
```

`init` ограничивает присваивание initialization phase. Это shallow immutability: referenced mutable objects всё ещё могут изменяться.

## 7. `required`

```csharp
public sealed class CreateUserCommand
{
    public required string Name { get; init; }
    public required string Email { get; init; }
}
```

Compiler требует инициализацию required members:

```csharp
var command = new CreateUserCommand
{
    Name = "Ada",
    Email = "ada@example.test"
};
```

`required` не валидирует пустую строку и может быть обойдён reflection/deserialization. Это compile-time construction contract, не полноценная validation.

## 8. Constructor или object initializer

Constructor лучше для небольшого набора обязательных invariants:

```csharp
public Money(decimal amount, string currency) { }
```

Required properties удобны для DTO/configuration с большим набором именованных полей. Не смешивайте способы так, чтобы consumer не понимал, что обязательно.

## 9. Indexers

Indexer позволяет обращаться к объекту через `[]`:

```csharp
public sealed class Headers
{
    private readonly Dictionary<string, string> _values =
        new(StringComparer.OrdinalIgnoreCase);

    public string? this[string name]
    {
        get => _values.GetValueOrDefault(name);
        set
        {
            if (value is null)
                _values.Remove(name);
            else
                _values[name] = value;
        }
    }
}
```

Indexers могут принимать несколько parameters и перегружаться. API должен ясно определять поведение отсутствующего key: исключение, default или `TryGet`.

## 10. `readonly` и immutable object

```csharp
public sealed class Invoice
{
    private readonly List<Line> _lines;

    public Invoice(IEnumerable<Line> lines)
    {
        _lines = [.. lines];
    }

    public IReadOnlyList<Line> Lines => _lines;
}
```

Этот вариант всё ещё может утечь через downcast, если consumer получит исходную `List`. Копирование constructor input защищает от внешней mutation, но `IReadOnlyList<T>` не гарантирует глубокую immutability элементов.

Для сильного контракта рассмотрите `ImmutableArray<T>` или defensive copies на boundary.

## 11. Caching computed property

```csharp
private string? _normalized;

public string Normalized => _normalized ??= ExpensiveNormalize(Name);
```

Lazy cache безопасен только если source state не меняется либо cache invalidated. В concurrent code простое `??=` может вычислить значение несколько раз; допустимость зависит от side effects.

## 12. Interface properties

```csharp
public interface IEntity
{
    Guid Id { get; }
}
```

Интерфейс задаёт access contract, не требует auto-property. Реализация может вычислять значение или хранить его иначе.

## 13. Сравнение с Java

```java
user.getName();
user.setName("Ada");
```

```csharp
string name = user.Name;
user.Name = "Ada";
```

C# property присутствует в metadata как property и распознаётся reflection, serializers, model binding и UI frameworks. Это не просто соглашение имён методов. При этом accessor всё равно является выполняемым кодом и может бросить исключение.

Java records и C# records похожи целью, но C# properties, `init`, `required` и `with` образуют другую object-initialization model.

## 14. Антипримеры

- Setter выполняет сетевой запрос.
- Getter меняет состояние объекта.
- Public setter позволяет нарушить lifecycle transition.
- `required` считается заменой runtime validation.
- Возвращается mutable collection под видом полной immutability.
- Indexer молча возвращает default там, где отсутствие — ошибка.

## 15. Практические выводы

- Property подходит для наблюдаемой характеристики объекта.
- Method подходит для операции, особенно дорогой или меняющей состояние.
- `private set` и domain methods выражают разрешённые transitions.
- `init` и `required` улучшают construction contract, но не валидируют внешние данные.
- Иммутабельность нужно рассматривать по всему object graph.

## Самопроверка

1. Чем auto-property отличается от public field?
2. Гарантирует ли `required string` непустое значение?
3. Почему `init` даёт только shallow immutability?
4. Когда getter не должен выполнять вычисление?
5. Как определить contract отсутствующего key в indexer?

