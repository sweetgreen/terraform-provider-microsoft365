package schema

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sharedValidators "github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/validators/graph_beta/device_management"
)

// WindowsUpdateAssignments defines the schema for multiple assignments with explicit targets.
func WindowsUpdateAssignments() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "Assignments for Windows Quality Update policies, specifying groups to include or exclude.",
		Validators: []validator.List{
			sharedValidators.SingleIncludeExcludeAssignment(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"target": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Specifies whether the assignment is 'include' or 'exclude'.",
					Validators: []validator.String{
						stringvalidator.OneOf("include", "exclude"),
					},
				},
				"group_ids": schema.SetAttribute{
					Required:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Set of Microsoft Entra ID group IDs to apply for this assignment.",
					Validators: []validator.Set{
						setvalidator.ValueStringsAre(
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$`),
								"must be a valid GUID in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
							),
						),
					},
				},
			},
		},
	}
}
