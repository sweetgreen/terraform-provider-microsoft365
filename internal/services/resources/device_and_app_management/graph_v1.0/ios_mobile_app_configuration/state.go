package graphV1IosMobileAppConfiguration

import (
	"context"

	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// MapRemoteResourceToTerraform maps the remote iOS mobile app configuration to Terraform state
func MapRemoteResourceToTerraform(ctx context.Context, data *IosMobileAppConfigurationResourceModel, remoteResource graphmodels.ManagedDeviceMobileAppConfigurationable) {
	if remoteResource == nil {
		tflog.Debug(ctx, "Remote resource is nil")
		return
	}

	// Cast to iOS mobile app configuration
	iosConfig, ok := remoteResource.(graphmodels.IosMobileAppConfigurationable)
	if !ok {
		tflog.Warn(ctx, "Remote resource is not an iOS mobile app configuration")
		return
	}

	tflog.Debug(ctx, "Starting to map remote iOS mobile app configuration to Terraform state")

	// Map base properties
	data.ID = convert.GraphToFrameworkString(iosConfig.GetId())
	data.DisplayName = convert.GraphToFrameworkString(iosConfig.GetDisplayName())
	data.Description = convert.GraphToFrameworkString(iosConfig.GetDescription())
	if iosConfig.GetVersion() != nil {
		data.Version = types.Int64Value(int64(*iosConfig.GetVersion()))
	} else {
		data.Version = types.Int64Null()
	}
	data.CreatedDateTime = convert.GraphToFrameworkTime(iosConfig.GetCreatedDateTime())
	data.LastModifiedDateTime = convert.GraphToFrameworkTime(iosConfig.GetLastModifiedDateTime())

	// Map targeted mobile apps
	if iosConfig.GetTargetedMobileApps() != nil {
		data.TargetedMobileApps = convert.GraphToFrameworkStringList(iosConfig.GetTargetedMobileApps())
	} else {
		data.TargetedMobileApps = types.ListNull(types.StringType)
	}

	// Map encoded setting XML
	encodedXml := iosConfig.GetEncodedSettingXml()
	if encodedXml != nil && len(encodedXml) > 0 {
		data.EncodedSettingXml = types.StringValue(string(encodedXml))
	} else {
		data.EncodedSettingXml = types.StringNull()
	}

	// Map settings
	if iosConfig.GetSettings() != nil && len(iosConfig.GetSettings()) > 0 {
		settingsList := make([]attr.Value, 0, len(iosConfig.GetSettings()))
		for _, setting := range iosConfig.GetSettings() {
			// Map key type
			var keyTypeStr string
			if setting.GetAppConfigKeyType() != nil {
				keyType := *setting.GetAppConfigKeyType()
				switch keyType {
				case graphmodels.STRINGTYPE_MDMAPPCONFIGKEYTYPE:
					keyTypeStr = "stringType"
				case graphmodels.INTEGERTYPE_MDMAPPCONFIGKEYTYPE:
					keyTypeStr = "integerType"
				case graphmodels.REALTYPE_MDMAPPCONFIGKEYTYPE:
					keyTypeStr = "realType"
				case graphmodels.BOOLEANTYPE_MDMAPPCONFIGKEYTYPE:
					keyTypeStr = "booleanType"
				case graphmodels.TOKENTYPE_MDMAPPCONFIGKEYTYPE:
					keyTypeStr = "tokenType"
				}
			}

			settingObj, _ := types.ObjectValue(
				map[string]attr.Type{
					"app_config_key":       types.StringType,
					"app_config_key_type":  types.StringType,
					"app_config_key_value": types.StringType,
				},
				map[string]attr.Value{
					"app_config_key":       convert.GraphToFrameworkString(setting.GetAppConfigKey()),
					"app_config_key_type":  types.StringValue(keyTypeStr),
					"app_config_key_value": convert.GraphToFrameworkString(setting.GetAppConfigKeyValue()),
				},
			)
			settingsList = append(settingsList, settingObj)
		}

		settingsListType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"app_config_key":       types.StringType,
				"app_config_key_type":  types.StringType,
				"app_config_key_value": types.StringType,
			},
		}
		data.Settings, _ = types.ListValue(settingsListType, settingsList)
	} else {
		// Set to null when there are no settings from the remote resource
		data.Settings = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"app_config_key":       types.StringType,
				"app_config_key_type":  types.StringType,
				"app_config_key_value": types.StringType,
			},
		})
	}

	tflog.Debug(ctx, "Finished mapping remote iOS mobile app configuration to Terraform state")
}

// MapRemoteAssignmentsToTerraform maps remote assignments to Terraform state
func MapRemoteAssignmentsToTerraform(ctx context.Context, data *IosMobileAppConfigurationResourceModel, assignments []graphmodels.ManagedDeviceMobileAppConfigurationAssignmentable) {
	if assignments == nil || len(assignments) == 0 {
		data.Assignments = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id": types.StringType,
				"target": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"odata_type": types.StringType,
						"group_id":   types.StringType,
					},
				},
			},
		})
		return
	}

	tflog.Debug(ctx, "Starting to map remote assignments to Terraform state")

	assignmentsList := make([]attr.Value, 0, len(assignments))
	for _, assignment := range assignments {
		var targetValue attr.Value

		// Map target
		if assignment.GetTarget() != nil {
			target := assignment.GetTarget()
			var odataType string
			var groupId types.String = types.StringNull()

			// Determine the type and map accordingly
			switch target.(type) {
			case graphmodels.AllLicensedUsersAssignmentTargetable:
				odataType = "#microsoft.graph.allLicensedUsersAssignmentTarget"

			case graphmodels.AllDevicesAssignmentTargetable:
				odataType = "#microsoft.graph.allDevicesAssignmentTarget"

			case graphmodels.GroupAssignmentTargetable:
				odataType = "#microsoft.graph.groupAssignmentTarget"
				groupTarget := target.(graphmodels.GroupAssignmentTargetable)
				groupId = convert.GraphToFrameworkString(groupTarget.GetGroupId())

			case graphmodels.ExclusionGroupAssignmentTargetable:
				odataType = "#microsoft.graph.exclusionGroupAssignmentTarget"
				exclusionTarget := target.(graphmodels.ExclusionGroupAssignmentTargetable)
				groupId = convert.GraphToFrameworkString(exclusionTarget.GetGroupId())
			}

			targetValue, _ = types.ObjectValue(
				map[string]attr.Type{
					"odata_type": types.StringType,
					"group_id":   types.StringType,
				},
				map[string]attr.Value{
					"odata_type": types.StringValue(odataType),
					"group_id":   groupId,
				},
			)
		}

		assignmentObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"id": types.StringType,
				"target": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"odata_type": types.StringType,
						"group_id":   types.StringType,
					},
				},
			},
			map[string]attr.Value{
				"id":     convert.GraphToFrameworkString(assignment.GetId()),
				"target": targetValue,
			},
		)
		assignmentsList = append(assignmentsList, assignmentObj)
	}

	assignmentsListType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id": types.StringType,
			"target": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"odata_type": types.StringType,
					"group_id":   types.StringType,
				},
			},
		},
	}
	data.Assignments, _ = types.ListValue(assignmentsListType, assignmentsList)

	tflog.Debug(ctx, "Finished mapping remote assignments to Terraform state")
}
