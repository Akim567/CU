# Лекция: Структуры и интерфейсы в Go

## Зачем нужны структуры?

Представьте, что вы пишете программу для управления сотрудниками компании. Для каждого сотрудника нужно хранить имя, возраст, должность и зарплату. Можно использовать отдельные переменные:

```go
var (
    employeeName     = "Анна Иванова"
    employeeAge      = 28
    employeePosition = "Разработчик"
    employeeSalary   = 120000
)
```

Но что если сотрудников сотни? Как передать все эти связанные данные в функцию и не потерять между ними связь? Как организовать код так, чтобы он был читаемым и поддерживаемым?

Именно для этого в Go существуют **структуры** — составной тип данных, который позволяет объединять связанные переменные в одну единицу данных.

---

## Структуры: группировка данных

**Структура** в Go — это составной тип данных, который позволяет объединить именованные поля (возможно, разных типов) под одним именем. Структуры помогают организовать связанные данные и служат основой моделирования данных в Go.

### Объявление структуры

Структура объявляется с помощью ключевых слов `type` и `struct`:

```go
type Employee struct {
    Name     string
    Age      int
    Position string
    Salary   int
}
```

### Создание значений структурного типа

Существует несколько способов задать значение структуре:

**Способ 1: Литерал структуры с именованными полями**
```go
emp1 := Employee{
    Name:     "Анна Иванова",
    Age:      28,
    Position: "Разработчик",
    Salary:   120000,
}
```

**Способ 2: Литерал структуры с полями по порядку (он же сокращенный вариант)**
```go
emp2 := Employee{"Петр Петров", 32, "Тестировщик", 100000}
```

**Способ 3: Нулевое значение с последующим заполнением**
```go
var emp3 Employee
emp3.Name = "Мария Сидорова"
emp3.Age = 25
emp3.Position = "Аналитик"
emp3.Salary = 110000
```

> [!NOTE]
> У структуры каждое поле по умолчанию имеет нулевое значение по типу (для `string` - `""`, для `int` - `0`, для `bool` - `false`, для составных типов - `nil` и т.д.).

### Доступ к полям структуры

Для доступа к полям структуры используется точечная нотация:

```go
fmt.Println("Имя:", emp1.Name)
fmt.Println("Зарплата:", emp1.Salary)

// Изменение значений
emp1.Salary = 130000
emp1.Position = "Старший разработчик"
```

## Критически важные особенности структур

### Структуры — это значения!

Перед тем как рассматривать передачу структур в функции, важно понимать фундаментальную особенность структур в Go: **структуры являются типами-значениями**.

Это означает, что при присваивании одной структуры другой или при передаче структуры в функцию **создается полная копия** всех полей структуры:

```go
func main() {
    emp1 := Employee{Name: "Анна", Salary: 100000}
    emp2 := emp1 // создается копия значения!
    
    emp2.Salary = 200000 // изменяем копию
    
    fmt.Println(emp1.Salary) // 100000 - оригинал не изменился
    fmt.Println(emp2.Salary) // 200000 - изменилась только копия
}
```

### Передача структур в функции: по значению vs по указателю

Давайте рассмотрим пример:

```go
func printEmployee(emp Employee) {
    fmt.Printf("Сотрудник: %s, Возраст: %d, Должность: %s, Зарплата: %d\n",
        emp.Name, emp.Age, emp.Position, emp.Salary)
}

// Эта функция НЕ изменит оригинальную структуру!
func tryToIncreaseSalary(emp Employee, amount int) {
    emp.Salary += amount // изменяем только копию
    fmt.Printf("Внутри функции зарплата: %d\n", emp.Salary)
}

// Эта функция изменит оригинальную структуру
func increaseSalary(emp *Employee, amount int) {
    emp.Salary += amount // изменяем оригинал через указатель
}

func main() {
    emp := Employee{
        Name:     "Анна Иванова",
        Age:      28,
        Position: "Разработчик",
        Salary:   120000,
    }

    printEmployee(emp)
    
    // Попытка изменить через копию значения - НЕ влияет на оригинал
    tryToIncreaseSalary(emp, 10000)
    fmt.Printf("После tryToIncreaseSalary: %d\n", emp.Salary) // 120000 - не изменилось!
    
    // Изменение через указатель - влияет на оригинал
    increaseSalary(&emp, 10000)
    fmt.Printf("После increaseSalary: %d\n", emp.Salary) // 130000 - изменилось!
}
```

**Почему так происходит?**

Когда мы вызываем `tryToIncreaseSalary(emp, 10000)`, Go создает полную копию структуры `emp` и передает эту копию в функцию. Все изменения происходят с копией, а оригинальная структура остается нетронутой.

Когда мы вызываем `increaseSalary(&emp, 10000)`, мы передаем **адрес** структуры (указатель). Функция получает доступ к той же области памяти и меняет исходное значение.

> [!NOTE]
> Важно понимать следующее:
> 
> **Производительность**: Копирование больших структур может быть дорогим. Если структура содержит много полей крупных типов, эффективней передавать указатель на структуру, а не саму структуру.
>
> **Безопасность**: Передача по значению гарантирует, что функция не может случайно изменить исходные данные. Это делает код более предсказуемым, но требует осознанного выбора между копией и указателем.

## Вложенные структуры

Структуры могут содержать другие структуры в качестве полей. Это позволяет создавать составные типы любой сложности:

```go
type Address struct {
    Street string
    City   string
    Zip    string
}

type Person struct {
    Name    string
    Age     int
    Address Address // вложенная структура
}

func main() {
    person := Person{
        Name: "Анна Иванова",
        Age:  28,
        Address: Address{
            Street: "ул. Ленина, 10",
            City:   "Москва",
            Zip:    "101000",
        },
    }

    fmt.Println(person.Name)
    fmt.Println(person.Address.City)
}
```

### Встраивание структур (Embedding)

Go поддерживает встраивание структур, что позволяет создавать композицию типов. **Встраивание (embedding)** — это механизм, при котором поля и методы одной структуры становятся доступными в другой структуре без явного указания имени поля.

Это ключевое отличие от наследования в объектно-ориентированных языках: вместо отношения "является" (is-a) мы создаем отношение "содержит" (has-a), с удобным синтаксисом доступа:

```go
type Contact struct {
    Email string
    Phone string
}

type Employee struct {
    Name     string
    Position string
    Contact  // встроенная структура (анонимное поле)
}

func main() {
    emp := Employee{
        Name:     "Анна Иванова",
        Position: "Разработчик",
        Contact: Contact{
            Email: "anna@company.com",
            Phone: "+7-123-456-7890",
        },
    }

    // Доступ к полям встроенной структуры (короткий синтаксис)
    fmt.Println(emp.Email) // anna@company.com
    fmt.Println(emp.Phone) // +7-123-456-7890
    
    // Доступ к полям встроенной структуры (полный синтаксис)
    fmt.Println(emp.Contact.Email) // anna@company.com - также работает
    fmt.Println(emp.Contact.Phone) // +7-123-456-7890
}
```

### Конфликты имен при embedding

Что происходит, когда у встраиваемой структуры и внешней структуры есть поля с одинаковыми именами? В таких случаях поле внешней структуры **скрывает** поле встроенной:

```go
package main

import "fmt"

type Base struct {
	Name string
	ID   int
}

type Extended struct {
	Name string // скрывает Base.Name
	Age  int    //
	Base        // встроенная структура
}

func main() {
	ext := Extended{
		Name: "Основное имя",
		Age:  25,
		Base: Base{
			Name: "Встроенное имя",
			ID:   123,
		},
	}

	fmt.Println(ext.Name)      // "Основное имя" - поле Extended.Name
	fmt.Println(ext.Base.Name) // "Встроенное имя" - поле Base.Name
	fmt.Println(ext.ID)        // 123 - поле Base.ID (доступно напрямую)
}
```

**Правила разрешения конфликтов:**
1. Имена полей, объявленные в самой структуре, имеют приоритет над именами из встраиваемых структур.
2. Если одинаковое имя встречается только у встроенных, выбирается поле/метод с меньшей глубиной (ближе к внешней структуре).
3. Если одинаковые имена приходят из нескольких встраиваний на одной глубине, обращение не по полному пути приведёт к ошибке компиляции из-за неоднозначности: потребуется использовать полный путь (`ext.Base.Name`).
4. Методы подчиняются тем же правилам: разрешение конфликтов методов работает аналогично полям при помощи продвижения методов.

```go
type A struct {
	Value string
}

type B struct {
	Value string
}

type Conflicted struct {
	A
	B
}

func main() {
	c := Conflicted{
		A: A{Value: "от A"},
		B: B{Value: "от B"},
	}

	fmt.Println(c.Value)   // Ошибка компиляции: неоднозначность!
	fmt.Println(c.A.Value) // "Поле Value от A"
	fmt.Println(c.B.Value) // "Поле Value от B"
}
```

## Методы структур

В Go можно определять методы для структур. Метод — это функция с получателем (receiver):

```go
type Rectangle struct {
    Width  float64
    Height float64
}

// Метод с получателем по значению
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Метод с получателем по указателю
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

// Метод для красивого вывода (реализует fmt.Stringer)
func (r Rectangle) String() string {
    return fmt.Sprintf("Прямоугольник: %.2f x %.2f", r.Width, r.Height)
}

func main() {
    rect := Rectangle{Width: 10, Height: 5}

    fmt.Println("Площадь:", rect.Area())
    fmt.Println(rect) // используется метод String()

    rect.Scale(2)
    fmt.Println("После масштабирования:", rect)
    fmt.Println("Новая площадь:", rect.Area())
}
```

### Получатель по значению vs по указателю

**Получатель по значению** — метод получает копию структуры:
- Используется когда методу не требуется изменять исходные поля структуры
- Безопасен для конкурентного доступа
- Может быть дороже для больших структур из-за копирования

**Получатель по указателю** — метод получает указатель на структуру:
- Используется когда методу требуется изменять исходные поля структуры
- Избегает копирования, эффективен для передачи больших структур
- Требует осторожности при конкурентном доступе

### Публичные и приватные методы

Как и обычные функции, методы структур могут быть **публичными** (доступными из других пакетов) или **приватными** (доступными только внутри текущего пакета). В терминологии Go это называют **экспортируемыми** и **неэкспортируемыми** идентификаторами. Правило простое: если имя метода начинается с заглавной буквы — метод **экспортируемый**, если с маленькой — **неэкспортируемыми**.

```go
type BankAccount struct {
    balance float64 // Неэкспортируемое поле
    Owner   string  // Экспортируемое поле
}

// Экспортируемый метод - доступен из других пакетов
func (ba *BankAccount) Deposit(amount float64) {
    if amount > 0 {
        ba.balance += amount
    }
}

// Экспортируемый метод - доступен из других пакетов
func (ba *BankAccount) GetBalance() float64 {
    return ba.balance
}

// Неэкспортируемый метод - доступен только внутри текущего пакета
func (ba *BankAccount) validateTransaction(amount float64) bool {
    return amount > 0 && amount <= ba.balance
}

// Экспортируемый метод, использующий неэкспортируемый метод
func (ba *BankAccount) Withdraw(amount float64) bool {
    if ba.validateTransaction(amount) {
        ba.balance -= amount
        return true
    }
    return false
}
```

**Практическая польза:**
- **Экспортируемые методы** формируют публичный API вашего типа — то, как другие пакеты могут взаимодействовать с вашей структурой
- **Неэкспортируемые методы** инкапсулируют внутреннюю логику, которую не нужно выставлять наружу
- Это позволяет изменять внутреннюю реализацию, не ломая код, который использует экспортируемые методы вашей структуры

```go
type Counter struct {
    value int
}

// Метод по значению - НЕ изменяет оригинал
func (c Counter) IncrementWrong() {
    c.value++ // изменяем копию
}

// Метод по указателю - изменяет оригинал
func (c *Counter) Increment() {
    c.value++ // изменяем оригинал
}

func (c Counter) Value() int {
    return c.value
}

func main() {
    counter := Counter{value: 0}

    counter.IncrementWrong()
    fmt.Println(counter.Value()) // 0 - не изменилось

    counter.Increment()
    fmt.Println(counter.Value()) // 1 - изменилось
}
```

### Конфликт методов при embedding

Как и с полями, при встраивании структур могут возникать конфликты методов. Go применяет те же правила приоритета, что и для полей:

```go
type Writer struct{}

func (w Writer) Write() {
	fmt.Println("Writer.Write() вызван")
}

func (w Writer) Process() {
	fmt.Println("Writer.Process() вызван")
}

type Logger struct{}

func (l Logger) Write() {
	fmt.Println("Logger.Write() вызван")
}

func (l Logger) Log() {
	fmt.Println("Logger.Log() вызван")
}

type FileHandler struct {
	Writer
	Logger
}

// Собственный метод FileHandler "скрывает" одноимённые методы встроенных структур
func (fh FileHandler) Write() {
	fmt.Println("FileHandler.Write() вызван")
	// Можно вызвать методы встроенных структур явно
	fh.Writer.Write()
	fh.Logger.Write()
}

func main() {
	fh := FileHandler{}

	fh.Write()   // FileHandler.Write() - собственный метод
	fh.Process() // Writer.Process() - метод Writer, доступный через встраивание
	fh.Log()     // Logger.Log() - метод Logger, доступный через встраивание

	// Явный вызов методов встроенных структур
	fh.Writer.Write() // Writer.Write()
	fh.Logger.Write() // Logger.Write()
}
```

**Важные моменты при работе с методами:**

1. **Продвижение методов**: Методы встроенной структуры автоматически становятся доступны у внешней (они "продвигаются")
2. **Приоритет внешней**: Метод, объявленный у внешнего типа, скрывает одноимённый метод встроенной структуры
3. **Множественное встраивание и неоднозначность**: Если два встроенных типа на одной глубине имеют одинаковый метод, обращение `fh.Write()` без уточнения — ошибка компиляции. Используйте явный путь (`fh.Writer.Write()`/`fh.Logger.Write()`). Само объявление допустимо — ошибка возникает при использовании селектора

```go
type Printer interface {
	Print()
}

type BasicPrinter struct{}

func (bp BasicPrinter) Print() {
	fmt.Println("Basic print")
}

type AdvancedDevice struct {
	BasicPrinter // встраиваем структуру с методом Print()
}

func main() {
	device := AdvancedDevice{}

	// Благодаря embedding, AdvancedDevice удовлетворяет Printer:
	var printer Printer = device
	printer.Print() // "Basic print"

	// Также можно вызвать напрямую:
	device.Print() // "Basic print"
}
```

## Интерфейсы: определение поведения

**Интерфейс** в Go — это тип, который определяет набор методов. Интерфейсы описывают поведение, а не данные. Любой тип, у которого есть все методы интерфейса, неявно реализует этот интерфейс.

### Объявление интерфейса

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}
```

### Реализация интерфейса

В Go не нужно явно указывать, что тип реализует интерфейс. Если тип имеет все методы интерфейса, он автоматически удовлетворяет интерфейсу:

```go
type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return 3.14159 * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * 3.14159 * c.Radius
}
```

### Использование интерфейсов

```go
func printShapeInfo(s Shape) {
    fmt.Printf("Площадь: %.2f\n", s.Area())
    fmt.Printf("Периметр: %.2f\n", s.Perimeter())
}

func main() {
    rect := Rectangle{Width: 10, Height: 5}
    circle := Circle{Radius: 3}

    // Один и тот же код работает с разными типами
    fmt.Println("Прямоугольник:")
    printShapeInfo(rect)

    fmt.Println("\nКруг:")
    printShapeInfo(circle)
}
```

## Пустой интерфейс и type assertion

### Пустой интерфейс interface{} и тип any

**Пустой интерфейс** `interface{}` не содержит ни одного метода в своем определении. Поскольку любой тип в Go может иметь ноль методов, **любой тип удовлетворяет пустому интерфейсу**. Это делает `interface{}` универсальным типом, способным хранить значения любого типа.

```go
func printAnything(value interface{}) {
	fmt.Printf("Значение: %v, Тип: %T\n", value, value)
}

func main() {
	printAnything(42)                              // int
	printAnything("Hello")                         // string
	printAnything([]int{1, 2, 3})                  // []int
	printAnything(Rectangle{Width: 10, Height: 5}) // Rectangle
	printAnything(nil)                             // <nil>
}
```

### Тип any (Go 1.18+)

Начиная с `Go 1.18`, был введен псевдоним `any` для `interface{}`. Это было сделано для улучшения читаемости кода:

```go
// Старый способ
func processOld(value interface{}) { }

// Новый способ (Go 1.18+)
func processNew(value any) { }

// Это абсолютно одинаковые функции!
```

**Когда использовать `any`:**
- В новом коде лучше использовать `any` — он короче и яснее выражает намерения
- В библиотеках, которые должны работать с версиями Go до 1.18, используйте `interface{}` для обратной совместимости

```go
// Пример использования any в структуре данных
type Cache struct {
    data map[string]any
}

func (c *Cache) Set(key string, value any) {
    c.data[key] = value
}

func (c *Cache) Get(key string) (any, bool) {
    v, ok := c.data[key]
    return v, ok
}

func main() {
    cache := Cache{data: make(map[string]any)}
    
    cache.Set("user_id", 12345)
    cache.Set("username", "john_doe")
    cache.Set("scores", []int{95, 87, 92})
    cache.Set("active", true)
    
    if value, ok := cache.Get("username"); ok {
        fmt.Printf("Username: %v\n", value)
    }
}
```

### Ограничения и особенности пустого интерфейса

**1. Потеря статической типизации**
```go
func demonstrateTypeLoss() {
    var x any = 42
    
    fmt.Println(x + 10) // Ошибка компиляции! Go не знает, что x — это число
    
    // Нужно приведение типа через type assertion
    if num, ok := x.(int); ok {
        fmt.Println(num + 10) // 52
    }
}
```

**2. Производительность**
```go
// Медленнее - нужны проверка типов и приведения
func slowSum(values []any) int {
    sum := 0
    for _, v := range values {
        if num, ok := v.(int); ok {
            sum += num
        }
    }
    return sum
}

// Быстрее - работа с конкретным типом
func fastSum(values []int) int {
    sum := 0
    for _, v := range values {
        sum += v
    }
    return sum
}
```

**3. Отсутствие проверки типов на этапе компиляции**
```go
func riskyCode() {
    var data any = "строка"
    
    // Код ниже скомпилируется, но упадет в runtime!
    number := data.(int) // panic: interface conversion: interface {} is string, not int
    
    // Безопасный вариант
    if number, ok := data.(int); ok {
        fmt.Println("Число:", number)
    } else {
        fmt.Println("Это не число")
    }
}
```

### Type Assertion

Для извлечения конкретного типа из значения интерфейса используется **type assertion**:

```go
func processValue(value interface{}) {
    // Безопасная проверка через идиому "comma, ok"
    if str, ok := value.(string); ok {
        fmt.Printf("Это строка: %s (длина: %d)\n", str, len(str))
    } else if num, ok := value.(int); ok {
        fmt.Printf("Это число: %d (квадрат: %d)\n", num, num*num)
    } else {
        fmt.Println("Неизвестный тип")
    }
}

func main() {
    processValue("Hello")
    processValue(42)
    processValue(3.14)
}
```

### Type Switch

Более элегантный способ работы с разными типами через type switch:

```go
func describe(value interface{}) {
    switch v := value.(type) {
    case string:
        fmt.Printf("Строка длиной %d: %s\n", len(v), v)
    case int:
        fmt.Printf("Целое число: %d\n", v)
    case float64:
        fmt.Printf("Число с плавающей точкой: %.2f\n", v)
    case Rectangle:
        fmt.Printf("Прямоугольник: площадь = %.2f\n", v.Area())
    default:
        fmt.Printf("Неизвестный тип: %T\n", v)
    }
}
```

## Method set и правила реализации интерфейсов

**Method set** — это набор методов, доступных у значения определенного типа. Понимание method set критически важно для работы с интерфейсами в Go, особенно когда речь идет о методах с получателями-указателями и получателями-значениями.

### Правила method set

**Для типа T (значение):**
- Method set содержит методы с получателем `(t T)`
- НЕ содержит методы с получателем `(t *T)`

**Для типа *T (указатель):**
- Method set содержит методы с получателем `(t T)` И `(t *T)`

```go
type Document struct {
    title string
    content string
}

// Метод с получателем по значению
func (d Document) GetTitle() string {
    return d.title
}

// Метод с получателем по указателю
func (d *Document) SetTitle(title string) {
    d.title = title
}

// Интерфейс только с методом по значению
type Reader interface {
    GetTitle() string
}

// Интерфейс с методами по значению и указателю
type Writer interface {
    GetTitle() string
    SetTitle(string)
}

func main() {
    doc := Document{title: "Заголовок", content: "Содержимое"}
    docPtr := &Document{title: "Заголовок", content: "Содержимое"}
    
    // Document (значение) удовлетворяет Reader
    var r1 Reader = doc    // OK
    var r2 Reader = docPtr // OK (указатель содержит методы значения)
    
    // Writer требует SetTitle (указательный метод)
    var w1 Writer = doc    // Ошибка компиляции: method set T не включает SetTitle
    var w2 Writer = docPtr // OK
}
```

### Практические последствия method set

**1. Слайсы значений vs слайсы указателей с интерфейсами**
```go
type Processor interface {
    Process()
}

type Task struct {
    name string
}

func (t Task) Process() {
    fmt.Printf("Обработка задачи: %s\n", t.name)
}

func (t *Task) SetName(name string) {
    t.name = name
}

func processTasks(processors []Processor) {
    for _, p := range processors {
        p.Process()
    }
}

func main() {
    tasks := []Task{
        {name: "Задача 1"},
        {name: "Задача 2"},
    }
    
    // Нельзя преобразовать []Task в []Processor напрямую (разные представления)
    // Заполняем вручную:
    var processors []Processor
    for _, task := range tasks {
        processors = append(processors, task) // OK: Task имеет метод Process() по значению
    }
    
    processTasks(processors)
}
```

**2. Когда нужны указатели в интерфейсах**
```go
type Validator interface {
    Validate() error
    Fix() // метод изменяет объект - нужен получатель-указатель
}

type User struct {
    Email string
    Age   int
}

func (u User) Validate() error {
    if u.Age < 18 {
        return fmt.Errorf("возраст должен быть не менее 18")
    }
    if !strings.Contains(u.Email, "@") {
        return fmt.Errorf("некорректный email")
    }
    return nil
}

// Метод изменяет значение оригинальной структуры - нужен получатель по указателю
func (u *User) Fix() {
    if u.Age < 18 {
        u.Age = 18
    }
    if !strings.Contains(u.Email, "@") {
        u.Email += "@example.com"
    }
}

func validateAndFix(v Validator) {
    if err := v.Validate(); err != nil {
        fmt.Printf("Ошибка валидации: %v\n", err)
        v.Fix()
        fmt.Println("Данные исправлены")
    }
}

func main() {
    user := User{Email: "john", Age: 16}
    
    validateAndFix(&user) // OK: Только *User удовлетворяет Validator
    validateAndFix(user) // Ошибка компиляции!
    
    fmt.Printf("Результат: %+v\n", user)
}
```

### Автоматическое разыменование

Go автоматически преобразует между указателями и значениями при вызове методов:

```go
type Counter struct {
    value int
}

func (c *Counter) Increment() {
    c.value++
}

func (c Counter) GetValue() int {
    return c.value
}

func main() {
    // Значение
    counter1 := Counter{value: 0}
    counter1.Increment() // Компилятор автоматически подставит (&counter1).Increment()
    
    // Указатель  
    counter2 := &Counter{value: 0}
    value := counter2.GetValue() // Компилятор автоматически подставит (*counter2).GetValue()
    
    fmt.Println(counter1.GetValue(), value)
}
```
> [!IMPORTANT]
> Авто-разыменование работает только в момент вызова метода, но не изменяет method set типа. Поэтому при присваивании интерфейсу учитывается именно method set. Но это НЕ работает при присваивании интерфейсам!

## Композиция интерфейсов

Интерфейсы можно комбинировать для создания более сложных контрактов:

```go
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

type Closer interface {
    Close() error
}

// Композиция интерфейсов
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

Пример простой реализации:

```go
type File struct {
    name string
    data []byte
}

func (f *File) Read(p []byte) (int, error) {
    copy(p, f.data)
    return len(f.data), nil
}

func (f *File) Write(p []byte) (int, error) {
    f.data = append(f.data, p...)
    return len(p), nil
}

func (f *File) Close() error {
    fmt.Printf("Файл %s закрыт\n", f.name)
    return nil
}

func processFile(rwc ReadWriteCloser) {
    data := []byte("Hello, World!")
    rwc.Write(data)

    buffer := make([]byte, len(data))
    rwc.Read(buffer)
    fmt.Printf("Прочитано: %s\n", string(buffer))

    rwc.Close()
}
```

### Проектирование интерфейсов

1. **Держите интерфейсы маленькими** — лучше много маленьких интерфейсов, чем один большой
2. **Определяйте интерфейсы там, где используете** — а не в пакете с реализацией. Так делается потому что интерфейс отражает потребности клиента, а не возможности провайдера
3. **Используйте осмысленные имена** — `Writer`, `Reader`, `Validator`
4. **Принимайте интерфейсы — возвращайте конкретные типы** — вызывающему полезны реальные поля/методы результата: абстракция нужна на входе, где вы "потребляете" поведение
5. **Не вводите интерфейс "на всякий случай"** — если у вас единственная реализация и нет места, где нужна подмена, начните с конкретного типа. Интерфейс добавьте, когда появится нужда в вариативности
6. **Иногда лучше дженерики, чем interface{}/any** — если цель — обобщить алгоритм над типами, чище использовать дженерики

## Заключение

Структуры и интерфейсы — это основные строительные блоки для создания хорошо организованного и масштабируемого кода в Go:

- **Структуры** помогают группировать связанные данные и определять методы для работы с ними
- **Интерфейсы** определяют поведение и позволяют писать гибкий, тестируемый код
- **Встраивание структур (embedding)** обеспечивает композицию без наследования
- **Неявное удовлетворение интерфейсов** делает код более гибким и менее связанным

Правильное использование структур и интерфейсов поможет вам писать код, который легко читать, тестировать и расширять.
