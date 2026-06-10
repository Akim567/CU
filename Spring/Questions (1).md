## Блок 1.

1. Фреймворк. Отличие библиотеки от фреймворка. Преимущества и недостатки использования фреймворков. Что такое Spring. Что такое Spring Boot.
2. Инверсия управления. Внедрение зависимостей, типы: конструктор, поле, сеттер. @Autowired, @Qualifier. Конфигурирование: XML-based, annotation-based, java-based.
3. Виды архитектур на уровне приложения: слоистая, модульная, чистая.
4. AOP. Понятия: Join Point, Pointcut, Advice, Aspect.
5. AOP. Фреймворк Spring AOP и библиотеки CGLIB & JDK Dynamic proxies. Фреймворк AspectJ и понятие Weaving.
6. Bean. Жизненный цикл.
7. Bean. Стереотипные аннотации. Скоупы бинов. Управление бинами посредством: ObjectFactory, ObjectProvider, Provider.
8. Аннотации конфигураций: @Configuration, @ComponentScan, @Import, @Profile, @Primary, @DependsOn, @Order.
9. Properties. @Value. @PropertySource. @ConfigurationProperties. Файлы конфигурации application.properties и application.yaml.
10. Servlet. DispatcherServlet. Путь запроса в Spring MVC. Полный цикл обработки HTTP-запроса.
11. Контейнеры сервлетов: Tomcat, Jetty, Undertow, Netty.
12. Формирование ответов. ResponseEntity. Кастомные хедеры в ответах. CORS (cross-origin resource sharing).
13. @Controller, @RestController, @RequestMapping, @RequestBody, @ResponseBody, @PathVariable, @RequestParam. Обработка path- и query-параметров. Понятие DTO. Маппинг DTO на Entity. Валидация на уровне контроллеров.
14. Загрузка файлов: raw и MultipartFile. Работа с состоянием: заголовки, куки, сессии.
15. Обработка ошибок. @ControllerAdvice. @ExceptionHandler.
16. HTTP. REST. OpenAPI + Spring. Генерация контроллеров и клиентов по спецификации.
17. Логирование. SLF4J и реализации. Уровни логирования, конфигурация. MDC.
18. Тестирование. Unit-тестирование. Интеграционное тестирование. E2E-тестирование. @SpringBootTest, @WebMvcTest, @MockBean, @SpyBean, MockMvc. Testcontainers.

## Блок 2.

1. JDBC: Statement, PreparedStatement, ResultSet. Выполнение запросов. Транзакции. Driver.
2. Spring JDBC. Connection pool. DataSource. JdbcTemplate.
3. Императивные и декларативные транзакции в Spring JDBC. Уровни изоляции. Propagation.
4. Миграции: Flyway (Liquibase).
5. JPA и Hibernate. Принцип работы Hibernate. @Table, @Id, @GeneratedValue, @Column. JPA-репозитории в Spring.
6. Hibernate. Жизненный цикл сущности Hibernate.
7. Hibernate. Связи в сущностях и каскадные операции.
8. Hibernate. Производительность: кеширование, проблема N + 1.
9. Hibernate. Пагинация и сортировка.
10. Spring Data Repository. Добавление новых методов. JPQL, HQL.
11. Spring Security. Архитектура и конфигурирование. AuthenticationManager. PasswordEncoder. Аутентификация и авторизация. JWT.
12. RestClient. Составление HTTP-запросов. Отправка хедеров. Базовая авторизация, отправка хедеров авторизации. Обработка ответов и исключения.
13. Rate limiter и Circuit breaker. Resilience4j в Spring Boot.

## Блок 3.

1. Реактивное программирование. Project Reactor. Reactive Streams. Тестирование.
2. Реактивное программирование. Spring WebFlux. Netty и EventLoop. Тестирование.
3. Реактивное программирование. R2DBC. Проблема блокировки потоков в доступе к БД. Проблемы JDBC в реактивном стеке.
4. Реактивное программирование. Реактивные транзакции. Полный путь обработки реактивного запроса.
5. Apache Kafka. Понятие: topic, partition, offset, consumer group. Понятие Replication и In-Sync-Replicas (ISR). Отличие от других инструментов.
6. Kafka консьюмер. Архитектура консьюмера. Конфигурации консьюмера. Batch консьюмер. Увеличение пропускной способности.
7. Kafka продьюсер. Архитектура продьюсера. KafkaTemplate, RoutingKafkaTemplate, ReplyKafkaTemplate. Увеличение пропускной способности.
8. 4 золотых сигнала мониторинга. Типы метрик по Prometheus (Counter, Gauge, Histogram и Summary). Квантили и персентили.
9. Spring Actuator: эндпоинты, конфигурации. Micrometer: SDK, интеграция с Prometheus и Grafana. Производительность и оверхед при использовании метрик.