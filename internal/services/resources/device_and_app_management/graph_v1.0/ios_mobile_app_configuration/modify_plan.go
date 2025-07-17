package graphV1IosMobileAppConfiguration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ModifyPlan handles plan modifications for the resource
func (r *IosMobileAppConfigurationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// No plan modifications needed for this resource currently
	// This method is here to satisfy the ResourceWithModifyPlan interface
}
