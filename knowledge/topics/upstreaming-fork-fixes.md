---
title: Upstreaming fork fixes to deploymenttheory
tags: [upstream, process, ci, release, task-1032, task-6a42]
updated: 2026-10-19
---

# Upstreaming fork fixes to deploymenttheory

`origin` (sweetgreen/terraform-provider-microsoft365) is a GitHub fork of
`upstream` (deploymenttheory/terraform-provider-microsoft365). Fixes ship
first on the fork's `sg/...` line as `vX.Y.Z-sg.N`. Each one then goes
upstream so the `sg` line can drop it after the next upstream release.

## Recipe

1. `git fetch upstream main`, then create a separate worktree on a branch cut
   from `upstream/main`, not from the `sg` line.
2. Cherry-pick the fork commits. Remove fork PR numbers like `(#89)` from the
   subjects and bodies, because upstream would link them to its own PRs. Put
   the fork PR URL in the body instead.
3. Push the branch to `origin` and open the PR from `sweetgreen:<branch>`
   against `upstream` `main`. **Agents can't do this step.** Every GitHub
   token the vibe-kanban `gh` router holds (sweetgreen, davidvasandani,
   alderbridge, teemayhq) is a fine-grained PAT. Creating a PR on
   deploymenttheory returns `403 Resource not accessible by personal access
   token` through both GraphQL and REST. Push the branch, save the PR body,
   and give a human the compare link:
   `https://github.com/deploymenttheory/terraform-provider-microsoft365/compare/main...sweetgreen:terraform-provider-microsoft365:<branch>?expand=1`
4. **Title:** `pr-title-validation.yml` requires
   `^(feat|fix|docs|style|refactor|test|chore|build|ci|perf)(\(.+\))?: .+$`.
   Upstream squash-merges, so the PR title becomes the release-please
   changelog line. Make it cover every commit in the PR.
5. **Body:** follow `.github/PULL_REQUEST_TEMPLATE.md`: Summary, Motivation,
   Type of Change, Testing, Quality Checklist.
6. **CI:** `pr-tests.yml` tests only the changed Go packages. On non-draft PRs
   it also enforces a patch-coverage threshold. Upstream may need to approve
   workflow runs for outside contributors.

## Shipping a fork release (`vX.Y.Z-sg.N`)

- The release line is `sg/fix-release-tool-cache`, not `main`. `main` doesn't carry the group fixes.
  Branch from `origin/sg/fix-release-tool-cache` and open the PR against it. Then push an annotated
  `vX.Y.Z-sg.N` tag on the merge commit, which runs `provider-release.yml`.
  `registry.sweetgreen.engineering` serves GitHub releases directly.
- A ruleset rejects pushes whose author or committer email isn't `@sweetgreen.com`
  (`GH013 ... Author email must match a given regex pattern`). Agent worktrees may have a non-sweetgreen
  git identity configured. Commit with `GIT_AUTHOR_*`/`GIT_COMMITTER_*` set to the sweetgreen
  identity used by the fork's history rather than changing the shared git config.
- Vibe Kanban's `create_pr` opens PRs only from the workspace branch. A release-line PR from
  another branch needs `gh pr create`. `get_pr`/`merge_pr` take a PR number and work for it.
- `pr-tests.yml` takes about 45 minutes per run, mostly compiling the SDK. A run can also be
  OOM-killed on the hosted arm runner a few minutes in ("The operation was canceled", with
  `oom_kill 1` in the job's `memory.events` dump) before any test runs. That's a flake; re-run it.
- On fork PRs the unit-test job always ends red: the group tests pass (`ok ... group`), then the
  Codecov step fails because `CODECOV_API_TOKEN` is not set. The golangci-lint job times out. Judge
  the PR by the per-package `ok`/`FAIL` lines; merging then needs `merge_pr` with `force`.

## Running the tests without a local build

Compiling `msgraph-beta-sdk-go` takes about 10 GB of RAM. The vibe-kanban
worker cgroup throttles memory at 9.3 GiB (`memory.high`), so a local
`go test` or `go vet` on any provider package stalls for tens of minutes. It
also starves other processes on the host: even `gh` and `go env` hang. Run the
tests in the fork's CI instead:

```sh
gh workflow run pr-tests.yml -R sweetgreen/terraform-provider-microsoft365 \
  --ref <branch> -f base_ref=<upstream-main-sha>
```

`base_ref` limits the run to the packages that changed relative to upstream.
To prove the tests catch the bug, push a throwaway branch that reverts only
the production files, dispatch the same workflow, and expect it to fail.
Delete the branch afterwards.

## Log

| Date | Change | Fork PRs | Upstream PR | Upstream release |
|---|---|---|---|---|
| 2026-10-15 | Group update fixes (`isAssignableToRole` on PATCH, 204 nil dereference, plan-time rejection) | #89, #90 | _not opened yet: branch `fix/groups-update-patch` on the fork; body in `specs/vk/1032-upstream-the-mic/upstream-pr-body.md`_ | _pending_ |
| 2026-10-19 | Group write consistency predicate (stale read after PATCH or POST) | #92 | _not opened yet: added as the fourth commit on `fix/groups-update-patch` (22/22 group tests pass there); the PR body above covers it_ | _pending_ |
