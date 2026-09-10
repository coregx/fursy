# Contributing to FURSY HTTP Router

Thank you for considering contributing to FURSY! This document outlines the development workflow and guidelines.

## Git Workflow (GitHub Flow)

This project uses GitHub Flow — a simple branch-based workflow.

### Branch Structure

```
main                 # Production-ready code (tagged releases)
  ├─ feat/*          # New features
  ├─ fix/*           # Bug fixes
  └─ chore/*         # Maintenance (deps, docs, CI)
```

### Rules

- **`main`** is the only long-lived branch. All PRs target `main`.
- **Never push directly to `main`** — all changes go through Pull Requests.
- Small, self-contained changes (typos, dep bumps) can be a single commit.
- Larger features should use a feature branch.

### Contributing a Feature or Fix

```bash
# 1. Fork the repo and clone your fork
git clone https://github.com/YOUR_USERNAME/fursy.git
cd fursy

# 2. Create a branch from main
git checkout main
git pull origin main
git checkout -b feat/my-new-feature

# 3. Work on your changes...
git add .
git commit -m "feat: add my new feature"

# 4. Push to your fork
git push origin feat/my-new-feature

# 5. Open a Pull Request targeting main on coregx/fursy
```

### After PR is Merged

Your branch is automatically deleted. Tags and releases are created by maintainers.

## Semantic Versioning

FURSY follows [Semantic Versioning 2.0.0](https://semver.org/):

### For 0.x.y versions (pre-1.0):
- **0.y.0** - New features (minor bump)
- **0.y.z** - Bug fixes, hotfixes (patch bump)

### For 1.x.y+ versions (stable API):
- **Major (x.0.0)** - Breaking changes
- **Minor (x.y.0)** - New features (backwards-compatible)
- **Patch (x.y.z)** - Bug fixes only

**Note**: v1.0.0 will only be released after 6-12 months of production usage and full API stabilization. Breaking changes are allowed in 0.x versions.

## Commit Message Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (build, dependencies, etc.)
- **perf**: Performance improvements

### Examples

```bash
feat(router): add wildcard route support
fix(context): resolve parameter extraction edge case
docs: update README with OpenAPI examples
refactor(radix): simplify tree traversal logic
test(middleware): add benchmarks for chain execution
perf(pool): optimize context pooling strategy
chore: update golangci-lint to v2.13
```

## Code Quality Standards

### Before Committing

Run the pre-commit checks:

```bash
gofmt -l .                  # Verify formatting (must be empty)
golangci-lint run            # Lint (0 issues required)
go test ./...                # Run tests
```

### Pull Request Requirements

- [ ] Code is formatted (`go fmt ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] All tests pass with race detector (`go test -race ./...`)
- [ ] New code has tests (minimum 85% coverage for Phase 1, 90%+ for Phase 2+)
- [ ] Benchmarks for performance-critical code
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow Conventional Commits
- [ ] No sensitive data (credentials, tokens, etc.)
- [ ] Uses `log/slog` for logging
- [ ] No external dependencies in core package

## Development Setup

### Prerequisites

- **Go 1.27 or later** (required for generic methods on concrete types)
- **golangci-lint** (for code quality checks)
- **git** (for version control)

### Install Dependencies

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify installation
golangci-lint --version
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.txt ./...

# Run with race detector (always use before commit!)
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkRouter_StaticRoute -benchmem ./...
```

### Running Linter

```bash
# Run linter
golangci-lint run

# Run with verbose output
golangci-lint run -v

# Run and save report
golangci-lint run --out-format=colored-line-number > lint-report.txt
```

## Project Structure

```
fursy/
├── .golangci.yml         # Linter configuration
├── .github/
│   └── workflows/        # CI/CD pipelines
├── docs/                 # Public documentation
│   └── PERFORMANCE.md   # Benchmark results
├── examples/             # Usage examples
│   ├── 01-hello-world/
│   ├── 02-rest-api-crud/
│   ├── validation/      # Validation examples (01-basic through 06-production)
│   └── ...              # More examples (middleware, SSE, WebSocket, DB)
├── internal/             # Internal implementation (not in Go docs)
│   ├── radix/           # Radix tree routing engine
│   ├── binding/         # Request body binding (JSON, XML, form)
│   ├── negotiate/       # Content negotiation
│   └── validation/      # Validation internals
├── middleware/           # Built-in middleware (flat package, not subdirs)
│   ├── logger.go        # log/slog logging
│   ├── recovery.go      # Panic recovery
│   ├── cors.go          # CORS headers
│   ├── basicauth.go     # Basic authentication
│   ├── jwt.go           # JWT authentication
│   ├── ratelimit.go     # Rate limiting
│   ├── circuitbreaker.go # Circuit breaker
│   └── secure.go        # Security headers (OWASP)
├── plugins/              # Optional plugins (can have dependencies)
│   ├── opentelemetry/   # Tracing + metrics
│   ├── validator/       # go-playground/validator integration
│   ├── stream/          # SSE + WebSocket
│   └── database/        # SQL with transactions
├── router.go             # Public API - Router
├── context.go            # Public API - Context (non-generic)
├── box.go                # Public API - Box[Req, Res] (type-safe generic context)
├── handler_generic.go    # Generic handler adapter
├── group.go              # Route groups
├── problem.go            # RFC 9457 Problem Details
├── openapi.go            # OpenAPI 3.1 generation
└── go.mod                # Go module
```

## Architecture Principles

### Clean Public API (Wrapper Pattern)

FURSY uses a **wrapper architecture** to keep the public API clean:

```
github.com/coregx/fursy/          ← Public API (in Go docs)
├── router.go                    ← Router + route registration
├── context.go                   ← Non-generic HTTP context
├── box.go                       ← Box[Req, Res] generic context
└── problem.go                   ← RFC 9457 Problem Details

github.com/coregx/fursy/internal/ ← Implementation (NOT in Go docs)
├── radix/                       ← Radix tree routing engine
├── binding/                     ← Request body binding
└── negotiate/                   ← Content negotiation
```

**Why?**
- `internal/` packages cannot be imported by external modules
- Go docs show ONLY clean, simple public API
- Implementation details hidden from users
- Allows changing internals without breaking changes

### Zero Dependencies (Core)

- **Core package** (`fursy/`) must use ONLY stdlib
- **Plugins** (`plugins/`) can have dependencies
- Never add external dependencies to core without discussion

### Type Safety

Use Go generics for type-safe handlers:

```go
type Handler[Req, Res any] func(*Box[Req, Res]) error

type Box[Req, Res any] struct {
    *Context  // Embedded base context
    ReqBody  *Req
    ResBody  *Res
}
```

## Adding New Features

1. Check if issue exists, if not create one
2. Discuss approach in the issue
3. Create feature branch from `main`
4. Write tests FIRST (TDD approach)
5. Implement feature
6. Add benchmarks for performance-critical code
7. Update documentation
8. Run quality checks (`gofmt -l . && golangci-lint run && go test ./...`)
9. Create pull request targeting `main`
10. Wait for code review and CI
11. Address feedback
12. Merge when approved

## Code Style Guidelines

### General Principles

- Follow Go conventions and idioms
- Write self-documenting code
- Add comments for complex logic (especially in radix tree)
- Keep functions small and focused (<50 lines ideal)
- Use meaningful variable names
- **TDD approach** - write tests first!

### Naming Conventions

- **Public types/functions**: `PascalCase` (e.g., `Router`, `ServeHTTP`)
- **Private types/functions**: `camelCase` (e.g., `findRoute`, `extractParams`)
- **Constants**: `PascalCase` (e.g., `StatusOK`, `MethodGet`)
- **Test functions**: `Test*` (e.g., `TestRouter_GET`)
- **Benchmark functions**: `Benchmark*` (e.g., `BenchmarkRouter_StaticRoute`)

### Required Standards

#### 1. Use log/slog

```go
import "log/slog"

// Structured logging
slog.Info("request processed",
    "method", req.Method,
    "path", req.URL.Path,
    "duration", duration,
)
```

#### 2. Error Handling with RFC 9457

```go
// Use RFC 9457 Problem Details
return c.Problem(fursy.NotFound("User not found"))
return c.Problem(fursy.BadRequest("Invalid email"))
```

### Testing

- Use table-driven tests when appropriate
- Test both success and error cases
- Use `testing.T.Run()` for subtests
- **Minimum coverage**: 85% (Phase 1), 90%+ (Phase 2+)
- Always run with race detector: `go test -race`

### Benchmarking

Performance is a core goal. Always benchmark critical paths:

```go
func BenchmarkRouter_StaticRoute(b *testing.B) {
    r := New()
    r.Handle("GET", "/users/:id", handler)

    req := httptest.NewRequest("GET", "/users/123", nil)
    w := httptest.NewRecorder()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        r.ServeHTTP(w, req)
    }
}
```

**Performance goals**:
- Route lookup: <500ns (parametric), <300ns (static)
- Allocations: 1 alloc/op (routing hot path)
- Throughput: >1M req/s (with middleware)

## Getting Help

- Check [existing issues](https://github.com/coregx/fursy/issues)
- Read documentation in `docs/`
- Review examples in `examples/`
- Ask questions in GitHub Issues

## Performance Benchmarking

FURSY prioritizes performance. See [docs/PERFORMANCE.md](docs/PERFORMANCE.md) for:
- Current benchmark results
- Performance optimization techniques
- Comparison with other routers

**Current metrics** (v0.4.0):
- Static routes: 256 ns/op, 1 alloc/op
- Parametric routes: 326 ns/op, 1 alloc/op
- Coverage: 94.6%

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to FURSY HTTP Router!** 🚀
