# Bean Scopes

## **Задание 1: Singleton vs Prototype - демонстрация разницы**

**Задача:** Показать разное поведение singleton и prototype scope.

**Решение:**
```java
@Component
@Scope("singleton")
public class SingletonService {
    private int count = 0;
    
    public int getAndIncrement() {
        return count++;
    }
}

@Component
@Scope("prototype")
public class PrototypeService {
    private int count = 0;
    
    public int getAndIncrement() {
        return count++;
    }
}

// Тест
@SpringBootTest
class ScopeTest {
    @Autowired SingletonService singleton1;
    @Autowired SingletonService singleton2;
    @Autowired PrototypeService prototype1;
    @Autowired PrototypeService prototype2;
    
    @Test
    void testScopes() {
        System.out.println("Singleton1: " + singleton1.getAndIncrement()); // 0
        System.out.println("Singleton2: " + singleton2.getAndIncrement()); // 1
        System.out.println("Prototype1: " + prototype1.getAndIncrement()); // 0
        System.out.println("Prototype2: " + prototype2.getAndIncrement()); // 0
    }
}
```

---

## **Задание 2: Проблема - внедрение prototype в singleton**

**Задача:** Показать проблему, когда prototype ведет себя как singleton.

**Решение:**
```java
@Component
@Scope("singleton")
public class ShoppingCartService {
    @Autowired
    private CartItem cartItem; // Prototype бин!
    
    public void addItem() {
        System.out.println("CartItem: " + cartItem); // Всегда один объект!
    }
}

@Component
@Scope("prototype")
public class CartItem {
    private final UUID id = UUID.randomUUID();
}

// Проблема: CartItem всегда один и тот же, хотя мы хотим новый при каждом вызове
```

---

## **Задание 3: Решение через ScopedProxyMode**

**Задача:** Использовать прокси для создания нового prototype при каждом обращении.

**Решение:**
```java
// Вариант 1: @Scope с proxyMode
@Component
@Scope(value = "prototype", proxyMode = ScopedProxyMode.TARGET_CLASS)
public class CartItem {
    private final UUID id = UUID.randomUUID();
}

// Вариант 2: XML/Java Config аналог
@Configuration
public class ProxyConfig {
    @Bean
    @Scope(proxyMode = ScopedProxyMode.INTERFACES)
    public CartItem cartItem() {
        return new CartItem();
    }
}

// Теперь ShoppingCartService будет получать новый CartItem при каждом обращении
```

---

## **Задание 4: Решение через ApplicationContext**

**Задача:** Получать prototype бин напрямую из контекста.

**Решение:**
```java
@Component
@Scope("singleton")
public class ShoppingCartService {
    @Autowired
    private ApplicationContext context;
    
    public void addItem() {
        CartItem newItem = context.getBean(CartItem.class);
        System.out.println("New CartItem: " + newItem);
    }
}
```

---

## **Задание 5: Решение через ObjectFactory/Provider**

**Задача:** Использовать ObjectFactory или Provider для ленивого получения бина.

**Решение:**
```java
@Component
@Scope("singleton")
public class ShoppingCartService {
    // Вариант 1: ObjectFactory
    @Autowired
    private ObjectFactory<CartItem> cartItemFactory;
    
    // Вариант 2: Provider (JSR-330)
    @Autowired
    private Provider<CartItem> cartItemProvider;
    
    public void addItem() {
        CartItem item1 = cartItemFactory.getObject();
        CartItem item2 = cartItemProvider.get();
        System.out.println("Different items: " + (item1 != item2)); // true
    }
}
```

---

## **Задание 6: Решение через @Lookup**

**Задача:** Использовать lookup method injection.

**Решение:**
```java
@Component
@Scope("singleton")
public abstract class ShoppingCartService {
    
    public void addItem() {
        CartItem item = createCartItem();
        System.out.println("New CartItem: " + item);
    }
    
    @Lookup
    protected abstract CartItem createCartItem();
}
// Spring создаст подкласс с реализацией метода createCartItem()
```

---

## **Задание 7: Request и Session scope в веб-приложении**

**Задача:** Создать бины с request и session scope.

**Решение:**
```java
// 1. Конфигурация
@Configuration
@EnableWebMvc
public class WebConfig implements WebMvcConfigurer {
    // Включение request/session scope
}

// 2. Request-scoped бин (новый для каждого HTTP запроса)
@Component
@Scope(value = WebApplicationContext.SCOPE_REQUEST, 
       proxyMode = ScopedProxyMode.TARGET_CLASS)
public class RequestData {
    private final String requestId = UUID.randomUUID().toString();
    private final Instant created = Instant.now();
}

// 3. Session-scoped бин (один на сессию пользователя)
@Component
@Scope(value = WebApplicationContext.SCOPE_SESSION,
       proxyMode = ScopedProxyMode.INTERFACES)
public class UserSession {
    private String userId;
    private List<String> cartItems = new ArrayList<>();
    
    public void addToCart(String item) {
        cartItems.add(item);
    }
}

// 4. Контроллер
@RestController
public class CartController {
    @Autowired private RequestData requestData;
    @Autowired private UserSession userSession;
    
    @PostMapping("/add")
    public String addToCart(@RequestParam String item) {
        userSession.addToCart(item);
        return "Added in request: " + requestData.getRequestId();
    }
}
```

## **Задание 8: Кастомный scope "TwoPhase"**

**Задача:** Реализовать собственный scope с переключением между двумя фазами.

**Шаги выполнения:**

1. **Реализуем интерфейс `Scope`:**
    - Создаем два Map для хранения бинов разных фаз
    - Реализуем метод `get()`, который возвращает бин из текущей фазы
    - Добавляем логику переключения между фазами

2. **Регистрируем кастомный scope:**
    - Создаем конфигурационный класс
    - Регистрируем scope через `CustomScopeConfigurer`
    - Задаем имя scope как "twoPhase"

3. **Создаем бин с кастомным scope:**
    - Аннотируем класс `@Scope("twoPhase")`
    - Добавляем счетчик для отслеживания созданных экземпляров

4. **Тестируем работу:**
    - Получаем бин несколько раз
    - Переключаем фазу и снова получаем бин
    - Наблюдаем создание разных экземпляров для разных фаз

```java
// Шаг 1: Реализуем кастомный scope
public class TwoPhaseScope implements Scope {
    private Map<String, Object> phase1 = new HashMap<>();
    private Map<String, Object> phase2 = new HashMap<>();
    private boolean usePhase1 = true;
    
    @Override
    public Object get(String name, ObjectFactory<?> factory) {
        Map<String, Object> current = usePhase1 ? phase1 : phase2;
        if (!current.containsKey(name)) {
            current.put(name, factory.getObject());
            System.out.println("Создан бин в фазе " + (usePhase1 ? "1" : "2"));
        }
        return current.get(name);
    }
    
    public void switchPhase() { 
        usePhase1 = !usePhase1; 
        System.out.println("Переключено на фазу " + (usePhase1 ? "1" : "2"));
    }
    // остальные методы интерфейса...
}

// Шаг 2: Регистрируем scope
@Configuration
public class ScopeConfig {
    @Bean
    public static CustomScopeConfigurer scopeConfigurer() {
        CustomScopeConfigurer c = new CustomScopeConfigurer();
        c.addScope("twoPhase", new TwoPhaseScope());
        return c;
    }
}

// Шаг 3: Создаем бин с кастомным scope
@Component 
@Scope("twoPhase")
public class TwoPhaseBean {
    private static int counter = 0;
    private final int id = ++counter;
    
    public int getId() { return id; }
}
```
