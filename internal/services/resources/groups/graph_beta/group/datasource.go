package graphBetaGroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	msgraphbetasdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/client"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/constants"
	commonschema "github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/schema"
)

const (
	DataSourceName        = "graph_beta_groups_group"
	DataSourceReadTimeout = 180
)

var (
	// Basic data source interface (Read operations)
	_ datasource.DataSource = &GroupDataSource{}

	// Allows the data source to be configured with the provider client
	_ datasource.DataSourceWithConfigure = &GroupDataSource{}
)

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{
		ReadPermissions: []string{
			"Group.Read.All",
			"Directory.Read.All",
		},
	}
}

type GroupDataSource struct {
	client           *msgraphbetasdk.GraphServiceClient
	ProviderTypeName string
	TypeName         string
	ReadPermissions  []string
}

// Metadata returns the data source type name.
func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	d.ProviderTypeName = req.ProviderTypeName
	d.TypeName = DataSourceName
	resp.TypeName = req.ProviderTypeName + "_" + DataSourceName
}

// Configure sets the client for the data source.
func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = client.SetGraphBetaClientForDataSource(ctx, req, resp, constants.PROVIDER_NAME+"_"+DataSourceName)
}

// Schema defines the schema for the data source.
func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves Azure AD/Entra groups using the `/groups` endpoint. This data source enables querying of security groups, Microsoft 365 groups, and distribution groups with support for filtering by ID, display name, or retrieving all groups. Use this to lookup group information for organizational identity and access management configurations.",
		Attributes: map[string]schema.Attribute{
			"filter_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of filter to apply when querying groups. Valid values are: 'all' (retrieve all groups), 'id' (filter by specific group ID), or 'display_name' (filter by group display name).",
				Validators: []validator.String{
					stringvalidator.OneOf("all", "id", "display_name"),
				},
			},
			"filter_value": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The value to use for filtering. Required when filter_type is 'id' or 'display_name'. When filter_type is 'id', provide the group's unique identifier. When filter_type is 'display_name', provide the exact display name of the group.",
			},
			"groups": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "A list of groups matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the group.",
						},
						"display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name for the group.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "An optional description for the group.",
						},
						"mail_nickname": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The mail alias for the group.",
						},
						"mail_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Specifies whether the group is mail-enabled.",
						},
						"security_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Specifies whether the group is a security group.",
						},
						"group_types": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Specifies the group type and its membership.",
						},
						"visibility": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Specifies the group join policy and group content visibility for groups.",
						},
						"is_assignable_to_role": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Indicates whether this group can be assigned to a Microsoft Entra role.",
						},
						"membership_rule": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The rule that determines members for this group if the group is a dynamic group.",
						},
						"membership_rule_processing_state": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Indicates whether the dynamic membership processing is on or paused.",
						},
						"created_date_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timestamp of when the group was created.",
						},
						"mail": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The SMTP address for the group.",
						},
						"proxy_addresses": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Email addresses for the group that direct to the same group mailbox.",
						},
						"on_premises_sync_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "true if this group is synced from an on-premises directory.",
						},
						"preferred_data_location": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The preferred data location for the Microsoft 365 group.",
						},
						"preferred_language": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The preferred language for a Microsoft 365 group.",
						},
						"theme": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Specifies a Microsoft 365 group's color theme.",
						},
						"classification": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Describes a classification for the group.",
						},
						"expiration_date_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timestamp of when the group is set to expire.",
						},
						"renewed_date_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timestamp of when the group was last renewed.",
						},
						"security_identifier": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Security identifier of the group, used in Windows scenarios.",
						},
					},
				},
			},
			"timeouts": commonschema.Timeouts(ctx),
		},
	}
}
