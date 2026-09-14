# Конспект лекции 15 — Delegates, lambdas и events

## 1. Поведение как значение

Delegate — типобезопасная ссылка на один или несколько методов совместимой сигнатуры.

```csharp
public delegate decimal Discount(decimal total);

static decimal Premium(decimal total) => total * 0.10m;

Discount discount = Premium;
decimal value = discount(1_000m);
```

Delegate хранит method и, для instance method, target object.

## 2. Стандартные delegates

```csharp
Func<int, int, int> add = (left, right) => left + right;
Action<string> print = Console.WriteLine;
Predicate<int> positive = value => value > 0;
```

- `Func<...>` возвращает последний type parameter;
- `Action<...>` возвращает `void`;
- `Predicate<T>` представляет `bool (T)`.

Собственный delegate type нужен, если имя является частью domain API, нужны `ref` parameters или специализированные annotations.

## 3. Lambda syntax

```csharp
Func<int, int> square = x => x * x;
Func<int, int, int> max = (x, y) => x > y ? x : y;
Action announce = () => Console.WriteLine("ready");

Func<int, int> absolute = value =>
{
    if (value == int.MinValue)
        throw new OverflowException();
    return Math.Abs(value);
};
```

Expression lambda вычисляет expression; statement lambda содержит block.

C# 14 позволяет modifiers у simple lambda parameters без обязательного явного типа в поддерживаемых формах, но сложная сигнатура часто лучше читается через named method.

## 4. Method groups

```csharp
static bool IsActive(User user) => user.IsActive;

Func<User, bool> predicate = IsActive;
var active = users.Where(IsActive);
```

Compiler выбирает подходящий overload по target delegate type. При неоднозначности lambda с explicit parameters делает выбор ясным.

## 5. Closures

```csharp
int threshold = 10;
Func<int, bool> greater = value => value > threshold;

threshold = 20;
Console.WriteLine(greater(15)); // False
```

Lambda захватывает переменную, а не snapshot значения. Compiler может вынести captured locals в closure object.

Классическая ошибка цикла сегодня зависит от формы declaration, но mutation общего captured state всё равно опасна:

```csharp
var actions = new List<Action>();
for (int i = 0; i < 3; i++)
{
    int copy = i;
    actions.Add(() => Console.WriteLine(copy));
}
```

Static lambda запрещает capture:

```csharp
Func<int, int> square = static x => x * x;
```

## 6. Multicast delegates

```csharp
Action<string> handlers = WriteConsole;
handlers += WriteAudit;
handlers("created");
handlers -= WriteAudit;
```

Invocation идёт по invocation list. Если handler бросает exception, следующие обычно не вызываются. Для non-void multicast delegate возвращается результат последнего handler — это редко хороший business contract.

## 7. Events

Event ограничивает управление delegate: внешний consumer может только подписаться или отписаться.

```csharp
public sealed class OrderService
{
    public event EventHandler<OrderCreatedEventArgs>? OrderCreated;

    public void Create(Order order)
    {
        // save...
        OrderCreated?.Invoke(this, new OrderCreatedEventArgs(order.Id));
    }
}

public sealed class OrderCreatedEventArgs(Guid orderId) : EventArgs
{
    public Guid OrderId { get; } = orderId;
}
```

Только declaring type может вызвать event напрямую.

## 8. Подписки и утечки

Publisher хранит ссылки на subscribers через delegate. Если long-lived publisher переживает subscriber, забытая подписка удерживает объект в памяти.

```csharp
service.OrderCreated += OnOrderCreated;
try
{
    // use subscriber
}
finally
{
    service.OrderCreated -= OnOrderCreated;
}
```

UI и long-lived singleton events особенно требуют lifecycle discipline. Возможны disposable subscription abstractions или weak-event patterns.

## 9. Async delegates и events

`async void` допустим прежде всего для event handlers, потому что event contract возвращает void:

```csharp
private async void OnClicked(object? sender, EventArgs args)
{
    try
    {
        await SaveAsync();
    }
    catch (Exception ex)
    {
        ShowError(ex);
    }
}
```

Для application callbacks предпочитайте `Func<CancellationToken, Task>`, чтобы caller мог await и обработать exception.

Обычный multicast `Func<Task>` не await-ит handlers как группу автоматически. Для async event bus определите явный iteration/error policy.

## 10. Variance

Delegates могут объявлять `in`/`out` type parameters. `Func<Animal>` можно использовать там, где ожидается producer `object`; `Action<Animal>` — там, где consumer `Cat`.

## 11. Expression tree boundary

Одна lambda может компилироваться либо в executable delegate, либо в data structure:

```csharp
Func<User, bool> code = user => user.IsActive;
Expression<Func<User, bool>> tree = user => user.IsActive;
```

Первое вызывается CLR. Второе можно исследовать и переводить, например, в SQL. Не каждый C# construct представим/поддерживается provider.

## 12. Сравнение с Java

| Java | C# |
|---|---|
| functional interface | delegate type |
| `Function`, `Consumer`, `Predicate` | `Func`, `Action`, `Predicate` |
| method reference `::` | method group |
| effectively final local capture | captured variable может мутировать |
| listener interfaces | language-level events/delegates |
| lambdas реализуют interface target | lambdas convert to delegate или expression tree |

## 13. Практические рекомендации

- Используйте named method для сложной логики.
- Делайте lambda `static`, если capture не нужен.
- Не полагайтесь на return multicast delegate.
- Определяйте error/ordering policy для набора handlers.
- Управляйте временем жизни подписок.
- Для async callback возвращайте `Task`, а не `void`.

## Самопроверка

1. Что хранит delegate instance?
2. Захватывает lambda значение или переменную?
3. Почему event безопаснее public delegate field?
4. Кто удерживает subscriber в памяти?
5. Чем delegate lambda отличается от expression-tree lambda?

