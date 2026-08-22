# client — Business (Customer) Domain: Logical Design Guide

> **Purpose.** This is the conceptual/logical reference for the customer-identity and
> access domain of the client EMI platform. It defines the modules, the entities inside
> them, the fields each entity is *responsible for holding*, and how the entities relate.
> It is written to be handed to a code-level spec pass — it deliberately stops short of
> schemas, types, indexes, or any implementation detail.

> **Starter scope (this document).**
> - **Included:** the legal-entity model, operator access, machine actors, and the
    >   relationships that bind them.
> - **Excluded for now:** KYC/KYB *verification and screening machinery* (Nafath/Wathq/
    >   sanctions integration, PEP checks, risk rating, periodic review, onboarding workflows)
    >   and **OPA** (fine-grained authorization / entitlement projection).
> - The two entities that KYB will later populate (`Person`, `BusinessRole`) are still
    >   described, in a clearly marked **deferred** section, because one *seam* between them
    >   and the access layer must exist from day one or adding KYB later becomes a rewrite.
    >   You can leave that section unbuilt for v1; just keep the seam.

---

## 0. Systems this domain references but does not own

This domain is a *master of record for identity and access*. Several facts it points at
are owned elsewhere. Getting this boundary right is the first design decision.

| Concern | Owner (not this domain) | This domain holds |
|---|---|---|
| Login, password, MFA, sessions | **Keycloak** | only the `sub` reference |
| Money, balances, safeguarding | **TigerBeetle** | only account-set references |
| Rail state, payment lifecycle | **Hyperswitch** | nothing (references live in the payment domain) |
| The per-payer virtual IBANs your collections product mints | **payment layer** (as ISO counterparties) | nothing — these are *not* `Business` records |
| A scheduled payment's recurrence rule / next-run / amount | **payment domain** | nothing — this domain only answers *who may initiate, and who initiated this one* |

**Rule of thumb.** If a fact changes a balance, defines a rail, or is a login credential,
it is a *reference out*, never a field this domain accumulates.

---

## 1. The load-bearing distinction

Everything below is shaped by one fact: **a human shows up in two unrelated capacities**,
and welding them into one record is the expensive mistake.

- **As someone who *controls* the business** — a UBO, signatory, or director you must
  *know* for compliance. This is a **control fact**. It exists whether or not the human
  ever logs in. *(Deferred in starter scope — this is the KYB door.)*
- **As someone who *operates* the account** — a founder, a finance clerk, an outsourced
  accountant who authenticates and acts. This is an **access fact**. It exists whether or
  not the human controls anything.

The two sets overlap but are **not equal**, and that inequality is the entire reason to
keep them separate:

- A UBO who owns 40% and never opens the app → *controls*, does not *operate*.
- A hired bookkeeper with payment rights → *operates*, does not *control*.

If "human" were one record you would be forced to either invent a fake 0%-ownership row
for the bookkeeper, or screen your silent UBO as if they were a user. Both are wrong.
**Two concepts, reconciled by one key** (the verified national ID) — never merged into one
row.

In the starter scope you build only the *operate* door. The *control* door is stubbed, but
the **reconciliation seam** (a nullable link from the operator edge to a natural person)
is present from the first migration.

---

## 2. Module map

Two bounded contexts, one deferred spine.

```
business/     ← the customer-identity context ("customer" in AML vocabulary, "Business" in code)
              Organization, Business, Classification
              └─ [deferred] Person, BusinessRole   (the compliance spine)

access/       ← sibling supporting context: who/what may operate an account
              Actor (abstraction), Membership, Invitation, ServiceAccount, User(ref)
```

**Why `business/` and not `customer/`, `party/`, `merchant/`, `client/`:**

- `Party` is retired — it collides with your ISO 20022 payment layer (`InitgPty`, `Dbtr`,
  `Cdtr`).
- `Merchant` collides with Hyperswitch's first-class `merchant` object and is semantically
  wrong (a merchant accepts cards; your product is an operating account where card
  acceptance is one surface among IBANs, collections, and wallet).
- `Customer` overloads your *own* payer concept — your collections product already mints
  per-*customer* virtual IBANs, where "customer" means the business's payers. Keeping
  `Customer` as the root would repeat the `Party` trap one level up.
- `Company` is too narrow — it excludes sole establishments (مؤسسة) and freelancers, which
  are businesses but not companies.
- `Business` is the right superset and is collision-free at the top level of ISO 20022,
  TigerBeetle, Temporal, and Kafka. Regulatory vocabulary may still say "customer"
  (CDD/KYC are terms of art) — code vocabulary and regulatory vocabulary are allowed to
  differ.

---

## 3. The `business/` module

### 3.1 `Organization` — the grouping layer *(optional / collapsible for v1)*

**Responsibility.** Groups one or more `Business` legal entities that belong to the same
client group (a holding company with multiple registered entities). It is a convenience
grouping, **not** the aggregate root and **not** where money or access attaches.

**Holds**
- `name` — display name of the group.
- (that is nearly all it holds — it is intentionally thin.)

**Deferral note.** If your starter only serves single-legal-entity clients, you may omit
`Organization` entirely and let `Business` stand alone; add it later when a client first
brings a second entity. Nothing below depends on it existing.

---

### 3.2 `Business` — the aggregate root / legal entity

**Responsibility.** The SME legal entity — "the client" on your books. This is the
aggregate root of the domain: lifecycle, classification, and every access grant hang off
it. It **holds references out** to things it does not own and never accumulates money or
rail state.

**Holds — legal identity**
- `legal_name` — the name on the CR; the name compliance and screening use.
- `trade_name` — the name the business trades under and the UI shows (SMEs often differ
  from the CR); display-only.
- `cr_number` — commercial registration number (the registration fact, not a KYB process).
- `legal_form` — company / sole establishment / freelancer, etc. Drives *which registration
  fields apply* (see 3.3, "the one structural split").
- `incorporation_date` — registration date.

**Holds — lifecycle**
- `status` — where the entity sits in its state machine (see §7). In starter scope this is
  simplified because there is no screening state to pass through.

**Holds — references out** *(pointers, never owned data)*
- `organization_id` — parent group, if `Organization` is in use.
- `keycloak_org_id` — the one Keycloak Organization this Business maps to (1:1).
- ledger account-set reference(s) — the TigerBeetle accounts for this Business.
- virtual IBAN reference(s) — minted in the payment layer.
- product-enrollment reference(s) — which client products this Business is enrolled in.

**Deferred (KYB) fields — do not build in v1**
- `risk_rating`, `last_assessed_at`, `next_review_due_at` — periodic KYB refresh state.

**Naming note.** `legal_name` vs `trade_name` is a real split, not a nicety: screening
keys on legal, the UI keys on trade. Keep both from the start.

---

### 3.3 `Classification` — labeled values, not new types

**The problem it solves.** A `Business` is classified along several independent axes at
once — size segment (micro / SME / corporate), commercial service tier (gold / silver,
owned by marketing/growth), and later others. The naive move is to encode each combination
as a distinct subtype. That produces a Cartesian product of variants that grows
*multiplicatively* with every new axis. **Types multiply; labels add.**

**The shape.** Keep a single `Business` entity and attach classifications as labeled values
on a shared classification concept using an **`(axis, value)`** pattern.

**Holds (per classification record)**
- `business_id` — the entity being classified.
- `axis` — the classification dimension, e.g. `size_segment`, `service_tier`.
- `value` — the value on that axis, e.g. `sme`, `gold`.

**Two guardrails that keep this from becoming EAV:**
1. **Closed vocabulary.** Each `axis` has a constrained set of allowed `value`s, enforced
   against a vocabulary reference — invalid combinations cannot be written.
2. **Analytics, not OLTP.** Multi-axis filtering ("all gold SMEs in Riyadh") is a
   reporting concern and lives in the data warehouse, not in transactional queries here.

**The one *structural* split that is justified.** Registration *form* — a freelancer and a
CR-holder genuinely require different fields at capture. So `legal_form` may branch the
*registration input*, but it is still **one `Business` entity** afterward; the difference
is in what's collected, not in a subtype hierarchy.

**Promotion rule.** An axis earns its own dedicated record (with its own fields and
lifecycle) *only when it accumulates correlated fields that live and change together* — not
preemptively. Until then it stays a label.

> **Scope note.** A `kyb_tier` axis belongs to the compliance world and is out of the
> starter scope. `size_segment` and `service_tier` are operational/commercial and are in
> scope.

---

## 4. The `access/` module

This module answers exactly one question: **who, or what, may operate a given Business —
and who initiated a given action?** It is where the operator door lives.

### 4.1 `Actor` — the initiation abstraction *(the single most important idea here)*

**Responsibility.** `Actor` is the abstract supertype for **anything the system
authenticates, authorizes, and attributes an action to**. It has two concrete forms:

- a **human operator** → a `Membership` (backed by a Keycloak `User`), and
- a **machine** → a `ServiceAccount` (backed by a credential).

**Why it exists.** Every downstream consumer — the payment hub, the Temporal workflow, the
TigerBeetle transfer, the audit log — should reference **`actor_id`**, never a `staff_id`
or a `user_id`. A scheduled payment fired by automation is then *"initiated by actor X"* in
exactly the same shape as a human-initiated one. Nothing downstream branches on
human-vs-machine, so when the machine surface expands (your BYOP roadmap), it slots in
**additively** with no rewrite of the payment path.

**Holds.** `Actor` is thin — often just an identity and a discriminator of which concrete
form backs it. The substance lives in `Membership` / `ServiceAccount`.

**Naming note (why not `Principal`, `Agent`, `Party`).**
- `Principal` collides with the **principal amount** in your ledger — a `principal_id`
  column next to a `principal` amount welds ambiguity into the schema. Rejected.
- `Agent` is taken — in ISO 20022 the *agents* are the banks (`DebtorAgent`,
  `CreditorAgent`). Rejected.
- `Party` already retired.
- **`Actor`** says what the thing *does* — it acts — which is the one property a human
  operator and a machine genuinely share.

---

### 4.2 `Membership` — the human operator grant

**Responsibility.** The join that says *this human may operate this Business, in this
capacity.* It is a many-to-many edge between a Keycloak `User` and a `Business`, and it is
the human concrete form of `Actor`.

**Holds**
- `business_id` — the account being operated.
- `keycloak_sub` — the login identity (owned by Keycloak; this is a reference).
- `person_id` — **nullable**; set only when the operator is *also* a natural person you
  hold for compliance. **This is the reconciliation seam** — keep it from day one even
  while `Person` is unbuilt. The founder's membership points to their `Person`; the
  outsourced accountant's points to null.
- `role` — the *coarse* role: `owner | admin | finance_operator | viewer`, with
  **maker / checker** treated as first-class for payment flows.
- `status` — `invited → active → suspended → revoked` (see §7).

**Boundary — what does *not* live here.** Fine-grained entitlements (per-product
visibility, exact maker/checker wiring) are **not** columns on `Membership`. In the full
system they are authored as authorization relations and projected to OPA; `Membership`
carries only the coarse role and is the origin record that projection reads. **In starter
scope, OPA is excluded** — so `Membership.role` is the whole access story for now, and the
fine-grained layer is a later addition that reads this same record. Model `role` so it can
stay coarse without repainting later.

---

### 4.3 `User` — the external identity *(reference only)*

**Responsibility.** Named explicitly so the boundary is unambiguous: the `User` is **not
yours**. Username, password, MFA, and sessions all live in Keycloak. This domain only ever
holds the `sub` (via `Membership`). It appears in this document as a box on the *far side*
of the boundary, not as a table you own.

**Holds (in this domain):** nothing beyond the `sub` reference carried by `Membership`.

---

### 4.4 `Invitation` — the state *before* a `User` exists

**Responsibility.** Models the gap between "an admin invited a colleague by email" and
"that colleague has a Keycloak account." At invite time there is no `sub` yet, so a
`Membership` cannot exist — `Invitation` is that pre-user state. It is the concrete
mechanic of *how a human becomes a user*.

**Holds**
- `business_id` — the account they're being invited to.
- `invited_email` — where the invite is sent.
- `intended_role` — the role the resulting `Membership` will carry.
- `token` — the single-use acceptance secret.
- `status` — `pending → accepted | expired | revoked`.
- `invited_by` — the actor who issued the invite.
- `expires_at` — invite expiry.

**Interaction.** On acceptance → a Keycloak user is created → a `Membership` is
materialized (carrying the `sub` and the `intended_role`) → the `Invitation` moves to
`accepted`. `Invitation` is transient; `Membership` is the durable record.

---

### 4.5 `ServiceAccount` — the machine operator

**Responsibility.** The machine concrete form of `Actor`: a non-human that acts on a
Business (e.g. runs its scheduled payments). Backed by a **credential** (API key / OAuth
client), not by a `Person`.

**Holds**
- `business_id` — the account it acts on (see the two topologies below).
- `label` / `name` — human-readable identifier for the credential.
- `status` — active / suspended / revoked.
- credential reference — a pointer to the secret material (the secret itself is not stored
  as domain data).

**Two topologies (build the first now):**
1. **A business's own automation.** The Business generates a credential to run *its own*
   scheduled payments. The service account is scoped to exactly one Business — a machine
   "operator" of that Business, nothing more. **This is where you start.**
2. **A third-party platform acting for many businesses.** *(Deferred — your BYOP surface.)*
   An ERP or marketplace integrates once and acts for every client business that authorizes
   it. Here the external *application* is one deduplicated thing (like `Person`) and each
   Business grants it access separately (like `Membership`) — the GitHub App model: one
   app, many installs, a scoped token per install. You do **not** build this now; you only
   ensure the initiation seam references `actor_id` so it slots in additively.

**Boundary.** The *scheduled-payment instruction itself* — recurrence rule, next-run time,
amount — lives in the **payment domain**, not here. This domain answers only "who may
initiate, and who initiated this one." The `ServiceAccount` is that *who*; the schedule is
a separate concern that merely names an actor as its initiator.

---

## 5. Deferred — the compliance identity spine

> **Not part of the starter build.** This section exists so the seam in §4.2
> (`Membership.person_id`) has a defined other end, and so adding KYB later is *additive*.
> You may skip building these entities for v1. Do **not** skip the nullable `person_id`.

### 5.1 `Person` — the natural person

**Responsibility.** Holds a human's identity **once**, decoupled from any single Business,
so the same human can control one Business and operate another without duplication. In the
full system this is the KYC subject; the screening result lives here, run once, and every
relationship for that human inherits it.

**Would hold** *(when built)*
- `national_id` — the identity anchor (Nafath in the full system).
- `full_name`, `nationality`, `date_of_birth`.
- *(deferred KYC machinery: `is_pep`, `screening_status`, verification provenance.)*

**Why it's a hub.** One human → one screening record → every relationship linked by
construction rather than by hoping national IDs match. It is also the cleaner answer under
PDPL: a data-subject erasure or rectification touches one `Person`, not scattered copies.

**Naming note.** `Person` / `Human` / *natural person* all work — your `Business` is the
*legal person*, the human is the *natural* one. Standard lean is `Person`.

### 5.2 `BusinessRole` — the control edge

**Responsibility.** The typed edge from a `Person` to a `Business` that records *how they
control it*. This is what a single "related person" collapses into once identity is pulled
out into `Person`.

**Would hold** *(when built)*
- `person_id`, `business_id` — the two ends.
- `role` — one of `ubo | signatory | director | legal_representative`.
- role-specific attributes — e.g. `ownership_pct` + `control_basis` for a UBO; signing
  authority (sole/joint) + evidence reference for a signatory. **Ownership % stays on this
  edge; screening stays on the `Person`.**
- verification state — `verified_at`, `verification_source` *(deferred machinery)*.

**Key shape.** One `Person` may hold **several** `BusinessRole`s to the **same** `Business`
— a founder is `ubo` + `signatory` + `director`, i.e. three edges — and each edge
re-verifies on its own trigger (ownership on share transfer; signing authority on board
resolution).

### 5.3 The reconciliation seam

The operate-door and the control-door meet in exactly one place: `Membership.person_id`.

- When an operator is also a controlling person, their `Membership` points at their
  `Person`, and from that `Person` you can read their `BusinessRole` edges.
- The match is by **verified national ID**, never by a shared row.
- Build the nullable column now; populate it only once `Person` exists.

---

## 6. How the pieces interact

### 6.1 Relationships & cardinalities

| From | To | Cardinality | Meaning |
|---|---|---|---|
| `Organization` | `Business` | 1 → * | a group holds many legal entities *(optional layer)* |
| `Business` | `Membership` | 1 → * | an account has many operators |
| `Business` | `Invitation` | 1 → * | an account has many pending invites |
| `Business` | `ServiceAccount` | 1 → * | an account has many machine operators |
| `Business` | Classification | 1 → * | an account carries many labels across axes |
| `Membership` | `User` | * → 1 | every operator maps to exactly one Keycloak user |
| `Membership` | `Person` | * → 0..1 | an operator *may* also be a controlling person *(nullable seam)* |
| `Actor` | `Membership` / `ServiceAccount` | 1 → 1 | the abstract initiator resolves to one concrete form |
| **[deferred]** `Business` | `BusinessRole` | 1 → * | an account has many control edges |
| **[deferred]** `Person` | `BusinessRole` | 1 → * | a human controls many businesses |
| **[deferred]** `Person` | `Membership` | 1 → 0..1 *(per business)* | a human operates a given business via at most one membership |

### 6.2 The four canonical human scenarios *(the test the model must pass)*

| # | Real-world human | Control edge *(deferred)* | Operator edge | `Membership.person_id` |
|---|---|---|---|---|
| 1 | **Founder** — owns, signs, operates | `BusinessRole[ubo, signatory]` | `Membership[owner]` → User | set |
| 2 | **Silent UBO** — owns, never logs in | `BusinessRole[ubo]` | *(none)* | — |
| 3 | **Outsourced accountant** — operates, owns nothing | *(none)* | `Membership[finance_operator]` → User | null |
| 4 | **Approving signatory** — signs, doesn't own | `BusinessRole[signatory]` | `Membership[checker]` → User | set |

In the starter build, only the "operator edge" column is materialized; scenarios 1–4
still resolve correctly on the access side, and the `person_id` column reserves the link.

### 6.3 The initiation path

```
        (human)                         (machine)
      Membership  ─┐                 ServiceAccount
                   ├──►  Actor  ──►  actor_id  ──►  payment hub / Temporal / TigerBeetle / audit log
   User (Keycloak)─┘        ▲
                            └─ downstream never branches on human-vs-machine
```

Everything that records "who did this" records an `actor_id`. That is the whole payoff of
the `Actor` abstraction.

---

## 7. Lifecycles (state machines)

State transitions are logical here — the code-spec pass defines guards and events.

**`Business.status`** *(simplified — no screening states in starter scope)*
```
draft ──► active ──► suspended ──► active
                          └──────► offboarded
```
*(When KYB is added, a screening/pending state sits between `draft` and `active`.)*

**`Membership.status`**
```
invited ──► active ──► suspended ──► active
                            └──────► revoked
```

**`Invitation.status`**
```
pending ──► accepted
   ├──────► expired
   └──────► revoked
```

---

## 8. Design rules & boundaries

### 8.1 The rule that sorts everything

When something new appears and you're unsure how to model it, apply this — **nothing
defaults to "new type":**

| Situation | Model it as |
|---|---|
| Different capture journey **/ compliance treatment** | a **type** *(compliance dimension deferred)* |
| A value some policy reads | an **attribute / label** (a Classification value) |
| Different backing — human vs machine | a **sibling entity** under `Actor` |
| Controls but does not act | a **separate relationship** (the `BusinessRole`) *(deferred)* |

### 8.2 What lives elsewhere (recap — do not let these bleed in)

- Money and balances → **TigerBeetle**.
- Rail / payment lifecycle → **Hyperswitch**.
- Login and credentials → **Keycloak** (outside this domain).
- Per-payer virtual IBANs (collections) → **payment layer** as counterparties, *not*
  `Business` records.
- A schedule's recurrence/next-run/amount → **payment domain**; this domain only names the
  initiating `Actor`.

### 8.3 No independent sums

This domain is a master of *identity and access*, not of value. It must never accumulate an
independent financial total — any service producing money totals outside TigerBeetle is a
shadow ledger and is out of bounds by construction. (Named here only so the line stays
bright as the domain grows.)

---

## 9. Deliberately out of scope for the starter guide

Kept out on purpose; each has a defined slot for when you add it:

- **KYC/KYB verification & screening** — populates `Person` and `BusinessRole`; adds a
  screening state to `Business.status`; introduces provider integrations and periodic
  review. Seam already reserved (`Membership.person_id`).
- **OPA / fine-grained authorization** — reads `Membership.role` as its origin record and
  projects entitlements (maker/checker wiring, per-product visibility). Keep `role` coarse
  so this is additive.
- **Open role catalog** — for now `Membership.role` is a small closed set; a governed,
  extensible catalog comes later (with vocabulary constraints, breaking-change gating on
  the API enum, and a compliance-owned merge gate).
- **BYOP multi-tenant service accounts** — topology 2 in §4.5 (the GitHub App model).
  Reserved by having initiation reference `actor_id`.

---

### One-line summary of the whole shape

> `Organization` groups `Business`; a `Business` is operated by `Actor`s — humans via
> `Membership` (backed by a Keycloak `User`, born from an `Invitation`) or machines via
> `ServiceAccount` — and, once KYB is added, is *controlled* by `Person`s through
> `BusinessRole` edges, reconciled to operators through the nullable `Membership.person_id`
> seam that you build from day one.