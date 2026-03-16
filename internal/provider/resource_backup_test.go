package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBackupResource_Database(t *testing.T) {
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
				Config: testAccBackupResourceConfig_Database("test-backup-project", "test-backup-env", "test-backup-db", "testbkapp", "testbkdb", "testbkuser", "test-backup-dest", "0 2 * * *", true, "db-backup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_backup.test", "schedule", "0 2 * * *"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "prefix", "db-backup"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "database_type", "postgres"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "backup_type", "database"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "keep_latest_count", "30"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test", "destination_id"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test", "database_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccBackupResourceConfig_Database("test-backup-project", "test-backup-env", "test-backup-db", "testbkapp", "testbkdb", "testbkuser", "test-backup-dest", "0 3 * * *", false, "updated-backup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_backup.test", "schedule", "0 3 * * *"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "enabled", "false"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "prefix", "updated-backup"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_backup.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupResource_Compose(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for compose backup
			{
				Config: testAccBackupResourceConfig_Compose("test-compose-backup-project", "test-compose-backup-env", "test-compose-backup-dest", "0 4 * * *", true, "compose-backup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "schedule", "0 4 * * *"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "enabled", "true"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "prefix", "compose-backup"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "backup_type", "compose"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "service_name", "db"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "database_type", "postgres"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "keep_latest_count", "10"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test_compose", "id"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test_compose", "destination_id"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test_compose", "compose_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccBackupResourceConfig_Compose("test-compose-backup-project", "test-compose-backup-env", "test-compose-backup-dest", "0 5 * * *", false, "updated-compose-backup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "schedule", "0 5 * * *"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "enabled", "false"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose", "prefix", "updated-compose-backup"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_backup.test_compose",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupResource_ComposePostgresService(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupResourceConfig_ComposePostgresService(
					"test-compose-pg-backup-project",
					"test-compose-pg-backup-env",
					"test-compose-pg-backup-dest",
					"0 2 * * *",
					true,
					"compose-pg-backup",
					"infisical",
					"postgres",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_backup.test_compose_pg", "backup_type", "compose"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose_pg", "database_type", "postgres"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose_pg", "service_name", "postgres"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose_pg", "database", "infisical"),
				),
			},
			{
				Config: testAccBackupResourceConfig_ComposePostgresService(
					"test-compose-pg-backup-project",
					"test-compose-pg-backup-env",
					"test-compose-pg-backup-dest",
					"0 3 * * *",
					false,
					"compose-pg-backup-updated",
					"infisical",
					"postgres",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_backup.test_compose_pg", "schedule", "0 3 * * *"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose_pg", "enabled", "false"),
					resource.TestCheckResourceAttr("dokploy_backup.test_compose_pg", "prefix", "compose-pg-backup-updated"),
				),
			},
		},
	})
}

func testAccBackupResourceConfig_Database(projectName, envName, dbName, appName, dbDbName, dbUser, destName, schedule string, enabled bool, prefix string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for backup tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_postgres" "test" {
  name              = "%s"
  app_name          = "%s"
  database_name     = "%s"
  database_user     = "%s"
  database_password = "test_password_123"
  environment_id    = dokploy_environment.test.id
}

resource "dokploy_destination" "test" {
  name              = "%s"
  storage_provider  = "s3"
  access_key        = "test-access-key"
  secret_access_key = "test-secret-key"
  bucket            = "test-backups"
  region            = "us-east-1"
  endpoint          = "https://s3.amazonaws.com"
}

resource "dokploy_backup" "test" {
  destination_id    = dokploy_destination.test.id
  database_id       = dokploy_postgres.test.id
  database_type     = "postgres"
  backup_type       = "database"
  schedule          = "%s"
  enabled           = %t
  prefix            = "%s"
  database          = "%s"
  keep_latest_count = 30
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, dbName, appName, dbDbName, dbUser, destName, schedule, enabled, prefix, dbDbName)
}

func testAccBackupResourceConfig_Compose(projectName, envName, destName, schedule string, enabled bool, prefix string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test_compose" {
  name        = "%s"
  description = "Test project for compose backup tests"
}

resource "dokploy_environment" "test_compose" {
  project_id = dokploy_project.test_compose.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  name           = "test-compose-backup"
  environment_id = dokploy_environment.test_compose.id
  source_type    = "raw"
  compose_file_content = <<-EOT
version: '3.8'
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: testpass123
      POSTGRES_DB: testdb
    volumes:
      - db_data:/var/lib/postgresql/data

volumes:
  db_data:
EOT
}

resource "dokploy_destination" "test_compose" {
  name              = "%s"
  storage_provider  = "s3"
  access_key        = "test-access-key"
  secret_access_key = "test-secret-key"
  bucket            = "test-compose-backups"
  region            = "us-east-1"
  endpoint          = "https://s3.amazonaws.com"
}

resource "dokploy_backup" "test_compose" {
  destination_id    = dokploy_destination.test_compose.id
  compose_id        = dokploy_compose.test.id
  backup_type       = "compose"
  service_name      = "db"
  database_type     = "postgres"
  metadata = {
    postgres = {
      database_user = "postgres"
    }
  }
  schedule          = "%s"
  enabled           = %t
  prefix            = "%s"
  database          = "testdb"
  keep_latest_count = 10
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, destName, schedule, enabled, prefix)
}

func testAccBackupResourceConfig_ComposePostgresService(projectName, envName, destName, schedule string, enabled bool, prefix, databaseName, serviceName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test_compose_pg" {
  name        = "%s"
  description = "Test project for compose postgres backup metadata"
}

resource "dokploy_environment" "test_compose_pg" {
  project_id = dokploy_project.test_compose_pg.id
  name       = "%s"
}

resource "dokploy_compose" "test_compose_pg" {
  name           = "test-compose-pg-backup"
  environment_id = dokploy_environment.test_compose_pg.id
  source_type    = "raw"
  compose_file_content = <<-EOT
version: '3.8'
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: testpass123
      POSTGRES_DB: %s
      POSTGRES_USER: %s
EOT
}

resource "dokploy_destination" "test_compose_pg" {
  name              = "%s"
  storage_provider  = "s3"
  access_key        = "test-access-key"
  secret_access_key = "test-secret-key"
  bucket            = "test-compose-pg-backups"
  region            = "us-east-1"
  endpoint          = "https://s3.amazonaws.com"
}

resource "dokploy_backup" "test_compose_pg" {
  destination_id    = dokploy_destination.test_compose_pg.id
  compose_id        = dokploy_compose.test_compose_pg.id
  backup_type       = "compose"
  service_name      = "%s"
  database_type     = "postgres"
  metadata = {
    postgres = {
      database_user = "%s"
    }
  }
  schedule          = "%s"
  enabled           = %t
  prefix            = "%s"
  database          = "%s"
  keep_latest_count = 7
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, databaseName, databaseName, destName, serviceName, databaseName, schedule, enabled, prefix, databaseName)
}
