# Семинар 8 — Производительность EF Core

## Задание 1. Создайте измеримый N+1

В тестовой среде включите lazy loading или явно выполните query lines в loop. Seed: 100 orders по 5 lines. Подсчитайте SQL commands и total duration.

## Задание 2. Исправьте projection

Верните summary с line count/total одним server query. Сравните SQL, command count и payload.

## Задание 3. Include и cartesian explosion

Добавьте вторую collection navigation (например, audit entries), загрузите обе single query. Оцените duplicated rows. Сравните:

- single query;
- `AsSplitQuery`;
- DTO projection.

Объясните consistency trade-off split query.

## Задание 4. Tracking benchmark

На representative dataset сравните tracking, no-tracking и projection по времени/allocations. Не делайте вывод по одной итерации Debug build.

## Задание 5. Offset pagination

Реализуйте page/pageSize с limit 100 и deterministic `CreatedAt, Id`. Создайте index, проверьте plan на большой offset.

## Задание 6. Keyset pagination

Cursor кодирует last `CreatedAt` + `Id` (не доверяйте неподписанному cursor для sensitive data). Реализуйте next page и докажите отсутствие duplicates при вставке новой более поздней записи.

## Задание 7. Optimistic concurrency

Два contexts загружают order version 0. Первый сохраняет version 1, второй получает concurrency exception. Преобразуйте его в 412 при `If-Match` либо 409 без HTTP precondition.

## Задание 8. Set-based update

Архивируйте expired orders через `ExecuteUpdateAsync`. Покажите, что tracked entity в уже открытом context может остаться stale; очистите tracker/reload или не смешивайте contexts.

## Задание 9. Query inspection

Для трёх горячих запросов сохраните:

- `ToQueryString`;
- query tag;
- `EXPLAIN ANALYZE` на non-production data;
- используемый index;
- row estimate vs actual.

## Проверка

- N+1 воспроизведён числами и устранён;
- pagination deterministic;
- index обоснован plan, не догадкой;
- concurrency conflict не перетирает данные;
- query имеет cancellation/limit;
- performance вывод основан на representative Release run.

