# Лекция 3 — Архитектура, конфигурация, Options и Hosted Services

## 1. Архитектура уровня приложения

Типичная layered dependency direction:

```text
API → Application → Domain
Infrastructure ───────┘
API composition root → Infrastructure registrations
```

- Domain хранит business rules и не зависит от web/EF.
- Application координирует use cases и abstractions.
- Infrastructure реализует database, brokers, files, external clients.
- API переводит HTTP contract в application calls.

Количество projects зависит от размера. Layers можно соблюдать namespaces/folders в одном assembly. Архитектура — правила dependencies, а не дерево каталогов.

## 2. Modular monolith

Modules группируют feature/domain capabilities:

```text
Modules/
├─ Orders/
│  ├─ Domain
│  ├─ Application
│  ├─ Infrastructure
│  └─ Api
└─ Payments/
```

Module boundaries уменьшают случайное coupling. Shared project не должен превращаться в склад domain entities всех modules.

## 3. Configuration pipeline

Default `WebApplicationBuilder` добавляет providers в определённом порядке. Поздний source перекрывает ранний key.

```json
{
  "Payments": {
    "BaseUrl": "https://sandbox.example",
    "TimeoutSeconds": 10
  }
}
```

Environment variable для nested key:

```text
Payments__TimeoutSeconds=5
```

Double underscore переносимее colon между shells/containers.

## 4. Typed Options

```csharp
public sealed class PaymentOptions
{
    public const string SectionName = "Payments";

    [Required, Url]
    public required string BaseUrl { get; init; }

    [Range(1, 120)]
    public int TimeoutSeconds { get; init; } = 10;
}
```

Registration:

```csharp
services.AddOptions<PaymentOptions>()
    .BindConfiguration(PaymentOptions.SectionName)
    .ValidateDataAnnotations()
    .Validate(options => Uri.TryCreate(
        options.BaseUrl,
        UriKind.Absolute,
        out _), "BaseUrl must be absolute")
    .ValidateOnStart();
```

## 5. `IOptions`, `IOptionsSnapshot`, `IOptionsMonitor`

| API | Lifetime/поведение | Использование |
|---|---|---|
| `IOptions<T>` | singleton-style cached value | config не требуется перечитывать |
| `IOptionsSnapshot<T>` | scoped, пересчитывается на scope | request получает snapshot |
| `IOptionsMonitor<T>` | singleton-friendly, change notifications | long-lived services/reload |

Reload зависит от provider. Изменение config не гарантирует, что внешняя система поддерживает безопасную hot reconfiguration.

```csharp
public sealed class MonitorConsumer(IOptionsMonitor<PaymentOptions> options)
{
    public string CurrentUrl => options.CurrentValue.BaseUrl;
}
```

Подписку `OnChange` нужно dispose, если consumer живёт меньше monitor.

## 6. Named options

```csharp
services.AddOptions<PaymentOptions>("primary")
    .BindConfiguration("Payments:Primary");
services.AddOptions<PaymentOptions>("backup")
    .BindConfiguration("Payments:Backup");
```

Named options подходят для нескольких configurations одного type. Key должен быть controlled application value.

## 7. Environments

```csharp
if (app.Environment.IsDevelopment())
{
    app.MapOpenApi();
}
```

Обычно используются Development, Staging, Production. Environment — deployment choice, не authorization mechanism. Нельзя скрывать опасный endpoint только условием среды, если сеть всё равно может быть неверно настроена.

## 8. Secrets

- локально: user secrets или environment variables;
- production: secret manager/vault/workload identity;
- repository: только безопасные defaults/templates.

Configuration API хранит values как strings и не является secret vault. Не логируйте весь configuration dump.

## 9. Registration modules

```csharp
public static class OrdersModule
{
    public static IServiceCollection AddOrders(
        this IServiceCollection services,
        IConfiguration configuration)
    {
        services.AddScoped<OrderService>();
        services.AddScoped<IOrderRepository, EfOrderRepository>();
        return services;
    }
}
```

Extension method группирует composition без скрытого component scanning.

## 10. Hosted services

```csharp
public sealed class QueueWorker(
    IServiceScopeFactory scopeFactory,
    ILogger<QueueWorker> logger) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            await using AsyncServiceScope scope = scopeFactory.CreateAsyncScope();
            var job = scope.ServiceProvider.GetRequiredService<ProcessNextJob>();
            await job.ExecuteAsync(stoppingToken);
        }
    }
}
```

Hosted service зарегистрирован singleton. Scoped dependency создаётся внутри work-unit scope.

## 11. Startup readiness

`StartAsync` влияет на startup. Долгая initialization задерживает readiness. Подходы:

- validate critical config on start;
- выполнить короткий обязательный init;
- долгий warmup вести с readiness false до завершения;
- schema migration координировать отдельным deployment job при нескольких replicas.

## 12. Shutdown

`stoppingToken` сигнализирует graceful stop. Worker должен:

- прекратить принимать новую работу;
- завершить/отменить текущую по policy;
- commit/abandon broker message корректно;
- закрыть scope;
- уложиться в shutdown timeout.

## 13. Сравнение со Spring

| Spring | ASP.NET Core |
|---|---|
| `application.yml` + property sources | configuration providers |
| `@Value` | прямой `IConfiguration`, но typed options предпочтительнее |
| `@ConfigurationProperties` | Options binding/validation |
| `@Profile` | environment/config-driven registration |
| `@Configuration`/`@Import` | registration extension modules |
| scheduled/background beans | `IHostedService`/`BackgroundService` |

## 14. Best practices

- Validate required config на startup.
- Domain не читает global configuration.
- Передавайте typed options или уже вычисленный domain value.
- Не ветвите business rules напрямую по environment name.
- Background service создаёт scope на unit of work.
- Reloadable setting должен иметь определённую concurrency/error policy.

## Самопроверка

1. Что важнее: количество projects или dependency direction?
2. Какой provider выигрывает при одинаковом key?
3. Чем snapshot отличается от monitor?
4. Почему hosted service не может напрямую держать scoped DbContext?
5. Где хранить production secrets?

