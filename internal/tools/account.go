package tools

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

// ProfileParams defines the parameters for the myshows_profile tool.
type ProfileParams struct {
	Login string `json:"login,omitempty" jsonschema:"Username to look up; empty returns your own profile"`
}

// ProfileTool returns the MCP tool definition for myshows_profile.
func ProfileTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_profile",
		Description: "Fetch a user profile with watch-time statistics; defaults to the authenticated user",
		Annotations: readOnly("Profile"),
	}
}

// NewProfileHandler creates a handler for the myshows_profile tool.
func NewProfileHandler(api myshows.API) mcp.ToolHandlerFor[ProfileParams, myshows.Profile] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params ProfileParams,
	) (*mcp.CallToolResult, myshows.Profile, error) {
		profile, err := pointerResult("profile failed", func() (*myshows.Profile, error) {
			return api.Profile(ctx, params.Login)
		})

		return nil, profile, err
	}
}

// MyShowsParams defines the parameters for the myshows_my_shows tool. The whole
// tracked list can be large, so the filters and pagination (applied
// client-side; the API has no server-side paging) keep the result small.
type MyShowsParams struct {
	Login  string `json:"login,omitempty"  jsonschema:"Username to look up; empty returns your own tracked shows"`
	ShowID int    `json:"showId,omitempty" jsonschema:"Return only the tracked show with this ID"`
	Query  string `json:"query,omitempty"  jsonschema:"Case-insensitive substring filter on the show title"`
	Status string `json:"status,omitempty" jsonschema:"Filter by watch status: watching, later, cancelled, finished"`
	Limit  int    `json:"limit,omitempty"  jsonschema:"Maximum number of shows to return (0 = all)"`
	Offset int    `json:"offset,omitempty" jsonschema:"Skip this many shows before returning, for pagination"`
}

// MyShowsResult is the output of the myshows_my_shows tool. Total is the number
// of shows matching the filters before pagination; Count is the number actually
// returned in this page.
type MyShowsResult struct {
	Count int                   `json:"count"`
	Total int                   `json:"total"`
	Shows []myshows.ProfileShow `json:"shows"`
}

// MyShowsTool returns the MCP tool definition for myshows_my_shows.
func MyShowsTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "myshows_my_shows",
		Description: "List a user's tracked shows with watch status, rating, and progress; defaults to the " +
			"authenticated user. The full list is large -- use showId, query, status, limit, and offset to narrow it.",
		Annotations: readOnly("My Shows"),
	}
}

// NewMyShowsHandler creates a handler for the myshows_my_shows tool.
func NewMyShowsHandler(api myshows.API) mcp.ToolHandlerFor[MyShowsParams, MyShowsResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params MyShowsParams,
	) (*mcp.CallToolResult, MyShowsResult, error) {
		shows, err := api.MyShows(ctx, params.Login)
		if err != nil {
			return nil, MyShowsResult{}, myshowsErr("my shows failed", err)
		}

		filtered := filterProfileShows(shows, &params)
		page := paginate(filtered, params.Offset, params.Limit)

		return nil, MyShowsResult{Count: len(page), Total: len(filtered), Shows: page}, nil
	}
}

// MyEpisodesParams defines the parameters for the myshows_my_episodes tool.
type MyEpisodesParams struct {
	ShowID int `json:"showId" jsonschema:"Show ID to list watched episodes for"`
}

// MyEpisodesResult is the output of the myshows_my_episodes tool.
type MyEpisodesResult struct {
	Count    int                      `json:"count"`
	Episodes []myshows.WatchedEpisode `json:"episodes"`
}

// MyEpisodesTool returns the MCP tool definition for myshows_my_episodes.
func MyEpisodesTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_my_episodes",
		Description: "List the authenticated user's watched episodes for a show, with watch dates and ratings",
		Annotations: readOnly("My Episodes"),
	}
}

// NewMyEpisodesHandler creates a handler for the myshows_my_episodes tool.
func NewMyEpisodesHandler(api myshows.API) mcp.ToolHandlerFor[MyEpisodesParams, MyEpisodesResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params MyEpisodesParams,
	) (*mcp.CallToolResult, MyEpisodesResult, error) {
		if params.ShowID <= 0 {
			return nil, MyEpisodesResult{}, validationErr(ErrIDRequired)
		}

		episodes, err := api.MyEpisodes(ctx, params.ShowID)
		if err != nil {
			return nil, MyEpisodesResult{}, myshowsErr("my episodes failed", err)
		}

		return nil, MyEpisodesResult{Count: len(episodes), Episodes: episodes}, nil
	}
}

// UnwatchedParams defines the parameters for the myshows_unwatched tool.
type UnwatchedParams struct {
	List string `json:"list,omitempty" jsonschema:"Which list: unwatched (default, all pending) or next (next episode per show)"`
}

// UnwatchedResult is the output of the myshows_unwatched tool.
type UnwatchedResult struct {
	Count    int                   `json:"count"`
	Episodes []myshows.NextEpisode `json:"episodes"`
}

// UnwatchedTool returns the MCP tool definition for myshows_unwatched.
func UnwatchedTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_unwatched",
		Description: "List episodes the authenticated user has not watched yet, paired with their show",
		Annotations: readOnly("Unwatched Episodes"),
	}
}

// NewUnwatchedHandler creates a handler for the myshows_unwatched tool.
func NewUnwatchedHandler(api myshows.API) mcp.ToolHandlerFor[UnwatchedParams, UnwatchedResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params UnwatchedParams,
	) (*mcp.CallToolResult, UnwatchedResult, error) {
		list, listErr := resolveList(params.List)
		if listErr != nil {
			return nil, UnwatchedResult{}, validationErr(listErr)
		}

		episodes, err := api.NextEpisodes(ctx, list)
		if err != nil {
			return nil, UnwatchedResult{}, myshowsErr("unwatched failed", err)
		}

		return nil, UnwatchedResult{Count: len(episodes), Episodes: episodes}, nil
	}
}

// CountersParams has no parameters.
type CountersParams struct{}

// CountersTool returns the MCP tool definition for myshows_counters.
func CountersTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_counters",
		Description: "Report the authenticated user's pending-item counts (unwatched episodes, new comments, achievements)",
		Annotations: readOnly("Counters"),
	}
}

// NewCountersHandler creates a handler for the myshows_counters tool.
func NewCountersHandler(api myshows.API) mcp.ToolHandlerFor[CountersParams, myshows.Counters] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ CountersParams,
	) (*mcp.CallToolResult, myshows.Counters, error) {
		counters, err := pointerResult("counters failed", func() (*myshows.Counters, error) {
			return api.Counters(ctx)
		})

		return nil, counters, err
	}
}

// RecommendationsParams defines the parameters for the myshows_recommendations tool.
type RecommendationsParams struct {
	Count int `json:"count,omitempty" jsonschema:"Maximum number of recommendations"`
}

// RecommendationsResult is the output of the myshows_recommendations tool.
type RecommendationsResult struct {
	Count           int                      `json:"count"`
	Recommendations []myshows.Recommendation `json:"recommendations"`
}

// RecommendationsTool returns the MCP tool definition for myshows_recommendations.
func RecommendationsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_recommendations",
		Description: "List personalised show recommendations for the authenticated user",
		Annotations: readOnly("Recommendations"),
	}
}

// NewRecommendationsHandler creates a handler for the myshows_recommendations tool.
func NewRecommendationsHandler(api myshows.API) mcp.ToolHandlerFor[RecommendationsParams, RecommendationsResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params RecommendationsParams,
	) (*mcp.CallToolResult, RecommendationsResult, error) {
		recommendations, err := api.Recommendations(ctx, params.Count)
		if err != nil {
			return nil, RecommendationsResult{}, myshowsErr("recommendations failed", err)
		}

		return nil, RecommendationsResult{Count: len(recommendations), Recommendations: recommendations}, nil
	}
}

// filterProfileShows applies the optional showId/query/status filters. With no
// filter set it returns the input unchanged.
func filterProfileShows(shows []myshows.ProfileShow, params *MyShowsParams) []myshows.ProfileShow {
	if params.ShowID == 0 && params.Query == "" && params.Status == "" {
		return shows
	}

	query := strings.ToLower(params.Query)
	out := make([]myshows.ProfileShow, 0, len(shows))

	for i := range shows {
		entry := &shows[i]

		switch {
		case params.ShowID != 0 && entry.Show.ID != params.ShowID:
			continue
		case params.Status != "" && !strings.EqualFold(entry.WatchStatus, params.Status):
			continue
		case query != "" && !titleContains(&entry.Show, query):
			continue
		}

		out = append(out, *entry)
	}

	return out
}

// titleContains reports whether either title holds the lower-cased query.
func titleContains(show *myshows.Show, lowerQuery string) bool {
	return strings.Contains(strings.ToLower(show.Title), lowerQuery) ||
		strings.Contains(strings.ToLower(show.TitleOriginal), lowerQuery)
}

// paginate returns the offset/limit window of shows. A limit of 0 means no cap.
func paginate(shows []myshows.ProfileShow, offset, limit int) []myshows.ProfileShow {
	if offset < 0 {
		offset = 0
	}

	if offset >= len(shows) {
		return []myshows.ProfileShow{}
	}

	shows = shows[offset:]

	if limit > 0 && limit < len(shows) {
		shows = shows[:limit]
	}

	return shows
}

// resolveList validates the episode list name, defaulting to unwatched.
func resolveList(list string) (string, error) {
	switch list {
	case "":
		return "unwatched", nil
	case "unwatched", "next":
		return list, nil
	default:
		return "", ErrInvalidList
	}
}
