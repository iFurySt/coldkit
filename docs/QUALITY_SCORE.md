# Quality Score

Track quality by product area and architectural layer so agents can prioritize
the weakest parts of the system.

## Suggested Scale

- `A`: strong coverage, stable behavior, clear docs, low operational risk.
- `B`: acceptable but still has known gaps.
- `C`: works but needs targeted hardening.
- `D`: fragile or underspecified.

## Current Score

| Area | Score | Why | Next Step |
| --- | --- | --- | --- |
| Product surface | B | First TRON CLI and MCP surface exists with offline generation, validation, watch-only balances, JSON output, and secret-tool gating. | Add packaged releases and documented MCP client examples. |
| Architecture docs | B | Repository boundaries are now project-specific and chain extensibility is documented. | Add a focused MCP protocol note if the server grows. |
| Testing | B | Unit tests cover TRON vectors, vanity matching, watch-only balance parsing, and MCP tool exposure. CLI smoke commands have been run locally. | Add CI running `go test ./...` and release builds. |
| Observability | C | CLI/MCP are simple and mostly synchronous; there is no structured logging convention yet. | Add debug/progress output for long vanity searches without leaking secrets. |
| Security | B | Cold/watch-only boundary and MCP secret gating are explicit. | Add reproducible release checksums and supply-chain scanning. |
