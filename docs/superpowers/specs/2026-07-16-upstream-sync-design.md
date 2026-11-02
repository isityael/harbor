# Harbor upstream sync design

## Objective

Synchronise the fork with `goharbor/harbor` `upstream/main` while preserving
only the deliberate `isityael` build, dependency-hardening, and release
customisations. Keep `main` as an exact upstream tracking branch and keep
`progressed` as the fork integration branch.

The audited starting points are:

- `main`: `13532877a1d097124fb8627ce81d821ce6ce1f1d`
- `progressed`: `668a2155ba55a6af18d594804d9e26a7f5da79d8`
- `upstream/main`: `a0bacace646b02559f9aa2b189c5cafc2029d19c`
- common base for `progressed` and `upstream/main`:
  `8b8223313bc874058a0e064428f970919a2a6206`

## Branch and publication strategy

1. Fast-forward `main` to the fetched `upstream/main` commit without adding
   fork commits.
2. Preserve the old `progressed` tip with an explicit backup reference before
   rewriting history.
3. Build a temporary sync branch from the updated `main` rather than replaying
   all 224 fork-side commits and 70 merge commits mechanically.
4. Reapply the intentional current fork delta in reviewable groups.
5. Replace `progressed` only after all validation gates pass, using
   `--force-with-lease` against the audited remote tip.
6. Push Forgejo (`origin`) first, verify its Woodpecker pipeline, then push the
   identical `main`, `progressed`, and release tag state to the GitHub mirror.

This approach preserves the repository rule that the fork rebases on upstream
without retaining obsolete release bumps, merge commits, or superseded
workarounds.

## Fork surface to retain

Retain these repository-level integrations:

- `.woodpecker/lint.yaml`
- `.woodpecker/release.yaml`
- `renovate.json`
- `build/docker/Dockerfile.core`
- `build/docker/Dockerfile.db`
- `build/docker/Dockerfile.exporter`
- `build/docker/Dockerfile.jobservice`
- `build/docker/Dockerfile.nginx`
- `build/docker/Dockerfile.portal`
- `build/docker/Dockerfile.registry`
- `build/docker/Dockerfile.registryctl`
- `build/docker/Dockerfile.trivy-adapter`
- `build/docker/Dockerfile.valkey`
- `VERSION` with an `isityael` suffix appropriate for the new release
- GHCR publication, Trivy scanning, Cosign signing, and Notation signing
- DHI non-root runtime setup and explicit privilege drops
- Forgejo-first Renovate routing to `progressed`
- removal of upstream GitHub Actions workflows because Woodpecker is the
  authoritative CI and release system for the fork

Retain source-level fork changes only when current `upstream/main` does not
provide the same behaviour. This includes reviewing the Redis failure mapping,
Docker Hub proxy behaviour, redirect validation, dependency retirement, and
test-harness adaptations individually.

## Upstream changes to adopt

Adopt all current upstream functional and security work, including:

- registry proxy configuration support and replication fixes
- retention path correction and manifest upload limits
- registry ping URL restrictions
- nil-safe audit event handling
- SQL injection and timing hardening
- encrypted configuration recovery behaviour
- Angular 21, Clarity 18, and the associated portal test changes
- current Go 1.26.4 source state and upstream dependency changes
- current distribution, Trivy, and Trivy adapter component versions
- upstream multi-architecture and build corrections where they also apply to
  the DHI image implementations

## Conflict policy

The pre-implementation merge simulation identified 24 conflicts across 43
overlapping files.

### CI workflows

Keep the fork deletion for the eleven `.github/workflows/` conflicts. Review
new upstream workflows separately rather than deleting them by pattern.

### Go dependency state

Use upstream `src/go.mod` and `src/go.sum` as the base, then reapply the fork's
source-backed dependency retirement and compatible upgrades. Never resolve
these files by choosing one side wholesale. Run `go mod tidy` only after all
source imports compile, and review the resulting module graph for retracted or
renamed modules.

### Portal dashboard

Prefer the upstream Angular 21 dashboard implementation and tests. Reapply the
fork's July dashboard changes only where an upstream test or a local portal
build demonstrates that the underlying issue remains. Restore upstream Robot
Framework coverage unless it is incompatible with the final dashboard.

### Component configuration

Resolve the PostgreSQL, jobservice, registryctl, chart operator, and related
test conflicts by preserving upstream behaviour first, then adapting only the
fork dependency APIs required by retained dependency upgrades.

## Image alignment

Update fork-owned Dockerfiles because upstream changes cannot modify these new
files automatically:

- use stable `goharbor/distribution` `v2.8.3-harbor.1` instead of
  `v2.8.3-harbor.1-rc.3`
- align the Trivy binary with upstream `v0.72.0`
- align the Trivy adapter with the upstream v2.16 development line
- retain DHI builders and runtimes, numeric users, ownership, and Linux/AMD64
  output
- review every copied upstream entrypoint or configuration file changed since
  the common base

The release pipeline must use one exact release identifier consistently in
`VERSION`, the tag trigger, tag validation, image tags, and signing loops.

## Validation gates

Validation is ordered so inexpensive source failures stop the process before
image publication:

1. `git diff --check`
2. generated Swagger server code using the same command as Woodpecker
3. `go mod tidy` followed by a clean module diff review
4. targeted tests for every retained fork source change
5. the four authoritative Go builds: `core`, `jobservice`, `registryctl`, and
   `cmd/exporter`
6. portal dependency installation and production build
7. Linux/AMD64 builds for all ten custom Dockerfiles
8. runtime smoke checks for user identity, entrypoints, and required files
9. Woodpecker lint pipeline on the rewritten `progressed`
10. release-tag pipeline, image availability, vulnerability scan results, and
    signatures

Failures must be fixed on the temporary sync branch. `progressed` is not
rewritten and no release tag is created until the relevant local gates pass.

## Recovery and safety

- Preserve the user's untracked `/Users/yaelmeya/git/m0sh1.cc/harbor/AGENTS.md`.
- Do not include unrelated local files in commits.
- Record the old `origin/progressed` object ID before force-pushing.
- Use `--force-with-lease`, never an unconditional force push.
- If Forgejo, Woodpecker, GHCR, or DHI is unavailable, stop publication at the
  last verified local state and report the external blocker.
- Do not alter the infra or harbor-helm consumers until the new image tag has
  built, scanned, and been signed successfully.

## Completion criteria

The sync is complete when:

- Forgejo and GitHub `main` equal the audited `upstream/main` commit
- Forgejo and GitHub `progressed` point to the same validated rewritten tip
- the fork delta contains only reviewed intentional customisations
- all required local validation gates pass
- Woodpecker validates `progressed`
- the chosen release tag builds all ten GHCR images and completes signing
- the previous `progressed` tip remains recoverable from the backup reference
