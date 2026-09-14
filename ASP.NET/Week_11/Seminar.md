# Семинар 11 — Unit, integration и E2E testing

## Задание 1. Domain tests

Покройте state machine order:

- Draft → Paid;
- Paid → Cancelled по правилу;
- Shipped нельзя отменить;
- repeated Pay имеет определённый idempotent/conflict contract;
- total invariant.

Без DI/Host/EF.

## Задание 2. Application service

Используйте fake repository/payment gateway. Проверьте:

- missing order не вызывает gateway;
- payment failure не меняет status;
- successful payment сохраняется один раз;
- cancellation не переводится в domain failure.

## Задание 3. API factory

Через `WebApplicationFactory<Program>` проверьте 400/404/409/201 и Problem Details content type/code/traceId.

## Задание 4. DI override

Замените external PaymentClient fake. Удалите production registration и докажите, какой instance resolved.

## Задание 5. Test authentication

Создайте handler, читающий безопасные test headers и формирующий claims. Протестируйте 401, 403, owner/support scenarios. Handler существует только в test host.

## Задание 6. Real JWT validation

Отдельно выпустите tokens test key и проверьте:

- wrong issuer;
- wrong audience;
- invalid signature;
- expired;
- correct scope.

## Задание 7. PostgreSQL Testcontainer

Поднимите PostgreSQL, примените migrations и проверьте unique constraint, decimal mapping, concurrency token, keyset query.

## Задание 8. Isolation

Запустите DB tests параллельно. Исправьте collisions через unique data/schema/reset strategy. Не полагайтесь на порядок tests.

## Задание 9. HTTP stub

Fake payment server должен уметь 200/409/429/delay/invalid JSON. Проверьте outbound DTO, idempotency header и cancellation.

## Задание 10. Fake time

Проверьте token/order expiration без sleep через `TimeProvider`.

## Задание 11. E2E journey

Один тест: login → create → pay → GET paid. Оставьте минимум happy path и один critical failure; не дублируйте всю unit matrix.

## Проверка

- tests независимы и deterministic;
- async operations awaited;
- PostgreSQL не заменён in-memory provider для provider-specific поведения;
- auth matrix покрыта;
- external failures управляемы;
- failure output понятен.

