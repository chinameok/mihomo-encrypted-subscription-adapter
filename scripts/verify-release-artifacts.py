#!/usr/bin/env python3
"""Validate the complete release package set without extracting archives."""

import hashlib
from pathlib import Path
import re
import sys
import tarfile


TARGETS = ("darwin-arm64", "linux-amd64", "windows-amd64.exe")


def verify(directory: Path, version: str) -> None:
    version_pattern = (
        r"v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)"
        r"|validation-[0-9a-f]{12}"
    )
    if not re.fullmatch(version_pattern, version):
        raise ValueError("invalid release or validation version")
    packages = {
        f"mihomo-subscription-adapter-{target}-{version}.tar.gz":
        f"mihomo-subscription-adapter-{target}"
        for target in TARGETS
    }
    expected = set(packages) | {f"{name}.sha256" for name in packages}
    if {entry.name for entry in directory.iterdir()} != expected:
        raise ValueError("expected exactly three archives and their three checksums")
    for name, binary in packages.items():
        archive = directory / name
        checksum = directory / f"{name}.sha256"
        if any(path.is_symlink() or not path.is_file() for path in (archive, checksum)):
            raise ValueError("release artifacts must be regular files")
        digest = hashlib.sha256()
        with archive.open("rb") as stream:
            for block in iter(lambda: stream.read(1024 * 1024), b""):
                digest.update(block)
        if checksum.read_text(encoding="utf-8") != f"{digest.hexdigest()}  {name}\n":
            raise ValueError(f"checksum mismatch: {name}")
        with tarfile.open(archive, "r:gz") as package:
            members = package.getmembers()
            if len(members) != 1 or members[0].name != binary:
                raise ValueError(f"unexpected archive contents: {name}")
            if not members[0].isfile() or members[0].size == 0:
                raise ValueError(f"expected a nonempty regular binary: {name}")
        print(f"Verified package and checksum: {name}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit("usage: verify-release-artifacts.py DIRECTORY VERSION")
    try:
        verify(Path(sys.argv[1]), sys.argv[2])
    except (ValueError, OSError, tarfile.TarError) as error:
        sys.exit(str(error))
