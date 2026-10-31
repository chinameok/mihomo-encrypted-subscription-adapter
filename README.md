# Mihomo Encrypted Subscription Adapter

`mihomo-encrypted-subscription-adapter` is a small, local command-line adapter
for subscription services that return an application-encrypted Clash/Mihomo
profile. It downloads a provider response, decrypts it with the password that
the user supplies to the service, validates the resulting profile, and writes
an independently managed Mihomo provider file.

The project is intentionally local-first. It does not proxy traffic itself,
replace Mihomo, or provide a hosted subscription conversion service.

> **Status:** experimental. The CLI and offline validation path are available;
> Mac Keychain, Mihomo-fork controller, and real traffic acceptance remain
> environment-specific gates. Treat pre-1.0 releases as unsuitable for
> unattended production use.

## Security boundary

The subscription URL is a bearer credential. Anyone who obtains it may be able
to retrieve the account's subscription. The website password is used locally
to decrypt the provider response; this tool does not bypass the provider's
authentication or recover a password.

The adapter must never send the password to a third party, print it, put it in
an argument list, or write it to a repository. On macOS the supported storage
is Keychain; a mode-600 file is an explicit fallback. Logs and errors are
redacted by design.

Never commit any of the following:

- a real subscription URL or token;
- a website password, private key, certificate, or VPN profile;
- a decrypted provider, node list, or real proxy endpoint;
- screenshots, logs, or issue attachments containing the above.

CI and release jobs use synthetic encrypted fixtures and fake HTTP/controller
servers only. They never contact a supplier endpoint.

## Threat model and non-goals

The adapter assumes the local user controls the host and can protect its
Keychain, filesystem, and process list. It protects secrets from accidental
repository, log, and error exposure, but it cannot protect a host that is
already compromised or a subscription URL that the user has intentionally
shared.

The adapter does not:

- defeat subscription encryption or provider access controls;
- validate that a proxy is trustworthy or safe to use;
- create a VPN tunnel or alter system routes;
- expose a LAN-wide proxy listener;
- modify Mihomo's core or silently switch the active traffic policy.

Decryption success, provider refresh success, and real traffic success are
separate claims. A release can be `decrypt_validated` without being
`provider_refreshed` or `traffic_verified`.

## Data flow

```text
subscription endpoint
        |
        v
bounded HTTP fetch -> decrypt -> YAML/profile validation
        |                            |
        |                            +--> redacted diagnostics
        v
atomic last-known-good provider file -> optional Mihomo controller refresh
```

Candidates are validated before replacement. A failed fetch, decryption,
validation, or controller refresh leaves the previous last-known-good provider
available. The adapter never replaces the active Mihomo configuration as
part of a routine refresh.

## Installation and usage

Release artifacts are published on the [GitHub Releases page][releases].
Download the artifact for your platform, verify its accompanying SHA-256
checksum, and place the binary in a user-owned directory. The exact packaging
and launchd integration are documented in [DEVELOP.md](DEVELOP.md) and the
generic [macOS LaunchAgent template](packaging/macos/README.md).

The safe command contract is:

```text
adapter status
adapter refresh --dry-run
adapter refresh
adapter set-password        # interactive; never a command-line secret
adapter set-url             # interactive; URL is redacted in output
adapter rollback
adapter disable
```

Use `--help` for the installed version's options. Do not paste a real URL or
password into shell history, an issue, or a chat transcript.

## Supported environments

| Environment | Source/build status | Runtime verification |
| --- | --- | --- |
| macOS Apple Silicon (`darwin/arm64`) | Primary target | Requires a user-authorized Mac acceptance run |
| Linux (`linux/amd64`) | CI cross-build target | No hosted service or VPN validation |
| Windows (`windows/amd64`) | CI cross-build target | No Windows runtime commitment in 1.x |

The adapter is designed to work with a Mihomo provider file and controller,
but provider field names and controller endpoints are validated against the
running Mihomo fork during acceptance. A green CI build does not prove that a
particular VPN, supplier, or network is reachable.

When controller refresh is enabled, set these environment variables in the
protected launchd environment or shell that runs the adapter:

```text
MEOK_MIHOMO_CONTROLLER_URL=http://127.0.0.1:19090
MEOK_MIHOMO_CONTROLLER_SECRET=<local secret>
MEOK_MIHOMO_PROVIDER_NAME=meok-subscription
```

`MEOK_MIHOMO_PROVIDER_REFRESH_URL` may override the provider refresh endpoint
when a Mihomo fork uses a different route. If the controller URL is unset, the
adapter stops after atomically publishing the provider file. A controller
failure rolls the file back to the last-known-good version when one exists.

## Development

Read [DEVELOP.md](DEVELOP.md) before changing code. The normal local checks
are:

```bash
gofmt -w .
go test -count=1 ./...
go test -race ./...
go vet ./...
go mod verify
```

Tests must use synthetic encrypted payloads and `httptest`; live subscription
URLs are prohibited. The CI workflow also runs a short fuzz smoke test,
`govulncheck`, secret-pattern checks, and cross-builds.

## Project governance

- [CONTRIBUTING.md](CONTRIBUTING.md) describes commits, tests, and review.
- [SECURITY.md](SECURITY.md) describes private vulnerability reporting.
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) sets the community standard.
- [docs/adr/](docs/adr/) records durable architecture decisions.
- [CHANGELOG.md](CHANGELOG.md) records user-visible changes.

This project uses Conventional Commit subjects and squash merges. Tagged
releases are built from source, include checksums, and state whether signing or
notarization is available. Checksums detect download corruption; they do not
prove a build's provenance.

## License

The project is available under the [MIT License](LICENSE). See
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for dependency and clean-room
attribution boundaries.

[releases]: https://github.com/chinameok/mihomo-encrypted-subscription-adapter/releases
