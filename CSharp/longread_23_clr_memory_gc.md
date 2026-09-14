# Конспект лекции 23 — CLR: assemblies, память и Garbage Collector

## 1. Путь выполнения

```text
C# source
  → Roslyn
  → PE file: CIL + metadata
  → AssemblyLoadContext loads assemblies
  → CLR verifies/resolves types
  → JIT compiles used methods
  → GC manages reachable objects
```

Assembly обычно содержит manifest, metadata tables, IL methods и resources. Type identity включает assembly и load context, а не только полное имя.

## 2. Metadata и managed execution

Metadata описывает:

- types и inheritance;
- fields, methods, signatures;
- generic parameters;
- references на assemblies;
- custom attributes.

CLR использует её для loading, verification, reflection, JIT и interop. Managed code не означает «никогда не работает с native memory», а означает выполнение под контрактами runtime.

## 3. Assembly loading

Default `AssemblyLoadContext` загружает обычные dependencies. Custom context нужен plugins/isolation/unloading.

```csharp
var context = new AssemblyLoadContext("plugin", isCollectible: true);
Assembly plugin = context.LoadFromAssemblyPath(fullPath);
```

Unload только инициируется: context освободится, когда не останется сильных ссылок на types, instances, delegates, threads и другие объекты из него.

## 4. Runtime memory areas

Упрощённо процесс содержит:

- managed heap;
- stacks потоков;
- JIT code heaps;
- loader/metadata structures;
- native allocations runtime и libraries;
- thread pool и OS resources.

Метрика managed heap не равна всей resident memory процесса.

## 5. Object layout

Managed object conceptually имеет runtime header, type information и fields с alignment. Точный layout зависит от runtime, architecture и attributes. Не закладывайтесь на «фиксированный размер заголовка» без измерения конкретной версии.

Reference field хранит ссылку; value-type field встраивается inline. Массив value types хранит элементы inline, массив references — ссылки на отдельные объекты.

## 6. Reachability

GC начинает с roots:

- stack/register references;
- static references;
- GC handles;
- runtime structures.

Объект жив, если достижим по цепочке references. Циклы сами по себе не мешают collection:

```text
root → A ↔ B   // живы

A ↔ B          // без пути от root: могут быть собраны
```

## 7. Generational GC

Managed heap логически делится на generations:

- Gen 0 — новые короткоживущие объекты;
- Gen 1 — промежуточный буфер;
- Gen 2 — долгоживущие объекты;
- Large Object Heap — крупные allocations;
- pinned-object mechanisms/heap — отдельные аспекты для pinned data.

Идея generational hypothesis: большинство новых объектов быстро умирает, поэтому young collection может быть частой и относительно дешёвой.

Survivor может быть promoted. Частые promotions и большие live sets делают GC дороже.

## 8. Allocation

Small-object allocation часто очень быстра благодаря thread-local allocation contexts: pointer bump, а не общий expensive allocator. Цена проявляется позже в GC work, promotion, cache pressure.

«Allocation быстрый» не означает «allocation бесплатный».

## 9. LOH

Large objects обычно попадают в Large Object Heap. Threshold является runtime detail, исторически около 85,000 bytes. LOH собирается с Gen 2 и может влиять на pauses/fragmentation.

Не дробите objects вслепую. Pooling больших buffers оправдан при повторении и контролируемом ownership.

## 10. Finalization

Finalizable object требует дополнительного lifecycle: после недостижимости finalizer должен быть запланирован, а память освобождается позже.

```csharp
public void Dispose()
{
    ReleaseUnmanaged();
    GC.SuppressFinalize(this);
}
```

Предпочитайте `SafeHandle` для native handles. Обычные managed classes finalizer не требуют.

## 11. Pinned memory

GC обычно перемещает objects при compaction. Pinned object нельзя перемещать, что может создавать fragmentation.

```csharp
fixed (byte* pointer = buffer)
{
    NativeCall(pointer, buffer.Length);
}
```

Pinning должен быть коротким и обоснованным interop. Современные APIs предлагают memory pinning abstractions, но правила lifetime остаются.

## 12. GC modes

Workstation/server GC и latency modes оптимизируют разные workloads. ASP.NET Core production обычно получает server-oriented runtime configuration в зависимости от deployment. Не переключайте режим только по названию; измеряйте throughput, pauses, heap size и CPU.

## 13. Weak references

`WeakReference<T>` не удерживает object живым. Он полезен для некоторых caches/metadata mappings, но состояние может исчезнуть между проверкой и использованием. Weak reference не заменяет обычную cache eviction policy.

## 14. Memory leaks в managed code

Managed leak — ненужный объект всё ещё достижим:

- static collection растёт без очистки;
- event publisher удерживает subscribers;
- cache не имеет bounds;
- background task удерживает closure;
- timer/handle не disposed;
- AssemblyLoadContext удерживается delegate/type reference.

GC не может угадать, что достижимый объект больше не нужен бизнесу.

## 15. Сравнение с JVM

Общие идеи: generational collection, roots, compacting collectors, JIT и managed heap. Отличаются object layout, collectors, loader identity, metadata, tuning flags и implementation evolution.

- JVM `ClassLoader` ближе по роли к loader contexts, но не идентичен `AssemblyLoadContext`.
- JVM metaspace и CLR loader/native structures нельзя сопоставлять один к одному.
- LOH — характерная часть .NET heap model.
- И в Java, и в .NET `finalize`/finalizers не заменяют deterministic resource management.

## 16. Диагностика

Измеряйте:

- allocation rate;
- heap size по generations;
- GC pause time и frequency;
- promoted bytes;
- LOH size;
- thread-pool и native memory отдельно.

Инструменты: `dotnet-counters`, `dotnet-trace`, `dotnet-gcdump`, `dotnet-dump`, профилировщики IDE/production APM.

## Самопроверка

1. Является ли размер managed heap всей памятью процесса?
2. Почему cycles могут быть собраны?
3. Что такое promotion?
4. Почему finalizer задерживает освобождение?
5. Как event создаёт managed memory leak?

