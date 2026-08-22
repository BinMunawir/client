package access_activities

import (
	"context"
	"fmt"
	"time"

	"github.com/BinMunawir/client/internal/access/store"
	"github.com/BinMunawir/client/internal/core"
)

type MaterializeMembershipInput struct {
	Invitation core.Invitation
	Sub        string // the Keycloak `sub` from ProvisionUser
}

// MaterializeMembership creates the durable Membership from the accepted invitation
// (design §4.4). It first creates the Actor (the initiation abstraction, §4.1) then the
// Membership that backs it. Skeleton: map → persist → classify (standard §6.2).
func MaterializeMembership(ctx context.Context, in MaterializeMembershipInput) (core.Membership, error) {
	actor, mbr := materializeInputToEntity(in)
	mbr, err := materializePersist(ctx, actor, mbr)
	if err != nil {
		return core.Membership{}, fmt.Errorf("materializePersist: %w", err)
	}
	return mbr, nil
}

func materializeInputToEntity(in MaterializeMembershipInput) (core.Actor, core.Membership) {
	now := time.Now().UTC()
	actor := core.Actor{
		ID:        core.ActorID(),
		Type:      core.EnumActorTypeMembership,
		UpdatedAt: now,
		CreatedAt: now,
	}
	mbr := core.Membership{
		ID:          core.MembershipID(),
		CorrID:      "mbr-" + in.Invitation.ID, // deterministic idempotency key from the invitation
		ActorID:     actor.ID,
		BusinessID:  in.Invitation.BusinessID,
		KeycloakSub: in.Sub,
		PersonID:    nil, // the reconciliation seam — null for a pure operator (design §6.2, scenario 3)
		Role:        in.Invitation.IntendedRole,
		Status:      core.EnumMembershipStatusActive,
		UpdatedAt:   now,
		CreatedAt:   now,
	}
	return actor, mbr
}

func materializePersist(ctx context.Context, actor core.Actor, mbr core.Membership) (core.Membership, error) {
	if _, err := store.InsertActor(ctx, actor); err != nil {
		// The actor id is a fresh random; a unique-violation here is only possible on a replay
		// that already created it — tolerate it and proceed to the membership.
		if !isUniqueViolation(err) {
			return mbr, fmt.Errorf("store.InsertActor: %w", err)
		}
	}

	inserted, err := store.InsertMembership(ctx, mbr)
	if err != nil {
		if isUniqueViolation(err) {
			// Idempotent replay: membership already materialized for this invitation (standard §8).
			existing, ferr := store.FetchMembershipByCorrID(ctx, mbr.CorrID)
			if ferr != nil {
				return mbr, fmt.Errorf("materializePersist: reconcile: %w", ferr)
			}
			return existing, nil
		}
		return mbr, fmt.Errorf("store.InsertMembership: %w", err)
	}
	return inserted, nil
}
