# Лекция 12 — Async HTTP, streaming, Kestrel и cancellation

## 1. Request concurrency model

Kestrel принимает connections и обрабатывает I/O асинхронно. ASP.NET Core не выделяет гарантированный отдельный thread на весь request. Continuation после `await` может выполняться другим ThreadPool thread.

```text
request → sync work → await DB
                     thread свободен
          DB completion → continuation → response
```

## 2. Scalability async I/O

```csharp
[HttpGet("{id:guid}")]
public async Task<ActionResult<OrderResponse>> Get(Guid id, CancellationToken ct)
{
    OrderResponse? order = await queries.FindAsync(id, ct);
    return order is null ? NotFound() : Ok(order);
}
```

Преимущество — освобождение thread во время I/O. Если provider на деле синхронно блокирует, async keyword сам не исправит ситуацию.

## 3. ThreadPool starvation

Причины:

- `.Result`/`.Wait()`;
- sync database/file/network APIs под нагрузкой;
- долгий CPU work на request threads;
- lock contention;
- unbounded fan-out.

Симптомы: растут queue length, thread count и latency, CPU может не быть полностью занят.

## 4. Cancellation

Action token связан с `HttpContext.RequestAborted`:

```csharp
await db.Orders.SingleOrDefaultAsync(x => x.Id == id, ct);
await client.SendAsync(request, ct);
```

Client disconnect — сигнал остановиться, но:

- network/proxy может сообщить не сразу;
- side effect уже мог commit;
- cancellation cooperative;
- cleanup должен всё равно выполниться.

## 5. CPU-bound work

Большой image/report calculation блокирует CPU независимо от `async`.

Варианты:

- оптимизировать;
- bounded parallel processing;
- вынести в background queue/service;
- вернуть 202 + operation resource;
- масштабировать workers.

`Task.Run` внутри controller только переносит работу в тот же process ThreadPool и часто ухудшает capacity.

## 6. Streaming response

Async sequence endpoint:

```csharp
[HttpGet("stream")]
public IAsyncEnumerable<OrderResponse> Stream(CancellationToken ct) =>
    queries.StreamAsync(ct);
```

Serialization может выдавать JSON array постепенно, но proxy/client buffering и serializer behavior влияют на perceived streaming.

NDJSON/SSE задают другие wire contracts. Выбирайте content type и framing явно.

## 7. Server-Sent Events

SSE — однонаправленный stream server → browser по HTTP:

```text
event: order-status
data: {"id":"...","status":"Paid"}

```

Требуются heartbeat, disconnect cancellation, bounded subscriber buffers, replay/last-event-id policy и proxy timeouts.

Для двунаправленного real-time взаимодействия подходит SignalR/WebSocket stack.

## 8. Request streaming

Большой upload обрабатывайте chunk-by-chunk:

```csharp
await storage.SaveAsync(Request.Body, Request.ContentLength, ct);
```

Необходимо:

- server/proxy body limits;
- content-length/chunked handling;
- temp/quota policy;
- hash/virus scan;
- partial cleanup при cancellation;
- не доверять content type/name.

## 9. Backpressure boundaries

Task-based async не является Reactive Streams protocol. `IAsyncEnumerable<T>` — pull-based: consumer запрашивает следующий element. Но между broker → channel → endpoint могут быть buffers.

Bounded `Channel<T>` задаёт явную capacity/full-mode. При медленном client producer должен wait/drop/disconnect по policy, иначе memory растёт.

## 10. Response has started

После отправки headers/status нельзя безопасно заменить response на Problem Details. Ошибка в середине stream приводит к оборванному payload/stream-specific error event.

До начала:

```csharp
if (!Response.HasStarted) { /* map error */ }
```

Streaming contract должен описывать partial failure.

## 11. Timeouts и deadlines

Есть request timeout, reverse-proxy timeout, outbound timeout, database command timeout. Несогласованные значения дают «загадочные» отмены.

Deadline budget должен уменьшаться по call chain. Retry attempts помещаются внутри общего budget.

## 12. Kestrel limits

Настраиваются request body, headers, rates/timeouts и connections. Значения согласуются с reverse proxy. Слишком широкие limits повышают DoS risk, слишком узкие ломают legit clients.

## 13. Async disposal

Resources живут до завершения streaming enumeration. `await using`/iterator finally должны сработать при normal completion, exception и cancellation.

## 14. Сравнение с WebFlux/Reactor

| Spring WebFlux | ASP.NET Core |
|---|---|
| Reactor `Mono`/`Flux` | `Task<T>`/`IAsyncEnumerable<T>` |
| Reactive Streams demand | async pull и отдельные bounded buffers; не тот же protocol |
| Netty event loop | Kestrel + async sockets + ThreadPool continuations |
| `WebClient` | `HttpClient` |

ASP.NET Core обычный async stack масштабируется без обязательной reactive operator model. Для streams/backpressure всё равно требуется design.

## 15. Best practices

- Async all the way для I/O.
- No blocking waits.
- Cancellation propagated.
- Bounded concurrency/buffers.
- CPU work вне latency-sensitive request.
- Streaming errors спроектированы до headers start.
- Load test с медленными clients и disconnects.

## Самопроверка

1. Привязан ли request к одному thread?
2. Почему `Task.Run` не решает server CPU capacity?
3. Что происходит при exception после начала response?
4. Является ли `IAsyncEnumerable` Reactive Streams?
5. Где нужен bounded buffer?

