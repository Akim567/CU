# Конспект лекции 12 — Исключения и управление ресурсами

## Введение

Exception сообщает, что операция не смогла выполнить контракт нормальным способом. Но managed memory и внешние ресурсы имеют разные жизненные циклы: GC освобождает память, а файлы, sockets и database connections необходимо закрывать детерминированно.

## 1. Иерархия

В C# бросать можно объекты, наследующие `System.Exception`.

Частые стандартные типы:

| Тип | Когда уместен |
|---|---|
| `ArgumentNullException` | обязательный argument равен `null` |
| `ArgumentException` | argument имеет недопустимое содержание |
| `ArgumentOutOfRangeException` | число/enum вне диапазона |
| `InvalidOperationException` | состояние объекта не допускает operation |
| `NotSupportedException` | operation принципиально не поддерживается |
| `KeyNotFoundException` | обязательный key отсутствует |
| `OperationCanceledException` | операция отменена через cancellation |

Не создавайте custom exception, если стандартный type полностью передаёт смысл.

## 2. `throw`

```csharp
public void Withdraw(decimal amount)
{
    if (amount <= 0)
        throw new ArgumentOutOfRangeException(nameof(amount));

    if (amount > Balance)
        throw new InsufficientFundsException(Id, amount, Balance);

    Balance -= amount;
}
```

Custom domain exception:

```csharp
public sealed class InsufficientFundsException : Exception
{
    public InsufficientFundsException(
        Guid accountId,
        decimal requested,
        decimal available)
        : base($"Account {accountId} has insufficient funds")
    {
        AccountId = accountId;
        Requested = requested;
        Available = available;
    }

    public Guid AccountId { get; }
    public decimal Requested { get; }
    public decimal Available { get; }
}
```

Не помещайте секреты и персональные данные в message.

## 3. `try`, `catch`, `finally`

```csharp
try
{
    await ProcessAsync(cancellationToken);
}
catch (ValidationException ex)
{
    logger.LogWarning(ex, "Input rejected");
}
finally
{
    activity?.Stop();
}
```

`finally` выполняется при normal return и большинстве exception paths. Он предназначен для cleanup, а не для замены результата или сокрытия ошибки.

## 4. Порядок `catch`

Ловите от specific к general:

```csharp
try
{
    await client.SendAsync(request, cancellationToken);
}
catch (TaskCanceledException) when (!cancellationToken.IsCancellationRequested)
{
    throw new ExternalServiceTimeoutException();
}
catch (HttpRequestException ex)
{
    throw new ExternalServiceException("Request failed", ex);
}
```

Derived exception после base exception был бы недостижим.

## 5. Exception filters

```csharp
catch (HttpRequestException ex) when (ex.StatusCode is HttpStatusCode.NotFound)
{
    return null;
}
```

Filter решает, подходит ли handler, до входа в `catch`. Это лучше, чем поймать broadly и повторно бросить по условию: сохраняется более точная exception-dispatch semantics.

## 6. Повторный throw

```csharp
catch (Exception ex)
{
    logger.LogError(ex, "Operation failed");
    throw; // сохраняет исходный stack trace
}
```

❌ `throw ex;` начинает stack trace с текущего места и теряет важный контекст.

Оборачивание сохраняет cause через inner exception:

```csharp
catch (NpgsqlException ex)
{
    throw new OrderStorageException("Cannot save order", ex);
}
```

## 7. Не ловить всё без стратегии

```csharp
catch (Exception)
{
    return default;
}
```

Такой код смешивает validation, infrastructure failure, bugs и cancellation. На application boundary общий handler может преобразовать исключение в response и залогировать его. Внутри domain/service слоя catch должен иметь конкретную цель: восстановить, добавить meaningful context, компенсировать или перевести abstraction.

Не ловите `StackOverflowException`, `OutOfMemoryException` и другие catastrophic failures как обычные business errors.

## 8. Исключения и обычное отсутствие

Ожидаемое отсутствие лучше выразить через API:

```csharp
if (cache.TryGetValue(key, out User? user))
{
    return user;
}
```

`FindAsync` может вернуть `User?`, а `GetRequiredAsync` — бросить domain-specific not-found exception. Название должно отражать contract.

## 9. Stack unwinding

При `throw` runtime ищет подходящий handler вверх по цепочке вызовов. Frames покидаются, выполняются `finally` blocks. Если handler не найден, exception достигает process/framework boundary.

```text
Controller → Service → Repository → Driver
     ▲                         │
     └──── exception unwind ───┘
```

Логировать одну ошибку на каждом уровне не нужно: получится несколько одинаковых записей. Обычно логируют там, где принимают решение или завершают request/job.

## 10. `IDisposable`

GC управляет memory, но не временем закрытия OS handles.

```csharp
using FileStream stream = File.OpenRead(path);
using var reader = new StreamReader(stream);
string content = reader.ReadToEnd();
```

`using` разворачивается в `try/finally` с вызовом `Dispose`, даже если чтение бросает исключение.

Scope statement:

```csharp
using (var connection = new NpgsqlConnection(connectionString))
{
    await connection.OpenAsync(cancellationToken);
}
```

## 11. Ownership

Кто создал resource, тот обычно и dispose его, если ownership явно не передан.

```csharp
public sealed class ReportWriter(Stream destination)
{
    public Task WriteAsync(Report report, CancellationToken ct) =>
        JsonSerializer.SerializeAsync(destination, report, cancellationToken: ct);
}
```

Здесь writer не обязан закрывать внешний stream: consumer может продолжить запись. Contract ownership надо документировать или выразить parameter/configuration.

## 12. `IAsyncDisposable`

Async cleanup нужен, если закрытие само выполняет asynchronous I/O:

```csharp
await using DbContext context = await factory.CreateDbContextAsync(ct);
await context.SaveChangesAsync(ct);
```

`await using` вызывает `DisposeAsync`.

## 13. Finalizer

Finalizer нужен почти исключительно type, напрямую владеющему unmanaged resource:

```csharp
~NativeBuffer()
{
    ReleaseUnmanagedMemory();
}
```

Он недетерминирован, усложняет GC и должен сочетаться с standard dispose pattern/SafeHandle. Обычный application class с managed members finalizer не требует.

## 14. Cancellation — не ошибка бизнеса

Cancellation обычно выражается `OperationCanceledException`, связанной с token. Не превращайте её в HTTP 500 и не retry отменённую пользователем операцию.

```csharp
catch (OperationCanceledException) when (cancellationToken.IsCancellationRequested)
{
    throw;
}
```

## 15. Сравнение с Java

| Java | C# |
|---|---|
| checked и unchecked exceptions | checked exceptions отсутствуют |
| `throws` в сигнатуре | нет language-level `throws` contract |
| try-with-resources / `AutoCloseable` | `using` / `IDisposable`, `await using` / `IAsyncDisposable` |
| `throw e` обычно сохраняет Java trace semantics | в C# для сохранения нужен `throw;`, не `throw ex;` |
| suppressed exceptions | dispose exceptions имеют свои правила; не полагайтесь на полную идентичность |

Отсутствие checked exceptions не отменяет documentation contract. Публичный API должен описывать значимые failure modes.

## 16. Best practices

- Исключения — для exceptional paths, а не обычного branching.
- Ловите только то, что можете осмысленно обработать.
- Сохраняйте inner exception при переводе abstraction.
- Используйте `throw;` для повторного проброса.
- Dispose resources детерминированно.
- Явно определяйте ownership.
- Не логируйте один exception на каждом слое.
- Не поглощайте cancellation.

## Самопроверка

1. Чем `throw;` отличается от `throw ex;`?
2. Когда exception filter лучше проверки внутри catch?
3. Освобождает ли GC файл сразу после потери ссылки?
4. Когда нужен `await using`?
5. Где уместен общий `catch (Exception)`?

