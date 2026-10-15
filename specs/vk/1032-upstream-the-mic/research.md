# Research: Graph behaviour for group PATCH

## Evidence (sweetgreen-test tenant, 2026-10-14/15, terraform-infrastructure #1899)

| Probe: `PATCH /beta/groups/{id}` with only… | Result |
|---|---|
| `isAssignableToRole` (`false`, unchanged) | `400 An error occurred while processing the write request` (category: validation) |
| `description` | `204 No Content` |
| `mailNickname` | `204 No Content` |
| `mailEnabled` + `securityEnabled` | `204 No Content` |
| `visibility` | `204 No Content` |
| `groupTypes` | `204 No Content` |

- Microsoft Graph docs for `isAssignableToRole` say it "can only be set while
  creating the group and is immutable". The provider's own schema description
  (`resource.go:213`) says the same.
- A successful group PATCH returns `204 No Content`. With no body, Kiota's
  `Groups().ByGroupId(id).Patch` returns `(nil, nil)`, so `*nil.GetId()`
  panics. In Terraform this shows up as "Plugin did not respond", after the
  write has already landed in Graph.
- Groups that were never role-assignable return `"isAssignableToRole": null`.

## Decisions
- **Omit the property on update instead of sending it only when it changed.**
  Graph rejects it even when the value is unchanged, so there is never a
  valid case for sending it in a PATCH.
- **Map null to `false` on read instead of using a plan modifier.** This
  matches the schema default and Graph's meaning (null means not assignable).
  It is the smallest change that prevents a spurious diff after import.
- **Keep the ID from state instead of doing a GET after the PATCH.** The ID is
  immutable, and `Update` already calls `ReadWithRetry` right afterwards to
  refresh everything else.
- **No `RequiresReplace`.** See the plan's Risks section.

## Alternatives rejected
- Using a separate PATCH model per field: unnecessary churn.
- Adding a new mock for the 204 response: the existing responder in
  `mocks/responders.go` already returns 204.
