# Конспект лекции 11 — Интерфейсы, композиция и расширение типов

## 1. Интерфейс как контракт возможностей

```csharp
public interface IClock
{
    DateTimeOffset UtcNow { get; }
}

public sealed class SystemClock : IClock
{
    public DateTimeOffset UtcNow => DateTimeOffset.UtcNow;
}
```

Интерфейс сообщает, что умеет объект, не фиксируя его concrete storage и construction. Тип может реализовать несколько interfaces.

## 2. Маленькие role interfaces

```csharp
public interface IOrderReader
{
    Task<Order?> FindAsync(Guid id, CancellationToken cancellationToken);
}

public interface IOrderWriter
{
    Task SaveAsync(Order order, CancellationToken cancellationToken);
}
```

Разделение read/write contract снижает coupling consumers. Но дробить каждый метод в отдельный interface без реальных независимых consumers тоже не нужно.

## 3. Explicit implementation

```csharp
public interface ITextSerializer
{
    string Serialize(object value);
}

public sealed class JsonSerializerAdapter : ITextSerializer
{
    string ITextSerializer.Serialize(object value) =>
        System.Text.Json.JsonSerializer.Serialize(value);
}
```

Explicit member доступен только через interface reference:

```csharp
var adapter = new JsonSerializerAdapter();
ITextSerializer serializer = adapter;
string json = serializer.Serialize(new { Id = 1 });
```

Это помогает устранить collision members или скрыть operation из основного concrete API.

## 4. Default interface members

```csharp
public interface IRetryPolicy
{
    int MaxAttempts => 3;
    TimeSpan GetDelay(int attempt) => TimeSpan.FromSeconds(attempt);
}
```

Default implementation помогает эволюции интерфейса, но не превращает его в удобный base class. State экземпляра интерфейс обычно не хранит, versioning и dispatch имеют тонкости, а consumers могут не увидеть default member через concrete type.

## 5. Static abstract members и generic math

```csharp
public interface IFactory<TSelf>
    where TSelf : IFactory<TSelf>
{
    static abstract TSelf CreateDefault();
}
```

Static abstract interface members позволяют generic algorithm вызывать static operations у type parameter. На этой основе построены generic math interfaces (`INumber<TSelf>`).

## 6. Interface segregation и dependency inversion

Высокоуровневая логика зависит от abstraction:

```csharp
public sealed class ExpiringTokenService(IClock clock, ITokenStore store)
{
    public async Task<bool> IsValidAsync(
        string token,
        CancellationToken cancellationToken)
    {
        Token? found = await store.FindAsync(token, cancellationToken);
        return found is { } && found.ExpiresAt > clock.UtcNow;
    }
}
```

Constructor injection делает dependencies явными. Но интерфейс нужен не «потому что DI», а когда существует boundary, substitution, test seam или несколько реализаций.

## 7. Composition

Decorator реализует тот же contract и оборачивает component:

```csharp
public sealed class LoggingOrderReader(
    IOrderReader inner,
    ILogger<LoggingOrderReader> logger) : IOrderReader
{
    public async Task<Order?> FindAsync(Guid id, CancellationToken ct)
    {
        logger.LogInformation("Reading order {OrderId}", id);
        return await inner.FindAsync(id, ct);
    }
}
```

Decorator явно представляет cross-cutting behavior и легче тестируется, чем скрытая глобальная магия.

## 8. Extension methods и members

```csharp
public static class EnumerableValidationExtensions
{
    public static IEnumerable<T> NotNull<T>(
        this IEnumerable<T?> source) where T : class =>
        source.Where(item => item is not null)!;
}
```

Extensions выбираются compile-time по namespaces и overload resolution. Они:

- не меняют фактический type;
- не имеют доступа к private members;
- уступают настоящему instance member с подходящей сигнатурой;
- могут создавать ambiguity при чрезмерном распространении.

C# 14 extension blocks добавляют extension properties и static extensions, но основное правило остаётся: это static members с instance-like syntax.

## 9. User-defined operators

```csharp
public readonly record struct Money(decimal Amount, string Currency)
{
    public static Money operator +(Money left, Money right)
    {
        if (left.Currency != right.Currency)
            throw new InvalidOperationException("Currencies differ");
        return left with { Amount = left.Amount + right.Amount };
    }
}
```

Operator overload должен сохранять ожидаемую математическую/доменную интуицию. Не используйте `+` для отправки HTTP или mutation database.

C# 14 позволяет user-defined compound assignment operators. Они особенно чувствительны к ожиданиям consumers и не должны вводить отличную от основного operator semantics.

## 10. Conversion operators

```csharp
public readonly record struct UserId(Guid Value)
{
    public static explicit operator Guid(UserId id) => id.Value;
    public static explicit operator UserId(Guid value) => new(value);
}
```

`implicit` безопасен только для очевидного преобразования без потери данных и неожиданных исключений. Domain wrappers обычно лучше преобразовывать явно или через named factory.

## 11. Сравнение с Java

- Оба языка поддерживают multiple interface implementation.
- Default interface methods существуют в обоих, но runtime/versioning details различаются.
- C# поддерживает properties, events, indexers и operators в interface contracts.
- C# static abstract interface members дают compile-time polymorphism для generic math.
- Extension methods/members — встроенный механизм C#; в Java похожий эффект достигается static utilities или defaults.
- Explicit interface implementation не имеет прямого повседневного аналога в Java.

## 12. Антипримеры

- Интерфейс `IUserService` создан только для единственного class и ни одного boundary.
- Один «god interface» содержит десятки несвязанных операций.
- Default methods хранят сложную доменную логику и скрывают dependencies.
- Extension method принимает почти любой `object` и загрязняет autocomplete.
- Implicit conversion делает I/O или может часто бросать исключение.

## Практические выводы

- Проектируйте interfaces от потребностей consumers.
- Используйте composition и decorators для заменяемого поведения.
- Предпочитайте explicit named operations неожиданным operators/conversions.
- Не создавайте abstraction заранее без реальной оси изменения.

## Самопроверка

1. Когда explicit implementation полезна?
2. Видит ли concrete variable default interface member автоматически?
3. Почему extension не является настоящим instance member?
4. Когда interface помогает DI, а когда только добавляет файл?
5. Каким должен быть implicit conversion?

