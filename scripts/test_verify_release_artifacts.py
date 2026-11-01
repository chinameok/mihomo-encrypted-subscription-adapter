"""Synthetic package fixtures; no supplier input or runtime is used."""

import contextlib
import hashlib
import importlib.util
import io
from pathlib import Path
import tarfile
import tempfile
import unittest


SPEC = importlib.util.spec_from_file_location(
    "release_verifier", Path(__file__).with_name("verify-release-artifacts.py")
)
VERIFIER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VERIFIER)


class VerifyReleaseArtifactsTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.directory = Path(self.temporary.name)
        self.version = "validation-0123456789ab"
        self.archives = []
        for target in VERIFIER.TARGETS:
            binary = f"mihomo-subscription-adapter-{target}"
            archive = self.directory / f"{binary}-{self.version}.tar.gz"
            self.write_archive(archive, binary)
            self.archives.append(archive)

    @staticmethod
    def write_archive(archive, name, *, kind=tarfile.REGTYPE, content=b"synthetic binary"):
        with tarfile.open(archive, "w:gz") as package:
            member = tarfile.TarInfo(name)
            member.type = kind
            member.size = len(content) if kind == tarfile.REGTYPE else 0
            package.addfile(member, io.BytesIO(content))
        archive.with_name(archive.name + ".sha256").write_text(
            f"{hashlib.sha256(archive.read_bytes()).hexdigest()}  {archive.name}\n",
            encoding="utf-8",
        )

    def verify(self):
        with contextlib.redirect_stdout(io.StringIO()):
            VERIFIER.verify(self.directory, self.version)

    def test_complete_set(self):
        self.verify()

    def test_missing_artifact(self):
        self.archives[0].unlink()
        with self.assertRaisesRegex(ValueError, "exactly three"):
            self.verify()

    def test_extra_staged_binary(self):
        (self.directory / "stage").mkdir()
        with self.assertRaisesRegex(ValueError, "exactly three"):
            self.verify()

    def test_wrong_checksum(self):
        self.archives[0].with_name(self.archives[0].name + ".sha256").write_text("bad\n")
        with self.assertRaisesRegex(ValueError, "checksum mismatch"):
            self.verify()

    def test_windows_binary_name(self):
        self.write_archive(self.archives[-1], "mihomo-subscription-adapter-windows-amd64")
        with self.assertRaisesRegex(ValueError, "unexpected archive contents"):
            self.verify()

    def test_archive_traversal(self):
        self.write_archive(self.archives[0], "../mihomo-subscription-adapter-darwin-arm64")
        with self.assertRaisesRegex(ValueError, "unexpected archive contents"):
            self.verify()

    def test_archive_link(self):
        self.write_archive(
            self.archives[0], "mihomo-subscription-adapter-darwin-arm64", kind=tarfile.SYMTYPE
        )
        with self.assertRaisesRegex(ValueError, "nonempty regular binary"):
            self.verify()

    def test_empty_binary(self):
        self.write_archive(self.archives[0], "mihomo-subscription-adapter-darwin-arm64", content=b"")
        with self.assertRaisesRegex(ValueError, "nonempty regular binary"):
            self.verify()

    def test_invalid_version(self):
        self.version = "../main"
        with self.assertRaisesRegex(ValueError, "invalid release"):
            self.verify()


if __name__ == "__main__":
    unittest.main()
