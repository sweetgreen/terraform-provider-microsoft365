package graphBetaGroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ModifyPlan handles plan modification for groups
func (r *GroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Don't validate on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	tflog.Debug(ctx, "Starting plan modification for group resource")

	var plan GroupResourceModel

	// Get the planned configuration
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Perform cross-attribute validation
	resp.Diagnostics.Append(ValidateGroupConfiguration(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// isAssignableToRole can only be set when a group is created and is never sent on update, so
	// reject a change to an existing group at plan time rather than failing after apply.
	if !req.State.Raw.IsNull() {
		var state GroupResourceModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !state.IsAssignableToRole.IsNull() && !state.IsAssignableToRole.IsUnknown() &&
			!plan.IsAssignableToRole.IsNull() && !plan.IsAssignableToRole.IsUnknown() &&
			!state.IsAssignableToRole.Equal(plan.IsAssignableToRole) {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_assignable_to_role"),
				"Cannot change is_assignable_to_role on an existing group",
				fmt.Sprintf("Microsoft Graph only accepts isAssignableToRole when a group is created, and this group "+
					"has is_assignable_to_role = %t. Set it to match, or remove the group from the configuration, "+
					"apply, then add it back to create a new group.", state.IsAssignableToRole.ValueBool()),
			)
			return
		}
	}

	tflog.Debug(ctx, "Finished plan modification for group resource")
}
