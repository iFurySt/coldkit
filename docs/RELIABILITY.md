# Reliability

`coldkit` is a local CLI/MCP tool, so reliability is mostly about deterministic
commands, bounded network calls, clear errors, and repeatable validation.

## Local Validation

Required before release-oriented changes:

```sh
go test ./...
make build
go run ./cmd/ck self
go run ./cmd/ck tron gen -s 2 -s 3 -n 2 --max 10000 --pub -j
```

Watch-only validation can be run when network access is acceptable:

```sh
go run ./cmd/ck tron bal TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3 -j
go run ./cmd/ck tron resource TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3 -j
```

## Timeouts

- Watch-only balance and resource queries use a 20 second default timeout.
- Cold commands should not need timeouts, but vanity search should support `--max`
  and later `--progress`.

## Error Handling

- CLI errors should exit non-zero and print concise errors to stderr.
- JSON output should stay parseable on success.
- MCP errors should return JSON-RPC errors, not panic or emit mixed stdout text.
- Watch-only TRON queries should fall back across configured full node endpoints
  for transient HTTP 429, HTTP 5xx, network, and timeout failures.

## Long Vanity Searches

Long vanity searches can run for a long time. Operators should use `--max` for
unattended runs. Future progress output must avoid printing private keys unless
the final result is intentionally secret-bearing.

## Operational Risks

- Third-party public RPC/API outages and unauthenticated rate limits can break
  watch-only balance checks if every configured endpoint fails.
- Offline machines may lack a recent Go toolchain if building from source.
- Users may misunderstand public preview output as spendable; docs and tool
  descriptions should keep stating that private keys are required to use funds.

CI/CD status and future integration guidance live in `docs/CICD.md`.
