# coldkit Apple Setup

Generated: 2026-07-11

## Developer Account

- Team name: `Yifan PANG`
- Team ID: `J9P29FA5BX`

## Bundle Identifier

- Name: `coldkit`
- Bundle ID: `com.ifuryst.coldkit`
- App Store Connect / Developer Portal resource id: `22R3D6N87Q`
- Platform: `UNIVERSAL`
- Capabilities: none enabled

## Direct Distribution Signing

The npm release workflow signs Darwin `ck` and `ck-mcp` binaries with Developer
ID Application when the GitHub Actions secrets are configured.

Configured secret names:

- `COLDKIT_CODESIGN_P12_BASE64`
- `COLDKIT_CODESIGN_P12_PASSWORD`
- `COLDKIT_CODESIGN_KEYCHAIN_PASSWORD`
- `COLDKIT_CODESIGN_IDENTITY`
- `APPLE_DEVELOPER_TEAM_ID`
- `APPLE_NOTARY_API_KEY_P8_BASE64`
- `APPLE_NOTARY_KEY_ID`
- `APPLE_NOTARY_ISSUER_ID`

Current signing identity:

```text
Developer ID Application: Yifan PANG (J9P29FA5BX)
SHA-1: 188DEC067AF15D425FCA6B08FD101B9ABD01F571
```

## Notes

- Do not commit certificates, private keys, provisioning profiles, p12 exports,
  or App Store Connect `.p8` keys.
- The npm packages currently ship signed Darwin binaries. Notarized `.zip`,
  `.pkg`, or `.dmg` artifacts should be added as a separate GitHub Release
  workflow before claiming notarized direct downloads.
