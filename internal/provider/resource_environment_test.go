package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccEnvironmentResource(t *testing.T) {
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
				Config: testAccEnvironmentResourceConfig("test-env-project", "staging"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment.test", "name", "staging"),
					resource.TestCheckResourceAttrSet("dokploy_environment.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_environment.test", "project_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccEnvironmentResourceConfig("test-env-project", "production"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment.test", "name", "production"),
				),
			},
			// ImportState testing with composite ID (project_id:environment_id)
			{
				ResourceName:      "dokploy_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["dokploy_environment.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}

					projectID := rs.Primary.Attributes["project_id"]
					environmentID := rs.Primary.ID

					// Format: project_id:environment_id
					return fmt.Sprintf("%s:%s", projectID, environmentID), nil
				},
			},
		},
	})
}

func testAccEnvironmentResourceConfig(projectName, envName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for environment tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName)
}

// TestAccEnvironmentResourceWithDescription tests environment with description field.
func TestAccEnvironmentResourceWithDescription(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with description
			{
				Config: testAccEnvironmentResourceWithDescConfig("test-env-desc-project", "dev-env", "Development environment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment.test", "name", "dev-env"),
					resource.TestCheckResourceAttr("dokploy_environment.test", "description", "Development environment"),
					resource.TestCheckResourceAttrSet("dokploy_environment.test", "id"),
				),
			},
			// Update description
			{
				Config: testAccEnvironmentResourceWithDescConfig("test-env-desc-project", "dev-env", "Updated development environment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment.test", "name", "dev-env"),
					resource.TestCheckResourceAttr("dokploy_environment.test", "description", "Updated development environment"),
				),
			},
		},
	})
}

func testAccEnvironmentResourceWithDescConfig(projectName, envName, description string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for environment with description"
}

resource "dokploy_environment" "test" {
  project_id  = dokploy_project.test.id
  name        = "%s"
  description = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, description)
}
