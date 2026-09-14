# Семинар 4 — Middleware, controllers, binding и validation

## Задание 1. Correlation middleware

Принимайте `X-Correlation-Id`, но разрешайте только 1–64 символа `[A-Za-z0-9._-]`. При invalid value создавайте новый id. Верните его в response header и logging scope.

## Задание 2. Timing middleware

Измеряйте request через `Stopwatch.GetTimestamp`, логируйте method, path template/endpoint display name, status и elapsed. Не включайте query целиком: там могут быть tokens.

## Задание 3. CRUD controller

Реализуйте:

```text
POST   /api/orders
GET    /api/orders/{id:guid}
GET    /api/orders?status=Paid&page=1&pageSize=20
PUT    /api/orders/{id:guid}
DELETE /api/orders/{id:guid}
```

Используйте request/response DTO. Ограничьте `pageSize` диапазоном 1–100.

## Задание 4. Validation layers

DTO rules:

- number обязателен, длина 3–40;
- total 0.01–1,000,000;
- хотя бы одна line.

Domain rules:

- нельзя менять paid order;
- number unique;
- line quantity positive.

Добавьте unique check в repository, но объясните, почему без DB unique constraint остаётся race.

## Задание 5. Binding sources

Добавьте endpoint:

```csharp
[HttpPatch("{id:guid}/status")]
public Task<IActionResult> ChangeStatus(
    [FromRoute] Guid id,
    [FromHeader(Name = "If-Match")] string? version,
    [FromBody] ChangeStatusRequest request,
    CancellationToken ct)
```

Проверьте отсутствие header, invalid enum и malformed JSON как разные cases.

## Задание 6. File upload

Добавьте attachment до 5 MiB. Storage filename — generated GUID, original name хранится только как metadata после нормализации. Проверьте empty file, превышение лимита и cancellation.

## Задание 7. Cookie и session experiment

Создайте два отдельных endpoints:

- безопасная preference cookie с `HttpOnly`, `Secure`, `SameSite`;
- session counter.

Объясните, где физически находится state в каждом варианте и почему session потребует distributed store при нескольких replicas.

## Задание 8. Pipeline order

Временно поставьте timing/correlation middleware после endpoints так, чтобы часть запросов его не проходила. Зафиксируйте поведение и исправьте порядок.

## Проверка

- correlation id присутствует в response/log;
- invalid model получает стабильный 400;
- unknown resource — 404, conflict — 409;
- transport DTO отделён от domain;
- файлы имеют size limit и безопасное имя;
- request cancellation передаётся вниз;
- middleware order проверен минимум одним integration test позднее.

