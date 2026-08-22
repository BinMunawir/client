package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const defaultUserAgent = "maal-business-keycloak-go/0.1"

// Client is the transport core for the Keycloak Admin REST API — a typed mirror, no
// business logic (api-adapter-standard §1). The realm is client-level config; endpoint
// functions in the sub-package build paths as /admin/realms/{realm}/...
type Client struct {
	base  *url.URL
	realm string
	hc    *http.Client
	log   *slog.Logger
	ua    string
}

func ClientNew(cfg Config, opts ...Option) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("keycloak: base URL is required")
	}
	base, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("keycloak: invalid base URL %q: %w", cfg.BaseURL, err)
	}
	if base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("keycloak: base URL %q must be absolute (scheme + host)", cfg.BaseURL)
	}
	if cfg.Realm == "" {
		return nil, errors.New("keycloak: realm is required")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	c := &Client{
		base:  base,
		realm: cfg.Realm,
		hc:    &http.Client{Timeout: timeout},
		log:   slog.Default(),
		ua:    defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Realm returns the realm this client targets; endpoint functions use it to build paths.
func (c *Client) Realm() string { return c.realm }

func (c *Client) newRequest(ctx context.Context, method, path string, body any, rc *reqConfig) (*http.Request, error) {
	u, err := c.base.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("keycloak: invalid path %q: %w", path, err)
	}
	if len(rc.query) > 0 {
		u.RawQuery = rc.query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("keycloak: encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("keycloak: build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.ua)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Per-request headers win over the defaults above.
	for k, vs := range rc.headers {
		req.Header.Del(k)
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	return req, nil
}

func (c *Client) send(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := c.hc.Do(req)
	dur := time.Since(start)
	if err != nil {
		c.log.DebugContext(req.Context(), "keycloak request failed",
			"method", req.Method, "url", req.URL.String(), "dur", dur, "err", err)
		return nil, err
	}
	c.log.DebugContext(req.Context(), "keycloak request",
		"method", req.Method, "url", req.URL.String(), "status", resp.StatusCode, "dur", dur)
	return resp, nil
}

// Call is the single generic entrypoint every endpoint function delegates to.
func Call[Out any](ctx context.Context, c *Client, method, path string, body any, opts ...ReqOption) (*Out, error) {
	rc := newReqConfig(opts)

	req, err := c.newRequest(ctx, method, path, body, rc)
	if err != nil {
		return nil, err
	}

	resp, err := c.send(req)
	if err != nil {
		return nil, fmt.Errorf("keycloak: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("keycloak: read response body: %w", err)
	}

	out := new(Out)
	if resp.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(raw)) == 0 {
		return out, nil // nothing to decode (e.g. 201 Created with empty body)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil, fmt.Errorf("keycloak: decode response (http %d): %w", resp.StatusCode, err)
	}
	return out, nil
}
