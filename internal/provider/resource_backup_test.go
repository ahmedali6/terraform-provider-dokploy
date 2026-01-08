package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBackupResource(t *testing.T) {
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
				Config: testAccBackupResourceConfig("test-backup-project", "test-backup-env", "test-backup-db", "test-backup-dest", "0 2 * * *", true, "db-backup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_backup.test", "schedule", "0 2 * * *"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "prefix", "db-backup"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "database_type", "postgres"),
					resource.TestCheckResourceAttr("dokploy_backup.test", "keep_latest_count", "30"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test", "destination_id"),
					resource.TestCheckResourceAttrSet("dokploy_backup.test", "database_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccBackupResourceConfig("test-backup-project", "test-backup-env", "test-backup-db", "test-backup-dest", "0 3 * * *", false, "updated-backup"),
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

func testAccBackupResourceConfig(projectName, envName, dbName, destName, schedule string, enabled bool, prefix string) string {
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

resource "dokploy_database" "test" {
  project_id     = dokploy_project.test.id
  environment_id = dokploy_environment.test.id
  name           = "%s"
  type           = "postgres"
  password       = "test_password_123"
  version        = "16"
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
  database_id       = dokploy_database.test.id
  database_type     = "postgres"
  schedule          = "%s"
  enabled           = %t
  prefix            = "%s"
  database          = "postgres"
  keep_latest_count = 30
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, dbName, destName, schedule, enabled, prefix)
}
