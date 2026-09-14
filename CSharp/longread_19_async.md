# Конспект лекции 19 — Асинхронность: `Task`, `async`/`await` и cancellation

## 1. Асинхронность не равна новому потоку

Асинхронная I/O-операция позволяет не блокировать thread, пока ОС/сеть/БД ожидает событие. Параллельность означает одновременное вычисление, часто на нескольких CPU cores. Это разные задачи.

```csharp
public async Task<string> DownloadAsync(
    HttpClient client,
    Uri uri,
    CancellationToken cancellationToken)
{
    return await client.GetStringAsync(uri, cancellationToken);
}
```

Во время ожидания метод не обязан удерживать thread.

## 2. `Task` как обещание завершения

- `Task` — операция без result;
- `Task<T>` — операция с result;
- состояния включают успешное завершение, fault и cancellation.

Вызов async method начинает выполняться синхронно до первого незавершённого await. Затем compiler-generated state machine сохраняет state и возвращает Task caller.

## 3. Что делает `await`

```csharp
Task<User?> pending = repository.FindAsync(id, ct);
User? user = await pending;
```

Если task завершён, continuation может продолжиться сразу. Если нет, метод регистрирует continuation и возвращает control caller. `await` не означает `Task.Wait()`.

## 4. Async all the way

❌ Blocking over async:

```csharp
User user = repository.FindAsync(id, ct).Result!;
```

Проблемы:

- thread блокируется;
- возможны deadlocks в synchronization-context environments;
- thread-pool starvation под нагрузкой;
- exception wrapping/diagnostics хуже.

✅ Распространяйте `async` до application boundary:

```csharp
public async Task<User> GetRequiredAsync(Guid id, CancellationToken ct)
{
    return await repository.FindAsync(id, ct)
        ?? throw new UserNotFoundException(id);
}
```

## 5. `async void`

`async void` нельзя await; caller не получает completion/error contract. Он допустим преимущественно для event handlers.

```csharp
public async Task HandleAsync() { }
```

Library/application callbacks должны возвращать `Task`.

## 6. Concurrency через `WhenAll`

Последовательное ожидание независимых операций:

```csharp
User user = await LoadUserAsync(id, ct);
Orders orders = await LoadOrdersAsync(id, ct);
```

Concurrent start:

```csharp
Task<User> userTask = LoadUserAsync(id, ct);
Task<Orders> ordersTask = LoadOrdersAsync(id, ct);

await Task.WhenAll(userTask, ordersTask);
User user = await userTask;
Orders orders = await ordersTask;
```

Так можно делать только для независимых operations и dependencies, допускающих concurrency. Один EF `DbContext` не thread-safe.

## 7. Exceptions

Exception до первого asynchronous suspension может быть captured в returned Task согласно форме метода; при `await` faulted task исходное exception повторно бросается в continuation.

```csharp
try
{
    await operation;
}
catch (HttpRequestException ex)
{
    // handle
}
```

При нескольких faults `WhenAll` хранит aggregate information в task, а поведение `await` требует внимательного diagnostics всех child tasks, если это важно.

## 8. CancellationToken

Cancellation cooperative:

```csharp
public async Task ProcessAsync(
    IEnumerable<Item> items,
    CancellationToken cancellationToken)
{
    foreach (Item item in items)
    {
        cancellationToken.ThrowIfCancellationRequested();
        await SaveAsync(item, cancellationToken);
    }
}
```

Token нужно передавать вниз во все поддерживающие I/O calls. Не создавайте `CancellationToken.None` посередине call chain без причины.

Cancellation сообщает «результат больше не нужен/операцию надо прекратить», но не откатывает уже совершённые side effects. Для atomicity нужны transaction/compensation/idempotency.

## 9. Timeout

```csharp
using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(5));
using var linked = CancellationTokenSource.CreateLinkedTokenSource(
    requestToken,
    timeout.Token);

await operation(linked.Token);
```

Timeout policy должна различать caller cancellation и истечение собственного deadline, если response/logging зависят от причины.

## 10. `ConfigureAwait`

В ASP.NET Core нет классического request `SynchronizationContext`, поэтому deadlock pattern старого ASP.NET/UI не переносится буквально. В reusable libraries `ConfigureAwait(false)` может по-прежнему выражать отсутствие требования возвращаться в captured context; в application code не превращайте его в механический шум.

## 11. `ValueTask<T>`

`ValueTask<T>` может уменьшить allocation, если операция часто завершается синхронно. Цена — более сложные правила consumption и больший state machine.

```csharp
public ValueTask<User?> FindCachedAsync(Guid id)
{
    if (_cache.TryGetValue(id, out User? user))
        return ValueTask.FromResult<User?>(user);

    return new ValueTask<User?>(LoadAsync(id));
}
```

По умолчанию возвращайте `Task<T>`. Переходите к `ValueTask<T>` после measurement и при стабильном API contract.

## 12. Async streams

```csharp
public async IAsyncEnumerable<Event> ReadEventsAsync(
    [EnumeratorCancellation] CancellationToken cancellationToken = default)
{
    while (await HasNextAsync(cancellationToken))
        yield return await ReadNextAsync(cancellationToken);
}

await foreach (Event item in ReadEventsAsync(ct).WithCancellation(ct))
{
    await HandleAsync(item, ct);
}
```

`IAsyncEnumerable<T>` позволяет асинхронно получать элементы по одному. Это не автоматическая бесконечная buffer/backpressure guarantee: producer/provider design всё ещё важен.

## 13. Resource scope

```csharp
await using Stream stream = await OpenAsync(ct);
await JsonSerializer.SerializeAsync(stream, payload, cancellationToken: ct);
```

Не возвращайте task изнутри `using` без await, если resource должен жить до завершения:

```csharp
// Опасно: stream disposed до фактического завершения returned task
Task CopyWrong(Stream input)
{
    using var output = File.Create(path);
    return input.CopyToAsync(output);
}
```

## 14. Сравнение с Java

- `Task<T>` ближе по роли к `CompletableFuture<T>`, но `async`/`await` встроены в язык и compiler state machines.
- `CancellationToken` — явный cooperative contract; Java часто использует interruption/framework cancellation.
- Virtual threads делают blocking style дешевле в ряде Java-сценариев; ASP.NET Core по-прежнему строит scalable I/O вокруг async APIs.
- `IAsyncEnumerable<T>` соответствует asynchronous pull-stream model, не Java Stream.

## 15. Best practices

- Async suffix для публичных asynchronous methods.
- Не используйте `.Result`, `.Wait()` и `Task.Run` вокруг естественного async I/O.
- Передавайте cancellation token.
- Запускайте параллельно только независимые thread-safe operations.
- Не retry cancellation.
- Не выбирайте `ValueTask` без measurement.
- Учитывайте partial side effects при отмене.

## Самопроверка

1. Создаёт ли async I/O отдельный thread на всё ожидание?
2. Почему `async void` трудно обрабатывать?
3. Когда `WhenAll` безопасен?
4. Откатывает ли cancellation выполненную запись?
5. Почему `ValueTask` не является бесплатной заменой `Task`?

