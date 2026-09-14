# Семинар 5 — Единый HTTP и error contract

## Задание 1. Матрица responses

Для каждого Order endpoint выпишите success/error statuses. Реализуйте:

- POST: 201 + Location;
- GET missing: 404 Problem Details;
- invalid DTO: 400 validation problem;
- duplicate number: 409;
- successful DELETE: 204.

Проверьте, что 204 не сериализует `null`.

## Задание 2. Domain exception handler

Создайте `OrderNotFoundException`, `OrderConflictException` и handlers. Не используйте exception для malformed input, которое model binding умеет вернуть как 400.

Добавьте extensions:

```csharp
problem.Extensions["traceId"] = context.TraceIdentifier;
problem.Extensions["code"] = "order_conflict";
```

`code` стабилен для клиента, localized `title/detail` — нет.

## Задание 3. Fallback 500

Создайте fallback handler, который логирует exception и возвращает generic Problem Details. В Development детали можно видеть в контролируемом developer UI/log, но публичный body не содержит stack trace.

## Задание 4. ETag

Добавьте `Version` order и:

1. GET возвращает ETag.
2. PUT требует `If-Match`.
3. Несовпадение возвращает 412.
4. Успешный update меняет version.

Объясните race, если check и update не атомарны в repository.

## Задание 5. CORS policy

Разрешите только `https://localhost:5173`, методы GET/POST/PUT и headers Content-Type/Authorization/If-Match. Проверьте preflight через curl с `Origin` и `Access-Control-Request-Method`.

Убедитесь, что origin `https://evil.example` не получает allow-origin.

## Задание 6. OpenAPI completeness

Для каждого endpoint опишите:

- request/response schemas;
- status codes;
- Problem Details;
- bearer security для protected operations;
- pagination query parameters.

Сохраните generated JSON только если repository policy требует artifact; отдельный generated client не нужен.

## Задание 7. Custom response header

Добавьте middleware, возвращающий `X-Request-Duration-Ms`. Не пытайтесь менять headers после начала body; проверьте `Response.HasStarted` для failure diagnostics.

## Задание 8. Ошибочные решения

Исправьте:

```csharp
catch (Exception ex)
{
    return Ok(new { error = ex.ToString() });
}
```

Найдите минимум четыре нарушения: status, data exposure, duplicated local handling, unstable contract/logging.

## Проверка

- единый Problem Details shape;
- stable machine code ошибки;
- trace id коррелирует с log;
- 201/204/401/403/404/409/412 используются корректно;
- CORS policy точечная;
- OpenAPI совпадает с фактическими responses.

