## [2026-07-19 23:15] | Task: TRC20 transfer dry-run safeguards

### Execution Context

- Agent ID: `codex`
- Base Model: `GPT-5`
- Runtime: `local CLI`

### User Query

> Investigate a TRC20 transfer parameter encoding error that caused a `REVERT`,
> then update coldkit to reduce future call-data mistakes and catch failures
> before broadcast.

### Changes Overview

- Area: TRON CLI and watch-only safety tooling.
- Key actions:
  - Added `ck tron trc20-transfer` to build USDT/TRC20 transfer call data
    without signing or broadcasting.
  - Added strict decimal amount parsing, ABI parameter length checks, and
    address/amount round-trip validation.
  - Added optional `--owner` dry-run simulation through TRON full node constant
    contract execution, surfacing `REVERT` and `FAILED` responses before
    callers hand data to an external signer or broadcaster.
  - Updated docs, release notes, quality notes, and npm package version.

### Design Intent

The failure mode came from hand-written ABI hex outside the tool boundary. The
fix moves TRC20 transfer parameter generation into a deterministic local command
and adds a network dry-run as an explicit watch-only step. The command still
does not sign or broadcast, preserving coldkit's existing signer boundary while
removing a high-risk manual encoding step from agent workflows.

### Files Modified

- `internal/tron/trc20.go`
- `internal/tron/trc20_test.go`
- `internal/cli/command.go`
- `README.md`
- `docs/CLI.md`
- `docs/ARCHITECTURE.md`
- `docs/SECURITY.md`
- `docs/QUALITY_SCORE.md`
- `docs/releases/feature-release-notes.md`
- `package.json`
