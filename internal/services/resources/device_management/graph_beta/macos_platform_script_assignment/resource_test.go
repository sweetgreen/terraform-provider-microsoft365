package graphBetaMacosPlatformScriptAssignment_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jarcoal/httpmock"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/mocks"
	localMocks "github.com/sweetgreen/terraform-provider-microsoft365/internal/services/resources/device_management/graph_beta/macos_platform_script_assignment/mocks"
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
	return unitTestProviderConfig + string(content)
}

func testConfigMaximal() string {
	content, err := os.ReadFile(filepath.Join("mocks", "terraform", "resource_maximal.tf"))
	if err != nil {
		return ""
	}
	return unitTestProviderConfig + string(content)
}

func testConfigMinimalToMaximal() string {
	// For minimal to maximal test, we need to use the maximal config
	// but with the minimal resource name and script_id to simulate an update

	// Read the maximal config
	maximalContent, err := os.ReadFile(filepath.Join("mocks", "terraform", "resource_maximal.tf"))
	if err != nil {
		return ""
	}

	// Replace the resource name to match the minimal one
	updatedMaximal := strings.Replace(string(maximalContent), "test_maximal", "test_minimal", 1)

	// Replace the script_id to match the minimal one
	updatedMaximal = strings.Replace(updatedMaximal, "00000000-0000-0000-0000-000000000004", "00000000-0000-0000-0000-000000000003", 1)

	return unitTestProviderConfig + updatedMaximal
}

// Helper function to set up the test environment
func setupTestEnvironment(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("TF_ACC", "0")
	os.Setenv("MS365_TEST_MODE", "true")
}

// Helper function to set up the mock environment
func setupMockEnvironment() (*mocks.Mocks, *localMocks.MacosPlatformScriptAssignmentMock) {
	// Activate httpmock
	httpmock.Activate()

	// Create a new Mocks instance and register authentication mocks
	mockClient := mocks.NewMocks()
	mockClient.AuthMocks.RegisterMocks()

	// Register local mocks directly
	scriptAssignmentMock := &localMocks.MacosPlatformScriptAssignmentMock{}
	scriptAssignmentMock.RegisterMocks()

	return mockClient, scriptAssignmentMock
}

// Helper function to check if a resource exists
func testCheckExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		return nil
	}
}

// TestUnitMacosPlatformScriptAssignmentResource_Create_Minimal tests the creation of a script assignment with minimal configuration
func TestUnitMacosPlatformScriptAssignmentResource_Create_Minimal(t *testing.T) {
	// Set up mock environment
	_, _ = setupMockEnvironment()
	defer httpmock.DeactivateAndReset()

	// Set up the test environment
	setupTestEnvironment(t)

	// Run the test
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfigMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testCheckExists("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "macos_platform_script_id", "00000000-0000-0000-0000-000000000003"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "target.target_type", "allDevices"),
					resource.TestCheckResourceAttrSet("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "id"),
				),
			},
		},
	})
}

// TestUnitMacosPlatformScriptAssignmentResource_Create_Maximal tests the creation of a script assignment with maximal configuration
func TestUnitMacosPlatformScriptAssignmentResource_Create_Maximal(t *testing.T) {
	// Set up mock environment
	_, _ = setupMockEnvironment()
	defer httpmock.DeactivateAndReset()

	// Set up the test environment
	setupTestEnvironment(t)

	// Run the test
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfigMaximal(),
				Check: resource.ComposeTestCheckFunc(
					testCheckExists("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_maximal"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_maximal", "macos_platform_script_id", "00000000-0000-0000-0000-000000000004"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_maximal", "target.target_type", "groupAssignment"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_maximal", "target.group_id", "44444444-4444-4444-4444-444444444444"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_maximal", "target.device_and_app_management_assignment_filter_id", "55555555-5555-5555-5555-555555555555"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_maximal", "target.device_and_app_management_assignment_filter_type", "include"),
					resource.TestCheckResourceAttrSet("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_maximal", "id"),
				),
			},
		},
	})
}

// TestUnitMacosPlatformScriptAssignmentResource_Update_MinimalToMaximal tests updating from minimal to maximal configuration
func TestUnitMacosPlatformScriptAssignmentResource_Update_MinimalToMaximal(t *testing.T) {
	// Set up mock environment
	_, _ = setupMockEnvironment()
	defer httpmock.DeactivateAndReset()

	// Set up the test environment
	setupTestEnvironment(t)

	// Run the test
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Start with minimal configuration
			{
				Config: testConfigMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testCheckExists("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "macos_platform_script_id", "00000000-0000-0000-0000-000000000003"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "target.target_type", "allDevices"),
				),
			},
			// Update to maximal configuration (with the same resource name)
			{
				Config: testConfigMinimalToMaximal(),
				Check: resource.ComposeTestCheckFunc(
					testCheckExists("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "macos_platform_script_id", "00000000-0000-0000-0000-000000000003"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "target.target_type", "groupAssignment"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "target.group_id", "44444444-4444-4444-4444-444444444444"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "target.device_and_app_management_assignment_filter_id", "55555555-5555-5555-5555-555555555555"),
					resource.TestCheckResourceAttr("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal", "target.device_and_app_management_assignment_filter_type", "include"),
				),
			},
		},
	})
}

// TestUnitMacosPlatformScriptAssignmentResource_Delete tests the deletion of a script assignment
func TestUnitMacosPlatformScriptAssignmentResource_Delete(t *testing.T) {
	// Set up mock environment
	_, _ = setupMockEnvironment()
	defer httpmock.DeactivateAndReset()

	// Set up the test environment
	setupTestEnvironment(t)

	// Run the test
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a script assignment
			{
				Config: testConfigMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testCheckExists("microsoft365_graph_beta_device_management_macos_platform_script_assignment.test_minimal"),
				),
			},
			// Delete the script assignment by removing the configuration
			{
				Config: unitTestProviderConfig,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

// TestUnitMacosPlatformScriptAssignmentResource_Import tests the import functionality
func TestUnitMacosPlatformScriptAssignmentResource_Import(t *testing.T) {
	// Skip import test for now - this resource requires a composite import ID
	// which is not yet implemented
	t.Skip("Import functionality requires composite ID implementation")
}

// Helper function to get the import ID
func testImportStateId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}

// TestUnitMacosPlatformScriptAssignmentResource_Error tests error handling
func TestUnitMacosPlatformScriptAssignmentResource_Error(t *testing.T) {
	// Set up mock environment
	_, scriptAssignmentMock := setupMockEnvironment()
	defer httpmock.DeactivateAndReset()

	// Register error mocks
	scriptAssignmentMock.RegisterErrorMocks()

	// Set up the test environment
	setupTestEnvironment(t)

	// Run the test
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_macos_platform_script_assignment" "test_error" {
  macos_platform_script_id = "99999999-9999-9999-9999-999999999999"
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

// TestUnitMacosPlatformScriptAssignmentResource_NotFoundScript tests handling of not found script
func TestUnitMacosPlatformScriptAssignmentResource_NotFoundScript(t *testing.T) {
	// Set up mock environment
	_, scriptAssignmentMock := setupMockEnvironment()
	defer httpmock.DeactivateAndReset()

	// Register error mocks
	scriptAssignmentMock.RegisterErrorMocks()

	// Set up the test environment
	setupTestEnvironment(t)

	// Run the test
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: unitTestProviderConfig + `
resource "microsoft365_graph_beta_device_management_macos_platform_script_assignment" "test_not_found" {
  macos_platform_script_id = "ffffffff-ffff-ffff-ffff-ffffffffffff"
  target = {
    target_type = "allDevices"
  }
}
`,
				ExpectError: regexp.MustCompile("Script not found"),
			},
		},
	})
}
