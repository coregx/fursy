# AGENTS.md — working in fursy

fursy is a Go 1.27 HTTP router (`github.com/coregx/fursy`) with type-safe generic handlers,
RFC 9457 problem responses and OpenAPI 3.1 generation.

## Start here

- `README.md` — overview, quick start, middleware and examples.
- `llms.txt` — the detailed guide for AI agents (this file is the short entry).
- `CHANGELOG.md` — current version and changes.
- `CONTRIBUTING.md` — contribution workflow and checks.

## Rules that matter

- `go 1.27` is the module floor. Core routing has zero external dependencies; middleware adds only
  `golang-jwt/jwt` (JWT) and `golang.org/x/time` (rate limit).
- Plain handlers use `router.Handle("GET", path, fn)` with `*fursy.Context`; typed handlers use
  `router.GET(path, func(c *fursy.Box[Req, Res]) error)`.
- `*fursy.Context` is pooled: never keep a reference after the handler returns.
- Errors are RFC 9457 problems by default; JSON fields are camelCase.
- `middleware.Timeout` is cooperative: it sets a context deadline and returns 503 when the handler
  returns the deadline error before writing. Register header-setting middleware (CORS, Secure)
  before `Timeout`, or their headers are discarded with the handler's on a 503.
- Fixes and features start with a failing test. Before a PR: `gofmt`, `go test ./...`,
  `golangci-lint run` (0 issues), and the example/plugin modules still build.
