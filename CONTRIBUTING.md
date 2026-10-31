# Contributing

Thank you for improving the adapter. Contributions are welcome when they keep
the local-only security boundary intact.

## Before you start

Read README.md, DEVELOP.md, and relevant architecture decision records. For a
behavior change, open an issue first unless the fix is small and obvious. Do
not attach a real supplier response, provider, URL, password, VPN profile, or
log.

## Contribution rules

1. Use a focused branch and a Conventional Commit subject.
2. Add or update synthetic tests for behavior changes.
3. Run formatting, unit, race, vet, module verification, and relevant fuzz
   smoke checks locally.
4. Keep credentials out of code, fixtures, logs, issue comments, and PR
   descriptions.
5. Update the changelog for user-visible behavior and an ADR for durable
   architectural choices.
6. Explain compatibility, security impact, and rollback in the PR body.

Clean-room compatibility work must be based on publicly observable protocol
behavior and independently written code. Do not copy implementation code from
FlClash or another client into this MIT-licensed repository. See
THIRD_PARTY_NOTICES.md.

## Pull request checklist

- [ ] No real URL, token, password, provider, VPN material, or secret-bearing
      log is included.
- [ ] Tests use synthetic fixtures or local fakes only.
- [ ] gofmt, unit tests, race tests, go vet, and go mod verify pass.
- [ ] Fuzz or malformed-input behavior was considered where parsing changes.
- [ ] Error and log output remains redacted.
- [ ] README, changelog, ADR, or security documentation is updated as needed.
- [ ] The PR states whether this affects decrypt_validated,
      provider_refreshed, or traffic_verified.

Maintainers may request a smaller change, additional tests, or a private
reproduction. Public issue discussion is never a place to exchange credentials.
