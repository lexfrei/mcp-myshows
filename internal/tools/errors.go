// Package tools provides MCP tool definitions and handlers for MyShows.
package tools

import "github.com/cockroachdb/errors"

// ErrValidation indicates invalid parameters provided by the caller.
var ErrValidation = errors.New("validation error")

// ErrMyShows indicates a failure talking to the MyShows API.
var ErrMyShows = errors.New("myshows request error")

// ErrQueryRequired is returned when a search query is empty.
var ErrQueryRequired = errors.New("query is required")

// ErrIDRequired is returned when an entity ID is missing or non-positive.
var ErrIDRequired = errors.New("id must be a positive integer")

// ErrExternalIDRequired is returned when an external lookup is missing its id.
var ErrExternalIDRequired = errors.New("id is required")

// ErrInvalidSource is returned when an unknown external source is requested.
var ErrInvalidSource = errors.New("source must be one of: imdb, kinopoisk, thetvdb")

// ErrInvalidStatus is returned when an unknown show status is requested.
var ErrInvalidStatus = errors.New("status must be one of: watching, later, cancelled, remove")

// ErrInvalidRating is returned when a rating is out of range.
var ErrInvalidRating = errors.New("rating must be between 0 and 5")

// ErrInvalidList is returned when an unknown episode list is requested.
var ErrInvalidList = errors.New("list must be one of: unwatched, next")

// ErrEmptyResponse indicates the client returned neither data nor an error,
// which a well-behaved client never does but the interface does not forbid.
var ErrEmptyResponse = errors.New("myshows returned no data")

// validationErr marks an error as a validation error.
func validationErr(err error) error {
	//nolint:wrapcheck // Mark adds a sentinel category; the caller supplies the message.
	return errors.Mark(err, ErrValidation)
}

// myshowsErr wraps a message and underlying error as a MyShows API error.
func myshowsErr(msg string, err error) error {
	//nolint:wrapcheck // Mark adds a sentinel category on top of Wrap which adds context.
	return errors.Mark(errors.Wrap(err, msg), ErrMyShows)
}
