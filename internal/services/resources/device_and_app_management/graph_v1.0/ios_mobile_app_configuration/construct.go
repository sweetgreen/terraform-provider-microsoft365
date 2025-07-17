package graphV1IosMobileAppConfiguration

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/convert"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// constructResource constructs an iOS mobile app configuration resource from Terraform model
func constructResource(ctx context.Context, data *IosMobileAppConfigurationResourceModel) (graphmodels.IosMobileAppConfigurationable, error) {
	resource := graphmodels.NewIosMobileAppConfiguration()

	// Set display name
	displayName := data.DisplayName.ValueString()
	resource.SetDisplayName(&displayName)

	// Set description if provided
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		description := data.Description.ValueString()
		resource.SetDescription(&description)
	}

	// Set targeted mobile apps if provided
	if !data.TargetedMobileApps.IsNull() && !data.TargetedMobileApps.IsUnknown() {
		var targetedApps []string
		err := convert.FrameworkToGraphStringList(ctx, data.TargetedMobileApps, func(val []string) {
			targetedApps = val
		})
		if err != nil {
			return nil, fmt.Errorf("error converting targeted mobile apps: %v", err)
		}
		resource.SetTargetedMobileApps(targetedApps)
	}

	// Set encoded setting XML if provided
	if !data.EncodedSettingXml.IsNull() && !data.EncodedSettingXml.IsUnknown() {
		encodedXml := []byte(data.EncodedSettingXml.ValueString())
		resource.SetEncodedSettingXml(encodedXml)
	}

	// Set settings if provided
	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var settingsList []AppConfigurationSettingItemModel
		data.Settings.ElementsAs(ctx, &settingsList, false)

		if len(settingsList) > 0 {
			settings := make([]graphmodels.AppConfigurationSettingItemable, 0, len(settingsList))
			for _, setting := range settingsList {
				appConfigSetting := graphmodels.NewAppConfigurationSettingItem()

				// Set app config key
				appConfigKey := setting.AppConfigKey.ValueString()
				appConfigSetting.SetAppConfigKey(&appConfigKey)

				// Set app config key type
				keyType := setting.AppConfigKeyType.ValueString()
				switch keyType {
				case "stringType":
					keyTypeEnum := graphmodels.STRINGTYPE_MDMAPPCONFIGKEYTYPE
					appConfigSetting.SetAppConfigKeyType(&keyTypeEnum)
				case "integerType":
					keyTypeEnum := graphmodels.INTEGERTYPE_MDMAPPCONFIGKEYTYPE
					appConfigSetting.SetAppConfigKeyType(&keyTypeEnum)
				case "realType":
					keyTypeEnum := graphmodels.REALTYPE_MDMAPPCONFIGKEYTYPE
					appConfigSetting.SetAppConfigKeyType(&keyTypeEnum)
				case "booleanType":
					keyTypeEnum := graphmodels.BOOLEANTYPE_MDMAPPCONFIGKEYTYPE
					appConfigSetting.SetAppConfigKeyType(&keyTypeEnum)
				case "tokenType":
					keyTypeEnum := graphmodels.TOKENTYPE_MDMAPPCONFIGKEYTYPE
					appConfigSetting.SetAppConfigKeyType(&keyTypeEnum)
				}

				// Set app config key value
				appConfigKeyValue := setting.AppConfigKeyValue.ValueString()
				appConfigSetting.SetAppConfigKeyValue(&appConfigKeyValue)

				settings = append(settings, appConfigSetting)
			}
			resource.SetSettings(settings)
		}
	}

	return resource, nil
}

// constructResourceForUpdate constructs an iOS mobile app configuration resource for update
func constructResourceForUpdate(ctx context.Context, data *IosMobileAppConfigurationResourceModel) (graphmodels.IosMobileAppConfigurationable, error) {
	// For updates, we use the same construction as create
	return constructResource(ctx, data)
}

// constructAssignment constructs an assignment from Terraform model
func constructAssignment(ctx context.Context, data *ManagedDeviceMobileAppConfigurationAssignmentModel) (graphmodels.ManagedDeviceMobileAppConfigurationAssignmentable, error) {
	assignment := graphmodels.NewManagedDeviceMobileAppConfigurationAssignment()

	// Construct target
	target, err := constructAssignmentTarget(ctx, &data.Target)
	if err != nil {
		return nil, err
	}
	assignment.SetTarget(target)

	return assignment, nil
}

// constructAssignmentTarget constructs an assignment target based on OData type
func constructAssignmentTarget(ctx context.Context, data *DeviceAndAppManagementAssignmentTargetModel) (graphmodels.DeviceAndAppManagementAssignmentTargetable, error) {
	odataType := data.ODataType.ValueString()

	switch odataType {
	case "#microsoft.graph.allLicensedUsersAssignmentTarget":
		target := graphmodels.NewAllLicensedUsersAssignmentTarget()
		return target, nil

	case "#microsoft.graph.allDevicesAssignmentTarget":
		target := graphmodels.NewAllDevicesAssignmentTarget()
		return target, nil

	case "#microsoft.graph.groupAssignmentTarget":
		target := graphmodels.NewGroupAssignmentTarget()
		if !data.GroupId.IsNull() && !data.GroupId.IsUnknown() {
			groupId := data.GroupId.ValueString()
			target.SetGroupId(&groupId)
		}
		// Note: Filter settings are not available in v1.0 API
		return target, nil

	case "#microsoft.graph.exclusionGroupAssignmentTarget":
		target := graphmodels.NewExclusionGroupAssignmentTarget()
		if !data.GroupId.IsNull() && !data.GroupId.IsUnknown() {
			groupId := data.GroupId.ValueString()
			target.SetGroupId(&groupId)
		}
		// Note: Filter settings are not available in v1.0 API
		return target, nil

	default:
		return nil, fmt.Errorf("unsupported assignment target type: %s", odataType)
	}
}
