# Экзаменационные вопросы по ASP.NET Core и EF Core

## Блок 1. Host, DI и конфигурация

1. Framework против library. Inversion of control в ASP.NET Core.
2. Generic Host, Kestrel, middleware и endpoint model.
3. `IServiceCollection`, service provider и composition root.
4. Constructor injection, несколько implementations и keyed services.
5. Transient, Scoped, Singleton: creation, ownership и disposal.
6. Captive dependency и scope validation.
7. Manual scope, background singleton и scoped DbContext.
8. Decorator, middleware, filter, interceptor/proxy: критерии выбора.
9. Configuration providers и precedence.
10. Options, Snapshot, Monitor, named options и validation.
11. Environments, secrets и reloadable configuration.
12. Layered architecture, modular monolith и dependency direction.
13. `IHostedService`, startup, readiness и graceful shutdown.

## Блок 2. HTTP и Web API

14. Путь запроса через Kestrel и middleware pipeline.
15. Middleware order, short-circuit и response path.
16. Routing, route constraints и model binding sources.
17. Controllers против Minimal APIs.
18. DTO, entity, over-posting и mapping.
19. Syntactic, DTO, domain и database validation.
20. Upload/download, body buffering и size limits.
21. Cookies, session и distributed state.
22. HTTP status semantics: 201, 204, 400, 401, 403, 404, 409, 412, 429.
23. Problem Details и exception handling boundary.
24. Middleware против MVC exception filter.
25. CORS, preflight, credentials и ограничения защиты.
26. OpenAPI generation, contract testing и API evolution.
27. ETag/If-Match и связь с optimistic concurrency.

## Блок 3. Базы данных

28. ADO.NET abstractions и Npgsql provider.
29. Connection pooling, short-lived connection и exhaustion.
30. Parameterized SQL и невозможность parameterize identifiers.
31. DataReader lifetime и mapping.
32. Transactions, isolation и retry whole transaction.
33. Unknown outcome, idempotency и external call внутри transaction.
34. EF Core model и DbContext unit of work.
35. ChangeTracker states и identity map.
36. Attributes против Fluent configuration.
37. Relations, FK, navigation и cascades.
38. Tracking/no-tracking и projection.
39. Migrations и безопасный deployment.
40. N+1, lazy/eager/explicit loading.
41. Single/split query и cartesian explosion.
42. Offset/keyset pagination и indexes.
43. EF optimistic concurrency и resolution strategies.
44. Set-based updates и stale tracked state.

## Блок 4. Security и resilience

45. Authentication schemes, handlers и middleware order.
46. ClaimsPrincipal, issuer/subject и permissions.
47. Policy/resource-based authorization.
48. Password hashing, salt, cost, pepper и Identity.
49. JWT validation: signature, issuer, audience, lifetime, algorithm.
50. Access/refresh token, rotation, reuse detection и revocation.
51. Cookie/CSRF, bearer/XSS и роль CORS.
52. Structured logs, scopes, audit и redaction.
53. `IHttpClientFactory`, handlers и typed clients.
54. URI/header/body/error response contracts.
55. Timeout против caller cancellation.
56. Retry safety, idempotency key и jitter.
57. Circuit breaker states и policy scope.
58. Rate limiter, concurrency limiter и distributed replicas.

## Блок 5. Testing и async

59. Unit, integration, contract и E2E tests.
60. State vs interaction testing.
61. `WebApplicationFactory` и замена registrations.
62. Test authentication против real JWT validation tests.
63. Почему EF InMemory/SQLite не заменяют PostgreSQL tests.
64. Testcontainers lifecycle и isolation.
65. TimeProvider и deterministic async tests.
66. ASP.NET Core request/thread/async model.
67. ThreadPool starvation и blocking-over-async.
68. Request cancellation и partial side effects.
69. Streaming response/request, NDJSON, SSE и `Response.HasStarted`.
70. Backpressure и bounded channels.
71. CPU-bound work, background queue и 202 operation resource.

## Блок 6. Messaging и observability

72. Durable job, lease, retry и poison handling.
73. Transactional outbox и причина duplicates.
74. Inbox/idempotent consumer и unknown outcome.
75. Kafka topic, partition, key, offset и ordering.
76. Consumer group, rebalance и parallelism.
77. Producer acknowledgements, idempotence, batching и durability.
78. Commit before/after processing: delivery semantics.
79. Retry topic, blocking retry, DLQ и ordering.
80. Schema evolution и integration event contracts.
81. Kafka transactions против PostgreSQL atomicity.
82. Logs, metrics и traces: роли.
83. Четыре золотых сигнала и SLO.
84. Counter, gauge, histogram, summary и percentiles.
85. Metric cardinality и route templates.
86. Activity/OpenTelemetry и context propagation.
87. Liveness против readiness.
88. Prometheus/Grafana, alerts и telemetry overhead.

