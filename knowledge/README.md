# Knowledge base

Reusable knowledge from work on Sweetgreen's fork of
terraform-provider-microsoft365. Each topic page lists the task ids that
contributed to it in its `tags`. The registry docs live in `docs/`, which is
generated, so this knowledge base is kept outside it.

## Index

| Topic | Summary | Tags |
|---|---|---|
| [Microsoft Graph group PATCH behaviour](topics/graph-group-patch-behaviour.md) | `isAssignableToRole` is create-only (400 on PATCH); a successful PATCH returns 204 with no body, and the next read can be stale | graph, groups, task-1032, task-6a42 |
| [Upstreaming fork fixes to deploymenttheory](topics/upstreaming-fork-fixes.md) | Branch and title conventions, shipping `sg` releases (release line, sweetgreen commit identity), running CI in the fork, upstream PR log | upstream, process, ci, release, task-1032, task-6a42 |
| [Reading back after a Graph write](topics/entra-read-after-write-consistency.md) | Entra replicas lag. Poll with a `ConsistencyPredicate` built from the request body, name stale attributes, and test with a stale-read mock | graph, entra, eventual-consistency, testing, task-6a42 |
| [Entra provisioning schema through Graph](topics/graph-synchronization-schema.md) | Job schema is GET/PUT whole-document; parseExpression 404s for jobs, so expressions are parsed locally with `schema/functions` signatures; `version` works as a plan-to-apply concurrency token | graph, entra, provisioning, task-6aec |
