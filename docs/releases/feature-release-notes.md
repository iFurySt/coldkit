# Feature Release Notes

## 2026-07

| Date | Area | User Impact | Change Summary |
| --- | --- | --- | --- |
| 2026-07-19 | TRON TRC20 Safety | Operators can generate USDT/TRC20 transfer call data with local ABI round-trip checks and optional full-node dry-run simulation before signing or broadcasting elsewhere. | Added `ck tron trc20-transfer` with amount precision validation, transfer parameter generation, call data output, and `--owner` dry-run handling for `REVERT`/`FAILED` responses. |
| 2026-07-16 | TRON Watch-Only | Developers can query mainnet, Nile, or Shasta with `--network` instead of copying testnet endpoint URLs into every command. | Added built-in endpoint pools for `mainnet`, `nile`, and `shasta`; exposed network selection in CLI and MCP watch-only tools while keeping custom endpoint overrides. |
| 2026-07-15 | TRON Watch-Only | TRON balance and resource checks are less likely to fail from public API rate limits because they use full node calls with endpoint fallback. | Replaced the indexed account lookup with `/wallet/*` full node calls, added retry-style fallback across endpoints, and made `--endpoint` repeatable for custom pools. |
| 2026-07-15 | TRON Watch-Only | TRON balance checks now include Energy and Bandwidth, and operators can query resources directly with `ck tron resource` or MCP `tron_resource`. | Added account resource parsing through TronGrid, merged resources into `tron_balance`, and documented the expanded CLI/MCP surface. |

## 2026-04

| Date | Area | User Impact | Change Summary |
| --- | --- | --- | --- |
| 2026-04-08 | Template | Introduced the base harness repository template for future services and products. | Added agent entry docs, execution-plan scaffolding, change-history templates, and docs checks. |
