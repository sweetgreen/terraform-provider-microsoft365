package graphBetaCloudPcProvisioningPolicy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/devicemanagement"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/constructors"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// constructAssignmentsRequestBody creates the request body for assigning a Cloud PC provisioning policy to groups
func constructAssignmentsRequestBody(ctx context.Context, assignments []CloudPcProvisioningPolicyAssignmentModel) (*devicemanagement.VirtualEndpointProvisioningPoliciesItemAssignPostRequestBody, error) {
	tflog.Debug(ctx, "Constructing assignments request body")

	requestBody := devicemanagement.NewVirtualEndpointProvisioningPoliciesItemAssignPostRequestBody()

	// If no assignments are provided, return an empty assignments array
	// This is used for removing all assignments
	if len(assignments) == 0 {
		requestBody.SetAssignments([]models.CloudPcProvisioningPolicyAssignmentable{})
		tflog.Debug(ctx, "No assignments provided, setting empty assignments array")
		return requestBody, nil
	}

	graphAssignments := []models.CloudPcProvisioningPolicyAssignmentable{}

	for i, assignment := range assignments {
		if assignment.GroupId.IsNull() || assignment.GroupId.ValueString() == "" {
			tflog.Warn(ctx, fmt.Sprintf("Skipping assignment %d with empty group ID", i))
			continue
		}

		tflog.Debug(ctx, fmt.Sprintf("Creating assignment %d for group ID: %s", i, assignment.GroupId.ValueString()))

		graphAssignment := models.NewCloudPcProvisioningPolicyAssignment()
		if !assignment.ID.IsNull() && assignment.ID.ValueString() != "" {
			convert.FrameworkToGraphString(assignment.ID, graphAssignment.SetId)
		}

		// Create the target with the proper @odata.type
		target := models.NewCloudPcManagementAssignmentTarget()

		// Always set the @odata.type for consistency
		odataType := "#microsoft.graph.cloudPcManagementGroupAssignmentTarget"
		target.SetOdataType(&odataType)

		// Set up additional data with the groupId
		additionalData := target.GetAdditionalData()

		// The groupId must be set directly in the additionalData map
		additionalData["groupId"] = assignment.GroupId.ValueString()

		tflog.Debug(ctx, fmt.Sprintf("Setting groupId in additionalData: %s", assignment.GroupId.ValueString()))

		// Handle Frontline-specific fields
		if !assignment.ServicePlanId.IsNull() && assignment.ServicePlanId.ValueString() != "" {
			// Frontline (dedicated/shared)
			tflog.Debug(ctx, fmt.Sprintf("Setting frontline-specific fields for assignment %d", i))
			additionalData["servicePlanId"] = assignment.ServicePlanId.ValueString()

			if !assignment.AllotmentDisplayName.IsNull() && assignment.AllotmentDisplayName.ValueString() != "" {
				additionalData["allotmentDisplayName"] = assignment.AllotmentDisplayName.ValueString()
			} else {
				additionalData["allotmentDisplayName"] = "Default Allotment"
			}

			if !assignment.AllotmentLicenseCount.IsNull() {
				additionalData["allotmentLicensesCount"] = int32(assignment.AllotmentLicenseCount.ValueInt64())
			} else {
				additionalData["allotmentLicensesCount"] = int32(1)
			}
		}

		graphAssignment.SetTarget(target)
		graphAssignments = append(graphAssignments, graphAssignment)

		tflog.Debug(ctx, fmt.Sprintf("Successfully created assignment %d", i))
	}

	requestBody.SetAssignments(graphAssignments)

	if err := constructors.DebugLogGraphObject(ctx, fmt.Sprintf("Final JSON to be sent to Graph API for resource %s assignments", ResourceName), requestBody); err != nil {
		tflog.Error(ctx, "Failed to debug log object", map[string]interface{}{
			"error": err.Error(),
		})
	}

	tflog.Debug(ctx, fmt.Sprintf("Constructed assignments request body with %d assignments", len(graphAssignments)))
	return requestBody, nil
}
