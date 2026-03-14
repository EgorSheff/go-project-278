### Hexlet tests and linter status:
[![Actions Status](https://github.com/EgorSheff/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/EgorSheff/go-project-278/actions)

# URL Shortener

Сервис сокращения ссылок на Go (Gin + PostgreSQL + sqlc).

## Запуск

```bash
docker compose up
```

Приложение будет доступно на порту `80`. Требуемые переменные окружения (`DATABASE_URL`, `BASE_URL`) уже настроены в `docker-compose.yml`.

## Разработка

```bash
make build   # Сборка бинарника
make test    # Запуск тестов
make lint    # Линтер
make sqlc    # Перегенерация кода из SQL
```

## API

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/links` | Список ссылок (пагинация) |
| POST | `/api/links` | Создать ссылку |
| GET | `/api/links/:id` | Получить ссылку |
| PUT | `/api/links/:id` | Обновить ссылку |
| DELETE | `/api/links/:id` | Удалить ссылку |
| GET | `/api/link_visits` | Список визитов |
| GET | `/r/:code` | Редирект по короткому коду (307) |
