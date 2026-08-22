package keycloak

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const defaultTimeout = 30 * time.Second

type Config struct {
	BaseURL string
	Realm   string
	Timeout time.Duration
}

type Option func(*Client)

func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		if l != nil {
			c.log = l
		}
	}
}

// ReqOption customises a single request.
type ReqOption func(*reqConfig)

type reqConfig struct {
	query   url.Values
	headers http.Header
}

func newReqConfig(opts []ReqOption) *reqConfig {
	rc := &reqConfig{}
	for _, o := range opts {
		o(rc)
	}
	return rc
}

func WithQuery(key, value string) ReqOption {
	return func(rc *reqConfig) {
		if rc.query == nil {
			rc.query = url.Values{}
		}
		rc.query.Add(key, value)
	}
}

func WithHeader(key, value string) ReqOption {
	return func(rc *reqConfig) {
		if rc.headers == nil {
			rc.headers = http.Header{}
		}
		rc.headers.Set(key, value)
	}
}

// WithBearerToken sets the Keycloak Admin API bearer token for a single request. Token
// acquisition (client-credentials grant) is a caller/infra concern — the adapter stays a
// thin, faithful mirror of the API and does not embed an auth flow.
func WithBearerToken(token string) ReqOption {
	return WithHeader("Authorization", "Bearer "+token)
}
