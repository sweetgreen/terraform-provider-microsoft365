# Prior knowledge

## Project knowledge base

This repository has no knowledge base yet: no `wiki/`, `knowledge/` or similar
directory exists on `main` or in this worktree. This is the first task to
record one. The Sweetgreen shared KB (`kb__search_knowledge`, queried for
"terraform-provider-microsoft365 fork sg release groups group update") had no
relevant entries; the hits were unrelated Confluence release plans.

## What the repo and its history already tell us

- `origin` is `sweetgreen/terraform-provider-microsoft365`, a GitHub fork of
  `deploymenttheory/terraform-provider-microsoft365` (`upstream`). The token
  has ADMIN on the fork, so upstream PRs can come from a branch pushed there.
- The fork ships `vX.Y.Z-sg.N` releases from an `sg/...` branch. The two group
  fixes are fork commits `07905410` (#89) and `221ac610` (#90).
- Upstream conventions (read from `upstream/main`):
  - `pr-title-validation.yml` requires the PR title to match
    `^(feat|fix|docs|style|refactor|test|chore|build|ci|perf)(\(.+\))?: .+$`.
    Release-please builds the changelog from these titles.
  - `.github/PULL_REQUEST_TEMPLATE.md` sets out the PR body: Summary, Motivation, Type
    of Change, Testing, and Quality Checklist.
  - `pr-tests.yml` runs unit tests only for changed Go packages and enforces a
    patch-coverage threshold on non-draft PRs.
  - Recent outside contributors, such as kenchan0130 and ndhillon-openai, used
    `fix: …` titles and had them merged.
- The group package's unit tests use `httpmock` mocks in `group/mocks`. The
  PATCH mock already answers `204 No Content`, which is why the 204 regression
  test needs no new mock.
