package business_activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/BinMunawir/maal_business/internal/business/idp"
	"github.com/BinMunawir/maal_business/internal/business/store"
	"github.com/BinMunawir/maal_business/internal/core"
	"github.com/go-jet/jet/v2/qrm"
)

type ProvisionOrgInput struct {
	Biz core.Business
}

// ProvisionOrg creates the Keycloak Organization this Business maps to (1:1, design §3.2)
// and records the reference-out. Skeleton: port call → persist → classify (standard §6.2).
// The side effect (Keycloak) lives behind the idp anti-corruption port.
func ProvisionOrg(ctx context.Context, in ProvisionOrgInput) (core.Business, error) {
	orgID, err := idp.CreateOrganization(ctx, in.Biz)
	if err != nil {
		return in.Biz, fmt.Errorf("idp.CreateOrganization: %w", err)
	}

	biz := in.Biz
	biz.KeycloakOrgID = &orgID

	biz, err = provisionOrgPersist(ctx, biz)
	if err != nil {
		return biz, fmt.Errorf("provisionOrgPersist: %w", err)
	}
	return biz, nil
}

func provisionOrgPersist(ctx context.Context, biz core.Business) (core.Business, error) {
	updated, err := store.SetBusinessKeycloakOrg(ctx, biz)
	if err == nil {
		return updated, nil
	}
	if errors.Is(err, qrm.ErrNoRows) {
		// The row keyed by id must exist by this point in the flow; a no-match is unexpected.
		return biz, fmt.Errorf("store.SetBusinessKeycloakOrg: UNEXPECTED: %w", err)
	}
	return biz, fmt.Errorf("store.SetBusinessKeycloakOrg: %w", err)
}
