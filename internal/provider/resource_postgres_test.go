package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPostgresResource(t *testing.T) {
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
				Config: testAccPostgresResourceConfig("test-postgres-project", "test-postgres-env", "test-postgres", "testpgapp", "testdb", "testuser"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_postgres.test", "name", "test-postgres"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "environment_id"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "database_name", "testdb"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "database_user", "testuser"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "app_name_prefix", "testpgapp"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "app_name"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "internal_connection"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "internal_port"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "internal_port", "5432"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "external_connection", ""),
				),
			},
			// Update and Read testing
			{
				Config: testAccPostgresResourceConfigWithDescription("test-postgres-project", "test-postgres-env", "test-postgres-updated", "testpgapp", "testdb", "testuser", "Updated PostgreSQL instance"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_postgres.test", "name", "test-postgres-updated"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "description", "Updated PostgreSQL instance"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_postgres.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_password", "app_name_prefix", "internal_connection", "external_connection"},
			},
		},
	})
}

func testAccPostgresResourceConfig(projectName, envName, pgName, appNamePrefix, dbName, dbUser string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for PostgreSQL tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_postgres" "test" {
  name              = "%s"
  app_name_prefix    = "%s"
  database_name     = "%s"
  database_user     = "%s"
  database_password = "test_postgres_password_123"
  environment_id    = dokploy_environment.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, pgName, appNamePrefix, dbName, dbUser)
}

func testAccPostgresResourceConfigWithDescription(projectName, envName, pgName, appNamePrefix, dbName, dbUser, description string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for PostgreSQL tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_postgres" "test" {
  name              = "%s"
  app_name_prefix    = "%s"
  database_name     = "%s"
  database_user     = "%s"
  database_password = "test_postgres_password_123"
  environment_id    = dokploy_environment.test.id
  description       = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, pgName, appNamePrefix, dbName, dbUser, description)
}

// TestAccPostgresResourceExtended tests PostgreSQL with extended settings.
func TestAccPostgresResourceExtended(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with extended settings
			{
				Config: testAccPostgresResourceExtendedConfig("test-pg-ext-project", "test-pg-ext-env", "test-pg-ext", "testpgext", "testdb", "testuser", "128", "256"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_postgres.test", "name", "test-pg-ext"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "memory_reservation", "128"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "memory_limit", "256"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "id"),
				),
			},
			// Update extended settings
			{
				Config: testAccPostgresResourceExtendedConfig("test-pg-ext-project", "test-pg-ext-env", "test-pg-ext-updated", "testpgext", "testdb", "testuser", "256", "512"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_postgres.test", "name", "test-pg-ext-updated"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "memory_reservation", "256"),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "memory_limit", "512"),
				),
			},
		},
	})
}

func testAccPostgresResourceExtendedConfig(projectName, envName, pgName, appNamePrefix, dbName, dbUser, memReserve, memLimit string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for PostgreSQL extended tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_postgres" "test" {
  name               = "%s"
  app_name_prefix     = "%s"
  database_name      = "%s"
  database_user      = "%s"
  database_password  = "test_postgres_password_123"
  environment_id     = dokploy_environment.test.id
  memory_reservation = "%s"
  memory_limit       = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, pgName, appNamePrefix, dbName, dbUser, memReserve, memLimit)
}
