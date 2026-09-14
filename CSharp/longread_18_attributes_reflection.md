# Конспект лекции 18 — Attributes, reflection и `dynamic`

## 1. Metadata-driven code

Assembly содержит IL и metadata о types/members. Attributes добавляют декларативные metadata, а reflection позволяет исследовать их во время выполнения.

```csharp
[Obsolete("Use CreateAsync instead")]
public void Create() { }
```

Compiler распознаёт `ObsoleteAttribute` и выдаёт diagnostic.

## 2. Attribute — type

```csharp
[AttributeUsage(AttributeTargets.Method, AllowMultiple = false, Inherited = true)]
public sealed class AuditAttribute : Attribute
{
    public AuditAttribute(string action) => Action = action;
    public string Action { get; }
    public bool IncludeArguments { get; init; }
}
```

Применение:

```csharp
[Audit("order.create", IncludeArguments = false)]
public Task CreateOrderAsync() => Task.CompletedTask;
```

Constructor arguments и named property values должны быть representable в attribute metadata: constants, `typeof`, enum, arrays допустимых values и т. п.

## 3. `AttributeUsage`

- `ValidOn` ограничивает targets;
- `AllowMultiple` разрешает несколько экземпляров;
- `Inherited` влияет на поиск у derived types/methods согласно reflection API.

Attribute сам ничего не делает. Нужен consumer: compiler, analyzer, source generator, runtime framework или ваш код.

## 4. Получение `Type`

```csharp
Type compileTime = typeof(User);
Type runtime = user.GetType();
Type? byName = Type.GetType("MyApp.User, MyApp");
```

- `typeof` не требует instance;
- `GetType()` возвращает exact runtime type;
- string lookup хрупок к assembly/name/version и может вернуть `null`.

## 5. Members и BindingFlags

```csharp
PropertyInfo[] properties = typeof(User).GetProperties(
    BindingFlags.Instance | BindingFlags.Public);

MethodInfo? method = typeof(User).GetMethod(
    nameof(User.Rename),
    [typeof(string)]);
```

Явно задавайте binding scope. Иначе inherited/public/static overloads могут дать неожиданный результат или ambiguity.

## 6. Чтение attributes

```csharp
MethodInfo method = typeof(OrderService)
    .GetMethod(nameof(OrderService.CreateOrderAsync))!;

AuditAttribute? audit = method.GetCustomAttribute<AuditAttribute>();

if (audit is not null)
    Console.WriteLine(audit.Action);
```

`CustomAttributeData` позволяет исследовать metadata без создания attribute instances, что важно для tooling.

## 7. Создание и invocation

```csharp
Type type = typeof(User);
object instance = Activator.CreateInstance(type, "Ada")
    ?? throw new InvalidOperationException();

MethodInfo rename = type.GetMethod("Rename")
    ?? throw new MissingMethodException();

rename.Invoke(instance, ["Grace"]);
```

Reflection переносит многие ошибки в runtime, использует object arrays/boxing и оборачивает exception target method в `TargetInvocationException`.

## 8. Generic reflection

```csharp
Type listType = typeof(List<>);
Type closed = listType.MakeGenericType(typeof(User));
object list = Activator.CreateInstance(closed)!;
```

Можно исследовать generic arguments и constraints:

```csharp
foreach (Type argument in closed.GetGenericArguments())
    Console.WriteLine(argument);
```

## 9. Performance и caching

Reflection lookup дороже прямого вызова. В hot path cache `Type`/`MemberInfo` или создавайте compiled delegate:

```csharp
MethodInfo method = typeof(User).GetMethod(nameof(User.Rename))!;
var rename = method.CreateDelegate<Action<User, string>>();
rename(user, "Grace");
```

Не оптимизируйте редкую startup discovery без измерений.

## 10. `dynamic`

```csharp
dynamic plugin = LoadPlugin();
plugin.Execute("input");
```

Runtime binder выбирает member. Ошибки типа отсутствующего method проявятся при выполнении. `dynamic` распространяется по expression, если результат не присвоен statically typed variable.

Подходящие scenarios:

- COM interop;
- dynamic-language interop;
- узкий adapter над runtime-defined object;
- tooling/prototyping.

Не подходит для core domain и обычного JSON. Для JSON используйте DTO, DOM (`JsonDocument`/`JsonNode`) или explicit mapping.

## 11. Reflection, source generators и trimming

Runtime reflection удобна, но:

- ошибки поздние;
- metadata lookup может быть дорогим;
- Native AOT/trimming не всегда знает, какие members надо сохранить;
- private reflection связывает код с implementation details.

Source generator анализирует compilation и генерирует strongly typed code заранее. Для serialization, DI registration или mapping это может убрать runtime discovery и улучшить AOT compatibility.

Если reflection всё же нужна в trimmed app, используйте предусмотренные annotations/descriptors и тестируйте published artifact.

## 12. Assembly loading

```csharp
Assembly assembly = Assembly.LoadFrom(path);
Type[] exported = assembly.GetExportedTypes();
```

Plugin loading требует понимания `AssemblyLoadContext`, dependency isolation и unloadability. Один и тот же type name из разных load contexts не гарантирует type identity.

Не загружайте недоверенный код в процесс, считая reflection sandbox. Современный .NET не предоставляет старую Code Access Security модель как безопасную изоляцию. Нужен process/container boundary.

## 13. Сравнение с Java

| Java | C#/.NET |
|---|---|
| annotation | attribute |
| `Class<?>` | `System.Type` |
| `Method.invoke` | `MethodInfo.Invoke` |
| annotation processors | Roslyn source generators/analyzers |
| MethodHandles | delegates/expression compilation/dynamic methods в соответствующих сценариях |
| ClassLoader identity | AssemblyLoadContext участвует в loading/isolation |

.NET generics reified, поэтому reflection видит конкретные arguments `List<int>`. В Java основная модель сталкивается с erasure.

## 14. Best practices

- Определите consumer attribute до её создания.
- Используйте `nameof`, а не строку member name, где возможно.
- Cache repeated reflection lookup.
- На hot path создавайте typed delegate.
- Не используйте private reflection как стабильный API.
- Для AOT/trimming предпочитайте source generation.
- Не считайте process-internal loading security boundary.

## Самопроверка

1. Выполняет ли attribute действие самостоятельно?
2. Чем `typeof(T)` отличается от `obj.GetType()`?
3. Почему `MethodInfo.Invoke` сложнее прямого вызова?
4. Когда `CustomAttributeData` полезнее создания attribute?
5. Почему reflection конфликтует с trimming?

