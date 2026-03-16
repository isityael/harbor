# CLAUDE — Runtime Instructions

This is a **fork of [goharbor/harbor](https://github.com/goharbor/harbor)** maintained by sm-moshi.

## Purpose

Build custom Harbor Go component images with cherry-picked high-priority fixes
(Redis stability, OIDC, GC performance, security) that upstream hasn't released yet.
Images are pushed to `harbor.m0sh1.cc/apps/` and deployed via the infra repo's
harbor-helm wrapper chart.

## Key rules

- **Rebase on upstream, don't diverge.** All changes should be cherry-picks from upstream PRs
  or minimal fixes that can be upstreamed. Avoid custom features that make rebasing harder.
- **`main` stays synced with upstream** — never commit directly to `main`.
  `sm-moshi/v2.15` is the patched branch for the v2.15 series.
- **Version scheme:** `VERSION` file is set to `v2.15.0-rc2-smN` (e.g. `-sm1`, `-sm2`).
  Tags follow the same pattern. Bump the `-smN` suffix for each batch of cherry-picks.
- **No secrets or credentials** in this repo.
- **Only build Go components** that are affected by cherry-picked fixes. Components without
  relevant fixes should use upstream `goharbor/*` images.

## Components built

| Component | Binary path | Image |
|-----------|------------|-------|
| core | `src/core` | `harbor.m0sh1.cc/apps/harbor-core` |
| jobservice | `src/jobservice` | `harbor.m0sh1.cc/apps/harbor-jobservice` |
| registryctl | `src/registryctl` | `harbor.m0sh1.cc/apps/harbor-registryctl` |
| exporter | `src/cmd/exporter` | `harbor.m0sh1.cc/apps/harbor-exporter` |

Components NOT built (use upstream images): portal, registry (distribution), trivy, db, redis.

## Workflow

### Adding a new upstream PR fix
```bash
git checkout sm-moshi/v2.15
git fetch upstream pull/<PR>/head:pr-<PR>
git cherry-pick <commit-hash>
# Resolve conflicts, then test:
cd src && go vet ./... && go build ./core ./jobservice ./registryctl ./cmd/exporter
```

### Rebasing on a new upstream release
```bash
git fetch upstream --tags
# Create new branch from new tag:
git checkout -b sm-moshi/v2.15-ga v2.15.0
# Re-cherry-pick only PRs not yet merged upstream
# Update VERSION, tag, push
```

### Releasing
```bash
# Update VERSION to v2.15.0-rc2-smN
git add -A && git commit -m "build: prepare v2.15.0-rc2-smN"
git tag v2.15.0-rc2-smN
git push origin sm-moshi/v2.15 --tags
# Woodpecker release pipeline builds and pushes 4 images
```

## Validation
```bash
cd src && go vet ./...
cd src && staticcheck $(go list ./... | grep -v '\.pb\.go')
# Local Docker build test:
docker build -f .woodpecker/docker/Dockerfile.core -t harbor-core-test .
```

## Cherry-picked PRs (v2.15.0-rc2-sm1)

| PR | Title | Status |
|----|-------|--------|
| #22637 | Fix broken migration in 0180_2.15.0 schema | Open |
| #22679 | Fix Redis errors treated as rate-limit 429 | Open |
| #22572 | Distributed lock for execution status refresh | Open |
| #22556 | Disallow empty robot_name_prefix (OIDC fix) | Open |
| #22871 | Validate redirect URIs (open redirect fix) | Open |
| #22994 | Fix full table scan in blob GC (4min to 40ms) | Open |

Skipped:
- #22879 (audit log redaction) — not applicable to v2.15.0-rc2 base; superseded by #22913

## CI/CD

- **Woodpecker CI** via GitHub forge (id=2)
- **Lint pipeline** (`.woodpecker/lint.yaml`): `go vet` + `staticcheck` on push/PR
- **Release pipeline** (`.woodpecker/release.yaml`): Build 4 images, trivy scan,
  cosign + notation signing. Triggered by `v*-sm*` tags.
- **Remote builder:** Builds are offloaded to the Mac via SSH (`remote_builders` in the
  buildx plugin). Dockerfiles use `--platform=$BUILDPLATFORM` for native-speed Go
  cross-compilation — only the Photon runtime stage uses QEMU.
- **Secrets used:** `harbor_username`, `harbor_password`, `mac_builder_ssh_key`,
  `cosign_key`, `cosign_password`, `notation_key`, `notation_cert` (org-level on sm-moshi)

## Consumer

The infra repo (`sm-moshi/infra`) overrides image repos/tags in
`apps/user/harbor/values.yaml` to point to our fork images.
The harbor-helm chart (`sm-moshi/harbor-helm`) is a separate fork.

## Gotchas

- **registryctl needs the distribution binary** (`registry_DO_NOT_USE_GC`). The Dockerfile
  builds it from `goharbor/distribution` at `v2.8.3-patch-redis`.
- **Circular dependency:** Harbor images are pushed to Harbor. If Harbor is down during
  an upgrade, fall back to building locally and pushing via `docker push`.
- **Go 1.25.7** — pinned to match upstream `go.mod`.
- **Photon 5.0** base image — matches upstream. Needs `tdnf` for package installs.
- **Cross-compilation:** Builder stage uses `--platform=$BUILDPLATFORM` with
  `GOOS=${TARGETOS} GOARCH=${TARGETARCH}`. Go compiles natively on the builder
  (ARM64 Mac), producing AMD64 binaries. The Photon runtime stage is the only
  part that runs under QEMU.
- **Remote builder requires SSH:** The Mac needs `Remote Login` enabled in System Settings
  and the `mac_builder_ssh_key` secret must contain the corresponding SSH private key.
  Mac must be reachable from k8s pods on VLAN 20 → VLAN 10 port 22.

## Style

- Use British English in all prose.
