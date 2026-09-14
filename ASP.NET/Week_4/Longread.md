# Лекция 4 — HTTP pipeline, middleware, routing и Web API

## 1. Путь запроса

```text
client
 → reverse proxy/load balancer
 → Kestrel
 → middleware 1 → middleware 2 → routing/auth
 → endpoint (controller/minimal handler)
 ← response через middleware в обратном порядке
```

Middleware образуют ordered pipeline. Каждый компонент может:

- обработать request и вызвать `next`;
- выполнить работу после downstream;
- завершить pipeline досрочно (short-circuit).

## 2. Middleware

```csharp
app.Use(async (context, next) =>
{
    long started = Stopwatch.GetTimestamp();
    try
    {
        await next(context);
    }
    finally
    {
        TimeSpan elapsed = Stopwatch.GetElapsedTime(started);
        context.RequestServices
            .GetRequiredService<ILoggerFactory>()
            .CreateLogger("RequestTiming")
            .LogInformation("{Method} {Path} took {ElapsedMs} ms",
                context.Request.Method,
                context.Request.Path,
                elapsed.TotalMilliseconds);
    }
});
```

Порядок важен: exception handler должен стоять достаточно рано, authentication до authorization, endpoint execution после routing decisions.

Reusable middleware:

```csharp
public sealed class CorrelationMiddleware(
    RequestDelegate next,
    ILogger<CorrelationMiddleware> logger)
{
    public async Task InvokeAsync(HttpContext context)
    {
        string correlationId = context.Request.Headers
            .TryGetValue("X-Correlation-Id", out var value)
                ? value.ToString()
                : Guid.NewGuid().ToString("N");

        context.Response.Headers["X-Correlation-Id"] = correlationId;
        using (logger.BeginScope(new Dictionary<string, object>
        {
            ["CorrelationId"] = correlationId
        }))
        {
            await next(context);
        }
    }
}
```

Входной id нужно валидировать по длине/формату, иначе клиент управляет log field.

## 3. Routing

```csharp
[ApiController]
[Route("api/orders")]
public sealed class OrdersController : ControllerBase
{
    [HttpGet("{id:guid}")]
    public ActionResult Get(Guid id) => Ok();
}
```

Route constraint `:guid` участвует в matching, но не заменяет domain validation.

Route template должен описывать resource hierarchy, а не implementation method names.

## 4. Model binding sources

```csharp
[HttpPut("{id:guid}")]
public async Task<IActionResult> Update(
    [FromRoute] Guid id,
    [FromQuery] bool notify,
    [FromHeader(Name = "If-Match")] string? etag,
    [FromBody] UpdateOrderRequest request,
    CancellationToken ct)
```

Sources:

- route values;
- query string;
- headers;
- body formatters;
- form/files;
- services.

Complex body обычно один. Не смешивайте domain entity с transport DTO.

## 5. `[ApiController]`

Включает API-oriented conventions, включая inference binding sources и automatic 400 для invalid model state. Конкретный error shape настраивается; не полагайтесь на неявный default как на вечный public contract.

## 6. DTO и over-posting

❌ Принимать EF entity:

```csharp
public Task Create(Order entity) // клиент может прислать внутренние поля
```

✅ Отдельный input DTO:

```csharp
public sealed record CreateOrderRequest(
    [Required, StringLength(40)] string Number,
    [Range(0.01, 1_000_000)] decimal Total);
```

DTO задаёт wire contract. Domain model повторно защищает business invariant, потому что command может прийти не только из HTTP.

## 7. Validation

Уровни:

1. syntactic/binding — JSON и types;
2. DTO validation — required/range/format;
3. application/domain rules — уникальность, переход состояния, права;
4. database constraints — последняя защита consistency.

Async database validation не стоит помещать в property attribute. Выполняйте её в use case с cancellation и race-safe database constraint.

## 8. Request body и buffering

Body обычно forward-only stream. Model binder читает его один раз. Middleware, которое читает body для logging, должно осознанно включать buffering, ограничивать размер и не логировать secrets/files.

## 9. File upload

Small buffered upload:

```csharp
[HttpPost("{id:guid}/attachment")]
[RequestSizeLimit(10 * 1024 * 1024)]
public async Task<IActionResult> Upload(
    Guid id,
    IFormFile file,
    CancellationToken ct)
{
    if (file.Length == 0) return BadRequest();

    await using Stream input = file.OpenReadStream();
    await storage.SaveAsync(id, input, ct);
    return NoContent();
}
```

Безопасность:

- лимит размера;
- generated storage name;
- content scanning при необходимости;
- не доверять `FileName` и `ContentType`;
- хранить вне web root;
- streaming для больших files.

## 10. Cookies и session

Cookie хранится у клиента и отправляется с requests по domain/path/SameSite/Secure rules. Не помещайте sensitive data в незашифрованную cookie.

Session обычно хранит id в cookie, data — server-side/cache. Session создаёт stateful concerns: distributed cache, affinity, expiry, privacy. Для REST API чаще используют explicit tokens/resources, но session допустима для определённых apps.

## 11. Request cancellation

`CancellationToken` action parameter связан с request aborted token. Передавайте его в application, DB и HTTP calls. Client disconnect не гарантирует rollback side effects.

## 12. Minimal APIs

```csharp
app.MapPost("/api/orders", async (
    CreateOrderRequest request,
    OrderService service,
    CancellationToken ct) =>
{
    OrderDto created = await service.CreateAsync(request, ct);
    return TypedResults.Created($"/api/orders/{created.Id}", created);
});
```

Endpoint filters дают local cross-cutting behavior. Для large controller-oriented course дальше используется MVC, но principles одинаковы.

## 13. Сравнение со Spring MVC

| Spring MVC | ASP.NET Core |
|---|---|
| servlet container | Kestrel/reverse proxy stack |
| DispatcherServlet | endpoint routing + MVC middleware/services |
| servlet filters/interceptors | middleware/MVC filters |
| `@RestController` | `[ApiController] : ControllerBase` |
| `@PathVariable` | `[FromRoute]` |
| `@RequestParam` | `[FromQuery]` |
| `@RequestBody` | `[FromBody]` |
| `MultipartFile` | `IFormFile` |

## 14. Best practices

- Pipeline order фиксируйте и тестируйте.
- Thin controllers переводят HTTP в application command/result.
- DTO не равен entity.
- Ограничивайте body/file sizes.
- Не логируйте body по умолчанию.
- Validation разделяйте по слоям.
- Cancellation передавайте, но consistency решайте transaction/idempotency.

## Самопроверка

1. Почему response проходит middleware в обратном порядке?
2. Чем route constraint отличается от validation?
3. Почему entity опасна как input model?
4. Можно ли читать request body несколько раз автоматически?
5. Что гарантирует request cancellation?

