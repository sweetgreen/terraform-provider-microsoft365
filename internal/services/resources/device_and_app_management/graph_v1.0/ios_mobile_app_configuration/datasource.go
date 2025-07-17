package graphV1IosMobileAppConfiguration

import (
	"context"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/client"
	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/constants"
	commonschema "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/schema"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

const (
	DataSourceName        = "graph_v1_device_and_app_management_ios_mobile_app_configuration"
	DataSourceReadTimeout = 180
)

var (
	// Basic data source interface (Read operations)
	_ datasource.DataSource = &IosMobileAppConfigurationDataSource{}

	// Allows the data source to be configured with the provider client
	_ datasource.DataSourceWithConfigure = &IosMobileAppConfigurationDataSource{}
)

func NewIosMobileAppConfigurationDataSource() datasource.DataSource {
	return &IosMobileAppConfigurationDataSource{
		ReadPermissions: []string{
			"DeviceManagementApps.Read.All",
		},
	}
}

type IosMobileAppConfigurationDataSource struct {
	client           *msgraphsdk.GraphServiceClient
	ProviderTypeName string
	TypeName         string
	ReadPermissions  []string
}

// Metadata returns the data source type name.
func (d *IosMobileAppConfigurationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	d.ProviderTypeName = req.ProviderTypeName
	d.TypeName = DataSourceName
	resp.TypeName = req.ProviderTypeName + "_" + DataSourceName
}

// Configure sets the client for the data source.
func (d *IosMobileAppConfigurationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = client.SetGraphStableClientForDataSource(ctx, req, resp, constants.PROVIDER_NAME+"_"+DataSourceName)
}

// Schema defines the schema for the data source.
func (d *IosMobileAppConfigurationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves iOS mobile app configuration policies from Microsoft Intune. This data source allows you to query app-specific settings for managed iOS applications.",
		Attributes: map[string]schema.Attribute{
			"filter_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of filter to apply when querying iOS mobile app configurations. Valid values are: 'all' (retrieve all configurations), 'id' (filter by specific configuration ID), or 'display_name' (filter by configuration display name).",
				Validators: []validator.String{
					stringvalidator.OneOf("all", "id", "display_name"),
				},
			},
			"filter_value": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The value to use for filtering. Required when filter_type is 'id' or 'display_name'. When filter_type is 'id', provide the configuration's unique identifier. When filter_type is 'display_name', provide the exact display name of the configuration.",
			},
			"ios_mobile_app_configurations": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "A list of iOS mobile app configurations matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the iOS mobile app configuration.",
						},
						"display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name of the iOS mobile app configuration.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The description of the iOS mobile app configuration.",
						},
						"version": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Version of the device configuration.",
						},
						"created_date_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DateTime the object was created.",
						},
						"last_modified_date_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DateTime the object was last modified.",
						},
						"targeted_mobile_apps": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "The list of targeted mobile app IDs.",
						},
						"encoded_setting_xml": schema.StringAttribute{
							Computed:            true,
							Sensitive:           true,
							MarkdownDescription: "Base64 encoded configuration XML.",
						},
						"settings": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "iOS app configuration settings.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"app_config_key": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The application configuration key.",
									},
									"app_config_key_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The application configuration key type.",
									},
									"app_config_key_value": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The application configuration key value.",
									},
								},
							},
						},
						"assignments": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "The list of assignments for this iOS mobile app configuration.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Key of the entity.",
									},
									"target": schema.SingleNestedAttribute{
										Computed:            true,
										MarkdownDescription: "Target for the assignment.",
										Attributes: map[string]schema.Attribute{
											"odata_type": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "The type of assignment target.",
											},
											"group_id": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "The group Id that is the target of the assignment.",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"timeouts": commonschema.Timeouts(ctx),
		},
	}
}
