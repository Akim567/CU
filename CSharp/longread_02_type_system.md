# Конспект лекции 2 — Система типов, переменные и преобразования

## Цели

После лекции вы сможете объяснить, что означает статический тип выражения, выбирать числовой тип, отличать `var` от `dynamic`, выполнять безопасные преобразования и предсказывать поведение arithmetic overflow.

## 1. Зачем языку типы

Тип задаёт:

- множество допустимых значений;
- разрешённые операции;
- представление данных;
- правила преобразования;
- контракт между вызывающим и вызываемым кодом.

```csharp
int count = 10;
string title = "CLR";
bool published = false;
```

Компилятор знает статический тип каждой переменной и отклоняет бессмысленные операции до запуска.

```csharp
// int result = title - count; // compile-time error
```

## 2. Common Type System

Все типы .NET участвуют в общей системе типов — **CTS**. На верхнем уровне находится `System.Object`. Однако из этого не следует, что все значения всегда представлены объектом в куче. Value types имеют собственную семантику хранения и копирования; преобразование value type к `object` обычно вызывает boxing.

Две большие категории:

- **value types**: числовые типы, `bool`, `char`, `struct`, `enum`, nullable value types;
- **reference types**: `class`, `record class`, `interface`, `delegate`, массивы, `string`.

Подробная семантика копирования рассматривается в лекции 6.

## 3. Встроенные типы и CLR-типы

Ключевые слова C# являются псевдонимами типов BCL:

| C# | CLR/BCL | Размер | Пример |
|---|---|---:|---|
| `sbyte` | `System.SByte` | 8 бит | `-10` |
| `byte` | `System.Byte` | 8 бит | `255` |
| `short` | `System.Int16` | 16 бит | `-300` |
| `ushort` | `System.UInt16` | 16 бит | `600` |
| `int` | `System.Int32` | 32 бита | `42` |
| `uint` | `System.UInt32` | 32 бита | `42U` |
| `long` | `System.Int64` | 64 бита | `42L` |
| `ulong` | `System.UInt64` | 64 бита | `42UL` |
| `nint`/`nuint` | native-sized integer | размер указателя | индексы native API |
| `float` | `System.Single` | 32 бита | `1.5F` |
| `double` | `System.Double` | 64 бита | `1.5` |
| `decimal` | `System.Decimal` | 128 бит | `1.5M` |
| `bool` | `System.Boolean` | логический | `true` |
| `char` | `System.Char` | 16 бит | `'A'` |
| `string` | `System.String` | reference | `"text"` |
| `object` | `System.Object` | reference | любое значение после допустимого преобразования |

### `decimal`, `double` и деньги

`double` представляет двоичную плавающую точку и подходит для научных расчётов, измерений и графики. `decimal` хранит десятичную мантиссу и обычно предпочтительнее для денежных величин.

```csharp
double binary = 0.1 + 0.2;
decimal money = 0.1m + 0.2m;

Console.WriteLine(binary == 0.3); // обычно False
Console.WriteLine(money == 0.3m); // True
```

⚠️ Выбор `decimal` не решает правила округления бизнеса. Способ и момент округления всё равно должны быть определены явно.

### `char` — не всегда пользовательский символ

`char` представляет один UTF-16 code unit. Некоторые Unicode-символы занимают surrogate pair из двух `char`. Для обработки Unicode scalar values существует `System.Text.Rune`.

## 4. Объявление и инициализация

```csharp
int attempts = 3;
string message = "ready";
DateTime startedAt = DateTime.UtcNow;
```

Локальная переменная должна быть определённо присвоена до чтения:

```csharp
int result;
// Console.WriteLine(result); // ошибка компиляции
result = 42;
```

Поля получают default values, но полагаться на неявный `0` или `null` вместо явного инварианта часто неудачно.

## 5. `var`: вывод статического типа

```csharp
var users = new List<string>(); // List<string>
var total = 10m;                // decimal
var name = "Ada";               // string
```

`var` не означает «тип можно менять». Тип выводится один раз при компиляции.

```csharp
var value = 10;
// value = "ten"; // int нельзя присваивание string
```

Хорошее применение — когда тип очевиден справа или слишком громоздок. Явный тип полезен, когда он передаёт смысл и не виден из выражения.

```csharp
Dictionary<string, List<Order>> ordersByCustomer = LoadOrders();
var ordersByCustomer2 = LoadOrders();
```

Вторая строка читается только при понятном имени метода и переменной.

## 6. `const`, `readonly` и неизменяемость

`const` — compile-time constant:

```csharp
const int MaxAttempts = 5;
const string Protocol = "https";
```

Значение `const` встраивается в код потребителя. Публичные constants между независимо обновляемыми assemblies требуют осторожности.

`readonly` применяется к полю и разрешает присваивание в объявлении или конструкторе:

```csharp
public sealed class RetryPolicy
{
    private readonly TimeSpan _delay;

    public RetryPolicy(TimeSpan delay)
    {
        _delay = delay;
    }
}
```

`readonly` запрещает заменить ссылку, но не делает сам объект глубоко неизменяемым.

## 7. Литералы

```csharp
int million = 1_000_000;
int binary = 0b_1010;
int hex = 0x_FF;
long distance = 9_000_000_000L;
uint mask = 0xFFFFU;
decimal price = 19.99m;
float ratio = 0.5f;
```

Integer literal без suffix выбирает первый подходящий тип по правилам языка. Для публичных контрактов лучше указывать ожидаемый тип явно.

## 8. Неявные и явные преобразования

Безопасное расширение обычно не требует cast:

```csharp
int count = 100;
long wide = count;
double approximate = wide;
```

Потенциально теряющее данные преобразование требует явного cast:

```csharp
double source = 42.9;
int truncated = (int)source; // 42, не округление
```

Для округления используйте выбранное правило:

```csharp
int rounded = checked((int)Math.Round(source, MidpointRounding.AwayFromZero));
```

## 9. `checked` и overflow

В unchecked integer arithmetic переполнение может «обернуться» по диапазону:

```csharp
int max = int.MaxValue;
int wrapped = unchecked(max + 1);
Console.WriteLine(wrapped); // int.MinValue
```

`checked` требует исключения `OverflowException`:

```csharp
int next = checked(max + 1);
```

Overflow context можно задавать блоком или настройкой проекта. На границах данных — размеры, деньги, внешние числа — молчаливое переполнение особенно опасно.

## 10. Парсинг и форматирование

```csharp
using System.Globalization;

string raw = "1234.50";
bool ok = decimal.TryParse(
    raw,
    NumberStyles.Number,
    CultureInfo.InvariantCulture,
    out decimal amount);
```

Культура — часть контракта. Точка и запятая означают разное в разных locales. Для machine-to-machine форматов обычно используют invariant culture или стандартный сериализатор.

```csharp
string transport = amount.ToString("F2", CultureInfo.InvariantCulture);
```

## 11. `object` и `dynamic`

При `object` компилятор разрешает только члены, известные у `object`, пока значение не будет проверено и преобразовано:

```csharp
object value = "hello";

if (value is string text)
{
    Console.WriteLine(text.Length);
}
```

`dynamic` откладывает разрешение операции до runtime:

```csharp
dynamic data = "hello";
Console.WriteLine(data.Length); // работает
data = 10;
// Console.WriteLine(data.Length); // RuntimeBinderException
```

`dynamic` полезен на границах с динамическими API, COM или некоторыми serializers, но убирает compile-time safety. Это точечный инструмент, а не замена нормальной модели типов.

## 12. Default values

```csharp
Console.WriteLine(default(int));       // 0
Console.WriteLine(default(bool));      // False
Console.WriteLine(default(DateTime));  // 01.01.0001 ...
string? text = default;                // null
```

В generic-коде используется `default(T)`, а современный синтаксис часто позволяет `default` без типа.

⚠️ Default value может быть формально допустим, но семантически некорректен. Например, default `DateTime` редко означает реальную дату.

## 13. Сравнение с Java

| Java | C# |
|---|---|
| только signed integer primitives | есть signed и unsigned integer types |
| `BigDecimal` для точной десятичной арифметики | встроенный value type `decimal` |
| локальный `var` — вывод типа | `var` — вывод типа |
| отдельные primitives и wrappers | единая CTS; value types могут boxing в `object` |
| `Object` | `object`/`System.Object` |
| нет общего `dynamic` keyword | `dynamic` с runtime binding |
| overflow обычно wraparound | `checked`/`unchecked` управляют проверкой |

Ключевое отличие: C# позволяет вызывать методы непосредственно у value type (`42.ToString()`), потому что они являются полноценными CTS types. Это не означает, что каждое такое значение уже boxed.

## 14. Практические рекомендации

- Включайте `<Nullable>enable</Nullable>` с первого проекта.
- Для денег начинайте с `decimal`, затем фиксируйте currency и rounding policy.
- Используйте `TryParse` для ожидаемо невалидного внешнего ввода.
- Не применяйте `dynamic` внутри основной доменной модели.
- Включайте checked arithmetic там, где переполнение нарушает инварианты.
- Не делайте публичный mutable state только ради удобства.

## Самопроверка

1. Почему `var` сохраняет статическую типизацию?
2. Чем `decimal` принципиально отличается от `double`?
3. Когда conversion должен быть explicit?
4. Что произойдёт при boxing `int` в `object`?
5. Почему `readonly List<int>` не является immutable collection?

## Полезные материалы

- [Built-in types](https://learn.microsoft.com/dotnet/csharp/language-reference/builtin-types/built-in-types)
- [Casting and conversions](https://learn.microsoft.com/dotnet/csharp/programming-guide/types/casting-and-type-conversions)
- [Nullable reference types](https://learn.microsoft.com/dotnet/csharp/nullable-references)
- [`checked` and `unchecked`](https://learn.microsoft.com/dotnet/csharp/language-reference/statements/checked-and-unchecked)
