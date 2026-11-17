package graphBetaGroup

import (
	"context"

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

	// Refuse to plan a new group whose display name is already taken, so the plan fails rather than
	// the apply. Only a create has no prior state; update, refresh and import never check, so an
	// existing group can still be adopted. An unknown name is checked again in Create.
	if req.State.Raw.IsNull() && shouldCheckDuplicateName(&plan) && r.client != nil {
		r.checkDuplicateDisplayName(ctx, plan.DisplayName.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Finished plan modification for group resource")
}

// shouldCheckDuplicateName reports whether the duplicate-name check applies to a planned group: the
// flag is known and true, and the display name is known.
func shouldCheckDuplicateName(plan *GroupResourceModel) bool {
	if plan.PreventDuplicateNames.IsNull() || plan.PreventDuplicateNames.IsUnknown() || !plan.PreventDuplicateNames.ValueBool() {
		return false
	}
	return !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown()
}
