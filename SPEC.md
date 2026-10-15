# SPEC: Upstream the `microsoft365_graph_beta_groups_group` update fixes

## Problem

`microsoft365_graph_beta_groups_group` cannot update an existing group on
upstream `deploymenttheory/terraform-provider-microsoft365` `main` (checked at
`93ce6c1e`, 2026-10-15). Two bugs stack:

1. **`isAssignableToRole` in the update PATCH.** `constructResource`
   (`construct.go:32`) always calls
   `convert.FrameworkToGraphBool(data.IsAssignableToRole, requestBody.SetIsAssignableToRole)`.
   The schema defaults `is_assignable_to_role` to `false`, so every PATCH sends
   the property. Graph accepts it only at creation and answers
   `400 An error occurred while processing the write request` (validation).
   Single-field PATCH probes: `isAssignableToRole` → 400; `description`,
   `mailNickname`, `mailEnabled`/`securityEnabled`, `visibility`, `groupTypes`
   → 204.
2. **Nil dereference after a successful PATCH.** Graph answers a successful
   group PATCH with `204 No Content`; the Kiota SDK returns a nil `Groupable`,
   and `Update` (`crud.go:215`) runs `*updatedResource.GetId()` and panics
   after the write already happened.

A related read-side issue: Graph returns `null` for `isAssignableToRole` on
groups that were never role-assignable. `MapRemoteStateToTerraform` maps that
to a null bool, so an imported group plans a change from null to the `false`
default on an immutable property.

## Fix (as shipped in the fork, sweetgreen#89 / #90)

- `construct.go`: set `isAssignableToRole` only when `!isUpdate`.
- `state.go`: map a nil `isAssignableToRole` to `types.BoolValue(false)`.
- `crud.go`: discard the PATCH response; keep `plan.ID = groupId` from state.

## Tests

- New `construct_is_assignable_test.go`:
  `TestConstructResource_UpdateOmitsIsAssignableToRole`,
  `TestConstructResource_CreateSendsIsAssignableToRole`,
  `TestMapRemoteState_NullIsAssignableToRoleReadsAsFalse`.
- `resource_test.go`: `TestUnitResourceGroup_11_UpdateWithNoContentResponse`
  applies a group, changes its description, and asserts the update succeeds
  against the package's existing 204 PATCH mock.
- Command:
  `go test ./internal/services/resources/groups/graph_beta/group/ -run 'TestConstructResource_|TestMapRemoteState_|TestUnitResourceGroup_' -count=1`

## Delivery

- Branch `fix/groups-update-patch` from `upstream/main`, two commits (one per
  fix), pushed to the `sweetgreen/terraform-provider-microsoft365` fork (a
  GitHub fork of upstream).
- One upstream PR against `deploymenttheory/terraform-provider-microsoft365`
  `main`. Title must match upstream `pr-title-validation.yml`
  (`^(fix|feat|…)(\(.+\))?: .+$`), e.g.
  `fix(groups): allow updating existing groups`.
- PR body follows upstream `.github/PULL_REQUEST_TEMPLATE.md` and includes the
  Graph evidence, the reproduction (import any security group and change its
  description) and links to sweetgreen#89 and sweetgreen#90.

## Out of scope

- PLA-361 (`prevent_duplicate_names`).
- Any change to the fork's `sg` line; it drops the patches after an upstream
  release ships them.

## Acceptance

- Upstream PR open with both fixes and tests; group package unit tests pass
  locally and in upstream CI.
- Merge or upstream feedback recorded; release containing the fixes noted for
  retiring the fork patches. (Merge is upstream maintainers' decision and
  cannot be completed inside this task.)
