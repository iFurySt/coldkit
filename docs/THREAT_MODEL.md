# Threat Model

`coldkit` is a local tool. It cannot make an infected or online machine safe for
private-key generation.

## Assets

- Private keys.
- Generated addresses intended for receiving funds.
- User confidence that watch-only workflows do not touch private keys.

## Defended Against

- Accidental use of private keys in watch-only balance checks.
- Accidental exposure of private-key-returning MCP tools to agents.
- Invalid TRON address input.
- Ambiguous CLI output for scripts and agents.
- Dependency sprawl in cold key-generation paths.

## Not Defended Against

- Malware on the machine running secret-generation commands.
- Compromised terminal, shell history, screen recording, or clipboard.
- User copying private keys into chat, tickets, cloud sync, or an online machine.
- Broken operating-system entropy sources.
- Supply-chain attacks outside the dependencies and binaries the user actually runs.
- Malicious forks or modified binaries.

## Trust Boundaries

- Cold CLI commands: local CPU and local entropy only; no network.
- Watch-only CLI commands: public address input, public network APIs, no secrets.
- MCP default mode: public data and public previews only.
- MCP secret mode: private keys may be returned to the MCP client; use offline only.

## Required Review For New Features

Any feature that handles secret material must answer:

- Does it perform network I/O?
- Does it write private keys to disk?
- Does it expose private keys to an AI agent, browser, clipboard, logs, or shell history?
- Is the feature disabled by default in MCP?
- Is there a public-only mode for previews and tests?

Any feature that performs network I/O must answer:

- Can it be called with a private key?
- Does it leak address metadata to a third-party API?
- Is the endpoint documented?
- Does it have a timeout?
