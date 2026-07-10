# Supply Chain Security

This document records the current dependency and release posture for `coldkit`.

## Current State

- Go dependencies are declared in `go.mod` and locked in `go.sum`.
- Runtime dependencies are intentionally small: Cobra for CLI wiring and `golang.org/x/crypto/sha3` for legacy Keccak-256.
- Cold key-generation paths do not depend on npm, pip, external services, or generated code.
- There is no automated CI, SBOM, vulnerability scanning, or release provenance workflow yet.

## Rules

- Keep `go.mod` and `go.sum` committed.
- Prefer Go standard library or small, auditable dependencies.
- Do not add telemetry, webhook export, remote signing, or secret-upload dependencies.
- Pin GitHub Actions to immutable commit SHAs when CI is introduced.

## Tooling To Add Later

- `go test ./...` in GitHub Actions.
- OSV scanning for Go dependencies.
- Release checksums for `ck` and `ck-mcp`.
- SPDX SBOM artifacts for release builds.
- Signed build provenance after release automation exists.

## Release Assumptions

The first release should produce static-ish single-file binaries for macOS,
Linux, and Windows where possible. Every release should publish checksums and
clearly mark `ck-mcp --enable-secret-tools` as offline-only.
