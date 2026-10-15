# Implementation plan

1. Create the upstream branch `fix/groups-update-patch` from `upstream/main` in a
   separate worktree, leaving this task branch alone.
2. Cherry-pick `07905410` (#89: construct.go, state.go,
   construct_is_assignable_test.go). Resolve any conflicts against upstream
   and drop fork-only PR numbers from the commit subject.
3. Cherry-pick `221ac610` (#90: crud.go, resource_test.go case 11) the same way.
4. Run `gofmt -l` and `go vet` on the package, then run
   `go test ./internal/services/resources/groups/graph_beta/group/ -run 'TestConstructResource_|TestMapRemoteState_|TestUnitResourceGroup_' -count=1`.
5. Prove the tests catch the bugs: revert each production change locally and
   confirm the matching test fails. Discard the reverts afterwards.
6. Run a Codex review of `upstream/main...fix/groups-update-patch` and fix any
   confirmed findings.
7. Push the branch to `origin` (the sweetgreen fork). Open the upstream PR with
   title `fix(groups): allow updating existing groups`, using a body that
   follows the upstream template and includes the evidence, the reproduction
   and links to the fork PRs.
8. Watch upstream CI, including title validation and PR tests, and record the
   results.
9. Record knowledge in `knowledge/` on this task branch. Open the task PR
   against fork `main` and merge it.
10. Follow-ups outside this session: respond to upstream review, then note the
    upstream release version so the `sg` line can drop the patches.
