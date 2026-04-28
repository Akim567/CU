# Семинар 8: Оптимизация и продвинутые возможности JPA
**Цель:** Решить проблему производительности N+1 запросов, внедрить кэширование второго уровня, настроить пагинацию и сортировку, реализовать оптимистичные и пессимистичные блокировки для конкурентного доступа.

---

## Задание 1. Проблема N+1 — обнаружение и измерение 

**Конкретная цель:** Создать код, который демонстрирует проблему N+1, и подсчитать количество SQL-запросов.

**Пошаговая инструкция:**
1. Создайте 10 пользователей через `CommandLineRunner` при старте
2. Создайте по 5 постов для каждого пользователя (всего 50 постов)
3. Напишите метод который выводит все посты с именами авторов
4. Включите `show-sql: true` и `format_sql: true`
5. Посчитайте количество SQL-запросов в логах

**Ожидаемый код:**
```java
@Component
@RequiredArgsConstructor
public class DataInitializer implements CommandLineRunner {
    
    private final UserRepository userRepository;
    private final PostRepository postRepository;
    
    @Override
    @Transactional
    public void run(String... args) {
        // Создаём 10 пользователей
        for (int i = 0; i < 10; i++) {
            User user = new User("user" + i, "user" + i + "@example.com");
            userRepository.save(user);
            
            // Создаём 5 постов для каждого
            for (int j = 0; j < 5; j++) {
                Post post = new Post("Post " + i + "-" + j, "Content");
                post.setUser(user);
                postRepository.save(post);
            }
        }
        System.out.println("Created 10 users with 50 posts");
    }
}

@Component
@RequiredArgsConstructor
public class NPlusOneDemo implements CommandLineRunner {
    
    private final PostRepository postRepository;
    
    @Override
    @Transactional(readOnly = true)
    public void run(String... args) {
        // Загружаем все посты
        List<Post> posts = postRepository.findAll();
        System.out.println("Loaded " + posts.size() + " posts");
        
        // Для каждого поста загружаем автора (N+1 проблема!)
        for (Post post : posts) {
            String authorName = post.getUser().getUsername();
            System.out.println(post.getTitle() + " by " + authorName);
        }
        // Ожидаем: 1 запрос для постов + 50 запросов для пользователей = 51 запрос
    }
}
```

**Задание для студента:**
1. Запустите приложение
2. Откройте логи и найдите все SQL-запросы
3. Посчитайте количество `SELECT` для таблицы `users`

**Критерий выполнения:** Студент видит 51 запрос в логах, понимает причину проблемы N+1.

---

## Задание 2. Решение N+1 через `JOIN FETCH` в JPQL 

**Конкретная цель:** Исправить проблему N+1 через явное указание загрузки связи.

**Пошаговая инструкция:**
1. Добавьте метод в `PostRepository` с аннотацией `@Query`
2. Используйте `JOIN FETCH` для предзагрузки `user`
3. Добавьте `DISTINCT` для устранения дубликатов
4. Сравните количество запросов с Заданием 1

**Ожидаемый код:**
```java
@Repository
public interface PostRepository extends JpaRepository<Post, Long> {
    
    @Query("SELECT DISTINCT p FROM Post p JOIN FETCH p.user")
    List<Post> findAllWithUsers();
}
```

**Тестирование:**
```java
@Component
@RequiredArgsConstructor
public class JoinFetchDemo implements CommandLineRunner {
    
    private final PostRepository postRepository;
    
    @Override
    @Transactional(readOnly = true)
    public void run(String... args) {
        List<Post> posts = postRepository.findAllWithUsers();
        System.out.println("Loaded " + posts.size() + " posts");
        
        for (Post post : posts) {
            // Нет дополнительных запросов!
            String authorName = post.getUser().getUsername();
            System.out.println(post.getTitle() + " by " + authorName);
        }
        // Ожидаем: 1 запрос с LEFT JOIN
    }
}
```

**Задание для студента:**
1. Запустите приложение
2. Найдите SQL-запрос в логах
3. Убедитесь что есть `LEFT JOIN` между `posts` и `users`

**Критерий выполнения:** В логах только 1 SQL-запрос с `JOIN`, проблема N+1 решена.

---

## Задание 3. Решение N+1 через `@BatchSize` (15 мин)

**Конкретная цель:** Исправить проблему N+1 через пакетную загрузку Hibernate.

**Пошаговая инструкция:**
1. Добавьте аннотацию `@BatchSize(size = 10)` на поле `user` в сущности `Post`
2. Используйте обычный метод `findAll()` (без `@Query`)
3. Посчитайте количество запросов (должно быть ~6 вместо 51)

**Ожидаемый код:**
```java
@Entity
@Table(name = "posts")
public class Post {
    // ... другие поля
    
    @ManyToOne(fetch = FetchType.LAZY)
    @BatchSize(size = 10)  // <-- Оптимизация
    @JoinColumn(name = "user_id")
    private User user;
}
```

**Тестирование:**
```java
@Component
@RequiredArgsConstructor
public class BatchSizeDemo implements CommandLineRunner {
    
    private final PostRepository postRepository;
    
    @Override
    @Transactional(readOnly = true)
    public void run(String... args) {
        List<Post> posts = postRepository.findAll();
        
        for (Post post : posts) {
            post.getUser().getUsername();
        }
        // Ожидаем: 1 запрос для постов + 5 запросов для пользователей (50/10 = 5 пакетов)
    }
}
```

**Задание для студента:**
1. Запустите приложение
2. Посчитайте количество `SELECT` для таблицы `users`

**Критерий выполнения:** В логах ~6 запросов вместо 51, студент понимает принцип пакетной загрузки.

---

## Задание 4. Решение N+1 через `@EntityGraph` 

**Конкретная цель:** Использовать стандартный JPA механизм для описания графа загрузки.

**Пошаговая инструкция:**
1. Добавьте `@NamedEntityGraph` на сущность `Post`
2. Укажите атрибуты для загрузки (`user`)
3. Используйте `@EntityGraph` в репозитории
4. Сравните с `JOIN FETCH`

**Ожидаемый код:**
```java
@Entity
@Table(name = "posts")
@NamedEntityGraph(
    name = "Post.withUser",
    attributeNodes = @NamedAttributeNode("user")
)
public class Post {
    // ... поля
}

@Repository
public interface PostRepository extends JpaRepository<Post, Long> {
    
    @EntityGraph(value = "Post.withUser", type = EntityGraphType.LOAD)
    List<Post> findAll();
}
```

**Задание для студента:**
1. Запустите приложение
2. Сравните SQL с `JOIN FETCH`

**Критерий выполнения:** `@EntityGraph` работает, студент понимает когда использовать каждый подход.

---

## Задание 5. Пагинация `Page` vs `Slice` 

**Конкретная цель:** Реализовать пагинацию двумя способами и сравнить производительность.

**Пошаговая инструкция:**
1. Создайте метод с возвратом `Page<Post>`
2. Создайте метод с возвратом `Slice<Post>`
3. Сравните количество SQL-запросов (Page делает дополнительный COUNT)
4. Протестируйте сортировку

**Ожидаемый код:**
```java
@Repository
public interface PostRepository extends JpaRepository<Post, Long> {
    
    Page<Post> findAll(Pageable pageable);
    Slice<Post> findSliceBy(Pageable pageable);
}

@Component
@RequiredArgsConstructor
public class PaginationDemo implements CommandLineRunner {
    
    private final PostRepository postRepository;
    
    @Override
    @Transactional(readOnly = true)
    public void run(String... args) {
        // Pageable: страница 0, размер 10, сортировка по createdAt DESC + id ASC
        Pageable pageable = PageRequest.of(
            0, 10,
            Sort.by(Sort.Order.desc("createdAt"), Sort.Order.asc("id"))
        );
        
        // Page — делает SELECT + COUNT
        Page<Post> page = postRepository.findAll(pageable);
        System.out.println("Page: " + page.getTotalElements() + " total");
        
        // Slice — только SELECT (быстрее)
        Slice<Post> slice = postRepository.findSliceBy(pageable);
        System.out.println("Slice: hasNext = " + slice.hasNext());
    }
}
```

**Задание для студента:**
1. Запустите приложение
2. Найдите в логах запрос `SELECT count(*)`

**Критерий выполнения:** Студент понимает разницу между `Page` и `Slice`, знает когда использовать каждый.


---

## Задание 6. Блокировки — оптимистичная и пессимистичная 

### 6.1. Оптимистичная блокировка через `@Version`
**Конкретная цель:** Реализовать контроль версий для предотвращения конфликтов.

**Пошаговая инструкция:**
1. Добавьте поле `@Version` в сущность `Post`
2. Создайте метод обновления с обработкой `OptimisticLockException`
3. Добавьте `@Retryable` для автоматических повторных попыток

**Ожидаемый код:**
```java
@Entity
public class Post {
    // ... другие поля
    
    @Version
    private Long version;
}

@Service
public class PostService {
    
    @Retryable(
        value = OptimisticLockingFailureException.class,
        maxAttempts = 3,
        backoff = @Backoff(delay = 100, multiplier = 2)
    )
    @Transactional
    public void updatePost(Long id, String newTitle) {
        Post post = postRepository.findById(id).orElseThrow();
        post.setTitle(newTitle);
        // При конфликте будет OptimisticLockingFailureException
    }
}
```

### 6.2. Пессимистичная блокировка через `@Lock`
**Конкретная цель:** Реализовать блокировку на уровне БД.

**Пошаговая инструкция:**
1. Добавьте метод в репозиторий с `@Lock(PESSIMISTIC_WRITE)`
2. Используйте `@Query` с явным запросом
3. Протестируйте блокировку

**Ожидаемый код:**
```java
@Repository
public interface PostRepository extends JpaRepository<Post, Long> {
    
    @Lock(LockModeType.PESSIMISTIC_WRITE)
    @Query("SELECT p FROM Post p WHERE p.id = :id")
    Optional<Post> findByIdWithLock(@Param("id") Long id);
}
```

**Критерий выполнения:** Оба типа блокировок работают, студент понимает различия.

---

## Задание 7. Продвинутый уровень 

### 7.1. Кастомный репозиторий с `EntityManager`
**Конкретная цель:** Создать репозиторий со сложной логикой через `EntityManager`.

**Задача:**
1. Создайте интерфейс `CustomPostRepository`
2. Реализуйте `CustomPostRepositoryImpl` с суффиксом `Impl`
3. Используйте `EntityManager` для динамического запроса

**Ожидаемый код:**
```java
public interface CustomPostRepository {
    List<Post> findPostsByComplexCriteria(String title, Long userId);
}

@Repository
@Transactional(readOnly = true)
public class CustomPostRepositoryImpl implements CustomPostRepository {
    
    @PersistenceContext
    private EntityManager em;
    
    @Override
    public List<Post> findPostsByComplexCriteria(String title, Long userId) {
        String jpql = "SELECT p FROM Post p WHERE p.title LIKE :title";
        if (userId != null) {
            jpql += " AND p.user.id = :userId";
        }
        
        TypedQuery<Post> query = em.createQuery(jpql, Post.class);
        query.setParameter("title", "%" + title + "%");
        if (userId != null) {
            query.setParameter("userId", userId);
        }
        
        return query.getResultList();
    }
}

@Repository
public interface PostRepository extends JpaRepository<Post, Long>, CustomPostRepository {
}
```

### 7.2. QueryHints для оптимизации
**Конкретная цель:** Добавить подсказки для запросов.

**Задача:**
1. Добавьте `@QueryHints` на метод репозитория
2. Используйте `readOnly`, `fetchSize`, `timeout`

**Ожидаемый код:**
```java
@QueryHints({
    @QueryHint(name = "org.hibernate.readOnly", value = "true"),
    @QueryHint(name = "org.hibernate.fetchSize", value = "100"),
    @QueryHint(name = "org.hibernate.timeout", value = "5000")
})
List<Post> findByTitle(String title);
```

**Критерий выполнения:** Все три задания компилируются, студент понимает инструменты отладки и оптимизации.

---
