# Конспект лекции 1 — C# и платформа .NET

> Материал рассчитан на первое знакомство с C# и служит картой платформы для разработчиков, переходящих с Java.

## Из этого лонгрида вы узнаете

- чем язык C# отличается от платформы .NET;
- что делают SDK, runtime, CLR и Base Class Library;
- как исходный код превращается в выполняющуюся программу;
- как создать, собрать и запустить проект через `dotnet` CLI;
- зачем нужен файл проекта `.csproj`;
- почему для курса выбран .NET 10 LTS и C# 14.

## 1. Язык, платформа и реализация

**C#** — статически типизированный язык программирования. Его синтаксис и семантика описывают классы, методы, выражения, generics, pattern matching, async и другие конструкции.

**.NET** — платформа, включающая:

- runtime для выполнения управляемого кода;
- стандартные библиотеки;
- SDK и инструменты сборки;
- компиляторы языков;
- семейство прикладных фреймворков: ASP.NET Core, EF Core, .NET MAUI и другие.

Язык и платформа связаны, но не тождественны. C# можно компилировать для разных реализаций .NET, а для .NET существуют и другие языки, например F# и Visual Basic.

### Термины

| Термин | Роль |
|---|---|
| **SDK** | компилятор, CLI, MSBuild, шаблоны и инструменты разработки |
| **Runtime** | компоненты, необходимые для запуска приложения |
| **CLR/CoreCLR** | виртуальная машина: загрузка типов, JIT, GC, исключения, потоки |
| **BCL** | базовая библиотека: строки, коллекции, I/O, сеть, threading |
| **Roslyn** | компилятор и API анализа C# и Visual Basic |
| **IL/CIL** | промежуточные инструкции в сборке .NET |
| **Assembly** | единица развертывания и версионирования: обычно `.dll` или `.exe` |

## 2. Первая программа

Создадим каталог и консольный проект:

```bash
dotnet new console -n HelloCSharp
cd HelloCSharp
dotnet run
```

Современный шаблон использует **top-level statements**:

```csharp
Console.WriteLine("Hello, C#!");
```

Компилятор всё равно формирует тип и точку входа. Top-level syntax убирает церемониальный код, но не отменяет модель методов и классов.

Эквивалентная явная форма:

```csharp
namespace HelloCSharp;

internal static class Program
{
    public static void Main(string[] args)
    {
        Console.WriteLine("Hello, C#!");
    }
}
```

Разберём её:

1. `namespace HelloCSharp;` задаёт пространство имён в file-scoped форме.
2. `internal` делает тип доступным внутри текущей сборки.
3. `static class` запрещает создание экземпляров и требует статических членов.
4. `Main` — точка входа. Допустимы несколько поддерживаемых сигнатур, в том числе возвращающие `int` или `Task`.
5. `string[] args` содержит аргументы командной строки.
6. `Console.WriteLine` вызывает статический метод BCL.

## 3. Пример с вводом, проверкой и ветвлением

```csharp
Console.Write("Введите целое число: ");
string? input = Console.ReadLine();

if (!int.TryParse(input, out int number))
{
    Console.Error.WriteLine("Ошибка: введено не целое число.");
    return;
}

string parity = number % 2 == 0 ? "чётное" : "нечётное";
Console.WriteLine($"Число {number} — {parity}.");
```

### Пошаговый разбор

- `Console.ReadLine()` возвращает `string?`: поток может завершиться и вернуть `null`.
- `int.TryParse` не бросает исключение на обычной ошибке ввода. Он возвращает `bool`.
- `out int number` одновременно объявляет переменную и позволяет методу записать результат.
- `Console.Error` — стандартный поток ошибок, отличный от стандартного вывода.
- `return` в top-level программе завершает её.
- `?:` — условный оператор, вычисляющий значение.
- `$"..."` — интерполированная строка.

Этот пример показывает важную черту C#: API часто предлагает пару `Parse`/`TryParse`. Если невалидный ввод ожидаем, обычно выбирают `TryParse`; исключение не должно заменять обычное ветвление.

## 4. Что происходит при сборке

Команда:

```bash
dotnet build
```

упрощённо запускает такую цепочку:

```text
.cs + .csproj
      │
      ▼
  MSBuild оценивает проект
      │
      ▼
 Roslyn проверяет и компилирует C#
      │
      ▼
 assembly: IL + metadata + resources
      │
      ▼
 CLR загружает сборку и JIT-компилирует методы
      │
      ▼
 машинный код выполняется процессором
```

Компилятор обычно не создаёт окончательный машинный код для каждой поддерживаемой платформы. Он формирует IL и метаданные. Во время исполнения JIT преобразует используемые методы в инструкции текущей архитектуры. Для отдельных сценариев доступна Ahead-of-Time compilation, но это отдельная модель развертывания с ограничениями.

## 5. Структура минимального проекта

```text
HelloCSharp/
├── HelloCSharp.csproj
├── Program.cs
├── obj/
└── bin/
```

- `Program.cs` содержит исходный код.
- `.csproj` — декларативное описание проекта для MSBuild.
- `obj/` содержит промежуточные результаты, восстановленные зависимости и generated files.
- `bin/` содержит итоговые артефакты по configuration и target framework.

Минимальный проект:

```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net10.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
  </PropertyGroup>
</Project>
```

### Ключевые свойства

- `Project Sdk` подключает стандартные цели сборки .NET.
- `OutputType=Exe` означает запускаемое приложение.
- `TargetFramework=net10.0` определяет доступный API и runtime contract.
- `ImplicitUsings` добавляет набор обычных `using` автоматически.
- `Nullable` включает статический анализ nullable reference types.

## 6. Основные команды CLI

| Команда | Назначение |
|---|---|
| `dotnet --info` | показать SDK, runtime, ОС и архитектуру |
| `dotnet new list` | показать шаблоны |
| `dotnet new console` | создать консольный проект |
| `dotnet restore` | разрешить и загрузить NuGet-зависимости |
| `dotnet build` | восстановить зависимости и собрать |
| `dotnet run` | собрать при необходимости и запустить |
| `dotnet test` | собрать и запустить тесты |
| `dotnet publish` | подготовить артефакты для развертывания |
| `dotnet clean` | удалить результаты предыдущей сборки |

`dotnet run` удобен при разработке. Для production сначала выполняют `publish`, а затем запускают опубликованный артефакт в контролируемой среде.

## 7. SDK и runtime — не одно и то же

На машине разработчика нужен **SDK**. На сервере возможны варианты:

- **framework-dependent deployment** — приложение использует установленный runtime;
- **self-contained deployment** — runtime публикуется вместе с приложением;
- **single-file** — файлы упаковываются для удобства доставки;
- **Native AOT** — приложение заранее компилируется в native binary, но reflection и dynamic-loading сценарии требуют особого внимания.

Наличие runtime не означает, что можно собирать код. Наличие более нового SDK обычно позволяет таргетировать поддерживаемые framework versions, если установлены необходимые reference packs.

## 8. Версии языка и платформы

Версия C# и версия .NET связаны шаблонами и SDK, но являются разными понятиями:

```xml
<TargetFramework>net10.0</TargetFramework>
<LangVersion>14.0</LangVersion>
```

Не следует без необходимости ставить `LangVersion=preview`: preview-синтаксис может измениться. Для учебного курса используется финальный C# 14.

## 9. Где применяется C#

- серверные приложения и микросервисы на ASP.NET Core;
- облачные сервисы и фоновые workers;
- desktop через WPF, WinForms и кроссплатформенные решения;
- игры в Unity;
- мобильные и desktop-приложения на .NET MAUI;
- инструменты, компиляторы, тестовые и data-processing приложения.

Язык не привязан только к Windows. Современный .NET работает на Linux, macOS и Windows, а контейнеризация ASP.NET Core является обычным production-сценарием.

## 10. Сравнение с Java

| Java | C#/.NET |
|---|---|
| JDK | .NET SDK |
| JRE/JVM | .NET Runtime/CLR |
| `javac` | Roslyn compiler (`csc`, вызываемый SDK) |
| bytecode в `.class` | IL и metadata в assembly |
| JAR | assembly и набор publish-артефактов |
| Maven/Gradle | MSBuild + NuGet + `dotnet` CLI |
| `public static void main` | `Main` или top-level statements |
| package | namespace; он не обязан совпадать с каталогом |

Главная аналогия — управляемый runtime, промежуточное представление, JIT и garbage collection. Но JVM и CLR имеют разные type systems, metadata model, generics и runtime APIs. Нельзя автоматически переносить внутренние правила JVM на CLR.

## 11. Частые ошибки новичков

### ❌ Путать namespace и assembly

Namespace организует имена в исходном коде. Assembly — физическая единица компиляции и загрузки. Один namespace может быть распределён по нескольким assemblies, а одна assembly может содержать много namespaces.

### ❌ Править `bin/` и `obj/`

Это генерируемые каталоги. Изменения исчезнут при следующей сборке.

### ❌ Считать `var` динамической типизацией

```csharp
var count = 10; // статический тип — int
// count = "ten"; // ошибка компиляции
```

### ❌ Учить только синтаксис

Для надёжного C# необходимы модель типов, nullable analysis, управление ресурсами и async. Они влияют на любой ASP.NET Core-код.

## Выводы

- C# — язык, .NET — платформа.
- SDK нужен для разработки, runtime — для выполнения.
- C# компилируется в IL и metadata внутри assembly; CLR выполняет код и управляет памятью.
- `.csproj` определяет target framework, настройки компиляции и зависимости.
- `dotnet` CLI даёт воспроизводимый способ создать, собрать, проверить и опубликовать приложение.

## Самопроверка

1. Чем SDK отличается от runtime?
2. Что находится внутри assembly?
3. Почему namespace нельзя считать аналогом отдельного JAR?
4. Что делает `dotnet build`, а что — `dotnet publish`?
5. Почему top-level statements не отменяют точку входа?

## Полезные материалы

- [Документация C#](https://learn.microsoft.com/dotnet/csharp/)
- [.NET CLI overview](https://learn.microsoft.com/dotnet/core/tools/)
- [.NET application publishing](https://learn.microsoft.com/dotnet/core/deploying/)
- [.NET support policy](https://dotnet.microsoft.com/platform/support/policy)
