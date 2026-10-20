# Implementation plan (task 6a42)

Code and docs go to different branches:

- **Code:** branch `sg/group-write-consistency` from `origin/sg/fix-release-tool-cache`
  (`v1.3.0-sg.3`), with a PR into `sg/fix-release-tool-cache`. Released as `v1.3.0-sg.4`.
- **Docs and knowledge:** the task branch `vk/6a42-microsoft365-pro` (from `main`), with a PR
  into `main`. It holds SPEC, PRIOR_KNOWLEDGE, this plan, `specs/vk/6a42-…` and `knowledge/`
  updates, as #91 did.
- **Pin bump:** terraform-infrastructure branch `vk/6a42-microsoft365-pro`.

## Steps

1. Create a worktree for `sg/group-write-consistency` from `origin/sg/fix-release-tool-cache`.
2. `predicate.go`: replace `groupConsistencyPredicate` with `groupConsistency`:
   - `newGroupConsistency(ctx, sent graphmodels.Groupable)` maps the request body through
     `MapRemoteStateToTerraform` into `expected`, and records which attributes were sent
     (getter non-nil; `groupTypes` non-empty).
   - `predicate(ctx, state)` reads the state into a model. It requires a non-empty ID, compares
     each sent attribute with `Equal`, stores the mismatches (`name (sent X, read Y)`), and
     returns true when there are none.
   - `staleAttributes()` returns the last mismatches.
3. `crud.go`:
   - Create: build `consistency := newGroupConsistency(ctx, requestBody)` and set
     `opts.ConsistencyPredicate = consistency.predicate`. On error, add the stale attributes to
     the diagnostic.
   - Update: do the same. Error summary "Error reading resource state after update"; the detail
     names the group ID and the stale attributes and says the write was accepted.
   - Share the detail builder (`consistencyErrorDetail`) between the two.
4. `mocks/responders.go`: add `SetStaleReadsAfterPatch(n int)`. In the PATCH responder, snapshot
   the group before merging and arm the stale counter. In the GET responder, serve the snapshot
   while the counter is non-zero (decrement when > 0; -1 means forever). Reset in
   `RegisterMocks` and `CleanupMockState`.
5. Tests:
   - `resource_test.go` case 12: one stale read after the PATCH. The update succeeds and the new
     rule is in state.
   - Case 13: reads never catch up, `timeouts = { update = "6s" }`, `ExpectError` matching
     `membership_rule`.
   - `predicate_test.go` (internal package): the tracker ignores attributes that weren't sent,
     compares sets regardless of order, reports the mismatch with both values, and requires an ID.
6. Format with `gofmt` (a single-file run is cheap). Don't run `go vet` or `go test` locally;
   see PRIOR_KNOWLEDGE.
7. Push, open the PR into `sg/fix-release-tool-cache`, and let CI run `pr-tests.yml` on the changed
   packages. Also prove case 12 fails without the fix: on a throwaway branch, drop the predicate
   from Update and dispatch `pr-tests.yml`.
8. Run a Codex review, fix confirmed findings, and re-run CI.
9. Merge, tag `v1.3.0-sg.4` (annotated) on the merge commit, push the tag, and watch
   `provider-release.yml` publish the GitHub release.
10. terraform-infrastructure: bump the pin to `1.3.0-sg.4` and open a PR. Expect the Atlantis plan
    to show no resource changes. Then approve, apply and merge, per the Atlantis workflow.
11. Knowledge base: update the `graph-group-patch-behaviour` topic (or add an
    eventual-consistency topic) and the index, tag them `task-6a42`, and add a row to the
    upstreaming log.
