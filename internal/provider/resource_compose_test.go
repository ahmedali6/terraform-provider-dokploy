package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccComposeResource(t *testing.T) {
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
				Config: testAccComposeResourceConfig("test-compose-project", "test-env", "test-compose", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-compose"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "project_id"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "environment_id"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "deploy_on_create", "false"),
				),
			},
			// Update and Read testing
			{
				Config: testAccComposeResourceConfig("test-compose-project", "test-env", "test-compose-updated", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-compose-updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_compose.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deploy_on_create"},
			},
		},
	})
}

func testAccComposeResourceConfig(projectName, envName, composeName string, deployOnCreate bool) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for compose tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  project_id     = dokploy_project.test.id
  environment_id = dokploy_environment.test.id
  name           = "%s"
  source_type    = "raw"
  compose_file_content = <<EOF
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
EOF
  deploy_on_create = %t
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, composeName, deployOnCreate)
}
