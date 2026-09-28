# Потоки приложения

## Жизненный цикл запроса

```mermaid
sequenceDiagram
    participant UI as React UI
    participant NG as Nginx
    participant MW as Middleware
    participant H as Handler
    participant S as Domain service
    participant R as Repository
    participant DB as PostgreSQL

    UI->>NG: HTTP /api/...
    NG->>MW: proxy request
    MW->>MW: requestId, timer, panic recovery
    MW->>H: request context
    H->>H: decode + authenticate
    H->>S: use case
    S->>R: repository interface
    R->>DB: parameterized SQL
    DB-->>R: data
    R-->>S: domain result
    S-->>H: result / domain error
    H-->>UI: JSON + X-Request-ID
    MW->>MW: access log + metrics
```

## Регистрация и вход

```mermaid
flowchart TD
    A[Email и пароль] --> B[HTTP handler]
    B --> C{Валидация}
    C -- Ошибка --> D[4xx + error code]
    C -- OK --> E[Hash/проверка bcrypt]
    E --> F[Создать случайный session token]
    F --> G[Сохранить hash токена в PostgreSQL]
    G --> H[Установить HttpOnly cookie]
    H --> I[Открыть список журналов]
```

## Изменение ячейки

```mermaid
flowchart TD
    A[Пользователь меняет ячейку] --> B[PATCH cell]
    B --> C[Проверить сессию и владение]
    C --> D[Получить тип и роль колонки]
    D --> E{JSON соответствует типу и роли?}
    E -- Нет --> F[400 INVALID_CELL_VALUE]
    E -- Да --> G[Сохранить source=manual]
    G --> H[Загрузить торговый контекст строки]
    H --> I[Пересчитать calculated Risk/RR/PnL/Result]
    I --> J[Вернуть 204]
    J --> K[Frontend инвалидирует rows и analytics]
```

Подробные формулы описаны в [правилах торговых расчётов](trading-calculation-rules.md).

## Изменение схемы журнала

Создание, изменение роли или удаление колонки меняет смысл существующих строк. После операции backend запускает пересчёт всех строк журнала. Frontend обновляет columns, rows и analytics.

## Загрузка изображения

```mermaid
flowchart TD
    A[Выбор файла] --> B[Nginx проверяет размер HTTP body]
    B --> C[Backend проверяет размер и MIME]
    C --> D[Проверить journal/row/column/user]
    D --> E[Записать файл в storage]
    E --> F[Записать metadata в PostgreSQL]
    F --> G[Обновить строку на frontend]
    C -- Ошибка --> H[Структурированная 4xx ошибка]
    E -- Ошибка --> I[500 + requestId + backend log]
```

Порядок записи файла и metadata требует компенсации при частичном сбое: незавершённая операция не должна оставлять недоступный metadata record или бесхозный файл.

## Аналитика

`GET /journals/{id}/analytics` загружает согласованный dataset: системные колонки, строки, значения и настройки. Domain-сервис вычисляет метрики и data-quality issues, после чего frontend только форматирует отчёт.

```mermaid
flowchart LR
    C[Columns + roles] --> A[Analytics service]
    R[Rows + cells] --> A
    S[Deposit + Risk + RR] --> A
    A --> M[Metrics]
    A --> D[Distributions]
    A --> Q[Data quality]
```
