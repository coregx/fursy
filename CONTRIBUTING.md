# Contributing to FURSY HTTP Router

Thank you for considering contributing to FURSY! This document outlines the development workflow and guidelines.

## Git Workflow (Git-Flow)

This project uses Git-Flow branching model for development.

### Branch Structure

```
main                 # Production-ready code (tagged releases)
  └─ develop         # Integration branch for next release
       ├─ feature/*  # New features
       ├─ bugfix/*   # Bug fixes
       └─ hotfix/*   # Critical fixes from main
```

### Branch Purposes

- **main**: Production-ready code. Only releases are merged here.
- **develop**: Active development branch. All features merge here first.
- **feature/\***: New features. Branch from `develop`, merge back to `develop`.
- **bugfix/\***: Bug fixes. Branch from `develop`, merge back to `develop`.
- **hotfix/\***: Critical production fixes. Branch from `main`, merge to both `main` and `develop`.

### Workflow Commands

#### Starting a New Feature

```bash
# Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/my-new-feature

# Work on your feature...
git add .
git commit -m "feat: add my new feature"

# When done, merge back to develop
git checkout develop
git merge --no-ff feature/my-new-feature
git branch -d feature/my-new-feature
git push origin develop
```

#### Fixing a Bug

```bash
# Create bugfix branch from develop
git checkout develop
git pull origin develop
git checkout -b bugfix/fix-issue-123

# Fix the bug...
git add .
git commit -m "fix: resolve issue #123"

# Merge back to develop
git checkout develop
git merge --no-ff bugfix/fix-issue-123
git branch -d bugfix/fix-issue-123
git push origin develop
```

#### Creating a Release

```bash
# Create release branch from develop
git checkout develop
git pull origin develop
git checkout -b release/v0.4.0

# Update version numbers, CHANGELOG, etc.
git add .
git commit -m "chore: prepare release v0.4.0"

# Merge to main and tag
git checkout main
git merge --no-ff release/v0.4.0
git tag -a v0.4.0 -m "Release v0.4.0"

# Merge back to develop
git checkout develop
git merge --no-ff release/v0.4.0

# Delete release branch
git branch -d release/v0.4.0

# Push everything
git push origin main develop --tags
```

#### Hotfix (Critical Production Bug)

```bash
# Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug

# Fix the bug...
git add .
git commit -m "fix: critical production bug"

# Merge to main and tag
git checkout main
git merge --no-ff hotfix/critical-bug
git tag -a v0.3.1 -m "Hotfix v0.3.1"

# Merge to develop
git checkout develop
git merge --no-ff hotfix/critical-bug

# Delete hotfix branch
git branch -d hotfix/critical-bug

# Push everything
git push origin main develop --tags
```

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
chore: update golangci-lint to v1.60
```

## Code Quality Standards

### Before Committing

Run the pre-commit checks:

```bash
bash scripts/pre-release-check.sh
```

This script runs:
1. `go fmt ./...` - Format code
2. `golangci-lint run` - Lint code
3. `go test -race -coverprofile=coverage.txt ./...` - Run tests with race detector
4. Coverage check (>85% for Phase 1, >90% for Phase 2+)

### Pull Request Requirements

- [ ] Code is formatted (`go fmt ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] All tests pass with race detector (`go test -race ./...`)
- [ ] New code has tests (minimum 85% coverage for Phase 1, 90%+ for Phase 2+)
- [ ] Benchmarks for performance-critical code
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow Conventional Commits
- [ ] No sensitive data (credentials, tokens, etc.)
- [ ] Uses `encoding/json/v2` (NOT `encoding/json`)
- [ ] Uses `log/slog` for logging
- [ ] No external dependencies in core package

## Development Setup

### Prerequisites

- **Go 1.25 or later** (required for generics and modern stdlib)
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
├── examples/             # Usage examples
│   ├── hello-world/
│   └── rest-api/
├── internal/             # Internal implementation (not in Go docs)
│   ├── radix/           # Radix tree routing engine
│   └── pool/            # Context pooling
├── middleware/           # Built-in middleware
│   ├── logger/
│   ├── recovery/
│   ├── cors/
│   ├── ratelimit/
│   └── auth/
├── plugins/              # Optional plugins (can have dependencies)
│   ├── opentelemetry/
│   ├── validator/
│   └── openapi/
├── scripts/              # Development scripts
│   └── pre-release-check.sh
├── router.go             # Public API - Router (wrapper over internal/radix)
├── context.go            # Public API - Context (HTTP context)
├── box.go                # Public API - Box[Req, Res] (type-safe container)
├── handler.go            # Public API - Handler types
├── group.go              # Public API - Route groups
├── error.go              # Public API - RFC 9457 Problem Details
├── CONTRIBUTING.md       # This file
├── README.md             # Main documentation
└── go.mod                # Go module
```

## Architecture Principles

### Clean Public API (Wrapper Pattern)

FURSY uses a **wrapper architecture** to keep the public API clean:

```
github.com/coregx/fursy/          ← Public API (in Go docs)
├── router.go                    ← Wrapper over internal/radix
├── context.go                   ← Public API
└── handler.go                   ← Public types

github.com/coregx/fursy/internal/ ← Implementation (NOT in Go docs)
├── radix/                       ← Real routing implementation
└── pool/                        ← Context pooling
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
3. Create feature branch from `develop`
4. Write tests FIRST (TDD approach)
5. Implement feature
6. Add benchmarks for performance-critical code
7. Update documentation
8. Run quality checks (`bash scripts/pre-release-check.sh`)
9. Create pull request to `develop`
10. Wait for code review
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

#### 1. Use encoding/json/v2

```go
// ✅ CORRECT
import "encoding/json/v2"

// ❌ WRONG - Do NOT use old version
import "encoding/json"
```

#### 2. Use log/slog

```go
import "log/slog"

// Structured logging
slog.Info("request processed",
    "method", req.Method,
    "path", req.URL.Path,
    "duration", duration,
)
```

#### 3. Error Handling with RFC 9457

```go
// Always use RFC 9457 Problem Details
return c.Error(404, fursy.NotFound("User not found"))
return c.Error(400, fursy.BadRequest("Invalid email"))
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
func BenchmarkRouter_SimpleRoute(b *testing.B) {
    r := New()
    r.GET("/users/:id", handler)

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
- Check [.claude/STATUS.md](.claude/STATUS.md) for current project status

## Performance Benchmarking

FURSY prioritizes performance. See [PERFORMANCE.md](PERFORMANCE.md) for:
- Current benchmark results
- Performance optimization techniques
- Comparison with other routers

**Current metrics** (Phase 3):
- Static routes: 256 ns/op, 1 alloc/op
- Parametric routes: 326 ns/op, 1 alloc/op
- Deep nesting (4 params): 561 ns/op, 1 alloc/op
- Coverage: 91.7%

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to FURSY HTTP Router!** 🚀
