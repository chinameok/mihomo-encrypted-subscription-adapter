---
status: accepted
date: 2026-10-31
---

# ADR-0004: Refresh Mihomo through a staged provider integration

## Decision

Integrate the adapter provider beside the existing Mihomo configuration first.
Validate a staged configuration, install the provider definition, and refresh
it through the local controller while the current traffic strategy remains the
default. A public proxy-group cutover is a separately authorized operation.

Company VPN entries, company-domain rules, LAN binding, TUN state, and existing
rollback backups are outside the adapter's automatic scope.

## Rationale

The adapter must not turn a subscription update into an unreviewed network
migration. Staging isolates provider parsing and controller compatibility from
traffic-policy changes and gives the user a simple rollback.

## Consequences

- Provider visibility and controller acceptance are tested before cutover.
- The UI/controller endpoint is treated as a local integration, not a hosted
  service.
- Real Mac acceptance is required for traffic_verified.
- CI can claim only synthetic decrypt and provider behavior.
