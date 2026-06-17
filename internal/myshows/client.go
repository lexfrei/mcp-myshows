package myshows

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// defaultTimeout bounds a single HTTP request/response cycle.
const defaultTimeout = 30 * time.Second

// defaultUserAgent identifies the client to MyShows when none is configured.
const defaultUserAgent = "Mozilla/5.0 (compatible; mcp-myshows)"

// Fallback endpoints used when Options leaves them empty.
const (
	fallbackAPIURL     = "https://api.myshows.me/v2/rpc/"
	fallbackSessionURL = "https://myshows.me/api/session"
)

// JSON-RPC parameter keys, kept as constants so every param map is built in one
// uniform style.
const (
	paramID           = "id"
	paramQuery        = "query"
	paramShowID       = "showId"
	paramWithEpisodes = "withEpisodes"
	paramSource       = "source"
	paramMode         = "mode"
	paramCount        = "count"
	paramLogin        = "login"
	paramList         = "list"
	paramStatus       = "status"
	paramRating       = "rating"
)

// API is the behaviour the MCP tools depend on. It is satisfied by *Client and
// mocked in tests.
//
//nolint:interfacebloat // one method per MCP tool; the MyShows surface is intentionally broad.
type API interface {
	// Public, read-only methods (no authentication required).
	Search(ctx context.Context, query string) ([]Show, error)
	GetShow(ctx context.Context, showID int, withEpisodes bool) (*ShowDetails, error)
	GetShowByExternal(ctx context.Context, externalID, source string) (*ShowDetails, error)
	GetEpisode(ctx context.Context, episodeID int) (*Episode, error)
	Top(ctx context.Context, mode string, count int) ([]RankedShow, error)
	Genres(ctx context.Context) ([]Genre, error)

	// Account, read-only methods (require authentication).
	Profile(ctx context.Context, login string) (*Profile, error)
	MyShows(ctx context.Context, login string) ([]ProfileShow, error)
	MyEpisodes(ctx context.Context, showID int) ([]WatchedEpisode, error)
	NextEpisodes(ctx context.Context, list string) ([]NextEpisode, error)
	Counters(ctx context.Context) (*Counters, error)
	Recommendations(ctx context.Context, count int) ([]Recommendation, error)

	// Write methods (require authentication, mutate the account).
	CheckEpisode(ctx context.Context, episodeID, rating int) error
	UnCheckEpisode(ctx context.Context, episodeID int) error
	SetShowStatus(ctx context.Context, showID int, status string) error
	RateShow(ctx context.Context, showID, rating int) error
	RateEpisode(ctx context.Context, episodeID, rating int) error
}

// Options configures a Client. Username+Password (or a raw Token) are required
// for the account and write methods; public methods work without either.
type Options struct {
	// APIURL is the JSON-RPC endpoint; empty falls back to the documented v2 API.
	APIURL string
	// SessionURL is the login endpoint; empty falls back to the default.
	SessionURL string
	// Username and Password authenticate via the session endpoint.
	Username string
	Password string
	// Token is a pre-obtained Bearer token used instead of a username login.
	// When set, the client never logs in and never persists a token.
	Token string
	// TokenPath persists the session token between runs (empty disables it).
	TokenPath string
	// UserAgent overrides defaultUserAgent.
	UserAgent string
	// Transport overrides the HTTP round-tripper (e.g. for a proxy).
	Transport http.RoundTripper
}

// Ensure *Client satisfies the API interface.
var _ API = (*Client)(nil)

// Client is the concrete myshows.API backed by net/http.
type Client struct {
	apiURL     string
	sessionURL string
	http       Doer
	username   string
	password   string
	tokenPath  string
	userAgent  string

	authMu sync.Mutex   // serialises logins
	mu     sync.RWMutex // guards token
	token  string
}

// New builds a Client from opts. It seeds the token from the explicit override
// or the on-disk cache, but defers the actual login until the first authed call.
func New(opts *Options) (*Client, error) {
	if opts == nil {
		opts = &Options{}
	}

	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	client := &Client{
		apiURL:     orDefault(opts.APIURL, fallbackAPIURL),
		sessionURL: orDefault(opts.SessionURL, fallbackSessionURL),
		http: &http.Client{
			Timeout:   defaultTimeout,
			Transport: opts.Transport,
		},
		username:  opts.Username,
		password:  opts.Password,
		tokenPath: opts.TokenPath,
		userAgent: userAgent,
	}

	if opts.Token != "" {
		client.token = opts.Token
	} else {
		client.loadToken()
	}

	return client, nil
}

// orDefault returns value, or fallback when value is empty.
func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

// currentToken returns the active Bearer token, or "" when unauthenticated.
func (c *Client) currentToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.token
}

// setToken replaces the active Bearer token.
func (c *Client) setToken(token string) {
	c.mu.Lock()
	c.token = token
	c.mu.Unlock()
}
