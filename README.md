# Разработка банковской системы

### [Модели](./src/models/)

На уровне моделей происходит сериализация/десериализация.
Валидация полей и проверка на уникальность осуществляется средствами SQL, а также на уровне обработчиков.

### [Репозитории](./src/repository/)

На уровне репозиториев происходит инкапсуляция и параметризация SQL запрсоов. 
Ошибки обрабатываются и пробрасываются на уровень выше. 

### [Сервисы](./src/services/)

Слой сервисов реализовывает основную бизнес-логику приложения:

- Регистрация и аутентификация
- Создание счетов, пополнение баланса
- Переводы между счетами
- Генерация карт
- Кредиты
- Интеграция с SMTP
- Интеграция с ЦБ РФ
- Шедулер для списания платежей

Также производится логирование через logrus

### [Обработчики](./src/handlers/)

На уровне обработчиков происходит валидация входных данных и формирование HTTP-ответов посредством вызова методов сервисов. Проверка прав доступа в ресурсам осуществляется через Middleware.

### [Маршрутизация](./main.go)

Реализована на уровне корня приложения и использует слой обработчиков.

### [Middleware](./src/middlewares/)

Middleware выполняет несколько функций - проверка токена, блокировку неавторизованных запросов, а также добавление ID пользователя в контекст.

### [База данных](./src/config/db.go)

В качестве СУБД используется PostgreSQL