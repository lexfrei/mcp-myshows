package myshows

import "context"

// CheckEpisode marks an episode as watched, optionally recording a rating
// (0-5; 0 leaves it unrated, so the rating param is sent only when positive).
// Requires authentication.
func (c *Client) CheckEpisode(ctx context.Context, episodeID, rating int) error {
	params := map[string]any{paramID: episodeID}
	if rating > 0 {
		params[paramRating] = rating
	}

	return c.callAuthed(ctx, "manage.CheckEpisode", params, nil)
}

// UnCheckEpisode marks a previously watched episode as unwatched. Requires
// authentication.
func (c *Client) UnCheckEpisode(ctx context.Context, episodeID int) error {
	return c.callAuthed(ctx, "manage.UnCheckEpisode", map[string]any{paramID: episodeID}, nil)
}

// SetShowStatus changes the user's tracking status for a show. Status is one of
// watching, later, cancelled, or remove. Requires authentication.
func (c *Client) SetShowStatus(ctx context.Context, showID int, status string) error {
	return c.callAuthed(ctx, "manage.SetShowStatus", map[string]any{
		paramID:     showID,
		paramStatus: status,
	}, nil)
}

// RateShow sets the user's rating (0-5) for a show. Requires authentication.
func (c *Client) RateShow(ctx context.Context, showID, rating int) error {
	return c.callAuthed(ctx, "manage.RateShow", map[string]any{
		paramID:     showID,
		paramRating: rating,
	}, nil)
}

// RateEpisode sets the user's rating (0-5) for an episode. Requires
// authentication.
func (c *Client) RateEpisode(ctx context.Context, episodeID, rating int) error {
	return c.callAuthed(ctx, "manage.RateEpisode", map[string]any{
		paramID:     episodeID,
		paramRating: rating,
	}, nil)
}
