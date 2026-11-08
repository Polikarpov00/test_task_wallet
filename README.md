# Wallet Service

## Запуск

```
docker compose up --build
```

Сервис будет доступен на `http://localhost:8081`, PostgreSQL — на `localhost:5433`.
Миграции применяются автоматически при старте. Настройки лежат в `config.env`.

## API

## Принял решение назвать поле walletId вместо valletId, т.к. посчитал опечаткой !

### Изменить баланс

`POST /api/v1/wallet`

```json
{
  "walletId": "7b2f5c1e-3d4a-4e8b-9f6c-1a2b3c4d5e6f",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

`operationType` принимает значения `DEPOSIT` или `WITHDRAW`. `amount` — целое положительное число в минимальных единицах валюты (копейках).

Ответ `200`:
```json
{ "walletId": "7b2f5c1e-...", "balance": 1000 }
```

### Получить баланс

`GET /api/v1/wallets/{walletId}`

Ответ `200`:
```json
{ "walletId": "7b2f5c1e-...", "balance": 1000 }
```

### Коды ответов

| Код | Когда |
|-----|-------|
| 200 | Успех |
| 400 | Невалидный JSON или UUID, неизвестный `operationType`, `amount <= 0` |
| 404 | Кошелёк не найден (GET или WITHDRAW) |
| 422 | Недостаточно средств |
| 500 | Внутренняя ошибка (детали пишутся в лог, клиенту не отдаются) |

Первый `DEPOSIT` на новый `walletId` создаёт кошелёк.

## Тесты

Тесты конкурентности работают на реальном PostgreSQL (нужен запущенный `docker compose`):

```bash
go test -race -count=1 ./...
```

- `ConcurrentMixed`: 200 параллельных депозитов и снятий; итоговый баланс сходится.
- `ConcurrentOverdraw`: на балансе 100, 50 параллельных снятий по 10. Ровно 10 успешных, 40 получают «недостаточно средств», итоговый баланс 0.

Если БД недоступна, тесты пропускаются. DSN можно переопределить через `TEST_DB_DSN`.

## Нагрузочный тест

Нужен [vegeta](https://github.com/tsenart/vegeta) (`brew install vegeta`):

```bash
./loadtest/run.sh 

RATE=2000 DURATION=10s ./loadtest/run.sh
```

## Структура

```
cmd/wallet/            точка входа, сборка зависимостей, graceful shutdown
internal/config/       загрузка конфигурации из env / config.env
internal/domain/       модели и доменные ошибки
internal/service/      бизнес-валидация
internal/handler/      HTTP: JSON, маршруты, маппинг ошибок в коды ответов
internal/repository/   PostgreSQL: пул соединений, миграции, запросы
migrations/            SQL-миграции (goose, встроены в бинарник через embed)
loadtest/              нагрузочный тест на vegeta
```

## Стек

Go 1.26, PostgreSQL 16, pgx/v5, goose, caarlos0/env