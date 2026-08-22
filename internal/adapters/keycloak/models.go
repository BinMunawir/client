package keycloak

// Ptr returns a pointer to v, for setting pointer-typed request-model fields inline
// (api-adapter-standard §5.2). Anti-corruption ports in this service prefer the Go 1.26
// builtin new(expr) per service-standard §6.4, but Ptr is part of the adapter's public
// surface as the standard prescribes.
func Ptr[T any](v T) *T { return &v }
