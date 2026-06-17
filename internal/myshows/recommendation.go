package myshows

import "context"

// Recommendations returns personalised show suggestions. Count caps the result.
// Requires authentication.
func (c *Client) Recommendations(ctx context.Context, count int) ([]Recommendation, error) {
	params := map[string]any{}
	if count > 0 {
		params[paramCount] = count
	}

	var recommendations []Recommendation

	err := c.callAuthed(ctx, "recommendation.Get", params, &recommendations)

	return recommendations, err
}
