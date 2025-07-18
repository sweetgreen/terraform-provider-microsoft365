package graphM365AppsInstallationOptions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

func MapRemoteStateToTerraform(ctx context.Context, data *M365AppsInstallationOptionsResourceModel, remoteResource graphmodels.M365AppsInstallationOptionsable) {
	if remoteResource == nil {
		tflog.Debug(ctx, "Remote resource is nil")
		return
	}

	tflog.Debug(ctx, "Starting to map remote state to Terraform state for M365AppsInstallationOptions")

	updateChannel := remoteResource.GetUpdateChannel()
	if updateChannel != nil {
		data.UpdateChannel = convert.GraphToFrameworkEnum(updateChannel)
	}

	if remoteWindows := remoteResource.GetAppsForWindows(); remoteWindows != nil {
		data.AppsForWindows = &AppsInstallationOptionsForWindows{
			IsMicrosoft365AppsEnabled: convert.GraphToFrameworkBool(remoteWindows.GetIsMicrosoft365AppsEnabled()),
			IsSkypeForBusinessEnabled: convert.GraphToFrameworkBool(remoteWindows.GetIsSkypeForBusinessEnabled()),
		}
	} else {
		data.AppsForWindows = nil
	}

	if remoteMac := remoteResource.GetAppsForMac(); remoteMac != nil {
		data.AppsForMac = &AppsInstallationOptionsForMac{
			IsMicrosoft365AppsEnabled: convert.GraphToFrameworkBool(remoteMac.GetIsMicrosoft365AppsEnabled()),
			IsSkypeForBusinessEnabled: convert.GraphToFrameworkBool(remoteMac.GetIsSkypeForBusinessEnabled()),
		}
	} else {
		data.AppsForMac = nil
	}

	tflog.Debug(ctx, fmt.Sprintf("Finished stating resource %s with id %s", ResourceName, data.ID.ValueString()))

}
