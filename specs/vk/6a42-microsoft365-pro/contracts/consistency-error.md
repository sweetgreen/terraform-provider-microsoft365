# Contract: error when reads never catch up

Summary (unchanged): `Error reading resource state after update` (or `after create`).

Detail when at least one attribute is stale:

```
Graph accepted the update of microsoft365_graph_beta_groups_group <id>, but reads still return
a different value for: membership_rule (sent "<new>", read "<old>"). The change may still be
replicating; run terraform plan to check whether it has landed. (<ReadWithRetry error>)

Several stale attributes are joined with "; ". The create path says "the create of".
```

When no attribute was recorded as stale (for example, every read failed), the detail keeps
today's text: `Could not read resource state: <ResourceName>: <err>`.
