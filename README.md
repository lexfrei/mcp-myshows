# mcp-myshows

[![CI](https://github.com/lexfrei/mcp-myshows/actions/workflows/ci.yml/badge.svg)](https://github.com/lexfrei/mcp-myshows/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/lexfrei/mcp-myshows?sort=semver)](https://github.com/lexfrei/mcp-myshows/releases) [![Go Report Card](https://goreportcard.com/badge/github.com/lexfrei/mcp-myshows)](https://goreportcard.com/report/github.com/lexfrei/mcp-myshows) [![Go](https://img.shields.io/github/go-mod/go-version/lexfrei/mcp-myshows)](go.mod) [![License](https://img.shields.io/github/license/lexfrei/mcp-myshows)](LICENSE)

MCP server for [MyShows.me](https://myshows.me). Search shows, inspect a show or episode, and — when you sign in — drive your personal tracker: watchlist, unwatched queue, marking episodes watched, ratings, and recommendations — all from any MCP-compatible client.

## Highlights

- **Hybrid auth.** Public tools (search, show/episode details, top chart, genres) need no credentials. The personal-tracker tools light up only when you provide a username and password.
- **No OAuth AppID needed.** Authentication uses the same session login the MyShows website does, so your username and password are enough — there is nothing to request from the MyShows team.
- **Typed, trimmed results.** Responses are mapped to compact structs (id, title, year, status, rating, watch progress, …) rather than dumped raw, keeping tool output token-cheap.
- Distroless multi-arch container image, signed with cosign.

## Features

### Public (no credentials)

- **myshows_search** — search shows by title or keywords.
- **myshows_show** — full show details (title, year, status, ratings, network), optionally with the episode list.
- **myshows_show_by_external** — look a show up by an external id from imdb, kinopoisk, or thetvdb.
- **myshows_episode** — details of a single episode.
- **myshows_top** — the top-ranked shows chart.
- **myshows_genres** — the list of genres.
- **myshows_server_version** — report the server version, revision, and Go runtime.

### Personal tracker (requires sign-in)

- **myshows_profile** — a user profile with watch-time statistics.
- **myshows_my_shows** — tracked shows with watch status, rating, and progress.
- **myshows_my_episodes** — your watched episodes for a show, with dates and ratings.
- **myshows_unwatched** — episodes you have not watched yet (all pending, or the next one per show).
- **myshows_counters** — pending-item counts (unwatched episodes, new comments, achievements).
- **myshows_recommendations** — personalised show recommendations.
- **myshows_check_episode** / **myshows_uncheck_episode** — mark an episode watched (optionally rated) or unwatched.
- **myshows_set_show_status** — set a show to watching, later, cancelled, or remove it from the tracker.
- **myshows_rate_show** / **myshows_rate_episode** — rate a show or episode (0–5).

The write tools (`check`, `uncheck`, `set_show_status`, `rate_*`) modify your MyShows account and are annotated as non-read-only so clients can gate them.

## Authentication

MyShows' documented API uses OAuth 2.0, which requires an application id (`client_id`/`client_secret`) granted on request. To keep setup to just a username and password, this server currently authenticates the way the MyShows website does: it posts the credentials to the session endpoint, receives a Bearer token, and calls the documented v2 JSON-RPC API with it. The token is cached on disk and refreshed by logging in again when it expires.

> **Interim approach.** The session login is a pragmatic stand-in. Once an OAuth application id is granted, authentication will move to the official OAuth 2.0 flow; the token-acquisition code is isolated in one file (`internal/myshows/auth.go`) so the switch is contained.

Alternatively, set `MYSHOWS_TOKEN` to a Bearer token you already hold to skip the login entirely (the server cannot refresh such a token on its own).

## Configuration

Configuration is read from environment variables. Credentials are optional: without them the public tools still work.

| Variable | Description | Default |
| --- | --- | --- |
| `MYSHOWS_USERNAME` | Account username (enables the personal-tracker tools) | — |
| `MYSHOWS_PASSWORD` | Account password | — |
| `MYSHOWS_TOKEN` | Pre-obtained Bearer token, used instead of a username login | — |
| `MYSHOWS_TOKEN_FILE` | Path to persist the session token between runs | `~/.mcp-myshows/token.json` (bare process); `/home/nobody/.mcp-myshows/token.json` (container, set in the image) |
| `MYSHOWS_API_URL` | Override the JSON-RPC endpoint | `https://api.myshows.me/v2/rpc/` |
| `MYSHOWS_SESSION_URL` | Override the session login endpoint | `https://myshows.me/api/session` |
| `MYSHOWS_USER_AGENT` | Override the User-Agent | a generic browser UA |
| `MYSHOWS_PROXY` | HTTP/SOCKS5 proxy URL | — |
| `MCP_HTTP_PORT` | Enable the HTTP transport on this port | stdio only |
| `MCP_HTTP_HOST` | HTTP bind host | `127.0.0.1` |
| `MCP_HTTP_TOKEN` | Bearer token required on every HTTP request | — |

### HTTP transport security

The default transport is stdio, which is local and needs no network exposure. The optional HTTP transport (`MCP_HTTP_PORT`) serves the full tool set — including the account write tools — and has no built-in per-request authentication. To prevent accidentally handing your MyShows account to anyone who can reach the port, the server refuses to start when the HTTP transport would bind to a non-loopback host (`MCP_HTTP_HOST` other than a loopback address) unless `MCP_HTTP_TOKEN` is set. When the token is set, every HTTP request must carry `Authorization: Bearer <token>`. Keep the transport on `127.0.0.1`, or set a token and place it behind TLS, for any non-local use.

## Usage

With Claude Code, via the bundled `.mcp.json` (Docker):

```json
{
  "mcpServers": {
    "mcp-myshows": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "MYSHOWS_USERNAME",
        "-e", "MYSHOWS_PASSWORD",
        "-e", "MYSHOWS_TOKEN",
        "-v", "mcp-myshows-session:/home/nobody/.mcp-myshows",
        "ghcr.io/lexfrei/mcp-myshows:latest"
      ],
      "env": {
        "MYSHOWS_USERNAME": "your-username",
        "MYSHOWS_PASSWORD": "your-password"
      }
    }
  }
}
```

Leave `MYSHOWS_USERNAME`/`MYSHOWS_PASSWORD` empty to run in public-only mode — search and show/episode lookups work without an account. The named volume persists the session token across `--rm` container runs; drop it if you do not want persistence.

## Development

```bash
go build ./cmd/mcp-myshows
go test -race ./...
golangci-lint run
```

Opt-in live integration tests exercise the real API:

```bash
# public path, no credentials
go test -tags integration -run TestLive_Search -count=1 ./internal/myshows/

# authenticated path
MYSHOWS_USERNAME=... MYSHOWS_PASSWORD=... \
  go test -tags integration -run TestLive_Account -count=1 ./internal/myshows/
```

## Support

If this project is useful to you, you can support its development via [GitHub Sponsors](https://github.com/sponsors/lexfrei).

## See also

Another MyShows MCP server worth knowing about: [zeloras/myshows_mcp](https://github.com/zeloras/myshows_mcp).

## License

BSD 3-Clause. See [LICENSE](LICENSE).
