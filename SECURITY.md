# Security Policy

## Reporting a Vulnerability

If you find a security vulnerability in AgeForge, **please do not open a public GitHub issue.**

Instead, report it privately via one of these channels:

- **Discord**: DM `espresso20` on the [AgeForge Discord](https://discord.gg/EPvyd5vjpj)
- **GitHub**: Use [GitHub's private vulnerability reporting](https://github.com/espresso20/ageforge/security/advisories/new)

Please include:
- A description of the vulnerability and its potential impact
- Steps to reproduce it
- Any relevant logs, screenshots, or proof-of-concept

I'll acknowledge the report within 48 hours and aim to release a fix within 14 days for valid issues.

## Scope

AgeForge is a local terminal application — it reads/writes save files to disk and makes outbound-only network requests to check for updates (GitHub releases API). There is no server, no user accounts, and no data collection.

The main attack surface is:

| Area | Notes |
|------|-------|
| **Save file parsing** | Malformed JSON save files could cause crashes |
| **Auto-updater** | Downloads binaries from GitHub releases; SHA256 is verified before install |
| **Website** | Static Netlify site; no server-side code |

## Supported Versions

Only the latest release is actively maintained. If you're on an older version, please update before reporting.
