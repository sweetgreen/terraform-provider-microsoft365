package graphBetaMacOSDmgApp

import (
	"context"
	"regexp"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/client"
	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/constants"
	planmodifiers "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/plan_modifiers"
	commonschema "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/schema"
	commonschemagraphbeta "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/schema/graph_beta/device_and_app_management"
	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	msgraphbetasdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
)

const (
	ResourceName  = "graph_beta_device_and_app_management_macos_dmg_app"
	CreateTimeout = 180
	UpdateTimeout = 180
	ReadTimeout   = 180
	DeleteTimeout = 180
)

var (
	// Basic resource interface (CRUD operations)
	_ resource.Resource = &MacOSDmgAppResource{}

	// Allows the resource to be configured with the provider client
	_ resource.ResourceWithConfigure = &MacOSDmgAppResource{}

	// Enables import functionality
	_ resource.ResourceWithImportState = &MacOSDmgAppResource{}

	// Enables plan modification/diff suppression
	_ resource.ResourceWithModifyPlan = &MacOSDmgAppResource{}
)

func NewMacOSDmgAppResource() resource.Resource {
	return &MacOSDmgAppResource{
		ReadPermissions: []string{
			"DeviceManagementApps.Read.All",
			"DeviceManagementConfiguration.Read.All",
		},
		WritePermissions: []string{
			"DeviceManagementApps.ReadWrite.All",
			"DeviceManagementConfiguration.ReadWrite.All",
		},
		ResourcePath: "/deviceAppManagement/mobileApps",
	}
}

type MacOSDmgAppResource struct {
	client           *msgraphbetasdk.GraphServiceClient
	ProviderTypeName string
	TypeName         string
	ReadPermissions  []string
	WritePermissions []string
	ResourcePath     string
}

// Metadata returns the resource type name.
func (r *MacOSDmgAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.ProviderTypeName = req.ProviderTypeName
	r.TypeName = ResourceName
	resp.TypeName = r.FullTypeName()
}

// FullTypeName returns the full resource type name in the format "providername_resourcename".
func (r *MacOSDmgAppResource) FullTypeName() string {
	return r.ProviderTypeName + "_" + r.TypeName
}

// Configure sets the client for the resource.
func (r *MacOSDmgAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = client.SetGraphBetaClientForResource(ctx, req, resp, constants.PROVIDER_NAME+"_"+ResourceName)
}

// ImportState imports the resource state.
func (r *MacOSDmgAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema returns the schema for the resource.
func (r *MacOSDmgAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a macOS DMG (Apple Disk Image) app deployment in Microsoft Intune using the `/deviceAppManagement/mobileApps` endpoint. DMG files contain one or more .app files that are automatically extracted and installed to the /Applications folder on managed macOS devices via the Intune management agent.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.UseStateForUnknownString(),
				},
				MarkdownDescription: "The unique identifier of the macOS DMG app.",
			},
			"is_featured": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "The value indicating whether the app is marked as featured by the admin.",
			},
			"privacy_information_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The privacy statement Url.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(constants.HttpOrHttpsUrlRegex),
						"must be a valid URL starting with http:// or https://",
					),
				},
			},
			"information_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The more information Url.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(constants.HttpOrHttpsUrlRegex),
						"must be a valid URL starting with http:// or https://",
					),
				},
			},
			"owner": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The owner of the app.",
			},
			"developer": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The developer of the app.",
			},
			"notes": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Notes for the app.",
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The title of the Intune macOS DMG application.",
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A detailed description of the Intune macOS DMG application.",
			},
			"publisher": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The publisher of the Intune macOS DMG application.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 1024),
				},
			},
			"categories": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of category names to associate with this application. You can use either the predefined Intune category names like 'Business', 'Productivity', etc., or provide specific category UUIDs. Predefined values include: 'Other apps', 'Books & Reference', 'Data management', 'Productivity', 'Business', 'Development & Design', 'Photos & Media', 'Collaboration & Social', 'Computer management'.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^(Other apps|Books & Reference|Data management|Productivity|Business|Development & Design|Photos & Media|Collaboration & Social|Computer management|[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$`),
							"must be either a predefined category name or a valid GUID in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
						),
					),
				},
			},
			"relationships": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "List of relationships associated with this application.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the relationship.",
						},
						"source_display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name of the source application in the relationship.",
						},
						"source_display_version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display version of the source application in the relationship.",
						},
						"source_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the source application in the relationship.",
						},
						"source_publisher_display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name of the source application's publisher in the relationship.",
						},
						"target_display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name of the target application in the relationship.",
						},
						"target_display_version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display version of the target application in the relationship.",
						},
						"target_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the target application in the relationship.",
						},
						"target_publisher": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The publisher of the target application in the relationship.",
						},
						"target_publisher_display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name of the target application's publisher in the relationship.",
						},
						"target_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The type of the target in the relationship.",
						},
					},
				},
			},
			"created_date_time": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.UseStateForUnknownString(),
				},
				MarkdownDescription: "The date and time the app was created. This property is read-only.",
			},
			"upload_state": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					planmodifiers.UseStateForUnknownInt32(),
				},
				MarkdownDescription: "The upload state. Possible values are: 0 - Not Ready, 1 - Ready, 2 - Processing. This property is read-only.",
			},
			"publishing_state": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					planmodifiers.UseStateForUnknownString(),
				},
				MarkdownDescription: "The publishing state for the app. The app cannot be assigned unless the app is published. " +
					"Possible values are: notPublished, processing, published.",
			},
			"is_assigned": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					planmodifiers.UseStateForUnknownBool(),
				},
				MarkdownDescription: "The value indicating whether the app is assigned to at least one group. This property is read-only.",
			},
			"role_scope_tag_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set of scope tag IDs for this Settings Catalog template profile.",
				PlanModifiers: []planmodifier.Set{
					planmodifiers.DefaultSetValue(
						[]attr.Value{types.StringValue("0")},
					),
				},
			},
			"dependent_app_count": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					planmodifiers.UseStateForUnknownInt32(),
				},
				MarkdownDescription: "The total number of dependencies the child app has. This property is read-only.",
			},
			"superseding_app_count": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					planmodifiers.UseStateForUnknownInt32(),
				},
				MarkdownDescription: "The total number of apps this app directly or indirectly supersedes. This property is read-only.",
			},
			"superseded_app_count": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					planmodifiers.UseStateForUnknownInt32(),
				},
				MarkdownDescription: "The total number of apps this app is directly or indirectly superseded by. This property is read-only.",
			},
			"macos_dmg_app": schema.SingleNestedAttribute{
				MarkdownDescription: "The path to the DMG file to be uploaded. Resource supports both local file sources and URL based sources.",
				Optional:            true,
				Validators: []validator.Object{
					validators.ExactlyOneOf("installer_file_path_source", "installer_url_source"),
				},
				Attributes: map[string]schema.Attribute{
					"ignore_version_detection": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "When TRUE, indicates that the app's version will NOT be used to detect if the app is installed on a device. When FALSE, indicates that the app's version will be used to detect if the app is installed on a device. Set this to true for apps that use a self update feature. The default value is FALSE.",
					},
					"included_apps": schema.SetNestedAttribute{
						Required: true,
						PlanModifiers: []planmodifier.Set{
							planmodifiers.UseStateForUnknownSet(),
						},
						MarkdownDescription: "Define the app bundle identifiers and version numbers to be used to detect the presence of the macOS app installation. This list is dynamically populated based on the DMG metadata, and users can also append additional entries. Maximum of 500 apps.\n\n" +
							"### Notes:\n" +
							"- Included app bundle IDs (`CFBundleIdentifier`) and build numbers (`CFBundleShortVersionString`) are used for detecting and monitoring app installation status of the uploaded file.\n" +
							"- The list should **only** contain the application(s) installed by the uploaded file in the `/Applications` folder on macOS.\n" +
							"- Any other type of file that is not an application or is not installed in the `/Applications` folder should **not** be included.\n" +
							"- If the list contains files that are not applications or none of the listed apps are installed, app installation status will **not** report success.\n" +
							"- When multiple apps are present in the DMG, the **first app** in the list is used to identify the application.\n\n" +
							"### Example:\n" +
							"To retrieve the `CFBundleIdentifier` and `CFBundleShortVersionString` of an installed application, you can use the macOS Terminal:\n\n" +
							"```bash\n" +
							"# Retrieve the Bundle Identifier\n" +
							"defaults read /Applications/Company\\ Portal.app/Contents/Info CFBundleIdentifier\n\n" +
							"# Retrieve the Short Version String\n" +
							"defaults read /Applications/Company\\ Portal.app/Contents/Info CFBundleShortVersionString\n" +
							"```\n\n" +
							"Alternatively, these values can also be located in the `<app_name>.app/Contents/Info.plist` file inside the mounted DMG.\n\n" +
							"For apps added to Intune, the Intune admin center can also provide the app bundle ID.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"bundle_id": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "The `CFBundleIdentifier` of the app as defined in the DMG metadata or appended manually.",
									PlanModifiers: []planmodifier.String{
										planmodifiers.UseStateForUnknownString(),
									},
								},
								"bundle_version": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "The `CFBundleShortVersionString` of the app as defined in the DMG metadata or appended manually.",
									PlanModifiers: []planmodifier.String{
										planmodifiers.UseStateForUnknownString(),
									},
								},
							},
						},
					},
					"primary_bundle_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The bundleId of the primary app in the DMG. Maps to CFBundleIdentifier in the app's bundle configuration.",
						PlanModifiers: []planmodifier.String{
							planmodifiers.UseStateForUnknownString(),
						},
					},
					"primary_bundle_version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The version of the primary app in the DMG. Maps to CFBundleShortVersion in the app's bundle configuration.",
						PlanModifiers: []planmodifier.String{
							planmodifiers.UseStateForUnknownString(),
						},
					},
					"minimum_supported_operating_system": schema.SingleNestedAttribute{
						Required:            true,
						MarkdownDescription: "Specifies the minimum macOS version required for the application. Each field indicates whether the version is supported.",
						Attributes: map[string]schema.Attribute{
							"v10_7": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.7 or later. Defaults to `false`.",
							},
							"v10_8": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.8 or later. Defaults to `false`.",
							},
							"v10_9": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.9 or later. Defaults to `false`.",
							},
							"v10_10": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.10 or later. Defaults to `false`.",
							},
							"v10_11": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.11 or later. Defaults to `false`.",
							},
							"v10_12": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.12 or later. Defaults to `false`.",
							},
							"v10_13": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.13 or later. Defaults to `false`.",
							},
							"v10_14": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.14 or later. Defaults to `false`.",
							},
							"v10_15": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 10.15 or later. Defaults to `false`.",
							},
							"v11_0": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 11.0 or later. Defaults to `false`.",
							},
							"v12_0": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 12.0 or later. Defaults to `false`.",
							},
							"v13_0": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 13.0 or later. Defaults to `false`.",
							},
							"v14_0": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 14.0 or later. Defaults to `false`.",
							},
							"v15_0": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Application supports macOS 15.0 or later. Defaults to `false`.",
							},
						},
					},
				},
			},
			"content_version": commonschemagraphbeta.MobileAppContentVersionSchema(),
			"app_installer":   commonschemagraphbeta.MobileAppDmgInstallerMetadataSchema(),
			"app_icon":        commonschemagraphbeta.MobileAppIconSchema(),
			"timeouts":        commonschema.Timeouts(ctx),
		},
	}
}
