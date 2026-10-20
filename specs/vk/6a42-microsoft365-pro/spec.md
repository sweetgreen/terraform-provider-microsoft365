# Feature Specification: Group writes wait for Graph to return the new values

**Feature dir**: `specs/vk/6a42-microsoft365-pro/`
**Status**: Draft
**Repos**: `sweetgreen/terraform-provider-microsoft365` (provider fix, release `1.3.0-sg.4`),
`sweetgreen/terraform-infrastructure` (pin bump)
**Technical spec**: `terraform-provider-microsoft365/SPEC.md`

## Summary

`microsoft365_graph_beta_groups_group` (provider `1.3.0-sg.3`) can report a failed update after
Microsoft Graph has already accepted it. After writing, the provider reads the group once.
Microsoft Entra is eventually consistent, so that read can still return the old values.
Terraform then fails with "Provider produced inconsistent result after apply". On 2026-10-19,
9 of the 12 `membership_rule` changes in terraform-infrastructure #1907 failed this way in
sweetgreen-test, although every write had landed. The recovery was a re-plan and a second
apply. Production adoption (#1892) changes the same rules. The provider must keep reading until
Graph returns what was written, and fail clearly when it never does. The fix ships as
`1.3.0-sg.4`, and the sweetgreen-test unit pins it.

## User Stories

- As an operator applying group changes through Atlantis, I want an apply whose writes all
  succeeded to finish green in one run, so that I don't have to re-plan, prove there are no
  changes and apply again, and outputs are saved.
- As an operator whose change really hasn't replicated, I want the error to name the attribute
  that is still stale, with the value sent and the value read, so that I know what to check.
- As the owner of production adoption (#1892), I want the provider pin to include this fix before
  production rules change, so that production doesn't repeat the sweetgreen-test failure.

## Functional Requirements

- FR-1: After a successful group update, the provider keeps reading the group until every
  attribute the update sent reads back with the sent value. Only then does it record state.
- FR-2: FR-1 also applies after a group is created.
- FR-3: Attributes covered: `display_name`, `description`, `mail_nickname`, `mail_enabled`,
  `security_enabled`, `group_types`, `visibility`, `membership_rule`,
  `membership_rule_processing_state`, and `is_assignable_to_role` on create. An attribute is
  checked only when the request sent it.
- FR-4: The comparison uses the same normalisation as a normal refresh. A value that a refresh
  would accept as matching is never waited on. Set attributes compare regardless of order.
- FR-5: Waiting keeps the existing limits: up to 30 re-reads, 2 seconds apart, and never past
  the operation's timeout (default 180 s, configurable with `timeouts`).
- FR-6: If the limit is reached and an attribute still differs, the operation fails with an
  error that says Graph accepted the write. The error names each stale attribute with the sent
  and last-read values.
- FR-7: Behaviour when the first read is already consistent is unchanged: no extra reads, no
  extra delay.
- FR-8: The fix is released as `1.3.0-sg.4` from the fork's release line, which carries the
  earlier `sg.2`/`sg.3` group fixes.
- FR-9: The `retail-automox-routing` sweetgreen-test unit pins `1.3.0-sg.4`. Its Atlantis plan
  shows no resource changes.

## Out of Scope

- Changing the shared read-retry helper for every resource. The fix is limited to the group
  resource.
- Clearing `membership_rule` by removing it from config. The provider doesn't send a null today;
  that's a separate defect.
- Owners and members, which are set only at creation and aren't part of the group read.
- Opening the upstream PR on deploymenttheory (needs a human; tracked under PLA-362).
- Merging #1907 or editing #1892's branch.

## Acceptance Criteria

- [ ] A unit test with a mocked Graph whose first read after the PATCH returns the old
      `membershipRule`, and whose second returns the new one, passes. The update succeeds and
      state holds the new rule. The same test fails on `1.3.0-sg.3` code.
- [ ] A unit test whose reads never catch up fails within its update timeout, with an error
      matching `membership_rule`.
- [ ] The existing group unit tests still pass in the fork's CI.
- [ ] GitHub release `v1.3.0-sg.4` exists with signed artifacts, and the registry serves it.
- [ ] The terraform-infrastructure PR pins `1.3.0-sg.4`, and its Atlantis plan for
      `environments/microsoft365/sweetgreen-test/groups/retail-automox-routing` shows no resource
      changes.

## Clarifications

### Session 2026-10-19

No answers were supplied, so each question was resolved with the default that best fits the
constitution:

- Q: #1907 holds the Atlantis lock and edits the same `terragrunt.hcl`. Should the pin bump wait
  or be folded in? → A: Use a separate PR (constitution III, IV: a pin bump is its own PR). It
  touches only the pin block, which #1907 doesn't change, so the two merge cleanly. Its Atlantis
  plan can run only after #1907 merges and releases the lock. If #1907 is still open when this
  task finishes, the pin PR stays open with that dependency stated, and isn't merged unplanned.
- Q: Who bumps #1892's production pin? → A: #1892's owner. This task doesn't push to another
  task's branch. Once `1.3.0-sg.4` is published, it posts a comment on #1892 naming the release
  and the one-line change. That covers the "pins the new release before production adoption"
  criterion.
- Q: Should the error include full values? → A: Yes. Show the sent and last-read values
  quoted, untruncated. Membership rules aren't secret, and the full text is what an operator
  compares.

FR-9 is refined: the pin PR's Atlantis plan is the evidence, and it's taken after #1907 merges.

## Open Questions

None.
