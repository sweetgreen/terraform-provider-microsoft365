# Knowledge base

Reusable knowledge from work on Sweetgreen's fork of
terraform-provider-microsoft365. Each topic page lists the task ids that
contributed to it in its `tags`. The registry docs live in `docs/`, which is
generated, so this knowledge base is kept outside it.

## Index

| Topic | Summary | Tags |
|---|---|---|
| [Microsoft Graph group PATCH behaviour](topics/graph-group-patch-behaviour.md) | `isAssignableToRole` is create-only (400 on PATCH); a successful PATCH returns 204 with no body | graph, groups, task-1032 |
| [Upstreaming fork fixes to deploymenttheory](topics/upstreaming-fork-fixes.md) | Branch and title conventions, running CI in the fork instead of a local build, upstream PR log | upstream, process, ci, task-1032 |
