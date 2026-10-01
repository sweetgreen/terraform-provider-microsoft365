---
title: Syncing upstream deploymenttheory into the sweetgreen fork
tags: [upstream-sync, release, ci, task:cde1-sync-microsoft36]
updated: 2026-09-28
---

# Syncing upstream into the fork

This page covers how to bring `deploymenttheory/terraform-provider-microsoft365@main` into `sweetgreen/terraform-provider-microsoft365@main`.

## Procedure
1. Set up remotes: `origin` = sweetgreen, `upstream` = deploymenttheory. Run `git fetch upstream origin`.
2. Measure the drift: `git rev-list --count origin/main..upstream/main` shows how far behind the fork is, and `upstream/main..origin/main` shows the fork-only commits.
3. Cut a branch from `origin/main` and run `git merge --no-ff upstream/main`. **Never rebase, squash, or cherry-pick.** Upstream SHAs must stay reachable so the next sync is incremental.
4. Resolve conflicts using the rules below.
5. Verify the fork delta: `git diff --stat upstream/main HEAD` should list **only** the fork-owned files in the table below. If it does, and `go.mod`/`go.sum` and non-test `*.go` files are unchanged, the provider code is identical to that upstream release and builds the same way. There is no need for a full local `go build ./...`, which takes more than 10 minutes on the msgraph beta SDK. Let PR CI build it.
6. Open the PR and **merge with a merge commit** (the repo allows merge commits).

## Fork-owned files (the entire fork delta as of 2026-10-01)
| File | Fork behaviour |
|---|---|
| `.github/workflows/tf-registry-goreleaser.yml` | `ubuntu-xlarge` runner; `use-ssm-signing-keys` path that pulls the GPG key from SSM `/terraform-registry-api-github-proxy/*`; exports `gpg-public-key.asc` |
| `.github/workflows/provider-release.yml` | `ubuntu-xlarge` pre-release job; `use-ssm-signing-keys: true` in place of GPG secrets |
| `.github/workflows/release-please.yml` | `runs-on: ubuntu-xlarge` (everything else follows upstream) |
| `.github/workflows/migrate-signing-key-to-ssm.yml` | fork-only, one-time migration workflow |
| `.goreleaser.yaml` | uploads `gpg-public-key.asc` as a release asset |
| `.gitignore` | ignores `gpg-public-key.asc` |
| `CHANGELOG.md` | fork 0.43.1–0.43.3-alpha entries on top |
| `.github/workflows/go-lint.yml`, `.github/workflows/pr-tests.yml` | A **Resolve comparison baseline** step compares a sync PR against the newest upstream commit it contains, so the checks cover fork-owned code (the `ci:full-scope` label overrides this). Lint uses `--new-from-rev=<baseline>` on the changed packages instead of `only-new-issues`. Job timeouts are 90 min (lint) and 300 min (tests). |
| `scripts/pipeline/pr/lib/{git_operations,go_tests}.py`, `steps/detect_changes.py` | skip deleted package dirs; batch `go test` (25 packages, `-p 1`, `-ldflags=-s -w`, `-skip=^TestAcc`); fail the job on test failures; emit `packages-args` for lint. **Upstream #4037 (v1.3.0) rewrote these files.** On the next sync, take upstream's scripts and workflows, then re-apply only the runner expression and the baseline step. |
| 2 `list_resource_test.go` files (conditional access policy, users) + 4 `resource_acceptance_test.go` files (agent identity blueprint ×3, application) | test fixes for upstream bugs. If upstream fixes them too, take upstream's version on conflict |

## Conflict rules
- **Runner lines** (`runs-on`): keep the fork's `ubuntu-xlarge` on the release path only. Upstream-only workflows keep upstream's runners.
- **release-please**: take upstream wholesale except `runs-on`. Since 2026-09, upstream uses a GitHub App token gated on `vars.RP_APP_ID`, falling back to `secrets.PAT_TOKEN` with a visible warning. The fork uses the PAT until that variable is set.
- **CHANGELOG.md**: take the fork header plus the fork 0.43.x blocks, then upstream's file from its first `## [` heading onward.
- **go.sum**: take upstream, then confirm with `go mod tidy` that there's no diff.
- The fork keeps its own 0.43.x-alpha release line and does not adopt upstream version numbers.

## Gotchas
- The vibe-kanban harness makes `WIP: run interrupted…` commits that sweep in untracked pipeline files (`.specify/`, `specs/`, `SPEC.md`…). Check `git log` after any restart and `git reset HEAD~1` those commits before pushing.
- Known pre-existing issues in the fork's SSM signing step, flagged by Codex review during the 2026-09 sync and not fixed by it:
  - `::add-mask::` on the multi-line armored key only masks the first line.
  - `PRESET_PASSPHRASE` needs `allow-preset-passphrase` in `gpg-agent.conf`, which the SSM path never configures.

## CI on sync PRs
- **Dependency Review** flags whatever vulnerable deps upstream pins. For example, grpc 1.79.x in 2026-09. Fix it with a separate bump, not in the merge.
- **Memory (blocker for lint).** golangci-lint on provider packages needs more than 16 GiB, because it computes analysis facts for the msgraph beta SDK. Hosted `ubuntu-24.04-arm` (16 GiB) is shut down with exit 143. The self-hosted `ubuntu-xlarge` pool, a shared ARM64 org pool of 7, was tried on 2026-10-01: package loading took 14.5 min, then analysis was OOM-killed (exit 137). That pool is no bigger, so lint needs a larger runner (≥32 GiB) before it can pass. Unit tests fit on 16 GiB only when packages build one at a time (`-p 1`). Two at a time gets the runner killed.
- **Scope.** A sync touches about 300 packages. Comparing against the merged upstream commit limits PR checks to fork-owned changes. For a full unit-test run of the merged code, add the `ci:full-scope` label to the PR.
- **Upstream test failures** fixed in the fork during the 2026-09 sync:
  - **List-resource permission tests.** Tests went stale after upstream's Graph-permissions bot rewrote `ReadPermissions` (#2754). The fix aligns the tests with the constructors.
  - **Acceptance tests with no `PreCheck`.** `resource.Test` runs whenever `TF_ACC` is non-empty, **including `"0"`**. Every `TestAcc*` needs `PreCheck: func() { mocks.TestAccPreCheck(t) }`. CI also passes `-skip=^TestAcc`.
  - On the next sync, recheck both classes. Use a `go/ast` scan for `TestAcc*` functions whose `resource.Test` has no `PreCheck`, and compare each test's `expectedPermissions` with its constructor.
- **Merging needs a second person.** Ruleset #6835010 sets `require_extra_approval_for_unattributed_changes`, so an agent-authored PR needs another human's approval before it can merge. `gh pr merge` fails with "base branch policy prohibits the merge". `gh pr merge --merge --auto` queues the merge until someone approves.

## Blocker: org commit-email ruleset (2026-10)
Org ruleset #11785376 applies to all branches. It requires author and committer emails to match `@sweetgreen.com`, `noreply@github.com` or `[bot]@users.noreply.github.com`. Upstream authors use personal GitHub noreply addresses (e.g. `62835948+ShocOne@users.noreply.github.com`), so pushing **new** upstream commits is rejected (GH013). The `merge-upstream` (Sync fork) API is rejected too (422).

The 2026-09 sync's 368 commits were not flagged, for unknown reasons. A local merge with upstream v1.3.0 (`cde1/upstream-1.3.0-merge`) could not be pushed.

**Fix:** an org admin either adds `@users\.noreply\.github\.com` to the pattern or adds a bypass for upstream syncs. Re-authoring upstream commits would break upstream fidelity (Constitution IV) and misattribute the work.

## History
- PR #69 (2026-03): v0.43 → v0.49.1-alpha, 180 commits.
- 2026-09-28 (task cde1-sync-microsoft36): v0.49.1-alpha → v1.2.0, 368 commits, 2 conflicts (release-please.yml, CHANGELOG.md).
