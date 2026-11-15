# AI Buddy — обзор проекта

Помощник стажёра: чат с ассистентом, профиль, команда, каталог сервисов, база знаний,
чек-лист адаптации, заметки и доски. Один веб-интерфейс, за ним несколько Go-сервисов,
Python-RAG и общая обвязка в Docker Compose.

Документ описывает, что проект умеет и из чего состоит. Контракты и детали по сервисам —
рядом: `docs/api/openapi-*.yaml`, `docs/specs/*.md`, `docs/service-boundaries.md`.

---

## 1. Экраны фронтенда

Интерфейс — один SPA (React + Vite), навигация слева, экраны переключаются без перезагрузки.

| Экран | Что делает |
|---|---|
| **Твой AI Buddy** | чат с ассистентом: история чатов, сообщения, потоковые ответы через WebSocket, источники и действия под ответом |
| **Мой профиль** | профиль сотрудника: имя, должность, команда, инструменты |
| **Команды** | список команд, состав, у каждой — свой бадди и цвет |
| **Сервисы** | внутренние технические сервисы: описание, ответственные, заявки на доступ |
| **База знаний** | документы онбординга, которые подмешиваются ассистенту как контекст |
| **Заметки** | личные заметки: создать, изменить, удалить (оптимистичные версии) |
| **Доски** | личные доски на Excalidraw: холст, автосохранение, версии |

Чек-лист адаптации живёт на экране профиля: отмечаешь задачи — растёт питомец
(до 25% он «животик», к 50% получает шляпу), прогресс сохраняется на бэкенде.

---

## 2. Сервисы

Все сервисы поднимаются одним `deploy/compose.yaml`. Наружу опубликованы только
loopback-адреса (`127.0.0.1`), остальное видно лишь внутри docker-сети.

| Сервис | Порт | Что делает |
|---|---|---|
| `frontend` | 8080 | SPA + nginx: отдаёт статику, проксирует `/api/` в шлюз, `/notes-api/` и `/board-api/` — прямо в note-service. Имя сервиса резолвится на каждый запрос, поэтому после пересоздания note-service перезапускать фронт не нужно |
| `gateway` | 8080 (внутр.) | единая точка входа API: маршрутизация по путям, проверка JWT, служебный токен для внутренних вызовов |
| `auth` | 8081 | вход: пароль + код из письма, Google OAuth, обновление сессии, сброс пароля |
| `employee` | 8082 | профили, люди, команды, технические сервисы, инструменты, заявки на доступ (БД `info_db`) |
| `chat` | 8083 | единственный чат: REST + WebSocket, контекст, источники, действия |
| `ragmock` | 8084 | детерминированные ответы ассистента без внешнего AI (для демо) |
| `corpmock` | 8085 | имитация корпоративных заявок |
| `note_service` | 8079 | заметки, доски, документы, чек-лист и прогресс; единый API данных для RAG |
| `rag` | профиль `rag` | Python RAG: индексация Note → embeddings → pgvector → LLM |
| `mailpit` | 8025 | локальный перехватчик почты: коды входа и письма видны в вебе |
| `postgres` | внутр. | общая БД: `auth_db`, `info_db` и базы чата |
| `note_postgres` | внутр. | своя БД `notes` для note-service |
| `rag_postgres` | профиль `rag` | pgvector-хранилище для RAG |
| `prometheus`, `grafana` | 9090, 3000 | метрики и дашборды (см. раздел 7) |

`LLM-service` в репозитории есть, но в compose не подключён — это заготовка (пока только `src/`).

---

## 3. API

Все публичные пути идут через шлюз (`/api/v1/...`). Шлюз сам проверяет токен, вырезает
клиентские `X-User-ID`/`X-Role`/`X-Forwarded-*` и подставляет служебный `X-Service-Token`
для внутренних вызовов; недоступный сервис превращается в `502 upstream_unavailable`.

### auth
```
POST /api/v1/auth/register          POST /api/v1/auth/login
POST /api/v1/auth/confirm           POST /api/v1/auth/refresh
POST /api/v1/auth/logout            GET  /api/v1/auth/me
POST /api/v1/auth/password/reset    POST /api/v1/auth/password/confirm
GET  /api/v1/auth/google/login      GET  /api/v1/auth/google/callback
```

### employee
```
GET   /api/v1/profile                     PATCH /api/v1/profile
GET   /api/v1/teams        /api/v1/teams/{id}
GET   /api/v1/services     /api/v1/services/{id}
GET   /api/v1/tools        GET   /api/v1/access-requests
POST  /api/v1/access-requests
```
Служебные (без публичной авторизации, по сервисному токену): `GET /internal/v1/context`,
`/internal/v1/services`, `/internal/v1/teams` и их `/{id}`.

### chat
```
GET    /api/v1/chats                POST  /api/v1/chats
GET    /api/v1/chats/{id}           PATCH /api/v1/chats/{id}      DELETE /api/v1/chats/{id}
GET    /api/v1/chats/{id}/messages  POST  /api/v1/chats/{id}/messages
POST   /api/v1/chats/{id}/messages/{messageId}/answer
GET    /api/v1/chat-ws/  (WebSocket)
```

### note_service
```
POST   /api/v1/notes                 GET    /api/v1/notes            (limit/offset)
GET    /api/v1/notes/{id}            PATCH  /api/v1/notes/{id}      DELETE /api/v1/notes/{id}
POST   /api/v1/boards                GET    /api/v1/boards           (список без сцен)
GET    /api/v1/boards/{id}           PATCH  /api/v1/boards/{id}     DELETE /api/v1/boards/{id}
GET    /api/v1/checklist             PATCH  /api/v1/checklist/tasks/{id}
GET    /api/v1/information           GET    /api/v1/information/{id}
GET    /internal/*path               (данные для RAG и индексатора)
GET    /healthz                      GET    /metrics                GET /swagger/index.html
```

### ragmock / corpmock
```
POST /internal/v1/generate            (ragmock: детерминированный ответ ассистента)
POST /internal/v1/access-requests     (corpmock: имитация согласования заявки)
```

---

## 4. Данные

- **Общая БД** `postgres` с миграциями по наборам: `shared/migrations/auth`, `shared/migrations/chat`,
  `shared/migrations/info`. Миграции встроены в бинари через `go:embed`, применяются при старте
  сервиса; версия — порядковый номер в списке, а не имя файла.
- **note-service** держит свою БД `notes` и свои миграции в `note-service/migrations`
  (накатывает сервис `note_migrate`):
  - `note.notes` — заметки (заголовок, текст, версия, владелец);
  - `board.boards` — доски: `title`, `scene jsonb` (до 8 МиБ), версия;
  - `onboarding.*` — документы, чек-лист, прогресс.
- **Владение данными** — по границам сервисов (`docs/service-boundaries.md`): заметки и доски
  личные, владелец берётся из подписанного токена, а не из параметров запроса.
- **Демо-данные** наполняются при первом подъёме: команды, люди, сервисы, инструменты,
  документы онбординга.

Ключевая деталь защиты: `user_id` в запросе сверяется с владельцем из JWT на уровне сервиса,
иначе получился бы доступ к чужим записям по подмене параметра.

---

## 5. Инфраструктура

```
deploy/compose.yaml        весь стек одним файлом
deploy/configure.sh        генерирует .env при первом запуске, существующий не трогает
deploy/vpn-route.sh        правило маршрутизации для докера (см. грабли ниже)
Makefile                   цели запуска
.shared/platform/          общий Go-код: конфиг, httpapi, security, postgres, metrics, server
```

Цели `make` из корня репозитория:

| Цель | Что делает |
|---|---|
| `make up` | поднять всё: конфиг, сборка образов, основной стек, затем Prometheus и Grafana |
| `make down` | погасить: сначала метрики, потом основной стек |
| `make logs` | логи основного стека |
| `make build` / `make config` | только сборка / только проверка конфигов |
| `make metrics-up` / `metrics-down` / `metrics-logs` | управлять только обвязкой метрик |
| `make rag` / `make rag-import` | поднять RAG-профиль / импортировать в него знания |

Сети: `app` (сервисы между собой) и `database` (БД, изолирована от основного трафика).
Сервисы обращаются друг к другу по именам (`gateway`, `note_service`), не по `127.0.0.1`.

---

## 6. Вход и демо-доступы

- Демо-аккаунты: `backend@buddy.test`, `frontend@buddy.test`, `analytics@buddy.test`,
  `qa@buddy.test`, `backend2@buddy.test`. Пароль: `BuddyDemo2026!`.
- После пароля сервис присылает **код из письма** (одноразовый, TTL 7 минут). В локальном
  запуске письма не уходят наружу — они видны в Mailpit: http://localhost:8025
- Вход через Google работает при настроенных Google-ключах; аккаунт создаётся только после
  подтверждения email.
- Войти можно и по Google, и по паролю — это два способа к одной учётной записи.

Открывать сайт нужно **на одном хостнейме** — `http://localhost:8080`
(см. грабли: куки сессии и состояния OAuth привязаны к хосту, `127.0.0.1` и `localhost` для них разные).

---

## 7. Наблюдаемость

Метрики собираются со **всех** сервисов и описаны кодом, а не руками.

- Считалки лежат в `shared/platform/metrics`: счётчик запросов по сервису/методу/пути/коду,
  гистограмма длительности, число запросов в обработке. Путь берётся **шаблоном маршрута**
  (`/api/v1/notes/:id`), а не сырым URL — иначе каждый id создавал бы новую серию.
- Включаются в одном месте — `shared/platform/server.Run(name, build)`: поэтому `/metrics`
  есть сразу у auth, employee, chat и gateway, а имя сервиса становится меткой `service`.
  note-service работает на gin и подключает те же считалки своим адаптером.
- `/metrics` доступен без авторизации, но только внутри сети — его скрейпит Prometheus.
- Ограничитель частоты в note-service: 20 запросов/с пачкой до 40 на владельца, отказ — `429`
  с `Retry-After`; настраивается `RATE_LIMIT_RPS` и `RATE_LIMIT_BURST`.

Стек обвязки:

```
observability/compose.yaml                 отдельный проект ai-buddy-observability
observability/prometheus/prometheus.yml    цели через file_sd: targets/*.json
observability/prometheus/targets/*.json    по файлу на сервис — владеет команда сервиса
observability/grafana/provisioning/**      источник данных и провайдер дашбордов
<service>/observability/dashboards/*.json  дашборд сервиса живёт рядом с сервисом
```

Дашборды в Grafana разложены по папкам с именами сервисов; домашняя страница Grafana сразу
открывает дашборд note-service. Адреса: Grafana http://127.0.0.1:3000 (`admin`/`admin`),
Prometheus http://127.0.0.1:9090 (`/targets` — состояние целей).

---

## 8. Доски (Excalidraw)

- Хранится сцена холста: `elements` + `files`. Состояние приложения (`appState`) не хранится,
  из него сохраняются только фон и сетка — иначе автосохранение срабатывало бы на каждое движение.
- Сохранение с дебаунсом, при закрытии вкладки досылается; при конфликте версий — `409` и
  сообщение «изменили в другой вкладке».
- Сцена шлётся целиком, поэтому поднят лимит тела: nginx `12m`, сервис 8 МиБ, и такой же
  CHECK в базе.
- Шрифты Excalidraw лежат в репозитории (самохостинг), внешний CDN не используется.

---

## 9. Тесты и проверки

| Что | Где |
|---|---|
| Юнит-тесты Go | рядом с кодом (`shared/platform/...`, `note-service/internal/core/...`) |
| Интеграционные тесты основного приложения | `tests/integration/` — поднимают реальные БД, ходят по HTTP |
| Тесты фронтенда | `frontend/src/**/*.test.tsx` (vitest) |
| Тесты RAG | `rag-service/tests/` (pytest) |
| Контракты | `docs/api/openapi-*.yaml` + примеры `docs/api/info-examples.http` |

Интеграционные тесты умеют выгружать реальные ответы в `reports/contract-fixtures.json`
(переменная `CONTRACT_FIXTURES_OUT`) — по ним ответы сверяются со спеками в пайплайне
команды, поэтому изменение формы ответа требует правки спеки вместе с кодом. Захватывается
только трафик, идущий через шлюз.

---

## 10. Как запустить локально

```bash
cd /path/to/AI-buddy
make up
```

Что дальше:

- сайт — http://localhost:8080
- коды входа — http://localhost:8025 (Mailpit)
- employee — http://localhost:8082, note-service — http://localhost:8079
- метрики — Prometheus http://127.0.0.1:9090, Grafana http://127.0.0.1:3000

Порядок старта не случайный: `postgres` → `auth`/`employee` → разовые задачи (`note_migrate`,
`note_import`) → `note_service` → `gateway` → `frontend`. Первый запуск занимает несколько минут
(сборка образов), стек поднимается за 30–60 секунд.

---

## 11. Грабли, о которые уже спотыкались

- **VPN и докер.** VPN перехватывает маршруты, из-за этого контейнеры перестают резолвить имена
  (внутренние и внешние): сборка падает на скачивании пакетов, nginx отдаёт `502`, а вход через
  Google не проходит. Лечится `sudo sh deploy/vpn-route.sh` после `make up` (правило живёт до
  перезагрузки) и явным DNS для докера в `/etc/docker/daemon.json`.
- **Хостнейм браузера.** `localhost` и `127.0.0.1` — разные хосты для кук. Сессия и состояние
  входа Google привязаны к хосту, поэтому открывать сайт надо тем адресом, который прописан в
  `APP_ORIGIN`/`GOOGLE_REDIRECT_URL`.
- **Пересборка после правок.** Изменения в коде не появляются от `docker compose up -d`:
  контейнер пересоздаётся из старого образа. Нужна сборка (`make up` собирает с `--build`),
  а фронт после этого просит жёсткого обновления страницы (Ctrl+Shift+R).
- **Три Go-модуля.** Корневой, `chat-service`, `note-service`. Ошибка `missing go.sum entry`
  лечится `go mod tidy` в том модуле, чей образ падает, — в остальных она не проходит.
- **Файловые логи note-service** пишутся внутри контейнера и исчезают при пересоздании; смотреть
  надо `make logs`.

---

## 12. Что ещё в планах

- подтверждение почты письмом с токеном, привязка Google-аккаунта;
- на продакшене — `COOKIE_SECURE=true` и домен в куках;
- для Google из инфраструктуры нужен белый IP или прокси (сейчас в dev-окружении помогает
  `HTTPS_PROXY` в `.env`);
- SMTP-переменные для реальной отправки писем (сейчас письмо пишется в лог, а локально видно
  в Mailpit);
- `LLM-service` — заготовка под отдельный LLM-сервис.
