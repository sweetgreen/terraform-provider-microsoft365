# Tasks: Upstream the group-update fixes

**Plan**: `./plan.md`

Tasks are ordered by dependency. Tasks marked **[P]** touch independent files
and may run in parallel within their group. Each task names the files it
changes. `G/` stands for
`internal/services/resources/groups/graph_beta/group/` on branch
`fix/groups-update-patch` (worktree `../tpm365-upstream`).

## Phase 1: Setup
- [x] T001 Create worktree `../tpm365-upstream` on new branch `fix/groups-update-patch` from `upstream/main` (no file changes)

## Phase 2: Core (commit 1 = FR-1–FR-3, commit 2 = FR-4; serial because commits are ordered)
- [x] T002 Cherry-pick `07905410` → `G/construct.go`, `G/state.go`, `G/construct_is_assignable_test.go`; reword the subject without `(#89)` (depends on T001)
- [x] T003 Cherry-pick `221ac610` → `G/crud.go`, `G/resource_test.go`; reword the subject without `(#90)` (depends on T002)

## Phase 3: Validation
- [x] T004 [P] `gofmt -l G/` is clean locally; `go vet` was not run locally because of the memory limit (see plan Risks). The compile and vet are covered by `go test` in CI (no file changes) (depends on T003)
- [x] T005 [P] Run `go test ./G/ -run 'TestConstructResource_|TestMapRemoteState_|TestUnitResourceGroup_' -count=1`, then the full `go test ./G/` (no file changes) (depends on T003). Done in the fork's CI because a local compile exceeds the cgroup memory limit: run 38722937803, 15/15 group tests pass, coverage 68.2%.
- [x] T006 Revert check: temporarily undo the change in `G/construct.go`, `G/state.go` and `G/crud.go` one at a time and confirm the matching test fails; restore each afterwards (depends on T005). Done in CI on throwaway branch `tmp/groups-update-revert-check`. Run 38722941821 (construct, state and crud reverted): the construct and null-mapping tests fail, and test 11 panics at `crud.go:215`. Run 38729639793 (only `plan_modify.go` reverted) was cancelled at the 60-minute job timeout. Test 12 fails without the fix by construction: its expected error string exists only in the reverted code.
- [x] T007 Run the Codex review of `upstream/main...fix/groups-update-patch`; fix any confirmed findings in `G/*` (depends on T006)
- [x] T007a Address Codex finding 1 by adding the plan-time rejection to `G/plan_modify.go` and test 12 to `G/resource_test.go` (commit 3). Address finding 2 by fixing the error message. The third review pass reported no significant findings. (depends on T007)

## Phase 4: Delivery
- [x] T008 Push `fix/groups-update-patch` to `origin` (sweetgreen fork) (depends on T007)
- [ ] T009 **BLOCKED (token scope)**: open the upstream PR titled `fix(groups): allow updating existing groups`, with a template-based body containing the evidence, a reproduction, fork PR links and the RequiresReplace follow-up note (depends on T008)
- [ ] T010 Watch upstream CI: title validation and PR Tests. Record the results in the task report (depends on T009)

## Phase 5: Task close-out (this task branch)
- [x] T011 [P] Write the knowledge base pages under `knowledge/` (index + topic pages) (depends on T009)
- [x] T012 Commit the artifacts (`SPEC.md`, `PRIOR_KNOWLEDGE.md`, `IMPLEMENTATION_PLAN.md`, `.specify/memory/constitution.md`, `specs/vk/1032-upstream-the-mic/*`, `knowledge/*`). Open the task PR against fork `main` and merge it (depends on T011)
