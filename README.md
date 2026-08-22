# client — Business (Customer) Identity & Access Service

A Go service implementing the **client Business (Customer) domain** (`service.md`) built to the
**Go Service Standard** (`blueprint.md`). It is the master of record for *identity and access* —
the customer legal entity and who/what may operate it — and deliberately holds **no money**
(balances live in TigerBeetle; login lives in Keycloak; rail state lives in Hyperswitch).

```
module github.com/BinMunawir/client   ·   Go 1.26   ·   Temporal · Postgres (jet) · Keycloak
```

---

## What it models (starter scope)

Two bounded contexts (`service.md` §2), plus the deferred compliance spine kept only as a seam:

| Context | Entities | Where in code |
|---|---|---|
| **business/** | `Organization`, `Business` (aggregate root), `Classification` | `internal/core/business.go`, slice `internal/client/` |
| **access/** | `Actor` (abstraction), `Membership`, `Invitation`, `ServiceAccount`, `User`(ref) | `internal/core/access.go`, slice `internal/access/` |
| **deferred** | `Person`, `BusinessRole` | `internal/core/person.go` (seam only: `Membership.PersonID`) |

The **operator door** is built; the **control door** (KYB) is stubbed. The one thing kept from
day one is the reconciliation seam — the nullable `memberships.person_id` column — so adding KYB
later is additive, not a rewrite (`service.md` §4.2, §5.3). The four canonical human scenarios
(`service.md` §6.2) all resolve on the access side today.

### Capabilities (Temporal workflows)

- **`business` queue** — `Onboard`: `Register` (draft + classifications) → `ProvisionOrg`
  (Keycloak org, 1:1) → `Activate` (draft → active).
- **`access` queue** — `Invite` (issue invitation) · `AcceptInvitation` (`Validate` →
  `ProvisionUser` (Keycloak) → `MaterializeMembership` (Actor + Membership) → `Accept`) ·
  `ProvisionServiceAccount` (machine operator, topology 1).

Every "who did this" is an **`actor_id`** — a human `Membership` and a machine `ServiceAccount`
resolve to the same abstraction, so downstream never branches on human-vs-machine (`service.md` §4.1).

---

## Layout (blueprint.md §2)

```
cmd/{worker,starter,local}      thin entrypoints: config → dial → wire → run
config/                         cleanenv struct + <env>.yml; one global CNF
internal/
  core/                         entities, enums, IDs, closed vocab — stdlib only, no I/O
  business/                     vertical slice: workflows/ activities/ store/ idp(port)
  access/                       vertical slice: workflows/ activities/ store/ idp(port)
  adapters/
    pg/                         shared pool (pg.DB()), Pagination[T], migrations/, .gen/
    temporal/                   Dial()
    keycloak/                   typed HTTP adapter (api-adapter-standard.md) + keycloak_admin suite
justfile  docker-compose.yaml
```

Hexagonal: `core` imports nothing outward; adapters are shared and feature-agnostic; slices own
their whole stack. What crosses the workflow boundary is a flat serializable DTO; domain types
live inside the slice (blueprint.md §1).

---

## Run it

**Prereqs:** Go 1.26+, Docker, [`goose`], [`jet`], [`just`], and the Temporal CLI (for the durable path).

```sh
docker compose up -d            # Postgres on :54322 (+ CloudBeaver on :8978)
just mg-up                      # apply migrations
# just mg-gen                   # (optional) regenerate internal/adapters/pg/.gen from the live schema
```

**Local end-to-end (no Temporal, no Keycloak) — the fastest way to see it work:**

```sh
just run-local                  # onboard → invite → accept → provision service account, on Postgres alone
```

`cmd/local` drives the **pure-function twins** directly (blueprint.md §6.1). Keycloak isn't
required: when `KEYCLOAK__ADMIN_TOKEN` is unset the idp ports return deterministic dev-stub
references (`kc-org-dev-…`, `kc-sub-dev-…`) so the whole domain runs against just Postgres.

**Durable path (Temporal):**

```sh
temporal server start-dev       # Temporal dev server on :7233 (+ UI :8233)
just run-worker                 # hosts one worker per task queue (business + access)
just run-starter                # starts an Onboard run; workflow ID = CorrID (idempotency key)
```

**Hit real Keycloak** (optional): set `KEYCLOAK__BASE_URL`, `KEYCLOAK__REALM`, and
`KEYCLOAK__ADMIN_TOKEN` (a bearer for the Admin API). The `keycloak` adapter then creates the
Organization/User for real; the ports treat an "already exists" conflict as idempotent.

---

## Deliberate decisions & deviations

- **No money, by construction** (`service.md` §8.3). Unlike the reference service, there is no
  `Amount` value object and no `decimal` at the DB boundary — this domain never accumulates a
  financial total. It holds only *references out* (`keycloak_org_id`, `ledger_ref`, …).
- **Shared connection pool via `pg.DB()`** (blueprint.md §6.3: *"one shared pool per process —
  never open and close a connection per call"*). Implemented as a lazily-opened singleton; stores
  call `pg.DB()` inline in `QueryContext`.
- **Idempotency** (blueprint.md §8): each entity carries a caller `correlation_id` under a `UNIQUE`
  constraint; a create that violates it is reconciled by re-fetch. Status transitions are guarded
  (`WHERE status <> target … RETURNING`); a guarded no-match surfaces as `qrm.ErrNoRows` and is
  classified explicitly — for a replayed transition it re-reads and returns the row as a no-op.
- **Generated `.gen` is a committed snapshot.** It was produced in jet's exact output shape and is
  compiled via explicit import (Go ignores `.gen/` for `./...` wildcards but resolves explicit
  imports). Run `just mg-gen` against a live DB to regenerate it after any migration.
- **Keycloak is outside this domain** (`service.md` §0, §4.3). The service holds only the `sub`;
  the `keycloak` adapter is a thin, faithful mirror of just the Admin endpoints the flows need.

---

## Adding a capability (blueprint.md §10)

Copy a slice under `internal/`, then in order: define the entity/enums in `core`; write the
migration and `just mg-gen`; implement `store` (mapper, insert, guarded transitions, reads); write
each activity (map → port → persist → classify); wrap any outbound integration in an
anti-corruption port; compose the Temporal workflow and its twin; register both in `cmd/worker`.
Shared adapters are reused, not modified.

[`goose`]: https://github.com/pressly/goose
[`jet`]: https://github.com/go-jet/jet
[`just`]: https://github.com/casey/just
