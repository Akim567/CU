# Лекция 9 — Authentication, authorization, Identity, JWT и logging

## 1. Authentication и authorization

- **Authentication** устанавливает identity: кто делает запрос.
- **Authorization** решает, разрешено ли identity выполнить operation над resource.

401 означает, что подходящая authentication отсутствует/неуспешна. 403 — identity установлена, но permission недостаточно.

## 2. Middleware order

```csharp
app.UseAuthentication();
app.UseAuthorization();
app.MapControllers();
```

Authentication создаёт `HttpContext.User`; authorization применяет policies/metadata endpoint.

## 3. Schemes и handlers

Scheme связывает имя с authentication handler/configuration. В одном приложении могут быть cookie, bearer, API key или external schemes. Не выбирайте scheme по недоверенному input без allow-list/policy scheme design.

Bearer registration conceptually:

```csharp
services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddJwtBearer(options =>
    {
        options.Authority = configuration["Auth:Authority"];
        options.Audience = "orders-api";
    });
```

Production validation должна проверять issuer, audience, signature, lifetime и допустимые algorithms/keys.

## 4. Claims и principal

`ClaimsPrincipal` содержит identities и claims. Claim — утверждение issuer, не автоматически истинный permission.

```csharp
string? subject = User.FindFirstValue("sub");
```

Не используйте display name/email как стабильный primary key. Subject + issuer обычно образуют более надёжную identity boundary.

## 5. Authorization policies

```csharp
services.AddAuthorization(options =>
{
    options.AddPolicy("orders.write", policy =>
        policy.RequireAuthenticatedUser()
              .RequireClaim("scope", "orders.write"));
});
```

```csharp
[Authorize(Policy = "orders.write")]
[HttpPost]
public Task<IActionResult> Create(...) { }
```

Role — один вид claim-based rule. Policies лучше выражают permissions и комбинированные requirements.

## 6. Resource-based authorization

Право может зависеть от самого order:

```csharp
AuthorizationResult allowed = await authorization.AuthorizeAsync(
    User,
    order,
    "orders.edit");
```

Handler сравнивает owner/tenant/status. Сначала нужно безопасно загрузить resource в tenant boundary, затем авторизовать operation. Иногда для предотвращения enumeration forbidden ресурс маскируют как 404 по policy.

## 7. Password storage

Пароль не шифруют обратимо и не хешируют быстрым SHA-256. Нужен password hashing/KDF с уникальной salt и cost: ASP.NET Core `PasswordHasher<TUser>`/Identity управляет форматом и migration rehash.

Pepper, если используется, хранится отдельно в secret manager и требует rotation strategy. Он не заменяет salt/KDF.

## 8. ASP.NET Core Identity

Identity предоставляет user/role stores, password hashing, tokens, lockout, confirmation и cookie integration. Он не является обязательным для API, использующего внешний OIDC provider.

Не создавайте самодельную authentication систему, если Identity/OIDC provider покрывает требования.

## 9. JWT

JWT — подписанный набор claims, не «зашифрованный токен» по умолчанию. Base64url payload читается клиентом.

Проверки:

- signature/key;
- `iss`;
- `aud`;
- `exp`/`nbf` с ограниченным clock skew;
- algorithm policy;
- key rotation;
- token type/use.

Access token короткоживущий. Refresh token — credential высокой ценности: хранить защищённо, rotation/reuse detection, revoke per session/device.

## 10. Revocation trade-off

Self-contained access JWT быстро проверяется локально, но плохо отзывается мгновенно. Варианты:

- короткий TTL;
- introspection/reference tokens;
- denylist с operational cost;
- session/version claim, проверяемый по store;
- key rotation для массового revoke с большим blast radius.

## 11. CSRF и XSS

- Cookie authentication автоматически отправляется browser и требует CSRF защиты для state-changing requests.
- Bearer token в Authorization header не отправляется автоматически, но XSS может украсть token из browser storage.
- CORS не является CSRF/XSS защитой целиком.

Выбор storage/flow делайте по browser architecture и threat model.

## 12. Structured logging

```csharp
logger.LogInformation(
    "User {UserId} created order {OrderId}",
    userId,
    orderId);
```

Уровни:

- Trace/Debug — диагностика высокой детализации;
- Information — значимое normal event;
- Warning — degraded/unexpected, но обработанное;
- Error — operation failed;
- Critical — process/system-level failure.

Не логируйте passwords, raw tokens, authorization headers, connection strings, payment data и полный personal payload.

## 13. Scopes и trace correlation

```csharp
using (logger.BeginScope(new Dictionary<string, object>
{
    ["OrderId"] = orderId,
    ["UserId"] = userId
}))
{
    await service.ProcessAsync(ct);
}
```

OpenTelemetry `Activity.TraceId` должен коррелировать logs/traces. Не создавайте новый correlation id на каждом layer.

## 14. Audit log

Audit отличается от diagnostic log:

- определённая schema;
- actor/action/resource/time/outcome;
- retention/access controls;
- tamper-evidence по требованиям;
- избегание sensitive payload;
- не терять запись молча.

Обычный application logger не всегда удовлетворяет compliance audit.

## 15. Сравнение со Spring Security

| Spring Security | ASP.NET Core |
|---|---|
| security filter chain | auth middleware + handlers/policies |
| Authentication | ClaimsPrincipal/AuthenticationTicket |
| `@PreAuthorize` | `[Authorize(Policy=...)]`/service authorization |
| UserDetailsService | Identity user store/custom service/provider |
| PasswordEncoder | PasswordHasher/Identity password services |
| MDC | logging scopes + Activity baggage/tags |

## 16. Checklist

- HTTPS required.
- Validate issuer/audience/signature/lifetime.
- Secrets outside source/logs.
- Permission policy, не только UI hiding.
- Resource/tenant authorization на server.
- Rate limit login/token endpoints.
- Lockout/brute-force monitoring без user enumeration.
- Logs structured и redacted.

## Самопроверка

1. Когда вернуть 401, а когда 403?
2. Почему JWT payload не секретен?
3. Зачем проверять audience?
4. Почему cookie auth требует CSRF consideration?
5. Чем audit log отличается от diagnostic log?

