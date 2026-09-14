# Лекция 8 — EF Core: производительность, N+1, paging и concurrency

## 1. Сначала увидеть SQL

LINQ — source code, SQL — фактическая database operation.

```csharp
IQueryable<Order> query = db.Orders.Where(x => x.Status == status);
string sql = query.ToQueryString();
```

В production наблюдайте duration, rows, errors и command category, но не логируйте sensitive parameter values без политики.

## 2. N+1

```csharp
foreach (Order order in orders)
{
    Console.WriteLine(order.Lines.Count); // lazy loading может дать отдельный query
}
```

Один query orders + N queries lines. Latency растёт по round trips.

Решения:

- projection aggregation;
- `Include` для graph;
- explicit batch query;
- отключение lazy loading;
- DataLoader-like batching в соответствующем API.

Выбор зависит от нужной shape, не от правила «всегда Include».

## 3. Projection first

```csharp
var page = await db.Orders
    .AsNoTracking()
    .Where(x => x.Status == status)
    .OrderByDescending(x => x.CreatedAt)
    .Select(x => new OrderSummary(
        x.Id,
        x.Number,
        x.Lines.Count,
        x.Lines.Sum(l => l.Price * l.Quantity)))
    .Take(50)
    .ToListAsync(ct);
```

Database выполняет count/sum, сеть передаёт только DTO columns.

## 4. Tracking cost

Tracking нужен для update graph. Read-only queries:

```csharp
db.Orders.AsNoTracking();
```

`AsNoTrackingWithIdentityResolution` устраняет duplicate instances в result graph без долгосрочного tracking, но имеет собственную стоимость.

## 5. Single и split queries

Несколько collection `Include` в одном SQL могут создать cartesian explosion и повторить parent columns.

```csharp
var orders = await db.Orders
    .Include(x => x.Lines)
    .AsSplitQuery()
    .ToListAsync(ct);
```

Split query делает несколько SQL requests и имеет consistency/round-trip trade-offs. Projection часто лучше обоих вариантов.

## 6. Pagination

Offset pagination:

```csharp
query.OrderBy(x => x.CreatedAt).ThenBy(x => x.Id)
    .Skip((page - 1) * pageSize)
    .Take(pageSize);
```

Проблемы: большой offset сканируется, concurrent inserts сдвигают страницы.

Keyset pagination:

```csharp
query.Where(x =>
        x.CreatedAt > cursorTime ||
        (x.CreatedAt == cursorTime && x.Id.CompareTo(cursorId) > 0))
    .OrderBy(x => x.CreatedAt)
    .ThenBy(x => x.Id)
    .Take(pageSize);
```

Нужен deterministic unique ordering и соответствующий composite index.

## 7. Indexes

Index проектируется под WHERE/JOIN/ORDER BY и selectivity. Слишком много indexes замедляет writes и занимает место.

```csharp
builder.HasIndex(x => new { x.Status, x.CreatedAt });
```

Порядок columns важен. Проверяйте `EXPLAIN (ANALYZE, BUFFERS)` на representative data.

## 8. Optimistic concurrency

```csharp
builder.Property(x => x.Version).IsConcurrencyToken();
```

EF включает original version в update predicate. Если rows affected = 0, бросается `DbUpdateConcurrencyException`.

Стратегии:

- сообщить 409/412;
- reload и попросить повторить;
- merge non-conflicting fields;
- controlled retry whole use case, если semantically safe.

Нельзя просто бесконечно retry, перетирая новое значение.

## 9. Pessimistic locking

Provider-specific SQL `FOR UPDATE` удерживает locks до конца transaction. Нужен для некоторых critical sections, но увеличивает contention/deadlock risk. Не держите lock во время external call.

## 10. Bulk updates

```csharp
int updated = await db.Orders
    .Where(x => x.Status == OrderStatus.Expired)
    .ExecuteUpdateAsync(setters => setters
        .SetProperty(x => x.Status, OrderStatus.Archived), ct);
```

Set-based command не загружает entities и не синхронизирует уже tracked instances автоматически. Domain callbacks/in-memory invariants не вызываются.

## 11. Compiled queries и context pooling

Оба механизма нужны редко: EF уже cache-ит query compilation shapes, а context creation сравнительно дёшево. Применяйте после profiling high-throughput hot path и внимательно учитывайте tenant/state reset.

## 12. Client evaluation и translation

Непереводимый predicate обычно приводит к exception, а не скрытой загрузке всего набора в современных EF Core. Явный `AsEnumerable` переносит остаток в memory — делайте это после строгого server filter/projection.

## 13. Query tags

```csharp
db.Orders.TagWith("Orders.ListByStatus");
```

Tag помогает связать SQL с use case. Не включайте user input/PII.

## 14. Сравнение с Hibernate

Проблемы N+1, identity map, eager/lazy loading, batching и optimistic locking общие. Отличаются fetch defaults, APIs и generated SQL.

- `Include` не равен во всех деталях `JOIN FETCH`.
- `AsNoTracking` — явный EF read optimization.
- LINQ translation отличается от JPQL/HQL.
- `ExecuteUpdate` обходит tracked domain graph подобно bulk DML caveats ORM.

## 15. Checklist

- SQL и plan просмотрены.
- Projection before Include для API reads.
- Page имеет deterministic order и limit.
- Index соответствует query.
- Context не живёт долго и не используется concurrent.
- Concurrency conflict обработан явно.
- Bulk operation не оставляет stale tracked state.

## Самопроверка

1. Почему Include не универсальное решение N+1?
2. Чем split query платит за уменьшение cartesian explosion?
3. Почему keyset требует unique ordering?
4. Что означает `DbUpdateConcurrencyException`?
5. Почему bulk update может нарушить ожидания tracked graph?

