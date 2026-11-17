package graphBetaServicePrincipalSynchronizationJobSchema

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	msgraphbetasdk "github.com/microsoftgraph/msgraph-beta-sdk-go"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/client"
	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/constants"
	commonschema "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/schema"
)

const (
	ResourceName  = "microsoft365_graph_beta_applications_service_principal_synchronization_job_schema"
	CreateTimeout = 180
	ReadTimeout   = 180
	UpdateTimeout = 180
	DeleteTimeout = 180
)

var (
	_ resource.Resource                   = &SynchronizationJobSchemaResource{}
	_ resource.ResourceWithConfigure      = &SynchronizationJobSchemaResource{}
	_ resource.ResourceWithImportState    = &SynchronizationJobSchemaResource{}
	_ resource.ResourceWithValidateConfig = &SynchronizationJobSchemaResource{}
)

// SynchronizationJobSchemaResource manages the attribute mappings of an Entra
// provisioning (synchronization) job.
type SynchronizationJobSchemaResource struct {
	client           *msgraphbetasdk.GraphServiceClient
	ReadPermissions  []string
	WritePermissions []string
}

func NewSynchronizationJobSchemaResource() resource.Resource {
	return &SynchronizationJobSchemaResource{
		ReadPermissions: []string{
			"Synchronization.Read.All",
		},
		WritePermissions: []string{
			"Synchronization.ReadWrite.All",
		},
	}
}

func (r *SynchronizationJobSchemaResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = ResourceName
}

func (r *SynchronizationJobSchemaResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = client.SetGraphBetaClientForResource(ctx, req, resp, ResourceName)
}

// ImportState accepts service_principal_id/synchronization_job_id/synchronization_rule_id.
// The rule ID may be left out when the schema has exactly one rule.
func (r *SynchronizationJobSchemaResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, "/")
	if (len(parts) != 2 && len(parts) != 3) || parts[0] == "" || parts[1] == "" || (len(parts) == 3 && parts[2] == "") {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Use service_principal_id/synchronization_job_id[/synchronization_rule_id], got %q.", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_principal_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("synchronization_job_id"), parts[1])...)
	if len(parts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("synchronization_rule_id"), parts[2])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}

func (r *SynchronizationJobSchemaResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	requiresReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the attribute mappings of a Microsoft Entra provisioning job " +
			"(the object and attribute mappings shown under **Provisioning > Mappings** of an enterprise application) " +
			"using `/servicePrincipals/{id}/synchronization/jobs/{jobId}/schema`.\n\n" +
			"The resource is authoritative for the object mappings of one synchronization rule: object and attribute " +
			"mappings that exist in Entra but not in the configuration are removed on apply. Directory definitions, " +
			"other rules, and rule and object mapping metadata are sent back unchanged.\n\n" +
			"The job's schema always exists, so creating this resource overwrites the live mappings with the " +
			"configuration. Import it first so the first plan shows the real difference. Destroying the resource only " +
			"removes it from state; the mappings in Entra are left as they are.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "`service_principal_id/synchronization_job_id/synchronization_rule_id`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"service_principal_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Object ID of the enterprise application's service principal.",
				PlanModifiers:       requiresReplace,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(constants.GuidRegex), "must be a valid GUID"),
				},
			},
			"synchronization_job_id": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "ID of the provisioning job, for example " +
					"`ServiceNowOutDelta.<tenant id without dashes>.<application id>`. " +
					"List them with `GET /servicePrincipals/{id}/synchronization/jobs`.",
				PlanModifiers: requiresReplace,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"synchronization_rule_id": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "ID of the synchronization rule whose object mappings are managed " +
					"(`synchronizationRules[].id` in the schema). Most applications have a single rule.",
				PlanModifiers: requiresReplace,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"object_mappings": schema.MapNestedAttribute{
				Required: true,
				MarkdownDescription: "Object mappings of the rule, keyed by target object name (for example `sys_user`). " +
					"Mappings in Entra that are not listed here are removed on apply.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Display name of the mapping, for example `Provision Microsoft Entra ID Users`. Disabled template mappings often have none.",
						},
						"enabled": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Whether objects are provisioned with this mapping.",
						},
						"source_object_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Source object in Microsoft Entra ID, for example `User` or `Group`. Disabled template mappings may have an empty string.",
						},
						"flow_types": schema.SetAttribute{
							Required:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Actions the mapping takes on target objects: any of `Add`, `Update` and `Delete`.",
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(stringvalidator.OneOf("Add", "Update", "Delete")),
							},
						},
						"scope": schema.StringAttribute{
							Optional: true,
							MarkdownDescription: "Source object scoping filter (`filter` resource in Graph) as JSON, for example from `jsonencode()`. " +
								"Leave unset for no scoping filter.",
							Validators: []validator.String{jsonValidator{}},
						},
						"attribute_mappings": schema.MapNestedAttribute{
							Required: true,
							MarkdownDescription: "Attribute mappings keyed by target attribute name (for example `active`). " +
								"Mappings in Entra that are not listed here are removed on apply.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"source_expression": schema.StringAttribute{
										Optional: true,
										MarkdownDescription: "Value written to the target attribute, as an Entra expression: an attribute such as `[mail]`, " +
											"a constant such as `\"1\"`, or a function such as `Switch([IsSoftDeleted], , \"False\", \"1\", \"True\", \"0\")`. " +
											"Write it as Entra stores it: arguments separated by a comma and a space, and an empty argument as nothing " +
											"(`, ,`); validation shows the expected form. Leave unset for a mapping with no source.",
									},
									"default_value": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: "Value used when the expression returns null. Entra distinguishes null (unset) from the empty string.",
									},
									"flow_type": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString("Always"),
										MarkdownDescription: "When the attribute is written: `Always` (default, create and update) or `ObjectAddOnly` (create only), or `MultiValueAddOnly`, `ValueAddOnly`, `AttributeAddOnly`.",
										Validators: []validator.String{
											stringvalidator.OneOf("Always", "ObjectAddOnly", "MultiValueAddOnly", "ValueAddOnly", "AttributeAddOnly"),
										},
									},
									"flow_behavior": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString("FlowWhenChanged"),
										MarkdownDescription: "`FlowWhenChanged` (default) or `FlowAlways`.",
										Validators: []validator.String{
											stringvalidator.OneOf("FlowWhenChanged", "FlowAlways"),
										},
									},
									"matching_priority": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             int64default.StaticInt64(0),
										MarkdownDescription: "Priority of this attribute when matching source and target objects. `0` (default) means it is not used for matching; `1` is the first matching attribute.",
									},
									"export_missing_references": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             booldefault.StaticBool(false),
										MarkdownDescription: "Whether to export references to objects that are not yet provisioned. Defaults to `false`.",
									},
								},
							},
						},
					},
				},
			},
			"version": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "Schema version Entra returned, which changes on every schema change. An apply fails without " +
					"writing anything if the version in Entra no longer matches the one read when the plan was made.",
			},
			"timeouts": commonschema.ResourceTimeouts(ctx),
		},
	}
}

// ValidateConfig checks expression syntax and form at plan time, without calling Graph.
func (r *SynchronizationJobSchemaResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config SynchronizationJobSchemaResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.ObjectMappings.IsNull() || config.ObjectMappings.IsUnknown() {
		return
	}
	objects := map[string]ObjectMappingModel{}
	if config.ObjectMappings.ElementsAs(ctx, &objects, false).HasError() {
		return
	}
	for _, target := range sortedKeys(objects) {
		om := objects[target]
		if om.AttributeMappings.IsNull() || om.AttributeMappings.IsUnknown() {
			continue
		}
		attrs := map[string]AttributeMappingModel{}
		if om.AttributeMappings.ElementsAs(ctx, &attrs, false).HasError() {
			continue
		}
		for _, name := range sortedKeys(attrs) {
			expr := attrs[name].SourceExpression
			if expr.IsNull() || expr.IsUnknown() {
				continue
			}
			p := path.Root("object_mappings").AtMapKey(target).AtName("attribute_mappings").AtMapKey(name).AtName("source_expression")
			canonical, err := canonicalExpression(expr.ValueString())
			if err != nil {
				resp.Diagnostics.AddAttributeError(p, "Invalid provisioning expression", fmt.Sprintf("%q: %s", expr.ValueString(), err))
				continue
			}
			if canonical != expr.ValueString() {
				resp.Diagnostics.AddAttributeError(
					p,
					"Provisioning expression is not in the form Entra stores",
					fmt.Sprintf(
						"Entra stores this expression as\n\n  %s\n\nWrite it exactly that way, or every plan after the apply will show a change.",
						canonical,
					),
				)
			}
		}
	}
}

type jsonValidator struct{}

func (jsonValidator) Description(context.Context) string { return "value must be valid JSON" }

func (v jsonValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (jsonValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !json.Valid([]byte(req.ConfigValue.ValueString())) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", "scope must be a JSON document, for example from jsonencode().")
	}
}
