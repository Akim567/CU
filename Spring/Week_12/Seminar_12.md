# Семинар 12. Практикум по Spring WebFlux и Project Reactor

**Цель**: закрепить теоретические знания лекции через практическое написание реактивных цепочек, управление потоками, обработку отказов и тестирование в стеке Spring WebFlux.

---

## Подготовка окружения (единожды)

Создай проект Spring Boot с зависимостями:

```xml
<!-- pom.xml -->
<dependencies>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-webflux</artifactId>
    </dependency>
    <dependency>
        <groupId>io.projectreactor</groupId>
        <artifactId>reactor-test</artifactId>
        <scope>test</scope>
    </dependency>
    <dependency>
        <groupId>io.projectreactor.tools</groupId>
        <artifactId>blockhound</artifactId>
        <version>...</version>
        <scope>test</scope>
    </dependency>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-test</artifactId>
        <scope>test</scope>
    </dependency>
</dependencies>
```

---

## Задание 1. Основы Project Reactor: ленивость, map, filter, log

### Описание
Напиши `main`-метод, создающий `Flux` из чисел 1–10. Отфильтруй чётные, умножь на 2, добавь логирование и подписку. Продемонстрируй, что без `subscribe()` код не выполняется.

### Критерии выполнения
1. Использованы `Flux.range`, `filter`, `map`, `log`, `doOnSubscribe`, `subscribe`.
2. В консоли виден порядок вызова операторов.
3. Добавлен комментарий, доказывающий ленивое выполнение.

### Решение
```java
import reactor.core.publisher.Flux;

public class Task1Basics {
    public static void main(String[] args) {
        System.out.println("=== Создание цепочки (ещё не выполняется) ===");
        
        Flux<Integer> chain = Flux.range(1, 10)
            .doOnSubscribe(s -> System.out.println("Подписка установлена"))
            .log("Reactor-Log")
            .filter(i -> {
                System.out.println("Filter: проверяю " + i);
                return i % 2 == 0;
            })
            .map(i -> {
                System.out.println("Map: умножаю " + i);
                return i * 2;
            })
            .doOnNext(res -> System.out.println("Результат: " + res))
            .doOnComplete(() -> System.out.println("Поток завершён"));

        System.out.println("=== Цепочка создана. Вызов subscribe() запускает выполнение ===");
        chain.subscribe(); // Только здесь начинается выполнение
    }
}
```

### Ожидаемый вывод
```
=== Создание цепочки (ещё не выполняется) ===
=== Цепочка создана. Вызов subscribe() запускает выполнение ===
Подписка установлена
Reactor-Log | onSubscribe([Fuseable] FluxPeekFuseable...)
Reactor-Log | request(unbounded)
Reactor-Log | onNext(1)
Filter: проверяю 1
Reactor-Log | onNext(2)
Filter: проверяю 2
Map: умножаю 2
Результат: 4
... (продолжение для 3..10)
Reactor-Log | onComplete()
Поток завершён
```

### Теоретическая справка
- **Ленивое выполнение (Lazy Evaluation).** Операторы не создают потоки данных до `subscribe()`. Reactor анализирует цепочку и применяет **Fusion** (оптимизацию), объединяя синхронные шаги.
- **Backpressure.** По умолчанию `request(unbounded)` означает, что подписчик готов принять всё. В реальных системах количество ограничивается явно.

---

## Задание 2. Асинхронность и Schedulers: subscribeOn vs publishOn

### Описание
Сымитируй загрузку пользователей из «базы данных» (задержка `Thread.sleep(200)`). Обработай список ID параллельно, не блокируя Event Loop. Продемонстрируйте переключение потоков.

### Критерии выполнения
1. Использован `Mono.fromCallable` с эмуляцией I/O.
2. Применён `Schedulers.boundedElastic()` для изоляции блокировки.
3. В логах видны разные имена потоков до и после переключения.
4. Объяснено, почему `parallel()` здесь недопустим.

### Решение
```java
import reactor.core.publisher.Flux;
import reactor.core.scheduler.Schedulers;

import java.util.List;

public class Task2Schedulers {
    public static void main(String[] args) {
        List<Long> ids = List.of(10L, 20L, 30L, 40L);

        Flux<String> users = Flux.fromIterable(ids)
            .log("1. Source")
            .flatMap(id -> 
                Mono.fromCallable(() -> simulateDbCall(id)) // Эмуляция блокирующего I/O
                    .subscribeOn(Schedulers.boundedElastic()) // Переключаем I/O в worker-пул
                    .map(user -> "User-" + user + " [" + Thread.currentThread().getName() + "]")
            )
            .publishOn(Schedulers.parallel()) // Дальнейшая обработка в CPU-пуле
            .map(s -> s + " (обработано)");

        users.subscribe(System.out::println);
        
        // Блокируем main, чтобы дождаться асинхронного завершения
        try { Thread.sleep(2000); } catch (InterruptedException e) {}
    }

    private static String simulateDbCall(Long id) {
        try { Thread.sleep(200); } catch (InterruptedException ignored) {}
        return id.toString();
    }
}
```

### Ожидаемый вывод
```
1. Source | onSubscribe(...)
1. Source | request(unbounded)
1. Source | onNext(10)
1. Source | onNext(20)
...
User-10 [boundedElastic-1] (обработано)
User-30 [boundedElastic-3] (обработано)
User-20 [boundedElastic-2] (обработано)
User-40 [boundedElastic-4] (обработано)
```
*(Порядок может отличаться из-за параллельного выполнения.)*

### Теоретическая справка
- `subscribeOn` влияет на **источник** и выполняется один раз на цепочку.
- `publishOn` переключает контекст для **всех нижележащих** операторов.
- `boundedElastic()` предназначен для I/O. Начиная с Reactor 3.6 + Java 21, автоматически использует **виртуальные потоки**, снижая overhead.
- Важно: `Thread.sleep` в реальных WebFlux-приложениях запрещён. Вместо него используй `WebClient`, `R2DBC` или `Mono.delay()`.

---

## Задание 3. Обработка ошибок и Retry с экспоненциальной задержкой

### Описание
Создай `Mono`, имитирующий нестабильный внешний API (падает 2 раза, затем возвращает успех). Реализуйте механизм повторных попыток с экспоненциальной задержкой (`backoff`) и случайным разбросом (`jitter`). Добавь fallback на случай исчерпания попыток.

### Критерии выполнения
1. Использован `Retry.backoff()` с `filter` и `jitter`.
2. Реализован `onErrorResume` как fallback.
3. В логах видны номера попыток и задержки.
4. Код устойчив к `TimeoutException`, но сразу падает на других ошибках.

### Решение
```java
import reactor.core.publisher.Mono;
import reactor.util.retry.Retry;
import java.time.Duration;
import java.util.concurrent.atomic.AtomicInteger;

public class Task3ErrorHandling {
    private static final AtomicInteger attempts = new AtomicInteger(0);

    public static void main(String[] args) {
        Mono<String> resilientCall = Mono.defer(() -> {
            int attempt = attempts.incrementAndGet();
            System.out.println("Попытка #" + attempt);
            if (attempt <= 2) {
                return Mono.error(new RuntimeException("API unavailable"));
            }
            return Mono.just("Данные получены");
        })
        .retryWhen(Retry.backoff(3, Duration.ofMillis(100))
            .maxBackoff(Duration.ofMillis(500))
            .jitter(0.5) // Случайный разброс +-50% для предотвращения thundering herd
            .filter(e -> e instanceof RuntimeException) // Ретраим только runtime-ошибки
            .doBeforeRetry(signal -> 
                System.out.println("Повтор через " + signal.totalDelay() + "мс")
            )
        )
        .onErrorResume(RuntimeException.class, e -> 
            Mono.just("Fallback: использованы кешированные данные")
        );

        resilientCall.subscribe(System.out::println);
        try { Thread.sleep(2000); } catch (InterruptedException e) {}
    }
}
```

### Ожидаемый вывод
```
Попытка #1
Повтор через ~100 ms
Попытка #2
Повтор через ~250 ms
Попытка #3
Данные получены
```

### Теоретическая справка
- Простой `.retry(3)` опасен: мгновенные повторы перегружают падающий сервис.
- `Retry.backoff` + `jitter` – production-стандарт. Разброс предотвращает синхронизацию запросов от тысяч клиентов.
- `onErrorResume` заменяет терминальный поток на альтернативный `Publisher`, сохраняя контракт реактивности.

---

## Задание 4. Spring WebFlux: реактивный REST-контроллер

### Описание
Создай приложение Spring Boot с реактивным контроллером. Реализуй эндпоинт `GET /api/users/{id}`, возвращающий `Mono<ResponseEntity<User>>`. Обработай случай отсутствия пользователя (HTTP 404). Добавь эндпоинт `GET /api/users`, возвращающий `Flux<User>`.

### Критерии выполнения
1. Использованы `@RestController`, `Mono`, `Flux`, `ResponseEntity`.
2. Реализована обработка `defaultIfEmpty` или `switchIfEmpty`.
3. Код не содержит блокирующих вызовов.
4. Продемонстрирована разница между `Mono.just()` и `Mono.empty()`.

### Решение
```java
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.util.Map;

@RestController
@RequestMapping("/api/users")
public class UserController {

    // Имитация реактивного репозитория
    private final Map<Long, String> db = Map.of(1L, "Alice", 2L, "Bob");

    @GetMapping("/{id}")
    public Mono<ResponseEntity<String>> getUser(@PathVariable Long id) {
        return Mono.justOrEmpty(db.get(id))
            .map(name -> ResponseEntity.ok("User: " + name))
            .defaultIfEmpty(ResponseEntity.notFound().build());
    }

    @GetMapping
    public Flux<String> getAllUsers() {
        return Flux.fromIterable(db.values())
            .delayElements(java.time.Duration.ofMillis(100)); // Имитация стриминга
    }
}
```
*Запусти `SpringApplication.run(...)`, открой `http://localhost:8080/api/users/1` и `http://localhost:8080/api/users/99`.*

### Теоретическая справка
- WebFlux автоматически подписывается на возвращаемый `Mono/Flux`, сериализует результат в JSON и отправляет клиенту.
- `Mono<Void>` используется для операций без тела ответа (например, `DELETE`).
- В отличие от Spring MVC, здесь **нет пула потоков на запрос**. Один Event Loop обрабатывает тысячи соединений, поэтому блокировка в контроллере = остановка всего сервера.

---

## Задание 5. Тестирование: StepVerifier, WebTestClient, BlockHound

### Описание
Напиши три теста.
1. Юнит-тест реактивной цепочки через `StepVerifier` с проверкой backpressure.
2. Интеграционный тест контроллера через `@WebFluxTest` + `WebTestClient`.
3. Подключи `BlockHound` для детекции скрытых блокировок.

### Критерии выполнения
1. Тесты не используют `.block()`.
2. `StepVerifier` проверяет последовательность, ошибку и завершение.
3. `WebTestClient` проверяет статус 200 и тело ответа.
4. `BlockHound` корректно интегрирован и выбрасывает `BlockingOperationError` при намеренной блокировке.

### Решение
```java
import io.projectreactor.test.StepVerifier;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.reactive.WebFluxTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.test.web.reactive.server.WebTestClient;
import reactor.core.publisher.Mono;
import reactor.test.StepVerifier;
import reactor.core.scheduler.Schedulers;
import reactor.blockhound.BlockingOperationError;
import reactor.blockhound.BlockHound;

import static org.mockito.Mockito.when;
import static org.junit.jupiter.api.Assertions.assertEquals;

// 1. StepVerifier
class ReactorUnitTest {
    @Test
    void testFluxWithBackpressure() {
        Flux<String> flux = Flux.just("A", "B", "C").delayElements(java.time.Duration.ofMillis(10));
        
        StepVerifier.create(flux, 1) // Подписчик с буфером 1
            .thenRequest(2)          // Запрашиваем 2 элемента
            .expectNext("A", "B")
            .thenRequest(1)
            .expectNext("C")
            .verifyComplete();
    }
}

// 2. WebTestClient + @WebFluxTest
@WebFluxTest(controllers = UserController.class)
class UserControllerIntegrationTest {
    @Autowired private WebTestClient webTestClient;
    @MockBean private UserRepository userRepository; // Заглушка

    @Test
    void shouldReturnUser() {
        when(userRepository.findById(1L)).thenReturn(Mono.just("Alice"));
        
        webTestClient.get()
            .uri("/api/users/1")
            .exchange()
            .expectStatus().isOk()
            .expectBody(String.class)
            .value(result -> assertEquals("User: Alice", result));
    }
}

// 3. BlockHound Integration
class BlockHoundTest {
    @BeforeAll
    static void installBlockHound() {
        BlockHound.install();
    }

    @Test
    void shouldDetectBlockingCall() {
        Mono<String> mono = Mono.fromCallable(() -> {
            Thread.sleep(100); // Запрещено в реактивном потоке
            return "OK";
        }).subscribeOn(Schedulers.boundedElastic());

        StepVerifier.create(mono)
            .expectError(BlockingOperationError.class) // BlockHound перехватит sleep
            .verify();
    }
}
```

### Теоретическая справка
- **`StepVerifier`** имитирует реального подписчика, проверяет сигналы `onNext/onError/onComplete` и управляет `request(n)`.
- **`WebTestClient`** тестирует полный HTTP-цикл без запуска реального сервера. `@WebFluxTest` загружает только веб-слой, ускоряя запуск в 5–10 раз.
- **`BlockHound`** патчит байткод на старте JVM. Перед каждым потенциально блокирующим вызовом проверяет контекст потока. Если поток реактивный, кидает `BlockingOperationError`. Компилятор этого не видит, поэтому инструмент критичен для CI.

---