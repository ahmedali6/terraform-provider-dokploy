package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegistryResource(t *testing.T) {
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
				Config: testAccRegistryResourceConfig("test-registry-project", "test-registry-env", "test-registry-app", "docker.io", "testuser"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_registry.test", "registry_url", "docker.io"),
					resource.TestCheckResourceAttr("dokploy_registry.test", "username", "testuser"),
					resource.TestCheckResourceAttrSet("dokploy_registry.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_registry.test", "application_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccRegistryResourceConfig("test-registry-project", "test-registry-env", "test-registry-app", "ghcr.io", "updateduser"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_registry.test", "registry_url", "ghcr.io"),
					resource.TestCheckResourceAttr("dokploy_registry.test", "username", "updateduser"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_registry.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccRegistryResourceConfig(projectName, envName, appName, registryURL, username string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for registry tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  project_id     = dokploy_project.test.id
  environment_id = dokploy_environment.test.id
  name           = "%s"
  build_type     = "nixpacks"
}

resource "dokploy_registry" "test" {
  application_id = dokploy_application.test.id
  registry_url   = "%s"
  username       = "%s"
  password       = "test_password_123"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, registryURL, username)
}
