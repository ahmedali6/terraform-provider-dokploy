package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentVariablesResource(t *testing.T) {
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
				Config: testAccEnvironmentVariablesResourceConfig("test-env-vars-project", "test-env-vars-env", "test-env-vars-app", "ENV1=value1\nENV2=value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment_variables.test", "env", "ENV1=value1\nENV2=value2"),
					resource.TestCheckResourceAttrSet("dokploy_environment_variables.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_environment_variables.test", "application_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccEnvironmentVariablesResourceConfig("test-env-vars-project", "test-env-vars-env", "test-env-vars-app", "ENV1=updated_value1\nENV3=value3"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment_variables.test", "env", "ENV1=updated_value1\nENV3=value3"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_environment_variables.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEnvironmentVariablesResourceWithBuildArgs(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with build args
			{
				Config: testAccEnvironmentVariablesResourceWithBuildArgsConfig("test-build-args-project", "test-build-args-env", "test-build-args-app"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment_variables.test", "env", "ENV1=value1"),
					resource.TestCheckResourceAttr("dokploy_environment_variables.test", "build_args", "ARG1=buildvalue1"),
					resource.TestCheckResourceAttrSet("dokploy_environment_variables.test", "id"),
				),
			},
		},
	})
}

func testAccEnvironmentVariablesResourceConfig(projectName, envName, appName, envVars string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for environment variables tests"
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

resource "dokploy_environment_variables" "test" {
  application_id = dokploy_application.test.id
  env            = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, envVars)
}

func testAccEnvironmentVariablesResourceWithBuildArgsConfig(projectName, envName, appName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for build args tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  project_id     = dokploy_project.test.id
  environment_id = dokploy_environment.test.id
  name           = "%s"
  build_type     = "dockerfile"
}

resource "dokploy_environment_variables" "test" {
  application_id = dokploy_application.test.id
  env            = "ENV1=value1"
  build_args     = "ARG1=buildvalue1"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName)
}
