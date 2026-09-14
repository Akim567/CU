# Семинар 7 — Основы EF Core

## Задание 1. Подключение

Добавьте EF Core provider Npgsql версии 10.x и design package. Зарегистрируйте `OrderDbContext` через validated connection string.

## Задание 2. Entities и configuration

Создайте `Order` и `OrderLine` с private setters и domain methods. Mapping вынесите в `IEntityTypeConfiguration<T>`.

Требования:

- UUID primary keys;
- unique order number;
- `numeric(18,2)`;
- status conversion;
- required one-to-many;
- cascade policy задана явно.

## Задание 3. Первая migration

Создайте migration, прочитайте `Up` и `Down`, SQL script. Убедитесь, что unique index/check constraints соответствуют model.

## Задание 4. CRUD

Реализуйте:

- create aggregate;
- find с lines;
- update status через domain method;
- delete/soft-delete по выбранному contract.

Один request использует один DbContext и один `SaveChangesAsync` для aggregate operation.

## Задание 5. Entity states

До/после Add, query, property mutation и SaveChanges выведите `Entry.State`. Объясните переходы.

## Задание 6. Identity map

В одном context дважды запросите тот же tracking entity и сравните `ReferenceEquals`. Затем выполните `AsNoTracking` и сравните.

## Задание 7. Relations

Добавьте line через domain method, сохраняя обе стороны graph по необходимости. Проверьте generated FK и поведение удаления order.

## Задание 8. Projection

Создайте list endpoint, который возвращает summary без `Include`:

```csharp
var items = await db.Orders
    .AsNoTracking()
    .OrderByDescending(x => x.CreatedAt)
    .Select(x => new OrderSummary(x.Id, x.Number, x.Lines.Count))
    .ToListAsync(ct);
```

Посмотрите generated SQL.

## Задание 9. Ошибка concurrency использования

Запустите две async queries одновременно на одном context и зафиксируйте ошибку/недопустимость. Исправьте последовательным await либо отдельными contexts, если операции действительно независимы.

## Проверка

- schema и model согласованы;
- constraints присутствуют в DB;
- DbContext short-lived;
- entities не возвращаются из controller;
- read projection no-tracking;
- relation/cascade проверены интеграционно;
- migration review выполнен.

