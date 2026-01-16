package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMongoDBResource(t *testing.T) {
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
				Config: testAccMongoDBResourceConfig("test-mongo-project", "test-mongo-env", "test-mongo", "testmongoapp", "testuser"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mongo.test", "name", "test-mongo"),
					resource.TestCheckResourceAttrSet("dokploy_mongo.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_mongo.test", "environment_id"),
					resource.TestCheckResourceAttr("dokploy_mongo.test", "database_user", "testuser"),
					resource.TestCheckResourceAttr("dokploy_mongo.test", "replica_sets", "false"),
					resource.TestCheckResourceAttr("dokploy_mongo.test", "app_name_prefix", "testmongoapp"),
					resource.TestCheckResourceAttrSet("dokploy_mongo.test", "app_name"),
					resource.TestCheckResourceAttrSet("dokploy_mongo.test", "internal_connection"),
				),
			},
			// Update and Read testing
			{
				Config: testAccMongoDBResourceConfigWithDescription("test-mongo-project", "test-mongo-env", "test-mongo-updated", "testmongoapp", "testuser", "Updated MongoDB instance"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mongo.test", "name", "test-mongo-updated"),
					resource.TestCheckResourceAttr("dokploy_mongo.test", "description", "Updated MongoDB instance"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_mongo.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_password", "app_name_prefix", "internal_connection", "external_connection"},
			},
		},
	})
}

func testAccMongoDBResourceConfig(projectName, envName, mongoName, appNamePrefix, dbUser string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for MongoDB tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_mongo" "test" {
  name              = "%s"
  app_name_prefix    = "%s"
  database_user     = "%s"
  database_password = "test_mongo_password_123"
  environment_id    = dokploy_environment.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, mongoName, appNamePrefix, dbUser)
}

func testAccMongoDBResourceConfigWithDescription(projectName, envName, mongoName, appNamePrefix, dbUser, description string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for MongoDB tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_mongo" "test" {
  name              = "%s"
  app_name_prefix    = "%s"
  database_user     = "%s"
  database_password = "test_mongo_password_123"
  environment_id    = dokploy_environment.test.id
  description       = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, mongoName, appNamePrefix, dbUser, description)
}

// TestAccMongoDBResourceWithReplicaSets tests MongoDB with replica sets enabled.
func TestAccMongoDBResourceWithReplicaSets(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with replica sets
			{
				Config: testAccMongoDBResourceWithReplicaSetsConfig("test-mongo-rs-project", "test-mongo-rs-env", "test-mongo-rs", "testmongors", "testuser"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_mongo.test", "name", "test-mongo-rs"),
					resource.TestCheckResourceAttr("dokploy_mongo.test", "replica_sets", "true"),
					resource.TestCheckResourceAttrSet("dokploy_mongo.test", "id"),
				),
			},
		},
	})
}

func testAccMongoDBResourceWithReplicaSetsConfig(projectName, envName, mongoName, appNamePrefix, dbUser string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for MongoDB replica sets tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_mongo" "test" {
  name              = "%s"
  app_name_prefix    = "%s"
  database_user     = "%s"
  database_password = "test_mongo_password_123"
  environment_id    = dokploy_environment.test.id
  replica_sets      = true
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, mongoName, appNamePrefix, dbUser)
}
