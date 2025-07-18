package graphBetaLinuxPlatformScript

import (
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToDataSource maps a Linux Platform Script to a model
func MapRemoteStateToDataSource(data graphmodels.DeviceManagementConfigurationPolicyable) LinuxPlatformScriptModel {
	model := LinuxPlatformScriptModel{
		ID:          convert.GraphToFrameworkString(data.GetId()),
		DisplayName: convert.GraphToFrameworkString(data.GetName()),
		Description: convert.GraphToFrameworkString(data.GetDescription()),
	}

	return model
}
