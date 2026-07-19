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
| Product surface | B | First TRON CLI and MCP surface exists with offline generation, validation, watch-only balances/resources, TRC20 call-data previews with dry-run simulation, JSON output, npm packaging, MCP agent install helpers, and macOS Keychain-backed digest signing. | Add full unsigned TRON transaction parsing and signed transaction assembly on top of the signer boundary. |
| Architecture docs | B | Repository boundaries are now project-specific and chain extensibility is documented. | Add a focused MCP protocol note if the server grows. |
| Testing | B | Unit tests cover TRON vectors, vanity matching, watch-only balance/resource parsing, digest signing, MCP tool exposure, MCP config installation, and CI runs the local gates. | Add release build checksums, broader install smoke tests, and manual macOS Keychain prompt verification. |
| Observability | C | CLI/MCP are simple and mostly synchronous; there is no structured logging convention yet. | Add debug/progress output for long vanity searches without leaking secrets. |
| Security | B | Cold/watch-only boundary, MCP secret gating, and the signer-not-secret-return boundary are explicit. | Add full transaction preview/confirmation semantics before expanding signing beyond digest inputs. |
