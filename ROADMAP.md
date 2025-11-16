# FURSY HTTP Router - Development Roadmap

> **Strategic Advantage**: Modern Go 1.25+ features + proven routing patterns!
> **Approach**: Combine best practices from httprouter, Gin, Echo with type-safe generics

**Last Updated**: 2025-11-16 | **Current Version**: v0.1.0 (Production Ready) | **Phase**: Phase 3 Complete ✅ → Phase 4 Ready (Ecosystem) | **Target**: v1.0.0 LTS (TBD, after full API stabilization)

---

## 🎯 Vision

Build a **next-generation HTTP router for Go 1.25+** with type-safe handlers, RFC 9457 error responses, and built-in OpenAPI 3.1 generation - all with **minimal dependencies** (core routing: stdlib only).

### Key Advantages

✅ **Type-Safe Generic Handlers**
- Box[Req, Res any] for compile-time safety
- No interface{} casting needed
- Modern Go 1.25+ features

✅ **Production-Ready Performance**
- <500ns parametric route lookup
- 1 allocation per request (routing hot path)
- ~10M requests/sec throughput
- Zero-allocation context pooling

✅ **RFC 9457 Problem Details**
- Standardized error responses
- Native support, no external library
- OpenAPI-compatible error schema

✅ **Zero Dependencies (Core)**
- Core package = stdlib only
- Uses `encoding/json/v2` and `log/slog`
- Plugins can have dependencies

---

## 🚀 Version Strategy

### Philosophy: Foundation → API Excellence → Production → Ecosystem → Stable

```
Phase 0 (Setup) ✅ COMPLETE
      ↓
v0.1.0 (Foundation + Features) ← CURRENT RELEASE (Phase 0-3)
      ↓ Radix tree, middleware, groups, generics, RFC 9457,
      ↓ Auth, rate limiting, circuit breaker, pooling
v0.2.0+ (New Features) ← FUTURE (Phase 4 - Q1-Q3 2026)
      ↓ Plugins, docs, community
v0.x.x (Stable Maintenance)
      ↓ (6-9 months production validation)
v1.0.0 LTS → Long-term support (Q3 2026)
```

### Critical Milestones

**v0.1.0 (Phase 0-3)** = Foundation + Production Features ← CURRENT (First Public Release)
- Radix tree routing with zero-allocation lookup
- Middleware pipeline (Next/Abort pattern)
- Route groups with middleware inheritance
- Generic type-safe Box[Req, Res]
- RFC 9457 Problem Details (native)
- OpenAPI spec foundations
- **JWT Authentication** (94.2% coverage)
- **Rate Limiting** (94.4% coverage, token bucket algorithm)
- **Security Headers** (100% coverage, OWASP 2025)
- **Circuit Breaker** (95.5% coverage, zero deps, production-ready)
- **Graceful Shutdown** (100% core methods, Kubernetes-ready)
- **Context Pooling** (1 alloc/op, ~10M req/s throughput)
- **Performance benchmarks** (PERFORMANCE.md with 19 benchmarks)

**Performance Metrics** (Phase 3):
- Static routes: 256 ns/op, 1 alloc/op ✅
- Parametric routes: 326 ns/op, 1 alloc/op ✅
- Deep nesting (4 params): 561 ns/op, 1 alloc/op ✅
- Coverage: 91.7% ✅
- Status: **Production-ready** 🚀

**v0.2.0+ (Phase 4)** = Ecosystem (Future - Q1-Q3 2026)
- Database middleware (PostgreSQL, MySQL, SQLite)
- Cache middleware (Redis, Memcached)
- Documentation website (fursy.coregx.dev)
- Migration guides (from Gin, Echo, Chi, ozzo-routing)
- Examples repository (Hello World, REST API, Microservices)
- Community building (Discord, GitHub Discussions)
- Releases driven by feature readiness (NOT schedule!)

**v1.0.0** = Production-validated with LTS guarantee (Q3 2026)
- API stability commitment (no breaking changes in v1.x.x)
- 6-12 months production usage validation
- Long-term support (3+ years)
- Comprehensive ecosystem

**Why 0.x.y strategy?**:
- Allows API evolution without major version bumps
- v1.0.0 = serious long-term API stability commitment
- 0.y.0 = new features, 0.y.z = bug fixes
- Avoids v2.0.0 (requires new Go module path)

**See**: `.claude/STATUS.md` and `docs/dev/04_IMPLEMENTATION_PLAN.md` for complete details

---

## 📊 Current Status (v0.1.0 - Production Ready)

**Phase**: 🚀 Phase 3 Complete → Ready for Phase 4 (Ecosystem)
**Performance**: Production-ready! (256-326 ns/op, 1 alloc/op)
**Coverage**: 91.7% (target: >85% Phase 1, >90% Phase 2+)

**What Works**:
- ✅ **Radix tree routing** (zero-allocation lookup, <500ns parametric routes)
- ✅ **Generic type-safe handlers** (Box[Req, Res any])
- ✅ **Middleware pipeline** (Next/Abort, pre-allocated buffers)
- ✅ **Route groups** (nested groups, middleware inheritance)
- ✅ **RFC 9457 Problem Details** (standardized error responses)
- ✅ **JWT Authentication** (token validation, claims extraction)
- ✅ **Rate Limiting** (token bucket, per-IP/per-user)
- ✅ **Security Headers** (OWASP 2025, CSP, HSTS, X-Frame-Options)
- ✅ **Circuit Breaker** (failure threshold, auto-recovery, zero deps)
- ✅ **Graceful Shutdown** (connection draining, Kubernetes signals)
- ✅ **Context Pooling** (sync.Pool, 1 alloc/op, memory-efficient)

**Performance**:
- ✅ Static routes: 256 ns/op, 1 alloc/op
- ✅ Parametric routes: 326 ns/op, 1 alloc/op
- ✅ Throughput: ~10M req/s (simple routes)
- ✅ Memory efficient: context pooling prevents leaks

**Validation**:
- ✅ 91.7% test coverage (exceeded Phase 2 target of 88%)
- ✅ 0 linter issues (34+ linters via golangci-lint)
- ✅ Race detector clean
- ✅ Cross-platform (Linux, macOS, Windows)

**Documentation**:
- ✅ PERFORMANCE.md (19 benchmarks, optimization details)
- ✅ CONTRIBUTING.md (git-flow workflow)
- ✅ RELEASE_GUIDE.md (release process)
- ✅ SECURITY.md (security best practices)

**History**: See [CHANGELOG.md](CHANGELOG.md) for complete release history

---

## 📅 What's Next

### **Phase 4 (Current) - Ecosystem Building** (2025-2026, Feature-Driven)

**Goal**: Community, ecosystem, and production adoption

**Duration**: 8-12 weeks (as features are ready, NOT on schedule!)

**Scope**:

**Week 1-3: Plugin Ecosystem**
- [ ] Plugin development guide
- [ ] Plugin registry system
- [ ] Database middleware (PostgreSQL, MySQL, SQLite)
  - Connection pooling
  - Transaction management
  - Query builder integration
- [ ] Cache middleware (Redis, Memcached)
  - TTL support
  - Cache invalidation
  - Distributed caching
- [x] OpenTelemetry plugin (92% coverage, tracing + metrics)
- [x] Prometheus metrics plugin (HTTP semantic conventions)
- [x] Validator plugin (go-playground/validator/v10)

**Week 4-6: Documentation & Examples**
- [ ] Documentation website (fursy.coregx.dev)
  - Getting started guide
  - API reference (auto-generated from godoc)
  - Middleware guide
  - Best practices & FAQ
  - Performance tuning guide
- [ ] Examples repository
  - Hello World (minimal setup)
  - REST API (CRUD operations)
  - gRPC Gateway (protocol buffers)
  - Microservice template (production-ready)
  - Production setup (Docker, Kubernetes)
- [ ] Migration guides
  - From Gin (5-minute migration)
  - From Echo (10-minute migration)
  - From Chi (seamless transition)
  - From ozzo-routing (direct replacement)

**Week 7-9: Community Building**
- [ ] GitHub Discussions setup
- [ ] Discord server (real-time support)
- [ ] Twitter/X account (announcements)
- [ ] Contributing guide (enhanced)
- [ ] Blog post announcement
- [ ] Reddit r/golang post
- [ ] HackerNews submission
- [ ] Conference talk submission
- [ ] Video tutorials (YouTube)
- [ ] Sample projects (real-world use cases)
- [ ] Benchmarks vs competitors (Gin, Echo, Fiber, httprouter)

**Week 10-12: Ecosystem Launch**
- [ ] Security audit (third-party)
- [ ] Final performance benchmarks
- [ ] Documentation review
- [ ] Example testing
- [ ] CHANGELOG updates
- [ ] Release notes (v0.4.0+)
- [ ] Migration guides published
- [ ] Announcement blog post
- [ ] Tag ecosystem releases
- [ ] Publish to pkg.go.dev
- [ ] Social media announcements
- [ ] Submit to Awesome Go

**Versioning Strategy** (Phase 4):
- **0.y.0** - New features (Database middleware = v0.2.0, Cache middleware = v0.3.0, etc.)
- **0.y.z** - Bug fixes, hotfixes, security patches
- **v1.0.0** = TBD (after full API stabilization, 6-12+ months)

**Important**: Releases are feature-driven, NOT schedule-driven! Phase 4 ≠ "mandatory 6 releases".

**Success Criteria** (Phase 4):
- ✅ Active community (Discord, GitHub Discussions)
- ✅ 1,000+ GitHub stars
- ✅ 10+ production deployments
- ✅ 5+ community contributors
- ✅ Comprehensive documentation website
- ✅ Migration guides from popular frameworks
- ✅ Ecosystem releases as features are ready (v0.y.0)

**Note on v1.0.0**:
- v1.0.0 = separate milestone, NOT a Phase 4 goal!
- Will release when API is fully stable
- After minimum 6-12 months of production usage
- When we're confident no breaking changes are needed

---

### **v0.x.x - Stable Maintenance Phase** (Ongoing)

**Goal**: Production validation, stability, and community support

**Scope**:
- 🐛 Bug fixes from production use (high priority)
- 🛡️ Security updates (critical priority)
- ⚡ Performance optimizations based on profiling
- 📝 Documentation improvements from user feedback
- ✨ Minor feature enhancements (community-driven)
- ⛔ NO breaking API changes

**Community Adoption**:
- 👥 Real-world project validation
- 📊 Performance benchmarks and profiling
- 🔍 Edge case discovery and handling
- 💬 API refinement suggestions
- 🌐 Community engagement (Discord, Discussions)

**Quality Focus**:
- 📈 Maintain >90% test coverage
- 🔒 Zero security vulnerabilities
- ✅ All middleware production-tested
- 📋 Responsive issue triage and resolution

---

### **v1.0.0 - Long-Term Support Release** (TBD - After Full API Stabilization)

**Goal**: LTS release with stability guarantees

**Requirements**:
- v0.x.x stable for 6+ months
- Positive community feedback
- No critical bugs
- API proven in production
- 10+ production deployments
- Active community (Discord, Discussions)

**LTS Guarantees**:
- ✅ API stability (no breaking changes in v1.x.x)
- ✅ Long-term support (3+ years)
- ✅ Semantic versioning strictly followed
- ✅ Security updates and bug fixes
- ✅ Performance improvements (non-breaking)
- ✅ Backwards compatibility

**What v1.0.0 Means**:
- Serious commitment to API stability
- Breaking changes only in v2.0.0 (requires new module path)
- Production-ready for mission-critical applications
- Enterprise support considerations

---

## 🔮 Future Enhancements (Post-v1.0.0)

**Potential Focus Areas** (priority TBD based on community feedback):

**Performance Optimizations**:
- ⚡ Zero-allocation JSON encoding (custom encoder)
- 🧠 Header pooling (reduce response allocations)
- 📊 Parallel route matching (for very large route tables)
- 🔄 Custom memory allocator (arena-style)

**Advanced Features**:
- 📐 Websocket support
- 🗂️ Server-Sent Events (SSE)
- 🔗 HTTP/2 Server Push
- 📦 gRPC-Web gateway

**Developer Experience**:
- 🛠️ Code generation for type-safe clients
- 📚 Interactive examples (playground)
- 🧪 Testing utilities for users
- 📖 Video tutorials and courses

**Enterprise Features**:
- 🔍 Distributed tracing (OpenTelemetry deep integration)
- 📊 Advanced metrics (custom collectors)
- 🔒 OAuth2/OIDC provider integration
- 📈 Health checks and readiness probes

**Note**: Features will be prioritized based on:
1. Community requests and votes
2. Production use case needs
3. HTTP standards evolution
4. Maintainability and complexity

---

## 📚 Resources

**fursy Documentation**:
- README.md - Project overview
- CONTRIBUTING.md - How to contribute
- PERFORMANCE.md - Benchmark results
- RELEASE_GUIDE.md - Release process
- SECURITY.md - Security best practices

**Development**:
- .claude/STATUS.md - Current project status
- docs/dev/ - Development documentation
- examples/ - Usage examples

**Inspiration & References**:
- [httprouter](https://github.com/julienschmidt/httprouter) - Radix tree routing
- [Gin](https://github.com/gin-gonic/gin) - Middleware patterns
- [Echo](https://github.com/labstack/echo) - Context design
- [Fiber](https://github.com/gofiber/fiber) - Performance benchmarks
- [ozzo-routing](https://github.com/go-ozzo/ozzo-routing) - Sister project philosophy

**Standards**:
- [RFC 9457](https://datatracker.ietf.org/doc/html/rfc9457) - Problem Details for HTTP APIs
- [OpenAPI 3.1](https://spec.openapis.org/oas/v3.1.0) - API specification
- [OWASP 2025](https://owasp.org/) - Security best practices

---

## 📞 Support

**Documentation**:
- README.md - Overview and quick start
- examples/ - Working code examples
- CHANGELOG.md - Release history
- .claude/STATUS.md - Current development status

**Feedback**:
- GitHub Issues - Bug reports and feature requests
- GitHub Discussions - Questions and community help
- Discord - Real-time support (coming in Phase 4)

---

## 🔬 Development Approach

**Modern Go Best Practices**:
- Go 1.25+ features (generics, any type)
- stdlib v2 (`encoding/json/v2`, `log/slog`)
- Minimal dependencies (core routing: stdlib only)
- Pure Go (no CGo)
- TDD approach (tests first!)

**Performance-First**:
- Benchmark everything
- Profile before optimizing
- Zero-allocation hot paths
- Memory-efficient pooling
- Production-ready from day one

**Community-Driven**:
- Open development (all in GitHub)
- Responsive to feedback
- Transparent roadmap
- Collaborative decision-making

---

## 🎯 Success Metrics

**Phase 1 (Foundation)** ✅ ACHIEVED:
- ✅ Working radix tree routing
- ✅ <100ns route lookup
- ✅ >85% test coverage
- ✅ 0 linter issues

**Phase 2 (API Excellence)** ✅ ACHIEVED:
- ✅ Generic type-safe handlers
- ✅ RFC 9457 support
- ✅ >88% test coverage
- ✅ Enhanced middleware API

**Phase 3 (Production Features)** ✅ ACHIEVED:
- ✅ Auth, rate limiting, circuit breaker
- ✅ Context pooling (1 alloc/op)
- ✅ Graceful shutdown
- ✅ 91.7% test coverage
- ✅ Production-ready performance

**Phase 4 (Ecosystem)** - IN PROGRESS:
- [ ] 1,000+ GitHub stars
- [ ] 10+ production deployments
- [ ] 5+ community contributors
- [ ] Documentation website live
- [ ] Migration guides published
- [ ] Active Discord community

**v1.0.0 (LTS)**:
- [ ] 6+ months stable
- [ ] 10,000+ downloads
- [ ] No critical bugs
- [ ] Positive community sentiment
- [ ] Production-validated API

---

*Version 1.0 (Created 2025-11-16)*
*Current: v0.1.0 Production Ready | Phase: Phase 0-3 Complete ✅ | Next: v0.2.0+ (feature-driven) | Target: v1.0.0 LTS (TBD, after API stabilization)*
