package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationResource(t *testing.T) {
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
				Config: testAccApplicationResourceConfig("test-app-project", "test-app-env", "test-app", "nixpacks"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-app"),
					resource.TestCheckResourceAttr("dokploy_application.test", "build_type", "nixpacks"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "project_id"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "environment_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccApplicationResourceConfig("test-app-project", "test-app-env", "test-app-updated", "dockerfile"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-app-updated"),
					resource.TestCheckResourceAttr("dokploy_application.test", "build_type", "dockerfile"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccApplicationResourceWithGit(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with Git
			{
				Config: testAccApplicationResourceWithGitConfig("test-app-git-project", "test-app-git-env", "test-git-app"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-git-app"),
					resource.TestCheckResourceAttr("dokploy_application.test", "custom_git_url", "https://github.com/dokploy/dokploy"),
					resource.TestCheckResourceAttr("dokploy_application.test", "custom_git_branch", "main"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "id"),
				),
			},
		},
	})
}

func testAccApplicationResourceConfig(projectName, envName, appName, buildType string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for application tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  project_id     = dokploy_project.test.id
  environment_id = dokploy_environment.test.id
  name           = "%s"
  build_type     = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, buildType)
}

func testAccApplicationResourceWithGitConfig(projectName, envName, appName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for application git tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  project_id         = dokploy_project.test.id
  environment_id     = dokploy_environment.test.id
  name               = "%s"
  build_type         = "nixpacks"
  custom_git_url     = "https://github.com/dokploy/dokploy"
  custom_git_branch  = "main"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName)
}
