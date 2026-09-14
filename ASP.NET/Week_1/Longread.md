# Лекция 1 — Введение в ASP.NET Core, Generic Host и Dependency Injection

## Что вы изучите

- отличие framework от library;
- состав ASP.NET Core приложения;
- Generic Host и жизненный цикл процесса;
- встроенный DI container;
- регистрацию и constructor injection;
- controller-based и Minimal API models.

## 1. Framework и inversion of control

Library вызывает ваш код только тогда, когда вы явно обращаетесь к её функции. Framework задаёт lifecycle и вызывает зарегистрированные компоненты в нужные моменты.

```text
Library:   Application → Library
Framework: Host → Application callbacks/components
```

ASP.NET Core управляет startup, configuration, logging, DI, web server, middleware pipeline, routing и graceful shutdown.

## 2. Что входит в ASP.NET Core

ASP.NET Core — web framework поверх .NET. Основные части:

- **Kestrel** — cross-platform HTTP server;
- **Generic Host** — configuration, logging, DI и lifetime;
- **middleware pipeline** — обработка запроса по цепочке;
- **routing/endpoints** — выбор обработчика;
- controllers/Minimal APIs/Razor/SignalR/gRPC — прикладные programming models;
- authentication, authorization, CORS, rate limiting, health checks;
- integration with OpenTelemetry, configuration providers и background services.

ASP.NET Core не равен старому ASP.NET Framework и не требует IIS, хотя может работать за reverse proxy IIS/nginx/Envoy.

## 3. Минимальное приложение

```csharp
var builder = WebApplication.CreateBuilder(args);

builder.Services.AddControllers();

var app = builder.Build();

app.MapControllers();

app.Run();
```

### По шагам

1. `CreateBuilder` создаёт host builder, загружает default configuration и logging.
2. `builder.Services` — `IServiceCollection`, список registrations.
3. `Build` формирует service provider и web application.
4. `MapControllers` добавляет controller endpoints.
5. `Run` запускает server и ожидает shutdown.

После `Build` composition root считается собранным. Registrations обычно выполняются до него.

## 4. Dependency Injection

Проблема жёсткой связности:

```csharp
public sealed class OrderService
{
    private readonly PostgresOrderRepository _repository = new();
}
```

Service сам выбирает storage, connection configuration и lifetime. Заменить implementation трудно.

Dependency injection:

```csharp
public interface IOrderRepository
{
    Task<Order?> FindAsync(Guid id, CancellationToken ct);
}

public sealed class OrderService(IOrderRepository repository)
{
    public Task<Order?> FindAsync(Guid id, CancellationToken ct) =>
        repository.FindAsync(id, ct);
}
```

Registration:

```csharp
builder.Services.AddScoped<IOrderRepository, PostgresOrderRepository>();
builder.Services.AddScoped<OrderService>();
```

Container строит object graph по constructor parameters.

## 5. Constructor injection

```csharp
[ApiController]
[Route("api/orders")]
public sealed class OrdersController(OrderService service) : ControllerBase
{
    [HttpGet("{id:guid}")]
    public async Task<ActionResult<OrderDto>> Get(
        Guid id,
        CancellationToken cancellationToken)
    {
        Order? order = await service.FindAsync(id, cancellationToken);
        return order is null ? NotFound() : Ok(OrderDto.From(order));
    }
}
```

Constructor injection:

- делает required dependencies явными;
- позволяет `readonly` design;
- облегчает unit testing;
- обнаруживает часть ошибок composition при startup/resolution.

Property/service-locator injection скрывает contract и обычно не рекомендуется.

## 6. Registration variants

```csharp
services.AddTransient<IFormatter, JsonFormatter>();
services.AddScoped<IOrderRepository, EfOrderRepository>();
services.AddSingleton<IClock, SystemClock>();

services.AddSingleton(new FeatureFlags { NewCheckout = true });

services.AddScoped<IPriceCalculator>(provider =>
{
    var clock = provider.GetRequiredService<IClock>();
    return new PriceCalculator(clock);
});
```

Factory полезна при conditional construction, но не должна превращаться в service locator с десятками `GetRequiredService`.

## 7. Несколько implementations

```csharp
services.AddScoped<INotifier, EmailNotifier>();
services.AddScoped<INotifier, SmsNotifier>();
```

Consumer получает все registrations:

```csharp
public sealed class NotificationService(IEnumerable<INotifier> notifiers)
{
    public Task NotifyAllAsync(Message message, CancellationToken ct) =>
        Task.WhenAll(notifiers.Select(x => x.SendAsync(message, ct)));
}
```

Для выбора по ключу .NET DI поддерживает keyed services:

```csharp
services.AddKeyedScoped<INotifier, EmailNotifier>("email");
```

Не заменяйте ясную strategy abstraction строковыми keys повсюду.

## 8. Options вместо ручного чтения configuration

```csharp
builder.Services
    .AddOptions<PaymentOptions>()
    .BindConfiguration("Payment")
    .ValidateDataAnnotations()
    .ValidateOnStart();
```

Typed options делают config contract видимым. Подробно — неделя 3.

## 9. Controllers и Minimal APIs

Minimal endpoint:

```csharp
app.MapGet("/api/orders/{id:guid}", async (
    Guid id,
    OrderService service,
    CancellationToken ct) =>
{
    Order? order = await service.FindAsync(id, ct);
    return order is null
        ? Results.NotFound()
        : Results.Ok(OrderDto.From(order));
});
```

Controllers дают attributes, filters, conventions и привычную группировку большого API. Minimal APIs уменьшают ceremony и подходят для небольших endpoints/infrastructure services. Оба варианта используют routing, DI и middleware.

## 10. Environments и configuration defaults

Default builder читает configuration из нескольких sources, включая `appsettings.json`, environment-specific JSON, environment variables и command-line. Более поздний provider обычно переопределяет более ранний key.

Никогда не храните production secrets в committed `appsettings.json`.

## 11. Logging

```csharp
public sealed class OrderService(
    IOrderRepository repository,
    ILogger<OrderService> logger)
{
    public async Task<Order?> FindAsync(Guid id, CancellationToken ct)
    {
        logger.LogInformation("Loading order {OrderId}", id);
        return await repository.FindAsync(id, ct);
    }
}
```

Message template сохраняет structured fields. Не интерполируйте secret/PII.

## 12. Сравнение со Spring

| Spring | ASP.NET Core |
|---|---|
| `@SpringBootApplication` | `WebApplication.CreateBuilder` + composition root |
| `ApplicationContext` | built `IServiceProvider` внутри Host |
| `@Bean`/component scanning | explicit `IServiceCollection` registrations |
| constructor `@Autowired` | constructor injection без attribute |
| `@Qualifier` | keyed services или explicit factory/strategy |
| Spring MVC controller | MVC/API controller |
| embedded Tomcat | Kestrel |

Встроенный container намеренно проще Spring. Component scanning и сложные post-processors не являются default model; explicit registration облегчает понимание composition.

## 13. Частые ошибки

- Вызывать `BuildServiceProvider()` во время registrations: появляется второй container и дублируются singletons.
- Получать dependencies через `app.Services` в каждом method.
- Регистрировать mutable singleton без thread-safety.
- Смешивать domain model и API DTO.
- Создавать HTTP/DB clients вручную на каждый request без framework factories/pooling.

## Итоги

ASP.NET Core application — Host, DI container, server и request pipeline. `Program.cs` является composition root: здесь связываются abstractions и implementations. Constructor injection — основной способ выражения dependencies.

## Самопроверка

1. Чем framework отличается от library по потоку управления?
2. Что происходит до и после `builder.Build()`?
3. Почему `BuildServiceProvider` внутри registration опасен?
4. Что общего у controllers и Minimal APIs?
5. Когда `IEnumerable<T>` используется для нескольких implementations?

