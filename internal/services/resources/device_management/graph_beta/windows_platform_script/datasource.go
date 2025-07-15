package graphBetaWindowsPlatformScript

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
	msgraphbetasdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
)

const (
	DataSourceName        = "graph_beta_device_management_windows_platform_script"
	DataSourceReadTimeout = 180
)

var (
	// Basic data source interface (Read operations)
	_ datasource.DataSource = &WindowsPlatformScriptDataSource{}

	// Allows the data source to be configured with the provider client
	_ datasource.DataSourceWithConfigure = &WindowsPlatformScriptDataSource{}
)

func NewWindowsPlatformScriptDataSource() datasource.DataSource {
	return &WindowsPlatformScriptDataSource{
		ReadPermissions: []string{
			"DeviceManagementConfiguration.Read.All",
			"DeviceManagementScripts.Read.All",
			"DeviceManagementManagedDevices.Read.All",
		},
	}
}

type WindowsPlatformScriptDataSource struct {
	client           *msgraphbetasdk.GraphServiceClient
	ProviderTypeName string
	TypeName         string
	ReadPermissions  []string
}

// Metadata returns the data source type name.
func (d *WindowsPlatformScriptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	d.ProviderTypeName = req.ProviderTypeName
	d.TypeName = DataSourceName
	resp.TypeName = req.ProviderTypeName + "_" + DataSourceName
}

// Configure sets the client for the data source.
func (d *WindowsPlatformScriptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = client.SetGraphBetaClientForDataSource(ctx, req, resp, constants.PROVIDER_NAME+"_"+DataSourceName)
}

// Schema defines the schema for the data source.
func (d *WindowsPlatformScriptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves Windows PowerShell scripts from Microsoft Intune using the `/deviceManagement/deviceManagementScripts` endpoint. This data source enables querying of Windows platform scripts for automated deployment and execution on managed Windows devices.",
		Attributes: map[string]schema.Attribute{
			"filter_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of filter to apply when querying Windows platform scripts. Valid values are: 'all' (retrieve all scripts), 'id' (filter by specific script ID), or 'display_name' (filter by script display name).",
				Validators: []validator.String{
					stringvalidator.OneOf("all", "id", "display_name"),
				},
			},
			"filter_value": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The value to use for filtering. Required when filter_type is 'id' or 'display_name'. When filter_type is 'id', provide the script's unique identifier. When filter_type is 'display_name', provide the exact display name of the script.",
			},
			"windows_platform_scripts": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "A list of Windows platform scripts matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for this Intune windows platform script.",
						},
						"display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the windows platform script.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Optional description for the windows platform script.",
						},
						"script_content": schema.StringAttribute{
							Computed:            true,
							Sensitive:           true,
							MarkdownDescription: "The script content.",
						},
						"run_as_account": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Indicates the type of execution context. Possible values are: `system`, `user`.",
						},
						"enforce_signature_check": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Indicate whether the script signature needs be checked.",
						},
						"file_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Script file name.",
						},
						"role_scope_tag_ids": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Set of scope tag IDs for this Settings Catalog template profile.",
						},
						"run_as_32_bit": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "A value indicating whether the PowerShell script should run as 32-bit.",
						},
						"assignments": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "The list of group assignments for this Windows platform script.",
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
											"collection_id": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "The collection Id that is the target of the assignment.(ConfigMgr).",
											},
											"device_and_app_management_assignment_filter_id": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "The Id of the filter for the target assignment.",
											},
											"device_and_app_management_assignment_filter_type": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "The type of filter of the target assignment i.e. Exclude or Include.",
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
