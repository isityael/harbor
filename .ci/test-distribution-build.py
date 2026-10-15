#!/usr/bin/env python3
"""Exercise the release registry source, ACL reader and real optional-driver build."""

import argparse
from contextlib import nullcontext
import os
from pathlib import Path
import platform
import re
import shutil
import subprocess
import tempfile


ROOT = Path(__file__).resolve().parent.parent


def run(*args: str, cwd: Path, env=None, capture=False):
    print("+", " ".join(args), flush=True)
    return subprocess.run(args, cwd=cwd, env=env, check=True, text=True,
                          stdout=subprocess.PIPE if capture else None)


def verify(source: Path) -> None:
    run("python3", str(ROOT / ".ci/test-distribution-vex.py"),
        "--source-dir", str(source), cwd=ROOT)
    run("sh", str(ROOT / "build/docker/prepare-distribution.sh"), cwd=source)
    acl_test = source / "vendor/google.golang.org/cloud/storage/harbor_acl_compat_test.go"
    shutil.copyfile(ROOT / ".ci/distribution-gcs-acl_test.go", acl_test)
    dns_test = source / "vendor/github.com/xenolf/lego/acme/harbor_dns_compat_test.go"
    shutil.copyfile(ROOT / ".ci/distribution-dns_test.go", dns_test)
    run("go", "test", "-mod=vendor", "-count=1", "-v",
        "./vendor/google.golang.org/cloud/storage", "./vendor/github.com/xenolf/lego/acme", cwd=source)
    # Exercise the unchanged registry-specific uploadpurging patch from its
    # actual Dockerfile, rather than maintaining a second copy in this test.
    dockerfile = (ROOT / "build/docker/Dockerfile.registry").read_text()
    upload_patch = re.search(r"RUN python3 - <<'PY'\n(.*?)\nPY", dockerfile, re.S)
    if not upload_patch:
        raise RuntimeError("registry uploadpurging build patch was not found")
    run("python3", "-c", upload_patch.group(1), cwd=source)
    run("python3", str(ROOT / ".ci/test-distribution-vex.py"),
        "--source-dir", str(source), "--run-regressions", cwd=ROOT)
    build_env = os.environ | {"CGO_ENABLED": "0", "GOOS": "linux", "GOARCH": "amd64",
                              "BUILDTAGS": "include_oss include_gcs"}
    run("make", "PREFIX=/go", "clean", "bin/registry", cwd=source, env=build_env)
    metadata = run("go", "version", "-m", "bin/registry", cwd=source, capture=True).stdout
    for setting in ("GOOS=linux", "GOARCH=amd64", "-tags=include_oss,include_gcs"):
        if setting not in metadata:
            raise RuntimeError(f"registry build is missing required setting: {setting}")
    if platform.system() == "Linux" and platform.machine() in ("x86_64", "amd64"):
        run(str(source / "bin/registry"), "--version", cwd=source)
    print("Distribution GCS ACL regressions and Linux/amd64 GCS+OSS registry build passed")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-dir", type=Path, help="reuse an isolated distribution checkout")
    args = parser.parse_args()
    context = nullcontext(args.source_dir.resolve()) if args.source_dir else tempfile.TemporaryDirectory(prefix="harbor-distribution-test-")
    with context as directory:
        source = Path(directory)
        if not args.source_dir:
            dockerfile = (ROOT / "build/docker/Dockerfile.registry").read_text()
            tag = re.search(r"^ARG DISTRIBUTION_VERSION=(\S+)$", dockerfile, re.M).group(1)
            url = re.search(r"^ARG DISTRIBUTION_SRC=(\S+)$", dockerfile, re.M).group(1)
            other = (ROOT / "build/docker/Dockerfile.registryctl").read_text()
            if f"ARG DISTRIBUTION_VERSION={tag}\n" not in other or f"ARG DISTRIBUTION_SRC={url}\n" not in other:
                raise RuntimeError("registry and registryctl must use the same distribution source")
            run("git", "clone", "--depth", "1", "--branch", tag, url, str(source), cwd=ROOT)
        verify(source)


if __name__ == "__main__":
    main()
