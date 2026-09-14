# Лекция 5 — HTTP-ответы, Problem Details, CORS и OpenAPI

## 1. Response contract

HTTP response состоит из status line, headers и optional body. API contract должен согласовывать их, а не возвращать `200` с текстом ошибки.

| Сценарий | Типичный status |
|---|---:|
| успешный GET/PUT | 200 |
| создан resource | 201 + `Location` |
| успешно без body | 204 |
| malformed/invalid input | 400 |
| нет authentication | 401 |
| недостаточно прав | 403 |
| resource отсутствует | 404 |
| conflict/invariant | 409 |
| precondition failed | 412 |
| rate limit | 429 |
| unexpected server error | 500 |

Выбор зависит от contract; таблица не заменяет API design.

## 2. `ActionResult<T>`

```csharp
[HttpGet("{id:guid}")]
public async Task<ActionResult<OrderResponse>> Get(Guid id, CancellationToken ct)
{
    Order? order = await service.FindAsync(id, ct);
    if (order is null) return NotFound();
    return Ok(OrderResponse.From(order));
}
```

`ActionResult<T>` выражает typed success body и альтернативные HTTP results.

## 3. Created и Location

```csharp
return CreatedAtAction(
    nameof(Get),
    new { id = created.Id },
    created);
```

Client может использовать `Location`. Route/action generation лучше string concatenation.

## 4. Headers, caching и concurrency

```csharp
Response.Headers.ETag = $"\"{order.Version}\"";
```

ETag + `If-Match` реализуют optimistic HTTP precondition. Это не автоматически EF concurrency token: application связывает два contracts.

Не кэшируйте персональные responses публично. Cache-Control, Vary, authorization и reverse proxies должны рассматриваться вместе.

## 5. Problem Details

RFC 9457-style Problem Details предоставляет machine-readable error shape:

```json
{
  "type": "https://example.test/problems/order-conflict",
  "title": "Order conflict",
  "status": 409,
  "detail": "The order is already paid",
  "instance": "/api/orders/...",
  "traceId": "..."
}
```

Registration/handler:

```csharp
builder.Services.AddProblemDetails();

app.UseExceptionHandler();
app.UseStatusCodePages();
```

Для domain mapping используйте `IExceptionHandler`:

```csharp
public sealed class DomainExceptionHandler(
    IProblemDetailsService problems) : IExceptionHandler
{
    public async ValueTask<bool> TryHandleAsync(
        HttpContext context,
        Exception exception,
        CancellationToken ct)
    {
        if (exception is not OrderConflictException conflict)
            return false;

        context.Response.StatusCode = StatusCodes.Status409Conflict;
        return await problems.TryWriteAsync(new ProblemDetailsContext
        {
            HttpContext = context,
            ProblemDetails = new ProblemDetails
            {
                Status = 409,
                Title = "Order conflict",
                Detail = conflict.PublicMessage,
                Type = "https://example.test/problems/order-conflict"
            }
        });
    }
}
```

Не раскрывайте stack trace, SQL, connection strings и внутренние ids.

## 6. Exception boundary

Handlers располагаются от specific к fallback. Unexpected exception:

- логируется один раз с trace context;
- выдаёт generic 500;
- не поглощает process-corrupting state;
- не маскирует cancellation как server error.

Validation errors могут формироваться без exceptions через model state/result types.

## 7. MVC filters

Filters работают вокруг MVC stages/actions:

- authorization filters (редко custom вместо policies);
- resource/action/result/exception filters.

Выбор:

- middleware — concern любого HTTP endpoint;
- filter — MVC-specific action metadata/context;
- decorator — application service concern.

Exception middleware охватывает больше pipeline, чем MVC exception filter.

## 8. CORS

CORS — browser policy для cross-origin requests, не authentication и не firewall.

```csharp
services.AddCors(options =>
{
    options.AddPolicy("frontend", policy =>
        policy.WithOrigins("https://app.example.test")
              .WithMethods("GET", "POST", "PUT", "DELETE")
              .WithHeaders("Content-Type", "Authorization"));
});

app.UseCors("frontend");
```

Credentials нельзя безопасно сочетать с wildcard origin. Origin comparison и proxy scheme/host должны быть настроены корректно.

Preflight `OPTIONS` проверяет permission before non-simple request. Он не вызывает business endpoint.

## 9. OpenAPI

OpenAPI описывает paths, operations, parameters, schemas, responses и security schemes. Для ASP.NET Core добавьте соответствующие built-in/OpenAPI services выбранной версии и map document endpoint.

Описание должно включать все значимые responses:

```csharp
[ProducesResponseType<OrderResponse>(StatusCodes.Status200OK)]
[ProducesResponseType<ProblemDetails>(StatusCodes.Status404NotFound)]
[HttpGet("{id:guid}")]
public Task<ActionResult<OrderResponse>> Get(...) { }
```

Generated document нужно contract-test-ить: attributes и inferred metadata могут расходиться с реальным handler behavior.

## 10. API evolution

- Добавление optional response field обычно совместимо для tolerant clients.
- Удаление/переименование field — breaking.
- Изменение meaning или error status может быть breaking даже без schema change.
- Enum expansion ломает clients с exhaustive switch.
- Pagination/order defaults являются contract.

Versioning strategy выбирается заранее: URL/header/media type или эволюция без формальных versions.

## 11. Сравнение со Spring

| Spring | ASP.NET Core |
|---|---|
| `ResponseEntity<T>` | `ActionResult<T>`/`IResult` |
| `@ControllerAdvice` | exception middleware + `IExceptionHandler`/ProblemDetails |
| `@ExceptionHandler` | typed exception handler/filter |
| CORS config | CORS middleware/policies |
| springdoc/OpenAPI | ASP.NET Core OpenAPI ecosystem/built-in services |

## 12. Best practices

- Один стабильный error media type/shape.
- Public detail не содержит internals.
- 401 и 403 не смешиваются.
- 201 имеет корректный `Location`.
- 204 не содержит body.
- CORS разрешает точные origins/methods/headers.
- OpenAPI проверяется как artifact, а не считается автоматически истинным.

## Самопроверка

1. Чем 401 отличается от 403?
2. Что должен содержать 201 response?
3. Почему CORS не защищает API от server-to-server client?
4. Когда middleware лучше exception filter?
5. Может ли schema-compatible изменение быть behavioral breaking change?

