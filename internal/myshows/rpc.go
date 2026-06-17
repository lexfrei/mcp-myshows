package myshows

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
)

// httpUnauthorized is the status (and JSON-RPC error code) MyShows returns when
// the Bearer token is missing or expired.
const httpUnauthorized = 401

// Doer is the subset of *http.Client the client relies on, so tests can inject
// a custom round-tripper.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// rpcRequest is a JSON-RPC 2.0 request envelope.
type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

// rpcError is the JSON-RPC error object.
type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// rpcResponse is a JSON-RPC 2.0 response envelope.
type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

// call issues a single JSON-RPC request to the API endpoint. When out is
// non-nil the result is unmarshalled into it. A 401 (HTTP status or JSON-RPC
// error code) maps to ErrNotAuthenticated so callers can re-authenticate.
func (c *Client) call(ctx context.Context, method string, params, out any) error {
	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1})
	if err != nil {
		return errors.Wrap(err, "marshal request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "build request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	token := c.currentToken()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return apiErr(err, "POST %s", method)
	}
	defer func() { _ = resp.Body.Close() }()

	return decodeResponse(method, resp, out)
}

// decodeResponse classifies the HTTP status and JSON-RPC envelope, unmarshalling
// the result into out when the call succeeded.
func decodeResponse(method string, resp *http.Response, out any) error {
	if resp.StatusCode == httpUnauthorized {
		return ErrNotAuthenticated
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return errors.Wrapf(ErrAPI, "%s: status %d", method, resp.StatusCode)
	}

	var envelope rpcResponse

	decErr := json.NewDecoder(resp.Body).Decode(&envelope)
	if decErr != nil {
		return apiErr(decErr, "decode %s response", method)
	}

	if envelope.Error != nil {
		// MyShows mirrors an expired/missing session into the JSON-RPC error code
		// (returned with HTTP 200), not just the HTTP status, so check it here too.
		if envelope.Error.Code == httpUnauthorized {
			return ErrNotAuthenticated
		}

		return errors.Wrapf(ErrAPI, "%s: %s (code %d)", method, envelope.Error.Message, envelope.Error.Code)
	}

	if out != nil {
		unErr := json.Unmarshal(envelope.Result, out)
		if unErr != nil {
			return apiErr(unErr, "unmarshal %s result", method)
		}
	}

	return nil
}

// callPublic runs a method that does not require authentication. A configured
// token, if present, is still attached -- harmless, and it lets the API
// personalise responses when the operator is logged in.
func (c *Client) callPublic(ctx context.Context, method string, params, out any) error {
	return c.call(ctx, method, params, out)
}

// callAuthed ensures a session, runs the method, and retries once after
// re-authenticating if the token was rejected mid-flight.
func (c *Client) callAuthed(ctx context.Context, method string, params, out any) error {
	authErr := c.ensureAuth(ctx)
	if authErr != nil {
		return authErr
	}

	// Capture the token this attempt uses so reauth can tell whether another
	// goroutine already refreshed it after a concurrent 401.
	token := c.currentToken()

	err := c.call(ctx, method, params, out)
	if errors.Is(err, ErrNotAuthenticated) {
		reErr := c.reauth(ctx, token)
		if reErr != nil {
			return reErr
		}

		return c.call(ctx, method, params, out)
	}

	return err
}
