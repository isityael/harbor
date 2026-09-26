#!/usr/bin/env bash
# harbor-db ships two PostgreSQL majors: the old one only for pg_upgrade
# (copied from /opt/postgresql/<old>) and the runtime one. The entrypoint's
# "<old>" "<new>" arguments must match the majors of those two images, or
# the build fails (missing /opt/postgresql/<old>) or upgrades break.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail=0

major_of() { sed -nE 's#.*dhi\.io/postgres:([0-9]+)\..*#\1#p' <<<"$1"; }

check() {
  local file="$1" old_ref="$2" new_ref="$3"
  local entry old new
  entry="$(grep -E '^ENTRYPOINT' "$4")"
  old="$(sed -nE 's#.*"([0-9]+)", "([0-9]+)"\].*#\1#p' <<<"${entry}")"
  new="$(sed -nE 's#.*"([0-9]+)", "([0-9]+)"\].*#\2#p' <<<"${entry}")"
  if [[ "$(major_of "${old_ref}")" != "${old}" ]]; then
    echo "${file}: upgrade-source image is PostgreSQL $(major_of "${old_ref}"), entrypoint expects ${old}" >&2
    fail=1
  fi
  if [[ "$(major_of "${new_ref}")" != "${new}" ]]; then
    echo "${file}: runtime image is PostgreSQL $(major_of "${new_ref}"), entrypoint expects ${new}" >&2
    fail=1
  fi
  if ! grep -Fq "/opt/postgresql/${old} " "${file}"; then
    echo "${file}: does not copy /opt/postgresql/${old} from the upgrade-source image" >&2
    fail=1
  fi
}

release="${root}/build/docker/Dockerfile.db"
check "${release}" \
  "$(grep -E '^ARG POSTGRES_15_IMAGE=' "${release}")" \
  "$(grep -E '^ARG RUNTIME_IMAGE=' "${release}")" \
  "${release}"

base="${root}/make/photon/db/Dockerfile.base"
check "${base}" \
  "$(grep -E '^FROM dhi\.io/postgres:.* AS pg15' "${base}")" \
  "$(grep -E '^FROM dhi\.io/postgres:' "${base}" | grep -v ' AS pg15')" \
  "${root}/make/photon/db/Dockerfile"

grep -Fq '"matchPackageNames": ["dhi.io/postgres"]' "${root}/renovate.json" || {
  echo "renovate.json must hold dhi.io/postgres to its current major" >&2
  fail=1
}

exit "${fail}"
