# Лекция 6 — ADO.NET, Npgsql, pooling и транзакции

## 1. Слои доступа к PostgreSQL

```text
application
  → ADO.NET abstractions (DbConnection, DbCommand, DbDataReader)
  → Npgsql provider
  → PostgreSQL protocol/server
```

ADO.NET задаёт provider-neutral базовые contracts, Npgsql реализует их для PostgreSQL. EF Core строится выше, но понимание connection/command/reader остаётся обязательным.

## 2. Connection

```csharp
await using var connection = new NpgsqlConnection(connectionString);
await connection.OpenAsync(cancellationToken);
```

`DisposeAsync` обычно возвращает physical connection в pool, а не обязательно закрывает TCP. Поэтому connection object должен быть короткоживущим: open late, dispose early.

Connection string содержит secrets и operational settings. Храните её в configuration secret provider, не логируйте полностью.

## 3. Parameterized command

```csharp
await using var command = connection.CreateCommand();
command.CommandText = """
    select id, number, total
    from orders
    where id = @id
    """;
command.Parameters.AddWithValue("id", id);

await using NpgsqlDataReader reader =
    await command.ExecuteReaderAsync(cancellationToken);
```

Parameters отделяют code SQL от values и защищают от injection. Имена table/column/order direction нельзя передать обычным value parameter; разрешайте их по allow-list.

## 4. DataReader

```csharp
if (!await reader.ReadAsync(ct))
    return null;

var order = new Order(
    reader.GetGuid(reader.GetOrdinal("id")),
    reader.GetString(reader.GetOrdinal("number")),
    reader.GetDecimal(reader.GetOrdinal("total")));
```

Reader forward-only и связан с connection/command. Ordinals можно вычислить один раз перед большим loop. Обрабатывайте database null через `IsDBNull`/nullable getters по provider API.

## 5. Execute methods

- `ExecuteNonQueryAsync` — affected rows для insert/update/delete/DDL;
- `ExecuteScalarAsync` — первая column первой row;
- `ExecuteReaderAsync` — result set.

Affected rows — часть optimistic concurrency check:

```sql
update orders
set total = @total, version = version + 1
where id = @id and version = @expected_version;
```

`0` rows означает missing или concurrency conflict — различайте по contract.

## 6. Pooling

Pool обычно разделяется по нормализованной connection string. Ограниченный pool создаёт backpressure: при нехватке свободной connection caller ждёт/получает timeout.

Причины exhaustion:

- connection не disposed;
- long transaction;
- slow query;
- слишком высокая request concurrency;
- database limit ниже суммарных pools replicas;
- streaming reader удерживается долго.

Увеличение pool size может перегрузить БД сильнее. Сначала найдите причину.

## 7. Transaction

```csharp
await using NpgsqlTransaction transaction =
    await connection.BeginTransactionAsync(ct);

try
{
    await InsertOrderAsync(connection, transaction, order, ct);
    await InsertAuditAsync(connection, transaction, audit, ct);
    await transaction.CommitAsync(ct);
}
catch
{
    await transaction.RollbackAsync(CancellationToken.None);
    throw;
}
```

Commands должны использовать ту же connection/transaction. Rollback при cancellation часто выполняют с отдельным cleanup token/deadline, чтобы caller cancellation не помешала освобождению.

## 8. Isolation

Isolation определяет видимость concurrent changes и anomalies. Уровни .NET мапятся на capabilities PostgreSQL; фактическая semantics определяется database.

- Read Committed — default PostgreSQL, каждый statement видит свой snapshot.
- Repeatable Read — transaction snapshot, PostgreSQL предотвращает больше anomalies, чем некоторые общие описания.
- Serializable — database обнаруживает опасные dependencies; transaction может завершиться serialization failure и требовать полного retry.

Retry только безопасной whole transaction и с ограничением attempts/backoff.

## 9. Transaction boundaries

Не держите transaction во время HTTP call. Иначе locks/connections живут в ожидании сети, а атомарность между DB и external service всё равно не появляется.

Для DB + broker используют outbox/inbox, idempotency и eventual consistency.

## 10. Batch и COPY

Много отдельных round trips медленно. В зависимости от сценария применяют:

- multi-row insert;
- prepared/batched commands;
- Npgsql batch API;
- PostgreSQL `COPY` для bulk load.

Batch size ограничивают по memory, query size, locks и latency.

## 11. Migrations

ADO.NET не диктует migration tool. Возможны EF migrations, FluentMigrator, DbUp или SQL migrations. Главное:

- versioned immutable scripts;
- backward-compatible rollout;
- отдельно оценивать locks/table rewrites;
- не запускать конкурентно из каждой replica без coordination.

## 12. Сравнение с JDBC/Spring JDBC

| Java | .NET |
|---|---|
| JDBC Driver | ADO.NET provider/Npgsql |
| `Connection` | `DbConnection`/`NpgsqlConnection` |
| `PreparedStatement` | parameterized `DbCommand` |
| `ResultSet` | `DbDataReader` |
| DataSource/HikariCP | provider data source/connection pooling |
| `JdbcTemplate` | встроенного точного аналога нет; Dapper часто даёт lightweight mapping |

## 13. Best practices

- Parameterize values.
- Dispose connection/command/reader.
- Передавайте cancellation.
- Короткие transactions.
- Проверяйте affected rows.
- Не retry non-idempotent fragment.
- Настраивайте pool относительно database и replica count.
- Наблюдайте query duration, waits и pool saturation.

## Самопроверка

1. Закрывает ли dispose physical connection?
2. Почему нельзя parameterize column name?
3. Что означает 0 affected rows при versioned update?
4. Почему HTTP call внутри DB transaction опасен?
5. Что нужно retry при serialization failure?

