## Запуск

Требуется установленный Docker и Docker Compose.

Перед запуском можно настроить `.env` (порты, DSN к БД, лог‑уровень).

Из корневой директории:

```bash
docker-compose up --build
```

По умолчанию сервис доступен на `http://localhost:${APP_PORT}` (по умолчанию `8080`), Postgres — на `localhost:${POSTGRES_PORT}` (по умолчанию `5432`).
