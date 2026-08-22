// Mirrors the Keycloak Admin REST API — Users (https://www.keycloak.org/docs-api, retrieved 2026-08).

package keycloak_admin

import (
	"context"
	"net/http"
	"net/url"

	"github.com/BinMunawir/maal_business/internal/adapters/keycloak"
)

// UsersPost - POST: /admin/realms/{realm}/users  (Create a new user)
func UsersPost(ctx context.Context, c *keycloak.Client, in ModelUserRepresentation, opts ...keycloak.ReqOption) (*UserCreateRes, error) {
	return keycloak.Call[UserCreateRes](ctx, c, http.MethodPost, "/admin/realms/"+url.PathEscape(c.Realm())+"/users", in, opts...)
}

// UsersGet - GET: /admin/realms/{realm}/users  (Get users — filter with query params, e.g. email + exact)
func UsersGet(ctx context.Context, c *keycloak.Client, opts ...keycloak.ReqOption) (*[]ModelUserRepresentation, error) {
	return keycloak.Call[[]ModelUserRepresentation](ctx, c, http.MethodGet, "/admin/realms/"+url.PathEscape(c.Realm())+"/users", nil, opts...)
}
