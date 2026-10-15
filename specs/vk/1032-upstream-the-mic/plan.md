# Implementation Plan: Upstream the group-update fixes

**Spec**: `./spec.md`
**Status**: Ready

## Technical Context
- Go 1.26 (from `go.mod`; local toolchain go1.26.3). Uses terraform-plugin-framework
  v1.19, terraform-plugin-testing v1.16, msgraph-beta-sdk-go v0.166 (Kiota)
  and jarcoal/httpmock v1.4.2. No dependency changes.
- Base: `upstream/main` @ `93ce6c1e`. Package:
  `internal/services/resources/groups/graph_beta/group/` (Go package
  `graphBetaGroup`).
- Upstream CI: `pr-title-validation.yml` (conventional-commit title) and
  `pr-tests.yml` (unit tests for changed packages, patch-coverage gate on
  non-draft PRs).

## Architecture & Approach
Three small, independent changes in the package, carried as two commits:

| Req | File | Change |
|-----|------|--------|
| FR-1, FR-2 | `construct.go` (`constructResource`, line 32) | Wrap the `SetIsAssignableToRole` call in `if !isUpdate`. |
| FR-3 | `state.go` (`MapRemoteStateToTerraform`, line 31) | If `GetIsAssignableToRole()` is nil, set `types.BoolValue(false)`; otherwise keep the existing conversion. Adds the `types` import. |
| FR-4 | `crud.go` (`Update`, lines 204–215) | Discard the `Patch` result (`_, err = …`) and set `plan.ID = types.StringValue(groupId)`. |
| FR-5 | `construct_is_assignable_test.go` (new) | Pure-function tests for FR-1 to FR-3. |
| FR-5 | `resource_test.go` | `TestUnitResourceGroup_11_UpdateWithNoContentResponse`: a two-step `resource.UnitTest` (description "before" → "after") against the existing mock in `mocks/responders.go`, whose PATCH responder returns `204` with an empty body (lines 161–179). |

| FR-7 | `plan_modify.go` (`ModifyPlan`) | On update, if the state and plan values are both known and differ, add an attribute error. Commit 3. |
| FR-7 | `resource_test.go` | `TestUnitResourceGroup_12_IsAssignableToRoleChangeRejected`: create with `false`, then change it to `true` and expect the error. |

Delivery mechanism: cherry-pick fork commits `07905410` (commit 1: FR-1 to FR-3)
and `221ac610` (commit 2: FR-4) onto a branch cut from `upstream/main`. The
upstream files touched are byte-identical to the fork's pre-fix versions at
those hunks, as confirmed by diffing `git show upstream/main:<file>`. The picks
should therefore apply cleanly, and any conflict is resolved by hand. Commit
subjects drop the fork PR numbers `(#89)`/`(#90)` so they don't resolve to
unrelated upstream PRs.

Work happens in a separate git worktree (`../tpm365-upstream`) so that this
task branch, which carries only the planning and knowledge artifacts, keeps
its own history.

## Data Model
Not applicable. No schema or model changes: `is_assignable_to_role` keeps its
schema, its `false` default and its `RequiresReplace` behaviour.

## Contracts
Not applicable. The resource's Terraform contract is unchanged. The Graph
contract that matters is recorded in `research.md`.

## Research Notes
See `./research.md`.

## Constitution Check
- I. Clarity: three local edits, each with a one-line comment explaining the Graph
  behaviour. ✅
- II. Test the contract: the acceptance criteria are concrete, and each FR has a test. ✅
- III. Small steps: two commits, no refactor. ✅
- IV. Upstream first: the branch is cut from `upstream/main`, contains only this
  change, and uses a conventional-commit title. ✅
- V. Graph evidence: the probe results go in research.md and the PR body. ✅
- VI. Mock unit tests: they run offline, and the "revert → fail" check is in
  the tasks. ✅
- VII. Retire fork patches: recording the release version is a follow-up on
  the issue. ✅

## Risks & Dependencies
- Upstream may already have changed these lines, or may do so before review.
  Mitigation: re-check at PR time and rebase if needed.
- Patch-coverage gate: the changed lines are all exercised by the new tests,
  so the risk is low.
- `is_assignable_to_role` has no `RequiresReplace` (checked in `resource.go:209`
  and `plan_modify.go`). If a user flips it on an existing group, Terraform
  still plans an in-place update. Before this change that update failed with
  Graph's 400. After it, the PATCH succeeds without the property, the
  read-back returns the old value, and Terraform reports an inconsistent
  result. The change is still rejected, just with a different error. Adding
  `RequiresReplace` would destroy and recreate groups, which changes user
  behaviour, so it stays out of this bug-fix PR. **Resolved after the Codex
  review:** commit 3 adds a plan-time error instead (FR-7). The message does
  not suggest `-replace`, because Terraform plans against the existing state
  first and the error would block that path. `RequiresReplace` is left to the
  maintainers.
- Local `go test` cannot run here: compiling msgraph-beta-sdk-go needs about
  10 GB, and this worker's cgroup sets `memory.high` at 9.3 GiB. Tests run in
  the fork's `PR Tests` workflow (dispatched with `base_ref=93ce6c1e`) and in
  upstream CI.
- Merge timing is up to upstream maintainers. The fork keeps
  `v1.3.0-sg.3` until upstream releases.
