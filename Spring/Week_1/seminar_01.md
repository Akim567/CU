
# Список практических заданий по Spring Boot (от голого проекта до продвинутого DI)

## Задание 1: Создание Spring Boot проекта без Spring Initializr
**Цель:** Понимание минимальной конфигурации Spring Boot

**Задача:**
1. Создайте новый Maven проект вручную (без Spring Initializr)
2. Добавьте в `pom.xml`:
    - Родительский проект `spring-boot-starter-parent`
    - Зависимость `spring-boot-starter`
    - Plugin `spring-boot-maven-plugin`
3. Создайте основной класс `Application.java` с аннотацией `@SpringBootApplication`
4. Реализуйте метод `main`, который запускает SpringApplication
5. Убедитесь, что приложение запускается (должен быть виден лог Spring)

**Ожидаемый результат:** Пустое Spring Boot приложение, которое успешно запускается и останавливается.

---

## Задание 2: Добавление и конфигурация модулей
**Цель:** Работа с разными Spring Boot Starter зависимостями

**Задача:**
1. К проекту добавьте:
    - `spring-boot-starter-web` (для веб-приложения)
    - `spring-boot-starter-test` (для тестирования)

2. Настройте `application.yaml` для:
    - Порта сервера (не 8080)
    - Логирования уровня DEBUG для Spring

**Ожидаемый результат:** Многомодульное приложение, которое запускается на кастомном порту.

## Задание 3: Создание и внедрение простых бинов
**Цель:** Базовое понимание DI в Spring Boot

**Задача:**
1. Создайте интерфейс `NotificationService` с методом `send(String message)`
2. Создайте две реализации:
    - `EmailNotificationService` (выводит "Email: {message}")
    - `SmsNotificationService` (выводит "SMS: {message}")

3. Создайте контроллер `NotificationController`, который:
    - Получает `NotificationService` через constructor injection
    - Имеет эндпоинт `POST /notify` с параметром `message`
    - Вызывает `notificationService.send(message)`

**Ожидаемый результат:** Контроллер, который использует конкретную реализацию.

---

## Задание 4: Внедрение нескольких бинов одного типа (Field Injection)
**Цель:** Понимание проблем при field injection

**Задача:**
1. Создайте три реализации интерфейса `PaymentProcessor`:
    - `CreditCardProcessor` ("Processing via Credit Card")
    - `PayPalProcessor` ("Processing via PayPal")
    - `CryptoProcessor` ("Processing via Cryptocurrency")

2. Создайте сервис `PaymentService` с тремя полями:
   ```java
   @Autowired
   private PaymentProcessor creditCardProcessor;
   
   @Autowired
   private PaymentProcessor payPalProcessor;
   
   @Autowired
   private PaymentProcessor cryptoProcessor;
   ```

3. Попробуйте запустить приложение. Получите и разберите ошибку `NoUniqueBeanDefinitionException`

4. Исправьте ошибку с помощью `@Qualifier` на полях и на классах реализаций

**Ожидаемый результат:** Понимание проблемы ambiguous dependency и способов её решения.

---

## Задание 5: Внедрение бина через setter-инъекцию
**Цель:**  Освоить setter-based dependency injection.

**Задача:**

1. Создать интерфейс Calculator с методом int add(int a, int b).
2. Создать реализацию SimpleCalculator.
3. В сервисе MathService внедрить Calculator через setter с @Autowired.
4. Убедиться, что бин корректно инжектится.

**Ожидаемый результат:** Понимание различных способов инъекции.

---

## Задание 6: Внедрение через конструктор с @Qualifier
**Цель:** Лучшая практика DI с квалификаторами

**Задача:**
1. Создайте интерфейс `ReportGenerator` с методом `generate(String data)`
2. Создайте три реализации:
    - `PdfReportGenerator` ("Generating PDF report: {data}")
    - `ExcelReportGenerator` ("Generating Excel report: {data}")
    - `HtmlReportGenerator` ("Generating HTML report: {data}")

3. Создайте три контроллера, каждый из которых получает свою реализацию через constructor injection:
   ```java
   // PdfReportController
   public PdfReportController(@Qualifier("pdfReportGenerator") ReportGenerator generator)
   
   // ExcelReportController
   public ExcelReportController(@Qualifier("excelReportGenerator") ReportGenerator generator)
   
   // HtmlReportController
   public HtmlReportController(@Qualifier("htmlReportGenerator") ReportGenerator generator)
   ```

4. Создайте эндпоинты в каждом контроллере для генерации отчетов

**Ожидаемый результат:** Раздельное использование разных реализаций одного интерфейса.

---

## Задание 7: Использование @Primary для выбора реализации по умолчанию
**Цель:** Работа с приоритетными бинами

**Задача:**
1. Создайте интерфейс `CacheService` с методами `put(String key, String value)` и `get(String key)`
2. Создайте две реализации:
    - `MemoryCacheService` (хранит в HashMap) - пометьте `@Primary`
    - `RedisCacheService` (заглушка, которая пишет "Would save to Redis") - без `@Primary`

3. Создайте сервис `DataService`, который зависит от `CacheService` (через constructor без @Qualifier)

4. Создайте контроллер, который использует `DataService` и проверьте, что используется `MemoryCacheService`

5. Уберите `@Primary` и добавьте `@Qualifier("redisCacheService")` в `DataService`

**Ожидаемый результат:** Понимание разницы между `@Primary` и `@Qualifier`.

---
## Задание 8: Внедрение коллекций бинов
**Цель:** Работа с List, Set и Map внедрений

**Задача:**
1. Создайте интерфейс `ValidationRule` с методом `boolean validate(String input)`
2. Создайте 4-5 реализаций (например: `NotEmptyRule`, `EmailFormatRule`, `MinLengthRule`, `MaxLengthRule`, `NoSpacesRule`)

3. Создайте `ValidationService`, который внедряет:
   ```java
   @Autowired
   private List<ValidationRule> allRules;
   
   @Autowired
   private Set<ValidationRule> uniqueRules;
   
   @Autowired
   private Map<String, ValidationRule> ruleMap;
   ```

4. Реализуйте метод `validateAll(String input)`, который применяет все правила и возвращает список ошибок

5. Добавьте `@Order` к некоторым правилам и посмотрите, как меняется порядок в List

**Ожидаемый результат:** Использование паттерна "Chain of Responsibility" через Spring DI.

---
## Задание 9: Внедрение значений из конфигурационных файлов с помощью @Value
**Цель:**  Научиться использовать @Value для инъекции значений из конфигурации.

**Задача:**

1. Создайте application.properties или application.yaml со следующими свойствами:

```yaml
app:
   name: "My Spring App"
   version: "1.0.0"
   timeout: 5000
```

2. Создайте компонент AppInfoService, который внедряет эти значения с помощью @Value.
3. Добавьте метод String getAppInfo(), возвращающий строку с названием и версией.
4. Создайте REST-эндпоинт, который использует этот сервис.

**Ожидаемый результат:** Умение работать с внешней конфигурацией через @Value.

---
## Задание 10: Внедрение списка бинов простого типа
**Цель:**  Научиться инжектить коллекции простых значений.

**Задача:**
1. В application.yaml определить список строк:

```yaml
app:
   features:
      - "auth"
      - "logging"
      - "caching"
```
2. Создать компонент FeatureService, который через @Value внедряет этот список.
3. Вывести все фичи в лог при старте.

**Ожидаемый результат:** Работа с коллекциями значений из конфигурации.

---

## Задание 11: Одновременное использование @Primary и @Qualifier
**Цель:**  Научиться комбинировать @Primary и @Qualifier.

**Задача:**
1. Создайте интерфейс Logger с методом void log(String message).
2. Создайте две реализации:

> ConsoleLogger (пометить @Primary)  
> FileLogger (с @Qualifier("fileLogger"))

3. Создайте сервис LoggingService, который внедряет Logger через конструктор без @Qualifier.
4. Создайте второй сервис AuditService, который внедряет Logger с @Qualifier("fileLogger").
5. Убедитесь, что в LoggingService используется ConsoleLogger, а в AuditService — FileLogger.

**Ожидаемый результат:** Понимание приоритета @Qualifier над @Primary.

---
## Задание 12: Создание бина через @Configuration и @Bean с зависимостью
**Цель:** Понять разницу между @Component и @Bean.

**Задача:**
1. Создайте класс ExternalApiClient (без аннотаций) с конструктором, принимающим String apiUrl.
2. Создайте @Configuration класс ApiConfig, в котором определите бин:

```java
@Bean
public ExternalApiClient apiClient(@Value("${app.api.url}") String url) {
    return new ExternalApiClient(url);
}
```
4. Настройте свойство app.api.url в конфигурации.
5. Внедрите ExternalApiClient в сервис и используйте его.

**Ожидаемый результат:** Понимание Java-based конфигурации и создания бинов из сторонних классов.

---

## Задание 13: Проблема циклических зависимостей и как DI помогает их обнаружить
**Цель:** Понять, как DI помогает выявлять проблемы архитектуры

**Задача:**
1. Создайте циклическую зависимость: ServiceA зависит от ServiceB, ServiceB зависит от ServiceA
2. Покажите, как Spring обнаруживает эту проблему при запуске
3. Решите проблему через рефакторинг (выделение общего функционала в третий сервис)
4. Покажите, как сложно было бы обнаружить эту проблему без IoC контейнера

**Ожидаемый результат:** Понимание, как DI помогает поддерживать чистую архитектуру.

---
## Задание 14: Демонстрация проблемы жесткой связности (Tight Coupling)
**Цель:** Почувствовать проблему, которую решает DI

**Задача:**
1. Создайте класс OrderService без Spring:

```java
public class OrderService {
private EmailService emailService = new EmailService();

    public void processOrder(Order order) {
        // логика обработки заказа
        emailService.sendConfirmation(order);
    }
}
```

2. Создайте класс EmailService с методом sendConfirmation
3. Попробуйте протестировать OrderService без отправки реальных писем
4. Попробуйте заменить EmailService на SmsService

**Ожидаемый результат:** Понимание, почему жесткая связность мешает тестированию и модификации кода.

---
## Задание 15: Рефакторинг кода с использованием DI
**Цель:** Увидеть, как DI решает проблему жесткой связности

**Задача:**
1. Возьмите код из задания 14
2. Создайте интерфейс NotificationService
3. Переделайте OrderService:

```java
@Service
public class OrderService {
private final NotificationService notificationService;

    public OrderService(NotificationService notificationService) {
        this.notificationService = notificationService;
    }
    
    public void processOrder(Order order) {
        notificationService.sendConfirmation(order);
    }
}
```

4. Создайте тест, где подменяете NotificationService моком

**Ожидаемый результат:** Понимание, как DI делает код гибким и тестируемым.
