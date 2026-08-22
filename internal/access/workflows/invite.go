package workflows

import (
	"fmt"
	"time"

	access_activities "github.com/BinMunawir/client/internal/access/activities"
	"github.com/BinMunawir/client/internal/core"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// TaskQueueAccess is the task queue for the access capability (invitations, memberships,
// service accounts).
const TaskQueueAccess = "access"

type InviteInput struct {
	CorrID       string
	BusinessID   string
	InvitedEmail string
	IntendedRole string
	InvitedBy    string
	TTLHours     int
}

type InviteOutput struct {
	Invitation core.Invitation
}

// Invite issues a pending Invitation. It is a single durable step today; modelling it as a
// workflow keeps the whole access surface uniform and leaves room for a notification step
// (sending the invite email) to be added additively.
func Invite(ctx workflow.Context, in InviteInput) (InviteOutput, error) {
	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	logger.Info("Invite workflow started", "corr_id", in.CorrID, "business_id", in.BusinessID)

	var out InviteOutput
	var inv core.Invitation

	issueIn := access_activities.IssueInvitationInput{
		CorrID:       in.CorrID,
		BusinessID:   in.BusinessID,
		InvitedEmail: in.InvitedEmail,
		IntendedRole: in.IntendedRole,
		InvitedBy:    in.InvitedBy,
		TTLHours:     in.TTLHours,
	}
	if err := workflow.ExecuteActivity(ctx, access_activities.IssueInvitation, issueIn).Get(ctx, &inv); err != nil {
		return out, fmt.Errorf("access_activities.IssueInvitation: %w", err)
	}

	out.Invitation = inv
	logger.Info("Invite workflow completed", "invitation_id", inv.ID)
	return out, nil
}
