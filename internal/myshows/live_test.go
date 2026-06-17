//go:build integration

package myshows_test

import (
	"os"
	"testing"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

// TestLive_Search exercises the public search path against the real API. It
// needs no credentials.
//
// Run with:
//
//	go test -tags integration -run TestLive_Search -count=1 ./internal/myshows/
func TestLive_Search(t *testing.T) {
	client, err := myshows.New(&myshows.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	shows, err := client.Search(t.Context(), "breaking bad")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(shows) == 0 {
		t.Fatal("Search returned no results")
	}

	t.Logf("first result: %d %q (%d)", shows[0].ID, shows[0].Title, shows[0].Year)
}

// TestLive_Account exercises the authenticated read path against the real API.
// It is skipped unless MYSHOWS_USERNAME and MYSHOWS_PASSWORD are set.
//
// Run with:
//
//	MYSHOWS_USERNAME=... MYSHOWS_PASSWORD=... \
//	  go test -tags integration -run TestLive_Account -count=1 ./internal/myshows/
func TestLive_Account(t *testing.T) {
	username := os.Getenv("MYSHOWS_USERNAME")
	password := os.Getenv("MYSHOWS_PASSWORD")

	if username == "" || password == "" {
		t.Skip("set MYSHOWS_USERNAME and MYSHOWS_PASSWORD to run the authenticated live test")
	}

	client, err := myshows.New(&myshows.Options{
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	counters, err := client.Counters(t.Context())
	if err != nil {
		t.Fatalf("Counters: %v", err)
	}

	t.Logf("unwatched episodes: %d", counters.UnwatchedEpisodes)

	shows, err := client.MyShows(t.Context(), "")
	if err != nil {
		t.Fatalf("MyShows: %v", err)
	}

	t.Logf("tracked shows: %d", len(shows))

	// Pin the nested response shape: the unit tests mock at the API boundary, so
	// a wrong JSON tag on ProfileShow.Show would slip past CI. Here a populated
	// nested show proves the real wire format still matches the struct.
	if len(shows) > 0 {
		first := shows[0]
		if first.Show.ID == 0 || first.Show.Title == "" {
			t.Errorf("nested show not populated: %+v", first)
		}

		t.Logf("first tracked show: %d %q (status %q)", first.Show.ID, first.Show.Title, first.WatchStatus)
	}
}
