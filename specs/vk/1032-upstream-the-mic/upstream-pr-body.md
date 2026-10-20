# Pull Request Description

## Summary

`microsoft365_graph_beta_groups_group` can't update an existing group. Two bugs combine to cause this. The first two commits fix them, the third handles a consequence of the first fix, and the fourth stops a stale read after a successful update from failing the apply:

1. **`isAssignableToRole` is sent on update.** `constructResource` builds the update PATCH from every attribute. The schema defaults `is_assignable_to_role` to `false`, so every PATCH includes `isAssignableToRole`. Graph accepts that property only when a group is created, so every update fails with:

   ```
   400 An error occurred while processing the write request (category: validation)
   ```

   This happens even when nothing role-related changed. The fix sends `isAssignableToRole` only on create.
   `MapRemoteStateToTerraform` now also reads a `null` `isAssignableToRole` as `false`, the schema default. Without this, an imported group that was never role-assignable plans a spurious `null -> false` change on an immutable property.

2. **Nil pointer dereference after a successful update.** Graph answers a successful group PATCH with **204 No Content**, so the SDK returns a nil group. `Update` then ran `*updatedResource.GetId()` (`crud.go:215`) and crashed the plugin ("Plugin did not respond") after the change had already been written. The fix keeps the ID from state; the ID can't change on update. The second bug was hidden while the first one was present, because every PATCH failed with 400 before reaching this point.

3. **Reject `is_assignable_to_role` changes on existing groups at plan time.** After fix 1, the attribute is never sent on update. It has no `RequiresReplace`, so flipping it on an existing group would still plan an in-place update. The PATCH would then succeed without it, and Terraform would report an inconsistent result after apply. `ModifyPlan` now fails the plan with a message that gives the group's current value and explains how to recreate the group.

4. **Keep reading after a write until Graph returns the new values.** Microsoft Entra is eventually consistent, so the GET straight after a successful PATCH can come from a replica that still holds the old values. `Update` passed no `ConsistencyPredicate` to `crud.ReadWithRetry` and stored that stale read, and Terraform failed with "Provider produced inconsistent result after apply" although the write had landed. `Create`'s predicate checked only `display_name` and `mail_nickname`. Both now use a predicate built from the request body. It is mapped through `MapRemoteStateToTerraform`, so it applies the same normalisation as `Read`, and it checks every attribute the request sent. Limits are the existing `DefaultReadWithRetryOptions()`, within the operation timeout. If the reads never catch up, the error says Graph accepted the write and names each stale attribute with the sent and read values. This follows Microsoft's guidance in [Designing for eventual consistency for Microsoft Entra](https://devblogs.microsoft.com/identity/designing-for-eventual-consistency-for-microsoft-entra/).

### Issue Reference

No upstream issue. Fixes 1, 2 and 4 have been running in Sweetgreen's fork (fix 3 is new in this PR):

- sweetgreen/terraform-provider-microsoft365#89: never send `isAssignableToRole` on update
- sweetgreen/terraform-provider-microsoft365#90: don't read the group ID from the update PATCH response
- sweetgreen/terraform-provider-microsoft365#92: keep reading after a group write until Graph returns the new values. Without it, 9 of 12 `membership_rule` updates in one apply failed this way on a live tenant, though every write had landed.

### Motivation and Context

**Graph behaviour.** We sent single-field `PATCH /beta/groups/{id}` probes to an existing security group:

| PATCH body contains only | Response |
|---|---|
| `isAssignableToRole` (value unchanged, `false`) | `400` validation error (above) |
| `description` | `204 No Content` |
| `mailNickname` | `204 No Content` |
| `mailEnabled` + `securityEnabled` | `204 No Content` |
| `visibility` | `204 No Content` |
| `groupTypes` | `204 No Content` |

This matches the attribute's own documentation: "This property can only be set while creating the group and is immutable."

**Reproduce on `main`:**

1. Import any existing security group into `microsoft365_graph_beta_groups_group`.
2. Change its `description` and run `terraform apply`.
3. Graph returns 400 (bug 1). With only bug 1 fixed, the apply crashes with "Plugin did not respond" after the description has been written (bug 2).

### Dependencies

None.

## Type of Change

- [x] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update (Wiki/README/Code comments)
- [ ] ♻️ Refactor (code improvement without functional changes)
- [ ] 🎨 Style update (formatting, renaming)
- [ ] 🔧 Configuration change
- [ ] 📦 Dependency update

## Testing

- [x] I have added unit tests that prove my fix is effective or that my feature works
  - `TestConstructResource_UpdateOmitsIsAssignableToRole` and `TestConstructResource_CreateSendsIsAssignableToRole` check the request bodies.
  - `TestMapRemoteState_NullIsAssignableToRoleReadsAsFalse` checks the null mapping.
  - `TestUnitResourceGroup_11_UpdateWithNoContentResponse` applies a group, changes its description and updates it against the package's existing 204 PATCH mock.
  - `TestUnitResourceGroup_12_IsAssignableToRoleChangeRejected` creates a group with `false`, changes it to `true`, and expects the plan-time error.
  - `TestUnitResourceGroup_13_UpdateWaitsForStaleMembershipRule` serves the pre-PATCH group for the first GET after the PATCH (new `SetStaleReadsAfterPatch` mock switch). The update succeeds and state holds the new rule. `TestUnitResourceGroup_14_UpdateNeverConsistentNamesAttribute` never catches up, with `timeouts.update = "6s"`, and expects the new error naming `membership_rule`. `predicate_test.go` covers the predicate on its own.
  - Each test fails when its fix is reverted. We checked this in CI on throwaway branches: reverting the production code makes the construct and null-mapping tests fail, and test 11 panics at `crud.go:215`. Test 12 matches on an error string that exists only in the new `plan_modify.go` code, so it cannot pass with that code reverted. With the Update predicate removed, tests 13 and 14 fail with "Provider produced inconsistent result after apply: .membership_rule". A CI run reverting only that file hit the workflow's 60-minute job timeout before reporting.
- [x] New and existing unit tests pass locally with my changes
  (`go test ./internal/services/resources/groups/graph_beta/group/ -count=1`, run through this repo's `PR Tests` workflow on our fork: 22/22 group tests pass)
- [x] I have tested this code in the following browsers/environments: a live Entra tenant through Sweetgreen's fork (release `v1.3.0-sg.3`, fixes 1 and 2; `v1.3.0-sg.4` adds fix 4), updating existing security groups

## Quality Checklist

- [x] I have reviewed my own code before requesting review
- [x] I have verified there are no other open Pull Requests for the same update/change
- [ ] All CI/CD pipelines pass without errors or warnings
- [x] My code follows the established style guidelines of this project
- [x] I have added necessary documentation (if appropriate)
- [x] I have commented my code, particularly in complex areas
- [ ] I have made corresponding changes to the README and other relevant documentation
- [x] My changes generate no new warnings
- [x] I have performed a self-review of my own code
- [x] My code is properly formatted according to project standards

## Additional Notes

**Why not `RequiresReplace`?** The attribute is `Optional`+`Computed` with `Default: false`. An existing role-assignable group, imported with a config that omits `is_assignable_to_role`, would plan a destroy and recreate. For a group that holds role assignments, that is a dangerous plan to approve by accident. The plan-time error is the conservative choice. If you'd prefer `RequiresReplace`, I can change the PR.

The error message deliberately does not suggest `terraform apply -replace`. Terraform plans against the existing state before applying the forced replacement, so the error would block that path as well.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
