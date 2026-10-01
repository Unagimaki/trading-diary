# Trade Diary

Production authentication (Railway backend / Vercel frontend):

- Railway: `APP_ENV=production`, `FRONTEND_ORIGIN=https://trading-diary-frontend.vercel.app` (no trailing slash or wildcard).
- Vercel: `VITE_API_URL=https://trading-diary-production.up.railway.app`.
- Redeploy both services after changing code/environment variables.
- Production session cookies use `HttpOnly; Secure; SameSite=None; Path=/`, including logout deletion. Development defaults to `APP_ENV=development` and uses `HttpOnly; SameSite=Lax; Path=/` without Secure for local HTTP.
- API requests include credentials centrally in `frontend/src/shared/api/http.ts`; file uploads also include credentials.
- Verify in browser DevTools: login sets the cookie, then `/auth/me` and `/journals` send it and return 200. Browser policies blocking third-party cookies can still prevent cross-site sessions.

Trade Diary — веб-приложение для ведения гибкого дневника торговых наблюдений. Пользователь может создавать журналы с собственной структурой колонок, фиксировать сделки, потенциальные входы, ошибки и рыночные гипотезы, а затем анализировать накопленные данные.

Проект состоит из личного кабинета на React, REST API на Go и базы данных PostgreSQL. Интерфейс построен по Feature-Sliced Design, backend — по принципам Clean Architecture.

Сейчас реализованы регистрация, пользовательские сессии, управление журналами, динамические колонки, типизированные ячейки, скриншоты, контроль качества торговых данных и автоматическая аналитика.

Подробные продуктовые и технические правила находятся в [docs](docs/README.md).
