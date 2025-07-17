package graphV1IosMobileAppConfiguration

import (
	"context"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/convert"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// MapRemoteConfigurationToDataSource maps a Microsoft Graph iOS mobile app configuration to the data source model
func MapRemoteConfigurationToDataSource(ctx context.Context, data *IosMobileAppConfigurationItemModel, remoteConfig graphmodels.IosMobileAppConfigurationable) {
	if remoteConfig == nil {
		tflog.Debug(ctx, "Remote configuration is nil")
		return
	}

	tflog.Debug(ctx, "Starting to map remote iOS mobile app configuration to data source")

	// Map base properties
	data.ID = convert.GraphToFrameworkString(remoteConfig.GetId())
	data.DisplayName = convert.GraphToFrameworkString(remoteConfig.GetDisplayName())
	data.Description = convert.GraphToFrameworkString(remoteConfig.GetDescription())

	// Map version
	if remoteConfig.GetVersion() != nil {
		data.Version = types.Int64Value(int64(*remoteConfig.GetVersion()))
	} else {
		data.Version = types.Int64Null()
	}

	data.CreatedDateTime = convert.GraphToFrameworkTime(remoteConfig.GetCreatedDateTime())
	data.LastModifiedDateTime = convert.GraphToFrameworkTime(remoteConfig.GetLastModifiedDateTime())

	// Map targeted mobile apps
	if remoteConfig.GetTargetedMobileApps() != nil {
		data.TargetedMobileApps = convert.GraphToFrameworkStringList(remoteConfig.GetTargetedMobileApps())
	} else {
		data.TargetedMobileApps = types.ListNull(types.StringType)
	}

	// Map encoded setting XML
	if remoteConfig.GetEncodedSettingXml() != nil {
		data.EncodedSettingXml = types.StringValue(string(remoteConfig.GetEncodedSettingXml()))
	} else {
		data.EncodedSettingXml = types.StringNull()
	}

	// Map settings
	if remoteConfig.GetSettings() != nil && len(remoteConfig.GetSettings()) > 0 {
		data.Settings = make([]AppConfigurationSettingItemDataSourceModel, 0, len(remoteConfig.GetSettings()))
		for _, setting := range remoteConfig.GetSettings() {
			settingModel := AppConfigurationSettingItemDataSourceModel{
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
		data.Settings = []AppConfigurationSettingItemDataSourceModel{}
	}

	tflog.Debug(ctx, "Finished mapping remote iOS mobile app configuration to data source")
}

// MapRemoteAssignmentsToDataSource maps remote assignments to data source model
func MapRemoteAssignmentsToDataSource(ctx context.Context, data *IosMobileAppConfigurationItemModel, assignments []graphmodels.ManagedDeviceMobileAppConfigurationAssignmentable) {
	if assignments == nil || len(assignments) == 0 {
		data.Assignments = []ManagedDeviceMobileAppConfigurationAssignmentDataSourceModel{}
		return
	}

	tflog.Debug(ctx, "Starting to map remote assignments to data source")

	data.Assignments = make([]ManagedDeviceMobileAppConfigurationAssignmentDataSourceModel, 0, len(assignments))
	for _, assignment := range assignments {
		assignmentModel := ManagedDeviceMobileAppConfigurationAssignmentDataSourceModel{
			ID: convert.GraphToFrameworkString(assignment.GetId()),
		}

		// Map target
		if assignment.GetTarget() != nil {
			target := assignment.GetTarget()
			targetModel := DeviceAndAppManagementAssignmentTargetDataSourceModel{}

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

			case graphmodels.ExclusionGroupAssignmentTargetable:
				targetModel.ODataType = types.StringValue("#microsoft.graph.exclusionGroupAssignmentTarget")
				exclusionTarget := target.(graphmodels.ExclusionGroupAssignmentTargetable)
				targetModel.GroupId = convert.GraphToFrameworkString(exclusionTarget.GetGroupId())
			}

			assignmentModel.Target = targetModel
		}

		data.Assignments = append(data.Assignments, assignmentModel)
	}

	tflog.Debug(ctx, "Finished mapping remote assignments to data source")
}
