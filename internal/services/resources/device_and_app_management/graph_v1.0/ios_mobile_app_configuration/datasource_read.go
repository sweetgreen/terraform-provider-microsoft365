package graphV1IosMobileAppConfiguration

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/errors"
)

// Read fetches iOS mobile app configuration data from the Microsoft Graph API
func (d *IosMobileAppConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IosMobileAppConfigurationDataSourceModel

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

	// Determine filter type and fetch configurations accordingly
	filterType := data.FilterType.ValueString()
	filterValue := data.FilterValue.ValueString()

	tflog.Debug(ctx, "Reading iOS mobile app configurations with filter", map[string]interface{}{
		"filter_type":  filterType,
		"filter_value": filterValue,
	})

	var configurations []models.ManagedDeviceMobileAppConfigurationable
	var err error

	switch filterType {
	case "all":
		configurations, err = d.fetchAllConfigurations(ctx)
	case "id":
		if filterValue == "" {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"filter_value is required when filter_type is 'id'",
			)
			return
		}
		configuration, fetchErr := d.fetchConfigurationByID(ctx, filterValue)
		if fetchErr != nil {
			err = fetchErr
		} else if configuration != nil {
			configurations = []models.ManagedDeviceMobileAppConfigurationable{configuration}
		}
	case "display_name":
		if filterValue == "" {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"filter_value is required when filter_type is 'display_name'",
			)
			return
		}
		configurations, err = d.fetchConfigurationsByDisplayName(ctx, filterValue)
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

	// Map the fetched configurations to the data source model
	data.IosMobileAppConfigurations = make([]IosMobileAppConfigurationItemModel, 0, len(configurations))
	for _, configuration := range configurations {
		// Only process iOS mobile app configurations
		if iosConfig, ok := configuration.(models.IosMobileAppConfigurationable); ok {
			configItem := IosMobileAppConfigurationItemModel{}
			MapRemoteConfigurationToDataSource(ctx, &configItem, iosConfig)

			// Fetch assignments for this configuration
			if iosConfig.GetId() != nil {
				assignments, err := d.fetchConfigurationAssignments(ctx, *iosConfig.GetId())
				if err != nil {
					tflog.Warn(ctx, "Failed to fetch assignments for configuration", map[string]interface{}{
						"config_id": *iosConfig.GetId(),
						"error":     err.Error(),
					})
				} else {
					MapRemoteAssignmentsToDataSource(ctx, &configItem, assignments)
				}
			}

			data.IosMobileAppConfigurations = append(data.IosMobileAppConfigurations, configItem)
		}
	}

	tflog.Debug(ctx, "Successfully read iOS mobile app configurations", map[string]interface{}{
		"count": len(data.IosMobileAppConfigurations),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// fetchAllConfigurations retrieves all mobile app configurations from Microsoft Graph
func (d *IosMobileAppConfigurationDataSource) fetchAllConfigurations(ctx context.Context) ([]models.ManagedDeviceMobileAppConfigurationable, error) {
	result, err := d.client.DeviceAppManagement().MobileAppConfigurations().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	if result != nil && result.GetValue() != nil {
		return result.GetValue(), nil
	}

	return []models.ManagedDeviceMobileAppConfigurationable{}, nil
}

// fetchConfigurationByID retrieves a specific configuration by ID
func (d *IosMobileAppConfigurationDataSource) fetchConfigurationByID(ctx context.Context, configID string) (models.ManagedDeviceMobileAppConfigurationable, error) {
	configuration, err := d.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(configID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	return configuration, nil
}

// fetchConfigurationsByDisplayName retrieves configurations filtered by display name
func (d *IosMobileAppConfigurationDataSource) fetchConfigurationsByDisplayName(ctx context.Context, displayName string) ([]models.ManagedDeviceMobileAppConfigurationable, error) {
	// First get all configurations
	allConfigs, err := d.fetchAllConfigurations(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by display name
	var filteredConfigs []models.ManagedDeviceMobileAppConfigurationable
	for _, config := range allConfigs {
		if config.GetDisplayName() != nil && *config.GetDisplayName() == displayName {
			filteredConfigs = append(filteredConfigs, config)
		}
	}

	return filteredConfigs, nil
}

// fetchConfigurationAssignments retrieves assignments for a specific configuration
func (d *IosMobileAppConfigurationDataSource) fetchConfigurationAssignments(ctx context.Context, configID string) ([]models.ManagedDeviceMobileAppConfigurationAssignmentable, error) {
	assignments, err := d.client.DeviceAppManagement().MobileAppConfigurations().ByManagedDeviceMobileAppConfigurationId(configID).Assignments().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	if assignments != nil && assignments.GetValue() != nil {
		return assignments.GetValue(), nil
	}

	return []models.ManagedDeviceMobileAppConfigurationAssignmentable{}, nil
}
