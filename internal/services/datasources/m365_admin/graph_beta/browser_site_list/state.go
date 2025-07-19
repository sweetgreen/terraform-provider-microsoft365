package graphBetaBrowserSiteList

import (
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

// MapRemoteStateToDataSource maps a Browser Site List to a model
func MapRemoteStateToDataSource(data graphmodels.BrowserSiteListable) BrowserSiteListResourceModel {
	model := BrowserSiteListResourceModel{
		ID:          convert.GraphToFrameworkString(data.GetId()),
		DisplayName: convert.GraphToFrameworkString(data.GetDisplayName()),
	}

	return model
}
