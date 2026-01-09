package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortResource(t *testing.T) {
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
				Config: testAccPortResourceConfig("test-port-project", "test-port-env", "test-port-app", 8080, 3000, "tcp"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_port.test", "published_port", "8080"),
					resource.TestCheckResourceAttr("dokploy_port.test", "target_port", "3000"),
					resource.TestCheckResourceAttr("dokploy_port.test", "protocol", "tcp"),
					resource.TestCheckResourceAttrSet("dokploy_port.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_port.test", "application_id"),
				),
			},
			// Update testing - change target_port (in-place update, not replace)
			{
				Config: testAccPortResourceConfig("test-port-project", "test-port-env", "test-port-app", 8080, 4000, "tcp"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_port.test", "published_port", "8080"),
					resource.TestCheckResourceAttr("dokploy_port.test", "target_port", "4000"),
					resource.TestCheckResourceAttr("dokploy_port.test", "protocol", "tcp"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_port.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPortResourceConfig(projectName, envName, appName string, publishedPort, targetPort int, protocol string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for port tests"
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

resource "dokploy_port" "test" {
  application_id = dokploy_application.test.id
  published_port = %d
  target_port    = %d
  protocol       = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, publishedPort, targetPort, protocol)
}
