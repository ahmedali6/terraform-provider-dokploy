package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDomainResource(t *testing.T) {
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
				Config: testAccDomainResourceConfig("test-domain-project", "test-domain-env", "test-domain-app", "example.com", 3000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "host", "example.com"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "3000"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "application_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDomainResourceConfig("test-domain-project", "test-domain-env", "test-domain-app", "updated.example.com", 8080),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "host", "updated.example.com"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "8080"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_domain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDomainResourceWithTraefikMe(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with Traefik.me
			{
				Config: testAccDomainResourceWithTraefikMeConfig("test-traefik-project", "test-traefik-env", "test-traefik-app"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "generate_traefik_me", "true"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "3000"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "id"),
				),
			},
		},
	})
}

func testAccDomainResourceConfig(projectName, envName, appName, host string, port int) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for domain tests"
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

resource "dokploy_domain" "test" {
  application_id = dokploy_application.test.id
  host           = "%s"
  port           = %d
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, host, port)
}

func testAccDomainResourceWithTraefikMeConfig(projectName, envName, appName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for traefik.me domain tests"
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

resource "dokploy_domain" "test" {
  application_id      = dokploy_application.test.id
  generate_traefik_me = true
  port                = 3000
  https               = true
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName)
}
