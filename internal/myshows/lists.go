package myshows

import "context"

// NextEpisodes returns episodes from a personal episode list. List is one of
// unwatched or next. Requires authentication.
func (c *Client) NextEpisodes(ctx context.Context, list string) ([]NextEpisode, error) {
	var episodes []NextEpisode

	err := c.callAuthed(ctx, "lists.Episodes", map[string]any{paramList: list}, &episodes)

	return episodes, err
}
