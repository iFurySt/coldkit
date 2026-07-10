# Security

`coldkit` handles wallet key material. Secure defaults are part of the product,
not an optional hardening pass.

## Secret Handling

- Private keys must never be sent to watch-only commands or MCP tools.
- `ck-mcp` must not expose private-key-returning tools unless started with `--enable-secret-tools`.
- Users should run secret-generating commands only on offline machines.
- Do not add clipboard, webhook, POST, telemetry, or remote export paths for secret material.

## Network Boundaries

- Cold commands must not perform network I/O.
- Watch-only commands may use public chain APIs, but must accept public addresses only.
- Any new external API endpoint must be documented in `README.md` or a focused docs page.

## Current Sensitive Commands

- `ck tron gen` prints private keys unless `--pub` is used.
- `ck tron from-private` accepts a private key and should be used only offline.
- `ck-mcp --enable-secret-tools` exposes `tron_generate_secret`; default MCP mode does not expose it.

## Dependency Rules

- Keep dependency manifests committed.
- Prefer Go standard library or small, auditable dependencies.
- Do not add npm/pip runtime dependencies for cold key generation paths.

The detailed threat model lives in `docs/THREAT_MODEL.md`.
