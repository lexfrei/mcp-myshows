package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// minRating and maxRating bound the rating accepted by the rate/check tools.
const (
	minRating = 0
	maxRating = 5
)

// ptrBool returns a pointer to value, for the *bool annotation hint fields.
func ptrBool(value bool) *bool { return &value }

// readOnly builds annotations for a tool that only reads remote state.
func readOnly(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		Title:         title,
		ReadOnlyHint:  true,
		OpenWorldHint: ptrBool(true),
	}
}

// write builds annotations for a tool that performs additive, idempotent
// updates to the account. IdempotentHint follows the MCP definition -- calling
// the tool again with the SAME arguments has no additional effect: re-marking
// an episode watched is a no-op, and re-applying the same rating leaves the
// rating unchanged. (Calling with different arguments, e.g. a new rating, of
// course changes state; that is not what the hint claims.)
func write(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		Title:           title,
		ReadOnlyHint:    false,
		DestructiveHint: ptrBool(false),
		IdempotentHint:  true,
		OpenWorldHint:   ptrBool(true),
	}
}

// writeDestructive builds annotations for a write tool that can drop data, such
// as removing a show from the tracker.
func writeDestructive(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		Title:           title,
		ReadOnlyHint:    false,
		DestructiveHint: ptrBool(true),
		IdempotentHint:  true,
		OpenWorldHint:   ptrBool(true),
	}
}

// pointerResult applies the nil guard and error mapping shared by the tools
// that return a single pointer value. It returns only (value, error): the
// tool's *mcp.CallToolResult is always nil because the SDK builds the
// client-facing result from the error on failure and from the structured output
// on success, so handlers prepend nil themselves.
func pointerResult[R any](failMessage string, fetch func() (*R, error)) (R, error) {
	var zero R

	result, err := fetch()
	if err != nil {
		return zero, myshowsErr(failMessage, err)
	}

	if result == nil {
		return zero, myshowsErr(failMessage, ErrEmptyResponse)
	}

	return *result, nil
}

// idLookup validates a positive ID and runs an ID-keyed fetch that returns a
// pointer result, reusing pointerResult for the nil guard and error mapping.
func idLookup[R any](
	ctx context.Context,
	entityID int,
	failMessage string,
	fetch func(context.Context, int) (*R, error),
) (R, error) {
	if entityID <= 0 {
		var zero R

		return zero, validationErr(ErrIDRequired)
	}

	return pointerResult(failMessage, func() (*R, error) {
		return fetch(ctx, entityID)
	})
}

// validateRating rejects ratings outside the inclusive 0-5 range.
func validateRating(rating int) error {
	if rating < minRating || rating > maxRating {
		return ErrInvalidRating
	}

	return nil
}
