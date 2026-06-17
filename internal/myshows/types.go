package myshows

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

	Description     string    `json:"description,omitempty"`
	Country         string    `json:"country,omitempty"`
	CountryTitle    string    `json:"countryTitle,omitempty"`
	Started         string    `json:"started,omitempty"`
	Ended           string    `json:"ended,omitempty"`
	KinopoiskRating float64   `json:"kinopoiskRating,omitempty"`
	ImdbID          string    `json:"imdbId,omitempty"`
	ImdbRating      float64   `json:"imdbRating,omitempty"`
	Runtime         int       `json:"runtime,omitempty"`
	GenreIDs        []int     `json:"genreIds,omitempty"`
	Network         *Network  `json:"network,omitempty"`
	Episodes        []Episode `json:"episodes,omitempty"`
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
