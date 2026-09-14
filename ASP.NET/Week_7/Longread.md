# Лекция 7 — Entity Framework Core: модель, tracking и связи

## 1. Что такое EF Core

EF Core — ORM и data access framework. Он:

- строит model между .NET types и database schema;
- переводит LINQ expression trees в provider queries;
- отслеживает entity state;
- генерирует INSERT/UPDATE/DELETE через `SaveChanges`;
- управляет relations, migrations и concurrency mechanisms.

EF Core не отменяет SQL, indexes, transactions и database constraints.

## 2. `DbContext` как unit of work

```csharp
public sealed class OrderDbContext(DbContextOptions<OrderDbContext> options)
    : DbContext(options)
{
    public DbSet<Order> Orders => Set<Order>();
    public DbSet<OrderLine> OrderLines => Set<OrderLine>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(OrderDbContext).Assembly);
    }
}
```

Context:

- short-lived;
- scoped per request/unit of work by default;
- не thread-safe;
- dispose после unit;
- не должен быть global repository/cache.

## 3. Registration

```csharp
services.AddDbContext<OrderDbContext>(options =>
    options.UseNpgsql(configuration.GetConnectionString("Orders")));
```

Connection pooling provider и `DbContext` pooling — разные механизмы. Context pooling reuse-ит instance state и требует особой осторожности с tenant/user-specific fields.

## 4. Entity mapping

```csharp
public sealed class Order
{
    private readonly List<OrderLine> _lines = [];

    public Guid Id { get; private set; }
    public string Number { get; private set; } = null!;
    public OrderStatus Status { get; private set; }
    public long Version { get; private set; }
    public IReadOnlyCollection<OrderLine> Lines => _lines;
}
```

Fluent configuration:

```csharp
public sealed class OrderConfiguration : IEntityTypeConfiguration<Order>
{
    public void Configure(EntityTypeBuilder<Order> builder)
    {
        builder.ToTable("orders");
        builder.HasKey(x => x.Id);
        builder.Property(x => x.Number).HasMaxLength(40).IsRequired();
        builder.HasIndex(x => x.Number).IsUnique();
        builder.Property(x => x.Version).IsConcurrencyToken();
    }
}
```

Attributes удобны для простого mapping; Fluent API поддерживает полный provider model и держит infrastructure concerns вне domain type.

## 5. Entity states

| State | Смысл при SaveChanges |
|---|---|
| Detached | context не отслеживает |
| Unchanged | отслеживает, изменений нет |
| Added | INSERT |
| Modified | UPDATE изменённых properties |
| Deleted | DELETE |

```csharp
EntityEntry<Order> entry = db.Entry(order);
Console.WriteLine(entry.State);
```

## 6. Identity map

Tracking query обычно возвращает один instance для одной primary key внутри context. Это помогает relationship fix-up, но context, живущий слишком долго, накапливает entities и stale state.

## 7. Query

```csharp
Order? order = await db.Orders
    .Include(x => x.Lines)
    .SingleOrDefaultAsync(x => x.Id == id, ct);
```

`SingleOrDefault` выражает uniqueness key. LINQ строит query до terminal async call.

Для read-only projection:

```csharp
OrderResponse? result = await db.Orders
    .Where(x => x.Id == id)
    .Select(x => new OrderResponse(
        x.Id,
        x.Number,
        x.Lines.Sum(line => line.Price * line.Quantity)))
    .SingleOrDefaultAsync(ct);
```

## 8. `SaveChangesAsync`

```csharp
db.Orders.Add(order);
await db.SaveChangesAsync(ct);
```

SaveChanges:

- detects changes;
- orders commands с учётом graph;
- использует transaction при необходимости;
- обновляет generated values/state.

Это не distributed transaction с broker/HTTP.

## 9. Relationships

```csharp
builder.HasMany(order => order.Lines)
    .WithOne(line => line.Order)
    .HasForeignKey(line => line.OrderId)
    .OnDelete(DeleteBehavior.Cascade);
```

Различайте:

- navigation property;
- foreign key property;
- principal/dependent;
- required/optional relation;
- database cascade и client-side behavior.

## 10. Loading strategies

- eager: `Include`;
- explicit: `Entry(...).Collection(...).LoadAsync()`;
- lazy: proxy/navigation access автоматически query.

Lazy loading удобно для demos, но скрывает I/O и создаёт N+1. В server code чаще применяют explicit projection/eager loading.

## 11. Owned/complex value objects

Value object вроде Address/Money можно мапить в columns/complex structure в зависимости от EF version/provider. Сохраняйте domain equality/invariants и учитывайте ограничения query translation/migrations.

## 12. Migrations

```bash
dotnet ef migrations add InitialCreate
dotnet ef database update
```

Migration — generated proposal, её надо review:

- удаление/переименование распознано верно?
- нужен ли backfill?
- будет ли table lock/rewrite?
- совместимы ли старая и новая versions приложения?

Production migrations лучше выполнять контролируемым job/step, а не каждой replica на startup.

## 13. Repository нужен не всегда

`DbContext` уже unit of work/repository-like abstraction. Дополнительный repository полезен, если:

- скрывает aggregate-specific queries;
- защищает architecture boundary;
- даёт domain-oriented methods;
- меняет storage implementation реально возможно.

Generic CRUD repository часто только прячет сильные EF capabilities и возвращает `IQueryable` без правил.

## 14. Сравнение с JPA/Hibernate

| JPA/Hibernate | EF Core |
|---|---|
| EntityManager/Session | DbContext |
| persistence context | ChangeTracker/identity map |
| JPQL/HQL/Criteria | LINQ expression queries |
| entity states | Added/Unchanged/Modified/Deleted/Detached |
| `persist`/flush | Add + SaveChanges |
| annotations | attributes/Fluent API |

EF Core не является реализацией JPA. Fetch defaults, cascades, proxies, dirty tracking и query language различаются.

## 15. Best practices

- Один short unit of work.
- Никогда не используйте context concurrent.
- Project DTO для reads.
- Domain/database constraints защищают invariants.
- Review migrations.
- Avoid lazy loading по умолчанию.
- Не оборачивайте EF механически generic repository.

## Самопроверка

1. Почему DbContext scoped, но не thread-safe?
2. Что даёт identity map?
3. Чем connection pooling отличается от context pooling?
4. Когда projection лучше Include?
5. Почему generated migration требует review?

