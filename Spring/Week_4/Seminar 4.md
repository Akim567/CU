# Seminar 4: RESTful API. Spring MVC. Обработка запросов

## Задание 0. Создание проекта и подключение зависимостей

**Цель:** подготовить рабочее окружение: создать Spring Boot проект и добавить все необходимые зависимости.

**Задача:**
1. Создайте новый Spring Boot проект (можно использовать[start.spring.io](https://start.spring.io/)или IDE).
2. Добавьте следующие зависимости:
    - **Spring Web**(`spring-boot-starter-web`)
    - **Validation**(`spring-boot-starter-validation`)
    - **MapStruct**(и его процессор)
3. Запустите приложение и проверьте, что оно стартует без ошибок.
```xml
<dependencies>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-web</artifactId>
    </dependency>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-validation</artifactId>
    </dependency>
    <dependency>
        <groupId>org.mapstruct</groupId>
        <artifactId>mapstruct</artifactId>
        <version>1.6.3</version>
    </dependency>
</dependencies>

<build>
   <plugins>
      <plugin>
         <groupId>org.apache.maven.plugins</groupId>
         <artifactId>maven-compiler-plugin</artifactId>
         <version>3.8.1</version>
         <configuration>
            <annotationProcessorPaths>
               <path>
                  <groupId>org.mapstruct</groupId>
                  <artifactId>mapstruct-processor</artifactId>
                  <version>1.6.3</version>
               </path>
            </annotationProcessorPaths>
         </configuration>
      </plugin>
   </plugins>
</build>
```

```groovy
plugins {
   id 'java'
   id 'org.springframework.boot' version '3.5.11'  
   id 'io.spring.dependency-management' version '1.1.7'
}

group = 'com.example'
version = '0.0.1-SNAPSHOT'
sourceCompatibility = '17' 

repositories {
   mavenCentral()
}

dependencies {
   implementation group: 'org.springframework.boot', name: 'spring-boot-starter-web'
   implementation group: 'org.springframework.boot', name: 'spring-boot-starter-validation'
   
   implementation 'org.mapstruct:mapstruct:1.6.3'
   annotationProcessor 'org.mapstruct:mapstruct-processor:1.6.3'
}
```

## Задание 1. Создание базового REST-контроллера
**Цель:** научиться использовать `@RestController` и `@GetMapping`.

**Задача:** Создайте класс `PingController` с методом, который обрабатывает GET-запрос на `/api/v1/ping` и возвращает строку `"pong"`.

**Ожидаемый код:**
```java
@RestController
@RequestMapping("/api/v1")
public class PingController {
    @GetMapping("/ping")
    public String ping() {
        return "pong";
    }
}
```
*Проверка:* после запуска приложения GET-запрос на `/api/v1/ping` должен вернуть `"pong"`.

---

## Задание 2. Использование @PathVariable и @RequestParam
**Цель:** знакомство с path-variables и query-params.

**Задача:**
1. В том же контроллере добавьте метод, обрабатывающий `GET /api/v1/users/{id}`, который возвращает строку `"User id: " + id`.
2. Добавьте метод для поиска пользователей: `GET /api/v1/users/search?name=...&page=...` (параметр `page` необязательный, по умолчанию 0). Метод должен возвращать строку с переданными значениями.

**Ожидаемый код:**
```java
@GetMapping("/users/{id}")
public String getUserById(@PathVariable Long id) {
    return "User id: " + id;
}

@GetMapping("/users/search")
public String searchUsers(@RequestParam String name,
                          @RequestParam(defaultValue = "0") int page) {
    return "Searching for " + name + ", page " + page;
}
```

---

## Задание 3. Создание DTO и использование @RequestBody
**Цель:** научиться разделять входные и выходные данные через DTO, применять `@RequestBody`.

**Задача:**
1. Создайте два класса-record (или обычных POJO) в пакете `dto`:
   - `UserCreateDto` с полями `username` (String), `email` (String), `password` (String).
   - `UserResponseDto` с полями `id` (Long), `username` (String), `email` (String).
2. В контроллере `UserController` создайте POST-метод на `/api/v1/users`, который принимает `@RequestBody UserCreateDto` и возвращает `ResponseEntity<UserResponseDto>`. В теле метода сгенерируйте фиктивный `id` (например, 42L) и верните объект `UserResponseDto` со статусом `201 CREATED`.

**Ожидаемый код:**
```java
@PostMapping("/users")
public ResponseEntity<UserResponseDto> createUser(@RequestBody UserCreateDto createDto) {
    // имитация сохранения
    UserResponseDto response = new UserResponseDto(42L, createDto.getUsername(), createDto.getEmail());
    return ResponseEntity.status(HttpStatus.CREATED).body(response);
}
```

---
## Задание 4. Маппинг DTO ↔ Entity с помощью MapStruct
**Цель:** применить библиотеку MapStruct для преобразования между DTO и Entity.

**Задача:**
1. Создайте простую POJO `UserEntity` с полями `id`, `username`, `email`, `passwordHash`
2. Добавьте зависимость MapStruct
3. Создайте интерфейс-маппер `UserMapper` с методами `toResponseDto(UserEntity entity)` и `toEntity(UserCreateDto createDto)`.
4. В сервисе (или прямо в контроллере) используйте маппер для преобразования.

**Пример интерфейса:**
```java
@Mapper(componentModel = "spring")
public interface UserMapper {
    UserResponseDto toResponseDto(UserEntity entity);
    UserEntity toEntity(UserCreateDto createDto);
}
```

Прежде чем приступить к написанию маппера, убедитесь, что в ваш проект добавлены необходимые зависимости.

> **Важно:** `mapstruct-processor` необходим для генерации кода маппера во время компиляции. Без него ваш интерфейс маппера не будет иметь реализации.

### Создание сущности (Entity)
Создайте класс `UserEntity` (можно в пакете `entity`), который будет представлять данные, хранящиеся в базе (или просто бизнес-объект). Для простоты используем обычный POJO:

```java
public class UserEntity {
    private Long id;
    private String username;
    private String email;
    private String passwordHash; // допустим, храним хеш пароля

    // стандартные конструкторы, геттеры и сеттеры
}
```

### Создание интерфейса маппера
В пакете `mapper` создайте интерфейс `UserMapper` с аннотацией `@Mapper`. Укажите `componentModel = "spring"`, чтобы Spring мог внедрять сгенерированную реализацию как бин.

```java
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;

@Mapper(componentModel = "spring")
public interface UserMapper {
    
    // Преобразование Entity -> ResponseDto
    UserResponseDto toResponseDto(UserEntity entity);
    
    // Преобразование CreateDto -> Entity
    // Поле passwordHash мы не маппим напрямую, его нужно будет заполнить отдельно (хеширование)
    @Mapping(target = "id", ignore = true)           // id генерируется автоматически
    @Mapping(target = "passwordHash", ignore = true) // будем задавать вручную
    UserEntity toEntity(UserCreateDto createDto);
}
```

### Использование маппера в сервисе (или контроллере)
В сервисе (или прямо в контроллере) внедрите `UserMapper` и используйте его для преобразований.

```java
@RestController
@RequestMapping("/api/v1/users")
public class UserController {

    private final UserMapper userMapper;

    public UserController(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    @PostMapping
    public ResponseEntity<UserResponseDto> createUser(@RequestBody UserCreateDto createDto) {
        // 1. Преобразуем DTO в Entity
        UserEntity entity = userMapper.toEntity(createDto);
        // 2. Устанавливаем хеш пароля (в реальности — хешируем)
        entity.setPasswordHash("hashed_" + createDto.getPassword());
        // 3. Сохраняем (здесь пропустим сохранение, просто присвоим ID)
        entity.setId(42L);
        // 4. Преобразуем Entity в ResponseDto
        UserResponseDto response = userMapper.toResponseDto(entity);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }
}
```

**Проверка:** После компиляции проекта убедитесь, что код собирается без ошибок, и при отправке POST-запроса с валидным JSON возвращается корректный ответ. MapStruct автоматически сгенерирует реализацию вашего интерфейса.

---

## Задание 5. Потоковая загрузка большого файла
**Цель:** избежать загрузки всего файла в память, используя InputStream.

**Задача:**
Модифицируйте метод так, чтобы копирование происходило напрямую из `InputStream` в файл. Добавьте проверку размера файла до копирования (если размер превышает 10MB — вернуть ошибку 400).

**Ожидаемый фрагмент:**
```java
if (file.getSize() > 10 * 1024 * 1024) {
    return ResponseEntity.badRequest().body("File too large");
}
try (InputStream is = file.getInputStream()) {
    Files.copy(is, uploadPath, StandardCopyOption.REPLACE_EXISTING);
}
```

---

## Задание 6. Скачивание файла
**Цель:** реализовать endpoint для скачивания ранее загруженного файла.

**Задача:**
1. Создайте GET-метод `/api/v1/files/download/{filename}`.
2. Используйте `Resource` (например, `FileSystemResource`) для передачи файла.
3. Установите заголовок `Content-Disposition: attachment; filename="..."` и верните `ResponseEntity<Resource>`.

**Пример:**
```java
@GetMapping("/download/{filename}")
public ResponseEntity<Resource> downloadFile(@PathVariable String filename) throws IOException {
    Path filePath = Paths.get("uploads/").resolve(filename);
    Resource resource = new FileSystemResource(filePath);
    return ResponseEntity.ok()
            .header(HttpHeaders.CONTENT_DISPOSITION, "attachment; filename=\"" + resource.getFilename() + "\"")
            .contentType(MediaType.APPLICATION_OCTET_STREAM)
            .body(resource);
}
```

---

## Задание 7. Работа с куками
**Цель:** научиться устанавливать и читать куки.

**Задача:**
1. В контроллере создайте метод `POST /api/v1/login`, который принимает `LoginRequest` (поля `username`, `password` — для простоты без проверки) и устанавливает httpOnly-куку с именем `authToken` и значением `"example-token"` (срок жизни 1 день).
2. Создайте метод `GET /api/v1/profile`, который читает значение куки `authToken` с помощью `@CookieValue` и возвращает его в ответе (или сообщение, что кука отсутствует).

**Пример установки куки:**
```java
ResponseCookie cookie = ResponseCookie.from("authToken", "dummy-token")
        .httpOnly(true)
        .path("/")
        .maxAge(Duration.ofDays(1))
        .build();
        
return ResponseEntity.ok().header(HttpHeaders.SET_COOKIE, cookie.toString()).build();
```

**Пример чтения:**
```java
@GetMapping("/profile")
public String getProfile(@CookieValue(value = "authToken", defaultValue = "no-cookie") String token) {
    return "Token: " + token;
}
```

---

## Задание 8. Работа с сессией
**Цель:** научиться сохранять данные в HTTP-сессии.

**Задача:**
1. В контроллере создайте метод `POST /api/v1/cart/add`, который принимает `@RequestParam Long itemId` и сохраняет список идентификаторов товаров в сессии (атрибут `"cart"`). Если корзины нет — создать новую.
2. Создайте метод `GET /api/v1/cart`, который возвращает содержимое корзины из сессии.

**Пример:**
```java
@PostMapping("/cart/add")
public String addToCart(@RequestParam Long itemId, HttpSession session) {
    List<Long> cart = (List<Long>) session.getAttribute("cart");
    if (cart == null) {
        cart = new ArrayList<>();
        session.setAttribute("cart", cart);
    }
    cart.add(itemId);
    
    return "Item added";
}

@GetMapping("/cart")
public List<Long> viewCart(HttpSession session) {
    List<Long> cart = (List<Long>) session.getAttribute("cart");
    return cart != null ? cart : List.of();
}
```

---

## Задание 9. Базовая валидация DTO
**Цель:** добавить аннотации валидации в DTO и активировать её в контроллере.

**Задача:**
1. Добавьте в `UserCreateDto` следующие ограничения:
   - `username`: `@NotBlank`, `@Size(min=3, max=50)`
   - `email`: `@NotBlank`, `@Email`
   - `password`: `@NotBlank`, `@Size(min=8)`
2. В контроллере добавьте `@Valid` перед `@RequestBody`.
3. На классе контроллера добавьте `@Validated`.
4. Попробуйте отправить некорректные данные и убедитесь, что приходит статус 400.

---

## Задание 10. Группы валидации
**Цель:** применить разные правила для создания и обновления.

**Задача:**
1. Создайте интерфейсы-маркеры `OnCreate`, `OnUpdate`.
2. Измените `UserCreateDto`, добавив поле `id`:
   - Для `id`: `@Null(groups = OnCreate.class)`, `@NotNull(groups = OnUpdate.class)`.
3. Создайте метод `PUT /api/v1/users/{id}`, который принимает `@RequestBody` с группой `OnUpdate.class` (используйте `@Validated(OnUpdate.class)`).
4. В методе `POST` укажите группу `OnCreate.class`.

**Пример:**
```java
@PutMapping("/users/{id}")
public UserResponseDto updateUser(@PathVariable Long id,
                                  @RequestBody @Validated(OnUpdate.class) UserCreateDto dto) {
    // ...
}
```

---

## Задание 11. Обработка ошибок валидации
**Цель:** написать глобальный обработчик для возврата структурированных ошибок.

**Задача:**
1. Создайте класс `@RestControllerAdvice`.
2. Добавьте метод, обрабатывающий `MethodArgumentNotValidException`.
3. Сформируйте ответ в виде списка полей и сообщений об ошибках (можно использовать `Map` или специальный DTO).

**Пример:**
```java
@ExceptionHandler(MethodArgumentNotValidException.class)
public ResponseEntity<Map<String, String>> handleValidationExceptions(MethodArgumentNotValidException ex) {
    Map<String, String> errors = new HashMap<>();
    
    ex.getBindingResult().getFieldErrors().forEach(error -> 
        errors.put(error.getField(), error.getDefaultMessage()));
    
    return ResponseEntity.badRequest().body(errors);
}

@ExceptionHandler(MethodArgumentNotValidException.class)
public ResponseEntity<Map<String, String>> handleValidationExceptions(HandlerMethodValidationException ex) {
        Map<String, String> errors = new HashMap<>();
        
        ex.getBindingResult().getFieldErrors().forEach(error ->
        errors.put(error.getField(), error.getDefaultMessage()));
        
        return ResponseEntity.badRequest().body(errors);
}
```
