# Семинар 13. Практикум по теме «Асинхронная работа с базами данных»

**Цель семинара**: научиться заменять блокирующие JDBC-вызовы на неблокирующие R2DBC, понимать разницу в моделях транзакций и управлять backpressure.

---

### Подготовка окружения

Перед началом создай новый проект Spring Initializr со следующими зависимостями:
*   `Spring Reactive Web` (WebFlux);
*   `R2DBC` (Spring Data R2DBC);
*   `PostgreSQL Driver` (R2DBC);
*   `Lombok` (опционально, для упрощения кода);
*   `Validation`.

В `application.yml` настрой подключение:
```yaml
spring:
  r2dbc:
    url: r2dbc:postgresql://localhost:5432/seminar_db
    username: postgres
    password: secret
  flyway:
    enabled: false # Для простоты семинара будем создавать таблицы через код или вручную
```

*Примечание: если у тебя нет Flyway, выполни следующий SQL-скрипт в базе данных перед запуском:*
```sql
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id),
    quantity INT NOT NULL,
    total_price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

### Задание 1. Базовый CRUD на R2DBC

**Легенда**: ты мигрируешь сервис каталога товаров с JDBC на R2DBC. Твоя задача — создать реактивный репозиторий и сервис для управления товарами.

**Задача**
1.  Создать Entity-класс `Product`.
2.  Создать `ReactiveCrudRepository`.
3.  Написать Service, который возвращает `Mono` и `Flux`.
4.  Написать Controller, который отдаёт данные в JSON.

**Решение (полный код)**

**1. Entity**
```java
package com.example.seminar.entity;

import org.springframework.data.annotation.Id;
import org.springframework.data.relational.core.mapping.Table;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import java.math.BigDecimal;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Table("products")
public class Product {
    @Id
    private Long id;
    private String name;
    private BigDecimal price;
    private int stock;
}
```

**2. Repository**
```java
package com.example.seminar.repository;

import com.example.seminar.entity.Product;
import org.springframework.data.r2dbc.repository.R2dbcRepository;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import java.math.BigDecimal;

public interface ProductRepository extends R2dbcRepository<Product, Long> {
    
    // Поиск товаров дороже определённой суммы
    Flux<Product> findByPriceGreaterThan(BigDecimal price);
    
    // Поиск по имени (частичное совпадение)
    Flux<Product> findByNameContainingIgnoreCase(String name);
}
```

**3. Service**
```java
package com.example.seminar.service;

import com.example.seminar.entity.Product;
import com.example.seminar.repository.ProductRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import java.math.BigDecimal;

@Service
@RequiredArgsConstructor
public class ProductService {

    private final ProductRepository productRepository;

    public Flux<Product> getAllProducts() {
        return productRepository.findAll();
    }

    public Mono<Product> getProductById(Long id) {
        return productRepository.findById(id)
                .switchIfEmpty(Mono.error(new RuntimeException("Product not found")));
    }

    public Mono<Product> createProduct(Product product) {
        // ID генерируется БД, поэтому обнуляем его перед сохранением
        product.setId(null); 
        return productRepository.save(product);
    }

    public Flux<Product> getExpensiveProducts(BigDecimal minPrice) {
        return productRepository.findByPriceGreaterThan(minPrice);
    }
}
```

**4. Controller**
```java
package com.example.seminar.controller;

import com.example.seminar.entity.Product;
import com.example.seminar.service.ProductService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import java.math.BigDecimal;

@RestController
@RequestMapping("/api/products")
@RequiredArgsConstructor
public class ProductController {

    private final ProductService productService;

    @GetMapping
    public Flux<Product> getAll() {
        return productService.getAllProducts();
    }

    @GetMapping("/{id}")
    public Mono<ResponseEntity<Product>> getById(@PathVariable Long id) {
        return productService.getProductById(id)
                .map(ResponseEntity::ok)
                .defaultIfEmpty(ResponseEntity.notFound().build());
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public Mono<Product> create(@RequestBody Product product) {
        return productService.createProduct(product);
    }

    @GetMapping("/expensive")
    public Flux<Product> getExpensive(@RequestParam BigDecimal minPrice) {
        return productService.getExpensiveProducts(minPrice);
    }
}
```

---

### Задание 2. Реактивные транзакции (оформление заказа)

**Легенда**: нужно реализовать оформление заказа. Это атомарная операция:
1.  Проверить наличие товара (`stock > 0`).
2.  Уменьшить количество товара на складе (`stock - 1`).
3.  Создать запись о заказе.
    Если любой шаг падает, все изменения должны откатиться.

**Важно**: в R2DBC `@Transactional` работает через `Reactor Context`. Мы используем `TransactionalOperator` для явного контроля, так как это более наглядно для обучения.

**Задача**
1.  Создать Entity `Order`.
2.  Создать Repository для `Order`.
3.  Реализовать метод `placeOrder` в сервисе с использованием `TransactionalOperator`.

**Решение (полный код)**

**1. Entity Order**
```java
package com.example.seminar.entity;

import org.springframework.data.annotation.Id;
import org.springframework.data.relational.core.mapping.Table;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Table("orders")
public class Order {
    @Id
    private Long id;
    private Long productId;
    private int quantity;
    private BigDecimal totalPrice;
    private LocalDateTime createdAt;
}
```

**2. Repository Order**
```java
package com.example.seminar.repository;

import com.example.seminar.entity.Order;
import org.springframework.data.r2dbc.repository.R2dbcRepository;
import org.springframework.data.r2dbc.repository.Query;
import reactor.core.publisher.Mono;

public interface OrderRepository extends R2dbcRepository<Order, Long> {
    
    // Кастомный запрос для создания заказа с возвратом сгенерированного ID
    @Query("INSERT INTO orders (product_id, quantity, total_price, created_at) VALUES (:productId, :quantity, :totalPrice, NOW()) RETURNING *")
    Mono<Order> saveCustom(Long productId, int quantity, BigDecimal totalPrice);
}
```

**3. Service с транзакцией**
```java
package com.example.seminar.service;

import com.example.seminar.entity.Order;
import com.example.seminar.entity.Product;
import com.example.seminar.repository.OrderRepository;
import com.example.seminar.repository.ProductRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.r2dbc.connection.R2dbcTransactionManager;
import org.springframework.stereotype.Service;
import org.springframework.transaction.reactive.TransactionalOperator;
import reactor.core.publisher.Mono;

import java.math.BigDecimal;

@Service
@RequiredArgsConstructor
public class OrderService {

    private final ProductRepository productRepository;
    private final OrderRepository orderRepository;
    private final TransactionalOperator transactionalOperator;

    /**
     * Оформляет заказ в транзакции
     */
    public Mono<Order> placeOrder(Long productId, int quantity) {
        
        // Начало реактивной цепочки, которая будет обёрнута в транзакцию
        return productRepository.findById(productId)
                .switchIfEmpty(Mono.error(new RuntimeException("Product not found")))
                .flatMap(product -> {
                    if (product.getStock() < quantity) {
                        return Mono.error(new RuntimeException("Not enough stock"));
                    }

                    // Обновляем остаток
                    product.setStock(product.getStock() - quantity);
                    BigDecimal totalPrice = product.getPrice().multiply(BigDecimal.valueOf(quantity));

                    // Сохраняем обновлённый продукт И создаём заказ
                    // Важно: мы возвращаем Mono<Order>, но внутри делаем два действия
                    return productRepository.save(product)
                            .then(orderRepository.saveCustom(productId, quantity, totalPrice));
                })
                // Оборачиваем всю цепочку в транзакцию
                .as(transactionalOperator::transactional);
    }
}
```

**4. Controller для заказов**
```java
package com.example.seminar.controller;

import com.example.seminar.entity.Order;
import com.example.seminar.service.OrderService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Mono;

@RestController
@RequestMapping("/api/orders")
@RequiredArgsConstructor
public class OrderController {

    private final OrderService orderService;

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public Mono<Order> createOrder(@RequestBody OrderRequest request) {
        return orderService.placeOrder(request.getProductId(), request.getQuantity());
    }

    // DTO для запроса
    public static class OrderRequest {
        private Long productId;
        private int quantity;

        public Long getProductId() { return productId; }
        public void setProductId(Long productId) { this.productId = productId; }
        public int getQuantity() { return quantity; }
        public void setQuantity(int quantity) { this.quantity = quantity; }
    }
}
```

---

### Задание 3. Backpressure и потоковая передача данных

**Легенда**: администратор хочет выгрузить весь каталог товаров. Товаров миллионы. Если мы просто сделаем `findAll()`, мы можем перегрузить память клиента или сеть. Нужно реализовать серверную отправку событий (SSE) или просто поток с искусственной задержкой, чтобы продемонстрировать, как R2DBC и Reactor обрабатывают запросы данных порциями.

**Задача**
1.  Добавить в контроллер endpoint, который отдаёт `Flux<Product>` с задержкой между элементами.
2.  Продемонстрировать, что данные не грузятся все сразу в память, а идут потоком.

**Решение (полный код)**

Добавьте этот метод в `ProductController`:

```java
    @GetMapping(value = "/stream", produces = "text/event-stream")
    public Flux<Product> streamProducts() {
        // Получаем поток из БД
        return productService.getAllProducts()
                // Имитируем медленную обработку или отправку на клиент (backpressure test)
                // Если клиент не успевает читать, R2DBC-драйвер остановит чтение из БД
                .delayElements(java.time.Duration.ofMillis(100)) 
                .log(); // .log() покажет в консоли сигналы request(n), onNext, onComplete
    }
```

**Как проверить**
1.  Заполни базу данными (можно использовать простой скрипт или вставить 10–20 товаров вручную).
2.  Открой в браузере или через `curl`: `curl http://localhost:8080/api/products/stream`.
3.  Ты увидишь, как товары приходят по одному с задержкой.
4.  Посмотри в логи консоли приложения. Ты увидишь сообщения от Reactor:
    *   `request(32)` — клиент запросил порцию;
    *   `onNext(Product(...))` — элемент отправлен.
 
 Если бы клиент был очень медленным, запросы к БД бы приостанавливались.

---

### Задание 4 (дополнительное). Сравнение с Virtual Threads (концептуальное)

**Легенда**: твой тимлид спрашивает: «А почему мы не используем обычные JDBC-репозитории с Virtual Threads? Это же проще».

**Задача**

Напиши короткий комментарий или тестовый метод, который показывает, чем отличается код R2DBC от кода с Virtual Threads (гипотетически).

**Ответ для студента (вставь в код как комментарий):**

```java
/*
 * СРАВНЕНИЕ ПОДХОДОВ
 * 
 * 1. R2DBC (реактивный):
 *    - код асинхронный (Mono/Flux);
 *    - поток Event Loop не блокируется;
 *    - поддерживает backpressure (клиент управляет скоростью чтения);
 *    - сложнее отлаживать, сложнее писать (цепочки операторов);
 *    - идеально для высоконагруженных I/O-операций с большим количеством одновременных соединений.
 * 
 * 2. Virtual Threads (Imperative):
 *    - код синхронный (обычные методы, JDBC);
 *    - поток «виртуальный», при I/O он отмонтируется от carrier-потока;
 *    - нет встроенного backpressure на уровне драйвера JDBC (данные читаются блоками);
 *    - проще писать, легче миграция legacy-кода;
 *    - идеально для большинства бизнес-приложений, где нет экстремальных нагрузок.
 * 
 * ВЫВОД
 * Если нужен строгий контроль над ресурсами и backpressure (например, стриминг больших данных) — выбирай R2DBC.
 * Если нужна простота разработки и высокая конкурентность без переписывания всего кода — выбирай Virtual Threads + JDBC.
 */
```