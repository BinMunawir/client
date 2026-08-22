package access_activities

import (
	"context"
	"fmt"

	"github.com/BinMunawir/client/internal/access/idp"
	"github.com/BinMunawir/client/internal/core"
)

type ProvisionUserInput struct {
	Invitation core.Invitation
}

// ProvisionUser creates the Keycloak user for the accepted invitation and returns its
// `sub` (design §4.4). This is a port-only step — the side effect lives behind the idp
// anti-corruption port; the `sub` is persisted by MaterializeMembership next.
func ProvisionUser(ctx context.Context, in ProvisionUserInput) (string, error) {
	sub, err := idp.CreateUser(ctx, in.Invitation)
	if err != nil {
		return "", fmt.Errorf("idp.CreateUser: %w", err)
	}
	return sub, nil
}
