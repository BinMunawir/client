package workflows

import (
	"context"
	"fmt"

	activities "github.com/BinMunawir/client/internal/client/activities"
	"github.com/BinMunawir/client/internal/core"
)

// OnboardLocal is the pure-function twin of Onboard: the identical sequence executed with
// direct calls, for local runs and tests (standard §6.1). Keep it step-for-step in sync
// with the Temporal workflow. Because the twin calls the same activity functions, those
// activities are plain func(ctx, In) (Out, error) that never read workflow.Context.
func OnboardLocal(ctx context.Context, in OnboardInput) (OnboardOutput, error) {
	var out OnboardOutput
	var biz core.Business
	var err error

	// 1. Register
	registerIn := activities.RegisterInput{
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
	biz, err = activities.Register(ctx, registerIn)
	if err != nil {
		return out, fmt.Errorf("activities.Register: %w", err)
	}

	// 2. ProvisionOrg
	biz, err = activities.ProvisionOrg(ctx, activities.ProvisionOrgInput{Biz: biz})
	if err != nil {
		return out, fmt.Errorf("activities.ProvisionOrg: %w", err)
	}

	// 3. Activate
	biz, err = activities.Activate(ctx, activities.ActivateInput{Biz: biz})
	if err != nil {
		return out, fmt.Errorf("activities.Activate: %w", err)
	}

	out.Business = biz
	return out, nil
}
