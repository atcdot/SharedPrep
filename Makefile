.PHONY: build run run-bin test proto web/install web/dev web/build docker/up docker/down

# --- Backend ---

build:
	go build -o bin/api ./cmd/api/

run:
	go run ./cmd/api/

run-bin: build
	./bin/api

test:
	go test ./... -count=1

# --- Proto ---

proto:
	buf generate

# --- Frontend ---

web/install:
	cd web && pnpm install

web/dev:
	cd web && pnpm dev

web/build:
	cd web && pnpm build

# --- Docker ---

docker/up:
	docker compose -f docker/docker-compose.yml up --build -d

docker/down:
	docker compose -f docker/docker-compose.yml down

docker/logs:
	docker compose -f docker/docker-compose.yml logs -f api

# --- Full stack (local) ---

dev:
	@make run & PID=$$!; cd web && pnpm dev; kill $$PID 2>/dev/null
