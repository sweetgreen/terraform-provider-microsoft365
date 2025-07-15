// REF: https://learn.microsoft.com/en-us/graph/api/resources/intune-apps-iosmobileappconfiguration?view=graph-rest-1.0
package graphV1IosMobileAppConfiguration

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IosMobileAppConfigurationResourceModel represents the Terraform resource model
type IosMobileAppConfigurationResourceModel struct {
	ID                   types.String                                         `tfsdk:"id"`
	DisplayName          types.String                                         `tfsdk:"display_name"`
	Description          types.String                                         `tfsdk:"description"`
	Version              types.Int64                                          `tfsdk:"version"`
	CreatedDateTime      types.String                                         `tfsdk:"created_date_time"`
	LastModifiedDateTime types.String                                         `tfsdk:"last_modified_date_time"`
	TargetedMobileApps   types.List                                           `tfsdk:"targeted_mobile_apps"`
	EncodedSettingXml    types.String                                         `tfsdk:"encoded_setting_xml"`
	Settings             []AppConfigurationSettingItemModel                   `tfsdk:"settings"`
	Assignments          []ManagedDeviceMobileAppConfigurationAssignmentModel `tfsdk:"assignments"`
	Timeouts             timeouts.Value                                       `tfsdk:"timeouts"`
}

// AppConfigurationSettingItemModel represents an app configuration setting
type AppConfigurationSettingItemModel struct {
	AppConfigKey      types.String `tfsdk:"app_config_key"`
	AppConfigKeyType  types.String `tfsdk:"app_config_key_type"`
	AppConfigKeyValue types.String `tfsdk:"app_config_key_value"`
}

// ManagedDeviceMobileAppConfigurationAssignmentModel represents an assignment
type ManagedDeviceMobileAppConfigurationAssignmentModel struct {
	ID     types.String                                `tfsdk:"id"`
	Target DeviceAndAppManagementAssignmentTargetModel `tfsdk:"target"`
}

// DeviceAndAppManagementAssignmentTargetModel represents the assignment target
type DeviceAndAppManagementAssignmentTargetModel struct {
	ODataType types.String `tfsdk:"odata_type"`
	GroupId   types.String `tfsdk:"group_id"`
}
