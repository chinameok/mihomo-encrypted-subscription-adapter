# Agent and automation guide

This repository contains a local security-sensitive adapter. Read this file
and CONTEXT.md before making changes.

## Non-negotiable boundaries

- Never read, request, print, commit, or upload a real subscription URL,
  password, decrypted provider, VPN profile, private key, or secret-bearing
  log.
- Use synthetic fixtures and fake HTTP/controller servers in tests.
- Do not contact a supplier endpoint from CI.
- Do not modify Mihomo core, system routes, VPN configuration, or a user's
  active traffic policy as part of a code change.
- Keep credential handling, decryption, provider output, and controller calls
  behind explicit package boundaries.

## Verification language

Use the project states precisely:

- decrypt_validated: synthetic or authorized local response decrypted and
  validated.
- provider_refreshed: a provider candidate was atomically installed and the
  local controller accepted it.
- traffic_verified: an authorized runtime probe confirmed a particular network
  path. CI cannot claim this state.

## Change process

Read DEVELOP.md, CONTRIBUTING.md, and relevant docs/adr/ records. Keep
changes focused, run the documented checks, and update the changelog or ADR
when the public contract changes. Do not weaken a check to hide a failure.

The repository owner makes release and runtime decisions. Agents may prepare
changes and evidence; they must not silently publish credentials, alter a live
Mac, or claim real traffic success from a build-only result.
