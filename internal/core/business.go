package core

import "time"

// The business/ bounded context: the customer-identity domain ("customer" in AML
// vocabulary, "Business" in code — design §2). Business is the aggregate root.

// EnumBusinessStatus is the simplified starter-scope lifecycle (design §7). There is
// no screening state because KYB is deferred; a pending/screening state slots in
// between Draft and Active when KYB is added.
//
//	draft ──► active ──► suspended ──► active
//	                         └──────► offboarded
type EnumBusinessStatus string

const (
	EnumBusinessStatusDraft      EnumBusinessStatus = "Draft"
	EnumBusinessStatusActive     EnumBusinessStatus = "Active"
	EnumBusinessStatusSuspended  EnumBusinessStatus = "Suspended"
	EnumBusinessStatusOffboarded EnumBusinessStatus = "Offboarded"
)

// EnumLegalForm drives which registration fields apply at capture (design §3.3, "the
// one structural split"). It branches the registration *input*, but the result is
// still a single Business entity — never a subtype hierarchy.
type EnumLegalForm string

const (
	EnumLegalFormCompany           EnumLegalForm = "Company"
	EnumLegalFormSoleEstablishment EnumLegalForm = "SoleEstablishment"
	EnumLegalFormFreelancer        EnumLegalForm = "Freelancer"
)

// Organization is the optional/collapsible grouping layer (design §3.1): it groups
// one or more Business legal entities of the same client group. It is intentionally
// thin — it is NOT the aggregate root and NOT where money or access attaches.
type Organization struct {
	ID        string
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
}

// Business is the SME legal entity — the aggregate root of the domain (design §3.2).
// It holds legal identity and lifecycle, and holds *references out* to things it does
// not own (Keycloak org, ledger account-set, virtual IBANs). It never accumulates
// money or rail state (design §8.3, "No independent sums").
type Business struct {
	ID     string
	CorrID string // caller-supplied correlation / idempotency key (UNIQUE, standard §8)

	// legal identity
	LegalName         string // the name on the CR; the name screening keys on
	TradeName         string // the name the business trades under; the UI keys on this
	CRNumber          string // commercial registration number (a registration fact)
	LegalForm         EnumLegalForm
	IncorporationDate *time.Time // registration date; absent for some forms until captured

	// lifecycle
	Status EnumBusinessStatus

	// references out — pointers, never owned data (design §3.2). nil until provisioned.
	OrganizationID *string // parent group, if the Organization layer is in use
	KeycloakOrgID  *string // the one Keycloak Organization this Business maps to (1:1)
	LedgerRef      *string // TigerBeetle account-set reference (money lives in TigerBeetle)

	// Deferred (KYB) fields — do NOT build in v1: risk_rating, last_assessed_at,
	// next_review_due_at. See design §3.2 and core.Person.

	UpdatedAt time.Time
	CreatedAt time.Time
}

// EnumClassificationAxis is the classification *dimension*. Types multiply; labels add
// (design §3.3) — so instead of encoding every (size × tier × …) combination as a
// distinct subtype, a Business carries labeled (axis, value) records on a shared
// concept. size_segment and service_tier are in starter scope; kyb_tier is deferred.
type EnumClassificationAxis string

const (
	EnumClassificationAxisSizeSegment EnumClassificationAxis = "SizeSegment"
	EnumClassificationAxisServiceTier EnumClassificationAxis = "ServiceTier"
)

// Classification is a labeled value on a Business, using the (axis, value) pattern
// (design §3.3). The closed vocabulary (see vocab.go) is the guardrail that keeps this
// from degrading into EAV; multi-axis analytical filtering is a warehouse concern, not
// a transactional query here.
type Classification struct {
	ID         string
	BusinessID string
	Axis       EnumClassificationAxis
	Value      string
	UpdatedAt  time.Time
	CreatedAt  time.Time
}
