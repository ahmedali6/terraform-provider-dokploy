package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccComposeResource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	composeContentV1 := `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"`

	composeContentV2 := `version: '3.8'
services:
  web:
    image: nginx:alpine
    ports:
      - "8081:80"
  redis:
    image: redis:latest`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccComposeResourceConfig("test-compose-project", "test-env", "test-compose", composeContentV1, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-compose"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "environment_id"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "deploy_on_create", "false"),
				),
			},
			// Update and Read testing - change name and compose_file_content
			{
				Config: testAccComposeResourceConfig("test-compose-project", "test-env", "test-compose-updated", composeContentV2, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-compose-updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_compose.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deploy_on_create", "branch", "trigger_type"}, // deploy_on_create is write-only; branch/trigger_type have API defaults that don't apply to raw source type in this test
			},
		},
	})
}

func testAccComposeResourceConfig(projectName, envName, composeName, composeContent string, deployOnCreate bool) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for compose tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  source_type    = "raw"
  compose_file_content = <<EOF
%s
EOF
  deploy_on_create = %t
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, composeName, composeContent, deployOnCreate)
}

// TestAccComposeResourceInferRawType tests source type inference for raw compose
func TestAccComposeResourceInferRawType(t *testing.T) {
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
			// Create without explicit source_type - should infer "raw" from compose_file_content
			{
				Config: testAccComposeResourceInferRawConfig("test-compose-infer-raw", "test-env-infer-raw", "test-infer-raw", composeContent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-infer-raw"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "source_type", "raw"),
				),
			},
		},
	})
}

func testAccComposeResourceInferRawConfig(projectName, envName, composeName, composeContent string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for compose infer raw tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  # source_type omitted - should be inferred as "raw" because compose_file_content is set
  compose_file_content = <<EOF
%s
EOF
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, composeName, composeContent)
}

// TestAccComposeResourceInferGitType tests source type inference for git compose
func TestAccComposeResourceInferGitType(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without explicit source_type - should infer "git" from custom_git_url
			{
				Config: testAccComposeResourceInferGitConfig("test-compose-infer-git", "test-env-infer-git", "test-infer-git"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-infer-git"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "source_type", "git"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "custom_git_url", "https://github.com/dokploy/dokploy"),
				),
			},
		},
	})
}

func testAccComposeResourceInferGitConfig(projectName, envName, composeName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for compose infer git tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  environment_id    = dokploy_environment.test.id
  name              = "%s"
  # source_type omitted - should be inferred as "git" because custom_git_url is set
  custom_git_url    = "https://github.com/dokploy/dokploy"
  custom_git_branch = "main"
  compose_path      = "./docker-compose.yml"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, composeName)
}

// TestAccComposeResourceExtended tests compose with extended settings
func TestAccComposeResourceExtended(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
    environment:
      - APP_ENV=production`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with extended settings
			{
				Config: testAccComposeResourceExtendedConfig("test-compose-ext-project", "test-compose-ext-env", "test-compose-ext", composeContent, "Test compose description", "ENV_VAR=value1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-compose-ext"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "description", "Test compose description"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "env", "ENV_VAR=value1"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "id"),
				),
			},
			// Update extended settings
			{
				Config: testAccComposeResourceExtendedConfig("test-compose-ext-project", "test-compose-ext-env", "test-compose-ext-updated", composeContent, "Updated compose description", "ENV_VAR=value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", "test-compose-ext-updated"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "description", "Updated compose description"),
					resource.TestCheckResourceAttr("dokploy_compose.test", "env", "ENV_VAR=value2"),
				),
			},
		},
	})
}

func testAccComposeResourceExtendedConfig(projectName, envName, composeName, composeContent, description, env string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for compose extended tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_compose" "test" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  description    = "%s"
  source_type    = "raw"
  compose_file_content = <<EOF
%s
EOF
  env = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, composeName, description, composeContent, env)
}
