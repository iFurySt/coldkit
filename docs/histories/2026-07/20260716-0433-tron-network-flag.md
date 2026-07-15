## [2026-07-16 04:33] | Task: Add TRON network selection

### Execution Context

- Agent ID: `codex`
- Base Model: `GPT-5`
- Runtime: `Codex CLI`

### User Query

> Add `--network` so local development and testing can use TRON test chains.

### Changes Overview

- Area: TRON watch-only CLI and MCP tools.
- Key actions:
  - Added built-in endpoint pools for `mainnet`, `nile`, and `shasta`.
  - Added `--network` to `ck tron bal` and `ck tron resource`.
  - Added optional `network` arguments for MCP `tron_balance` and
    `tron_resource`.
  - Kept repeated `--endpoint` as an override for custom node pools.
  - Updated README, CLI, MCP, and architecture docs.

### Design Intent

Local development needs a short, repeatable way to target public TRON testnets
without copying endpoint URLs into every command. Network selection chooses a
known endpoint pool, while explicit endpoints remain available for local private
chains or provider-specific infrastructure.

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
