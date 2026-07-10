# CLI

The CLI binary is `ck`. Commands are chain-first so the project can grow beyond
TRON without renaming the tool.

## Shape

```sh
ck <chain> <command> [flags]
```

Current TRON commands:

```sh
ck tron gen
ck tron val T...
ck tron bal T...
ck tron from-private <PRIVATE_KEY_HEX>
ck tron self
```

Top-level command:

```sh
ck self
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

## Future Chains

Future chains should follow the same shape:

```sh
ck eth gen
ck eth val 0x...
ck sol gen
ck sol bal ...
```

Avoid chain-specific binary names unless a chain requires a separate runtime.
