# Лекция 10 — `HttpClientFactory`, resilience и rate limiting

## 1. Почему не `new HttpClient()` на каждый запрос

Частое создание/disposal handlers/connections может вызвать socket exhaustion и потерю pooling. Один вечный client без handler rotation может плохо реагировать на DNS changes. `IHttpClientFactory` управляет handler lifetime и configuration.

## 2. Typed client

```csharp
public sealed class PaymentClient(HttpClient httpClient)
{
    public async Task<PaymentResponse> CreateAsync(
        PaymentRequest request,
        CancellationToken ct)
    {
        using HttpResponseMessage response = await httpClient.PostAsJsonAsync(
            "payments",
            request,
            cancellationToken: ct);

        if (response.StatusCode == HttpStatusCode.Conflict)
            throw new PaymentConflictException();

        response.EnsureSuccessStatusCode();

        return await response.Content.ReadFromJsonAsync<PaymentResponse>(ct)
            ?? throw new InvalidDataException("Empty payment response");
    }
}
```

Registration:

```csharp
services.AddHttpClient<PaymentClient>((sp, client) =>
{
    var options = sp.GetRequiredService<IOptions<PaymentOptions>>().Value;
    client.BaseAddress = new Uri(options.BaseUrl);
    client.Timeout = TimeSpan.FromSeconds(options.TimeoutSeconds);
    client.DefaultRequestHeaders.UserAgent.ParseAdd("orders-service/1.0");
});
```

## 3. URI и encoding

Не склеивайте query user input строкой:

```csharp
string uri = QueryHelpers.AddQueryString(
    "payments",
    new Dictionary<string, string?> { ["customerId"] = customerId });
```

Path segments тоже требуют escaping/route contract. BaseAddress trailing slash и relative URI resolution имеют правила `System.Uri`, которые надо тестировать.

## 4. Headers и authentication

Per-request dynamic token:

```csharp
using var request = new HttpRequestMessage(HttpMethod.Get, "payments/123");
request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", token);
```

Лучше delegating handler, получающий short-lived token и добавляющий его. Не задавайте user-specific token в `DefaultRequestHeaders` shared client.

## 5. Error body

Ограничивайте размер чтения/логирования error response. Попытайтесь распарсить Problem Details, но сохраняйте fallback для invalid/unexpected content type.

Не вызывайте `EnsureSuccessStatusCode` до custom mapping, если нужно различить 404/409/429.

## 6. Timeout и cancellation

Различайте:

- caller cancellation;
- client timeout;
- connection timeout;
- server 408;
- proxy timeout/status.

Единый `TaskCanceledException` без context недостаточен для SLO diagnostics. Используйте policy telemetry и linked deadlines.

## 7. Retry

Retry допустим для transient failure и idempotent/replay-safe operation.

Безопаснее:

- GET/HEAD при transient transport/5xx/429 с policy;
- PUT/DELETE если API/idempotency contract это гарантирует;
- POST только с idempotency key/server deduplication.

Используйте exponential backoff + jitter, `Retry-After`, общий deadline и ограничение attempts.

Retry не исправляет validation 400, auth 401/403 или стабильный 404.

## 8. Circuit breaker

States conceptually:

```text
Closed --failure threshold--> Open
Open --break duration--> HalfOpen
HalfOpen --probe success--> Closed
HalfOpen --probe failure--> Open
```

Breaker защищает caller/resources от постоянного медленного failure и даёт dependency восстановиться. Он не делает ответ успешным: нужен fallback/error propagation.

Policy scope/key важен. Один breaker на все hosts/tenants может создать слишком большой blast radius.

## 9. Rate limiter

Rate limiting защищает capacity и fairness. Алгоритмы:

- fixed window;
- sliding window;
- token bucket;
- concurrency limiter.

Inbound ASP.NET limiter:

```csharp
services.AddRateLimiter(options =>
{
    options.AddFixedWindowLimiter("login", limiter =>
    {
        limiter.PermitLimit = 5;
        limiter.Window = TimeSpan.FromMinutes(1);
        limiter.QueueLimit = 0;
    });
});

app.UseRateLimiter();
```

Distributed replicas требуют shared/coordinated limit или осознанный per-instance budget.

## 10. Bulkhead/concurrency limit

Даже при низком RPS медленные requests могут исчерпать connections. Concurrency limiter ограничивает одновременно выполняемые operations и очередь.

Bounded queue + 429/503 лучше unbounded memory growth. Выбор status/retry-after зависит от boundary.

## 11. Resilience pipeline

Современный .NET Http resilience stack строится вокруг resilience handlers/pipelines (Polly integration). Порядок policies важен:

- общий timeout;
- retry;
- attempt timeout;
- circuit breaker;
- rate/concurrency limiter.

Не копируйте порядок механически: определите, что считается одной попыткой и что учитывает breaker.

## 12. Idempotency key

```http
POST /payments
Idempotency-Key: 8f...
```

Server атомарно связывает key + operation scope + request hash с результатом. Повтор с тем же key и другим payload — conflict. TTL/storage/race handling являются частью contract.

## 13. Observability

Для outbound calls измеряйте:

- dependency name/route template;
- latency;
- result/status category;
- attempt count;
- timeout/cancellation reason;
- breaker state transitions;
- rate limit rejections.

Не используйте raw URL с IDs как metric label — cardinality explosion.

## 14. Сравнение со Spring

| Spring | .NET |
|---|---|
| RestClient/WebClient | HttpClient/typed clients |
| interceptor | DelegatingHandler |
| Resilience4j | resilience pipelines/Polly ecosystem |
| rate limiter filter | ASP.NET Core rate limiting middleware |
| circuit breaker annotation | configured pipeline/decorator |

## 15. Best practices

- Factory-managed clients.
- Typed DTO и explicit status mapping.
- Cancellation/deadline передаются.
- Retry только replay-safe operations.
- Jitter и `Retry-After`.
- Ограниченная очередь/concurrency.
- Metrics labels low-cardinality.
- Secrets/bodies redacted.

## Самопроверка

1. Почему вечный manual handler может иметь DNS проблему?
2. Когда POST можно retry?
3. Что делает half-open state?
4. Чем rate limit отличается от concurrency limit?
5. Почему raw URL плох как metric label?

