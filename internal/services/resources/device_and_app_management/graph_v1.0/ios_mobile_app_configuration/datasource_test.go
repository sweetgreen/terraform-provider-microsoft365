package graphV1IosMobileAppConfiguration_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/mocks"
)

func TestAccIosMobileAppConfigurationDataSource_Basic(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIosMobileAppConfigurationDataSourceConfig_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test", "ios_mobile_app_configurations.#"),
				),
			},
		},
	})
}

func TestAccIosMobileAppConfigurationDataSource_FilterByID(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIosMobileAppConfigurationDataSourceConfig_FilterByID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test", "ios_mobile_app_configurations.#", "1"),
					resource.TestCheckResourceAttrSet("data.microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test", "ios_mobile_app_configurations.0.display_name"),
				),
			},
		},
	})
}

func TestAccIosMobileAppConfigurationDataSource_FilterByDisplayName(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIosMobileAppConfigurationDataSourceConfig_FilterByDisplayName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test", "ios_mobile_app_configurations.#"),
					resource.TestCheckResourceAttr("data.microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test", "ios_mobile_app_configurations.0.display_name", "Test iOS Config"),
				),
			},
		},
	})
}

func testAccIosMobileAppConfigurationDataSourceConfig_Basic() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

data "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration" "test" {
  filter_type = "all"
}
`)
}

func testAccIosMobileAppConfigurationDataSourceConfig_FilterByID() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

resource "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration" "test" {
  display_name = "Test iOS Config for Data Source"
  description  = "Test iOS mobile app configuration for data source"
}

data "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration" "test" {
  filter_type  = "id"
  filter_value = microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test.id
}
`)
}

func testAccIosMobileAppConfigurationDataSourceConfig_FilterByDisplayName() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

resource "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration" "test" {
  display_name = "Test iOS Config"
  description  = "Test iOS mobile app configuration"
}

data "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration" "test" {
  filter_type  = "display_name"
  filter_value = "Test iOS Config"
  
  depends_on = [microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test]
}
`)
}
