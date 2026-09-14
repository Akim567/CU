# Курс ASP.NET Core 10 и EF Core 10

Это пятнадцатинедельный курс серверной разработки на .NET 10 LTS. По глубине и формату он соответствует `Spring/`: каждая неделя содержит теоретический лонгрид и практический семинар с заданиями, решениями и критериями проверки.

Основная сквозная предметная область — сервис заказов (**Order Service**). Проекты для семинаров намеренно не добавлены: фрагменты предназначены для самостоятельной сборки учебного приложения.

## Предпосылки

- пройдены основы из [`CSharp/`](../CSharp/README.md);
- установлен .NET 10 SDK;
- для недель 6–8 доступен PostgreSQL;
- Docker удобен для PostgreSQL, Kafka, Prometheus, Grafana и Testcontainers, но способ установки остаётся на выбор студента.

## Программа

| Неделя | Лекция | Семинар |
|---:|---|---|
| 1 | [Введение, Generic Host и DI](Week_1/Longread.md) | [Первое приложение и DI](Week_1/Seminar.md) |
| 2 | [Lifetimes, lifecycle, disposal и decorators](Week_2/Longread.md) | [Жизненный цикл и interception](Week_2/Seminar.md) |
| 3 | [Архитектура и конфигурация](Week_3/Longread.md) | [Options, environments и hosted services](Week_3/Seminar.md) |
| 4 | [HTTP pipeline, middleware и controllers](Week_4/Longread.md) | [Request processing и validation](Week_4/Seminar.md) |
| 5 | [Ответы, ошибки, CORS и OpenAPI](Week_5/Longread.md) | [Единый HTTP contract](Week_5/Seminar.md) |
| 6 | [ADO.NET, Npgsql и транзакции](Week_6/Longread.md) | [Работа с PostgreSQL без ORM](Week_6/Seminar.md) |
| 7 | [Основы EF Core](Week_7/Longread.md) | [Entities, tracking и relations](Week_7/Seminar.md) |
| 8 | [Производительность EF Core](Week_8/Longread.md) | [N+1, paging и concurrency](Week_8/Seminar.md) |
| 9 | [Security, Identity, JWT и logging](Week_9/Longread.md) | [Authentication и policies](Week_9/Seminar.md) |
| 10 | [HTTP clients и resilience](Week_10/Longread.md) | [HttpClientFactory, retry, rate limit, circuit breaker](Week_10/Seminar.md) |
| 11 | [Тестирование](Week_11/Longread.md) | [Unit, integration и E2E](Week_11/Seminar.md) |
| 12 | [Async HTTP, streaming и Kestrel](Week_12/Longread.md) | [Cancellation и streaming endpoints](Week_12/Seminar.md) |
| 13 | [Асинхронная БД и background processing](Week_13/Longread.md) | [Queues, workers и outbox](Week_13/Seminar.md) |
| 14 | [Event-driven architecture и Kafka](Week_14/Longread.md) | [Producer, consumer, retry и idempotency](Week_14/Seminar.md) |
| 15 | [Observability](Week_15/Longread.md) | [OpenTelemetry, Prometheus и health checks](Week_15/Seminar.md) |

Дополнительно: [экзаменационные вопросы](exam_questions.md).

## Базовый проект

```xml
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
  </PropertyGroup>
</Project>
```

## Как выполнять семинары

1. Сначала прочитайте цель и критерии.
2. Реализуйте задание без просмотра решения.
3. Проверьте normal, invalid и failure scenarios.
4. Сравните API и lifetime behavior с решением.
5. Не копируйте package version вслепую: используйте последнюю совместимую patch-версию линии 10.x.

## Сравнение терминов со Spring

| Spring | ASP.NET Core |
|---|---|
| ApplicationContext/BeanFactory | `IServiceProvider` и Generic Host |
| bean | зарегистрированный service/component instance |
| singleton/prototype/request scope | Singleton/Transient/Scoped |
| BeanPostProcessor/AOP proxy | decorators, middleware, filters, interceptors/proxies |
| Spring MVC DispatcherServlet | ASP.NET Core middleware + endpoint routing |
| `@ControllerAdvice` | exception handler middleware + Problem Details |
| JdbcTemplate | ADO.NET/provider APIs; часто Dapper как дополнительная библиотека |
| JPA/Hibernate | EF Core |
| Spring Security filter chain | authentication/authorization middleware + handlers |
| WebFlux/Reactor | Task-based async, async streams, Kestrel pipeline |
| Actuator/Micrometer | health checks, `System.Diagnostics.Metrics`, OpenTelemetry |

Сходство по роли не означает идентичное устройство. Каждый лонгрид отдельно отмечает важные различия.

