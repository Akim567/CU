# Лекция 14 — Event-Driven Architecture и Kafka в .NET

## 1. Event-driven architecture

Event сообщает о свершившемся факте:

```text
OrderPaid v1
messageId, orderId, amount, currency, occurredAt
```

Producer не вызывает конкретный consumer напрямую. Это снижает temporal coupling, но добавляет eventual consistency, duplicates, ordering и schema evolution.

Command просит выполнить действие (`CapturePayment`), event сообщает факт (`PaymentCaptured`). Имена events обычно в прошедшем времени.

## 2. Broker comparison

- RabbitMQ силён routing/queues/ack-oriented messaging.
- NATS предлагает лёгкую messaging модель и JetStream durability.
- Kafka — distributed append log с partitions, retention и replay.

Выбор по ordering, retention, throughput, routing, operations и consumer model, а не популярности.

## 3. Topic, partition, record

- **topic** — логический поток;
- **partition** — упорядоченный append log;
- **record** — key, value, headers, timestamp;
- **offset** — позиция record внутри partition.

Global ordering всего topic не гарантируется. Ordering существует внутри partition.

## 4. Key и partitioning

Key `orderId` обычно направляет события одного order в одну partition, сохраняя их порядок. Hot key создаёт skew. Изменение partition count может изменить mapping keys для будущих messages.

## 5. Replication и ISR

Partition имеет leader и replicas. In-sync replicas достаточно близки к leader. Durability producer зависит от acknowledgements, replication factor и `min.insync.replicas`.

`acks=all` не значит абсолютную невозможность потери при неверной cluster configuration/unclean operations, но является основой сильного durability contract.

## 6. Producer

Концептуальная registration:

```csharp
var config = new ProducerConfig
{
    BootstrapServers = options.BootstrapServers,
    Acks = Acks.All,
    EnableIdempotence = true
};
```

Producer batching/compression повышают throughput ценой latency/CPU. Reuse producer: создание на каждое сообщение дорого.

```csharp
DeliveryResult<string, byte[]> result = await producer.ProduceAsync(
    "order-events",
    new Message<string, byte[]>
    {
        Key = orderId.ToString(),
        Value = payload,
        Headers = headers
    },
    ct);
```

Delivery success означает broker acknowledgement по config, не обработку consumer.

## 7. Consumer group

В group каждая partition назначается максимум одному активному consumer group member. Разные groups читают независимо.

Максимальный полезный parallelism одного group для topic примерно ограничен числом partitions. Больше consumers → часть idle.

Rebalance перераспределяет partitions. Долгая processing/неправильные poll settings могут вызвать повторную доставку и ownership changes.

## 8. Offset и commit

Offset commit — позиция, с которой group продолжит после restart/rebalance.

- commit до processing → риск at-most-once/loss при crash;
- processing, затем commit → at-least-once/duplicates;
- exactly-once end-to-end требует transaction/idempotency на всех side effects, не одного флага.

## 9. Consumer BackgroundService

Confluent consumer `Consume` традиционно blocking. Его размещают в dedicated long-running loop/worker с корректным shutdown, а не делают вид, что он Task-native I/O.

```csharp
protected override Task ExecuteAsync(CancellationToken stoppingToken) =>
    Task.Factory.StartNew(
        () => ConsumeLoop(stoppingToken),
        stoppingToken,
        TaskCreationOptions.LongRunning,
        TaskScheduler.Default);
```

Конкретную threading model сверяйте с client version. Consumer instance обычно не thread-safe для arbitrary concurrent calls.

## 10. Processing pipeline

```text
consume
 → validate envelope/schema
 → idempotency/inbox
 → domain transaction
 → commit DB
 → store/commit offset
```

Если DB commit прошёл, а offset commit нет, сообщение повторится — inbox защищает effect.

## 11. Batch consumption

Kafka client выдаёт records по poll. Application может формировать bounded batches по size/time. Batch увеличивает throughput, но:

- одно плохое сообщение усложняет partial success;
- latency растёт;
- memory/transaction duration растёт;
- commit position нельзя продвинуть через необработанный offset без policy.

## 12. Retry и DLQ

Immediate retry блокирует partition и сохраняет order, но задерживает следующие records. Retry topic разгружает основную partition, но меняет ordering.

DLQ record должен включать:

- original topic/partition/offset/key;
- message id/type/schema version;
- bounded safe error category;
- attempt/timestamp;
- original payload по privacy/retention policy.

DLQ — не мусорная корзина: нужны alert, ownership, replay и исправление причины.

## 13. Schema evolution

Правила:

- stable event name/version;
- additive optional fields обычно безопаснее удаления;
- consumers игнорируют неизвестные fields;
- producer не меняет meaning существующего field;
- registry/compatibility checks для Avro/Protobuf/JSON Schema;
- upcasters/adapters для старых versions при необходимости.

Не публикуйте EF entity JSON.

## 14. Transactions и exactly-once

Kafka transactions могут атомарно publish records и offsets внутри Kafka workflow. Они не делают атомарной запись в PostgreSQL. Для DB + Kafka обычно нужны outbox/inbox.

Термин exactly-once всегда уточняйте: доставка, processing effect, конкретный boundary.

## 15. Partition count

Определяют по:

- required consumer parallelism;
- throughput per partition;
- key cardinality/skew;
- retention/storage;
- broker/metadata overhead;
- future growth.

Слишком мало ограничивает scale, слишком много увеличивает overhead/rebalance complexity.

## 16. Observability

- consumer lag по partition/group;
- processing latency/error/retry/DLQ;
- rebalance count/duration;
- producer delivery/error/batch/compression;
- outbox age;
- throughput и partition skew.

Не используйте message/order id как metric label.

## 17. Сравнение со Spring Kafka

Spring даёт `KafkaTemplate`, `@KafkaListener`, listener containers и error handlers. В .NET обычно используется Confluent.Kafka + собственный hosted service/обёртка или higher-level library.

Kafka semantics одинаковы: abstractions фреймворка не отменяют poll, partitions, offsets, rebalances и duplicates.

## Самопроверка

1. Где Kafka гарантирует порядок?
2. Почему consumers больше partitions не увеличивают parallelism group?
3. Что ломается при commit до processing?
4. Почему DB + Kafka не становятся atomic от Kafka transaction?
5. Как retry topic влияет на ordering?

