# Logger

## Задание 1: Базовый логгер в сервисе
**Цель:** научиться писать параметризованные логи через SLF4J.

**Сделать:**
- Создать `HelloService` с методом `greet(String name)`.
- Внутри логировать `INFO` вида: `Greeting requested: name={}` и `Greeting produced: message={}`.
- Создать контроллер `GET /api/hello?name=...`, который вызывает сервис.

**Готово, если:**
- В консоли видно 2 строки `INFO` на каждый запрос.
- Логи используют `{}` вместо конкатенации строк.

---

## Задание 2: Проверка уровней DEBUG/INFO
**Цель:** увидеть, как уровни влияют на вывод.

**Сделать:**
- Добавить в `HelloService` один `log.debug(...)`.
- В `application.yml` выставить:
    - `logging.level.root=INFO`
    - `logging.level.<ваш.пакет>=DEBUG`

```java
log.debug("Debug details: nameLength={}", name != null ? name.length() : 0);
```
**Готово, если:**
- DEBUG сообщения вашего пакета видны, а DEBUG от остальных библиотек — нет.

---

## Задание 3: Настроить pattern и добавить thread + logger
**Цель:** сделать лог читабельнее.

**Сделать:**
- В `application.yml` задать `logging.pattern.console`, чтобы в логе были:
    - Дата/время с миллисекундами
    - Уровень
    - Thread name
    - Logger name (укороченный)
    - Сообщение

```yaml
server:
  port: 8080

management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus
  endpoint:
    health:
      show-details: always

logging:
  level:
    root: INFO
    com.example.demo: DEBUG
    org.springframework: INFO
    org.hibernate.SQL: WARN
  pattern:
    console: "%d{yyyy-MM-dd HH:mm:ss.SSS} %-5level [%thread] %logger{36} trace=%X{traceId} - %msg%n"
```

**Готово, если:**
- Каждая строка лога содержит thread и имя логгера.

---

## Задание 4: Access log (метод, путь, статус, время)
**Цель:** минимальный access log через фильтр.

**Сделать:**
- Создать `OncePerRequestFilter`, который:
    - Замеряет время выполнения
    - после `filterChain.doFilter` пишет `INFO` вида:
      `HTTP {} {} -> status={} timeMs={}`

```java
package com.example.demo.logging;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

@Component
public class RequestTimingFilter extends OncePerRequestFilter {

    private static final Logger log = LoggerFactory.getLogger(RequestTimingFilter.class);

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {

        long start = System.currentTimeMillis();
        try {
            filterChain.doFilter(request, response);
        } finally {
            long ms = System.currentTimeMillis() - start;
            log.info("HTTP {} {} -> status={} timeMs={}",
                    request.getMethod(),
                    request.getRequestURI(),
                    response.getStatus(),
                    ms);
        }
    }
}
```
**Готово, если:**
- На любой запрос появляется одна строка access log.

---

## Задание 5: MDC traceId (корреляция логов)
**Цель:** связать события одного запроса.

**Сделать:**
- Создать фильтр `TraceIdFilter`:
    - берёт `X-Trace-Id` из заголовка или генерирует UUID
    - кладёт в `MDC.put("traceId", ...)`
    - добавляет header `X-Trace-Id` в ответ
- Обновить pattern так, чтобы был `trace=%X{traceId}`.

```java 
package com.example.demo.logging;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.Optional;
import java.util.UUID;

@Component
public class TraceIdFilter extends OncePerRequestFilter {

    private static final String TRACE_ID = "traceId";
    private static final String HEADER = "X-Trace-Id";

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {

        String traceId = Optional.ofNullable(request.getHeader(HEADER))
                .filter(s -> !s.isBlank())
                .orElse(UUID.randomUUID().toString());

        MDC.put(TRACE_ID, traceId);
        response.setHeader(HEADER, traceId);

        try {
            filterChain.doFilter(request, response);
        } finally {
            MDC.remove(TRACE_ID);
        }
    }
}
```

```java 
// TraceIdFilter
@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class TraceIdFilter extends OncePerRequestFilter { ... }
```

```java
// RequestTimingFilter
@Component
@Order(Ordered.HIGHEST_PRECEDENCE + 10)
public class RequestTimingFilter extends OncePerRequestFilter { ... }
```
**Готово, если:**
- При запросе без заголовка сервер возвращает `X-Trace-Id`.
- В access log и в логах контроллера/сервиса присутствует один и тот же `traceId`.

---

## Задание 6: Понизить шумные логгеры
**Цель:** научиться “успокаивать” библиотеки.

**Сделать:**
- Включить `logging.level.org.springframework=INFO`
- Включить `logging.level.org.hibernate.SQL=WARN` (даже если Hibernate нет — просто как пример)
- Проверить, что логи Spring не стали “простынёй”.

```java 
logging:
  level:
    org.springframework: INFO
    org.hibernate.SQL: WARN
```
**Готово, если:**
- Вы понимаете, какие пакеты стали тише/громче (и видите это по выводу).
---

# Passwords
---

## Общие предпосылки (минимальные зависимости)

```gradle
dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.boot:spring-boot-starter-security'

    // JWT (jjwt)
    implementation 'io.jsonwebtoken:jjwt-api:0.12.6'
    runtimeOnly 'io.jsonwebtoken:jjwt-impl:0.12.6'
    runtimeOnly 'io.jsonwebtoken:jjwt-jackson:0.12.6'
}
```

---

## Задание PWD-1: Регистрация с BCrypt и проверкой пароля
**Цель:** хранить только хеш пароля и уметь проверять пароль.

### Сделать
1) Создать “пользователей” в памяти (Map).
2) Реализовать регистрацию `POST /auth/register` (username + password).
3) Пароль сохранять только как `BCrypt` hash.
4) Добавить сервис проверки пароля `POST /auth/check` (username + password) → `true/false`.

### Код, который должен получиться

#### PasswordConfig. Java
```java
package com.example.demo.auth;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

@Configuration
class PasswordConfig {
    @Bean
    PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder(); // можно BCryptPasswordEncoder(12)
    }
}
```

#### UserRecord. Java
```java
package com.example.demo.auth;

public record UserRecord(String username, String passwordHash) {}
```

#### InMemoryUserStore. Java
```java
package com.example.demo.auth;

import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Component
class InMemoryUserStore {
    private final Map<String, UserRecord> users = new ConcurrentHashMap<>();

    public Optional<UserRecord> find(String username) {
        return Optional.ofNullable(users.get(username));
    }

    public void save(UserRecord user) {
        users.put(user.username(), user);
    }

    public boolean exists(String username) {
        return users.containsKey(username);
    }
}
```

#### DTO
```java
package com.example.demo.auth;

public record RegisterRequest(String username, String password) {}
public record CheckPasswordRequest(String username, String password) {}
public record CheckPasswordResponse(boolean valid) {}
```

#### AuthPasswordService. Java
```java
package com.example.demo.auth;

import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
class AuthPasswordService {

    private final InMemoryUserStore store;
    private final PasswordEncoder encoder;

    AuthPasswordService(InMemoryUserStore store, PasswordEncoder encoder) {
        this.store = store;
        this.encoder = encoder;
    }

    public void register(String username, String rawPassword) {
        if (username == null || username.isBlank()) throw new IllegalArgumentException("username required");
        if (rawPassword == null || rawPassword.isBlank()) throw new IllegalArgumentException("password required");
        if (store.exists(username)) throw new IllegalArgumentException("user already exists");

        String hash = encoder.encode(rawPassword);
        store.save(new UserRecord(username, hash));
    }

    public boolean checkPassword(String username, String rawPassword) {
        UserRecord user = store.find(username).orElseThrow(() -> new IllegalArgumentException("user not found"));
        return encoder.matches(rawPassword, user.passwordHash());
    }
}
```

#### AuthPasswordController. Java
```java
package com.example.demo.auth;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/auth")
class AuthPasswordController {

    private final AuthPasswordService service;

    AuthPasswordController(AuthPasswordService service) {
        this.service = service;
    }

    @PostMapping("/register")
    public ResponseEntity<Void> register(@RequestBody RegisterRequest req) {
        service.register(req.username(), req.password());
        return ResponseEntity.ok().build();
    }

    @PostMapping("/check")
    public CheckPasswordResponse check(@RequestBody CheckPasswordRequest req) {
        return new CheckPasswordResponse(service.checkPassword(req.username(), req.password()));
    }
}
```

#### SecurityConfig (разрешим /auth/** без логина)
```java
package com.example.demo.security;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.Customizer;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.web.SecurityFilterChain;

@Configuration
class SecurityConfig {
    @Bean
    SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        return http
                .csrf(csrf -> csrf.disable())
                .authorizeHttpRequests(auth -> auth
                        .requestMatchers("/auth/**").permitAll()
                        .anyRequest().authenticated()
                )
                .httpBasic(Customizer.withDefaults())
                .build();
    }
}
```

### Готово, если
- В хранилище у пользователя лежит не пароль, а строка вида `$2a$...`.
- `/auth/check` возвращает `true` только для верного пароля.
## Задание PWD-2: Смена пароля с проверкой старого и запретом повторного
**Цель:** правильно менять пароль: проверить текущий, захешировать новый, обновить.

### Сделать
1) Добавить endpoint `POST /auth/change-password`
2) Требования:
    - Пользователь существует
    - Старый пароль верный
    - новый пароль не равен старому (через `matches`)
    - Сохранить новый хеш

### Код

#### DTO
```java
package com.example.demo.auth;

public record ChangePasswordRequest(String username, String oldPassword, String newPassword) {}
```

#### Дополнение в AuthPasswordService
```java
public void changePassword(String username, String oldPassword, String newPassword) {
    UserRecord user = store.find(username).orElseThrow(() -> new IllegalArgumentException("user not found"));

    if (!encoder.matches(oldPassword, user.passwordHash())) {
        throw new IllegalArgumentException("old password invalid");
    }
    if (encoder.matches(newPassword, user.passwordHash())) {
        throw new IllegalArgumentException("new password must be different");
    }

    store.save(new UserRecord(username, encoder.encode(newPassword)));
}
```

#### Дополнение в контроллер
```java
@PostMapping("/change-password")
public ResponseEntity<Void> changePassword(@RequestBody ChangePasswordRequest req) {
    service.changePassword(req.username(), req.oldPassword(), req.newPassword());
    return ResponseEntity.ok().build();
}
```

### Готово, если
- При неверном старом пароле смена не проходит.
- Если новый пароль такой же — смена не проходит.
- После смены `/auth/check` работает только с новым паролем.

---


Ниже 2 базовых задания про **соль** и **перец** (salt & pepper) в паролях, в формате лабораторной: цель → что сделать → код → критерии готовности. Я опираюсь на Spring Security `PasswordEncoder` и BCrypt.

Ключевая идея:
- **Соль (salt)** — уникальная для каждого пароля, хранится вместе с хешем. В BCrypt соль встроена в итоговую строку хеша, отдельно хранить её не надо.
- **Перец (pepper)** — общий секрет приложения (как ключ), хранится **не в БД**, а в конфигурации/секрет-хранилище. Его добавляют к паролю перед хешированием/проверкой.

---

## Задание SALT-1: Доказать, что BCrypt “солит” сам (разные хеши для одного пароля)

### Цель
Убедиться, что при одинаковом пароле BCrypt создаёт **разные хеши** (за счёт соли), и оба хеша проходят `matches()`.

### Что сделать
1) Создать endpoint `GET /lab/salt?password=...`
2) Он должен:
    - дважды захешировать один и тот же пароль через `BCryptPasswordEncoder.encode`
    - вернуть JSON с `hash1`, `hash2`, `equalHashes`, `matches1`, `matches2`

### Код

### PasswordConfig. Java
```java
package com.example.demo.auth;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

@Configuration
class PasswordConfig {
    @Bean
    PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }
}
```

### SaltLabController. Java
```java
package com.example.demo.lab;

import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/lab")
class SaltLabController {

    private final PasswordEncoder encoder;

    SaltLabController(PasswordEncoder encoder) {
        this.encoder = encoder;
    }

    record SaltLabResponse(String hash1, String hash2, boolean equalHashes, boolean matches1, boolean matches2) {}

    @GetMapping("/salt")
    public SaltLabResponse salt(@RequestParam String password) {
        String hash1 = encoder.encode(password);
        String hash2 = encoder.encode(password);

        boolean equal = hash1.equals(hash2); // почти всегда false
        boolean matches1 = encoder.matches(password, hash1);
        boolean matches2 = encoder.matches(password, hash2);

        return new SaltLabResponse(hash1, hash2, equal, matches1, matches2);
    }
}
```

### Критерии готовности
- `equalHashes = false` (в большинстве запусков; фактически должен быть false всегда, если BCrypt корректный).
- `matches1 = true` и `matches2 = true`.
- Понимаете и можете объяснить: “соль хранится в строке BCrypt-хеша”.

---

## Задание PEPPER-1: Добавить “перец” (pepper) поверх BCrypt через делегирующий PasswordEncoder

### Цель
Реализовать pepper так, чтобы:
- при регистрации пароль хешировался как `BCrypt( password + pepper )`
- При проверке использовался тот же pepper
- pepper **не хранится** в БД, а берётся из `application.yml` (или env)

> Важно: pepper — это секрет. В реальном проде хранится в Vault/KMS/секретах окружения, а не в репозитории.

### Что сделать
1) Добавить в конфиг `security.password.pepper`
2) Реализовать `PepperPasswordEncoder`, который оборачивает BCrypt:
    - `encode(raw) -> delegate.encode(raw + pepper)`
    - `matches(raw, encoded) -> delegate.matches(raw + pepper, encoded)`
3) Подключить его как `PasswordEncoder` бин.
4) Проверить:
    - Регистрация создаёт хеш
    - Проверка пароля работает
    - Если изменить pepper в конфиге — проверка старых паролей перестанет проходить

### Код
```yaml
security:
  password:
    pepper: "CHANGE_ME_PEPPER_STORE_IN_ENV_OR_VAULT"
```

### PasswordProps. Java
```java
package com.example.demo.auth;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "security.password")
public record PasswordProps(String pepper) {}
```

### PasswordPropsConfig. Java
```java
package com.example.demo.auth;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@EnableConfigurationProperties(PasswordProps.class)
class PasswordPropsConfig {}
```

### PepperPasswordEncoder. Java
```java
package com.example.demo.auth;

import org.springframework.security.crypto.password.PasswordEncoder;

import java.util.Objects;

public class PepperPasswordEncoder implements PasswordEncoder {

    private final PasswordEncoder delegate;
    private final String pepper;

    public PepperPasswordEncoder(PasswordEncoder delegate, String pepper) {
        this.delegate = Objects.requireNonNull(delegate);
        this.pepper = Objects.requireNonNull(pepper);
    }

    @Override
    public String encode(CharSequence rawPassword) {
        return delegate.encode(rawPasswordWithPepper(rawPassword));
    }

    @Override
    public boolean matches(CharSequence rawPassword, String encodedPassword) {
        return delegate.matches(rawPasswordWithPepper(rawPassword), encodedPassword);
    }

    private String rawPasswordWithPepper(CharSequence rawPassword) {
        // важно: pepper добавляем, но в логи никогда не пишем
        return rawPassword + pepper;
    }
}
```

### PasswordConfig. Java (обновлённая)
```java
package com.example.demo.auth;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

@Configuration
class PasswordConfig {

    @Bean
    PasswordEncoder passwordEncoder(PasswordProps props) {
        PasswordEncoder bcrypt = new BCryptPasswordEncoder();
        return new PepperPasswordEncoder(bcrypt, props.pepper());
    }
}
```

### (Используйте Ваш сервис регистрации/проверки)
Если у вас был сервис из прошлых заданий (register/check), код менять не нужно — он уже использует `PasswordEncoder`, а значит pepper включится автоматически.

### Критерии готовности
- После регистрации пароль проверяется успешно.
- Хеш в “БД” выглядит как BCrypt (`$2a$...`), pepper нигде не хранится.
- При смене pepper в `application.yml` (и перезапуске) проверка старого пароля **ломается** (это ожидаемо и важно понимать как операционный риск).

---

---
# JWT

## Общие классы для JWT (подготовка)

### JwtProperties. Java
```java
package com.example.demo.jwt;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "security.jwt")
public record JwtProperties(String secret, String issuer, long accessTtlSeconds) {}
```

Важно: secret должен быть достаточно длинным (для HS256 желательно 32+ байта).
```yaml
security:
  jwt:
    secret: "CHANGE_ME_CHANGE_ME_CHANGE_ME_CHANGE_ME_32bytes"
    issuer: "demo-app"
    access-ttl-seconds: 900
```

### JwtConfig. Java
```java
package com.example.demo.jwt;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@EnableConfigurationProperties(JwtProperties.class)
class JwtConfig {}
```

---

## Задание JWT-1: JwtService (создание и проверка токена)
**Цель:** научиться генерировать JWT и валидировать подпись/exp/iss.

### Код: JwtService. Java
```java
package com.example.demo.jwt;

import io.jsonwebtoken.*;
import io.jsonwebtoken.security.Keys;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Date;
import java.util.Map;

@Service
public class JwtService {

    private final JwtProperties props;
    private final SecretKey key;

    public JwtService(JwtProperties props) {
        this.props = props;
        this.key = Keys.hmacShaKeyFor(props.secret().getBytes(StandardCharsets.UTF_8));
    }

    public String generateAccessToken(String username, Map<String, Object> claims) {
        Instant now = Instant.now();
        Instant exp = now.plusSeconds(props.accessTtlSeconds());

        return Jwts.builder()
                   .issuer(props.issuer())
                   .subject(username)
                   .issuedAt(Date.from(now))
                   .expiration(Date.from(exp))
                   .claims(claims)
                   .signWith(key, Jwts.SIG.HS256)
                   .compact();
    }

    public Jws<Claims> parseAndValidate(String jwt) {
        JwtParser parser = Jwts.parser()
                               .verifyWith(key)
                               .requireIssuer(props.issuer())
                               .build();

        return parser.parseSignedClaims(jwt);
    }

    public String extractUsername(String jwt) {
        return parseAndValidate(jwt).getPayload().getSubject();
    }
}
```

### Готово, если
- Токен создаётся.
- При изменении одного символа токен перестаёт валидироваться (SignatureException).

---

## Задание JWT-2: /auth/login возвращает JWT по паролю
**Цель:** связать проверку пароля (из PWD-1) и выдачу JWT.

### Сделать
- Endpoint `POST /auth/login` принимает username/password.
- Если пароль верный — вернуть JSON с `accessToken`.

### Код

#### DTO
```java
package com.example.demo.jwt;

public record LoginRequest(String username, String password) {}
public record TokenResponse(String accessToken) {}
```

#### AuthJwtController. Java
```java
package com.example.demo.jwt;

import com.example.demo.auth.AuthPasswordService;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/auth")
class AuthJwtController {

    private final AuthPasswordService passwordService;
    private final JwtService jwtService;

    AuthJwtController(AuthPasswordService passwordService, JwtService jwtService) {
        this.passwordService = passwordService;
        this.jwtService = jwtService;
    }

    @PostMapping("/login")
    public TokenResponse login(@RequestBody LoginRequest req) {
        boolean ok = passwordService.checkPassword(req.username(), req.password());
        if (!ok) throw new IllegalArgumentException("invalid credentials");

        // минимальный пример claims
        String token = jwtService.generateAccessToken(req.username(), Map.of("roles", "USER"));
        return new TokenResponse(token);
    }
}
```

### Готово, если
- После регистрации можно залогиниться и получить JWT.

---

## Задание JWT-3: JwtAuthFilter — принимать JWT в Authorization: Bearer
**Цель:** на каждом запросе (кроме /auth/**) доставать JWT, валидировать, выставлять Authentication.

### Код: JwtAuthenticationFilter. Java
```java
package com.example.demo.jwt;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jws;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.HttpHeaders;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.List;

@Component
public class JwtAuthenticationFilter extends OncePerRequestFilter {

    private final JwtService jwtService;

    public JwtAuthenticationFilter(JwtService jwtService) {
        this.jwtService = jwtService;
    }

    @Override
    protected boolean shouldNotFilter(HttpServletRequest request) {
        return request.getRequestURI().startsWith("/auth/");
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
        throws ServletException, IOException {

        String header = request.getHeader(HttpHeaders.AUTHORIZATION);
        if (header == null || !header.startsWith("Bearer ")) {
            filterChain.doFilter(request, response);
            return;
        }

        String token = header.substring("Bearer ".length()).trim();

        try {
            Jws<Claims> jws = jwtService.parseAndValidate(token);
            String username = jws.getPayload().getSubject();

            // упрощённо: берём role из claims
            String roles = jws.getPayload().get("roles", String.class);
            List<SimpleGrantedAuthority> authorities =
                roles == null ? List.of() : List.of(new SimpleGrantedAuthority("ROLE_" + roles));

            var auth = new UsernamePasswordAuthenticationToken(username, null, authorities);
            SecurityContextHolder.getContext().setAuthentication(auth);

        } catch (Exception e) {
            // токен плохой — чистим контекст и отдаём 401
            SecurityContextHolder.clearContext();
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            return;
        }

        filterChain.doFilter(request, response);
    }
}
```

### Подключить фильтр в SecurityConfig
```java
package com.example.demo.security;

import com.example.demo.jwt.JwtAuthenticationFilter;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

@Configuration
class SecurityConfig {

    @Bean
    SecurityFilterChain securityFilterChain(HttpSecurity http,
                                            JwtAuthenticationFilter jwtFilter) throws Exception {
        return http
                   .csrf(csrf -> csrf.disable())
                   .authorizeHttpRequests(auth -> auth
                                                      .requestMatchers("/auth/**").permitAll()
                                                      .anyRequest().authenticated()
                   )
                   // basic можно убрать, если чисто JWT
                   //.httpBasic(Customizer.withDefaults())
                   .addFilterBefore(jwtFilter, UsernamePasswordAuthenticationFilter.class)
                   .build();
    }
}
```

### Готово, если
- Без токена `/api/**` даёт 401/403 (в зависимости от конфигурации).
- С `Authorization: Bearer <token>` запрос проходит.

---

## Задание JWT-4: Защищённый endpoint + кто я (principal)
**Цель:** показать, что SecurityContext заполнен.

### Код: ProfileController. Java
```java
package com.example.demo.profile;

import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ProfileController {

    @GetMapping("/api/me")
    public String me(Authentication authentication) {
        return "You are: " + authentication.getName() + ", authorities=" + authentication.getAuthorities();
    }
}
```

### Готово, если
- С JWT `/api/me` возвращает username из `sub`.

---

## Задание JWT-5: Роли из JWT и ограничение доступа
**Цель:** использовать `GrantedAuthority` из JWT для авторизации.

### Сделать
1) Добавить два endpoint:
    - `/api/user` доступен ROLE_USER
    - `/api/admin` доступен ROLE_ADMIN
2) В логине для разных пользователей выдавать разные роли (например, если username == "admin" → ADMIN)

### Код

#### Роли при логине (обновите AuthJwtController)
```java
@PostMapping("/login")
public TokenResponse login(@RequestBody LoginRequest req) {
    boolean ok = passwordService.checkPassword(req.username(), req.password());
    if (!ok) throw new IllegalArgumentException("invalid credentials");

    String role = req.username().equalsIgnoreCase("admin") ? "ADMIN" : "USER";
    String token = jwtService.generateAccessToken(req.username(), Map.of("roles", role));
    return new TokenResponse(token);
    }
```

#### SecuredController. Java
```java
package com.example.demo.profile;

import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SecuredController {

    @PreAuthorize("hasRole('USER')")
    @GetMapping("/api/user")
    public String user() {
        return "user area";
    }

    @PreAuthorize("hasRole('ADMIN')")
    @GetMapping("/api/admin")
    public String admin() {
        return "admin area";
    }
}
```

#### Включить method security
```java
package com.example.demo.security;

import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;

@Configuration
@EnableMethodSecurity
class MethodSecurityConfig {}
```

### Готово, если
- Токен пользователя `USER` открывает `/api/user`, но запрещает `/api/admin`.
- Токен `admin` открывает оба.

---

## Быстрый чек-лист ручной проверки (curl)

1) Регистрация:
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret"}'
```

2) Логин → взять accessToken:
```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret"}'
```

3) Вызов защищённого:
```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/me
```

---
