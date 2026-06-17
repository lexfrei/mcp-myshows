package config_test

import (
	"testing"

	"github.com/cockroachdb/errors"

	"github.com/lexfrei/mcp-myshows/internal/config"
)

const (
	testHTTPPort   = "8080"
	testPublicHost = "0.0.0.0"
)

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.APIURL != config.DefaultAPIURL {
		t.Errorf("APIURL = %q, want %q", cfg.APIURL, config.DefaultAPIURL)
	}

	if cfg.SessionURL != config.DefaultSessionURL {
		t.Errorf("SessionURL = %q, want %q", cfg.SessionURL, config.DefaultSessionURL)
	}

	if cfg.HTTPHost != "127.0.0.1" {
		t.Errorf("HTTPHost = %q, want 127.0.0.1", cfg.HTTPHost)
	}

	if cfg.HTTPEnabled() {
		t.Error("HTTPEnabled() = true, want false when MCP_HTTP_PORT unset")
	}
}

func TestLoad_OverridesEndpoints(t *testing.T) {
	clearEnv(t)
	t.Setenv("MYSHOWS_API_URL", "https://example.test/rpc/")
	t.Setenv("MYSHOWS_SESSION_URL", "https://example.test/session")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.APIURL != "https://example.test/rpc/" {
		t.Errorf("APIURL = %q", cfg.APIURL)
	}

	if cfg.SessionURL != "https://example.test/session" {
		t.Errorf("SessionURL = %q", cfg.SessionURL)
	}
}

func TestHasAuth(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{name: "nothing", cfg: config.Config{}, want: false},
		{name: "token only", cfg: config.Config{Token: "abc"}, want: true},
		{name: "username only", cfg: config.Config{Username: "u"}, want: false},
		{name: "password only", cfg: config.Config{Password: "p"}, want: false},
		{name: "username and password", cfg: config.Config{Username: "u", Password: "p"}, want: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.cfg.HasAuth(); got != testCase.want {
				t.Errorf("HasAuth() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestLoad_InvalidHTTPPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("MCP_HTTP_PORT", "notaport")

	_, err := config.Load()
	if !errors.Is(err, config.ErrInvalidHTTPPort) {
		t.Fatalf("err = %v, want ErrInvalidHTTPPort", err)
	}
}

func TestLoad_InvalidProxy(t *testing.T) {
	clearEnv(t)
	t.Setenv("MYSHOWS_PROXY", "://missing-scheme")

	_, err := config.Load()
	if !errors.Is(err, config.ErrInvalidProxy) {
		t.Fatalf("err = %v, want ErrInvalidProxy", err)
	}
}

func TestHTTPAddr(t *testing.T) {
	cfg := config.Config{HTTPHost: testPublicHost, HTTPPort: testHTTPPort}

	if got := cfg.HTTPAddr(); got != "0.0.0.0:8080" {
		t.Errorf("HTTPAddr() = %q, want 0.0.0.0:8080", got)
	}
}

func TestLoad_TokenFileOverride(t *testing.T) {
	clearEnv(t)
	t.Setenv("MYSHOWS_TOKEN_FILE", "/tmp/custom-token.json")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.TokenFile != "/tmp/custom-token.json" {
		t.Errorf("TokenFile = %q, want /tmp/custom-token.json", cfg.TokenFile)
	}
}

func TestValidateHTTP(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{name: "disabled", cfg: config.Config{}, wantErr: false},
		{name: "loopback no token", cfg: config.Config{HTTPPort: testHTTPPort, HTTPHost: "127.0.0.1"}, wantErr: false},
		{name: "localhost no token", cfg: config.Config{HTTPPort: testHTTPPort, HTTPHost: "localhost"}, wantErr: false},
		{name: "ipv6 loopback no token", cfg: config.Config{HTTPPort: testHTTPPort, HTTPHost: "::1"}, wantErr: false},
		{name: "public no token", cfg: config.Config{HTTPPort: testHTTPPort, HTTPHost: testPublicHost}, wantErr: true},
		{name: "hostname no token", cfg: config.Config{HTTPPort: testHTTPPort, HTTPHost: "example.test"}, wantErr: true},
		{
			name:    "public with token",
			cfg:     config.Config{HTTPPort: testHTTPPort, HTTPHost: testPublicHost, HTTPToken: "secret"},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.cfg.ValidateHTTP()
			if testCase.wantErr && !errors.Is(err, config.ErrInsecureHTTP) {
				t.Errorf("err = %v, want ErrInsecureHTTP", err)
			}

			if !testCase.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// clearEnv unsets every MyShows and MCP environment variable so each test
// starts from a known baseline regardless of the host environment.
func clearEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"MYSHOWS_USERNAME", "MYSHOWS_PASSWORD", "MYSHOWS_TOKEN", "MYSHOWS_TOKEN_FILE",
		"MYSHOWS_API_URL", "MYSHOWS_SESSION_URL", "MYSHOWS_USER_AGENT", "MYSHOWS_PROXY",
		"MCP_HTTP_PORT", "MCP_HTTP_HOST", "MCP_HTTP_TOKEN",
	} {
		t.Setenv(key, "")
	}
}
