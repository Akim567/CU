# Семинар 7: Основы JPA и связь сущностей
**Цель:** Перейти на уровень ORM с JPA/Hibernate, понять жизненный цикл сущностей (Transient/Managed/Detached/Removed), настроить связи между таблицами (@ManyToOne, @OneToMany), использовать репозитории Spring Data и каскадные операции.

---

## Задание 1. Переход на JPA — замена JDBC зависимостей

**Конкретная цель:** Мигрировать проект с Spring JDBC на Spring Data JPA.

**Пошаговая инструкция:**
1. Откройте `build.gradle` проекта из Семинара 1
2. Удалите зависимость `spring-boot-starter-jdbc`
3. Добавьте зависимость `spring-boot-starter-data-jpa`
4. Обновите `application.yml` с настройками JPA
5. Удалите классы `UserRepository`, `PostRepository` из Семинара 6 (будем создавать новые)
6. Запустите приложение — убедитесь, что нет ошибок

**Изменения в `build.gradle`:**
```gradle
dependencies {
    // УДАЛИТЬ:
    // implementation 'org.springframework.boot:spring-boot-starter-jdbc'
    
    // ДОБАВИТЬ:
    implementation 'org.springframework.boot:spring-boot-starter-data-jpa'
    
    // ОСТАВИТЬ:
    implementation 'org.postgresql:postgresql:...'
    implementation 'org.flywaydb:flyway-core:...'
    implementation 'org.projectlombok:lombok:...'
    annotationProcessor 'org.projectlombok:lombok:...'
}
```

**Изменения в `application.yml`:**
```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/seminar_db
    username: postgres
    password: postgres
  jpa:
    show-sql: true
    open-in-view: false
    hibernate:
      ddl-auto: validate
    properties:
      hibernate:
        format_sql: true
```

**Критерий выполнения:** Приложение запускается, в логах видны настройки Hibernate, нет ошибок подключения.

---

## Задание 2. Создание сущности `User` с аннотациями JPA 

**Конкретная цель:** Создать JPA-сущность с правильными аннотациями для маппинга на таблицу.

**Пошаговая инструкция:**
1. Создайте класс `User` в пакете `entity`
2. Добавьте аннотации `@Entity`, `@Table`
3. Настройте первичный ключ с `@Id` и `@GeneratedValue`
4. Настройте поля с `@Column` (укажите `nullable`, `length`)
5. Добавьте защищённый конструктор без аргументов (требование Hibernate)
6. Добавьте `@PrePersist` для автоматического заполнения `createdAt`

**Ожидаемый код:**
```java
@Entity
@Table(name = "users")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class User {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(name = "username", nullable = false, length = 100)
    private String username;
    
    @Column(name = "email", nullable = false, length = 100, unique = true)
    private String email;
    
    @Column(name = "created_at", updatable = false)
    private Instant createdAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
    }
}
```

**Вопросы для самопроверки:**
- Почему конструктор должен быть `protected`, а не `private`?
- Что произойдёт если убрать `@Entity`?
- Зачем нужна `@PrePersist` вместо установки даты в конструкторе?

**Критерий выполнения:** При запуске приложения нет ошибок валидации схемы, Hibernate распознаёт сущность.

---

## Задание 3. Репозиторий и CRUD операции через Spring Data 

**Конкретная цель:** Создать репозиторий и протестировать все CRUD операции с наблюдением за SQL-запросами.

**Пошаговая инструкция:**
1. Создайте интерфейс `UserRepository` в пакете `repository`
2. Расширьте `JpaRepository<User, Long>`
3. Добавьте custom query methods через Naming Convention (минимум 4 метода)
4. Добавьте метод с `@Query` для поиска по домену email
5. Создайте `CommandLineRunner` для тестирования всех операций
6. Включите `show-sql: true` и наблюдайте за запросами

**Ожидаемый код репозитория:**
```java
@Repository
public interface UserRepository extends JpaRepository<User, Long> {
    
    // Naming Convention - поиск по email
    Optional<User> findByEmail(String email);
    
    // Naming Convention - поиск по части username
    List<User> findByUsernameContaining(String username);
    
    // Naming Convention - проверка существования
    boolean existsByEmail(String email);
    
    // Naming Convention - подсчёт
    long countByCreatedAtAfter(Instant date);
    
    // @Query - JPQL запрос
    @Query("SELECT u FROM User u WHERE u.email LIKE %:domain")
    List<User> findByEmailDomain(@Param("domain") String domain);
}
```

**Ожидаемый код тестирования:**
```java
@Component
@RequiredArgsConstructor
public class JpaCrudRunner implements CommandLineRunner {
    
    private final UserRepository userRepository;
    
    @Override
    @Transactional
    public void run(String... args) {
        // CREATE
        User user = new User("alice", "alice@example.com");
        user = userRepository.save(user);
        System.out.println("Created user with id: " + user.getId());
        
        // READ (findById)
        User found = userRepository.findById(user.getId()).orElseThrow();
        System.out.println("Found user: " + found.getUsername());
        
        // READ (custom query)
        List<User> byDomain = userRepository.findByEmailDomain("example.com");
        System.out.println("Users with domain: " + byDomain.size());
        
        // UPDATE (Dirty Checking - без вызова save!)
        found.setUsername("alice_updated");
        // При коммите Hibernate сам выполнит UPDATE
        System.out.println("Username updated (Dirty Checking)");
        
        // DELETE
        userRepository.delete(found);
        System.out.println("User deleted");
        
        // COUNT
        System.out.println("Total users: " + userRepository.count());
    }
}
```

**Критерий выполнения:** Все CRUD операции работают, в логах видны SQL-запросы (INSERT, SELECT, UPDATE, DELETE), Dirty Checking срабатывает без явного `save()`.

---

## Задание 4. Связь `@ManyToOne` — Post ссылается на User

**Конкретная цель:** Создать сущность `Post` со связью на `User` и протестировать ленивую загрузку.

**Пошаговая инструкция:**
1. Создайте класс `Post` в пакете `entity`
2. Добавьте поля: `id`, `title`, `content`, `user`
3. Настройте связь `@ManyToOne(fetch = FetchType.LAZY)`
4. Добавьте `@JoinColumn(name = "user_id")` для внешнего ключа
5. Создайте `PostRepository extends JpaRepository<Post, Long>`
6. Напишите тест для демонстрации `LazyInitializationException`

**Ожидаемый код сущности:**
```java
@Entity
@Table(name = "posts")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class Post {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false, length = 200)
    private String title;
    
    @Column(columnDefinition = "TEXT")
    private String content;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;
}
```

**Ожидаемый код репозитория:**
```java
@Repository
public interface PostRepository extends JpaRepository<Post, Long> {
    List<Post> findByUserId(Long userId);
}
```

**Тест ленивой загрузки:**
```java
@Component
@RequiredArgsConstructor
public class LazyLoadingRunner implements CommandLineRunner {
    
    private final PostRepository postRepository;
    
    @Override
    @Transactional
    public void run(String... args) {
        // 1. Загружаем пост (внутри транзакции)
        Post post = postRepository.findById(1L).orElseThrow();
        System.out.println("Post loaded: " + post.getTitle());
        
        // 2. Обращаемся к пользователю (выполняется второй SQL)
        String authorName = post.getUser().getUsername();
        System.out.println("Author: " + authorName);
        // В логах: 2 SQL запроса (SELECT post, SELECT user)
    }
}

// Отдельный метод БЕЗ @Transactional
public void testOutsideTransaction() {
    Post post = postRepository.findById(1L).orElseThrow();
    // Транзакция завершена → сессия закрыта
    try {
        String name = post.getUser().getUsername();
        //  LazyInitializationException!
    } catch (LazyInitializationException e) {
        System.out.println("Expected exception: " + e.getMessage());
    }
}
```

**Критерий выполнения:** Связь работает, ленивая загрузка срабатывает внутри транзакции, `LazyInitializationException` выбрасывается вне транзакции.

---

## Задание 5. Связь `@OneToMany` — User владеет Post_ами 

**Конкретная цель:** Добавить обратную связь в `User` и настроить каскадные операции.

**Пошаговая инструкция:**
1. Добавьте поле `List<Post> posts` в сущность `User`
2. Настройте `@OneToMany(mappedBy = "user")`
3. Добавьте `cascade = {PERSIST, MERGE}` и `orphanRemoval = true`
4. Создайте helper-методы `addPost()` и `removePost()` для управления связью
5. Протестируйте каскадное сохранение

**Ожидаемый код:**
```java
@Entity
@Table(name = "users")
public class User {
    // ... другие поля из Задания 2
    
    @OneToMany(
        mappedBy = "user",
        cascade = {CascadeType.PERSIST, CascadeType.MERGE},
        orphanRemoval = true
    )
    private List<Post> posts = new ArrayList<>();
    
    // Helper методы для управления двусторонней связью
    public void addPost(Post post) {
        posts.add(post);
        post.setUser(this);
    }
    
    public void removePost(Post post) {
        posts.remove(post);
        post.setUser(null);
    }
}
```

**Тест каскада:**
```java
@Component
@RequiredArgsConstructor
public class CascadeRunner implements CommandLineRunner {
    
    private final UserRepository userRepository;
    
    @Override
    @Transactional
    public void run(String... args) {
        // 1. Создаём пользователя
        User user = new User("bob", "bob@example.com");
        
        // 2. Создаём посты и добавляем к пользователю
        Post post1 = new Post("Post 1", "Content 1");
        Post post2 = new Post("Post 2", "Content 2");
        
        user.addPost(post1);
        user.addPost(post2);
        
        // 3. Сохраняем только пользователя (посты сохранятся каскадом)
        userRepository.save(user);
        System.out.println("User and posts saved (cascade PERSIST)");
        
        // 4. Удаляем пост из коллекции (orphanRemoval удалит из БД)
        user.removePost(post1);
        userRepository.save(user);
        System.out.println("Post removed (orphanRemoval)");
    }
}
```

**Критерий выполнения:** Посты сохраняются автоматически при сохранении пользователя, удаление из коллекции приводит к DELETE в БД.

---

## Задание 6. Жизненный цикл сущностей — 4 состояния

**Конкретная цель:** Продемонстрировать все 4 состояния сущности и переходы между ними.

**Пошаговая инструкция:**
1. Создайте метод для демонстрации каждого состояния
2. Наблюдайте за SQL-запросами в логах
3. Запишите когда выполняется INSERT, UPDATE, DELETE

**Ожидаемый код:**
```java
@Component
@RequiredArgsConstructor
public class LifecycleRunner implements CommandLineRunner {
    
    private final UserRepository userRepository;
    
    @Override
    public void run(String... args) {
        demonstrateLifecycle();
    }
    
    @Transactional
    public void demonstrateLifecycle() {
        // 1. TRANSIENT (новый объект, не в БД, не отслеживается)
        User user = new User("charlie", "charlie@example.com");
        System.out.println("State: TRANSIENT (no SQL yet)");
        
        // 2. MANAGED (после save, отслеживается Hibernate)
        user = userRepository.save(user);
        System.out.println("State: MANAGED (INSERT executed)");
        // В логах: INSERT INTO users ...
        
        // 3. DIRTY CHECKING (изменения фиксируются автоматически)
        user.setUsername("charlie_updated");
        // UPDATE выполнится при commit без вызова save()
        System.out.println("Dirty Checking активен");
    }
    
    // 4. DETACHED (вне транзакции, изменения не отслеживаются)
    public void testDetached() {
        User user = userRepository.findById(1L).orElseThrow();
        // Транзакция завершена → DETACHED
        user.setUsername("detached_user");
        // Изменения НЕ сохранятся!
        userRepository.save(user); // Это сделает merge
        System.out.println("State: DETACHED (merge required)");
    }
    
    // 5. REMOVED (помечен на удаление)
    @Transactional
    public void testRemoved() {
        User user = userRepository.findById(1L).orElseThrow();
        userRepository.delete(user);
        System.out.println("State: REMOVED (DELETE at commit)");
    }
}
```

---

## Задание 7. Продвинутый уровень 

### 7.1. Каскад `REMOVE` vs `ON DELETE CASCADE` в БД
**Конкретная цель:** Сравнить два подхода к удалению связанных записей.

**Задача:**
1. Реализуйте удаление пользователя с постами через JPA `cascade = REMOVE`
2. Реализуйте через `ON DELETE CASCADE` в миграции Flyway
3. Сравните количество SQL-запросов в логах

**Ожидаемый результат:**
- JPA cascade: N+1 запросов (1 на каждый пост + 1 на пользователя)
- БД cascade: 1 запрос (СУБД удаляет связанные записи сама)

### 7.2. Аудит с `@PrePersist` и `@PreUpdate`
**Конкретная цель:** Добавить автоматическое заполнение полей аудита.

**Задача:**
1. Добавьте поля `createdAt` и `updatedAt` в сущность
2. Используйте `@PrePersist` и `@PreUpdate`
3. Протестируйте что даты заполняются автоматически

**Ожидаемый код:**
```java
@PrePersist
protected void onCreate() {
    createdAt = Instant.now();
    updatedAt = Instant.now();
}

@PreUpdate
protected void onUpdate() {
    updatedAt = Instant.now();
}
```

### 7.3. Проекции DTO через интерфейс
**Конкретная цель:** Загружать только нужные поля без полной сущности.

**Задача:**
1. Создайте интерфейс-проекцию `UserSummary`
2. Добавьте метод в репозиторий возвращающий проекцию
3. Сравните SQL с загрузкой полной сущности

**Ожидаемый код:**
```java
public interface UserSummary {
    Long getId();
    String getUsername();
    String getEmail();
}

@Repository
public interface UserRepository extends JpaRepository<User, Long> {
    UserSummary findByUsername(String username);
}
```


