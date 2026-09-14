# Лекция 15 — Observability: logs, metrics, traces и health checks

## 1. Monitoring и observability

Monitoring отвечает на заранее известные вопросы с dashboards/alerts. Observability помогает исследовать внутреннее состояние системы по внешним сигналам, включая неизвестные заранее failure modes.

Три основных сигнала:

- logs — discrete events;
- metrics — числовые time series;
- traces — путь операции через services.

Они дополняют, а не заменяют друг друга.

## 2. Четыре золотых сигнала

1. **Latency** — распределение времени, отдельно success/failure.
2. **Traffic** — запросы/операции в секунду.
3. **Errors** — доля и категории failures.
4. **Saturation** — исчерпание ограниченного ресурса: CPU, pool, queue, threads.

Среднее скрывает tail latency. Для SLO нужны percentiles/histograms.

## 3. Logs

```csharp
logger.LogInformation(
    "Order {OrderId} changed from {OldStatus} to {NewStatus}",
    orderId,
    oldStatus,
    newStatus);
```

Structured fields позволяют фильтровать без парсинга текста. Хороший log event имеет стабильное имя/shape, level и context.

Запрещено:

- raw tokens/passwords/connection strings;
- бесконтрольные request/response bodies;
- stack trace в каждом layer;
- user/order id как logger category;
- PII без retention/access policy.

## 4. Trace context

.NET instrumentation строится вокруг `System.Diagnostics.Activity`. W3C trace context распространяет trace/span ids между services.

```csharp
private static readonly ActivitySource Source = new("Orders.Application");

using Activity? activity = Source.StartActivity("Order.Pay");
activity?.SetTag("order.status", order.Status.ToString());
```

Не записывайте unique order id как metric label. В trace tag он допустим по privacy/sampling policy, потому что trace — отдельная event-like запись, но volume/PII всё равно контролируются.

## 5. Metrics instruments

```csharp
private static readonly Meter Meter = new("Orders.Application");
private static readonly Counter<long> Created =
    Meter.CreateCounter<long>("orders.created");
private static readonly Histogram<double> Duration =
    Meter.CreateHistogram<double>("orders.processing.duration", "s");
```

Основные виды:

- Counter — монотонное количество событий;
- UpDownCounter — может расти/уменьшаться;
- Histogram — distribution measurements;
- ObservableGauge — текущее наблюдаемое значение.

Prometheus exposition имеет counters, gauges, histograms, summaries; OpenTelemetry mapping/export details зависят от exporter.

## 6. Cardinality

Плохие labels:

```text
order_id, user_id, full_url, exception_message
```

Они создают почти уникальную series на событие и перегружают storage/query.

Хорошие bounded dimensions:

```text
http.method, route template, status class,
operation, result, dependency
```

Даже комбинация bounded labels может умножиться; оценивайте cardinality budget.

## 7. Histogram и percentiles

Histogram распределяет observations по buckets и позволяет aggregation across instances, rate и server-side quantiles. Buckets выбирают вокруг SLO, иначе p95/p99 будут грубыми.

Summary/локально вычисленные quantiles обычно хуже агрегируются между replicas. Не усредняйте percentiles.

## 8. OpenTelemetry setup

Концептуально:

```csharp
services.AddOpenTelemetry()
    .WithTracing(tracing => tracing
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddSource("Orders.Application")
        .AddOtlpExporter())
    .WithMetrics(metrics => metrics
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddMeter("Orders.Application")
        .AddPrometheusExporter());
```

Конкретные packages/extensions сверяйте с line 10.x. Instrumentation создаёт telemetry, exporter отправляет backend/collector.

## 9. Sampling

Head sampling принимает решение в начале trace; tail sampling — после получения spans в collector и может сохранить errors/slow traces. Sampling снижает volume, но влияет на статистические выводы и debugging.

Metrics не следует «семплировать как traces» без иной aggregation design.

## 10. Health checks

```csharp
services.AddHealthChecks()
    .AddNpgSql(connectionString, tags: ["ready"]);

app.MapHealthChecks("/health/live", new HealthCheckOptions
{
    Predicate = _ => false
});

app.MapHealthChecks("/health/ready", new HealthCheckOptions
{
    Predicate = registration => registration.Tags.Contains("ready")
});
```

- Liveness: process способен продолжать работу; failure вызывает restart.
- Readiness: instance готов принимать traffic; dependencies/config/warmup могут влиять.

Не делайте liveness зависимой от внешней БД: общий outage вызовет restart storm.

## 11. Prometheus

Prometheus обычно pull-ит exposition endpoint по scrape interval и хранит time series. Counter анализируют через `rate`, histogram — через bucket series и `histogram_quantile`.

Примеры концептуальных запросов:

```promql
sum(rate(http_server_request_duration_seconds_count[5m]))

sum(rate(http_server_request_duration_seconds_bucket{le="0.5"}[5m]))
/
sum(rate(http_server_request_duration_seconds_count[5m]))
```

Второй запрос показывает долю requests ≤ 0.5s при согласованных labels.

## 12. Grafana

Dashboard строится от вопроса:

- соблюдается ли SLO;
- какой route/dependency деградирует;
- есть ли saturation;
- изменилось ли после deploy;
- растёт ли Kafka lag/outbox age.

График без единиц, legend, aggregation и threshold может вводить в заблуждение.

## 13. Alerts

Alert должен быть actionable. Предпочитайте symptoms/user impact:

- error-budget burn;
- sustained high latency;
- queue age/lag;
- saturation.

Один краткий spike не всегда требует paging. Alert содержит owner/runbook/dashboard и severity.

## 14. Overhead

Telemetry потребляет CPU, allocations, network и storage. Источники overhead:

- слишком подробные spans;
- high-cardinality attributes;
- synchronous exporter;
- logging больших bodies;
- слишком частые gauges/callbacks;
- exception as flow.

Измеряйте с instrumentation и без неё на representative load. Используйте batching/sampling/bounds.

## 15. Correlation workflow

Типичный incident:

```text
alert: p99 вырос
 → dashboard: route /orders/{id}, DB saturation
 → exemplar/trace: slow SQL span
 → logs по traceId: concurrency retries
 → query plan/DB metrics: missing index
```

Связь signals важнее количества dashboards.

## 16. Сравнение со Spring

| Spring | ASP.NET Core/.NET |
|---|---|
| Actuator health/metrics | Health Checks + metrics endpoints/exporters |
| Micrometer | `System.Diagnostics.Metrics` + OpenTelemetry metrics |
| Observation/Tracing | Activity/OpenTelemetry tracing |
| MDC | logging scopes/Activity context |
| Prometheus/Grafana | те же backend concepts |

## 17. Production checklist

- Route templates вместо raw URLs в metric labels.
- Trace propagation через HTTP/Kafka headers.
- Redaction/PII policy.
- Liveness и readiness разделены.
- Histograms aligned to SLO.
- Error-budget alerts с runbook.
- Kafka lag/outbox age/DB pool saturation видны.
- Telemetry pipeline имеет own monitoring и bounded queues.

## Самопроверка

1. Почему средняя latency недостаточна?
2. Чем trace tag отличается от metric label по cardinality?
3. Почему liveness не должна падать вместе с БД?
4. Можно ли усреднять p95 разных replicas?
5. Как telemetry сама создаёт overhead?

