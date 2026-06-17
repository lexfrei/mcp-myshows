package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

// ShowStatusParams defines the parameters for the myshows_show_status tool.
type ShowStatusParams struct {
	ShowIDs []int `json:"showIds" jsonschema:"Show IDs to look up the watch status for"`
}

// ShowStatusResult is the output of the myshows_show_status tool.
type ShowStatusResult struct {
	Count    int                  `json:"count"`
	Statuses []myshows.ShowStatus `json:"statuses"`
}

// ShowStatusTool returns the MCP tool definition for myshows_show_status.
func ShowStatusTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "myshows_show_status",
		Description: "Get the authenticated user's watch status (watching, later, cancelled, finished) for one or " +
			"more shows by ID -- a lightweight alternative to my_shows when only the status is needed",
		Annotations: readOnly("Show Status"),
	}
}

// NewShowStatusHandler creates a handler for the myshows_show_status tool.
func NewShowStatusHandler(api myshows.API) mcp.ToolHandlerFor[ShowStatusParams, ShowStatusResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params ShowStatusParams,
	) (*mcp.CallToolResult, ShowStatusResult, error) {
		idErr := validateShowIDs(params.ShowIDs)
		if idErr != nil {
			return nil, ShowStatusResult{}, validationErr(idErr)
		}

		statuses, err := api.ShowStatuses(ctx, params.ShowIDs)
		if err != nil {
			return nil, ShowStatusResult{}, myshowsErr("show status failed", err)
		}

		return nil, ShowStatusResult{Count: len(statuses), Statuses: statuses}, nil
	}
}

// validateShowIDs rejects an empty list or any non-positive ID.
func validateShowIDs(ids []int) error {
	if len(ids) == 0 {
		return ErrIDRequired
	}

	for _, showID := range ids {
		if showID <= 0 {
			return ErrIDRequired
		}
	}

	return nil
}
