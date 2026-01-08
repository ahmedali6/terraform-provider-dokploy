package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedirectResource(t *testing.T) {
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
				Config: testAccRedirectResourceConfig("test-redirect-project", "test-redirect-env", "test-redirect-app", "/old-path", "/new-path", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_redirect.test", "regex", "/old-path"),
					resource.TestCheckResourceAttr("dokploy_redirect.test", "replacement", "/new-path"),
					resource.TestCheckResourceAttr("dokploy_redirect.test", "permanent", "false"),
					resource.TestCheckResourceAttrSet("dokploy_redirect.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_redirect.test", "application_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccRedirectResourceConfig("test-redirect-project", "test-redirect-env", "test-redirect-app", "/old-updated", "/new-updated", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_redirect.test", "regex", "/old-updated"),
					resource.TestCheckResourceAttr("dokploy_redirect.test", "replacement", "/new-updated"),
					resource.TestCheckResourceAttr("dokploy_redirect.test", "permanent", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_redirect.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRedirectResourceConfig(projectName, envName, appName, regex, replacement string, permanent bool) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for redirect tests"
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

resource "dokploy_redirect" "test" {
  application_id = dokploy_application.test.id
  regex          = "%s"
  replacement    = "%s"
  permanent      = %t
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, regex, replacement, permanent)
}
