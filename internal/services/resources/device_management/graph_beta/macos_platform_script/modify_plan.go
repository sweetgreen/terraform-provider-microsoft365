package graphBetaMacOSPlatformScript

import (
	"context"

	sharedmodels "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/shared_models/graph_beta/device_management"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ModifyPlan handles plan modification for diff suppression
func (r *MacOSPlatformScriptResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Get the planned and current state
	var plan MacOSPlatformScriptResourceModel
	var state MacOSPlatformScriptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Handle assignment consistency
	// If assignments were configured in the plan but state shows nil, maintain the configured assignments
	if plan.Assignments != nil && state.Assignments == nil {
		tflog.Debug(ctx, "Assignments were configured but state is nil, maintaining configured assignments in plan")
		
		// Ensure the plan maintains the configured assignments structure
		if plan.Assignments.AllDevices.IsNull() {
			plan.Assignments.AllDevices = types.BoolValue(false)
		}
		if plan.Assignments.AllUsers.IsNull() {
			plan.Assignments.AllUsers = types.BoolValue(false)
		}
		
		// Set the modified plan
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}

	// If both plan and state have assignments but state shows all false values and plan has explicit false values,
	// ensure consistency
	if plan.Assignments != nil && state.Assignments != nil {
		planAllDevices := plan.Assignments.AllDevices.ValueBool()
		planAllUsers := plan.Assignments.AllUsers.ValueBool()
		stateAllDevices := state.Assignments.AllDevices.ValueBool()
		stateAllUsers := state.Assignments.AllUsers.ValueBool()
		
		// If plan has explicit false values and state also has false values, maintain consistency
		if !planAllDevices && !planAllUsers && !stateAllDevices && !stateAllUsers {
			// Check if include/exclude groups are empty in both
			planIncludeEmpty := plan.Assignments.IncludeGroupIds == nil || len(plan.Assignments.IncludeGroupIds) == 0
			planExcludeEmpty := plan.Assignments.ExcludeGroupIds == nil || len(plan.Assignments.ExcludeGroupIds) == 0
			stateIncludeEmpty := state.Assignments.IncludeGroupIds == nil || len(state.Assignments.IncludeGroupIds) == 0
			stateExcludeEmpty := state.Assignments.ExcludeGroupIds == nil || len(state.Assignments.ExcludeGroupIds) == 0
			
			if planIncludeEmpty && planExcludeEmpty && stateIncludeEmpty && stateExcludeEmpty {
				tflog.Debug(ctx, "Both plan and state have empty assignments with false values, maintaining consistency")
				
				// Ensure the plan maintains the same structure as configured
				plan.Assignments = &sharedmodels.DeviceManagementScriptAssignmentResourceModel{
					AllDevices: types.BoolValue(false),
					AllUsers:   types.BoolValue(false),
				}
				
				resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
			}
		}
	}

	tflog.Debug(ctx, "ModifyPlan completed for assignment consistency")
}
