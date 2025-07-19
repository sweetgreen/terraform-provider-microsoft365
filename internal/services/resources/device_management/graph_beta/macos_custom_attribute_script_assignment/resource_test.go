package graphBetaMacosCustomAttributeScriptAssignment_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jarcoal/httpmock"
	frameworkMocks "github.com/sweetgreen/terraform-provider-microsoft365/internal/mocks"
	localMocks "github.com/sweetgreen/terraform-provider-microsoft365/internal/services/resources/device_management/graph_beta/macos_custom_attribute_script_assignment/mocks"
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

// Helper functions to return the test configurations by reading from files
func testConfigMinimal() string {
	content, err := os.ReadFile(filepath.Join("mocks", "terraform", "resource_minimal.tf"))
	if err != nil {
		return ""
	}
	return string(content)
}

func testConfigMaximal() string {
	content, err := os.ReadFile(filepath.Join("mocks", "terraform", "resource_maximal.tf"))
	if err != nil {
		return ""
	}
	return string(content)
}

// setupMockEnvironment initializes the HTTP mock environment for testing
func setupMockEnvironment() {
	httpmock.Activate()
	httpmock.Reset()

	// Create a new Mocks instance and register authentication mocks
	mockClient := frameworkMocks.NewMocks()
	mockClient.AuthMocks.RegisterMocks()

	// Register resource-specific mocks
	macosCustomAttributeScriptAssignmentMock := &localMocks.MacosCustomAttributeScriptAssignmentMock{}
	macosCustomAttributeScriptAssignmentMock.RegisterMocks()
	macosCustomAttributeScriptAssignmentMock.RegisterErrorMocks()
}

// setupTestEnvironment prepares the test environment
func setupTestEnvironment(t *testing.T) func() {
	if os.Getenv("TF_ACC") != "" {
		t.Skip("Skipping unit test in acceptance test mode")
	}

	setupMockEnvironment()

	return func() {
		httpmock.DeactivateAndReset()
	}
}

func TestUnitMacosCustomAttributeScriptAssignmentResource_Create(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
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
					resource.TestCheckResourceAttrSet("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "id"),
				),
			},
		},
	})
}

func TestUnitMacosCustomAttributeScriptAssignmentResource_CreateWithGroup(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000002"
  target = {
    target_type = "groupAssignment"
    group_id = "22222222-2222-2222-2222-222222222222"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "macos_custom_attribute_script_id", "00000000-0000-0000-0000-000000000002"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.target_type", "groupAssignment"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.group_id", "22222222-2222-2222-2222-222222222222"),
					resource.TestCheckResourceAttrSet("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "id"),
				),
			},
		},
	})
}

func TestUnitMacosCustomAttributeScriptAssignmentResource_CreateWithFilter(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000002"
  target = {
    target_type = "groupAssignment"
    group_id = "22222222-2222-2222-2222-222222222222"
    device_and_app_management_assignment_filter_id = "33333333-3333-3333-3333-333333333333"
    device_and_app_management_assignment_filter_type = "include"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "macos_custom_attribute_script_id", "00000000-0000-0000-0000-000000000002"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.target_type", "groupAssignment"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.group_id", "22222222-2222-2222-2222-222222222222"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.device_and_app_management_assignment_filter_id", "33333333-3333-3333-3333-333333333333"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.device_and_app_management_assignment_filter_type", "include"),
					resource.TestCheckResourceAttrSet("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "id"),
				),
			},
		},
	})
}

func TestUnitMacosCustomAttributeScriptAssignmentResource_Update(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
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
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.target_type", "allDevices"),
				),
			},
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000002"
  target = {
    target_type = "allLicensedUsers"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test", "target.target_type", "allLicensedUsers"),
				),
			},
		},
	})
}

func TestUnitMacosCustomAttributeScriptAssignmentResource_MinimalConfig(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + testConfigMinimal(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_minimal", "macos_custom_attribute_script_id", "00000000-0000-0000-0000-000000000002"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_minimal", "target.target_type", "allDevices"),
					resource.TestCheckResourceAttrSet("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_minimal", "id"),
				),
			},
		},
	})
}

func TestUnitMacosCustomAttributeScriptAssignmentResource_MaximalConfig(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + testConfigMaximal(),
				Check: resource.ComposeTestCheckFunc(
					// Check first assignment
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_maximal", "macos_custom_attribute_script_id", "00000000-0000-0000-0000-000000000004"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_maximal", "target.target_type", "groupAssignment"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_maximal", "target.group_id", "22222222-2222-2222-2222-222222222222"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_maximal", "target.device_and_app_management_assignment_filter_id", "33333333-3333-3333-3333-333333333333"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_maximal", "target.device_and_app_management_assignment_filter_type", "include"),
					resource.TestCheckResourceAttrSet("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_maximal", "id"),

					// Check all licensed users assignment
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_all_licensed_users", "target.target_type", "allLicensedUsers"),

					// Check exclude filter assignment
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test_exclude_filter", "target.device_and_app_management_assignment_filter_type", "exclude"),
				),
			},
		},
	})
}

func TestUnitMacosCustomAttributeScriptAssignmentResource_ErrorHandling(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test" {
  macos_custom_attribute_script_id = "99999999-9999-9999-9999-999999999999"
  target = {
    target_type = "allDevices"
  }
}
`,
				ExpectError: regexp.MustCompile("Error creating assignment"),
			},
		},
	})
}

// TODO: Import state test requires additional mock setup to handle the Read operation after import
// func TestUnitMacosCustomAttributeScriptAssignmentResource_ImportState(t *testing.T) {
// 	cleanup := setupTestEnvironment(t)
// 	defer cleanup()

// 	resource.Test(t, resource.TestCase{
// 		IsUnitTest:               true,
// 		ProtoV6ProviderFactories: frameworkMocks.TestUnitTestProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: unitTestProviderConfig + `
// resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test" {
//   macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000002"
//   target = {
//     target_type = "allDevices"
//   }
// }
// `,
// 			},
// 			{
// 				ResourceName:      "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test",
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 				ImportStateIdFunc: func(s *terraform.State) (string, error) {
// 					rs, ok := s.RootModule().Resources["microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment.test"]
// 					if !ok {
// 						return "", fmt.Errorf("resource not found")
// 					}
// 					scriptId := rs.Primary.Attributes["macos_custom_attribute_script_id"]
// 					assignmentId := rs.Primary.ID
// 					return fmt.Sprintf("%s/%s", scriptId, assignmentId), nil
// 				},
// 			},
// 		},
// 	})
// }
