# Research: group write consistency

## R1. Where to get the "expected" values

- **Chosen:** map the request body through `MapRemoteStateToTerraform`. This is the
  normalisation `Read` uses (nil→null, empty set→null, nil `isAssignableToRole`→false), so the
  predicate waits only for differences that would make Terraform fail anyway. The set of sent
  attributes falls out of the same body (getter non-nil).
- Rejected: comparing with the plan model directly. Unknown Optional+Computed values
  (`description`, `group_types`, `membership_rule_processing_state`) would need special
  handling, and `Read`'s normalisation would have to be duplicated.

## R2. How to name the stale attribute

`ReadWithRetry` only takes `func(ctx, tfsdk.State) bool` and returns a generic error.

- **Chosen:** a tracker struct in the group package. Its predicate method records the
  mismatches, and `crud.go` adds them to the diagnostic. The shared helper isn't changed, so
  there's nothing to coordinate upstream with the other resources that use the helper.
- Rejected: changing the option to `func(...) error`. That touches every resource that uses
  `ConsistencyPredicate` (about 10 packages) and widens the upstream PR.

## R3. Timing in tests

`ReadWithRetry` caps the number of retries at `(deadline - 1s) / interval`. With
`timeouts = { update = "6s" }` and a 2 s interval it reads at 0, 2 and 4 s, then stops. The
never-consistent test therefore finishes in about 5 s without a test-only hook for the interval.

## R4. Evidence that the read lags the write (constitution V)

- terraform-infrastructure #1907 apply output (comment 6277927974): 9 of 12 `membership_rule`
  updates failed with inconsistent result. The re-plan showed every rule as configured.
- Microsoft, "Designing for eventual consistency for Microsoft Entra": writes and reads can hit
  different replicas, so clients should poll until the read reflects the write. The helper's own
  doc comment cites the same guidance.

## R5. Release mechanics

`v1.3.0-sg.3` is an annotated tag on `sg/fix-release-tool-cache`. Pushing a `v*` tag runs
`provider-release.yml`: goreleaser, with the GPG key from SSM. `registry.sweetgreen.engineering`
proxies GitHub releases, so no registry step is needed.

No new dependencies.
