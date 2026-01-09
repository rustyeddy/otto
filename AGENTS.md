# AGENTS.md — OttO

## Purpose
OttO is the core runtime/framework. Keep application-specific logic out of this repo.
Add tests and docs with minimal disruption to public APIs.

## Non-goals
- Do NOT add/require external services (no real MQTT broker, DB, HTTP server) for unit tests.
- Do NOT access hardware, GPIO, serial, or OS-specific devices in tests.
- Avoid broad refactors. Prefer small, additive changes that enable testing.

## Go conventions
- Run `gofmt` on all changed files.
- Keep package boundaries clean; avoid circular deps.
- Prefer context-aware APIs for long-running work.
- Avoid goroutine leaks: every goroutine must have a deterministic shutdown path.

## Testing (required)
Use **testify**:
- Use `github.com/stretchr/testify/require` for must-pass assertions.
- Use `github.com/stretchr/testify/assert` for non-fatal checks.
- Use `github.com/stretchr/testify/mock` only when a small fake isn’t practical.

Test rules:
- Tests must be hermetic: no network, no filesystem writes outside `t.TempDir()`.
- Avoid `time.Sleep` for synchronization. Use channels, WaitGroups, or context cancellation.
- Any test that could block must use `context.WithTimeout` and fail fast.
- Prefer table-driven tests.
- Keep tests deterministic and non-flaky.

Commands:
- Run: `go test ./...`
- When adding new packages or helpers, keep them internal: `internal/testutil` is allowed.

## What to test first (priority)
1. Pure logic: parsing, validation, topic naming/handling, config processing.
2. Concurrency: cancellation, shutdown behavior, channel fan-in/out, race-prone areas.
3. Interface contracts: error propagation and edge cases.

## Review checklist (before proposing changes)
- Does this change keep OttO independent of app/device-specific logic?
- Are new tests hermetic and deterministic?
- Any new goroutines? Where do they stop?
- Any sleeps/time-based flakiness introduced? Remove it.

## Commit guidance (if committing)
Prefer small commits:
- `test: <pkg> baseline`
- `testutil: add <helper>`
- `docs: godoc for <pkg>`
