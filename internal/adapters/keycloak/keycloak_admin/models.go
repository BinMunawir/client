package keycloak_admin

// Request/response models for the Keycloak Admin API — Users and Organizations. Field
// names mirror the wire JSON (api-adapter-standard §5). These models are used on both the
// request and response side, so they follow the request convention: pointer fields.

// ModelUserRepresentation mirrors Keycloak's UserRepresentation (the subset used here).
type ModelUserRepresentation struct {
	Id            *string             `json:"id,omitempty"`
	Username      *string             `json:"username,omitempty"`
	Email         *string             `json:"email,omitempty"`
	FirstName     *string             `json:"firstName,omitempty"`
	LastName      *string             `json:"lastName,omitempty"`
	Enabled       *bool               `json:"enabled,omitempty"`
	EmailVerified *bool               `json:"emailVerified,omitempty"`
	Attributes    map[string][]string `json:"attributes,omitempty"`
}

// ModelOrganizationRepresentation mirrors Keycloak's OrganizationRepresentation (subset).
type ModelOrganizationRepresentation struct {
	Id      *string                   `json:"id,omitempty"`
	Name    *string                   `json:"name,omitempty"`
	Alias   *string                   `json:"alias,omitempty"`
	Domains []ModelOrganizationDomain `json:"domains,omitempty"`
}

type ModelOrganizationDomain struct {
	Name     *string `json:"name,omitempty"`
	Verified *bool   `json:"verified,omitempty"`
}
