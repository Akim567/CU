# Семинар 15 — OpenTelemetry, Prometheus и operational readiness

## Задание 1. Structured logs

Добавьте stable events для create/pay/cancel. Fields: operation, order status, result; ids — в log scope, secrets/DTO body отсутствуют.

## Задание 2. Trace propagation

Настройте ASP.NET Core и HttpClient instrumentation. Вручную создайте Activity для application use case. Проверьте один trace через inbound → DB/outbound → response.

## Задание 3. Kafka context

Inject W3C trace context в Kafka headers, extract у consumer и создайте consumer span. Некорректные headers не должны ломать processing.

## Задание 4. Custom metrics

Создайте:

- counter `orders.created`;
- counter `orders.processing.failures` с bounded `reason`;
- histogram duration;
- observable gauge channel depth;
- histogram outbox message age.

Никаких order/user/message ids в tags.

## Задание 5. Cardinality audit

Найдите и исправьте labels: raw path, exception message, customer id, SQL text. Оцените верхнюю границу series комбинаций.

## Задание 6. Health checks

Создайте `/health/live` без external dependencies и `/health/ready` с PostgreSQL + обязательным startup state. Kafka readiness добавляйте только если невозможность publish/consume действительно означает «не принимать traffic» для роли process.

## Задание 7. Prometheus

Подключите exporter/scrape. Постройте queries для RPS, error rate, доли requests быстрее SLO и p95 histogram.

## Задание 8. Grafana dashboard

Панели:

- traffic/error/latency;
- DB duration/pool saturation;
- outbound dependency;
- Kafka lag;
- outbox age;
- channel depth;
- deployment annotation/version.

У каждой панели units, legend и useful aggregation.

## Задание 9. Alerts

Создайте три правила:

1. sustained error-budget burn;
2. outbox oldest age выше threshold;
3. readiness failure/saturation.

Добавьте owner и runbook action. Избегайте paging на единичный краткий spike.

## Задание 10. Overhead experiment

Под load сравните telemetry off/on и aggressive body logging. Измерьте CPU, allocations, throughput, exporter queue/drop. Удалите body logging и настройте batching/sampling.

## Финальный production drill

Искусственно замедлите DB query и пройдите цепочку:

1. alert;
2. dashboard;
3. trace;
4. correlated logs;
5. query plan;
6. fix;
7. подтверждение восстановления SLO.

## Проверка

- signals коррелируют trace id;
- metrics low-cardinality;
- liveness/readiness разделены;
- SLO выражен histogram buckets/query;
- alerts actionable;
- telemetry overhead измерен;
- secrets/PII не экспортируются.

