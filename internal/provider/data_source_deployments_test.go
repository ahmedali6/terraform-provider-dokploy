package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentsDataSource_Application(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentsDataSourceConfig_Application(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_deployments.test", "deployments.#"),
				),
			},
		},
	})
}

func TestAccDeploymentsDataSource_Compose(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentsDataSourceConfig_Compose(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_deployments.test", "deployments.#"),
				),
			},
		},
	})
}

func testAccDeploymentsDataSourceConfig_Application() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "test-deployments-ds-project"
  description = "Test project for deployments data source"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "test-deployments-ds-env"
}

resource "dokploy_application" "test" {
  environment_id = dokploy_environment.test.id
  name           = "test-deployments-ds-app"
  source_type    = "docker"
  docker_image   = "nginx:latest"
}

data "dokploy_deployments" "test" {
  application_id = dokploy_application.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"))
}

func testAccDeploymentsDataSourceConfig_Compose() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "test-deployments-ds-compose-project"
  description = "Test project for compose deployments data source"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "test-deployments-ds-compose-env"
}

resource "dokploy_compose" "test" {
  environment_id       = dokploy_environment.test.id
  name                 = "test-deployments-ds-compose"
  source_type          = "raw"
  compose_file_content = <<-EOT
version: '3.8'
services:
  web:
    image: nginx:latest
EOT
}

data "dokploy_deployments" "test" {
  compose_id = dokploy_compose.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"))
}
