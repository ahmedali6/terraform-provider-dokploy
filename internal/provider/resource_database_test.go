package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabaseResource(t *testing.T) {
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
				Config: testAccDatabaseResourceConfig("test-db-project", "test-db-env", "test-postgres-db", "postgres", "16"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_database.test", "name", "test-postgres-db"),
					resource.TestCheckResourceAttr("dokploy_database.test", "type", "postgres"),
					resource.TestCheckResourceAttr("dokploy_database.test", "version", "16"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "project_id"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "environment_id"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "app_name"),
				),
			},
			// Update testing - changing name should trigger replacement
			{
				Config: testAccDatabaseResourceConfig("test-db-project", "test-db-env", "updated-postgres-db", "postgres", "16"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_database.test", "name", "updated-postgres-db"),
					resource.TestCheckResourceAttr("dokploy_database.test", "type", "postgres"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "project_id", "version", "app_name"}, // project_id, version, app_name sometimes not returned on import
			},
		},
	})
}

func TestAccDatabaseResourceMySQL(t *testing.T) {
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
				Config: testAccDatabaseResourceConfig("test-db-mysql-project", "test-db-mysql-env", "test-mysql-db", "mysql", "8"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_database.test", "name", "test-mysql-db"),
					resource.TestCheckResourceAttr("dokploy_database.test", "type", "mysql"),
					resource.TestCheckResourceAttr("dokploy_database.test", "version", "8"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "app_name"),
				),
			},
			// Update testing - changing name should trigger replacement
			{
				Config: testAccDatabaseResourceConfig("test-db-mysql-project", "test-db-mysql-env", "updated-mysql-db", "mysql", "8"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_database.test", "name", "updated-mysql-db"),
					resource.TestCheckResourceAttr("dokploy_database.test", "type", "mysql"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "project_id", "version", "app_name"},
			},
		},
	})
}

func TestAccDatabaseResourceMongoDB(t *testing.T) {
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
				Config: testAccDatabaseResourceConfig("test-db-mongo-project", "test-db-mongo-env", "test-mongo-db", "mongo", "7"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_database.test", "name", "test-mongo-db"),
					resource.TestCheckResourceAttr("dokploy_database.test", "type", "mongo"),
					resource.TestCheckResourceAttr("dokploy_database.test", "version", "7"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "app_name"),
				),
			},
			// Update testing - changing name should trigger replacement
			{
				Config: testAccDatabaseResourceConfig("test-db-mongo-project", "test-db-mongo-env", "updated-mongo-db", "mongo", "7"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_database.test", "name", "updated-mongo-db"),
					resource.TestCheckResourceAttr("dokploy_database.test", "type", "mongo"),
					resource.TestCheckResourceAttrSet("dokploy_database.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "project_id", "version", "app_name"},
			},
		},
	})
}

func testAccDatabaseResourceConfig(projectName, envName, dbName, dbType, version string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for database tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_database" "test" {
  project_id     = dokploy_project.test.id
  environment_id = dokploy_environment.test.id
  name           = "%s"
  type           = "%s"
  password       = "test_password_123"
  version        = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, dbName, dbType, version)
}
