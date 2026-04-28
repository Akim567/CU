# Семинар 11. Тестирование Spring-приложений

**Цель**: научиться писать различные виды тестов для Spring-приложений, освоить Unit-тестирование сервисов, Integration-тестирование репозиториев, Slice-тесты контроллеров, настроить Testcontainers для реальной БД и создать E2E-тесты.

---

## Задание 1. Подготовка проекта для тестирования

**Конкретная цель**: создать проект Spring Boot с правильными зависимостями для тестирования (или используй старые наработки по семинарам).

### Пошаговая инструкция

1. Создай новый проект Spring Boot через [start.spring.io](https://start.spring.io/) или в IDE.
2. Выбери: Java 17+, Spring Boot 3+.x, Gradle/Maven.
3. Добавь зависимости согласно таблице.

| Зависимость                                | Назначение                                             |
|--------------------------------------------|--------------------------------------------------------|
| spring-boot-starter-web                    | Веб-слой для контроллеров                              |
| spring-boot-starter-data-jpa               | Работа с БД через JPA                                  |
| spring-boot-starter-test                   | Основной модуль тестирования (JUnit, Mockito, AssertJ) |
| postgresql                                 | JDBC-драйвер для PostgreSQL                            |
| flyway-core                                | Миграции базы данных                                   |
| lombok                                     | Генерация boilerplate-кода                             |
| testcontainers + testcontainers-postgresql | Тестирование с реальной БД в Docker                    |

4. Создай файл `application-test.yml` для тестового профиля

### Файл `build.gradle`

```groovy
dependencies {
    // Основное приложение
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.boot:spring-boot-starter-data-jpa'
    implementation 'org.postgresql:postgresql:42.6.0'
    implementation 'org.flywaydb:flyway-core:9.22.0'
    implementation 'org.projectlombok:lombok:1.18.30'
    annotationProcessor 'org.projectlombok:lombok:1.18.30'
    
    // Тестирование
    testImplementation 'org.springframework.boot:spring-boot-starter-test'
    testImplementation 'org.testcontainers:junit-jupiter:1.19.0'
    testImplementation 'org.testcontainers:postgresql:1.19.0'
    testRuntimeOnly 'org.junit.platform:junit-platform-launcher'
}
```

### Файл `application-test.yml`

```yaml
spring:
  datasource:
    # Для юнит-тестов используем H2 (быстро)
    # Для интеграционных — Testcontainers (реальная БД)
    url: jdbc:h2:mem:testdb;DB_CLOSE_DELAY=-1;DB_CLOSE_ON_EXIT=FALSE
    username: sa
    password: 
    driver-class-name: org.h2.Driver
  jpa:
    hibernate:
      ddl-auto: create-drop
    show-sql: true
    properties:
      hibernate:
        format_sql: true
  flyway:
    enabled: false # Отключаем миграции для юнит-тестов
```

### Файл `application.yml` (основной)

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/seminar8_db
    username: postgres
    password: postgres
    driver-class-name: org.postgresql.Driver
  jpa:
    hibernate:
      ddl-auto: validate
  flyway:
    enabled: true
    locations: classpath:db/migration
```

### Файл `src/main/resources/db/migration/V1__create_users_table.sql`

```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    age INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Критерий выполнения**: приложение запускается без ошибок, тестовый профиль активируется при запуске тестов.

---

## Задание 2. Создание сущности и репозитория

**Конкретная цель**: создать JPA-сущность User и репозиторий для последующего тестирования.

### Пошаговая инструкция

1. Создай класс `User` в пакете `entity`.
2. Добавь аннотации JPA (`@Entity`, `@Table`, `@Id`, `@Column` ...).
3. Создай интерфейс `UserRepository` в пакете `repository`.
4. Расширь `JpaRepository<User, Long>`.
5. Добавь методы поиска через Naming Convention.

### Ожидаемый код сущности

```java
import jakarta.persistence.*;
import lombok.*;
import java.time.Instant;

@Entity
@Table(name = "users")
@Data // Осторожно с этой аннотацией — в prod её можно использовать с осторожностью
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class User {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(name = "username", nullable = false, length = 100, unique = true)
    private String username;
    
    @Column(name = "email", nullable = false, length = 100, unique = true)
    private String email;
    
    @Column(name = "age")
    private Integer age;
    
    @Column(name = "created_at", updatable = false)
    private Instant createdAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
    }
}
```

### Ожидаемый код репозитория

```java
import com.example.seminar8.entity.User;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;
import java.util.List;

@Repository
public interface UserRepository extends JpaRepository<User, Long> {
    
    // Поиск по email (Naming Convention)
    Optional<User> findByEmail(String email);
    
    // Поиск по username (Naming Convention)
    Optional<User> findByUsername(String username);
    
    // Проверка существования
    boolean existsByEmail(String email);
    
    // Поиск по части username
    List<User> findByUsernameContaining(String keyword);
}
```

**Критерий выполнения**: приложение компилируется, Hibernate распознаёт сущность, репозиторий создаётся без ошибок.

---

## Задание 3. Создание сервиса с бизнес-логикой

**Конкретная цель**: создать сервис с бизнес-логикой для последующего Unit-тестирования.

### Пошаговая инструкция

1. Создай класс `UserService` в пакете `service`.
2. Добавь аннотацию `@Service`.
3. Внедри `UserRepository` через конструктор.
4. Реализуй методы: `createUser`, `getUserById`, `getAllUsers`, `deleteUser`.
5. Добавь проверку на дубликат email.

### Ожидаемый код сервиса

```java
import com.example.seminar8.entity.User;
import com.example.seminar8.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.util.List;
import java.util.Optional;

@Service
@RequiredArgsConstructor
public class UserService {
    
    private final UserRepository userRepository;
    
    @Transactional
    public User createUser(String username, String email, Integer age) {
        // Проверка на дубликат
        if (userRepository.existsByEmail(email)) {
            throw new RuntimeException("Email already exists: " + email);
        }
        
        final User user = User.builder()
            .username(username)
            .email(email)
            .age(age)
            .build();
        
        return userRepository.save(user);
    }
    
    @Transactional(readOnly = true)
    public Optional<User> getUserById(Long id) {
        return userRepository.findById(id);
    }
    
    @Transactional(readOnly = true)
    public List<User> getAllUsers() {
        return userRepository.findAll();
    }
    
    @Transactional
    public void deleteUser(Long id) {
        userRepository.deleteById(id);
    }
    
    @Transactional
    public User updateUserEmail(Long id, String newEmail) {
        final User user = userRepository.findById(id)
            .orElseThrow(() -> new RuntimeException("User not found"));
        
        if (userRepository.existsByEmail(newEmail)) {
            throw new RuntimeException("Email already exists: " + newEmail);
        }
        
        user.setEmail(newEmail);
        return userRepository.save(user);
    }
}
```

**Критерий выполнения**: сервис компилируется, все методы реализованы, транзакции настроены.

---

## Задание 4. Unit-тестирование сервиса (@SpringBootTest + Mockito)

**Конкретная цель**: написать Unit-тесты для сервиса с использованием моков зависимостей.

### Пошаговая инструкция

1. Создай тестовый класс `UserServiceTest` в `src/test/java`.
2. Используй аннотацию `@SpringBootTest` с `webEnvironment = NONE`.
3. Замени репозиторий на мок через `@MockitoBean`.
4. Настрой поведение мока через `Mockito.doReturn().when()`.
5. Проверь взаимодействие через `verify()`.
6. Протестируй успешный сценарий и сценарий с ошибкой.

### Ожидаемый код теста

```java
import com.example.seminar8.entity.User;
import com.example.seminar8.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockitoBean;
import org.springframework.test.context.ActiveProfiles;

import java.util.Optional;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@SpringBootTest(
    webEnvironment = SpringBootTest.WebEnvironment.NONE,
    properties = {"any.property=example_value"}
)
@ActiveProfiles("test")
class UserServiceTest {
    
    @Autowired
    private UserService userService;
    
    @MockitoBean
    private UserRepository userRepository;
    
    @BeforeEach
    void setUp() {
        // Очистка моков перед каждым тестом
        reset(userRepository);
    }
    
    @Test
    void shouldCreateUserSuccessfully() {
        // Given
        final User user = User.builder()
            .username("john_doe")
            .email("john@example.com")
            .age(25)
            .build();
        
        // Настраиваем поведение мока
        doReturn(false).when(userRepository).existsByEmail(anyString());
        doReturn(user).when(userRepository).save(any(User.class));
        
        // When
        final User savedUser = userService.createUser("john_doe", "john@example.com", 25);
        
        // Then
        assertThat(savedUser.getUsername()).isEqualTo("john_doe");
        assertThat(savedUser.getEmail()).isEqualTo("john@example.com");
        assertThat(savedUser.getAge()).isEqualTo(25);
        
        // Проверяем взаимодействие (London School)
        verify(userRepository, times(1)).existsByEmail("john@example.com");
        verify(userRepository, times(1)).save(any(User.class));
    }
    
    @Test
    void shouldThrowExceptionWhenEmailExists() {
        // Given
        doReturn(true).when(userRepository).existsByEmail("duplicate@example.com");
        
        // When & Then
        assertThatThrownBy(() -> 
            userService.createUser("john", "duplicate@example.com", 25)
        )
        .isInstanceOf(RuntimeException.class)
        .hasMessageContaining("Email already exists");
        
        // save не должен быть вызван
        verify(userRepository, never()).save(any(User.class));
    }
    
    @Test
    void shouldFindUserById() {
        // Given
        final User user = User.builder()
            .id(1L)
            .username("alice")
            .email("alice@example.com")
            .age(30)
            .build();
        
        doReturn(Optional.of(user)).when(userRepository).findById(1L);
        
        // When
        final Optional<User> found = userService.getUserById(1L);
        
        // Then
        assertThat(found).isPresent();
        assertThat(found.get().getUsername()).isEqualTo("alice");
        verify(userRepository, times(1)).findById(1L);
    }
    
    @Test
    void shouldReturnEmptyWhenUserNotFound() {
        // Given
        doReturn(Optional.empty()).when(userRepository).findById(999L);
        
        // When
        final Optional<User> found = userService.getUserById(999L);
        
        // Then
        assertThat(found).isEmpty();
    }
}
```

**Вопросы для самопроверки**
1. В чём разница между `@Mock` и `@MockitoBean`?
2. Почему мы используем `anyString()` и `any(User.class)` вместо конкретных значений?
3. Что проверяет `verify()` и когда это полезно?
4. Почему `@SpringBootTest` с `webEnvironment = NONE` быстрее, чем с `RANDOM_PORT`?

**Критерий выполнения**: все 4 теста проходят, в логах видно выполнение тестов, моки работают корректно.

---

## Задание 5. Slice-тестирование контроллера (@WebMvcTest)

**Конкретная цель**: написать тесты для контроллера без загрузки всего контекста приложения.

### Пошаговая инструкция

1. Создай REST-контроллер `UserController`.
2. Создай тестовый класс `UserControllerTest`.
3. Используй аннотацию `@WebMvcTest` для загрузки только MVC-слоя.
4. Замокай сервис через `@MockitoBean`.
5. Используй `MockMvc` для виртуальных HTTP-запросов.
6. Проверь статусы ответов и тело ответа (JSON).

### Ожидаемый код контроллера

```java
import com.example.seminar8.entity.User;
import com.example.seminar8.service.UserService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.util.List;

@RestController
@RequestMapping("/api/v1/users")
@RequiredArgsConstructor
public class UserController {
    
    private final UserService userService;
    
    @PostMapping
    public ResponseEntity<User> createUser(
            @RequestBody UserCreateRequest request) {
        User user = userService.createUser(
            request.getUsername(), 
            request.getEmail(), 
            request.getAge()
        );
        return ResponseEntity.status(201).body(user);
    }
    
    @GetMapping("/{id}")
    public ResponseEntity<User> getUserById(@PathVariable Long id) {
        return userService.getUserById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }
    
    @GetMapping
    public ResponseEntity<List<User>> getAllUsers() {
        return ResponseEntity.ok(userService.getAllUsers());
    }
    
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        userService.deleteUser(id);
        return ResponseEntity.noContent().build();
    }
}

// DTO для создания пользователя
record UserCreateRequest(String username, String email, Integer age) {}
```

### Ожидаемый код теста контроллера

```java
import com.example.seminar8.entity.User;
import com.example.seminar8.service.UserService;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockitoBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.util.List;
import java.util.Optional;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(controllers = {UserController.class})
class UserControllerTest {
    
    @Autowired
    private MockMvc mockMvc;
    
    @Autowired
    private ObjectMapper objectMapper;
    
    @MockitoBean
    private UserService userService;
    
    @Test
    void shouldCreateUserAndReturn201() throws Exception {
        // Given
        UserCreateRequest request = new UserCreateRequest("john", "john@test.com", 25);
        User savedUser = User.builder()
            .id(1L)
            .username("john")
            .email("john@test.com")
            .age(25)
            .build();
        
        doReturn(savedUser).when(userService).createUser(any(), any(), any());
        
        // When & Then
        mockMvc.perform(post("/api/v1/users")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
            .andExpect(status().isCreated()) // 201
            .andExpect(jsonPath("$.id").value(1))
            .andExpect(jsonPath("$.username").value("john"))
            .andExpect(jsonPath("$.email").value("john@test.com"));
        
        verify(userService, times(1)).createUser(any(), any(), any());
    }
    
    @Test
    void shouldReturn404WhenUserNotFound() throws Exception {
        // Given
        doReturn(Optional.empty()).when(userService).getUserById(999L);
        
        // When & Then
        mockMvc.perform(get("/api/v1/users/999"))
            .andExpect(status().isNotFound()); // 404
        
        verify(userService, times(1)).getUserById(999L);
    }
    
    @Test
    void shouldReturnAllUsers() throws Exception {
        // Given
        List<User> users = List.of(
            User.builder().id(1L).username("alice").email("a@test.com").age(20).build(),
            User.builder().id(2L).username("bob").email("b@test.com").age(25).build()
        );
        doReturn(users).when(userService).getAllUsers();
        
        // When & Then
        mockMvc.perform(get("/api/v1/users"))
            .andExpect(status().isOk()) // 200
            .andExpect(jsonPath("$.length()").value(2))
            .andExpect(jsonPath("$[0].username").value("alice"))
            .andExpect(jsonPath("$[1].username").value("bob"));
    }
    
    @Test
    void shouldDeleteUserAndReturn204() throws Exception {
        // Given
        doNothing().when(userService).deleteUser(1L);
        
        // When & Then
        mockMvc.perform(delete("/api/v1/users/1"))
            .andExpect(status().isNoContent()); // 204
        
        verify(userService, times(1)).deleteUser(1L);
    }
}
```

**Вопросы для самопроверки**
1. Что загружает `@WebMvcTest` и что НЕ загружает?
2. Зачем нужен `MockMvc` вместо реального HTTP-клиента?
3. Почему все зависимости контроллера нужно мокать вручную?
4. Как проверить тело JSON-ответа в тесте?

**Критерий выполнения**: все 4 теста контроллера проходят, HTTP-статусы проверяются корректно, JSON парсится правильно.

---

## Задание 6. Интеграционное тестирование репозитория (@DataJpaTest)

**Конкретная цель**: написать интеграционные тесты для репозитория с in-memory базой данных.

### Пошаговая инструкция

1. Создай тестовый класс `UserRepositoryTest`.
2. Используй аннотацию `@DataJpaTest` для ограниченного контекста.
3. Внедри `TestEntityManager` для удобной работы с сущностями.
4. Протестируй CRUD-операции.
5. Проверь работу custom query methods.

### Ожидаемый код теста

```java
import com.example.seminar8.entity.User;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;
import org.springframework.boot.test.autoconfigure.orm.jpa.TestEntityManager;
import org.springframework.test.context.ActiveProfiles;

import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.*;

@DataJpaTest
@ActiveProfiles("test")
class UserRepositoryTest {
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private TestEntityManager entityManager;
    
    @Test
    void shouldSaveAndFindUser() {
        // Given
        User user = User.builder()
            .username("ivan_dev")
            .email("ivan@example.com")
            .age(28)
            .build();
        
        // Сохраняем и сбрасываем в БД
        User savedUser = entityManager.persistAndFlush(user);
        
        // When
        Optional<User> foundUser = userRepository.findById(savedUser.getId());
        
        // Then
        assertThat(foundUser).isPresent();
        assertThat(foundUser.get().getId()).isEqualTo(savedUser.getId());
        assertThat(foundUser.get().getEmail()).isEqualTo("ivan@example.com");
        assertThat(foundUser.get().getAge()).isEqualTo(28);
    }
    
    @Test
    void shouldFindUserByEmail() {
        // Given
        User user = User.builder()
            .username("alice")
            .email("alice@example.com")
            .age(25)
            .build();
        entityManager.persistAndFlush(user);
        
        // When
        Optional<User> found = userRepository.findByEmail("alice@example.com");
        
        // Then
        assertThat(found).isPresent();
        assertThat(found.get().getUsername()).isEqualTo("alice");
    }
    
    @Test
    void shouldReturnEmptyWhenEmailNotFound() {
        // When
        Optional<User> found = userRepository.findByEmail("notexist@example.com");
        
        // Then
        assertThat(found).isEmpty();
    }
    
    @Test
    void shouldCheckEmailExists() {
        // Given
        User user = User.builder()
            .username("bob")
            .email("bob@example.com")
            .age(30)
            .build();
        entityManager.persistAndFlush(user);
        
        // When & Then
        assertThat(userRepository.existsByEmail("bob@example.com")).isTrue();
        assertThat(userRepository.existsByEmail("notexist@example.com")).isFalse();
    }
    
    @Test
    void shouldFindUsersByUsernameContaining() {
        // Given
        entityManager.persistAndFlush(User.builder()
            .username("john_doe").email("john1@test.com").age(20).build());
        entityManager.persistAndFlush(User.builder()
            .username("john_smith").email("john2@test.com").age(25).build());
        entityManager.persistAndFlush(User.builder()
            .username("alice").email("alice@test.com").age(30).build());
        
        // When
        List<User> found = userRepository.findByUsernameContaining("john");
        
        // Then
        assertThat(found).hasSize(2);
        assertThat(found).extracting("username")
            .containsExactlyInAnyOrder("john_doe", "john_smith");
    }
    
    @Test
    void shouldDeleteUser() {
        // Given
        User user = User.builder()
            .username("to_delete")
            .email("delete@test.com")
            .age(22)
            .build();
        User saved = entityManager.persistAndFlush(user);
        
        // When
        userRepository.deleteById(saved.getId());
        entityManager.clear(); // Очищаем кэш первого уровня
        
        // Then
        Optional<User> found = userRepository.findById(saved.getId());
        assertThat(found).isEmpty();
    }
}
```

**Вопросы для самопроверки**
1. Что делает `@DataJpaTest` и чем отличается от `@SpringBootTest`?
2. Зачем нужен `TestEntityManager` вместо обычного `EntityManager`?
3. Почему тесты с `@DataJpaTest` автоматически откатывают транзакции?
4. Почему не нужно очищать базу между тестами вручную?

**Критерий выполнения**: все 6 тестов проходят, в логах видны SQL-запросы, транзакции откатываются автоматически.

---

## Задание 7. Testcontainers — тестирование с реальной БД

**Конкретная цель**: настроить интеграционные тесты с PostgreSQL в Docker-контейнере через Testcontainers.

### Пошаговая инструкция

1. Добавь зависимости Testcontainers в `build.gradle`.
2. Создай тестовый класс `UserRepositoryIntegrationTest`.
3. Используй аннотации `@Testcontainers` и `@Container`.
4. Настрой динамические свойства через `@DynamicPropertySource`.
5. Протестируй репозиторий с реальной PostgreSQL.

### Ожидаемый код теста

```java
import com.example.seminar8.entity.User;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;
import org.springframework.boot.test.autoconfigure.orm.jpa.TestEntityManager;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.util.Optional;

import static org.assertj.core.api.Assertions.*;

@Testcontainers
@DataJpaTest
class UserRepositoryIntegrationTest {
    
    // Контейнер запускается один раз для всех тестов в классе
    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:15-alpine");
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private TestEntityManager entityManager;
    
    // Динамическая подмена свойств подключения
    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
        registry.add("spring.datasource.driver-class-name", postgres::getDriverClassName);
    }
    
    @Test
    void shouldSaveAndFindUserInRealPostgres() {
        // Given
        User user = User.builder()
            .username("postgres_test_user")
            .email("postgres@test.com")
            .age(35)
            .build();
        
        User savedUser = entityManager.persistAndFlush(user);
        
        // When
        Optional<User> found = userRepository.findById(savedUser.getId());
        
        // Then
        assertThat(found).isPresent();
        assertThat(found.get().getEmail()).isEqualTo("postgres@test.com");
    }
    
    @Test
    void shouldFindUserByEmailInRealPostgres() {
        // Given
        User user = User.builder()
            .username("email_test")
            .email("unique@email.com")
            .age(40)
            .build();
        entityManager.persistAndFlush(user);
        
        // When
        Optional<User> found = userRepository.findByEmail("unique@email.com");
        
        // Then
        assertThat(found).isPresent();
        assertThat(found.get().getUsername()).isEqualTo("email_test");
    }
}
```

**Вопросы для самопроверки**
1. Почему поле `postgres` должно быть `static`?
2. Что делает аннотация `@Testcontainers`?
3. Зачем нужен `@DynamicPropertySource`?
4. Какие преимущества у Testcontainers перед H2?

**Критерий выполнения**: тесты проходят с реальной PostgreSQL в Docker, контейнер запускается и останавливается автоматически.

---

## Задание 8. Оптимизация Testcontainers через JUnit Extension

**Конкретная цель**: создать кастомную аннотацию и Extension для запуска контейнера один раз на все тесты.

### Пошаговая инструкция

1. Создай кастомную аннотацию `@WithPostgres`.
2. Создай класс `PostgresExtension`, реализующий `BeforeAllCallback` и `CloseableResource`.
3. Реализуй singleton-контейнер с блокировкой.
4. Примени аннотацию в тестах.

### Ожидаемый код кастомной аннотации

```java
import org.junit.jupiter.api.extension.ExtendWith;
import java.lang.annotation.*;

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.TYPE)
@ExtendWith({PostgresExtension.class})
public @interface WithPostgres {
}
```

### Ожидаемый код Extension

```java
import org.junit.jupiter.api.extension.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.utility.DockerImageName;

import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

public class PostgresExtension implements BeforeAllCallback, ExtensionContext.Store.CloseableResource {
    
    private static final Logger log = LoggerFactory.getLogger(PostgresExtension.class);
    private static final Lock LOCK = new ReentrantLock();
    private static final AtomicBoolean STARTED = new AtomicBoolean(false);
    
    // Singleton-контейнер — один на все тесты
    private static final PostgreSQLContainer<?> POSTGRES =
        new PostgreSQLContainer<>(DockerImageName.parse("postgres:15-alpine"))
            .withDatabaseName("testdb")
            .withUsername("test")
            .withPassword("test");
    
    @Override
    public void beforeAll(ExtensionContext context) {
        LOCK.lock();
        try {
            // Запускаем, только если ещё не запущен
            if (!STARTED.compareAndExchange(false, true)) {
                log.info("Starting PostgreSQL Container...");
                POSTGRES.start();
                
                // Устанавливаем системные свойства для Spring
                System.setProperty("spring.datasource.url", POSTGRES.getJdbcUrl());
                System.setProperty("spring.datasource.username", POSTGRES.getUsername());
                System.setProperty("spring.datasource.password", POSTGRES.getPassword());
                System.setProperty("spring.datasource.driver-class-name", POSTGRES.getDriverClassName());
                
                // Регистрируем ресурс для закрытия
                context.getRoot().getStore(ExtensionContext.Store.GLOBAL)
                    .put("POSTGRES Container", this);
                
                log.info("PostgreSQL Container started on port: {}", POSTGRES.getFirstMappedPort());
            }
        } finally {
            LOCK.unlock();
        }
    }
    
    @Override
    public void close() {
        log.info("Stopping PostgreSQL Container...");
        POSTGRES.stop();
        STARTED.set(false);
    }
}
```

### Использование в тесте

```java
@WithPostgres
@DataJpaTest
class UserRepositoryOptimizedTest {
    
    @Autowired
    private UserRepository userRepository;
    
    @Test
    void shouldWorkWithSharedContainer() {
        // Тест использует тот же контейнер, что и другие тесты
        User user = User.builder()
            .username("optimized_test")
            .email("opt@test.com")
            .age(33)
            .build();
        userRepository.save(user);
        
        assertThat(userRepository.findByEmail("opt@test.com")).isPresent();
    }
}
```

**Вопросы для самопроверки**
1. Почему используется `AtomicBoolean` и `Lock`?
2. Что даёт реализация `CloseableResource`?
3. Почему контейнер должен быть `static`?
4. Какие преимущества у этого подхода в CI/CD?

**Критерий выполнения**: Extension работает, контейнер запускается один раз на все тесты, закрывается после всех тестов.

---

## Задание 9. E2E-тестирование (End-to-End)

**Конкретная цель**: написать сквозной тест, проверяющий всю систему от HTTP-запроса до базы данных.

### Пошаговая инструкция

1. Создай кастомную аннотацию `@E2ETest`.
2. Создай тестовый класс `UserControllerE2ETest`.
3. Используй `@SpringBootTest` с `RANDOM_PORT`.
4. Протестируй полный сценарий: HTTP-запрос → сервис → БД → проверка в БД.

### Ожидаемый код кастомной аннотации

```java
import org.springframework.boot.test.autoconfigure.orm.jpa.AutoConfigureTestEntityManager;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import java.lang.annotation.*;

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.TYPE)
@AutoConfigureTestEntityManager
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@ActiveProfiles({"test"})
@WithPostgres
public @interface E2ETest {
    @AliasFor(annotation = SpringBootTest.class, attribute = "properties")
    String[] properties() default {};
}
```

### Ожидаемый код E2E-теста

```java
import com.example.seminar8.config.E2ETest;
import com.example.seminar8.entity.User;
import com.example.seminar8.repository.UserRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.util.Optional;

import static org.assertj.core.api.Assertions.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@E2ETest
@AutoConfigureMockMvc
class UserControllerE2ETest {
    
    @Autowired
    private MockMvc mockMvc;
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private ObjectMapper objectMapper;
    
    @Test
    void shouldCreateUserAndPersistInDatabase() throws Exception {
        // Given
        UserCreateRequest inputUser = new UserCreateRequest(
            "e2e_test_user", 
            "e2e@test.com", 
            27
        );
        
        // When + Then (цепочка проверок)
        final String responseJson = mockMvc.perform(post("/api/v1/users")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(inputUser)))
            .andExpect(status().isCreated()) // Статус 201
            .andExpect(jsonPath("$.id").exists()) // ID сгенерирован
            .andExpect(jsonPath("$.username").value("e2e_test_user"))
            .andExpect(jsonPath("$.email").value("e2e@test.com"))
            .andExpect(jsonPath("$.age").value(27))
            .andReturn()
            .getResponse()
            .getContentAsString();
        
        User createdUser = objectMapper.readValue(responseJson, User.class);
        
        // Проверка в БД (интеграционная часть)
        Optional<User> userFromDb = userRepository.findById(createdUser.getId());
        assertThat(userFromDb)
            .as("Пользователь должен физически существовать в базе данных PostgreSQL")
            .isPresent();
        assertThat(userFromDb.get().getEmail()).isEqualTo("e2e@test.com");
        assertThat(userFromDb.get().getId()).isGreaterThan(0L);
    }
    
    @Test
    void shouldGetUserById() throws Exception {
        // Given — сначала создаём пользователя
        User user = userRepository.save(User.builder()
            .username("get_test")
            .email("get@test.com")
            .age(29)
            .build());
        
        // When + Then
        mockMvc.perform(get("/api/v1/users/" + user.getId()))
            .andExpect(status().isOk())
            .andExpect(jsonPath("$.username").value("get_test"))
            .andExpect(jsonPath("$.email").value("get@test.com"));
    }
    
    @Test
    void shouldReturn404ForNonExistentUser() throws Exception {
        // When + Then
        mockMvc.perform(get("/api/v1/users/99999"))
            .andExpect(status().isNotFound());
    }
}
```

**Вопросы для самопроверки**
1. Чем E2E-тест отличается от интеграционного?
2. Почему E2E-тесты медленнее других видов тестов?
3. Зачем проверять данные в БД после HTTP-запроса?
4. Почему используется `RANDOM_PORT` вместо `DEFINED_PORT`?

**Критерий выполнения**: E2E-тесты проходят, полный сценарий работает от HTTP до БД, данные действительно сохраняются в PostgreSQL.

---

## Задание 10. Продвинутый уровень — обработка ошибок и валидация

**Конкретная цель**: настроить глобальную обработку исключений и тестировать сценарии ошибок.

### Пошаговая инструкция

1. Создай класс `GlobalExceptionHandler` с `@RestControllerAdvice`.
2. Добавь обработчики для `DataAccessException` и `RuntimeException`.
3. Создай тесты для проверки обработки ошибок.
4. Проверь HTTP-статусы и тело ответа при ошибках.

### Ожидаемый код обработчика

```java
import org.springframework.dao.DataAccessException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

@RestControllerAdvice
public class GlobalExceptionHandler {
    
    @ExceptionHandler(DataAccessException.class)
    public ResponseEntity<Map<String, Object>> handleDbException(DataAccessException ex) {
        Map<String, Object> body = new HashMap<>();
        body.put("timestamp", Instant.now());
        body.put("status", HttpStatus.INTERNAL_SERVER_ERROR.value());
        body.put("error", "Database Error");
        body.put("message", ex.getMostSpecificCause().getMessage());
        
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(body);
    }
    
    @ExceptionHandler(RuntimeException.class)
    public ResponseEntity<Map<String, Object>> handleRuntimeException(RuntimeException ex) {
        Map<String, Object> body = new HashMap<>();
        body.put("timestamp", Instant.now());
        body.put("status", HttpStatus.BAD_REQUEST.value());
        body.put("error", "Business Logic Error");
        body.put("message", ex.getMessage());
        
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(body);
    }
    
    @ExceptionHandler(Exception.class)
    public ResponseEntity<Map<String, Object>> handleGenericException(Exception ex) {
        Map<String, Object> body = new HashMap<>();
        body.put("timestamp", Instant.now());
        body.put("status", HttpStatus.INTERNAL_SERVER_ERROR.value());
        body.put("error", "Internal Server Error");
        body.put("message", ex.getMessage());
        
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(body);
    }
}
```

### Тест обработки ошибок

```java
@WebMvcTest(controllers = {UserController.class})
class UserControllerErrorTest {
    
    @Autowired
    private MockMvc mockMvc;
    
    @Autowired
    private ObjectMapper objectMapper;
    
    @MockitoBean
    private UserService userService;
    
    @Test
    void shouldReturn400WhenEmailExists() throws Exception {
        // Given
        UserCreateRequest request = new UserCreateRequest("dup", "dup@test.com", 25);
        doThrow(new RuntimeException("Email already exists: dup@test.com"))
            .when(userService).createUser(any(), any(), any());
        
        // When & Then
        mockMvc.perform(post("/api/v1/users")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
            .andExpect(status().isBadRequest()) // 400
            .andExpect(jsonPath("$.error").value("Business Logic Error"))
            .andExpect(jsonPath("$.message").value("Email already exists: dup@test.com"));
    }
    
    @Test
    void shouldReturn500WhenDatabaseError() throws Exception {
        // Given
        doThrow(new DataAccessException("DB connection failed") {})
            .when(userService).getAllUsers();
        
        // When & Then
        mockMvc.perform(get("/api/v1/users"))
            .andExpect(status().isInternalServerError()) // 500
            .andExpect(jsonPath("$.error").value("Database Error"));
    }
}
```

**Критерий выполнения**: обработчик исключений работает, тесты проверяют корректные HTTP-статусы и тело ответа при ошибках.

---

## Дополнительные вопросы для закрепления

1. **Пирамида тестирования.** Почему Unit-тестов должно быть больше, чем E2E?
2. **Школы тестирования.** В чём разница между лондонской и детройтской школой? Какую используешь ты?
3. **Mockito.** Когда лучше использовать `verify()`, а когда — проверять состояние объекта?
4. **Testcontainers.** Какие проблемы решает Testcontainers по сравнению с H2?
5. **Контекст Spring.** Почему кеширование контекста может быть опасным в тестах?
6. **Транзакции.** Почему тесты с `@DataJpaTest` автоматически откатывают изменения?
7. **Производительность.** Как ускорить запуск тестов в CI-/CD-пайплайне?
8. **Best Practices.** Какие аннотации тестирования ты будешь использовать в своём проекте и почему?

---

## Рекомендации для самостоятельной работы

1. **Добавь тесты для сервиса Post** (если есть связь User-Post).
2. **Настрой параллельный запуск тестов** через JUnit 5.
3. **Добавь покрытие кода тестами** через JaCoCo и настрой минимальный порог.
4. **Создай тестовые данные** через `@Sql` или паттерн Test Data Builder.
5. **Настрой запуск тестов в CI/CD** (GitHub Actions, GitLab CI).
6. **Изучи архетипы тестов:** Given-When-Then, AAA (Arrange-Act-Assert).
7. **Попробуй ArchUnit** для тестирования архитектуры проекта.
