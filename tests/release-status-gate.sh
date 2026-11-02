#!/bin/sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
gate="$repo_root/hack/wait-for-woodpecker-success.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$tmp/bin"
cat >"$tmp/bin/curl" <<'EOF'
#!/bin/sh
cat "$STATUS_FIXTURE"
EOF
chmod +x "$tmp/bin/curl"

run_gate() {
  PATH="$tmp/bin:$PATH" \
    FORGEJO_API_URL=https://forgejo.invalid/api/v1 \
    FORGEJO_REPOSITORY=isityael/example \
    FORGEJO_TOKEN=secret \
    WOODPECKER_GATE_MAX_ATTEMPTS=1 \
    WOODPECKER_GATE_SLEEP_SECONDS=0 \
    STATUS_FIXTURE="$1" \
    "$gate" deadbeef ci/woodpecker/push/lint
}

cat >"$tmp/success.json" <<'EOF'
[{"context":"ci/woodpecker/push/lint","status":"success"}]
EOF
run_gate "$tmp/success.json"

cat >"$tmp/pending.json" <<'EOF'
[{"context":"ci/woodpecker/push/lint","status":"pending"}]
EOF
if run_gate "$tmp/pending.json"; then
  echo "pending status unexpectedly passed" >&2
  exit 1
fi

cat >"$tmp/failure.json" <<'EOF'
[{"context":"ci/woodpecker/push/lint","status":"failure"}]
EOF
if run_gate "$tmp/failure.json"; then
  echo "failure status unexpectedly passed" >&2
  exit 1
fi

cat >"$tmp/missing.json" <<'EOF'
[]
EOF
if run_gate "$tmp/missing.json"; then
  echo "missing status unexpectedly passed" >&2
  exit 1
fi

echo "release status gate tests passed"
