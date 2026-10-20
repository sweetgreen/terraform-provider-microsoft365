---
title: Reading back after a Graph write (Entra eventual consistency)
tags: [graph, entra, eventual-consistency, groups, testing, task-6a42]
updated: 2026-10-19
---

# Reading back after a Graph write

## The failure

Entra is eventually consistent. The GET straight after a successful POST or PATCH can come from
a replica that still has the old object. `crud.ReadWithRetry` with no `ConsistencyPredicate`
returns on the first read without an error, so the provider stores the stale values. Terraform
then fails with **"Provider produced inconsistent result after apply"**, though the write landed.

How to recognise it: the next `plan` refreshes the resource as configured and shows no resource
changes. On 2026-10-19 this happened to 9 of the 12 `membership_rule` updates in
terraform-infrastructure #1907 on sweetgreen-test, with `microsoft365_graph_beta_groups_group`
`1.3.0-sg.3`. Fixed in `1.3.0-sg.4` (fork PR #92).

## The pattern (group resource, reusable for any resource)

- Build the tracker from the **request body, before sending it**. Map the body through the
  resource's own `MapRemoteStateToTerraform` to get the expected values. That applies exactly the
  normalisation `Read` does, so the predicate only waits for differences Terraform would reject
  anyway.
- Check only attributes the body set (getter non-nil, or a non-empty slice). Unknown
  Optional+Computed plan values are never sent and must not be compared.
- Compare with `attr.Value.Equal`. `types.Set` equality ignores order.
- `ConsistencyPredicate` returns only a `bool`, and `ReadWithRetry` returns a generic
  "consistency predicate not yet satisfied" error. To name the stale attribute, let the
  predicate method record its mismatches on the tracker and add them to the diagnostic in
  `Create`/`Update`. Leave the shared helper unchanged.
- Limits come from `DefaultReadWithRetryOptions()`: about 30 reads, 2 s apart, capped at
  `(deadline - 1 s) / interval`. A user's `timeouts` block bounds the wait.

See `internal/services/resources/groups/graph_beta/group/predicate.go` on the `sg` line.

## Testing stale reads offline

- The group mock has `SetStaleReadsAfterPatch(n)`. The PATCH responder snapshots the group
  before merging, and the next `n` GETs serve the snapshot (`-1` = forever). Arm it from the
  update step's `PreConfig`.
- DELETE must clear the stale state. Otherwise a never-consistent test hangs in teardown, because
  delete verification keeps reading the snapshot.
- For a never-consistent test, set `timeouts = { update = "6s" }` in the config. The retry cap
  then gives about 3 reads and the step fails in about 5 s, with no test hook for the interval.
- `ExpectError` must match text the old code can't produce. Terraform's own
  inconsistent-result error also contains the attribute name, so also match the summary and a
  word from the new detail, such as `(?s)Error reading resource state after update.*accepted.*membership_rule`.
