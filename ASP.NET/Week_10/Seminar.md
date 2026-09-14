# Семинар 10 — Устойчивый HTTP-клиент

## Подготовка

Создайте локальный fake Payment API с endpoints: success, delayed, 429 + Retry-After, 500, invalid JSON и idempotent POST store.

## Задание 1. Typed client

Настройте BaseAddress, timeout, User-Agent, JSON DTO. Не создавайте `HttpClient` в method.

## Задание 2. Query/path/header

Соберите URI через encoding helper, добавьте `X-Correlation-Id` на request. Проверьте special characters.

## Задание 3. Bearer delegating handler

Создайте token provider и handler. Token должен быть per-request/appropriately cached, не static default header. Не логируйте его.

## Задание 4. Status mapping

Отобразите:

- 404 → nullable/not-found result;
- 409 → `PaymentConflictException`;
- 429 → transient rate-limit result с Retry-After;
- other non-success → bounded Problem Details parsing + fallback.

## Задание 5. Timeout classification

Проверьте caller cancellation и policy timeout отдельно. В logs/metrics причины должны различаться.

## Задание 6. Retry GET

Fake endpoint два раза даёт 503, затем 200. Настройте 3 attempts, exponential delay + jitter. Зафиксируйте attempt count.

## Задание 7. Не retry 400

Докажите одним test/metric, что validation 400 выполняется один раз.

## Задание 8. Idempotent POST

Сгенерируйте idempotency key один раз на logical operation и reuse при retries. Fake server атомарно возвращает исходный result. Повтор с иным payload → 409.

## Задание 9. Circuit breaker

Вызовите failing endpoint до Open, убедитесь в fast rejection, дождитесь HalfOpen и восстановите dependency. Логируйте transitions без high-cardinality state.

## Задание 10. Inbound rate limit

Добавьте policy на `/auth/login`: 5/min, queue 0, 429 + Retry-After. Объясните ограничение per-instance memory limiter при нескольких replicas.

## Задание 11. Concurrency limiter

Ограничьте 3 одновременных payment calls и маленькую очередь. Load test должен показать controlled rejection вместо неограниченного роста latency.

## Проверка

- client factory-managed;
- auth header не утёк;
- 400 не retry;
- POST retry имеет idempotency;
- timeout/cancellation различаются;
- breaker transitions наблюдаемы;
- очереди bounded;
- error body чтение ограничено.

