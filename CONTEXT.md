# Project context

## Product boundary

This repository provides a local adapter for application-encrypted
Clash/Mihomo subscription responses. Mihomo remains the proxy engine. The
adapter fetches, decrypts, validates, and atomically publishes a provider
file; it does not become a hosted converter or a VPN client.

## Terms

- Subscription URL: bearer credential capable of retrieving a subscription.
- Website password: local input used by the supplier protocol to decrypt a
  response. It is not a command-line argument or a bypass mechanism.
- Provider candidate: validated output written to a temporary location before
  atomic replacement.
- Last-known-good: the most recent provider accepted by local validation;
  failures preserve it.
- decrypt_validated: decryption and profile validation succeeded.
- provider_refreshed: Mihomo accepted a refreshed provider through its local
  controller.
- traffic_verified: an authorized runtime probe confirmed a particular network
  path. This is never inferred from CI.

## Source of truth

Code and tests define current behavior. ADRs define durable architecture
choices. README and DEVELOP describe the public contract. Private machine
paths, VPN material, live subscription data, and acceptance notes stay outside
this repository.

## Version and release policy

The project uses Conventional Commits, squash merges, semantic tags, and a
curated changelog. It is pre-1.0 and may change compatibility in minor
releases. Dependabot is the sole automated dependency updater.
