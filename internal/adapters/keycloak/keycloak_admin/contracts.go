package keycloak_admin

import "github.com/BinMunawir/client/internal/adapters/keycloak"

// Keycloak's create endpoints return 201 with an empty body on success and an error
// object on failure. These envelopes capture exactly that (api-adapter-standard §4): on
// success the embedded *ErrorApi stays nil; on failure it is populated and the caller
// checks res.ErrorApi. GET (search) endpoints return a bare JSON array and skip the
// envelope entirely — see users.go / organizations.go.
type UserCreateRes struct {
	*keycloak.ErrorApi
}

type OrganizationCreateRes struct {
	*keycloak.ErrorApi
}
