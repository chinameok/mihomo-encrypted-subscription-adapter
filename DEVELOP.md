# Development guide

This document is for contributors and maintainers. It keeps terminal-user
instructions in README.md and all real subscription credentials outside the
repository.

## Prerequisites

- Go 1.22 or newer for local development. CI and release builds pin Go
  1.26.10 so the vulnerability scan uses a patched standard library.
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

The `main` ruleset requires a pull request, resolved review conversations,
and the latest base branch before merging. Required GitHub Actions checks are
`Go checks`, `Cross-build darwin/arm64`, `Cross-build linux/amd64`, and
`Cross-build windows/amd64`. Force pushes and deletion are blocked, with no
bypass. A single maintainer can merge without a second approval; successful
checks are still required.

## Cross-builds and release artifacts

The check workflow builds the supported matrix and uploads seven-day artifacts
for maintainer testing. The release workflow builds the CLI with
`-trimpath -buildvcs=false` from the pushed version tag, emits SHA-256 checksum
files, and publishes GitHub build provenance for each archive. A tagged CI
release artifact is the preferred input for the Mac canary; it keeps the
installed binary tied to a reviewed source tag and the workflow evidence.

Both validation and publishing download the combined three-platform package
set and verify the exact six files (three archives and three checksums). Each
archive must contain only its correctly named, nonempty binary; Windows uses
`.exe`. The verification job also checks GitHub provenance against this
repository, `.github/workflows/release.yml`, and the selected source commit
and ref. The publish job runs only after this verification succeeds.

Before merging workflow or dependency changes, exercise the complete release
path without creating a formal release:

~~~bash
gh workflow run release.yml --repo chinameok/mihomo-encrypted-subscription-adapter \
  --ref YOUR_REVIEW_BRANCH -f publish=false
gh run list --repo chinameok/mihomo-encrypted-subscription-adapter \
  --workflow release.yml --branch YOUR_REVIEW_BRANCH --limit 1
gh run watch RUN_ID --repo chinameok/mihomo-encrypted-subscription-adapter --exit-status
~~~

With no `tag` input, validation builds the exact commit selected by the
workflow dispatch, with archive version `validation-SHORT_SHA`. This works for
a review branch or `main`; it does not silently build a release tag. Manual
runs default to `publish=false`. Validation still writes seven-day Actions
artifacts and GitHub build provenance attestations, which are public in this
public repository. It does not create or move Git tags, create a GitHub
Release, contact a supplier, or install a binary on a user machine.

The package verifier can be checked locally using synthetic fixtures:

~~~bash
python3 -B -m unittest discover -s scripts -p 'test_*.py'
~~~

A local equivalent for the primary target is:

~~~bash
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64   go build -trimpath -buildvcs=false   -o dist/mihomo-subscription-adapter-darwin-arm64   ./cmd/mihomo-subscription-adapter
~~~

Do not claim that a cross-build is a Mac runtime acceptance. Real Keychain,
launchd, Mihomo controller, and traffic checks require the authorized Mac
procedure described by the private deployment runbook.

To publish a release, first merge the reviewed change, create an annotated
semantic version tag, and push that tag. The tag push starts the release
workflow; no supplier URL, password, provider, or VPN material is needed:

~~~bash
git tag -a v0.1.0 -m "release v0.1.0"
git push origin v0.1.0
~~~

The release job verifies that the tag exists before creating the GitHub
Release. It does not run for ordinary local builds or `main` pushes.

To manually validate or publish an existing tag, both the workflow ref and
the `tag` input must name the same existing `vMAJOR.MINOR.PATCH` tag. This
binds the provenance source revision to the actual code being built. For
example, for a future tag containing this workflow:

~~~bash
gh workflow run release.yml --repo chinameok/mihomo-encrypted-subscription-adapter \
  --ref v0.2.0 -f tag=v0.2.0 -f publish=false
# Change publish=false to publish=true only when a formal release is intended.
~~~

Manual publishing rejects an absent/invalid tag, a missing tag, or a workflow
ref that differs from the requested tag. It never interprets the tag input as
an arbitrary branch name. Pushing a valid version tag continues to publish
automatically after the same package and provenance checks. Re-running an
already published tag does not overwrite its release. CI artifacts are build
evidence; passing these checks alone does not require a runtime binary update.

## Versioning and dependency updates

Tags use vMAJOR.MINOR.PATCH. Before 1.0, compatibility may change between
minor versions and the changelog must call that out. Dependabot is the only
automated dependency updater; do not add Renovate or a second updater for the
same ecosystems.

Release notes use GitHub's generated notes from the source tag and the curated
CHANGELOG.md remains the human-maintained user-facing record. A release
contains source tag, binaries, checksums, provenance attestations, and a note
about signing/notarization status. No release job contacts a subscription
endpoint.

## Maintainer-only acceptance boundary

The repository CI and tests stop at decrypt_validated and synthetic
provider_refreshed. A real Mac test may continue to traffic_verified, but it
must be performed interactively with credentials stored in Keychain and must
not upload logs, providers, or URLs. Keep private acceptance notes outside this
repository.
