package graphV1IosMobileAppConfiguration_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/mocks"
)

func TestAccIosMobileAppConfigurationResource_Basic(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resourceName := "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIosMobileAppConfigurationResourceConfig_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "Test iOS App Config"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test iOS mobile app configuration"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIosMobileAppConfigurationResource_WithSettings(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resourceName := "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIosMobileAppConfigurationResourceConfig_WithSettings(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "Test iOS App Config with Settings"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.app_config_key", "server_url"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.app_config_key_type", "stringType"),
					resource.TestCheckResourceAttr(resourceName, "settings.1.app_config_key", "enable_feature"),
					resource.TestCheckResourceAttr(resourceName, "settings.1.app_config_key_type", "booleanType"),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	// Add any required pre-checks here
}

func testAccIosMobileAppConfigurationResourceConfig_Basic() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

resource "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration" "test" {
  display_name = "Test iOS App Config"
  description  = "Test iOS mobile app configuration"
}
`)
}

func testAccIosMobileAppConfigurationResourceConfig_WithSettings() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

resource "microsoft365_graph_v1_device_and_app_management_ios_mobile_app_configuration" "test" {
  display_name = "Test iOS App Config with Settings"
  description  = "Test iOS mobile app configuration with settings"
  
  settings {
    app_config_key       = "server_url"
    app_config_key_type  = "stringType"
    app_config_key_value = "https://api.example.com"
  }
  
  settings {
    app_config_key       = "enable_feature"
    app_config_key_type  = "booleanType"
    app_config_key_value = "true"
  }
}
`)
}
