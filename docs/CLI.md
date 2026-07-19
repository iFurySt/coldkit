# CLI

The CLI binary is available as `ck` and `coldkit`. Both names run the same
command. Commands are chain-first so the project can grow beyond TRON without
renaming the tool.

## Shape

```sh
ck <chain> <command> [flags]
coldkit <chain> <command> [flags]
```

Current TRON commands:

```sh
ck tron gen
ck tron val T...
ck tron bal T...
ck tron resource T...
ck tron trc20-transfer T... 30 --owner T...
ck tron from-private <PRIVATE_KEY_HEX>
ck tron sign-hash <DIGEST_HEX> --key main
ck tron self
```

Top-level command:

```sh
ck self
ck add-mcp codex
ck add-mcp claude-code
```

## Short Flags

- `-j, --json`: machine-readable output.
- `-n, --count`: number of matching generated addresses.
- `-p, --prefix`: required Base58 address prefix; repeatable.
- `-s, --suffix`: required Base58 address suffix; repeatable.
- `--max`: maximum total attempts for vanity search; `0` means unlimited.
- `--pub`: omit private keys from generated output.

## Vanity Matching

Prefix and suffix flags are OR lists within each side and ANDed across sides.

This matches addresses ending in either `888` or `xyz`:

```sh
ck tron gen -s 888 -s xyz -n 3 -j
```

This matches addresses starting with `T` and ending in either `2` or `3`:

```sh
ck tron gen -p T -s 2 -s 3 -n 2 --max 10000 -j
```

Long patterns are exponentially expensive. A three-character suffix averages
about `58^3` attempts. Use `--max` when running unattended.

## Output Rules

- Human output is for local terminal use.
- JSON output is stable enough for scripts and agents.
- Commands that print private keys must be run only in trusted offline
  environments.
- Use `--pub` for previews, demos, tests, logs, and agent-visible output.

## Network Rules

Cold commands must not perform network I/O:

- `ck tron gen`
- `ck tron val`
- `ck tron from-private`
- `ck tron self`
- `ck self`

Watch-only commands may perform network I/O:

- `ck tron bal`
- `ck tron resource`
- `ck tron trc20-transfer ... --owner T...`

Watch-only commands must never accept private keys.

Watch-only queries use a small TRON full node endpoint pool and automatically
fall back when an endpoint is rate limited, unavailable, or times out.
`--network` selects the built-in pool:

- `mainnet`: production TRON network, the default.
- `nile`: Nile testnet.
- `shasta`: Shasta testnet.

Testnet balances and history do not carry over from mainnet. Repeat
`--endpoint` to override the selected network pool with your own node or
provider:

```sh
ck tron bal T... --network nile -j
ck tron resource T... --network shasta -j
ck tron bal T... --endpoint https://api.trongrid.io --endpoint http://127.0.0.1:8090 -j
```

## TRC20 Transfer Preview

`ck tron trc20-transfer` builds TRC20 `transfer(address,uint256)` call data
without signing or broadcasting a transaction. It defaults to USDT/TRC20 and
6 decimals:

```sh
ck tron trc20-transfer TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J 30 -j
```

The command validates the destination address, converts the decimal amount to
raw token units, ABI-encodes the transfer parameter, and round-trips the encoded
address and amount before printing output. This avoids hand-written hex such as
retaining the TRON `41` address prefix in the wrong ABI word.

Pass `--owner` to dry-run the contract call through `/wallet/triggerconstantcontract`
before any external signer or broadcaster uses the call data:

```sh
ck tron trc20-transfer TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J 30 --owner T... -j
```

Dry-run mode uses the same `--network`, repeated `--endpoint`, and `--timeout`
flags as other watch-only network commands. A dry-run failure returns a non-zero
exit status and includes the node message, such as a `REVERT` response.

For non-USDT tokens, provide the token metadata explicitly:

```sh
ck tron trc20-transfer T... 12.5 --token TOKEN --contract T... --decimals 6 -j
```

## macOS Keychain Signing

`ck keychain import-tron NAME` stores a TRON private key in macOS Keychain and
prints only public metadata. By default it prompts for the private key without
echoing it:

```sh
ck keychain import-tron main
```

For scripted local import, pass the key through stdin explicitly:

```sh
printf '%s\n' "$TRON_PRIVATE_KEY_HEX" | ck keychain import-tron main --private-key-stdin
```

Sign a 32-byte digest with the stored key:

```sh
ck tron sign-hash 1111111111111111111111111111111111111111111111111111111111111111 --key main -j
```

The signing command asks macOS Keychain for the secret. Depending on local
Keychain policy, macOS may require Touch ID, Apple Watch, or account password.
Only the signature result is printed.

## MCP Installation

`ck add-mcp <agent>` installs the local `ck-mcp` stdio server into an agent's
MCP configuration. It defaults to user-level config and writes the nearest
sibling `ck-mcp` binary path when one exists, falling back to `ck-mcp` from
`PATH`.

Supported agents:

- `codex`: writes `~/.codex/config.toml`, or `$CODEX_HOME/config.toml` when
  `CODEX_HOME` is set.
- `claude-code`: writes `~/.claude.json`.

Aliases:

- `claude` is accepted as an alias for `claude-code`.

Examples:

```sh
ck add-mcp codex
ck add-mcp claude-code
ck add-mcp codex --project
ck add-mcp claude-code --command /absolute/path/to/ck-mcp
```

Project installs write `.codex/config.toml` for Codex and `.mcp.json` for
Claude Code.

## Future Chains

Future chains should follow the same shape:

```sh
ck eth gen
ck eth val 0x...
ck sol gen
ck sol bal ...
```

Avoid chain-specific binary names unless a chain requires a separate runtime.
