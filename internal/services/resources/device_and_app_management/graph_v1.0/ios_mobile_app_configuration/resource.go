package graphV1IosMobileAppConfiguration

import (
	"context"

	"github.com/sweetgreen/terraform-provider-microsoft365/internal/client"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/constants"
	planmodifiers "github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/plan_modifiers"
	commonschema "github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/schema"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

const (
	ResourceName  = "graph_v1_device_and_app_management_ios_mobile_app_configuration"
	CreateTimeout = 180
	UpdateTimeout = 180
	ReadTimeout   = 180
	DeleteTimeout = 180
)

var (
	// Basic resource interface (CRUD operations)
	_ resource.Resource = &IosMobileAppConfigurationResource{}

	// Allows the resource to be configured with the provider client
	_ resource.ResourceWithConfigure = &IosMobileAppConfigurationResource{}

	// Enables import functionality
	_ resource.ResourceWithImportState = &IosMobileAppConfigurationResource{}

	// Enables plan modification/diff suppression
	_ resource.ResourceWithModifyPlan = &IosMobileAppConfigurationResource{}
)

func NewIosMobileAppConfigurationResource() resource.Resource {
	return &IosMobileAppConfigurationResource{
		ReadPermissions: []string{
			"DeviceManagementApps.Read.All",
		},
		WritePermissions: []string{
			"DeviceManagementApps.ReadWrite.All",
		},
		ResourcePath: "/deviceAppManagement/mobileAppConfigurations",
	}
}

type IosMobileAppConfigurationResource struct {
	client           *msgraphsdk.GraphServiceClient
	ProviderTypeName string
	TypeName         string
	ReadPermissions  []string
	WritePermissions []string
	ResourcePath     string
}

// Metadata returns the resource type name.
func (r *IosMobileAppConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.ProviderTypeName = req.ProviderTypeName
	r.TypeName = ResourceName
	resp.TypeName = req.ProviderTypeName + "_" + ResourceName
}

// Configure sets the client for the resource.
func (r *IosMobileAppConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = client.SetGraphStableClientForResource(ctx, req, resp, constants.PROVIDER_NAME+"_"+ResourceName)
}

// ImportState imports the resource state.
func (r *IosMobileAppConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *IosMobileAppConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages iOS mobile app configuration policies in Microsoft Intune. These policies allow you to configure app-specific settings for managed iOS applications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.UseStateForUnknownString(),
				},
				MarkdownDescription: "The unique identifier for the iOS mobile app configuration. Read-only.",
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the iOS mobile app configuration.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the iOS mobile app configuration.",
			},
			"version": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Version of the device configuration. Read-only.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DateTime the object was created. Read-only.",
			},
			"last_modified_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DateTime the object was last modified. Read-only.",
			},
			"targeted_mobile_apps": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "The list of targeted mobile app IDs.",
			},
			"encoded_setting_xml": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Computed:            true,
				MarkdownDescription: "Base64 encoded configuration XML.",
				PlanModifiers: []planmodifier.String{
					planmodifiers.UseStateForUnknownString(),
				},
			},
			"settings": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "iOS app configuration settings.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"app_config_key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The application configuration key.",
						},
						"app_config_key_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The application configuration key type. Possible values are: stringType, integerType, realType, booleanType, tokenType.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"stringType",
									"integerType",
									"realType",
									"booleanType",
									"tokenType",
								),
							},
						},
						"app_config_key_value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The application configuration key value.",
						},
					},
				},
			},
			"assignments": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The list of assignments for this iOS mobile app configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Key of the entity. Read-only.",
						},
						"target": schema.SingleNestedAttribute{
							Required:            true,
							MarkdownDescription: "Target for the assignment.",
							Attributes: map[string]schema.Attribute{
								"odata_type": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "The type of assignment target. Possible values are: #microsoft.graph.allLicensedUsersAssignmentTarget, #microsoft.graph.allDevicesAssignmentTarget, #microsoft.graph.exclusionGroupAssignmentTarget, #microsoft.graph.groupAssignmentTarget.",
									Validators: []validator.String{
										stringvalidator.OneOf(
											"#microsoft.graph.allLicensedUsersAssignmentTarget",
											"#microsoft.graph.allDevicesAssignmentTarget",
											"#microsoft.graph.exclusionGroupAssignmentTarget",
											"#microsoft.graph.groupAssignmentTarget",
										),
									},
								},
								"group_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "The group Id that is the target of the assignment. Required when odata_type is groupAssignmentTarget or exclusionGroupAssignmentTarget.",
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
