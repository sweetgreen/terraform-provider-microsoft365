package graphV1IosMobileAppConfiguration

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/crud"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/errors"
)

// Create handles the Create operation.
func (r *IosMobileAppConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var object IosMobileAppConfigurationResourceModel

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

	requestBody, err := constructResource(ctx, &object)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error constructing resource for creation",
			fmt.Sprintf("Could not construct resource: %s", err.Error()),
		)
		return
	}

	resource, err := r.client.DeviceAppManagement().MobileAppConfigurations().Post(ctx, requestBody, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Create", r.WritePermissions)
		return
	}

	MapRemoteResourceToTerraform(ctx, &object, resource)

	// Handle assignments if provided
	if !object.Assignments.IsNull() && !object.Assignments.IsUnknown() {
		var assignmentsList []ManagedDeviceMobileAppConfigurationAssignmentModel
		object.Assignments.ElementsAs(ctx, &assignmentsList, false)

		if len(assignmentsList) > 0 {
			err = r.updateAssignments(ctx, object.ID.ValueString(), assignmentsList)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating assignments",
					fmt.Sprintf("Could not update assignments: %s", err.Error()),
				)
				return
			}
		}
	}

	// Re-read the resource to ensure we have the exact state from the API
	createdResource, err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Get(ctx, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Read after Create", r.ReadPermissions)
		return
	}

	MapRemoteResourceToTerraform(ctx, &object, createdResource)

	// Always read assignments to get the current state
	assignments, err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Assignments().Get(ctx, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Read assignments after Create", r.ReadPermissions)
		return
	}

	if assignments != nil && assignments.GetValue() != nil {
		MapRemoteAssignmentsToTerraform(ctx, &object, assignments.GetValue())
	} else {
		// Ensure assignments is set to empty list if no assignments exist
		MapRemoteAssignmentsToTerraform(ctx, &object, nil)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished creating resource %s with ID: %s", ResourceName, object.ID.ValueString()))
}

// Read handles the Read operation.
func (r *IosMobileAppConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var object IosMobileAppConfigurationResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Starting read of resource: %s", ResourceName))

	resp.Diagnostics.Append(req.State.Get(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading %s with ID: %s", ResourceName, object.ID.ValueString()))

	ctx, cancel := crud.HandleTimeout(ctx, object.Timeouts.Read, ReadTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	resource, err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Get(ctx, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Read", r.ReadPermissions)
		return
	}

	MapRemoteResourceToTerraform(ctx, &object, resource)

	// Read assignments
	assignments, err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Assignments().Get(ctx, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Read", r.ReadPermissions)
		return
	}

	if assignments != nil && assignments.GetValue() != nil {
		MapRemoteAssignmentsToTerraform(ctx, &object, assignments.GetValue())
	} else {
		// Ensure assignments is set to empty list if no assignments exist
		MapRemoteAssignmentsToTerraform(ctx, &object, nil)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished reading resource %s with ID: %s", ResourceName, object.ID.ValueString()))
}

// Update handles the Update operation.
func (r *IosMobileAppConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var object IosMobileAppConfigurationResourceModel

	tflog.Debug(ctx, fmt.Sprintf("Starting update of resource: %s", ResourceName))

	resp.Diagnostics.Append(req.Plan.Get(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := crud.HandleTimeout(ctx, object.Timeouts.Update, UpdateTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	requestBody, err := constructResourceForUpdate(ctx, &object)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error constructing resource for update",
			fmt.Sprintf("Could not construct resource: %s", err.Error()),
		)
		return
	}

	_, err = r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Patch(ctx, requestBody, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Update", r.WritePermissions)
		return
	}

	// Handle assignments update
	if !object.Assignments.IsNull() && !object.Assignments.IsUnknown() {
		var assignmentsList []ManagedDeviceMobileAppConfigurationAssignmentModel
		object.Assignments.ElementsAs(ctx, &assignmentsList, false)

		err = r.updateAssignments(ctx, object.ID.ValueString(), assignmentsList)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating assignments",
				fmt.Sprintf("Could not update assignments: %s", err.Error()),
			)
			return
		}
	}

	// Read the updated resource
	resource, err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Get(ctx, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Read", r.ReadPermissions)
		return
	}

	MapRemoteResourceToTerraform(ctx, &object, resource)

	// Read assignments
	assignments, err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Assignments().Get(ctx, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Read", r.ReadPermissions)
		return
	}

	if assignments != nil && assignments.GetValue() != nil {
		MapRemoteAssignmentsToTerraform(ctx, &object, assignments.GetValue())
	} else {
		// Ensure assignments is set to empty list if no assignments exist
		MapRemoteAssignmentsToTerraform(ctx, &object, nil)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &object)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished updating resource %s with ID: %s", ResourceName, object.ID.ValueString()))
}

// Delete handles the Delete operation.
func (r *IosMobileAppConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var object IosMobileAppConfigurationResourceModel

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

	err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(object.ID.ValueString()).Delete(ctx, nil)
	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Delete", r.WritePermissions)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished deleting resource %s with ID: %s", ResourceName, object.ID.ValueString()))
}

// updateAssignments handles updating assignments for the mobile app configuration
func (r *IosMobileAppConfigurationResource) updateAssignments(ctx context.Context, configId string, assignments []ManagedDeviceMobileAppConfigurationAssignmentModel) error {
	// First, get current assignments
	currentAssignments, err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(configId).Assignments().Get(ctx, nil)
	if err != nil {
		return err
	}

	// Delete all current assignments
	if currentAssignments != nil && currentAssignments.GetValue() != nil {
		for _, assignment := range currentAssignments.GetValue() {
			if assignment.GetId() != nil {
				err := r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(configId).Assignments().ByManagedDeviceMobileAppConfigurationAssignmentId(*assignment.GetId()).Delete(ctx, nil)
				if err != nil {
					return err
				}
			}
		}
	}

	// Create new assignments
	for _, assignment := range assignments {
		assignmentBody, err := constructAssignment(ctx, &assignment)
		if err != nil {
			return err
		}

		_, err = r.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(configId).Assignments().Post(ctx, assignmentBody, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
