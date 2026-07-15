## [2026-07-15 03:49] | Task: Add TRON endpoint fallback

### Execution Context

- Agent ID: `codex`
- Base Model: `GPT-5`
- Runtime: `Codex CLI`

### User Query

> TronGrid account queries often return HTTP 429 in real use. Reset the API-key
> environment-variable direction and use multiple endpoints instead.

### Changes Overview

- Area: TRON watch-only network reliability.
- Key actions:
  - Replaced TronGrid indexed account lookups with TRON full node HTTP calls.
  - Added default full node endpoint fallback for transient 429, 5xx, network,
    and timeout failures.
  - Made `--endpoint` repeatable so users can supply their own fallback pool.
  - Updated CLI, reliability, and security docs.

### Design Intent

The default user experience should stay zero-configuration. Public endpoints can
still be rate limited or unavailable, so `coldkit` now treats endpoint diversity
as the first reliability layer while keeping custom provider or local-node
configuration available through explicit flags.

### Files Modified

- `internal/tron/watch.go`
- `internal/tron/account_test.go`
- `internal/cli/command.go`
- `internal/mcp/server.go`
- `README.md`
- `docs/CLI.md`
- `docs/RELIABILITY.md`
- `docs/SECURITY.md`
