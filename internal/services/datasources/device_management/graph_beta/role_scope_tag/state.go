package graphBetaRoleScopeTag

import (
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToDataSource maps a Role Scope Tag to a model
func MapRemoteStateToDataSource(data graphmodels.RoleScopeTagable) RoleScopeTagModel {
	model := RoleScopeTagModel{
		ID:          convert.GraphToFrameworkString(data.GetId()),
		DisplayName: convert.GraphToFrameworkString(data.GetDisplayName()),
		Description: convert.GraphToFrameworkString(data.GetDescription()),
	}

	return model
}
