# Tasks: Group writes wait for Graph to return the new values

**Plan**: `./plan.md`

Tasks are ordered by dependency. Tasks marked **[P]** touch independent files
and may run in parallel within their group. Each task names the file(s) it
changes. `P/` = `terraform-provider-microsoft365/` (worktree on branch
`sg/group-write-consistency` from `origin/sg/fix-release-tool-cache`);
`G/ = P/internal/services/resources/groups/graph_beta/group/`;
`I/` = `terraform-infrastructure/`.

## Phase 1: Setup
- [x] T001 Create the provider worktree and branch `sg/group-write-consistency` from
      `origin/sg/fix-release-tool-cache` (no file changes)

## Phase 2: Core (depends on T001)
- [x] T002 [P] Replace `groupConsistencyPredicate` with the `groupConsistency` tracker
      (`newGroupConsistency`, `predicate`, `staleAttributes`) in `G/predicate.go`
- [x] T003 [P] Add `SetStaleReadsAfterPatch` plus the PATCH snapshot and GET stale serving and
      their reset in `G/mocks/responders.go`
- [x] T004 Wire the tracker into `Create` and `Update` and add `consistencyErrorDetail` in
      `G/crud.go` (depends on T002)

## Phase 3: Validation
- [x] T005 [P] Add the tracker unit tests in `G/predicate_test.go` (depends on T002)
- [x] T006 [P] Add `TestUnitResourceGroup_12_UpdateWaitsForStaleMembershipRule` and
      `TestUnitResourceGroup_13_UpdateNeverConsistentNamesAttribute` in `G/resource_test.go`
      (depends on T003, T004)
- [x] T007 Run `gofmt -l` on the changed files, commit, push and open a PR into
      `sg/fix-release-tool-cache`, then wait for `pr-tests.yml` to pass (depends on T005, T006)
- [x] T008 Prove that case 12 fails without the fix: on a throwaway branch, revert `G/crud.go`'s
      Update predicate wiring, dispatch `pr-tests.yml` and expect a failure, then delete the branch
      (depends on T007)

## Phase 4: Review and release
- [x] T009 Run a Codex review of the provider diff and fix confirmed findings in `G/*` (depends on T007)
- [ ] T010 Merge the PR, tag `v1.3.0-sg.4` (annotated) on the merge commit, push it, and confirm
      `provider-release.yml` published the release (depends on T008, T009)

## Phase 5: Consumers (depends on T010)
- [ ] T011 [P] Bump the pin to `1.3.0-sg.4` and update its comment in
      `I/environments/microsoft365/sweetgreen-test/groups/retail-automox-routing/terragrunt.hcl`;
      open a PR and get an Atlantis plan showing no changes after #1907 releases its lock
- [ ] T012 [P] Comment on terraform-infrastructure #1892 with the release and the one-line pin
      change (no file changes)

## Phase 6: Knowledge
- [ ] T013 [P] Record what was learned (tagged `task-6a42`) in
      `P/knowledge/topics/graph-group-patch-behaviour.md`,
      `P/knowledge/topics/upstreaming-fork-fixes.md` and `P/knowledge/README.md`, and commit the
      SPEC, PRIOR_KNOWLEDGE and IMPLEMENTATION_PLAN on `vk/6a42-microsoft365-pro`; open a PR
      into the provider `main`
- [ ] T014 [P] Update `I/.project-knowledge/topics/entra-dynamic-group-adoption.md` and
      `I/.project-knowledge/INDEX.md` with the `1.3.0-sg.4` requirement for group rule changes

<!--
Conventions:
- `T001` … task ids are stable and referenced by the dependency graph.
- `[P]` … parallel-safe (independent files). Omit for tasks that must be serial.
- `[ ]` / `[x]` … completion checkbox, toggled from the workbench.
-->
