package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["dokploy_domain.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}

					appID := rs.Primary.Attributes["application_id"]
					domainID := rs.Primary.ID

					// Format: application:<app-id>:<domain-id>
					return fmt.Sprintf("application:%s:%s", appID, domainID), nil
				},
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
				Config: testAccDomainResourceWithTraefikMeConfig("test-traefik-project", "test-traefik-env", "test-traefik-app", 3000, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "generate_traefik_me", "true"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "3000"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "https", "true"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "id"),
				),
			},
			// Update testing - change port and https
			{
				Config: testAccDomainResourceWithTraefikMeConfig("test-traefik-project", "test-traefik-env", "test-traefik-app", 8080, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "generate_traefik_me", "true"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "8080"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "https", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_domain.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["dokploy_domain.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}

					appID := rs.Primary.Attributes["application_id"]
					domainID := rs.Primary.ID

					// Format: application:<app-id>:<domain-id>
					return fmt.Sprintf("application:%s:%s", appID, domainID), nil
				},
				ImportStateVerifyIgnore: []string{"generate_traefik_me"}, // This is a creation-time flag, not stored
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
  environment_id = dokploy_environment.test.id
  name           = "%s"
  build_type     = "nixpacks"
  source_type    = "docker"
  docker_image   = "nginx:latest"
}

resource "dokploy_domain" "test" {
  application_id = dokploy_application.test.id
  host           = "%s"
  port           = %d
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, host, port)
}

func testAccDomainResourceWithTraefikMeConfig(projectName, envName, appName string, port int, https bool) string {
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
  environment_id = dokploy_environment.test.id
  name           = "%s"
  build_type     = "nixpacks"
  source_type    = "docker"
  docker_image   = "nginx:latest"
}

resource "dokploy_domain" "test" {
  application_id      = dokploy_application.test.id
  generate_traefik_me = true
  port                = %d
  https               = %t
  path                = "/"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, port, https)
}

// TestAccDomainResourceWithCompose tests domain resource attached to a compose service.
func TestAccDomainResourceWithCompose(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create domain attached to compose
			{
				Config: testAccDomainResourceWithComposeConfig("test-domain-compose-project", "test-domain-compose-env", "test-domain-compose", composeContent, "compose.example.com", 80),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "host", "compose.example.com"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "80"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "compose_id"),
				),
			},
			// Update domain
			{
				Config: testAccDomainResourceWithComposeConfig("test-domain-compose-project", "test-domain-compose-env", "test-domain-compose", composeContent, "updated-compose.example.com", 8080),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "host", "updated-compose.example.com"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "8080"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_domain.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["dokploy_domain.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}

					composeID := rs.Primary.Attributes["compose_id"]
					domainID := rs.Primary.ID

					// Format: compose:<compose-id>:<domain-id>
					return fmt.Sprintf("compose:%s:%s", composeID, domainID), nil
				},
			},
		},
	})
}

func testAccDomainResourceWithComposeConfig(projectName, envName, composeName, composeContent, host string, port int) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for compose domain tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  environment_id       = dokploy_environment.test.id
  name                 = "%s"
  source_type          = "raw"
  compose_file_content = <<EOF
%s
EOF
}

resource "dokploy_domain" "test" {
  compose_id   = dokploy_compose.test.id
  service_name = "web"
  host         = "%s"
  port         = %d
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, composeName, composeContent, host, port)
}

// TestAccDomainResourceWithComposeTraefikMe tests traefik.me domain for compose.
func TestAccDomainResourceWithComposeTraefikMe(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create traefik.me domain attached to compose
			{
				Config: testAccDomainResourceWithComposeTraefikMeConfig("test-compose-traefik-project", "test-compose-traefik-env", "test-compose-traefik", composeContent, 80),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_domain.test", "generate_traefik_me", "true"),
					resource.TestCheckResourceAttr("dokploy_domain.test", "port", "80"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "compose_id"),
					resource.TestCheckResourceAttrSet("dokploy_domain.test", "host"),
				),
			},
		},
	})
}

func testAccDomainResourceWithComposeTraefikMeConfig(projectName, envName, composeName, composeContent string, port int) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for compose traefik.me domain tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  environment_id       = dokploy_environment.test.id
  name                 = "%s"
  source_type          = "raw"
  compose_file_content = <<EOF
%s
EOF
}

resource "dokploy_domain" "test" {
  compose_id          = dokploy_compose.test.id
  service_name        = "web"
  generate_traefik_me = true
  port                = %d
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, composeName, composeContent, port)
}
