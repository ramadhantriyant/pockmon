.PHONY: run build generate migrate migrate-status migrate-down lint tidy clean

# ── Dev ──────────────────────────────────────────────────────────────────────

run:
	go run ./cmd/api/

build:
	go build -o bin/pockmon ./cmd/api/

# ── Database ─────────────────────────────────────────────────────────────────

migrate:
	goose -dir sql/schema postgres "$(DB_URL)" up

migrate-status:
	goose -dir sql/schema postgres "$(DB_URL)" status

migrate-down:
	goose -dir sql/schema postgres "$(DB_URL)" down

# ── Code generation ──────────────────────────────────────────────────────────

generate:
	sqlc generate

# ── Quality ──────────────────────────────────────────────────────────────────

lint:
	go vet ./...

tidy:
	go mod tidy

# ── Misc ─────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/
