package graphBetaServicePrincipalSynchronizationJobSchema

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mappingsToTerraform converts the mappings read from Graph into object_mappings.
// prior is the configuration or state the read follows; a scope that is the same JSON
// as the prior value keeps the prior text, so key order and spacing don't show as a change.
func mappingsToTerraform(
	ctx context.Context,
	mappings map[string]objectMapping,
	prior map[string]ObjectMappingModel,
) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	objects := map[string]attr.Value{}
	for target, om := range mappings {
		attrs := map[string]attr.Value{}
		for name, am := range om.AttributeMappings {
			value, d := types.ObjectValue(attributeMappingAttrTypes, map[string]attr.Value{
				"source_expression":         types.StringPointerValue(am.SourceExpression),
				"default_value":             types.StringPointerValue(am.DefaultValue),
				"flow_type":                 types.StringValue(am.FlowType),
				"flow_behavior":             types.StringValue(am.FlowBehavior),
				"matching_priority":         types.Int64Value(am.MatchingPriority),
				"export_missing_references": types.BoolValue(am.ExportMissingReferences),
			})
			diags.Append(d...)
			attrs[name] = value
		}
		attrMap, d := types.MapValue(attributeMappingType, attrs)
		diags.Append(d...)

		flowTypes := make([]attr.Value, 0, len(om.FlowTypes))
		for _, f := range om.FlowTypes {
			flowTypes = append(flowTypes, types.StringValue(f))
		}
		flowSet, d := types.SetValue(types.StringType, flowTypes)
		diags.Append(d...)

		scope := types.StringPointerValue(om.Scope)
		if p, ok := prior[target]; ok && om.Scope != nil && !p.Scope.IsNull() && !p.Scope.IsUnknown() &&
			sameJSONText(p.Scope.ValueString(), *om.Scope) {
			scope = p.Scope
		}

		value, d := types.ObjectValue(objectMappingAttrTypes, map[string]attr.Value{
			"name":               types.StringPointerValue(om.Name),
			"enabled":            types.BoolValue(om.Enabled),
			"source_object_name": types.StringValue(om.SourceObjectName),
			"flow_types":         flowSet,
			"scope":              scope,
			"attribute_mappings": attrMap,
		})
		diags.Append(d...)
		objects[target] = value
	}
	result, d := types.MapValue(objectMappingType, objects)
	diags.Append(d...)
	return result, diags
}

// mappingsFromTerraform converts object_mappings from a plan into the form written to Graph.
func mappingsFromTerraform(ctx context.Context, value types.Map) (map[string]objectMapping, diag.Diagnostics) {
	var diags diag.Diagnostics
	objects := map[string]ObjectMappingModel{}
	diags.Append(value.ElementsAs(ctx, &objects, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make(map[string]objectMapping, len(objects))
	for target, om := range objects {
		var flowTypes []string
		diags.Append(om.FlowTypes.ElementsAs(ctx, &flowTypes, false)...)
		sort.Strings(flowTypes)
		attrs := map[string]AttributeMappingModel{}
		diags.Append(om.AttributeMappings.ElementsAs(ctx, &attrs, false)...)
		mapping := objectMapping{
			Name:              om.Name.ValueStringPointer(),
			Enabled:           om.Enabled.ValueBool(),
			SourceObjectName:  om.SourceObjectName.ValueString(),
			FlowTypes:         flowTypes,
			Scope:             om.Scope.ValueStringPointer(),
			AttributeMappings: make(map[string]attributeMapping, len(attrs)),
		}
		for name, am := range attrs {
			mapping.AttributeMappings[name] = attributeMapping{
				SourceExpression:        am.SourceExpression.ValueStringPointer(),
				DefaultValue:            am.DefaultValue.ValueStringPointer(),
				FlowType:                am.FlowType.ValueString(),
				FlowBehavior:            am.FlowBehavior.ValueString(),
				MatchingPriority:        am.MatchingPriority.ValueInt64(),
				ExportMissingReferences: am.ExportMissingReferences.ValueBool(),
			}
		}
		out[target] = mapping
	}
	return out, diags
}

// objectMappingModels returns the elements of object_mappings, or nil if it is null or unknown.
func objectMappingModels(ctx context.Context, value types.Map) map[string]ObjectMappingModel {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	out := map[string]ObjectMappingModel{}
	if value.ElementsAs(ctx, &out, false).HasError() {
		return nil
	}
	return out
}

func sameJSONText(a, b string) bool {
	var x, y any
	if json.Unmarshal([]byte(a), &x) != nil || json.Unmarshal([]byte(b), &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}
