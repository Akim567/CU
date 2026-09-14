# Курс C#: от первого выражения до CLR

Этот каталог — самостоятельный курс по C# 14 и .NET 10 LTS. Он повторяет глубину и последовательность материалов из `Java/`, но строится вокруг собственной модели C#: value types, properties, delegates, LINQ, `async`/`await`, nullable reference types и CLR.

> Курс рассчитан и на первое знакомство с C#, и на переход с Java. Блоки **«Сравнение с Java»** объясняют не только различия синтаксиса, но и отличия CLR, BCL и типичных инженерных практик.

## Как проходить курс

1. Материалы 01–06 дают синтаксис и корректную модель типов и памяти.
2. Материалы 07–14 посвящены объектной модели, generics и коллекциям.
3. Материалы 15–18 разбирают функциональные возможности, LINQ и метаданные.
4. Материалы 19–21 посвящены async, concurrency и I/O.
5. Материалы 22–24 объясняют сборку и устройство CLR.

Не перепрыгивайте сразу к ASP.NET Core. Особенно важны темы `IDisposable`, generics, LINQ, `IQueryable<T>`, async и cancellation: фреймворк использует их повсеместно.

## Содержание

| № | Материал | Результат |
|---:|---|---|
| 01 | [C# и платформа .NET](longread_01_csharp_and_dotnet.md) | понять SDK, runtime, компиляцию и запуск |
| 02 | [Система типов и переменные](longread_02_type_system.md) | освоить built-in types, conversions, `var`, `dynamic` |
| 03 | [Операторы и управление потоком](longread_03_control_flow.md) | писать условия, циклы и pattern matching |
| 04 | [Методы и параметры](longread_04_methods.md) | применять overload, `ref`/`out`/`in`, tuples и local functions |
| 05 | [Массивы, строки и текст](longread_05_arrays_and_strings.md) | корректно работать с коллекциями фиксированного размера и Unicode |
| 06 | [Value и reference semantics](longread_06_value_reference_nullability.md) | понимать копирование, boxing и nullable types |
| 07 | [Классы и объекты](longread_07_classes.md) | проектировать состояние, конструкторы и доступ |
| 08 | [Свойства и инкапсуляция](longread_08_properties.md) | применять properties, indexers, `init`, `required` |
| 09 | [Records, structs, tuples и enums](longread_09_data_types.md) | выбирать форму модели данных |
| 10 | [Наследование и полиморфизм](longread_10_inheritance.md) | понимать virtual dispatch, abstract и sealed types |
| 11 | [Интерфейсы и композиция](longread_11_interfaces.md) | строить контракты и композицию |
| 12 | [Исключения и ресурсы](longread_12_exceptions_and_resources.md) | обрабатывать ошибки и управлять временем жизни ресурсов |
| 13 | [Generics](longread_13_generics.md) | использовать constraints, variance и generic math |
| 14 | [Коллекции и равенство](longread_14_collections.md) | выбирать коллекцию и соблюдать equality contracts |
| 15 | [Delegates, lambdas и events](longread_15_delegates_lambdas_events.md) | передавать поведение и проектировать события |
| 16 | [LINQ и IEnumerable](longread_16_linq.md) | строить ленивые конвейеры обработки данных |
| 17 | [IQueryable и expression trees](longread_17_iqueryable_expressions.md) | понимать перевод запросов провайдерами |
| 18 | [Attributes, reflection и dynamic](longread_18_attributes_reflection.md) | читать метаданные и выполнять динамический код осознанно |
| 19 | [Асинхронность](longread_19_async.md) | применять `Task`, cancellation и async streams |
| 20 | [Многопоточность](longread_20_concurrency.md) | различать concurrency и parallelism, синхронизировать доступ |
| 21 | [I/O и сериализация](longread_21_io_serialization.md) | работать с потоками, файлами и JSON |
| 22 | [Проекты и зависимости](longread_22_projects_msbuild_nuget.md) | владеть `dotnet`, MSBuild и NuGet |
| 23 | [CLR: загрузка, память и GC](longread_23_clr_memory_gc.md) | понимать assemblies, type loading и сборку мусора |
| 24 | [CLR: JIT и производительность](longread_24_clr_jit_performance.md) | понимать tiered JIT, оптимизации и измерения |

После курса проверьте себя по [экзаменационным вопросам](exam_questions.md), затем переходите к [`ASP.NET/`](../ASP.NET/README.md).

## Версионность

Основной целевой фреймворк во всех новых проектах:

```xml
<TargetFramework>net10.0</TargetFramework>
<Nullable>enable</Nullable>
<ImplicitUsings>enable</ImplicitUsings>
<LangVersion>14.0</LangVersion>
```

Если пример использует возможность, появившуюся именно в C# 14, это отмечено явно. Большая часть фундаментального синтаксиса применима и к .NET 8/9, но поведение библиотек и доступный API всегда следует сверять с целевой версией.

## Обозначения

- ✅ рекомендуемый подход;
- ⚠️ важное ограничение или компромисс;
- ❌ антипример;
- **Сравнение с Java** — перенос знаний между экосистемами;
- **Самопроверка** — вопросы, на которые стоит ответить без подсказок.

