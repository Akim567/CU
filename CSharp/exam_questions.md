# Экзаменационные вопросы по C# и .NET

## Блок 1. Основы языка

1. C#, .NET SDK, runtime, CLR, BCL, Roslyn: роли и границы ответственности.
2. Путь от `.cs` до выполнения метода: project evaluation, compilation, IL, metadata, loading, JIT.
3. Built-in types. Signed/unsigned integers, floating point и `decimal`.
4. `var`, `object` и `dynamic`: static/runtime binding и типичные ошибки.
5. Implicit/explicit conversions, parsing, `checked` и `unchecked`.
6. Statements и expressions. `if`, switch statement, switch expression.
7. Type, relational, property и list patterns. Порядок сопоставления.
8. Циклы, `break`, `continue`, guard clauses и short-circuit logic.
9. Методы: overload resolution, optional/named parameters.
10. Передача по значению. `ref`, `out`, `in`, `params`.
11. Tuples, deconstruction, local functions и extension members.
12. Массивы: одномерные, rectangular, jagged, indices и ranges.
13. `string`, immutability, interpolation, culture и Unicode.
14. `Span<T>`/`ReadOnlySpan<T>`: view, ownership и ref-safety.

## Блок 2. Типы и объектная модель

15. Value/reference semantics без упрощения «stack/heap».
16. Boxing/unboxing. Источники скрытого boxing.
17. Nullable value types и nullable reference analysis.
18. Classes, fields, constructors, primary constructors и object initializers.
19. Access modifiers, assembly boundary, nested и partial types.
20. Properties, backing field, `field`, `init`, `required` и indexers.
21. Иммутабельность: shallow/deep, defensive copy, immutable collections.
22. Class, record class, struct, record struct и tuple: критерии выбора.
23. Record equality, `with`, default struct value.
24. Enum и flags enum. Валидация внешних значений.
25. Наследование, virtual dispatch, `virtual`, `abstract`, `override`, `new`.
26. Constructors и inheritance. Опасность virtual call из constructor.
27. Upcast, downcast, `is`, `as`, pattern matching.
28. `sealed`, protected API, Liskov substitution.
29. Члены `object`: `ToString`, `Equals`, `GetHashCode`, `GetType`.
30. Composition over inheritance.
31. Interfaces, explicit implementation и default members.
32. Static abstract interface members и generic math.
33. Extension methods/blocks, operators и conversion operators.

## Блок 3. Ошибки, generics и коллекции

34. Exception hierarchy и выбор стандартного/custom exception.
35. Stack unwinding, `try`/`catch`/`finally`, filters.
36. `throw;` против `throw ex;`, inner exception и logging boundaries.
37. `IDisposable`, `IAsyncDisposable`, `using`, ownership и finalization.
38. Generic methods/types, constraints и runtime reification.
39. Invariance, covariance и contravariance.
40. Отличия generics C# от Java type erasure/wildcards.
41. Interfaces коллекций и критерии выбора implementation.
42. Внутренняя модель `List<T>`, `Dictionary<TKey,TValue>`, `HashSet<T>`.
43. Equality/hash contract, comparers и mutable keys.
44. Read-only, immutable, frozen и concurrent collections.
45. Enumeration versioning и изменение коллекции при обходе.

## Блок 4. Функциональный и динамический C#

46. Delegate, method group, lambda и closure.
47. Multicast delegates, events и lifetime подписок.
48. Async callbacks и проблема `async void`.
49. LINQ to Objects, deferred execution и materialization.
50. `Where`, `Select`, `SelectMany`, ordering, grouping, aggregation.
51. `First`/`Single` variants и absence/uniqueness contracts.
52. Iterators, `yield`, resource lifetime и repeated enumeration.
53. `IEnumerable<T>` против `IQueryable<T>`.
54. Expression trees и provider translation boundary.
55. Server/client evaluation, projection и dynamic query composition.
56. Attributes, `AttributeUsage` и consumers metadata.
57. Reflection: `Type`, `MemberInfo`, construction и invocation.
58. Reflection performance, delegates, trimming, Native AOT, source generators.
59. `dynamic`: runtime binding, допустимые границы и риски.

## Блок 5. Async, concurrency и I/O

60. `Task`/`Task<T>`, async state machine и смысл `await`.
61. Async I/O против parallel CPU work.
62. Blocking over async, ThreadPool starvation и `ConfigureAwait`.
63. `Task.WhenAll`, exceptions и ограничение concurrency.
64. Cooperative cancellation, linked tokens, timeout и side effects.
65. `ValueTask<T>` и `IAsyncEnumerable<T>`.
66. Race condition, atomicity и `Interlocked`.
67. `lock`, `Monitor`, `SemaphoreSlim`, deadlock.
68. Concurrent collections, Channels и backpressure.
69. Parallel APIs и PLINQ. Degree of parallelism.
70. Streams, capabilities, encoding и asynchronous disposal.
71. Buffers, `ArrayPool<T>`, `Span<T>`/`Memory<T>` и Pipelines.
72. `System.Text.Json`, DTO contracts и source generation.

## Блок 6. Tooling и CLR

73. `.csproj`, solution, SDK-style defaults и MSBuild evaluation.
74. ProjectReference, PackageReference, transitive dependencies и restore.
75. Central package management, lock files и reproducible builds.
76. Build, test, pack, publish и deployment modes.
77. Assembly structure, metadata и AssemblyLoadContext.
78. Runtime memory areas и object reachability.
79. Generational GC, LOH, promotion, pinning и finalization.
80. Managed memory leaks и инструменты диагностики.
81. Tiered compilation, dynamic PGO, inlining и devirtualization.
82. Bounds-check elimination, allocation trade-offs и structs.
83. ReadyToRun и Native AOT.
84. Корректный benchmark и отличие от production load test.

