// Package myshows is a client for the MyShows.me JSON-RPC API. It authenticates
// via the session endpoint (username/password -> Bearer token, no OAuth AppID)
// and calls the documented v2 JSON-RPC API with a standard Authorization header.
package myshows

import "github.com/cockroachdb/errors"

// ErrNoCredentials indicates no authentication method is configured, so
// account and write methods cannot run.
var ErrNoCredentials = errors.New("no credentials configured")

// ErrLoginFailed indicates the session endpoint rejected the credentials.
var ErrLoginFailed = errors.New("login failed")

// ErrNotAuthenticated indicates the API rejected the request because the
// session token is missing or expired.
var ErrNotAuthenticated = errors.New("not authenticated")

// ErrAPI indicates the API returned a JSON-RPC error or an unexpected response.
var ErrAPI = errors.New("myshows API error")

// apiErr wraps err with a message and marks it as an API failure.
func apiErr(err error, format string, args ...any) error {
	//nolint:wrapcheck // Mark only adds a sentinel category; Wrapf already added the message.
	return errors.Mark(errors.Wrapf(err, format, args...), ErrAPI)
}

// loginErr wraps err with a message and marks it as a login failure.
func loginErr(err error, msg string) error {
	//nolint:wrapcheck // Mark only adds a sentinel category; Wrap already added the message.
	return errors.Mark(errors.Wrap(err, msg), ErrLoginFailed)
}
