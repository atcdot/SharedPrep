# SharedPrep — Task Breakdown

## 1. Project Bootstrap

- [x] **1.1** Структура директорий (api/, cmd/api/, internal/, pkg/, migrations/, docker/)
- [x] **1.2** Docker Compose: PostgreSQL + service
- [x] **1.3** go.mod зависимости: connect-go, pgx, goose
- [x] **1.4** cmd/api/main.go: graceful shutdown, конфиг через env

## 2. Proto Contracts

- [x] **2.1** api/event/v1/event.proto: CRUD + shareable link
- [x] **2.2** api/participant/v1/participant.proto: join by link, list
- [x] **2.3** api/item/v1/item.proto: CRUD, assignment
- [x] **2.4** buf.yaml + buf.gen.yaml, code generation

## 3. Database

- [x] **3.1** Миграция: events, participants, items
- [x] **3.2** internal/storage/postgres.go: pgxpool, goose migrations

## 4. Business Logic

- [x] **4.1** internal/service/event.go: CreateEvent, GetEvent, UpdateEvent, DeleteEvent
- [x] **4.2** internal/service/participant.go: JoinEvent, ListParticipants, ResolveByToken
- [x] **4.3** internal/service/item.go: CreateItem, UpdateItem, DeleteItem, ClaimItem, UnclaimItem, ListItems

## 5. API Layer

- [x] **5.1** internal/api/handler_event.go
- [x] **5.2** internal/api/handler_participant.go
- [x] **5.3** internal/api/handler_item.go
- [x] **5.4** internal/api/middleware.go: auth middleware (cookie token)
- [x] **5.5** cmd/api/main.go: wiring handlers + middleware

## 6. Testing & Polish

- [x] **6.1** Integration тесты (18 тестов, service слой)
- [x] **6.2** Handler-level тесты (18 тестов, Connect-RPC)
- [x] **6.3** Примеры API в docs/api-examples.md

## 7. Docker

- [x] **7.1** docker/Dockerfile: multi-stage build
- [x] **7.2** docker/docker-compose.yml: db + api service
