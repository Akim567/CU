# Семинар 6: Работа с базами данных на уровне JDBC и Spring JDBC
**Цель:** Научиться работать с базами данных на низком уровне (JDBC), понять проблемы ручного управления ресурсами, 
освоить Spring JDBC для устранения шаблонного кода, настроить миграции через Flyway и управлять транзакциями декларативно.

---

## Задание 1. Подготовка проекта 

**Конкретная цель:** Создать Spring Boot проект с правильными зависимостями для работы с JDBC.

**Пошаговая инструкция:**
1. Создайте новый или используйте предыдущий Spring Boot проект через [start.spring.io](https://start.spring.io/) или в IDE
2. Выберите: Java 17+, Spring Boot 3.x, Gradle/Maven
3. Добавьте зависимости согласно таблице:

| Зависимость | Назначение |
|-------------|------------|
| `spring-boot-starter-jdbc` | Основной модуль Spring JDBC |
| `postgresql` | JDBC-драйвер для PostgreSQL |
| `flyway-core` | Миграции базы данных |
| `lombok` | Генерация boilerplate-кода |

**Файл `build.gradle`:**
```gradle
dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-jdbc'
    implementation 'org.postgresql:postgresql:...'
    implementation 'org.flywaydb:flyway-core:...'
    implementation 'org.projectlombok:lombok:...'
    annotationProcessor 'org.projectlombok:lombok:...'
}
```

**Файл `application.yml`:**
```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/seminar_db
    username: postgres
    password: postgres
    driver-class-name: org.postgresql.Driver
  flyway:
    enabled: true
    locations: classpath:db/migration
```

**Критерий выполнения:** Приложение запускается без ошибок, в логах нет исключений подключения.

---

## Задание 2. Используй «голый» JDBC  

**Конкретная цель:** Написать код подключения и вставки без Spring, чтобы понять объём шаблонного кода и риски утечек ресурсов.

**Пошаговая инструкция:**
1. Создайте класс `RawJdbcExample` в пакете `example`
2. Добавьте метод `main` с выбрасыванием `SQLException`
3. Реализуйте подключение через `DriverManager`
4. Выполните INSERT в таблицу `users`
5. Выполните SELECT и выведите результаты
6. Закройте все ресурсы в блоке `finally` или `try-with-resources`

**Ожидаемый код:**
```java
public class RawJdbcExample {
    public static void main(String[] args) throws SQLException {
        Connection conn = null;
        PreparedStatement pstmt = null;
        ResultSet rs = null;
        
        try {
            // 1. Подключение
            conn = DriverManager.getConnection(
                "jdbc:postgresql://localhost:5432/seminar_db",
                "postgres", "postgres"
            );
            
            // 2. Вставка данных
            pstmt = conn.prepareStatement(
                "INSERT INTO users (username, email) VALUES (?, ?)"
            );
            pstmt.setString(1, "test_user");
            pstmt.setString(2, "test@example.com");
            int rowsInserted = pstmt.executeUpdate();
            System.out.println("Вставлено строк: " + rowsInserted);
            
            // 3. Чтение данных
            pstmt = conn.prepareStatement("SELECT * FROM users");
            rs = pstmt.executeQuery();
            
            while (rs.next()) {
                System.out.printf(
                    "id=%d, username=%s, email=%s%n",
                    rs.getLong("id"),
                    rs.getString("username"),
                    rs.getString("email")
                );
            }
        } finally {
            // 4. Закрытие ресурсов в обратном порядке
            if (rs != null) rs.close();
            if (pstmt != null) pstmt.close();
            if (conn != null) conn.close();
        }
    }
}
```

**Вопросы для самопроверки:**
- Сколько строк кода потребовалось для одной операции INSERT+SELECT?
- Что будет, если забыть закрыть `ResultSet`?
- Где здесь защита от SQL-инъекций?
- Почему `PreparedStatement` лучше `Statement`?

**Критерий выполнения:** Код компилируется, данные вставляются и читаются, ресурсы закрываются корректно.

---

## Задание 3. Переход на `JdbcTemplate`

**Конкретная цель:** Рефакторинг кода с использованием Spring JDBC для устранения шаблонного кода.

**Пошаговая инструкция:**
1. Создайте класс `User` в пакете `model` с полями: `id`, `username`, `email`, `createdAt`
2. Создайте класс `UserRepository` с аннотацией `@Repository`
3. Внедрите `JdbcTemplate` через конструктор (используйте Lombok `@RequiredArgsConstructor`)
4. Реализуйте 4 метода: `save()`, `findById()`, `findAll()`, `deleteById()`
5. Используйте `RowMapper` для маппинга `ResultSet` в объект `User`
6. Создайте `CommandLineRunner` для тестирования всех методов

**Ожидаемый код репозитория:**
```java
@Repository
@RequiredArgsConstructor
public class UserRepository {
    
    private final JdbcTemplate jdbcTemplate;
    
    private final RowMapper<User> rowMapper = (rs, rowNum) -> new User(
        rs.getLong("id"),
        rs.getString("username"),
        rs.getString("email"),
        rs.getTimestamp("created_at").toLocalDateTime()
    );
    
    // CREATE
    public User save(String username, String email) {
        jdbcTemplate.update(
            "INSERT INTO users (username, email) VALUES (?, ?)",
            username, email
        );
        return findByUsername(username);
    }
    
    // READ (один объект)
    public User findById(Long id) {
        return jdbcTemplate.queryForObject(
            "SELECT * FROM users WHERE id = ?",
            rowMapper, id
        );
    }
    
    // READ (список)
    public List<User> findAll() {
        return jdbcTemplate.query("SELECT * FROM users", rowMapper);
    }
    
    // DELETE
    public int deleteById(Long id) {
        return jdbcTemplate.update(
            "DELETE FROM users WHERE id = ?", id
        );
    }
    
    // Helper метод
    private User findByUsername(String username) {
        return jdbcTemplate.queryForObject(
            "SELECT * FROM users WHERE username = ?",
            rowMapper, username
        );
    }
}
```

**Ожидаемый код для локального тестирования:**
```java
@Component
@RequiredArgsConstructor
public class JdbcTestRunner implements CommandLineRunner {
    
    private final UserRepository userRepository;
    
    @Override
    public void run(String... args) {
        // 1. Создаём пользователя
        User user = userRepository.save("john", "john@example.com");
        System.out.println("Created user with id: " + user.getId());
        
        // 2. Находим по ID
        User found = userRepository.findById(user.getId());
        System.out.println("Found: " + found.getUsername());
        
        // 3. Получаем всех
        List<User> all = userRepository.findAll();
        System.out.println("Total users: " + all.size());
        
        // 4. Удаляем
        int deleted = userRepository.deleteById(user.getId());
        System.out.println("Deleted rows: " + deleted);
    }
}
```

**Критерий выполнения:** Все 4 CRUD-операции работают, в консоли видны результаты, нет утечек ресурсов.

---

## Задание 4. `NamedParameterJdbcTemplate` для сложных запросов 

**Конкретная цель:** Использовать именованные параметры для запросов с большим числом параметров и оператором `IN`.

**Пошаговая инструкция:**
1. Добавьте в `UserRepository` поле `NamedParameterJdbcTemplate`
2. Создайте метод `findByEmails(List<String> emails)` для поиска по списку email
3. Создайте метод `updateUserEmail(Long id, String newEmail)` с именованными параметрами
4. Протестируйте методы в `CommandLineRunner`

**Ожидаемый код:**
```java
@Repository
@RequiredArgsConstructor
public class UserRepository {
    
    private final JdbcTemplate jdbcTemplate;
    private final NamedParameterJdbcTemplate namedJdbcTemplate;
    
    // ... rowMapper и другие методы из Задания 3
    
    // Поиск по списку email (оператор IN)
    public List<User> findByEmails(List<String> emails) {
        MapSqlParameterSource params = new MapSqlParameterSource();
        params.addValue("emails", emails);
        
        String sql = """
            SELECT * FROM users 
            WHERE email IN (:emails)
        """;
        
        return namedJdbcTemplate.query(sql, params, rowMapper);
    }
    
    // Обновление с именованными параметрами
    public void updateUserEmail(Long id, String newEmail) {
        Map<String, Object> params = Map.of(
            "id", id,
            "email", newEmail
        );
        
        namedJdbcTemplate.update(
            "UPDATE users SET email = :email WHERE id = :id",
            params
        );
    }
}
```

**Тестирование:**
```java
@Override
public void run(String... args) {
    // Создаём нескольких пользователей
    userRepository.save("alice", "alice@example.com");
    userRepository.save("bob", "bob@example.com");
    userRepository.save("charlie", "charlie@example.com");
    
    // Ищем по списку email
    List<User> found = userRepository.findByEmails(
        List.of("alice@example.com", "bob@example.com")
    );
    System.out.println("Found " + found.size() + " users");
    
    // Обновляем email
    User alice = userRepository.findByUsername("alice");
    userRepository.updateUserEmail(alice.getId(), "alice_new@example.com");
}
```

**Критерий выполнения:** Запрос с `IN` работает корректно, именованные параметры читаются легче чем `?`.

---

## Задание 5. Миграции Flyway — версионирование схемы

**Конкретная цель:** Настроить автоматическое применение миграций при старте приложения.

**Пошаговая инструкция:**
1. Создайте папку `src/main/resources/db/migration`
2. Создайте файл `V1__create_users_table.sql` с созданием таблицы `users`
3. Создайте файл `V2__create_posts_table.sql` с созданием таблицы `posts`
4. Создайте файл `V3__add_index_on_email.sql` с добавлением индекса
5. Удалите таблицы из БД вручную (через pgAdmin или psql)
6. Запустите приложение — таблицы должны создаться автоматически
7. Проверьте таблицу `flyway_schema_history`

**Файл `V1__create_users_table.sql`:**
```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Файл `V2__create_posts_table.sql`:**
```sql
CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Файл `V3__add_index_on_email.sql`:**
```sql
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

**Проверка:**
```sql
-- В pgAdmin или psql выполните:
SELECT * FROM flyway_schema_history;
-- Должны быть 3 записи с версиями 1, 2, 3
```

**Критерий выполнения:** При старте приложения в логах видны сообщения Flyway о применённых миграциях, таблица `flyway_schema_history` содержит 3 записи.

---

## Задание 6. Транзакции `@Transactional` — атомарность операций 

**Конкретная цель:** Реализовать атомарную операцию с откатом при ошибке.

**Пошаговая инструкция:**
1. Создайте класс `UserService` с аннотацией `@Service`
2. Внедрите `UserRepository` и создайте `PostRepository` (аналогично `UserRepository`)
3. Создайте метод `registerUserWithPost()` с аннотацией `@Transactional`
4. Внутри метода: создайте пользователя, создайте пост, выбросьте исключение если email содержит "error"
5. Протестируйте с обычным email и с email содержащим "error"
6. Проверьте, что при ошибке ни пользователь, ни пост не сохранились

**Ожидаемый код:**
```java
@Service
@RequiredArgsConstructor
public class UserService {
    
    private final UserRepository userRepository;
    private final PostRepository postRepository;
    
    @Transactional
    public void registerUserWithPost(String username, String email, String postTitle) {
        // 1. Создаём пользователя
        User user = userRepository.save(username, email);
        System.out.println("User created: " + user.getId());
        
        // 2. Имитация ошибки
        if (email.contains("error")) {
            throw new IllegalArgumentException("Invalid email");
        }
        
        // 3. Создаём пост
        postRepository.create(postTitle, user.getId());
        System.out.println("Post created for user: " + user.getId());
    }
}
```

**Тестирование:**
```java
@Component
@RequiredArgsConstructor
public class TransactionTestRunner implements CommandLineRunner {
    
    private final UserService userService;
    private final UserRepository userRepository;
    
    @Override
    public void run(String... args) {
        try {
            // Этот вызов должен откатиться
            userService.registerUserWithPost("test", "test@error.com", "Test Post");
        } catch (IllegalArgumentException e) {
            System.out.println("Error caught: " + e.getMessage());
        }
        
        // Проверяем, что пользователь не создался
        int count = userRepository.findAll().size();
        System.out.println("Total users after failed transaction: " + count);
        // Ожидаем: 0 (или прежнее количество)
    }
}
```

**Критерий выполнения:** При ошибке транзакция откатывается, данные не сохраняются, в логах виден `ROLLBACK`.

---

## Задание 7. Продвинутый уровень 

### 7.1. Кастомный `ResultSetExtractor` для сложного маппинга
**Конкретная цель:** Реализовать маппинг результата JOIN-запроса с группировкой.

**Задача:**
1. Напишите запрос с JOIN `users` и `posts`
2. Используйте `ResultSetExtractor` вместо `RowMapper`
3. Сгруппируйте результат в `Map<User, List<Post>>`

**Ожидаемый код:**
```java
public Map<User, List<Post>> findAllUsersWithPosts() {
    String sql = """
        SELECT u.id as u_id, u.username, u.email,
               p.id as p_id, p.title, p.content
        FROM users u
        LEFT JOIN posts p ON u.id = p.user_id
    """;
    
    return jdbcTemplate.query(sql, rs -> {
        Map<User, List<Post>> result = new HashMap<>();
        
        while (rs.next()) {
            User user = new User(
                rs.getLong("u_id"),
                rs.getString("username"),
                rs.getString("email"),
                null
            );
            
            Post post = null;
            if (rs.getObject("p_id") != null) {
                post = new Post(
                    rs.getLong("p_id"),
                    rs.getString("title"),
                    rs.getString("content")
                );
            }
            
            result.computeIfAbsent(user, k -> new ArrayList<>());
            if (post != null) {
                result.get(user).add(post);
            }
        }
        
        return result;
    });
}
```

### 7.2. Глобальная обработка исключений БД
**Конкретная цель:** Создать `@RestControllerAdvice` для обработки `DataAccessException`.

**Задача:**
1. Создайте класс `DatabaseExceptionHandler` с `@RestControllerAdvice`
2. Добавьте метод с `@ExceptionHandler(DataAccessException.class)`
3. Возвращайте понятное сообщение вместо 500 ошибки

**Ожидаемый код:**
```java
@RestControllerAdvice
public class DatabaseExceptionHandler {
    
    @ExceptionHandler(DataAccessException.class)
    public ResponseEntity<String> handleDbException(DataAccessException ex) {
        return ResponseEntity
            .status(HttpStatus.INTERNAL_SERVER_ERROR)
            .body("Database error: " + ex.getMostSpecificCause().getMessage());
    }
}
```

### 7.3. Настройка пула соединений HikariCP
**Конкретная цель:** Оптимизировать параметры пула соединений.

**Задача:**
1. Добавьте в `application.yml` секцию `spring.datasource.hikari`
2. Настройте `maximum-pool-size`, `minimum-idle`, `connection-timeout`

**Ожидаемая конфигурация:**
```yaml
spring:
  datasource:
    hikari:
      maximum-pool-size: 10
      minimum-idle: 5
      connection-timeout: 20000
      idle-timeout: 300000
      max-lifetime: 600000
```

**Критерий выполнения:** Все три задания компилируются и работают, в логах видна настройка HikariCP.

---

## Чек-лист завершения Семинара 1
- [ ] Проект создан с правильными зависимостями
- [ ] «Голый» JDBC код написан и понятны его недостатки
- [ ] `JdbcTemplate` используется для CRUD операций
- [ ] `NamedParameterJdbcTemplate` используется для сложных запросов
- [ ] Flyway миграции применяются при старте
- [ ] `@Transactional` обеспечивает атомарность
- [ ] [Опционально] Продвинутые задания выполнены

---
