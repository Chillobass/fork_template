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
