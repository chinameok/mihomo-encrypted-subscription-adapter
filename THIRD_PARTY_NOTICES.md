# Third-party notices

The implementation is intended to be an independent, clean-room
compatibility adapter. It does not copy source code from FlClash, Mihomo, or
another client. Protocol behavior is implemented from publicly observable
responses and independently written tests.

## Dependencies

The final dependency list is generated from go.mod and go.sum when the
implementation lands. Every non-standard dependency must retain its license
notice here or through the dependency documented notice mechanism. The Go
standard library is distributed under its own license terms.

Do not add a dependency whose license is incompatible with this repository MIT
distribution without documenting the decision and obtaining maintainer
approval. Do not vendor copied implementation code merely to avoid attribution.

## References

Public protocol and compatibility references may be linked from ADRs. A
reference is not permission to copy its source code. If a future dependency
requires a notice, add the exact version, project, license, and required text
before publishing a release.
