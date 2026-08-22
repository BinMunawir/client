package access_activities

import (
	"context"
	"fmt"
	"time"

	"github.com/BinMunawir/maal_business/internal/access/store"
	"github.com/BinMunawir/maal_business/internal/core"
)

const defaultInviteTTLHours = 72

type IssueInvitationInput struct {
	CorrID       string
	BusinessID   string
	InvitedEmail string
	IntendedRole string
	InvitedBy    string // the actor_id issuing the invite
	TTLHours     int
}

// IssueInvitation creates a pending Invitation with a single-use token (design §4.4).
// Skeleton: map → persist → classify (standard §6.2).
func IssueInvitation(ctx context.Context, in IssueInvitationInput) (core.Invitation, error) {
	inv, err := issueInputToEntity(in)
	if err != nil {
		return core.Invitation{}, fmt.Errorf("issueInputToEntity: %w", err)
	}
	inv, err = issuePersist(ctx, inv)
	if err != nil {
		return core.Invitation{}, fmt.Errorf("issuePersist: %w", err)
	}
	return inv, nil
}

func issueInputToEntity(in IssueInvitationInput) (core.Invitation, error) {
	role := core.EnumMembershipRole(in.IntendedRole)
	if !validMembershipRole(role) {
		return core.Invitation{}, fmt.Errorf("invalid intended_role %q", in.IntendedRole)
	}
	ttl := in.TTLHours
	if ttl <= 0 {
		ttl = defaultInviteTTLHours
	}
	now := time.Now().UTC()
	return core.Invitation{
		ID:           core.InvitationID(),
		CorrID:       in.CorrID,
		BusinessID:   in.BusinessID,
		InvitedEmail: in.InvitedEmail,
		IntendedRole: role,
		Token:        core.NewInviteToken(),
		Status:       core.EnumInvitationStatusPending,
		InvitedBy:    in.InvitedBy,
		ExpiresAt:    now.Add(time.Duration(ttl) * time.Hour),
		UpdatedAt:    now,
		CreatedAt:    now,
	}, nil
}

func issuePersist(ctx context.Context, inv core.Invitation) (core.Invitation, error) {
	inserted, err := store.InsertInvitation(ctx, inv)
	if err != nil {
		if isUniqueViolation(err) {
			// Idempotent: this invite was already issued for this correlation key (standard §8).
			existing, ferr := store.FetchInvitationByCorrID(ctx, inv.CorrID)
			if ferr != nil {
				return inv, fmt.Errorf("issuePersist: reconcile: %w", ferr)
			}
			return existing, nil
		}
		return inv, fmt.Errorf("store.InsertInvitation: %w", err)
	}
	return inserted, nil
}

func validMembershipRole(r core.EnumMembershipRole) bool {
	switch r {
	case core.EnumMembershipRoleOwner, core.EnumMembershipRoleAdmin, core.EnumMembershipRoleFinanceOperator,
		core.EnumMembershipRoleViewer, core.EnumMembershipRoleMaker, core.EnumMembershipRoleChecker:
		return true
	default:
		return false
	}
}
