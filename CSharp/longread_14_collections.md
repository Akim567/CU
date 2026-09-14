# Конспект лекции 14 — Коллекции, равенство и хеширование

## 1. Карта коллекций .NET

| Задача | Основной выбор |
|---|---|
| индексированный изменяемый список | `List<T>` |
| FIFO queue | `Queue<T>` |
| LIFO stack | `Stack<T>` |
| deque | `LinkedList<T>` или специализированное решение |
| поиск по key | `Dictionary<TKey,TValue>` |
| uniqueness | `HashSet<T>` |
| сортированный key/value | `SortedDictionary<TKey,TValue>` |
| sorted compact list | `SortedList<TKey,TValue>` |
| immutable snapshot | `ImmutableArray<T>`, `ImmutableDictionary<,>` |
| read-heavy frozen lookup | `FrozenSet<T>`, `FrozenDictionary<,>` |
| concurrent access | `ConcurrentDictionary<,>`, `Channel<T>` и другие |

Выбирают по операциям и invariants, не по привычке.

## 2. Интерфейсы

```text
IEnumerable<T>
  └─ ICollection<T>
      └─ IList<T>

IReadOnlyCollection<T>
  └─ IReadOnlyList<T>
```

`IEnumerable<T>` обещает только enumeration. `ICollection<T>` добавляет count/mutation contract, `IList<T>` — индекс. Read-only interface запрещает mutation через этот reference, но не доказывает, что underlying object никогда не меняется.

## 3. `List<T>`

`List<T>` использует внутренний массив и capacity:

```csharp
var users = new List<User>(capacity: 100);
users.Add(new User("Ada"));
users.Insert(0, new User("Grace"));
users.RemoveAt(1);
```

- индекс — O(1);
- append — amortized O(1);
- insertion/removal в середине — O(n);
- рост capacity требует нового массива и копирования.

Предварительная capacity полезна, если размер действительно известен.

## 4. Queue и Stack

```csharp
var queue = new Queue<Job>();
queue.Enqueue(job);
if (queue.TryDequeue(out Job? next)) Process(next);

var stack = new Stack<Command>();
stack.Push(command);
if (stack.TryPop(out Command? latest)) Undo(latest);
```

`Try...` избегает exception для ожидаемо пустой коллекции.

## 5. Dictionary

```csharp
var byEmail = new Dictionary<string, User>(
    StringComparer.OrdinalIgnoreCase);

byEmail[user.Email] = user;

if (byEmail.TryGetValue(input, out User? found))
{
    Console.WriteLine(found.Name);
}
```

Hash table вычисляет hash, выбирает bucket и проверяет equality при collisions. Средняя сложность lookup близка к O(1), но contract не гарантирует отсутствие pathological cases.

## 6. Equality contract

Если `Equals(x, y)` возвращает true, `GetHashCode()` обязан вернуть одинаковое значение для обоих объектов в пределах процесса.

```csharp
public sealed record EmailAddress
{
    private EmailAddress(string value) => Value = value;
    public string Value { get; }

    public static EmailAddress Parse(string value)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(value);
        return new(value.Trim().ToLowerInvariant());
    }
}
```

Normalization задаёт equality semantics. Для пользовательских email case rules могут быть бизнес-специфичны; пример не универсальная email specification.

## 7. Mutable key — авария логики

Если поля, влияющие на hash code, изменятся после insertion, dictionary будет искать key в другом bucket.

✅ Keys должны быть immutable по equality components.

## 8. Comparers

```csharp
var tags = new HashSet<string>(StringComparer.OrdinalIgnoreCase)
{
    "dotnet",
    "DOTNET"
};

Console.WriteLine(tags.Count); // 1
```

Comparer является частью коллекции. Это лучше, чем менять исходные строки только ради lookup.

Для своих типов можно реализовать `IEquatable<T>` или передать `IEqualityComparer<T>`. Records синтезируют equality, но custom business comparison всё равно иногда нужен.

## 9. Sorted collections

`SortedDictionary` обычно построен на сбалансированном дереве и даёт O(log n) update/lookup. `SortedList` хранит compact sorted arrays: lookup быстрый, insertion в середину O(n).

```csharp
var schedule = new SortedDictionary<DateTimeOffset, Job>();
```

`SortedSet<T>` поддерживает ordered unique values и range operations.

## 10. HashSet

```csharp
var required = new HashSet<string>(["read", "write"]);
var actual = new HashSet<string>(["read", "write", "admin"]);

bool enough = required.IsSubsetOf(actual);
actual.IntersectWith(required);
```

Set operations яснее вложенных `Contains` loops.

## 11. Immutable и frozen

Immutable collections возвращают новую структуру при update и могут structural sharing:

```csharp
ImmutableArray<int> first = [1, 2];
ImmutableArray<int> second = first.Add(3);
```

Frozen collections оптимизируются один раз для последующих reads:

```csharp
FrozenDictionary<string, Handler> handlers =
    source.ToFrozenDictionary(StringComparer.Ordinal);
```

- immutable — нужны функциональные updates/snapshots;
- frozen — строим один раз, много читаем;
- read-only wrapper — запрещает mutation через wrapper, но source может меняться.

## 12. Enumeration и versioning

```csharp
foreach (var item in list)
{
    // list.Remove(item); // обычно InvalidOperationException
}
```

Enumerator обычной mutable collection отслеживает version. Правильные варианты: `RemoveAll`, отдельный список изменений, reverse index loop или специализированная concurrent collection.

Fail-fast — диагностика, не thread-safety guarantee.

## 13. Collection expressions

```csharp
int[] array = [1, 2, 3];
List<int> list = [1, 2, 3];
ReadOnlySpan<int> span = [1, 2, 3];
int[] combined = [0, .. array, 4];
```

Target type определяет конкретное построение. Учитывайте allocations и overload resolution в generic/public APIs.

## 14. Concurrent collections

Обычный `Dictionary` не поддерживает concurrent writes. `ConcurrentDictionary` предоставляет атомарные compound operations:

```csharp
User cached = cache.GetOrAdd(id, static key => Load(key));
```

Factory может быть вызвана более одного раза при race; только опубликованное значение выбирается атомарно. Side effects внутри factory требуют осторожности.

## 15. Сравнение с Java

| Java | .NET |
|---|---|
| `ArrayList` | `List<T>` |
| `HashMap` | `Dictionary<TKey,TValue>` |
| `HashSet` | `HashSet<T>` |
| `ArrayDeque` | `Queue<T>`/`Stack<T>`; общего array deque в BCL долго не было |
| `TreeMap` | `SortedDictionary<TKey,TValue>` |
| `ConcurrentHashMap` | `ConcurrentDictionary<TKey,TValue>` |
| `List.of` immutable | immutable/frozen collections; collection expressions сами immutability не обещают |

.NET naming часто опускает implementation detail (`List`, а не `ArrayList`). Не делайте вывод, что интерфейсы/complexity одинаковы по одному похожему названию.

## 16. Выбор коллекции

Спросите:

1. Нужен ли порядок?
2. Нужна ли uniqueness?
3. Какой основной lookup: index, key, range?
4. Какая частота mutations?
5. Нужны ли snapshots?
6. Есть ли concurrent access?
7. Каковы equality/culture rules?

## Самопроверка

1. Почему append в `List<T>` amortized O(1)?
2. Чем read-only interface отличается от immutable collection?
3. Почему mutable dictionary key опасен?
4. Когда frozen лучше immutable?
5. Гарантирует ли fail-fast thread safety?

