package tools_test

import (
	"context"

	"github.com/lexfrei/mcp-myshows/internal/myshows"
)

// mockAPI is a configurable myshows.API for handler tests. Unset returns are the
// zero value; write methods capture their arguments.
type mockAPI struct {
	searchResult []myshows.Show
	searchErr    error

	showResult *myshows.ShowDetails
	showErr    error

	episodeResult *myshows.Episode
	episodeErr    error

	countersResult *myshows.Counters
	countersErr    error

	nextResult []myshows.NextEpisode
	nextErr    error
	lastList   string

	lastMode  string
	lastLogin string

	myShowsResult  []myshows.ProfileShow
	myShowsErr     error
	statusesResult []myshows.ShowStatus
	statusesErr    error
	lastShowIDs    []int

	checkID     int
	checkRating int
	checkErr    error

	statusID    int
	statusValue string
	statusErr   error

	rateID     int
	rateRating int
	rateErr    error
}

func (m *mockAPI) Search(_ context.Context, _ string) ([]myshows.Show, error) {
	return m.searchResult, m.searchErr
}

func (m *mockAPI) GetShow(_ context.Context, _ int, _ bool) (*myshows.ShowDetails, error) {
	return m.showResult, m.showErr
}

func (m *mockAPI) GetShowByExternal(_ context.Context, _, _ string) (*myshows.ShowDetails, error) {
	return m.showResult, m.showErr
}

func (m *mockAPI) GetEpisode(_ context.Context, _ int) (*myshows.Episode, error) {
	return m.episodeResult, m.episodeErr
}

func (m *mockAPI) Top(_ context.Context, mode string, _ int) ([]myshows.RankedShow, error) {
	m.lastMode = mode

	return nil, nil
}

func (m *mockAPI) Genres(_ context.Context) ([]myshows.Genre, error) {
	return nil, nil
}

func (m *mockAPI) Profile(_ context.Context, _ string) (*myshows.Profile, error) {
	return &myshows.Profile{}, nil
}

func (m *mockAPI) MyShows(_ context.Context, login string) ([]myshows.ProfileShow, error) {
	m.lastLogin = login

	return m.myShowsResult, m.myShowsErr
}

func (m *mockAPI) ShowStatuses(_ context.Context, showIDs []int) ([]myshows.ShowStatus, error) {
	m.lastShowIDs = showIDs

	return m.statusesResult, m.statusesErr
}

func (m *mockAPI) MyEpisodes(_ context.Context, _ int) ([]myshows.WatchedEpisode, error) {
	return nil, nil
}

func (m *mockAPI) NextEpisodes(_ context.Context, list string) ([]myshows.NextEpisode, error) {
	m.lastList = list

	return m.nextResult, m.nextErr
}

func (m *mockAPI) Counters(_ context.Context) (*myshows.Counters, error) {
	return m.countersResult, m.countersErr
}

func (m *mockAPI) Recommendations(_ context.Context, _ int) ([]myshows.Recommendation, error) {
	return nil, nil
}

func (m *mockAPI) CheckEpisode(_ context.Context, episodeID, rating int) error {
	m.checkID = episodeID
	m.checkRating = rating

	return m.checkErr
}

func (m *mockAPI) UnCheckEpisode(_ context.Context, _ int) error {
	return nil
}

func (m *mockAPI) SetShowStatus(_ context.Context, showID int, status string) error {
	m.statusID = showID
	m.statusValue = status

	return m.statusErr
}

func (m *mockAPI) RateShow(_ context.Context, showID, rating int) error {
	m.rateID = showID
	m.rateRating = rating

	return m.rateErr
}

func (m *mockAPI) RateEpisode(_ context.Context, _, _ int) error {
	return nil
}
