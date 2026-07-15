# coldkit

`coldkit` is an offline-first wallet safety toolkit for humans and AI agents.

The first supported chain is TRON. The CLI binary is available as `ck` and
`coldkit`; the MCP server binary is `ck-mcp`.

The project is intentionally not a full wallet. It focuses on address
generation, validation, watch-only balance checks, and agent-safe MCP tools with
clear cold/hot boundaries.

## Security Model

- Cold commands such as `ck tron gen`, `ck tron val`, and `ck tron self` do not perform network I/O.
- Watch-only commands such as `ck tron bal` accept public addresses only.
- `ck-mcp` hides private-key-returning tools by default.
- Use private-key generation only on an offline machine.

## Install From Source

```sh
go build -o bin/ck ./cmd/ck
go build -o bin/ck-mcp ./cmd/ck-mcp
```

## Install With npm

```sh
npm install -g coldkit
```

The npm package installs prebuilt `ck`, `coldkit`, and `ck-mcp` commands for
macOS, Linux, and Windows on x64 and arm64 through platform-specific optional
packages.

For offline use after installation, make sure npm installs the matching
platform package:

```sh
npm install -g --include=optional coldkit
```

If npm skips optional packages during a global install, the launcher falls back
to fetching the matching platform package with `npm exec` on first run.

## CLI

Generate a normal TRON account offline:

```sh
ck tron gen -j
```

Generate vanity addresses with multiple suffixes:

```sh
ck tron gen -s 888 -s xyz -n 3 -j
```

Generate a public preview without printing private keys:

```sh
ck tron gen -s 888 -n 1 --pub -j
```

Validate a public TRON address offline:

```sh
ck tron val TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3 -j
```

Check TRX, USDT/TRC20, Energy, and Bandwidth for a public address:

```sh
ck tron bal TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3 -j
```

Check only Energy and Bandwidth resources:

```sh
ck tron resource TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3 -j
```

Watch-only queries use a small full node endpoint pool and automatically fall
back when an endpoint is rate limited or unavailable. To use your own node or
provider, repeat `--endpoint`:

```sh
ck tron bal T... --endpoint https://api.trongrid.io --endpoint http://127.0.0.1:8090
```

Run deterministic crypto test vectors:

```sh
ck self
```

Detailed CLI conventions live in [docs/CLI.md](docs/CLI.md).

## MCP

Safe watch-only/default mode:

```sh
ck-mcp
```

Offline secret mode:

```sh
ck-mcp --enable-secret-tools
```

Default tools:

- `tron_validate`
- `tron_balance`
- `tron_resource`
- `tron_generate_preview`

Secret tool, only exposed with `--enable-secret-tools`:

- `tron_generate_secret`

Detailed MCP conventions live in [docs/MCP.md](docs/MCP.md).

## Project Docs

- [Architecture](docs/ARCHITECTURE.md)
- [Threat Model](docs/THREAT_MODEL.md)
- [Security](docs/SECURITY.md)
- [Product Sense](docs/PRODUCT_SENSE.md)
- [Roadmap](docs/ROADMAP.md)
- [Reliability](docs/RELIABILITY.md)
- [Supply Chain Security](docs/SUPPLY_CHAIN_SECURITY.md)

## Development

```sh
go test ./...
go run ./cmd/ck self
go run ./cmd/ck tron gen -s 2 -s 3 -n 2 --max 10000 --pub -j
```

## License

[MIT](LICENSE)
