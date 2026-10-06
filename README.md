# Trip Service — лабораторная работа 1

Сервис поездок TripGo: HTTP API по контракту из `contracts/openapi`, данные в
PostgreSQL. Умеет создавать поездку, отдавать её и завершать.

## Требования

- Go 1.27.1
- `tripgoctl` (локальное окружение курса)
- `make`
- `golangci-lint` — только для `make lint`

## Первый запуск

Один раз после клонирования поставить goose как tool модуля:

```bash
go get -tool github.com/pressly/goose/v3/cmd/goose
go mod tidy
```

Дальше окружение, миграции и сервис:

```bash
tripgoctl cluster start        # один раз на машине
tripgoctl environment start    # поднимает PostgreSQL и создаёт .env
make migrate
make run
```

`Makefile` сам подхватывает `.env`, экспортировать переменные руками не нужно.

## Команды

| Команда | Что делает |
|---|---|
| `make generate` | генерирует `internal/generated/api.gen.go` по OpenAPI |
| `make migrate` | накатывает все миграции |
| `make migrate-down` | откатывает последнюю миграцию |
| `make migrate-reset` | откатывает все миграции |
| `make migrate-status` | состояние миграций |
| `make build` | собирает `bin/trip-service` |
| `make run` | собирает и запускает сервис |
| `make test` | `go test -race ./...` (тестов в ЛР1 нет, появятся в ЛР2) |
| `make lint` | `golangci-lint` с конфигом курса |

## Переменные окружения

Все переменные, кроме `HTTP_*_TIMEOUT`, обязательные. Если какой-то нет или
значение не разбирается, сервис падает на старте и перечисляет все проблемы
сразу. Пример — `.env.example`.

| Переменная | Пример | Что задаёт |
|---|---|---|
| `HTTP_ADDR` | `:8080` | адрес HTTP-сервера |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `SHUTDOWN_TIMEOUT` | `10s` | весь бюджет на graceful shutdown |
| `DATABASE_URL` | `postgres://...` | строка подключения, берётся из `.env` |
| `DATABASE_MAX_CONNS` | `10` | максимум соединений в пуле |
| `DATABASE_MIN_CONNS` | `2` | минимум соединений в пуле |
| `DATABASE_MAX_CONN_LIFETIME` | `30m` | сколько живёт соединение |
| `DATABASE_CONNECT_TIMEOUT` | `5s` | таймаут подключения и стартового `Ping` |
| `DATABASE_QUERY_TIMEOUT` | `3s` | таймаут каждого запроса, `BEGIN`/`COMMIT`/`ROLLBACK` и проверки в `/ready` |
| `HTTP_READ_TIMEOUT` | `10s` | необязательная, по умолчанию `10s` |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | необязательная, по умолчанию `5s` |
| `HTTP_WRITE_TIMEOUT` | `15s` | необязательная, по умолчанию `15s` |
| `HTTP_IDLE_TIMEOUT` | `60s` | необязательная, по умолчанию `60s` |

Таймауты HTTP-сервера сделаны необязательными, потому что их нет в базовом
наборе курса и `tripgoctl` не кладёт их в `.env`.

## Структура

```
cmd/trip-service/       main: конфиг, зависимости, сервер, graceful shutdown
internal/config/        чтение и проверка переменных окружения
internal/domain/        поездка, переход статуса, доменные ошибки
internal/postgres/      пул и менеджер транзакций
internal/repository/    запросы к trips и trip_status_history (pgx + squirrel)
internal/service/       бизнес-логика: создание, получение, завершение
internal/httpapi/       хендлеры, валидация, problem+json, роутер
internal/generated/     код, сгенерированный по OpenAPI, и конфиг генератора
migrations/             миграции goose
contracts/              копия контрактов курса, не редактируется
```

Сгенерированный код лежит в `internal/generated/`, как в команде из задания ЛР1.
Руками его не правим: `make generate` перезаписывает файл.

## Как проверить

```bash
# создание
curl -sS -X POST localhost:8080/api/v1/trips \
  -H 'content-type: application/json' \
  -d '{"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c",
       "driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5",
       "start_point":{"latitude":59.9398,"longitude":30.3146},
       "end_point":{"latitude":59.9290,"longitude":30.3626},
       "price":1450}'

curl -sS localhost:8080/api/v1/trips/$TRIP_ID                 # получение
curl -sS -X POST localhost:8080/api/v1/trips/$TRIP_ID/finish  # 200, повтор — 409 trip_completed
```

Гонка на создание (в `trip.json` тело запроса выше), ожидаем ровно один `201`:

```bash
seq 20 | xargs -P20 -I{} curl -s -o /dev/null -w '%{http_code}\n' \
  -X POST localhost:8080/api/v1/trips -H 'content-type: application/json' \
  -d @trip.json | sort | uniq -c
```

Гонка на завершение, ожидаем ровно один `200`:

```bash
seq 20 | xargs -P20 -I{} curl -s -o /dev/null -w '%{http_code}\n' \
  -X POST localhost:8080/api/v1/trips/$TRIP_ID/finish | sort | uniq -c
```

Атомарность создания. Код менять не нужно, достаточно временно сломать вторую
вставку ограничением, которое не проходит ни одна строка:

```bash
psql "$DATABASE_URL" -c "ALTER TABLE trip_status_history ADD CONSTRAINT break_inserts CHECK (false) NOT VALID"
# POST /api/v1/trips → 500 internal_error, в trips новой строки нет
psql "$DATABASE_URL" -c "ALTER TABLE trip_status_history DROP CONSTRAINT break_inserts"
```

`/ready`: остановить PostgreSQL — `503 {"status":"unavailable"}`, запустить снова —
`200`. `/health` всё это время отвечает `200`.

Graceful shutdown: `kill -TERM $(pgrep -f bin/trip-service)`. Сервис перестаёт
принимать соединения, дожидается текущих запросов и закрывает пул. Если не
укладывается в `SHUTDOWN_TIMEOUT`, пишет об этом в лог и закрывает соединения
принудительно.

## Решения

### Уровень изоляции — Read Committed

Все транзакции открываются с `READ COMMITTED`. Ни одно бизнес-правило не держится
на том, что транзакция «видит снимок» данных:

- запрет двух активных поездок держит уникальный индекс, а он работает на любом
  уровне изоляции;
- завершение поездки проверяет статус условием в самом `UPDATE`. В Read Committed
  второй `UPDATE` той же строки ждёт коммита первого и перепроверяет `WHERE` уже
  на новой версии строки.

`REPEATABLE READ` или `SERIALIZABLE` ничего бы здесь не добавили, зато на
конкурентных запросах давали бы ошибки сериализации (`40001`), которые пришлось
бы ловить и повторять. К тому же Read Committed — уровень по умолчанию в
PostgreSQL, поведение предсказуемое.

### Менеджер транзакций

`internal/postgres/txmanager.go`, интерфейс в бизнес-коде:

```go
type TxManager interface {
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}
```

- `Do` открывает транзакцию (`BEGIN ... ISOLATION LEVEL READ COMMITTED`), кладёт
  `pgx.Tx` в `context` под приватным ключом и вызывает `fn` с этим контекстом.
- `fn` вернула `nil` — `COMMIT`. Вернула ошибку — `ROLLBACK`, ошибка уходит
  наружу (если упал и сам откат, обе ошибки объединяются через `errors.Join`).
  Паника — `ROLLBACK` в `defer` с `recover`, затем паника летит дальше.
- Репозиторий берёт исполнителя через `postgres.QuerierFromContext(ctx, pool)`:
  если в контексте есть транзакция, запрос идёт в неё, иначе в пул. Транзакция
  не передаётся аргументом метода репозитория.
- **Вызов репозитория вне `Do`** просто идёт через пул: каждый запрос выполняется
  в своей неявной транзакции (autocommit). Так работает `GET /trips/{id}`.
- **Вложенный `Do`** видит транзакцию в контексте и новую не открывает: `fn`
  выполняется в той же транзакции, а коммитит или откатывает внешний `Do`.
- Сервис (`internal/service`) знает только интерфейс `TxManager` и ничего не знает
  про `pgx`, транзакции и пулы.
- `ROLLBACK` выполняется с контекстом без отмены (`context.WithoutCancel`) и своим
  таймаутом: если клиент ушёл и контекст запроса отменён, транзакцию всё равно
  нужно откатить и вернуть соединение в пул.

Создание поездки — один `Do` с двумя вставками: в `trips` и в
`trip_status_history`. Завершение — тоже один `Do`: `UPDATE trips` и запись
перехода `active → completed` в журнал.

### Запрет двух активных поездок у водителя

Частичный уникальный индекс в миграции `00001_create_trips.sql`:

```sql
CREATE UNIQUE INDEX trips_driver_active_uniq ON trips (driver_id) WHERE status = 'active';
```

В индекс попадают только активные поездки, поэтому завершённых у водителя может
быть сколько угодно, а активная — одна. При двух одновременных `INSERT` второй
ждёт, пока закоммитится первый, и получает ошибку `23505`. Репозиторий
распознаёт её по коду и имени ограничения и возвращает `domain.ErrDriverBusy`,
хендлер отвечает `409 driver_busy`.

Проверка `SELECT` перед `INSERT` здесь не годится: оба параллельных запроса
пройдут проверку раньше, чем кто-то из них вставит строку.

### Конкурентное завершение поездки

```sql
-- так squirrel собирает запрос в TripRepository.Finish
-- ($1 = 'completed', $5 = 'active')
UPDATE trips SET status = $1, finished_at = $2, updated_at = $3
WHERE id = $4 AND status = $5
RETURNING id, user_id, ...
```

Проверка статуса и запись — одна атомарная операция, окна между ними нет. При двух
одновременных `finish` первый `UPDATE` блокирует строку. Второй ждёт его коммита,
перечитывает строку, видит `status = 'completed'` и ничего не обновляет: один
`200`, один `409`, `finished_at` не перезаписывается.

Ноль обновлённых строк означает «поездки нет» или «уже завершена». Чтобы различить
эти случаи, репозиторий дочитывает поездку: нет строки — `404 trip_not_found`,
есть — `409 trip_completed`.

Если убрать `AND status = 'active'` (или сделать «прочитали, проверили в Go,
записали» без блокировки), оба запроса увидят `active`, оба обновят строку и
вернут `200`. Второй перезапишет `finished_at`, а в журнале окажутся два перехода
`active → completed`.

### `/health` и `/ready`

- `/health` — liveness: «процесс жив». В бд не ходит. Если бы ходил, то при
  падении PostgreSQL оркестратор начал бы перезапускать все живые инстансы
  сервиса, хотя перезапуск их не починит, а только добавит нагрузки при
  восстановлении.
- `/ready` — readiness: «можно ли слать трафик». Делает `Ping` пула с таймаутом.
  Если бд недоступна, отвечает `503`, и балансировщик временно не шлёт на
  инстанс запросы, но процесс не убивает.

Обе ручки отвечают `application/json` со схемой `HealthResponse`.

### Ошибки и валидация

- Все ошибки отдаются в `application/problem+json` (RFC 9457) с полем `code`.
  Перевод доменных ошибок в HTTP собран в одном месте —
  `internal/httpapi/errors.go`, функция `writeError`.
- Неизвестная ошибка логируется целиком, а клиенту уходит `500 internal_error`
  без деталей: текст ошибки бд и SQL наружу не попадают.
- Тело запроса разбирается в сгенерированный `api.TripData` с
  `DisallowUnknownFields` (в контракте `additionalProperties: false`). Наличие
  обязательных полей проверяется отдельно: после разбора `"price": 0` не
  отличить от отсутствующего `price`. Затем проверяются значения: UUID не нулевые,
  координаты в диапазонах, `price >= 0`.
- Если `tripId` не UUID, сгенерированный код вызывает наш обработчик ошибок
  параметров, и тот отвечает `400 invalid_request`.

### Прочее

- `id` поездки и `started_at` выставляет сервис. Время обрезается до микросекунд,
  как его хранит PostgreSQL, чтобы ответ на создание совпадал с последующим `GET`.
- `finished_at` и `last_position_at` в ответе всегда присутствуют и равны `null`,
  пока не заданы (опция `nullable-type` генератора). `last_position_at` появится
  в ЛР3.
- Цена хранится в целых рублях (`BIGINT`) — учебное упрощение из задания.
- Каждый запрос в бд выполняется с таймаутом `DATABASE_QUERY_TIMEOUT`.

## Что не сделано

Задания со звёздочкой (Idempotency-Key и Dockerfile) в этой версии не реализованы.
Заголовок `Idempotency-Key` принимается по контракту, но пока игнорируется.
