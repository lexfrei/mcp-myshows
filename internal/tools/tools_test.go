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

const (
	titleBreakingBad = "Breaking Bad"
	statusWatching   = "watching"
	statusFinished   = "finished"
	statusLater      = "later"
)

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

	api := &mockAPI{showResult: &myshows.ShowDetails{Show: myshows.Show{ID: 187, Title: titleBreakingBad}}}
	handler := tools.NewShowHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ShowParams{ShowID: 187})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.ID != 187 || result.Title != titleBreakingBad {
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
		tools.SetShowStatusParams{ShowID: 187, Status: statusWatching})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if !result.Success || api.statusID != 187 || api.statusValue != statusWatching {
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

func sampleProfileShows() []myshows.ProfileShow {
	return []myshows.ProfileShow{
		{Show: myshows.Show{ID: 1, Title: titleBreakingBad}, WatchStatus: statusFinished},
		{Show: myshows.Show{ID: 2, Title: "Better Call Saul"}, WatchStatus: statusWatching},
		{Show: myshows.Show{ID: 3, Title: "The Wire"}, WatchStatus: statusFinished},
	}
}

func TestMyShowsHandler_FilterByShowID(t *testing.T) {
	t.Parallel()

	handler := tools.NewMyShowsHandler(&mockAPI{myShowsResult: sampleProfileShows()})

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyShowsParams{ShowID: 3})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.Count != 1 || result.Total != 1 || result.Shows[0].Show.ID != 3 {
		t.Errorf("unexpected: count=%d total=%d shows=%+v", result.Count, result.Total, result.Shows)
	}
}

func TestMyShowsHandler_FilterByStatusAndQuery(t *testing.T) {
	t.Parallel()

	handler := tools.NewMyShowsHandler(&mockAPI{myShowsResult: sampleProfileShows()})

	_, byStatus, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyShowsParams{Status: statusFinished})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if byStatus.Total != 2 {
		t.Errorf("status filter total = %d, want 2", byStatus.Total)
	}

	_, byBoth, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.MyShowsParams{Status: statusFinished, Query: "wire"})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if byBoth.Total != 1 || byBoth.Shows[0].Show.ID != 3 {
		t.Errorf("status+query filter = %+v", byBoth.Shows)
	}
}

func TestMyShowsHandler_Pagination(t *testing.T) {
	t.Parallel()

	handler := tools.NewMyShowsHandler(&mockAPI{myShowsResult: sampleProfileShows()})

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyShowsParams{Offset: 1, Limit: 1})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.Total != 3 || result.Count != 1 || result.Shows[0].Show.ID != 2 {
		t.Errorf("pagination = total %d count %d shows %+v", result.Total, result.Count, result.Shows)
	}
}

func TestMyShowsHandler_PaginationEdges(t *testing.T) {
	t.Parallel()

	handler := tools.NewMyShowsHandler(&mockAPI{myShowsResult: sampleProfileShows()})

	_, beyond, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyShowsParams{Offset: 99})
	if err != nil {
		t.Fatalf("offset beyond: %v", err)
	}

	if beyond.Count != 0 || beyond.Total != 3 || len(beyond.Shows) != 0 {
		t.Errorf("offset beyond list: count=%d total=%d shows=%d", beyond.Count, beyond.Total, len(beyond.Shows))
	}

	_, noCap, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyShowsParams{Limit: 0})
	if err != nil {
		t.Fatalf("limit 0: %v", err)
	}

	if noCap.Count != 3 {
		t.Errorf("limit 0 (no cap) count = %d, want 3", noCap.Count)
	}

	_, negative, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.MyShowsParams{Offset: -5})
	if err != nil {
		t.Fatalf("negative offset: %v", err)
	}

	if negative.Count != 3 {
		t.Errorf("negative offset count = %d, want 3 (should clamp to 0)", negative.Count)
	}
}

func TestSearchHandler_WithStatusPartial(t *testing.T) {
	t.Parallel()

	api := &mockAPI{
		searchResult:   []myshows.Show{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}},
		statusesResult: []myshows.ShowStatus{{ShowID: 1, WatchStatus: statusWatching}},
	}
	handler := tools.NewSearchHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.SearchParams{Query: "x", WithStatus: true})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.Results[0].WatchStatus != statusWatching {
		t.Errorf("matched hit status = %q, want %q", result.Results[0].WatchStatus, statusWatching)
	}

	if result.Results[1].WatchStatus != "" {
		t.Errorf("unmatched hit status = %q, want empty", result.Results[1].WatchStatus)
	}
}

func TestShowStatusHandler_EmptyIDs(t *testing.T) {
	t.Parallel()

	handler := tools.NewShowStatusHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ShowStatusParams{})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestShowStatusHandler_InvalidID(t *testing.T) {
	t.Parallel()

	handler := tools.NewShowStatusHandler(&mockAPI{})

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ShowStatusParams{ShowIDs: []int{1, -2}})
	if !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestShowStatusHandler_Success(t *testing.T) {
	t.Parallel()

	api := &mockAPI{statusesResult: []myshows.ShowStatus{{ShowID: 45534, WatchStatus: statusLater}}}
	handler := tools.NewShowStatusHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{}, tools.ShowStatusParams{ShowIDs: []int{45534}})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.Count != 1 || result.Statuses[0].WatchStatus != statusLater {
		t.Errorf("unexpected statuses: %+v", result.Statuses)
	}

	if len(api.lastShowIDs) != 1 || api.lastShowIDs[0] != 45534 {
		t.Errorf("lastShowIDs = %v, want [45534]", api.lastShowIDs)
	}
}

func TestSearchHandler_WithStatusEnriches(t *testing.T) {
	t.Parallel()

	api := &mockAPI{
		searchResult: []myshows.Show{{ID: 187, Title: titleBreakingBad}, {ID: 45534, Title: "Westworld"}},
		statusesResult: []myshows.ShowStatus{
			{ShowID: 187, WatchStatus: statusFinished},
			{ShowID: 45534, WatchStatus: statusLater},
		},
	}
	handler := tools.NewSearchHandler(api)

	_, result, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.SearchParams{Query: "x", WithStatus: true})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if result.Results[0].WatchStatus != statusFinished || result.Results[1].WatchStatus != statusLater {
		t.Errorf("status not merged: %+v", result.Results)
	}

	if len(api.lastShowIDs) != 2 {
		t.Errorf("lastShowIDs = %v, want 2 ids", api.lastShowIDs)
	}
}

func TestSearchHandler_WithStatusNotAuthenticated(t *testing.T) {
	t.Parallel()

	api := &mockAPI{
		searchResult: []myshows.Show{{ID: 1, Title: "A"}},
		statusesErr:  myshows.ErrNotAuthenticated,
	}
	handler := tools.NewSearchHandler(api)

	_, _, err := handler(t.Context(), &mcp.CallToolRequest{},
		tools.SearchParams{Query: "x", WithStatus: true})
	if !errors.Is(err, tools.ErrStatusNeedsAuth) || !errors.Is(err, tools.ErrValidation) {
		t.Fatalf("expected ErrStatusNeedsAuth/ErrValidation, got %v", err)
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
