// Mirrors the Keycloak Admin REST API — Organizations (https://www.keycloak.org/docs-api, retrieved 2026-08).

package keycloak_admin

import (
	"context"
	"net/http"
	"net/url"

	"github.com/BinMunawir/maal_business/internal/adapters/keycloak"
)

// OrganizationsPost - POST: /admin/realms/{realm}/organizations  (Create a new organization)
func OrganizationsPost(ctx context.Context, c *keycloak.Client, in ModelOrganizationRepresentation, opts ...keycloak.ReqOption) (*OrganizationCreateRes, error) {
	return keycloak.Call[OrganizationCreateRes](ctx, c, http.MethodPost, "/admin/realms/"+url.PathEscape(c.Realm())+"/organizations", in, opts...)
}

// OrganizationsGet - GET: /admin/realms/{realm}/organizations  (Get organizations — filter with the `search` query param)
func OrganizationsGet(ctx context.Context, c *keycloak.Client, opts ...keycloak.ReqOption) (*[]ModelOrganizationRepresentation, error) {
	return keycloak.Call[[]ModelOrganizationRepresentation](ctx, c, http.MethodGet, "/admin/realms/"+url.PathEscape(c.Realm())+"/organizations", nil, opts...)
}
