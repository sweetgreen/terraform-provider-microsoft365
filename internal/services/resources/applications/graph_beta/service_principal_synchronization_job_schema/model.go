package graphBetaServicePrincipalSynchronizationJobSchema

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SynchronizationJobSchemaResourceModel is the Terraform model of one synchronization
// rule's object mappings in a provisioning job's schema.
type SynchronizationJobSchemaResourceModel struct {
	ID                    types.String   `tfsdk:"id"`
	ServicePrincipalID    types.String   `tfsdk:"service_principal_id"`
	SynchronizationJobID  types.String   `tfsdk:"synchronization_job_id"`
	SynchronizationRuleID types.String   `tfsdk:"synchronization_rule_id"`
	ObjectMappings        types.Map      `tfsdk:"object_mappings"`
	Version               types.String   `tfsdk:"version"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

// ObjectMappingModel is one element of object_mappings, keyed by target object name.
type ObjectMappingModel struct {
	Name              types.String `tfsdk:"name"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	SourceObjectName  types.String `tfsdk:"source_object_name"`
	FlowTypes         types.Set    `tfsdk:"flow_types"`
	Scope             types.String `tfsdk:"scope"`
	AttributeMappings types.Map    `tfsdk:"attribute_mappings"`
}

// AttributeMappingModel is one element of attribute_mappings, keyed by target attribute name.
type AttributeMappingModel struct {
	SourceExpression        types.String `tfsdk:"source_expression"`
	DefaultValue            types.String `tfsdk:"default_value"`
	FlowType                types.String `tfsdk:"flow_type"`
	FlowBehavior            types.String `tfsdk:"flow_behavior"`
	MatchingPriority        types.Int64  `tfsdk:"matching_priority"`
	ExportMissingReferences types.Bool   `tfsdk:"export_missing_references"`
}

var attributeMappingAttrTypes = map[string]attr.Type{
	"source_expression":         types.StringType,
	"default_value":             types.StringType,
	"flow_type":                 types.StringType,
	"flow_behavior":             types.StringType,
	"matching_priority":         types.Int64Type,
	"export_missing_references": types.BoolType,
}

var attributeMappingType = types.ObjectType{AttrTypes: attributeMappingAttrTypes}

var objectMappingAttrTypes = map[string]attr.Type{
	"name":               types.StringType,
	"enabled":            types.BoolType,
	"source_object_name": types.StringType,
	"flow_types":         types.SetType{ElemType: types.StringType},
	"scope":              types.StringType,
	"attribute_mappings": types.MapType{ElemType: attributeMappingType},
}

var objectMappingType = types.ObjectType{AttrTypes: objectMappingAttrTypes}
