package graphBetaMacOSPlatformScript

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ModifyPlan handles plan modification for diff suppression
func (r *MacOSPlatformScriptResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state MacOSPlatformScriptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Handle assignments consistency
	// If assignments were configured but API returns empty, maintain the configured structure
	if !req.Config.Raw.IsNull() {
		var config MacOSPlatformScriptResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if !resp.Diagnostics.HasError() && config.Assignments != nil && state.Assignments == nil {
			// If config has assignments but state is nil (empty from API),
			// use the config assignments to maintain consistency
			plan.Assignments = config.Assignments
			resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		}
	}

	tflog.Debug(ctx, "Completed ModifyPlan for macOS platform script")
}
