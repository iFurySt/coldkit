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
| Product surface | B | First TRON CLI and MCP surface exists with offline generation, validation, watch-only balances, JSON output, npm packaging, and MCP agent install helpers. | Expand `ck add-mcp` to more agent clients after validating their config formats. |
| Architecture docs | B | Repository boundaries are now project-specific and chain extensibility is documented. | Add a focused MCP protocol note if the server grows. |
| Testing | B | Unit tests cover TRON vectors, vanity matching, watch-only balance parsing, MCP tool exposure, MCP config installation, and CI runs the local gates. | Add release build checksums and broader install smoke tests. |
| Observability | C | CLI/MCP are simple and mostly synchronous; there is no structured logging convention yet. | Add debug/progress output for long vanity searches without leaking secrets. |
| Security | B | Cold/watch-only boundary and MCP secret gating are explicit. | Add reproducible release checksums and supply-chain scanning. |
