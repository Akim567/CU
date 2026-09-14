# Семинар 12 — Async request processing и streaming

## Задание 1. Найдите blocking-over-async

Добавьте endpoint с `.Result`, дайте concurrent load и наблюдайте latency/ThreadPool counters. Затем замените на `await` и сравните.

## Задание 2. Cancellation chain

Передайте request token controller → service → repository → Npgsql. Используйте `pg_sleep`, отмените client request и проверьте прекращение command/cleanup.

## Задание 3. CPU endpoint

Создайте тяжёлое вычисление. Сравните:

- выполнение inline;
- `Task.Run`;
- bounded background queue + 202/job status.

Объясните, почему второй вариант не увеличивает CPU capacity.

## Задание 4. `IAsyncEnumerable`

Верните orders batch-by-batch с cancellation. Запрещено materialize всё до первого response byte. Проверьте медленного client.

## Задание 5. NDJSON

Реализуйте `application/x-ndjson`: один JSON object + newline, flush по разумной batch policy. При partial failure stream закрывается; HTTP status уже не меняется.

## Задание 6. SSE

Endpoint отправляет status events и heartbeat. На disconnect отписывается от publisher. Buffer bounded; задайте policy для медленного consumer.

## Задание 7. Streaming upload

Сохраните большой file через Request.Body без полного buffering. Добавьте size/quota, SHA-256, temporary file и cleanup при cancel/error.

## Задание 8. Deadline budget

Request имеет 5 секунд общего budget, outbound dependency — attempt timeout 1 секунда и максимум 2 retries. Докажите, что retries не переживают request deadline.

## Задание 9. Load test

Сравните normal и slow clients. Наблюдайте:

- request latency;
- active requests/connections;
- ThreadPool queue;
- allocation rate;
- cancellation count;
- channel depth.

## Проверка

- нет `.Result`/`.Wait`;
- cancellation проходит весь chain;
- buffers/concurrency bounded;
- partial streaming failure определён;
- upload cleanup надёжен;
- CPU job отделён от immediate HTTP lifecycle.

