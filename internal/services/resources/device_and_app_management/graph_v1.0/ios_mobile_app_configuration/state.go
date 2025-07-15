package graphV1IosMobileAppConfiguration

import (
	"context"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/convert"
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
	if iosConfig.GetEncodedSettingXml() != nil {
		data.EncodedSettingXml = types.StringValue(string(iosConfig.GetEncodedSettingXml()))
	} else {
		data.EncodedSettingXml = types.StringNull()
	}

	// Map settings
	if iosConfig.GetSettings() != nil && len(iosConfig.GetSettings()) > 0 {
		data.Settings = make([]AppConfigurationSettingItemModel, 0, len(iosConfig.GetSettings()))
		for _, setting := range iosConfig.GetSettings() {
			settingModel := AppConfigurationSettingItemModel{
				AppConfigKey:      convert.GraphToFrameworkString(setting.GetAppConfigKey()),
				AppConfigKeyValue: convert.GraphToFrameworkString(setting.GetAppConfigKeyValue()),
			}

			// Map key type
			if setting.GetAppConfigKeyType() != nil {
				keyType := *setting.GetAppConfigKeyType()
				switch keyType {
				case graphmodels.STRINGTYPE_MDMAPPCONFIGKEYTYPE:
					settingModel.AppConfigKeyType = types.StringValue("stringType")
				case graphmodels.INTEGERTYPE_MDMAPPCONFIGKEYTYPE:
					settingModel.AppConfigKeyType = types.StringValue("integerType")
				case graphmodels.REALTYPE_MDMAPPCONFIGKEYTYPE:
					settingModel.AppConfigKeyType = types.StringValue("realType")
				case graphmodels.BOOLEANTYPE_MDMAPPCONFIGKEYTYPE:
					settingModel.AppConfigKeyType = types.StringValue("booleanType")
				case graphmodels.TOKENTYPE_MDMAPPCONFIGKEYTYPE:
					settingModel.AppConfigKeyType = types.StringValue("tokenType")
				default:
					settingModel.AppConfigKeyType = types.StringNull()
				}
			} else {
				settingModel.AppConfigKeyType = types.StringNull()
			}

			data.Settings = append(data.Settings, settingModel)
		}
	} else {
		data.Settings = []AppConfigurationSettingItemModel{}
	}

	tflog.Debug(ctx, "Finished mapping remote iOS mobile app configuration to Terraform state")
}

// MapRemoteAssignmentsToTerraform maps remote assignments to Terraform state
func MapRemoteAssignmentsToTerraform(ctx context.Context, data *IosMobileAppConfigurationResourceModel, assignments []graphmodels.ManagedDeviceMobileAppConfigurationAssignmentable) {
	if assignments == nil || len(assignments) == 0 {
		data.Assignments = []ManagedDeviceMobileAppConfigurationAssignmentModel{}
		return
	}

	tflog.Debug(ctx, "Starting to map remote assignments to Terraform state")

	data.Assignments = make([]ManagedDeviceMobileAppConfigurationAssignmentModel, 0, len(assignments))
	for _, assignment := range assignments {
		assignmentModel := ManagedDeviceMobileAppConfigurationAssignmentModel{
			ID: convert.GraphToFrameworkString(assignment.GetId()),
		}

		// Map target
		if assignment.GetTarget() != nil {
			target := assignment.GetTarget()
			targetModel := DeviceAndAppManagementAssignmentTargetModel{}

			// Determine the type and map accordingly
			switch target.(type) {
			case graphmodels.AllLicensedUsersAssignmentTargetable:
				targetModel.ODataType = types.StringValue("#microsoft.graph.allLicensedUsersAssignmentTarget")

			case graphmodels.AllDevicesAssignmentTargetable:
				targetModel.ODataType = types.StringValue("#microsoft.graph.allDevicesAssignmentTarget")

			case graphmodels.GroupAssignmentTargetable:
				targetModel.ODataType = types.StringValue("#microsoft.graph.groupAssignmentTarget")
				groupTarget := target.(graphmodels.GroupAssignmentTargetable)
				targetModel.GroupId = convert.GraphToFrameworkString(groupTarget.GetGroupId())
				// Note: Filter settings are not available in v1.0 API

			case graphmodels.ExclusionGroupAssignmentTargetable:
				targetModel.ODataType = types.StringValue("#microsoft.graph.exclusionGroupAssignmentTarget")
				exclusionTarget := target.(graphmodels.ExclusionGroupAssignmentTargetable)
				targetModel.GroupId = convert.GraphToFrameworkString(exclusionTarget.GetGroupId())
				// Note: Filter settings are not available in v1.0 API
			}

			assignmentModel.Target = targetModel
		}

		data.Assignments = append(data.Assignments, assignmentModel)
	}

	tflog.Debug(ctx, "Finished mapping remote assignments to Terraform state")
}
