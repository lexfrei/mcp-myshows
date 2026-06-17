package myshows

import "context"

// Search returns shows matching a free-text query. No authentication required.
func (c *Client) Search(ctx context.Context, query string) ([]Show, error) {
	var shows []Show

	err := c.callPublic(ctx, "shows.Search", map[string]any{paramQuery: query}, &shows)

	return shows, err
}

// GetShow returns the full details for a show, optionally including its episode
// list. No authentication required.
func (c *Client) GetShow(ctx context.Context, showID int, withEpisodes bool) (*ShowDetails, error) {
	var show ShowDetails

	err := c.callPublic(ctx, "shows.GetById", map[string]any{
		paramShowID:       showID,
		paramWithEpisodes: withEpisodes,
	}, &show)
	if err != nil {
		return nil, err
	}

	return &show, nil
}

// GetShowByExternal looks a show up by an external identifier. Source is one of
// imdb, kinopoisk, or thetvdb. No authentication required.
func (c *Client) GetShowByExternal(ctx context.Context, externalID, source string) (*ShowDetails, error) {
	var show ShowDetails

	err := c.callPublic(ctx, "shows.GetByExternalId", map[string]any{
		paramID:     externalID,
		paramSource: source,
	}, &show)
	if err != nil {
		return nil, err
	}

	return &show, nil
}

// GetEpisode returns the details of a single episode. No authentication required.
func (c *Client) GetEpisode(ctx context.Context, episodeID int) (*Episode, error) {
	var episode Episode

	err := c.callPublic(ctx, "shows.Episode", map[string]any{paramID: episodeID}, &episode)
	if err != nil {
		return nil, err
	}

	return &episode, nil
}

// Top returns the ranked chart of shows. Mode selects the chart (e.g. all,
// year, month); count caps the result. No authentication required.
func (c *Client) Top(ctx context.Context, mode string, count int) ([]RankedShow, error) {
	params := map[string]any{}
	if mode != "" {
		params[paramMode] = mode
	}

	if count > 0 {
		params[paramCount] = count
	}

	var ranked []RankedShow

	err := c.callPublic(ctx, "shows.Top", params, &ranked)

	return ranked, err
}

// Genres returns the full list of genres. No authentication required.
func (c *Client) Genres(ctx context.Context) ([]Genre, error) {
	var genres []Genre

	err := c.callPublic(ctx, "shows.Genres", map[string]any{}, &genres)

	return genres, err
}
