# Семинар 6 — PostgreSQL через ADO.NET и Npgsql

## Подготовка схемы

```sql
create table orders (
    id uuid primary key,
    number text not null unique,
    total numeric(18,2) not null check (total > 0),
    status text not null,
    version bigint not null default 0,
    created_at timestamptz not null
);
```

Connection string получите из secret/configuration, не вставляйте в source.

## Задание 1. Data source

Создайте singleton `NpgsqlDataSource` из validated options и repository, открывающий connection на operation. Не создавайте новый data source на request.

## Задание 2. Insert

Выполните parameterized insert. `created_at` передавайте как UTC `DateTimeOffset`/совместимый provider type. Верните affected rows и требуйте `1`.

## Задание 3. Read mapping

Реализуйте `FindAsync(Guid)`. Обработайте no rows. Не используйте `SELECT *`; перечислите columns и сопоставьте типы.

## Задание 4. SQL injection experiment

Сначала покажите опасный код (не оставляйте его):

```csharp
command.CommandText = $"select ... where number = '{input}'";
```

Введите quote-based payload, затем замените parameter и докажите, что payload стал обычным value.

## Задание 5. Dynamic sorting allow-list

Разрешите только `createdAt` и `total`, `asc`/`desc`. Сопоставьте enum с заранее заданным SQL fragment. Values фильтра/limit/offset остаются parameters.

## Задание 6. Transaction

Одной transaction вставьте order и audit row. Искусственно вызовите exception между commands и проверьте отсутствие обеих rows.

## Задание 7. Optimistic update

```sql
update orders
set status = @status, version = version + 1
where id = @id and version = @expected_version;
```

При 0 rows различите 404 и 409 дополнительной проверкой либо определите единый conflict contract.

## Задание 8. Pool exhaustion

В тестовой среде задайте небольшой pool, намеренно удержите несколько open connections, наблюдайте timeout. Затем исправьте `await using`. Не повышайте pool как окончательное решение.

## Задание 9. Cancellation

Запустите `select pg_sleep(...)`, отмените token и убедитесь, что command прекращается, connection возвращается в пригодное состояние/pool согласно provider behavior.

## Проверка

- SQL values parameterized;
- resources закрываются;
- transaction atomic;
- version conflict обнаруживается;
- sorting fragments allow-listed;
- cancellation доходит до provider;
- connection string не логируется.

