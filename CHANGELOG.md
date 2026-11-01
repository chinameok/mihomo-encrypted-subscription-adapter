# Changelog

All notable user-visible changes are recorded here. The project is pre-1.0;
minor releases may contain compatibility changes. Release notes are also
published on GitHub.

## [Unreleased]

### Added

- Manual release validation for all three platform archives, checksums, and
  GitHub provenance, with publishing disabled by default.

### Changed

- Release publishing now requires package and provenance verification and an
  unambiguous existing version tag. Validation creates only Actions artifacts
  and attestations, without creating a tag or formal release.

## [0.1.0] - 2026-10-31

### Added

- Initial local adapter for FlClash-encrypted Clash/Mihomo subscriptions.
- Bounded HTTPS fetch, AES-128-CBC/PKCS7 decryption, profile validation, and
  atomic last-known-good provider updates.
- Optional Mihomo controller refresh, protected local credentials, rollback,
  status, disable/enable commands, and a macOS LaunchAgent template.
- Public CI, cross-build and release workflows with checksums.

[unreleased]: https://github.com/chinameok/mihomo-encrypted-subscription-adapter/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/chinameok/mihomo-encrypted-subscription-adapter/releases/tag/v0.1.0
