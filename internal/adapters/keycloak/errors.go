package keycloak

// ErrorApi mirrors the error bodies the Keycloak Admin API returns. The API is not
// consistent — OAuth-style endpoints return {error, error_description}, resource endpoints
// return {errorMessage} — so all keys are modelled and Msg() picks the most specific one.
type ErrorApi struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorMessage     string `json:"errorMessage"`
}

func (e *ErrorApi) Msg() string {
	switch {
	case e == nil:
		return ""
	case e.ErrorMessage != "":
		return e.ErrorMessage
	case e.ErrorDescription != "":
		return e.ErrorDescription
	default:
		return e.Error
	}
}
