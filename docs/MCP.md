# MCP

The MCP server binary is `ck-mcp`. It is designed for AI agents and defaults to
safe, public-data-only tools.

## Start Modes

Default mode:

```sh
ck-mcp
```

Default mode exposes only tools that either use public addresses or return
public previews.

Secret mode:

```sh
ck-mcp --enable-secret-tools
```

Secret mode exposes private-key-returning tools and should only be used on an
offline machine.

## Current Tools

Default tools:

- `tron_validate`: validate a public TRON address offline.
- `tron_balance`: check public TRX and USDT/TRC20 balances.
- `tron_generate_preview`: generate public address previews without returning private keys.

Secret tools:

- `tron_generate_secret`: generate TRON accounts and return private keys. Disabled by default.

## Tool Naming

MCP tools use `<chain>_<action>`:

```text
tron_validate
tron_balance
tron_generate_preview
tron_generate_secret
```

Future examples:

```text
eth_validate
eth_generate_preview
sol_balance
```

## Agent Safety Defaults

- Do not expose secret tools by default.
- Do not accept private keys in watch-only tools.
- Tool descriptions must clearly say when a tool performs network I/O.
- Tool descriptions must clearly say when a tool can return private keys.
- Prefer JSON content in text responses so agents can parse deterministic fields.

## Non-Goals

- No send/broadcast transaction tools in the MVP.
- No remote signing.
- No seed phrase export over MCP by default.
- No clipboard helpers.
- No webhooks or POST-to-URL output paths.

## Protocol Scope

The current server implements a small stdio JSON-RPC MCP surface:

- `initialize`
- `notifications/initialized`
- `tools/list`
- `tools/call`

If the server grows beyond simple local tools, add tests and document the
supported MCP protocol version and client compatibility here.
