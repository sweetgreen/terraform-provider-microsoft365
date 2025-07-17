package graphBetaWindowsPlatformScript

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WindowsPlatformScriptDataSourceModel represents the Terraform data source model
type WindowsPlatformScriptDataSourceModel struct {
	FilterType             types.String                     `tfsdk:"filter_type"`
	FilterValue            types.String                     `tfsdk:"filter_value"`
	WindowsPlatformScripts []WindowsPlatformScriptItemModel `tfsdk:"windows_platform_scripts"`
	Timeouts               timeouts.Value                   `tfsdk:"timeouts"`
}

// WindowsPlatformScriptItemModel represents an individual Windows platform script
type WindowsPlatformScriptItemModel struct {
	ID                    types.String                                      `tfsdk:"id"`
	DisplayName           types.String                                      `tfsdk:"display_name"`
	Description           types.String                                      `tfsdk:"description"`
	ScriptContent         types.String                                      `tfsdk:"script_content"`
	RunAsAccount          types.String                                      `tfsdk:"run_as_account"`
	EnforceSignatureCheck types.Bool                                        `tfsdk:"enforce_signature_check"`
	FileName              types.String                                      `tfsdk:"file_name"`
	RoleScopeTagIds       types.Set                                         `tfsdk:"role_scope_tag_ids"`
	RunAs32Bit            types.Bool                                        `tfsdk:"run_as_32_bit"`
	Assignments           []DeviceManagementScriptAssignmentDataSourceModel `tfsdk:"assignments"`
}

// DeviceManagementScriptAssignmentDataSourceModel represents an assignment in the data source
type DeviceManagementScriptAssignmentDataSourceModel struct {
	ID     types.String                                          `tfsdk:"id"`
	Target DeviceAndAppManagementAssignmentTargetDataSourceModel `tfsdk:"target"`
}

// DeviceAndAppManagementAssignmentTargetDataSourceModel represents the assignment target in the data source
type DeviceAndAppManagementAssignmentTargetDataSourceModel struct {
	CollectionId                               types.String `tfsdk:"collection_id"`
	DeviceAndAppManagementAssignmentFilterId   types.String `tfsdk:"device_and_app_management_assignment_filter_id"`
	DeviceAndAppManagementAssignmentFilterType types.String `tfsdk:"device_and_app_management_assignment_filter_type"`
	GroupId                                    types.String `tfsdk:"group_id"`
}
