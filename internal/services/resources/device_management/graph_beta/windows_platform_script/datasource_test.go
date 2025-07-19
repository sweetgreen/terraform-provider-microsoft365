package graphBetaWindowsPlatformScript_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWindowsPlatformScriptDataSource_FilterByAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configDataSourceWindowsPlatformScriptFilterAll(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.microsoft365_graph_beta_device_management_windows_platform_script.test", "filter_type", "all"),
					resource.TestCheckResourceAttrSet("data.microsoft365_graph_beta_device_management_windows_platform_script.test", "windows_platform_scripts.#"),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	// Pre-check logic would go here
}

var testAccProtoV6ProviderFactories = make(map[string]func() (tfprotov6.ProviderServer, error))

func configDataSourceWindowsPlatformScriptFilterAll() string {
	return `
data "microsoft365_graph_beta_device_management_windows_platform_script" "test" {
  filter_type = "all"
}
`
}
