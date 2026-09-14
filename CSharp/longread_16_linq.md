# Конспект лекции 16 — LINQ и `IEnumerable<T>`

## Введение

LINQ (Language Integrated Query) задаёт общий vocabulary запросов к последовательностям: filter, projection, ordering, grouping, aggregation. Это не одна коллекция и не только SQL.

## 1. Императивная и декларативная обработка

```csharp
var result = new List<string>();
foreach (User user in users)
{
    if (user.IsActive)
        result.Add(user.Name.ToUpperInvariant());
}
result.Sort();
```

LINQ:

```csharp
List<string> result = users
    .Where(user => user.IsActive)
    .Select(user => user.Name.ToUpperInvariant())
    .Order()
    .ToList();
```

Pipeline описывает намерение. Он не всегда быстрее и не отменяет знание complexity/allocations.

## 2. Deferred execution

```csharp
IEnumerable<User> active = users.Where(user => user.IsActive);
```

`Where` обычно не перебирает source сразу. Работа начинается при enumeration:

```csharp
foreach (User user in active) { }
```

Если source изменился до enumeration, результат может измениться. Каждый повторный enumeration часто повторяет computation.

```csharp
List<User> snapshot = active.ToList();
```

Materialization фиксирует snapshot на этот момент и расходует память.

## 3. Intermediate и terminal operations

Ленивые sequence-returning operations:

- `Where`, `Select`, `SelectMany`;
- `Take`, `Skip`, `Distinct`;
- `OrderBy`, `ThenBy` (буферизуют при enumeration);
- `Chunk`, `Append`, `Concat`.

Terminal/materializing operations:

- `ToList`, `ToArray`, `ToDictionary`;
- `Count`, `Any`, `All`;
- `First`, `Single`, `Last` variants;
- `Sum`, `Average`, `Min`, `Max`, `Aggregate`.

## 4. `Where` и `Select`

```csharp
var summaries = orders
    .Where(order => order.Status == OrderStatus.Paid)
    .Select(order => new
    {
        order.Id,
        Total = order.Lines.Sum(line => line.Price * line.Quantity)
    });
```

Anonymous type удобен внутри метода/query. На публичной границе нужен именованный DTO/record.

## 5. `SelectMany`

```csharp
IEnumerable<Line> allLines = orders.SelectMany(order => order.Lines);
```

Он преобразует каждый element в sequence и flatten-ит sequences на один уровень.

```csharp
var pairs = orders.SelectMany(
    order => order.Lines,
    (order, line) => new { order.Id, Line = line });
```

## 6. `First`, `Single` и absence contract

| Метод | Пусто | Более одного |
|---|---|---|
| `First` | exception | первый |
| `FirstOrDefault` | default | первый |
| `Single` | exception | exception |
| `SingleOrDefault` | default | exception |

`Single` утверждает uniqueness invariant. Не выбирайте его только потому, что «ожидается один» без готовности диагностировать duplicate.

## 7. `Any` против `Count`

```csharp
bool exists = users.Any();
```

Для общего `IEnumerable<T>` `Any` может остановиться после первого элемента. `Count()` может пройти всю последовательность, хотя оптимизированные sources умеют сообщить count без enumeration.

Для проверки predicate:

```csharp
bool hasAdmin = users.Any(user => user.Role == "admin");
```

## 8. Ordering

```csharp
var ordered = users
    .OrderBy(user => user.LastName, StringComparer.Ordinal)
    .ThenBy(user => user.FirstName, StringComparer.Ordinal);
```

`ThenBy` сохраняет предыдущие keys. Второй `OrderBy` начинает новое primary ordering.

## 9. Grouping и lookup

```csharp
var totals = orders
    .GroupBy(order => order.CustomerId)
    .Select(group => new
    {
        CustomerId = group.Key,
        Total = group.Sum(order => order.Total)
    });
```

`ToLookup` создаёт one-to-many lookup, где отсутствующий key даёт пустую sequence. `ToDictionary` требует unique keys и обычно бросает exception на duplicate.

## 10. Aggregation

```csharp
decimal total = orders.Sum(order => order.Total);
decimal max = orders.Max(order => order.Total);

decimal product = numbers.Aggregate(
    seed: 1m,
    (acc, value) => acc * value);
```

`Aggregate` мощный, но specialized operation часто понятнее.

## 11. Set operations

```csharp
var common = first.Intersect(second);
var combined = first.Union(second);
var missing = required.Except(actual);
```

Equality определяется default или переданным comparer.

## 12. Собственные iterators и `yield`

```csharp
static IEnumerable<int> Positive(IEnumerable<int> source)
{
    foreach (int value in source)
    {
        if (value > 0)
            yield return value;
    }
}
```

Compiler строит state machine. Код до первого `yield return` фактически выполняется при enumeration, а не обязательно при вызове метода.

Resource внутри iterator должен быть корректно disposed, если enumeration прекращена:

```csharp
static IEnumerable<string> ReadLines(string path)
{
    using var reader = File.OpenText(path);
    while (reader.ReadLine() is { } line)
        yield return line;
}
```

`foreach` dispose-ит enumerator, что запускает iterator `finally`.

## 13. Multiple enumeration

```csharp
IEnumerable<User> query = LoadUsersFromRemoteSource();
int count = query.Count();
User? first = query.FirstOrDefault();
```

Source может быть вычислен дважды. Если повторение дорого или имеет side effects, materialize один раз. Но не добавляйте `ToList` в каждый pipeline: streaming и early termination могут быть важнее.

## 14. Сравнение с Java Stream API

| Java Stream | LINQ to Objects |
|---|---|
| stream обычно одноразовый | `IEnumerable<T>` обычно можно запросить повторно, но source contract может отличаться |
| `filter` | `Where` |
| `map` | `Select` |
| `flatMap` | `SelectMany` |
| `collect(toList())` / `toList()` | `ToList()` |
| `anyMatch` | `Any` |
| `reduce` | `Aggregate` |
| primitive streams | generic algorithms + specialized APIs; value-type generics |

LINQ operators могут работать с `IEnumerable<T>` как executable code или `IQueryable<T>` как expression tree. Это принципиальная граница следующей лекции.

## 15. Best practices

- Держите pipeline читаемым; сложную lambda выносите в named method.
- Осознавайте момент enumeration.
- Не делайте side effects в `Where`/`Select`.
- Выбирайте `Single` только для реального invariant.
- Передавайте comparer для строк и domain keys.
- Не смешивайте `IEnumerable` и `IQueryable` без понимания места выполнения.

## Самопроверка

1. Когда начинает выполняться `Where`?
2. Чем `FirstOrDefault` отличается от `SingleOrDefault`?
3. Почему два enumeration могут быть дорогими?
4. Чем `ToLookup` отличается от `ToDictionary`?
5. Что compiler создаёт для iterator с `yield`?

