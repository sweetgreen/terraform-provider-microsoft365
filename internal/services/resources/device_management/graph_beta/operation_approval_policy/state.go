package graphBetaOperationApprovalPolicy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToTerraform maps a remote operation approval policy to the Terraform resource model
func MapRemoteStateToTerraform(ctx context.Context, data OperationApprovalPolicyResourceModel, policy graphmodels.OperationApprovalPolicyable) OperationApprovalPolicyResourceModel {
	if policy == nil {
		tflog.Debug(ctx, "Remote policy is nil")
		return data
	}

	data.ID = convert.GraphToFrameworkString(policy.GetId())
	data.DisplayName = convert.GraphToFrameworkString(policy.GetDisplayName())
	data.Description = convert.GraphToFrameworkString(policy.GetDescription())

	if lastModified := policy.GetLastModifiedDateTime(); lastModified != nil {
		data.LastModifiedDateTime = types.StringValue(lastModified.Format("2006-01-02T15:04:05Z07:00"))
	}

	if policyType := policy.GetPolicyType(); policyType != nil {
		data.PolicyType = types.StringValue(policyType.String())
	}

	if policyPlatform := policy.GetPolicyPlatform(); policyPlatform != nil {
		data.PolicyPlatform = types.StringValue(policyPlatform.String())
	}

	if policySet := policy.GetPolicySet(); policySet != nil {
		data.PolicySet = mapRemotePolicySetToTerraform(policySet)
	}

	if approverGroupIds := policy.GetApproverGroupIds(); approverGroupIds != nil {
		data.ApproverGroupIds = convert.GraphToFrameworkStringSet(ctx, approverGroupIds)
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished stating resource %s with id %s", ResourceName, data.ID.ValueString()))

	return data
}

// mapRemotePolicySetToTerraform maps a remote policy set to a Terraform policy set
func mapRemotePolicySetToTerraform(remotePolicySet graphmodels.OperationApprovalPolicySetable) OperationApprovalPolicySetResourceModel {
	policySet := OperationApprovalPolicySetResourceModel{}

	if policyType := remotePolicySet.GetPolicyType(); policyType != nil {
		policySet.PolicyType = types.StringValue(policyType.String())
	}

	if policyPlatform := remotePolicySet.GetPolicyPlatform(); policyPlatform != nil {
		policySet.PolicyPlatform = types.StringValue(policyPlatform.String())
	}

	return policySet
}
