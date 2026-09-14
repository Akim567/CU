# Семинар 13 — Background jobs и transactional outbox

## Задание 1. Найдите concurrent DbContext bug

Запустите две EF operations на одном scoped context. Исправьте последовательно либо через `IDbContextFactory` с независимыми contexts. Объясните влияние на consistency.

## Задание 2. In-process queue

Создайте bounded channel capacity 100, endpoint enqueue и BackgroundService. При full buffer выберите wait либо 429/503; решение документируйте.

## Задание 3. Scoped job

Worker создаёт async scope на каждый job, получает DbContext/application handler, корректно закрывает scope при success/error/cancel.

## Задание 4. Operation resource

POST `/reports` возвращает 202 + Location. GET `/operations/{id}` показывает state/progress/result. Unknown operation — 404, failed — stable error code.

## Задание 5. Crash loss

Поставьте breakpoint/crash после HTTP 202 до in-memory processing. Покажите потерю job и сформулируйте, когда это недопустимо.

## Задание 6. Outbox

В одной EF transaction измените order и добавьте `OutboxMessage`. Payload содержит versioned integration event DTO, не entity.

## Задание 7. Publisher

Читает batch, использует lease/`SKIP LOCKED`, публикует, отмечает processed. Ограничьте batch и parallelism. Добавьте exponential backoff.

## Задание 8. Duplicate scenario

Имитируйте crash после broker publish до mark processed. Следующий запуск публикует повтор. Consumer с inbox unique message id применяет effect один раз.

## Задание 9. Poison message

После ограниченного числа попыток пометьте failed, сохраните безопасный error summary и поднимите metric/alert. Предусмотрите controlled replay.

## Задание 10. Shutdown

Остановите worker во время job. Проверьте token, lease/ack и отсутствие silent loss.

## Проверка

- job buffer bounded;
- обязательная работа durable;
- order + outbox atomic;
- duplicate воспроизводится и обезвреживается inbox;
- retries bounded;
- poison виден оператору;
- shutdown policy определена.

