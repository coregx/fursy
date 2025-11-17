# Release Guide - Git-Flow Standard Process

**CRITICAL**: Read this guide BEFORE creating any release!

> **Universal guide** for releasing Go projects using git-flow workflow

---

## 🔴 CRITICAL: Backup Before Any Operation

**ALWAYS create a backup before any serious operations!**

**Linux/macOS**:
```bash
# Directory backup
cp -r fursy fursy-backup-$(date +%Y%m%d-%H%M%S)

# Git bundle (portable, cross-platform)
git bundle create ../fursy-backup.bundle --all
```

**Windows (PowerShell)**:
```powershell
# Directory backup
Copy-Item -Recurse fursy "fursy-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"

# Git bundle (portable, cross-platform)
git bundle create ..\fursy-backup.bundle --all
```

**Git bundle** is recommended - portable, cross-platform, space-efficient!

**Dangerous operations (require backup)**:
- `git reset --hard`
- `git branch -D`
- `git tag -d`
- `git push -f`
- `git rebase`
- Any rollback/deletion operations

---

## 🎯 Git Flow Strategy

### Branch Structure

```
main        - Production-ready code ONLY (protected, green CI always)
  ↑
release/*   - Release candidates (RC)
  ↑
develop     - Active development (default branch for PRs)
  ↑
feature/*   - Feature branches
```

### Branch Rules

#### `main` Branch
- ✅ **ALWAYS** production-ready
- ✅ **ALWAYS** green CI (all tests passing)
- ✅ **ONLY** accepts merges from `release/*` branches
- ❌ **NEVER** commit directly to main
- ❌ **NEVER** push without green CI
- ❌ **NEVER** force push
- 🏷️ **Tags created ONLY after CI passes**

#### `develop` Branch
- Default branch for development
- Accepts feature branches
- May contain work-in-progress code
- Should pass tests, but can have warnings
- **Current default branch**

#### `release/*` Branches
- Format: `release/v0.4.0`, `release/v0.5.0`
- Created from `develop`
- Only bug fixes and documentation updates allowed
- No new features
- Merges to both `main` and `develop`

#### `feature/*` Branches
- Format: `feature/openapi-generation`, `feature/prometheus-middleware`
- Created from `develop`
- Merged back to `develop` with `--squash` (1 clean commit per feature)

---

## 🔀 Merge Strategy (Git-Flow Standard)

### When to Use --squash vs --no-ff

**Use `--squash` (feature → develop)**:
```bash
# Feature branches: many WIP commits → 1 clean commit
git checkout develop
git merge --squash feature/my-feature
git commit -m "feat: implement my feature

- Component 1
- Component 2
- Component 3"
```

**Why squash features?**:
- Keeps develop history clean (5-10 commits per release)
- Prevents 100+ WIP commits cluttering develop
- Each feature = 1 logical commit
- Makes git log readable

**Use `--no-ff` (release → main, main → develop)**:
```bash
# Release branches: preserve complete history
git checkout main
git merge --no-ff release/v0.4.0
# Note: -m message is optional here, git will auto-generate merge commit

# Merge back to develop
git checkout develop
git merge --no-ff main -m "Merge release v0.4.0 back to develop"
```

**Why --no-ff for releases?**:
- Standard git-flow practice (official workflow)
- Preserves all release preparation commits
- Allows proper tag placement
- Enables clean merge back to develop
- Shows clear release boundaries in history

**NEVER Use --squash (release → main)**:
- ❌ Breaks git-flow (main ← develop merge conflicts)
- ❌ Loses release preparation history
- ❌ Makes merge back to develop difficult
- ❌ Not standard practice

---

## 🔧 Pre-Release Validation Script

### Location
`scripts/pre-release-check.sh`

### Purpose
Runs **all quality checks locally** before creating a release, matching CI requirements exactly.

### When to Use

#### 1. Before Every Commit (Recommended)
```bash
# Quick validation before committing
bash scripts/pre-release-check.sh

# If script passes (green/yellow), safe to commit:
git add .
git commit -m "..."
git push
```

#### 2. Before Creating Release Branch (Mandatory)
```bash
# MUST pass before starting release process
bash scripts/pre-release-check.sh

# Only proceed if output shows:
# ✅ "All checks passed! Ready for release."
```

#### 3. Before Merging to Main (Mandatory)
```bash
# Final validation on release branch
git checkout release/v0.4.0
bash scripts/pre-release-check.sh

# If errors found, fix them before merging to main
```

#### 4. After Major Changes (Recommended)
- After refactoring
- After dependency updates
- After documentation updates
- After fixing bugs

### What the Script Validates

1. **Go version**: 1.25+ required
2. **Code formatting**: `gofmt -l .` must be clean
3. **Static analysis**: `go vet ./...` must pass
4. **Build**: `go build ./...` must succeed
5. **go.mod**: `go mod verify` and `go mod tidy` check
6. **Tests**: All tests passing (with race detector)
7. **Coverage**: >85% (Phase 1), >90% (Phase 2+)
8. **golangci-lint**: 0 issues required (strict linting)
9. **Benchmarks**: Performance regression check
10. **Documentation**: All critical files present

### Exit Codes

- **0 (green)**: All checks passed, ready for release
- **0 (yellow)**: Checks passed with warnings (review recommended)
- **1 (red)**: Checks failed with errors (must fix before release)

### Example Output

```bash
$ bash scripts/pre-release-check.sh

========================================
  FURSY HTTP Router - Pre-Release Check
========================================

[INFO] Checking Go version...
[SUCCESS] Go version: go1.25.1

[INFO] Checking code formatting (gofmt -l .)...
[SUCCESS] All files are properly formatted

[INFO] Running golangci-lint...
[SUCCESS] golangci-lint passed with 0 issues

[INFO] Running tests with race detector...
[SUCCESS] All tests passed (88.9% coverage)

[INFO] Running benchmarks...
[SUCCESS] Static routes: 256 ns/op, 1 alloc/op ✅
[SUCCESS] Parametric routes: 326 ns/op, 1 alloc/op ✅

========================================
  Summary
========================================

[SUCCESS] All checks passed! Ready for release.

Next steps (from RELEASE_GUIDE.md):
  1. Create release branch: git checkout -b release/v0.4.0
  2. Update CHANGELOG.md with version details
  ...
```

### Warnings vs Errors

**Warnings (yellow)** - Non-blocking, but review recommended:
- Uncommitted changes detected
- Test coverage slightly below target
- Minor benchmark regressions (<5%)

**Errors (red)** - Blocking, must fix:
- Code not formatted
- go vet failures
- Build failures
- Test failures
- golangci-lint issues
- Coverage significantly below target
- Performance regressions (>10%)
- Missing documentation files

---

## 📋 Version Naming

### Semantic Versioning

Format: `MAJOR.MINOR.PATCH[-PRERELEASE]`

**For 0.x.y versions (pre-1.0)**:
- `0.y.0` - New features (minor bump)
- `0.y.z` - Bug fixes, hotfixes (patch bump)

**For 1.x.y+ versions (stable API)**:
- `x.0.0` - Breaking changes (requires v2+ module path)
- `x.y.0` - New features (backwards-compatible)
- `x.y.z` - Bug fixes only

Examples:
- `v0.4.0` - New features (Database middleware, Cache middleware)
- `v0.4.1` - Bug fix for v0.4.0
- `v0.5.0` - Next feature release
- `v1.0.0` - TBD (after 6-12 months production usage)

### FURSY Versioning Strategy

**Current Path**: `v0.y.z` until `v1.0.0`

**Released versions**:
- `v0.1.0` - Phase 0-3 (Foundation + Production Features) - Radix tree routing, middleware pipeline, RFC 9457, generics, JWT auth, rate limiting, circuit breaker, context pooling, graceful shutdown
- `v0.2.0` - Phase 4.1 (Documentation & Examples) - Validator plugin, 11 comprehensive examples, middleware/validation/content-negotiation/observability documentation, AI agent guide (llms.md)

**Future releases** (Phase 4 - Ecosystem):
- `v0.3.0+` - Future ecosystem components (Database middleware, Cache middleware, Documentation website, Migration guides)
- `v1.0.0` - TBD (stable API, production-proven, 6-12 months validation)

**Rationale**:
- Stay on 0.x.y to allow API evolution
- v1.0.0 = long-term stability commitment (6-12 months production usage first)
- Breaking changes allowed in 0.x versions
- Avoid v2.0.0 (requires new import path in Go)

---

## ✅ Pre-Release Checklist

**CRITICAL**: Complete ALL items before creating release branch!

### 1. Automated Quality Checks

**Run our pre-release validation script**:

```bash
# ONE COMMAND runs ALL checks (matches CI exactly)
bash scripts/pre-release-check.sh
```

This script validates:
- ✅ Go version (1.25+)
- ✅ Code formatting (gofmt)
- ✅ Static analysis (go vet)
- ✅ All tests passing
- ✅ Race detector
- ✅ Coverage >85% (Phase 1), >90% (Phase 2+)
- ✅ golangci-lint (0 issues required)
- ✅ go.mod integrity
- ✅ Benchmarks (performance regression check)
- ✅ All documentation present

**Manual checks** (if script not available):

```bash
# Format code
go fmt ./...

# Verify formatting
if [ -n "$(gofmt -l .)" ]; then
  echo "ERROR: Code not formatted"
  gofmt -l .
  exit 1
fi

# Static analysis
go vet ./...

# Linting (strict)
golangci-lint run --timeout=5m ./...
# Must show: "0 issues."

# All tests
go test -race ./...
# All must PASS

# Coverage check
go test -coverprofile=coverage.txt ./...
go tool cover -func=coverage.txt | tail -1
# Minimum: >85% (Phase 1), >90% (Phase 2+)

# Benchmarks
go test -bench=. -benchmem ./...
# Check for regressions
```

### 2. Dependencies

```bash
# Verify modules
go mod verify

# Tidy and check diff
go mod tidy
git diff go.mod go.sum
# Should show NO changes

# Check dependencies (core = stdlib only!)
go list -m all | grep -v indirect
# Core package should show NO external dependencies
# Plugins may have dependencies
```

### 3. Documentation

- [ ] README.md updated with latest features
- [ ] CHANGELOG.md entry created for this version
- [ ] All public APIs have godoc comments
- [ ] Examples are up-to-date and tested
- [ ] PERFORMANCE.md updated with latest benchmarks
- [ ] Migration guide (if breaking changes in 0.x)
- [ ] Known limitations documented

### 4. GitHub Actions

- [ ] `.github/workflows/*.yml` exist
- [ ] All workflows tested on `develop`
- [ ] CI passes on latest `develop` commit
- [ ] Coverage badge updated (if changed)

### 5. Project-Specific Checks

**FURSY HTTP Router Requirements**:
- [ ] All Phase tasks complete (check STATUS.md)
- [ ] Test coverage >85% (Phase 1), >90% (Phase 2+)
- [ ] Benchmarks show no regressions
- [ ] Performance targets met:
  - Static routes: <300 ns/op
  - Parametric routes: <500 ns/op
  - Allocations: 1 alloc/op (routing hot path)
- [ ] Minimal dependencies verified (core routing: stdlib only, middleware: jwt + x/time)
- [ ] Uses `encoding/json/v2` (NOT `encoding/json`)
- [ ] Uses `log/slog` for logging
- [ ] RFC 9457 compliance (error responses)
- [ ] No regressions in existing features

---

## 🚀 Release Process (Git-Flow Standard 2025)

### 🔴 CRITICAL: Documentation Updates Location

**IMPORTANT**: All documentation updates (README.md, CHANGELOG.md, PERFORMANCE.md, docs/*.md) **MUST be done in the release branch**, NOT in develop!

**Why?**:
- Release branch is for preparing the release (bumping versions, updating docs, bug fixes)
- Develop is for feature development
- This is standard git-flow practice

**Workflow**:
```
develop (features only)
   ↓
release/v0.4.0 (version bumps, docs, fixes ONLY)
   ↓
main (production)
```

### Step 1: Pre-Release Validation (In Develop)

```bash
# BEFORE creating release branch, validate develop
git checkout develop
git pull origin develop

# Run ALL pre-release checks (CRITICAL!)
bash scripts/pre-release-check.sh
# Script must exit with: "All checks passed! Ready for release."
# If errors: FIX THEM in develop before proceeding!

# Verify develop is clean
git status
# Should show: "nothing to commit, working tree clean"
```

### Step 2: Create Release Branch

```bash
# Create release branch from develop (example: v0.4.0)
git checkout -b release/v0.4.0 develop

# ⚠️ IMPORTANT: Now update ALL documentation IN THIS BRANCH:
# - README.md (version badges, features)
# - CHANGELOG.md (add v0.4.0 section with date)
# - PERFORMANCE.md (update benchmarks)
# - STATUS.md (update current version, phase status)
# - docs/*.md (update version references)

# Example documentation updates:
# 1. README.md - Update badge versions
# 2. CHANGELOG.md - Add release section:
#    ## [0.4.0] - 2025-02-15
#    ### Added
#    - Database middleware with connection pooling
#    - Documentation website at fursy.coregx.dev
#    ### Fixed
#    - Bug in parameter extraction for edge cases
# 3. PERFORMANCE.md - Update benchmarks
# 4. STATUS.md - Update current version and phase

# Commit ALL documentation changes in release branch
git add .
git commit -m "chore: prepare v0.4.0 release

- Update README.md version badges and features
- Add CHANGELOG.md entry for v0.4.0
- Update PERFORMANCE.md with latest benchmarks
- Update STATUS.md current version
- Update version references throughout documentation"

# Push release branch
git push origin release/v0.4.0
```

### Step 3: Wait for CI (CRITICAL!)

```bash
# Go to GitHub Actions and WAIT for green CI
# URL: https://github.com/coregx/fursy/actions
```

**⏸️ STOP HERE! Do NOT proceed until CI is GREEN!**

✅ **All checks must pass:**
- Unit tests (Linux, macOS, Windows)
- Linting (golangci-lint)
- Code formatting (gofmt)
- Coverage check (>85%)
- Race detector

❌ **If CI fails:**
1. Fix issues in `release/v0.4.0` branch
2. Commit fixes
3. Push and wait for CI again
4. Repeat until GREEN

### Step 4: Merge to Main (After Green CI)

**🔴 CRITICAL RULE**: Release branches merge to main with `--no-ff` (NOT --squash!)

**Why `--no-ff` for releases?**:
- Preserves complete release history
- Standard git-flow practice
- Allows proper merge back to develop
- Tags point to actual release commits

**Why NOT `--squash`?**:
- Squash is for feature branches → develop
- Release → main uses `--no-ff` to preserve history
- This is the official git-flow standard

```bash
# ONLY after CI is green on release branch!
git checkout main
git pull origin main

# ⚠️ IMPORTANT: Use --no-ff (NOT --squash) for release merges!
git merge --no-ff release/v0.4.0 -m "Release v0.4.0

Database Middleware & Documentation Website

Features:
- Database middleware with connection pooling (PostgreSQL, MySQL, SQLite)
- Documentation website at fursy.coregx.dev
- Migration guides from Gin, Echo, Chi
- OpenAPI 3.1 integration examples

Bug Fixes:
- Fixed parameter extraction for edge cases with special characters
- Fixed context pooling race condition
- Fixed middleware chain execution order

Performance:
- Static routes: 256 ns/op, 1 alloc/op ✅
- Parametric routes: 326 ns/op, 1 alloc/op ✅
- Coverage: 90.2% (target: >90%) ✅

Quality Metrics:
- Linter: 0 issues ✅
- Tests: All passing ✅
- Race detector: Clean ✅
- Benchmarks: No regressions ✅

Technical Details:
- Minimal dependencies (core routing: stdlib only, middleware: jwt + x/time)
- Uses encoding/json/v2 and log/slog
- RFC 9457 compliant error responses
- Full backwards compatibility with v0.3.x"

# Push to main
git push origin main
```

### Step 5: Wait for CI on Main

```bash
# Go to GitHub Actions and verify main branch CI
# https://github.com/coregx/fursy/actions

# WAIT for green CI on main branch!
```

**⏸️ STOP! Do NOT create tag until main CI is GREEN!**

### Step 6: Create Tag (After Green CI on Main)

```bash
# ONLY after main CI is green!

# Create annotated tag
git tag -a v0.4.0 -m "Release v0.4.0

FURSY HTTP Router v0.4.0 - Database Middleware & Ecosystem

Features:
- Database middleware with connection pooling
  - PostgreSQL, MySQL, SQLite support
  - Transaction management
  - Query builder integration
- Documentation website (fursy.coregx.dev)
  - Getting started guide
  - API reference
  - Best practices & FAQ
- Migration guides from popular frameworks
  - From Gin (5-minute migration)
  - From Echo (10-minute migration)
  - From Chi (seamless transition)

Performance:
- Static routes: 256 ns/op, 1 alloc/op
- Parametric routes: 326 ns/op, 1 alloc/op
- Deep nesting (4 params): 561 ns/op, 1 alloc/op
- Throughput: ~10M req/s (simple routes)
- Memory-efficient context pooling

Quality:
- Test coverage: 90.2% overall
- golangci-lint: 0 issues
- Race detector: Clean
- Minimal dependencies (core routing: stdlib only)
- Production-ready stability

API Stability:
- Backwards compatible with v0.3.x
- No breaking changes
- Deprecated APIs: None
- New APIs: Database middleware, plugin system

Known Limitations:
- Database middleware requires plugin installation
- OpenAPI generation requires validator plugin
- Write support for complex types (planned for v0.5.0)

Next Release (v0.5.0):
- Cache middleware (Redis, Memcached)
- Rate limiting improvements
- Advanced monitoring (Prometheus, OpenTelemetry)

See CHANGELOG.md for complete details."

# Push tag
git push origin v0.4.0
```

### Step 7: Merge Back to Develop

```bash
# Keep develop in sync
git checkout develop
git merge --no-ff release/v0.4.0 -m "Merge release v0.4.0 back to develop"
git push origin develop

# Delete release branch (optional, after confirming release is good)
git branch -d release/v0.4.0
git push origin --delete release/v0.4.0
```

### Step 8: Create GitHub Release

1. Go to: https://github.com/coregx/fursy/releases/new
2. Select tag: `v0.4.0`
3. Release title: `v0.4.0 - Database Middleware & Documentation Website`
4. Description: Copy from CHANGELOG.md
5. **Do NOT check "Set as a pre-release"** (0.x versions are development, not pre-release)
6. Click "Publish release"

---

## 🔥 Hotfix Process

For critical bugs in production (`main` branch):

```bash
# Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/v0.4.1

# Fix the bug
# ... make changes ...

# Test thoroughly
go test -race ./...
golangci-lint run ./...
bash scripts/pre-release-check.sh

# Commit
git add .
git commit -m "fix: critical bug in route matching"

# Push and wait for CI
git push origin hotfix/v0.4.1

# WAIT FOR GREEN CI!

# Merge to main
git checkout main
git merge --no-ff hotfix/v0.4.1 -m "Hotfix v0.4.1 - Fix critical route matching bug"
git push origin main

# WAIT FOR GREEN CI ON MAIN!

# Create tag
git tag -a v0.4.1 -m "Hotfix v0.4.1

Critical bug fix for route matching with special characters.

Bug Fixes:
- Fixed route matching failure when path contains URL-encoded characters
- Fixed parameter extraction for paths with trailing slashes

Impact: High (affects all parametric routes with special characters)
Urgency: Critical (production issue)

Tested:
- All existing tests passing
- Added regression test for URL-encoded paths
- Verified with production-like workload

See CHANGELOG.md for details."

git push origin v0.4.1

# Merge back to develop
git checkout develop
git merge --no-ff hotfix/v0.4.1 -m "Merge hotfix v0.4.1"
git push origin develop

# Delete hotfix branch
git branch -d hotfix/v0.4.1
git push origin --delete hotfix/v0.4.1
```

---

## 📊 CI Requirements

### Must Pass Before Release

All GitHub Actions workflows must be GREEN:

1. **Unit Tests** (3 platforms)
   - Linux (ubuntu-latest)
   - macOS (macos-latest)
   - Windows (windows-latest)
   - Go version: 1.25

2. **Code Quality**
   - go vet (no errors)
   - golangci-lint (0 issues)
   - gofmt (all files formatted)

3. **Coverage**
   - Phase 1: ≥85%
   - Phase 2+: ≥90%
   - Core packages: ≥95%

4. **Race Detection**
   - go test -race ./... (no data races)

5. **Benchmarks**
   - No performance regressions (>10% slower)
   - Allocation counts within target (1 alloc/op)

---

## 🚫 NEVER Do This

❌ **NEVER commit directly to main**
```bash
# WRONG!
git checkout main
git commit -m "quick fix"  # ❌ NO!
```

❌ **NEVER push to main without green CI**
```bash
# WRONG!
git push origin main  # ❌ WAIT for CI first!
```

❌ **NEVER create tags before CI passes**
```bash
# WRONG!
git tag v0.4.0  # ❌ WAIT for green CI on main!
git push origin v0.4.0
```

❌ **NEVER force push to main or develop**
```bash
# WRONG!
git push -f origin main  # ❌ NEVER!
```

❌ **NEVER skip lint or format checks**
```bash
# WRONG!
git commit -m "skip CI" --no-verify  # ❌ NO!
```

❌ **NEVER push without running lint locally**
```bash
# WRONG WORKFLOW:
git commit -m "feat: something"
git push  # ❌ Run lint FIRST!

# CORRECT WORKFLOW:
bash scripts/pre-release-check.sh  # ✅ Check FIRST
git commit -m "feat: something"
git push
```

❌ **NEVER add dependencies to core package**
```bash
# WRONG!
go get github.com/some/library  # ❌ Core = stdlib only!

# CORRECT (if needed):
# Add to plugins/ package, NOT core
```

---

## ✅ Always Do This

✅ **ALWAYS run checks before commit**
```bash
# Recommended: Use our pre-release script
bash scripts/pre-release-check.sh

# Or manual workflow:
go fmt ./...
golangci-lint run ./...
go test -race ./...
git add .
git commit -m "..."
git push
```

✅ **ALWAYS wait for green CI before proceeding**
```bash
# Correct workflow:
git push origin release/v0.4.0
# ⏸️ WAIT for green CI
git checkout main
git merge --no-ff release/v0.4.0
git push origin main
# ⏸️ WAIT for green CI on main
git tag -a v0.4.0 -m "..."
git push origin v0.4.0
```

✅ **ALWAYS use annotated tags**
```bash
# Good
git tag -a v0.4.0 -m "Release v0.4.0"

# Bad
git tag v0.4.0  # Lightweight tag
```

✅ **ALWAYS update CHANGELOG.md**
- Document all changes
- Include breaking changes (if any in 0.x)
- Add performance metrics
- Reference task completion

✅ **ALWAYS test on all platforms locally if possible**
```bash
# At minimum:
go test -race ./...
golangci-lint run ./...
go mod verify
bash scripts/pre-release-check.sh
```

✅ **ALWAYS use encoding/json/v2**
```go
// ✅ CORRECT
import "encoding/json/v2"

// ❌ WRONG
import "encoding/json"
```

✅ **ALWAYS use log/slog**
```go
// ✅ CORRECT
import "log/slog"

slog.Info("request processed", "method", req.Method)

// ❌ WRONG
import "log"
log.Printf("request processed")
```

---

## 📝 Release Checklist Template

Copy this for each release:

```markdown
## Release v0.4.0 Checklist

### Pre-Release
- [ ] All tests passing locally (`go test -race ./...`)
- [ ] Code formatted (`go fmt ./...`, `gofmt -l .` = empty)
- [ ] Linter clean (`golangci-lint run ./...` = 0 issues)
- [ ] Dependencies verified (`go mod verify`)
- [ ] Pre-release script passed (`bash scripts/pre-release-check.sh`)
- [ ] CHANGELOG.md updated
- [ ] PERFORMANCE.md updated with latest benchmarks
- [ ] STATUS.md updated with current version
- [ ] README.md updated (if needed)
- [ ] Phase tasks complete (check STATUS.md)

### Release Branch
- [ ] Created release/v0.4.0 from develop
- [ ] Pushed to GitHub
- [ ] CI GREEN on release branch
- [ ] All checks passed (tests, lint, format, coverage, benchmarks)

### Main Branch
- [ ] Merged release branch to main (`--no-ff`)
- [ ] Pushed to origin
- [ ] CI GREEN on main
- [ ] All checks passed

### Tagging
- [ ] Created annotated tag v0.4.0
- [ ] Tag message includes full changelog
- [ ] Pushed tag to origin
- [ ] GitHub release created

### Cleanup
- [ ] Merged back to develop
- [ ] Deleted release branch
- [ ] Verified pkg.go.dev updated
- [ ] Announced release (if applicable)
```

---

## 🎯 Summary: Golden Rules

1. **main = Production ONLY** - Always green CI, always stable
2. **Wait for CI** - NEVER proceed without green CI
3. **Tags LAST** - Only after main CI is green
4. **No Direct Commits** - Use release branches
5. **Annotated Tags** - Always use `git tag -a`
6. **Full Testing** - Run `bash scripts/pre-release-check.sh` before commit
7. **Document Everything** - Update CHANGELOG.md, PERFORMANCE.md, STATUS.md
8. **Git Flow** - develop → release/* → main → tag
9. **Check Lint ALWAYS** - `golangci-lint run ./...` before every push
10. **Zero Dependencies** - Core = stdlib only (plugins can have deps)

---

## 🔧 FURSY-Specific Guidelines

### Before Release

**Performance Validation**:
- [ ] Benchmarks show no regressions (>10% slower = blocker)
- [ ] Static routes: <300 ns/op
- [ ] Parametric routes: <500 ns/op
- [ ] Allocations: 1 alloc/op (routing hot path)
- [ ] Throughput: >1M req/s (with middleware)

**Code Quality**:
- [ ] Uses `encoding/json/v2` (NOT `encoding/json`)
- [ ] Uses `log/slog` for all logging
- [ ] Minimal dependencies verified (core routing: stdlib only, middleware: jwt + x/time)
- [ ] All exported APIs have godoc comments
- [ ] RFC 9457 compliant error responses

**Testing**:
- [ ] Test coverage >85% (Phase 1), >90% (Phase 2+)
- [ ] Race detector clean (`go test -race ./...`)
- [ ] All examples tested and working
- [ ] Cross-platform tests passing (Linux, macOS, Windows)

**Documentation**:
- [ ] PERFORMANCE.md updated with latest benchmarks
- [ ] API changes documented in CHANGELOG.md
- [ ] Migration guide (if breaking changes in 0.x)
- [ ] Examples updated for new features

---

**Remember**: A release can always wait. A broken production release cannot be undone.

**When in doubt, wait for CI!**

**Always run lint before push!**

---

*Last Updated: 2025-11-16*
*FURSY HTTP Router Release Process*
*Git-Flow Standard 2025 - Adapted from scigolibs/hdf5*
