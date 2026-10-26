---
title: Syncing upstream deploymenttheory into the sweetgreen fork
tags: [upstream-sync, release, ci, task:cde1-sync-microsoft36, task:2a0f-investigate-why]
updated: 2026-10-06
---

# Syncing upstream into the fork

This page covers how to bring a `deploymenttheory/terraform-provider-microsoft365` release into
`sweetgreen/terraform-provider-microsoft365@main`.

## Current shape (since 2026-10-06)

Fork `main` is an upstream release tag with a short stack of fork-owned commits on top. The
2026-10-06 sync rebased the fork commits onto upstream `v1.3.0` (`93ff739b`) and force-pushed
`main`. The previous merge-based `main` is kept as branch `sg/main-pre-v1.3.0-rebase` (`5244c187`).

Upstream SHAs are never rewritten. Only the fork's own commits are replayed. That keeps every
upstream commit reachable, so the next sync is still incremental. The older "never rebase" rule was
about rewriting upstream history, and that still applies.

## Procedure (rebase onto an upstream release tag)
1. Set up remotes: `origin` = sweetgreen, `upstream` = deploymenttheory. Upstream tags clash with
   fork tag names (both repos have `v0.44.0-alpha`, for example). Fetch the target tag into its own
   namespace: `git fetch upstream refs/tags/vX.Y.Z:refs/upstream-tags/vX.Y.Z`.
2. List the fork-owned commits:
   `git log --no-merges refs/upstream-tags/vX.Y.Z..origin/main`.
3. Create `sg/rebase-upstream-vX.Y.Z` from the tag and cherry-pick the fork commits with
   `-x`. Use the conflict rules below.
4. Set the committer to a `@sweetgreen.com` address (`GIT_COMMITTER_EMAIL`). The org email
   ruleset rejects anything else (see below).
5. Verify the fork delta: `git diff --stat refs/upstream-tags/vX.Y.Z HEAD` should list only the
   files in the table below. If `go.mod`, `go.sum` and non-test `*.go` files are unchanged, the
   provider code matches that upstream release. Its CI result is your build evidence. A full local
   `go build ./...` takes more than 10 minutes.
6. Push the branch. Then an org admin adds a temporary **Organization admin** bypass to ruleset
   #11785376. Back up the old `main` as a branch, and run
   `git push --force-with-lease=refs/heads/main:<old-sha> origin <new-sha>:refs/heads/main`.
   Remove the bypass straight away and check it is gone with
   `gh api repos/sweetgreen/terraform-provider-microsoft365/rulesets/11785376`.

## Fork-owned files (the whole fork delta as of 2026-10-06)
| File | Fork behaviour |
|---|---|
| `.github/workflows/tf-registry-goreleaser.yml` | `ubuntu-xlarge` runner; a `use-ssm-signing-keys` path that pulls the GPG key from SSM `/terraform-registry-api-github-proxy/*`; exports `gpg-public-key.asc` |
| `.github/workflows/provider-release.yml` | `ubuntu-xlarge` pre-release job; `use-ssm-signing-keys: true` instead of GPG secrets |
| `.github/workflows/release-please.yml` | `runs-on: ubuntu-xlarge`; everything else follows upstream |
| `.github/workflows/migrate-signing-key-to-ssm.yml` | fork-only, one-time migration workflow |
| `.goreleaser.yaml` | uploads `gpg-public-key.asc` as a release asset |
| `.gitignore` | ignores `gpg-public-key.asc` |
| `.github/workflows/pr-tests.yml` | unit-test job timeout of 300 min, instead of upstream's 60 |
| 2 `list_resource_test.go` files (conditional access policy, users) + 4 `resource_acceptance_test.go` files (agent identity blueprint ×3, application) | fixes for upstream test bugs; take upstream's version on conflict once upstream fixes them |
| `wiki/` | this knowledge base |

Dropped in the 2026-10-06 rebase, and why:
- The fork 0.43.x `CHANGELOG.md` entries are obsolete on a 1.x base.
- An old `go mod tidy` commit: running `go mod tidy` on v1.3.0 gives no diff.
- A whitespace-only formatting commit whose files upstream had rewritten or deleted.
- The disk-diagnostics step, which fork `main` had already reverted to `free-disk-space`.
- `6516de95` and `aacece9a` (the CI baseline step, batched `go test`, and runner switching):
  - Upstream #4037 (v1.3.0) rewrote `go-lint.yml`, `pr-tests.yml` and `scripts/pipeline/pr/*`. It
    lints changed packages only, fails on findings, skips `TestAcc*`, and fails on test failures.
  - The baseline step only made sense for merge-based sync PRs, and syncs are now rebases.

## Conflict rules
- **Runner lines** (`runs-on`): keep the fork's `ubuntu-xlarge` on the release path only.
  Upstream-only workflows keep upstream's runners.
- **release-please**: take upstream wholesale except `runs-on`. Upstream uses a GitHub App
  token gated on `vars.RP_APP_ID`, falling back to `secrets.PAT_TOKEN` with a visible warning. The
  fork uses the PAT until that variable is set.
- **CI scripts and lint/test workflows**: take upstream's.
- **go.sum**: take upstream, then confirm with `go mod tidy` that there's no diff.

## Gotchas
- The vibe-kanban harness makes `WIP: run interrupted…` commits that sweep in untracked pipeline
  files (`.specify/`, `specs/`, `SPEC.md`…). Check `git log` after any restart and
  `git reset HEAD~1` those commits before pushing.
- The agent sandbox's git identity is `codex@openai.com`, which the email ruleset rejects. Always
  set `GIT_COMMITTER_NAME`/`GIT_COMMITTER_EMAIL` when cherry-picking.
- Known issues in the fork's SSM signing step, flagged by Codex review during the 2026-09 sync and
  still not fixed:
  - `::add-mask::` on the multi-line armored key only masks the first line.
  - `PRESET_PASSPHRASE` needs `allow-preset-passphrase` in `gpg-agent.conf`, which the SSM path
    never configures.

## CI
- **Dependency Review** flags whatever vulnerable dependencies upstream pins. Fix them with a
  separate bump.
- **Memory (blocker for lint).** golangci-lint on provider packages needs more than 16 GiB, because
  it computes analysis facts for the msgraph beta SDK.
  - Hosted `ubuntu-24.04-arm` (16 GiB) gets killed (exit 143).
  - The self-hosted `ubuntu-xlarge` pool is no bigger: on 2026-10-01 analysis was OOM-killed there
    (exit 137).
  - Lint needs a runner with at least 32 GiB.
  - Unit tests fit on 16 GiB only when packages build one at a time (`-p 1`), which is what
    upstream's scripts do.
- **Upstream test failures** fixed in the fork during the 2026-09 sync and carried into v1.3.0:
  - **List-resource permission tests.** These went stale after upstream's Graph-permissions bot
    rewrote `ReadPermissions` (#2754). The fork's fix aligns the tests with the constructors.
  - **Acceptance tests with no `PreCheck`.** `resource.Test` runs whenever `TF_ACC` is non-empty,
    **including `"0"`**. Every `TestAcc*` needs `PreCheck: func() { mocks.TestAccPreCheck(t) }`.
    CI also passes `-skip=^TestAcc`.
  - On each sync, recheck both classes. Use a `go/ast` scan for `TestAcc*` functions whose
    `resource.Test` has no `PreCheck`, and compare each test's `expectedPermissions` with its
    constructor.
- **Merging needs a second person.** Ruleset #6835010 sets
  `require_extra_approval_for_unattributed_changes`, so an agent-authored PR needs another human's
  approval before it can merge.

## Org commit-email ruleset
Org ruleset #11785376 applies to all branches. It requires author and committer emails to match
`@sweetgreen.com`, `noreply@github.com` or `[bot]@users.noreply.github.com`. Upstream authors
use personal GitHub noreply addresses (e.g. `62835948+ShocOne@users.noreply.github.com`).

- Pushing a branch that only *adds* fork commits on top of an upstream tag passes. GitHub only
  flagged the fork's own 7 commits, apparently because the upstream objects are already in the
  fork network.
- Moving `main` across new upstream commits is rejected (GH013): 282 author and 124 committer
  violations for v1.3.0.
- The `merge-upstream` (Sync fork) API is rejected too (422).
- The agent's fine-grained PAT has no `organization_administration` permission, so a human org
  admin has to add the temporary bypass in the UI.

Re-authoring upstream commits would misattribute the work, so don't do it.

## History
- PR #69 (2026-03): v0.43 → v0.49.1-alpha, 180 commits, merge.
- 2026-09-28 (task cde1-sync-microsoft36): v0.49.1-alpha → v1.2.0, 368 commits, merge (PR #86,
  never merged; closed 2026-10-06 as superseded).
- 2026-10-06 (task 2a0f-investigate-why): rebased onto v1.3.0, so `main` = v1.3.0 + 7 fork
  commits (`dc752824`), force-pushed under a temporary org-admin bypass. The test fixes and this wiki
  were carried over from PR #86's branch in a follow-up PR.
