package myshows

import "context"

// Profile returns a user's profile. An empty login returns the authenticated
// user's own profile. Requires authentication.
func (c *Client) Profile(ctx context.Context, login string) (*Profile, error) {
	params := map[string]any{}
	if login != "" {
		params[paramLogin] = login
	}

	var profile Profile

	err := c.callAuthed(ctx, "profile.Get", params, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// MyShows returns a user's tracked shows with watch status, rating, and
// progress. An empty login returns the authenticated user's list. Requires
// authentication.
func (c *Client) MyShows(ctx context.Context, login string) ([]ProfileShow, error) {
	params := map[string]any{}
	if login != "" {
		params[paramLogin] = login
	}

	var shows []ProfileShow

	err := c.callAuthed(ctx, "profile.Shows", params, &shows)

	return shows, err
}

// ShowStatuses returns the authenticated user's watch status for the given
// shows -- a lightweight alternative to MyShows when only the status is needed.
// Requires authentication.
func (c *Client) ShowStatuses(ctx context.Context, showIDs []int) ([]ShowStatus, error) {
	var statuses []ShowStatus

	err := c.callAuthed(ctx, "profile.ShowStatuses", map[string]any{paramShowIDs: showIDs}, &statuses)

	return statuses, err
}

// MyEpisodes returns the authenticated user's watched episodes for a show.
// Requires authentication.
func (c *Client) MyEpisodes(ctx context.Context, showID int) ([]WatchedEpisode, error) {
	var episodes []WatchedEpisode

	err := c.callAuthed(ctx, "profile.Episodes", map[string]any{paramShowID: showID}, &episodes)

	return episodes, err
}

// Counters returns the authenticated user's pending-item counts. Requires
// authentication.
func (c *Client) Counters(ctx context.Context) (*Counters, error) {
	var counters Counters

	err := c.callAuthed(ctx, "profile.Counters", map[string]any{}, &counters)
	if err != nil {
		return nil, err
	}

	return &counters, nil
}
