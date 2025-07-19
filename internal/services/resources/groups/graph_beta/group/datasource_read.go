package graphBetaGroup

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/errors"
)

// Read fetches group data from the Microsoft Graph API
func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel

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

	// Determine filter type and fetch groups accordingly
	filterType := data.FilterType.ValueString()
	filterValue := data.FilterValue.ValueString()

	tflog.Debug(ctx, "Reading groups with filter", map[string]interface{}{
		"filter_type":  filterType,
		"filter_value": filterValue,
	})

	var groups []models.Groupable
	var err error

	switch filterType {
	case "all":
		groups, err = d.fetchAllGroups(ctx)
	case "id":
		if filterValue == "" {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"filter_value is required when filter_type is 'id'",
			)
			return
		}
		group, fetchErr := d.fetchGroupByID(ctx, filterValue)
		if fetchErr != nil {
			err = fetchErr
		} else if group != nil {
			groups = []models.Groupable{group}
		}
	case "display_name":
		if filterValue == "" {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"filter_value is required when filter_type is 'display_name'",
			)
			return
		}
		groups, err = d.fetchGroupsByDisplayName(ctx, filterValue)
	default:
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			fmt.Sprintf("Invalid filter_type: %s", filterType),
		)
		return
	}

	if err != nil {
		errors.HandleGraphError(ctx, err, resp, "groups", d.ReadPermissions)
		return
	}

	// Map the fetched groups to the data source model
	data.Groups = make([]GroupItemModel, 0, len(groups))
	for _, group := range groups {
		groupItem := MapRemoteGroupToDataSource(ctx, group, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Groups = append(data.Groups, groupItem)
	}

	tflog.Debug(ctx, "Successfully read groups", map[string]interface{}{
		"count": len(data.Groups),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// fetchAllGroups retrieves all groups from Microsoft Graph
func (d *GroupDataSource) fetchAllGroups(ctx context.Context) ([]models.Groupable, error) {
	pageIterator, err := d.client.Groups().Get(ctx, &groups.GroupsRequestBuilderGetRequestConfiguration{})
	if err != nil {
		return nil, err
	}

	var allGroups []models.Groupable
	if pageIterator != nil && pageIterator.GetValue() != nil {
		allGroups = append(allGroups, pageIterator.GetValue()...)

		// TODO: Implement proper pagination handling if needed
	}

	return allGroups, nil
}

// fetchGroupByID retrieves a specific group by ID
func (d *GroupDataSource) fetchGroupByID(ctx context.Context, groupID string) (models.Groupable, error) {
	group, err := d.client.Groups().ByGroupId(groupID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// fetchGroupsByDisplayName retrieves groups filtered by display name
func (d *GroupDataSource) fetchGroupsByDisplayName(ctx context.Context, displayName string) ([]models.Groupable, error) {
	filter := fmt.Sprintf("displayName eq '%s'", displayName)
	requestConfig := &groups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: &groups.GroupsRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	}

	result, err := d.client.Groups().Get(ctx, requestConfig)
	if err != nil {
		return nil, err
	}

	if result != nil && result.GetValue() != nil {
		return result.GetValue(), nil
	}

	return []models.Groupable{}, nil
}
