package graphBetaWindowsAutopilotDeploymentProfileAssignment

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/constants"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/crud"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/errors"
)

// Create handles the Create operation for Windows Autopilot Deployment Profile Assignment resources.
func (r *WindowsAutopilotDeploymentProfileAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var object WindowsAutopilotDeploymentProfileAssignmentResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Starting creation of resource: %s", ResourceName))

	resp.Diagnostics.Append(req.Plan.Get(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := crud.HandleTimeout(ctx, object.Timeouts.Create, CreateTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	requestBody, err := ConstructWindowsAutopilotDeploymentProfileAssignment(ctx, r.client, object, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error constructing resource",
			fmt.Sprintf("Could not construct resource: %s: %s", ResourceName, err.Error()),
		)
		return
	}

	createdResource, err := r.client.
		DeviceManagement().
		WindowsAutopilotDeploymentProfiles().
		ByWindowsAutopilotDeploymentProfileId(object.WindowsAutopilotDeploymentProfileId.ValueString()).
		Assignments().
		Post(ctx, requestBody, nil)

	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Create", r.WritePermissions)
		return
	}

	object.ID = types.StringValue(*createdResource.GetId())

	resp.Diagnostics.Append(resp.State.Set(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readReq := resource.ReadRequest{State: resp.State, ProviderMeta: req.ProviderMeta}
	stateContainer := &crud.CreateResponseContainer{CreateResponse: resp}

	opts := crud.DefaultReadWithRetryOptions()
	opts.Operation = "Create"
	opts.ResourceTypeName = constants.PROVIDER_NAME + "_" + ResourceName

	err = crud.ReadWithRetry(ctx, r.Read, readReq, stateContainer, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading resource state after create",
			fmt.Sprintf("Could not read resource state: %s: %s", ResourceName, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished Create Method: %s", ResourceName))
}

// Read handles the Read operation for Windows Autopilot Deployment Profile Assignment resources.
func (r *WindowsAutopilotDeploymentProfileAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var object WindowsAutopilotDeploymentProfileAssignmentResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Starting Read method for: %s", ResourceName))

	operation := "Read"
	if ctxOp := ctx.Value("retry_operation"); ctxOp != nil {
		if opStr, ok := ctxOp.(string); ok {
			operation = opStr
		}
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &object)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading %s with ID: %s for Windows Autopilot Deployment Profile: %s",
		ResourceName, object.ID.ValueString(), object.WindowsAutopilotDeploymentProfileId.ValueString()))

	ctx, cancel := crud.HandleTimeout(ctx, object.Timeouts.Read, ReadTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	// Get by ID doesnt exist, despite what the schema says, so we need to get all assignments and find the one with matching ID or criteria
	assignments, err := r.client.
		DeviceManagement().
		WindowsAutopilotDeploymentProfiles().
		ByWindowsAutopilotDeploymentProfileId(object.WindowsAutopilotDeploymentProfileId.ValueString()).
		Assignments().
		Get(ctx, nil)

	if err != nil {
		errors.HandleGraphError(ctx, err, resp, operation, r.ReadPermissions)
		return
	}

	// Find the assignment with matching ID or criteria
	var foundAssignment graphmodels.WindowsAutopilotDeploymentProfileAssignmentable
	if assignments != nil && assignments.GetValue() != nil {
		for _, assignment := range assignments.GetValue() {
			// If we have an ID, match by ID, otherwise match by criteria
			if !object.ID.IsNull() && !object.ID.IsUnknown() {
				if assignment.GetId() != nil && *assignment.GetId() == object.ID.ValueString() {
					foundAssignment = assignment
					break
				}
			} else if matchesAssignment(ctx, object, assignment) {
				foundAssignment = assignment
				object.ID = types.StringValue(*assignment.GetId())
				break
			}
		}
	}

	if foundAssignment == nil {
		tflog.Debug(ctx, fmt.Sprintf("Assignment with ID %s not found in collection", object.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	object = MapRemoteStateToTerraform(ctx, object, foundAssignment)

	resp.Diagnostics.Append(resp.State.Set(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished Read Method: %s", ResourceName))
}

// Update handles the Update operation for Windows Autopilot Deployment Profile Assignment resources.
func (r *WindowsAutopilotDeploymentProfileAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WindowsAutopilotDeploymentProfileAssignmentResourceModel
	var state WindowsAutopilotDeploymentProfileAssignmentResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Updating %s with ID: %s", ResourceName, state.ID.ValueString()))

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := crud.HandleTimeout(ctx, plan.Timeouts.Update, UpdateTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	requestBody, err := ConstructWindowsAutopilotDeploymentProfileAssignment(ctx, r.client, plan, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error constructing resource for update",
			fmt.Sprintf("Could not construct resource: %s: %s", ResourceName, err.Error()),
		)
		return
	}

	// Update the assignment
	updatedResource, err := r.client.
		DeviceManagement().
		WindowsAutopilotDeploymentProfiles().
		ByWindowsAutopilotDeploymentProfileId(plan.WindowsAutopilotDeploymentProfileId.ValueString()).
		Assignments().
		ByWindowsAutopilotDeploymentProfileAssignmentId(state.ID.ValueString()).
		Patch(ctx, requestBody, nil)

	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Update", r.WritePermissions)
		return
	}

	plan.ID = types.StringValue(*updatedResource.GetId())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readReq := resource.ReadRequest{State: resp.State, ProviderMeta: req.ProviderMeta}
	stateContainer := &crud.UpdateResponseContainer{UpdateResponse: resp}

	opts := crud.DefaultReadWithRetryOptions()
	opts.Operation = "Update"
	opts.ResourceTypeName = constants.PROVIDER_NAME + "_" + ResourceName

	err = crud.ReadWithRetry(ctx, r.Read, readReq, stateContainer, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading resource state after update",
			fmt.Sprintf("Could not read resource state: %s: %s", ResourceName, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished updating %s with ID: %s", ResourceName, state.ID.ValueString()))
}

// Delete handles the Delete operation for Windows Autopilot Deployment Profile Assignment resources.
func (r *WindowsAutopilotDeploymentProfileAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var object WindowsAutopilotDeploymentProfileAssignmentResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Starting deletion of resource: %s", ResourceName))

	resp.Diagnostics.Append(req.State.Get(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := crud.HandleTimeout(ctx, object.Timeouts.Delete, DeleteTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	err := r.client.
		DeviceManagement().
		WindowsAutopilotDeploymentProfiles().
		ByWindowsAutopilotDeploymentProfileId(object.WindowsAutopilotDeploymentProfileId.ValueString()).
		Assignments().
		ByWindowsAutopilotDeploymentProfileAssignmentId(object.ID.ValueString()).
		Delete(ctx, nil)

	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Delete", r.WritePermissions)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Removing %s from Terraform state", ResourceName))

	resp.State.RemoveResource(ctx)

	tflog.Debug(ctx, fmt.Sprintf("Finished Delete Method: %s", ResourceName))
}
