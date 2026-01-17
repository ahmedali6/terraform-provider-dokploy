package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig("Test Project", "Initial Description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", "Test Project"),
					resource.TestCheckResourceAttr("dokploy_project.test", "description", "Initial Description"),
					resource.TestCheckResourceAttrSet("dokploy_project.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccProjectResourceConfig("Test Project Updated", "Updated Description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", "Test Project Updated"),
					resource.TestCheckResourceAttr("dokploy_project.test", "description", "Updated Description"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectResourceConfig(name, description string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "%s"
}
 `, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, description)
}

func TestAccProjectResourceMinimal(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without description
			{
				Config: testAccProjectResourceConfigNoDescription("test-project-minimal"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", "test-project-minimal"),
					resource.TestCheckResourceAttr("dokploy_project.test", "description", ""),
					resource.TestCheckResourceAttrSet("dokploy_project.test", "id"),
				),
			},
			// Update with description
			{
				Config: testAccProjectResourceConfig("test-project-minimal", "Added Description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", "test-project-minimal"),
					resource.TestCheckResourceAttr("dokploy_project.test", "description", "Added Description"),
				),
			},
		},
	})
}

func testAccProjectResourceConfigNoDescription(name string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name = "%s"
}
 `, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name)
}
