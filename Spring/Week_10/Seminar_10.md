
# Подготовка

## 0) RestClientConfig для локального дебага
```java 
@Configuration
public class RestClientConfig {
    @Bean
    RestClient restClient(RestClient.Builder builder) {
        return builder
                .baseUrl("http://localhost:8080")
                .defaultHeader(HttpHeaders.USER_AGENT, "seminar-app/1.0")
                .build();
    }
}
```
## 1) DTO (те же, что использует клиент)

```java

public record TaskDto(Long id, String title, boolean completed) {}

public record TaskResponse(Long id, String title, String description, boolean completed) {}

public record TaskCreateRequest(String title, String description) {}

public record TaskCreateResponse(Long id, String title, String description) {}

public record TaskPutRequest(String title, String description, boolean completed) {}

public record TaskPatchRequest(Boolean completed) {}

// RFC7807-like
public record ProblemDetails(String type, String title, int status, String detail) {}
```

## 2) InMemory “хранилище”
```java 
@Component
public class TaskStore {
    private final Map<Long, TaskResponse> tasks = new ConcurrentHashMap<>();
    private final AtomicLong seq = new AtomicLong(40);

    public TaskStore() {
        tasks.put(1L, new TaskResponse(1L, "Buy milk", "2 liters", false));
        tasks.put(2L, new TaskResponse(2L, "Read book", "Spring HTTP", true));
    }

    public List<TaskResponse> all() {
        return tasks.values().stream()
                .sorted(Comparator.comparing(TaskResponse::id))
                .toList();
    }

    public Optional<TaskResponse> get(long id) {
        return Optional.ofNullable(tasks.get(id));
    }

    public TaskResponse create(TaskCreateRequest req) {
        long id = seq.incrementAndGet();
        TaskResponse tr = new TaskResponse(id, req.title(), req.description(), false);
        tasks.put(id, tr);
        return tr;
    }

    public TaskResponse put(long id, TaskPutRequest req) {
        TaskResponse tr = new TaskResponse(id, req.title(), req.description(), req.completed());
        tasks.put(id, tr);
        return tr;
    }

    public TaskResponse patch(long id, TaskPatchRequest req) {
        TaskResponse old = tasks.get(id);
        if (old == null) return null;
        boolean completed = (req.completed() != null) ? req.completed() : old.completed();
        TaskResponse upd = new TaskResponse(id, old.title(), old.description(), completed);
        tasks.put(id, upd);
        return upd;
    }

    public boolean delete(long id) {
        return tasks.remove(id) != null;
    }
}
```

## 3) Controller: “внешнее API для RestClient” (все задания)
```java 
@RestController
@RequestMapping("/v1")
public class DebugExternalApiController {

    private final TaskStore store;

    public DebugExternalApiController(TaskStore store) {
        this.store = store;
    }

    // --------- Задания 2, 13, 14, 18: GET /tasks/{id} ----------
    @GetMapping(value = "/tasks/{id}", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<?> getTask(
            @PathVariable long id,
            @RequestHeader(value = "X-Request-Id", required = false) String requestId,
            @RequestHeader(value = HttpHeaders.ACCEPT, required = false) String accept
    ) {
        // имитация 406 если клиент просит XML
        if (accept != null && accept.contains(MediaType.APPLICATION_XML_VALUE)) {
            return ResponseEntity.status(HttpStatus.NOT_ACCEPTABLE).build();
        }

        return store.get(id)
                .<ResponseEntity<?>>map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .contentType(MediaType.APPLICATION_PROBLEM_JSON)
                        .body(new ProblemDetails(
                                "https://example.com/errors/not-found",
                                "Not Found",
                                404,
                                "Task " + id + " not found"
                        )));
    }

    // --------- Задания 3, 4: GET /tasks?completed=&limit= ----------
    @GetMapping(value = "/tasks", produces = MediaType.APPLICATION_JSON_VALUE)
    public List<TaskDto> listTasks(
            @RequestParam(required = false) Boolean completed,
            @RequestParam(required = false, defaultValue = "100") int limit
    ) {
        Stream<TaskResponse> s = store.all().stream();
        if (completed != null) {
            s = s.filter(t -> t.completed() == completed);
        }
        return s.limit(limit)
                .map(t -> new TaskDto(t.id(), t.title(), t.completed()))
                .toList();
    }

    // --------- Задание 5, 19: POST /tasks -> 201 + Location ----------
    @PostMapping(
            value = "/tasks",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE
    )
    public ResponseEntity<TaskCreateResponse> create(@RequestBody TaskCreateRequest req) {
        // имитация 409 conflict по названию
        if ("exists".equalsIgnoreCase(req.title())) {
            return ResponseEntity.status(HttpStatus.CONFLICT).build();
        }

        TaskResponse created = store.create(req);
        URI location = URI.create("/v1/tasks/" + created.id());
        return ResponseEntity.created(location)
                .body(new TaskCreateResponse(created.id(), created.title(), created.description()));
    }

    // --------- Задание 6: PUT /tasks/{id} ----------
    @PutMapping(
            value = "/tasks/{id}",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE
    )
    public ResponseEntity<TaskResponse> put(@PathVariable long id, @RequestBody TaskPutRequest req) {
        TaskResponse saved = store.put(id, req);
        return ResponseEntity.ok(saved);
    }

    // --------- Задание 7: PATCH /tasks/{id} ----------
    @PatchMapping(
            value = "/tasks/{id}",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE
    )
    public ResponseEntity<?> patch(@PathVariable long id, @RequestBody TaskPatchRequest req) {
        TaskResponse patched = store.patch(id, req);
        if (patched == null) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .contentType(MediaType.APPLICATION_PROBLEM_JSON)
                    .body(new ProblemDetails(
                            "https://example.com/errors/not-found",
                            "Not Found",
                            404,
                            "Task " + id + " not found"
                    ));
        }
        return ResponseEntity.ok(patched);
    }

    // --------- Задание 8: DELETE /tasks/{id} -> 204 ----------
    @DeleteMapping("/tasks/{id}")
    public ResponseEntity<Void> delete(@PathVariable long id) {
        store.delete(id); // даже если не было — всё равно 204 (частая практика)
        return ResponseEntity.noContent().build();
    }

    // --------- Задание 10: Basic Auth пример GET /secure ----------
    @GetMapping(value = "/secure", produces = MediaType.TEXT_PLAIN_VALUE)
    public ResponseEntity<String> secureBasic(@RequestHeader(value = HttpHeaders.AUTHORIZATION, required = false) String auth) {
        if (auth == null || !auth.startsWith("Basic ")) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .header(HttpHeaders.WWW_AUTHENTICATE, "Basic realm=\"debug\"")
                    .body("Missing Basic auth");
        }

        String b64 = auth.substring("Basic ".length());
        String decoded = new String(Base64.getDecoder().decode(b64), StandardCharsets.UTF_8);

        if (!"alice:secret".equals(decoded)) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body("Bad credentials");
        }
        return ResponseEntity.ok("OK (basic)");
    }

    // --------- Задание 11/12: Bearer Auth пример GET /tasks-auth/{id} ----------
    @GetMapping(value = "/tasks-auth/{id}", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<?> getTaskBearer(
            @PathVariable long id,
            @RequestHeader(value = HttpHeaders.AUTHORIZATION, required = false) String auth
    ) {
        if (auth == null || !auth.startsWith("Bearer ")) { // важен пробел
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        String token = auth.substring("Bearer ".length());
        if (!"valid-token".equals(token)) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        return getTask(id, null, MediaType.APPLICATION_JSON_VALUE);
    }

    // --------- Задания 17: 415 Unsupported Media Type (если не JSON) ----------
    // Это автоматически сделает Spring, если вы отправите не application/json в endpoints с consumes=JSON.
    // Но добавим отдельный endpoint для демонстрации "сервер ждет форму":
    @PostMapping(value = "/tasks-form", consumes = MediaType.APPLICATION_FORM_URLENCODED_VALUE)
    public ResponseEntity<String> createForm(@RequestParam String title) {
        return ResponseEntity.ok("created-form title=" + title);
    }

    // --------- Задание 20: бинарный файл ----------
    @GetMapping(value = "/files/{id}", produces = MediaType.APPLICATION_OCTET_STREAM_VALUE)
    public ResponseEntity<byte[]> download(@PathVariable String id) {
        byte[] bytes = ("file-" + id + "\n").getBytes(StandardCharsets.UTF_8);
        return ResponseEntity.ok()
                .header(HttpHeaders.CONTENT_DISPOSITION, "attachment; filename=\""+id+".txt\"")
                .body(bytes);
    }

    // --------- Доп: 429 Too Many Requests (из лекции, без resil4j) ----------
    @GetMapping(value = "/limited", produces = MediaType.TEXT_PLAIN_VALUE)
    public ResponseEntity<String> limited() {
        return ResponseEntity.status(429)
                .header("Retry-After", "3")
                .body("Too many requests, retry after 3 seconds");
    }

    // --------- Доп: “сервер вернул HTML вместо JSON” ----------
    @GetMapping(value = "/html-error", produces = MediaType.TEXT_HTML_VALUE)
    public ResponseEntity<String> htmlError() {
        return ResponseEntity.status(502).body("<html><body><h1>Bad Gateway</h1></body></html>");
    }
}
```

# RestClient

Ниже — 20 заданий для семинара по HTTP-запросам/ответам и Spring `RestClient` (без темы Circuit Breaker/Rate Limiter). В каждом задании дан итоговый код (готовое решение). Для краткости предполагается Spring Boot 3 / Spring 6.

---

## 1) Базовая конфигурация RestClient с baseUrl и User-Agent
**Цель:** сделать единый `RestClient` bean.

**Итоговый код:**
```java
@Configuration
public class RestClientConfig {
    @Bean
    RestClient restClient(RestClient.Builder builder) {
        return builder
                .baseUrl("https://api.example.com")
                .defaultHeader(HttpHeaders.USER_AGENT, "seminar-app/1.0")
                .build();
    }
}
```

---

## 2) GET по path variable + Accept: JSON
**Цель:** корректно просить JSON и подставлять `{id}`.

**Итоговый код:**
```java
public record TaskResponse(Long id, String title, String description, boolean completed) {}

@Service
public class TasksClient {
    private final RestClient restClient;

    public TasksClient(RestClient restClient) {
        this.restClient = restClient;
    }

    public TaskResponse getTask(long id) {
        return restClient.get()
                .uri("/v1/tasks/{id}", id)
                .accept(MediaType.APPLICATION_JSON)
                .retrieve()
                .body(TaskResponse.class);
    }
}
```

---

## 3) GET со списком и ParameterizedTypeReference
**Цель:** получить `List<TaskDto>`.

**Итоговый код:**
```java
public record TaskDto(Long id, String title, boolean completed) {}

public List<TaskDto> getTasks() {
    return restClient.get()
            .uri("/v1/tasks")
            .accept(MediaType.APPLICATION_JSON)
            .retrieve()
            .body(new ParameterizedTypeReference<List<TaskDto>>() {});
}
```

---

## 4) GET с query params через uriBuilder (без “склейки строк”)
**Цель:** правильный encoding query-параметров.

**Итоговый код:**
```java
public List<TaskDto> getTasks(boolean completed, int limit) {
    return restClient.get()
            .uri(ub -> ub.path("/v1/tasks")
                    .queryParam("completed", completed)
                    .queryParam("limit", limit)
                    .build())
            .accept(MediaType.APPLICATION_JSON)
            .retrieve()
            .body(new ParameterizedTypeReference<List<TaskDto>>() {});
}
```

---

## 5) POST JSON: выставить Content-Type и Accept
**Цель:** не забыть `Content-Type: application/json`.

**Итоговый код:**
```java
public record TaskCreateRequest(String title, String description) {}
public record TaskCreateResponse(Long id, String title, String description) {}

public TaskCreateResponse create(TaskCreateRequest req) {
    return restClient.post()
            .uri("/v1/tasks")
            .contentType(MediaType.APPLICATION_JSON)
            .accept(MediaType.APPLICATION_JSON)
            .body(req)
            .retrieve()
            .body(TaskCreateResponse.class);
}
```

---

## 6) PUT: “полная замена” ресурса (DTO целиком)
**Цель:** показать семантику PUT.

**Итоговый код:**
```java
public record TaskPutRequest(String title, String description, boolean completed) {}

public TaskResponse put(long id, TaskPutRequest req) {
    return restClient.put()
            .uri("/v1/tasks/{id}", id)
            .contentType(MediaType.APPLICATION_JSON)
            .accept(MediaType.APPLICATION_JSON)
            .body(req)
            .retrieve()
            .body(TaskResponse.class);
}
```

---

## 7) PATCH: частичное обновление
**Цель:** обновить только `completed`.

**Итоговый код:**
```java
public record TaskPatchRequest(Boolean completed) {}

public TaskResponse patch(long id, TaskPatchRequest req) {
    return restClient.patch()
            .uri("/v1/tasks/{id}", id)
            .contentType(MediaType.APPLICATION_JSON)
            .accept(MediaType.APPLICATION_JSON)
            .body(req)
            .retrieve()
            .body(TaskResponse.class);
}
```

---

## 8) DELETE и правильная обработка 204 No Content
**Цель:** не парсить тело при 204.

**Итоговый код:**
```java
public void delete(long id) {
    restClient.delete()
            .uri("/v1/tasks/{id}", id)
            .retrieve()
            .toBodilessEntity();
}
```

---

## 9) Кастомный заголовок X-Request-Id на один запрос
**Цель:** проброс корреляции.

**Итоговый код:**
```java
public TaskResponse getTask(long id, String requestId) {
    return restClient.get()
            .uri("/v1/tasks/{id}", id)
            .header("X-Request-Id", requestId)
            .accept(MediaType.APPLICATION_JSON)
            .retrieve()
            .body(TaskResponse.class);
}
```

---

## 10) Basic Auth: сформировать Authorization: Basic base64 (user:pass)
**Цель:** понять что base64 — это кодировка.

**Итоговый код:**
```java
private static String basicAuth(String user, String pass) {
    String token = user + ":" + pass;
    String b64 = Base64.getEncoder().encodeToString(token.getBytes(StandardCharsets.UTF_8));
    return "Basic " + b64;
}

public String callBasic() {
    return restClient.get()
            .uri("/v1/secure")
            .header(HttpHeaders.AUTHORIZATION, basicAuth("alice", "secret"))
            .retrieve()
            .body(String.class);
}
```

---

## 11) Bearer Token: “Bearer ” + пробел
**Цель:** не допустить типовую ошибку.

**Итоговый код:**
```java
public TaskResponse getTaskAuthorized(long id, String token) {
    return restClient.get()
            .uri("/v1/tasks/{id}", id)
            .header(HttpHeaders.AUTHORIZATION, "Bearer " + token)
            .accept(MediaType.APPLICATION_JSON)
            .retrieve()
            .body(TaskResponse.class);
}
```

---

## 12) Обёртка AuthorizedRestClient (токен динамический)
**Цель:** не писать Authorization в каждом методе.

**Итоговый код:**
```java
public interface TokenProvider {
    String getAccessToken();
}

@Component
public class AuthorizedRestClient {
    private final RestClient restClient;
    private final TokenProvider tokenProvider;

    public AuthorizedRestClient(RestClient restClient, TokenProvider tokenProvider) {
        this.restClient = restClient;
        this.tokenProvider = tokenProvider;
    }

    public RestClient.RequestHeadersSpec<?> get(String uriTemplate, Object... vars) {
        return restClient.get()
                .uri(uriTemplate, vars)
                .header(HttpHeaders.AUTHORIZATION, "Bearer " + tokenProvider.getAccessToken());
    }

    public RestClient.RequestBodySpec post(String uriTemplate, Object... vars) {
        return restClient.post()
                .uri(uriTemplate, vars)
                .header(HttpHeaders.AUTHORIZATION, "Bearer " + tokenProvider.getAccessToken());
    }
}
```

---

## 13) GET, который возвращает Optional: 404 → Optional.Empty ()
**Цель:** маппинг 404 в “нет данных”.

**Итоговый код:**
```java
public Optional<TaskResponse> findTask(long id) {
    try {
        TaskResponse resp = restClient.get()
                .uri("/v1/tasks/{id}", id)
                .accept(MediaType.APPLICATION_JSON)
                .retrieve()
                .body(TaskResponse.class);
        return Optional.ofNullable(resp);
    } catch (HttpClientErrorException.NotFound e) {
        return Optional.empty();
    }
}
```

---

## 14) Problem Details: распарсить тело ошибки и бросить своё исключение
**Цель:** не “глотать” body ошибки.

**Итоговый код:**
```java
public record ProblemDetails(String type, String title, int status, String detail) {}

public static class TaskNotFoundException extends RuntimeException {
    public TaskNotFoundException(String message) { super(message); }
}
public static class ExternalApiException extends RuntimeException {
    public ExternalApiException(String message, Throwable cause) { super(message, cause); }
}

private final ObjectMapper objectMapper = new ObjectMapper();

public TaskResponse getTaskOrThrow(long id) {
    try {
        return restClient.get()
                .uri("/v1/tasks/{id}", id)
                .accept(MediaType.APPLICATION_JSON)
                .retrieve()
                .body(TaskResponse.class);
    } catch (HttpClientErrorException e) {
        ProblemDetails pd = tryReadProblem(e);
        if (e.getStatusCode().value() == 404) {
            throw new TaskNotFoundException(pd != null ? pd.detail() : ("Task not found: " + id));
        }
        throw new ExternalApiException("Client error: " + e.getStatusCode() + " detail=" + (pd != null ? pd.detail() : ""), e);
    } catch (HttpServerErrorException e) {
        throw new ExternalApiException("Server error: " + e.getStatusCode(), e);
    }
}

private ProblemDetails tryReadProblem(RestClientResponseException e) {
    try {
        byte[] body = e.getResponseBodyAsByteArray();
        if (body == null || body.length == 0) return null;
        return objectMapper.readValue(body, ProblemDetails.class);
    } catch (Exception ex) {
        return null;
    }
}
```

---

## 15) “Безопасное” Логирование тела ошибки (лимит + без секретов)
**Цель:** подготовить safeBody.

**Итоговый код:**
```java
private String safeBody(RestClientResponseException e) {
    byte[] b = e.getResponseBodyAsByteArray();
    if (b == null || b.length == 0) return "";
    String s = new String(b, StandardCharsets.UTF_8);
    int max = 2000;
    if (s.length() > max) s = s.substring(0, max) + "...(truncated)";
    // Пример примитивного маскирования
    s = s.replaceAll("(?i)\"access_token\"\\s*:\\s*\"[^\"]+\"", "\"access_token\":\"***\"");
    s = s.replaceAll("(?i)\"token\"\\s*:\\s*\"[^\"]+\"", "\"token\":\"***\"");
    return s;
}
```

---

## 16) Настройка таймаутов (connect + read) через JDK HttpClient
**Цель:** не “висеть” бесконечно.

**Итоговый код:**
```java
@Configuration
public class RestClientTimeoutConfig {

    @Bean
    public RestClient restClient(RestClient.Builder builder) {
        HttpClient httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(2))
                .build();

        JdkClientHttpRequestFactory rf = new JdkClientHttpRequestFactory(httpClient);
        rf.setReadTimeout(Duration.ofSeconds(3));

        return builder
                .requestFactory(rf)
                .baseUrl("https://api.example.com")
                .defaultHeader(HttpHeaders.USER_AGENT, "seminar-app/1.0")
                .build();
    }
}
```

---

## 17) Обработка 415 Unsupported Media Type: исправить POST, добавив Content-Type
**Цель:** типичная диагностика/лечение 415.

**Итоговый код (правильный запрос):**
```java
public TaskCreateResponse createFixed(TaskCreateRequest req) {
    return restClient.post()
            .uri("/v1/tasks")
            .contentType(MediaType.APPLICATION_JSON) // ключевое
            .accept(MediaType.APPLICATION_JSON)
            .body(req)
            .retrieve()
            .body(TaskCreateResponse.class);
}
```

---

## 18) Обработка 406 Not Acceptable: выставить Accept: JSON
**Цель:** не получить HTML/XML вместо JSON.

**Итоговый код:**
```java
public TaskResponse getTaskAcceptJson(long id) {
    return restClient.get()
            .uri("/v1/tasks/{id}", id)
            .accept(MediaType.APPLICATION_JSON) // ключевое
            .retrieve()
            .body(TaskResponse.class);
}
```

---

## 19) 201 Created и чтение Location из заголовков
**Цель:** уметь доставать `Location`, даже если тело есть/нет.

**Итоговый код:**
```java
public record CreatedWithLocation(URI location, TaskCreateResponse body) {}

public CreatedWithLocation createAndReadLocation(TaskCreateRequest req) {
    ResponseEntity<TaskCreateResponse> entity = restClient.post()
            .uri("/v1/tasks")
            .contentType(MediaType.APPLICATION_JSON)
            .accept(MediaType.APPLICATION_JSON)
            .body(req)
            .retrieve()
            .toEntity(TaskCreateResponse.class);

    return new CreatedWithLocation(entity.getHeaders().getLocation(), entity.getBody());
}
```

---

## 20) Скачать бинарный файл (bytes) с Accept: octet-stream
**Цель:** не пытаться парсить как JSON.

**Итоговый код:**
```java
public byte[] downloadFile(String fileId) {
    return restClient.get()
            .uri("/v1/files/{id}", fileId)
            .accept(MediaType.APPLICATION_OCTET_STREAM)
            .retrieve()
            .body(byte[].class);
}
```

# CB и RL
## 0) Подготовка проекта (общие зависимости)

Build. Gradle (добавить к вашему)
```gradle
dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.boot:spring-boot-starter-actuator'

    implementation 'io.github.resilience4j:resilience4j-spring-boot3'
    implementation 'io.github.resilience4j:resilience4j-micrometer'

    // опционально: для более удобных тестов
    testImplementation 'org.springframework.boot:spring-boot-starter-test'
}
```

Application.yaml (базовая конфигурация, будем дополнять)
```yaml
management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus

resilience4j:
  circuitbreaker:
    instances: {}
  ratelimiter:
    instances: {}
```

---

## Задание 1: “Ненадёжный” внешний сервис-имитатор

**Цель:** иметь endpoint, который иногда падает/тормозит, чтобы тестировать CB.

### UnstableController. Java
```java
package com.example.demo.unstable;

import org.springframework.web.bind.annotation.*;

import java.util.concurrent.ThreadLocalRandom;

@RestController
@RequestMapping("/unstable")
public class UnstableController {

    // Пример: иногда 500, иногда 200
    @GetMapping("/random-fail")
    public String randomFail(@RequestParam(defaultValue = "0.3") double failProbability) {
        if (ThreadLocalRandom.current().nextDouble() < failProbability) {
            throw new RuntimeException("Random failure");
        }
        return "OK";
    }

    // Пример: искусственная задержка
    @GetMapping("/slow")
    public String slow(@RequestParam(defaultValue = "300") long ms) throws InterruptedException {
        Thread.sleep(ms);
        return "SLOW_OK " + ms;
    }
}
```

### GlobalExceptionHandler (если ещё нет)
Чтобы видеть нормальные ответы, можно вернуть 500 JSON (не обязательно, но удобно).

**Готово, если:** `/unstable/random-fail` иногда отвечает 500.

---

## Задание 2: Клиент, который вызывает “внешний сервис”

**Цель:** отделить “наш API” от “внешнего сервиса”.

### ExternalClient. Java
```java
package com.example.demo.client;

import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

@Component
public class ExternalClient {
    private final RestClient restClient;

    public ExternalClient(RestClient.Builder builder) {
        this.restClient = builder.baseUrl("http://localhost:8080").build();
    }

    public String callRandomFail(double p) {
        return restClient.get()
                .uri("/unstable/random-fail?failProbability={p}", p)
                .retrieve()
                .body(String.class);
    }

    public String callSlow(long ms) {
        return restClient.get()
                .uri("/unstable/slow?ms={ms}", ms)
                .retrieve()
                .body(String.class);
    }
}
```

### ApiController. Java
```java
package com.example.demo.api;

import com.example.demo.client.ExternalClient;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api")
public class ApiController {

    private final ExternalClient external;

    public ApiController(ExternalClient external) {
        this.external = external;
    }

    @GetMapping("/proxy/random-fail")
    public String proxyRandomFail(@RequestParam(defaultValue = "0.3") double p) {
        return external.callRandomFail(p);
    }

    @GetMapping("/proxy/slow")
    public String proxySlow(@RequestParam(defaultValue = "300") long ms) {
        return external.callSlow(ms);
    }
}
```

**Готово, если:** `/api/proxy/random-fail` падает, когда падает внешний.

---

## Задание 3: Circuit Breaker через аннотацию + fallback

**Цель:** включить CB на вызов внешнего сервиса и сделать graceful degradation.

 Application. Yml
```yaml
resilience4j:
  circuitbreaker:
    instances:
      externalService:
        slidingWindowType: COUNT_BASED
        slidingWindowSize: 10
        failureRateThreshold: 50
        waitDurationInOpenState: 5s
        permittedNumberOfCallsInHalfOpenState: 3
        minimumNumberOfCalls: 5
```

### ExternalFacade. Java
```java
package com.example.demo.facade;

import com.example.demo.client.ExternalClient;
import io.github.resilience4j.circuitbreaker.annotation.CircuitBreaker;
import org.springframework.stereotype.Service;

@Service
public class ExternalFacade {

    private final ExternalClient client;

    public ExternalFacade(ExternalClient client) {
        this.client = client;
    }

    @CircuitBreaker(name = "externalService", fallbackMethod = "fallbackRandomFail")
    public String callRandomFail(double p) {
        return client.callRandomFail(p);
    }

    // fallback signature: same args + Throwable at end (или Exception)
    private String fallbackRandomFail(double p, Throwable t) {
        return "FALLBACK: external unavailable (" + t.getClass().getSimpleName() + ")";
    }
}
```

### ApiController (обновить)
```java
private final ExternalFacade facade;

public ApiController(ExternalFacade facade) { this.facade = facade; }

@GetMapping("/proxy/random-fail")
public String proxyRandomFail(@RequestParam(defaultValue = "0.7") double p) {
    return facade.callRandomFail(p);
}
```

**Готово, если:** после серии ошибок `/api/proxy/random-fail` начинает быстро отдавать `FALLBACK...` (CB open).

---


## Задание 4: Circuit Breaker на “медленные ответы” (slow call)

**Цель:** считать “слишком долго” как проблему.

```yaml
resilience4j:
  circuitbreaker:
    instances:
      externalSlow:
        slidingWindowType: COUNT_BASED
        slidingWindowSize: 10
        minimumNumberOfCalls: 5
        failureRateThreshold: 50
        slowCallRateThreshold: 50
        slowCallDurationThreshold: 200ms
        waitDurationInOpenState: 5s
```
### ExternalSlowFacade. Java
```java
package com.example.demo.facade;

import com.example.demo.client.ExternalClient;
import io.github.resilience4j.circuitbreaker.annotation.CircuitBreaker;
import org.springframework.stereotype.Service;

@Service
public class ExternalSlowFacade {

    private final ExternalClient client;

    public ExternalSlowFacade(ExternalClient client) {
        this.client = client;
    }

    @CircuitBreaker(name = "externalSlow", fallbackMethod = "fallbackSlow")
    public String callSlow(long ms) {
        return client.callSlow(ms);
    }

    private String fallbackSlow(long ms, Throwable t) {
        return "FALLBACK_SLOW: degraded mode";
    }
}
```

### ApiController (добавить endpoint)
```java
private final ExternalSlowFacade slowFacade;

@GetMapping("/proxy/slow")
public String proxySlow(@RequestParam(defaultValue = "300") long ms) {
    return slowFacade.callSlow(ms);
}
```

**Готово, если:** при `ms=300` (больше 200ms) после нескольких вызовов CB открывается и начинает отдавать fallback.

---

## Задание 5: RateLimiter через аннотацию

**Цель:** ограничить частоту запросов на endpoint.

```yaml
resilience4j:
  ratelimiter:
    instances:
      helloLimiter:
        limitForPeriod: 5
        limitRefreshPeriod: 10s
        timeoutDuration: 0
```

### RateLimitedHelloController. Java
```java
package com.example.demo.ratelimit;

import io.github.resilience4j.ratelimiter.annotation.RateLimiter;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class RateLimitedHelloController {

    @RateLimiter(name = "helloLimiter", fallbackMethod = "fallbackHello")
    @GetMapping("/api/limited/hello")
    public String hello() {
        return "HELLO";
    }

    private String fallbackHello(Throwable t) {
        return "TOO_MANY_REQUESTS";
    }
}
```

**Готово, если:** 6-й запрос в течение 10 секунд начинает возвращать `TOO_MANY_REQUESTS`.

---


## Задание 6: Правильный HTTP статус 429 для RateLimiter

**Цель:** вместо просто строки возвращать 429 Too Many Requests.

### RateLimitedHelloController (заменить возвращаемый тип)
```java
package com.example.demo.ratelimit;

import io.github.resilience4j.ratelimiter.annotation.RateLimiter;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class RateLimitedHelloController {

    @RateLimiter(name = "helloLimiter", fallbackMethod = "fallbackHello")
    @GetMapping("/api/limited/hello")
    public ResponseEntity<String> hello() {
        return ResponseEntity.ok("HELLO");
    }

    private ResponseEntity<String> fallbackHello(Throwable t) {
        // Resilience4j кинет RequestNotPermitted
        return ResponseEntity.status(429).body("TOO_MANY_REQUESTS");
    }
}
```

**Готово, если:** при превышении лимита ответ имеет статус 429.

---

## Задание 7: Композиция RateLimiter + CircuitBreaker

**Цель:** защитить внешний сервис и от перегрузки, и от аварий.

### ExternalProtectedFacade. Java
```java
package com.example.demo.facade;

import com.example.demo.client.ExternalClient;
import io.github.resilience4j.circuitbreaker.annotation.CircuitBreaker;
import io.github.resilience4j.ratelimiter.annotation.RateLimiter;
import org.springframework.stereotype.Service;

@Service
public class ExternalProtectedFacade {

    private final ExternalClient client;

    public ExternalProtectedFacade(ExternalClient client) {
        this.client = client;
    }

    @RateLimiter(name = "externalLimiter", fallbackMethod = "fallback")
    @CircuitBreaker(name = "externalService", fallbackMethod = "fallback")
    public String callRandomFail(double p) {
        return client.callRandomFail(p);
    }

    private String fallback(double p, Throwable t) {
        // Одна точка деградации: и на лимит, и на circuit open, и на ошибки
        return "DEGRADED: " + t.getClass().getSimpleName();
    }
}
```

```yaml
resilience4j:
  ratelimiter:
    instances:
      externalLimiter:
        limitForPeriod: 3
        limitRefreshPeriod: 5s
        timeoutDuration: 0
```

**Готово, если:** при частых вызовах сначала срабатывает лимит (DEGRADED: RequestNotPermitted), а при ошибках — CB.

---

## Задание 8: Наблюдаемость — метрики Actuator для CB и RL

**Цель:** увидеть, что Resilience4j отдаёт метрики в Micrometer.

### Проверка
Сделайте несколько запросов, затем откройте:

- `/actuator/metrics/resilience4j.circuitbreaker.state`
- `/actuator/metrics/resilience4j.circuitbreaker.calls`
- `/actuator/metrics/resilience4j.ratelimiter.calls`

**Готово, если:** в метриках есть tags с `name=externalService` / `helloLimiter`, и значения меняются после запросов.

> Примечание: названия метрик могут слегка отличаться по версии, но ключевое — вы видите circuit/rate limiter в actuator metrics.

---

## Задание 9: События Circuit Breaker в лог (EventPublisher)

**Цель:** логировать переходы состояний: CLOSED → OPEN → HALF_OPEN.

### CircuitBreakerLoggingConfig. Java
```java
package com.example.demo.config;

import io.github.resilience4j.circuitbreaker.CircuitBreakerRegistry;
import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Configuration;

@Configuration
public class CircuitBreakerLoggingConfig {

    private static final Logger log = LoggerFactory.getLogger(CircuitBreakerLoggingConfig.class);
    private final CircuitBreakerRegistry registry;

    public CircuitBreakerLoggingConfig(CircuitBreakerRegistry registry) {
        this.registry = registry;
    }

    @PostConstruct
    void init() {
        registry.getAllCircuitBreakers().forEach(cb ->
            cb.getEventPublisher()
              .onStateTransition(e -> log.warn("CB transition: name={} {} -> {}",
                      cb.getName(),
                      e.getStateTransition().getFromState(),
                      e.getStateTransition().getToState()))
              .onFailureRateExceeded(e -> log.warn("CB failureRateExceeded: name={} rate={}",
                      cb.getName(), e.getFailureRate()))
        );
    }
}
```

**Готово, если:** в логах появляются строки о переходах CB после серии ошибок/медленных вызовов.

---

## Задание 10: Кастомный “ключ” для rate limiting (простая сегментация)

**Цель:** базово ограничивать по “клиенту” (например, по заголовку `X-Api-Key`) — без сложных gateway’ев.

Resilience4j `@RateLimiter` из коробки ограничивает “по экземпляру” (по name), а не по ключу клиента. На базовом уровне сделаем простой вариант: **разные limiter instances** по ключам (например, free vs premium). Это не идеальная multi-tenant реализация, но хороший учебный шаг.

```yaml
resilience4j:
  ratelimiter:
    instances:
      freeLimiter:
        limitForPeriod: 2
        limitRefreshPeriod: 10s
        timeoutDuration: 0
      premiumLimiter:
        limitForPeriod: 10
        limitRefreshPeriod: 10s
        timeoutDuration: 0
```
### PlanBasedRateLimitedController. Java
```java
package com.example.demo.ratelimit;

import io.github.resilience4j.ratelimiter.RateLimiterRegistry;
import io.github.resilience4j.ratelimiter.RequestNotPermitted;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/plan")
public class PlanBasedRateLimitedController {

    private final RateLimiterRegistry registry;

    public PlanBasedRateLimitedController(RateLimiterRegistry registry) {
        this.registry = registry;
    }

    @GetMapping("/resource")
    public ResponseEntity<String> resource(@RequestHeader(value = "X-Plan", defaultValue = "free") String plan) {
        var limiterName = plan.equalsIgnoreCase("premium") ? "premiumLimiter" : "freeLimiter";
        var limiter = registry.rateLimiter(limiterName);

        // “ручное” применение лимитера
        if (!limiter.acquirePermission()) {
            throw new RequestNotPermitted(limiter);
        }

        return ResponseEntity.ok("OK plan=" + plan);
    }

    @ExceptionHandler(RequestNotPermitted.class)
    public ResponseEntity<String> tooMany(RequestNotPermitted e) {
        return ResponseEntity.status(429).body("TOO_MANY_REQUESTS limiter=" + e.getRateLimiterName());
    }
}
```

**Готово, если:**
- Без заголовка `X-Plan` лимит маленький (2/10s).
- С `X-Plan: premium` лимит больше (10/10s).

---

### Как быстро прогнать руками (шпаргалка curl)

1) Rate limit:
```bash
for i in {1..7}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/limited/hello; done
```

2) Circuit breaker (поднять вероятность падения):
```bash
for i in {1..20}; do curl -s http://localhost:8080/api/proxy/random-fail?p=0.9; echo; done
```

3) Slow calls:
```bash
for i in {1..15}; do curl -s http://localhost:8080/api/proxy/slow?ms=300; echo; done
```

---


