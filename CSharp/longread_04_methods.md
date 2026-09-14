# Конспект лекции 4 — Методы, параметры и возвращаемые значения

## 1. Метод как контракт

Метод имеет имя, type parameters, параметры, return type, modifiers и тело. Хорошая сигнатура сообщает, какие данные нужны и какой результат гарантируется.

```csharp
public static decimal CalculateTotal(decimal price, int quantity)
{
    if (price < 0)
        throw new ArgumentOutOfRangeException(nameof(price));
    if (quantity < 0)
        throw new ArgumentOutOfRangeException(nameof(quantity));

    return price * quantity;
}
```

`return` должен быть совместим с объявленным return type. `void` означает отсутствие возвращаемого значения.

## 2. Expression-bodied members

Одно выражение можно записать компактно:

```csharp
static int Square(int value) => value * value;
```

Краткая форма полезна, пока сохраняет читаемость. Валидация, логирование и несколько шагов обычно требуют block body.

## 3. Передача аргументов по значению

По умолчанию параметр передаётся **по значению**. Копируется значение переменной:

```csharp
static void Increment(int value) => value++;

int number = 10;
Increment(number);
Console.WriteLine(number); // 10
```

Для reference type копируется ссылка. Метод может изменить объект, но переназначение локальной копии ссылки не меняет переменную вызывающего:

```csharp
static void Rename(User user) => user.Name = "Grace";

static void Replace(User user)
{
    user = new User("New"); // только локальная копия ссылки
}
```

## 4. `ref`, `out` и `in`

### `ref`

`ref` передаёт переменную по ссылке. Она должна быть инициализирована до вызова:

```csharp
static void Swap<T>(ref T left, ref T right)
{
    (left, right) = (right, left);
}
```

И вызывающий, и метод явно пишут `ref`.

### `out`

`out` предназначен для результата. Входное значение не требуется, но метод обязан присвоить параметр на каждом нормальном пути:

```csharp
if (Guid.TryParse(raw, out Guid id))
{
    Console.WriteLine(id);
}
```

### `in`

`in` передаёт readonly reference. Это может снизить copying большого struct, но не является автоматической оптимизацией:

```csharp
static decimal LengthSquared(in Vector vector) =>
    vector.X * vector.X + vector.Y * vector.Y;
```

Для маленьких structs обычная передача часто быстрее и проще. Измеряйте.

## 5. `params`

```csharp
static int Sum(params int[] values)
{
    int total = 0;
    foreach (int value in values)
        total += value;
    return total;
}

int total = Sum(1, 2, 3);
```

`params` должен быть последним параметром. Вызов может создавать массив; это важно на горячем пути. C# 13+ расширил допустимые collection types для `params`, но публичный API должен оставаться понятным и совместимым.

## 6. Именованные и optional arguments

```csharp
static void Connect(
    string host,
    int port = 443,
    bool useTls = true)
{
}

Connect("example.com", useTls: false);
Connect(host: "localhost", port: 8080, useTls: false);
```

Optional default должен быть compile-time constant. Значение встраивается в call site, поэтому изменение default в библиотеке не меняет уже скомпилированного потребителя.

Именованные аргументы повышают понятность нескольких параметров одного типа:

```csharp
Move(source: firstPath, destination: secondPath, overwrite: true);
```

## 7. Перегрузка

Методы можно перегружать по числу и типам параметров:

```csharp
static string Format(int value) => value.ToString();
static string Format(DateTime value) => value.ToString("O");
```

Return type не участвует в выборе overload:

```csharp
// Нельзя объявить одновременно:
// int Parse(string text)
// decimal Parse(string text)
```

Опасайтесь неоднозначностей с optional parameters, `null`, implicit conversions и generics.

## 8. Tuples и несколько результатов

```csharp
static (int Min, int Max) FindRange(IEnumerable<int> values)
{
    using IEnumerator<int> iterator = values.GetEnumerator();
    if (!iterator.MoveNext())
        throw new ArgumentException("Sequence is empty", nameof(values));

    int min = iterator.Current;
    int max = iterator.Current;

    while (iterator.MoveNext())
    {
        min = Math.Min(min, iterator.Current);
        max = Math.Max(max, iterator.Current);
    }

    return (min, max);
}

var (min, max) = FindRange([3, 1, 9]);
```

Value tuples удобны для локального, очевидного результата. Если результат является доменной сущностью, развивается независимо или требует invariants, создайте именованный `record`.

## 9. Local functions

```csharp
static int Factorial(int value)
{
    if (value < 0)
        throw new ArgumentOutOfRangeException(nameof(value));

    return Calculate(value);

    static int Calculate(int n) => n <= 1 ? 1 : n * Calculate(n - 1);
}
```

Local function скрывает helper внутри единственного consumer. `static` запрещает неявный захват внешних переменных, уменьшая риск closure allocation.

## 10. Extension methods и extension members

Классический extension method:

```csharp
public static class StringExtensions
{
    public static bool HasValue(this string? value) =>
        !string.IsNullOrWhiteSpace(value);
}
```

Использование выглядит как instance call:

```csharp
if (name.HasValue()) { }
```

Это статический вызов, выбранный компилятором. Extension не получает доступ к private state и не переопределяет настоящий instance member.

C# 14 добавляет **extension blocks**, включая extension properties и static extension members:

```csharp
public static class SequenceExtensions
{
    extension<T>(IEnumerable<T> source)
    {
        public bool IsEmpty => !source.Any();
    }
}
```

Используйте новый синтаксис, когда он действительно улучшает API; для библиотек учитывайте language version потребителей.

## 11. Generic methods

```csharp
static T? FirstOrDefault<T>(IEnumerable<T> source)
{
    foreach (T item in source)
        return item;
    return default;
}
```

Компилятор часто выводит `T` из аргумента. Constraints и variance рассматриваются отдельно.

## 12. Method group и delegate

Имя метода можно преобразовать в подходящий delegate:

```csharp
static bool IsEven(int value) => value % 2 == 0;

Predicate<int> predicate = IsEven;
Console.WriteLine(predicate(4));
```

Это основа callbacks, LINQ, events и многих ASP.NET Core APIs.

## 13. Argument validation

```csharp
static User Load(string id)
{
    ArgumentException.ThrowIfNullOrWhiteSpace(id);
    // ...
    return new User(id);
}
```

Используйте стандартные exception types и helper methods на boundary. Но не дублируйте проверки во всех приватных функциях, если invariant уже гарантирован вызывающим кодом.

## 14. Возврат ссылок и `ref` safety

C# поддерживает `ref return` и `ref local` для сценариев без копирования:

```csharp
static ref int Find(int[] values, int target)
{
    for (int i = 0; i < values.Length; i++)
    {
        if (values[i] == target)
            return ref values[i];
    }

    throw new KeyNotFoundException();
}

ref int place = ref Find(numbers, 10);
place = 20;
```

Это продвинутый инструмент: возвращаемая ссылка не должна пережить storage, на который указывает. Компилятор применяет ref-safety rules.

## 15. Сравнение с Java

| Возможность | Java | C# |
|---|---|---|
| pass-by-value | всегда, включая копию object reference | по умолчанию; дополнительно `ref`/`out`/`in` |
| optional/named args | нет общего механизма | встроены в язык |
| несколько результатов | record/class/array | value tuple или отдельный type |
| extension methods | нет | есть; C# 14 расширяет модель |
| local functions | локальные/анонимные альтернативы ограничены | именованные local functions |
| checked exceptions в сигнатуре | есть | нет |

Важно: Java тоже передаёт ссылку по значению, а не «объект по ссылке». C# default ведёт себя так же; только `ref` меняет саму переменную вызывающего.

## 16. Частые ошибки

- `ref` используется для сокрытия плохо спроектированного результата вместо return type.
- публичный метод имеет много optional booleans: вызов трудно понимать без имён;
- overloads различаются неочевидными numeric conversions;
- tuple из пяти полей заменяет нормальную доменную модель;
- extension method становится свалкой несвязанных helpers;
- async-метод объявлен `async void` вне event handler.

## Практические выводы

- По умолчанию возвращайте результат через `return`.
- Применяйте `Try...` + `out` для ожидаемого отсутствия/невалидности, если это idiom API.
- Используйте named arguments для повышения ясности, не для компенсации чрезмерной сигнатуры.
- Выбирайте record вместо tuple, если данные имеют долгоживущий смысл.
- Делайте local function `static`, если захват не нужен.

## Самопроверка

1. Что именно копируется при передаче reference type без `ref`?
2. Чем контракт `out` отличается от `ref`?
3. Почему нельзя перегрузить метод только по return type?
4. Когда tuple уступает record?
5. Почему изменение optional default в библиотеке может не повлиять на старого клиента?
