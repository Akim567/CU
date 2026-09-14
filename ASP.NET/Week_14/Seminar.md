# Семинар 14 — Kafka producer, consumer, retry и idempotency

## Задание 1. Topic design

Создайте `order-events` минимум с несколькими partitions и replication settings, подходящими тестовому cluster. Обоснуйте partition key `orderId`.

## Задание 2. Event envelope

```csharp
public sealed record IntegrationEvent<T>(
    Guid MessageId,
    string Type,
    int SchemaVersion,
    DateTimeOffset OccurredAt,
    string CorrelationId,
    T Data);
```

Не включайте CLR assembly-qualified type name.

## Задание 3. Singleton producer

Настройте `acks=all`, idempotence и delivery report. Publish через abstraction с key/orderId, trace headers и cancellation. Producer не создаётся на message.

## Задание 4. Outbox publisher

Прочитайте outbox batch, publish, пометьте processed. Имитируйте crash после acknowledgement и покажите duplicate на следующем запуске.

## Задание 5. Consumer group

Запустите 1, затем несколько instances. Наблюдайте partition assignment/rebalance. Запустите consumers больше partitions и объясните idle members.

## Задание 6. Manual processing/commit

Отключите auto-commit по выбранной model. Обработайте message в DB transaction с inbox id, затем продвиньте offset. Имитируйте crash в каждой точке.

## Задание 7. Idempotent consumer

Unique constraint на `(consumer_name, message_id)`. Duplicate не повторяет domain effect, но считается успешно обработанным для offset progression.

## Задание 8. Ordering

Отправьте последовательные versions одного order с одним key и с разными keys. Consumer отклоняет/паркует stale version и не применяет events вне domain order.

## Задание 9. Retry

Transient DB error: bounded backoff. Permanent schema/domain error: DLQ. Определите, будет ли partition блокироваться и почему.

## Задание 10. DLQ и replay

Сохраните original coordinates, message id, schema version и safe error. Создайте controlled replay command, который не обходит inbox/validation.

## Задание 11. Metrics

Добавьте low-cardinality metrics: processed, failed, duplicate, DLQ, processing duration, outbox age. Lag берите из Kafka monitoring, а не вычисляйте наивно одним consumer.

## Проверка

- ordering key выбран осознанно;
- producer reuse и acknowledgements настроены;
- duplicates воспроизводятся и безопасны;
- offset/DB order documented;
- retry bounded;
- DLQ операционно обслуживается;
- schema version присутствует;
- metrics без high-cardinality ids.

