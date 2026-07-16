# Harbor Upstream Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the `progressed` fork branch on current `goharbor/harbor` `upstream/main`, retain only intentional fork behaviour, validate it, and publish identical Forgejo and GitHub branch state.

**Architecture:** Keep `main` as an exact upstream pointer. Construct a clean `codex/upstream-sync-2026-07-16` branch from `upstream/main`, squash-apply the current fork delta, resolve the known overlaps with upstream-first semantics, and split the resolved result into reviewable commits. Replace `progressed` only after local validation, using an explicit backup reference and `--force-with-lease`.

**Tech Stack:** Git, Go 1.26.4, Go modules, go-swagger 0.33.1, Angular 21, Node.js, Docker Buildx, Woodpecker CI, DHI images, GHCR, Trivy, Cosign, Notation.

## Global Constraints

- Forgejo remote `origin` is canonical; remote `github` is a push mirror.
- `main` must equal `upstream/main` and contain no fork commits.
- Fork work belongs on `progressed`; never commit fork changes directly to `main`.
- Preserve `/Users/yaelmeya/git/m0sh1.cc/harbor/AGENTS.md` and all unrelated user files.
- Use British English in prose.
- Retain DHI non-root images, GHCR publication, Woodpecker, Trivy, Cosign, and Notation.
- Use `--force-with-lease` for the `progressed` rewrite; never use an unconditional force push.
- Do not change `/Users/yaelmeya/git/m0sh1.cc/infra` or `/Users/yaelmeya/git/m0sh1.cc/harbor-helm` until the new images have built and been signed.

---

### Task 1: Preserve branch state and create the clean sync branch

**Files:**
- Read: `/Users/yaelmeya/git/m0sh1.cc/harbor/docs/superpowers/specs/2026-07-16-upstream-sync-design.md`
- Read: `/Users/yaelmeya/git/m0sh1.cc/harbor/docs/superpowers/plans/2026-07-16-upstream-sync.md`

**Interfaces:**
- Consumes: fetched `upstream/main`, local `progressed`, `origin/progressed`, and `github/progressed` refs.
- Produces: backup branch `backup/progressed-pre-upstream-sync-2026-07-16` and worktree branch `codex/upstream-sync-2026-07-16`.

- [ ] **Step 1: Refresh and record all relevant object IDs**

Run:

```bash
git fetch upstream --tags
git fetch origin --tags
git fetch github --tags
git rev-parse main progressed upstream/main origin/main origin/progressed github/main github/progressed
```

Expected: `origin/progressed` and `github/progressed` are identical; `upstream/main` resolves successfully.

- [ ] **Step 2: Create a recoverable backup reference**

Run:

```bash
git branch backup/progressed-pre-upstream-sync-2026-07-16 progressed
git show -s --format='%H %s' backup/progressed-pre-upstream-sync-2026-07-16
```

Expected: the backup points to the local `progressed` tip containing the approved design and plan.

- [ ] **Step 3: Create the isolated worktree from upstream**

Run from `/Users/yaelmeya/git/m0sh1.cc/harbor` after verifying `.worktrees` is ignored:

```bash
git worktree add /Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16 -b codex/upstream-sync-2026-07-16 upstream/main
```

Expected: the new branch tip equals `upstream/main` and the original checkout remains untouched.

- [ ] **Step 4: Establish the upstream baseline**

Run inside the worktree:

```bash
git status --short --branch
git diff --check
cd src && go mod download
```

Expected: clean status and successful module download.

### Task 2: Apply and split the repository-level fork integration

**Files:**
- Create: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/.woodpecker/lint.yaml`
- Create: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/.woodpecker/release.yaml`
- Create: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/renovate.json`
- Create: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/build/docker/Dockerfile.*`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/VERSION`
- Delete: selected upstream files under `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/.github/workflows/`

**Interfaces:**
- Consumes: net tree delta from `backup/progressed-pre-upstream-sync-2026-07-16`.
- Produces: Woodpecker/DHI/GHCR integration on the clean upstream base.

- [ ] **Step 1: Apply the net fork delta without committing**

Run:

```bash
git merge --squash backup/progressed-pre-upstream-sync-2026-07-16
```

Expected: known conflicts are reported; no merge commit is created.

- [ ] **Step 2: Resolve CI workflow policy**

Remove the eleven upstream GitHub workflows previously replaced by Woodpecker, but retain and review any new upstream workflows not present in the conflict list.

Run:

```bash
git status --short
git diff --name-only --diff-filter=U
```

Expected: the known `.github/workflows/` conflicts are resolved as deletions.

- [ ] **Step 3: Preserve documentation and build integration**

Ensure the approved files are present, then stage and commit only repository-level CI, release, Renovate, Dockerfile, and design/plan files.

Run:

```bash
git diff --check
git commit -m "build: restore fork image and release integration"
```

Expected: one focused commit without Go source or portal changes.

### Task 3: Reconcile source and Go dependency hardening

**Files:**
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/src/go.mod`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/src/go.sum`
- Modify: retained fork source files under `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/src/`

**Interfaces:**
- Consumes: upstream Go 1.26.4 source and the fork's source-backed dependency replacements.
- Produces: a coherent Go module graph and compiling retained fork behaviour.

- [ ] **Step 1: Resolve source conflicts using upstream as the behavioural base**

For `src/common/dao/pgsql.go`, `src/jobservice/config/config.go`, `src/pkg/chart/operator.go`, and `src/registryctl/config/config.go`, preserve upstream functional changes and adapt only APIs required by retained dependency replacements.

- [ ] **Step 2: Audit each retained fork behaviour**

Use file history and targeted diffs to decide whether Redis error mapping, Docker Hub proxy handling, redirect validation, dependency retirement, and test-harness changes remain absent upstream. Remove superseded patches.

- [ ] **Step 3: Reconcile modules from upstream state**

Start with upstream direct requirements, add only modules required by retained source, then run:

```bash
cd src
go mod tidy
go list -m -u all
go mod verify
```

Expected: tidy and verify pass; renamed or retracted modules are explicitly reviewed rather than blindly upgraded.

- [ ] **Step 4: Generate Swagger and run focused tests**

Run the Woodpecker-equivalent Swagger command, then targeted tests for every retained source package.

Expected: generated code succeeds and all targeted tests pass or an environment-only dependency is identified with exact evidence.

- [ ] **Step 5: Build the four authoritative Go components**

Run:

```bash
cd src
CGO_ENABLED=0 go build -o /dev/null ./core
CGO_ENABLED=0 go build -o /dev/null ./jobservice
CGO_ENABLED=0 go build -o /dev/null ./registryctl
CGO_ENABLED=0 go build -o /dev/null ./cmd/exporter
```

Expected: all four commands exit zero.

- [ ] **Step 6: Commit source reconciliation**

Run `git diff --check`, review the staged diff, and commit with:

```bash
git commit -m "fix: reconcile fork hardening with upstream"
```

### Task 4: Adopt the upstream portal implementation

**Files:**
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/src/portal/package-lock.json`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/src/portal/app-swagger-ui/package-lock.json`
- Modify only if demonstrated necessary: job-service dashboard files under `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/src/portal/src/app/base/left-side-nav/job-service-dashboard/`
- Restore or deliberately adapt: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/tests/resources/Harbor-Pages/Job_Service_Dashboard.robot`

**Interfaces:**
- Consumes: upstream Angular 21/Clarity 18 dashboard and tests.
- Produces: portal build and test state without obsolete legacy-template workarounds.

- [ ] **Step 1: Prefer upstream portal conflict resolutions**

Accept upstream dashboard templates and lockfiles first. Reapply a fork dashboard change only after reproducing its original failure on the upstream implementation.

- [ ] **Step 2: Install and build the portal**

Run the repository-supported clean dependency installation and production build from `src/portal`.

Expected: dependency lockfiles remain deterministic and the production build exits zero.

- [ ] **Step 3: Run focused dashboard tests and commit**

Run the relevant Angular dashboard unit tests. Commit the resolved portal state with:

```bash
git commit -m "fix: align portal customisations with Angular 21"
```

### Task 5: Align custom image sources and release metadata

**Files:**
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/build/docker/Dockerfile.registry`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/build/docker/Dockerfile.registryctl`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/build/docker/Dockerfile.trivy-adapter`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/.woodpecker/release.yaml`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/VERSION`

**Interfaces:**
- Consumes: upstream distribution `v2.8.3-harbor.1`, Trivy `v0.72.0`, and current adapter version.
- Produces: one consistent new `isityael` release identifier and ten aligned image builds.

- [ ] **Step 1: Update component sources**

Set both distribution Dockerfiles to `v2.8.3-harbor.1`, set Trivy to `v0.72.0`, and derive the adapter version from current upstream component metadata.

- [ ] **Step 2: Select one release identifier**

Use the next unique `v2.16.0-yael.N` identifier after checking local, Forgejo, GitHub, and GHCR tag state. Apply it consistently to `VERSION` and every tag trigger/validation expression.

- [ ] **Step 3: Validate pipeline consistency**

Run exact searches for old component and release versions. Expected: no stale release trigger, distribution RC, or Trivy 0.69.3 remains in active build configuration.

- [ ] **Step 4: Commit image alignment**

Run `git diff --check` and commit with:

```bash
git commit -m "build: align custom images with current upstream"
```

### Task 6: Validate all custom images

**Files:**
- Test: all ten files under `/Users/yaelmeya/git/m0sh1.cc/harbor/.worktrees/upstream-sync-2026-07-16/build/docker/`

**Interfaces:**
- Consumes: reconciled source, portal, entrypoints, and image metadata.
- Produces: locally validated Linux/AMD64 image set.

- [ ] **Step 1: Build Go-backed images first**

Use the available Linux/AMD64 Buildx builder to build core, jobservice, registryctl, and exporter without pushing.

Expected: all four images build successfully.

- [ ] **Step 2: Build the remaining six images**

Build nginx, portal, registry, database, Trivy adapter, and Valkey without pushing.

Expected: all six images build successfully.

- [ ] **Step 3: Run runtime smoke checks**

Inspect configured users, entrypoints, file ownership, executable bits, and component version output for each image.

Expected: runtime users and required files match the DHI non-root design.

- [ ] **Step 4: Run final repository checks**

Run:

```bash
git diff --check
git status --short
git log --oneline --decorate upstream/main..HEAD
```

Expected: no unstaged build artefacts and only intentional fork commits above upstream.

### Task 7: Publish safely and verify CI and release

**Files:**
- No new source files; operates on validated Git refs and CI/release state.

**Interfaces:**
- Consumes: validated `codex/upstream-sync-2026-07-16` tip and recorded old remote object IDs.
- Produces: synchronised `main`, rewritten `progressed`, release tag, ten signed GHCR images, and matching Forgejo/GitHub refs.

- [ ] **Step 1: Fast-forward and publish `main`**

Update local `main` to exact `upstream/main`, then push `main` to `origin` and `github`. Verify both remote object IDs equal upstream.

- [ ] **Step 2: Replace and publish `progressed` to Forgejo**

Move local `progressed` to the validated sync tip and push with an explicit lease against the recorded old `origin/progressed` object ID.

Expected: Forgejo accepts the rewrite without overriding an unexpected concurrent update.

- [ ] **Step 3: Verify the Woodpecker branch pipeline**

Inspect the pipeline for the exact new `progressed` commit and wait for both Swagger generation and four-component build-check to complete.

- [ ] **Step 4: Mirror `progressed` to GitHub**

Push with an explicit lease against the recorded old `github/progressed` object ID, then verify both remotes point to the same commit.

- [ ] **Step 5: Create and publish the release tag**

Create the selected signed release tag only after the branch pipeline passes. Push it to Forgejo, verify the Woodpecker release pipeline, then mirror the identical tag to GitHub.

- [ ] **Step 6: Verify release outputs**

Confirm all ten GHCR tags exist, inspect Trivy results, and verify both Cosign and Notation signatures. Do not update downstream consumers unless all required outputs are present.

- [ ] **Step 7: Report exact completion state**

Report final `main`, `progressed`, backup, and tag object IDs; local validation commands; Woodpecker pipeline identifiers; GHCR image list; and any non-blocking HIGH vulnerability findings.
