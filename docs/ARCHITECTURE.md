# Architecture

`coldkit` is an offline-first wallet safety toolkit. It exposes a short CLI for
humans and scripts, plus an MCP server for AI agents.

## Repository Shape

- `cmd/ck/`: main CLI entry point. The public command shape is chain-first, for example `ck tron gen`.
- `cmd/ck-mcp/`: stdio MCP server entry point. It exposes safe watch-only tools by default and secret-returning tools only when explicitly enabled.
- `internal/cli/`: Cobra command wiring and human/JSON output formatting.
- `internal/keychain/`: local key storage adapters, currently macOS Keychain for TRON signing keys.
- `internal/mcp/`: minimal JSON-RPC stdio MCP server implementation and tool dispatch.
- `internal/mcpinstall/`: agent-specific MCP config installers for `ck-mcp`.
- `internal/tron/`: TRON address generation, Base58Check validation, vanity matching, deterministic self-tests, TRC20 transfer call-data previews, and watch-only balance/resource lookup.
- `docs/`: project rules, architecture, histories, and security notes.

## Boundaries

- Chain-specific cryptography and API clients live under `internal/<chain>/`.
- CLI and MCP layers call chain packages; chain packages do not import CLI or MCP code.
- Cold paths must not perform network I/O. Today that includes `ck tron gen`, `ck tron val`, `ck tron from-private`, and `ck tron self`.
- Watch-only paths can perform network I/O, but must accept public addresses only.
- TRC20 transfer preview paths may build ABI call data, but they do not sign or
  broadcast transactions. Network dry-runs are watch-only simulations using
  public owner and recipient addresses.
- Secret-returning MCP tools remain disabled unless `ck-mcp --enable-secret-tools` is used.
- Signer paths may use local OS secret storage and authorization, but must not
  return private keys to CLI or MCP callers.

## Current Product Surface

- `ck tron gen`: generate TRON accounts offline; supports repeated `-p/--prefix`, repeated `-s/--suffix`, `-n/--count`, `--max`, `--pub`, and `-j/--json`.
- `ck tron val`: validate public TRON addresses offline.
- `ck tron bal`: query public TRX, USDT/TRC20, Energy, and Bandwidth; supports `--network mainnet|nile|shasta`.
- `ck tron resource`: query public TRON Energy and Bandwidth; supports `--network mainnet|nile|shasta`.
- `ck tron trc20-transfer`: build TRC20 `transfer(address,uint256)` call data
  offline, with optional `--owner` dry-run simulation before signing or
  broadcasting elsewhere.
- `ck tron from-private`: derive an address from a private key for verification.
- `ck keychain import-tron`: store a TRON private key in macOS Keychain.
- `ck tron sign-hash`: sign a 32-byte digest using a macOS Keychain-backed TRON key.
- `ck add-mcp`: install `ck-mcp` into supported agent MCP configs.
- `ck-mcp`: expose `tron_validate`, `tron_balance`, `tron_resource`, `tron_generate_preview`,
  and Keychain-backed `tron_sign_hash`; optionally expose
  `tron_generate_secret`.

More CLI details live in `docs/CLI.md`; MCP details live in `docs/MCP.md`.

## Future Chains

New chains should follow the existing `internal/tron` boundary and add commands
under `ck <chain> ...`. MCP tools should use the same prefix style, such as
`eth_validate` or `sol_balance`.
