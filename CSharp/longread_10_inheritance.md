# Конспект лекции 10 — Наследование и полиморфизм

## 1. Отношение «является»

```csharp
public abstract class Payment
{
    protected Payment(decimal amount)
    {
        if (amount <= 0)
            throw new ArgumentOutOfRangeException(nameof(amount));
        Amount = amount;
    }

    public decimal Amount { get; }
    public abstract string Kind { get; }
    public virtual decimal Fee() => 0m;
}

public sealed class CardPayment(decimal amount) : Payment(amount)
{
    public override string Kind => "card";
    public override decimal Fee() => Amount * 0.02m;
}
```

`CardPayment` является `Payment` и может использоваться через base reference.

## 2. Virtual dispatch

```csharp
Payment payment = new CardPayment(100m);
Console.WriteLine(payment.Fee()); // 2.00
```

Static type переменной — `Payment`, runtime type объекта — `CardPayment`. Для `virtual` member CLR выбирает override по runtime type.

В C# member не виртуален по умолчанию. Base должен явно написать `virtual` или `abstract`, derived — `override`.

## 3. `new` hiding — не override

```csharp
public class Base
{
    public void Print() => Console.WriteLine("base");
}

public class Derived : Base
{
    public new void Print() => Console.WriteLine("derived");
}

Base value = new Derived();
value.Print(); // base
```

`new` скрывает member для static lookup и обычно создаёт путаницу. Он нужен лишь в редких compatibility scenarios.

## 4. Abstract types

Abstract class нельзя создать напрямую. Она может иметь state, constructors, implemented members и abstract contract.

```csharp
public abstract class Document
{
    public Guid Id { get; } = Guid.NewGuid();
    public abstract byte[] Render();
}
```

Abstract member не имеет implementation и обязан быть реализован concrete derived type, если тот сам не abstract.

## 5. Base constructor

Перед телом derived constructor должен завершиться base constructor:

```csharp
public sealed class PdfDocument(string title) : Document
{
    public string Title { get; } = title;
    public override byte[] Render() => [];
}
```

Не вызывайте virtual members из constructors: override может обратиться к derived state, которое ещё не инициализировано.

## 6. `protected`

`protected` расширяет API для наследников. Это долгосрочный контракт, а не просто «почти private». Derived types начинают зависеть от деталей base implementation.

Предпочитайте private state и protected operations с ясными invariants.

## 7. Upcast и downcast

Upcast безопасен:

```csharp
Payment payment = new CardPayment(100m);
```

Downcast требует проверки:

```csharp
if (payment is CardPayment card)
{
    Console.WriteLine(card.Kind);
}
```

Явный неверный cast бросит `InvalidCastException`. Оператор `as` возвращает `null` и применим к reference/nullable types:

```csharp
CardPayment? card = payment as CardPayment;
```

Patterns обычно яснее, особенно при нескольких cases.

## 8. `sealed`

```csharp
public sealed class ApiKeyCredential { }
```

Sealed class запрещает наследование. Отдельный override можно sealed:

```csharp
public sealed override string ToString() => "fixed";
```

Sealing фиксирует behavioral contract и иногда помогает JIT, но основной мотив — design.

## 9. Члены `object`

Каждый type совместим с `object`. Ключевые virtual methods:

- `ToString()`;
- `Equals(object?)`;
- `GetHashCode()`.

```csharp
public override string ToString() => $"{Kind}: {Amount}";
```

`GetType()` не virtual и возвращает точный runtime type.

Equality contracts должны сохранять reflexivity, symmetry, transitivity и согласованность с hash code. В inheritance hierarchy value equality особенно сложна; composition или sealed types часто безопаснее.

## 10. Covariant return

Override может уточнить reference return type:

```csharp
public abstract class Factory
{
    public abstract Payment Create();
}

public sealed class CardFactory : Factory
{
    public override CardPayment Create() => new(100m);
}
```

## 11. Composition over inheritance

```csharp
public interface IFeePolicy
{
    decimal Calculate(decimal amount);
}

public sealed class PaymentService(IFeePolicy policy)
{
    public decimal Total(decimal amount) => amount + policy.Calculate(amount);
}
```

Composition позволяет заменить policy без наследования service, уменьшает coupling к base internals и сочетается с DI.

Наследование оправдано, когда:

- отношение «является» стабильно;
- Liskov substitution действительно соблюдается;
- base type намеренно спроектирован для extension;
- shared implementation не является единственным аргументом.

## 12. Sealed hierarchy и patterns

C# не имеет полного аналога Java sealed hierarchy с исчерпывающей compiler-проверкой каждого switch для произвольных classes. Можно закрыть конкретные leaf types и использовать patterns, но добавление нового subtype не всегда вызовет compile error во всех switches. Не считайте `_` доказательством исчерпывающей модели.

## 13. Сравнение с Java

| Java | C# |
|---|---|
| methods virtual по умолчанию, кроме ограничений | требуется `virtual`/`abstract` + `override` |
| `final class/method` | `sealed class/override` |
| `extends` | `:` |
| `super` | `base` |
| `instanceof` patterns | `is` patterns |
| checked override exception rules | checked exceptions отсутствуют |
| sealed `permits` hierarchy | sealed types есть, но модель исчерпывающего ADT отличается |

Явность `override` в C# предотвращает случайное переопределение нового base member и предупреждает скрытие.

## 14. Частые ошибки

- Наследование используется только ради повторного использования кода.
- Base constructor вызывает virtual member.
- Derived class усиливает preconditions или ослабляет guarantees.
- `new` принимают за polymorphic override.
- Base exposes protected mutable fields.
- Equality переопределена в открытой hierarchy без продуманного symmetry contract.

## Выводы

Полиморфизм C# явный: base объявляет extension point, derived подтверждает override. `sealed` и composition должны быть нормальным выбором, а открытое наследование — осознанным публичным контрактом.

## Самопроверка

1. Почему метод без `virtual` не участвует в virtual dispatch?
2. Чем `new` отличается от `override`?
3. Почему virtual call из constructor опасен?
4. Когда `as` возвращает `null`?
5. Какие условия нужны для корректной substitution?

