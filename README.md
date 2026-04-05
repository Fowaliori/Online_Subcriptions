# Сервис для агрегации данных об онлайн подписках пользователей 

REST-сервис для хранения подписок пользователей и расчёта суммарных расходов за период.

---

## Функционал

- CRUD подписок
- Подсчёт расходов за период (`/subscriptions/summary`)
- Фильтрация по `user_id` и `service_name`

---

## Стек

- Go
- PostgreSQL
- Goose (миграции)
- Docker / docker-compose

---

## Запуск

```bash
make run
```

Сервис: http://localhost:8080

---

## Пример

```bash
curl "http://localhost:8080/subscriptions/summary?start=01-2024&end=12-2024"
```

---
