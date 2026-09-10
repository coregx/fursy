# FURSY Performance Report

> **Generated**: 2026-09-10 (v0.5.0)
> **Platform**: Windows AMD64, Intel Core i7-1255U (12th Gen)
> **Go Version**: 1.27.1
> **Benchmark Duration**: default (1s+ per test, 3-5 runs)

---

## Key Metrics

FURSY achieves **true zero-allocation routing** on all route types. The param buffer is pre-allocated in `sync.Pool` and reused across requests — no heap allocation per request.

| Route type | Latency | Alloc | Throughput |
|---|---|---|---|
| Static | **53 ns/op** | **0 B, 0 alloc** | ~19M req/s |
| Root path | **46 ns/op** | **0 B, 0 alloc** | ~22M req/s |
| Parametric (1 param) | **64 ns/op** | **0 B, 0 alloc** | ~16M req/s |
| Multi-param (2 params) | **88 ns/op** | **0 B, 0 alloc** | ~11M req/s |
| Deep nesting (4 params) | **143 ns/op** | **0 B, 0 alloc** | ~7M req/s |
| Wildcard | **58 ns/op** | **0 B, 0 alloc** | ~17M req/s |
| Mixed routes (13 routes) | **84-109 ns/op** | **0 B, 0 alloc** | ~9-12M req/s |
| Context.Param() | **1.5 ns/op** | **0 B, 0 alloc** | — |

404 Not Found and 405 Method Not Allowed responses allocate 2x (response body write) — expected and unavoidable.

---

## Raw Benchmark Data (v0.5.0)

### Routing (full ServeHTTP pipeline)

```
BenchmarkRouter_StaticRoute-12                       20,122,753    53 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_RootPath-12                          24,247,959    46 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_LongStaticPath-12                    24,079,848    55 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_ParameterRoute-12                    17,183,679    64 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_ParameterRoute_MultipleParams-12     13,164,300    88 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_DeepNesting-12 (4 params)             8,562,687   143 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_WildcardRoute-12                     21,221,322    58 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_MultipleRoutes-12                    15,712,173    84 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_MixedRoutes-12                       10,354,315   109 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_TrailingSlash_ExactMatch-12          19,302,285    54 ns/op     0 B/op   0 allocs/op
BenchmarkRouter_TrailingSlash_Strip-12                5,695,316   229 ns/op   256 B/op   1 allocs/op
```

### Error Paths

```
BenchmarkRouter_NotFound-12                           6,361,592   192 ns/op    53 B/op   2 allocs/op
BenchmarkRouter_MethodNotAllowed-12                   7,119,540   209 ns/op    77 B/op   2 allocs/op
```

### Context Operations

```
BenchmarkContext_Param-12                           695,830,063   1.5 ns/op     0 B/op   0 allocs/op
BenchmarkContext_Query-12                           100,000,000  21.8 ns/op     0 B/op   0 allocs/op
```

---

## How Zero-Alloc Works

The routing hot path achieves 0 allocations through:

1. **Context pooling** (`sync.Pool`): Context structs are reused across requests
2. **Caller-provided param buffer**: `Lookup(path, c.radixBuf[:0])` — the param slice is pre-allocated in the pooled Context and passed to the radix tree. No `make()` on the hot path.
3. **Pre-allocated handler chain**: middleware slice is reused from pool
4. **Radix tree `Contains()`**: existence checks for 405 responses use a dedicated method that doesn't allocate params at all

The 1 alloc in `TrailingSlash_Strip` comes from the fallback `Lookup` in `tryTrailingSlashLookup` which uses a fresh buffer (non-hot path).

---

## v0.4.0 → v0.5.0 Comparison

| Route type | v0.4.0 | v0.5.0 | Speedup |
|---|---|---|---|
| Static | 256 ns, 1 alloc | **53 ns, 0 alloc** | **4.8x** |
| Parametric | 326 ns, 1 alloc | **64 ns, 0 alloc** | **5.1x** |
| Root path | 256 ns, 1 alloc | **46 ns, 0 alloc** | **5.6x** |
| Wildcard | 539 ns, 1 alloc | **58 ns, 0 alloc** | **9.3x** |
| Deep nesting | 561 ns, 1 alloc | **143 ns, 0 alloc** | **3.9x** |
