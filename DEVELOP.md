# Development guide

This document is for contributors and maintainers. It keeps terminal-user
instructions in README.md and all real subscription credentials outside the
repository.

## Prerequisites

- Go version declared by go.mod (CI uses that module version).
- Git and a working shell.
- Optional: mise or another local tool version manager.

The portable build stores the URL and password in a mode-600 local directory.
The macOS deployment may replace that store with a native Keychain backend;
that backend is an environment-specific packaging decision and is not used by
the Linux CI build.

Do not install a supplier profile or log in to a supplier account to run unit
tests. A live acceptance run is a separate, user-authorized operation on the
target Mac.

## Repository layout

~~~text
cmd/mihomo-subscription-adapter/  CLI entry point
internal/decrypt/                  encrypted response handling
internal/fetch/                    bounded HTTP transport
  internal/provider/                 YAML validation and provider output
  internal/credentials/              Keychain/file credential boundaries
  internal/runtime/                  atomic files, locking, controller calls
  packaging/macos/                   generic LaunchAgent template
  docs/adr/                          accepted architecture decisions
~~~

Keep supplier-specific behavior behind small interfaces. Do not copy code
from FlClash or another client; use public protocol observations and
clean-room tests instead. See ADR-0001 in docs/adr/.

## Local checks

Run these from the repository root:

~~~bash
gofmt -w .
go test -count=1 ./...
go test -race ./...
go vet ./...
go mod verify
~~~

If fuzz targets exist, run a bounded smoke test:

~~~bash
go test ./internal/decrypt -run=^$ -fuzz=Fuzz -fuzztime=10s
~~~

For vulnerability checks, install the pinned govulncheck version used by
.github/workflows/check.yml, then run:

~~~bash
govulncheck ./...
~~~

The workflow is authoritative for exact tool versions and the build matrix.
Never make a check pass by weakening a test, removing a fixture, or embedding
a real provider response.

## Synthetic fixtures only

Fixtures must be generated from fake node names, fake endpoints, and synthetic
passwords. Exercise successful decrypt, wrong password, malformed Base64/PKCS7,
missing provider fields, duplicate names, oversized input, and atomic rollback.

Before opening a PR, inspect the diff and search for credential-shaped values:

~~~bash
git diff --check
git grep -nE 'AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9_]{20,}' -- . ':!go.sum' || true
~~~

The command is a prompt for review, not a substitute for human inspection.
Remove any match that is not a deliberately documented test marker before
submitting the PR.

## Commit and pull request workflow

Use a Conventional Commit subject, for example:

~~~text
feat: add provider normalization
fix: keep last known good provider on refresh failure
docs: clarify credential storage
~~~

Keep commits focused and prefer squash merge. Pull requests should explain
the user-visible behavior, tests run, security impact, and rollback path. A
PR must state explicitly whether it changes the decryption contract,
credential handling, provider output, or release artifacts.

## Cross-builds and release artifacts

The release workflow builds the CLI with -trimpath -buildvcs=false for the
supported matrix and emits SHA-256 checksum files. A local equivalent for the
primary target is:

~~~bash
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64   go build -trimpath -buildvcs=false   -o dist/mihomo-subscription-adapter-darwin-arm64   ./cmd/mihomo-subscription-adapter
~~~

Do not claim that a cross-build is a Mac runtime acceptance. Real Keychain,
launchd, Mihomo controller, and traffic checks require the authorized Mac
procedure described by the private deployment runbook.

## Versioning and dependency updates

Tags use vMAJOR.MINOR.PATCH. Before 1.0, compatibility may change between
minor versions and the changelog must call that out. Dependabot is the only
automated dependency updater; do not add Renovate or a second updater for the
same ecosystems.

Release notes are derived from Conventional Commit subjects and the curated
CHANGELOG.md. A release contains source tag, binaries, checksums, and a note
about signing/notarization status. No release job contacts a subscription
endpoint.

## Maintainer-only acceptance boundary

The repository CI and tests stop at decrypt_validated and synthetic
provider_refreshed. A real Mac test may continue to traffic_verified, but it
must be performed interactively with credentials stored in Keychain and must
not upload logs, providers, or URLs. Keep private acceptance notes outside this
repository.
