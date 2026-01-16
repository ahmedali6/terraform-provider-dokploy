package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMySQLResource(t *testing.T) {
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
				Config: testAccMySQLResourceConfig("test-mysql-project", "test-mysql-env", "test-mysql", "testmysqlapp", "testdb", "testuser"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mysql.test", "name", "test-mysql"),
					resource.TestCheckResourceAttrSet("dokploy_mysql.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_mysql.test", "environment_id"),
					resource.TestCheckResourceAttr("dokploy_mysql.test", "database_name", "testdb"),
					resource.TestCheckResourceAttr("dokploy_mysql.test", "database_user", "testuser"),
					resource.TestCheckResourceAttr("dokploy_mysql.test", "app_name_prefix", "testmysqlapp"),
					resource.TestCheckResourceAttrSet("dokploy_mysql.test", "app_name"),
					resource.TestCheckResourceAttrSet("dokploy_mysql.test", "internal_connection"),
				),
			},
			// Update and Read testing
			{
				Config: testAccMySQLResourceConfigWithDescription("test-mysql-project", "test-mysql-env", "test-mysql-updated", "testmysqlapp", "testdb", "testuser", "Updated MySQL instance"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mysql.test", "name", "test-mysql-updated"),
					resource.TestCheckResourceAttr("dokploy_mysql.test", "description", "Updated MySQL instance"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_mysql.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_password", "database_root_password", "app_name_prefix", "internal_connection", "external_connection"},
			},
		},
	})
}

func testAccMySQLResourceConfig(projectName, envName, mysqlName, appNamePrefix, dbName, dbUser string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for MySQL tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_mysql" "test" {
  name                   = "%s"
  app_name_prefix        = "%s"
  database_name          = "%s"
  database_user          = "%s"
  database_password      = "test_mysql_password_123"
  database_root_password = "test_mysql_root_password_123"
  environment_id         = dokploy_environment.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, mysqlName, appNamePrefix, dbName, dbUser)
}

func testAccMySQLResourceConfigWithDescription(projectName, envName, mysqlName, appNamePrefix, dbName, dbUser, description string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for MySQL tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_mysql" "test" {
  name                   = "%s"
  app_name_prefix        = "%s"
  database_name          = "%s"
  database_user          = "%s"
  database_password      = "test_mysql_password_123"
  database_root_password = "test_mysql_root_password_123"
  environment_id         = dokploy_environment.test.id
  description            = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, mysqlName, appNamePrefix, dbName, dbUser, description)
}
