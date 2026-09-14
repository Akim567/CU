# Конспект лекции 5 — Массивы, строки и текст

## 1. Массив как объект фиксированного размера

```csharp
int[] scores = new int[3];
scores[0] = 10;
scores[1] = 20;
scores[2] = 30;
```

Массив:

- является reference type;
- хранит элементы одного совместимого типа;
- имеет неизменяемую длину;
- проверяет границы индекса;
- реализует набор collection interfaces.

```csharp
int[] compact = [10, 20, 30]; // collection expression
Console.WriteLine(compact.Length);
```

`Length` — количество элементов, последний индекс равен `Length - 1`.

## 2. Инициализация и default values

```csharp
var numbers = new int[3];       // [0, 0, 0]
var flags = new bool[2];        // [false, false]
var names = new string?[2];     // [null, null]
```

Все slots инициализируются default value типа. `new string[2]` при включённом nullable analysis создаёт неудобное состояние: runtime заполнит массив `null`, хотя element type объявлен non-nullable. Обычно массив заполняют сразу либо используют `string?[]` до завершения построения.

## 3. Одномерные, прямоугольные и jagged arrays

### Одномерный

```csharp
int[] row = [1, 2, 3];
```

### Прямоугольный

```csharp
int[,] matrix =
{
    { 1, 2, 3 },
    { 4, 5, 6 }
};

Console.WriteLine(matrix[1, 2]); // 6
Console.WriteLine(matrix.GetLength(0)); // rows
```

### Jagged

```csharp
int[][] triangle =
[
    [1],
    [2, 3],
    [4, 5, 6]
];
```

Jagged array — массив ссылок на массивы; строки могут иметь разную длину. Rectangular array — один объект с несколькими измерениями.

## 4. Индексы и диапазоны

Оператор `^` считает с конца:

```csharp
int last = scores[^1];
int beforeLast = scores[^2];
```

`^0` означает позицию после последнего элемента и не подходит для чтения.

Range `..` задаёт полуоткрытый диапазон `[start, end)`:

```csharp
int[] middle = scores[1..^1];
```

Для массива slicing создаёт новый массив и копирует элементы. Для `Span<T>` slice обычно является view без копирования.

## 5. Перебор и изменение

```csharp
for (int i = 0; i < scores.Length; i++)
{
    scores[i] *= 2;
}

foreach (int score in scores)
{
    Console.WriteLine(score);
}
```

Обычная iteration variable `foreach` не предназначена для переназначения элемента. Для mutation массива используйте индекс либо подходящий `Span<T>`.

## 6. Полезные операции

```csharp
Array.Sort(scores);
int position = Array.BinarySearch(scores, 20);
Array.Fill(scores, -1);
Array.Clear(scores);
```

`BinarySearch` требует данные, отсортированные совместимым comparer. Иначе результат не имеет смысла.

## 7. `string`: immutable reference type

```csharp
string text = "hello";
string upper = text.ToUpperInvariant();

Console.WriteLine(text);  // hello
Console.WriteLine(upper); // HELLO
```

`string` — reference type, но immutable. Операции не меняют исходную строку, а возвращают результат.

Иммутабельность даёт:

- безопасное совместное использование;
- стабильный hash code;
- возможность interning;
- предсказуемость API.

Цена — новые allocations при многократной конкатенации.

## 8. Строковые литералы

```csharp
string escaped = "line1\nline2";
string verbatim = @"C:\data\files";
string raw = """
    {
      "name": "Ada"
    }
    """;
```

- обычный literal обрабатывает escape sequences;
- verbatim literal `@""` сохраняет backslashes, а quote удваивается;
- raw string literal удобен для JSON, SQL и многострочного текста.

## 9. Интерполяция и форматирование

```csharp
decimal price = 19.5m;
int quantity = 3;
string message = $"Total: {price * quantity:F2}";
```

Format зависит от culture, если не задан provider. Для протоколов используйте invariant culture:

```csharp
using System.Globalization;

string wire = string.Create(
    CultureInfo.InvariantCulture,
    $"{price:F2}");
```

Для логирования с structured logging не интерполируйте заранее:

```csharp
logger.LogInformation("Order {OrderId} costs {Total}", orderId, total);
```

Шаблон сохраняет отдельные поля для поиска и метрик.

## 10. Сравнение строк

```csharp
bool exact = string.Equals(left, right, StringComparison.Ordinal);
bool ignoreCase = string.Equals(
    left,
    right,
    StringComparison.OrdinalIgnoreCase);
```

Выбор comparison — часть контракта:

- `Ordinal` — machine identifiers, tokens, protocol values;
- `OrdinalIgnoreCase` — case-insensitive identifiers без linguistic rules;
- culture-aware варианты — пользовательский текст, когда этого требует UX.

❌ `ToLower() == other.ToLower()` создаёт строки, зависит от culture и хуже передаёт намерение.

## 11. `null`, empty и whitespace

Это три разных состояния:

```csharp
string? missing = null;
string empty = "";
string spaces = "   ";

bool absent = string.IsNullOrWhiteSpace(spaces);
```

Не объединяйте состояния автоматически, если `null` означает «не передано», а empty — «очистить значение», как в PATCH contract.

## 12. Unicode без иллюзий

```csharp
string emoji = "😀";
Console.WriteLine(emoji.Length); // 2 UTF-16 code units
```

`Length` не равен числу grapheme clusters и иногда не равен числу Unicode scalar values.

Перебор scalar values:

```csharp
foreach (System.Text.Rune rune in emoji.EnumerateRunes())
{
    Console.WriteLine(rune.Value);
}
```

Для пользовательски воспринимаемых символов нужны text-element APIs; grapheme cluster может состоять из нескольких code points.

## 13. `StringBuilder`

```csharp
var builder = new System.Text.StringBuilder();

for (int i = 0; i < 1_000; i++)
{
    builder.Append(i).AppendLine();
}

string result = builder.ToString();
```

`StringBuilder` полезен при неизвестном количестве последовательных mutations. Для нескольких простых частей compiler/runtime часто оптимизирует concatenation; не применяйте builder автоматически.

## 14. `Span<T>` и `ReadOnlySpan<T>`

`Span<T>` — stack-only view на contiguous memory:

```csharp
Span<int> view = scores.AsSpan(1, 2);
view[0] = 99; // меняет исходный массив

ReadOnlySpan<char> prefix = text.AsSpan(0, 3);
```

Span не владеет памятью. Он только описывает участок. Благодаря ref-safety его нельзя произвольно сохранить в обычное поле или пережить async suspension.

Когда применять:

- parsing без создания substrings;
- hot paths с большим числом slices;
- interop и buffer-oriented APIs.

Когда не применять: в обычной бизнес-логике без измеренной проблемы allocations.

## 15. Сравнение с Java

| Java | C# |
|---|---|
| `array.length` | `array.Length` |
| `String` immutable | `string` immutable |
| `StringBuilder` | `StringBuilder` |
| substring создаёт отдельную строку в современных JDK | range над string также создаёт string; span даёт view |
| UTF-16 `char` | UTF-16 `char`; `Rune` для scalar values |
| `Arrays.*` | `Array.*` |
| только jagged через arrays of arrays | есть jagged и rectangular arrays |

## 16. Частые ошибки

- Off-by-one при ручных индексах.
- Предположение, что `Length` строки — число видимых символов.
- Culture-sensitive сравнение identifiers.
- Конкатенация в большом цикле.
- Возврат mutable array без копии из API, который обещает инкапсуляцию.
- Применение `Span<T>` до профилирования.

## Самопроверка

1. Чем rectangular array отличается от jagged?
2. Копирует ли `array[1..3]` элементы?
3. Почему emoji может иметь `Length == 2`?
4. Когда нужен `OrdinalIgnoreCase`?
5. Кто владеет памятью, на которую смотрит `Span<T>`?

