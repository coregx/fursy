# Changelog

All notable changes to FURSY will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Future features and enhancements (Phase 4: Ecosystem)

## [0.3.0] - 2025-01-19

### Added

#### Real-Time Communications (plugins/stream)
- **Stream Library Integration**: Official `plugins/stream` for SSE and WebSocket support
  - `SSEHub[T]` middleware for type-safe Server-Sent Events broadcasting
  - `WebSocketHub` middleware for WebSocket connection management
  - Context helpers: `stream.SSEUpgrade()`, `stream.WebSocketUpgrade()`
  - Type-safe helpers: `stream.GetSSEHub[T]()`, `stream.GetWebSocketHub()`
  - Integration with `github.com/coregx/stream` v0.1.0 (314 tests, 84.3% coverage)
  - **83.3% test coverage** (13 tests, 700+ lines)

#### Database Integration (plugins/database)
- **Database Plugin**: Official `plugins/database` for database/sql integration
  - `Middleware()` for injecting database into request context
  - `TxMiddleware()` for automatic transaction management
  - Context helper: `c.DB()` for convenient database access
  - Enhanced helpers: `MustGetDB()`, `GetDBOrError()`, `MustGetTx()`, `GetTxOrError()`
  - dbcontext pattern with 3 approaches (production-ready + prototyping + custom)
  - Repository pattern integration examples
  - Transaction support: `BeginTx()`, `Commit()`, `Rollback()`, `WithTx()`
  - Pure Go SQLite support (modernc.org/sqlite, no CGO)
  - Works with any database/sql driver (PostgreSQL, MySQL, SQLite, SQL Server)
  - **89.7% test coverage** (17 tests including CRUD integration test)

#### Production Boilerplate (examples/10-production-boilerplate)
- **Complete DDD Example**: Domain-Driven Design production reference implementation
  - Domain-Driven Design with Rich Models (User entity with business logic)
  - Clean Architecture: domain/application/infrastructure/interfaces layers
  - Value Objects: Email, Password, Role, Status (enforcing invariants)
  - User management module:
    - Registration with password hashing (bcrypt)
    - JWT authentication (login, token refresh)
    - Profile management (get, update, delete)
    - Role-based authorization (user/admin)
  - Real-time features:
    - Server-Sent Events for notifications
    - WebSocket chat (with join/leave messages)
    - Multi-client broadcasting
  - Database:
    - SQLite with migrations
    - Transaction support
    - Repository pattern
  - Infrastructure:
    - Docker support (Dockerfile + docker-compose.yml)
    - Graceful shutdown (signal handling)
    - Structured logging (log/slog)
    - RFC 9457 error responses
  - Testing:
    - **36 comprehensive tests** (1,766 lines)
    - **60.6% test coverage**
    - Unit tests for all layers
    - Integration tests for API endpoints
  - **Complete README** (14.8KB, API documentation, setup instructions)

#### Examples
- **SSE Notifications**: `examples/07-sse-notifications` - Real-time server push
  - Multi-client notification hub
  - POST endpoint to broadcast notifications
  - Complete working example
- **WebSocket Chat**: `examples/08-websocket-chat` - Bidirectional communication
  - Real-time chat server with broadcasting
  - HTML/CSS/JS web client included
  - Join/leave notifications
  - Username support
- **Database CRUD**: `examples/09-rest-api-with-db` - Database integration
  - Complete REST API with SQLite
  - CRUD operations (Create, Read, Update, Delete)
  - Transaction examples
  - Error handling with RFC 9457
- **Examples Navigation**: `examples/README.md` with learning path
  - 10 total examples (from basic to production)
  - Progressive learning path (Beginner → Intermediate → Advanced)
  - Quick navigation table
  - Common patterns reference

#### Context Methods
- **Plugin Integration Methods** in Context:
  - `c.DB()` - Access database connection (requires plugins/database)
  - `c.SSE()` stub method - Returns ErrStreamNotImported if plugin not used
  - `c.WebSocket()` stub method - Returns ErrStreamNotImported if plugin not used
  - New error: `ErrStreamNotImported` for clear error messages
  - **4 integration tests** in `context_integration_test.go`

### Changed
- **Coverage**: Increased from 88.9% to **93.1% overall** (+4.2%)
  - Core: 93.1% (maintained high quality)
  - plugins/database: 89.7% (17 tests)
  - plugins/stream: 83.3% (13 tests)
  - examples/10-production-boilerplate: 60.6% (36 tests)
- **Test Count**: Added **70+ tests** across new features
  - Database: 17 tests
  - Stream plugin: 13 tests
  - Production boilerplate: 36 tests
  - Integration: 4 tests
  - **Total: 650+ tests** across all packages
- **Dependencies**: Added `github.com/coregx/stream` as ecosystem dependency
  - Core routing: Still ZERO dependencies (stdlib only)
  - Stream plugin: Depends on github.com/coregx/stream v0.1.0
  - Database plugin: Zero external dependencies (database/sql only)

### Ecosystem
- **stream v0.1.0 Released**: Companion library for real-time communications
  - **SSE**: 92.3% coverage, 61 tests, RFC text/event-stream compliant
  - **WebSocket**: 84.3% coverage, 253 tests, RFC 6455 compliant
  - Zero external dependencies (pure stdlib)
  - High performance: 11 μs/op (10 WS clients), 8 μs/op (10 SSE clients)
  - Production-ready with comprehensive documentation

### Documentation
- **README.md**: Updated with v0.3.0 features
  - New "Plugin Integration Methods" section
  - Real-time communications examples
  - Database integration patterns
  - Updated version badges (v0.3.0, 93.1% coverage)
  - Updated roadmap with v0.3.0 completion
- **Plugin Documentation**:
  - `plugins/database/README.md` - Complete database integration guide (558 lines)
  - `plugins/stream/README.md` - Complete stream integration guide (373 lines)
  - dbcontext pattern with 3 approaches
  - Repository pattern integration
  - Best practices and examples
- **Examples Documentation**:
  - `examples/README.md` - Navigation guide with learning path (121 lines)
  - `examples/10-production-boilerplate/README.md` - Production guide (14.8KB)
  - Progressive learning path (Beginner → Intermediate → Advanced)

### Performance
- **Maintained Routing Performance**: 256 ns/op (static), 326 ns/op (parametric)
- **Real-time Performance** (via stream v0.1.0):
  - SSE Hub broadcast (10 clients): 8 μs/op
  - WebSocket Hub broadcast (10 clients): 11 μs/op
- All benchmarks validated and documented
- No performance regressions

### Testing
- **Total Tests**: 650+ across all packages (up from ~580)
- **Coverage**:
  - Core: 93.1% (maintained high quality)
  - Database plugin: 89.7%
  - Stream plugin: 83.3%
  - Production boilerplate: 60.6%
- Comprehensive integration tests for all new features
- Production boilerplate demonstrates testing best practices (1,766 lines of tests)

### Contributors
- Enhanced by Claude (fursy-senior-architect agent)

## [0.2.0] - 2025-01-18

### Added

**Validator Plugin**
- New `plugins/validator` package with go-playground/validator/v10 integration
- Automatic request validation with 100+ validation tags
- RFC 9457 Problem Details error conversion
- 40+ default error messages for common validation tags
- Custom error messages support
- 94.3% test coverage

**Documentation Enhancements**
- **Middleware Section** in README.md (288 lines)
  - All 8 built-in middleware documented with examples
  - Configuration options for each middleware
  - Comparison table vs Gin/Echo/Fiber
  - OPUS visibility fix - all middleware now discoverable
- **Automatic Validation Section** in README.md
  - Complete guide to request validation
  - Type-safe Box[Req, Res] validation examples
  - RFC 9457 error responses
  - Comparison with other routers
- **Content Negotiation Section** in README.md
  - RFC 9110 compliant content negotiation
  - AI agent support (text/markdown)
  - Q-value priority examples
  - Multi-format response examples
- **Observability Section** in README.md
  - OpenTelemetry integration guide
  - Distributed tracing with Jaeger
  - Metrics collection with Prometheus
  - Custom spans examples
- **llms.md** - Complete guide for AI agents (1,716 lines)
  - Project architecture and structure
  - Development standards (encoding/json/v2, log/slog, minimal deps)
  - Testing requirements and git workflow
  - All 8 middleware documented
  - Common gotchas with fixes
  - Contributing guidelines

**Examples** (11 Total)
- **Basic Examples**
  - `01-hello-world` - Minimal fursy application (<30 lines)
  - `02-rest-api-crud` - Complete CRUD API (385 lines)
- **Advanced Examples**
  - `04-content-negotiation` - Multi-format responses (1,507 lines)
    - Accepts(), AcceptsAny(), Markdown() methods
    - Q-value priority handling
    - AI agent friendly responses
  - `05-middleware` - All middleware + custom patterns (1,512 lines)
    - All 8 built-in middleware demonstrated
    - 8 custom middleware patterns
    - Production-ready configurations
  - `06-opentelemetry` - Distributed tracing (1,270 lines)
    - OTLP/HTTP integration with Jaeger
    - Custom spans for DB queries and external calls
    - Docker Compose for Jaeger + Prometheus
    - 7 endpoints demonstrating tracing patterns
- **Validation Examples** (6 examples in validation/ directory)
  - `validation/01-basic` - Simple validation demo
  - `validation/02-rest-api-crud` - Full CRUD with validator
  - `validation/03-custom-validator` - Custom validation functions
  - `validation/04-nested-structs` - Nested struct validation
  - `validation/05-custom-messages` - Custom error messages
  - `validation/06-production` - Production-ready setup
- **Examples Index**
  - `examples/README.md` - Navigation guide (602 lines)
  - Progressive learning path (Beginner → Intermediate → Advanced)
  - Total learning time: ~3.5 hours

### Improvements

**Developer Experience**
- Progressive learning path from basic to advanced (11 examples)
- Complete navigation guide in examples/README.md
- AI agent friendly documentation (llms.md)
- All middleware now visible in README (OPUS discoverability fix)

**Statistics**
- 7,000+ lines of documentation added
- 5,000+ lines of example code
- 11 complete working examples
- 3 new plugins documented (Validator, OpenTelemetry, Metrics)

### Documentation

**Updated Files**
- README.md - Added 4 major sections (Middleware, Validation, Content Negotiation, Observability)
- llms.md - Created comprehensive AI agent guide
- examples/README.md - Created examples navigation index

**Coverage**
- All 8 middleware now documented
- All public APIs documented with examples
- Production patterns demonstrated in examples

## [0.1.0] - 2025-11-16

### Added

**Core HTTP Router**
- High-performance HTTP router with radix tree algorithm
- Support for three route types:
  - Static routes: `/users`
  - Named parameters: `/users/:id`
  - Wildcard routes: `/files/*path`
- All standard HTTP methods: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- Generic `Handle(method, path, handler)` for custom methods

**Context API**
- Context-based request/response handling
- URL parameter extraction: `Param(key)`
- Query parameter helpers: `Query()`, `QueryDefault()`, `QueryValues()`
- Form parameter helpers: `Form()`, `FormDefault()`, `PostForm()`
- Data storage for middleware: `Get()`, `Set()`, `GetString()`, `GetInt()`, `GetBool()`

**Response Helpers**
- Explicit methods (full control):
  - `String(code, text)` - Text responses
  - `JSON(code, data)` - JSON serialization (uses `encoding/json/v2`)
  - `JSONIndent(code, data, indent)` - Pretty JSON
  - `XML(code, data)` - XML serialization
  - `NoContent(code)` - Empty responses
  - `Redirect(code, url)` - HTTP redirects (301, 302, 307, 308)
  - `Blob(code, contentType, data)` - Binary responses
  - `Stream(code, contentType, reader)` - Streaming responses
  - `Error(status, problem)` - RFC 9457 Problem Details responses
- Convenience methods (REST best practices):
  - `OK(obj)` - 200 OK JSON response
  - `Created(obj)` - 201 Created (POST best practice)
  - `Accepted(obj)` - 202 Accepted (async operations)
  - `NoContentSuccess()` - 204 No Content (DELETE best practice)
  - `Text(s)` - 200 OK plain text response

**HTTP Headers**
- `SetHeader(key, value)` - Set response headers
- `GetHeader(key)` - Get request headers

**Middleware**
- Middleware pipeline with `Next()` and `Abort()` pattern
- Pre-allocated handlers buffer (capacity 16)
- Route groups with nested middleware inheritance
- JWT Authentication middleware (94.2% coverage)
- Rate Limiting middleware (94.4% coverage, token bucket algorithm)
- Security Headers middleware (100% coverage, OWASP 2025 compliant)
- Circuit Breaker middleware (95.5% coverage, zero dependencies)
- Recovery middleware (panic recovery)
- Logger middleware (structured logging with `log/slog`)
- CORS middleware (Cross-Origin Resource Sharing)

**Performance Optimizations**
- Context pooling with `sync.Pool` - 1 alloc/op
- Radix tree routing: 256 ns/op static, 326 ns/op parametric
- Zero-allocation parameter extraction
- Pre-allocated buffers (params: 8, handlers: 16)
- Memory leak prevention (max capacity limits: 32/64)
- Efficient memory usage (~10M req/s throughput)

**Error Handling**
- **RFC 9457 Problem Details** - Standardized error responses
- Automatic 404 Not Found for unregistered routes
- Automatic 405 Method Not Allowed (configurable)
- Error propagation from handlers
- Predefined errors: `ErrInvalidRedirectCode`
- Helper functions: `BadRequest()`, `Unauthorized()`, `Forbidden()`, `NotFound()`, etc.

**Production Features**
- **Graceful Shutdown** - Connection draining, signal handling (SIGTERM, SIGINT)
- **Circuit Breaker** - Failure threshold, auto-recovery, zero dependencies
- **Rate Limiting** - Token bucket, per-IP/per-user, configurable limits
- **Security Headers** - CSP, HSTS, X-Frame-Options, X-Content-Type-Options
- **JWT Authentication** - Token validation, claims extraction, secure defaults

**Testing & Quality**
- Comprehensive test suite: 100+ test functions, 300+ test cases
- **91.7% overall test coverage** (exceeded Phase 2 target of 88%)
- Race condition testing with `-race` flag (all tests pass)
- Benchmark suite with 19 benchmarks
- Memory allocation tracking (1 alloc/op routing hot path)
- golangci-lint configuration with 34+ enabled linters
- Pre-release check script (`scripts/pre-release-check.sh`)
- Cross-platform testing (Linux, macOS, Windows)

**Documentation**
- Complete package documentation with examples
- godoc comments on all exported types and functions
- README.md with quick start guide
- CHANGELOG.md (this file)
- CONTRIBUTING.md - Development workflow and git-flow
- RELEASE_GUIDE.md - Release process documentation
- SECURITY.md - Security policy and best practices
- PERFORMANCE.md - Detailed benchmark results and optimization guide
- ROADMAP.md - Project roadmap and version strategy
- Example code for all major features

### Performance

**Benchmarks** (on Intel Core i7-1255U, 12th Gen):
- Static routes: **256 ns/op**, 1 alloc/op, ~10.5M ops/s
- Parametric routes: **326 ns/op**, 1 alloc/op, ~7.2M ops/s
- Multiple params: **344 ns/op**, 1 alloc/op, ~6.3M ops/s
- Deep nesting (4 params): **561 ns/op**, 1 alloc/op, ~4.0M ops/s
- Wildcard routes: **539 ns/op**, 1 alloc/op, ~7.8M ops/s
- Context.Param(): **3.7 ns/op**, 0 allocs/op
- Context.Query(): **21.8 ns/op**, 0 allocs/op
- Middleware chain: **1805 ns/op**, 11 allocs/op, ~1.7M ops/s

**Coverage**:
- Overall: **91.7%** (exceeded Phase 1 target of 85%)
- Security middleware: 94-100%
- Circuit breaker: 95.5%
- JWT authentication: 94.2%
- Rate limiting: 94.4%

See [PERFORMANCE.md](PERFORMANCE.md) for detailed results.

### Technical Details

**Dependencies**:
- **Minimal external dependencies** - Core routing: stdlib only
- **Middleware dependencies**: JWT (`golang-jwt/jwt/v5`), Rate Limiting (`golang.org/x/time`)
- Go 1.25+ required (uses generics and modern features)
- Uses standard library v2: `encoding/json/v2`, `log/slog`
- Standard library: `net/http`, `encoding/xml`, `sync`, `time`
- Plugins (optional) may have additional dependencies

**Architecture**:
- Clean separation: public API (fursy) wraps internal implementation (internal/radix)
- Radix tree routing algorithm (based on httprouter design)
- Context pooling pattern for performance
- Interface-based design for extensibility

### Breaking Changes

- Initial release, no breaking changes
- API is production-ready but may evolve in 0.x versions
- v1.0.0 will guarantee API stability (planned Q3 2026)

### Notes

This release represents **Phase 0-3 completion**:
- **Phase 0**: Project setup, CI/CD, documentation infrastructure
- **Phase 1**: Foundation - Radix tree routing, basic middleware, route groups
- **Phase 2**: API Excellence - RFC 9457, improved error handling
- **Phase 3**: Production Features - Auth, rate limiting, circuit breaker, pooling

**Next Steps** (v0.2.0+):
- Additional middleware (Database, Cache)
- Documentation website (fursy.coregx.dev)
- Migration guides from popular frameworks
- Community building and feedback

**Version Strategy**:
- **0.y.0** - New features (e.g., v0.2.0, v0.3.0)
- **0.y.z** - Bug fixes, hotfixes
- **v1.0.0** - Long-term API stability (Q3 2026, after 6-12 months production validation)

### Credits

This is the first production-ready release of FURSY. All core features are complete:
- ✅ Generic type-safe handlers with `Box[Req, Res]`
- ✅ Middleware pipeline system with Next/Abort pattern
- ✅ RFC 9457 Problem Details for standardized errors
- ✅ OpenAPI 3.1 specification generation (built-in)
- ✅ Route groups with nested middleware inheritance
- ✅ High-performance routing (256-326 ns/op, 1 alloc/op)
- ✅ Production features (JWT, Rate Limiting, Circuit Breaker, Security Headers)
- ✅ Convenience methods for REST best practices
- ✅ 91.7% test coverage

See [ROADMAP.md](ROADMAP.md) for future ecosystem features (v0.2.0+).

---

## [Unreleased]

### Planned for Future Releases (Phase 4: Ecosystem)

**v0.2.0+ (Feature Releases)** - As features are ready:
- Database middleware (PostgreSQL, MySQL, SQLite)
- Cache middleware (Redis, Memcached)
- Additional plugins and integrations
- Community-requested features

**v1.0.0 LTS** - After 6-12+ months of production usage:
- API stability guarantees
- Long-term support (3+ years)
- Enterprise-grade reliability
- No breaking changes in v1.x.x

See [ROADMAP.md](ROADMAP.md) for detailed plans and timelines.

---

[0.1.0]: https://github.com/coregx/fursy/releases/tag/v0.1.0
