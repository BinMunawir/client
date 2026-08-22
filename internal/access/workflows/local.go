package workflows

import (
	"context"
	"fmt"

	access_activities "github.com/BinMunawir/client/internal/access/activities"
)

// The pure-function twins of the access workflows (standard §6.1): identical sequences with
// direct calls, for local runs and tests. Keep each step-for-step in sync with its Temporal
// workflow above.

func InviteLocal(ctx context.Context, in InviteInput) (InviteOutput, error) {
	var out InviteOutput
	inv, err := access_activities.IssueInvitation(ctx, access_activities.IssueInvitationInput{
		CorrID:       in.CorrID,
		BusinessID:   in.BusinessID,
		InvitedEmail: in.InvitedEmail,
		IntendedRole: in.IntendedRole,
		InvitedBy:    in.InvitedBy,
		TTLHours:     in.TTLHours,
	})
	if err != nil {
		return out, fmt.Errorf("access_activities.IssueInvitation: %w", err)
	}
	out.Invitation = inv
	return out, nil
}

func AcceptInvitationLocal(ctx context.Context, in AcceptInvitationInput) (AcceptInvitationOutput, error) {
	var out AcceptInvitationOutput

	inv, err := access_activities.ValidateInvitation(ctx, access_activities.ValidateInvitationInput{Token: in.Token})
	if err != nil {
		return out, fmt.Errorf("access_activities.ValidateInvitation: %w", err)
	}

	sub, err := access_activities.ProvisionUser(ctx, access_activities.ProvisionUserInput{Invitation: inv})
	if err != nil {
		return out, fmt.Errorf("access_activities.ProvisionUser: %w", err)
	}

	mbr, err := access_activities.MaterializeMembership(ctx, access_activities.MaterializeMembershipInput{Invitation: inv, Sub: sub})
	if err != nil {
		return out, fmt.Errorf("access_activities.MaterializeMembership: %w", err)
	}

	accepted, err := access_activities.AcceptInvite(ctx, access_activities.AcceptInviteInput{Invitation: inv})
	if err != nil {
		return out, fmt.Errorf("access_activities.AcceptInvite: %w", err)
	}

	out.Membership = mbr
	out.Invitation = accepted
	return out, nil
}

func ProvisionServiceAccountLocal(ctx context.Context, in ProvisionServiceAccountInput) (ProvisionServiceAccountOutput, error) {
	var out ProvisionServiceAccountOutput
	sa, err := access_activities.CreateServiceAccount(ctx, access_activities.CreateServiceAccountInput{
		CorrID:     in.CorrID,
		BusinessID: in.BusinessID,
		Label:      in.Label,
	})
	if err != nil {
		return out, fmt.Errorf("access_activities.CreateServiceAccount: %w", err)
	}
	out.ServiceAccount = sa
	return out, nil
}
