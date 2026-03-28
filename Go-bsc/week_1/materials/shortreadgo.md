# Введение в Go — Краткий справочник

## Установка

### macOS
```bash
brew install go
```

### Windows
- Скачать с https://go.dev/dl/ и запустить установщик
- Или через Chocolatey: `choco install golang`

### Linux
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install golang-go

# CentOS/RHEL
sudo yum install golang  # или dnf для новых версий
```

### Проверка установки
```bash
go version
```

---

## Первая программа и пакеты

### Hello World
```go
package main

import "fmt"

func main() {
    fmt.Print("Hello, world!")
}
```

### Структура программы:
1. **Пакет** — единица деления кода, область видимости приватных переменных
2. **package main** — обязательно для входной точки приложения  
3. **import** — импорт других пакетов (обязательно использовать все импорты!)
4. **func main()** — входная точка приложения

### Множественный импорт:
```go
import (
    "fmt"
    "os"
)
```

---

## Сборка и кроссплатформенность

### Основные команды сборки:
```bash
go build                   # собрать исполняемый файл
go build -o myapp          # указать имя файла
```

### Кроссплатформенная сборка:
```bash
go tool dist list          # посмотреть доступные платформы

# Сборка для Windows на других ОС:
GOOS=windows GOARCH=amd64 go build

# Примеры платформ:
# darwin/amd64, linux/amd64, windows/amd64
# android/arm64, ios/arm64, и многие другие
```

**Особенность Go**: статическая линковка — все зависимости упакованы в один исполняемый файл.

---

## Переменные и типы данных

### Основные типы:

**Строковый тип:**
- `string` — набор байт в кодировке UTF-8 (занимает 16 байт: указатель + длина)

**Целые числа:**
- `int8, int16, int32, int64` — знаковые целые
- `uint8, uint16, uint32, uint64` — беззнаковые целые  
- `int, uint` — зависят от архитектуры (32/64 бита)
- `byte` — алиас для `uint8`
- `rune` — алиас для `int32` (символ Unicode)

**Вещественные числа:**
- `float32, float64` — числа с плавающей запятой
- `complex64, complex128` — комплексные числа

**Логический тип:**
- `bool` — только `true` или `false`

### Объявление переменных:

```go
// Полное объявление
var i int = 1
var w float64 = 12.5

// С нулевыми значениями
var i int          // 0
var isPresent bool // false
var name string    // ""

// Выведение типа
var i, j, k = 1, 2, 3 // int

// Короткое объявление (только в функциях!)
name := "John Doe"
age := 34

// Множественное объявление
var i, j, k int = 1, 2, 3
var (
    name string = "John Doe"
    occupation string = "gardener"
)
```

### Константы и iota:
```go
const Pi = 3.14159

const (
    StatusOK = 200
    StatusNotFound = 404
)

// Enum с iota
type Direction int
const (
    North Direction = iota // 0
    East                   // 1
    South                  // 2
    West                   // 3
)
```

### Области видимости:
- **Пакет** — переменные уровня пакета (глобальные)
  - Верхний регистр = видны из других пакетов
  - Нижний регистр = только внутри пакета
- **Функция** — ограничена фигурными скобками

---

## Операции

### Арифметические:
```go
x := 10
y := 3
fmt.Println(x + y)  // 13 (сложение)
fmt.Println(x - y)  // 7  (вычитание)
fmt.Println(x * y)  // 30 (умножение)
fmt.Println(x / y)  // 3  (деление)
fmt.Println(x % y)  // 1  (остаток)
```

### Побитовые:
```go
x & y   // побитовое И
x | y   // побитовое ИЛИ  
x ^ y   // побитовое исключающее ИЛИ
x << 2  // сдвиг влево
x >> 2  // сдвиг вправо
```

### Строки:
```go
greeting := "Hello, " + "world!"  // конкатенация
len("Hello")  // 5 (длина в БАЙТАХ, не символах!)
```

### Сравнения:
```go
==, !=        // равенство/неравенство
>, <, >=, <=  // больше/меньше
```

### Логические (только для bool):
```go
a && b  // логическое И
a || b  // логическое ИЛИ
!a      // логическое НЕ
```

### Короткое присваивание:
```go
v += 1  // эквивалентно v = v + 1
v -= 1  // эквивалентно v = v - 1
// аналогично для *=, /=, %=
```

---

## Условные операторы

### if-else:
```go
if x > 5 {
    fmt.Println("x больше 5")
}

// С инициализацией в заголовке
if x := getValue(); x > 0 {
    fmt.Println("Положительное значение:", x)
}
// x здесь уже недоступна

// Полная форма
if x > 0 {
    fmt.Println("Положительное")
} else if x < 0 {
    fmt.Println("Отрицательное")  
} else {
    fmt.Println("Ноль")
}
```

**Важно:** В Go НЕТ тернарного оператора!

### switch:
```go
// Обычный switch
switch time.Now().Weekday() {
case time.Saturday:
    fmt.Println("Today is Saturday")
case time.Sunday:
    fmt.Println("Today is Sunday")
default:
    fmt.Println("Today is a weekday")
}

// С инициализацией
switch tnow := time.Now().Weekday(); tnow {
case time.Saturday, time.Sunday:
    fmt.Println("Today is", tnow)
}

// Без выражения (эквивалентно switch true)
hour := time.Now().Hour()
switch {
case hour < 12:
    fmt.Println("Good morning!")
case hour < 17:
    fmt.Println("Good afternoon!")
default:
    fmt.Println("Good evening!")
}
```

**Управление switch:**
- `fallthrough` — продолжить выполнение следующего case
- `break` — принудительный выход

**Зачем switch?** Более компактный синтаксис + оптимизация компилятора (таблица переходов).

---

## Циклы

В Go есть ТОЛЬКО цикл `for` (нет while, do-while и т.п.)

### Классический for:
```go
sum := 0
for i := 1; i < 5; i++ {
    sum += i
}
// 1. Инициализация: i := 1
// 2. Условие: i < 5  
// 3. Тело цикла
// 4. Финализация: i++
```

### Аналог while:
```go
n := 1
for n < 5 {
    n *= 2
}
```

### Бесконечный цикл:
```go
sum := 0
for {
    sum++
    if sum >= 100 {
        break  // выход из цикла
    }
}
```

### Управление циклами:
```go
// break с метками для вложенных циклов
loop001:
for {
    for j := 0; j < 10; j++ {
        if условие {
            break loop001  // выход из внешнего цикла
        }
    }
}

// continue — переход к следующей итерации
for i := 1; i < 5; i++ {
    if i%2 != 0 {
        continue // пропустить нечетные
    }
    sum += i
}
```

### Цикл for-range (по коллекциям):
```go
strings := []string{"hello", "world"}

// С индексом и значением
for i, s := range strings {
    fmt.Println(i, s)
}

// Только индекс
for i := range strings {
    // работаем только с индексом
}

// Только значение  
for _, s := range strings {
    // работаем только со значением
}

// По строке (по символам Unicode!)
for i, ch := range "日本語" {
    fmt.Printf("%#U starts at byte position %d\n", ch, i)
}
```

---

## Функции

**Функции в Go** — функции первого класса (можно присваивать переменным, передавать в параметрах, возвращать из функций).

### Базовый синтаксис:
```go
func имя(параметр тип) возвращаемыйТип {
    return значение
}
```

### Примеры функций:
```go
// Простая функция
func add(a, b int) int {  // сокращённая запись параметров одного типа
    return a + b
}

// Без параметров и возврата
func sayHello() {
    fmt.Println("Hello!")
}

// Множественный возврат
func divmod(a, b int) (int, int) {
    return a / b, a % b
}

// Именованные возвращаемые значения
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return // автоматически вернёт x и y
}
```

### Анонимные функции:
```go
// Присваивание в переменную
mul := func(x, y int) int {
    return x * y
}
println(mul(2, 2))

// Немедленное выполнение
sum := func(a, b, c int) int {
    return a + b + c
}(3, 5, 7)
```

### Variadic функции (переменное число параметров):
```go
func sum(begin int, nums ...int) int {
    res := begin
    for _, n := range nums {  // nums — это slice
        res += n
    }
    return res
}

println(sum(1))        // 1
println(sum(1, 1, 1))  // 3
```

### Типы функций:
```go
type Output func(string) string

// Функция высшего порядка (принимает функции как параметры)
func apply(x, y int, add func(int, int) int, sub func(int, int) int) (int, int) {
    r1 := add(x, y)
    r2 := sub(x, y)
    return r1, r2
}
```

### Рекурсия:
```go
func fact(n int) int {
    if n == 0 || n == 1 {
        return 1
    }
    return n * fact(n-1)
}
```

### Важно о параметрах:
```go
// ВСЕ параметры передаются по ЗНАЧЕНИЮ (копируются)
func inc(x int) {
    x++  // изменяется только локальная копия
}

func main() {
    a := 5
    inc(a)
    println(a) // 5 — исходное значение не изменилось
}
```
