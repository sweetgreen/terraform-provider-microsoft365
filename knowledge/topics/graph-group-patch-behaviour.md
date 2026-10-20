---
title: Microsoft Graph group PATCH behaviour
tags: [graph, groups, microsoft365_graph_beta_groups_group, task-1032, task-6a42]
updated: 2026-10-19
---

# Microsoft Graph group PATCH behaviour

These facts apply to `microsoft365_graph_beta_groups_group`
(`internal/services/resources/groups/graph_beta/group/`). They were observed on
the sweetgreen-test tenant on 2026-10-14 and 15 (terraform-infrastructure #1899).

## `isAssignableToRole` is create-only

- A `PATCH /beta/groups/{id}` that includes `isAssignableToRole` returns
  `400 An error occurred while processing the write request` (category:
  validation). This happens **even when the value is unchanged**.
- Single-field PATCH probes of `description`, `mailNickname`,
  `mailEnabled`+`securityEnabled`, `visibility` and `groupTypes` each returned
  `204`.
- The 400 doesn't name the property. To find the culprit, PATCH one field at
  a time.
- Graph returns `null` for `isAssignableToRole` on groups that were never
  role-assignable. The provider reads that as `false`, the schema default.
- The attribute has **no `RequiresReplace`**. `ModifyPlan` rejects a change on
  an existing group at plan time. Don't suggest `terraform apply -replace` for
  this: Terraform plans against the existing state first, so the error blocks
  that path too. `RequiresReplace` was not used because, with the `false`
  default, an imported role-assignable group whose config omits the attribute
  would plan a destroy and recreate.

## A successful PATCH returns `204 No Content`

- Kiota's `Groups().ByGroupId(id).Patch(...)` then returns `(nil, nil)`.
  Never read fields from the PATCH result. Keep the ID from state and let
  `crud.ReadWithRetry` refresh everything else.
- When this goes wrong, Terraform reports "Plugin did not respond" **after**
  the write has landed in Graph. State then lags behind Graph until the next
  refresh.
- The package's `mocks/responders.go` PATCH responder already returns 204, so
  unit tests exercise this path.

## The read after a successful PATCH can be stale

Entra replicas lag. A 204 PATCH can be followed by a GET that returns the old values. Since
`1.3.0-sg.4`, Create and Update poll until every attribute they sent reads back. See
[reading back after a Graph write](entra-read-after-write-consistency.md).

## Where the fixes live

- Fork: sweetgreen/terraform-provider-microsoft365#89 and #90, released as
  `v1.3.0-sg.2` and `v1.3.0-sg.3`; #92 (stale reads after a write) as `v1.3.0-sg.4`.
- Upstream: see [upstreaming fork fixes](upstreaming-fork-fixes.md) for the PR
  and its status.
