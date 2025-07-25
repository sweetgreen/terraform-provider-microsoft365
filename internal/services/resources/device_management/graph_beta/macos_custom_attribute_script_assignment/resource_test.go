package graphBetaMacosCustomAttributeScriptAssignment_test

import (
	"testing"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/mocks"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

func TestUnitMacosCustomAttributeScriptAssignmentResourceModel_Basic(t *testing.T) {
	t.Skip("Skipping test - mock implementation not available for this resource")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000002"
  target = {
    target_type = "allDevices"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "macos_custom_attribute_script_id", "00000000-0000-0000-0000-000000000002"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.target_type", "allDevices"),
				),
			},
		},
	})
}
