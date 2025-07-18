package graphBetaAssignmentFilter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToTerraform maps the base properties of an AssignmentFilterResourceModel to a Terraform state.
func MapRemoteStateToTerraform(ctx context.Context, data *AssignmentFilterResourceModel, remoteResource graphmodels.DeviceAndAppManagementAssignmentFilterable) {
	if remoteResource == nil {
		tflog.Debug(ctx, "Remote resource is nil")
		return
	}

	tflog.Debug(ctx, "Starting to map remote state to Terraform state", map[string]interface{}{
		"resourceId": convert.GraphToFrameworkString(remoteResource.GetId()),
	})

	data.ID = convert.GraphToFrameworkString(remoteResource.GetId())
	data.DisplayName = convert.GraphToFrameworkString(remoteResource.GetDisplayName())
	data.Description = convert.GraphToFrameworkString(remoteResource.GetDescription())
	data.Platform = convert.GraphToFrameworkEnum(remoteResource.GetPlatform())
	data.Rule = convert.GraphToFrameworkString(remoteResource.GetRule())
	data.AssignmentFilterManagementType = convert.GraphToFrameworkEnum(remoteResource.GetAssignmentFilterManagementType())
	data.CreatedDateTime = convert.GraphToFrameworkTime(remoteResource.GetCreatedDateTime())
	data.LastModifiedDateTime = convert.GraphToFrameworkTime(remoteResource.GetLastModifiedDateTime())
	data.RoleScopeTags = convert.GraphToFrameworkStringSet(ctx, remoteResource.GetRoleScopeTags())

	tflog.Debug(ctx, fmt.Sprintf("Finished stating resource %s with id %s", ResourceName, data.ID.ValueString()))

}
