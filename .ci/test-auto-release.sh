#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
workflow="${root}/.forgejo/workflows/release-tag.yaml"
release="${root}/.woodpecker/release.yaml"

test -f "${workflow}"
grep -Fq 'branch' "${workflow}"
grep -Fq 'progressed' "${workflow}"
grep -Fq 'build/**' "${workflow}"
grep -Fq '${{ github.server_url }}/api/v1' "${workflow}"
grep -Fq -- '-yael' "${workflow}"
if grep -Fq 'test "$(cat VERSION)" = "$CI_COMMIT_TAG"' "${release}"; then
  echo 'release pipeline still requires a manually bumped VERSION file' >&2
  exit 1
fi
printf 'Harbor automatic release contract passed\n'
