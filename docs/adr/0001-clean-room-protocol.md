---
status: accepted
date: 2026-10-31
---

# ADR-0001: Implement the encrypted subscription contract independently

## Decision

Implement the supplier response contract from public observations and
synthetic vectors. Keep the adapter code independent of FlClash, Mihomo, and
other client source trees.

The implementation may document compatible wire facts, such as response
encoding, key derivation, block cipher mode, padding, and profile shape, when
those facts are needed for interoperability. It must not copy source code,
tests containing real responses, or restricted-license implementation details.

## Rationale

The adapter is distributed under the MIT license and must keep a clear
clean-room boundary. A protocol is interoperable behavior; copied source code
and fixtures carry separate copyright and licensing obligations.

## Consequences

- Synthetic vectors are mandatory for CI.
- Any borrowed terminology or required notice is recorded in
  THIRD_PARTY_NOTICES.md.
- A compatibility change requires review of this ADR and a changelog entry.
- Supplier URLs, passwords, and decrypted nodes remain private.
