package tools_test

import (
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
	"github.com/lexfrei/mcp-myshows/internal/tools"
)

var errBoom = errors.New("boom")

const testIdent = "test"

func TestSearchHandler_Success(t *testing.T) {
	t.Parallel()

	api := &mockAPI{searchResult: []myshows.Show{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}}}
	handler := tools.NewSearchHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.SearchParams{Query: "x"})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.Count != 2 || len(result.Results) != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}
}

func TestSearchHandler_EmptyQuery(t *testing.T) {
	t.Parallel()

	handler := tools.NewSearchHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.SearchParams{})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSearchHandler_APIError(t *testing.T) {
	t.Parallel()

	handler := tools.NewSearchHandler(&mockAPI{searchErr: errBoom})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.SearchParams{Query: "x"})
	if !errors.Is(err, tools.ErrMyShows) {
		t.Fatalf("expected myshows error, got %v", err)
	}
}

func TestShowHandler_InvalidID(t *testing.T) {
	t.Parallel()

	handler := tools.NewShowHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ShowParams{ShowID: 0})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestShowHandler_Success(t *testing.T) {
	t.Parallel()

	api := &mockAPI{showResult: &myshows.ShowDetails{Show: myshows.Show{ID: 187, Title: "Breaking Bad"}}}
	handler := tools.NewShowHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ShowParams{ShowID: 187})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.ID != 187 || result.Title != "Breaking Bad" {
		t.Errorf("unexpected show: %+v", result)
	}
}

func TestShowHandler_EmptyResponse(t *testing.T) {
	t.Parallel()

	handler := tools.NewShowHandler(&mockAPI{showResult: nil})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ShowParams{ShowID: 187})
	if !errors.Is(err, tools.ErrMyShows) {
		t.Fatalf("expected myshows error, got %v", err)
	}
}

func TestShowByExternalHandler_InvalidSource(t *testing.T) {
	t.Parallel()

	handler := tools.NewShowByExternalHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.ShowByExternalParams{ID: "tt0903747", Source: "rotten"})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestShowByExternalHandler_MissingID(t *testing.T) {
	t.Parallel()

	handler := tools.NewShowByExternalHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.ShowByExternalParams{Source: "imdb"})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestEpisodeHandler_InvalidID(t *testing.T) {
	t.Parallel()

	handler := tools.NewEpisodeHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.EpisodeParams{EpisodeID: -1})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestEpisodeHandler_Success(t *testing.T) {
	t.Parallel()

	api := &mockAPI{episodeResult: &myshows.Episode{ID: 5, Title: "Pilot"}}
	handler := tools.NewEpisodeHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.EpisodeParams{EpisodeID: 5})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.ID != 5 {
		t.Errorf("ID = %d, want 5", result.ID)
	}
}

func TestUnwatchedHandler_DefaultsToUnwatched(t *testing.T) {
	t.Parallel()

	api := &mockAPI{}
	handler := tools.NewUnwatchedHandler(api)

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.UnwatchedParams{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if api.lastList != "unwatched" {
		t.Errorf("lastList = %q, want unwatched", api.lastList)
	}
}

func TestUnwatchedHandler_InvalidList(t *testing.T) {
	t.Parallel()

	handler := tools.NewUnwatchedHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.UnwatchedParams{List: "bogus"})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCountersHandler_Success(t *testing.T) {
	t.Parallel()

	api := &mockAPI{countersResult: &myshows.Counters{UnwatchedEpisodes: 42}}
	handler := tools.NewCountersHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.CountersParams{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.UnwatchedEpisodes != 42 {
		t.Errorf("UnwatchedEpisodes = %d, want 42", result.UnwatchedEpisodes)
	}
}

func TestCheckEpisodeHandler_Success(t *testing.T) {
	t.Parallel()

	api := &mockAPI{}
	handler := tools.NewCheckEpisodeHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.CheckEpisodeParams{EpisodeID: 99, Rating: 5})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if !result.Success || api.checkID != 99 || api.checkRating != 5 {
		t.Errorf("unexpected: success=%v id=%d rating=%d", result.Success, api.checkID, api.checkRating)
	}
}

func TestCheckEpisodeHandler_InvalidRating(t *testing.T) {
	t.Parallel()

	handler := tools.NewCheckEpisodeHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.CheckEpisodeParams{EpisodeID: 99, Rating: 7})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSetShowStatusHandler_InvalidStatus(t *testing.T) {
	t.Parallel()

	handler := tools.NewSetShowStatusHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.SetShowStatusParams{ShowID: 1, Status: "bogus"})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSetShowStatusHandler_Success(t *testing.T) {
	t.Parallel()

	api := &mockAPI{}
	handler := tools.NewSetShowStatusHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.SetShowStatusParams{ShowID: 187, Status: "watching"})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if !result.Success || api.statusID != 187 || api.statusValue != "watching" {
		t.Errorf("unexpected: %+v / id=%d status=%q", result, api.statusID, api.statusValue)
	}
}

func TestRateShowHandler_InvalidRating(t *testing.T) {
	t.Parallel()

	handler := tools.NewRateShowHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.RateShowParams{ShowID: 1, Rating: 99})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestProfileHandler_Success(t *testing.T) {
	t.Parallel()

	handler := tools.NewProfileHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ProfileParams{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
}

// TestMyShowsHandler_EmptyLoginForwarded pins that an empty login is intentional
// (it means "the authenticated user") and is forwarded to the client unchanged.
func TestMyShowsHandler_EmptyLoginForwarded(t *testing.T) {
	t.Parallel()

	api := &mockAPI{lastLogin: "sentinel"}
	handler := tools.NewMyShowsHandler(api)

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyShowsParams{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if api.lastLogin != "" {
		t.Errorf("lastLogin = %q, want empty (own profile)", api.lastLogin)
	}
}

func TestMyEpisodesHandler_InvalidID(t *testing.T) {
	t.Parallel()

	handler := tools.NewMyEpisodesHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyEpisodesParams{ShowID: 0})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestMyEpisodesHandler_Success(t *testing.T) {
	t.Parallel()

	handler := tools.NewMyEpisodesHandler(&mockAPI{})

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyEpisodesParams{ShowID: 187})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("Count = %d, want 0", result.Count)
	}
}

func TestRecommendationsHandler_Success(t *testing.T) {
	t.Parallel()

	handler := tools.NewRecommendationsHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.RecommendationsParams{Count: 5})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
}

func TestUnCheckEpisodeHandler_InvalidID(t *testing.T) {
	t.Parallel()

	handler := tools.NewUnCheckEpisodeHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.UnCheckEpisodeParams{EpisodeID: 0})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestUnCheckEpisodeHandler_Success(t *testing.T) {
	t.Parallel()

	handler := tools.NewUnCheckEpisodeHandler(&mockAPI{})

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.UnCheckEpisodeParams{EpisodeID: 99})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestRateEpisodeHandler_InvalidRating(t *testing.T) {
	t.Parallel()

	handler := tools.NewRateEpisodeHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.RateEpisodeParams{EpisodeID: 1, Rating: 99})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRateEpisodeHandler_Success(t *testing.T) {
	t.Parallel()

	handler := tools.NewRateEpisodeHandler(&mockAPI{})

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.RateEpisodeParams{EpisodeID: 5, Rating: 4})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

// TestTopHandler_PassesModeThrough documents that the mode is forwarded to the
// client unchanged -- the tool does not validate or rewrite it; the API does.
func TestTopHandler_PassesModeThrough(t *testing.T) {
	t.Parallel()

	api := &mockAPI{}
	handler := tools.NewTopHandler(api)

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.TopParams{Mode: "all"})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if api.lastMode != "all" {
		t.Errorf("lastMode = %q, want all", api.lastMode)
	}
}

// callTool registers one tool on an in-memory server and calls it the way a real
// MCP client does, returning the round-tripped result so tests can assert the
// contract the SDK actually presents to clients (rather than the discarded
// *CallToolResult the handler returns alongside an error).
func callTool(t *testing.T, args any) *mcp.CallToolResult {
	t.Helper()

	server := mcp.NewServer(&mcp.Implementation{Name: testIdent, Version: testIdent}, nil)
	mcp.AddTool(server, tools.SearchTool(), tools.NewSearchHandler(&mockAPI{
		searchResult: []myshows.Show{{ID: 1, Title: "A"}},
	}))

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer func() { _ = serverSession.Close() }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: testIdent}, nil)

	clientSession, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = clientSession.Close() }()

	result, err := clientSession.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "myshows_search",
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool transport error: %v", err)
	}

	return result
}

// TestErrorContract_RoundTrip pins the real error contract: a validation failure
// surfaces to the client as IsError with a populated message, even though the
// handler returns a nil *CallToolResult.
func TestErrorContract_RoundTrip(t *testing.T) {
	t.Parallel()

	result := callTool(t, map[string]any{"query": ""})

	if !result.IsError {
		t.Fatal("expected IsError on a validation failure")
	}

	if len(result.Content) == 0 {
		t.Error("expected error content to be populated")
	}
}

// TestSuccessContract_RoundTrip pins the success contract: structured output is
// delivered and IsError is false.
func TestSuccessContract_RoundTrip(t *testing.T) {
	t.Parallel()

	result := callTool(t, map[string]any{"query": "matrix"})

	if result.IsError {
		t.Fatalf("unexpected error result: %+v", result)
	}

	if result.StructuredContent == nil {
		t.Error("expected structured content on success")
	}
}
