package graphBetaWindowsPlatformScript

import (
	"context"
	"encoding/base64"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

// MapRemoteScriptToDataSource maps a DeviceManagementScriptable to WindowsPlatformScriptItemModel
func MapRemoteScriptToDataSource(ctx context.Context, data *WindowsPlatformScriptItemModel, script models.DeviceManagementScriptable) {
	if script == nil || data == nil {
		return
	}

	tflog.Debug(ctx, "Mapping Windows platform script from API to data source model", map[string]interface{}{
		"script_id": script.GetId(),
	})

	// Map basic properties
	if script.GetId() != nil {
		data.ID = types.StringValue(*script.GetId())
	} else {
		data.ID = types.StringNull()
	}

	if script.GetDisplayName() != nil {
		data.DisplayName = types.StringValue(*script.GetDisplayName())
	} else {
		data.DisplayName = types.StringNull()
	}

	if script.GetDescription() != nil {
		data.Description = types.StringValue(*script.GetDescription())
	} else {
		data.Description = types.StringNull()
	}

	// Handle script content - decode from base64
	if script.GetScriptContent() != nil {
		decodedContent, err := base64.StdEncoding.DecodeString(string(script.GetScriptContent()))
		if err != nil {
			tflog.Warn(ctx, "Failed to decode script content", map[string]interface{}{
				"error": err.Error(),
			})
			data.ScriptContent = types.StringNull()
		} else {
			data.ScriptContent = types.StringValue(string(decodedContent))
		}
	} else {
		data.ScriptContent = types.StringNull()
	}

	// Map run as account
	if script.GetRunAsAccount() != nil {
		data.RunAsAccount = types.StringValue(script.GetRunAsAccount().String())
	} else {
		data.RunAsAccount = types.StringNull()
	}

	// Map boolean properties
	if script.GetEnforceSignatureCheck() != nil {
		data.EnforceSignatureCheck = types.BoolValue(*script.GetEnforceSignatureCheck())
	} else {
		data.EnforceSignatureCheck = types.BoolNull()
	}

	if script.GetFileName() != nil {
		data.FileName = types.StringValue(*script.GetFileName())
	} else {
		data.FileName = types.StringNull()
	}

	// Map role scope tag IDs
	if script.GetRoleScopeTagIds() != nil && len(script.GetRoleScopeTagIds()) > 0 {
		roleScopeTagIds := make([]attr.Value, len(script.GetRoleScopeTagIds()))
		for i, tagId := range script.GetRoleScopeTagIds() {
			roleScopeTagIds[i] = types.StringValue(tagId)
		}
		data.RoleScopeTagIds = types.SetValueMust(types.StringType, roleScopeTagIds)
	} else {
		data.RoleScopeTagIds = types.SetNull(types.StringType)
	}

	if script.GetRunAs32Bit() != nil {
		data.RunAs32Bit = types.BoolValue(*script.GetRunAs32Bit())
	} else {
		data.RunAs32Bit = types.BoolNull()
	}

	tflog.Debug(ctx, "Successfully mapped Windows platform script", map[string]interface{}{
		"script_id":    data.ID.ValueString(),
		"display_name": data.DisplayName.ValueString(),
	})
}

// MapRemoteAssignmentsToDataSource maps DeviceManagementScriptAssignments to the data source model
func MapRemoteAssignmentsToDataSource(ctx context.Context, data *WindowsPlatformScriptItemModel, assignments []models.DeviceManagementScriptAssignmentable) {
	if assignments == nil || len(assignments) == 0 {
		data.Assignments = []DeviceManagementScriptAssignmentDataSourceModel{}
		return
	}

	tflog.Debug(ctx, "Mapping assignments from API to data source model", map[string]interface{}{
		"count": len(assignments),
	})

	assignmentModels := make([]DeviceManagementScriptAssignmentDataSourceModel, 0, len(assignments))
	for _, assignment := range assignments {
		if assignment == nil {
			continue
		}

		assignmentModel := DeviceManagementScriptAssignmentDataSourceModel{}

		// Map assignment ID
		if assignment.GetId() != nil {
			assignmentModel.ID = types.StringValue(*assignment.GetId())
		} else {
			assignmentModel.ID = types.StringNull()
		}

		// Map target
		if assignment.GetTarget() != nil {
			target := assignment.GetTarget()
			targetModel := DeviceAndAppManagementAssignmentTargetDataSourceModel{}

			// Map collection ID
			if collectionTarget, ok := target.(models.ConfigurationManagerCollectionAssignmentTargetable); ok && collectionTarget.GetCollectionId() != nil {
				targetModel.CollectionId = types.StringValue(*collectionTarget.GetCollectionId())
			} else {
				targetModel.CollectionId = types.StringNull()
			}

			// Map group-based assignments
			switch typedTarget := target.(type) {
			case models.GroupAssignmentTargetable:
				if typedTarget.GetGroupId() != nil {
					targetModel.GroupId = types.StringValue(*typedTarget.GetGroupId())
				} else {
					targetModel.GroupId = types.StringNull()
				}

				// Map filter properties
				if typedTarget.GetDeviceAndAppManagementAssignmentFilterId() != nil {
					targetModel.DeviceAndAppManagementAssignmentFilterId = types.StringValue(*typedTarget.GetDeviceAndAppManagementAssignmentFilterId())
				} else {
					targetModel.DeviceAndAppManagementAssignmentFilterId = types.StringNull()
				}

				if typedTarget.GetDeviceAndAppManagementAssignmentFilterType() != nil {
					targetModel.DeviceAndAppManagementAssignmentFilterType = types.StringValue(typedTarget.GetDeviceAndAppManagementAssignmentFilterType().String())
				} else {
					targetModel.DeviceAndAppManagementAssignmentFilterType = types.StringNull()
				}

			case models.ExclusionGroupAssignmentTargetable:
				if typedTarget.GetGroupId() != nil {
					targetModel.GroupId = types.StringValue(*typedTarget.GetGroupId())
				} else {
					targetModel.GroupId = types.StringNull()
				}

				// Map filter properties
				if typedTarget.GetDeviceAndAppManagementAssignmentFilterId() != nil {
					targetModel.DeviceAndAppManagementAssignmentFilterId = types.StringValue(*typedTarget.GetDeviceAndAppManagementAssignmentFilterId())
				} else {
					targetModel.DeviceAndAppManagementAssignmentFilterId = types.StringNull()
				}

				if typedTarget.GetDeviceAndAppManagementAssignmentFilterType() != nil {
					targetModel.DeviceAndAppManagementAssignmentFilterType = types.StringValue(typedTarget.GetDeviceAndAppManagementAssignmentFilterType().String())
				} else {
					targetModel.DeviceAndAppManagementAssignmentFilterType = types.StringNull()
				}

			default:
				// For other target types, set all fields to null
				targetModel.GroupId = types.StringNull()
				targetModel.DeviceAndAppManagementAssignmentFilterId = types.StringNull()
				targetModel.DeviceAndAppManagementAssignmentFilterType = types.StringNull()
			}

			assignmentModel.Target = targetModel
		}

		assignmentModels = append(assignmentModels, assignmentModel)
	}

	data.Assignments = assignmentModels

	tflog.Debug(ctx, "Successfully mapped assignments", map[string]interface{}{
		"mapped_count": len(assignmentModels),
	})
}
