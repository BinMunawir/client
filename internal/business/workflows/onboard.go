package workflows

import (
	"fmt"
	"time"

	business_activities "github.com/BinMunawir/maal_business/internal/business/activities"
	"github.com/BinMunawir/maal_business/internal/core"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// TaskQueueBusiness is the task queue for the business-onboarding capability.
const TaskQueueBusiness = "business"

// OnboardInput is the flat, serializable boundary DTO for the onboarding run. The run's
// idempotency key is CorrID (used as the workflow ID by the starter, standard §8).
type OnboardInput struct {
	CorrID            string
	LegalName         string
	TradeName         string
	CRNumber          string
	LegalForm         string
	IncorporationDate *time.Time
	OrganizationID    *string
	SizeSegment       string
	ServiceTier       string
}

type OnboardOutput struct {
	Business core.Business
}

// Onboard is the durable, ordered sequence that stands up a new Business:
//
//	Register (draft + classifications) → ProvisionOrg (Keycloak org 1:1) → Activate (→ active)
//
// Domain entities pass between activities (Temporal serializes them); the workflow's own
// input/output stays flat.
func Onboard(ctx workflow.Context, in OnboardInput) (OnboardOutput, error) {
	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	logger.Info("Onboard workflow started", "corr_id", in.CorrID)

	var out OnboardOutput
	var biz core.Business

	// 1. Register — create the Business in `draft` and attach its classifications.
	registerIn := business_activities.RegisterInput{
		CorrID:            in.CorrID,
		LegalName:         in.LegalName,
		TradeName:         in.TradeName,
		CRNumber:          in.CRNumber,
		LegalForm:         in.LegalForm,
		IncorporationDate: in.IncorporationDate,
		OrganizationID:    in.OrganizationID,
		SizeSegment:       in.SizeSegment,
		ServiceTier:       in.ServiceTier,
	}
	if err := workflow.ExecuteActivity(ctx, business_activities.Register, registerIn).Get(ctx, &biz); err != nil {
		return out, fmt.Errorf("business_activities.Register: %w", err)
	}

	// 2. ProvisionOrg — mint the Keycloak Organization and record the reference-out.
	if err := workflow.ExecuteActivity(ctx, business_activities.ProvisionOrg,
		business_activities.ProvisionOrgInput{Biz: biz}).Get(ctx, &biz); err != nil {
		return out, fmt.Errorf("business_activities.ProvisionOrg: %w", err)
	}

	// 3. Activate — draft → active.
	if err := workflow.ExecuteActivity(ctx, business_activities.Activate,
		business_activities.ActivateInput{Biz: biz}).Get(ctx, &biz); err != nil {
		return out, fmt.Errorf("business_activities.Activate: %w", err)
	}

	out.Business = biz
	logger.Info("Onboard workflow completed", "business_id", biz.ID, "status", biz.Status)
	return out, nil
}
