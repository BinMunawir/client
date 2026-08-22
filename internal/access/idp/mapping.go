package idp

import (
	"github.com/BinMunawir/client/internal/adapters/keycloak/keycloak_admin"
	"github.com/BinMunawir/client/internal/core"
)

// toUserRep is the translation boundary: it builds the Keycloak user request from the
// invitation. Pointer-typed request fields are set with the Go 1.26 builtin new(expr)
// (service standard §6.4). The email is used as the username; the account is created
// enabled and email-verified because the invite acceptance already proved control of it.
func toUserRep(inv core.Invitation) keycloak_admin.ModelUserRepresentation {
	return keycloak_admin.ModelUserRepresentation{
		Username:      new(inv.InvitedEmail),
		Email:         new(inv.InvitedEmail),
		Enabled:       new(true),
		EmailVerified: new(true),
	}
}
