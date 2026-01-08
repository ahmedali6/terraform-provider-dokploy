package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMountResource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - volume mount
			{
				Config: testAccMountResourceVolumeConfig("test-mount-project", "test-mount-env", "test-mount-app", "test-data"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mount.test", "type", "volume"),
					resource.TestCheckResourceAttr("dokploy_mount.test", "volume_name", "test-data"),
					resource.TestCheckResourceAttr("dokploy_mount.test", "mount_path", "/data"),
					resource.TestCheckResourceAttrSet("dokploy_mount.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_mount.test", "application_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_mount.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMountResourceBind(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - bind mount
			{
				Config: testAccMountResourceBindConfig("test-bind-mount-project", "test-bind-mount-env", "test-bind-mount-app"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mount.test", "type", "bind"),
					resource.TestCheckResourceAttr("dokploy_mount.test", "host_path", "/host/path"),
					resource.TestCheckResourceAttr("dokploy_mount.test", "mount_path", "/container/path"),
					resource.TestCheckResourceAttrSet("dokploy_mount.test", "id"),
				),
			},
		},
	})
}

func testAccMountResourceVolumeConfig(projectName, envName, appName, volumeName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for mount tests"
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

resource "dokploy_mount" "test" {
  application_id = dokploy_application.test.id
  type           = "volume"
  volume_name    = "%s"
  mount_path     = "/data"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, volumeName)
}

func testAccMountResourceBindConfig(projectName, envName, appName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for bind mount tests"
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

resource "dokploy_mount" "test" {
  application_id = dokploy_application.test.id
  type           = "bind"
  host_path      = "/host/path"
  mount_path     = "/container/path"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName)
}
