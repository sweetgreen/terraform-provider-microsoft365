package graphBetaWindowsPlatformScript

import (
	"context"
	"fmt"
	"time"

	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/errors"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

// Read fetches Windows platform script data from the Microsoft Graph API
func (d *WindowsPlatformScriptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WindowsPlatformScriptDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create timeout
	readTimeout, diags := data.Timeouts.Read(ctx, time.Duration(DataSourceReadTimeout)*time.Second)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	// Determine filter type and fetch scripts accordingly
	filterType := data.FilterType.ValueString()
	filterValue := data.FilterValue.ValueString()

	tflog.Debug(ctx, "Reading Windows platform scripts with filter", map[string]interface{}{
		"filter_type":  filterType,
		"filter_value": filterValue,
	})

	var scripts []models.DeviceManagementScriptable
	var err error

	switch filterType {
	case "all":
		scripts, err = d.fetchAllScripts(ctx)
	case "id":
		if filterValue == "" {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"filter_value is required when filter_type is 'id'",
			)
			return
		}
		script, fetchErr := d.fetchScriptByID(ctx, filterValue)
		if fetchErr != nil {
			err = fetchErr
		} else if script != nil {
			scripts = []models.DeviceManagementScriptable{script}
		}
	case "display_name":
		if filterValue == "" {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"filter_value is required when filter_type is 'display_name'",
			)
			return
		}
		scripts, err = d.fetchScriptsByDisplayName(ctx, filterValue)
	default:
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			fmt.Sprintf("Invalid filter_type: %s", filterType),
		)
		return
	}

	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "Read", d.ReadPermissions)
		return
	}

	// Map the fetched scripts to the data source model
	data.WindowsPlatformScripts = make([]WindowsPlatformScriptItemModel, 0, len(scripts))
	for _, script := range scripts {
		scriptItem := WindowsPlatformScriptItemModel{}
		MapRemoteScriptToDataSource(ctx, &scriptItem, script)

		// Fetch assignments for this script
		if script.GetId() != nil {
			assignments, err := d.fetchScriptAssignments(ctx, *script.GetId())
			if err != nil {
				tflog.Warn(ctx, "Failed to fetch assignments for script", map[string]interface{}{
					"script_id": *script.GetId(),
					"error":     err.Error(),
				})
			} else {
				MapRemoteAssignmentsToDataSource(ctx, &scriptItem, assignments)
			}
		}

		data.WindowsPlatformScripts = append(data.WindowsPlatformScripts, scriptItem)
	}

	tflog.Debug(ctx, "Successfully read Windows platform scripts", map[string]interface{}{
		"count": len(data.WindowsPlatformScripts),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// fetchAllScripts retrieves all device management scripts from Microsoft Graph
func (d *WindowsPlatformScriptDataSource) fetchAllScripts(ctx context.Context) ([]models.DeviceManagementScriptable, error) {
	result, err := d.client.DeviceManagement().DeviceManagementScripts().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	if result != nil && result.GetValue() != nil {
		return result.GetValue(), nil
	}

	return []models.DeviceManagementScriptable{}, nil
}

// fetchScriptByID retrieves a specific script by ID
func (d *WindowsPlatformScriptDataSource) fetchScriptByID(ctx context.Context, scriptID string) (models.DeviceManagementScriptable, error) {
	script, err := d.client.DeviceManagement().DeviceManagementScripts().ByDeviceManagementScriptId(scriptID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	return script, nil
}

// fetchScriptsByDisplayName retrieves scripts filtered by display name
func (d *WindowsPlatformScriptDataSource) fetchScriptsByDisplayName(ctx context.Context, displayName string) ([]models.DeviceManagementScriptable, error) {
	// First get all scripts
	allScripts, err := d.fetchAllScripts(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by display name
	var filteredScripts []models.DeviceManagementScriptable
	for _, script := range allScripts {
		if script.GetDisplayName() != nil && *script.GetDisplayName() == displayName {
			filteredScripts = append(filteredScripts, script)
		}
	}

	return filteredScripts, nil
}

// fetchScriptAssignments retrieves assignments for a specific script
func (d *WindowsPlatformScriptDataSource) fetchScriptAssignments(ctx context.Context, scriptID string) ([]models.DeviceManagementScriptAssignmentable, error) {
	assignments, err := d.client.DeviceManagement().DeviceManagementScripts().ByDeviceManagementScriptId(scriptID).Assignments().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	if assignments != nil && assignments.GetValue() != nil {
		return assignments.GetValue(), nil
	}

	return []models.DeviceManagementScriptAssignmentable{}, nil
}
