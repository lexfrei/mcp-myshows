package myshows

import "encoding/json"

// flexString is a string that also tolerates a JSON number (or null) on the
// wire. MyShows returns some id fields inconsistently -- e.g. imdbId is
// "tt0903747" for most shows but a bare number like 475784 for others.
type flexString string

// UnmarshalJSON accepts a JSON string, number, or null, normalising to a string.
// Any other token (bool, array, object) is rejected.
func (f *flexString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = ""

		return nil
	}

	if data[0] == '"' {
		var str string

		err := json.Unmarshal(data, &str)
		if err != nil {
			return err //nolint:wrapcheck // surfaced as a decode error by the caller.
		}

		*f = flexString(str)

		return nil
	}

	var num json.Number

	err := json.Unmarshal(data, &num)
	if err != nil {
		return err //nolint:wrapcheck // surfaced as a decode error by the caller.
	}

	*f = flexString(num.String())

	return nil
}

// Show is a trimmed show summary as returned by search, top, lists, profile,
// and recommendation methods. Fields absent from a given response stay zero.
type Show struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	TitleOriginal string  `json:"titleOriginal,omitempty"`
	Status        string  `json:"status,omitempty"`
	Year          int     `json:"year,omitempty"`
	TotalSeasons  int     `json:"totalSeasons,omitempty"`
	Watching      int     `json:"watching,omitempty"`
	Rating        float64 `json:"rating,omitempty"`
	KinopoiskID   int     `json:"kinopoiskId,omitempty"`
	Image         string  `json:"image,omitempty"`
}

// ShowDetails is the full view returned by shows.GetById and
// shows.GetByExternalId. It embeds Show so the shared summary fields are
// promoted to the top level when decoding.
type ShowDetails struct {
	Show

	Description     string     `json:"description,omitempty"`
	Country         string     `json:"country,omitempty"`
	CountryTitle    string     `json:"countryTitle,omitempty"`
	Started         string     `json:"started,omitempty"`
	Ended           string     `json:"ended,omitempty"`
	KinopoiskRating float64    `json:"kinopoiskRating,omitempty"`
	ImdbID          flexString `json:"imdbId,omitempty"`
	ImdbRating      float64    `json:"imdbRating,omitempty"`
	Runtime         int        `json:"runtime,omitempty"`
	GenreIDs        []int      `json:"genreIds,omitempty"`
	Network         *Network   `json:"network,omitempty"`
	Episodes        []Episode  `json:"episodes,omitempty"`
}

// Network is the broadcaster a show airs on.
type Network struct {
	ID      int    `json:"id,omitempty"`
	Title   string `json:"title,omitempty"`
	Country string `json:"country,omitempty"`
}

// Episode is a trimmed episode as returned by shows.Episode, the episodes list
// inside ShowDetails, and the next-to-watch list.
type Episode struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	ShowID        int     `json:"showId,omitempty"`
	SeasonNumber  int     `json:"seasonNumber,omitempty"`
	EpisodeNumber int     `json:"episodeNumber,omitempty"`
	AirDate       string  `json:"airDate,omitempty"`
	ShortName     string  `json:"shortName,omitempty"`
	Rating        float64 `json:"rating,omitempty"`
	IsSpecial     bool    `json:"isSpecial,omitempty"`
}

// Genre is a single genre from shows.Genres.
type Genre struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// RankedShow is one entry from shows.Top: a show with its chart position.
type RankedShow struct {
	Rank  int  `json:"rank"`
	Voted int  `json:"voted"`
	Show  Show `json:"show"`
}

// ProfileShow is one tracked show from profile.Shows, carrying the user's watch
// status, rating, and progress.
type ProfileShow struct {
	Show            Show   `json:"show"`
	WatchStatus     string `json:"watchStatus"`
	Rating          int    `json:"rating"`
	WatchedEpisodes int    `json:"watchedEpisodes"`
	TotalEpisodes   int    `json:"totalEpisodes"`
}

// ShowStatus is one entry from profile.ShowStatuses: a show's watch status for
// the authenticated user, without the full tracked-show payload.
type ShowStatus struct {
	ShowID      int    `json:"showId"`
	WatchStatus string `json:"watchStatus"`
}

// WatchedEpisode is one entry from profile.Episodes: an episode the user has
// marked watched, with the watch date and personal rating.
type WatchedEpisode struct {
	ID        int    `json:"id"`
	WatchDate string `json:"watchDate,omitempty"`
	Rating    int    `json:"rating,omitempty"`
}

// NextEpisode is one entry from lists.Episodes: an episode to watch next,
// paired with the show it belongs to.
type NextEpisode struct {
	Episode Episode `json:"episode"`
	Show    Show    `json:"show"`
}

// Recommendation is one entry from recommendation.Get.
type Recommendation struct {
	Show            Show    `json:"show"`
	Percent         float64 `json:"percent"`
	FriendsWatching int     `json:"friendsWatching"`
}

// Profile is the trimmed result of profile.Get.
type Profile struct {
	User  ProfileUser  `json:"user"`
	Stats ProfileStats `json:"stats"`
}

// ProfileUser holds the public-facing account fields from profile.Get.
type ProfileUser struct {
	Login      string `json:"login"`
	Gender     string `json:"gender,omitempty"`
	IsPro      bool   `json:"isPro,omitempty"`
	WastedTime int    `json:"wastedTime,omitempty"`
}

// ProfileStats holds the watch-time aggregates from profile.Get.
type ProfileStats struct {
	WatchedEpisodes   int     `json:"watchedEpisodes"`
	RemainingEpisodes int     `json:"remainingEpisodes"`
	TotalEpisodes     int     `json:"totalEpisodes"`
	WatchedHours      float64 `json:"watchedHours"`
	WatchedDays       float64 `json:"watchedDays"`
}

// Counters is the result of profile.Counters: a summary of pending items.
type Counters struct {
	UnwatchedEpisodes int `json:"unwatchedEpisodes"`
	UnwatchedMovies   int `json:"unwatchedMovies"`
	NewComments       int `json:"newComments"`
	NewCommentReplies int `json:"newCommentReplies"`
	NewAchievements   int `json:"newAchievements"`
}
