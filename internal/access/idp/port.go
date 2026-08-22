// Package idp is the anti-corruption port wrapping the Keycloak adapter for the access
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

// CreateUser materializes the Keycloak user for an accepted invitation and returns the
// `sub` (design §4.4: "a Keycloak user is created"). Domain in, the login-identity
// reference out — the `sub` is all this domain ever holds of a User (design §4.3). The
// call is idempotent: a user that already exists (a replayed activity) resolves the
// existing `sub` rather than failing.
func CreateUser(ctx context.Context, inv core.Invitation) (string, error) {
	token := config.CNF.Keycloak.AdminToken
	if token == "" {
		// Dev fallback (see business/idp): no Keycloak credentials configured → deterministic
		// stub `sub` so local/dev runs work against just Postgres.
		sub := "kc-sub-dev-" + inv.ID
		slog.WarnContext(ctx, "keycloak not configured; using dev stub user sub", "keycloak_sub", sub)
		return sub, nil
	}

	c, err := client()
	if err != nil {
		return "", fmt.Errorf("idp.client: %w", err)
	}

	res, err := keycloak_admin.UsersPost(ctx, c, toUserRep(inv), keycloak.WithBearerToken(token))
	if err != nil {
		return "", fmt.Errorf("keycloak.UsersPost: %w", err)
	}
	if res.ErrorApi != nil && !isConflict(res.ErrorApi) {
		return "", fmt.Errorf("keycloak create user: %s", res.ErrorApi.Msg())
	}

	sub, err := findUserSub(ctx, c, token, inv.InvitedEmail)
	if err != nil {
		return "", fmt.Errorf("findUserSub: %w", err)
	}
	return sub, nil
}

func findUserSub(ctx context.Context, c *keycloak.Client, token, email string) (string, error) {
	list, err := keycloak_admin.UsersGet(ctx, c,
		keycloak.WithQuery("email", email),
		keycloak.WithQuery("exact", "true"),
		keycloak.WithBearerToken(token))
	if err != nil {
		return "", fmt.Errorf("keycloak.UsersGet: %w", err)
	}
	for _, u := range *list {
		if u.Id != nil {
			return *u.Id, nil
		}
	}
	return "", fmt.Errorf("user %q not found after create", email)
}

func client() (*keycloak.Client, error) {
	return keycloak.ClientNew(keycloak.Config{
		BaseURL: config.CNF.Keycloak.BaseURL,
		Realm:   config.CNF.Keycloak.Realm,
	})
}

func isConflict(e *keycloak.ErrorApi) bool {
	return strings.Contains(strings.ToLower(e.Msg()), "exist")
}
