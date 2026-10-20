# Contract: group consistency tracker (`predicate.go`)

```go
// newGroupConsistency captures what a create or update sent, normalised as Read would map it.
func newGroupConsistency(ctx context.Context, sent graphmodels.Groupable) *groupConsistency

// predicate is passed as crud.ReadWithRetryOptions.ConsistencyPredicate.
// It returns true when the read state has a non-empty id and every sent attribute matches.
// Every call replaces the recorded mismatches.
func (c *groupConsistency) predicate(ctx context.Context, state tfsdk.State) bool

// staleAttributes lists the mismatches from the last predicate call, in schema order, as
// `<attribute> (sent <v>, read <v>)`. It is empty if the last read matched or none ran.
func (c *groupConsistency) staleAttributes() []string
```

Checked attributes, when the request body's getter is non-nil (or, for `groupTypes`, non-empty):
`display_name`, `description`, `mail_nickname`, `mail_enabled`, `security_enabled`,
`group_types`, `visibility`, `is_assignable_to_role`, `membership_rule`,
`membership_rule_processing_state`.

Values are rendered with `attr.Value.String()`, so strings are quoted and sets print as
`["A","B"]`.
