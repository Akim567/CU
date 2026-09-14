# Семинар 3 — Архитектура, Options и фоновые services

## Задание 1. Разделите слои

Разместите существующий Order Service по слоям или folders:

```text
Domain/Order.cs
Application/OrderService.cs
Application/IOrderRepository.cs
Infrastructure/InMemoryOrderRepository.cs
Api/OrdersController.cs
```

**Проверка:** Domain не содержит references на ASP.NET Core, configuration или logging abstractions.

## Задание 2. Registration module

```csharp
public static class OrderServiceRegistration
{
    public static IServiceCollection AddOrderModule(
        this IServiceCollection services)
    {
        services.AddSingleton<IOrderRepository, InMemoryOrderRepository>();
        services.AddScoped<OrderService>();
        return services;
    }
}
```

В `Program.cs` должна остаться строка `builder.Services.AddOrderModule()`.

## Задание 3. Bind и validate options

Создайте:

```json
{
  "OrderProcessing": {
    "BatchSize": 20,
    "IntervalSeconds": 10
  }
}
```

```csharp
public sealed class OrderProcessingOptions
{
    [Range(1, 500)]
    public int BatchSize { get; init; }

    [Range(1, 3600)]
    public int IntervalSeconds { get; init; }
}
```

Bind, validate annotations и включите `ValidateOnStart`. Проверьте, что `BatchSize=0` ломает startup с понятной ошибкой.

## Задание 4. Environment override

Задайте `OrderProcessing__BatchSize=5` через environment variable. Endpoint `/debug/options` должен показать effective value, но не должен возвращать secrets.

Объясните precedence.

## Задание 5. Snapshot против Monitor

Создайте:

- scoped consumer `IOptionsSnapshot<OrderProcessingOptions>`;
- singleton consumer `IOptionsMonitor<OrderProcessingOptions>`.

Если provider поддерживает reload, измените JSON и сравните поведение. Зафиксируйте, что live reload — capability provider, а не гарантия любого deployment.

## Задание 6. Background worker

Создайте bounded `Channel<Guid>`, endpoint добавления id и `BackgroundService`, который читает jobs.

```csharp
services.AddSingleton(Channel.CreateBounded<Guid>(100));
services.AddHostedService<OrderWorker>();
```

Worker на каждый id создаёт async scope и вызывает scoped `ProcessOrderJob`.

## Задание 7. Graceful shutdown

Добавьте задержку обработки, остановите process через Ctrl+C и проверьте:

- token приходит worker;
- новая работа не принимается;
- scope закрывается;
- cancellation не логируется как unexpected error.

## Задание 8. Опасная конфигурация

Создайте secret-like key и убедитесь, что он не попадает:

- в repository;
- в response `/debug/options`;
- в structured log state;
- в exception message.

## Итоговый чек-лист

- dependency direction соблюдён;
- registrations сгруппированы без второго provider;
- invalid config останавливает startup;
- environment override работает;
- worker не удерживает scoped dependency;
- channel bounded;
- shutdown и cancellation наблюдаемы.

