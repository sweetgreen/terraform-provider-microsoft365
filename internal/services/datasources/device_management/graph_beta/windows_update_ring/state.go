package graphBetaWindowsUpdateRing

import (
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToDataSource maps a Windows Update Ring to a model
func MapRemoteStateToDataSource(data graphmodels.DeviceConfigurationable) WindowsUpdateRingModel {
	model := WindowsUpdateRingModel{
		ID:          convert.GraphToFrameworkString(data.GetId()),
		DisplayName: convert.GraphToFrameworkString(data.GetDisplayName()),
		Description: convert.GraphToFrameworkString(data.GetDescription()),
	}

	return model
}
