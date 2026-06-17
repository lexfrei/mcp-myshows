// Command mcp-myshows is an MCP server exposing MyShows.me show search and
// personal-tracker tools over stdio and, optionally, HTTP.
package main

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/sync/errgroup"

	"github.com/lexfrei/mcp-myshows/internal/config"
	"github.com/lexfrei/mcp-myshows/internal/myshows"
	"github.com/lexfrei/mcp-myshows/internal/tools"
)

const (
	serverName        = "mcp-myshows"
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 5 * time.Second
)

// version and revision are set via ldflags at build time.
var (
	version  = "dev"
	revision = "unknown"
)

func main() {
	logger := newLogger()

	err := run(logger)
	if err != nil {
		logger.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}

// newLogger builds the structured JSON logger. Logs go to stderr because stdout
// carries the JSON-RPC stream.
func newLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func run(logger *slog.Logger) error {
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		return errors.Wrap(cfgErr, "invalid configuration")
	}

	httpErr := cfg.ValidateHTTP()
	if httpErr != nil {
		return errors.Wrap(httpErr, "invalid HTTP configuration")
	}

	transport, transportErr := cfg.ProxyTransport()
	if transportErr != nil {
		return errors.Wrap(transportErr, "invalid proxy configuration")
	}

	client, clientErr := myshows.New(&myshows.Options{
		APIURL:     cfg.APIURL,
		SessionURL: cfg.SessionURL,
		Username:   cfg.Username,
		Password:   cfg.Password,
		Token:      cfg.Token,
		TokenPath:  cfg.TokenFile,
		UserAgent:  cfg.UserAgent,
		Transport:  transport,
	})
	if clientErr != nil {
		return errors.Wrap(clientErr, "failed to create myshows client")
	}

	hasAuth := cfg.HasAuth()

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    serverName,
			Version: version + "+" + revision,
		},
		newServerOptions(logger, hasAuth),
	)

	registerTools(server, client, hasAuth)

	logger.Info("starting server", slog.Bool("authenticated", hasAuth))

	return serve(logger, server, cfg)
}

// newServerOptions wires the shared logger into the MCP server and describes the
// two operating modes in the instructions surfaced to clients.
func newServerOptions(logger *slog.Logger, hasAuth bool) *mcp.ServerOptions {
	instructions := "MCP server for MyShows.me. Public tools (search shows, show " +
		"and episode details, top chart, genres) work with no credentials. Set " +
		"MYSHOWS_USERNAME and MYSHOWS_PASSWORD to also enable the personal tracker: " +
		"your profile, tracked shows, watched and unwatched episodes, counters, " +
		"recommendations, and write actions (mark episodes watched, set show status, " +
		"rate shows and episodes)."

	if !hasAuth {
		instructions += " Authentication is not configured, so only the public tools are available."
	}

	return &mcp.ServerOptions{
		Instructions: instructions,
		Logger:       logger,
	}
}

// registerTools registers the public tools always, and the account and write
// tools only when authentication is configured.
func registerTools(server *mcp.Server, api myshows.API, hasAuth bool) {
	registerPublicTools(server, api)

	if hasAuth {
		registerAuthTools(server, api)
	}
}

// registerPublicTools registers the tools that require no credentials.
func registerPublicTools(server *mcp.Server, api myshows.API) {
	mcp.AddTool(server, tools.ServerVersionTool(),
		tools.NewServerVersionHandler(version, revision, runtime.Version()))
	mcp.AddTool(server, tools.SearchTool(), tools.NewSearchHandler(api))
	mcp.AddTool(server, tools.ShowTool(), tools.NewShowHandler(api))
	mcp.AddTool(server, tools.ShowByExternalTool(), tools.NewShowByExternalHandler(api))
	mcp.AddTool(server, tools.EpisodeTool(), tools.NewEpisodeHandler(api))
	mcp.AddTool(server, tools.TopTool(), tools.NewTopHandler(api))
	mcp.AddTool(server, tools.GenresTool(), tools.NewGenresHandler(api))
}

// registerAuthTools registers the account and write tools.
func registerAuthTools(server *mcp.Server, api myshows.API) {
	mcp.AddTool(server, tools.ProfileTool(), tools.NewProfileHandler(api))
	mcp.AddTool(server, tools.MyShowsTool(), tools.NewMyShowsHandler(api))
	mcp.AddTool(server, tools.MyEpisodesTool(), tools.NewMyEpisodesHandler(api))
	mcp.AddTool(server, tools.UnwatchedTool(), tools.NewUnwatchedHandler(api))
	mcp.AddTool(server, tools.CountersTool(), tools.NewCountersHandler(api))
	mcp.AddTool(server, tools.RecommendationsTool(), tools.NewRecommendationsHandler(api))
	mcp.AddTool(server, tools.CheckEpisodeTool(), tools.NewCheckEpisodeHandler(api))
	mcp.AddTool(server, tools.UnCheckEpisodeTool(), tools.NewUnCheckEpisodeHandler(api))
	mcp.AddTool(server, tools.SetShowStatusTool(), tools.NewSetShowStatusHandler(api))
	mcp.AddTool(server, tools.RateShowTool(), tools.NewRateShowHandler(api))
	mcp.AddTool(server, tools.RateEpisodeTool(), tools.NewRateEpisodeHandler(api))
}

// serve runs the stdio transport and, when configured, an HTTP transport.
func serve(logger *slog.Logger, server *mcp.Server, cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}

		signal.Stop(sigChan)
	}()

	group, groupCtx := errgroup.WithContext(ctx)
	httpEnabled := cfg.HTTPEnabled()

	group.Go(func() error {
		runErr := server.Run(groupCtx, &mcp.StdioTransport{})
		if runErr != nil && groupCtx.Err() == nil {
			return errors.Wrap(runErr, "stdio server failed")
		}

		if !httpEnabled {
			cancel()
		}

		return nil
	})

	if httpEnabled {
		group.Go(func() error {
			return runHTTPServer(groupCtx, logger, server, cfg.HTTPAddr(), cfg.HTTPToken)
		})
	}

	//nolint:wrapcheck // errors are already wrapped inside the group goroutines.
	return group.Wait()
}

// runHTTPServer starts an HTTP/SSE transport for the MCP server. Sharing a
// single *mcp.Server across transports is safe: the SDK guards internal state
// with a mutex. When token is set, every request must carry a matching Bearer
// token; otherwise the transport is unauthenticated and config validation has
// already confined it to a loopback host.
func runHTTPServer(ctx context.Context, logger *slog.Logger, server *mcp.Server, addr, token string) error {
	var handler http.Handler = mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return server },
		nil,
	)

	handler = bearerAuth(handler, token)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	//nolint:gosec // G118: shutdown uses a fresh context because ctx is already cancelled.
	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		shutdownErr := httpServer.Shutdown(shutdownCtx) //nolint:contextcheck // fresh context for graceful shutdown.
		if shutdownErr != nil {
			logger.Error("http server shutdown failed", slog.Any("error", shutdownErr))
		}
	}()

	logger.Info("http server listening", slog.String("addr", addr))

	listenErr := httpServer.ListenAndServe()
	if errors.Is(listenErr, http.ErrServerClosed) {
		return nil
	}

	return errors.Wrap(listenErr, "HTTP listen failed")
}

// bearerAuth wraps next so every request must present a matching
// "Authorization: Bearer <token>" header. An empty token disables the check;
// config validation has already confined that case to a loopback host.
func bearerAuth(next http.Handler, token string) http.Handler {
	if token == "" {
		return next
	}

	want := []byte("Bearer " + token)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		got := []byte(request.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(writer, request)
	})
}
