#!/usr/bin/env python3
"""Fail closed when the exact source behind Distribution's VEX changes."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
from unittest.mock import patch

ROOT = Path(__file__).resolve().parent.parent
SOURCE = "https://github.com/goharbor/distribution.git"
TAG = "v2.8.3-harbor.2"
COMMIT = "ac7c4cedc2f4d45347630beceaaa1639bb9c771a"
PURL = "pkg:golang/github.com/docker/distribution@v0.0.0-20260921091602-ac7c4cedc2f4%2Bdirty"
IDS = {"CVE-2017-11468", "CVE-2023-2253", "GHSA-qq97-vm5h-rrhg"}
# Fingerprints cover complete fixed implementations, not just matching snippets.
# Independently verified fixed commits 91c507a, 521ea3d and 10ade61 are ancestors
# of COMMIT. Reassess the primary advisories before changing these attestations.
FINGERPRINTS = {
    "registry/handlers/manifests.go": "50c3c924181025f0484fff3793fce254402cffb1e7ba0256787b1f889d9860bf",
    "registry/handlers/helpers.go": "a868019a3fb700a33e6b948dfdb27d18089f5db825f46f0e29f416fd051b2e90",
    "registry/handlers/catalog.go": "cd2526198e3baa1656231ab643a3e820dbb745ab7d8d9f312897b7e481279c6d",
    "registry/storage/blobstore.go": "afd4fc3b8bf30d536d58bfe3c08c82d7d82e373636207a39522de749e45abbce",
    "registry/storage/io.go": "8db74725a48b39b4777ef031637f68b5456a926b0c566c3cc56d71b284166144",
    "manifest/ocischema/manifest.go": "2fb8562d3924a969a2ff0a1c00f1328f7b4d2e0c1a994e7e97d9b3d5963aeffa",
    "manifest/manifestlist/manifestlist.go": "b931ce2ff27f45614720ddc33f0360f09fe8e82d197636b5c2b5fd4ef379386f",
    "configuration/configuration.go": "70167cc6c2a64cdb21095a5fa183d5c78403a0192e20582e4928de35884e8a39",
}
# Both reviewed variants retain identical catalog defaults: the release's
# existing uploadpurging map-conversion patch only changes that parser block.
APP_FINGERPRINTS = {
    "b32b27b33aafbaca6fb4e550ff5fa2241d7f730ae8442d2cc6cc1fb964058639",
    "595f2476bb32575a6bf0a6988ef815522bb65975bacebd8fa7c916236d7be61d",
}


def check_configuration(root=ROOT):
    for name in ("registry", "registryctl"):
        source = (root / f"build/docker/Dockerfile.{name}").read_text()
        for key, value in (("DISTRIBUTION_VERSION", TAG), ("DISTRIBUTION_SRC", SOURCE)):
            if re.findall(rf"^ARG {key}=(\S+)$", source, re.M) != [value]:
                raise ValueError(f"{name}: {key} drift requires fresh VEX review")
    document = json.loads((root / "build/docker/distribution.openvex.json").read_text())
    statements = document["statements"]
    if len(statements) != 3 or {s["vulnerability"]["name"] for s in statements} != IDS:
        raise ValueError("VEX must contain only the three reviewed source-fixed advisories")
    for statement in statements:
        if (statement["products"] != [{"@id": PURL}]
                or statement["status"] != "not_affected"
                or statement["justification"] != "vulnerable_code_not_present"
                or COMMIT not in statement["impact_statement"]):
            raise ValueError("VEX exact package, source or status drift requires review")


def check_source(source):
    revision = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=source, text=True).strip()
    if revision != COMMIT:
        raise ValueError(f"Distribution source changed to {revision}; VEX requires fresh review")
    for name, digest in FINGERPRINTS.items():
        if hashlib.sha256((source / name).read_bytes()).hexdigest() != digest:
            raise ValueError(f"security boundary changed: {name}; VEX requires fresh review")
    if hashlib.sha256((source / "registry/handlers/app.go").read_bytes()).hexdigest() not in APP_FINGERPRINTS:
        raise ValueError("registry app/default configuration changed; VEX requires fresh review")


def negative_controls(source):
    def rejects(operation):
        try:
            operation()
        except ValueError:
            return
        raise AssertionError("unsafe VEX/source drift unexpectedly passed")

    with tempfile.TemporaryDirectory(prefix="distribution-vex-negative-") as directory:
        temporary = Path(directory)
        for name in ("registry", "registryctl"):
            destination = temporary / f"build/docker/Dockerfile.{name}"
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(ROOT / destination.relative_to(temporary), destination)
        vex = temporary / "build/docker/distribution.openvex.json"
        original = (ROOT / vex.relative_to(temporary)).read_text()
        vex.write_text(original)
        check_configuration(temporary)
        for name in ("registry", "registryctl"):
            destination = temporary / f"build/docker/Dockerfile.{name}"
            unchanged = destination.read_text()
            destination.write_text(unchanged.replace(TAG, "v2.8.3-harbor.999"))
            rejects(lambda: check_configuration(temporary))
            destination.write_text(unchanged)
        for replacement in (PURL.replace("%2Bdirty", ""), "pkg:golang/github.com/docker/distribution"):
            vex.write_text(original.replace(PURL, replacement))
            rejects(lambda: check_configuration(temporary))
        vex.write_text(original.replace("CVE-2017-11468", "CVE-9999-0001"))
        rejects(lambda: check_configuration(temporary))
        for name in [*FINGERPRINTS, "registry/handlers/app.go"]:
            destination = temporary / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source / name, destination)
        with patch("subprocess.check_output", return_value=COMMIT + "\n"):
            check_source(temporary)
            for name in [*FINGERPRINTS, "registry/handlers/app.go"]:
                destination = temporary / name
                unchanged = destination.read_bytes()
                destination.write_bytes(unchanged + b"\n// unexpected source drift\n")
                rejects(lambda: check_source(temporary))
                destination.write_bytes(unchanged)
        with patch("subprocess.check_output", return_value="0" * 40 + "\n"):
            rejects(lambda: check_source(temporary))
    print("VEX negative controls rejected tag, PURL, advisory, source SHA and all security-file drift")


def regressions(source):
    fixtures = {
        "handlers": "registry/handlers/harbor_vex_test.go",
        "storage": "registry/storage/harbor_vex_test.go",
    }
    created = []
    try:
        for kind, target in fixtures.items():
            destination = source / target
            if destination.exists():
                raise ValueError(f"refusing to overwrite {destination}")
            shutil.copyfile(ROOT / f".ci/distribution-vex-{kind}_test.go", destination)
            created.append(destination)
        subprocess.run(["go", "test", "-mod=vendor", "-count=1", "-run",
                        "TestHarborVEX|TestValidateManifest|TestMediaTypes|TestCatalogAPI",
                        "./registry/handlers", "./registry/storage", "./manifest/ocischema",
                        "./manifest/manifestlist"], cwd=source, check=True)
    finally:
        for destination in created:
            destination.unlink(missing_ok=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-dir", type=Path)
    parser.add_argument("--fetch-source", action="store_true")
    parser.add_argument("--run-regressions", action="store_true")
    args = parser.parse_args()
    check_configuration()
    if args.fetch_source:
        if args.source_dir or args.run_regressions:
            parser.error("--fetch-source is only for immutable source verification")
        with tempfile.TemporaryDirectory(prefix="distribution-vex-") as directory:
            subprocess.run(["git", "clone", "--depth", "1", "--branch", TAG, SOURCE, directory], check=True)
            check_source(Path(directory))
    elif args.source_dir:
        check_source(args.source_dir)
        if args.run_regressions:
            negative_controls(args.source_dir)
            regressions(args.source_dir)
    elif args.run_regressions:
        parser.error("--run-regressions requires --source-dir")
    print("Distribution VEX exact configuration/source checks passed")


if __name__ == "__main__":
    main()
