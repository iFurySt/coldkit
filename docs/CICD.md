# CI/CD Guide

`coldkit` ships minimal GitHub Actions workflows for local gates and npm
publishing.

## Current Local Gates

```sh
go test ./...
make build
npm pack --dry-run
npm run build:npm-platform-packages
```

`make build` produces:

- `bin/ck`
- `bin/ck-mcp`

`bin/` is ignored and should not be committed.

## CI Workflow

`.github/workflows/ci.yml` runs on pushes to `main` and pull requests. It:

- sets up Go from `go.mod`;
- sets up Node.js 22;
- runs `go test ./...`;
- runs `make build`;
- builds platform-specific npm binary packages;
- runs `npm pack --dry-run` for each generated platform package and the root
  package.

Actions are pinned to immutable commit SHAs.

## npm Publishing Workflow

`.github/workflows/npm-publish.yml` runs manually or when a GitHub Release is
published. It uses npm Trusted Publishing through GitHub Actions OIDC, so the
repository should not store an npm publish token.

Configure the npm package trusted publisher with:

- Organization or user: `iFurySt`
- Repository: `coldkit`
- Workflow filename: `npm-publish.yml`
- Allowed action: `npm publish`

The workflow grants `id-token: write`, uses Node.js 22, updates npm to the
latest CLI, runs tests, publishes platform-specific optional binary packages,
and then publishes the root `coldkit` package with provenance enabled by
`package.json`.

## First Release Artifacts

The first release workflow should:

- build `ck` and `ck-mcp` for macOS, Linux, and Windows;
- publish checksums;
- include a short security note about offline secret generation;
- avoid any automatic secret generation in CI;
- update `docs/releases/` with user-facing notes.

## Later Supply-Chain Work

- OSV scanning.
- SPDX SBOM.
- Signed provenance for non-npm release artifacts.
- Reproducible build documentation.

If release automation is added, update `docs/SUPPLY_CHAIN_SECURITY.md` and
`docs/releases/README.md` in the same change.
