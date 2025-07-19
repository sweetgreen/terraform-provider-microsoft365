package graphV1IosMobileAppConfiguration

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IosMobileAppConfigurationDataSourceModel represents the Terraform data source model
type IosMobileAppConfigurationDataSourceModel struct {
	FilterType                 types.String                         `tfsdk:"filter_type"`
	FilterValue                types.String                         `tfsdk:"filter_value"`
	IosMobileAppConfigurations []IosMobileAppConfigurationItemModel `tfsdk:"ios_mobile_app_configurations"`
	Timeouts                   timeouts.Value                       `tfsdk:"timeouts"`
}

// IosMobileAppConfigurationItemModel represents an individual iOS mobile app configuration
type IosMobileAppConfigurationItemModel struct {
	ID                   types.String                                                   `tfsdk:"id"`
	DisplayName          types.String                                                   `tfsdk:"display_name"`
	Description          types.String                                                   `tfsdk:"description"`
	Version              types.Int64                                                    `tfsdk:"version"`
	CreatedDateTime      types.String                                                   `tfsdk:"created_date_time"`
	LastModifiedDateTime types.String                                                   `tfsdk:"last_modified_date_time"`
	TargetedMobileApps   types.List                                                     `tfsdk:"targeted_mobile_apps"`
	EncodedSettingXml    types.String                                                   `tfsdk:"encoded_setting_xml"`
	Settings             []AppConfigurationSettingItemDataSourceModel                   `tfsdk:"settings"`
	Assignments          []ManagedDeviceMobileAppConfigurationAssignmentDataSourceModel `tfsdk:"assignments"`
}

// AppConfigurationSettingItemDataSourceModel represents an app configuration setting in the data source
type AppConfigurationSettingItemDataSourceModel struct {
	AppConfigKey      types.String `tfsdk:"app_config_key"`
	AppConfigKeyType  types.String `tfsdk:"app_config_key_type"`
	AppConfigKeyValue types.String `tfsdk:"app_config_key_value"`
}

// ManagedDeviceMobileAppConfigurationAssignmentDataSourceModel represents an assignment in the data source
type ManagedDeviceMobileAppConfigurationAssignmentDataSourceModel struct {
	ID     types.String                                          `tfsdk:"id"`
	Target DeviceAndAppManagementAssignmentTargetDataSourceModel `tfsdk:"target"`
}

// DeviceAndAppManagementAssignmentTargetDataSourceModel represents the assignment target in the data source
type DeviceAndAppManagementAssignmentTargetDataSourceModel struct {
	ODataType types.String `tfsdk:"odata_type"`
	GroupId   types.String `tfsdk:"group_id"`
}
