package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMariaDBResource(t *testing.T) {
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
				Config: testAccMariaDBResourceConfig("test-mariadb-project", "test-mariadb-env", "test-mariadb", "testmariadbapp", "testdb", "testuser"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mariadb.test", "name", "test-mariadb"),
					resource.TestCheckResourceAttrSet("dokploy_mariadb.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_mariadb.test", "environment_id"),
					resource.TestCheckResourceAttr("dokploy_mariadb.test", "database_name", "testdb"),
					resource.TestCheckResourceAttr("dokploy_mariadb.test", "database_user", "testuser"),
					resource.TestCheckResourceAttr("dokploy_mariadb.test", "app_name_prefix", "testmariadbapp"),
					resource.TestCheckResourceAttrSet("dokploy_mariadb.test", "app_name"),
					resource.TestCheckResourceAttrSet("dokploy_mariadb.test", "internal_connection"),
				),
			},
			// Update and Read testing
			{
				Config: testAccMariaDBResourceConfigWithDescription("test-mariadb-project", "test-mariadb-env", "test-mariadb-updated", "testmariadbapp", "testdb", "testuser", "Updated MariaDB instance"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mariadb.test", "name", "test-mariadb-updated"),
					resource.TestCheckResourceAttr("dokploy_mariadb.test", "description", "Updated MariaDB instance"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_mariadb.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_password", "database_root_password", "app_name_prefix", "internal_connection", "external_connection"},
			},
		},
	})
}

func testAccMariaDBResourceConfig(projectName, envName, mariadbName, appNamePrefix, dbName, dbUser string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for MariaDB tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_mariadb" "test" {
  name                   = "%s"
  app_name_prefix        = "%s"
  database_name          = "%s"
  database_user          = "%s"
  database_password      = "test_mariadb_password_123"
  database_root_password = "test_mariadb_root_password_123"
  environment_id         = dokploy_environment.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, mariadbName, appNamePrefix, dbName, dbUser)
}

func testAccMariaDBResourceConfigWithDescription(projectName, envName, mariadbName, appNamePrefix, dbName, dbUser, description string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for MariaDB tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_mariadb" "test" {
  name                   = "%s"
  app_name_prefix        = "%s"
  database_name          = "%s"
  database_user          = "%s"
  database_password      = "test_mariadb_password_123"
  database_root_password = "test_mariadb_root_password_123"
  environment_id         = dokploy_environment.test.id
  description            = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, mariadbName, appNamePrefix, dbName, dbUser, description)
}
