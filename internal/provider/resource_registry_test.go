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
	dockerUsername := os.Getenv("DOCKER_USERNAME")
	dockerPassword := os.Getenv("DOCKER_PASSWORD")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	if dockerUsername == "" || dockerPassword == "" {
		t.Skip("DOCKER_USERNAME and DOCKER_PASSWORD must be set for registry tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRegistryResourceConfig("test-registry-project", "test-registry-env", "test-registry-app", "test-registry", "docker.io", dockerUsername, dockerPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_registry.test", "registry_url", "docker.io"),
					resource.TestCheckResourceAttr("dokploy_registry.test", "username", dockerUsername),
					resource.TestCheckResourceAttr("dokploy_registry.test", "registry_name", "test-registry"),
					resource.TestCheckResourceAttrSet("dokploy_registry.test", "id"),
				),
			},
			// Update and Read testing - change registry name
			{
				Config: testAccRegistryResourceConfig("test-registry-project", "test-registry-env", "test-registry-app", "updated-registry", "docker.io", dockerUsername, dockerPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_registry.test", "registry_url", "docker.io"),
					resource.TestCheckResourceAttr("dokploy_registry.test", "username", dockerUsername),
					resource.TestCheckResourceAttr("dokploy_registry.test", "registry_name", "updated-registry"),
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

func testAccRegistryResourceConfig(projectName, envName, appName, registryName, registryURL, username, password string) string {
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
  environment_id = dokploy_environment.test.id
  name           = "%s"
  build_type     = "nixpacks"
  source_type    = "docker"
  docker_image   = "nginx:latest"
}

resource "dokploy_registry" "test" {
  registry_name = "%s"
  registry_url  = "%s"
  username      = "%s"
  password      = "%s"
  image_prefix  = "%s/test"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, registryName, registryURL, username, password, registryURL)
}
