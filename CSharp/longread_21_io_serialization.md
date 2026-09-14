# Конспект лекции 21 — I/O, streams и JSON

## 1. Stream abstraction

`Stream` представляет последовательность bytes и capabilities: чтение, запись, seek. Он не обещает, что данные находятся в файле.

```csharp
await using FileStream input = File.OpenRead(path);
await using FileStream output = File.Create(destination);
await input.CopyToAsync(output, cancellationToken);
```

Один contract работает для files, memory, network и compression streams.

## 2. Capabilities и position

```csharp
if (stream.CanSeek)
    stream.Position = 0;
```

Network stream обычно нельзя rewind. API, читающий поток дважды, должен либо требовать seekable input, либо buffer content с лимитом.

## 3. Text поверх bytes

```csharp
await using var stream = File.OpenRead(path);
using var reader = new StreamReader(
    stream,
    Encoding.UTF8,
    detectEncodingFromByteOrderMarks: true,
    leaveOpen: false);

string content = await reader.ReadToEndAsync(cancellationToken);
```

Encoding — часть contract. «Текстовый файл» без encoding не определён полностью. UTF-8 — разумный default для новых protocols/files.

`leaveOpen` определяет ownership вложенного stream.

## 4. Не загружать большой файл целиком

```csharp
await foreach (string line in File.ReadLinesAsync(path, cancellationToken))
{
    await ProcessLineAsync(line, cancellationToken);
}
```

Streaming снижает peak memory и позволяет начать обработку раньше. Но lifetime файла продолжается до завершения enumeration/disposal.

## 5. Buffers

```csharp
byte[] buffer = ArrayPool<byte>.Shared.Rent(64 * 1024);
try
{
    int read = await stream.ReadAsync(buffer.AsMemory(), ct);
}
finally
{
    ArrayPool<byte>.Shared.Return(buffer, clearArray: true);
}
```

Pool уменьшает allocations, но rented array может быть больше запрошенного и содержать старые данные. Всегда используйте только valid slice и возвращайте buffer в `finally`. `clearArray` важен для secrets, но имеет стоимость.

## 6. `System.IO.Pipelines`

Pipelines помогают строить высокопроизводительные producer/consumer parsers с buffer management и backpressure. Это инфраструктурный инструмент Kestrel и protocols; для обычного JSON/file I/O `Stream` проще и надёжнее.

## 7. JSON serialization

```csharp
public sealed record CreateOrderRequest(Guid CustomerId, decimal Total);

var options = new JsonSerializerOptions(JsonSerializerDefaults.Web)
{
    WriteIndented = false
};

string json = JsonSerializer.Serialize(request, options);
CreateOrderRequest? copy =
    JsonSerializer.Deserialize<CreateOrderRequest>(json, options);
```

ASP.NET Core web defaults отличаются от default options, созданных без `JsonSerializerDefaults.Web`. Лучше иметь согласованную configuration.

## 8. Stream JSON

```csharp
await JsonSerializer.SerializeAsync(
    destination,
    payload,
    options,
    cancellationToken);

Payload? value = await JsonSerializer.DeserializeAsync<Payload>(
    source,
    options,
    cancellationToken);
```

Streaming API не гарантирует constant memory для любого object graph, но избегает обязательной промежуточной строки.

## 9. JSON contract

```csharp
public sealed record UserDto(
    [property: JsonPropertyName("id")] Guid Id,
    [property: JsonPropertyName("displayName")] string Name);
```

Serialization attributes связывают code с wire contract. Для независимых layers можно использовать configuration/source-generated context.

Не сериализуйте EF entities напрямую: navigation cycles, lazy loading, over-posting и случайная утечка полей делают contract нестабильным. Используйте DTO.

## 10. Source generation

```csharp
[JsonSerializable(typeof(UserDto))]
internal partial class AppJsonContext : JsonSerializerContext
{
}

string json = JsonSerializer.Serialize(dto, AppJsonContext.Default.UserDto);
```

Generated metadata снижает runtime reflection и помогает Native AOT. Нужно зарегистрировать все reachable contracts и проверить polymorphism.

## 11. Ошибки и безопасность

- ограничивайте размер input;
- не доверяйте file name из upload;
- не объединяйте пользовательский path с root без normalization/containment check;
- не десериализуйте произвольные type names;
- валидируйте DTO после syntactic parsing;
- передавайте cancellation;
- не логируйте полное тело с credentials/PII.

## 12. Atomic file replace

Для важного файла сначала пишут temporary sibling, flush-ят по требованиям и затем выполняют atomic replace/rename, поддерживаемый filesystem. Простая запись прямо в target оставляет partial file при crash.

## 13. Сравнение с Java

- `Stream` соответствует роли `InputStream`/`OutputStream`, но объединяет read/write capabilities в иерархии .NET.
- `TextReader`/`TextWriter` похожи на Reader/Writer.
- `using` соответствует try-with-resources.
- `System.Text.Json` — встроенный современный JSON stack; Java обычно выбирает библиотеку вроде Jackson.
- `Span`/`Memory` и Pipelines формируют специфичный .NET high-performance stack.

## Самопроверка

1. Можно ли всегда повторно прочитать Stream?
2. Что определяет `leaveOpen`?
3. Почему pooled buffer возвращают в `finally`?
4. Зачем API DTO вместо EF entity?
5. Что даёт JSON source generation?
