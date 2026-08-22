# justfile — migration + codegen recipes (service standard §9)

dir := "./internal/adapters/pg"
driver := "postgres"
dsn := env_var_or_default("DB_DSN", "postgres://maaladmin:maalpassword@localhost:54322/maalbizdb?sslmode=disable")

# ──────────────────────────────────────────────
# Migrations (goose) — write idempotently, with matching down steps
# ──────────────────────────────────────────────
mg-new name:
    goose -dir={{dir}}/migrations create {{name}} sql
mg-up:
    goose -dir={{dir}}/migrations {{driver}} "{{dsn}}" up
mg-down:
    goose -dir={{dir}}/migrations {{driver}} "{{dsn}}" down
mg-status:
    goose -dir={{dir}}/migrations {{driver}} "{{dsn}}" status

# ──────────────────────────────────────────────
# Codegen (go-jet) — regenerate model/table from the live schema after every migration.
# Generated code is committed and never hand-edited.
# ──────────────────────────────────────────────
mg-gen:
    jet -dsn={{dsn}} -schema=public -path={{dir}}/.gen

# ──────────────────────────────────────────────
# App
# ──────────────────────────────────────────────
run-worker:
    go run ./cmd/worker
run-starter:
    go run ./cmd/starter
run-local:
    go run ./cmd/local

build:
    go build ./...
tidy:
    go mod tidy
fmt:
    gofmt -w .
vet:
    go vet ./...
