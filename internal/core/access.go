package core

import "time"

// The access/ bounded context: who, or what, may operate a given Business — and who
// initiated a given action (design §4). This is the "operator door". The "control
// door" (Person/BusinessRole) is deferred; only the nullable seam is kept (see
// Membership.PersonID and core.Person).

// EnumActorType discriminates the concrete form backing an Actor. Actor is the abstract
// supertype for anything the system authenticates, authorizes, and attributes an action
// to (design §4.1). Every downstream consumer references actor_id — never a user_id or
// service_account_id — so the machine surface expands additively (the BYOP roadmap).
type EnumActorType string

const (
	EnumActorTypeMembership     EnumActorType = "Membership"
	EnumActorTypeServiceAccount EnumActorType = "ServiceAccount"
)

// Actor is thin — an identity plus a discriminator of which concrete form backs it.
// The substance lives in Membership / ServiceAccount (design §4.1).
type Actor struct {
	ID        string
	Type      EnumActorType
	UpdatedAt time.Time
	CreatedAt time.Time
}

// EnumMembershipRole is the *coarse* role (design §4.2). Fine-grained entitlements are
// deferred (OPA is out of starter scope): in the full system they are authored as
// authorization relations and projected to OPA, which reads this coarse role as its
// origin record. Maker/checker are first-class for payment flows. Keep this a small
// closed set so the fine-grained layer is additive.
type EnumMembershipRole string

const (
	EnumMembershipRoleOwner           EnumMembershipRole = "Owner"
	EnumMembershipRoleAdmin           EnumMembershipRole = "Admin"
	EnumMembershipRoleFinanceOperator EnumMembershipRole = "FinanceOperator"
	EnumMembershipRoleViewer          EnumMembershipRole = "Viewer"
	EnumMembershipRoleMaker           EnumMembershipRole = "Maker"
	EnumMembershipRoleChecker         EnumMembershipRole = "Checker"
)

// EnumMembershipStatus lifecycle (design §7):
//
//	invited ──► active ──► suspended ──► active
//	                           └──────► revoked
type EnumMembershipStatus string

const (
	EnumMembershipStatusInvited   EnumMembershipStatus = "Invited"
	EnumMembershipStatusActive    EnumMembershipStatus = "Active"
	EnumMembershipStatusSuspended EnumMembershipStatus = "Suspended"
	EnumMembershipStatusRevoked   EnumMembershipStatus = "Revoked"
)

// Membership is the human operator grant (design §4.2): the many-to-many edge saying
// "this human may operate this Business, in this capacity." It is the human concrete
// form of Actor.
type Membership struct {
	ID     string
	CorrID string // caller-supplied correlation / idempotency key (UNIQUE, standard §8)

	ActorID     string // the Actor this membership backs (1:1)
	BusinessID  string // the account being operated
	KeycloakSub string // the login identity — owned by Keycloak; this is a reference only

	// PersonID is THE reconciliation seam (design §4.2, §5.3). Nullable — set only when
	// the operator is also a natural person held for compliance. Built from day one even
	// while Person is unbuilt; the founder's membership points at their Person, the
	// outsourced accountant's points at null. Match is by verified national ID, never a
	// shared row.
	PersonID *string

	Role      EnumMembershipRole // the coarse role — the whole access story in starter scope
	Status    EnumMembershipStatus
	UpdatedAt time.Time
	CreatedAt time.Time
}

// EnumInvitationStatus lifecycle (design §7):
//
//	pending ──► accepted
//	   ├──────► expired
//	   └──────► revoked
type EnumInvitationStatus string

const (
	EnumInvitationStatusPending  EnumInvitationStatus = "Pending"
	EnumInvitationStatusAccepted EnumInvitationStatus = "Accepted"
	EnumInvitationStatusExpired  EnumInvitationStatus = "Expired"
	EnumInvitationStatusRevoked  EnumInvitationStatus = "Revoked"
)

// Invitation models the state *before* a User exists (design §4.4): the gap between "an
// admin invited a colleague by email" and "that colleague has a Keycloak account". At
// invite time there is no `sub` yet, so a Membership cannot exist. On acceptance a
// Keycloak user is created and a Membership is materialized; the Invitation is transient,
// the Membership is durable.
type Invitation struct {
	ID     string
	CorrID string // caller-supplied correlation / idempotency key (UNIQUE, standard §8)

	BusinessID   string
	InvitedEmail string             // where the invite is sent
	IntendedRole EnumMembershipRole // the role the resulting Membership will carry
	Token        string             // the single-use acceptance secret
	Status       EnumInvitationStatus
	InvitedBy    string    // the actor_id who issued the invite
	ExpiresAt    time.Time // invite expiry
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

// EnumServiceAccountStatus lifecycle: active ──► suspended ──► active; ──► revoked.
type EnumServiceAccountStatus string

const (
	EnumServiceAccountStatusActive    EnumServiceAccountStatus = "Active"
	EnumServiceAccountStatusSuspended EnumServiceAccountStatus = "Suspended"
	EnumServiceAccountStatusRevoked   EnumServiceAccountStatus = "Revoked"
)

// ServiceAccount is the machine concrete form of Actor (design §4.5): a non-human that
// acts on a Business (e.g. runs its scheduled payments), backed by a credential — not by
// a Person. Starter scope builds topology 1: scoped to exactly one Business. Topology 2
// (BYOP, the GitHub-App model — one app, many installs) is deferred; the initiation seam
// already references actor_id so it slots in additively.
type ServiceAccount struct {
	ID     string
	CorrID string // caller-supplied correlation / idempotency key (UNIQUE, standard §8)

	ActorID       string // the Actor this service account backs (1:1)
	BusinessID    string // the account it acts on
	Label         string // human-readable identifier for the credential
	Status        EnumServiceAccountStatus
	CredentialRef *string // pointer to secret material; the secret itself is not stored here
	UpdatedAt     time.Time
	CreatedAt     time.Time
}
