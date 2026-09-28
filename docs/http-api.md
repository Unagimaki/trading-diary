# HTTP API

## Общие правила

- В production frontend обращается к `/api`, Nginx удаляет префикс и передаёт запрос Go API.
- Формат тела — JSON, кроме загрузки изображений через `multipart/form-data`.
- Авторизация основана на HttpOnly cookie `trade_diary_session`.
- Идентификаторы ресурсов — UUID.
- Успешное удаление или обновление без тела возвращает `204 No Content`.

## Ресурсы

| Область | Маршруты |
|---|---|
| Auth | `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`, `GET /auth/me` |
| Journals | `GET/POST /journals`, `GET/PATCH/DELETE /journals/{id}` |
| Settings | `PATCH /journals/{id}/settings` |
| Columns | `GET/POST /journals/{journalID}/columns`, `PATCH/DELETE .../{id}` |
| Rows | `GET/POST /journals/{journalID}/rows`, `DELETE .../{rowID}` |
| Cells | `PATCH/DELETE .../rows/{rowID}/cells/{columnID}` |
| Images | `POST/DELETE .../cells/{columnID}/image`, `GET /attachments/{id}/content` |
| Analytics | `GET /journals/{journalID}/analytics` |
| Operations | `GET /health`, `GET /metrics` |

Актуальным источником маршрутов является `backend/internal/transport/http/router.go`. При добавлении endpoint эта таблица обновляется в той же задаче.

## Ошибки

```json
{
  "code": "INVALID_CELL_VALUE",
  "message": "invalid cell value",
  "requestId": "7e2d..."
}
```

`code` предназначен для логики frontend, `message` — безопасное описание, `requestId` — связь пользовательской ошибки с backend-логом. Внутренняя ошибка, SQL и stack trace клиенту не возвращаются.

Типовые статусы:

| Статус | Значение |
|---|---|
| `400` | Некорректный запрос или domain-валидация |
| `401` | Нет действующей сессии |
| `404` | Ресурс не найден или недоступен пользователю |
| `413` | Тело запроса отклонено прокси как слишком большое |
| `500` | Необработанная внутренняя ошибка |

## Сессия

В базе хранится SHA-256 hash случайного токена, а исходный токен находится только в HttpOnly cookie браузера. Пароль хранится как bcrypt hash. Каждый защищённый handler получает пользователя через сессию до вызова use case.

## Идемпотентность

GET не изменяет состояние. DELETE повторно может вернуть `404`. PATCH ячейки задаёт конечное значение, но может инициировать пересчёт зависимых ячеек.
