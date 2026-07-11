# Add MCP Install Command

## Request

Add a `ck add-mcp` subcommand that installs the coldkit MCP server into common
agent configuration directories, starting with Codex and Claude Code.
Also publish a patch release and expose `coldkit` as an npm command alias for
the existing `ck` CLI.

## Changes

- Added `internal/mcpinstall` for agent-specific config writes.
- Added `ck add-mcp AGENT` and `ck install-mcp AGENT` CLI entry points.
- Supported `codex`, `claude-code`, and the `cloud-code`/`claude` aliases.
- Wrote Codex config to TOML and Claude Code config to JSON.
- Added the `coldkit` npm bin alias for the same CLI launcher used by `ck`.
- Bumped npm package versions to `0.1.1` for the patch release.
- Updated pinned GitHub Actions to current Node 24-compatible releases after
  publish validation surfaced runtime deprecation annotations.
- Documented MCP installation in CLI and MCP docs.

## Intent

Users who install the npm CLI should be able to run one command to make
`ck-mcp` available to their preferred agent without manually editing config
files.

## Files

- `internal/mcpinstall/install.go`
- `internal/mcpinstall/install_test.go`
- `internal/cli/command.go`
- `package.json`
- `.github/workflows/ci.yml`
- `.github/workflows/npm-publish.yml`
- `README.md`
- `docs/CLI.md`
- `docs/MCP.md`
- `docs/ARCHITECTURE.md`
