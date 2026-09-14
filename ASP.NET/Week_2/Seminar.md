# Семинар 2 — Lifetimes, lifecycle и decorators

## Цель

Наблюдать создание/disposal services, воспроизвести captive dependency и реализовать безопасное cross-cutting behavior.

## Задание 1. Наблюдатель lifetime

Создайте три services с GUID и логированием constructors/disposal:

```csharp
public abstract class TrackedService(string name) : IDisposable
{
    public Guid InstanceId { get; } = Guid.NewGuid();

    public void Dispose() =>
        Console.WriteLine($"DISPOSE {name}: {InstanceId}");
}

public sealed class TransientMarker() : TrackedService("transient") { }
public sealed class ScopedMarker() : TrackedService("scoped") { }
public sealed class SingletonMarker() : TrackedService("singleton") { }
```

Зарегистрируйте соответствующие lifetimes. В endpoint разрешите каждый marker дважды через два consumer services и верните ids.

**Ожидание:** transient ids различаются, scoped совпадают в одном request, singleton совпадает между requests.

## Задание 2. Порядок disposal

Создайте `OuterService`, зависящий от `InnerService`; оба disposable и scoped. Завершите request и посмотрите порядок.

**Объяснение:** container dispose-ит graph в порядке, безопасном для ownership, обычно reverse creation/dependency order. Не стройте business protocol на недокументированном порядке между несвязанными services.

## Задание 3. Сломайте scope validation

```csharp
public sealed class BrokenSingleton(ScopedMarker marker)
{
}

services.AddScoped<ScopedMarker>();
services.AddSingleton<BrokenSingleton>();
```

Включите:

```csharp
builder.Host.UseDefaultServiceProvider(options =>
{
    options.ValidateScopes = true;
    options.ValidateOnBuild = true;
});
```

Зафиксируйте startup exception. Затем исправьте design: удалите зависимость либо создавайте scope на каждую operation через `IServiceScopeFactory`.

## Задание 4. Manual async scope

Создайте singleton `ReportWorker`, который по endpoint выполняет scoped `ReportJob`. Каждый запуск должен получить новый scoped marker и закрыть его даже при exception.

```csharp
await using AsyncServiceScope scope = scopeFactory.CreateAsyncScope();
var job = scope.ServiceProvider.GetRequiredService<ReportJob>();
await job.RunAsync(ct);
```

## Задание 5. Decorator repository

Реализуйте timing decorator:

```csharp
public sealed class TimedOrderRepository(
    IOrderRepository inner,
    ILogger<TimedOrderRepository> logger) : IOrderRepository
{
    public async Task<Order?> FindAsync(Guid id, CancellationToken ct)
    {
        long started = Stopwatch.GetTimestamp();
        try
        {
            return await inner.FindAsync(id, ct);
        }
        finally
        {
            TimeSpan elapsed = Stopwatch.GetElapsedTime(started);
            logger.LogInformation(
                "Order lookup took {ElapsedMs} ms",
                elapsed.TotalMilliseconds);
        }
    }
}
```

Поддержите все methods interface, не изменяя domain behavior.

## Задание 6. Exception decorator

Добавьте decorator, переводящий только storage-specific exception в `OrderStorageException`. Cancellation и programming errors не поглощайте.

## Задание 7. Выберите правильный interception level

Для каждого concern выберите middleware/filter/decorator/EF interceptor и объясните:

1. correlation id для каждого HTTP request;
2. authorization конкретного action;
3. retry repository command;
4. SQL command timing;
5. validation application command.

**Ориентир:** middleware, authorization policy/filter, decorator/resilience policy с идемпотентностью, EF interceptor, decorator/pipeline behavior соответственно.

## Задание 8. Self-invocation experiment

Если используете proxy library, перехватите virtual public method, затем вызовите его из другого method того же object. Объясните, почему внутренний call может не пройти через proxy.

## Проверка

- lifetime ids соответствуют contract;
- scope validation ловит captive dependency;
- manual scope закрывается при exception;
- decorator не меняет return/exception contract без явной цели;
- cancellation не переводится в 500/storage error;
- cross-cutting mechanisms выбраны по boundary.
