# npm Publishing Setup

## Request

Create and publish the public `iFurySt/coldkit` repository, keep the project
permissively licensed with no warranty, and set up npm distribution without
recording secrets in the repository.

## Changes

- Kept the MIT license because it permits broad reuse while disclaiming warranty
  and liability.
- Added npm package metadata for `coldkit` with `ck` and `ck-mcp` bin entries
  and platform-specific optional binary packages.
- Added Node.js launchers and a build script that cross-compiles Go binaries for
  macOS, Linux, and Windows on x64 and arm64.
- Added GitHub Actions CI and npm publishing workflows with pinned actions.
- Documented npm install, CI, npm Trusted Publishing, and supply-chain posture.

## Intent

The npm package should be installable with `npm install -g coldkit` and should
not require users to have Go installed. Splitting binaries into optional
platform packages keeps the root package small while preserving the single
install command. Future npm releases should publish from GitHub Actions through
OIDC Trusted Publishing rather than long-lived npm tokens.

## Files

- `package.json`
- `npm/bin/`
- `npm/lib/run-binary.js`
- `scripts/build-npm-platform-packages.js`
- `.github/workflows/ci.yml`
- `.github/workflows/npm-publish.yml`
- `README.md`
- `docs/CICD.md`
- `docs/SUPPLY_CHAIN_SECURITY.md`
