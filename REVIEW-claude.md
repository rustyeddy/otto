# Code Review — OttO IoT Framework

**Reviewer:** Claude (claude-sonnet-4-6)  
**Date:** 2026-05-27  
**Module:** `github.com/rustyeddy/otto` v0.0.12  
**Go version:** 1.24.5

---

## Summary

OttO is a Go framework for building IoT applications. It provides a clean MQTT-backed
messaging layer, a device wiring model (source/sink/duplex), a small rules engine, and
a runtime-configurable logging service. The active codebase (~2,200 lines) is lean and
well-structured. All tests pass and `go vet` is clean.

---

## Architecture Overview

```
cmd/otto/         - HTTP server entry point (cobra CLI, signals, /api/log)
messenger/        - Core messaging layer
  mqtt_client.go  - MQTT interface (Publish / Subscribe / SetWill)
  messenger.go    - Subscription manager (WantSub / ResubscribeAll)
  registry.go     - Device registry: wires devices to MQTT, state cache
  topics.go       - Canonical topic-path builder (TopicScheme)
  payloads.go     - StatusPayload / MetaPayload JSON structs
  wire_typed.go   - Generic WireSource / WireSink / WireDuplex helpers
  codec/          - Codec[T] interface + JSON[T] implementation
  mqtt/           - Paho MQTT adapter (implements messenger.MQTT)
rules/            - Concurrent rule runner + Follow + ToggleOnRisingEdge
logging/          - slog-based logging service with HTTP API
internal/testutils/ - Generic channel helpers (WaitRecv, CollectN, etc.)
version.go        - Module version constant
```

The `_archive/`, `_station/`, `_server/`, `_client/`, `_cmd/`, `_data/`, and `_utils/`
directories hold work-in-progress code that Go's toolchain ignores (leading underscore).

---

## Strengths

### 1. Clean interface boundaries

`messenger.MQTT` is a minimal three-method interface. The Paho adapter in `messenger/mqtt`
is the only place that imports the Paho library. Tests use a hand-rolled `fakeClient`,
keeping unit tests fast and dependency-free.

### 2. Reconnect-safe subscription model

`WantSub` records the *desired* subscription state; `ResubscribeAll` replays it on every
connect/reconnect. This is the correct pattern for MQTT clients — subscriptions survive
broker restarts without application-level intervention.

### 3. Generic device wiring

`WireSource[T]`, `WireSink[T]`, and `WireDuplex[T]` are clean generics that decouple the
publish/subscribe machinery from any particular value type. The `codec.Codec[T]` interface
keeps serialization swappable.

### 4. Good concurrency discipline

`Registry` separates subscription-map locking (`mu RWMutex`) from state-cache locking
(`stateMu RWMutex`), avoiding lock contention between the event-publishing hot path and
subscription management. `sync.Once` guards `Source.CloseOut()` correctly.

### 5. Solid test infrastructure

`internal/testutils` provides production-quality channel helpers (`WaitRecv`, `CollectN`,
`Eventually`, `Drain`) that make async tests deterministic. `Source[T]` / `Sink[T]` fakes
implement the `devices` interfaces cleanly.

### 6. Logging service is first-class

The `logging` package exposes a fully runtime-configurable `slog` setup (level, format,
output, file) via an HTTP handler, allowing log verbosity changes without restarts.

---

## Issues and Recommendations

### Critical

None. All tests pass and there are no data races visible from static analysis.

---

### High

#### H1 — `rand.Intn` in `mqtt/paho.go` uses a deprecated global source

**File:** `messenger/mqtt/paho.go:54`  
`rand.Intn` without a local source uses the global pseudo-random source seeded to 1 in
Go < 1.20; in Go 1.20+ it is automatically seeded, but using `math/rand/v2` is now
idiomatic and removes the deprecation warning.

```go
// Before
import "math/rand"
b[i] = letters[rand.Intn(len(letters))]

// After
import "math/rand/v2"
b[i] = letters[rand.IntN(len(letters))]
```

#### H2 — `SetWill` must be called before `Connect`

**File:** `messenger/mqtt/paho.go` (entire file)  
`Paho.SetWill` mutates `p.opts`, which is consumed when `p.c = paho.NewClient(p.opts)` is
called inside `Connect`. If a caller sets the will *after* `Connect`, it is silently
ignored. The `MQTT` interface does not document this ordering constraint; a comment or a
guard (return an error if `p.c != nil`) would prevent silent misconfiguration.

---

### Medium

#### M1 — `Messenger` and `Registry` duplicate the same subscription logic

Both types carry `mu sync.RWMutex`, `subscriptions/subs map[string]subSpec`,
`unsubs map[string]func() error`, `WantSub`, and `ResubscribeAll`. `Messenger` appears to
be an earlier, simpler version that `Registry` supersedes. If `Messenger` is still needed,
consider embedding or delegating to a shared `SubscriptionManager`; if not, remove it.

#### M2 — `Registry.Run` returns on the *first* device error, leaving others running

**File:** `messenger/registry.go:Run`  
When `errCh` fires, `Run` returns immediately. The device goroutines launched earlier
continue running until `ctx` is cancelled by the caller. If the caller ignores the error
and does not cancel the context, goroutines leak. Document this clearly, or cancel an
internal derived context on first error:

```go
ctx, cancel := context.WithCancel(ctx)
defer cancel()
// ... launch goroutines with the derived ctx
```

#### M3 — `WireSink` drops messages silently on timeout

**File:** `messenger/wire_typed.go:WireSink`  
When the device's input channel is full and the `CommandTimeout` elapses, the incoming
MQTT command is logged at `Warn` and dropped. For actuators (relays, pumps) this could
be a safety concern. Consider adding a metric counter or surfacing this as an event on
the device's event channel.

#### M4 — `ToggleOnRisingEdge` reads stale state from the Registry cache

**File:** `rules/toggle_on_rising.go:Run`  
The toggle reads `messenger.StateAs[bool](t.Registry, t.Relay.Name())` which is only
populated by `WireSource`. If `WireSource` has not yet published a state (e.g., at boot),
the default `false` is used, which may not match the physical relay state. A pull/read from
the device directly, or an explicit "unknown" state, would be safer.

#### M5 — `Makefile` `build` target is undefined

`make build` is listed in `.PHONY` and referenced in `make ci` and `make all`, but there
is no `build:` recipe in `Makefile`. Running `make build` or `make ci` will fail. Add:

```makefile
build:
	go build -o $(OTTO_BINARY) ./cmd/otto
```

#### M6 — CI pins Go 1.23 but `go.mod` requires Go 1.24.5

**File:** `.github/workflows/ci.yml`  
The workflow uses `go-version: '1.23'`; `go.mod` declares `go 1.24.5`. The build will
succeed (Go's toolchain is forward-compatible for downloads), but the mismatch can cause
confusing diagnostics. Pin CI to `'1.24'` or `'>=1.24'` to match.

---

### Low / Style

#### L1 — `_otto.go` / `_otto_test.go` are unreachable dead code

The files have a leading underscore and are excluded by the Go toolchain. If they are
being kept for reference, move them to `_archive/`; if they represent planned work,
rename them.

#### L2 — `examples/ledctl/ledctl` is a compiled binary committed to the repo

**File:** `examples/ledctl/ledctl`  
Compiled binaries should not be committed. Add `examples/ledctl/ledctl` (or `**/ledctl`)
to `.gitignore`.

#### L3 — `cover.out` and `coverage.html` are committed

Both files are build artifacts. They are not in `.gitignore` and appear in the repo. Add
them:

```gitignore
cover.out
coverage.html
```

#### L4 — `_mesh/otto.log` is a runtime log committed to the repo

Same issue as L3 — add `**/*.log` or the specific path to `.gitignore`.

#### L5 — Version is hardcoded in two places

`version.go` declares `Version = "0.0.12"` and `Makefile` declares `VERSION?= 0.0.12`.
The Makefile comment notes this should be read from `version.go`. A small shell snippet
(`$(shell grep 'Version = ' version.go | ...)`) or a `go generate` step would keep them
in sync.

#### L6 — `cmd/otto/main.go` `serveCmd` only exposes `/api/log`

The README advertises `/api/config`, `/api/data`, and `/api/stations`, but the running
server only mounts `/api/log`. This is fine for the current development state but the
README should reflect what is actually wired up to avoid confusion.

---

## Test Coverage Summary

| Package                    | Coverage |
|---------------------------|----------|
| `github.com/rustyeddy/otto` (version) | 100%     |
| `messenger/codec`          | 100%     |
| `logging`                  | 81%      |
| `messenger`                | 82%      |
| `messenger/mqtt`           | 80%      |
| `rules`                    | 48%      |
| `cmd/otto`                 | 0% (no tests) |
| `examples/ledctl`          | 0% (no tests) |
| `internal/testutils`       | 0% (test helpers; acceptable) |

`rules` at 48% is the most notable gap. `ToggleOnRisingEdge.Run` has no test coverage at
all, which is risky given the debounce and state-cache logic described in M4 above.

---

## Prioritized Action List

1. **Fix the `build` Makefile target** (M5) — CI and contributors are blocked without it.
2. **Align CI Go version with `go.mod`** (M6) — straightforward one-line fix.
3. **Add `.gitignore` entries** for `cover.out`, `coverage.html`, `*.log`, and compiled
   binaries (L2–L4).
4. **Clarify or remove `Messenger`** in favour of `Registry` (M1) — reduces surface area
   and confusion.
5. **Document `SetWill`-before-`Connect` ordering** or guard it (H2).
6. **Add tests for `ToggleOnRisingEdge`** to bring `rules` coverage above 80% (coverage gap).
7. **Migrate `math/rand` to `math/rand/v2`** (H1) — simple mechanical change.
8. **Address stale relay state at boot** (M4) if OttO is used with physical actuators.
