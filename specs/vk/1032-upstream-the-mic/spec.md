# Feature Specification: Upstream the group-update fixes

**Feature dir**: `specs/vk/1032-upstream-the-mic/`
**Status**: Clarified

## Summary
On upstream `deploymenttheory/terraform-provider-microsoft365` `main`,
`microsoft365_graph_beta_groups_group` cannot update an existing group. Every
update is rejected by Microsoft Graph (400) because the provider sends
`isAssignableToRole`, a property Graph accepts only at creation. If that is
fixed, a successful update then crashes the provider: Graph replies
`204 No Content`, and the code reads the group ID from a body that doesn't
exist. Sweetgreen fixed both bugs in its fork (sweetgreen#89, #90; released as
`v1.3.0-sg.2`/`-sg.3`). This feature contributes the fixes upstream so every
user of the provider can update groups and the fork can drop its patches.

## User Stories
- As a Terraform user managing Entra groups, I want to change an existing
  group's attributes (for example its description) so that `terraform apply`
  succeeds instead of failing with a Graph 400 or a plugin crash.
- As a user importing existing security groups, I want an unchanged import to
  plan no changes so that I'm not shown a spurious diff on an immutable
  property.
- As a Sweetgreen maintainer of the fork, I want these fixes merged and
  released upstream so that the `sg` line can retire its fork-only patches.

## Functional Requirements
- FR-1: An update request for a group must never include `isAssignableToRole`.
- FR-2: A create request must still send `isAssignableToRole` as configured.
- FR-3: A group whose `isAssignableToRole` Graph reports as null must read
  into state as `false` (the schema default). A `true` value still reads as
  `true`.
- FR-4: A successful update that returns no response body must complete
  without error, keeping the group's existing ID in state.
- FR-7: Changing `is_assignable_to_role` on an existing group must fail at plan
  time with a clear message. It must not be silently dropped from the update.
  (Added after the Codex review, 2026-10-15.)
- FR-5: Each of FR-1 to FR-4 and FR-7 has an offline unit test that fails without the
  fix.
- FR-6: The change goes to upstream as a PR against `main` that contains only
  these two fixes and their tests. The PR passes upstream title validation and
  the unit-test CI, and its description gives the Graph evidence, a
  reproduction and links to the fork PRs.

## Out of Scope
- PLA-361 / `prevent_duplicate_names`.
- Other group attributes that Graph may treat as create-only. No evidence
  exists for any beyond `isAssignableToRole`.
- Rebasing or changing the fork's `sg` line.
- Live-tenant acceptance tests.

## Acceptance Criteria
- [ ] `go test ./internal/services/resources/groups/graph_beta/group/ -run 'TestConstructResource_|TestMapRemoteState_|TestUnitResourceGroup_' -count=1`
      passes on the upstream branch.
- [ ] Reverting either production change makes its test fail.
- [ ] An upstream PR is open against `deploymenttheory/terraform-provider-microsoft365`
      `main`, its title passes `pr-title-validation.yml`, and upstream CI
      unit tests are green or any failure is explained.
- [ ] The upstream merge or feedback is recorded on the tracking issue, and so
      is the upstream release that contains the fixes.

## Clarifications (2026-10-15)
- **One PR or two?** One upstream PR with two commits, one per fix.
  Upstream squash-merges, with the commit subject taken from the PR title, so
  the title must cover both fixes: `fix(groups): allow updating existing groups`.
- **Which fork and branch?** Branch `fix/groups-update-patch` is cut from
  `upstream/main` and pushed to `sweetgreen/terraform-provider-microsoft365`,
  a GitHub fork of upstream that the token has admin rights on.
- **What does this task merge?** This task's own PR, against fork `main`,
  which carries the planning artifacts and the new knowledge base. The
  upstream PR is opened and left for upstream maintainers to review and merge.
  Its merge, and the release version that ships it, are tracked as follow-ups
  on the issue.

## Open Questions
- None blocking. Whether upstream maintainers accept the fixes as written is
  only known after review; that is tracked by the acceptance criteria.
