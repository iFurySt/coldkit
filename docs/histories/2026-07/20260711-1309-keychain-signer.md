# macOS Keychain Signer

## Request

Add a safer way for AI/MCP workflows to request signatures without receiving
private keys, starting with macOS Keychain authorization.

## Changes

- Added TRON secp256k1 digest signing with raw, recovery-id, and DER signature
  outputs.
- Added a macOS Keychain storage adapter for TRON private keys.
- Replaced the `security` CLI adapter with a native Security.framework and
  LocalAuthentication backend so signing authorization can use Touch ID,
  Apple Watch, or password.
- Added optional Developer ID code signing support for Darwin npm package
  binaries via `COLDKIT_CODESIGN_IDENTITY`.
- Moved npm publishing to a macOS runner so Darwin packages include the native
  Keychain backend.
- Added the imported CI signing keychain to the user search list before
  invoking `codesign`, avoiding direct keychain lookup failures on GitHub
  Actions.
- Registered Apple Developer Bundle ID `com.ifuryst.coldkit` under team
  `J9P29FA5BX`.
- Configured GitHub Actions secrets for Developer ID signing and notary
  credentials.
- Bumped npm packages to `0.1.4` for the native signer and signed Darwin
  binary release.
- Added `ck keychain import-tron`, `ck keychain show-tron`, and
  `ck keychain delete`.
- Added a post-write Keychain read-back check so imports fail immediately if
  macOS reports success but the stored item cannot be resolved from the active
  Keychain search list.
- Added `ck tron sign-hash --key NAME` for local Keychain-backed signing.
- Added MCP tool `tron_sign_hash`, which returns a signature result and never
  returns private keys.
- Documented the signer boundary in CLI, MCP, architecture, and security docs.

## Intent

AI agents should be able to request a signature while the local OS asks the
human to authorize key access. The private key may enter the local `coldkit`
process for signing, but it must not be printed, returned through MCP, or
written to logs.

This first step signs 32-byte digests. Full TRON unsigned transaction parsing
and signed transaction assembly should build on the same signer boundary later.

## Files

- `internal/tron/sign.go`
- `internal/keychain/`
- `internal/cli/command.go`
- `internal/mcp/server.go`
- `scripts/build-npm-platform-packages.js`
- `.github/workflows/npm-publish.yml`
- `.apple/README.md`
- `.apple/app-store-connect-summary.json`
- `.gitignore`
- `package.json`
- `docs/CLI.md`
- `docs/MCP.md`
- `docs/SECURITY.md`
- `docs/ARCHITECTURE.md`
