# Security policy

## Reporting a vulnerability

Do not report a credential leak, decryption weakness, or exploitable bug in a
public GitHub issue. Use GitHub Private vulnerability reporting for this
repository when it is available, or contact the repository owner through the
private contact channel listed in the repository GitHub profile.

If a report contains a subscription URL, password, key, provider, or log,
revoke or rotate the credential first when possible. Never paste those values
into a public issue or pull request. If you accidentally disclose one, close
the page, remove the content, and rotate the credential; GitHub history and
caches may retain it.

Include a concise description, affected version/commit, reproduction using
synthetic data where possible, impact, and a suggested mitigation. Do not send
a live supplier response to reproduce a bug.

## Scope and security promises

The adapter is local software. It attempts to prevent accidental disclosure in
its own arguments, logs, errors, provider files, and release artifacts. It
cannot protect a compromised host, a password already exposed elsewhere, or a
subscription URL deliberately shared with another party.

The tool does not bypass provider authentication. It accepts a password from
the local user to decrypt the provider response and does not transmit that
password to the supplier as an authentication workaround. It does not inspect,
modify, or route through another user's traffic.

## Supported versions

Security fixes target the latest tagged release and the default branch. The
project is pre-1.0, so compatibility and support boundaries may change; the
changelog records such changes.

## Release integrity

Release artifacts include SHA-256 checksums and the source tag. Checksums
detect corruption but do not replace code signing, notarization, or complete
build provenance. Those controls are tracked separately and must be added
without weakening the credential boundary.
