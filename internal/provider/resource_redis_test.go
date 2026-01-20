package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedisResource(t *testing.T) {
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
				Config: testAccRedisResourceConfig("test-redis-project", "test-redis-env", "test-redis", "testredisapp"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_redis.test", "name", "test-redis"),
					resource.TestCheckResourceAttrSet("dokploy_redis.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_redis.test", "environment_id"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "app_name_prefix", "testredisapp"),
					resource.TestCheckResourceAttrSet("dokploy_redis.test", "app_name"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "database_user", "default"),
					resource.TestCheckResourceAttrSet("dokploy_redis.test", "internal_connection"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "internal_port", "6379"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "external_connection", ""),
				),
			},
			// Update and Read testing
			{
				Config: testAccRedisResourceConfigWithDescription("test-redis-project", "test-redis-env", "test-redis-updated", "testredisapp", "Updated Redis instance"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_redis.test", "name", "test-redis-updated"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "description", "Updated Redis instance"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_redis.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_password", "app_name_prefix", "internal_connection", "external_connection"},
			},
		},
	})
}

func testAccRedisResourceConfig(projectName, envName, redisName, appNamePrefix string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for Redis tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_redis" "test" {
  name              = "%s"
  app_name_prefix   = "%s"
  database_password = "test_redis_password_123"
  environment_id    = dokploy_environment.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, redisName, appNamePrefix)
}

func testAccRedisResourceConfigWithDescription(projectName, envName, redisName, appNamePrefix string, description string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for Redis tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_redis" "test" {
  name              = "%s"
  app_name_prefix   = "%s"
  database_password = "test_redis_password_123"
  environment_id    = dokploy_environment.test.id
  description       = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, redisName, appNamePrefix, description)
}

// TestAccRedisResourceExtended tests Redis with extended settings that trigger needsUpdate path.
func TestAccRedisResourceExtended(t *testing.T) {
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
				Config: testAccRedisResourceExtendedConfig("test-redis-ext-project", "test-redis-ext-env", "test-redis-ext", "testredisext", "128", "256", "REDIS_ENV=test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_redis.test", "name", "test-redis-ext"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "memory_reservation", "128"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "memory_limit", "256"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "env", "REDIS_ENV=test"),
					resource.TestCheckResourceAttrSet("dokploy_redis.test", "id"),
				),
			},
			// Update extended settings
			{
				Config: testAccRedisResourceExtendedConfig("test-redis-ext-project", "test-redis-ext-env", "test-redis-ext-updated", "testredisext", "256", "512", "REDIS_ENV=production"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_redis.test", "name", "test-redis-ext-updated"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "memory_reservation", "256"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "memory_limit", "512"),
					resource.TestCheckResourceAttr("dokploy_redis.test", "env", "REDIS_ENV=production"),
				),
			},
		},
	})
}

func testAccRedisResourceExtendedConfig(projectName, envName, redisName, appNamePrefix, memReserve, memLimit, env string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for Redis extended tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_redis" "test" {
  name               = "%s"
  app_name_prefix    = "%s"
  database_password  = "test_redis_password_123"
  environment_id     = dokploy_environment.test.id
  memory_reservation = "%s"
  memory_limit       = "%s"
  env                = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, redisName, appNamePrefix, memReserve, memLimit, env)
}
