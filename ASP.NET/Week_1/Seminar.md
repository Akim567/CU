# Семинар 1 — Первое ASP.NET Core приложение и Dependency Injection

## Цель

Создать controller-based API без генераторов IDE, зарегистрировать несколько services и увидеть, как container строит object graph.

## Подготовка

```bash
dotnet new webapi -n OrderService --use-controllers
cd OrderService
dotnet run
```

Удалите демонстрационный WeatherForecast endpoint.

## Задание 1. Минимальный composition root

1. Добавьте controllers.
2. Соберите application.
3. Подключите controller endpoints.
4. Добавьте `/health/manual`, возвращающий `{ "status": "ok" }`.

**Решение:**

```csharp
var builder = WebApplication.CreateBuilder(args);
builder.Services.AddControllers();

var app = builder.Build();
app.MapControllers();
app.MapGet("/health/manual", () => Results.Ok(new { status = "ok" }));
app.Run();
```

## Задание 2. Repository abstraction

Создайте `Order`, `IOrderRepository` и in-memory implementation. Repository должен иметь `FindAsync` и `AddAsync` с cancellation token.

**Решение-каркас:**

```csharp
public sealed record Order(Guid Id, string Number, decimal Total);

public interface IOrderRepository
{
    Task<Order?> FindAsync(Guid id, CancellationToken ct);
    Task AddAsync(Order order, CancellationToken ct);
}

public sealed class InMemoryOrderRepository : IOrderRepository
{
    private readonly ConcurrentDictionary<Guid, Order> _orders = new();

    public Task<Order?> FindAsync(Guid id, CancellationToken ct)
    {
        ct.ThrowIfCancellationRequested();
        _orders.TryGetValue(id, out Order? order);
        return Task.FromResult(order);
    }

    public Task AddAsync(Order order, CancellationToken ct)
    {
        ct.ThrowIfCancellationRequested();
        if (!_orders.TryAdd(order.Id, order))
            throw new InvalidOperationException("Order already exists");
        return Task.CompletedTask;
    }
}
```

## Задание 3. Service и controller

Добавьте `OrderService`, POST `/api/orders` и GET `/api/orders/{id}`. Не возвращайте domain entity напрямую.

```csharp
public sealed record CreateOrderRequest(string Number, decimal Total);
public sealed record OrderResponse(Guid Id, string Number, decimal Total);
```

**Ключ решения:**

```csharp
builder.Services.AddSingleton<IOrderRepository, InMemoryOrderRepository>();
builder.Services.AddScoped<OrderService>();
```

Singleton repository выбран только для семинара, чтобы данные переживали requests. Concurrent collection обязательна из-за одновременных запросов.

## Задание 4. Несколько notifier implementations

1. Создайте `IOrderCreatedNotifier`.
2. Реализуйте console и audit notifier.
3. Зарегистрируйте оба.
4. Внедрите `IEnumerable<IOrderCreatedNotifier>` в service.
5. Уведомляйте после успешного добавления.

```csharp
builder.Services.AddScoped<IOrderCreatedNotifier, ConsoleNotifier>();
builder.Services.AddScoped<IOrderCreatedNotifier, AuditNotifier>();
```

Не используйте fire-and-forget: `Task.WhenAll` нужно await.

## Задание 5. Keyed services

Зарегистрируйте email/sms implementation по keys и создайте endpoint, выбирающий канал из ограниченного enum. Не передавайте arbitrary user string прямо в `GetRequiredKeyedService` без validation.

## Задание 6. Продемонстрируйте антипример

Временно вызовите `builder.Services.BuildServiceProvider()` два раза и получите singleton из каждого provider. Докажите, что instances разные. Затем удалите антипример.

## Проверка

- неизвестный order даёт 404;
- некорректный request не создаёт order;
- два одинаковых id не перезаписываются;
- оба notifier вызваны;
- `CancellationToken` доходит до repository;
- в production code нет ручного `BuildServiceProvider`.

## Вопросы

1. Почему in-memory repository зарегистрирован singleton?
2. Почему обычный `Dictionary` был бы опасен?
3. Что изменится при transient repository?
4. Где находится composition root?

