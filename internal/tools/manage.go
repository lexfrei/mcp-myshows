package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

// ActionResult is the output of the write tools.
type ActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CheckEpisodeParams defines the parameters for the myshows_check_episode tool.
type CheckEpisodeParams struct {
	EpisodeID int `json:"episodeId"        jsonschema:"Episode ID to mark watched"`
	Rating    int `json:"rating,omitempty" jsonschema:"Optional rating 0-5; 0 or omitted leaves it unrated"`
}

// CheckEpisodeTool returns the MCP tool definition for myshows_check_episode.
func CheckEpisodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_check_episode",
		Description: "Mark an episode as watched, optionally with a rating (0-5; 0 leaves it unrated)",
		Annotations: write("Mark Episode Watched"),
	}
}

// NewCheckEpisodeHandler creates a handler for the myshows_check_episode tool.
func NewCheckEpisodeHandler(api myshows.API) mcp.ToolHandlerFor[CheckEpisodeParams, ActionResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params CheckEpisodeParams,
	) (*mcp.CallToolResult, ActionResult, error) {
		validationFailure := firstErr(requirePositive(params.EpisodeID), validateRating(params.Rating))
		if validationFailure != nil {
			return nil, ActionResult{}, validationErr(validationFailure)
		}

		result, err := actionResult(
			"check episode failed",
			fmt.Sprintf("episode %d marked watched", params.EpisodeID),
			func() error { return api.CheckEpisode(ctx, params.EpisodeID, params.Rating) },
		)

		return nil, result, err
	}
}

// UnCheckEpisodeParams defines the parameters for the myshows_uncheck_episode tool.
type UnCheckEpisodeParams struct {
	EpisodeID int `json:"episodeId" jsonschema:"Episode ID to mark unwatched"`
}

// UnCheckEpisodeTool returns the MCP tool definition for myshows_uncheck_episode.
func UnCheckEpisodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_uncheck_episode",
		Description: "Mark a previously watched episode as unwatched",
		Annotations: write("Mark Episode Unwatched"),
	}
}

// NewUnCheckEpisodeHandler creates a handler for the myshows_uncheck_episode tool.
func NewUnCheckEpisodeHandler(api myshows.API) mcp.ToolHandlerFor[UnCheckEpisodeParams, ActionResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params UnCheckEpisodeParams,
	) (*mcp.CallToolResult, ActionResult, error) {
		reqErr := requirePositive(params.EpisodeID)
		if reqErr != nil {
			return nil, ActionResult{}, validationErr(reqErr)
		}

		result, err := actionResult(
			"uncheck episode failed",
			fmt.Sprintf("episode %d marked unwatched", params.EpisodeID),
			func() error { return api.UnCheckEpisode(ctx, params.EpisodeID) },
		)

		return nil, result, err
	}
}

// SetShowStatusParams defines the parameters for the myshows_set_show_status tool.
type SetShowStatusParams struct {
	ShowID int    `json:"showId" jsonschema:"Show ID"`
	Status string `json:"status" jsonschema:"New status: watching, later, cancelled, or remove"`
}

// SetShowStatusTool returns the MCP tool definition for myshows_set_show_status.
func SetShowStatusTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_set_show_status",
		Description: "Set the tracking status of a show: watching, later, cancelled, or remove (drops it from the tracker)",
		Annotations: writeDestructive("Set Show Status"),
	}
}

// NewSetShowStatusHandler creates a handler for the myshows_set_show_status tool.
func NewSetShowStatusHandler(api myshows.API) mcp.ToolHandlerFor[SetShowStatusParams, ActionResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params SetShowStatusParams,
	) (*mcp.CallToolResult, ActionResult, error) {
		validationFailure := firstErr(requirePositive(params.ShowID), validateStatus(params.Status))
		if validationFailure != nil {
			return nil, ActionResult{}, validationErr(validationFailure)
		}

		result, err := actionResult(
			"set show status failed",
			fmt.Sprintf("show %d status set to %s", params.ShowID, params.Status),
			func() error { return api.SetShowStatus(ctx, params.ShowID, params.Status) },
		)

		return nil, result, err
	}
}

// RateShowParams defines the parameters for the myshows_rate_show tool.
type RateShowParams struct {
	ShowID int `json:"showId" jsonschema:"Show ID"`
	Rating int `json:"rating" jsonschema:"Rating 0-5 (0 removes the rating)"`
}

// RateShowTool returns the MCP tool definition for myshows_rate_show.
func RateShowTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_rate_show",
		Description: "Set the authenticated user's rating (0-5) for a show",
		Annotations: write("Rate Show"),
	}
}

// NewRateShowHandler creates a handler for the myshows_rate_show tool.
func NewRateShowHandler(api myshows.API) mcp.ToolHandlerFor[RateShowParams, ActionResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params RateShowParams,
	) (*mcp.CallToolResult, ActionResult, error) {
		validationFailure := firstErr(requirePositive(params.ShowID), validateRating(params.Rating))
		if validationFailure != nil {
			return nil, ActionResult{}, validationErr(validationFailure)
		}

		result, err := actionResult(
			"rate show failed",
			fmt.Sprintf("show %d rated %d", params.ShowID, params.Rating),
			func() error { return api.RateShow(ctx, params.ShowID, params.Rating) },
		)

		return nil, result, err
	}
}

// RateEpisodeParams defines the parameters for the myshows_rate_episode tool.
type RateEpisodeParams struct {
	EpisodeID int `json:"episodeId" jsonschema:"Episode ID"`
	Rating    int `json:"rating"    jsonschema:"Rating 0-5 (0 removes the rating)"`
}

// RateEpisodeTool returns the MCP tool definition for myshows_rate_episode.
func RateEpisodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_rate_episode",
		Description: "Set the authenticated user's rating (0-5) for an episode",
		Annotations: write("Rate Episode"),
	}
}

// NewRateEpisodeHandler creates a handler for the myshows_rate_episode tool.
func NewRateEpisodeHandler(api myshows.API) mcp.ToolHandlerFor[RateEpisodeParams, ActionResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params RateEpisodeParams,
	) (*mcp.CallToolResult, ActionResult, error) {
		validationFailure := firstErr(requirePositive(params.EpisodeID), validateRating(params.Rating))
		if validationFailure != nil {
			return nil, ActionResult{}, validationErr(validationFailure)
		}

		result, err := actionResult(
			"rate episode failed",
			fmt.Sprintf("episode %d rated %d", params.EpisodeID, params.Rating),
			func() error { return api.RateEpisode(ctx, params.EpisodeID, params.Rating) },
		)

		return nil, result, err
	}
}

// actionResult runs a write action and maps the outcome to an ActionResult. The
// tool's *mcp.CallToolResult is always nil (the SDK builds it from the error or
// the structured output), so handlers prepend nil themselves.
func actionResult(failMessage, okMessage string, action func() error) (ActionResult, error) {
	err := action()
	if err != nil {
		return ActionResult{}, myshowsErr(failMessage, err)
	}

	return ActionResult{Success: true, Message: okMessage}, nil
}

// requirePositive rejects a non-positive entity ID.
func requirePositive(id int) error {
	if id <= 0 {
		return ErrIDRequired
	}

	return nil
}

// firstErr returns the first non-nil error, or nil.
func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// validateStatus rejects unknown show statuses.
func validateStatus(status string) error {
	switch status {
	case "watching", "later", "cancelled", "remove":
		return nil
	default:
		return ErrInvalidStatus
	}
}
