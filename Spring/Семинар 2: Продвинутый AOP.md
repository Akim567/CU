### **1. Load-Time Weaving: Аспект для конструкторов (Преодоление ограничений прокси)**

**Задача:** Преодолеть ключевое ограничение Spring AOP на основе прокси — невозможность перехватывать вызовы конструкторов и внутренние вызовы методов внутри класса.

**Действия:**

1. **Создайте простой доменный класс, не управляемый Spring:**
   ```java
   public class AuditRecord {
       private String id;
       private String action;
   
   // бойлерплейт
   }
   ```

2. **Создайте сервис, демонстрирующий проблему:**
   ```java
   @Service
   public class AuditService {
       
       public void createRecordDirectly() {
           // Этот конструктор НЕ будет перехвачен стандартным Spring AOP!
           AuditRecord record = new AuditRecord("direct-1", "CREATE");
           System.out.println("Запись создана напрямую: " + record.getId());
       }
       
       @Transactional // Метод с другой прокси-логикой
       public void createRecordInTransactional() {
           AuditRecord record = new AuditRecord("tx-1", "UPDATE");
           System.out.println("Запись создана в транзакционном методе: " + record.getId());
       }
   }
   ```

3. **Создайте аспект для перехвата конструкторов (AspectJ-аспект):**
   ```java
   @Aspect
   public class ConstructorAuditAspect {
       
       // Pointcut на выполнение конструктора AuditRecord
       @Pointcut("execution(com.example.demo.AuditRecord.new(..))")
       public void auditRecordConstructor() {}
       
       // Advice, который сработает ВОКРУГ конструктора
       @Around("auditRecordConstructor()")
       public Object logConstructorCall(ProceedingJoinPoint pjp) throws Throwable {
           System.out.println("[LTW] Начало создания AuditRecord с аргументами: " 
                              + Arrays.toString(pjp.getArgs()));
           // Вызов оригинального конструктора - ОБЯЗАТЕЛЕН
           Object result = pjp.proceed();
           System.out.println("[LTW] Завершено создание объекта: " + result);
           return result;
       }
   }
   ```
   **Важно:** Этот аспект НЕ помечаем `@Component` — он будет управляться AspectJ, не Spring!

4. **Настройте Load-Time Weaving:**
    - Добавьте зависимости в `pom.xml`:
      ```xml
      <dependency>
          <groupId>org.aspectj</groupId>
          <artifactId>aspectjrt</artifactId>
          <version>1.9.19</version>
      </dependency>
      <dependency>
          <groupId>org.aspectj</groupId>
          <artifactId>aspectjweaver</artifactId>
          <version>1.9.19</version>
      </dependency>
      ```
    - Создайте файл `src/main/resources/META-INF/aop.xml`:
      ```xml
      <!DOCTYPE aspectj PUBLIC "-//AspectJ//DTD//EN" "http://www.eclipse.org/aspectj/dtd/aspectj.dtd">
      <aspectj>
          <weaver>
              <include within="com.example.demo..*"/>
          </weaver>
          <aspects>
              <aspect name="com.example.demo.aspect.ConstructorAuditAspect"/>
          </aspects>
      </aspectj>
      ```
    - Включите LTW в конфигурации Spring:
      ```java
      @Configuration
      @EnableLoadTimeWeaving // Вместо @EnableAspectJAutoProxy
      public class AppConfig {
          // Конфигурация
      }
      ```

5. **Запустите приложение с Java Agent:**
    - В параметрах запуска JVM добавьте:
      ```
      -javaagent:"путь_к_вашему_мавен_репозиторию/org/aspectj/aspectjweaver/1.9.19/aspectjweaver-1.9.19.jar"
      ```
    - Вызовите методы `createRecordDirectly()` и `createRecordInTransactional()`
    - **Ожидаемый результат:** Аспект сработает для каждого конструктора `AuditRecord`, независимо от того, где и как он вызывается.

