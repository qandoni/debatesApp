# debatesApp

Социальная сеть с дебатами: посты, аргументированные дебаты со сторонами, комментарии,
оценки аргументов и голосование. Бэкенд на Go (gin + PostgreSQL + MinIO) с realtime-событиями
по WebSocket.

## Быстрый старт

```bash
cp .env.example .env      # заполнить значения (POSTGRES_*, HTTP_ADDR, MINIO_*, WS_*)
make env-up               # PostgreSQL (docker compose)
make env-port-forward     # проброс 127.0.0.1:5432 → Postgres (нужен приложению и тестам)
make minio-up             # MinIO (docker compose)
make migrate-up           # миграции
make debatesApp-run       # приложение → http://localhost:5050
```

У контейнера Postgres нет пробросов портов на хост: приложение из `make debatesApp-run` подключается к
`localhost:5432` через `socat`-контейнер `port-forwarder`, поэтому `make env-port-forward` обязателен
(без него падение на инициализации пула соединений).

Полезное:

| Команда | Что делает |
|---|---|
| `make env-port-forward` / `make env-port-close` | Проброс / остановка проброса порта Postgres на хост |
| `make env-down` / `make minio-down` | Остановить контейнеры |
| `make migrate-down` | Откатить миграции |
| `make migrate-create seq=name` | Новая пара миграций |
| `make ps` | Статус контейнеров |
| `make env-cleanup` | Полная очистка volume (опасно: удалит данные) |

## Проверки

```bash
gofmt -l cmd internal test
go vet ./cmd/... ./internal/... ./test/...
go test ./internal/... -race
```

## Конфигурация

Все переменные — в `.env` (шаблон: [.env.example](./.env.example)):

- `POSTGRES_*` — подключение к БД.
- `HTTP_ADDR`, `HTTP_SHUTDOWN_TIMEOUT` — HTTP-сервер (локально `:5050`).
- `LOGGER_LEVEL` — уровень логов.
- `MINIO_*` — хранилище файлов (аватары, картинки постов).
- `WS_*` — realtime-транспорт:
  - `WS_ALLOWED_ORIGINS` — whitelist Origin для WebSocket-handshake (локально `http://localhost:3000`);
  - `WS_READ_LIMIT` (4096) — максимум байт на входящий кадр;
  - `WS_WRITE_WAIT` (10s), `WS_PONG_WAIT` (60s), `WS_PING_PERIOD` (54s) — таймауты;
  - `WS_SEND_BUFFER_SIZE` (64) — очередь исходящих кадров на клиента;
  - `WS_AUTH_WAIT` (10s) — окно для первого `auth`-кадра.

## Документация

| Документ | Для кого |
|---|---|
| [docs/api.md](./docs/api.md) | REST API: маршруты, авторизация, форматы ошибок |
| [docs/websocket.md](./docs/websocket.md) | Realtime: подключение, протокол, 8 типов событий, реконнект, TS-типы |

Кратко о realtime: `GET /api/v1/ws` → кадр `auth` (токен из `/auth/login`) → `ready` →
`subscribe` на пост (`{"post_id": 7}`) → ack `{"type":"subscribed","data":{"post_id":7}}` → события вида
`{"type":"debate.vote.created","topic":"post:7","data":{…},"occurred_at":…}`.
Подробности и ограничения — в [websocket.md](./docs/websocket.md).

## Структура

```
cmd/debates_app/        точка входа
internal/core/          инфраструктура:
  auth, config, errors, logger, password   — базовые сервисы
  domain, enum                              — доменные модели
  realtime                                  — Hub событий (топики, публикация)
  repository, transport                     — репозитории и транспортный слой (HTTP + WS)
internal/features/      фичи: auth, comments, images, posts, statistics, storage, users
test/integration/       интеграционные тесты
migrations/             SQL-миграции
```

Доменные фичи (`comments`, `posts`, `statistics`, `users`) устроены одинаково: `service`
(бизнес-логика и публикация realtime-событий) + `transport` (HTTP-хендлеры и DTO) + `repository`
(SQL). Вспомогательные (`auth`, `images`, `storage`) — без полного набора слоёв. Подфичи живут
внутри: `comments/comment_ratings`, `posts/debate_votes`, `posts/debates` и т.д.
