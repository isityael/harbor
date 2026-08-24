#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
config="${repo_root}/renovate.json"

grep -Fq '"postUpdateOptions": ["gomodTidyE"]' "${config}" || {
  echo "Renovate must use error-tolerant Go module tidying for generated Harbor API packages." >&2
  exit 1
}

expected_dirs=(
  "./lib/..."
  "./pkg/..."
  "./common/..."
  "./cmd/exporter"
)

for dir in "${expected_dirs[@]}"; do
  grep -Fq "\"${dir}\"" "${config}" || {
    echo "Renovate goGetDirs is missing ${dir}." >&2
    exit 1
  }
done

if grep -Fq '"gomodTidy"' "${config}"; then
  echo "Strict gomodTidy cannot run before Harbor's generated API packages exist." >&2
  exit 1
fi

if [[ -d "${repo_root}/src/server/v2.0/models" || -d "${repo_root}/src/server/v2.0/restapi" ]]; then
  echo "Generated Harbor API packages already exist; the clean-checkout contract cannot be tested." >&2
  exit 1
fi

(
  cd "${repo_root}/src"
  go list -deps -test "${expected_dirs[@]}" >/dev/null
)

echo "Renovate Go artifact contract passed."
