# Конспект лекции 24 — CLR: JIT, оптимизации и производительность

## 1. Зачем JIT

IL сохраняет переносимость и metadata-rich execution. JIT компилирует method для текущей архитектуры и runtime conditions. Цена compilation платится при исполнении, зато runtime может использовать сведения о CPU и фактическом процессе.

## 2. Tiered compilation

Упрощённо runtime может:

1. быстро скомпилировать method с небольшими затратами;
2. наблюдать hot methods;
3. перекомпилировать важный код с более дорогими optimizations;
4. использовать dynamic profile-guided optimization.

Точные tiers и thresholds — implementation details конкретного runtime.

## 3. Inlining

JIT может встроить тело небольшого callee в caller:

```csharp
static int Square(int value) => value * value;
int result = Square(input);
```

После inlining открываются constant propagation, bounds-check elimination и dead-code elimination. Слишком большой method, complex exception handling или virtual uncertainty могут помешать.

`MethodImplOptions.AggressiveInlining` — hint, не приказ. Без benchmarks он часто вреден размеру code cache.

## 4. Devirtualization

Если JIT доказывает конкретный target virtual/interface call, он может заменить indirect dispatch direct call и затем inline.

Помогают sealed types, exact type knowledge и runtime profiles. Не делайте всё sealed только ради speculative micro-optimization; design остаётся главным.

## 5. Bounds-check elimination

```csharp
for (int i = 0; i < array.Length; i++)
    sum += array[i];
```

Canonical loop позволяет JIT доказать безопасность и убрать повторные checks. «Хитрый» index math может помешать. Сначала пишите понятный цикл.

## 6. Escape и stack allocation

`stackalloc` явно выделяет contiguous memory с lifetime текущего frame:

```csharp
Span<byte> buffer = stackalloc byte[256];
```

Размер должен быть ограничен. User-controlled large stackalloc может вызвать stack overflow.

JIT может оптимизировать allocations, но не полагайтесь на идентичный escape analysis JVM. Проверяйте generated code/allocation measurements.

## 7. ReadyToRun и Native AOT

- ReadyToRun публикует предварительно скомпилированный код и может улучшать startup ценой размера/части peak optimization.
- Native AOT создаёт native executable с быстрым startup и меньшей runtime dependency, но ограничивает dynamic code/reflection patterns и требует trimming-friendly dependencies.

JIT deployment остаётся нормальным выбором для server throughput; AOT — не универсально «быстрее».

## 8. Allocation hot paths

Источники:

- closures и captured lambdas;
- boxing;
- intermediate strings;
- LINQ iterators/materialization;
- arrays from `params`;
- async state/task objects;
- serialization buffers.

Сначала найдите allocation profile. Затем применяйте `Span`, pooling, static lambdas, struct enumerators или caching там, где benefit измерим и ownership ясен.

## 9. Struct trade-offs

Struct уменьшает отдельные heap allocations, но:

- копируется по значению;
- большой struct увеличивает traffic;
- interface conversion может boxing;
- mutable struct создаёт semantic bugs;
- generic specialization увеличивает native code size.

`readonly` помогает, но не превращает любую большую модель в хороший value type.

## 10. Benchmarking

Microbenchmark должен учитывать warmup, tiering, dead-code elimination, input distribution и noise. Используйте BenchmarkDotNet вместо ручного `Stopwatch` для микросравнений.

```csharp
[MemoryDiagnoser]
public class LookupBenchmarks
{
    private readonly Dictionary<int, string> _values =
        Enumerable.Range(0, 1_000).ToDictionary(x => x, x => x.ToString());

    [Benchmark]
    public bool Lookup() => _values.ContainsKey(500);
}
```

End-to-end performance требует load test и production telemetry: microbenchmark не измеряет database/network contention.

## 11. Big-O и constants

Algorithmic complexity важнее micro-optimization при росте данных. Затем имеют значение allocations, locality, vectorization, branch prediction и synchronization.

Правильный порядок:

1. определить SLO и workload;
2. измерить;
3. найти bottleneck;
4. изменить одну гипотезу;
5. проверить correctness и regression;
6. повторить.

## 12. Diagnostics

```bash
dotnet-counters monitor --process-id <pid>
dotnet-trace collect --process-id <pid>
dotnet-gcdump collect --process-id <pid>
```

Смотрите CPU samples, allocation stacks, exceptions, GC, locks/contention, ThreadPool queue, request latency и downstream calls.

## 13. Hardware intrinsics и SIMD

`Vector<T>` и hardware intrinsics ускоряют data-parallel computation, но привязывают code к supported types/architectures и увеличивают сложность fallback. Начинайте с library algorithms; используйте intrinsics для доказанного hotspot.

## 14. Сравнение с JVM JIT

Общие приёмы: tiering, profiling, inlining, devirtualization, constant folding, dead-code и bounds-check elimination. Но HotSpot C1/C2/Graal и CoreCLR RyuJIT имеют разные pipelines, thresholds и diagnostics.

Нельзя переносить JVM flag, размер object header или вывод о конкретной optimization на CLR. Сопоставлять полезно идеи, а подтверждать — disassembly и профиль конкретной версии.

## 15. Антиоптимизации

- Object pooling для дешёвых short-lived objects увеличивает retention/complexity.
- `Task.Run` вокруг I/O тратит ThreadPool.
- Ручной unsafe code без benchmark повышает риск.
- `AggressiveInlining` повсюду раздувает native code.
- `Span<T>` протекает через весь domain API без необходимости.
- Benchmark в Debug configuration даёт нерелевантный вывод.

## Выводы

Runtime оптимизирует понятный idiomatic code удивительно хорошо. Производительность — свойство workload и системы, а не количество «быстрых» keywords. Измерение важнее предположения.

## Самопроверка

1. Зачем нужны tiers JIT compilation?
2. Какие оптимизации открывает inlining?
3. Когда Native AOT имеет ограничения?
4. Почему маленький struct не всегда быстрее class?
5. Чем microbenchmark отличается от load test?

