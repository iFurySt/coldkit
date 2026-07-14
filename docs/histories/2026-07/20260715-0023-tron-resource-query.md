## [2026-07-15 00:23] | Task: Add TRON resource queries

### Execution Context

- Agent ID: `codex`
- Base Model: `GPT-5`
- Runtime: `Codex CLI`

### User Query

> Support querying TRON Energy and Bandwidth, and consider returning them with the balance tool.

### Changes Overview

- Area: TRON watch-only CLI and MCP tools.
- Key actions:
  - Added TRON account resource parsing from `wallet/getaccountresource`.
  - Extended balance results to include Energy and Bandwidth resources.
  - Added `ck tron resource` and MCP `tron_resource`.
  - Updated tests, release notes, package version, and repository docs for the
    new watch-only surface.

### Design Intent

The balance tool now returns resources in the same user-facing call because
operators usually need TRX, USDT, Energy, and Bandwidth together before sending
TRC20 transactions. A separate resource command/tool remains available for
callers that only need Energy and Bandwidth. Both paths remain watch-only and
accept public addresses only.

### Files Modified

- `internal/tron/watch.go`
- `internal/tron/account_test.go`
- `internal/cli/command.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `README.md`
- `docs/ARCHITECTURE.md`
- `docs/CLI.md`
- `docs/MCP.md`
- `docs/RELIABILITY.md`
- `docs/ROADMAP.md`
- `docs/QUALITY_SCORE.md`
- `docs/releases/feature-release-notes.md`
- `package.json`
