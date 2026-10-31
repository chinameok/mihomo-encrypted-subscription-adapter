---
status: accepted
date: 2026-10-31
---

# ADR-0003: Keep credentials in protected local stores

## Decision

Store the subscription URL and website password only in a user-protected
local store. macOS Keychain is the preferred store; a mode-600 file is an
explicit fallback for environments without Keychain support.

The CLI accepts secrets interactively or through a protected store reference.
It does not accept a password as a normal command-line argument, print the URL
in status output, or include secret values in structured logs and errors.

## Rationale

The URL is a bearer credential, and the password can decrypt the response.
Command-line arguments, shell history, crash reports, logs, and public CI are
poor locations for either value.

## Consequences

- Tests use fake credentials and verify redaction.
- URL rotation and password rotation are separate operations.
- A runtime diagnostic may show a host and a redacted identifier, never the
  complete URL or decrypted profile.
- Users must rotate a credential if it was pasted into a public issue or log.
