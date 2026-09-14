# Конспект лекции 13 — Generics, constraints и variance

## 1. Проблема кода для конкретного типа

```csharp
static int First(int[] values) => values[0];
static string First(string[] values) => values[0];
```

Логика одинакова, меняется тип. `object` убрал бы дублирование ценой casts, runtime errors и boxing. Generic parameter сохраняет статическую типизацию:

```csharp
static T First<T>(IReadOnlyList<T> values)
{
    if (values.Count == 0)
        throw new ArgumentException("Sequence is empty", nameof(values));
    return values[0];
}
```

## 2. Generic types

```csharp
public sealed class Result<T>
{
    private Result(T? value, string? error, bool success) =>
        (Value, Error, IsSuccess) = (value, error, success);

    public T? Value { get; }
    public string? Error { get; }
    public bool IsSuccess { get; }

    public static Result<T> Success(T value) => new(value, null, true);
    public static Result<T> Failure(string error) => new(default, error, false);
}
```

У хорошего result type invalid combinations должны быть недоступны. Для value types/nullability production-вариант может требовать более строгого union-like design.

## 3. Несколько параметров

```csharp
public sealed record Pair<TKey, TValue>(TKey Key, TValue Value);
```

Названия `TKey`, `TValue`, `TResult` передают роль. Один `T` достаточно только при очевидном смысле.

## 4. Reified generics

CLR сохраняет generic type information в runtime:

```csharp
Console.WriteLine(typeof(List<int>));
Console.WriteLine(typeof(List<string>));
Console.WriteLine(typeof(List<int>) == typeof(List<string>)); // False
```

Можно проверять constructed generic type:

```csharp
if (value is List<int> numbers) { }
```

Runtime обычно делит machine code между reference-type instantiations и специализирует для value types, где layout различается. Детали — оптимизация runtime, не contract языка.

## 5. Constraints

```csharp
static T Create<T>() where T : new() => new T();
```

Основные constraints:

| Constraint | Смысл |
|---|---|
| `where T : class` | non-nullable reference type |
| `where T : class?` | nullable или non-nullable reference type |
| `where T : struct` | non-nullable value type |
| `where T : notnull` | type argument не должен быть nullable |
| `where T : unmanaged` | unmanaged value type |
| `where T : Base` | наследник/сам `Base` |
| `where T : IFoo` | реализует interface |
| `where T : new()` | public parameterless constructor |

Constraints дают compiler доступ к operations:

```csharp
static string Describe<T>(T value) where T : IEntity =>
    $"{typeof(T).Name}: {value.Id}";
```

## 6. Несколько constraints

```csharp
public sealed class Repository<T>
    where T : class, IEntity, new()
{
}
```

Порядок ограничений регулируется grammar: primary constraint идёт раньше interfaces, `new()` обычно последним.

## 7. Generic math

```csharp
using System.Numerics;

static T Sum<T>(IEnumerable<T> values) where T : INumber<T>
{
    T result = T.Zero;
    foreach (T value in values)
        result += value;
    return result;
}
```

Static abstract interface members позволяют вызывать `T.Zero` и `+` без `dynamic` и overload на каждый numeric type.

## 8. Инвариантность

`List<Cat>` не является `List<Animal>`:

```csharp
// List<Animal> animals = new List<Cat>(); // error
```

Иначе в список кошек можно было бы добавить собаку. Mutable generic container обычно invariant.

## 9. Covariance (`out`)

Producer может быть covariant:

```csharp
IEnumerable<Cat> cats = GetCats();
IEnumerable<Animal> animals = cats;
```

`IEnumerable<out T>` выдаёт `T`, но не принимает его для mutation.

Собственный contract:

```csharp
public interface IProducer<out T>
{
    T Produce();
}
```

Variance поддерживается для interfaces и delegates с reference type arguments. Value types не получают такого reference conversion.

## 10. Contravariance (`in`)

Consumer может быть contravariant:

```csharp
public interface IConsumer<in T>
{
    void Consume(T value);
}

IConsumer<Animal> animalConsumer = new AnimalLogger();
IConsumer<Cat> catConsumer = animalConsumer;
```

Тот, кто умеет принять любое животное, умеет принять кошку.

## 11. Generic delegates

- `Func<in ..., out TResult>` — возвращает значение;
- `Action<in ...>` — ничего не возвращает;
- `Predicate<in T>` — возвращает `bool`.

Variance этих delegates позволяет безопасные method-group conversions.

## 12. `default(T)` и nullability

```csharp
static T? FindOrDefault<T>(IEnumerable<T> source, Predicate<T> predicate)
{
    foreach (T item in source)
        if (predicate(item)) return item;
    return default;
}
```

`default` может быть `null`, zeroed struct или zero. API должен различать «не найдено» и легитимное default value. Иногда лучше `(bool Found, T Value)`, `Try...`, exception или domain option type.

## 13. Open и constructed types

`List<>` — open generic type definition, `List<int>` — constructed type. Reflection умеет исследовать оба:

```csharp
Type definition = typeof(Dictionary<,>);
Type concrete = definition.MakeGenericType(typeof(string), typeof(int));
```

C# 14 разрешает `nameof(List<>)` для unbound generic type.

## 14. Сравнение с Java

| Java | C# |
|---|---|
| type erasure в основной модели generics | runtime сохраняет constructed generic types |
| primitives нельзя напрямую type argument | value types допустимы: `List<int>` |
| use-site variance `? extends`/`? super` | declaration-site `out`/`in` для interfaces/delegates |
| bounds `extends` | `where` constraints |
| нет `new T()` | `where T : new()` позволяет |
| numeric generics сложны | generic math через static abstract interfaces |

Мнемоника Java PECS помогает смыслом producer/consumer, но синтаксис и место объявления variance отличаются.

## 15. Частые ошибки

- Делать mutable container covariant концептуально.
- Добавлять `new()` только ради service locator construction.
- Возвращать `default` без понятного absence contract.
- Использовать reflection/dynamic там, где достаточно constraint.
- Создавать слишком generic API, теряя domain language.

## Выводы

Generics C# — runtime-visible и работают с value types без wrapper на каждый элемент. Constraints должны выражать минимальные необходимые возможности, а variance — направление безопасного потока значений.

## Самопроверка

1. Почему `List<Cat>` не является `List<Animal>`?
2. Чем `out T` отличается от `out` method parameter?
3. Видит ли runtime разницу `List<int>` и `List<string>`?
4. Что позволяет `new()` constraint?
5. Почему `default(T)` может быть неоднозначным результатом?

