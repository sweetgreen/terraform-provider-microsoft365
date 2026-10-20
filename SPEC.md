# SPEC: Keep reading after a group write until Graph returns the new values

Task: vk 6a42. Release target: `v1.3.0-sg.4` on the fork's `sg/fix-release-tool-cache` line.

## Problem

`microsoft365_graph_beta_groups_group` can fail an update that has already succeeded.
`Update` (`internal/services/resources/groups/graph_beta/group/crud.go`) PATCHes the group,
then calls `crud.ReadWithRetry` with `DefaultReadWithRetryOptions()` and no
`ConsistencyPredicate`. `ReadWithRetry` returns on the first read without an error. Microsoft
Entra is eventually consistent, so that first read can come from a replica that still holds
the old values. The provider writes the stale values to state, and Terraform reports
"Provider produced inconsistent result after apply".

Evidence (constitution V): on 2026-10-19 terraform-infrastructure #1907 applied 12
`membership_rule` changes in sweetgreen-test. Nine failed this way (POS Rings 1, 2, 3, 6;
QSR Rings 1, 3, 4; BOH Rings 1, 2). The re-plan refreshed every rule as configured, so every
PATCH had landed and only the read straight after it was stale. Microsoft documents this
behaviour and the remedy (trust the write, poll until the read reflects it):
<https://devblogs.microsoft.com/identity/designing-for-eventual-consistency-for-microsoft-entra/>.

`Create` already sets a predicate (`groupConsistencyPredicate`, `predicate.go`), but it only
checks `id`, `display_name` and `mail_nickname`. A create whose first read returns the group
without its `membership_rule` would fail the same way.

## Goals

1. After a successful PATCH (Update) or POST (Create), keep reading until every attribute the
   request sent reads back with the sent value.
2. Compare after the same normalisation `Read` applies, so the predicate only waits for
   differences that would otherwise make Terraform fail.
3. Keep the existing limits: `DefaultReadWithRetryOptions()` (30 retries, 2 s apart), capped
   by the operation's timeout (default 180 s; `timeouts.update` / `timeouts.create`).
4. If the reads never catch up, fail with an error that names each attribute still returning
   a different value, with the sent and the last-read values. Do not leave Terraform to report
   its generic inconsistent-result error.

## Non-goals

- Changing `crud.ReadWithRetry` or its options (shared by every resource; upstream code).
- Clearing `membership_rule` by removing it from config. The provider doesn't send a null
  today, which is a separate bug.
- Owners and members (create-time `@odata.bind` only; not part of the group object read).

## Design

### Which attributes are checked

An attribute is checked when the request body set it. Its expected value is the request
body passed through `MapRemoteStateToTerraform`, the function `Read` uses. So a value is
expected exactly as `Read` would produce it if Graph returned what was sent.

| Attribute | Sent when | Notes |
|---|---|---|
| `display_name` | always (required) | |
| `description` | plan value known and not null | Optional+Computed; unknown when unset |
| `mail_nickname` | always (required) | |
| `mail_enabled` | always (required) | |
| `security_enabled` | always (required) | |
| `group_types` | plan set known, not null, not empty | compared as a set (order-free) |
| `visibility` | always (defaults to `Private`) | |
| `is_assignable_to_role` | create only | never sent on update (#89) |
| `membership_rule` | plan value known and not null | |
| `membership_rule_processing_state` | plan value known and not null | |

`id` must also be non-empty, as the current create predicate checks.

### Shape

`predicate.go` replaces `groupConsistencyPredicate(expected *GroupResourceModel)` with a small
tracker built from the request body:

```go
type groupConsistency struct { /* expected values, checked attributes, last mismatches */ }
func newGroupConsistency(ctx context.Context, sent graphmodels.Groupable) *groupConsistency
func (c *groupConsistency) predicate(ctx context.Context, state tfsdk.State) bool // for opts.ConsistencyPredicate
func (c *groupConsistency) staleAttributes() []string // "membership_rule (sent \"x\", read \"y\")"
```

`predicate` records which checked attributes differ on each read. `ReadWithRetry` gives back
only a generic error. When it fails, `Create` and `Update` append the recorded mismatches to
the diagnostic:

> Error reading resource state after update: Graph accepted the update to group `<id>`, but
> after the retry limit it still returns a different value for: membership_rule (sent "...",
> read "..."). The change may still be replicating. Run `terraform plan` to check whether it
> has landed.

`Update` also stops reading the ID from the PATCH response, a fix that `v1.3.0-sg.3` already
carries. This task branches from that tag, so there's no change there.

## Tests (offline, `httpmock`; constitution VI)

The group mock gains a stale-read switch. After a PATCH, the next N GETs of that group
return the pre-PATCH body. N = -1 means it never catches up.

1. `TestUnitResourceGroup_12_UpdateWaitsForStaleMembershipRule`: create a dynamic group, then
   change `membership_rule` with one stale GET after the PATCH. The update succeeds and state
   holds the new rule. Without the fix this step fails with "inconsistent result after apply".
2. `TestUnitResourceGroup_13_UpdateNeverConsistentNamesAttribute`: same change, but GETs never
   catch up, with `timeouts = { update = "6s" }`. The step fails in about 5 s with an error
   matching `membership_rule`.
3. Pure-function tests of the tracker: it ignores attributes that weren't sent, compares sets
   regardless of order, and reports the stale attribute with both values.

## Rollout

1. Merge into `sg/fix-release-tool-cache`, then tag `v1.3.0-sg.4` (annotated). The tag push
   runs `provider-release.yml`. The Sweetgreen registry serves GitHub releases.
2. terraform-infrastructure: bump `provider_version_pin` in
   `environments/microsoft365/sweetgreen-test/groups/retail-automox-routing/terragrunt.hcl`
   to `1.3.0-sg.4`. Expect no resource changes in the Atlantis plan.
3. Upstream: add it to the PLA-362 upstream set (`fix/groups-update-patch`).

## Acceptance criteria

- [ ] Changing a group's `membership_rule` succeeds in one apply even when the first read after
      the PATCH is stale (test 1).
- [ ] A read that never catches up fails within the update timeout with an error naming the
      stale attribute (test 2).
- [ ] `v1.3.0-sg.4` is released, and terraform-infrastructure pins it before production adoption (#1892).
