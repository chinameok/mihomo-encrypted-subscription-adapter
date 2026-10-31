# macOS LaunchAgent template

The template in this directory is intentionally generic. It contains no
machine-specific path, subscription URL, password, controller secret, or
company VPN material. Copy it to a protected local runtime directory and
replace the placeholder values during an explicitly authorized Mac install.

The adapter can also be run manually. Install the LaunchAgent only after a
dry-run succeeds and the Mihomo provider path has been staged and reviewed.
The runtime directory must be user-owned and mode `0700`; the URL/password
files must be mode `0600`.

`MEOK_MIHOMO_CONTROLLER_SECRET` should be supplied by a protected local
environment mechanism or native Keychain integration. Do not put it in this
template or in a public repository.
