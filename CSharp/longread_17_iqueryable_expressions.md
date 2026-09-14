# Конспект лекции 17 — `IQueryable<T>` и expression trees

## 1. Одинаковый синтаксис, разное выполнение

```csharp
IEnumerable<User> local = users.Where(u => u.IsActive);
IQueryable<User> remote = db.Users.Where(u => u.IsActive);
```

В первом случае `Where` получает `Func<User,bool>` и выполняет compiled code над объектами. Во втором — `Expression<Func<User,bool>>`; provider анализирует tree и, например, переводит его в SQL.

## 2. Expression tree как данные

```csharp
Expression<Func<User, bool>> predicate =
    user => user.Age >= 18 && user.IsActive;
```

Упрощённая форма:

```text
Lambda(user)
└─ AndAlso
   ├─ GreaterThanOrEqual
   │  ├─ Member(user.Age)
   │  └─ Constant(18)
   └─ Member(user.IsActive)
```

Tree можно исследовать, переписывать и компилировать в delegate.

## 3. Query provider

`IQueryable<T>` содержит:

- `Expression` — описание запроса;
- `Provider` — умеет строить/выполнять запрос;
- element type.

LINQ methods из `Queryable` строят новые expression nodes. Реальная работа обычно начинается terminal operation (`ToListAsync`, `SingleAsync`, enumeration).

## 4. Translation boundary

Не любой .NET method имеет SQL equivalent:

```csharp
static bool IsCorporate(string email) => email.EndsWith("@company.test");

var query = db.Users.Where(user => IsCorporate(user.Email));
```

Provider может не перевести custom method. Правильные варианты:

- выразить условие поддерживаемыми operations;
- добавить provider-specific database function mapping;
- осознанно перейти к local execution после уменьшения result set.

```csharp
var candidates = await db.Users
    .Where(user => user.IsActive)
    .Select(user => user.Email)
    .ToListAsync(ct);

var corporate = candidates.Where(IsCorporate);
```

Сначала server filter/projection, потом local calculation. Нельзя случайно загрузить всю таблицу.

## 5. `AsEnumerable`

```csharp
IEnumerable<User> localPart = query.AsEnumerable();
```

`AsEnumerable` меняет дальнейший выбор LINQ operators на `Enumerable`, но само по себе не materialize. При enumeration remote query выполнится, затем остаток обработается локально.

## 6. Projection

```csharp
var summaries = await db.Orders
    .Where(order => order.Status == OrderStatus.Paid)
    .Select(order => new OrderSummary(
        order.Id,
        order.Lines.Sum(line => line.Price * line.Quantity)))
    .ToListAsync(ct);
```

Проекция выбирает только нужные columns и позволяет provider построить aggregation. `Include` не нужен, если данные сразу проецируются в DTO.

## 7. Dynamic composition

```csharp
IQueryable<Order> query = db.Orders;

if (filter.CustomerId is { } customerId)
    query = query.Where(order => order.CustomerId == customerId);

if (filter.From is { } from)
    query = query.Where(order => order.CreatedAt >= from);

query = filter.Descending
    ? query.OrderByDescending(order => order.CreatedAt)
    : query.OrderBy(order => order.CreatedAt);
```

Composition сохраняет единую tree до execution. Это обычно яснее ручного создания expression nodes.

## 8. Ручное построение expression

```csharp
ParameterExpression parameter = Expression.Parameter(typeof(User), "user");
MemberExpression age = Expression.Property(parameter, nameof(User.Age));
ConstantExpression limit = Expression.Constant(18);
BinaryExpression body = Expression.GreaterThanOrEqual(age, limit);

Expression<Func<User, bool>> adult =
    Expression.Lambda<Func<User, bool>>(body, parameter);
```

Нужно для query builders, rules engines, serializers и frameworks. В обычном application code lambda composition проще.

## 9. Captured variables и parameters

```csharp
int minimumAge = request.MinimumAge;
var query = db.Users.Where(user => user.Age >= minimumAge);
```

EF Core обычно parameterizes captured value, что помогает reuse query plan и защищает от injection. Не вставляйте user input в raw SQL string interpolation без provider parameterization API.

## 10. Repository boundary

Возвращать `IQueryable<T>` наружу удобно, но утечка provider abstraction позволяет любому consumer сформировать непредсказуемый запрос.

Варианты:

- вернуть готовый DTO/list;
- принять specification/expression с ограниченным contract;
- оставить `IQueryable` внутри application/data layer;
- осознанно раскрыть его в internal reporting module.

Решение зависит от архитектуры, а не от универсального запрета.

## 11. Expression limitations

Expression trees исторически поддерживают подмножество language constructs и сохраняют совместимость с consumers. Новая syntax может компилироваться только в delegate либо быть lowered. Всегда проверяйте provider capabilities и generated query.

## 12. Сравнение с Java

Java Stream содержит executable pipeline. JPA Criteria API или query DSL строит query model отдельно. В C# одинаковый LINQ syntax может связываться с `Enumerable` или `Queryable` благодаря overloads и expression trees.

Это мощно, но создаёт риск: корректный C# метод может оказаться непереводимым provider. Тип `IQueryable<T>` в сигнатуре — сигнал, что код пока описывает запрос, а не перебирает объекты.

## 13. Best practices

- Смотрите generated SQL для важных queries.
- Project only needed columns.
- Не вызывайте `ToList` до filters/sorting/paging.
- Явно отмечайте переход server → client.
- Не используйте side effects в query expression.
- Ограничивайте произвольный `IQueryable` на архитектурных границах.
- Передавайте cancellation token в async terminal operations.

## Самопроверка

1. Что получает `Enumerable.Where`, а что `Queryable.Where`?
2. Выполняет ли `AsEnumerable` запрос сразу?
3. Почему custom method может не работать в EF query?
4. Когда projection заменяет `Include`?
5. В чём риск публичного repository method, возвращающего `IQueryable`?

