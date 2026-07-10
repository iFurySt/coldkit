# CI/CD Guide

`coldkit` does not yet ship GitHub Actions workflows. The local project commands
are real and should become the first CI gate.

## Current Local Gates

```sh
go test ./...
make build
```

`make build` produces:

- `bin/ck`
- `bin/ck-mcp`

`bin/` is ignored and should not be committed.

## First CI Workflow

The first pull-request workflow should:

- set up Go;
- run `go test ./...`;
- run `make build`;
- avoid uploading binaries from pull requests.

Pin actions to immutable commit SHAs when the workflow is added.

## First Release Workflow

The first release workflow should:

- build `ck` and `ck-mcp` for macOS, Linux, and Windows;
- publish checksums;
- include a short security note about offline secret generation;
- avoid any automatic secret generation in CI;
- update `docs/releases/` with user-facing notes.

## Later Supply-Chain Work

- OSV scanning.
- SPDX SBOM.
- Signed provenance.
- Reproducible build documentation.

If release automation is added, update `docs/SUPPLY_CHAIN_SECURITY.md` and
`docs/releases/README.md` in the same change.
