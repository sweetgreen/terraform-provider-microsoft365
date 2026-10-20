# Prior knowledge (task 6a42)

Sources searched (read-only): this repo's `knowledge/` (2 topics, both from task 1032),
terraform-infrastructure `.project-knowledge/` (27 topics), and the Sweetgreen shared KB
(`kb__search_knowledge`, "microsoft365 terraform provider group membership_rule inconsistent
result after apply Entra eventual consistency"). The shared KB had no relevant entries; its
hits were guest-account and Intune SOPs.

## From `knowledge/topics/graph-group-patch-behaviour.md`

- A successful group PATCH returns `204 No Content`, and Kiota returns `(nil, nil)`. Keep
  the ID from state and let `crud.ReadWithRetry` refresh everything else. This is fixed on the
  `sg` line in `v1.3.0-sg.3` (#90). Note that `main` still has the old dereference, because
  releases come from `sg/fix-release-tool-cache`, not `main`.
- `isAssignableToRole` is create-only, so it is never sent on update (#89). Graph returns `null`
  for it on groups that were never role-assignable, and `Read` maps that to `false`.
- The group mock's PATCH responder already answers 204 and merges the body into the stored
  group. GETs return the stored group, so there's no stale-read mode yet.

## From `knowledge/topics/upstreaming-fork-fixes.md`

- Fixes ship first on the fork's `sg` line as `vX.Y.Z-sg.N`, then go upstream. PLA-362's
  upstream branch is `fix/groups-update-patch`. Agents can't open PRs on deploymenttheory
  (fine-grained PAT, 403), so a human has to open that PR.
- **Don't build or test locally.** Compiling `msgraph-beta-sdk-go` needs about 10 GB and the
  worker cgroup caps memory at 9.3 GiB, so `go test`/`go vet` stall and starve the host. Run
  the package tests in the fork's CI instead:
  `gh workflow run pr-tests.yml -R sweetgreen/terraform-provider-microsoft365 --ref <branch> -f base_ref=<base sha>`.
  To prove a test catches the bug, push a throwaway branch with only the production change
  reverted and expect CI to fail.

## Repo facts confirmed for this task

- Release line: `origin/sg/fix-release-tool-cache` (tip `221ac610` = `v1.3.0-sg.3`, an annotated
  tag). Fork PRs #89 and #90 merged into that branch. A tag push of `v*` runs
  `provider-release.yml`.
- `crud.ReadWithRetryOptions.ConsistencyPredicate` already exists. Several agent and application
  resources use it via a `predicate.go`. The group `Create` uses a weak predicate (`id`,
  `display_name`, `mail_nickname`). `Update` sets none.
- `ReadWithRetry` returns only a generic "consistency predicate not yet satisfied" error. To
  name the stale attribute, the group package has to record mismatches itself.
- `Read` normalisation (`MapRemoteStateToTerraform`): strings map nil→null, empty string sets map
  to null, a nil `isAssignableToRole` maps to `false`, and nothing else is rewritten.

## From terraform-infrastructure `.project-knowledge/`

- `atlantis-apply-workflow.md`: the order is approve → `atlantis apply` → merge. Atlantis does not
  apply on merge here, and a PR merged before apply leaves drift. You can't approve your own PR.
- `terraform-registry-github-token.md`: `registry.sweetgreen.engineering` proxies GitHub
  releases, so a published GitHub release is enough for `tofu init` to fetch the new version.
- `entra-dynamic-group-adoption.md`: retail routing groups in sweetgreen-test are managed with
  `modules/microsoft365/groups`. The rule strings must match Graph byte for byte.
- The `retail-automox-routing` unit pins the provider with a `generate "provider_version_pin"`
  override file (`zz_provider_version_override.tf`), currently at `1.3.0-sg.3`.
