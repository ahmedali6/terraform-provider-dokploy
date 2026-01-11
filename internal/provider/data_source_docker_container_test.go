package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDockerContainersDataSource_basic(t *testing.T) {
	// Skip in CI - requires running Dokploy instance with Docker containers
	t.Skip("Skipping acceptance test - requires running Dokploy instance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDockerContainersDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_docker_containers.test", "containers.#"),
				),
			},
		},
	})
}

const testAccDockerContainersDataSourceConfig_basic = `
data "dokploy_docker_containers" "test" {
}
`

func TestAccDockerContainersDataSource_withServerID(t *testing.T) {
	// Skip in CI - requires running Dokploy instance with remote server
	t.Skip("Skipping acceptance test - requires running Dokploy instance with remote server")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDockerContainersDataSourceConfig_withServerID,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_docker_containers.test", "containers.#"),
				),
			},
		},
	})
}

const testAccDockerContainersDataSourceConfig_withServerID = `
data "dokploy_docker_containers" "test" {
  server_id = "test-server-id"
}
`

func TestAccDockerContainersDataSource_withAppName(t *testing.T) {
	// Skip in CI - requires running Dokploy instance with matching app
	t.Skip("Skipping acceptance test - requires running Dokploy instance with matching app")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDockerContainersDataSourceConfig_withAppName,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_docker_containers.test", "containers.#"),
				),
			},
		},
	})
}

const testAccDockerContainersDataSourceConfig_withAppName = `
data "dokploy_docker_containers" "test" {
  app_name = "test-app"
  app_type = "application"
}
`

func TestAccDockerContainerDataSource_basic(t *testing.T) {
	// Skip in CI - requires running Dokploy instance with a known container
	t.Skip("Skipping acceptance test - requires running Dokploy instance with a known container")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDockerContainerDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_docker_container.test", "id"),
					resource.TestCheckResourceAttrSet("data.dokploy_docker_container.test", "name"),
					resource.TestCheckResourceAttrSet("data.dokploy_docker_container.test", "state_status"),
					resource.TestCheckResourceAttrSet("data.dokploy_docker_container.test", "config_json"),
				),
			},
		},
	})
}

const testAccDockerContainerDataSourceConfig_basic = `
# First get the list of containers
data "dokploy_docker_containers" "all" {
}

# Then get details for the first container
data "dokploy_docker_container" "test" {
  container_id = data.dokploy_docker_containers.all.containers[0].container_id
}
`

func TestAccDockerContainerDataSource_withServerID(t *testing.T) {
	// Skip in CI - requires running Dokploy instance with remote server
	t.Skip("Skipping acceptance test - requires running Dokploy instance with remote server")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDockerContainerDataSourceConfig_withServerID,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_docker_container.test", "id"),
					resource.TestCheckResourceAttrSet("data.dokploy_docker_container.test", "config_json"),
				),
			},
		},
	})
}

const testAccDockerContainerDataSourceConfig_withServerID = `
data "dokploy_docker_container" "test" {
  container_id = "test-container-id"
  server_id    = "test-server-id"
}
`
