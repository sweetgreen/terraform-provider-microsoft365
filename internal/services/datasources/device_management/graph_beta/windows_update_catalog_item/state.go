package graphBetaWindowsUpdateCatalogItem

import (
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToDataSource maps a Windows Update Catalog Item to a model
func MapRemoteStateToDataSource(data graphmodels.WindowsUpdateCatalogItemable) WindowsUpdateCatalogItemModel {
	model := WindowsUpdateCatalogItemModel{
		ID:               convert.GraphToFrameworkString(data.GetId()),
		DisplayName:      convert.GraphToFrameworkString(data.GetDisplayName()),
		ReleaseDateTime:  convert.GraphToFrameworkTime(data.GetReleaseDateTime()),
		EndOfSupportDate: convert.GraphToFrameworkTime(data.GetEndOfSupportDate()),
	}

	return model
}
