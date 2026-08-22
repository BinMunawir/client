// Package idp is the anti-corruption port wrapping the Keycloak adapter for the business
// slice. It speaks domain in and domain out (service standard §6.4), keeping Keycloak
// types out of core and the activities.
package idp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/BinMunawir/client/config"
	"github.com/BinMunawir/client/internal/adapters/keycloak"
	"github.com/BinMunawir/client/internal/adapters/keycloak/keycloak_admin"
	"github.com/BinMunawir/client/internal/core"
)

// CreateOrganization provisions the one Keycloak Organization this Business maps to (1:1,
// design §3.2) and returns its id, which the caller stores as Business.KeycloakOrgID. The
// call is idempotent: an already-existing organization (a replayed activity) resolves the
// existing id rather than failing.
func CreateOrganization(ctx context.Context, biz core.Business) (string, error) {
	token := config.CNF.Keycloak.AdminToken
	if token == "" {
		// Dev fallback: Keycloak admin credentials are not configured, so we do not call the
		// real Admin API. Return a deterministic stub reference so local/dev runs work
		// end-to-end against just Postgres. Set KEYCLOAK__ADMIN_TOKEN to hit real Keycloak.
		id := "kc-org-dev-" + biz.ID
		slog.WarnContext(ctx, "keycloak not configured; using dev stub organization id", "keycloak_org_id", id)
		return id, nil
	}

	c, err := client()
	if err != nil {
		return "", fmt.Errorf("idp.client: %w", err)
	}

	res, err := keycloak_admin.OrganizationsPost(ctx, c, toOrgRep(biz), keycloak.WithBearerToken(token))
	if err != nil {
		return "", fmt.Errorf("keycloak.OrganizationsPost: %w", err)
	}
	if res.ErrorApi != nil && !isConflict(res.ErrorApi) {
		return "", fmt.Errorf("keycloak create organization: %s", res.ErrorApi.Msg())
	}

	id, err := findOrganizationID(ctx, c, token, biz.LegalName)
	if err != nil {
		return "", fmt.Errorf("findOrganizationID: %w", err)
	}
	return id, nil
}

func findOrganizationID(ctx context.Context, c *keycloak.Client, token, name string) (string, error) {
	list, err := keycloak_admin.OrganizationsGet(ctx, c,
		keycloak.WithQuery("search", name),
		keycloak.WithBearerToken(token))
	if err != nil {
		return "", fmt.Errorf("keycloak.OrganizationsGet: %w", err)
	}
	for _, o := range *list {
		if o.Name != nil && *o.Name == name && o.Id != nil {
			return *o.Id, nil
		}
	}
	return "", fmt.Errorf("organization %q not found after create", name)
}

func client() (*keycloak.Client, error) {
	return keycloak.ClientNew(keycloak.Config{
		BaseURL: config.CNF.Keycloak.BaseURL,
		Realm:   config.CNF.Keycloak.Realm,
	})
}

// isConflict treats an "already exists" error as a benign conflict, so a replayed create
// is idempotent (the outbound idempotency story for Keycloak, which dedups by natural key).
func isConflict(e *keycloak.ErrorApi) bool {
	return strings.Contains(strings.ToLower(e.Msg()), "exist")
}
