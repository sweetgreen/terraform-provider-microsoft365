package graphBetaWindowsRemediationScript

import (
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToDataSource maps a Windows Remediation Script to a model
func MapRemoteStateToDataSource(data graphmodels.DeviceHealthScriptable) WindowsRemediationScriptModel {
	model := WindowsRemediationScriptModel{
		ID:          convert.GraphToFrameworkString(data.GetId()),
		DisplayName: convert.GraphToFrameworkString(data.GetDisplayName()),
		Description: convert.GraphToFrameworkString(data.GetDescription()),
	}

	return model
}
