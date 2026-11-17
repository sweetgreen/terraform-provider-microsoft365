---
title: Entra provisioning schema (synchronization jobs) through Graph
tags: [graph, entra, provisioning, synchronization, task-6aec]
updated: 2026-11-17
---

# Entra provisioning schema through Graph

`microsoft365_graph_beta_applications_service_principal_synchronization_job_schema`
(fork branch `sg/synchronization-job-schema`, first release `v1.3.0-sg.5`) manages the
attribute mappings of a provisioning job. azuread has no equivalent
(hashicorp/terraform-provider-azuread#1332). This page records what Graph actually does,
checked against the production ServiceNow job on 2026-11-17.

## Endpoints

| Call | Result |
|---|---|
| `GET /servicePrincipals/{sp}/synchronization/jobs` | Job IDs look like `ServiceNowOutDelta.<tenant id, no dashes>.<app id>`; `templateId` is the part before the first dot. |
| `GET .../jobs/{job}/schema` | Whole document (~140 KB for ServiceNow): `directories` (attribute definitions, large), `synchronizationRules[]` (each with `objectMappings[]`), `version`, `id`, plus `@odata.context` and `directories@odata.context`. |
| `PUT .../jobs/{job}/schema` | Replaces the whole document; answers 204. So the write is read-modify-write. The provider strips `@odata.*` keys and sends everything else back. |
| `GET .../jobs/{job}/schema/functions` | 59 expression functions with ordered parameters (`name`, `required`, `allowMultipleOccurrences`). Works. |
| `POST .../jobs/{job}/schema/parseExpression` | **404** from syncfabric ("No HTTP resource was found") on both beta and v1.0. The template path answers 401. Don't depend on it. |

`version` is a string like `Date:2026-03-25T23:42:02.1956460Z, ActivityId:<guid>`. It only
changes when the schema changes (ServiceNow's still showed the 2026-03-25 rollback in
November), so it works as an optimistic-concurrency token between plan and apply.

## Shape of a mapping

- `objectMappings[]`: `name` (null on disabled template mappings), `enabled`, `flowTypes`
  (a comma-separated string, `"Add,Update,Delete"`), `sourceObjectName` (`""` on disabled
  template mappings), `targetObjectName`, `scope` (filter or null), `metadata[]`,
  `attributeMappings[]`.
- `attributeMappings[]`: `targetAttributeName`, `source`, `defaultValue` (null and `""` both
  occur and are kept distinct), `flowType` (`Always`), `flowBehavior` (`FlowWhenChanged`),
  `matchingPriority` (0, or 1 for the matching attribute), `exportMissingReferences`.
- `source` is a tree: `{expression, name, type, parameters[{key, value}]}`. `type` is
  `Attribute` (`[mail]`, name `mail`), `Constant` (`"1"`, name `1`) or `Function`.
  Leaves have `parameters: []`. A mapping with only a default value has `source: null`.

## Building `source` from an expression

Because parseExpression is unavailable, the provider parses expressions itself
(`expression.go`):

- Function arguments map to the function's parameters by position. A parameter with
  `allowMultipleOccurrences` takes every remaining argument, so `Switch` gives
  `source, defaultValue, switchValue, switchValue, ...` and `Join` gives
  `separator, source, source, ...`.
- An empty argument (`Switch([IsSoftDeleted], , "False", ...)`, the omitted default) is left
  out of `parameters` but still uses up its position.
- Graph's stored form separates arguments with `", "` and writes an empty argument as
  nothing, giving `", ,"`. Inside a constant, `\"` is a quote and any other backslash is literal.
- Checked: every one of the 20 sources in the production ServiceNow schema rebuilds
  byte for byte from its expression, and reading then writing the unchanged 142 KB document
  gives back exactly what Graph returned, minus annotations.

## Permissions

`Synchronization.ReadWrite.All` (application) covers read and PUT. The Atlantis identity
("Entra + Intune Infrastructure as Code Integration", app `68224ed3-…`) did not have it on
2026-11-17. terraform-infrastructure's `scripts/manage-entra-entitlement-permission.sh
--permission Synchronization.ReadWrite.All` checks for it, and with `--apply` (Privileged Role
Administrator) grants it.

## Testing without a tenant write

Only the production tenant's credentials are available to agents, so nothing was PUT live
before release. The parser and document merge (`expression.go`, `schema_document.go`) use only
the standard library, so you can copy them with their tests into a scratch module and run
`go test` against a schema fetched with `az rest` without compiling the Graph SDK. The first
real write is the Atlantis apply after an import with a no-change plan.
