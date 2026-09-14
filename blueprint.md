# Go Service Standard
This document governs the service skeleton and layers.

---

## 1. Philosophy

- **Vertical slices.** A capability owns its whole stack — orchestration, use-case steps, persistence, and outbound ports — as a pkg under `internal/<feature>/`.
- **Shared Utilities.** Adapters are shared and feature-agnostic.
- **Temporal-first orchestration.** The durable Temporal workflow is the spine of every capability. Activities are the *only* place side effects happen.
- **Thin, faithful adapters.** Outbound integrations are typed mirrors of their API adapters or DB.
- **Serializable at the edges, domain types inside.** What crosses the workflow boundary is a flat, serializable DTO; domain types live within the slice.

## 2. Package layout

```
cmd/
  worker/          ← Temporal worker: registers workflows + activities on a task queue
  server/          ← starts HTTP server
config/
  config.go        ← cleanenv struct + Load()
  <env>.yml
internal/
  core/            ← entities, value objects, enums — no I/O, no framework imports
  <feature>/       ← one vertical slice per capability
    workflows/     ← Temporal workflow + pure-function twin + boundary DTOs
    activities/    ← use-case steps: map → port → persist → classify errors
    store/         ← persistence port (typed SQL): read.go, write.go, mapper.go
    <port>/        ← anti-corruption wrapper around one outbound adapter
  adapters/
    pg/            ← Postgres: connect, generic pagination, mappers, generated model+table (.gen)
    temporal/      ← Dial()
    <vendor>/      ← typed HTTP adapter (see api-adapter-standard.md)
justfile           ← migration + codegen recipes
```

- **`cmd/*` are thin.** Each `main` calls `config.Load()` first, dials infrastructure through an adapter, wires, and runs. No business logic in `cmd/`.
- **One directory per capability** under `internal/`. Adding a capability never edits another capability's package.
- **The parent `internal/adapters/<vendor>` package holds transport only.** Endpoint suites live in its sub-packages (adapter standard §2).

## 3. Entrypoints (`cmd/`)

Every `main` opens with configuration, then infrastructure:

```go
func main() {
	config.Load()

	c, err := temporal.Dial()
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	defer c.Close()
	// ... register or start ...
}
```

Entrypoints:
- **`worker`** — constructs a `worker.New(c, TaskQueue, …)`, registers the workflow and every activity, and blocks on `w.Run(worker.InterruptCh())`. This is the runtime that executes the durable flow.

## 4. Configuration (`config/`)

One `cleanenv` struct, one global, loaded once. Precedence is **env var → YAML(`<env>.yml`) → struct default**; `ENV` selects the YAML file.

```go
var CNF = &config{}

type config struct {
	Env string `env:"ENV" yaml:"Env" env-default:"dev"`
	Db  struct {
		Host string `env:"DB__HOST" yaml:"Host" env-default:"localhost"`
		Port int    `env:"DB__PORT" yaml:"Port" env-default:"5432"`
		// ...
	} `yaml:"Db"`
}

func Load() {
	cleanenv.ReadEnv(CNF)
	err := cleanenv.ReadConfig("config/"+CNF.Env+".yml", CNF)
	// tolerate an empty/missing env file; panic on a real parse error
}
```

- **Always give a dev-oriented default** so the service runs with zero configuration locally.
- **Prefer YAML** for non-secret values; keep secrets in env vars only.

## 5. Core domain (`internal/core/`)

Pure types only — no `context`, no framework, no adapter imports. Holds entities, value objects, enums, and ID generation for the domain.

### 5.1 Entities and value objects

```go
type Order struct {
	ID        string
	CorrID    string        // correlation / idempotency key from the caller
	Amt       Amount
	Status    EnumOrderStatus
	SettledAt *time.Time    // pointer: absent until the step that sets it runs
	UpdatedAt time.Time
	CreatedAt time.Time
}
```


### 5.2 Enums

The constant name is the full type name — `Enum` prefix included — plus the PascalCased value. Domain enum *values* are your own vocabulary (not a wire format), so choose them cleanly:

```go
type EnumOrderStatus string

const (
	EnumOrderStatusPlaced   EnumOrderStatus = "Placed"
	EnumOrderStatusReserved EnumOrderStatus = "Reserved"
	EnumOrderStatusCharged  EnumOrderStatus = "Charged"
	EnumOrderStatusSettled  EnumOrderStatus = "Settled"
)
```

## 6. Feature slice (`internal/<feature>/`)

### 6.1 Orchestration (`workflows/`)

The workflow is a **durable, ordered sequence of activities**. It defines the task-queue constant and the boundary DTOs, and chains activities with a shared retry policy. Domain entities may pass *between* activities (Temporal serializes them); the workflow's own input/output DTO stays flat and serializable.

```go
const TaskQueueOrder = "order"

type OrderInput struct {
	CorrID string
	AmtGross, AmtNet, AmtFee string // flat & serializable at the boundary
}
type OrderOutput struct{ Order core.Order }

func Order(ctx workflow.Context, in OrderInput) (OrderOutput, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})

	var order core.Order
	if err := workflow.ExecuteActivity(ctx, activities.Reserve,
		activities.ReserveInput{CorrID: in.CorrID /* … */}).Get(ctx, &order); err != nil {
		return OrderOutput{}, fmt.Errorf("activities.Reserve: %w", err)
	}
	if err := workflow.ExecuteActivity(ctx, activities.Charge,
		activities.ChargeInput{Order: order}).Get(ctx, &order); err != nil {
		return OrderOutput{}, fmt.Errorf("activities.Charge: %w", err)
	}
	return OrderOutput{Order: order}, nil
}
```

### 6.2 Activities (`activities/`)

One activity per step, each a free function with a fixed skeleton: **pure input→entity mapping → one port call → persist → return the updated entity.** Side effects live only here.

```go
type ReserveInput struct {
	CorrID                  string
	AmtGross, AmtNet, AmtFee string
}

func Reserve(ctx context.Context, in ReserveInput) (core.Order, error) {
	order, err := reserveInputToEntity(in) // pure; builds the domain entity
	if err != nil {
		return core.Order{}, fmt.Errorf("reserveInputToEntity: %w", err)
	}
	order, err = reservePersist(ctx, order) // one port call + error classification
	if err != nil {
		return core.Order{}, fmt.Errorf("reservePersist: %w", err)
	}
	return order, nil
}
```

**Infrastructure errors are classified at this boundary**, so ports below stay dumb and callers above stay domain-only. Match the driver's typed errors:

```go
func reservePersist(ctx context.Context, order core.Order) (core.Order, error) {
	order, err := store.Insert(ctx, order)
	if err == nil {
		return order, nil
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
		return reconcileOnConflict(ctx, order, pgErr) // idempotent replay (§8)
	}
	return order, fmt.Errorf("store.Insert: %w", err)
}
```

A guarded status update that matched no row surfaces as `qrm.ErrNoRows` — classify it explicitly rather than treating it as a generic failure.

### 6.3 Persistence port (`store/`)

Typed SQL over the generated `model`/`table` packages, split into `read.go`, `write.go`, `mapper.go`. One shared connection pool per process — never open and close a connection per call.

**Writes** go through a single guarded status-transition helper. The `WHERE status <> target` guard makes each transition idempotent; `RETURNING` re-reads the row so the caller gets canonical state:

```go
func updateStatus(ctx context.Context, o core.Order, to core.EnumOrderStatus,
	extras ...postgres.ColumnAssigment) (core.Order, error) {

	sets := append([]postgres.ColumnAssigment{
		table.Orders.UpdatedAt.SET(postgres.TimestampzT(time.Now().UTC())),
	}, extras...)

	stmt := table.Orders.UPDATE().
		SET(table.Orders.Status.SET(postgres.String(string(to))), sets...).
		WHERE(table.Orders.ID.EQ(postgres.String(o.ID)).
			AND(table.Orders.Status.NOT_EQ(postgres.String(string(to))))).
		RETURNING(table.Orders.AllColumns)

	var m model.Orders
	if err := stmt.QueryContext(ctx, pg.DB(), &m); err != nil {
		return o, fmt.Errorf("QueryContext: %w", err)
	}
	return OrderFromModel(m), nil
}

func Settle(ctx context.Context, o core.Order) (core.Order, error) {
	return updateStatus(ctx, o, core.EnumOrderStatusSettled,
		table.Orders.SettledAt.SET(postgres.TimestampzT(*o.SettledAt)))
}
```

**Reads** that paginate build the base statement **once** and derive the count from it — no second hand-built query, no throwaway scan struct. Scan the scalar count straight into the destination field, then chain the window onto the same statement:

```go
stmt := postgres.SELECT(table.Orders.AllColumns).FROM(table.Orders)

query, args := postgres.SELECT(postgres.COUNT(postgres.STAR)).
	FROM(stmt.AsTable(table.Orders.TableName())).Sql()
if err := pg.DB().QueryRowContext(ctx, query, args...).Scan(&page.Total); err != nil {
	return page, fmt.Errorf("count: %w", err)
}

stmt = stmt.ORDER_BY(table.Orders.CreatedAt.DESC()).LIMIT(limit).OFFSET(offset)
```

Count before chaining the window on — the statement mutates in place. Return rows in a generic `pg.Pagination[core.Order]` (§7).

**`mapper.go`** is the only place decimals appear — it converts between the domain's integer minor units and the DB `decimal` column, and nowhere else:

```go
m.AmountGross = decimal.New(int64(o.Amt.Gross), -2)         // to model
o.Amt.Gross  = int(m.AmountGross.Shift(2).IntPart())        // from model
```

### 6.4 Anti-corruption port (`<port>/`)

An outbound integration is wrapped in a small package that speaks domain in and domain out, keeping vendor types out of `core` and the activities. It builds the request from the entity, calls the adapter (with an idempotency key, §8), and maps the vendor's enums back to domain enums.

```go
func Charge(ctx context.Context, o core.Order) (core.Order, error) {
	c, err := acme.ClientNew(acme.Config{BaseURL: cfg.URL, APIKey: cfg.Key})
	if err != nil {
		return o, fmt.Errorf("acme.ClientNew: %w", err)
	}
	res, err := acme_payments.ChargesPost(ctx, c, toAcmeCharge(o), acme.WithIdempotencyKey(o.CorrID))
	if err != nil {
		return o, fmt.Errorf("acme.ChargesPost: %w", err)
	}
	if res.ErrorApi != nil {
		return o, fmt.Errorf("acme charge: %s: %s", res.Code, res.Message)
	}
	o.Status = fromAcmeStatus(res.Status)
	return o, nil
}
```

Pointer-typed request fields on the adapter model are set with the builtin `new(expr)` — `new(o.ID)`, `new(int64(o.Amt.Gross))`, `new(acme.EnumCurrencyUSD)` — not a `Ptr` helper. The `toAcmeCharge` / `fromAcmeStatus` mapping functions are the translation boundary; a vendor status enum is switched exhaustively onto the domain enum with an explicit `Unknown` default.

## 7. Outbound & infrastructure adapters (`internal/adapters/`)

- **Third-party HTTP APIs** are typed mirrors built per [`api-adapter-standard.md`](./api-adapter-standard.md): a parent package with `Client` / `ClientNew` / generic `Call[Out]` / `ErrorApi`, and one sub-package per docs suite. The service standard adds nothing to that; the anti-corruption port (§6.4) is the only code that imports these.
- **Database (`pg/`)** exposes a minimal surface: pool access, and a generic pager reused by every store.

  ```go
  type Pagination[T any] struct {
  	Items  []T
  	Total  int64
  	Limit  int
  	Offset int
  }
  ```

  Access goes through `database/sql` with the `pgx` stdlib driver; queries and the `model`/`table` code are generated by the SQL builder into `.gen/` (never hand-edited). `decimal` is the column type for money.
- **Orchestration (`temporal/`)** exposes `Dial()` — connect plus a health check — shared by the `worker` and `starter` entrypoints.

Each adapter is named for what it connects to and owns that concern completely; a capability reaches the outside world only through an adapter.

## 8. Cross-cutting conventions

- **Error wrapping.** Every returned error is wrapped with `%w` and a context prefix naming the function or operation (`fmt.Errorf("store.Insert: %w", err)`); outbound adapters additionally prefix the vendor name. Infrastructure errors are classified where the port is called (§6.2), not deep in the store.
- **Time.** All timestamps are UTC (`time.Now().UTC()`); persisted timestamp columns are `timestamptz` constrained to zero offset.

## 9. Data & tooling

- **Migrations** are versioned SQL applied by a migration tool, driven from the `justfile` (`mg-new`, `mg-up`, `mg-down`). Write them idempotently (`ADD COLUMN IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`) with matching down steps.
- **Codegen** regenerates the `model`/`table` packages from the live schema after every migration (`mg-gen`) into `internal/adapters/pg/.gen`. Generated code is committed and never edited.
