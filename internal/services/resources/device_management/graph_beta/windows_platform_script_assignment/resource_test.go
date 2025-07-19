package graphBetaWindowsPlatformScriptAssignment_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/mocks"
)

const unitTestProviderConfig = `
provider "microsoft365" {
  tenant_id = "00000000-0000-0000-0000-000000000001"
  auth_method = "client_secret"
  entra_id_options = {
    client_id = "11111111-1111-1111-1111-111111111111"
    client_secret = "mock-secret-value"
  }
  cloud = "public"
}
`

// TODO: Add mocks for this resource to enable unit testing
/*
func TestUnitWindowsPlatformScriptAssignmentResource_Basic(t *testing.T) {
	// Set up test environment
	t.Setenv("TF_ACC", "0")
	t.Setenv("MS365_TEST_MODE", "true")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_windows_platform_script_assignment" "test" {
  windows_platform_script_id = "00000000-0000-0000-0000-000000000002"
  target = {
    target_type = "allDevices"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_windows_platform_script_assignment.test", "windows_platform_script_id", "00000000-0000-0000-0000-000000000002"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_windows_platform_script_assignment.test", "target.target_type", "allDevices"),
				),
			},
		},
	})
}
*/
