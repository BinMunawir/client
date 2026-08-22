package idp

import (
	"github.com/BinMunawir/client/internal/adapters/keycloak/keycloak_admin"
	"github.com/BinMunawir/client/internal/core"
)

// toOrgRep is the translation boundary: it builds the Keycloak organization request from
// the domain entity. Pointer-typed request fields are set with the Go 1.26 builtin
// new(expr) rather than a Ptr helper (service standard §6.4).
func toOrgRep(biz core.Business) keycloak_admin.ModelOrganizationRepresentation {
	return keycloak_admin.ModelOrganizationRepresentation{
		Name:  new(biz.LegalName),
		Alias: new(biz.ID),
	}
}
