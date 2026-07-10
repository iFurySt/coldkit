# Product Sense

`coldkit` exists for users and agents who need wallet safety workflows without a
large wallet application.

## Primary Users

- Humans generating receiving addresses on offline machines.
- Operators checking public balances without touching private keys.
- AI agents that need deterministic, parseable wallet tooling.
- Developers who want a small, auditable foundation for cold/watch-only chain tools.

## Differentiation

- Short CLI: `ck`.
- Short MCP server: `ck-mcp`.
- Chain-first command shape that can grow beyond TRON.
- Cold and watch-only behavior is separated by design.
- MCP secret tools are disabled by default.
- JSON output is a first-class interface, not an afterthought.

## Product Priorities

1. Safety and clear trust boundaries.
2. Agent-friendly non-interactive behavior.
3. Small auditable implementation.
4. Cross-platform single-binary distribution.
5. Performance only after correctness and safety.

## Tradeoffs

- Prefer explicit commands over a broad interactive wallet shell.
- Prefer local binaries over hosted services.
- Prefer no UI for cold generation until CLI/MCP behavior is mature.
- Prefer public preview modes over exposing private keys to agents.
- Avoid transaction sending until address generation, validation, and watch-only behavior are boringly reliable.
