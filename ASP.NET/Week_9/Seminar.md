# Семинар 9 — JWT, policies и безопасное логирование

## Задание 1. Password hashing

Используйте `IPasswordHasher<AppUser>`:

1. хешируйте один пароль дважды;
2. убедитесь, что encoded hashes отличаются;
3. оба проходят verification;
4. raw password нигде не сохраняется/логируется.

Объясните роль salt и embedded format parameters.

## Задание 2. Login без enumeration

Endpoint принимает login/password. Для unknown user и wrong password возвращайте одинаковый public response и близкое поведение. Добавьте rate limit позже.

## Задание 3. JWT generation/validation

В учебной локальной схеме выпустите access token с `sub`, `iss`, `aud`, `exp`, `scope`, unique token id. Ключ получите из secret. Затем настройте JwtBearer validation с точным issuer/audience/algorithm.

Никогда не используйте демонстрационный symmetric key в production.

## Задание 4. Policies

Создайте:

- `orders.read`;
- `orders.write`;
- `orders.cancel`.

Защитите endpoints. Проверьте 401 без token и 403 с token без scope.

## Задание 5. Resource authorization

Только owner или support role может читать order. Только owner может отменить Draft/Paid по domain policy. Напишите `AuthorizationHandler` и unit tests для actor/resource combinations.

## Задание 6. Refresh sessions — дизайн

Без полного production auth реализуйте data model:

```text
session_id, user_id, refresh_hash, expires_at,
rotated_to, revoked_at, created_ip_hash, user_agent_hash
```

Опишите rotation transaction и reuse detection. В базе храните hash refresh token, не raw credential.

## Задание 7. Logging scope

Добавьте middleware scope с TraceId и после authentication — safe UserId. Убедитесь, что Authorization header/token отсутствуют в logs.

## Задание 8. Security log review

Создайте redaction checklist и проверьте:

- login DTO;
- HTTP logging middleware;
- exception handler;
- EF command logging;
- outbound HTTP logging.

## Проверка

- password verification через standard hasher;
- token validation строгая;
- 401/403 различаются;
- resource authorization server-side;
- refresh token не хранится открыто;
- logs содержат trace/user/order ids без credentials/PII payload.

