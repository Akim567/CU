# Конспект лекции 3 — Операторы и управление потоком

## Введение

Управление потоком определяет, какие выражения и сколько раз выполняются. Синтаксис C# знаком Java-разработчику, но современный C# активно использует expressions и pattern matching: ветвление не только выполняет команды, но и безопасно извлекает данные.

## 1. Арифметические операторы

```csharp
int a = 17;
int b = 5;

Console.WriteLine(a + b); // 22
Console.WriteLine(a - b); // 12
Console.WriteLine(a * b); // 85
Console.WriteLine(a / b); // 3: integer division
Console.WriteLine(a % b); // 2
```

Тип операндов определяет тип операции:

```csharp
double wrong = 1 / 2;      // сначала int division: 0
double right = 1.0 / 2;    // 0.5
double alsoRight = (double)1 / 2;
```

Унарные `++` и `--` отличаются значением выражения:

```csharp
int x = 10;
int before = x++; // before=10, x=11
int after = ++x;  // x=12, after=12
```

Не используйте несколько mutations одной переменной внутри сложного выражения: корректный код может быть трудно читаемым.

## 2. Сравнение и логика

```csharp
bool adult = age >= 18;
bool allowed = adult && hasTicket;
bool needsHelp = !allowed || hasAccessibilityPass;
```

`&&` и `||` выполняются с short-circuit:

```csharp
if (user is not null && user.IsActive)
{
    Console.WriteLine(user.Name);
}
```

Правая часть не вычисляется, если результат уже известен. Операторы `&` и `|` применимы и к `bool`, но вычисляют обе стороны; чаще они нужны для битовых операций.

## 3. Битовые операции

```csharp
[Flags]
enum Permission
{
    None = 0,
    Read = 1 << 0,
    Write = 1 << 1,
    Delete = 1 << 2
}

Permission access = Permission.Read | Permission.Write;
bool canWrite = (access & Permission.Write) != 0;
access &= ~Permission.Write;
```

Для `[Flags]` значения обычно задают степенями двойки. Современный код может использовать `access.HasFlag(...)`, но при интенсивных низкоуровневых вычислениях явная маска лучше показывает операцию.

## 4. `if`, `else` и guard clauses

```csharp
decimal CalculateDiscount(Customer customer, decimal total)
{
    if (customer is null)
        throw new ArgumentNullException(nameof(customer));

    if (total < 0)
        throw new ArgumentOutOfRangeException(nameof(total));

    if (!customer.IsActive)
        return 0;

    return customer.IsPremium ? total * 0.10m : total * 0.02m;
}
```

Ранние возвраты отделяют invalid/edge cases и уменьшают вложенность. Фигурные скобки рекомендуется сохранять в ветках с несколькими строками и в командном стиле проекта.

## 5. Условный оператор и null-aware операторы

```csharp
string label = score >= 60 ? "passed" : "failed";
string displayName = user?.Name ?? "anonymous";
cache[key] ??= LoadValue(key);
```

- `?.` прекращает цепочку при `null`.
- `??` выбирает правую часть, если левая `null`.
- `??=` присваивает только при `null`.

⚠️ Длинная цепочка `?.` может скрыть, где именно отсутствуют данные. Если разные причины `null` важны бизнесу, проверяйте их отдельно.

## 6. `switch` statement

```csharp
switch (command)
{
    case "start":
        Start();
        break;
    case "stop":
        Stop();
        break;
    default:
        Console.Error.WriteLine("Unknown command");
        break;
}
```

C# не разрешает обычный неявный fall-through между непустыми cases. Несколько labels можно объединить:

```csharp
case "quit":
case "exit":
    return;
```

## 7. `switch` expression

Когда ветвление вычисляет одно значение, expression обычно яснее:

```csharp
decimal rate = customer.Status switch
{
    CustomerStatus.Regular => 0.02m,
    CustomerStatus.Premium => 0.10m,
    CustomerStatus.Blocked => 0m,
    _ => throw new ArgumentOutOfRangeException()
};
```

Arms проверяются сверху вниз. `_` — discard pattern, то есть fallback.

## 8. Pattern matching

### Type pattern

```csharp
static decimal GetArea(Shape shape) => shape switch
{
    Circle c => Math.PI * c.Radius * c.Radius,
    Rectangle r => r.Width * r.Height,
    null => throw new ArgumentNullException(nameof(shape)),
    _ => throw new NotSupportedException(shape.GetType().Name)
};
```

Type pattern одновременно проверяет тип и создаёт переменную нужного типа — ручной cast не требуется.

### Relational и logical patterns

```csharp
string ClassifyTemperature(int value) => value switch
{
    < -20 => "extreme cold",
    >= -20 and < 0 => "cold",
    >= 0 and < 20 => "cool",
    >= 20 and <= 30 => "warm",
    > 30 => "hot"
};
```

### Property pattern

```csharp
bool CanShip(Order order) => order is
{
    Status: OrderStatus.Paid,
    Address.CountryCode: not null,
    Items.Count: > 0
};
```

Pattern проверяет форму объекта, не мутируя его.

### List pattern

```csharp
string Describe(int[] values) => values switch
{
    [] => "empty",
    [var only] => $"one: {only}",
    [0, ..] => "starts with zero",
    [var first, .., var last] => $"from {first} to {last}"
};
```

## 9. Циклы

### `for`

```csharp
for (int i = 0; i < items.Count; i++)
{
    Console.WriteLine($"{i}: {items[i]}");
}
```

Подходит, когда нужен индекс или точно контролируемый шаг.

### `foreach`

```csharp
foreach (Order order in orders)
{
    Process(order);
}
```

`foreach` работает через enumerable pattern/`IEnumerable<T>`. Изменение структуры обычной коллекции во время обхода часто приводит к `InvalidOperationException`.

### `while` и `do`

```csharp
while (queue.TryDequeue(out Job? job))
{
    Process(job);
}

do
{
    input = Console.ReadLine();
} while (input is not "quit");
```

`do` выполнит тело минимум один раз.

## 10. `break`, `continue`, `return`, `throw`

- `break` завершает ближайший цикл или `switch`.
- `continue` переходит к следующей итерации.
- `return` завершает метод.
- `throw` прерывает нормальный поток и запускает поиск обработчика.

```csharp
foreach (var item in items)
{
    if (item is null)
        continue;

    if (item.IsTerminal)
        break;

    Process(item);
}
```

## 11. Сравнение с Java

- Базовые `if`, `for`, `while`, `foreach` синтаксически близки.
- C# switch expressions и recursive patterns позволяют декларативно разбирать форму данных.
- В C# отсутствует неявный fall-through после непустого `case`.
- `is Type variable` соответствует проверке с безопасным связыванием переменной.
- `?.`, `??`, `??=` встроены в язык; Java обычно использует явные проверки или `Optional` в возвращаемых значениях.
- `[Flags] enum` — распространённая модель битовых комбинаций; Java чаще использует `EnumSet`.

## 12. Частые ошибки

### ❌ Assignment вместо сравнения

Для `bool` код `if (enabled = true)` компилируется и присваивает значение. Пишите `if (enabled)`.

### ❌ Слишком широкий default

Если enum должен быть исчерпывающим, молчаливый `_ => default` скрывает новые значения. Лучше бросить исключение либо явно определить forward-compatibility contract.

### ❌ Исключение как обычный цикл

Не используйте exception handling вместо `TryParse`, `TryGetValue` и других ожидаемых проверок.

## Выводы

Современный C# сочетает императивные statements и value-producing expressions. Pattern matching особенно полезен, когда он делает набор допустимых состояний явным. Главный критерий — не краткость, а предсказуемый поток управления.

## Самопроверка

1. Почему `1 / 2` нельзя исправить присваиванием результата в `double`?
2. Чем `&&` отличается от `&` для boolean operands?
3. В каком порядке проверяются arms switch expression?
4. Когда guard clauses улучшают метод?
5. Почему fallback для enum иногда опасен?

