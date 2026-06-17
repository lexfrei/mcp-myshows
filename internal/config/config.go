// Package config loads the mcp-myshows configuration from environment
// variables.
package config

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/errors"
)

const maxPort = 65535

// Default endpoints. All RPC traffic goes to the documented v2 JSON-RPC API;
// only the session login hits myshows.me directly.
const (
	// DefaultAPIURL is the documented MyShows JSON-RPC v2 endpoint.
	DefaultAPIURL = "https://api.myshows.me/v2/rpc/"
	// DefaultSessionURL is the login endpoint that exchanges username/password
	// for a Bearer token without an OAuth AppID.
	DefaultSessionURL = "https://myshows.me/api/session"
)

// ErrInvalidHTTPPort is returned when MCP_HTTP_PORT is not a valid port number.
var ErrInvalidHTTPPort = errors.New("MCP_HTTP_PORT must be a valid port number (1-65535)")

// ErrInvalidProxy is returned when MYSHOWS_PROXY is not a valid URL.
var ErrInvalidProxy = errors.New("MYSHOWS_PROXY must be a valid proxy URL")

// ErrInsecureHTTP is returned when the HTTP transport would bind to a
// non-loopback interface without an MCP_HTTP_TOKEN to authenticate requests.
var ErrInsecureHTTP = errors.New(
	"refusing to expose the unauthenticated HTTP transport on a non-loopback host; " +
		"set MCP_HTTP_TOKEN or bind MCP_HTTP_HOST to a loopback address")

// Config holds the application configuration loaded from environment variables.
type Config struct {
	// Username and Password authenticate against the MyShows session endpoint.
	Username string
	Password string
	// Token is a pre-obtained Bearer token used instead of a username login.
	Token string
	// TokenFile persists the session token between runs.
	TokenFile string
	// APIURL is the JSON-RPC endpoint; empty falls back to DefaultAPIURL.
	APIURL string
	// SessionURL is the login endpoint; empty falls back to DefaultSessionURL.
	SessionURL string
	// UserAgent overrides the default browser User-Agent.
	UserAgent string
	// Proxy is an optional HTTP/SOCKS5 proxy URL.
	Proxy string
	// HTTPPort and HTTPHost configure the optional HTTP transport.
	HTTPPort string
	HTTPHost string
	// HTTPToken, when set, is the Bearer token required on every HTTP request.
	HTTPToken string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	httpPort := os.Getenv("MCP_HTTP_PORT")
	if httpPort != "" {
		port, err := strconv.Atoi(httpPort)
		if err != nil || port < 1 || port > maxPort {
			return nil, ErrInvalidHTTPPort
		}
	}

	proxy := os.Getenv("MYSHOWS_PROXY")
	if proxy != "" {
		_, err := parseProxy(proxy)
		if err != nil {
			return nil, err
		}
	}

	httpHost := os.Getenv("MCP_HTTP_HOST")
	if httpHost == "" {
		httpHost = "127.0.0.1"
	}

	return &Config{
		Username:   os.Getenv("MYSHOWS_USERNAME"),
		Password:   os.Getenv("MYSHOWS_PASSWORD"),
		Token:      os.Getenv("MYSHOWS_TOKEN"),
		TokenFile:  resolveTokenFile(os.Getenv("MYSHOWS_TOKEN_FILE")),
		APIURL:     envOrDefault("MYSHOWS_API_URL", DefaultAPIURL),
		SessionURL: envOrDefault("MYSHOWS_SESSION_URL", DefaultSessionURL),
		UserAgent:  os.Getenv("MYSHOWS_USER_AGENT"),
		Proxy:      proxy,
		HTTPPort:   httpPort,
		HTTPHost:   httpHost,
		HTTPToken:  os.Getenv("MCP_HTTP_TOKEN"),
	}, nil
}

// envOrDefault returns the environment value for key, or fallback when unset.
func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// resolveTokenFile returns the configured token file, defaulting to
// ~/.mcp-myshows/token.json when unset and a home directory is available.
func resolveTokenFile(configured string) string {
	if configured != "" {
		return configured
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".mcp-myshows", "token.json")
}

// HasAuth reports whether any authentication method is configured. Public,
// read-only tools work without it; account and write tools require it.
func (c *Config) HasAuth() bool {
	return c.Token != "" || (c.Username != "" && c.Password != "")
}

// ProxyTransport builds an HTTP round-tripper honouring the configured proxy,
// or returns nil when no proxy is set.
func (c *Config) ProxyTransport() (http.RoundTripper, error) {
	if c.Proxy == "" {
		return nil, nil //nolint:nilnil // no proxy configured means no custom transport.
	}

	proxyURL, err := parseProxy(c.Proxy)
	if err != nil {
		return nil, err
	}

	// Clone the default transport so HTTP/2, connection pooling, and the
	// dial/TLS-handshake timeouts are preserved; only the proxy is overridden.
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Transport{Proxy: http.ProxyURL(proxyURL)}, nil
	}

	cloned := transport.Clone()
	cloned.Proxy = http.ProxyURL(proxyURL)

	return cloned, nil
}

// parseProxy validates and parses a proxy URL, requiring a scheme and host.
func parseProxy(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.Wrap(ErrInvalidProxy, raw)
	}

	return parsed, nil
}

// HTTPEnabled reports whether the HTTP transport should be started.
func (c *Config) HTTPEnabled() bool {
	return c.HTTPPort != ""
}

// ValidateHTTP guards the unauthenticated HTTP transport. The MCP server has no
// per-request auth of its own and exposes account write tools, so binding it to
// a non-loopback interface without a token would hand the account to anyone who
// can reach the port. It is allowed only on a loopback host or with a token set.
func (c *Config) ValidateHTTP() error {
	if !c.HTTPEnabled() || c.HTTPToken != "" || isLoopbackHost(c.HTTPHost) {
		return nil
	}

	return errors.Wrapf(ErrInsecureHTTP, "host %q", c.HTTPHost)
}

// isLoopbackHost reports whether host is the loopback interface.
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}

	parsed := net.ParseIP(host)

	return parsed != nil && parsed.IsLoopback()
}

// HTTPAddr returns the host:port address for the HTTP server.
func (c *Config) HTTPAddr() string {
	return net.JoinHostPort(c.HTTPHost, c.HTTPPort)
}
