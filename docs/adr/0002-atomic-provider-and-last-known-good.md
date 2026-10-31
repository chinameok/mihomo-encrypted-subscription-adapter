---
status: accepted
date: 2026-10-31
---

# ADR-0002: Publish validated providers atomically

## Decision

Fetch and decrypt into a bounded temporary location. Validate the profile and
provider entries before replacing the active provider using an atomic rename.
Retain a last-known-good copy and preserve it when any step fails.

The adapter may refresh a local Mihomo provider through its controller after
the file is safely installed. Routine refreshes never replace the full Mihomo
configuration and never silently change the active traffic policy.

## Rationale

A partial download, wrong password, malformed node, disk error, or controller
failure must not remove a working provider. Atomic replacement gives the
runtime a complete file, while the last-known-good copy supplies a predictable
rollback point.

## Consequences

- Temporary files must be created in the same filesystem as the destination.
- Candidate size, node count, and required fields are bounded and validated.
- Failure output contains categories and redacted context only.
- Runtime tests must distinguish provider refresh from real traffic success.
