package graphBetaGroup

import (
	"context"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/convert"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

// MapRemoteGroupToDataSource maps a Microsoft Graph group to the Terraform data source model
func MapRemoteGroupToDataSource(ctx context.Context, remoteGroup models.Groupable, diags *diag.Diagnostics) GroupItemModel {
	if remoteGroup == nil {
		tflog.Debug(ctx, "Remote group is nil")
		return GroupItemModel{}
	}

	var groupItem GroupItemModel

	// Map all properties using convert functions
	groupItem.ID = convert.GraphToFrameworkString(remoteGroup.GetId())
	groupItem.DisplayName = convert.GraphToFrameworkString(remoteGroup.GetDisplayName())
	groupItem.Description = convert.GraphToFrameworkString(remoteGroup.GetDescription())
	groupItem.MailNickname = convert.GraphToFrameworkString(remoteGroup.GetMailNickname())
	groupItem.MailEnabled = convert.GraphToFrameworkBool(remoteGroup.GetMailEnabled())
	groupItem.SecurityEnabled = convert.GraphToFrameworkBool(remoteGroup.GetSecurityEnabled())
	groupItem.GroupTypes = convert.GraphToFrameworkStringSet(ctx, remoteGroup.GetGroupTypes())
	groupItem.Visibility = convert.GraphToFrameworkString(remoteGroup.GetVisibility())
	groupItem.IsAssignableToRole = convert.GraphToFrameworkBool(remoteGroup.GetIsAssignableToRole())
	groupItem.MembershipRule = convert.GraphToFrameworkString(remoteGroup.GetMembershipRule())
	groupItem.MembershipRuleProcessingState = convert.GraphToFrameworkString(remoteGroup.GetMembershipRuleProcessingState())
	groupItem.CreatedDateTime = convert.GraphToFrameworkTime(remoteGroup.GetCreatedDateTime())
	groupItem.Mail = convert.GraphToFrameworkString(remoteGroup.GetMail())
	groupItem.ProxyAddresses = convert.GraphToFrameworkStringSet(ctx, remoteGroup.GetProxyAddresses())
	groupItem.OnPremisesSyncEnabled = convert.GraphToFrameworkBool(remoteGroup.GetOnPremisesSyncEnabled())
	groupItem.PreferredDataLocation = convert.GraphToFrameworkString(remoteGroup.GetPreferredDataLocation())
	groupItem.PreferredLanguage = convert.GraphToFrameworkString(remoteGroup.GetPreferredLanguage())
	groupItem.Theme = convert.GraphToFrameworkString(remoteGroup.GetTheme())
	groupItem.Classification = convert.GraphToFrameworkString(remoteGroup.GetClassification())
	groupItem.ExpirationDateTime = convert.GraphToFrameworkTime(remoteGroup.GetExpirationDateTime())
	groupItem.RenewedDateTime = convert.GraphToFrameworkTime(remoteGroup.GetRenewedDateTime())
	groupItem.SecurityIdentifier = convert.GraphToFrameworkString(remoteGroup.GetSecurityIdentifier())

	return groupItem
}
