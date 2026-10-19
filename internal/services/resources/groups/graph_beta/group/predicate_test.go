package graphBetaGroup

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

// consistencyTestModel is a dynamic security group as Read maps it.
func consistencyTestModel(rule string, groupTypes ...string) *GroupResourceModel {
	m := &GroupResourceModel{
		ID:                            types.StringValue("00000000-0000-0000-0000-000000000001"),
		DisplayName:                   types.StringValue("POS Ring 1"),
		Description:                   types.StringNull(),
		MailNickname:                  types.StringValue("pos-ring-1"),
		MailEnabled:                   types.BoolValue(false),
		SecurityEnabled:               types.BoolValue(true),
		GroupTypes:                    types.SetNull(types.StringType),
		Visibility:                    types.StringValue("Private"),
		IsAssignableToRole:            types.BoolValue(false),
		MembershipRule:                types.StringValue(rule),
		MembershipRuleProcessingState: types.StringValue("On"),
		GroupOwners:                   types.SetNull(types.StringType),
		GroupMembers:                  types.SetNull(types.StringType),
	}
	if len(groupTypes) > 0 {
		m.GroupTypes, _ = types.SetValueFrom(context.Background(), types.StringType, groupTypes)
	}
	return m
}

// sentUpdate builds the PATCH body the provider would send for the model.
func sentUpdate(t *testing.T, m *GroupResourceModel) graphmodels.Groupable {
	t.Helper()
	body, err := constructResource(context.Background(), m, true)
	if err != nil {
		t.Fatalf("constructResource returned an error: %v", err)
	}
	return body
}

func TestGroupConsistency_StaleMembershipRuleIsNamed(t *testing.T) {
	c := newGroupConsistency(context.Background(), sentUpdate(t, consistencyTestModel("(device.displayName -eq \"new\")", "DynamicMembership")))

	if c.matches(consistencyTestModel("(device.displayName -eq \"old\")", "DynamicMembership")) {
		t.Fatal("a read with the old membership_rule must not satisfy the predicate")
	}
	stale := strings.Join(c.staleAttributes(), "; ")
	for _, want := range []string{"membership_rule", `sent "(device.displayName -eq \"new\")"`, `read "(device.displayName -eq \"old\")"`} {
		if !strings.Contains(stale, want) {
			t.Errorf("stale attributes %q do not contain %q", stale, want)
		}
	}
	if strings.Contains(stale, "display_name") || strings.Contains(stale, "group_types") {
		t.Errorf("only membership_rule differs, got %q", stale)
	}

	if !c.matches(consistencyTestModel("(device.displayName -eq \"new\")", "DynamicMembership")) {
		t.Fatalf("a read with the new values must satisfy the predicate, stale: %v", c.staleAttributes())
	}
	if len(c.staleAttributes()) != 0 {
		t.Fatalf("a matching read must clear the stale attributes, got %v", c.staleAttributes())
	}
}

func TestGroupConsistency_GroupTypesCompareAsASet(t *testing.T) {
	c := newGroupConsistency(context.Background(), sentUpdate(t, consistencyTestModel("rule", "Unified", "DynamicMembership")))
	if !c.matches(consistencyTestModel("rule", "DynamicMembership", "Unified")) {
		t.Fatalf("group_types in a different order must match, stale: %v", c.staleAttributes())
	}
}

func TestGroupConsistency_IgnoresAttributesThatWereNotSent(t *testing.T) {
	plan := consistencyTestModel("rule")
	// Optional+Computed attributes left out of config are unknown in the plan and never sent.
	plan.Description = types.StringUnknown()
	plan.GroupTypes = types.SetUnknown(types.StringType)
	plan.MembershipRuleProcessingState = types.StringUnknown()
	c := newGroupConsistency(context.Background(), sentUpdate(t, plan))

	read := consistencyTestModel("rule", "DynamicMembership")
	read.Description = types.StringValue("set in the portal")
	read.MembershipRuleProcessingState = types.StringValue("Paused")
	// isAssignableToRole is never sent on update, so a role-assignable group still matches.
	read.IsAssignableToRole = types.BoolValue(true)
	if !c.matches(read) {
		t.Fatalf("attributes that were not sent must not be compared, stale: %v", c.staleAttributes())
	}
}

func TestGroupConsistency_RequiresAnID(t *testing.T) {
	c := newGroupConsistency(context.Background(), sentUpdate(t, consistencyTestModel("rule")))
	read := consistencyTestModel("rule")
	read.ID = types.StringNull()
	if c.matches(read) {
		t.Fatal("a read without an ID must not satisfy the predicate")
	}
}

func TestConsistencyErrorDetail(t *testing.T) {
	err := context.DeadlineExceeded
	named := consistencyErrorDetail("update", "group-id", err, []string{`membership_rule (sent "a", read "b")`})
	for _, want := range []string{"Graph accepted the update", "group-id", `membership_rule (sent "a", read "b")`} {
		if !strings.Contains(named, want) {
			t.Errorf("detail %q does not contain %q", named, want)
		}
	}
	if plain := consistencyErrorDetail("update", "group-id", err, nil); !strings.HasPrefix(plain, "Could not read resource state") {
		t.Errorf("without stale attributes the detail must keep the read error, got %q", plain)
	}
}
