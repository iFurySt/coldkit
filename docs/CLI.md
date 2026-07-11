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

Watch-only commands must never accept private keys.

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
