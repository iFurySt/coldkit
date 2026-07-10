# Bootstrap coldkit

## User Request

Create a standalone `coldkit` repository for a short CLI and MCP server. The CLI
should be chain-extensible, use `ck`, and initially support TRON cold-wallet and
USDT/TRC20 watch-only workflows.

## Changes

- Added Go module `github.com/ifuryst/coldkit`.
- Added `cmd/ck` CLI with `ck tron gen`, `ck tron val`, `ck tron bal`, `ck tron from-private`, and self-test commands.
- Added `cmd/ck-mcp` stdio MCP server with safe default tools and opt-in secret tool exposure.
- Added `internal/tron` for TRON key generation, Base58Check validation, vanity prefix/suffix matching, self-tests, and watch-only balance parsing.
- Added `internal/mcp` tests for tool exposure and validation tool behavior.
- Replaced template README, architecture, security, and quality docs with project-specific content.
- Added focused CLI, MCP, roadmap, threat model, product, reliability, and CI/CD docs so the intended open-source direction is versioned in the repository.

## Design Intent

- Keep cold key-generation paths separate from networked watch-only paths.
- Keep CLI and MCP short and agent-friendly.
- Use a chain-first command shape so future chains can be added without renaming the project.
- Disable private-key-returning MCP tools by default.

## Important Files

- `cmd/ck/main.go`
- `cmd/ck-mcp/main.go`
- `internal/cli/command.go`
- `internal/mcp/server.go`
- `internal/tron/account.go`
- `README.md`
- `docs/ARCHITECTURE.md`
- `docs/SECURITY.md`
