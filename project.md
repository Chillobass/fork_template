Project Path: tripgo-homework-02

Source Tree:

```txt
tripgo-homework-02
├── Makefile
├── README.md
├── cmd
│   └── trip-service
│       └── main.go
├── contracts
│   ├── asyncapi
│   │   └── driver-updates.asyncapi.yaml
│   ├── events
│   │   └── README.md
│   ├── openapi
│   │   ├── push-service.openapi.yaml
│   │   └── trip-service.openapi.yaml
│   ├── proto
│   │   ├── push
│   │   │   └── v1
│   │   │       └── push.proto
│   │   └── trip
│   │       └── v1
│   │           └── trip_service.proto
│   └── schema.md
├── docs
│   └── lab02
│       ├── grafana.png
│       └── jaeger.png
├── environment.toml
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   └── config.go
│   ├── container
│   │   └── container.go
│   ├── domain
│   │   └── trip
│   │       ├── errors.go
│   │       ├── trip.go
│   │       └── trip_test.go
│   ├── generated
│   │   ├── api.gen.go
│   │   └── oapi-codegen.yaml
│   ├── http
│   │   ├── domainerror
│   │   │   ├── domainerror.go
│   │   │   └── domainerror_test.go
│   │   ├── health
│   │   │   ├── contract.go
│   │   │   └── handler.go
│   │   ├── middleware
│   │   │   ├── http.go
│   │   │   └── http_test.go
│   │   ├── respond
│   │   │   └── json.go
│   │   ├── router.go
│   │   └── trip
│   │       ├── convert.go
│   │       ├── create
│   │       │   ├── contract.go
│   │       │   ├── convert.go
│   │       │   ├── handler.go
│   │       │   ├── handler_test.go
│   │       │   └── mocks_test.go
│   │       ├── finish
│   │       │   ├── contract.go
│   │       │   ├── handler.go
│   │       │   ├── handler_test.go
│   │       │   └── mocks_test.go
│   │       └── get
│   │           ├── contract.go
│   │           ├── handler.go
│   │           ├── handler_test.go
│   │           └── mocks_test.go
│   ├── infra
│   │   ├── postgres
│   │   │   ├── pool.go
│   │   │   ├── tracer.go
│   │   │   └── transaction
│   │   │       └── manager.go
│   │   └── telemetry
│   │       ├── log.go
│   │       ├── providers.go
│   │       └── trips.go
│   ├── model
│   │   └── status_change.go
│   ├── repository
│   │   └── trip
│   │       ├── contract.go
│   │       ├── convert.go
│   │       ├── fetch.go
│   │       ├── history.go
│   │       ├── repo.go
│   │       ├── repo_integration_test.go
│   │       └── save.go
│   └── usecase
│       └── trip
│           ├── create
│           │   ├── contract.go
│           │   ├── mocks_test.go
│           │   ├── usecase.go
│           │   └── usecase_test.go
│           ├── finish
│           │   ├── contract.go
│           │   ├── mocks_test.go
│           │   ├── usecase.go
│           │   └── usecase_test.go
│           └── get
│               ├── contract.go
│               ├── mocks_test.go
│               ├── usecase.go
│               └── usecase_test.go
└── migrations
    ├── 00001_create_trips.sql
    └── 00002_create_trip_status_history.sql

```

`Makefile`:

```
ifneq (,$(wildcard .env))
include .env
export
endif

OPENAPI_SPEC   := contracts/openapi/trip-service.openapi.yaml
OAPI_CONFIG    := internal/generated/oapi-codegen.yaml
MIGRATIONS_DIR := migrations
BIN            := bin/trip-service

GOOSE := go tool goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)"

.PHONY: generate migrate migrate-down migrate-reset migrate-status build run test test-integration lint

generate:
	go tool oapi-codegen -config $(OAPI_CONFIG) $(OPENAPI_SPEC)
	go generate ./internal/...

migrate:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-reset:
	$(GOOSE) reset

migrate-status:
	$(GOOSE) status

build:
	go build -o $(BIN) ./cmd/trip-service

run: build
	./$(BIN)

test:
	go test -race -cover ./...

test-integration:
	go test -race -tags=integration -count=1 ./internal/repository/trip

lint:
	golangci-lint run ./...

```

`README.md`:

```md
# Trip Service — лабораторная работа 2

Сервис создаёт, выдаёт и завершает поездки. Контракт HTTP хранится в
`contracts/openapi/trip-service.openapi.yaml`, состояние — в PostgreSQL.

## Запуск

Нужны Go 1.27.1, Make, Docker и [tripgoctl](https://github.com/course-go-autumn-2026/course-infra).
Для линтера нужен golangci-lint 2.15 или совместимая версия v2, собранная с Go 1.27.

```sh
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect
make migrate
make run
```

`tripgoctl` поднимает PostgreSQL, OTLP-коллектор, Jaeger и Grafana. Он создаёт
локальный `.env`, который Makefile загружает автоматически.

| Переменная | Назначение |
|---|---|
| `HTTP_ADDR` | Адрес сервера, например `:8080` |
| `HTTP_READ_TIMEOUT`, `HTTP_READ_HEADER_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT` | Таймауты сервера. По умолчанию 10s, 5s, 15s и 60s соответственно |
| `LOG_LEVEL` | Уровень логирования |
| `SHUTDOWN_TIMEOUT` | Общий бюджет остановки сервиса |
| `DATABASE_URL` | Строка подключения из `tripgoctl` |
| `DATABASE_MAX_CONNS`, `DATABASE_MIN_CONNS` | Размер пула |
| `DATABASE_MAX_CONN_LIFETIME` | Срок жизни соединения |
| `DATABASE_CONNECT_TIMEOUT`, `DATABASE_QUERY_TIMEOUT` | Таймауты подключения и каждой операции с БД |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Адрес OTLP из `tripgoctl` |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `http/protobuf` по умолчанию или `grpc` |
| `OTEL_SERVICE_NAME`, `OTEL_RESOURCE_ATTRIBUTES` | Имя сервиса и атрибуты ресурса OTel |
| `OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG` | Сэмплер и его аргумент по правилам OTel SDK |
| `OTEL_METRIC_EXPORT_INTERVAL` | Интервал отправки метрик, в миллисекундах |

Полный пример базовых настроек лежит в `.env.example`. `/health` отвечает,
если процесс запущен. `/ready` проверяет доступность PostgreSQL с таймаутом.

## Архитектура

```text
cmd/trip-service → container → HTTP, usecase, repository, infra
HTTP → интерфейс сценария → usecase → domain
                           usecase → интерфейсы хранилища и транзакции
                           repository → SQL → исполнитель из infra
```

- `internal/domain/trip` проверяет данные поездки и определяет переход
  `active → completed` и ошибки. `internal/model` содержит запись истории,
  которую сценарий передаёт в хранилище.
- `internal/usecase/trip/{create,get,finish}` управляет операциями. Каждый
  сценарий объявляет в своём `contract.go` только нужные ему методы.
- `internal/http/trip/{create,get,finish}` обрабатывает HTTP-запросы. Хендлеры
  объявляют собственные интерфейсы сценариев, преобразуют модели OpenAPI и
  используют `internal/http/domainerror` для ответов `application/problem+json`.
- `internal/repository/trip` хранит поездки и журнал. Его интерфейс
  исполнителя SQL позволяет отправить запрос как через пул, так и через
  транзакцию, выбранную менеджером по контексту.
- `internal/container` связывает реализации вручную. `main` загружает конфиг,
  запускает сервер и инициирует остановку по SIGINT или SIGTERM.

Создание записывает поездку и историю в одной транзакции. Частичный уникальный
индекс на активного водителя защищает от конкурентного создания.
Завершение выполняет `SELECT ... FOR UPDATE`, вызывает `Trip.Finish` и
сохраняет состояние и переход в журнал в той же транзакции. Ожидающий запрос
читает актуальное состояние строки, поэтому возвращает `409 trip_completed`.

## Тесты

```sh
make generate           # OpenAPI и моки gomock
make lint               # линтеры и форматтеры
make test               # unit-тесты с -race и покрытием по пакетам
make test-integration   # тесты на PostgreSQL с -race и тегом integration
```

Для интеграционных тестов нужны `DATABASE_URL` и применённые миграции
(`make migrate`). Тесты создают данные с уникальными UUID и удаляют их после
прогона. Проверяются конкурентное завершение и откат двух вставок, включая
случай отказа при вставке в журнал.

## Телеметрия

Порядок middleware — Recover → Tracing → Logging → Metrics → роутер.
HTTP-спан и метрики используют шаблон маршрута. Запросы PostgreSQL
инструментируются через otelpgx на пуле, без записи значений SQL-параметров.
JSON-логи идут в stdout, содержат `trace_id` и `span_id`; обработанная паника
пишется со стеком и возвращает `500 internal_error`. Метрики и трейсы
отправляются в коллектор асинхронно.

```sh
curl -sS -i -X POST http://localhost:8080/api/v1/trips \
  -H 'content-type: application/json' \
  -d '{"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c",
       "driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5",
       "start_point":{"latitude":59.9398,"longitude":30.3146},
       "end_point":{"latitude":59.9290,"longitude":30.3626},"price":1450}'
```

По `trace_id` из записи `request handled` откройте трейс в Jaeger. В нём
должны быть HTTP-спан и вложенные операции PostgreSQL. В Grafana Explore
выберите Prometheus и выполните запросы:

```promql
sum by (http_route, http_response_status_code) (rate(http_server_request_duration_seconds_count[2m]))
increase(tripgo_trips_created_total[2m])
```

### Проверка в окружении курса

- [Jaeger](docs/lab02/jaeger.png) показывает `POST /api/v1/trips` с ответом
  `201` и четырьмя дочерними спанами PostgreSQL. Идентификатор трассы
  `3b0a105be5ed7b73b7ceb150f2b04736` совпадает с полем `trace_id` в логе.
- [Grafana](docs/lab02/grafana.png) показывает p95 длительности по шаблонам
  маршрутов и счётчик созданных поездок после серии тестовых запросов.
- При остановленном коллекторе `/ready` ответил `200`, создание поездки —
  `201`, завершение — `200`.

```

`cmd/trip-service/main.go`:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Chillobass/fork_template/internal/config"
	"github.com/Chillobass/fork_template/internal/container"
	"github.com/Chillobass/fork_template/internal/infra/telemetry"
)

const serviceName = "trip-service"

func main() {
	cfg, err := config.Load()
	if err != nil {
		telemetry.NewLogger(serviceName, slog.LevelInfo).Error("configuration failed", "error", err)
		os.Exit(1)
	}
	logger := telemetry.NewLogger(serviceName, cfg.LogLevel)
	if err := run(cfg, logger); err != nil {
		logger.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := container.New(ctx, cfg, serviceName, logger)
	if err != nil {
		return err
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server started", "addr", cfg.HTTP.Addr)
		serverErr <- app.Server.ListenAndServe()
	}()

	select {
	case err = <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			err = fmt.Errorf("http server: %w", err)
		} else {
			err = nil
		}
	case <-ctx.Done():
	}

	stop()
	logger.Info("shutdown started", "timeout", cfg.ShutdownTimeout.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	err = errors.Join(err, app.Shutdown(shutdownCtx, logger))
	if err == nil {
		logger.Info("service stopped")
	}
	return err
}

```

`contracts/asyncapi/driver-updates.asyncapi.yaml`:

```yaml
asyncapi: 2.6.0
info:
  title: TripGo Driver Updates
  version: 1.0.0
  description: Опциональный WebSocket-контракт лабораторной работы 5.
servers:
  local:
    url: ws://localhost:8080
    protocol: ws
channels:
  /ws/v1/drivers/{driverId}:
    parameters:
      driverId:
        description: UUID водителя
        schema:
          type: string
          format: uuid
    subscribe:
      message:
        oneOf:
          - $ref: '#/components/messages/TripCreated'
          - $ref: '#/components/messages/TripCompleted'
components:
  messages:
    TripCreated:
      name: trip.created
      payload:
        type: object
        required: [type, trip_id, occurred_at]
        properties:
          type:
            const: trip.created
          trip_id:
            type: string
            format: uuid
          occurred_at:
            type: string
            format: date-time
    TripCompleted:
      name: trip.completed
      payload:
        type: object
        required: [type, trip_id, occurred_at]
        properties:
          type:
            const: trip.completed
          trip_id:
            type: string
            format: uuid
          occurred_at:
            type: string
            format: date-time

```

`contracts/events/README.md`:

```md
# Контракты событий

Оба топика — JSON. Формат выбран ради отладки: сообщение видно глазами в
Redpanda Console, туда же его можно положить руками, без утилит и генерации кода.

Общие правила:

- время — RFC 3339 в UTC: `2026-08-24T12:00:00Z`;
- деньги — целые рубли, как в HTTP-контракте;
- неизвестные поля потребитель игнорирует, а не падает: так мы сможем добавлять
  новые поля, не ломая тех, кто уже читает топик;
- удалять и переименовывать поля нельзя, это ломает потребителей.

---

## `trip.events.v1` — событие о завершении поездки

Публикует ваш сервис при переходе поездки в `completed`.

| Параметр | Значение |
|---|---|
| Ключ | `trip_id` |
| Заголовки | `event_type`, `event_id`, `content-type: application/json` |

```json
{
  "event_id": "9f1d2c8e-4b3a-4c1f-8e2d-77c0b5a1e934",
  "event_type": "trip.completed",
  "trip_id": "1f0a9c62-4a1c-4f2e-9d33-2a4bb0f0b111",
  "user_id": "5cb72c04-7650-45c9-a79b-bcdba0631e0c",
  "driver_id": "8860b315-ec86-42eb-a17c-7c163d721ff5",
  "price": 1450,
  "started_at": "2026-08-24T12:00:00Z",
  "finished_at": "2026-08-24T12:18:00Z",
  "occurred_at": "2026-08-24T12:18:00Z"
}
```

| Поле | Тип | Смысл |
|---|---|---|
| `event_id` | UUID | идентификатор события, новый на каждое сообщение |
| `event_type` | string | всегда `trip.completed` |
| `trip_id` | UUID | поездка |
| `user_id`, `driver_id` | UUID | участники |
| `price` | int | стоимость в целых рублях |
| `started_at`, `finished_at` | date-time | границы поездки |
| `occurred_at` | date-time | когда событие сформировано |

`event_id` нужен потребителю, чтобы отсеять повторную доставку: одно и то же
событие может прийти дважды.

---

## `trip.commands.v1` — команда создать поездку

Читает ваш сервис. В курсе сообщения кладутся руками через Redpanda Console.

| Параметр | Значение |
|---|---|
| Ключ | `driver_id` |
| Заголовки | `content-type: application/json` |
| DLQ | `trip.commands.v1.dlq` |
| Consumer-группа | `trip-service` |

```json
{
  "command_id": "3a7c1e55-2f90-4b6d-9a11-6d4e0f2b8c73",
  "user_id": "5cb72c04-7650-45c9-a79b-bcdba0631e0c",
  "driver_id": "8860b315-ec86-42eb-a17c-7c163d721ff5",
  "start_point": { "latitude": 59.9398, "longitude": 30.3146 },
  "end_point":   { "latitude": 59.9290, "longitude": 30.3626 },
  "price": 1450,
  "issued_at": "2026-08-24T12:00:00Z"
}
```

| Поле | Тип | Смысл |
|---|---|---|
| `command_id` | UUID | ключ идемпотентности, по нему отсекается повтор |
| `user_id`, `driver_id` | UUID | участники |
| `start_point`, `end_point` | object | координаты, как в HTTP-контракте |
| `price` | int | стоимость в целых рублях |
| `issued_at` | date-time | когда команда сформирована отправителем |

Валидация та же, что у `POST /api/v1/trips`. Сообщение, которое не прошло
валидацию или не разобралось как JSON, уезжает в DLQ.

```

`contracts/openapi/push-service.openapi.yaml`:

```yaml
openapi: 3.0.3
info:
  title: TripGo Push Service API
  version: 1.0.0
  description: >-
    HTTP-контракт сервиса-заглушки пушей. Используется в лабораторной работе 3.
    В лабораторной работе 4 тот же сервис вызывается по gRPC
    (contracts/proto/push/v1/push.proto).
servers:
- url: http://localhost:8090
tags:
- name: Push
- name: Admin
paths:
  /api/v1/push:
    post:
      tags:
      - Push
      summary: Отправить команду на пуш-уведомление
      operationId: sendPush
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PushRequest'
            example:
              request_id: 01J9Z6H7K8QW3M4N5P6R7S8T9V
              recipient_id: 8860b315-ec86-42eb-a17c-7c163d721ff5
              kind: REQUEST_POSITION
              data:
                trip_id: 1f0a9c62-4a1c-4f2e-9d33-2a4bb0f0b111
      responses:
        '202':
          description: Команда принята
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PushAccepted'
        '400':
          description: Некорректный запрос
        '429':
          description: Превышен лимит
          headers:
            Retry-After:
              description: Через сколько секунд повторять
              schema:
                type: integer
        '503':
          description: Временная недоступность
  /admin/behaviour:
    post:
      tags:
      - Admin
      summary: Настроить поведение заглушки на лету
      description: Используется в лабораторной работе 5 для проверки устойчивости.
      operationId: setBehaviour
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Behaviour'
      responses:
        '200':
          description: Настройки применены
        '400':
          description: Некорректные настройки
  /health:
    get:
      tags:
      - Push
      summary: Liveness
      operationId: pushHealth
      responses:
        '200':
          description: Процесс работает
components:
  schemas:
    PushKind:
      type: string
      enum:
      - REQUEST_POSITION
      - TRIP_CREATED
      - TRIP_COMPLETED
    PushRequest:
      type: object
      additionalProperties: false
      required:
      - request_id
      - recipient_id
      - kind
      properties:
        request_id:
          type: string
          description: Идентификатор запроса для корреляции логов.
        recipient_id:
          type: string
          format: uuid
          description: Кому шлём. Для REQUEST_POSITION — driver_id поездки.
        kind:
          $ref: '#/components/schemas/PushKind'
        title:
          type: string
        body:
          type: string
        data:
          type: object
          additionalProperties:
            type: string
    PushAccepted:
      type: object
      additionalProperties: false
      required:
      - message_id
      - accepted_at
      properties:
        message_id:
          type: string
        accepted_at:
          type: string
          format: date-time
    Behaviour:
      type: object
      additionalProperties: false
      properties:
        fail_rate:
          type: number
          format: double
          minimum: 0
          maximum: 1
        latency_ms:
          type: integer
          minimum: 0
        rate_limit_rps:
          type: integer
          minimum: 0

```

`contracts/openapi/trip-service.openapi.yaml`:

```yaml
openapi: 3.0.3
info:
  title: TripGo Trip Service API
  version: 1.1.0
  description: Кумулятивный контракт Trip Service начиная с лабораторной работы 2.
servers:
- url: http://localhost:8080
tags:
- name: Trips
- name: Positions
- name: Service
paths:
  /api/v1/trips:
    post:
      tags:
      - Trips
      summary: Создать поездку
      operationId: createTrip
      parameters:
      - $ref: '#/components/parameters/IdempotencyKey'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TripData'
            example:
              user_id: 5cb72c04-7650-45c9-a79b-bcdba0631e0c
              driver_id: 8860b315-ec86-42eb-a17c-7c163d721ff5
              start_point:
                latitude: 59.9398
                longitude: 30.3146
              end_point:
                latitude: 59.929
                longitude: 30.3626
              price: 1450
      responses:
        '201':
          description: Поездка создана
          headers:
            Location:
              description: URI созданной поездки
              schema:
                type: string
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Trip'
        '400':
          $ref: '#/components/responses/BadRequest'
        '409':
          $ref: '#/components/responses/CreateTripConflict'
        '500':
          $ref: '#/components/responses/InternalError'
  /api/v1/trips/{tripId}:
    get:
      tags:
      - Trips
      summary: Получить поездку
      operationId: getTrip
      parameters:
      - $ref: '#/components/parameters/TripId'
      responses:
        '200':
          description: Поездка
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Trip'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/TripNotFound'
        '500':
          $ref: '#/components/responses/InternalError'
  /api/v1/trips/{tripId}/finish:
    post:
      tags:
      - Trips
      summary: Завершить поездку
      description: Переводит активную поездку в статус completed.
      operationId: finishTrip
      parameters:
      - $ref: '#/components/parameters/TripId'
      responses:
        '200':
          description: Поездка завершена
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Trip'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/TripNotFound'
        '409':
          $ref: '#/components/responses/TripCompleted'
        '500':
          $ref: '#/components/responses/InternalError'
  /health:
    get:
      tags:
      - Service
      summary: Liveness
      operationId: health
      responses:
        '200':
          description: Процесс работает
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'
  /ready:
    get:
      tags:
      - Service
      summary: Readiness
      operationId: ready
      responses:
        '200':
          description: Сервис готов
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'
        '503':
          description: PostgreSQL недоступен
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'
  /api/v1/trips/{tripId}/positions:
    post:
      tags:
      - Positions
      summary: Сохранить координату поездки
      operationId: createTripPosition
      description: Координаты разрешено добавлять только в активную поездку.
      parameters:
      - $ref: '#/components/parameters/TripId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PositionData'
            example:
              coordinates:
                latitude: 59.9343
                longitude: 30.3351
              recorded_at: '2026-09-01T10:05:00Z'
      responses:
        '201':
          description: Координата сохранена
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TripPosition'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/TripNotFound'
        '409':
          $ref: '#/components/responses/TripCompleted'
        '500':
          $ref: '#/components/responses/InternalError'
    get:
      tags:
      - Positions
      summary: Получить маршрут поездки
      operationId: listTripPositions
      description: Возвращает все точки маршрута, отсортированные по recorded_at по возрастанию.
      parameters:
      - $ref: '#/components/parameters/TripId'
      responses:
        '200':
          description: Маршрут
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TripPositionsResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/TripNotFound'
        '500':
          $ref: '#/components/responses/InternalError'
components:
  parameters:
    IdempotencyKey:
      name: Idempotency-Key
      in: header
      required: false
      description: >-
        Ключ идемпотентности, генерирует клиент. Повторный запрос с тем же ключом
        и тем же телом возвращает уже созданную поездку и 200, с тем же ключом и
        другим телом — 409 idempotency_conflict. Заголовок относится к заданию со
        звёздочкой лабораторной работы 1, в основной части можно игнорировать.
      schema:
        type: string
        format: uuid
    TripId:
      name: tripId
      in: path
      required: true
      description: UUID поездки
      schema:
        type: string
        format: uuid
  schemas:
    Coordinates:
      type: object
      additionalProperties: false
      description: Географические координаты WGS 84.
      required:
      - latitude
      - longitude
      properties:
        latitude:
          type: number
          format: double
          minimum: -90
          maximum: 90
          example: 59.9398
        longitude:
          type: number
          format: double
          minimum: -180
          maximum: 180
          example: 30.3146
    TripData:
      type: object
      additionalProperties: false
      required:
      - user_id
      - driver_id
      - start_point
      - end_point
      - price
      properties:
        user_id:
          type: string
          format: uuid
        driver_id:
          type: string
          format: uuid
        start_point:
          $ref: '#/components/schemas/Coordinates'
        end_point:
          $ref: '#/components/schemas/Coordinates'
        price:
          type: integer
          format: int64
          minimum: 0
          example: 1450
          description: 'Стоимость поездки в целых рублях. Это учебное упрощение: копейки не используются.'
    TripStatus:
      type: string
      enum:
      - active
      - completed
    Trip:
      allOf:
      - $ref: '#/components/schemas/TripData'
      - type: object
        additionalProperties: false
        required:
        - id
        - status
        - started_at
        properties:
          id:
            type: string
            format: uuid
          status:
            $ref: '#/components/schemas/TripStatus'
          started_at:
            type: string
            format: date-time
          finished_at:
            type: string
            format: date-time
            nullable: true
          last_position_at:
            type: string
            format: date-time
            nullable: true
    PositionData:
      type: object
      additionalProperties: false
      required:
      - coordinates
      - recorded_at
      properties:
        coordinates:
          $ref: '#/components/schemas/Coordinates'
        recorded_at:
          type: string
          format: date-time
          description: Время фиксации координаты на мобильном устройстве.
    TripPosition:
      allOf:
      - $ref: '#/components/schemas/PositionData'
      - type: object
        additionalProperties: false
        required:
        - id
        - trip_id
        properties:
          id:
            type: integer
            format: int64
          trip_id:
            type: string
            format: uuid
    TripPositionsResponse:
      type: object
      additionalProperties: false
      required:
      - positions
      properties:
        positions:
          type: array
          items:
            $ref: '#/components/schemas/TripPosition'
    HealthResponse:
      type: object
      additionalProperties: false
      required:
      - status
      properties:
        status:
          type: string
          enum:
          - ok
          - unavailable
    Problem:
      type: object
      description: Problem Details for HTTP APIs (RFC 9457) с дополнительным полем code.
      required:
      - type
      - title
      - status
      - code
      properties:
        type:
          type: string
        title:
          type: string
        status:
          type: integer
          format: int32
        detail:
          type: string
        instance:
          type: string
        code:
          type: string
  responses:
    BadRequest:
      description: Некорректный запрос
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
          example:
            type: https://tripgo.example/problems/invalid-request
            title: Invalid request
            status: 400
            detail: Request validation failed
            code: invalid_request
    TripNotFound:
      description: Поездка не найдена
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
          example:
            type: https://tripgo.example/problems/trip-not-found
            title: Trip not found
            status: 404
            detail: Trip was not found
            code: trip_not_found
    TripCompleted:
      description: Поездка уже завершена
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
          example:
            type: https://tripgo.example/problems/trip-completed
            title: Trip completed
            status: 409
            detail: Operation is not allowed for a completed trip
            code: trip_completed
    CreateTripConflict:
      description: >-
        Конфликт при создании поездки: у водителя уже есть активная поездка либо
        ключ идемпотентности повторно использован с другим телом запроса.
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
          examples:
            driverBusy:
              summary: У водителя уже есть активная поездка
              value:
                type: https://tripgo.example/problems/driver-busy
                title: Driver busy
                status: 409
                detail: Driver already has an active trip
                code: driver_busy
            idempotencyConflict:
              summary: Тот же ключ идемпотентности с другим телом
              value:
                type: https://tripgo.example/problems/idempotency-conflict
                title: Idempotency conflict
                status: 409
                detail: Idempotency-Key was already used with a different request body
                code: idempotency_conflict
    InternalError:
      description: Внутренняя ошибка
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
          example:
            type: https://tripgo.example/problems/internal-error
            title: Internal Server Error
            status: 500
            detail: Internal server error
            code: internal_error

```

`contracts/proto/push/v1/push.proto`:

```proto
syntax = "proto3";

package tripgo.push.v1;

option go_package = "github.com/course-go-autumn-2026/homework/gen/push/v1;pushv1";

import "google/protobuf/timestamp.proto";

service PushService {
  // Принимает команду на отправку push-уведомления.
  // Реальная интеграция с APNs/FCM в учебный scope не входит.
  rpc SendPush(SendPushRequest) returns (SendPushResponse);
}

enum PushKind {
  PUSH_KIND_UNSPECIFIED = 0;
  PUSH_KIND_REQUEST_POSITION = 1;
  PUSH_KIND_TRIP_CREATED = 2;
  PUSH_KIND_TRIP_COMPLETED = 3;
}

message SendPushRequest {
  string request_id = 1;
  string recipient_id = 2;
  PushKind kind = 3;
  string title = 4;
  string body = 5;
  map<string, string> data = 6;
}

message SendPushResponse {
  string message_id = 1;
  google.protobuf.Timestamp accepted_at = 2;
}

```

`contracts/proto/trip/v1/trip_service.proto`:

```proto
syntax = "proto3";

package tripgo.trip.v1;

option go_package = "github.com/example/tripgo/gen/trip/v1;tripv1";

import "google/protobuf/timestamp.proto";

// Заготовка контракта для лабораторной работы 4.
// В отличие от push.proto и trip_events.proto, это ВАШ контракт:
// его можно и нужно дополнять. Требования — tasks/04-grpc-and-events, §4.1.
service TripService {
  rpc CreateTrip(CreateTripRequest) returns (Trip);
  rpc GetTrip(GetTripRequest) returns (Trip);
  rpc FinishTrip(FinishTripRequest) returns (Trip);
  rpc AddPosition(AddPositionRequest) returns (Position);
  rpc ListPositions(ListPositionsRequest) returns (ListPositionsResponse);

  // Задание со звёздочкой: подписка на координаты поездки.
  // rpc WatchTripPositions(WatchTripPositionsRequest) returns (stream Position);
}

message Coordinates {
  double latitude = 1;
  double longitude = 2;
}

enum TripStatus {
  TRIP_STATUS_UNSPECIFIED = 0;
  TRIP_STATUS_ACTIVE = 1;
  TRIP_STATUS_COMPLETED = 2;
}

message Trip {
  string id = 1;
  string user_id = 2;
  string driver_id = 3;
  Coordinates start_point = 4;
  Coordinates end_point = 5;
  int64 price = 6;
  TripStatus status = 7;
  google.protobuf.Timestamp started_at = 8;
  google.protobuf.Timestamp finished_at = 9;
  google.protobuf.Timestamp last_position_at = 10;
}

message Position {
  int64 id = 1;
  string trip_id = 2;
  Coordinates coordinates = 3;
  google.protobuf.Timestamp recorded_at = 4;
}

message CreateTripRequest {
  string user_id = 1;
  string driver_id = 2;
  Coordinates start_point = 3;
  Coordinates end_point = 4;
  int64 price = 5;
}

message GetTripRequest {
  string trip_id = 1;
}

message FinishTripRequest {
  string trip_id = 1;
}

message AddPositionRequest {
  string trip_id = 1;
  Coordinates coordinates = 2;
  google.protobuf.Timestamp recorded_at = 3;
}

message ListPositionsRequest {
  string trip_id = 1;
  google.protobuf.Timestamp from = 2;
  google.protobuf.Timestamp to = 3;
  int32 limit = 4;
}

message ListPositionsResponse {
  repeated Position positions = 1;
}

```

`contracts/schema.md`:

```md
# Схема данных

Здесь описаны все таблицы курса: что хранят, из каких полей состоят и какие
правила закреплены в самой схеме. Таблицы появляются по мере работ, в каждом
разделе указано, в какой.

Готовых миграций мы не даём — их вы пишете сами. `CREATE TABLE` ниже приведены,
чтобы не гадать про типы и ограничения: перенесите их в свои миграции, добавьте
`down` и, если считаете нужным, поправьте под своё решение. Имена таблиц и
колонок менять нельзя, на них завязаны следующие работы.

---

## `trips` — поездка

Появляется в работе 1. Колонка `last_position_at` добавляется в работе 3.

```sql
CREATE TABLE trips (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL,
    driver_id       UUID NOT NULL,

    start_latitude  DOUBLE PRECISION NOT NULL CHECK (start_latitude BETWEEN -90 AND 90),
    start_longitude DOUBLE PRECISION NOT NULL CHECK (start_longitude BETWEEN -180 AND 180),
    end_latitude    DOUBLE PRECISION NOT NULL CHECK (end_latitude BETWEEN -90 AND 90),
    end_longitude   DOUBLE PRECISION NOT NULL CHECK (end_longitude BETWEEN -180 AND 180),

    price           BIGINT NOT NULL CHECK (price >= 0),
    status          TEXT NOT NULL CHECK (status IN ('active', 'completed')),

    started_at      TIMESTAMPTZ NOT NULL,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (
        (status = 'active'    AND finished_at IS NULL) OR
        (status = 'completed' AND finished_at IS NOT NULL)
    )
);

CREATE INDEX trips_status_started_at_idx ON trips (status, started_at);
```

| Поле | Что хранит | Кто заполняет |
|---|---|---|
| `id` | идентификатор поездки | сервис при создании, из запроса не принимается |
| `user_id` | пассажир | приходит в запросе |
| `driver_id` | водитель | приходит в запросе |
| `start_*`, `end_*` | точки подачи и назначения, градусы | приходит в запросе |
| `price` | стоимость в целых рублях | приходит в запросе, см. упрощение в `domain.md` |
| `status` | `active` или `completed` | сервис |
| `started_at` | когда поездка создана | сервис, не клиент |
| `finished_at` | когда завершена, пусто пока активна | сервис при завершении |
| `last_position_at` | время последней координаты | сервис при сохранении координаты, работа 3 |
| `created_at`, `updated_at` | техполя | сервис |

Почему `price` целым числом: деньги во `float` не хранят, потеряете копейки на
округлении. У нас это ещё и целые рубли — учебное упрощение, в проде хранили бы
копейки.

Последний `CHECK` держит связку статуса и времени завершения: завершённая
поездка без `finished_at` или активная с ним в базу не попадут.

**Ограничение, которое надо добавить самим:** у одного водителя не может быть
двух активных поездок одновременно. Оно должно держаться при одновременных
запросах — два параллельных создания на одного водителя дают одну поездку и одну
ошибку. Проверка `SELECT` перед `INSERT` этого не даёт: оба запроса пройдут
проверку раньше, чем любой из них вставит строку.

В работе 3 к таблице добавляется колонка:

```sql
ALTER TABLE trips ADD COLUMN last_position_at TIMESTAMPTZ;
```

---

## `trip_status_history` — журнал переходов статуса

Появляется в работе 1. Нужен, чтобы создание поездки писало в две таблицы в одной
транзакции, — на этом отрабатывается менеджер транзакций.

```sql
CREATE TABLE trip_status_history (
    id          BIGSERIAL PRIMARY KEY,
    trip_id     UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status   TEXT NOT NULL CHECK (to_status IN ('active', 'completed')),
    reason      TEXT,
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX trip_status_history_trip_changed_idx
    ON trip_status_history (trip_id, changed_at);
```

`from_status` пуст при создании поездки: прежнего статуса нет. `reason` — текст
для человека, чем вызван переход.

---

## `trip_positions` — точки маршрута

Появляется в работе 3.

```sql
CREATE TABLE trip_positions (
    id          BIGSERIAL PRIMARY KEY,
    trip_id     UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    latitude    DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude   DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX trip_positions_trip_recorded_idx
    ON trip_positions (trip_id, recorded_at, id);
```

`recorded_at` — когда координата зафиксирована на устройстве, `created_at` —
когда её принял сервис. Это разные моменты: устройство копит точки в офлайне и
шлёт их пачкой позже. Маршрут строим по `recorded_at`, поэтому индекс по нему.

Индекс под выборку поездок для опроса вы добавляете сами — это отдельный пункт
чек-листа работы 3.

---

## `processed_commands` — обработанные команды

Появляется в работе 4. Хранит идентификаторы команд, которые сервис уже
обработал, чтобы повторная доставка не создала вторую поездку.

```sql
CREATE TABLE processed_commands (
    command_id   UUID PRIMARY KEY,
    trip_id      UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX processed_commands_processed_at_idx
    ON processed_commands (processed_at);
```

`command_id` — первичный ключ: попытка вставить его второй раз даст конфликт, по
нему и отличают повтор. Вставка идёт в одной транзакции с созданием поездки.

---

## `outbox_events` — исходящие события

Появляется в работе 5, только в варианте A.

```sql
CREATE TABLE outbox_events (
    id              UUID PRIMARY KEY,
    aggregate_type  TEXT NOT NULL,
    aggregate_id    UUID NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ,
    attempts        INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error      TEXT
);

CREATE INDEX outbox_events_pending_idx
    ON outbox_events (next_attempt_at, created_at)
    WHERE published_at IS NULL;
```

| Поле | Что хранит |
|---|---|
| `aggregate_type`, `aggregate_id` | к какой сущности относится событие: `trip` и его `id` |
| `event_type` | что произошло, например `trip.completed` |
| `payload` | тело события, которое уедет в топик |
| `published_at` | пусто, пока не опубликовано; по этому полю и ищут неотправленные |
| `attempts`, `next_attempt_at`, `last_error` | состояние повторов |

Индекс частичный, только по неопубликованным: опубликованных со временем станет
большинство, а publisher их никогда не читает.

```

`environment.toml`:

```toml
schema_version = 1
lab = 2
components = ["postgres", "observability"]

```

`go.mod`:

```mod
module github.com/Chillobass/fork_template

go 1.27.1

tool (
	github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	github.com/pressly/goose/v3/cmd/goose
	go.uber.org/mock/mockgen
)

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/exaring/otelpgx v0.13.0
	github.com/felixge/httpsnoop v1.1.0
	github.com/go-chi/chi/v5 v5.3.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.11.0
	github.com/oapi-codegen/nullable v1.2.0
	github.com/oapi-codegen/runtime v1.7.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0
	go.opentelemetry.io/otel v1.47.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.47.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.47.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.47.0
	go.opentelemetry.io/otel/metric v1.47.0
	go.opentelemetry.io/otel/sdk v1.47.0
	go.opentelemetry.io/otel/sdk/metric v1.47.0
	go.opentelemetry.io/otel/trace v1.47.0
	go.uber.org/mock v0.6.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.23.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.9.0 // indirect
	github.com/ClickHouse/ch-go v0.74.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.48.0 // indirect
	github.com/andybalholm/brotli v1.2.3 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.15 // indirect
	github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/elastic/go-sysinfo v1.15.5 // indirect
	github.com/elastic/go-windows v1.0.2 // indirect
	github.com/getkin/kin-openapi v0.142.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.8.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v1.0.0 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mfridman/xflag v0.1.0 // indirect
	github.com/microsoft/go-mssqldb v1.11.0 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.8.0 // indirect
	github.com/oasdiff/yaml v0.1.1 // indirect
	github.com/oasdiff/yaml3 v0.0.14 // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.29 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pressly/goose/v3 v3.28.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/sethvargo/go-retry v0.4.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/speakeasy-api/jsonpath v0.6.3 // indirect
	github.com/speakeasy-api/openapi v1.24.0 // indirect
	github.com/tursodatabase/libsql-client-go v0.0.0-20260528064733-9d5d30a29a60 // indirect
	github.com/vertica/vertica-sql-go v1.3.8 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/ydb-platform/ydb-go-genproto v0.0.0-20260810123728-f0c151ab31b9 // indirect
	github.com/ydb-platform/ydb-go-sdk/v3 v3.151.1 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/exp v0.0.0-20260824195058-e88cd73687aa // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261104224025-d1e00e9f33ab // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b // indirect
	google.golang.org/grpc v1.83.2 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	howett.net/plist v1.0.1 // indirect
	modernc.org/libc v1.75.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.57.0 // indirect
)

```

`go.sum`:

```sum
cloud.google.com/go v0.26.0/go.mod h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw=
cloud.google.com/go v0.34.0/go.mod h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw=
filippo.io/edwards25519 v1.2.0 h1:crnVqOiS4jqYleHd9vaKZ+HKtHfllngJIiOpNpoJsjo=
filippo.io/edwards25519 v1.2.0/go.mod h1:xzAOLCNug/yB62zG1bQ8uziwrIqIuxhctzJT18Q77mc=
github.com/Azure/azure-sdk-for-go/sdk/azcore v1.23.1 h1:zvXfGJCWvywnCA814d8ZiVyt+fm9nnTE8xSb99zRyfo=
github.com/Azure/azure-sdk-for-go/sdk/azcore v1.23.1/go.mod h1:iptorS+VYKFL2N6PnebpS91dubG35eAOEERnT4PJbQU=
github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.1 h1:u93s+zU2JD62im61Bm5CZIc1ZrOJaIAWEg0WOrMVkEo=
github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.1/go.mod h1:oXtinPO4OLj9d1DOTrqrL1oRwGhcqadvAmrl6wTeGlk=
github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache v0.4.0 h1:xFaZZ+IubdftrDHnGGwZ6QvQ3KHTtWl2MCK+GMt2vxs=
github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache v0.4.0/go.mod h1:mCBhUhlMjLLJKr5aqw2TNS/VqJOie8MzWq3DAMJeKso=
github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 h1:fhqpLE3UEXi9lPaBRpQ6XuRW0nU7hgg4zlmZZa+a9q4=
github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0/go.mod h1:7dCRMLwisfRH3dBupKeNCioWYUZ4SS09Z14H+7i8ZoY=
github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys v1.5.0 h1:MaKvxE6D0KkjOg6Wd9M00iqP5PR0kUxCfiezes4JweM=
github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys v1.5.0/go.mod h1:i2h9fsTFKZorh8RdV2IcSUf/Qj98GlTkrTvUbX/s8as=
github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 h1:nCYfgcSyHZXJI8J0IWE5MsCGlb2xp9fJiXyxWgmOFg4=
github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0/go.mod h1:ucUjca2JtSZboY8IoUqyQyuuXvwbMBVwFOm0vdQPNhA=
github.com/AzureAD/microsoft-authentication-extensions-for-go/cache v0.1.1 h1:WJTmL004Abzc5wDB5VtZG2PJk5ndYDgVacGqfirKxjM=
github.com/AzureAD/microsoft-authentication-extensions-for-go/cache v0.1.1/go.mod h1:tCcJZ0uHAmvjsVYzEFivsRTN00oz5BEsRgQHu5JZ9WE=
github.com/AzureAD/microsoft-authentication-library-for-go v1.9.0 h1:MDT4FxAPve5FnYn6vOL1r7RCRDG+l9cI7a5LlCuHsqA=
github.com/AzureAD/microsoft-authentication-library-for-go v1.9.0/go.mod h1:Y33QHnf0FfdVewFFISOGe20mkZbxX4H839o955/PoeI=
github.com/BurntSushi/toml v0.3.1/go.mod h1:xHWCNGjB5oqiDr8zfno3MHue2Ht5sIBksp03qcyfWMU=
github.com/ClickHouse/ch-go v0.74.0 h1:uYs2m4wIt0ZHSM1E72rg0maCfzhR2V3xWb/vZEgpeWE=
github.com/ClickHouse/ch-go v0.74.0/go.mod h1:sZ/r+8ttZMjyrP9PuFbgoVbth1ywIu2LIQNA2vgko6M=
github.com/ClickHouse/clickhouse-go/v2 v2.48.0 h1:auzd4VkapQYhQF8F2Gog7s3x78Bi1JZmByxGbrw3C+4=
github.com/ClickHouse/clickhouse-go/v2 v2.48.0/go.mod h1:lBjUCPRG6RpRQdMbkXq+JV8rY0/O5lw+Z7jShgReFjM=
github.com/Masterminds/squirrel v1.5.4 h1:uUcX/aBc8O7Fg9kaISIUsHXdKuqehiXAMQTYX8afzqM=
github.com/Masterminds/squirrel v1.5.4/go.mod h1:NNaOrjSoIDfDA40n7sr2tPNZRfjzjA400rg+riTZj10=
github.com/RaveNoX/go-jsoncommentstrip v1.0.0/go.mod h1:78ihd09MekBnJnxpICcwzCMzGrKSKYe4AqU6PDYYpjk=
github.com/andybalholm/brotli v1.2.3 h1:8H1qwOkl2LPfjf3YezB90JnCliZb6SInJ/OJkEbA5NQ=
github.com/andybalholm/brotli v1.2.3/go.mod h1:rzTDkvFWvIrjDXZHkuS16NPggd91W3kUSvPlQ1pLaKY=
github.com/antihax/optional v1.0.0/go.mod h1:uupD/76wgC+ih3iEmQUL+0Ugr19nfwCT1kdvxnR2qWY=
github.com/antlr4-go/antlr/v4 v4.13.1 h1:SqQKkuVZ+zWkMMNkjy5FZe5mr5WURWnlpmOuzYWrPrQ=
github.com/antlr4-go/antlr/v4 v4.13.1/go.mod h1:GKmUxMtwp6ZgGwZSva4eWPC5mS6vUAmOABFgjdkM7Nw=
github.com/apapsch/go-jsonmerge/v2 v2.0.0 h1:axGnT1gRIfimI7gJifB699GoE/oq+F2MU7Dml6nw9rQ=
github.com/apapsch/go-jsonmerge/v2 v2.0.0/go.mod h1:lvDnEdqiQrp0O42VQGgmlKpxL1AP2+08jFMw88y4klk=
github.com/bmatcuk/doublestar v1.1.1/go.mod h1:UD6OnuiIn0yFxxA2le/rnRU1G4RaI4UvFv1sNto9p6w=
github.com/cenkalti/backoff/v5 v5.0.3 h1:ZN+IMa753KfX5hd8vVaMixjnqRZ3y8CuJKRKj1xcsSM=
github.com/cenkalti/backoff/v5 v5.0.3/go.mod h1:rkhZdG3JZukswDf7f0cwqPNk4K0sa+F97BxZthm/crw=
github.com/census-instrumentation/opencensus-proto v0.2.1/go.mod h1:f6KPmirojxKA12rnyqOA5BBL4O983OfeGPqjHWSTneU=
github.com/cespare/xxhash/v2 v2.1.1/go.mod h1:VGX0DQ3Q6kWi7AoAeZDth3/j3BFtOZR5XLFGgcrjCOs=
github.com/cespare/xxhash/v2 v2.3.0 h1:UL815xU9SqsFlibzuggzjXhog7bL6oX9BbNZnL2UFvs=
github.com/cespare/xxhash/v2 v2.3.0/go.mod h1:VGX0DQ3Q6kWi7AoAeZDth3/j3BFtOZR5XLFGgcrjCOs=
github.com/chzyer/logex v1.1.10/go.mod h1:+Ywpsq7O8HXn0nuIou7OrIPyXbp3wmkHB+jjWRnGsAI=
github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e/go.mod h1:nSuG5e5PlCu98SY8svDHJxuZscDgtXS6KTTbou5AhLI=
github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1/go.mod h1:Q3SI9o4m/ZMnBNeIyt5eFwwo7qiLfzFZmjNmxjkiQlU=
github.com/client9/misspell v0.3.4/go.mod h1:qj6jICC3Q7zFZvVWo7KLAzC3yx5G7kyvSDkc90ppPyw=
github.com/cncf/udpa/go v0.0.0-20191209042840-269d4d468f6f/go.mod h1:M8M6+tZqaGXZJjfX53e64911xZQV5JYwmTeXPW+k8Sc=
github.com/cncf/udpa/go v0.0.0-20201120205902-5459f2c99403/go.mod h1:WmhPx2Nbnhtbo57+VJT5O0JRkEi1Wbu0z5j0R8u5Hbk=
github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4/go.mod h1:6pvJx4me5XPnfI9Z40ddWsdw2W/uZgQLFXToKeRcDiI=
github.com/cncf/xds/go v0.0.0-20210922020428-25de7278fc84/go.mod h1:eXthEFrGJvWHgFFCl3hGmgk+/aYT6PnTQLykKQRLhEs=
github.com/cncf/xds/go v0.0.0-20211001041855-01bcc9b48dfe/go.mod h1:eXthEFrGJvWHgFFCl3hGmgk+/aYT6PnTQLykKQRLhEs=
github.com/cncf/xds/go v0.0.0-20211011173535-cb28da3451f1/go.mod h1:eXthEFrGJvWHgFFCl3hGmgk+/aYT6PnTQLykKQRLhEs=
github.com/coder/websocket v1.8.15 h1:6B2JPeOGlpff2Uz6vOEH1Vzpi0iUz20A+lPVhPHtNUA=
github.com/coder/websocket v1.8.15/go.mod h1:NX3SzP+inril6yawo5CQXx8+fk145lPDC6pumgx0mVg=
github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/dlclark/regexp2 v1.11.4 h1:rPYF9/LECdNymJufQKmri9gV604RvvABwgOA8un7yAo=
github.com/dlclark/regexp2 v1.11.4/go.mod h1:DHkYz0B9wPfa6wondMfaivmHpzrQ3v9q8cnmRbL6yW8=
github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960/go.mod h1:9HQzr9D/0PGwMEbC3d5AB7oi67+h4TsQqItC1GVYG58=
github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 h1:PRxIJD8XjimM5aTknUK9w6DHLDox2r2M3DI4i2pnd3w=
github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936/go.mod h1:ttYvX5qlB+mlV1okblJqcSMtR4c52UKxDiX9GRBS8+Q=
github.com/dustin/go-humanize v1.0.1 h1:GzkhY7T5VNhEkwH0PVJgjz+fX1rhBrR7pRT3mDkpeCY=
github.com/dustin/go-humanize v1.0.1/go.mod h1:Mu1zIs6XwVuF/gI1OepvI0qD18qycQx+mFykh5fBlto=
github.com/elastic/go-sysinfo v1.15.5 h1:fCVUDmjHgljLUQCygherMnsRRJ9AkuAQIywTL7dEH28=
github.com/elastic/go-sysinfo v1.15.5/go.mod h1:ZBVXmqS368dOn/jvijV/zHLfakWTYHBZPk3G244lHrU=
github.com/elastic/go-windows v1.0.2 h1:yoLLsAsV5cfg9FLhZ9EXZ2n2sQFKeDYrHenkcivY4vI=
github.com/elastic/go-windows v1.0.2/go.mod h1:bGcDpBzXgYSqM0Gx3DM4+UxFj300SZLixie9u9ixLM8=
github.com/envoyproxy/go-control-plane v0.9.0/go.mod h1:YTl/9mNaCwkRvm6d1a2C3ymFceY/DCBVvsKhRF0iEA4=
github.com/envoyproxy/go-control-plane v0.9.1-0.20191026205805-5f8ba28d4473/go.mod h1:YTl/9mNaCwkRvm6d1a2C3ymFceY/DCBVvsKhRF0iEA4=
github.com/envoyproxy/go-control-plane v0.9.4/go.mod h1:6rpuAdCZL397s3pYoYcLgu1mIlRU8Am5FuJP05cCM98=
github.com/envoyproxy/go-control-plane v0.9.9-0.20201210154907-fd9021fe5dad/go.mod h1:cXg6YxExXjJnVBQHBLXeUAgxn2UodCpnH306RInaBQk=
github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1/go.mod h1:KJwIaB5Mv44NWtYuAOFCVOjcI94vtpEz2JU/D2v6IjE=
github.com/envoyproxy/protoc-gen-validate v0.1.0/go.mod h1:iSmxcyjqTsJpI2R4NaDN7+kN2VEUnK/pcBlmesArF7c=
github.com/exaring/otelpgx v0.13.0 h1:Vi0I2GWRS3nnVzgr73L/XbSLqthR0h6A1rm4Wt3W4aA=
github.com/exaring/otelpgx v0.13.0/go.mod h1:3OojrUKhhy3lTbYIMBijP3YjMey/jo14eHAW5cXcUdk=
github.com/felixge/httpsnoop v1.1.0 h1:3YtUj32ZZkqZtt3sZZsClsymw/QDuVfpNhoA31zeORc=
github.com/felixge/httpsnoop v1.1.0/go.mod h1:Zqxgdd+1Rkcz8euOqdr7lqgCRJztwr5hp9vDSi5UZCE=
github.com/fsnotify/fsnotify v1.4.7/go.mod h1:jwhsz4b93w/PPRr/qN1Yymfu8t87LnFCMoQvtojpjFo=
github.com/fsnotify/fsnotify v1.4.9 h1:hsms1Qyu0jgnwNXIxa+/V/PDsU6CfLf6CNO8H7IWoS4=
github.com/fsnotify/fsnotify v1.4.9/go.mod h1:znqG4EE+3YCdAaPaxE2ZRY/06pZUdp0tY4IgpuI1SZQ=
github.com/getkin/kin-openapi v0.142.0 h1:izj0vBdFprMhitfzaX8sTqztsEQyvwhssBoB6n8NO7w=
github.com/getkin/kin-openapi v0.142.0/go.mod h1:3BH9M9XDe/y9M5DSvEocVYAYq1w0qrhJHjC/vZi0AaY=
github.com/ghodss/yaml v1.0.0/go.mod h1:4dBDuWmgqj2HViK6kFavaiC9ZROes6MMH2rRYeMEF04=
github.com/go-chi/chi/v5 v5.3.2 h1:5YQkICvTCSZ25hoRsyJazN0scjzKGiu4VAUc7H1o1nY=
github.com/go-chi/chi/v5 v5.3.2/go.mod h1:R+tYY2hNuVUUjxoPtqUdgBqevM9s9njzkTLutVsOCto=
github.com/go-faster/city v1.0.1 h1:4WAxSZ3V2Ws4QRDrscLEDcibJY8uf41H6AhXDrNDcGw=
github.com/go-faster/city v1.0.1/go.mod h1:jKcUJId49qdW3L1qKHH/3wPeUstCVpVSXTM6vO3VcTw=
github.com/go-faster/errors v0.8.0 h1:9T9eJrM+72dFk7n4DfhuaDDe6cyuFCSW2oNUkN77Yqc=
github.com/go-faster/errors v0.8.0/go.mod h1:5ySTjWFiphBs07IKuiL69nxdfd5+fzh1u7FPGZP2quo=
github.com/go-logr/logr v1.2.2/go.mod h1:jdQByPbusPIv2/zmleS9BjJVeZ6kBagPoEUsqbVz/1A=
github.com/go-logr/logr v1.4.4 h1:tG4xh9yMsRCAiodLVTxyrkzSZ9+o0L1Kg/+cPVcbP/8=
github.com/go-logr/logr v1.4.4/go.mod h1:9T104GzyrTigFIr8wt5mBrctHMim0Nb2HLGrmQ40KvY=
github.com/go-logr/stdr v1.2.2 h1:hSWxHoqTgW2S2qGc0LTAI563KZ5YKYRhT3MFKZMbjag=
github.com/go-logr/stdr v1.2.2/go.mod h1:mMo/vtBO5dYbehREoey6XUKy/eSumjCCveDpRre4VKE=
github.com/go-openapi/jsonpointer v1.0.0 h1:kR9tHqY0CtZaOPVFm622dPVNhrvYpwr4uCxgL3h1H8s=
github.com/go-openapi/jsonpointer v1.0.0/go.mod h1:Z3rw7dWu1p9IgitXCFamSlA5lmDiklEB6vkaxcNZW5Y=
github.com/go-openapi/testify/v2 v2.6.0 h1:5PKH2HE7YJ/LuRPQGvSxBRlFXNQhSetBLlGAgUEu3ug=
github.com/go-openapi/testify/v2 v2.6.0/go.mod h1:SgsVHtfooshd0tublTtJ50FPKhujf47YRqauXXOUxfw=
github.com/go-sql-driver/mysql v1.10.0 h1:Q+1LV8DkHJvSYAdR83XzuhDaTykuDx0l6fkXxoWCWfw=
github.com/go-sql-driver/mysql v1.10.0/go.mod h1:M+cqaI7+xxXGG9swrdeUIoPG3Y3KCkF0pZej+SK+nWk=
github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0/go.mod h1:fyg7847qk6SyHyPtNmDHnmrv/HOrqktSC+C9fM+CJOE=
github.com/golang-jwt/jwt/v4 v4.5.2 h1:YtQM7lnr8iZ+j5q71MGKkNw9Mn7AjHM68uc9g5fXeUI=
github.com/golang-jwt/jwt/v4 v4.5.2/go.mod h1:m21LjoU+eqJr34lmDMbreY2eSTRJ1cv77w39/MY0Ch0=
github.com/golang-jwt/jwt/v5 v5.3.1 h1:kYf81DTWFe7t+1VvL7eS+jKFVWaUnK9cB1qbwn63YCY=
github.com/golang-jwt/jwt/v5 v5.3.1/go.mod h1:fxCRLWMO43lRc8nhHWY6LGqRcf+1gQWArsqaEUEa5bE=
github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 h1:au07oEsX2xN0ktxqI+Sida1w446QrXBRJ0nee3SNZlA=
github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9/go.mod h1:8vg3r2VgvsThLBIFL93Qb5yWzgyZWhEmBwUJWevAkK0=
github.com/golang-sql/sqlexp v0.1.0 h1:ZCD6MBpcuOVfGVqsEmY5/4FtYiKz6tSyUv9LPEDei6A=
github.com/golang-sql/sqlexp v0.1.0/go.mod h1:J4ad9Vo8ZCWQ2GMrC4UCQy1JpCbwU9m3EOqtpKwwwHI=
github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b/go.mod h1:SBH7ygxi8pfUlaOkMMuAQtPIUF8ecWP5IEl/CR7VP2Q=
github.com/golang/mock v1.1.1/go.mod h1:oTYuIxOrZwtPieC+H1uAHpcLFnEyAGVDL/k47Jfbm0A=
github.com/golang/protobuf v1.2.0/go.mod h1:6lQm79b+lXiMfvg/cZm0SGofjICqVBUtrP5yJMmIC1U=
github.com/golang/protobuf v1.3.2/go.mod h1:6lQm79b+lXiMfvg/cZm0SGofjICqVBUtrP5yJMmIC1U=
github.com/golang/protobuf v1.3.3/go.mod h1:vzj43D7+SQXF/4pzW/hwtAqwc6iTitCiVSaWz5lYuqw=
github.com/golang/protobuf v1.4.0-rc.1/go.mod h1:ceaxUfeHdC40wWswd/P6IGgMaK3YpKi5j83Wpe3EHw8=
github.com/golang/protobuf v1.4.0-rc.1.0.20200221234624-67d41d38c208/go.mod h1:xKAWHe0F5eneWXFV3EuXVDTCmh+JuBKY0li0aMyXATA=
github.com/golang/protobuf v1.4.0-rc.2/go.mod h1:LlEzMj4AhA7rCAGe4KMBDvJI+AwstrUpVNzEA03Pprs=
github.com/golang/protobuf v1.4.0-rc.4.0.20200313231945-b860323f09d0/go.mod h1:WU3c8KckQ9AFe+yFwt9sWVRKCVIyN9cPHBJSNnbL67w=
github.com/golang/protobuf v1.4.0/go.mod h1:jodUvKwWbYaEsadDk5Fwe5c77LiNKVO9IDvqG2KuDX0=
github.com/golang/protobuf v1.4.1/go.mod h1:U8fpvMrcmy5pZrNK1lt4xCsGvpyWQ/VVv6QDs8UjoX8=
github.com/golang/protobuf v1.4.2/go.mod h1:oDoupMAO8OvCJWAcko0GGGIgR6R6ocIYbsSw735rRwI=
github.com/golang/protobuf v1.4.3/go.mod h1:oDoupMAO8OvCJWAcko0GGGIgR6R6ocIYbsSw735rRwI=
github.com/golang/protobuf v1.5.0/go.mod h1:FsONVRAS9T7sI+LIUmWTfcYkHO4aIWwzhcaSAoJOfIk=
github.com/golang/protobuf v1.5.2/go.mod h1:XVQd3VNwM+JqD3oG2Ue2ip4fOMUkwXdXDdiuN0vRsmY=
github.com/golang/protobuf v1.5.4 h1:i7eJL8qZTpSEXOPTxNKhASYpMn+8e5Q6AdndVa1dWek=
github.com/golang/protobuf v1.5.4/go.mod h1:lnTiLA8Wa4RWRcIUkrtSVa5nRhsEGBg48fD6rSs7xps=
github.com/google/go-cmp v0.2.0/go.mod h1:oXzfMopK8JAjlY9xF4vHSVASa0yLyX7SntLO5aqRK0M=
github.com/google/go-cmp v0.3.0/go.mod h1:8QqcDgzrUqlUb/G2PQTWiueGozuR1884gddMywk6iLU=
github.com/google/go-cmp v0.3.1/go.mod h1:8QqcDgzrUqlUb/G2PQTWiueGozuR1884gddMywk6iLU=
github.com/google/go-cmp v0.4.0/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
github.com/google/go-cmp v0.5.0/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
github.com/google/go-cmp v0.5.5/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
github.com/google/go-cmp v0.5.6/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
github.com/google/go-cmp v0.7.0 h1:wk8382ETsv4JYUZwIsn6YpYiWiBsYLSJiTsyBybVuN8=
github.com/google/go-cmp v0.7.0/go.mod h1:pXiqmnSA92OHEEa9HXL2W4E7lf9JzCmGVUdgjX3N/iU=
github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
github.com/google/pprof v0.0.0-20260802141513-ef3492d7dac3 h1:LMLX+LgTNWpfvCBdFebv6EsYotImrt/Ppc5cXIriCSo=
github.com/google/pprof v0.0.0-20260802141513-ef3492d7dac3/go.mod h1:jl5iWTm0/hd5PjEYEOuwAJ57L/CibdZfrqZ5XA5GrCk=
github.com/google/uuid v1.1.2/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/grpc-ecosystem/grpc-gateway v1.16.0/go.mod h1:BDjrQk3hbvj6Nolgz8mAMFbcEtjT1g+wF4CSlocrBnw=
github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 h1:/Tnpcb2E0Pz/tN9s3bfEY2Q8ePCEX9iuS+cneUwncnw=
github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0/go.mod h1:zOBXOsUaBSjKgmH4OGzV1esUpR3oUSCPYVd2cUBjKYY=
github.com/hashicorp/golang-lru/v2 v2.0.7 h1:a+bsQ5rvGLjzHuww6tVxozPZFVghXaHOwFs4luLUK2k=
github.com/hashicorp/golang-lru/v2 v2.0.7/go.mod h1:QeFd9opnmA6QUJc5vARoKUSoFhyfM2/ZepoAG6RGpeM=
github.com/hpcloud/tail v1.0.0/go.mod h1:ab1qPbhIpdTxEkNHXyeSf5vhxWSCs/tWer42PpOxQnU=
github.com/ianlancetaylor/demangle v0.0.0-20200824232613-28f6c0f3b639/go.mod h1:aSSvb/t6k1mPoxDqO4vJh6VOCGPwU4O0C2/Eqndh1Sc=
github.com/jackc/pgpassfile v1.0.0 h1:/6Hmqy13Ss2zCq62VdNG8tM1wchn8zjSGOBJ6icpsIM=
github.com/jackc/pgpassfile v1.0.0/go.mod h1:CEx0iS5ambNFdcRtxPj5JhEz+xB6uRky5eyVu/W2HEg=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 h1:iCEnooe7UlwOQYpKFhBabPMi4aNAfoODPEFNiAnClxo=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761/go.mod h1:5TJZWKEWniPve33vlWYSoGYefn3gLQRzjfDlhSJ9ZKM=
github.com/jackc/pgx/v5 v5.11.0 h1:IzBBtyK9AHqf98cctWFifYSci2hgQR/cd56wB4p+ogg=
github.com/jackc/pgx/v5 v5.11.0/go.mod h1:mal1tBGAFfLHvZzaYh77YS/eC6IX9OWbRV1QIIM0Jn4=
github.com/jackc/puddle/v2 v2.2.2 h1:PR8nw+E/1w0GLuRFSmiioY6UooMp6KJv0/61nB7icHo=
github.com/jackc/puddle/v2 v2.2.2/go.mod h1:vriiEXHvEE654aYKXXjOvZM39qJ0q+azkZFrfEOc3H4=
github.com/jessevdk/go-flags v1.4.0/go.mod h1:4FA24M0QyGHXBuZZK/XkWh8h0e1EYbRYJSGM75WSRxI=
github.com/joho/godotenv v1.5.1 h1:7eLL/+HRGLY0ldzfGMeQkb7vMd0as4CfYvUVzLqw0N0=
github.com/joho/godotenv v1.5.1/go.mod h1:f4LDr5Voq0i2e/R5DDNOoa2zzDfwtkZa6DnEwAbqwq4=
github.com/jonboulle/clockwork v0.5.0 h1:Hyh9A8u51kptdkR+cqRpT1EebBwTn1oK9YfGYbdFz6I=
github.com/jonboulle/clockwork v0.5.0/go.mod h1:3mZlmanh0g2NDKO5TWZVJAfofYk64M7XN3SzBPjZF60=
github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d/go.mod h1:2PavIy+JPciBPrBUjwbNvtwB6RQlve+hkpll6QSNmOE=
github.com/keybase/go-keychain v0.0.1 h1:way+bWYa6lDppZoZcgMbYsvC7GxljxrskdNInRtuthU=
github.com/keybase/go-keychain v0.0.1/go.mod h1:PdEILRW3i9D8JcdM+FmY6RwkHGnhHxXwkPPMeUgOK1k=
github.com/klauspost/compress v1.19.2 h1:hMRETovs/pu/dVWN7zIT1PGG8t509MwT6bO7XSi26R8=
github.com/klauspost/compress v1.19.2/go.mod h1:cwPg85FWrGar70rWktvGQj8/hthj3wpl0PGDogxkrSQ=
github.com/kr/pretty v0.1.0/go.mod h1:dAy3ld7l9f0ibDNOQOHHMYYIIbhfbHSm3C4ZsoJORNo=
github.com/kr/pretty v0.3.1 h1:flRD4NNwYAUpkphVc1HcthR4KEIFJ65n8Mw5qdRn3LE=
github.com/kr/pretty v0.3.1/go.mod h1:hoEshYVHaxMs3cyo3Yncou5ZscifuDolrwPKZanG3xk=
github.com/kr/pty v1.1.1/go.mod h1:pFQYn66WHrOpPYNljwOMqo10TkYh1fy3cYio2l3bCsQ=
github.com/kr/text v0.1.0/go.mod h1:4Jbv+DJW3UT/LiOwJeYQe1efqtUx/iVham/4vfdArNI=
github.com/kr/text v0.2.0 h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=
github.com/kr/text v0.2.0/go.mod h1:eLer722TekiGuMkidMxC/pM04lWEeraHUUmBw8l2grE=
github.com/kylelemons/godebug v1.1.0 h1:RPNrshWIDI6G2gRW9EHilWtl7Z6Sb1BR0xunSBf0SNc=
github.com/kylelemons/godebug v1.1.0/go.mod h1:9/0rRGxNHcop5bhtWyNeEfOS8JIWk580+fNqagV/RAw=
github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 h1:SOEGU9fKiNWd/HOJuq6+3iTQz8KNCLtVX6idSoTLdUw=
github.com/lann/builder v0.0.0-20180802200727-47ae307949d0/go.mod h1:dXGbAdH5GtBTC4WfIxhKZfyBF/HBFgRZSWwZ9g/He9o=
github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 h1:P6pPBnrTSX3DEVR4fDembhRWSsG5rVo6hYhAB/ADZrk=
github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0/go.mod h1:vmVJ0l/dxyfGW6FmdpVm2joNMFikkuWg0EoCKLGUMNw=
github.com/mattn/go-isatty v0.0.24 h1:tGZZoVgT/KiqK1c8ocVLeDS8BSWMRd47J3Lbz7vsReI=
github.com/mattn/go-isatty v0.0.24/go.mod h1:nMCL3Zebbrt45jsMDgnfIwz6ydEQApk5oEI3HqDio6A=
github.com/mfridman/interpolate v0.0.2 h1:pnuTK7MQIxxFz1Gr+rjSIx9u7qVjf5VOoM/u6BbAxPY=
github.com/mfridman/interpolate v0.0.2/go.mod h1:p+7uk6oE07mpE/Ik1b8EckO0O4ZXiGAfshKBWLUM9Xg=
github.com/mfridman/xflag v0.1.0 h1:TWZrZwG1QklFX5S4j1vxfF1sZbZeZSGofMwPMLAF29M=
github.com/mfridman/xflag v0.1.0/go.mod h1:/483ywM5ZO5SuMVjrIGquYNE5CzLrj5Ux/LxWWnjRaE=
github.com/microsoft/go-mssqldb v1.11.0 h1:YbDqolEjGH9hBfvKzONTf5/dbl9RKXmizMJE93lVxNs=
github.com/microsoft/go-mssqldb v1.11.0/go.mod h1:goQLDOPlMN/l1REhnNPElMoY/yX+fUWn1+7UoFJPH9Y=
github.com/ncruces/go-strftime v1.0.0 h1:HMFp8mLCTPp341M/ZnA4qaf7ZlsbTc+miZjCLOFAw7w=
github.com/ncruces/go-strftime v1.0.0/go.mod h1:Fwc5htZGVVkseilnfgOVb9mKy6w1naJmn9CehxcKcls=
github.com/nxadm/tail v1.4.4/go.mod h1:kenIhsEOeOJmVchQTgglprH7qJGnHDVpk1VPCcaMI8A=
github.com/nxadm/tail v1.4.8 h1:nPr65rt6Y5JFSKQO7qToXr7pePgD6Gwiw05lkbyAQTE=
github.com/nxadm/tail v1.4.8/go.mod h1:+ncqLTQzXmGhMZNUePPaPqPvBxHAIsmXswZKocGu+AU=
github.com/oapi-codegen/nullable v1.2.0 h1:VflFkDW980KhBPiFF7nWSyjg+r4Obqj8lXipV0UkP5w=
github.com/oapi-codegen/nullable v1.2.0/go.mod h1:KUZ3vUzkmEKY90ksAmit2+5juDIhIZhfDl+0PwOQlFY=
github.com/oapi-codegen/oapi-codegen/v2 v2.8.0 h1:s4hxMxuqtR8jPzXkBTtFwY/SBuj3gEAYikmbBSdtLMM=
github.com/oapi-codegen/oapi-codegen/v2 v2.8.0/go.mod h1:yae2TI9IYB5vxQ35gFrpXh9L5H1eJv4MAUK1jumGMTo=
github.com/oapi-codegen/runtime v1.7.0 h1:t7358VYPvNbWJ9gdAkIK/smVeHpBf6yp8VTsaZsb/7k=
github.com/oapi-codegen/runtime v1.7.0/go.mod h1:GwV7hC2hviaMzj+ITfHVRESK5J2W/GefVwIND/bMGvU=
github.com/oasdiff/yaml v0.1.1 h1:6nHx+pn9gBRM6YpBlFZFQGCCd1nuvqOBtTD3KKTgGxY=
github.com/oasdiff/yaml v0.1.1/go.mod h1:EYJNoyktvWMJ0Hmhx+6qTaqMOsalUaRGT8Sj1hNcegU=
github.com/oasdiff/yaml3 v0.0.14 h1:aLJee3hxBK2H5wdXd9iPcIXb93Nty1Ge0pT171eHtkw=
github.com/oasdiff/yaml3 v0.0.14/go.mod h1:csto2xfDjYccdUn/yw/bPjj/cYTdp6HtFA0J4TWG+gg=
github.com/onsi/ginkgo v1.6.0/go.mod h1:lLunBs/Ym6LB5Z9jYTR76FiuTmxDTDusOGeTQH+WWjE=
github.com/onsi/ginkgo v1.10.2/go.mod h1:lLunBs/Ym6LB5Z9jYTR76FiuTmxDTDusOGeTQH+WWjE=
github.com/onsi/ginkgo v1.12.1/go.mod h1:zj2OWP4+oCPe1qIXoGWkgMRwljMUYCdkwsT2108oapk=
github.com/onsi/ginkgo v1.16.4 h1:29JGrr5oVBm5ulCWet69zQkzWipVXIol6ygQUe/EzNc=
github.com/onsi/ginkgo v1.16.4/go.mod h1:dX+/inL/fNMqNlz0e9LfyB9TswhZpCVdJM/Z6Vvnwo0=
github.com/onsi/ginkgo/v2 v2.1.3/go.mod h1:vw5CSIxN1JObi/U8gcbwft7ZxR2dgaR70JSE3/PpL4c=
github.com/onsi/gomega v1.7.0/go.mod h1:ex+gbHU/CVuBBDIJjb2X0qEXbFg53c61hWP/1CpauHY=
github.com/onsi/gomega v1.7.1/go.mod h1:XdKZgCCFLUoM/7CFJVPcG8C1xQ1AJ0vpAezJrB7JYyY=
github.com/onsi/gomega v1.10.1/go.mod h1:iN09h71vgCQne3DLsj+A5owkum+a2tYe+TOCB1ybHNo=
github.com/onsi/gomega v1.17.0/go.mod h1:HnhC7FXeEQY45zxNK3PPoIUhzk/80Xly9PcubAlGdZY=
github.com/onsi/gomega v1.19.0 h1:4ieX6qQjPP/BfC3mpsAtIGGlxTWPeA3Inl/7DtXw1tw=
github.com/onsi/gomega v1.19.0/go.mod h1:LY+I3pBVzYsTBU1AnDwOSxaYi9WoWiqgwooUqq9yPro=
github.com/paulmach/orb v0.13.0 h1:r7n7mQGGF+cj/CbcivEj9J3HGK+XR+yXnvzRdq9saIw=
github.com/paulmach/orb v0.13.0/go.mod h1:6scRWINywA2Jf05dcjOfLfxrUIMECvTSG2MVbRLxu/k=
github.com/pierrec/lz4/v4 v4.1.29 h1:CDQY6qZOLI4DW0Nx6R1vRrifrCeQHnNXkMb0hZWXFjg=
github.com/pierrec/lz4/v4 v4.1.29/go.mod h1:EoQMVJgeeEOMsCqCzqFm2O0cJvljX2nGZjcRIPL34O4=
github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c h1:+mdjkGKdHQG3305AYmdv1U2eRNDiU2ErMBj1gwrq8eQ=
github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c/go.mod h1:7rwL4CYBLnjLxUqIJNnCWiEdr3bn6IUYi15bNlnbCCU=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/pressly/goose/v3 v3.28.0 h1:D2M+iL31GmpZxSHOhX8mqyqAT3CXnokUmm0eKoSP+Vc=
github.com/pressly/goose/v3 v3.28.0/go.mod h1:v26MOuB8bL3kzzrt3Vqhb3R0PRVsl8hFQKdrht/L6Rk=
github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4/go.mod h1:xMI15A0UPsDsEKsMN9yxemIoYk6Tm2C1GtYGdfGttqA=
github.com/prometheus/procfs v0.22.0 h1:6q9+/JL9IKAPbCmBrv9n5O5Ty3NKnciV5X7YGw0oics=
github.com/prometheus/procfs v0.22.0/go.mod h1:CvmFr/GVhIjIvWJZW3tgkODBQMRIf0EyWMQLHCHab58=
github.com/rekby/fixenv v0.6.1 h1:jUFiSPpajT4WY2cYuc++7Y1zWrnCxnovGCIX72PZniM=
github.com/rekby/fixenv v0.6.1/go.mod h1:/b5LRc06BYJtslRtHKxsPWFT/ySpHV+rWvzTg+XWk4c=
github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec h1:W09IVJc94icq4NjY3clb7Lk8O1qJ8BdBEF8z0ibU0rE=
github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec/go.mod h1:qqbHyh8v60DhA7CoWK5oRCqLrMHRGoxYCSS9EjAz6Eo=
github.com/rogpeppe/fastuuid v1.2.0/go.mod h1:jVj6XXZzXRy/MSR5jhDC/2q6DgLz+nrA6LYCDYWNEvQ=
github.com/rogpeppe/go-internal v1.14.1 h1:UQB4HGPB6osV0SQTLymcB4TgvyWu6ZyliaW0tI/otEQ=
github.com/rogpeppe/go-internal v1.14.1/go.mod h1:MaRKkUm5W0goXpeCfT7UZI6fk/L7L7so1lCWt35ZSgc=
github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 h1:KRzFb2m7YtdldCEkzs6KqmJw4nqEVZGK7IN2kJkjTuQ=
github.com/santhosh-tekuri/jsonschema/v6 v6.0.2/go.mod h1:JXeL+ps8p7/KNMjDQk3TCwPpBy0wYklyWTfbkIzdIFU=
github.com/segmentio/asm v1.2.1 h1:DTNbBqs57ioxAD4PrArqftgypG4/qNpXoJx8TVXxPR0=
github.com/segmentio/asm v1.2.1/go.mod h1:BqMnlJP91P8d+4ibuonYZw9mfnzI9HfxselHZr5aAcs=
github.com/sergi/go-diff v1.1.0 h1:we8PVUC3FE2uYfodKH/nBHMSetSfHDR6scGdBi+erh0=
github.com/sergi/go-diff v1.1.0/go.mod h1:STckp+ISIX8hZLjrqAeVduY0gWCT9IjLuqbuNXdaHfM=
github.com/sethvargo/go-retry v0.4.0 h1:9qy1OoIAxBL+gBYnkTnTnWle5wlfsXQlwRzIbbpdqPw=
github.com/sethvargo/go-retry v0.4.0/go.mod h1:tvsjdKG6xfiCx4LSiUZ06kcv38xvdVQwv8R6/VnnVWg=
github.com/shopspring/decimal v1.4.0 h1:bxl37RwXBklmTi0C79JfXCEBD1cqqHt0bbgBAGFp81k=
github.com/shopspring/decimal v1.4.0/go.mod h1:gawqmDU56v4yIKSwfBSFip1HdCCXN8/+DMd9qYNcwME=
github.com/speakeasy-api/jsonpath v0.6.3 h1:c+QPwzAOdrWvzycuc9HFsIZcxKIaWcNpC+xhOW9rJxU=
github.com/speakeasy-api/jsonpath v0.6.3/go.mod h1:2cXloNuQ+RSXi5HTRaeBh7JEmjRXTiaKpFTdZiL7URI=
github.com/speakeasy-api/openapi v1.24.0 h1:opoD27rupX7zBVPq1HkIGLeMOzNNA7JalhYP8q34i04=
github.com/speakeasy-api/openapi v1.24.0/go.mod h1:g3+dIMe0AYgbbGvnlQZqesmjAVWSm9BmsjLevnefQrg=
github.com/spkg/bom v0.0.0-20160624110644-59b7046e48ad/go.mod h1:qLr4V1qq6nMqFKkMo8ZTx3f+BZEkzsRUY10Xsm2mwU0=
github.com/stretchr/objx v0.1.0/go.mod h1:HFkY916IF+rwdDfMAkV7OtwuqBVzrE8GR6GFx+wExME=
github.com/stretchr/objx v0.5.2 h1:xuMeJ0Sdp5ZMRXx/aWO6RZxdr3beISkG5/G/aIRr3pY=
github.com/stretchr/objx v0.5.2/go.mod h1:FRsXN1f5AsAjCGJKqEizvkpNtU+EGNCLh3NxZ/8L+MA=
github.com/stretchr/testify v1.2.2/go.mod h1:a8OnRcib4nhh0OaRAV+Yts87kKdq0PP7pXfy6kDkUVs=
github.com/stretchr/testify v1.3.0/go.mod h1:M5WIy9Dh21IEIfnGCwXGc5bZfKNJtfHm1UVUgZn+9EI=
github.com/stretchr/testify v1.4.0/go.mod h1:j7eGeouHqKxXV5pUuKE4zz7dFj8WfuZ+81PSLYec5m4=
github.com/stretchr/testify v1.5.1/go.mod h1:5W2xD1RspED5o8YsWQXVCued0rvSQ+mT+I5cxcmMvtA=
github.com/stretchr/testify v1.7.0/go.mod h1:6Fq8oRcR53rry900zMqJjRRixrwX3KX962/h/Wwjteg=
github.com/stretchr/testify v1.12.1 h1:EuwCh5fleGS7H32xRwO3wRGT7DxrDhLAT6FF8MpWDWE=
github.com/stretchr/testify v1.12.1/go.mod h1:MDEgiDPPsNp5cuIrHPPCyornHKgEVbtFUmoNlxoYthg=
github.com/tursodatabase/libsql-client-go v0.0.0-20260528064733-9d5d30a29a60 h1:TfQEwhr0Q9t+Bgs0TNk2eHZ9EGD107Mimic0kcoGS1M=
github.com/tursodatabase/libsql-client-go v0.0.0-20260528064733-9d5d30a29a60/go.mod h1:08inkKyguB6CGGssc/JzhmQWwBgFQBgjlYFjxjRh7nU=
github.com/vertica/vertica-sql-go v1.3.8 h1:FomjkM3cam9yE6zSic31flNWPLdsZbYGK9ihlLtbF1Y=
github.com/vertica/vertica-sql-go v1.3.8/go.mod h1:c4OZ8lq1Ztc18w8a0nG+dzQh69BzJRcKN2LZOnYbERI=
github.com/vmware-labs/yaml-jsonpath v0.3.2 h1:/5QKeCBGdsInyDCyVNLbXyilb61MXGi9NP674f9Hobk=
github.com/vmware-labs/yaml-jsonpath v0.3.2/go.mod h1:U6whw1z03QyqgWdgXxvVnQ90zN1BWz5V+51Ewf8k+rQ=
github.com/xyproto/randomstring v1.0.5 h1:YtlWPoRdgMu3NZtP45drfy1GKoojuR7hmRcnhZqKjWU=
github.com/xyproto/randomstring v1.0.5/go.mod h1:rgmS5DeNXLivK7YprL0pY+lTuhNQW3iGxZ18UQApw/E=
github.com/ydb-platform/ydb-go-genproto v0.0.0-20260810123728-f0c151ab31b9 h1:WcjfLaNwBZzyl0a/2Ms+tucKRZvulR1HMXM9V7evjLc=
github.com/ydb-platform/ydb-go-genproto v0.0.0-20260810123728-f0c151ab31b9/go.mod h1:Er+FePu1dNUieD+XTMDduGpQuCPssK5Q4BjF+IIXJ3I=
github.com/ydb-platform/ydb-go-sdk/v3 v3.151.1 h1:T+fB2ZDHpYIGC7DWjK+rzZHLoexBF0zG/n3Q9DGNskY=
github.com/ydb-platform/ydb-go-sdk/v3 v3.151.1/go.mod h1:dJXJ1u00IqO8Vsph8fWmYj8b1E7jyphnvJPqA69emBY=
github.com/yuin/goldmark v1.2.1/go.mod h1:3hX8gzYuyVAZsxl0MRgGTJEmQBFcNTphYh9decYSb74=
github.com/ziutek/mymysql v1.5.4 h1:GB0qdRGsTwQSBVYuVShFBKaXSnSnYYC2d9knnE1LHFs=
github.com/ziutek/mymysql v1.5.4/go.mod h1:LMSpPZ6DbqWFxNCHW77HeMg9I646SAhApZ/wKdgO/C0=
go.opentelemetry.io/auto/sdk v1.2.1 h1:jXsnJ4Lmnqd11kwkBV2LgLoFMZKizbCi5fNZ/ipaZ64=
go.opentelemetry.io/auto/sdk v1.2.1/go.mod h1:KRTj+aOaElaLi+wW1kO/DZRXwkF4C5xPbEe3ZiIhN7Y=
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 h1:LxwW/9ctSCv+QkE/cLR7M91ZIkXNMqJtEMi1vCw9U8s=
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0/go.mod h1:tOsftB4SslBwwErVEPaenU2RpThXWPIU8DoJHEC4dyw=
go.opentelemetry.io/otel v1.47.0 h1:j7ALJ/zgkS7Z6aeJW09p8VC9804bC+PpeTfCD4XPnOM=
go.opentelemetry.io/otel v1.47.0/go.mod h1:8wS9O2qfXrYrzp6hIF/HOYJJf/wIhFPhR2xLuP+iXQU=
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.47.0 h1:LaLVGJtIzHcvHjVs2OftmKcc+M9mdOqS8vZ1fBD1Yes=
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.47.0/go.mod h1:vF6H5sNDuQBJWgSX00iZgbIqk2MZZoVMH0ar5t5bOpw=
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.47.0 h1:7kszRN4Ze5GeAkUnBjBSUIunVBMA+mZIG04auueqguM=
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.47.0/go.mod h1:CLzakYy8jOdkBfcveCWYkeKJxujcF4xPhh2Z7IfE98U=
go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 h1:julhjPeUH/q/7hinbSdDdqt5h7Zw9YWmRlWRhI0jd54=
go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0/go.mod h1:Ao2mz688LH/tFf0yMAenidq6k2YNSx6SIY2q6jDACck=
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0 h1:UFpxOpYPmMNUtOWhdb+nC1WELIgVf8z9higURrbzYU4=
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0/go.mod h1:YBGjxe3lt0jtXoEXxGrGhv/jM7oUBUDVK7zhAvNyEjc=
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.47.0 h1:aXZXsZ012wOgVkqaiQv+Z/d8V0xKjmg70F6m4CE94lc=
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.47.0/go.mod h1:ziu1gUIJhAaxM3EV1e3Ralk5AAyL1UOYz3lJJ4VSI38=
go.opentelemetry.io/otel/log v1.47.0 h1:cOTS1CcLbSQeZKanGJ+0JpF/+t4PELi3O3bbl2lqCcI=
go.opentelemetry.io/otel/log v1.47.0/go.mod h1:9byitSQ5pLC6PpqwGXjqdMKya6ZTswHRZh2vvXT33nw=
go.opentelemetry.io/otel/metric v1.47.0 h1:4PptaldXx3Eat1XjMZ68pPJEs5wrhlemctZE9a3UdWY=
go.opentelemetry.io/otel/metric v1.47.0/go.mod h1:ADGSXxRrXM6bjbvLo535EstVFlPpPYZm4LBKixjDHwU=
go.opentelemetry.io/otel/metric/x v0.69.0 h1:DjRLr15H83v+hCW7JA9NoJvOkYTtmq5YoDRbe9deYpM=
go.opentelemetry.io/otel/metric/x v0.69.0/go.mod h1:uVvsMPMFFyj/HUQfrUnH3JjnOQ1dwFDorgFLRBasM0k=
go.opentelemetry.io/otel/sdk v1.47.0 h1:zWXEr4j2lFefG87TU6Yg8a7ngfohIKFZHKp0Hf5hC6I=
go.opentelemetry.io/otel/sdk v1.47.0/go.mod h1:VUc24kiOeoGsxG8G9ULx3fWKvB7jMhnGE8Oi607lgR0=
go.opentelemetry.io/otel/sdk/metric v1.47.0 h1:lfISg2j93VT6yqdk9OfUaZmw/GfcZqCCV3jdXtsPnKw=
go.opentelemetry.io/otel/sdk/metric v1.47.0/go.mod h1:ypLp+mW1Nt2x+Szt3b5/i1syodyts49lMOwxpDI3VGw=
go.opentelemetry.io/otel/trace v1.47.0 h1:JOjX/Oci8K94QHddo+bbfya/Ai/nf6/dt9ZfrFNWSrM=
go.opentelemetry.io/otel/trace v1.47.0/go.mod h1:jNaSLa2PZEYFG6fRjJABAu+bw4FS08uDmPg28lTghu0=
go.opentelemetry.io/proto/otlp v0.7.0/go.mod h1:PqfVotwruBrMGOCsRd/89rSnXhoiJIqeYNgFYFoEGnI=
go.opentelemetry.io/proto/otlp v1.11.0 h1:5rrYs0Ykyj50sdU/JU0x8etU+LubXWb+gED6TbEdMIk=
go.opentelemetry.io/proto/otlp v1.11.0/go.mod h1:SmVizdCOAm3XBtG1g1NnOdhW6jtddT72hLMhv8VwA8E=
go.uber.org/goleak v1.3.0 h1:2K3zAYmnTNqV73imy9J1T3WC+gmCePx2hEGkimedGto=
go.uber.org/goleak v1.3.0/go.mod h1:CoHD4mav9JJNrW/WLlf7HGZPjdw8EucARQHekz1X6bE=
go.uber.org/mock v0.6.0 h1:hyF9dfmbgIX5EfOdasqLsWD6xqpNZlXblLB/Dbnwv3Y=
go.uber.org/mock v0.6.0/go.mod h1:KiVJ4BqZJaMj4svdfmHM0AUx4NJYO8ZNpPnZn1Z+BBU=
go.uber.org/multierr v1.11.0 h1:blXXJkSxSSfBVBlC76pxqeO+LN3aDfLQo+309xJstO0=
go.uber.org/multierr v1.11.0/go.mod h1:20+QtiLqy0Nd6FdQB9TLXag12DsQkrbs3htMFfDN80Y=
go.yaml.in/yaml/v3 v3.0.5 h1:N6y/pJk8buWs9NY5ERU2HSMfm+IuD/OtfdAnq6kESPw=
go.yaml.in/yaml/v3 v3.0.5/go.mod h1:HVTZu1O7/Vkt2N+BFy8Zza+lnLsABggaTM2ZpNIGuKg=
golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2/go.mod h1:djNgcEr1/C05ACkg1iLfiJU5Ep61QUkGW8qpdssI0+w=
golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550/go.mod h1:yigFU9vqHzYiE8UmvKecakEJjdnWj3jj499lnFckfCI=
golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9/go.mod h1:LzIPMQfyMNhhGPhUkYOs5KpL4U8rLKemX1yGLhDgUto=
golang.org/x/crypto v0.57.0 h1:3ZVCjf8Ggz7zneR/EHRVx68Ctf+2pmIMP2UFhh9cC6M=
golang.org/x/crypto v0.57.0/go.mod h1:Fdz0i5U6CoizGwLda9DttjSk6qlZo25zYNtR+ycvuZA=
golang.org/x/exp v0.0.0-20190121172915-509febef88a4/go.mod h1:CJ0aWSM057203Lf6IL+f9T1iT9GByDxfZKAQTCR3kQA=
golang.org/x/exp v0.0.0-20260824195058-e88cd73687aa h1:QSyA8ishJCyT21kER9KwNt0b7BM3iRK4x9QXhjN5Fdk=
golang.org/x/exp v0.0.0-20260824195058-e88cd73687aa/go.mod h1:zeBbvyFKDaLwa7CH/zI8KXt7gTl14SF7sO08Pl5jBCM=
golang.org/x/lint v0.0.0-20181026193005-c67002cb31c3/go.mod h1:UVdnD1Gm6xHRNCYTkRU2/jEulfH38KcIWyp/GAMgvoE=
golang.org/x/lint v0.0.0-20190227174305-5b3e6a55c961/go.mod h1:wehouNa3lNwaWXcvxsM5YxQ5yQlVC4a0KAMCusXpPoU=
golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3/go.mod h1:6SW0HCj/g11FgYtHlgUYUwCkIfeOF89ocIRzGO/8vkc=
golang.org/x/mod v0.3.0/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
golang.org/x/mod v0.41.0 h1:qJmnOUb4YB+FsEuM3HcWucdZASCPGhsX6uljO6pog0c=
golang.org/x/mod v0.41.0/go.mod h1:Ek9pY8RKWXwsWvd3rQiHYtMqkjSUV+s1Rj7j4H5Ur6o=
golang.org/x/net v0.0.0-20180724234803-3673e40ba225/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
golang.org/x/net v0.0.0-20180826012351-8a410e7b638d/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
golang.org/x/net v0.0.0-20180906233101-161cd47e91fd/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
golang.org/x/net v0.0.0-20190108225652-1e06a53dbb7e/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
golang.org/x/net v0.0.0-20190213061140-3a22650c66bd/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
golang.org/x/net v0.0.0-20190311183353-d8887717615a/go.mod h1:t9HGtf8HONx5eT2rtn7q6eTqICYqUVnKs3thJo3Qplg=
golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3/go.mod h1:t9HGtf8HONx5eT2rtn7q6eTqICYqUVnKs3thJo3Qplg=
golang.org/x/net v0.0.0-20190620200207-3b0461eec859/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7/go.mod h1:qpuaurCH72eLCgpAm/N6yyVIVM9cpaDIP3A8BGJEC5A=
golang.org/x/net v0.0.0-20200822124328-c89045814202/go.mod h1:/O7V0waA8r7cgGh81Ro3o1hOxt32SMVPicZroKQ2sZA=
golang.org/x/net v0.0.0-20201021035429-f5854403a974/go.mod h1:sp8m0HH+o8qH0wwXwYZr8TS3Oi6o0r6Gce1SSxlDquU=
golang.org/x/net v0.0.0-20210428140749-89ef3d95e781/go.mod h1:OJAsFXCWl8Ukc7SiCT/9KSuxbyM7479/AVlXFRxuMCk=
golang.org/x/net v0.0.0-20220225172249-27dd8689420f/go.mod h1:CfG3xpIq0wQ8r1q4Su4UZFWDARRcnwPjda9FqA0JpMk=
golang.org/x/net v0.59.0 h1:5zfYln+w5XCxwrnMMJPufRgNoXEaGxl0wo5GqPXyues=
golang.org/x/net v0.59.0/go.mod h1:2DA/G1UfVbCpQPeWTmMPGY7Cs2PkBkwu743bVX5PIVg=
golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be/go.mod h1:N/0e6XlmueqKjAGxoOufVs8QHGRruUQn6yWY3a++T0U=
golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d/go.mod h1:gOpvHmFTYa4IltrdGE7lF6nIHvwfUNPOp7c8zoXwtLw=
golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20181108010431-42b317875d0f/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20190423024810-112230192c58/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
golang.org/x/sync v0.23.0 h1:KameEIfc1IkluZyXWLn39Wd4tURc6GbCiISGiZm2bQk=
golang.org/x/sync v0.23.0/go.mod h1:sUUOizhqBxiL6pEWpqNLUiaJn1ShEbZ6BBqskPbjZm0=
golang.org/x/sys v0.0.0-20180830151530-49385e6e1522/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
golang.org/x/sys v0.0.0-20180909124046-d0be0721c37e/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
golang.org/x/sys v0.0.0-20190412213103-97732733099d/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20190904154756-749cb33beabd/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20191005200804-aed5e4c7ecf9/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20191120155948-bd437916bb0e/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20191204072324-ce4227a45e2e/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20201119102817-f84b799fce68/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20210112080510-489259a85091/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20210423082822-04245dca01da/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.1.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.48.0 h1:bbX/i/6MgT9BVLM9RT1thmxL04yeTAhbEz4SyadbXoo=
golang.org/x/sys v0.48.0/go.mod h1:hNLxWAXmnKAxqDtdwIYC4bM9oQPEecfsnNMuSxOs3og=
golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1/go.mod h1:bj7SfCRtBDWHUb9snDiAeCFNEtKQo2Wmx5Cou7ajbmo=
golang.org/x/term v0.0.0-20210927222741-03fcf44c2211/go.mod h1:jbD1KX2456YbFQfuXm/mYQcufACuNUgVhRMnK/tPxf8=
golang.org/x/text v0.3.0/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
golang.org/x/text v0.3.3/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
golang.org/x/text v0.3.6/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
golang.org/x/text v0.3.7/go.mod h1:u+2+/6zg+i71rQMx5EYifcz6MCKuco9NR6JIITiCfzQ=
golang.org/x/text v0.42.0 h1:JbOZXgfeCPU9gacVtYliJqOhD+zhrEqK4LfdpmlUZqI=
golang.org/x/text v0.42.0/go.mod h1:ojzP1Z+2QtioaF8DTtO8K5q7JWVVYwZKenzujK0Zd0E=
golang.org/x/tools v0.0.0-20180917221912-90fa682c2a6e/go.mod h1:n7NCudcB/nEzxVGmLbDWY5pfWTLqBcC2KZ6jyYvM4mQ=
golang.org/x/tools v0.0.0-20190114222345-bf090417da8b/go.mod h1:n7NCudcB/nEzxVGmLbDWY5pfWTLqBcC2KZ6jyYvM4mQ=
golang.org/x/tools v0.0.0-20190226205152-f727befe758c/go.mod h1:9Yl7xja0Znq3iFh3HoIrodX9oNMXvdceNzlUR8zjMvY=
golang.org/x/tools v0.0.0-20190311212946-11955173bddd/go.mod h1:LCzVGOaR6xXOjkQ3onu1FJEFr0SW1gC7cKk1uF8kGRs=
golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135/go.mod h1:RgjU9mgBXZiqYHBnxXauZ1Gv1EHHAz9KjViQ78xBX0Q=
golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
golang.org/x/tools v0.0.0-20201224043029-2b0845dc783e/go.mod h1:emZCQorbCU4vsT4fOWvOPXz4eW1wZW4PmDk9uLelYpA=
golang.org/x/tools v0.49.0 h1:3NI7VXzL9+1WZD52Dx2ttoPwD5DWrFGpl9mFZDlmisI=
golang.org/x/tools v0.49.0/go.mod h1:SJNXV9DBKT0UbdttsQjbfJlAE/q+y36++zo3uL3N0Oo=
golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
gonum.org/v1/gonum v0.17.0 h1:VbpOemQlsSMrYmn7T2OUvQ4dqxQXU+ouZFQsZOx50z4=
gonum.org/v1/gonum v0.17.0/go.mod h1:El3tOrEuMpv2UdMrbNlKEh9vd86bmQ6vqIcDwxEOc1E=
google.golang.org/appengine v1.1.0/go.mod h1:EbEs0AVv82hx2wNQdGPgUI5lhzA/G0D9YwlJXL52JkM=
google.golang.org/appengine v1.4.0/go.mod h1:xpcJRLb0r/rnEns0DIKYYv+WjYCduHsrkT7/EB5XEv4=
google.golang.org/genproto v0.0.0-20180817151627-c66870c02cf8/go.mod h1:JiN7NxoALGmiZfu7CAH4rXhgtRTLTxftemlI0sWmxmc=
google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55/go.mod h1:DMBHOl98Agz4BDEuKkezgsaosCRResVns1a3J2ZsMNc=
google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013/go.mod h1:NbSheEEYHJ7i3ixzK3sjbqSGDJWnxyFXZblF3eUsNvo=
google.golang.org/genproto/googleapis/api v0.0.0-20261104224025-d1e00e9f33ab h1:n6IDWZ1RzI/NberLoLLFNrjl1H45uhc4+WYsUf5lPPc=
google.golang.org/genproto/googleapis/api v0.0.0-20261104224025-d1e00e9f33ab/go.mod h1:vHV3wJM9WZyfV7GyB+OQGq5dWVrFwCDxapPK4WlRD88=
google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b h1:6WNWHZk91dp4gZ25642OALkuG1UVNZbwENJD19DnB8k=
google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b/go.mod h1:OaIUM3+LpYcK2GXM4FTmhWoIq371Owdr+Cc7/BsYHHc=
google.golang.org/grpc v1.19.0/go.mod h1:mqu4LbDTu4XGKhr4mRzUsmM4RtVoemTSY81AxZiDr8c=
google.golang.org/grpc v1.23.0/go.mod h1:Y5yQAOtifL1yxbo5wqy6BxZv8vAUGQwXBOALyacEbxg=
google.golang.org/grpc v1.25.1/go.mod h1:c3i+UQWmh7LiEpx4sFZnkU36qjEYZ0imhYfXVyQciAY=
google.golang.org/grpc v1.27.0/go.mod h1:qbnxyOmOxrQa7FizSgH+ReBfzJrCY1pSN7KXBS8abTk=
google.golang.org/grpc v1.33.1/go.mod h1:fr5YgcSWrqhRRxogOsw7RzIpsmvOZ6IcH4kBYTpR3n0=
google.golang.org/grpc v1.36.0/go.mod h1:qjiiYl8FncCW8feJPdyg3v6XW24KsRHe+dy9BAGRRjU=
google.golang.org/grpc v1.47.0/go.mod h1:vN9eftEi1UMyUsIF80+uQXhHjbXYbm0uXoFCACuMGWk=
google.golang.org/grpc v1.83.2 h1:EManeRomTObA0BU7I8vXgg/78uE5MJ9M8B39EX2WscU=
google.golang.org/grpc v1.83.2/go.mod h1:YPI1hK3kDked6iHvgX3tR0y+nX/qpMFKhPgFsokw1S8=
google.golang.org/protobuf v0.0.0-20200109180630-ec00e32a8dfd/go.mod h1:DFci5gLYBciE7Vtevhsrf46CRTquxDuWsQurQQe4oz8=
google.golang.org/protobuf v0.0.0-20200221191635-4d8936d0db64/go.mod h1:kwYJMbMJ01Woi6D6+Kah6886xMZcty6N08ah7+eCXa0=
google.golang.org/protobuf v0.0.0-20200228230310-ab0ca4ff8a60/go.mod h1:cfTl7dwQJ+fmap5saPgwCLgHXTUD7jkjRqWcaiX5VyM=
google.golang.org/protobuf v1.20.1-0.20200309200217-e05f789c0967/go.mod h1:A+miEFZTKqfCUM6K7xSMQL9OKL/b6hQv+e19PK+JZNE=
google.golang.org/protobuf v1.21.0/go.mod h1:47Nbq4nVaFHyn7ilMalzfO3qCViNmqZ2kzikPIcrTAo=
google.golang.org/protobuf v1.22.0/go.mod h1:EGpADcykh3NcUnDUJcl1+ZksZNG86OlYog2l/sGQquU=
google.golang.org/protobuf v1.23.0/go.mod h1:EGpADcykh3NcUnDUJcl1+ZksZNG86OlYog2l/sGQquU=
google.golang.org/protobuf v1.23.1-0.20200526195155-81db48ad09cc/go.mod h1:EGpADcykh3NcUnDUJcl1+ZksZNG86OlYog2l/sGQquU=
google.golang.org/protobuf v1.25.0/go.mod h1:9JNX74DMeImyA3h4bdi1ymwjUzf21/xIlbajtzgsN7c=
google.golang.org/protobuf v1.26.0-rc.1/go.mod h1:jlhhOSvTdKEhbULTjvd4ARK9grFBp09yW+WbY/TyQbw=
google.golang.org/protobuf v1.26.0/go.mod h1:9q0QmTI4eRPtz6boOQmLYwt+qCgq0jsYwAQnmE0givc=
google.golang.org/protobuf v1.27.1/go.mod h1:9q0QmTI4eRPtz6boOQmLYwt+qCgq0jsYwAQnmE0givc=
google.golang.org/protobuf v1.28.0/go.mod h1:HV8QOd/L58Z+nl8r43ehVNZIU/HEI6OcFqwMG9pJV4I=
google.golang.org/protobuf v1.36.12 h1:pJOKDDOyeXErUroCihFAd5LQuwXBSpVnKGrj5o/fwxc=
google.golang.org/protobuf v1.36.12/go.mod h1:HTf+CrKn2C3g5S8VImy6tdcUvCska2kB7j23XfzDpco=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=
gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c/go.mod h1:JHkPIbrfpd72SG/EVd6muEfDQjcINNoR0C8j2r3qZ4Q=
gopkg.in/fsnotify.v1 v1.4.7/go.mod h1:Tz8NjZHkW78fSQdbUxIjBTcgA1z1m8ZHf0WmKUhAMys=
gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 h1:uRGJdciOHaEIrze2W8Q3AKkepLTh2hOroT7a+7czfdQ=
gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7/go.mod h1:dt/ZhP58zS4L8KSrWDmTeBkI65Dw0HsyUHuEVlX15mw=
gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0/go.mod h1:WDnlLJ4WF5VGsH/HVa3CI79GS0ol3YnhVnKP89i0kNg=
gopkg.in/yaml.v2 v2.2.1/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=
gopkg.in/yaml.v2 v2.2.2/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=
gopkg.in/yaml.v2 v2.2.3/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=
gopkg.in/yaml.v2 v2.2.4/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=
gopkg.in/yaml.v2 v2.3.0/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=
gopkg.in/yaml.v2 v2.4.0 h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=
gopkg.in/yaml.v2 v2.4.0/go.mod h1:RDklbk79AGWmwhnvt/jBztapEOGDOx6ZbXqjP6csGnQ=
gopkg.in/yaml.v3 v3.0.0-20191026110619-0b21df46bc1d/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
honnef.co/go/tools v0.0.0-20190102054323-c2f93a96b099/go.mod h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4=
honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc/go.mod h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4=
howett.net/plist v1.0.1 h1:37GdZ8tP09Q35o9ych3ehygcsL+HqKSwzctveSlarvM=
howett.net/plist v1.0.1/go.mod h1:lqaXoTrLY4hg8tnEzNru53gicrbv7rrk+2xJA/7hw9g=
modernc.org/cc/v4 v4.29.2 h1:h6+9ciCnPKutf4I03CvheAvDLX7+IHlqR6Iy6J+cgd8=
modernc.org/cc/v4 v4.29.2/go.mod h1:OnovgIhbbMXMu1aISnJ0wvVD1KnW+cAUJkIrAWh+kVI=
modernc.org/ccgo/v4 v4.35.0 h1:F+TUsmw09QxLzmi3aeYYGxjAXarmZaKgj3mKQHNaA8w=
modernc.org/ccgo/v4 v4.35.0/go.mod h1:qrVGs9S3Sr2Ztcg9ve+kTAYMp5a3YvWjo+SoN06kJ5I=
modernc.org/fileutil v1.4.0 h1:j6ZzNTftVS054gi281TyLjHPp6CPHr2KCxEXjEbD6SM=
modernc.org/fileutil v1.4.0/go.mod h1:EqdKFDxiByqxLk8ozOxObDSfcVOv/54xDs/DUHdvCUU=
modernc.org/gc/v2 v2.6.5 h1:nyqdV8q46KvTpZlsw66kWqwXRHdjIlJOhG6kxiV/9xI=
modernc.org/gc/v2 v2.6.5/go.mod h1:YgIahr1ypgfe7chRuJi2gD7DBQiKSLMPgBQe9oIiito=
modernc.org/gc/v3 v3.1.5 h1:21ldfPfRYE31Tb7B3mwAK8gy1AxP4+dKjrOQPfqakoc=
modernc.org/gc/v3 v3.1.5/go.mod h1:HFK/6AGESC7Ex+EZJhJ2Gni6cTaYpSMmU/cT9RmlfYY=
modernc.org/goabi0 v0.2.0 h1:HvEowk7LxcPd0eq6mVOAEMai46V+i7Jrj13t4AzuNks=
modernc.org/goabi0 v0.2.0/go.mod h1:CEFRnnJhKvWT1c1JTI3Avm+tgOWbkOu5oPA8eH8LnMI=
modernc.org/libc v1.75.6 h1:yKk8qo+Di4gkmvRboK8ocCqH22FiUCR6jRy2OwtCRus=
modernc.org/libc v1.75.6/go.mod h1:bO5o2ztHxBb2rjz0PgdHN0sSMw57CgxGFLZ3Qd/QpVQ=
modernc.org/mathutil v1.7.1 h1:GCZVGXdaN8gTqB1Mf/usp1Y/hSqgI2vAGGP4jZMCxOU=
modernc.org/mathutil v1.7.1/go.mod h1:4p5IwJITfppl0G4sUEDtCr4DthTaT47/N3aT6MhfgJg=
modernc.org/memory v1.12.1 h1:nFMiWrpStgZczNl6XI9GnIk/rWhYIyHGUaR04pGbp9g=
modernc.org/memory v1.12.1/go.mod h1:/JP4VbVC+K5sU2wZi9bHoq2MAkCnrt2r98UGeSK7Mjw=
modernc.org/opt v0.2.0 h1:tGyef5ApycA7FSEOMraay9SaTk5zmbx7Tu+cJs4QKZg=
modernc.org/opt v0.2.0/go.mod h1:03fq9lsNfvkYSfxrfUhZCWPk1lm4cq4N+Bh//bEtgns=
modernc.org/sortutil v1.2.1 h1:+xyoGf15mM3NMlPDnFqrteY07klSFxLElE2PVuWIJ7w=
modernc.org/sortutil v1.2.1/go.mod h1:7ZI3a3REbai7gzCLcotuw9AC4VZVpYMjDzETGsSMqJE=
modernc.org/sqlite v1.57.0 h1:qNQP6xnx5M0ISNtlnxoOX0+cD5bJ0/gr9aMmndFczzg=
modernc.org/sqlite v1.57.0/go.mod h1:yCJ2cmAaIkHQ25oXWrF8H4O1lIfPYPR26yCEDj2P3pQ=
modernc.org/strutil v1.2.1 h1:UneZBkQA+DX2Rp35KcM69cSsNES9ly8mQWD71HKlOA0=
modernc.org/strutil v1.2.1/go.mod h1:EHkiggD70koQxjVdSBM3JKM7k6L0FbGE5eymy9i3B9A=
modernc.org/token v1.1.0 h1:Xl7Ap9dKaEs5kLoOQeQmPWevfnk/DM5qcLcYlA8ys6Y=
modernc.org/token v1.1.0/go.mod h1:UGzOrNV1mAFSEB63lOFHIpNRUVMvYTc6yu1SMY/XTDM=

```

`internal/config/config.go`:

```go
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP            HTTPConfig
	Database        DatabaseConfig
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
}

const (
	defaultHTTPReadTimeout       = 10 * time.Second
	defaultHTTPReadHeaderTimeout = 5 * time.Second
	defaultHTTPWriteTimeout      = 15 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second
)

func Load() (Config, error) {
	r := &envReader{}

	cfg := Config{
		HTTP: HTTPConfig{
			Addr:              r.String("HTTP_ADDR"),
			ReadTimeout:       r.DurationOr("HTTP_READ_TIMEOUT", defaultHTTPReadTimeout),
			ReadHeaderTimeout: r.DurationOr("HTTP_READ_HEADER_TIMEOUT", defaultHTTPReadHeaderTimeout),
			WriteTimeout:      r.DurationOr("HTTP_WRITE_TIMEOUT", defaultHTTPWriteTimeout),
			IdleTimeout:       r.DurationOr("HTTP_IDLE_TIMEOUT", defaultHTTPIdleTimeout),
		},
		Database: DatabaseConfig{
			URL:             r.String("DATABASE_URL"),
			MaxConns:        r.Int32("DATABASE_MAX_CONNS"),
			MinConns:        r.Int32("DATABASE_MIN_CONNS"),
			MaxConnLifetime: r.Duration("DATABASE_MAX_CONN_LIFETIME"),
			ConnectTimeout:  r.Duration("DATABASE_CONNECT_TIMEOUT"),
			QueryTimeout:    r.Duration("DATABASE_QUERY_TIMEOUT"),
		},
		LogLevel:        r.LogLevel("LOG_LEVEL"),
		ShutdownTimeout: r.Duration("SHUTDOWN_TIMEOUT"),
	}

	if cfg.Database.MinConns > cfg.Database.MaxConns {
		r.addError("DATABASE_MIN_CONNS", errors.New("must not be greater than DATABASE_MAX_CONNS"))
	}

	if len(r.errs) > 0 {
		return Config{}, fmt.Errorf("invalid configuration: %w", errors.Join(r.errs...))
	}

	return cfg, nil
}

type envReader struct {
	errs []error
}

func (r *envReader) addError(name string, err error) {
	r.errs = append(r.errs, fmt.Errorf("%s: %w", name, err))
}

func (r *envReader) required(name string) (string, bool) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		r.addError(name, errors.New("is required but not set"))
		return "", false
	}

	return value, true
}

func (r *envReader) String(name string) string {
	value, _ := r.required(name)
	return value
}

func (r *envReader) Int32(name string) int32 {
	value, ok := r.required(name)
	if !ok {
		return 0
	}

	n, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		r.addError(name, fmt.Errorf("must be an integer, got %q", value))
		return 0
	}
	if n <= 0 {
		r.addError(name, fmt.Errorf("must be positive, got %d", n))
		return 0
	}

	return int32(n)
}

func (r *envReader) Duration(name string) time.Duration {
	value, ok := r.required(name)
	if !ok {
		return 0
	}

	return r.parseDuration(name, value)
}

func (r *envReader) DurationOr(name string, def time.Duration) time.Duration {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return def
	}

	return r.parseDuration(name, value)
}

func (r *envReader) parseDuration(name, value string) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		r.addError(name, fmt.Errorf("must be a duration like 5s or 30m, got %q", value))
		return 0
	}
	if d <= 0 {
		r.addError(name, fmt.Errorf("must be positive, got %s", d))
		return 0
	}

	return d
}

func (r *envReader) LogLevel(name string) slog.Level {
	value, ok := r.required(name)
	if !ok {
		return slog.LevelInfo
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		r.addError(name, fmt.Errorf("must be one of debug, info, warn, error, got %q", value))
		return slog.LevelInfo
	}

	return level
}

```

`internal/container/container.go`:

```go
package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Chillobass/fork_template/internal/config"
	transport "github.com/Chillobass/fork_template/internal/http"
	"github.com/Chillobass/fork_template/internal/http/health"
	"github.com/Chillobass/fork_template/internal/http/middleware"
	httpcreate "github.com/Chillobass/fork_template/internal/http/trip/create"
	httpfinish "github.com/Chillobass/fork_template/internal/http/trip/finish"
	httpget "github.com/Chillobass/fork_template/internal/http/trip/get"
	"github.com/Chillobass/fork_template/internal/infra/postgres"
	"github.com/Chillobass/fork_template/internal/infra/postgres/transaction"
	"github.com/Chillobass/fork_template/internal/infra/telemetry"
	repository "github.com/Chillobass/fork_template/internal/repository/trip"
	"github.com/Chillobass/fork_template/internal/usecase/trip/create"
	"github.com/Chillobass/fork_template/internal/usecase/trip/finish"
	"github.com/Chillobass/fork_template/internal/usecase/trip/get"
)

type App struct {
	Server *http.Server
	pool   *pgxpool.Pool
	otel   telemetry.Providers
}

func New(ctx context.Context, cfg config.Config, service string, logger *slog.Logger) (_ *App, err error) {
	providers := telemetry.New(ctx, service, logger)
	defer func() {
		if err != nil {
			cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
			defer cancel()
			if shutdownErr := providers.Shutdown(cleanup); shutdownErr != nil {
				logger.WarnContext(cleanup, "telemetry shutdown failed", "error", shutdownErr)
			}
		}
	}()

	meter := providers.Metrics.Meter(service)
	tracer, err := postgres.NewTracer(providers.Traces, meter)
	if err != nil {
		return nil, err
	}
	httpMetrics, err := middleware.New(logger, providers.Traces, meter)
	if err != nil {
		return nil, err
	}
	pool, err := postgres.NewPool(ctx, cfg.Database, tracer)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	tx := transaction.NewTxManager(pool, cfg.Database.QueryTimeout)
	trips := repository.New(tx, cfg.Database.QueryTimeout)
	events, err := telemetry.NewTrips(meter, trips, logger)
	if err != nil {
		pool.Close()
		return nil, err
	}
	handlers := &transport.Server{
		CreateHandler: httpcreate.New(create.New(tx, trips, trips, time.Now), logger, events.Created),
		GetHandler:    httpget.New(get.New(trips)),
		FinishHandler: httpfinish.New(finish.New(tx, trips, trips, time.Now), logger, events.Completed),
		HealthHandler: health.New(pool, cfg.Database.QueryTimeout, logger),
	}
	return &App{
		Server: &http.Server{
			Addr:              cfg.HTTP.Addr,
			Handler:           transport.NewRouter(handlers, httpMetrics),
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
			IdleTimeout:       cfg.HTTP.IdleTimeout,
			ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		},
		pool: pool,
		otel: providers,
	}, nil
}

func (a *App) Shutdown(ctx context.Context, logger *slog.Logger) error {
	err := a.Server.Shutdown(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "graceful HTTP shutdown failed, closing connections", "error", err)
		err = errors.Join(err, a.Server.Close())
	}
	if telemetryErr := a.otel.Shutdown(ctx); telemetryErr != nil {
		logger.WarnContext(ctx, "telemetry shutdown failed", "error", telemetryErr)
	}
	closed := make(chan struct{})
	go func() {
		a.pool.Close()
		close(closed)
	}()
	select {
	case <-closed:
		return err
	case <-ctx.Done():
		return errors.Join(err, fmt.Errorf("close database pool: %w", ctx.Err()))
	}
}

```

`internal/domain/trip/errors.go`:

```go
package trip

import "errors"

var (
	ErrTripNotFound  = errors.New("trip not found")
	ErrTripCompleted = errors.New("trip is already completed")
	ErrDriverBusy    = errors.New("driver already has an active trip")
)

type ValidationError struct {
	Field  string
	Detail string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Detail
}

```

`internal/domain/trip/trip.go`:

```go
package trip

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

type Data struct {
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartPoint Point
	EndPoint   Point
	Price      int64
}

type Trip struct {
	Data
	ID         uuid.UUID
	Status     Status
	StartedAt  time.Time
	FinishedAt *time.Time
}

func New(id uuid.UUID, data Data, at time.Time) (Trip, error) {
	if err := data.Validate(); err != nil {
		return Trip{}, err
	}
	return Trip{Data: data, ID: id, Status: StatusActive, StartedAt: at}, nil
}

func (t *Trip) Finish(at time.Time) error {
	if t.Status != StatusActive {
		return ErrTripCompleted
	}
	t.Status = StatusCompleted
	t.FinishedAt = &at
	return nil
}

func (d Data) Validate() error {
	switch {
	case d.UserID == uuid.Nil:
		return &ValidationError{Field: "user_id", Detail: "must be a non-empty UUID"}
	case d.DriverID == uuid.Nil:
		return &ValidationError{Field: "driver_id", Detail: "must be a non-empty UUID"}
	}
	if err := d.StartPoint.validate("start_point"); err != nil {
		return err
	}
	if err := d.EndPoint.validate("end_point"); err != nil {
		return err
	}
	if d.Price < 0 {
		return &ValidationError{Field: "price", Detail: "must be greater than or equal to 0"}
	}
	return nil
}

func (p Point) validate(field string) error {
	switch {
	case math.IsNaN(p.Latitude) || math.Abs(p.Latitude) > 90:
		return &ValidationError{Field: field + ".latitude", Detail: "must be between -90 and 90"}
	case math.IsNaN(p.Longitude) || math.Abs(p.Longitude) > 180:
		return &ValidationError{Field: field + ".longitude", Detail: "must be between -180 and 180"}
	default:
		return nil
	}
}

```

`internal/domain/trip/trip_test.go`:

```go
package trip_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

func TestValidate(t *testing.T) {
	valid := trip.Data{
		UserID: uuid.New(), DriverID: uuid.New(),
		StartPoint: trip.Point{Latitude: -90, Longitude: 180},
		EndPoint:   trip.Point{Latitude: 90, Longitude: -180},
	}
	cases := []struct {
		name   string
		change func(*trip.Data)
		field  string
	}{
		{name: "boundary values and zero price"},
		{name: "missing user", change: func(d *trip.Data) { d.UserID = uuid.Nil }, field: "user_id"},
		{name: "missing driver", change: func(d *trip.Data) { d.DriverID = uuid.Nil }, field: "driver_id"},
		{name: "start latitude", change: func(d *trip.Data) { d.StartPoint.Latitude = -91 }, field: "start_point.latitude"},
		{name: "end longitude", change: func(d *trip.Data) { d.EndPoint.Longitude = 181 }, field: "end_point.longitude"},
		{name: "NaN longitude", change: func(d *trip.Data) { d.StartPoint.Longitude = math.NaN() }, field: "start_point.longitude"},
		{name: "negative price", change: func(d *trip.Data) { d.Price = -1 }, field: "price"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := valid
			if tc.change != nil {
				tc.change(&data)
			}
			err := data.Validate()
			if tc.field == "" {
				if err != nil {
					t.Fatalf("validate: %v", err)
				}
				return
			}
			var validation *trip.ValidationError
			if !errors.As(err, &validation) || validation.Field != tc.field || validation.Detail == "" {
				t.Fatalf("expected validation error on %s, got %v", tc.field, err)
			}
		})
	}
}

func TestFinish(t *testing.T) {
	at := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		status     trip.Status
		finishedAt *time.Time
		wantErr    error
	}{
		{name: "active", status: trip.StatusActive},
		{name: "completed", status: trip.StatusCompleted, finishedAt: new(at.Add(-time.Hour)), wantErr: trip.ErrTripCompleted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := trip.Trip{Status: tc.status, FinishedAt: tc.finishedAt}
			if err := value.Finish(at); !errors.Is(err, tc.wantErr) {
				t.Fatalf("finish error = %v, want %v", err, tc.wantErr)
			}
			want := at
			if tc.finishedAt != nil {
				want = *tc.finishedAt
			}
			if value.Status != trip.StatusCompleted || value.FinishedAt == nil || !value.FinishedAt.Equal(want) {
				t.Fatalf("unexpected trip after finish: %+v", value)
			}
		})
	}
}

```

`internal/generated/api.gen.go`:

```go
// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.8.0 DO NOT EDIT.
package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/nullable"
	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Defines values for HealthResponseStatus.
const (
	Ok          HealthResponseStatus = "ok"
	Unavailable HealthResponseStatus = "unavailable"
)

// Valid indicates whether the value is a known member of the HealthResponseStatus enum.
func (e HealthResponseStatus) Valid() bool {
	switch e {
	case Ok:
		return true
	case Unavailable:
		return true
	default:
		return false
	}
}

// Defines values for TripStatus.
const (
	Active    TripStatus = "active"
	Completed TripStatus = "completed"
)

// Valid indicates whether the value is a known member of the TripStatus enum.
func (e TripStatus) Valid() bool {
	switch e {
	case Active:
		return true
	case Completed:
		return true
	default:
		return false
	}
}

// Coordinates Географические координаты WGS 84.
type Coordinates struct {
	// Latitude Example: 59.9398
	Latitude float64 `json:"latitude"`

	// Longitude Example: 30.3146
	Longitude float64 `json:"longitude"`
}

// HealthResponse defines model for HealthResponse.
type HealthResponse struct {
	Status HealthResponseStatus `json:"status"`
}

// HealthResponseStatus defines model for HealthResponse.Status.
type HealthResponseStatus string

// Problem Problem Details for HTTP APIs (RFC 9457) с дополнительным полем code.
type Problem struct {
	Code     string  `json:"code"`
	Detail   *string `json:"detail,omitempty"`
	Instance *string `json:"instance,omitempty"`
	Status   int32   `json:"status"`
	Title    string  `json:"title"`
	Type     string  `json:"type"`
}

// Trip defines model for Trip.
type Trip struct {
	DriverId openapi_types.UUID `json:"driver_id"`

	// EndPoint Географические координаты WGS 84.
	EndPoint       Coordinates                  `json:"end_point"`
	FinishedAt     nullable.Nullable[time.Time] `json:"finished_at,omitempty"`
	Id             openapi_types.UUID           `json:"id"`
	LastPositionAt nullable.Nullable[time.Time] `json:"last_position_at,omitempty"`

	// Price Стоимость поездки в целых рублях. Это учебное упрощение: копейки не используются.
	//
	// Example: 1450
	Price int64 `json:"price"`

	// StartPoint Географические координаты WGS 84.
	StartPoint Coordinates        `json:"start_point"`
	StartedAt  time.Time          `json:"started_at"`
	Status     TripStatus         `json:"status"`
	UserId     openapi_types.UUID `json:"user_id"`
}

// TripData defines model for TripData.
type TripData struct {
	DriverId openapi_types.UUID `json:"driver_id"`

	// EndPoint Географические координаты WGS 84.
	EndPoint Coordinates `json:"end_point"`

	// Price Стоимость поездки в целых рублях. Это учебное упрощение: копейки не используются.
	//
	// Example: 1450
	Price int64 `json:"price"`

	// StartPoint Географические координаты WGS 84.
	StartPoint Coordinates        `json:"start_point"`
	UserId     openapi_types.UUID `json:"user_id"`
}

// TripStatus defines model for TripStatus.
type TripStatus string

// IdempotencyKey defines model for IdempotencyKey.
type IdempotencyKey = openapi_types.UUID

// TripId defines model for TripId.
type TripId = openapi_types.UUID

// BadRequest Problem Details for HTTP APIs (RFC 9457) с дополнительным полем code.
type BadRequest = Problem

// CreateTripConflict Problem Details for HTTP APIs (RFC 9457) с дополнительным полем code.
type CreateTripConflict = Problem

// InternalError Problem Details for HTTP APIs (RFC 9457) с дополнительным полем code.
type InternalError = Problem

// TripCompleted Problem Details for HTTP APIs (RFC 9457) с дополнительным полем code.
type TripCompleted = Problem

// TripNotFound Problem Details for HTTP APIs (RFC 9457) с дополнительным полем code.
type TripNotFound = Problem

// CreateTripParams defines parameters for CreateTrip.
type CreateTripParams struct {
	// IdempotencyKey Ключ идемпотентности, генерирует клиент. Повторный запрос с тем же ключом и тем же телом возвращает уже созданную поездку и 200, с тем же ключом и другим телом — 409 idempotency_conflict. Заголовок относится к заданию со звёздочкой лабораторной работы 1, в основной части можно игнорировать.
	IdempotencyKey *IdempotencyKey `json:"Idempotency-Key,omitempty"`
}

// CreateTripJSONRequestBody defines body for CreateTrip for application/json ContentType.
type CreateTripJSONRequestBody = TripData

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// CreateTrip Создать поездку
	// (POST /api/v1/trips)
	CreateTrip(w http.ResponseWriter, r *http.Request, params CreateTripParams)
	// GetTrip Получить поездку
	// (GET /api/v1/trips/{tripId})
	GetTrip(w http.ResponseWriter, r *http.Request, tripId TripId)
	// FinishTrip Завершить поездку
	// (POST /api/v1/trips/{tripId}/finish)
	FinishTrip(w http.ResponseWriter, r *http.Request, tripId TripId)
	// Health Liveness
	// (GET /health)
	Health(w http.ResponseWriter, r *http.Request)
	// Ready Readiness
	// (GET /ready)
	Ready(w http.ResponseWriter, r *http.Request)
}

// Unimplemented server implementation that returns http.StatusNotImplemented for each endpoint.

type Unimplemented struct{}

// CreateTrip Создать поездку
// (POST /api/v1/trips)
func (_ Unimplemented) CreateTrip(w http.ResponseWriter, r *http.Request, params CreateTripParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GetTrip Получить поездку
// (GET /api/v1/trips/{tripId})
func (_ Unimplemented) GetTrip(w http.ResponseWriter, r *http.Request, tripId TripId) {
	w.WriteHeader(http.StatusNotImplemented)
}

// FinishTrip Завершить поездку
// (POST /api/v1/trips/{tripId}/finish)
func (_ Unimplemented) FinishTrip(w http.ResponseWriter, r *http.Request, tripId TripId) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Health Liveness
// (GET /health)
func (_ Unimplemented) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Ready Readiness
// (GET /ready)
func (_ Unimplemented) Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// CreateTrip operation middleware
func (siw *ServerInterfaceWrapper) CreateTrip(w http.ResponseWriter, r *http.Request) {

	var err error
	_ = err

	// Parameter object where we will unmarshal all parameters from the context
	var params CreateTripParams

	headers := r.Header

	// ------------- Optional header parameter "Idempotency-Key" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("Idempotency-Key")]; found {
		var IdempotencyKey IdempotencyKey
		n := len(valueList)
		if n != 1 {
			siw.ErrorHandlerFunc(w, r, &TooManyValuesForParamError{ParamName: "Idempotency-Key", Count: n})
			return
		}

		err = runtime.BindStyledParameterWithOptions("simple", "Idempotency-Key", valueList[0], &IdempotencyKey, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: false, Type: "string", Format: "uuid"})
		if err != nil {
			siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "Idempotency-Key", Err: err})
			return
		}

		params.IdempotencyKey = &IdempotencyKey

	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateTrip(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetTrip operation middleware
func (siw *ServerInterfaceWrapper) GetTrip(w http.ResponseWriter, r *http.Request) {

	var err error
	_ = err

	// ------------- Path parameter "tripId" -------------
	var tripId TripId

	err = runtime.BindStyledParameterWithOptions("simple", "tripId", chi.URLParam(r, "tripId"), &tripId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true, Type: "string", Format: "uuid", ValueIsUnescaped: r.URL.RawPath == ""})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "tripId", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetTrip(w, r, tripId)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// FinishTrip operation middleware
func (siw *ServerInterfaceWrapper) FinishTrip(w http.ResponseWriter, r *http.Request) {

	var err error
	_ = err

	// ------------- Path parameter "tripId" -------------
	var tripId TripId

	err = runtime.BindStyledParameterWithOptions("simple", "tripId", chi.URLParam(r, "tripId"), &tripId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true, Type: "string", Format: "uuid", ValueIsUnescaped: r.URL.RawPath == ""})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "tripId", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.FinishTrip(w, r, tripId)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// Health operation middleware
func (siw *ServerInterfaceWrapper) Health(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Health(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// Ready operation middleware
func (siw *ServerInterfaceWrapper) Ready(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Ready(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/api/v1/trips", wrapper.CreateTrip)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/api/v1/trips/{tripId}", wrapper.GetTrip)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/api/v1/trips/{tripId}/finish", wrapper.FinishTrip)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/health", wrapper.Health)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/ready", wrapper.Ready)
	})

	return r
}

```

`internal/generated/oapi-codegen.yaml`:

```yaml
package: api
output: internal/generated/api.gen.go
generate:
  models: true
  chi-server: true
output-options:
  include-operation-ids:
    - createTrip
    - getTrip
    - finishTrip
    - health
    - ready
  nullable-type: true

```

`internal/http/domainerror/domainerror.go`:

```go
package domainerror

import (
	"context"
	"errors"
	"net/http"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/respond"
)

const (
	contentTypeProblem = "application/problem+json"

	problemTypeBaseURI = "https://tripgo.example/problems/"
)

type problemKind struct {
	status        int32
	code          string
	slug          string
	title         string
	defaultDetail string
}

var (
	problemInvalidRequest = problemKind{
		status:        http.StatusBadRequest,
		code:          "invalid_request",
		slug:          "invalid-request",
		title:         "Invalid request",
		defaultDetail: "Request validation failed",
	}
	problemTripNotFound = problemKind{
		status:        http.StatusNotFound,
		code:          "trip_not_found",
		slug:          "trip-not-found",
		title:         "Trip not found",
		defaultDetail: "Trip was not found",
	}
	problemTripCompleted = problemKind{
		status:        http.StatusConflict,
		code:          "trip_completed",
		slug:          "trip-completed",
		title:         "Trip completed",
		defaultDetail: "Operation is not allowed for a completed trip",
	}
	problemDriverBusy = problemKind{
		status:        http.StatusConflict,
		code:          "driver_busy",
		slug:          "driver-busy",
		title:         "Driver busy",
		defaultDetail: "Driver already has an active trip",
	}
	problemInternal = problemKind{
		status:        http.StatusInternalServerError,
		code:          "internal_error",
		slug:          "internal-error",
		title:         "Internal Server Error",
		defaultDetail: "Internal server error",
	}
)

func Problem(err error) (int, api.Problem) {
	var validation *trip.ValidationError
	switch {
	case errors.As(err, &validation):
		return problemInvalidRequest.problem(validation.Error())
	case errors.Is(err, trip.ErrTripNotFound):
		return problemTripNotFound.problem("")
	case errors.Is(err, trip.ErrTripCompleted):
		return problemTripCompleted.problem("")
	case errors.Is(err, trip.ErrDriverBusy):
		return problemDriverBusy.problem("")
	default:
		return problemInternal.problem("")
	}
}

type reportKey struct{}

// WithReport gives request logging access to the original internal error.
func WithReport(ctx context.Context, report func(error)) context.Context {
	return context.WithValue(ctx, reportKey{}, report)
}

func Write(w http.ResponseWriter, r *http.Request, err error) {
	status, body := Problem(err)
	if status == http.StatusInternalServerError {
		if report, ok := r.Context().Value(reportKey{}).(func(error)); ok {
			report(err)
		}
	}
	write(w, r, status, body)
}

func InvalidRequest(w http.ResponseWriter, r *http.Request, detail string) {
	status, body := problemInvalidRequest.problem(detail)
	write(w, r, status, body)
}

func (k problemKind) problem(detail string) (int, api.Problem) {
	if detail == "" {
		detail = k.defaultDetail
	}
	return int(k.status), api.Problem{
		Type:   problemTypeBaseURI + k.slug,
		Title:  k.title,
		Status: k.status,
		Detail: &detail,
		Code:   k.code,
	}
}

func write(w http.ResponseWriter, r *http.Request, status int, body api.Problem) {
	instance := r.URL.Path
	body.Instance = &instance
	respond.Write(w, status, contentTypeProblem, body)
}

```

`internal/http/domainerror/domainerror_test.go`:

```go
package domainerror_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
)

func TestProblem(t *testing.T) {
	const private = "failed SQL query with confidential data"
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{trip.ErrTripNotFound, http.StatusNotFound, "trip_not_found"},
		{trip.ErrTripCompleted, http.StatusConflict, "trip_completed"},
		{trip.ErrDriverBusy, http.StatusConflict, "driver_busy"},
		{&trip.ValidationError{Field: "price", Detail: "must be greater than or equal to 0"}, http.StatusBadRequest, "invalid_request"},
		{errors.New(private), http.StatusInternalServerError, "internal_error"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			status, body := domainerror.Problem(fmt.Errorf("usecase: %w", tc.err))
			if status != tc.status || int(body.Status) != status || body.Code != tc.code || body.Detail == nil ||
				body.Title == "" || body.Type == "" || strings.Contains(*body.Detail, private) {
				t.Fatalf("problem = (%d, %+v)", status, body)
			}
			if status == http.StatusBadRequest && !strings.Contains(*body.Detail, "price") {
				t.Fatalf("validation detail = %q", *body.Detail)
			}
		})
	}
}

```

`internal/http/health/contract.go`:

```go
package health

import "context"

type pinger interface {
	Ping(context.Context) error
}

```

`internal/http/health/handler.go`:

```go
package health

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/respond"
)

type Handler struct {
	db      pinger
	timeout time.Duration
	logger  *slog.Logger
}

func New(db pinger, timeout time.Duration, logger *slog.Logger) *Handler {
	return &Handler{db: db, timeout: timeout, logger: logger}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	respond.JSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		h.logger.WarnContext(ctx, "readiness check failed", "error", err)
		respond.JSON(w, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}
	respond.JSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

```

`internal/http/middleware/http.go`:

```go
package middleware

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/Chillobass/fork_template/internal/http/domainerror"
)

type HTTP struct {
	logger   *slog.Logger
	tracer   trace.TracerProvider
	duration metric.Float64Histogram
	active   metric.Int64UpDownCounter
}

func New(logger *slog.Logger, tracer trace.TracerProvider, meter metric.Meter) (*HTTP, error) {
	duration, err := meter.Float64Histogram("http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 1, 2.5, 5, 10))
	if err != nil {
		return nil, fmt.Errorf("register HTTP duration: %w", err)
	}
	active, err := meter.Int64UpDownCounter("http.server.active_requests", metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("register active HTTP requests: %w", err)
	}
	return &HTTP{logger: logger, tracer: tracer, duration: duration, active: active}, nil
}

type requestKey struct{}

type observation struct {
	route   string
	started time.Time
	status  int
	ctx     context.Context
	failure error
}

func (o *observation) statusCode() int {
	if o.status != 0 {
		return o.status
	}
	return http.StatusInternalServerError
}

func (m *HTTP) Wrap(router *chi.Mux) http.Handler {
	return m.recover(router, m.tracing(m.logging(m.metrics(router))))
}

func (m *HTTP) recover(router chi.Routes, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := &observation{route: routePattern(router, r), started: time.Now(), ctx: r.Context()}
		ctx := context.WithValue(r.Context(), requestKey{}, o)
		ctx = domainerror.WithReport(ctx, func(err error) { o.failure = err })

		wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					if o.status == 0 && (code >= 200 || code == http.StatusSwitchingProtocols) {
						o.status = code
					}
					next(code)
				}
			},
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(body []byte) (int, error) {
					o.setOK()
					return next(body)
				}
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(reader io.Reader) (int64, error) {
					o.setOK()
					return next(reader)
				}
			},
			Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc {
				return func() {
					o.setOK()
					next()
				}
			},
		})
		defer func() {
			p := recover()
			if p == nil {
				return
			}
			if p == http.ErrAbortHandler {
				panic(p)
			}
			m.logRequest(r, o, slog.LevelError, "panic_type", fmt.Sprintf("%T", p), "stack", string(debug.Stack()))
			if o.status != 0 {
				panic(http.ErrAbortHandler)
			}
			domainerror.Write(w, r, fmt.Errorf("panic in HTTP handler"))
		}()
		next.ServeHTTP(wrapped, r.WithContext(ctx))
	})
}

func (m *HTTP) tracing(next http.Handler) http.Handler {
	traced := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := observed(r)
		o.ctx = r.Context()
		span := trace.SpanFromContext(o.ctx)
		span.SetAttributes(attribute.String("http.route", o.route))
		defer func() {
			status := o.statusCode()
			span.SetAttributes(attribute.Int("http.response.status_code", status))
			if status >= http.StatusBadRequest {
				span.SetStatus(codes.Error, "")
				span.SetAttributes(attribute.String("error.type", strconv.Itoa(status)))
			}
		}()
		next.ServeHTTP(w, r)
	})
	return otelhttp.NewHandler(traced, "http.request",
		otelhttp.WithTracerProvider(m.tracer),
		// The instruments below use the same route labels on success and panic.
		otelhttp.WithMeterProvider(noop.NewMeterProvider()),
		otelhttp.WithPropagators(propagation.TraceContext{}),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + observed(r).route
		}),
	)
}

func (m *HTTP) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		o := observed(r)
		level := slog.LevelInfo
		if o.statusCode() >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		if o.failure != nil {
			m.logRequest(r, o, level, "error", o.failure)
			return
		}
		m.logRequest(r, o, level)
	})
}

func (m *HTTP) logRequest(r *http.Request, o *observation, level slog.Level, details ...any) {
	fields := make([]any, 0, 8+len(details))
	fields = append(fields, "method", r.Method, "route", o.route, "status", o.statusCode(),
		"duration_ms", time.Since(o.started).Milliseconds())
	m.logger.Log(o.ctx, level, "request handled", append(fields, details...)...)
}

func (m *HTTP) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := observed(r)
		ctx := r.Context()
		method := attribute.String("http.request.method", r.Method)
		route := attribute.String("http.route", o.route)
		activeLabels := metric.WithAttributeSet(attribute.NewSet(method, route))
		m.active.Add(ctx, 1, activeLabels)
		defer func() {
			m.active.Add(ctx, -1, activeLabels)
			m.duration.Record(ctx, time.Since(o.started).Seconds(),
				metric.WithAttributeSet(attribute.NewSet(method, route, attribute.Int("http.response.status_code", o.statusCode()))))
		}()
		next.ServeHTTP(w, r)
		o.setOK()
	})
}

func (o *observation) setOK() {
	if o.status == 0 {
		o.status = http.StatusOK
	}
}

func observed(r *http.Request) *observation {
	return r.Context().Value(requestKey{}).(*observation)
}

func routePattern(router chi.Routes, r *http.Request) string {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	if pattern := router.Find(chi.NewRouteContext(), r.Method, path); pattern != "" {
		return pattern
	}
	return "<unmatched>"
}

```

`internal/http/middleware/http_test.go`:

```go
package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/middleware"
)

func TestRequestTelemetry(t *testing.T) {
	const pattern = "/api/v1/trips/{tripId}"
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		status  int
		level   string
	}{
		{name: "implicit success", handler: func(http.ResponseWriter, *http.Request) {}, status: http.StatusOK, level: "INFO"},
		{name: "client error", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}, status: http.StatusNotFound, level: "INFO"},
		{name: "internal error", handler: func(w http.ResponseWriter, r *http.Request) {
			domainerror.Write(w, r, errors.New("connection refused"))
		}, status: http.StatusInternalServerError, level: "ERROR"},
		{name: "panic", handler: func(http.ResponseWriter, *http.Request) {
			panic("private panic value")
		}, status: http.StatusInternalServerError, level: "ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			spans := tracetest.NewSpanRecorder()
			reader := sdkmetric.NewManualReader()
			tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
			meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			m, err := middleware.New(slog.New(slog.NewJSONHandler(&logs, nil)), tracer, meter.Meter("test"))
			if err != nil {
				t.Fatal(err)
			}
			router := chi.NewRouter()
			router.Get(pattern, tc.handler)
			rec := httptest.NewRecorder()
			m.Wrap(router).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/trips/example-id", nil))

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			var logEntry struct {
				Route  string
				Status int
				Level  string
				Error  string
				Stack  string
			}
			if err := json.Unmarshal(logs.Bytes(), &logEntry); err != nil {
				t.Fatalf("expected one request log, got %s: %v", logs.Bytes(), err)
			}
			if logEntry.Route != pattern || logEntry.Status != tc.status || logEntry.Level != tc.level {
				t.Fatalf("unexpected request log: %s", logs.Bytes())
			}
			if tc.name == "internal error" && logEntry.Error != "connection refused" {
				t.Fatalf("original error missing from log: %s", logs.Bytes())
			}
			if tc.name == "panic" && (logEntry.Stack == "" || !strings.Contains(rec.Body.String(), "internal_error")) {
				t.Fatalf("panic stack or problem missing: %s, %s", logs.Bytes(), rec.Body)
			}

			ended := spans.Ended()
			if len(ended) != 1 || ended[0].Name() != "GET "+pattern {
				t.Fatalf("unexpected spans: %v", ended)
			}
			var data metricdata.ResourceMetrics
			if err := reader.Collect(t.Context(), &data); err != nil {
				t.Fatal(err)
			}
			checkMetrics(t, data, pattern, tc.status)
		})
	}
}

func checkMetrics(t *testing.T, data metricdata.ResourceMetrics, pattern string, status int) {
	t.Helper()
	if len(data.ScopeMetrics) != 1 || len(data.ScopeMetrics[0].Metrics) != 2 {
		t.Fatalf("expected two instruments, got %+v", data.ScopeMetrics)
	}
	for _, metric := range data.ScopeMetrics[0].Metrics {
		switch points := metric.Data.(type) {
		case metricdata.Histogram[float64]:
			if len(points.DataPoints) != 1 {
				t.Fatalf("unexpected duration points: %+v", points)
			}
			point := points.DataPoints[0]
			route, _ := point.Attributes.Value("http.route")
			code, _ := point.Attributes.Value("http.response.status_code")
			if route.AsString() != pattern || code.AsInt64() != int64(status) || point.Count != 1 {
				t.Fatalf("unexpected duration point: %+v", point)
			}
		case metricdata.Sum[int64]:
			if len(points.DataPoints) != 1 || points.DataPoints[0].Value != 0 {
				t.Fatalf("active requests did not return to zero: %+v", points)
			}
		default:
			t.Fatalf("unexpected metric: %+v", metric)
		}
	}
}

```

`internal/http/respond/json.go`:

```go
package respond

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, body any) {
	Write(w, status, "application/json", body)
}

func Write(w http.ResponseWriter, status int, contentType string, body any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	// Headers are already committed. An interrupted write cannot be retried.
	_ = json.NewEncoder(w).Encode(body)
}

```

`internal/http/router.go`:

```go
package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/health"
	"github.com/Chillobass/fork_template/internal/http/middleware"
	"github.com/Chillobass/fork_template/internal/http/trip/create"
	"github.com/Chillobass/fork_template/internal/http/trip/finish"
	"github.com/Chillobass/fork_template/internal/http/trip/get"
)

type (
	CreateHandler = create.Handler
	GetHandler    = get.Handler
	FinishHandler = finish.Handler
	HealthHandler = health.Handler
)

type Server struct {
	*CreateHandler
	*GetHandler
	*FinishHandler
	*HealthHandler
}

func NewRouter(server api.ServerInterface, metrics *middleware.HTTP) http.Handler {
	router := chi.NewRouter()
	api.HandlerWithOptions(server, api.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			detail := "invalid request parameters"
			var formatErr *api.InvalidParamFormatError
			if errors.As(err, &formatErr) {
				detail = fmt.Sprintf("parameter %s has invalid format", formatErr.ParamName)
			}
			domainerror.InvalidRequest(w, r, detail)
		},
	})
	return metrics.Wrap(router)
}

```

`internal/http/trip/convert.go`:

```go
package trip

import (
	"time"

	"github.com/oapi-codegen/nullable"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
)

func Response(trip domain.Trip) api.Trip {
	resp := api.Trip{
		Id:       trip.ID,
		UserId:   trip.UserID,
		DriverId: trip.DriverID,
		StartPoint: api.Coordinates{
			Latitude:  trip.StartPoint.Latitude,
			Longitude: trip.StartPoint.Longitude,
		},
		EndPoint: api.Coordinates{
			Latitude:  trip.EndPoint.Latitude,
			Longitude: trip.EndPoint.Longitude,
		},
		Price:          trip.Price,
		Status:         api.TripStatus(trip.Status),
		StartedAt:      trip.StartedAt.UTC(),
		FinishedAt:     nullable.NewNullNullable[time.Time](),
		LastPositionAt: nullable.NewNullNullable[time.Time](),
	}

	if trip.FinishedAt != nil {
		resp.FinishedAt = nullable.NewNullableWithValue(trip.FinishedAt.UTC())
	}

	return resp
}

```

`internal/http/trip/create/contract.go`:

```go
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package create

import (
	"context"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type usecase interface {
	Create(ctx context.Context, input trip.Data) (trip.Trip, error)
}

```

`internal/http/trip/create/convert.go`:

```go
package create

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
)

const maxRequestBodyBytes = 1 << 20

var (
	tripDataRequiredFields    = []string{"user_id", "driver_id", "start_point", "end_point", "price"}
	coordinatesRequiredFields = []string{"latitude", "longitude"}
)

func decodeCreateTripRequest(w http.ResponseWriter, r *http.Request) (trip.Data, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	if err != nil {
		return trip.Data{}, errors.New("request body is too large or could not be read")
	}

	var req api.TripData
	if err := decodeStrict(body, &req); err != nil {
		return trip.Data{}, err
	}

	// после Unmarshal не отличить "price": 0 от отсутствующего price, поэтому наличие полей проверяем отдельно
	if err := checkRequiredFields(body); err != nil {
		return trip.Data{}, err
	}

	return trip.Data{
		UserID:     req.UserId,
		DriverID:   req.DriverId,
		StartPoint: trip.Point{Latitude: req.StartPoint.Latitude, Longitude: req.StartPoint.Longitude},
		EndPoint:   trip.Point{Latitude: req.EndPoint.Latitude, Longitude: req.EndPoint.Longitude},
		Price:      req.Price,
	}, nil
}

func decodeStrict(body []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return describeJSONError(err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func describeJSONError(err error) error {
	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
	)

	switch {
	case errors.As(err, &typeErr) && typeErr.Field == "":
		return errors.New("request body must be a JSON object")
	case errors.As(err, &typeErr):
		return fmt.Errorf("field %s has invalid type", typeErr.Field)
	case errors.As(err, &syntaxErr), errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("request body must be a valid JSON object")
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return errors.New(strings.TrimPrefix(err.Error(), "json: "))
	default:
		return errors.New("invalid request body")
	}
}

func checkRequiredFields(body []byte) error {
	var trip map[string]json.RawMessage
	if err := json.Unmarshal(body, &trip); err != nil || trip == nil {
		return errors.New("request body must be a JSON object")
	}
	if err := requireFields(trip, "", tripDataRequiredFields); err != nil {
		return err
	}

	for _, pointField := range []string{"start_point", "end_point"} {
		var point map[string]json.RawMessage
		if err := json.Unmarshal(trip[pointField], &point); err != nil {
			return fmt.Errorf("field %s must be an object", pointField)
		}
		if err := requireFields(point, pointField+".", coordinatesRequiredFields); err != nil {
			return err
		}
	}

	return nil
}

func requireFields(object map[string]json.RawMessage, prefix string, names []string) error {
	for _, name := range names {
		value, ok := object[name]
		if !ok || string(value) == "null" {
			return fmt.Errorf("field %s%s is required", prefix, name)
		}
	}

	return nil
}

```

`internal/http/trip/create/handler.go`:

```go
package create

import (
	"context"
	"log/slog"
	"net/http"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/respond"
	"github.com/Chillobass/fork_template/internal/http/trip"
)

type Handler struct {
	usecase usecase
	logger  *slog.Logger
	count   func(context.Context)
}

func New(uc usecase, logger *slog.Logger, count func(context.Context)) *Handler {
	return &Handler{usecase: uc, logger: logger, count: count}
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, _ api.CreateTripParams) {
	input, err := decodeCreateTripRequest(w, r)
	if err != nil {
		domainerror.InvalidRequest(w, r, err.Error())
		return
	}

	ctx := r.Context()
	created, err := h.usecase.Create(ctx, input)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	h.count(ctx)
	h.logger.InfoContext(ctx, "trip created", "trip_id", created.ID)
	w.Header().Set("Location", "/api/v1/trips/"+created.ID.String())
	respond.JSON(w, http.StatusCreated, trip.Response(created))
}

```

`internal/http/trip/create/handler_test.go`:

```go
package create_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/trip/create"
)

const validRequest = `{
	"user_id": "589b9e2a-52f2-45d0-a04c-f2823c48b9a7",
	"driver_id": "c0d1fc78-11b7-4cf8-a43e-f079eec2c5a5",
	"start_point": {"latitude": 0, "longitude": 3},
	"end_point": {"latitude": 4, "longitude": 5},
	"price": 0
}`

func TestCreateTrip(t *testing.T) {
	const private = "SQL: could not connect"
	const invalidRequest = "invalid_request"
	for _, tc := range []struct {
		name       string
		body       string
		expectCall bool
		err        error
		status     int
		code       string
	}{
		{name: "created", body: validRequest, expectCall: true, status: http.StatusCreated},
		{name: "busy driver", body: validRequest, expectCall: true, err: trip.ErrDriverBusy, status: http.StatusConflict, code: "driver_busy"},
		{name: "internal error", body: validRequest, expectCall: true, err: errors.New(private), status: http.StatusInternalServerError, code: "internal_error"},
		{name: "invalid JSON", body: "{", status: http.StatusBadRequest, code: invalidRequest},
		{name: "missing price", body: strings.Replace(validRequest, ",\n\t\"price\": 0", "", 1), status: http.StatusBadRequest, code: invalidRequest},
		{name: "null latitude", body: strings.Replace(validRequest, "\"latitude\": 0", "\"latitude\": null", 1), status: http.StatusBadRequest, code: invalidRequest},
		{name: "unknown field", body: strings.Replace(validRequest, "\"price\": 0", "\"price\": 0, \"other\": 1", 1), status: http.StatusBadRequest, code: invalidRequest},
		{name: "multiple objects", body: validRequest + "{}", status: http.StatusBadRequest, code: invalidRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewMockusecase(gomock.NewController(t))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(tc.body))
			data := trip.Data{
				UserID:     uuid.MustParse("589b9e2a-52f2-45d0-a04c-f2823c48b9a7"),
				DriverID:   uuid.MustParse("c0d1fc78-11b7-4cf8-a43e-f079eec2c5a5"),
				StartPoint: trip.Point{Longitude: 3},
				EndPoint:   trip.Point{Latitude: 4, Longitude: 5},
			}
			created := trip.Trip{ID: uuid.New(), Data: data, Status: trip.StatusActive, StartedAt: time.Now().UTC()}
			if tc.expectCall {
				uc.EXPECT().Create(req.Context(), data).Return(created, tc.err)
			}
			rec := httptest.NewRecorder()
			count := 0
			create.New(uc, slog.New(slog.DiscardHandler), func(_ context.Context) { count++ }).CreateTrip(rec, req, api.CreateTripParams{})

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.status == http.StatusCreated {
				var got api.Trip
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Id != created.ID || got.Status != api.Active || !got.FinishedAt.IsNull() ||
					count != 1 || rec.Header().Get("Location") != "/api/v1/trips/"+created.ID.String() {
					t.Fatalf("unexpected create response: %+v, count %d, headers %v", got, count, rec.Header())
				}
				return
			}
			var got api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Code != tc.code || strings.Contains(rec.Body.String(), private) || count != 0 ||
				rec.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("unexpected error response: %s, count %d, headers %v", rec.Body, count, rec.Header())
			}
		})
	}
}

```

`internal/http/trip/create/mocks_test.go`:

```go
// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go
//
// Generated by this command:
//
//	mockgen -source contract.go -destination mocks_test.go -package create_test
//

// Package create_test is a generated GoMock package.
package create_test

import (
	context "context"
	reflect "reflect"

	trip "github.com/Chillobass/fork_template/internal/domain/trip"
	gomock "go.uber.org/mock/gomock"
)

// Mockusecase is a mock of usecase interface.
type Mockusecase struct {
	ctrl     *gomock.Controller
	recorder *MockusecaseMockRecorder
	isgomock struct{}
}

// MockusecaseMockRecorder is the mock recorder for Mockusecase.
type MockusecaseMockRecorder struct {
	mock *Mockusecase
}

// NewMockusecase creates a new mock instance.
func NewMockusecase(ctrl *gomock.Controller) *Mockusecase {
	mock := &Mockusecase{ctrl: ctrl}
	mock.recorder = &MockusecaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockusecase) EXPECT() *MockusecaseMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *Mockusecase) Create(ctx context.Context, input trip.Data) (trip.Trip, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, input)
	ret0, _ := ret[0].(trip.Trip)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockusecaseMockRecorder) Create(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*Mockusecase)(nil).Create), ctx, input)
}

```

`internal/http/trip/finish/contract.go`:

```go
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type usecase interface {
	Finish(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}

```

`internal/http/trip/finish/handler.go`:

```go
package finish

import (
	"context"
	"log/slog"
	"net/http"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/respond"
	"github.com/Chillobass/fork_template/internal/http/trip"
)

type Handler struct {
	usecase usecase
	logger  *slog.Logger
	count   func(context.Context)
}

func New(uc usecase, logger *slog.Logger, count func(context.Context)) *Handler {
	return &Handler{usecase: uc, logger: logger, count: count}
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	ctx := r.Context()
	finished, err := h.usecase.Finish(ctx, id)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	h.count(ctx)
	h.logger.InfoContext(ctx, "trip completed", "trip_id", id)
	respond.JSON(w, http.StatusOK, trip.Response(finished))
}

```

`internal/http/trip/finish/handler_test.go`:

```go
package finish_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/trip/finish"
)

func TestFinishTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "completed", status: http.StatusOK},
		{name: "already completed", err: trip.ErrTripCompleted, status: http.StatusConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewMockusecase(gomock.NewController(t))
			id := uuid.New()
			at := time.Now().UTC()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
			uc.EXPECT().Finish(req.Context(), id).Return(trip.Trip{ID: id, Status: trip.StatusCompleted, FinishedAt: &at}, tc.err)
			rec := httptest.NewRecorder()
			count := 0
			finish.New(uc, slog.New(slog.DiscardHandler), func(_ context.Context) { count++ }).FinishTrip(rec, req, id)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.err == nil {
				var got api.Trip
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				finished, err := got.FinishedAt.Get()
				if err != nil || got.Id != id || got.Status != api.Completed || !finished.Equal(at) || count != 1 {
					t.Fatalf("unexpected trip: %+v, count %d", got, count)
				}
				return
			}
			var problem api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
				t.Fatal(err)
			}
			if problem.Code != "trip_completed" || count != 0 {
				t.Fatalf("unexpected problem: %s, count %d", rec.Body, count)
			}
		})
	}
}

```

`internal/http/trip/finish/mocks_test.go`:

```go
// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go
//
// Generated by this command:
//
//	mockgen -source contract.go -destination mocks_test.go -package finish_test
//

// Package finish_test is a generated GoMock package.
package finish_test

import (
	context "context"
	reflect "reflect"

	trip "github.com/Chillobass/fork_template/internal/domain/trip"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// Mockusecase is a mock of usecase interface.
type Mockusecase struct {
	ctrl     *gomock.Controller
	recorder *MockusecaseMockRecorder
	isgomock struct{}
}

// MockusecaseMockRecorder is the mock recorder for Mockusecase.
type MockusecaseMockRecorder struct {
	mock *Mockusecase
}

// NewMockusecase creates a new mock instance.
func NewMockusecase(ctrl *gomock.Controller) *Mockusecase {
	mock := &Mockusecase{ctrl: ctrl}
	mock.recorder = &MockusecaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockusecase) EXPECT() *MockusecaseMockRecorder {
	return m.recorder
}

// Finish mocks base method.
func (m *Mockusecase) Finish(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Finish", ctx, id)
	ret0, _ := ret[0].(trip.Trip)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Finish indicates an expected call of Finish.
func (mr *MockusecaseMockRecorder) Finish(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Finish", reflect.TypeOf((*Mockusecase)(nil).Finish), ctx, id)
}

```

`internal/http/trip/get/contract.go`:

```go
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package get

import (
	"context"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type usecase interface {
	Get(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}

```

`internal/http/trip/get/handler.go`:

```go
package get

import (
	"net/http"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/respond"
	"github.com/Chillobass/fork_template/internal/http/trip"
)

type Handler struct {
	usecase usecase
}

func New(uc usecase) *Handler {
	return &Handler{usecase: uc}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	found, err := h.usecase.Get(r.Context(), id)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, trip.Response(found))
}

```

`internal/http/trip/get/handler_test.go`:

```go
package get_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/trip/get"
)

func TestGetTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "found", status: http.StatusOK},
		{name: "missing", err: trip.ErrTripNotFound, status: http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewMockusecase(gomock.NewController(t))
			id := uuid.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
			uc.EXPECT().Get(req.Context(), id).Return(trip.Trip{ID: id, Status: trip.StatusActive}, tc.err)
			rec := httptest.NewRecorder()
			get.New(uc).GetTrip(rec, req, id)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.status, rec.Body)
			}
			if tc.err == nil {
				var got api.Trip
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Id != id || got.Status != api.Active || !got.LastPositionAt.IsNull() {
					t.Fatalf("unexpected trip: %+v", got)
				}
				return
			}
			var problem api.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
				t.Fatal(err)
			}
			if problem.Code != "trip_not_found" {
				t.Fatalf("unexpected problem: %s", rec.Body)
			}
		})
	}
}

```

`internal/http/trip/get/mocks_test.go`:

```go
// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go
//
// Generated by this command:
//
//	mockgen -source contract.go -destination mocks_test.go -package get_test
//

// Package get_test is a generated GoMock package.
package get_test

import (
	context "context"
	reflect "reflect"

	trip "github.com/Chillobass/fork_template/internal/domain/trip"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// Mockusecase is a mock of usecase interface.
type Mockusecase struct {
	ctrl     *gomock.Controller
	recorder *MockusecaseMockRecorder
	isgomock struct{}
}

// MockusecaseMockRecorder is the mock recorder for Mockusecase.
type MockusecaseMockRecorder struct {
	mock *Mockusecase
}

// NewMockusecase creates a new mock instance.
func NewMockusecase(ctrl *gomock.Controller) *Mockusecase {
	mock := &Mockusecase{ctrl: ctrl}
	mock.recorder = &MockusecaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockusecase) EXPECT() *MockusecaseMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *Mockusecase) Get(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(trip.Trip)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockusecaseMockRecorder) Get(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*Mockusecase)(nil).Get), ctx, id)
}

```

`internal/infra/postgres/pool.go`:

```go
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Chillobass/fork_template/internal/config"
)

func NewPool(ctx context.Context, cfg config.DatabaseConfig, tracer pgx.QueryTracer) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	poolCfg.ConnConfig.Tracer = tracer

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

```

`internal/infra/postgres/tracer.go`:

```go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

type queryTracer struct {
	tracer   *otelpgx.Tracer
	duration metric.Float64Histogram
}

func NewTracer(provider trace.TracerProvider, meter metric.Meter) (pgx.QueryTracer, error) {
	duration, err := meter.Float64Histogram("db.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5))
	if err != nil {
		return nil, fmt.Errorf("register database duration: %w", err)
	}
	return &queryTracer{
		tracer: otelpgx.NewTracer(
			otelpgx.WithTracerProvider(provider),
			// otelpgx lacks per-query collection labels on its histogram.
			otelpgx.WithMeterProvider(noop.NewMeterProvider()),
			otelpgx.WithDisableConnectionDetailsInAttributes(),
		),
		duration: duration,
	}, nil
}

type queryKey struct{}

type queryInfo struct {
	started    time.Time
	operation  attribute.KeyValue
	collection attribute.KeyValue
}

func (t *queryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	op := otelpgx.SQLOperationName(ctx, data.SQL)
	info := queryInfo{
		started:    time.Now(),
		operation:  attribute.String("db.operation.name", op),
		collection: attribute.String("db.collection.name", collection(data.SQL)),
	}
	ctx = t.tracer.TraceQueryStart(ctx, conn, data)
	trace.SpanFromContext(ctx).SetAttributes(info.operation, info.collection)
	return context.WithValue(ctx, queryKey{}, info)
}

func (t *queryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	info := ctx.Value(queryKey{}).(queryInfo)
	attrs := make([]attribute.KeyValue, 0, 3)
	attrs = append(attrs, info.operation, info.collection)
	if data.Err != nil && !errors.Is(data.Err, pgx.ErrNoRows) {
		errorType := attribute.String("error.type", dbErrorType(data.Err))
		attrs = append(attrs, errorType)
		trace.SpanFromContext(ctx).SetAttributes(errorType)
	}
	t.duration.Record(ctx, time.Since(info.started).Seconds(), metric.WithAttributeSet(attribute.NewSet(attrs...)))
	t.tracer.TraceQueryEnd(ctx, conn, data)
}

func dbErrorType(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return fmt.Sprintf("%T", err)
}

// The service emits single-table SQL. Keep only known table tokens to bound cardinality.
func collection(sql string) string {
	for token := range strings.FieldsSeq(sql) {
		switch table := strings.Trim(token, "(),\""); table {
		case "trips", "trip_status_history":
			return table
		}
	}
	return ""
}

```

`internal/infra/postgres/transaction/manager.go`:

```go
package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

func (m *TxManager) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.executor(ctx).Exec(ctx, sql, args...)
}

func (m *TxManager) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.executor(ctx).QueryRow(ctx, sql, args...)
}

func (m *TxManager) executor(ctx context.Context) executor {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return m.pool
}

type TxManager struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

func NewTxManager(pool *pgxpool.Pool, timeout time.Duration) *TxManager {
	return &TxManager{pool: pool, timeout: timeout}
}

const isoLevel = pgx.ReadCommitted

func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	beginCtx, cancel := context.WithTimeout(ctx, m.timeout)
	tx, err := m.pool.BeginTx(beginCtx, pgx.TxOptions{IsoLevel: isoLevel})
	cancel()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = m.rollback(ctx, tx)
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rbErr := m.rollback(ctx, tx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	commitCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	if err := tx.Commit(commitCtx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (m *TxManager) rollback(ctx context.Context, tx pgx.Tx) error {
	// контекст запроса может быть уже отменён, а откатиться и вернуть соединение в пул нужно всё равно
	rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), m.timeout)
	defer cancel()

	if err := tx.Rollback(rbCtx); err != nil {
		return fmt.Errorf("rollback transaction: %w", err)
	}

	return nil
}

```

`internal/infra/telemetry/log.go`:

```go
package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

func NewLogger(service string, level slog.Level) *slog.Logger {
	json := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(correlated{json}).With("service", service)
}

type correlated struct {
	slog.Handler
}

func (h correlated) Handle(ctx context.Context, record slog.Record) error {
	var traceID, spanID string
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		traceID, spanID = span.TraceID().String(), span.SpanID().String()
	}
	record.AddAttrs(slog.String("trace_id", traceID), slog.String("span_id", spanID))
	return h.Handler.Handle(ctx, record)
}

func (h correlated) WithAttrs(attrs []slog.Attr) slog.Handler {
	return correlated{h.Handler.WithAttrs(attrs)}
}

func (h correlated) WithGroup(name string) slog.Handler {
	return correlated{h.Handler.WithGroup(name)}
}

```

`internal/infra/telemetry/providers.go`:

```go
package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Providers struct {
	Traces  *sdktrace.TracerProvider
	Metrics *sdkmetric.MeterProvider
}

func New(ctx context.Context, service string, logger *slog.Logger) Providers {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.WarnContext(ctx, "telemetry export or collection failed", "error", err)
	}))

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", service),
			attribute.String("service.namespace", "tripgo"),
			attribute.String("service.version", buildVersion()),
			attribute.String("deployment.environment", "local"),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		logger.WarnContext(ctx, "some telemetry resource attributes were unavailable", "error", err)
	}

	traces := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	metrics := []sdkmetric.Option{sdkmetric.WithResource(res)}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	switch protocol {
	case "", "http/protobuf", "grpc":
		traceExporter, metricExporter := exporters(ctx, protocol == "grpc", logger)
		if traceExporter != nil {
			traces = append(traces, sdktrace.WithBatcher(traceExporter))
		}
		if metricExporter != nil {
			metrics = append(metrics, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
		}
	default:
		logger.WarnContext(ctx, "unsupported OTLP protocol, export disabled", "protocol", protocol)
	}

	return Providers{
		Traces:  sdktrace.NewTracerProvider(traces...),
		Metrics: sdkmetric.NewMeterProvider(metrics...),
	}
}

func exporters(ctx context.Context, grpc bool, logger *slog.Logger) (sdktrace.SpanExporter, sdkmetric.Exporter) {
	var (
		traces    sdktrace.SpanExporter
		metrics   sdkmetric.Exporter
		traceErr  error
		metricErr error
	)
	if grpc {
		traces, traceErr = otlptracegrpc.New(ctx)
		metrics, metricErr = otlpmetricgrpc.New(ctx)
	} else {
		traces, traceErr = otlptracehttp.New(ctx)
		metrics, metricErr = otlpmetrichttp.New(ctx)
	}
	if traceErr != nil {
		logger.WarnContext(ctx, "could not start trace exporter", "error", traceErr)
		traces = nil
	}
	if metricErr != nil {
		logger.WarnContext(ctx, "could not start metric exporter", "error", metricErr)
		metrics = nil
	}
	return traces, metrics
}

func (p Providers) Shutdown(ctx context.Context) error {
	return errors.Join(p.Metrics.Shutdown(ctx), p.Traces.Shutdown(ctx))
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return strings.TrimSpace(setting.Value)
			}
		}
	}
	return "dev"
}

```

`internal/infra/telemetry/trips.go`:

```go
package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
)

type activeTripCounter interface {
	CountActive(context.Context) (int64, error)
}

type Trips struct {
	created   metric.Int64Counter
	completed metric.Int64Counter
}

func NewTrips(meter metric.Meter, repo activeTripCounter, logger *slog.Logger) (Trips, error) {
	created, err := meter.Int64Counter("tripgo.trips.created", metric.WithUnit("{trip}"))
	if err != nil {
		return Trips{}, fmt.Errorf("register created trips: %w", err)
	}
	completed, err := meter.Int64Counter("tripgo.trips.completed", metric.WithUnit("{trip}"))
	if err != nil {
		return Trips{}, fmt.Errorf("register completed trips: %w", err)
	}
	_, err = meter.Int64ObservableGauge("tripgo.trips.active", metric.WithUnit("{trip}"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			count, err := repo.CountActive(ctx)
			if err != nil {
				logger.WarnContext(ctx, "could not measure active trips", "error", err)
				return nil
			}
			observer.Observe(count)
			return nil
		}))
	if err != nil {
		return Trips{}, fmt.Errorf("register active trips: %w", err)
	}
	return Trips{created: created, completed: completed}, nil
}

func (t Trips) Created(ctx context.Context) {
	t.created.Add(ctx, 1)
}

func (t Trips) Completed(ctx context.Context) {
	t.completed.Add(ctx, 1)
}

```

`internal/model/status_change.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type StatusChange struct {
	TripID    uuid.UUID
	From      *trip.Status
	To        trip.Status
	Reason    string
	ChangedAt time.Time
}

```

`internal/repository/trip/contract.go`:

```go
package trip

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

```

`internal/repository/trip/convert.go`:

```go
package trip

import (
	"github.com/jackc/pgx/v5"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
)

func scanTrip(row pgx.Row) (domain.Trip, error) {
	var (
		trip   domain.Trip
		status string
	)

	err := row.Scan(
		&trip.ID,
		&trip.UserID,
		&trip.DriverID,
		&trip.StartPoint.Latitude,
		&trip.StartPoint.Longitude,
		&trip.EndPoint.Latitude,
		&trip.EndPoint.Longitude,
		&trip.Price,
		&status,
		&trip.StartedAt,
		&trip.FinishedAt,
	)
	if err != nil {
		return domain.Trip{}, err
	}

	trip.Status = domain.Status(status)

	return trip, nil
}

```

`internal/repository/trip/fetch.go`:

```go
package trip

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
)

func (r *TripRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return r.fetch(ctx, id, false)
}

// GetForUpdate holds the row lock until the calling transaction ends.
func (r *TripRepository) GetForUpdate(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return r.fetch(ctx, id, true)
}

func (r *TripRepository) fetch(ctx context.Context, id uuid.UUID, lock bool) (domain.Trip, error) {
	statement := psql.Select(tripColumns...).From(tripsTable).Where(sq.Eq{"id": id})
	if lock {
		statement = statement.Suffix("FOR UPDATE")
	}
	query, args, err := statement.ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build select trip query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	trip, err := scanTrip(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Trip{}, domain.ErrTripNotFound
	}
	if err != nil {
		return domain.Trip{}, fmt.Errorf("select trip %s: %w", id, err)
	}

	return trip, nil
}

func (r *TripRepository) CountActive(ctx context.Context) (int64, error) {
	const query = "SELECT count(*) FROM trips WHERE status = $1"

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	var count int64
	if err := r.db.QueryRow(ctx, query, domain.StatusActive).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active trips: %w", err)
	}
	return count, nil
}

```

`internal/repository/trip/history.go`:

```go
package trip

import (
	"context"
	"fmt"

	"github.com/Chillobass/fork_template/internal/model"
)

func (r *TripRepository) Add(ctx context.Context, change model.StatusChange) error {
	var from *string
	if change.From != nil {
		s := string(*change.From)
		from = &s
	}

	query, args, err := psql.
		Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(change.TripID, from, string(change.To), change.Reason, change.ChangedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert status history query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert status history for trip %s: %w", change.TripID, err)
	}

	return nil
}

```

`internal/repository/trip/repo.go`:

```go
package trip

import (
	"time"

	sq "github.com/Masterminds/squirrel"
)

const (
	tripsTable                 = "trips"
	uniqueViolationCode        = "23505"
	driverActiveTripConstraint = "trips_driver_active_uniq"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var tripColumns = []string{
	"id",
	"user_id",
	"driver_id",
	"start_latitude",
	"start_longitude",
	"end_latitude",
	"end_longitude",
	"price",
	"status",
	"started_at",
	"finished_at",
}

type TripRepository struct {
	db           executor
	queryTimeout time.Duration
}

func New(db executor, queryTimeout time.Duration) *TripRepository {
	return &TripRepository{db: db, queryTimeout: queryTimeout}
}

```

`internal/repository/trip/repo_integration_test.go`:

```go
//go:build integration

package trip_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/infra/postgres/transaction"
	"github.com/Chillobass/fork_template/internal/model"
	repository "github.com/Chillobass/fork_template/internal/repository/trip"
	"github.com/Chillobass/fork_template/internal/usecase/trip/create"
	"github.com/Chillobass/fork_template/internal/usecase/trip/finish"
)

const queryTimeout = 5 * time.Second

func setup(t *testing.T) (*pgxpool.Pool, *transaction.TxManager, *repository.TripRepository, trip.Data) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required. Run tripgoctl environment start and make migrate.")
	}

	ctx, cancel := context.WithTimeout(t.Context(), queryTimeout)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	input := trip.Data{UserID: uuid.New(), DriverID: uuid.New(), Price: 120}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), queryTimeout)
		defer cancel()
		// Deleting the trip removes its history through the migration's foreign key.
		if _, err := pool.Exec(ctx, "DELETE FROM trips WHERE driver_id = $1", input.DriverID); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})

	tx := transaction.NewTxManager(pool, queryTimeout)
	return pool, tx, repository.New(tx, queryTimeout), input
}

func TestTripLifecycle(t *testing.T) {
	pool, tx, repo, input := setup(t)
	ctx, cancel := context.WithTimeout(t.Context(), 4*queryTimeout)
	defer cancel()
	creator := create.New(tx, repo, repo, time.Now)
	created, err := creator.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := creator.Create(ctx, input); !errors.Is(err, trip.ErrDriverBusy) {
		t.Fatalf("duplicate active trip: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil || got.Data != created.Data || got.Status != trip.StatusActive ||
		!got.StartedAt.Equal(created.StartedAt) || got.FinishedAt != nil {
		t.Fatalf("stored trip = (%+v, %v), created %+v", got, err, created)
	}

	finisher := finish.New(tx, repo, repo, time.Now)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	start := make(chan struct{})
	for range 2 {
		wg.Go(func() {
			<-start
			_, err := finisher.Finish(ctx, created.ID)
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	var completed, conflicts int
	for err := range results {
		switch {
		case err == nil:
			completed++
		case errors.Is(err, trip.ErrTripCompleted):
			conflicts++
		default:
			t.Fatalf("concurrent finish: %v", err)
		}
	}
	if completed != 1 || conflicts != 1 {
		t.Fatalf("finish outcomes: %d successes, %d conflicts", completed, conflicts)
	}

	got, err = repo.GetByID(ctx, created.ID)
	if err != nil || got.Status != trip.StatusCompleted || got.FinishedAt == nil {
		t.Fatalf("persisted completion = (%+v, %v)", got, err)
	}
	var transitions int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM trip_status_history WHERE trip_id = $1", created.ID).Scan(&transitions); err != nil {
		t.Fatal(err)
	}
	if transitions != 2 {
		t.Fatalf("expected creation and one completion, got %d transitions", transitions)
	}
	if _, err := creator.Create(ctx, input); err != nil {
		t.Fatalf("driver still busy after completion: %v", err)
	}
}

func TestRollbackInserts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status trip.Status
	}{
		{name: "second insert fails", status: "invalid-status"},
		{name: "both inserts succeed before rollback", status: trip.StatusActive},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool, tx, repo, input := setup(t)
			ctx, cancel := context.WithTimeout(t.Context(), 4*queryTimeout)
			defer cancel()
			at := time.Now().UTC()
			value, err := trip.New(uuid.New(), input, at)
			if err != nil {
				t.Fatal(err)
			}
			forced := errors.New("force rollback after both inserts")
			err = tx.Do(ctx, func(ctx context.Context) error {
				if err := repo.Create(ctx, value); err != nil {
					return err
				}
				if err := repo.Add(ctx, model.StatusChange{TripID: value.ID, To: tc.status, ChangedAt: at}); err != nil {
					return err
				}
				return forced
			})
			if err == nil || errors.Is(err, forced) != (tc.status == trip.StatusActive) {
				t.Fatalf("unexpected transaction error: %v", err)
			}
			if _, err := repo.GetByID(ctx, value.ID); !errors.Is(err, trip.ErrTripNotFound) {
				t.Fatalf("trip insert was not rolled back: %v", err)
			}
			var history int
			if err := pool.QueryRow(ctx, "SELECT count(*) FROM trip_status_history WHERE trip_id = $1", value.ID).Scan(&history); err != nil {
				t.Fatal(err)
			}
			if history != 0 {
				t.Fatalf("history insert was not rolled back: %d rows", history)
			}
		})
	}
}

```

`internal/repository/trip/save.go`:

```go
package trip

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
)

func (r *TripRepository) Create(ctx context.Context, trip domain.Trip) error {
	query, args, err := psql.
		Insert(tripsTable).
		Columns(tripColumns...).
		Values(
			trip.ID,
			trip.UserID,
			trip.DriverID,
			trip.StartPoint.Latitude,
			trip.StartPoint.Longitude,
			trip.EndPoint.Latitude,
			trip.EndPoint.Longitude,
			trip.Price,
			string(trip.Status),
			trip.StartedAt,
			trip.FinishedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	_, err = r.db.Exec(ctx, query, args...)
	if isUniqueViolation(err, driverActiveTripConstraint) {
		return fmt.Errorf("insert active trip: %w: %w", domain.ErrDriverBusy, err)
	}
	if err != nil {
		return fmt.Errorf("insert trip %s: %w", trip.ID, err)
	}

	return nil
}

func (r *TripRepository) Update(ctx context.Context, trip domain.Trip) error {
	query, args, err := psql.
		Update(tripsTable).
		Set("status", string(trip.Status)).
		Set("finished_at", trip.FinishedAt).
		Set("updated_at", trip.FinishedAt).
		Where(sq.Eq{"id": trip.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update trip query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("update trip %s: %w", trip.ID, err)
	}
	return nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == uniqueViolationCode &&
		pgErr.ConstraintName == constraint
}

```

`internal/usecase/trip/create/contract.go`:

```go
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package create

import (
	"context"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
)

type txManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}

type historyRepository interface {
	Add(ctx context.Context, change model.StatusChange) error
}

type tripRepository interface {
	Create(ctx context.Context, trip trip.Trip) error
}

```

`internal/usecase/trip/create/mocks_test.go`:

```go
// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go
//
// Generated by this command:
//
//	mockgen -source contract.go -destination mocks_test.go -package create_test
//

// Package create_test is a generated GoMock package.
package create_test

import (
	context "context"
	reflect "reflect"

	trip "github.com/Chillobass/fork_template/internal/domain/trip"
	model "github.com/Chillobass/fork_template/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MocktxManager is a mock of txManager interface.
type MocktxManager struct {
	ctrl     *gomock.Controller
	recorder *MocktxManagerMockRecorder
	isgomock struct{}
}

// MocktxManagerMockRecorder is the mock recorder for MocktxManager.
type MocktxManagerMockRecorder struct {
	mock *MocktxManager
}

// NewMocktxManager creates a new mock instance.
func NewMocktxManager(ctrl *gomock.Controller) *MocktxManager {
	mock := &MocktxManager{ctrl: ctrl}
	mock.recorder = &MocktxManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktxManager) EXPECT() *MocktxManagerMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MocktxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", ctx, fn)
	ret0, _ := ret[0].(error)
	return ret0
}

// Do indicates an expected call of Do.
func (mr *MocktxManagerMockRecorder) Do(ctx, fn any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MocktxManager)(nil).Do), ctx, fn)
}

// MockhistoryRepository is a mock of historyRepository interface.
type MockhistoryRepository struct {
	ctrl     *gomock.Controller
	recorder *MockhistoryRepositoryMockRecorder
	isgomock struct{}
}

// MockhistoryRepositoryMockRecorder is the mock recorder for MockhistoryRepository.
type MockhistoryRepositoryMockRecorder struct {
	mock *MockhistoryRepository
}

// NewMockhistoryRepository creates a new mock instance.
func NewMockhistoryRepository(ctrl *gomock.Controller) *MockhistoryRepository {
	mock := &MockhistoryRepository{ctrl: ctrl}
	mock.recorder = &MockhistoryRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockhistoryRepository) EXPECT() *MockhistoryRepositoryMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockhistoryRepository) Add(ctx context.Context, change model.StatusChange) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, change)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockhistoryRepositoryMockRecorder) Add(ctx, change any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockhistoryRepository)(nil).Add), ctx, change)
}

// MocktripRepository is a mock of tripRepository interface.
type MocktripRepository struct {
	ctrl     *gomock.Controller
	recorder *MocktripRepositoryMockRecorder
	isgomock struct{}
}

// MocktripRepositoryMockRecorder is the mock recorder for MocktripRepository.
type MocktripRepositoryMockRecorder struct {
	mock *MocktripRepository
}

// NewMocktripRepository creates a new mock instance.
func NewMocktripRepository(ctrl *gomock.Controller) *MocktripRepository {
	mock := &MocktripRepository{ctrl: ctrl}
	mock.recorder = &MocktripRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktripRepository) EXPECT() *MocktripRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MocktripRepository) Create(ctx context.Context, arg1 trip.Trip) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MocktripRepositoryMockRecorder) Create(ctx, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MocktripRepository)(nil).Create), ctx, arg1)
}

```

`internal/usecase/trip/create/usecase.go`:

```go
package create

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
)

type Usecase struct {
	tx      txManager
	trips   tripRepository
	history historyRepository
	now     func() time.Time
}

func New(tx txManager, trips tripRepository, history historyRepository, now func() time.Time) *Usecase {
	return &Usecase{tx: tx, trips: trips, history: history, now: now}
}

func (u *Usecase) Create(ctx context.Context, input trip.Data) (trip.Trip, error) {
	// Match PostgreSQL timestamp precision in the first response and in GET.
	now := u.now().UTC().Truncate(time.Microsecond)
	created, err := trip.New(uuid.New(), input, now)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("validate trip: %w", err)
	}

	err = u.tx.Do(ctx, func(ctx context.Context) error {
		if err := u.trips.Create(ctx, created); err != nil {
			return err
		}
		return u.history.Add(ctx, model.StatusChange{
			TripID: created.ID, To: created.Status, Reason: "trip created", ChangedAt: now,
		})
	})
	if err != nil {
		return trip.Trip{}, fmt.Errorf("create trip: %w", err)
	}
	return created, nil
}

```

`internal/usecase/trip/create/usecase_test.go`:

```go
package create_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
	"github.com/Chillobass/fork_template/internal/usecase/trip/create"
)

func TestCreate(t *testing.T) {
	dbFailure := errors.New("database unavailable")
	for _, tc := range []struct {
		name       string
		insertErr  error
		historyErr error
	}{
		{name: "success"},
		{name: "busy driver", insertErr: trip.ErrDriverBusy},
		{name: "history error", historyErr: dbFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMocktxManager(ctrl)
			repo := NewMocktripRepository(ctrl)
			history := NewMockhistoryRepository(ctrl)
			now := time.Date(2026, 10, 2, 8, 0, 0, 123456789, time.FixedZone("UTC+3", 10800))
			at := now.UTC().Truncate(time.Microsecond)
			input := trip.Data{UserID: uuid.New(), DriverID: uuid.New(), Price: 10}
			var inserted trip.Trip

			ctx := t.Context()
			type transactionKey struct{}
			txCtx := context.WithValue(ctx, transactionKey{}, true)
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(txCtx)
			})
			repo.EXPECT().Create(txCtx, gomock.Any()).DoAndReturn(func(_ context.Context, value trip.Trip) error {
				if value.ID == uuid.Nil || value.Data != input || value.Status != trip.StatusActive ||
					value.StartedAt != at || value.FinishedAt != nil {
					t.Fatalf("inserted trip = %+v", value)
				}
				inserted = value
				return tc.insertErr
			})
			if tc.insertErr == nil {
				history.EXPECT().Add(txCtx, gomock.Cond(func(change model.StatusChange) bool {
					return change.TripID == inserted.ID && change.From == nil && change.To == trip.StatusActive &&
						change.ChangedAt.Equal(at) && change.Reason == "trip created"
				})).Return(tc.historyErr)
			}

			uc := create.New(tx, repo, history, func() time.Time { return now })
			got, err := uc.Create(ctx, input)
			wantErr := tc.insertErr
			if wantErr == nil {
				wantErr = tc.historyErr
			}
			if wantErr != nil {
				if !errors.Is(err, wantErr) || got != (trip.Trip{}) {
					t.Fatalf("create = (%+v, %v), want zero trip and %v", got, err, wantErr)
				}
				return
			}
			if err != nil || got != inserted {
				t.Fatalf("create = (%+v, %v), want %+v", got, err, inserted)
			}
		})
	}
}

func TestCreateRejectsInvalidDataBeforeTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	uc := create.New(NewMocktxManager(ctrl), NewMocktripRepository(ctrl), NewMockhistoryRepository(ctrl), time.Now)
	var validation *trip.ValidationError
	if _, err := uc.Create(t.Context(), trip.Data{}); !errors.As(err, &validation) {
		t.Fatalf("create invalid trip: %v", err)
	}
}

```

`internal/usecase/trip/finish/contract.go`:

```go
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
)

type txManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}

type historyRepository interface {
	Add(ctx context.Context, change model.StatusChange) error
}

type tripRepository interface {
	GetForUpdate(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	Update(ctx context.Context, trip trip.Trip) error
}

```

`internal/usecase/trip/finish/mocks_test.go`:

```go
// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go
//
// Generated by this command:
//
//	mockgen -source contract.go -destination mocks_test.go -package finish_test
//

// Package finish_test is a generated GoMock package.
package finish_test

import (
	context "context"
	reflect "reflect"

	trip "github.com/Chillobass/fork_template/internal/domain/trip"
	model "github.com/Chillobass/fork_template/internal/model"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// MocktxManager is a mock of txManager interface.
type MocktxManager struct {
	ctrl     *gomock.Controller
	recorder *MocktxManagerMockRecorder
	isgomock struct{}
}

// MocktxManagerMockRecorder is the mock recorder for MocktxManager.
type MocktxManagerMockRecorder struct {
	mock *MocktxManager
}

// NewMocktxManager creates a new mock instance.
func NewMocktxManager(ctrl *gomock.Controller) *MocktxManager {
	mock := &MocktxManager{ctrl: ctrl}
	mock.recorder = &MocktxManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktxManager) EXPECT() *MocktxManagerMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MocktxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", ctx, fn)
	ret0, _ := ret[0].(error)
	return ret0
}

// Do indicates an expected call of Do.
func (mr *MocktxManagerMockRecorder) Do(ctx, fn any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MocktxManager)(nil).Do), ctx, fn)
}

// MockhistoryRepository is a mock of historyRepository interface.
type MockhistoryRepository struct {
	ctrl     *gomock.Controller
	recorder *MockhistoryRepositoryMockRecorder
	isgomock struct{}
}

// MockhistoryRepositoryMockRecorder is the mock recorder for MockhistoryRepository.
type MockhistoryRepositoryMockRecorder struct {
	mock *MockhistoryRepository
}

// NewMockhistoryRepository creates a new mock instance.
func NewMockhistoryRepository(ctrl *gomock.Controller) *MockhistoryRepository {
	mock := &MockhistoryRepository{ctrl: ctrl}
	mock.recorder = &MockhistoryRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockhistoryRepository) EXPECT() *MockhistoryRepositoryMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockhistoryRepository) Add(ctx context.Context, change model.StatusChange) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, change)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockhistoryRepositoryMockRecorder) Add(ctx, change any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockhistoryRepository)(nil).Add), ctx, change)
}

// MocktripRepository is a mock of tripRepository interface.
type MocktripRepository struct {
	ctrl     *gomock.Controller
	recorder *MocktripRepositoryMockRecorder
	isgomock struct{}
}

// MocktripRepositoryMockRecorder is the mock recorder for MocktripRepository.
type MocktripRepositoryMockRecorder struct {
	mock *MocktripRepository
}

// NewMocktripRepository creates a new mock instance.
func NewMocktripRepository(ctrl *gomock.Controller) *MocktripRepository {
	mock := &MocktripRepository{ctrl: ctrl}
	mock.recorder = &MocktripRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktripRepository) EXPECT() *MocktripRepositoryMockRecorder {
	return m.recorder
}

// GetForUpdate mocks base method.
func (m *MocktripRepository) GetForUpdate(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetForUpdate", ctx, id)
	ret0, _ := ret[0].(trip.Trip)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetForUpdate indicates an expected call of GetForUpdate.
func (mr *MocktripRepositoryMockRecorder) GetForUpdate(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetForUpdate", reflect.TypeOf((*MocktripRepository)(nil).GetForUpdate), ctx, id)
}

// Update mocks base method.
func (m *MocktripRepository) Update(ctx context.Context, arg1 trip.Trip) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MocktripRepositoryMockRecorder) Update(ctx, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MocktripRepository)(nil).Update), ctx, arg1)
}

```

`internal/usecase/trip/finish/usecase.go`:

```go
package finish

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
)

type Usecase struct {
	tx      txManager
	trips   tripRepository
	history historyRepository
	now     func() time.Time
}

func New(tx txManager, trips tripRepository, history historyRepository, now func() time.Time) *Usecase {
	return &Usecase{tx: tx, trips: trips, history: history, now: now}
}

func (u *Usecase) Finish(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	var finished trip.Trip
	err := u.tx.Do(ctx, func(ctx context.Context) error {
		current, err := u.trips.GetForUpdate(ctx, id)
		if err != nil {
			return err
		}

		from := current.Status
		now := u.now().UTC().Truncate(time.Microsecond)
		if err := current.Finish(now); err != nil {
			return err
		}
		if err := u.trips.Update(ctx, current); err != nil {
			return err
		}
		if err := u.history.Add(ctx, model.StatusChange{
			TripID: current.ID, From: &from, To: current.Status, Reason: "trip finished", ChangedAt: now,
		}); err != nil {
			return err
		}
		finished = current
		return nil
	})
	if err != nil {
		return trip.Trip{}, fmt.Errorf("finish trip: %w", err)
	}
	return finished, nil
}

```

`internal/usecase/trip/finish/usecase_test.go`:

```go
package finish_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/model"
	"github.com/Chillobass/fork_template/internal/usecase/trip/finish"
)

func TestFinish(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   trip.Status
		fetchErr error
		wantErr  error
	}{
		{name: "active", status: trip.StatusActive},
		{name: "completed", status: trip.StatusCompleted, wantErr: trip.ErrTripCompleted},
		{name: "missing", fetchErr: trip.ErrTripNotFound, wantErr: trip.ErrTripNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMocktxManager(ctrl)
			repo := NewMocktripRepository(ctrl)
			history := NewMockhistoryRepository(ctrl)
			now := time.Date(2026, 10, 3, 8, 0, 0, 123456789, time.UTC)
			at := now.Truncate(time.Microsecond)
			original := trip.Trip{ID: uuid.New(), Status: tc.status, StartedAt: at.Add(-time.Hour)}
			if original.Status == trip.StatusCompleted {
				original.FinishedAt = new(at.Add(-time.Minute))
			}

			ctx := t.Context()
			type transactionKey struct{}
			txCtx := context.WithValue(ctx, transactionKey{}, true)
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(txCtx)
			})
			fetch := repo.EXPECT().GetForUpdate(txCtx, original.ID).Return(original, tc.fetchErr)

			var want trip.Trip
			if tc.wantErr == nil {
				want = original
				want.Status = trip.StatusCompleted
				want.FinishedAt = &at
				gomock.InOrder(
					fetch,
					repo.EXPECT().Update(txCtx, want),
					history.EXPECT().Add(txCtx, model.StatusChange{
						TripID: original.ID, From: new(trip.StatusActive), To: trip.StatusCompleted,
						ChangedAt: at, Reason: "trip finished",
					}),
				)
			}

			got, err := finish.New(tx, repo, history, func() time.Time { return now }).Finish(ctx, original.ID)
			if !errors.Is(err, tc.wantErr) || got.ID != want.ID || got.Status != want.Status || got.StartedAt != want.StartedAt {
				t.Fatalf("finish = (%+v, %v), want (%+v, %v)", got, err, want, tc.wantErr)
			}
			if want.FinishedAt != nil && (got.FinishedAt == nil || !got.FinishedAt.Equal(*want.FinishedAt)) {
				t.Fatalf("finished_at = %v, want %v", got.FinishedAt, want.FinishedAt)
			}
		})
	}
}

```

`internal/usecase/trip/get/contract.go`:

```go
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package get

import (
	"context"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type tripRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}

```

`internal/usecase/trip/get/mocks_test.go`:

```go
// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go
//
// Generated by this command:
//
//	mockgen -source contract.go -destination mocks_test.go -package get_test
//

// Package get_test is a generated GoMock package.
package get_test

import (
	context "context"
	reflect "reflect"

	trip "github.com/Chillobass/fork_template/internal/domain/trip"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// MocktripRepository is a mock of tripRepository interface.
type MocktripRepository struct {
	ctrl     *gomock.Controller
	recorder *MocktripRepositoryMockRecorder
	isgomock struct{}
}

// MocktripRepositoryMockRecorder is the mock recorder for MocktripRepository.
type MocktripRepositoryMockRecorder struct {
	mock *MocktripRepository
}

// NewMocktripRepository creates a new mock instance.
func NewMocktripRepository(ctrl *gomock.Controller) *MocktripRepository {
	mock := &MocktripRepository{ctrl: ctrl}
	mock.recorder = &MocktripRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocktripRepository) EXPECT() *MocktripRepositoryMockRecorder {
	return m.recorder
}

// GetByID mocks base method.
func (m *MocktripRepository) GetByID(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(trip.Trip)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MocktripRepositoryMockRecorder) GetByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MocktripRepository)(nil).GetByID), ctx, id)
}

```

`internal/usecase/trip/get/usecase.go`:

```go
package get

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type Usecase struct {
	trips tripRepository
}

func New(trips tripRepository) *Usecase {
	return &Usecase{trips: trips}
}

func (u *Usecase) Get(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	found, err := u.trips.GetByID(ctx, id)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("get trip: %w", err)
	}
	return found, nil
}

```

`internal/usecase/trip/get/usecase_test.go`:

```go
package get_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/usecase/trip/get"
)

func TestGet(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "found"},
		{name: "missing", err: trip.ErrTripNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMocktripRepository(gomock.NewController(t))
			id := uuid.New()
			var want trip.Trip
			if tc.err == nil {
				want = trip.Trip{ID: id, Status: trip.StatusActive}
			}
			ctx := t.Context()
			repo.EXPECT().GetByID(ctx, id).Return(want, tc.err)

			got, err := get.New(repo).Get(ctx, id)
			if !errors.Is(err, tc.err) || got != want {
				t.Fatalf("get = (%+v, %v), want (%+v, %v)", got, err, want, tc.err)
			}
		})
	}
}

```

`migrations/00001_create_trips.sql`:

```sql
-- +goose Up
CREATE TABLE trips (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL,
    driver_id       UUID NOT NULL,

    start_latitude  DOUBLE PRECISION NOT NULL CHECK (start_latitude BETWEEN -90 AND 90),
    start_longitude DOUBLE PRECISION NOT NULL CHECK (start_longitude BETWEEN -180 AND 180),
    end_latitude    DOUBLE PRECISION NOT NULL CHECK (end_latitude BETWEEN -90 AND 90),
    end_longitude   DOUBLE PRECISION NOT NULL CHECK (end_longitude BETWEEN -180 AND 180),

    price           BIGINT NOT NULL CHECK (price >= 0),
    status          TEXT NOT NULL CHECK (status IN ('active', 'completed')),

    started_at      TIMESTAMPTZ NOT NULL,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (
        (status = 'active'    AND finished_at IS NULL) OR
        (status = 'completed' AND finished_at IS NOT NULL)
    )
);

CREATE INDEX trips_status_started_at_idx ON trips (status, started_at);

CREATE UNIQUE INDEX trips_driver_active_uniq ON trips (driver_id) WHERE status = 'active';

-- +goose Down
DROP TABLE trips;

```

`migrations/00002_create_trip_status_history.sql`:

```sql
-- +goose Up
CREATE TABLE trip_status_history (
    id          BIGSERIAL PRIMARY KEY,
    trip_id     UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status   TEXT NOT NULL CHECK (to_status IN ('active', 'completed')),
    reason      TEXT,
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX trip_status_history_trip_changed_idx
    ON trip_status_history (trip_id, changed_at);

-- +goose Down
DROP TABLE trip_status_history;

```