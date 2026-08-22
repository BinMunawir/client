package core

import (
	"crypto/rand"
	"encoding/hex"
)

// Domain IDs are generated here in core (service standard §8: "Domain IDs are
// generated in core; correlation keys come from the caller"). Each entity gets a
// short human-readable prefix so an id is self-describing in logs and audit trails.
// Correlation / idempotency keys are NOT generated here — those are supplied by the
// caller and persisted under a UNIQUE constraint.
func newID(prefix string) string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return prefix + "_" + hex.EncodeToString(b[:])
}

func OrganizationID() string   { return newID("org") }
func BusinessID() string       { return newID("biz") }
func ClassificationID() string { return newID("cls") }
func ActorID() string          { return newID("act") }
func MembershipID() string     { return newID("mbr") }
func InvitationID() string     { return newID("inv") }
func ServiceAccountID() string { return newID("svc") }

// NewInviteToken mints the single-use acceptance secret carried by an Invitation
// (design §4.4). It is high-entropy and unguessable.
func NewInviteToken() string {
	var b [24]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
