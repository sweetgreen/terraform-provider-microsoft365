package graphBetaGroup

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

func isAssignableTestModel() *GroupResourceModel {
	return &GroupResourceModel{
		DisplayName:        types.StringValue("Test Group"),
		MailNickname:       types.StringValue("test-group"),
		MailEnabled:        types.BoolValue(false),
		SecurityEnabled:    types.BoolValue(true),
		GroupTypes:         types.SetNull(types.StringType),
		IsAssignableToRole: types.BoolValue(false),
		GroupOwners:        types.SetNull(types.StringType),
		GroupMembers:       types.SetNull(types.StringType),
	}
}

// Graph rejects a PATCH that includes isAssignableToRole (400, validation), even when the value is
// unchanged, so updates must leave it out.
func TestConstructResource_UpdateOmitsIsAssignableToRole(t *testing.T) {
	group, err := constructResource(context.Background(), isAssignableTestModel(), true)
	if err != nil {
		t.Fatalf("constructResource returned an error: %v", err)
	}
	if group.GetIsAssignableToRole() != nil {
		t.Fatalf("update request included isAssignableToRole=%v; it must be omitted", *group.GetIsAssignableToRole())
	}
}

func TestConstructResource_CreateSendsIsAssignableToRole(t *testing.T) {
	model := isAssignableTestModel()
	model.IsAssignableToRole = types.BoolValue(true)
	group, err := constructResource(context.Background(), model, false)
	if err != nil {
		t.Fatalf("constructResource returned an error: %v", err)
	}
	if group.GetIsAssignableToRole() == nil || !*group.GetIsAssignableToRole() {
		t.Fatalf("create request must send isAssignableToRole=true")
	}
}

func TestMapRemoteState_NullIsAssignableToRoleReadsAsFalse(t *testing.T) {
	remote := graphmodels.NewGroup()
	id := "00000000-0000-0000-0000-000000000001"
	remote.SetId(&id)
	data := &GroupResourceModel{}
	MapRemoteStateToTerraform(context.Background(), data, remote)
	if data.IsAssignableToRole.IsNull() || data.IsAssignableToRole.ValueBool() {
		t.Fatalf("null isAssignableToRole must map to false, got %v", data.IsAssignableToRole)
	}

	assignable := true
	remote.SetIsAssignableToRole(&assignable)
	MapRemoteStateToTerraform(context.Background(), data, remote)
	if !data.IsAssignableToRole.ValueBool() {
		t.Fatalf("true isAssignableToRole must map to true")
	}
}
