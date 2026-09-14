# Конспект лекции 20 — Многопоточность и конкурентный доступ

## 1. Термины

- **Concurrency** — несколько задач прогрессируют в пересекающееся время.
- **Parallelism** — вычисления физически выполняются одновременно.
- **Asynchrony** — caller не блокируется в ожидании завершения.
- **Thread safety** — корректность при допустимом concurrent access.

Async I/O может быть concurrent без parallel CPU execution. CPU-bound работу параллелят осознанно.

## 2. Race condition

```csharp
private int _count;

public void Increment()
{
    _count++; // read-modify-write, не атомарный contract
}
```

Два потока могут прочитать одно значение и потерять increment.

```csharp
Interlocked.Increment(ref _count);
```

`Interlocked` предоставляет атомарные операции для простых state transitions.

## 3. `lock`

```csharp
private readonly object _gate = new();
private readonly Dictionary<Guid, Order> _orders = [];

public void Add(Order order)
{
    lock (_gate)
    {
        _orders.Add(order.Id, order);
    }
}
```

Все accesses к защищаемому invariant должны использовать тот же gate. Не lock-айте `this`, public object или interned string — внешний код тоже может их захватить.

Не выполняйте долгий I/O и `await` внутри обычного `lock`.

## 4. Monitor и новые lock abstractions

`lock` компилируется в корректный acquire/release pattern (для обычного object — через `Monitor`). Современные версии .NET/C# также оптимизируют работу со специализированным `System.Threading.Lock`. Используйте language statement, а не ручной `Enter` без `finally`.

## 5. `SemaphoreSlim`

Ограничивает concurrency и поддерживает async wait:

```csharp
private readonly SemaphoreSlim _gate = new(initialCount: 1);

public async Task UpdateAsync(CancellationToken ct)
{
    await _gate.WaitAsync(ct);
    try
    {
        await PerformUpdateAsync(ct);
    }
    finally
    {
        _gate.Release();
    }
}
```

Это async-compatible mutual exclusion внутри process, но не distributed lock и не database transaction.

## 6. Volatile и memory ordering

`volatile` влияет на visibility/order отдельных reads/writes, но не делает составные operations атомарными.

```csharp
private volatile bool _stopping;
```

Для coordination чаще предпочитайте cancellation, locks, interlocked operations или concurrent primitives. Memory model трудно проверять «на глаз».

## 7. Concurrent collections

```csharp
private readonly ConcurrentDictionary<Guid, Lazy<Task<User>>> _cache = new();

public Task<User> GetAsync(Guid id) =>
    _cache.GetOrAdd(
        id,
        key => new Lazy<Task<User>>(() => LoadAsync(key))).Value;
```

Даже concurrent collection не делает atomic весь многошаговый business process. Методы вроде `GetOrAdd`, `TryUpdate`, `AddOrUpdate` задают конкретные atomic boundaries; factories могут выполняться несколько раз.

## 8. Channels

```csharp
Channel<Job> channel = Channel.CreateBounded<Job>(new BoundedChannelOptions(100)
{
    FullMode = BoundedChannelFullMode.Wait,
    SingleReader = true,
    SingleWriter = false
});

await channel.Writer.WriteAsync(job, ct);

await foreach (Job next in channel.Reader.ReadAllAsync(ct))
{
    await ProcessAsync(next, ct);
}
```

Bounded channel выражает backpressure: producer ждёт при полном buffer. Unbounded queue может съесть память, если producer стабильно быстрее consumer.

In-memory channel теряет данные при process crash; durable work требует broker/database/outbox.

## 9. Parallel loops

```csharp
await Parallel.ForEachAsync(
    items,
    new ParallelOptions
    {
        MaxDegreeOfParallelism = Environment.ProcessorCount,
        CancellationToken = ct
    },
    async (item, token) => await ProcessAsync(item, token));
```

Degree должен соответствовать resource:

- CPU-bound — примерно cores;
- external service — его limits и connection pool;
- database — pool/transaction constraints.

Unlimited fan-out через `Task.WhenAll` над миллионами items создаёт tasks/memory pressure.

## 10. PLINQ

```csharp
var result = values
    .AsParallel()
    .WithCancellation(ct)
    .Where(IsPrime)
    .ToArray();
```

PLINQ полезен для CPU-heavy pure operations над достаточно большим dataset. Ordering, small work и synchronization могут сделать его медленнее sequential LINQ.

## 11. Deadlock

```text
Task A: держит Lock 1 → ждёт Lock 2
Task B: держит Lock 2 → ждёт Lock 1
```

Предотвращение:

- единый порядок захвата locks;
- маленькие critical sections;
- отсутствие external calls под lock;
- higher-level primitives/messages вместо shared mutable state;
- timeouts как diagnostics/containment, но не proof correctness.

## 12. ThreadPool starvation

Блокировка pool threads через `.Result`, sync I/O или длительный CPU work ухудшает throughput server. Метрики queue length/thread count/latency помогают отличить starvation от «медленной БД».

`Task.Run` переносит CPU work в ThreadPool, но не делает его дешевле. В ASP.NET не оборачивайте database/HTTP async call в `Task.Run`.

## 13. Immutability и ownership

Лучший shared mutable state — тот, которого нет. Immutable messages, partitioned state и single-reader queue часто проще locks.

Определите:

- кто владеет объектом;
- может ли он изменяться после публикации;
- какая operation atomic;
- что происходит при cancellation/failure.

## 14. Сравнение с Java

| Java | .NET |
|---|---|
| `synchronized` | `lock`/Monitor |
| `AtomicInteger` | `Interlocked` operations |
| `Semaphore` | `SemaphoreSlim` внутри process |
| concurrent collections | `System.Collections.Concurrent` |
| BlockingQueue | `Channel<T>` даёт async producer/consumer |
| parallel streams | PLINQ/Parallel APIs |
| virtual threads | прямого эквивалента как основной модели нет; async/await доминирует для I/O |

## 15. Best practices

- Сначала убирайте shared mutation, затем выбирайте primitive.
- Защищайте invariant, а не одну случайную строку.
- Ограничивайте concurrency.
- Не держите lock во время I/O.
- Не разделяйте EF `DbContext` между concurrent operations.
- Тесты не доказывают отсутствие races; design и stress tools дополняют их.

## Самопроверка

1. Почему `_count++` не является atomic increment?
2. Делает ли `volatile` составную operation атомарной?
3. Чем bounded channel полезнее unbounded queue?
4. Почему concurrent dictionary не делает transaction из трёх вызовов?
5. Когда PLINQ может ухудшить производительность?

