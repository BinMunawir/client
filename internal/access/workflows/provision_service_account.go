package workflows

import (
	"fmt"
	"time"

	access_activities "github.com/BinMunawir/client/internal/access/activities"
	"github.com/BinMunawir/client/internal/core"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type ProvisionServiceAccountInput struct {
	CorrID     string
	BusinessID string
	Label      string
}

type ProvisionServiceAccountOutput struct {
	ServiceAccount core.ServiceAccount
}

// ProvisionServiceAccount stands up a machine operator for a Business (design §4.5,
// topology 1). This flow touches no external service — it is pure persistence — so it is
// the one that runs end-to-end locally against just Postgres.
func ProvisionServiceAccount(ctx workflow.Context, in ProvisionServiceAccountInput) (ProvisionServiceAccountOutput, error) {
	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	logger.Info("ProvisionServiceAccount workflow started", "corr_id", in.CorrID, "business_id", in.BusinessID)

	var out ProvisionServiceAccountOutput
	var sa core.ServiceAccount

	createIn := access_activities.CreateServiceAccountInput{
		CorrID:     in.CorrID,
		BusinessID: in.BusinessID,
		Label:      in.Label,
	}
	if err := workflow.ExecuteActivity(ctx, access_activities.CreateServiceAccount, createIn).Get(ctx, &sa); err != nil {
		return out, fmt.Errorf("access_activities.CreateServiceAccount: %w", err)
	}

	out.ServiceAccount = sa
	logger.Info("ProvisionServiceAccount workflow completed", "service_account_id", sa.ID, "actor_id", sa.ActorID)
	return out, nil
}
