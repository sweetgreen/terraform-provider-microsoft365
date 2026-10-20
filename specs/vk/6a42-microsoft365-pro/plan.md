# Implementation Plan: Group writes wait for Graph to return the new values

**Spec**: `./spec.md`
**Status**: Draft

## Technical Context

- Provider: Go, terraform-plugin-framework, Kiota `msgraph-beta-sdk-go`. Fork
  `sweetgreen/terraform-provider-microsoft365`, release line `sg/fix-release-tool-cache`
  (tip = tag `v1.3.0-sg.3`, `221ac610`). Releases are cut by pushing an annotated `v*` tag
  (`.github/workflows/provider-release.yml`).
- The resource code is in `internal/services/resources/groups/graph_beta/group/`: `crud.go`,
  `predicate.go`, `state.go` (`MapRemoteStateToTerraform`), `construct.go` (`constructResource`)
  and `mocks/responders.go`.
- Shared helper `internal/services/common/crud/read_with_retry.go`:
  `ReadWithRetryOptions.ConsistencyPredicate func(ctx, tfsdk.State) bool`. The helper isn't
  changed.
- Constraint: no local `go build`/`go test`. The SDK compile exceeds the worker's 9.3 GiB cgroup.
  Tests run in the fork's `pr-tests.yml` (on PR, or by `workflow_dispatch` with `base_ref`).
  Locally only `gofmt` runs.
- Infra: OpenTofu 1.9 / Terragrunt 0.78, Atlantis. The pin is the
  `generate "provider_version_pin"` block in
  `environments/microsoft365/sweetgreen-test/groups/retail-automox-routing/terragrunt.hcl`.

## Architecture & Approach

1. **Consistency tracker** (`predicate.go`, which replaces `groupConsistencyPredicate`). It is
   built from the request body that was sent. The expected values are
   `MapRemoteStateToTerraform(ctx, &expected, requestBody)`, the same mapping `Read` applies
   (FR-4). The checked attributes are those whose getter on the request body is non-nil (FR-3).
   `predicate` loads the read state and requires a non-empty `id`. It compares each checked
   attribute with `attr.Value.Equal` (sets compare without regard to order) and saves the
   mismatches. `staleAttributes()` returns them for the error (FR-6). See
   `contracts/consistency-tracker.md`.
2. **`crud.go`**: `Create` and `Update` set `opts.ConsistencyPredicate = c.predicate` on
   `DefaultReadWithRetryOptions()`, so the limits are unchanged (FR-5). When the first read is
   already consistent, the helper returns at once (FR-7). On a `ReadWithRetry` error, the
   diagnostic detail comes from `consistencyErrorDetail(op, id, err, c.staleAttributes())`
   (FR-6; see `contracts/consistency-error.md`).
3. **Mock** (`mocks/responders.go`): `SetStaleReadsAfterPatch(n)`. The PATCH responder snapshots
   the pre-merge group. The GET responder serves the snapshot for the next `n` GETs (`-1` means
   forever). It is reset with the rest of the mock state.
4. **Tests**: `resource_test.go` gets cases 12 (one stale read, which succeeds) and 13 (never
   consistent, `timeouts.update = "6s"`, `ExpectError` `membership_rule`). `predicate_test.go`
   gets pure tests of the tracker.
5. **Release**: a PR into `sg/fix-release-tool-cache`, then CI, merge, and the annotated tag
   `v1.3.0-sg.4` on the merge commit. Watch `provider-release.yml`.
6. **Infra**: a one-line pin bump (`1.3.0-sg.3` to `1.3.0-sg.4`, and the comment naming the
   fix) on `vk/6a42-microsoft365-pro`. The PR's Atlantis plan has to wait for #1907 to merge.
   Comment on #1892 once the release is published.
7. **Docs**: SPEC/PRIOR_KNOWLEDGE/IMPLEMENTATION_PLAN and the knowledge base go into the
   provider's `main` through the task branch (precedent: #91).

## Data Model

No persistent data. The tracker's in-memory shape is in `contracts/consistency-tracker.md`.

## Contracts

- `contracts/consistency-tracker.md`: the tracker's constructor, predicate and stale report.
- `contracts/consistency-error.md`: the user-facing error text (FR-6).
- `contracts/mock-stale-reads.md`: the test mock's stale-read switch.

## Research Notes

See `./research.md`.

## Constitution Check

Infra constitution 1.17.0:

- I, II: The pin bump goes through a PR and an Atlantis plan. It isn't merged before the plan
  runs and shows no changes.
- III: The infra PR contains only the pin change; the provider work is in other repos and PRs.
- IV: Exact pin `1.3.0-sg.4`, with a comment naming the fix, as its own PR (the 1.17.0 amendment).
- V: The plan must show no resource changes (FR-9). Any change is a stop.
- IX: Terraform/OpenTofu terminology in titles and descriptions.
- X: Knowledge is recorded in the provider `knowledge/`, plus a terraform-infrastructure topic
  update if it is reusable there.

Provider constitution 1.1.0: II/VI (offline httpmock tests that fail without the fix, proven by
a revert run in CI), III (only the group package changes; the shared helper is untouched), IV
(goes upstream with PLA-362; branch noted), V (cites the #1907 evidence and Microsoft's guidance),
and the new post-write consistency constraint (implemented here). No deviations.

## Risks & Dependencies

- **Atlantis lock:** #1907 locks the sweetgreen-test project until it merges, which isn't in
  this task's control.
- **CI capacity:** `pr-tests.yml` runs on `ubuntu-24.04-arm` (hosted). The fork's scheduled runs
  have failed recently (docs and permissions jobs, not pr-tests). If the runners are unavailable,
  the tests can't be proven.
- **Release signing:** `provider-release.yml` signs with the GPG key from SSM (#79). `sg.3`
  released this way on 2026-10-15.
- **Slow test 13:** it takes about 5 s by design, which is acceptable.
- **Write that never lands:** if Graph accepted the write but the value really doesn't land
  (a server-side rewrite), the apply now takes about 60 s before failing. The old behaviour
  failed at once. The new error is clearer, so the trade is accepted.
