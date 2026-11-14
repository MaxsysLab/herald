# Security Policy

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues or pull
requests.**

Report privately through
[GitHub Private Vulnerability Reporting](https://github.com/MaxsysLab/herald/security/advisories/new)
(preferred), or by email to **security@maxsys.org**.

Please include a description of the issue, steps to reproduce or a proof of concept,
the affected version, and any fix you would suggest. If you would like credit, tell us
how to name you.

## What to expect

- An **acknowledgement** within **7 days** of your report.
- An initial **severity assessment** and a triage decision shortly after.
- We work on a fix under a coordinated **embargo**. The standard disclosure window is
  **90 days** from your report, or sooner once a fix is released, whichever comes first.
- On release we publish a **GitHub Security Advisory**, request a CVE where warranted,
  and credit reporters who want it.

## Known limits

Herald is an early release, and some limits are known and documented rather than
hidden. The most important one: **the management console has no sign-in yet**. These
are listed with their workarounds in [docs/SECURITY-MODEL.md](docs/SECURITY-MODEL.md).
You do not need to report them privately. Ideas and pull requests that close them are
welcome in the open.

## Supported versions

| Version | Supported |
|---|---|
| latest | yes |
| older | best effort |

## Safe harbor

We will not pursue or support legal action against researchers who act in good faith,
follow this policy, avoid privacy violations and service disruption, and give us
reasonable time to respond before any public disclosure.
