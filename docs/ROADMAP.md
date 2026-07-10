# Roadmap

This roadmap captures intended direction. It is not a commitment to add unsafe
wallet features.

## MVP: TRON Cold + Watch-Only

Implemented:

- `ck` CLI.
- `ck-mcp` MCP server.
- TRON private-key and address generation.
- TRON vanity prefix/suffix generation.
- TRON address validation.
- TRX and USDT/TRC20 watch-only balances.
- Default MCP secret-tool gating.

## Next

- Release binaries for macOS, Linux, and Windows.
- Add GitHub Actions for `go test ./...`.
- Add release checksums.
- Add `ck version`.
- Add `--progress` for long vanity searches without leaking secret material.
- Add README examples for configuring common MCP clients.
- Add benchmarks for vanity generation.

## Later

- Add `--output` with safe defaults:
  - refuse to overwrite files by default;
  - clearly mark files containing private keys;
  - support `--pub` public-only output.
- Add deterministic build notes and SBOM artifacts.
- Add optional stronger CPU implementations for secp256k1 if profiling justifies it.
- Add additional chains with the same cold/watch-only split.

## Candidate Chains

- ETH/EVM: address generation, validation, vanity, native/token watch-only balances.
- SOL: address generation, validation, native/SPL watch-only balances.
- TON: address validation and watch-only balance first, generation after format rules are documented.
- BTC: watch-only validation first; generation requires extra care around address types and HD derivation.

## Explicit Non-Goals

- No transaction sending in the MVP.
- No remote custody.
- No hosted service.
- No browser extension.
- No secret upload/export over network.
- No dependency-heavy web UI for cold key generation.
