package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// listToolNames registers tools for the given auth mode and returns the names
// the server advertises over an in-memory transport.
func listToolNames(t *testing.T, hasAuth bool) map[string]bool {
	t.Helper()

	client, err := myshows.New(&myshows.Options{})
	if err != nil {
		t.Fatalf("myshows.New: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{Name: serverName, Version: "test"}, newServerOptions(testLogger(), hasAuth))
	registerTools(server, client, hasAuth)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer func() { _ = serverSession.Close() }()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)

	clientSession, err := mcpClient.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = clientSession.Close() }()

	result, err := clientSession.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	names := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}

	return names
}

func TestRegisterTools_PublicOnly(t *testing.T) {
	t.Parallel()

	names := listToolNames(t, false)

	public := []string{
		"myshows_server_version",
		"myshows_search",
		"myshows_show",
		"myshows_show_by_external",
		"myshows_episode",
		"myshows_top",
		"myshows_genres",
	}

	for _, name := range public {
		if !names[name] {
			t.Errorf("public tool %q not registered", name)
		}
	}

	if len(names) != len(public) {
		t.Errorf("registered %d tools, want %d (auth tools must stay hidden)", len(names), len(public))
	}

	if names["myshows_counters"] {
		t.Error("auth tool myshows_counters registered without credentials")
	}
}

func TestRegisterTools_WithAuth(t *testing.T) {
	t.Parallel()

	names := listToolNames(t, true)

	want := []string{
		"myshows_server_version", "myshows_search", "myshows_show", "myshows_show_by_external",
		"myshows_episode", "myshows_top", "myshows_genres",
		"myshows_profile", "myshows_my_shows", "myshows_my_episodes", "myshows_unwatched",
		"myshows_counters", "myshows_recommendations",
		"myshows_check_episode", "myshows_uncheck_episode", "myshows_set_show_status",
		"myshows_rate_show", "myshows_rate_episode",
	}

	for _, name := range want {
		if !names[name] {
			t.Errorf("tool %q not registered", name)
		}
	}

	if len(names) != len(want) {
		t.Errorf("registered %d tools, want %d", len(names), len(want))
	}
}

func TestBearerAuth(t *testing.T) {
	t.Parallel()

	okHandler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	const token = "secret"

	tests := []struct {
		name   string
		token  string
		header string
		want   int
	}{
		{name: "no token disables the check", token: "", header: "", want: http.StatusOK},
		{name: "valid token", token: token, header: "Bearer " + token, want: http.StatusOK},
		{name: "missing header", token: token, header: "", want: http.StatusUnauthorized},
		{name: "wrong token", token: token, header: "Bearer nope", want: http.StatusUnauthorized},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			handler := bearerAuth(okHandler, testCase.token)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

			if testCase.header != "" {
				req.Header.Set("Authorization", testCase.header)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != testCase.want {
				t.Errorf("status = %d, want %d", rec.Code, testCase.want)
			}
		})
	}
}
