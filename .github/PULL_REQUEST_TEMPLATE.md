## Summary

<!-- Describe the user-visible behavior and why it is needed. -->

## Security and compatibility

- [ ] No real subscription URL, password, provider, VPN material, or secret-bearing log is included.
- [ ] Tests use synthetic fixtures or local fakes only.
- [ ] This change affects: [ ] decrypt_validated [ ] provider_refreshed [ ] traffic_verified
- [ ] An ADR or changelog entry is included when needed.

## Verification

<!-- List exact commands and their results. -->

- [ ] gofmt
- [ ] go test -count=1 ./...
- [ ] go test -race ./...
- [ ] go vet ./...
- [ ] go mod verify
- [ ] Fuzz/malformed input behavior considered
