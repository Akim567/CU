# Конспект лекции 6 — Value types, reference types и nullability

## Введение: модель важнее мнемоники

Фраза «value types живут на стеке, reference types — в куче» неверна как универсальное правило. Категория типа определяет прежде всего **семантику значения и копирования**. Место хранения зависит от контекста и решений runtime/JIT.

## 1. Копирование value type

```csharp
public struct Point
{
    public int X;
    public int Y;
}

Point first = new() { X = 1, Y = 2 };
Point second = first;
second.X = 99;

Console.WriteLine(first.X); // 1
```

Присваивание копирует значение `Point`. Поля первого экземпляра не меняются.

Value type может находиться:

- как local;
- внутри объекта reference type;
- элементом массива;
- внутри другого struct;
- boxed в управляемом объекте.

## 2. Копирование ссылки

```csharp
public sealed class User
{
    public string Name { get; set; } = "";
}

User first = new() { Name = "Ada" };
User second = first;
second.Name = "Grace";

Console.WriteLine(first.Name); // Grace
```

Присваивание копирует reference. Обе переменные указывают на один объект.

Переназначение второй переменной не меняет первую:

```csharp
second = new User { Name = "Linus" };
Console.WriteLine(first.Name); // Grace
```

## 3. Stack и managed heap

Упрощённая картина:

- каждый поток имеет call stack с frames вызванных методов;
- managed objects обычно размещаются в GC heap;
- локальные значения могут находиться в stack frame, registers или быть оптимизированы;
- поля value type встраиваются в containing storage;
- массив value types хранит значения inline.

JIT вправе менять физическое размещение, если наблюдаемая семантика сохраняется. Поэтому API проектируют по ownership и copying, а оптимизируют по измерениям.

## 4. Boxing и unboxing

Boxing создаёт object representation value type:

```csharp
int number = 42;
object boxed = number;
int restored = (int)boxed;
```

При boxing копируется значение. Последующее изменение `number` не меняет boxed value.

```csharp
number = 100;
Console.WriteLine(boxed); // 42
```

Unboxing требует совместимый boxed type:

```csharp
object boxedShort = (short)10;
// int wrong = (int)boxedShort; // InvalidCastException
int ok = (short)boxedShort;
```

Числовое расширение не совмещается с unboxing в одном cast.

### Где возникает скрытый boxing

- value type преобразуется к `object` или interface;
- non-generic collection хранит значения как `object`;
- вызов API с `params object[]`;
- некоторые interface calls и interpolation paths, если JIT/API не устранили boxing.

Generic collections `List<int>` хранят `int` без boxing каждого элемента.

## 5. `null` для reference types

`null` означает отсутствие object reference. Обращение к instance member через `null` приводит к `NullReferenceException`.

```csharp
User? user = FindUser(id);

if (user is null)
{
    return NotFound();
}

Console.WriteLine(user.Name);
```

## 6. Nullable reference types — анализ, а не новый runtime type

При `<Nullable>enable</Nullable>`:

```csharp
string title = "required";
string? description = null;
```

`string` и `string?` имеют одинаковый runtime type `System.String`. `?` добавляет compiler annotations и flow analysis.

```csharp
int Length(string? value)
{
    if (value is null)
        return 0;

    return value.Length; // анализ знает: здесь non-null
}
```

Warnings не предотвращают все NRE: reflection, legacy libraries, deserialization и `null!` могут нарушить обещание.

## 7. Nullable value types

`int?` — сокращение `Nullable<int>`:

```csharp
int? age = null;
age = 30;

if (age.HasValue)
{
    Console.WriteLine(age.Value);
}
```

Обычно удобнее patterns и `??`:

```csharp
if (age is int value)
    Console.WriteLine(value);

int effectiveAge = age ?? 0;
```

Nullable value type имеет два логических компонента: наличие и значение. Это не boxed `null` плюс boxed number в обычном storage.

## 8. Lifted operators

Операторы value types «поднимаются» к nullable variants:

```csharp
int? left = 10;
int? right = null;
int? sum = left + right; // null
```

Сравнения имеют специальные правила. Не переносите интуицию SQL `NULL` на C# автоматически.

## 9. Null operators

```csharp
int length = user?.Name?.Length ?? 0;
user ??= new User();
```

Null-forgiving operator `!` подавляет warning, но не проверяет runtime value:

```csharp
string? maybe = null;
Console.WriteLine(maybe!.Length); // компилятор молчит, runtime бросит NRE
```

✅ `!` уместен, когда invariant известен человеку и недоступен анализатору, например после framework initialization. Его наличие стоит объяснить.

## 10. Nullable attributes

Для сложных контрактов доступны attributes из `System.Diagnostics.CodeAnalysis`:

```csharp
using System.Diagnostics.CodeAnalysis;

static bool TryNormalize(
    string? input,
    [NotNullWhen(true)] out string? result)
{
    result = string.IsNullOrWhiteSpace(input)
        ? null
        : input.Trim();

    return result is not null;
}
```

После успешной проверки compiler знает, что `result` non-null.

## 11. Struct design

Struct хорош, когда значение:

- небольшое;
- представляет одно логическое значение;
- не требует identity;
- желательно immutable;
- часто хранится в массивах или вложенных объектах.

```csharp
public readonly record struct Money(decimal Amount, string Currency);
```

Большой mutable struct приводит к дорогому copying и неожиданным изменениям копий.

## 12. Defensive copies

Readonly context может заставить compiler копировать mutable struct перед вызовом метода, который потенциально меняет `this`. Поэтому `readonly struct` и readonly members важны не только для стиля, но и для предсказуемой производительности.

## 13. Equality preview

Value types по умолчанию имеют value-oriented equality implementation от `ValueType`, но reflection-based default может быть не оптимален. Records генерируют value equality. Обычные classes по умолчанию сравниваются по identity. Подробно — в лекциях 9 и 14.

## 14. Сравнение с Java

| Java | C# |
|---|---|
| primitives отдельно от object hierarchy | value types входят в CTS и могут boxing |
| wrappers: `Integer`, `Boolean` | `int?`, `bool?` через `Nullable<T>` |
| generics не принимают primitives напрямую | `List<int>` без wrapper-объекта на элемент |
| nullable annotations зависят от инструментов | nullable reference types встроены в compiler flow analysis |
| классы — reference semantics | classes — reference semantics |
| records — reference types | есть `record class` и `record struct` |

Java-разработчику особенно важно не считать, что `int?` аналогичен ссылке на `Integer` во всех деталях. Boxing, generics и memory layout различаются.

## 15. Антипримеры

### ❌ Отключить nullable warnings целиком

Так теряется описание контракта API. Лучше исправлять границы и точечно документировать исключения.

### ❌ Использовать `!` после каждого вызова

Это превращает анализатор в декорацию и возвращает NRE в runtime.

### ❌ Большой mutable struct

Копии расходятся, mutations теряются, interface conversions могут boxing.

### ❌ Оптимизировать stack против heap по догадке

Сначала определите allocations и hot path профилировщиком.

## Практические выводы

- Думайте о value/reference как о semantics копирования.
- Включайте nullable analysis в каждом новом проекте.
- Используйте patterns вместо безусловного `.Value`.
- Делайте structs небольшими и immutable.
- Избегайте boxing в массовых операциях, но не усложняйте обычный код без измерений.

## Самопроверка

1. Может ли value type находиться внутри GC heap?
2. Что копируется при присваивании class variable?
3. Создаёт ли `string?` новый runtime type?
4. Почему `(int)(object)(short)1` не работает?
5. Что реально делает null-forgiving operator?

