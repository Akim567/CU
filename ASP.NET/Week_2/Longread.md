# Лекция 2 — DI lifetimes, жизненный цикл, disposal и interception

## 1. Три lifetime

| Lifetime | Создание | Обычно живёт | Типичный пример |
|---|---|---|---|
| Transient | при каждом resolution | до конца owning scope/provider | stateless formatter |
| Scoped | один instance на scope | HTTP request либо manual scope | `DbContext` |
| Singleton | один на root provider | весь process | immutable cache metadata, clock |

```csharp
services.AddTransient<IFormatter, JsonFormatter>();
services.AddScoped<IOrderRepository, EfOrderRepository>();
services.AddSingleton<IClock, SystemClock>();
```

Lifetime — ownership contract, а не производственная оптимизация «singleton быстрее».

## 2. Жизненный цикл resolution

Упрощённо:

```text
registration
  → provider build
  → scope creation
  → constructor graph activation
  → component use
  → scope disposal (reverse dependency order)
  → root provider disposal
```

Встроенный container строит graph по constructors. У него нет прямого аналога полного `BeanPostProcessor` pipeline Spring.

## 3. Scoped request

ASP.NET Core создаёт service scope для HTTP request. Все scoped resolutions внутри него разделяют instance:

```csharp
public sealed class RequestMarker
{
    public Guid Id { get; } = Guid.NewGuid();
}
```

Два consumers в одном request получают один `RequestMarker`; следующий request — другой.

Scope не равен thread. Request может продолжиться на другом ThreadPool thread после `await`.

## 4. Captive dependency

```csharp
services.AddScoped<OrderDbContext>();
services.AddSingleton<OrderCache>();

public sealed class OrderCache(OrderDbContext db); // ошибка design
```

Singleton удерживает scoped dependency дольше scope. Последствия: stale state, memory retention, concurrent access к non-thread-safe `DbContext`.

Development validation:

```csharp
builder.Host.UseDefaultServiceProvider(options =>
{
    options.ValidateScopes = true;
    options.ValidateOnBuild = true;
});
```

Validation ловит многие, но не все dynamic/factory problems.

## 5. Singleton thread safety

Один instance используется concurrent requests. Он должен быть immutable либо синхронизирован.

```csharp
public sealed class Sequence
{
    private long _value;
    public long Next() => Interlocked.Increment(ref _value);
}
```

Transient не гарантирует отсутствия shared state: он может зависеть от static data или shared singleton.

## 6. Disposal container-owned services

```csharp
public sealed class TrackedConnection : IDisposable
{
    public void Dispose() => Console.WriteLine("disposed");
}

services.AddScoped<TrackedConnection>();
```

Instance, созданный container, dispose-ится с owning scope/provider. Не вызывайте `Dispose` вручную у injected service.

Instance, переданный registration как готовый объект, имеет особый ownership contract: container может не считать себя его создателем. Предпочитайте factory/type registration и сверяйте выбранный overload.

## 7. Async disposal

Service может реализовать `IAsyncDisposable`. Scope/provider поддерживают async disposal:

```csharp
await using AsyncServiceScope scope =
    app.Services.CreateAsyncScope();

var worker = scope.ServiceProvider.GetRequiredService<JobRunner>();
await worker.RunAsync(ct);
```

Вручную созданный scope всегда нужно закрывать.

## 8. Factory и manual scopes в singleton

Background service — singleton, но work unit может требовать scoped dependencies:

```csharp
public sealed class CleanupWorker(
    IServiceScopeFactory scopeFactory) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            await using AsyncServiceScope scope = scopeFactory.CreateAsyncScope();
            var job = scope.ServiceProvider.GetRequiredService<CleanupJob>();
            await job.RunAsync(stoppingToken);
            await Task.Delay(TimeSpan.FromMinutes(1), stoppingToken);
        }
    }
}
```

Это controlled scope factory, не разрешение service locator повсюду.

## 9. `IDbContextFactory<T>`

Для нескольких units of work, background processing или UI-like scopes:

```csharp
await using OrderDbContext db = await factory.CreateDbContextAsync(ct);
```

Каждый context имеет короткий lifetime; параллельные operations используют разные contexts.

## 10. Decorator

```csharp
public sealed class LoggingOrderRepository(
    IOrderRepository inner,
    ILogger<LoggingOrderRepository> logger) : IOrderRepository
{
    public async Task<Order?> FindAsync(Guid id, CancellationToken ct)
    {
        using IDisposable? scope = logger.BeginScope(new Dictionary<string, object>
        {
            ["OrderId"] = id
        });
        return await inner.FindAsync(id, ct);
    }
}
```

Decorator явно применяет cross-cutting behavior. Registration вручную требует различить inner concrete type:

```csharp
services.AddScoped<EfOrderRepository>();
services.AddScoped<IOrderRepository>(sp =>
    new LoggingOrderRepository(
        sp.GetRequiredService<EfOrderRepository>(),
        sp.GetRequiredService<ILogger<LoggingOrderRepository>>()));
```

Библиотеки вроде Scrutor упрощают decoration, но механизм остаётся composition.

## 11. Middleware, filters и interceptors

Cross-cutting concern выбирают по boundary:

- middleware — весь HTTP pipeline;
- endpoint/MVC filters — выбранные endpoints/actions;
- decorator — application interface calls;
- EF interceptors — database operations;
- `DispatchProxy`/Castle DynamicProxy — runtime proxy scenarios;
- source generators — compile-time wrapping/generation.

Один AOP hammer не нужен для всех уровней.

## 12. Proxy limitations

Runtime proxy перехватывает только calls, проходящие через proxy contract. Self-invocation внутри target может обойти interceptor. Non-virtual concrete members нельзя перехватить class proxy в обычной модели. Constructors уже выполнились до вызовов proxy methods.

Эти ограничения концептуально похожи на Spring proxy AOP, но конкретные libraries и dispatch rules различаются.

## 13. Startup и shutdown

Для initialization используйте:

- constructors только для дешёвых invariants;
- `IHostedService.StartAsync` для process-level startup;
- explicit initializer до `Run`;
- lazy initialization с thread-safe policy.

Shutdown:

- Host сигнализирует cancellation;
- hosted services получают `StopAsync`;
- provider dispose-ит services;
- есть ограниченный shutdown timeout.

Не помещайте долгую миграцию/сетевой вызов в constructor.

## 14. Сравнение со Spring Week 2

| Spring | ASP.NET Core/.NET DI |
|---|---|
| singleton bean | Singleton service |
| prototype | Transient |
| request scope | Scoped request service |
| `@PostConstruct` | explicit initializer/hosted service; универсального callback нет |
| `@PreDestroy` | `Dispose`/`DisposeAsync`/host shutdown |
| BeanPostProcessor | нет общего built-in эквивалента; decorators/factories/framework hooks |
| Spring AOP | middleware, filters, decorators, specialized interceptors/proxies |

## 15. Best practices

- Scoped `DbContext`, short unit of work.
- Singleton не зависит от scoped.
- Не храните scoped service в static field.
- Container-owned disposable не закрывайте вручную.
- Manual scope всегда dispose.
- Decorator регистрируйте с тем же lifetime, совместимым с inner dependencies.
- Cross-cutting механизм выбирайте по реальной границе.

## Самопроверка

1. Почему scope не равен thread?
2. Что такое captive dependency?
3. Кто dispose-ит injected scoped service?
4. Как background singleton использует `DbContext`?
5. Почему self-invocation может обойти proxy interceptor?

