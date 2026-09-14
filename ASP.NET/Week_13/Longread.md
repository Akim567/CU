# Лекция 13 — Асинхронная БД, background processing и согласованность

## 1. Async EF не делает БД быстрее

```csharp
Order? order = await db.Orders
    .SingleOrDefaultAsync(x => x.Id == id, ct);
```

Async освобождает caller thread во время network/database wait. Query plan, locks, indexes и server capacity остаются теми же.

## 2. Один DbContext — одна operation at a time

```csharp
Task<Order?> first = db.Orders.FirstOrDefaultAsync(...);
Task<int> second = db.Payments.CountAsync(...);
await Task.WhenAll(first, second); // один context использовать concurrent нельзя
```

Исправления:

- await последовательно;
- объединить в один query;
- отдельные contexts через factory, если concurrency действительно выгодна;
- не нарушать transaction consistency.

## 3. Cancellation и database state

Отмена command не означает, что server ничего не выполнил. Между commit и потерей response возникает **unknown outcome**. Client retry без idempotency может повторить side effect.

Для commands нужны natural/business key, idempotency record или retriable transaction semantics.

## 4. Background queue

HTTP request не должен fire-and-forget `Task`:

```csharp
_ = ProcessAsync(); // exception/lifetime/shutdown потеряны
```

In-process queue:

```csharp
public interface IBackgroundTaskQueue
{
    ValueTask EnqueueAsync(OrderJob job, CancellationToken ct);
    IAsyncEnumerable<OrderJob> ReadAllAsync(CancellationToken ct);
}
```

Реализация на bounded `Channel<T>`, worker — `BackgroundService`, scope — на job.

## 5. 202 Accepted

Для долгой операции:

```http
POST /reports
→ 202 Accepted
Location: /operations/{id}
```

Operation resource содержит Pending/Running/Succeeded/Failed/Cancelled, progress/result/error. 202 не обещает, что работа успешно завершится.

## 6. Durable queue

In-memory channel теряет jobs при crash/deploy и не делится между replicas. Durable alternatives:

- database jobs table с locking/lease;
- Kafka/RabbitMQ/cloud queue;
- scheduler/job framework.

Нужны retry, visibility/lease, poison-message policy, idempotency и monitoring.

## 7. Outbox problem

```text
transaction commits order
process crashes before broker publish
→ DB говорит Paid, event отсутствует
```

Или publish проходит, DB rollback — event лжёт.

Transactional outbox записывает domain change и outbox row в одной DB transaction:

```text
BEGIN
UPDATE orders ...
INSERT outbox(id, type, payload, occurred_at, status)
COMMIT
```

Отдельный publisher отправляет row и помечает processed. Publish может повториться, поэтому consumers idempotent.

## 8. Outbox schema

Минимальные fields:

- stable message id;
- aggregate id/type;
- event type/version;
- occurred time;
- payload;
- attempt/next-at/error;
- processed/lease info.

Не храните arbitrary runtime type name как вечный public event contract.

## 9. Polling и leasing

Несколько workers выбирают batch без двойной одновременной обработки через PostgreSQL locking (`FOR UPDATE SKIP LOCKED`) или lease update. Lease имеет expiry для crash recovery.

Даже при locking возможен повтор: broker accepted, worker crashed до mark processed.

## 10. Inbox/idempotent consumer

Consumer в одной transaction:

1. пытается вставить message id в inbox с unique key;
2. если уже есть — duplicate ignored;
3. применяет domain change;
4. commit;
5. затем подтверждает broker delivery.

Inbox не решает неправильный business ordering/versioning автоматически.

## 11. Retries

Классифицируйте:

- transient database/network;
- optimistic concurrency;
- serialization failure;
- permanent validation/schema error;
- poison payload.

Retry entire idempotent unit, bounded + backoff/jitter. Permanent отправляется в failed/DLQ process с alert/replay controls.

## 12. Schedulers и overlap

Periodic service может запустить новую итерацию до завершения старой. Определите:

- skip overlap;
- serialize;
- allow bounded concurrency;
- distributed leader/lease;
- catch-up missed schedules.

Timer сам не обеспечивает distributed singleton.

## 13. Graceful shutdown

Worker прекращает fetch, завершает/abandon current job, освобождает lease, закрывает scope. Broker ack/DB commit order должен предотвращать loss, принимая возможные duplicates.

## 14. Сравнение с R2DBC/WebFlux

EF Core/Npgsql async используют Task-based async I/O, а не Reactive Streams/R2DBC API. JDBC blocking mismatch, обсуждаемый в Spring WebFlux, не переносится буквально: ADO.NET providers имеют native async APIs, но конкретный provider должен реализовать их качественно.

Общее: нельзя держать transaction через внешнее ожидание; cancellation, pooling и backpressure требуют design.

## 15. Best practices

- One context operation at a time.
- No fire-and-forget request work.
- Durable queue для обязательной работы.
- Outbox + idempotent consumers.
- Bounded batch/concurrency/retry.
- Unknown outcome рассматривается явно.
- Metrics: queue depth, age, attempts, failures, throughput.

## Самопроверка

1. Ускоряет ли async плохой SQL plan?
2. Почему cancellation может дать unknown outcome?
3. Что гарантирует outbox?
4. Почему outbox publisher всё равно может отправить duplicate?
5. Чем in-memory channel уступает durable queue?

