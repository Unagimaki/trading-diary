# Trade Diary

MVP дневника торговых наблюдений с гибкой структурой колонок.

## Структура

- `frontend` — React-приложение по Feature-Sliced Design.
- `backend` — REST API на Go по принципам Clean Architecture.
- `compose.yaml` — локальный PostgreSQL.

## Локальный запуск

Весь стек в Docker:

```powershell
Copy-Item .env.example .env
docker compose up --build -d
```

После запуска frontend доступен на `http://localhost:3000`, а healthcheck backend — на `http://localhost:8080/health`.

Запуск приложений для разработки:

```powershell
docker compose up -d
cd backend
Copy-Item .env.example .env
go run ./cmd/api
```

В отдельном терминале:

```powershell
cd frontend
Copy-Item .env.example .env
npm install
npm run dev
```
