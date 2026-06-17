package tools

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

// SearchParams defines the parameters for the myshows_search tool.
type SearchParams struct {
	Query      string `json:"query"                jsonschema:"Search query: a show title or keywords"`
	WithStatus bool   `json:"withStatus,omitempty" jsonschema:"Also include your watch status for each result (requires authentication)"`
}

// SearchHit is one search result, optionally carrying the authenticated user's
// watch status. It embeds Show, so without withStatus it is identical to a Show.
type SearchHit struct {
	myshows.Show

	WatchStatus string `json:"watchStatus,omitempty"`
}

// SearchResult is the output of the myshows_search tool.
type SearchResult struct {
	Count   int         `json:"count"`
	Results []SearchHit `json:"results"`
}

// SearchTool returns the MCP tool definition for myshows_search.
func SearchTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "myshows_search",
		Description: "Search MyShows for TV shows by title or keywords. Set withStatus to also include your " +
			"watch status per result (requires authentication).",
		Annotations: readOnly("Search Shows"),
	}
}

// NewSearchHandler creates a handler for the myshows_search tool.
func NewSearchHandler(api myshows.API) mcp.ToolHandlerFor[SearchParams, SearchResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params SearchParams,
	) (*mcp.CallToolResult, SearchResult, error) {
		if params.Query == "" {
			return nil, SearchResult{}, validationErr(ErrQueryRequired)
		}

		shows, err := api.Search(ctx, params.Query)
		if err != nil {
			return nil, SearchResult{}, myshowsErr("search failed", err)
		}

		hits := make([]SearchHit, len(shows))
		for i := range shows {
			hits[i] = SearchHit{Show: shows[i]}
		}

		if params.WithStatus && len(hits) > 0 {
			statusErr := enrichWithStatus(ctx, api, hits)
			if statusErr != nil {
				return nil, SearchResult{}, statusErr
			}
		}

		return nil, SearchResult{Count: len(hits), Results: hits}, nil
	}
}

// enrichWithStatus fills each hit's WatchStatus from profile.ShowStatuses. When
// no credentials are configured it returns a clear ErrStatusNeedsAuth.
func enrichWithStatus(ctx context.Context, api myshows.API, hits []SearchHit) error {
	ids := make([]int, len(hits))
	for i := range hits {
		ids[i] = hits[i].ID
	}

	statuses, err := api.ShowStatuses(ctx, ids)
	if err != nil {
		if errors.Is(err, myshows.ErrNotAuthenticated) || errors.Is(err, myshows.ErrNoCredentials) {
			return validationErr(ErrStatusNeedsAuth)
		}

		return myshowsErr("search status enrichment failed", err)
	}

	byID := make(map[int]string, len(statuses))
	for i := range statuses {
		byID[statuses[i].ShowID] = statuses[i].WatchStatus
	}

	for i := range hits {
		hits[i].WatchStatus = byID[hits[i].ID]
	}

	return nil
}

// ShowParams defines the parameters for the myshows_show tool.
type ShowParams struct {
	ShowID       int  `json:"showId"                 jsonschema:"Show ID (from a search result)"`
	WithEpisodes bool `json:"withEpisodes,omitempty" jsonschema:"Include the full episode list"`
}

// ShowTool returns the MCP tool definition for myshows_show.
func ShowTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_show",
		Description: "Fetch detailed information about a show: title, year, status, ratings, network, and optionally its episodes",
		Annotations: readOnly("Show Details"),
	}
}

// NewShowHandler creates a handler for the myshows_show tool.
func NewShowHandler(api myshows.API) mcp.ToolHandlerFor[ShowParams, myshows.ShowDetails] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params ShowParams,
	) (*mcp.CallToolResult, myshows.ShowDetails, error) {
		if params.ShowID <= 0 {
			return nil, myshows.ShowDetails{}, validationErr(ErrIDRequired)
		}

		show, err := pointerResult("show lookup failed", func() (*myshows.ShowDetails, error) {
			return api.GetShow(ctx, params.ShowID, params.WithEpisodes)
		})

		return nil, show, err
	}
}

// ShowByExternalParams defines the parameters for the myshows_show_by_external tool.
type ShowByExternalParams struct {
	ID     string `json:"id"     jsonschema:"External identifier value (e.g. an IMDb ID like tt0903747)"`
	Source string `json:"source" jsonschema:"External source: imdb, kinopoisk, or thetvdb"`
}

// ShowByExternalTool returns the MCP tool definition for myshows_show_by_external.
func ShowByExternalTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_show_by_external",
		Description: "Look up a show by an external identifier from imdb, kinopoisk, or thetvdb",
		Annotations: readOnly("Show By External ID"),
	}
}

// NewShowByExternalHandler creates a handler for the myshows_show_by_external tool.
func NewShowByExternalHandler(api myshows.API) mcp.ToolHandlerFor[ShowByExternalParams, myshows.ShowDetails] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params ShowByExternalParams,
	) (*mcp.CallToolResult, myshows.ShowDetails, error) {
		if params.ID == "" {
			return nil, myshows.ShowDetails{}, validationErr(ErrExternalIDRequired)
		}

		srcErr := validateSource(params.Source)
		if srcErr != nil {
			return nil, myshows.ShowDetails{}, validationErr(srcErr)
		}

		show, err := pointerResult("show lookup failed", func() (*myshows.ShowDetails, error) {
			return api.GetShowByExternal(ctx, params.ID, params.Source)
		})

		return nil, show, err
	}
}

// EpisodeParams defines the parameters for the myshows_episode tool.
type EpisodeParams struct {
	EpisodeID int `json:"episodeId" jsonschema:"Episode ID"`
}

// EpisodeTool returns the MCP tool definition for myshows_episode.
func EpisodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_episode",
		Description: "Fetch detailed information about a single episode",
		Annotations: readOnly("Episode Details"),
	}
}

// NewEpisodeHandler creates a handler for the myshows_episode tool.
func NewEpisodeHandler(api myshows.API) mcp.ToolHandlerFor[EpisodeParams, myshows.Episode] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params EpisodeParams,
	) (*mcp.CallToolResult, myshows.Episode, error) {
		episode, err := idLookup(ctx, params.EpisodeID, "episode lookup failed", api.GetEpisode)

		return nil, episode, err
	}
}

// TopParams defines the parameters for the myshows_top tool.
type TopParams struct {
	Mode  string `json:"mode,omitempty"  jsonschema:"Chart mode; omit to use the API default. The known accepted value is 'all'. The API rejects unknown modes."`
	Count int    `json:"count,omitempty" jsonschema:"Maximum number of results"`
}

// TopResult is the output of the myshows_top tool.
type TopResult struct {
	Count   int                  `json:"count"`
	Results []myshows.RankedShow `json:"results"`
}

// TopTool returns the MCP tool definition for myshows_top.
func TopTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_top",
		Description: "List the top-ranked shows chart",
		Annotations: readOnly("Top Shows"),
	}
}

// NewTopHandler creates a handler for the myshows_top tool.
func NewTopHandler(api myshows.API) mcp.ToolHandlerFor[TopParams, TopResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		params TopParams,
	) (*mcp.CallToolResult, TopResult, error) {
		ranked, err := api.Top(ctx, params.Mode, params.Count)
		if err != nil {
			return nil, TopResult{}, myshowsErr("top failed", err)
		}

		return nil, TopResult{Count: len(ranked), Results: ranked}, nil
	}
}

// GenresParams has no parameters.
type GenresParams struct{}

// GenresResult is the output of the myshows_genres tool.
type GenresResult struct {
	Count  int             `json:"count"`
	Genres []myshows.Genre `json:"genres"`
}

// GenresTool returns the MCP tool definition for myshows_genres.
func GenresTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "myshows_genres",
		Description: "List all available genres",
		Annotations: readOnly("Genres"),
	}
}

// NewGenresHandler creates a handler for the myshows_genres tool.
func NewGenresHandler(api myshows.API) mcp.ToolHandlerFor[GenresParams, GenresResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ GenresParams,
	) (*mcp.CallToolResult, GenresResult, error) {
		genres, err := api.Genres(ctx)
		if err != nil {
			return nil, GenresResult{}, myshowsErr("genres failed", err)
		}

		return nil, GenresResult{Count: len(genres), Genres: genres}, nil
	}
}

// validateSource rejects unknown external-id sources.
func validateSource(source string) error {
	switch source {
	case "imdb", "kinopoisk", "thetvdb":
		return nil
	default:
		return ErrInvalidSource
	}
}
