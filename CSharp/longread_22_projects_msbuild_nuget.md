# Конспект лекции 22 — Проекты, MSBuild, NuGet и `dotnet` CLI

## 1. Уровни организации

```text
solution (.sln/.slnx)
├─ project A (.csproj)
├─ project B (.csproj)
└─ test project (.csproj)
```

Project компилируется в assembly. Solution группирует projects для разработки и orchestration, но не является runtime container.

## 2. SDK-style project

```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Npgsql" Version="10.0.0" />
  </ItemGroup>
</Project>
```

SDK добавляет defaults, targets и glob inclusion `.cs` files. MSBuild оценивает XML properties/items, затем выполняет target graph.

## 3. Restore, build, test, publish

```bash
dotnet restore
dotnet build --configuration Release
dotnet test --configuration Release --no-build
dotnet publish src/App/App.csproj --configuration Release
```

- restore вычисляет dependency graph и assets;
- build компилирует;
- test запускает test projects через выбранную platform;
- publish собирает deployable layout.

Build автоматически restore-ит по умолчанию; в CI explicit restore + `--no-restore` делает стадии и cache понятнее.

## 4. Project references

```xml
<ItemGroup>
  <ProjectReference Include="../Domain/Domain.csproj" />
</ItemGroup>
```

Это compile-time dependency и часть build graph. Архитектурное правило направления dependencies важнее количества projects.

## 5. Package references и transitive dependencies

```bash
dotnet add package FluentValidation
dotnet list package --include-transitive
```

NuGet package может принести transitive packages. Проверяйте graph, licenses, vulnerabilities и compatibility. Не добавляйте package ради одного тривиального helper.

## 6. Central package management

`Directory.Packages.props`:

```xml
<Project>
  <PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
  </PropertyGroup>
  <ItemGroup>
    <PackageVersion Include="Npgsql" Version="10.0.0" />
  </ItemGroup>
</Project>
```

Project:

```xml
<PackageReference Include="Npgsql" />
```

Централизация предотвращает случайный version drift. Она не гарантирует, что upgrade совместим.

## 7. Общие настройки

`Directory.Build.props` применяется к projects ниже по directory tree:

```xml
<Project>
  <PropertyGroup>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <AnalysisLevel>latest-recommended</AnalysisLevel>
  </PropertyGroup>
</Project>
```

`Directory.Build.targets` используется для targets, обычно после imports. Не помещайте туда магию без документации: влияние распространяется широко.

## 8. Configuration и target framework

`Debug`/`Release` — наборы MSBuild properties, а не просто включён/выключен debugger.

Multi-targeting:

```xml
<TargetFrameworks>net8.0;net10.0</TargetFrameworks>
```

Оправдано главным образом для libraries с реальными consumers разных frameworks. Application обычно имеет один target.

## 9. Locking versions

NuGet разрешает version ranges и transitive graph. Для повторяемости CI можно использовать lock file:

```xml
<RestorePackagesWithLockFile>true</RestorePackagesWithLockFile>
```

Далее repository хранит `packages.lock.json`, а locked mode запрещает неожиданный graph drift.

## 10. Packaging library

```xml
<PropertyGroup>
  <PackageId>Company.Contracts</PackageId>
  <Version>1.2.0</Version>
  <Description>Shared API contracts</Description>
  <GeneratePackageOnBuild>true</GeneratePackageOnBuild>
</PropertyGroup>
```

```bash
dotnet pack --configuration Release
dotnet nuget push package.nupkg --source internal
```

Public API compatibility, semantic versioning и source/package symbols становятся частью engineering contract.

## 11. Build lifecycle

Упрощённо:

1. evaluation — properties/items/imports превращаются в project model;
2. restore — dependency assets;
3. target execution — compile/test/pack/publish graphs;
4. incremental checks — inputs/outputs позволяют пропускать работу.

MSBuild declarative, но custom tasks/targets могут добавить произвольное поведение.

## 12. Reproducibility

- фиксируйте SDK через `global.json`, если команда требует одинаковый toolchain;
- не храните secrets в project files/NuGet config repository;
- pin container base images и packages по принятой политике;
- включайте deterministic builds/source link для libraries;
- очищайте не «на всякий случай», а при диагностике cache/input issue.

## 13. Сравнение с Gradle/Maven

| Java | .NET |
|---|---|
| Gradle/Maven project | `.csproj` MSBuild project |
| Maven Central | NuGet feeds/nuget.org |
| dependency coordinates | PackageReference id + version |
| wrapper/toolchains | `global.json`, installed SDK/container image |
| multi-module build | solution + ProjectReference graph |
| BOM/version catalog | central package management |

NuGet — dependency/package manager, MSBuild — build engine, `dotnet` — unified CLI. Нельзя называть всё это просто «аналогом Gradle» без разделения ролей.

## Самопроверка

1. Компилируется ли solution в одну assembly?
2. Чем ProjectReference отличается от PackageReference?
3. Зачем `Directory.Build.props`?
4. Что фиксирует package lock file?
5. Какие стадии выполняет `publish` сверх обычного coding workflow?

