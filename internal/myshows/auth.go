package myshows

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

// sessionResponse is the body returned by the session login endpoint. The
// endpoint also returns a refreshToken, but this client re-logs in with the
// stored username/password on expiry, so the refresh token is not used.
type sessionResponse struct {
	Token string `json:"token"`
}

// persistedToken is the on-disk token cache format.
type persistedToken struct {
	Token string `json:"token"`
}

// Permissions for the persisted token cache: private directory, private file.
const (
	tokenDirPerm  = 0o700
	tokenFilePerm = 0o600
)

// ensureAuth guarantees a Bearer token is available, logging in on first use.
func (c *Client) ensureAuth(ctx context.Context) error {
	if c.currentToken() != "" {
		return nil
	}

	c.authMu.Lock()
	defer c.authMu.Unlock()

	// Another goroutine may have logged in while we waited for the lock.
	if c.currentToken() != "" {
		return nil
	}

	return c.login(ctx)
}

// reauth refreshes a rejected token. stale is the token that just failed; if
// another goroutine already refreshed it while we waited for the lock, reauth is
// a no-op, so a burst of concurrent 401s triggers only one re-login. It fails
// when only a raw token override was supplied, since there is nothing to log in
// with.
func (c *Client) reauth(ctx context.Context, stale string) error {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	if c.currentToken() != stale {
		return nil
	}

	if c.username == "" || c.password == "" {
		return ErrNotAuthenticated
	}

	return c.login(ctx)
}

// login exchanges username/password for a Bearer token via the session
// endpoint and persists it. The caller must hold authMu.
func (c *Client) login(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		return ErrNoCredentials
	}

	body, err := json.Marshal(map[string]string{
		"login":    c.username,
		"password": c.password,
	})
	if err != nil {
		return errors.Wrap(err, "marshal login")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sessionURL, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "build login request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return loginErr(err, "login request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return errors.Wrapf(ErrLoginFailed, "status %d", resp.StatusCode)
	}

	var session sessionResponse

	decErr := json.NewDecoder(resp.Body).Decode(&session)
	if decErr != nil {
		return loginErr(decErr, "decode login response")
	}

	if session.Token == "" {
		return ErrLoginFailed
	}

	c.setToken(session.Token)
	c.saveToken(session.Token)

	return nil
}

// loadToken seeds the token from the on-disk cache, best-effort.
func (c *Client) loadToken() {
	if c.tokenPath == "" {
		return
	}

	data, err := os.ReadFile(c.tokenPath)
	if err != nil {
		return
	}

	var stored persistedToken

	unErr := json.Unmarshal(data, &stored)
	if unErr != nil {
		return
	}

	if stored.Token != "" {
		c.setToken(stored.Token)
	}
}

// saveToken persists the token cache, best-effort. A raw token override (no
// token path) is never written.
func (c *Client) saveToken(token string) {
	if c.tokenPath == "" {
		return
	}

	data, err := json.Marshal(persistedToken{Token: token})
	if err != nil {
		return
	}

	mkErr := os.MkdirAll(filepath.Dir(c.tokenPath), tokenDirPerm)
	if mkErr != nil {
		return
	}

	_ = os.WriteFile(c.tokenPath, data, tokenFilePerm)
}
