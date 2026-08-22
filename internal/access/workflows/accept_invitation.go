package workflows

import (
	"fmt"
	"time"

	access_activities "github.com/BinMunawir/maal_business/internal/access/activities"
	"github.com/BinMunawir/maal_business/internal/core"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type AcceptInvitationInput struct {
	Token string
}

type AcceptInvitationOutput struct {
	Membership core.Membership
	Invitation core.Invitation
}

// AcceptInvitation is the durable sequence that turns an accepted invite into a durable
// operator grant (design §4.4):
//
//	ValidateInvitation → ProvisionUser (Keycloak) → MaterializeMembership → AcceptInvite
//
// The `sub` returned by ProvisionUser passes between activities (Temporal serializes it),
// exactly as domain entities do.
func AcceptInvitation(ctx workflow.Context, in AcceptInvitationInput) (AcceptInvitationOutput, error) {
	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	logger.Info("AcceptInvitation workflow started")

	var out AcceptInvitationOutput
	var inv core.Invitation

	// 1. ValidateInvitation — pending + unexpired.
	if err := workflow.ExecuteActivity(ctx, access_activities.ValidateInvitation,
		access_activities.ValidateInvitationInput{Token: in.Token}).Get(ctx, &inv); err != nil {
		return out, fmt.Errorf("access_activities.ValidateInvitation: %w", err)
	}

	// 2. ProvisionUser — create the Keycloak user, returning its `sub`.
	var sub string
	if err := workflow.ExecuteActivity(ctx, access_activities.ProvisionUser,
		access_activities.ProvisionUserInput{Invitation: inv}).Get(ctx, &sub); err != nil {
		return out, fmt.Errorf("access_activities.ProvisionUser: %w", err)
	}

	// 3. MaterializeMembership — create the Actor + durable Membership carrying the `sub`.
	var mbr core.Membership
	if err := workflow.ExecuteActivity(ctx, access_activities.MaterializeMembership,
		access_activities.MaterializeMembershipInput{Invitation: inv, Sub: sub}).Get(ctx, &mbr); err != nil {
		return out, fmt.Errorf("access_activities.MaterializeMembership: %w", err)
	}

	// 4. AcceptInvite — invitation pending → accepted.
	var accepted core.Invitation
	if err := workflow.ExecuteActivity(ctx, access_activities.AcceptInvite,
		access_activities.AcceptInviteInput{Invitation: inv}).Get(ctx, &accepted); err != nil {
		return out, fmt.Errorf("access_activities.AcceptInvite: %w", err)
	}

	out.Membership = mbr
	out.Invitation = accepted
	logger.Info("AcceptInvitation workflow completed", "membership_id", mbr.ID, "actor_id", mbr.ActorID)
	return out, nil
}
