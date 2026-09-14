# C# и ASP.NET Core — спецификация учебного курса

**Статус:** implemented  
**Область:** образовательные материалы  
**Последнее обновление:** 2026-09-06  
**Связанные материалы:** `Java/`, `Spring/`  
**Источник истины:** `CSharp/README.md`, `ASP.NET/README.md`

**Реализовано в:** `CSharp/`, `ASP.NET/`  
**Известные пробелы:** отсутствуют в согласованном объёме; отдельные проекты семинаров намеренно исключены  
**Проверка:** 24 языковых лонгрида, 15 пар «лонгрид + семинар», два набора экзаменационных вопросов; проверены Markdown fences, локальные ссылки и структура всех недель

## 1. Контекст

Репозиторий содержит подробный вводный курс Java и пятнадцатинедельный курс Spring. Требуется создать самостоятельную траекторию изучения C# и серверной разработки на ASP.NET Core и EF Core. Новый курс сохраняет педагогическую манеру исходных материалов, но не является механическим переводом Java API в C#.

## 2. Зафиксированные решения

- Основной стек: .NET 10 LTS, C# 14, ASP.NET Core 10, EF Core 10.
- Языковой курс разбивается на более мелкие темы, чем исходные Java-лонгриды.
- Фреймворк-курс покрывает весь объём `Spring/`, а не только `Spring/Week_2/`.
- Семинары поставляются как Markdown с заданиями, критериями и решениями; отдельный проект для каждого семинара не создаётся.
- Основной язык текста — русский; устоявшиеся термины приводятся также на английском.
- В тематически значимых местах добавляются блоки «Сравнение с Java».
- Основной вариант серверного API — controller-based Web API. Minimal APIs рассматриваются и сравниваются отдельно.
- Основная реляционная СУБД — PostgreSQL; провайдер EF Core — Npgsql.

## 3. Формат одного языкового лонгрида

Каждый материал, где это применимо, содержит:

1. Назначение темы и ожидаемые результаты.
2. Проблему, которую решает механизм.
3. Термины и модель выполнения.
4. Синтаксис от минимального примера к инженерному.
5. Построчные либо поэтапные объяснения нетривиального кода.
6. Ограничения, частые ошибки и антипримеры.
7. Практические рекомендации.
8. Блок сравнения с Java.
9. Краткие выводы и вопросы для самопроверки.
10. Ссылки преимущественно на Microsoft Learn и спецификацию C#.

Код должен быть самодостаточным в границах примера. Если фрагмент намеренно неполный или демонстрирует ошибку компиляции, это явно подписывается.

## 4. Программа C#

| № | Тема | Ключевое содержание |
|---:|---|---|
| 01 | C# и платформа .NET | SDK/runtime, компиляция, CLI, первая программа, LTS |
| 02 | Система типов и переменные | built-in types, литералы, преобразования, `var`, `dynamic`, overflow |
| 03 | Операторы и управление потоком | условия, циклы, pattern matching, expressions |
| 04 | Методы и параметры | overload, named/optional, `ref`/`out`/`in`, `params`, tuples, local functions |
| 05 | Массивы, строки и текст | arrays, ranges, `string`, Unicode, interpolation, builders |
| 06 | Value и reference semantics | stack/heap без мифов, copying, boxing, nullable types |
| 07 | Классы и объекты | fields, constructors, access, partial types, nested types |
| 08 | Свойства и инкапсуляция | properties, indexers, `init`, `required`, immutability |
| 09 | Records, structs, tuples, enums | модели данных и семантика равенства |
| 10 | Наследование и полиморфизм | virtual dispatch, abstract/sealed, casting, `object` |
| 11 | Интерфейсы и композиция | default members, explicit implementation, extension members, operators |
| 12 | Исключения и управление ресурсами | hierarchy, filters, `finally`, `IDisposable`, `using` |
| 13 | Generics | constraints, variance, reification, generic math |
| 14 | Коллекции и равенство | collection families, hashing, comparers, immutable/frozen collections |
| 15 | Delegates, lambdas и events | closures, method groups, multicast delegates, event pattern |
| 16 | LINQ и `IEnumerable<T>` | deferred execution, operators, materialization, custom iterators |
| 17 | `IQueryable<T>` и expression trees | provider translation, client/server boundary, dynamic queries |
| 18 | Attributes, reflection и `dynamic` | metadata, invocation, attributes, source-generation boundary |
| 19 | Асинхронность | `Task`, `ValueTask`, `async`/`await`, cancellation, async streams |
| 20 | Многопоточность | memory model, synchronization, concurrent collections, channels, parallelism |
| 21 | I/O и сериализация | streams, files, pipelines overview, JSON, disposal, cancellation |
| 22 | Проекты и зависимости | solutions, project files, CLI, MSBuild, NuGet, central package management |
| 23 | CLR: загрузка, память и GC | assemblies, type loading, runtime areas, GC, finalization |
| 24 | CLR: JIT и производительность | tiered compilation, devirtualization, inlining, allocations, diagnostics |

Дополнительно создаётся `CSharp/exam_questions.md`.

## 5. Программа ASP.NET Core и EF Core

| Неделя | Лонгрид | Практика |
|---:|---|---|
| 01 | Framework, Generic Host, DI и первое приложение | создание проекта вручную, регистрации и варианты DI |
| 02 | DI lifetimes, lifecycle, disposal, decorators/interception | наблюдение жизненного цикла, captive dependency, декораторы |
| 03 | Архитектура, конфигурация, Options, environments | слои/модули, Options validation, providers, hosted services |
| 04 | HTTP pipeline, middleware, routing, controllers, binding, validation | Web API, DTO, файлы, cookies/session, собственный middleware |
| 05 | Ответы, Problem Details, exception handling, CORS, OpenAPI | единый error contract, headers, generated OpenAPI |
| 06 | ADO.NET, Npgsql, pooling, migrations и транзакции | raw SQL, parameters, mapping, transaction scenarios |
| 07 | EF Core: model, tracking, lifecycle, repositories, relations | CRUD, states, relations, cascade behavior |
| 08 | EF Core: LINQ translation, N+1, loading, paging, locks, performance | измерение запросов и исправление проблем |
| 09 | Security, Identity, password hashing, JWT и logging | login, bearer auth, roles/policies, structured logging |
| 10 | `HttpClientFactory`, resilience, rate limiting, circuit breaker | устойчивый HTTP-клиент и отказные сценарии |
| 11 | Unit, integration и E2E testing | xUnit, mocks, `WebApplicationFactory`, Testcontainers |
| 12 | Async request processing, streaming, Kestrel и cancellation | async endpoints, streaming, backpressure boundaries |
| 13 | Асинхронная БД, фоновые задачи и согласованность | EF async, cancellation, queues, outbox introduction |
| 14 | Event-driven architecture и Kafka | producer, consumer groups, offsets, retries, DLQ, idempotency |
| 15 | Observability | logs, metrics, traces, health checks, OpenTelemetry, Prometheus/Grafana |

Каждая неделя содержит `Longread.md` и `Seminar.md`. При необходимости сложная практика может быть разделена на несколько файлов, как в исходной второй неделе Spring.

## 6. Не входит в объём

- Готовый отдельный проект для каждого семинара.
- Desktop UI, MAUI, Unity и legacy ASP.NET Framework.
- Полноценный курс SQL, Kafka administration, Kubernetes или эксплуатации Grafana.
- Дословное повторение фактических ошибок и несогласованной нумерации исходных материалов.

## 7. Критерии качества

- Нет пустых заглушек и разделов, состоящих только из перечня терминов.
- Пример не выдаётся за полностью запускаемый, если ему не хватает контекста.
- Версионно-зависимые сведения соответствуют .NET 10/C# 14.
- В сравнении с Java различаются язык, BCL, CLR и фреймворк — эти уровни не смешиваются.
- В семинарах есть цель, предпосылки, последовательные действия, ожидаемый результат и решение либо проверяемая подсказка.
- Для БД, безопасности, транзакций и async отдельно разобраны ошибки и production-ограничения.

## 8. Проверка

- [x] Проверено дерево: 26 Markdown-файлов в `CSharp/` и 32 в `ASP.NET/`.
- [x] Проверены Markdown-заголовки и парность всех fenced code blocks.
- [x] Проверены относительные ссылки из обоих README; битых ссылок нет.
- [x] Все 15 недель содержат `Longread.md` и `Seminar.md`.
- [x] Каждый языковой лонгрид имеет блок сравнения и вопросы для самопроверки.
- [x] Каждый framework-лонгрид имеет сравнение со Spring/JVM-экосистемой, каждый семинар — проверяемый чек-лист.
- [x] Репрезентативный набор C# 12/.NET 8-совместимых примеров собран локальным SDK 8.0.401 без предупреждений и ошибок.
- [x] Конструкции C# 14 и .NET 10 сверены с актуальной официальной документацией Microsoft.
- [x] Выполнена содержательная сверка с `Java/` и всеми блоками `Spring/Questions (1).md`.

## 9. Известные ограничения среды

На момент старта локально установлен .NET SDK 8.0.401. Поэтому общая компиляционная проверка базового синтаксиса возможна, но конструкции C# 14 и ASP.NET Core 10 требуют проверки по официальной документации либо запуска в среде с .NET 10 SDK.

## 10. Журнал решений

- 2026-09-06: выбран актуальный LTS-стек .NET 10 вместо уходящего из поддержки .NET 8.
- 2026-09-06: языковой курс расширен до 24 более узких тем.
- 2026-09-06: ASP.NET Core оформляется как полный пятнадцатинедельный курс.
- 2026-09-06: отдельные проекты семинаров исключены по запросу пользователя.

## 11. Открытые вопросы

Нет блокирующих вопросов.
