package graphBetaGroup_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/mocks"
)

func TestAccGroupDataSource_Basic(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.microsoft365_graph_beta_groups_group.test", "groups.#"),
				),
			},
		},
	})
}

func TestAccGroupDataSource_FilterByID(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_FilterByID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.microsoft365_graph_beta_groups_group.test", "groups.#", "1"),
					resource.TestCheckResourceAttrSet("data.microsoft365_graph_beta_groups_group.test", "groups.0.display_name"),
				),
			},
		},
	})
}

func TestAccGroupDataSource_FilterByDisplayName(t *testing.T) {
	// Skip if not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: mocks.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_FilterByDisplayName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.microsoft365_graph_beta_groups_group.test", "groups.#"),
					resource.TestCheckResourceAttr("data.microsoft365_graph_beta_groups_group.test", "groups.0.display_name", "Test Group"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig_Basic() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

data "microsoft365_graph_beta_groups_group" "test" {
  filter_type = "all"
}
`)
}

func testAccGroupDataSourceConfig_FilterByID() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

resource "microsoft365_graph_beta_groups_group" "test" {
  display_name    = "Test Group for Data Source"
  mail_nickname   = "testgroupds"
  mail_enabled    = false
  security_enabled = true
}

data "microsoft365_graph_beta_groups_group" "test" {
  filter_type  = "id"
  filter_value = microsoft365_graph_beta_groups_group.test.id
}
`)
}

func testAccGroupDataSourceConfig_FilterByDisplayName() string {
	return fmt.Sprintf(`
provider "microsoft365" {
}

resource "microsoft365_graph_beta_groups_group" "test" {
  display_name    = "Test Group"
  mail_nickname   = "testgroup"
  mail_enabled    = false
  security_enabled = true
}

data "microsoft365_graph_beta_groups_group" "test" {
  filter_type  = "display_name"
  filter_value = "Test Group"
  
  depends_on = [microsoft365_graph_beta_groups_group.test]
}
`)
}
