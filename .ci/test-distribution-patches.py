#!/usr/bin/env python3
"""Check that dependency source drift cannot leave a partially patched vendor tree."""

import importlib.util
import json
from pathlib import Path
import re
import sys
import tempfile
import unittest

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
spec = importlib.util.spec_from_file_location("compat", ROOT / "build/docker/patch-distribution-compat.py")
compat = importlib.util.module_from_spec(spec)
spec.loader.exec_module(compat)


class DistributionPatchTests(unittest.TestCase):
    def test_renovate_tracks_every_shared_dependency_pin(self):
        configuration = json.loads((ROOT / "renovate.json").read_text())
        source = (ROOT / "build/docker/prepare-distribution.sh").read_text()
        manager = next(manager for manager in configuration["customManagers"]
                       if any("prepare-distribution" in pattern for pattern in manager["managerFilePatterns"]))
        # This manager uses the common RE2/Python regex subset; only named-group
        # notation differs. Exercise the actual checked-in extraction pattern.
        pattern = manager["matchStrings"][0].replace("(?<", "(?P<")
        extracted = {match["depName"]: match["currentValue"] for match in re.finditer(pattern, source)}
        annotated = re.findall(r"# renovate: datasource=go depName=(\S+)", source)
        self.assertEqual(set(extracted), set(annotated))
        self.assertIn("golang.org/x/oauth2", extracted)
        self.assertEqual(len(extracted), 11)

    def test_renovate_tracks_adapter_module_overrides(self):
        configuration = json.loads((ROOT / "renovate.json").read_text())
        source = (ROOT / "build/docker/Dockerfile.trivy-adapter").read_text()
        manager = next(manager for manager in configuration["customManagers"]
                       if any("Dockerfile" in pattern and "trivy-adapter" in pattern
                              for pattern in manager["managerFilePatterns"]))
        pattern = manager["matchStrings"][0].replace("(?<", "(?P<")
        extracted = {match["depName"] for match in re.finditer(pattern, source)}
        self.assertEqual(extracted, {"golang.org/x/net", "golang.org/x/crypto",
                                     "golang.org/x/sys", "golang.org/x/text",
                                     "github.com/redis/go-redis/v9"})

    def fixture(self, root):
        acl = root / "vendor/google.golang.org/cloud/storage/acl.go"
        dns = root / "vendor/github.com/xenolf/lego/acme/dns_challenge.go"
        acl.parent.mkdir(parents=True)
        dns.parent.mkdir(parents=True)
        acl.write_text(compat.LEGACY + "\n" + compat.LEGACY)
        dns.write_text("if err == dns.ErrTruncated {")
        return acl, dns

    def test_complete_patch_is_repeatable(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            acl, dns = self.fixture(root)
            compat.patch(root)
            first = (acl.read_text(), dns.read_text())
            self.assertEqual(first[0].count(compat.TYPED), 2)
            self.assertEqual(first[1], "if in != nil && in.Truncated {")
            compat.patch(root)
            self.assertEqual((acl.read_text(), dns.read_text()), first)

    def test_unexpected_acl_source_stops_both_patches(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            acl, dns = self.fixture(root)
            acl.write_text(compat.LEGACY)
            before = (acl.read_text(), dns.read_text())
            with self.assertRaisesRegex(ValueError, "refusing partial patch"):
                compat.patch(root)
            self.assertEqual((acl.read_text(), dns.read_text()), before)

    def test_unexpected_dns_source_stops_both_patches(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            acl, dns = self.fixture(root)
            dns.write_text("unexpected upstream source")
            before = (acl.read_text(), dns.read_text())
            with self.assertRaisesRegex(ValueError, "refusing partial patch"):
                compat.patch(root)
            self.assertEqual((acl.read_text(), dns.read_text()), before)


if __name__ == "__main__":
    unittest.main()
