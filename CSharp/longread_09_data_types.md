# Конспект лекции 9 — Records, structs, tuples и enums

## 1. Выбор формы данных

Не всякая модель должна быть class. C# предлагает несколько форм с разной семантикой:

| Форма | Категория | Равенство по умолчанию | Типичный смысл |
|---|---|---|---|
| `class` | reference type | identity | объект с состоянием и жизненным циклом |
| `record class` | reference type | по компонентам | immutable-oriented data model |
| `struct` | value type | value-based базовая реализация | небольшое значение |
| `record struct` | value type | по компонентам | компактная value model |
| tuple | value type | по элементам | локальный составной результат |
| `enum` | value type | underlying value | закрытый набор именованных констант |

## 2. Record class

```csharp
public sealed record Address(
    string Country,
    string City,
    string Street);

var first = new Address("RU", "Moscow", "Tverskaya");
var second = new Address("RU", "Moscow", "Tverskaya");

Console.WriteLine(first == second); // True
```

Positional record синтезирует constructor, init-only properties, deconstruction, value equality и readable `ToString`.

```csharp
Address moved = first with { Street = "Arbat" };
```

`with` выполняет shallow copy. Если record содержит mutable list, оба records могут ссылаться на неё.

## 3. Record не означает автоматически immutable

```csharp
public record Profile
{
    public string Name { get; set; } = "";
}
```

Это record с mutable property. Ключевое слово меняет equality/data-oriented synthesis, а не запрещает mutation.

Особенно опасно менять member, участвующий в hash code, после помещения record в `HashSet` или key position `Dictionary`.

## 4. Struct

```csharp
public readonly struct Percentage
{
    public Percentage(decimal value)
    {
        if (value is < 0 or > 100)
            throw new ArgumentOutOfRangeException(nameof(value));
        Value = value;
    }

    public decimal Value { get; }
    public decimal ApplyTo(decimal amount) => amount * Value / 100m;
}
```

Struct должен сохранять разумное поведение для `default`, потому что runtime может создать default value без вызова вашего constructor. Если `default(Percentage)` недопустим бизнесу, boundary должен это учитывать либо модель стоит изменить.

## 5. `readonly struct` и `ref struct`

`readonly struct` обещает отсутствие mutation instance state.

`ref struct` имеет ref-safety ограничения и предназначен для view над memory:

```csharp
Span<byte> buffer = stackalloc byte[128];
```

`Span<T>` — `ref struct`. Такие значения нельзя свободно boxing, хранить в heap object или использовать через обычные async suspension points. Это безопасность времени жизни, а не обычная доменная модель.

## 6. Record struct

```csharp
public readonly record struct Money(decimal Amount, string Currency)
{
    public Money Add(Money other)
    {
        if (!string.Equals(Currency, other.Currency, StringComparison.Ordinal))
            throw new InvalidOperationException("Currencies differ");

        return this with { Amount = Amount + other.Amount };
    }
}
```

Record struct сочетает value semantics и synthesized members. `readonly` защищает от mutation properties в positional form.

## 7. Tuples

```csharp
(string Name, int Count) result = ("books", 3);
Console.WriteLine(result.Name);

var (name, count) = result;
```

Имена tuple elements в основном помогают compile-time/readability и не являются надёжным долгоживущим serialization contract.

Используйте tuple для:

- private helper result;
- небольшого локального алгоритма;
- deconstruction.

Создайте отдельный type для public API, доменного смысла, validation или evolution.

## 8. Deconstruction

Собственный type может поддерживать deconstruction:

```csharp
public sealed class User
{
    public required Guid Id { get; init; }
    public required string Name { get; init; }

    public void Deconstruct(out Guid id, out string name) =>
        (id, name) = (Id, Name);
}

var (id, name) = user;
```

Не добавляйте слишком много `Deconstruct` overloads: позиционные значения хуже читаются при росте модели.

## 9. Enums

```csharp
public enum OrderStatus
{
    Draft = 0,
    Paid = 1,
    Shipped = 2,
    Cancelled = 3
}
```

Enum имеет underlying integral type. Любое значение underlying type можно cast к enum, даже если member не объявлен:

```csharp
OrderStatus unknown = (OrderStatus)999;
```

Поэтому внешние значения нужно валидировать:

```csharp
if (!Enum.IsDefined(status))
    throw new ArgumentOutOfRangeException(nameof(status));
```

## 10. Flags enums

```csharp
[Flags]
public enum FileAccessMode
{
    None = 0,
    Read = 1 << 0,
    Write = 1 << 1,
    Execute = 1 << 2,
    ReadWrite = Read | Write
}
```

Flags подходят для независимых комбинируемых возможностей, но не для mutually exclusive lifecycle states.

## 11. Equality и inheritance records

Records учитывают runtime type в equality, чтобы base и derived record с похожими components не считались случайно равными. Record hierarchies возможны, но data contracts обычно яснее с sealed records и composition.

## 12. Сравнение с Java

- Java `record` всегда class-like reference type; C# имеет record class и record struct.
- Java enum — class с instances, fields и methods; C# enum — именованный integral value type. Сложное поведение лучше вынести в extensions или отдельную модель.
- C# tuple встроен в язык через `ValueTuple` и deconstruction.
- C# struct — пользовательский value type; прямого общего аналога в Java нет.
- `with` у C# records поддерживает nondestructive mutation через shallow copy.

## 13. Практические рекомендации

- `class` — identity и изменяемый lifecycle.
- `record class` — data carrier с value equality.
- `readonly record struct` — маленькое часто копируемое значение.
- tuple — локальный результат, не долговечный contract.
- enum — небольшой закрытый набор; неизвестные внешние значения всё равно обрабатывайте.
- Не включайте mutable collection в value equality без ясного смысла.

## Самопроверка

1. Почему record не гарантирует глубокую immutability?
2. Что произойдёт с mutable record-key в dictionary?
3. Почему `default(TStruct)` важен для design?
4. Может ли enum хранить undeclared numeric value?
5. Когда tuple пора заменить record?

