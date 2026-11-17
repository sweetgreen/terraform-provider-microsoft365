package graphBetaGroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

// groupConsistencyAttribute is a group attribute that a create or update can send, with how to
// tell whether the request sent it and where its value lives in the model.
type groupConsistencyAttribute struct {
	name  string
	sent  func(graphmodels.Groupable) bool
	value func(*GroupResourceModel) attr.Value
}

// groupConsistencyAttributes lists, in schema order, the attributes checked after a write.
var groupConsistencyAttributes = []groupConsistencyAttribute{
	{"display_name", func(g graphmodels.Groupable) bool { return g.GetDisplayName() != nil }, func(m *GroupResourceModel) attr.Value { return m.DisplayName }},
	{"description", func(g graphmodels.Groupable) bool { return g.GetDescription() != nil }, func(m *GroupResourceModel) attr.Value { return m.Description }},
	{"mail_nickname", func(g graphmodels.Groupable) bool { return g.GetMailNickname() != nil }, func(m *GroupResourceModel) attr.Value { return m.MailNickname }},
	{"mail_enabled", func(g graphmodels.Groupable) bool { return g.GetMailEnabled() != nil }, func(m *GroupResourceModel) attr.Value { return m.MailEnabled }},
	{"security_enabled", func(g graphmodels.Groupable) bool { return g.GetSecurityEnabled() != nil }, func(m *GroupResourceModel) attr.Value { return m.SecurityEnabled }},
	{"group_types", func(g graphmodels.Groupable) bool { return len(g.GetGroupTypes()) > 0 }, func(m *GroupResourceModel) attr.Value { return m.GroupTypes }},
	{"visibility", func(g graphmodels.Groupable) bool { return g.GetVisibility() != nil }, func(m *GroupResourceModel) attr.Value { return m.Visibility }},
	{"is_assignable_to_role", func(g graphmodels.Groupable) bool { return g.GetIsAssignableToRole() != nil }, func(m *GroupResourceModel) attr.Value { return m.IsAssignableToRole }},
	{"membership_rule", func(g graphmodels.Groupable) bool { return g.GetMembershipRule() != nil }, func(m *GroupResourceModel) attr.Value { return m.MembershipRule }},
	{"membership_rule_processing_state", func(g graphmodels.Groupable) bool { return g.GetMembershipRuleProcessingState() != nil }, func(m *GroupResourceModel) attr.Value { return m.MembershipRuleProcessingState }},
}

// groupConsistency verifies that a group write has propagated across Microsoft Entra replicas
// before ReadWithRetry accepts a read as authoritative.
//
// Entra is eventually consistent: the read straight after a successful POST or PATCH can come
// from a replica that still holds the old values. Accepting that read writes stale values to
// state, and Terraform fails with "Provider produced inconsistent result after apply" although
// the write landed. The predicate keeps ReadWithRetry polling until every attribute the request
// sent reads back with the sent value.
//
// See: https://devblogs.microsoft.com/identity/designing-for-eventual-consistency-for-microsoft-entra/
type groupConsistency struct {
	expected GroupResourceModel
	checked  []groupConsistencyAttribute
	stale    []string
}

// newGroupConsistency captures what a create or update sent. The expected values are the request
// body mapped through MapRemoteStateToTerraform, the mapping Read applies, so the predicate only
// waits for differences that Terraform would also reject.
func newGroupConsistency(ctx context.Context, sent graphmodels.Groupable) *groupConsistency {
	c := &groupConsistency{}
	MapRemoteStateToTerraform(ctx, &c.expected, sent)
	for _, a := range groupConsistencyAttributes {
		if a.sent(sent) {
			c.checked = append(c.checked, a)
		}
	}
	return c
}

// predicate is a crud.ReadWithRetryOptions.ConsistencyPredicate. It returns true when the read
// state has an ID and every sent attribute matches, and records the mismatches otherwise.
func (c *groupConsistency) predicate(ctx context.Context, state tfsdk.State) bool {
	var actual GroupResourceModel
	if diags := state.Get(ctx, &actual); diags.HasError() {
		c.stale = nil
		return false
	}
	return c.matches(&actual)
}

// matches compares a read model with what was sent and records the mismatches.
func (c *groupConsistency) matches(actual *GroupResourceModel) bool {
	c.stale = nil

	if actual.ID.IsNull() || actual.ID.IsUnknown() || actual.ID.ValueString() == "" {
		c.stale = append(c.stale, "id (read empty)")
	}

	for _, a := range c.checked {
		want, got := a.value(&c.expected), a.value(actual)
		if !got.Equal(want) {
			c.stale = append(c.stale, fmt.Sprintf("%s (sent %s, read %s)", a.name, want, got))
		}
	}

	return len(c.stale) == 0
}

// staleAttributes returns the mismatches recorded by the last predicate call.
func (c *groupConsistency) staleAttributes() []string {
	return c.stale
}
