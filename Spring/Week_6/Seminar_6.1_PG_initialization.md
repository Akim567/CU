# Подготовка к семинару 6: Настройка окружения — Docker + PostgreSQL
**Цель:** Подготовить рабочее окружение для всех последующих семинаров. Установить Docker, поднять PostgreSQL, настроить подключение.

---

## Задание 1. Установка Docker

Установите согласно инструкции - https://docs.docker.com/engine/install/ 

**Проверка:**
```bash
docker --version
# Ожидаемый вывод: Docker version 24.x.x, build ...

docker-compose --version
# Ожидаемый вывод: Docker Compose version v2.x.x
```
---

## Задание 2. Настройка PostgreSQL через `docker-compose` 

### 2.1. Создание `docker-compose.yml`
**Задача:** Создать файл для управления контейнером через Docker Compose (рекомендуемый способ для проектов).

Пример можно взять тут https://github.com/khezen/compose-postgres/blob/master/docker-compose.yml 

### 2.2. Запуск и остановка
**Задача:** Научиться управлять контейнером через Compose.

**Команды:**
```bash
# Запуск
docker compose up -d

# Просмотр логов
docker compose logs -f postgres

# Остановка (данные сохраняются в volume)
docker compose down

# Остановка с удалением volume (данные удаляются!)
docker compose down -v

# Перезапуск
docker compose restart
```

---

## Задание 3. Инициализация схемы БД 

### 3.1. Создание скрипта миграции
**Задача:** Создать SQL-скрипт, который автоматически выполнится при первом запуске контейнера.

**Файл `init-scripts/01_init_tables.sql` (вам необходимо настроить `volume` в `docker-compose.yml`, добавив `- ./init-scripts:/docker-entrypoint-initdb.d`):**
```sql
-- Создаём таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаём таблицу постов
CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Добавляем тестовые данные
INSERT INTO users (username, email) VALUES 
    ('alice', 'alice@example.com'),
    ('bob', 'bob@example.com');

INSERT INTO posts (title, content, user_id) VALUES 
    ('Первый пост', 'Содержимое первого поста', 1),
    ('Второй пост', 'Содержимое второго поста', 2);
```

**Проверка:**
1. Удалите volume: `docker compose down -v`
2. Запустите заново: `docker compose up -d`
3. Подключитесь к БД и проверьте таблицы:
```bash
docker exec -it postgresdb psql -U postgres -d seminar_db -c "\dt"
docker exec -it postgresdb psql -U postgres -d seminar_db -c "SELECT * FROM users;"
```

---

## Задание 4. Подключение из приложения 

### 4.1. Конфигурация для Spring Boot
**Задача:** Подготовить конфигурацию для подключения из Java-приложения.

Подключите необходимые зависимости в проект `spring-boot-starter-jdbc`

**Файл `application.yml`:**
```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/seminar_db
    username: postgres
    password: postgres
    driver-class-name: org.postgresql.Driver
```

### 4.2. Тестовое подключение
**Задача:** Создать простой тест для проверки подключения.

**Код (CommandLineRunner):**
```java
@Component
@RequiredArgsConstructor
public class DatabaseTestRunner implements CommandLineRunner {
    private final JdbcTemplate jdbcTemplate;
    
    @Override
    public void run(String... args) throws Exception {
        // Пример простого запроса через JDBC
        // queryForObject возвращает результат запроса (количество строк)
        final Integer userCount = jdbcTemplate.queryForObject(
            "SELECT count(*) FROM users",
            Integer.class
        );

        System.out.println("==========================================");
        System.out.println("Подключение к БД успешно!");
        System.out.println("Найдено пользователей в базе: " + userCount);

        // Пример выборки данных
        final String username = jdbcTemplate.queryForObject(
            "SELECT username FROM users WHERE id = ?",
            String.class,
            1
        );
        System.out.println("Пользователь с ID=1: " + username);
        System.out.println("==========================================");
    }
}
```

**Проверка:** Запустите приложение — в консоли должно появиться сообщение об успешном подключении.

---

## Чек-лист готовности
- [ ] Docker установлен и работает (`docker --version`)
- [ ] Контейнер PostgreSQL запущен (`docker ps`)
- [ ] Базы данных `seminar_db` существует
- [ ] Таблицы `users` и `posts` созданы
- [ ] Подключение из Spring Boot работает
- [ ] [Опционально] PgAdmin доступен по `http://localhost:8080`

---
