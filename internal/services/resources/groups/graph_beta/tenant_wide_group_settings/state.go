package graphBetaTenantWideGroupSettings

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToTerraform maps the base properties of a DirectorySetting resource to a Terraform state.
func MapRemoteStateToTerraform(ctx context.Context, data *TenantWideGroupSettingsResourceModel, remoteResource graphmodels.DirectorySettingable) {
	if remoteResource == nil {
		tflog.Debug(ctx, "Remote resource is nil")
		return
	}

	tflog.Debug(ctx, "Starting to map remote state to Terraform state", map[string]interface{}{
		"resourceId": convert.GraphToFrameworkString(remoteResource.GetId()),
	})

	// Debug log for the raw ID value
	idVal := remoteResource.GetId()
	tflog.Debug(ctx, "MapRemoteStateToTerraform: remoteResource.GetId() value", map[string]interface{}{
		"raw_id": idVal,
	})

	data.ID = convert.GraphToFrameworkString(remoteResource.GetId())
	data.DisplayName = convert.GraphToFrameworkString(remoteResource.GetDisplayName())
	data.TemplateID = convert.GraphToFrameworkString(remoteResource.GetTemplateId())

	// Convert setting values to Terraform set
	if settingValues := remoteResource.GetValues(); settingValues != nil {
		var settingValueObjects []attr.Value

		for _, settingValue := range settingValues {
			if settingValue != nil {
				settingValueMap := map[string]attr.Value{
					"name":  convert.GraphToFrameworkString(settingValue.GetName()),
					"value": convert.GraphToFrameworkString(settingValue.GetValue()),
				}

				settingValueObj, diags := types.ObjectValue(map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				}, settingValueMap)

				if !diags.HasError() {
					settingValueObjects = append(settingValueObjects, settingValueObj)
				}
			}
		}

		if len(settingValueObjects) > 0 {
			setValue, diags := types.SetValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			}, settingValueObjects)

			if !diags.HasError() {
				data.Values = setValue
			}
		} else {
			data.Values = types.SetNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			})
		}
	} else {
		data.Values = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
			},
		})
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished mapping remote state to Terraform state %s with id %s", ResourceName, data.ID.ValueString()))
}
