# **План семинара: Практикум по Spring MVC & RESTful API**

**Общая цель:** Закрепить понимание жизненного цикла запроса в Spring MVC, научиться гибко управлять HTTP-ответами, централизованно обрабатывать ошибки, настраивать CORS и документировать API.

---

## **Блок 1: Разминка и DispatcherServlet (Основы)**

### **Задание 1: "Внедрение логгера"**
*   **Цель:** Добавить простое логирование входящих запросов через `HandlerInterceptor`.
*   **Пошаговая инструкция:** Создаем свой `HandlerInterceptor`, логируем метод и URI запроса в методе `preHandle`. Регистрируем его в `WebMvcConfigurer` для всех эндпоинтов `/api/**`.
*   **Код:**

```java
@Component
public class LoggingInterceptor implements HandlerInterceptor {
    private static final Logger log = LoggerFactory.getLogger(LoggingInterceptor.class);
    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        log.info(">>> Request: {} {}", request.getMethod(), request.getRequestURI());
        return true;
    }
}

@Configuration
public class WebConfig implements WebMvcConfigurer {
    private final LoggingInterceptor loggingInterceptor;

    @Autowired
    public WebConfig(LoggingInterceptor loggingInterceptor){
        this.loggingInterceptor=loggingInterceptor;
    }
    
    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(loggingInterceptor).addPathPatterns("/api/**");
    }
}
```
*   **Результат:** Понимание, как перехватывать запросы ДО того, как они попадут в контроллер, и видеть место интерцепторов в архитектуре.

### **Задание 2: "Смотрим маппинги"**
*   **Цель:** Научиться интроспекции Spring-контекста и просмотру зарегистрированных эндпоинтов.
*   **Пошаговая инструкция:** Создаем `@RestController` с эндпоинтом `GET /debug/mappings`. Инжектим `RequestMappingHandlerMapping` и возвращаем map всех зарегистрированных в системе URL-путей и методов.
```java
@RestController
@RequestMapping("internal/api/debug")
public class DebugController {
    @Autowired private RequestMappingHandlerMapping handlerMapping;

    @GetMapping("/mappings")
    public Map<String, String> getMappings() {
        Map<String, String> mappings = new HashMap<>();
        handlerMapping.getHandlerMethods().forEach((info, method) ->
            mappings.put(info.toString(), method.getBeanType().getName() + "#" + method.getMethod().getName())
        );
        return mappings;
    }
}
```
*   **Результат:** Понимание, как Spring хранит информацию о маршрутах, и использовать это для отладки.

### **Задание 3: "Разрешаем конфликт путей"**
*   **Цель:** Научиться разрешать неоднозначность маппинга URL с помощью регулярных выражений.
*   **Пошаговая инструкция:** Создаем два метода в контроллере: `@GetMapping("/{id:\\d+}")` и `@GetMapping("/{name:[a-zA-Z]+}")`. Первый принимает `Long`, второй — `String`.
```java 
@RestController
@RequestMapping("/api/items")
public class ItemController {
    @GetMapping("/{id:\\d+}")          // только цифры → Long
    public String getById(@PathVariable Long id) {
        return "Item by id: " + id;
    }

    @GetMapping("/{name:[a-zA-Z]+}")   // только буквы → String
    public String getByName(@PathVariable String name) {
        return "Item by name: " + name;
    }
}
```
*   **Результат:** Понимание, что просто `@GetMapping("/{param}")` для двух методов недопустим, и как использовать regexp для точного роутинга.

---

## **Блок 2: Управление ответами и ResponseEntity (Базовый уровень)**

### **Задание 4: "Создание ресурса с Location"**
*   **Цель:** Научиться правильно возвращать статус `201 Created` и заполнять заголовок `Location`.
*   **Пошаговая инструкция:** В методе `@PostMapping` создаем новый объект (имитация сохранения в БД с генерацией ID). Формируем URI с помощью `ServletUriComponentsBuilder.fromCurrentRequest().path("/{id}").buildAndExpand(createdId).toUri()`. Возвращаем `ResponseEntity.created(location).body(createdObject)`.
```java
@Data
public class User {
    private Long id;
    private String name;
    private String email;
}

@RestController
@RequestMapping("/api/users")
public class UserController {
    private final Map<Long, User> storage = new ConcurrentHashMap<>();
    private final AtomicLong idGen = new AtomicLong(1);

    @PostMapping
    public ResponseEntity<User> createUser(@RequestBody User user) {
        user.setId(idGen.incrementAndGet());
        storage.put(user.getId(), user);
        URI location = ServletUriComponentsBuilder.fromCurrentRequest()
                .path("/{id}")
                .buildAndExpand(user.getId())
                .toUri();
        return ResponseEntity.created(location).body(user);
    }
}
```
*   **Результат:** Следование REST-стандартам при создании ресурсов.

### **Задание 5: "Условный ответ"**
*   **Цель:** Научиться динамически менять HTTP-статус ответа.
*   **Пошаговая инструкция:** Создаем `@GetMapping("/{id}")`. Если объект найден — `ResponseEntity.ok(user)`. Если нет — `ResponseEntity.notFound().build()`.
```java 
@GetMapping("/{id}")
public ResponseEntity<User> getUser(@PathVariable Long id) {
    User user = storage.get(id);
    if (user == null) {
        return ResponseEntity.notFound().build();   // 404
    }
    return ResponseEntity.ok(user);                 // 200
}
```
*   **Результат:** Понимание базового паттерна использования `ResponseEntity` для управления статусами 200 и 404.

### **Задание 6: "Кастомный заголовок пагинации"**
*   **Цель:** Добавить кастомный нестандартный заголовок в ответ.
*   **Пошаговая инструкция:** Для метода `GET /api/users`, который возвращает список пользователей, добавляем в `ResponseEntity` заголовок `X-Total-Count` с общим количеством записей. Используем `.header("X-Total-Count", String.valueOf(totalCount))`.
```java 
@GetMapping
public ResponseEntity<List<User>> getUsers() {
    List<User> users = new ArrayList<>(storage.values());
    return ResponseEntity.ok()
            .header("X-Total-Count", String.valueOf(users.size()))
            .body(users);
}
```
*   **Результат:** Передача мета-информации (не в теле ответа) через заголовки.

---

## **Блок 3: Централизованная обработка ошибок (Продвинутый уровень)**

### **Задание 7: "Создаем кастомное исключение"**
*   **Цель:** Внедрить иерархию собственных исключений для бизнес-логики.
*   **Пошаговая инструкция:** Создаем абстрактный класс `BusinessException extends RuntimeException` и конкретное исключение `UserNotFoundException extends BusinessException`.
```java
public abstract class BusinessException extends RuntimeException {
    public BusinessException(String message) { super(message); }
}

public class UserNotFoundException extends BusinessException {
    public UserNotFoundException(Long id) { super("User not found with id: " + id); }
}
```
*   **Результат:** Выделение бизнес-ошибки в отдельную иерархию для более точной обработки.

### **Задание 8: "Локальный обработчик исключений"**
*   **Цель:** Обработать исключение прямо в контроллере с помощью `@ExceptionHandler`.
*   **Пошаговая инструкция:** В контроллере создаем метод, помеченный `@ExceptionHandler(UserNotFoundException.class)`, который возвращает `ResponseEntity` с телом ошибки и статусом 404.
```java 
// В UserController.java добавить:
@ExceptionHandler(UserNotFoundException.class)
public ResponseEntity<ErrorResponse> handleUserNotFound(UserNotFoundException ex) {
    ErrorResponse error = new ErrorResponse("USER_NOT_FOUND", ex.getMessage(), Instant.now());
    return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error);
}
```
*   **Результат:** Понимание разницы между локальной и глобальной обработкой.

### **Задание 9: "Глобальный обработчик с ControllerAdvice"**
*   **Цель:** Вынести логику обработки исключений в единый компонент.
*   **Пошаговая инструкция:** Создаем класс с аннотацией `@ControllerAdvice`. Внутри определяем метод для `UserNotFoundException` (возвращает 404) и метод для `Exception.class` (возвращает 500).
```java
@Data
@AllArgsConstructor
public class ErrorResponse {
    private String code;
    private String message;
    private Instant timestamp;
}

@ControllerAdvice
public class GlobalExceptionHandler {
    private static final Logger log = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(UserNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleUserNotFound(UserNotFoundException ex) {
        ErrorResponse error = new ErrorResponse("USER_NOT_FOUND", ex.getMessage(), Instant.now());
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleAll(Exception ex) {
        log.error("Unhandled exception", ex);
        ErrorResponse error = new ErrorResponse("INTERNAL_ERROR", "An unexpected error occurred", Instant.now());
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(error);
    }
}
```
*   **Результат:** Создание единой точки обработки ошибок для всего приложения.

### **Задание 10: "Обработка ошибок валидации"**
*   **Цель:** Красиво обрабатывать ошибки валидации (`@Valid`) и возвращать их список клиенту.
*   **Пошаговая инструкция:** Добавляем `@Valid` к `@RequestBody` в POST-методе. В `@ControllerAdvice` создаем обработчик для `MethodArgumentNotValidException`. Собираем все `FieldError` в список строк и возвращаем их в теле ответа со статусом 400.
```java
// Дополним ErrorResponse полем details
@Data
@AllArgsConstructor
public class ErrorResponse {
    private String code;
    private String message;
    private Instant timestamp;
    private List<String> details;          // новое поле

    public ErrorResponse(String code, String message, Instant timestamp) {
        this(code, message, timestamp, null);
    }
}

// В GlobalExceptionHandler добавить:
@ExceptionHandler(MethodArgumentNotValidException.class)
public ResponseEntity<ErrorResponse> handleValidation(MethodArgumentNotValidException ex) {
    List<String> details = ex.getBindingResult().getFieldErrors().stream()
            .map(err -> err.getField() + ": " + err.getDefaultMessage())
            .collect(Collectors.toList());
    ErrorResponse error = new ErrorResponse("VALIDATION_FAILED", "Validation failed", Instant.now(), details);
    return ResponseEntity.badRequest().body(error);
}

// В UserController пометим @Valid
@PostMapping
public ResponseEntity<User> createUser(@Valid @RequestBody User user) { ... }
```
*   **Результат:** Интеграция валидации Bean Validation с глобальным обработчиком ошибок.

---

## **Блок 4: Интеграция и безопасность (Экспертный уровень)**

### **Задание 11: "Настройка CORS под React"**
*   **Цель:** Научиться настраивать CORS для доступа с фронтенда.
*   **Пошаговая инструкция:** В глобальной конфигурации (`WebMvcConfigurer`) разрешаем доступ с `http://localhost:3000` (типичный адрес React-приложения) для методов GET, POST.
```java 
// В WebConfig.java добавить:
@Override
public void addCorsMappings(CorsRegistry registry) {
    registry.addMapping("/api/**")
            .allowedOrigins("http://localhost:3000")
            .allowedMethods("GET", "POST", "PUT", "DELETE");
}
```
*   **Результат:** Понимание, как разрешить браузеру взаимодействовать с API с другого домена.

### **Задание 12: "Точечное разрешение CORS"**
*   **Цель:** Научиться применять более тонкие настройки CORS, в т. Ч. На уровне метода.
*   **Пошаговая инструкция:** Для публичного эндпоинта `GET /api/public/info` ставим аннотацию `@CrossOrigin(origins = "*")`. Для приватного `POST /api/secure/data` оставляем глобальные настройки из Задания 11.
```java 
@RestController
@RequestMapping("/api")
public class PublicController {
    @GetMapping("/info")
    @CrossOrigin(origins = "*")   // разрешено всем
    public String info() {
        return "Public info";
    }
}

// В другом контроллере (например, SecureController) аннотация не нужна – используются глобальные правила
```
*   **Результат:** Использование аннотации `@CrossOrigin` для переопределения глобальных правил.

### **Задание 13: "Смена контейнера сервлетов"**
*   **Цель:** Понять, как легко менять "движок" приложения в Spring Boot.
*   **Пошаговая инструкция:** В `pom.xml` исключаем `spring-boot-starter-tomcat` и добавляем зависимость `spring-boot-starter-undertow`. Запускаем приложение и убеждаемся, что оно работает на Undertow (можно посмотреть логи запуска).
```xml 
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-web</artifactId>
    <exclusions>
        <exclusion>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-tomcat</artifactId>
        </exclusion>
    </exclusions>
</dependency>
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-jetty</artifactId>
</dependency>
```

```groovy
dependencies {
    implementation('org.springframework.boot:spring-boot-starter-web') {
        exclude group: 'org.springframework.boot', module: 'spring-boot-starter-tomcat'
    }
    implementation 'org.springframework.boot:spring-boot-starter-jetty'
}
```
*   **Результат:** Осознание, что Tomcat — не единственный выбор, и Spring Boot позволяет легко менять контейнеры.

### **Задание 14 Подключаем Swagger UI и кастомизируем документацию**

#### **Цель задания:**
Научиться интегрировать OpenAPI (Swagger) в Spring Boot, настроить отображение документации в браузере и обогатить её аннотациями для удобства фронтенд-разработчиков и тестировщиков.

---

### **Пошаговая инструкция**

#### **Шаг 1. Добавьте зависимость springdoc-openapi**

Откройте файл `pom.xml` вашего проекта и добавьте следующую зависимость в раздел `<dependencies>`:

```xml
<dependency>
    <groupId>org.springdoc</groupId>
    <artifactId>springdoc-openapi-starter-webmvc-ui</artifactId>
</dependency>
```

> **Почему именно эта зависимость?**  
> `springdoc-openapi` автоматически интегрируется с Spring Boot и генерирует спецификацию OpenAPI 3 на основе ваших контроллеров. Версия `-ui` дополнительно включает в себя встроенный Swagger UI, который будет доступен по умолчанию.

Если вы используете Gradle, добавьте:
```groovy
implementation 'org.springdoc:springdoc-openapi-starter-webmvc-ui'
```

#### **Шаг 2. Проверьте (или создайте) файл конфигурации `application.yml`**

Springdoc работает "из коробки" без дополнительной настройки, но для удобства можно задать пути к документации. Добавьте в `src/main/resources/application.yml` (или `application.properties`):

```yaml
springdoc:
  api-docs:
    path: /api-docs                 # URL для получения JSON-спецификации
  swagger-ui:
    path: /swagger-ui.html           # URL для отображения Swagger UI
    operations-sorter: method         # сортировка операций по HTTP-методу
    tags-sorter: alpha                # сортировка тегов по алфавиту
    display-request-duration: true    # показывать время выполнения запроса
```

После этих настроек спецификация будет доступна по адресу `http://localhost:8080/api-docs`, а UI — по `http://localhost:8080/swagger-ui.html`.

#### **Шаг 3. Аннотируйте ваши контроллеры для улучшения документации**

Чтобы документация была информативной, используйте аннотации из пакета `io.swagger.v3.oas.annotations`. Возьмём для примера контроллер `UserController` из предыдущих заданий и добавим к нему описание.

```java
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;

@RestController
@RequestMapping("/api/users")
@Tag(name = "Пользователи", description = "Управление пользователями")
public class UserController {

    @GetMapping("/{id}")
    @Operation(summary = "Получить пользователя по ID", description = "Возвращает пользователя по его уникальному идентификатору")
    @ApiResponses(value = {
        @ApiResponse(responseCode = "200", description = "Пользователь найден",
                     content = @Content(mediaType = "application/json",
                                        schema = @Schema(implementation = User.class))),
        @ApiResponse(responseCode = "404", description = "Пользователь не найден",
                     content = @Content)
    })
    public ResponseEntity<User> getUser(
            @Parameter(description = "ID пользователя", example = "1", required = true)
            @PathVariable Long id) {
        // ... реализация
    }

    @PostMapping
    @Operation(summary = "Создать нового пользователя")
    @ApiResponses(value = {
        @ApiResponse(responseCode = "201", description = "Пользователь создан",
                     content = @Content(schema = @Schema(implementation = User.class))),
        @ApiResponse(responseCode = "400", description = "Неверные входные данные")
    })
    public ResponseEntity<User> createUser(
            @io.swagger.v3.oas.annotations.parameters.RequestBody(
                description = "Данные пользователя", required = true,
                content = @Content(schema = @Schema(implementation = User.class)))
            @Valid @RequestBody User user) {
        // ... реализация
    }
}
```

Аннотации:
- `@Tag` — группирует операции по тегам (в UI будет отдельный раздел "Пользователи").
- `@Operation` — краткое описание метода.
- `@ApiResponses` — возможные ответы сервера.
- `@Parameter` — описание параметра пути или запроса.
- `@Schema` — указывает, какой DTO используется, можно также описать поля внутри DTO (см. следующий шаг).

#### **Шаг 4. Документируйте DTO (по желанию)**

Чтобы в Swagger UI корректно отображалась структура запросов и ответов, аннотируйте поля ваших DTO. Например, класс `User`:

```java
import io.swagger.v3.oas.annotations.media.Schema;

@Data
@Schema(description = "Объект пользователя")
public class User {
    @Schema(description = "Уникальный идентификатор", example = "1", accessMode = Schema.AccessMode.READ_ONLY)
    private Long id;

    @Schema(description = "Имя пользователя", example = "Иван Петров", required = true)
    @NotBlank
    private String name;

    @Schema(description = "Email пользователя", example = "ivan@example.com", required = true)
    @Email
    private String email;
}
```

Теперь в UI будет показано, какие поля обязательны и примеры значений.

#### **Шаг 5. Запустите приложение и откройте Swagger UI**

1. Запустите Spring Boot приложение (например, через `main` метод).
2. Откройте браузер и перейдите по адресу:  
   `http://localhost:8080/swagger-ui.html` (или тот путь, который вы указали в `application.yml`).
3. Вы увидите страницу Swagger UI со списком всех доступных эндпоинтов, разделённых по тегам.

#### **Шаг 6. Попробуйте интерактивное тестирование**

В Swagger UI вы можете:
- Развернуть любой эндпоинт.
- Нажать кнопку **Try it out**.
- Заполнить параметры (если требуется) и нажать **Execute**.
- Увидеть реальный HTTP-запрос, ответ, заголовки и код статуса.

Это позволяет быстро проверить работу API прямо из браузера, не прибегая к Postman или curl.

#### **Шаг 7. (Опционально) Кастомизируйте глобальную информацию об API**

Создайте конфигурационный бин, чтобы добавить описание, версию, контактную информацию и лицензию.

```java
import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.info.License;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class OpenApiConfig {

    @Bean
    public OpenAPI customOpenAPI() {
        return new OpenAPI()
                .info(new Info()
                        .title("User Management API")
                        .version("1.0.0")
                        .description("API для управления пользователями в системе")
                        .contact(new Contact()
                                .name("Support Team")
                                .email("support@example.com")
                                .url("https://example.com"))
                        .license(new License()
                                .name("Apache 2.0")
                                .url("https://www.apache.org/licenses/LICENSE-2.0")));
    }
}
```

После добавления этой конфигурации верхняя часть Swagger UI отобразит введённую информацию.

*   **Результат:** Создание конкретной документации

### **Задание 15: "Эхо-сервер с метриками"**
*   **Цель:** Комплексное задание, объединяющее все пройденные темы.
*   **Пошаговая инструкция:**
    1.  Создаем `@PostMapping("/api/echo")`, который принимает `String message`.
    2.  В ответе возвращаем `ResponseEntity` с телом `{"echo": message}` и заголовком `X-Processed-By`, равным имени хоста.
    3.  Если сообщение пустое или `null`, выбрасываем кастомное исключение `EmptyMessageException`.
    4.  В `@ControllerAdvice` обрабатываем это исключение и возвращаем красивую ошибку со статусом 400.
    5.  Добавляем в `@ControllerAdvice` логирование времени обработки запроса.
    6.  Документируем этот эндпоинт в OpenAPI с описанием возможной ошибки 400.

```java
public class EmptyMessageException extends BusinessException {
    public EmptyMessageException() { super("Message cannot be empty"); }
}

@RestController
@RequestMapping("/api")
public class EchoController {
    @PostMapping("/echo")
    public ResponseEntity<Map<String, String>> echo(@RequestBody(required = false) String message) {
        if (message == null || message.trim().isEmpty()) {
            throw new EmptyMessageException();
        }
        Map<String, String> response = Map.of("echo", message);
        String hostname;
        try {
            hostname = InetAddress.getLocalHost().getHostName();
        } catch (Exception e) {
            hostname = "unknown";
        }
        return ResponseEntity.ok()
                .header("X-Processed-By", hostname)
                .body(response);
    }
}

// В GlobalExceptionHandler добавить:
@ExceptionHandler(EmptyMessageException.class)
public ResponseEntity<ErrorResponse> handleEmptyMessage(EmptyMessageException ex) {
    ErrorResponse error = new ErrorResponse("EMPTY_MESSAGE", ex.getMessage(), Instant.now());
    return ResponseEntity.badRequest().body(error);
}
```
*   **Результат:** Собрание всех компонентов (ResponseEntity, исключения, Advice, заголовки, документацию) в одном работающем механизме.
