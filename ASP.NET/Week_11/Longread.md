# Лекция 11 — Тестирование ASP.NET Core приложений

## 1. Что проверяет тест

Тест имеет смысл, если защищает observable behavior, invariant или важную integration. Количество mocks/coverage процентов не равно качеству.

```text
много:  unit/domain tests
       application/component tests
меньше: integration tests with real infrastructure
мало:   end-to-end/system tests
```

Пирамида — heuristic. Если business logic в SQL/integration, одних unit tests недостаточно.

## 2. Unit test domain

```csharp
[Fact]
public void Pay_DraftOrder_ChangesStatus()
{
    var order = Order.Create("A-1", 100m);

    order.Pay();

    Assert.Equal(OrderStatus.Paid, order.Status);
}
```

Arrange-Act-Assert делает intent явным. Один test может иметь несколько assertions об одном outcome.

## 3. Tests и design

Pure domain logic не требует Host/DI/database. Если каждый rule test поднимает web server, architecture слишком связана с framework.

Используйте fakes/mocks для boundary, поведение которой контролируется test:

```csharp
repository.FindAsync(id, ct).Returns(order);
```

Не mock-айте сам объект тестирования и value types.

## 4. Interaction vs state

State assertion обычно устойчивее:

```csharp
Assert.Equal(OrderStatus.Paid, saved.Status);
```

Interaction нужен, если вызов и есть contract: сообщение опубликовано один раз, transaction committed, forbidden gateway не вызван.

Слишком точная проверка internal call order делает refactoring дорогим.

## 5. Integration test с `WebApplicationFactory`

```csharp
public sealed class OrdersApiTests(
    WebApplicationFactory<Program> factory) : IClassFixture<WebApplicationFactory<Program>>
{
    [Fact]
    public async Task Get_UnknownOrder_Returns404Problem()
    {
        HttpClient client = factory.CreateClient();
        HttpResponseMessage response = await client.GetAsync($"/api/orders/{Guid.NewGuid()}");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
        Assert.Equal("application/problem+json",
            response.Content.Headers.ContentType?.MediaType);
    }
}
```

Factory запускает application in-process с test server и real middleware/routing/serialization/DI.

## 6. Замена registrations

```csharp
factory.WithWebHostBuilder(builder =>
{
    builder.ConfigureTestServices(services =>
    {
        services.RemoveAll<IPaymentClient>();
        services.AddSingleton<IPaymentClient>(fake);
    });
});
```

Удаляйте исходную registration, иначе multiple registration resolution может выбрать неожиданный service.

## 7. Database integration

SQLite/InMemory provider не повторяет PostgreSQL types, transactions, indexes, constraints и SQL translation. Для важных EF queries используйте тот же database engine через Testcontainers.

Test isolation:

- transaction rollback, если app и test разделяют connection/transaction корректно;
- schema/database per test class;
- truncate/reset tool;
- unique data per test.

Parallel tests требуют независимого state.

## 8. Testcontainers

Container запускает disposable real PostgreSQL. Lifecycle обычно class/collection fixture, чтобы не платить startup на каждый method.

Migration применяется так же, как production artifact. Readiness — не просто container started, а database accepts connections.

## 9. Authentication tests

Для authorization matrix возможен test authentication handler, создающий controlled claims. Отдельные tests обязаны проверить реальную JWT validation configuration (issuer/audience/signature/expiry), иначе ложная уверенность.

## 10. External HTTP

Mocking `HttpMessageHandler`/local stub server позволяет проверить serialized request и failures. Не mock-айте `HttpClient` methods напрямую: они не являются обычным virtual contract.

Contract tests проверяют совместимость с OpenAPI/provider sandbox, но не заменяют failure simulation.

## 11. Time

Используйте `TimeProvider` вместо `DateTime.UtcNow` в business logic:

```csharp
var fakeTime = new FakeTimeProvider(
    new DateTimeOffset(2026, 1, 1, 0, 0, 0, TimeSpan.Zero));
```

Test управляет clock без sleep.

## 12. Async tests

Всегда await operation:

```csharp
await Assert.ThrowsAsync<OrderConflictException>(
    () => service.PayAsync(id, ct));
```

Не используйте `Thread.Sleep`; применяйте signals, fake time или bounded timeout только как safety net.

## 13. E2E

E2E запускает deployed-like system и реальные dependencies по scope. Он дорог и медленнее локализует failure, поэтому выбирайте critical journeys:

- authenticate → create order → pay → observe state;
- schema migration + old/new compatibility;
- broker publish/consume с idempotency.

## 14. Сравнение со Spring tests

| Spring | ASP.NET Core |
|---|---|
| JUnit | xUnit/NUnit/MSTest |
| Mockito | NSubstitute/Moq/FakeItEasy или hand-written fakes |
| `@SpringBootTest` | `WebApplicationFactory<Program>`/host integration |
| `@WebMvcTest` | focused factory/controller tests, но точного slice attribute нет |
| Testcontainers Java | Testcontainers for .NET |

## 15. Best practices

- Test name описывает scenario/outcome.
- Domain tests быстрые и framework-free.
- Integration проверяет real serialization/routing/SQL.
- PostgreSQL behavior тестируется PostgreSQL.
- Deterministic clock/random/id generators.
- No shared mutable fixture state.
- Failure message помогает понять contract.

## Самопроверка

1. Почему EF InMemory не доказывает PostgreSQL query correctness?
2. Когда interaction assertion оправдан?
3. Что реально проверяет WebApplicationFactory?
4. Зачем test authentication и real JWT tests нужны оба?
5. Почему `Task.Delay`/sleep делает tests хрупкими?

