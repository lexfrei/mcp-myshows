package myshows_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cockroachdb/errors"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

const (
	testUser = "user"
	testPass = "pass"
)

// rpcHandler is a configurable JSON-RPC test endpoint.
type rpcHandler struct {
	t            *testing.T
	result       string // raw JSON for the "result" field
	rpcError     string // raw JSON for the "error" field (overrides result)
	lastAuth     string
	rpcCalls     atomic.Int32
	sessionCalls atomic.Int32
	token        string // token handed out by the session endpoint
}

func (h *rpcHandler) handle() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/session":
			h.sessionCalls.Add(1)
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"token":"` + h.token + `","refreshToken":"refresh"}`))
		case "/rpc":
			h.rpcCalls.Add(1)
			h.lastAuth = req.Header.Get("Authorization")
			rw.Header().Set("Content-Type", "application/json")

			if h.rpcError != "" {
				_, _ = rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":` + h.rpcError + `}`))

				return
			}

			_, _ = rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":` + h.result + `}`))
		default:
			http.NotFound(rw, req)
		}
	}
}

func newServer(t *testing.T, handler *rpcHandler) *httptest.Server {
	t.Helper()
	handler.t = t
	server := httptest.NewServer(handler.handle())
	t.Cleanup(server.Close)

	return server
}

func TestSearch_NoAuthHeader(t *testing.T) {
	handler := &rpcHandler{result: `[{"id":187,"title":"Breaking Bad"}]`}
	server := newServer(t, handler)

	client, err := myshows.New(&myshows.Options{APIURL: server.URL + "/rpc"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	shows, err := client.Search(t.Context(), "breaking bad")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(shows) != 1 || shows[0].ID != 187 {
		t.Fatalf("unexpected results: %+v", shows)
	}

	if handler.lastAuth != "" {
		t.Errorf("public call sent Authorization header: %q", handler.lastAuth)
	}
}

func TestCounters_LogsInAndAttachesBearer(t *testing.T) {
	handler := &rpcHandler{
		token:  "session-token",
		result: `{"unwatchedEpisodes":593}`,
	}
	server := newServer(t, handler)

	client, err := myshows.New(&myshows.Options{
		APIURL:     server.URL + "/rpc",
		SessionURL: server.URL + "/session",
		Username:   testUser,
		Password:   testPass,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	counters, err := client.Counters(t.Context())
	if err != nil {
		t.Fatalf("Counters: %v", err)
	}

	if counters.UnwatchedEpisodes != 593 {
		t.Errorf("UnwatchedEpisodes = %d, want 593", counters.UnwatchedEpisodes)
	}

	if handler.sessionCalls.Load() != 1 {
		t.Errorf("sessionCalls = %d, want 1", handler.sessionCalls.Load())
	}

	if handler.lastAuth != "Bearer session-token" {
		t.Errorf("Authorization = %q, want Bearer session-token", handler.lastAuth)
	}
}

func TestCounters_ReauthOn401(t *testing.T) {
	handler := &reauthHandler{token: "fresh"}
	server := httptest.NewServer(handler.handle())
	t.Cleanup(server.Close)

	tokenFile := filepath.Join(t.TempDir(), "token.json")
	writeFile(t, tokenFile, `{"token":"stale"}`)

	client, err := myshows.New(&myshows.Options{
		APIURL:     server.URL + "/rpc",
		SessionURL: server.URL + "/session",
		Username:   testUser,
		Password:   testPass,
		TokenPath:  tokenFile,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.Counters(t.Context())
	if err != nil {
		t.Fatalf("Counters: %v", err)
	}

	if handler.rpcCalls.Load() != 2 {
		t.Errorf("rpcCalls = %d, want 2 (one rejected, one after re-login)", handler.rpcCalls.Load())
	}
}

func TestCounters_LoginFailure(t *testing.T) {
	handler := &rpcHandler{token: ""} // session returns no token
	server := newServer(t, handler)

	client, err := myshows.New(&myshows.Options{
		APIURL:     server.URL + "/rpc",
		SessionURL: server.URL + "/session",
		Username:   testUser,
		Password:   testPass,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.Counters(t.Context())
	if !errors.Is(err, myshows.ErrLoginFailed) {
		t.Fatalf("err = %v, want ErrLoginFailed", err)
	}
}

func TestCounters_TokenOverrideCannotReauth(t *testing.T) {
	handler := &rpcHandler{rpcError: `{"code":401,"message":"Internal error","data":"Unauthorized"}`}
	server := newServer(t, handler)

	client, err := myshows.New(&myshows.Options{
		APIURL: server.URL + "/rpc",
		Token:  "raw-token",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.Counters(t.Context())
	if !errors.Is(err, myshows.ErrNotAuthenticated) {
		t.Fatalf("err = %v, want ErrNotAuthenticated", err)
	}
}

func TestSearch_APIErrorMapping(t *testing.T) {
	handler := &rpcHandler{rpcError: `{"code":-32601,"message":"Method not found"}`}
	server := newServer(t, handler)

	client, err := myshows.New(&myshows.Options{APIURL: server.URL + "/rpc"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.Search(t.Context(), "x")
	if !errors.Is(err, myshows.ErrAPI) {
		t.Fatalf("err = %v, want ErrAPI", err)
	}
}

func TestLogin_PersistsToken(t *testing.T) {
	handler := &rpcHandler{token: "persisted-token", result: `{"unwatchedEpisodes":1}`}
	server := newServer(t, handler)

	tokenFile := filepath.Join(t.TempDir(), "token.json")

	client, err := myshows.New(&myshows.Options{
		APIURL:     server.URL + "/rpc",
		SessionURL: server.URL + "/session",
		Username:   testUser,
		Password:   testPass,
		TokenPath:  tokenFile,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.Counters(t.Context())
	if err != nil {
		t.Fatalf("Counters: %v", err)
	}

	data, readErr := os.ReadFile(tokenFile)
	if readErr != nil {
		t.Fatalf("read token file: %v", readErr)
	}

	var stored struct {
		Token string `json:"token"`
	}

	jsonErr := json.Unmarshal(data, &stored)
	if jsonErr != nil {
		t.Fatalf("unmarshal token file: %v", jsonErr)
	}

	if stored.Token != "persisted-token" {
		t.Errorf("persisted token = %q, want persisted-token", stored.Token)
	}
}

func TestCheckEpisode_RatingBody(t *testing.T) {
	var params []map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var envelope struct {
			Params map[string]any `json:"params"`
		}

		_ = json.NewDecoder(req.Body).Decode(&envelope)
		params = append(params, envelope.Params)
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":true}`))
	}))
	t.Cleanup(server.Close)

	client, err := myshows.New(&myshows.Options{APIURL: server.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	unratedErr := client.CheckEpisode(t.Context(), 99, 0)
	if unratedErr != nil {
		t.Fatalf("CheckEpisode unrated: %v", unratedErr)
	}

	ratedErr := client.CheckEpisode(t.Context(), 99, 3)
	if ratedErr != nil {
		t.Fatalf("CheckEpisode rated: %v", ratedErr)
	}

	if len(params) != 2 {
		t.Fatalf("got %d requests, want 2", len(params))
	}

	if _, ok := params[0]["rating"]; ok {
		t.Errorf("unrated check sent a rating param: %v", params[0])
	}

	if params[1]["rating"] != float64(3) {
		t.Errorf("rated check rating = %v, want 3", params[1]["rating"])
	}
}

func TestShowDetails_ImdbIDStringOrNumber(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "string", raw: `{"imdbId":"tt0903747"}`, want: "tt0903747"},
		{name: "number", raw: `{"imdbId":475784}`, want: "475784"},
		{name: "null", raw: `{"imdbId":null}`, want: ""},
		{name: "absent", raw: `{}`, want: ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var details myshows.ShowDetails

			err := json.Unmarshal([]byte(testCase.raw), &details)
			if err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if got := string(details.ImdbID); got != testCase.want {
				t.Errorf("ImdbID = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestShowDetails_ImdbIDRejectsNonScalar(t *testing.T) {
	for _, raw := range []string{`{"imdbId":true}`, `{"imdbId":[1,2]}`, `{"imdbId":{"x":1}}`} {
		var details myshows.ShowDetails

		err := json.Unmarshal([]byte(raw), &details)
		if err == nil {
			t.Errorf("%s: expected an unmarshal error, got none (ImdbID=%q)", raw, details.ImdbID)
		}
	}
}

func TestShowStatuses(t *testing.T) {
	var requested map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var envelope struct {
			Params map[string]any `json:"params"`
		}

		_ = json.NewDecoder(req.Body).Decode(&envelope)
		requested = envelope.Params
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"showId":45534,"watchStatus":"later"},{"showId":187,"watchStatus":"finished"}]}`))
	}))
	t.Cleanup(server.Close)

	client, err := myshows.New(&myshows.Options{APIURL: server.URL, Token: "tok"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	statuses, err := client.ShowStatuses(t.Context(), []int{45534, 187})
	if err != nil {
		t.Fatalf("ShowStatuses: %v", err)
	}

	if len(statuses) != 2 || statuses[0].ShowID != 45534 || statuses[0].WatchStatus != "later" {
		t.Fatalf("unexpected statuses: %+v", statuses)
	}

	if _, ok := requested["showIds"]; !ok {
		t.Errorf("request missing showIds param: %v", requested)
	}
}

func TestSearch_ForwardsConfiguredToken(t *testing.T) {
	handler := &rpcHandler{result: `[]`}
	server := newServer(t, handler)

	client, err := myshows.New(&myshows.Options{APIURL: server.URL + "/rpc", Token: "configured-token"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, searchErr := client.Search(t.Context(), "x")
	if searchErr != nil {
		t.Fatalf("Search: %v", searchErr)
	}

	if handler.lastAuth != "Bearer configured-token" {
		t.Errorf("Authorization = %q, want Bearer configured-token", handler.lastAuth)
	}
}

func TestReauth_ConcurrentRefreshLogsInOnce(t *testing.T) {
	const callers = 8

	handler := &reauthHandler{token: "fresh"}
	server := httptest.NewServer(handler.handle())
	t.Cleanup(server.Close)

	tokenFile := filepath.Join(t.TempDir(), "token.json")
	writeFile(t, tokenFile, `{"token":"stale"}`)

	client, err := myshows.New(&myshows.Options{
		APIURL:     server.URL + "/rpc",
		SessionURL: server.URL + "/session",
		Username:   testUser,
		Password:   testPass,
		TokenPath:  tokenFile,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var wg sync.WaitGroup

	for range callers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, _ = client.Counters(t.Context())
		}()
	}

	wg.Wait()

	if got := handler.sessionCalls.Load(); got != 1 {
		t.Errorf("sessionCalls = %d, want 1 (concurrent 401s should trigger one re-login)", got)
	}
}

// reauthHandler rejects the stale token, then accepts the fresh one.
type reauthHandler struct {
	token        string
	rpcCalls     atomic.Int32
	sessionCalls atomic.Int32
}

func (h *reauthHandler) handle() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		switch req.URL.Path {
		case "/session":
			h.sessionCalls.Add(1)
			_, _ = rw.Write([]byte(`{"token":"` + h.token + `"}`))
		case "/rpc":
			h.rpcCalls.Add(1)
			if req.Header.Get("Authorization") != "Bearer "+h.token {
				rw.WriteHeader(http.StatusUnauthorized)

				return
			}

			_, _ = rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"unwatchedEpisodes":1}}`))
		default:
			http.NotFound(rw, req)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
