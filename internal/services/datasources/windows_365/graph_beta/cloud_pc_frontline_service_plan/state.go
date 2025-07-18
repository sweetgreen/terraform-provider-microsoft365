package graphBetaCloudPcFrontlineServicePlan

import (
	"context"

	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/services/common/convert"
)

func MapRemoteStateToDataSource(ctx context.Context, data graphmodels.CloudPcFrontLineServicePlanable) CloudPcFrontlineServicePlanItemModel {
	model := CloudPcFrontlineServicePlanItemModel{
		ID:          convert.GraphToFrameworkString(data.GetId()),
		DisplayName: convert.GraphToFrameworkString(data.GetDisplayName()),
	}

	// Convert Int32 to Int64 for Terraform
	if totalCount := data.GetTotalCount(); totalCount != nil {
		model.TotalCount = convert.GraphToFrameworkInt32AsInt64(totalCount)
	}

	if usedCount := data.GetUsedCount(); usedCount != nil {
		model.UsedCount = convert.GraphToFrameworkInt32AsInt64(usedCount)
	}

	return model
}
