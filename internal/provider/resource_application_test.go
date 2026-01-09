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
				Config: testAccApplicationResourceConfig("test-app-project", "test-app-env", "test-app", "nginx:latest", "Test App", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-app"),
					resource.TestCheckResourceAttr("dokploy_application.test", "source_type", "docker"),
					resource.TestCheckResourceAttr("dokploy_application.test", "docker_image", "nginx:latest"),
					resource.TestCheckResourceAttr("dokploy_application.test", "title", "Test App"),
					resource.TestCheckResourceAttr("dokploy_application.test", "replicas", "1"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "environment_id"),
				),
			},
			// Update and Read testing - change name, docker_image, title, and replicas
			{
				Config: testAccApplicationResourceConfig("test-app-project", "test-app-env", "test-app-updated", "nginx:alpine", "Updated App", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-app-updated"),
					resource.TestCheckResourceAttr("dokploy_application.test", "source_type", "docker"),
					resource.TestCheckResourceAttr("dokploy_application.test", "docker_image", "nginx:alpine"),
					resource.TestCheckResourceAttr("dokploy_application.test", "title", "Updated App"),
					resource.TestCheckResourceAttr("dokploy_application.test", "replicas", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_application.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"branch", "owner", "repository", "github_id",
					"dockerfile_path", "docker_context_path", "docker_build_stage",
					"deploy_on_create", // Not returned by API
					"title",            // Not returned by API on import
				},
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
				Config: testAccApplicationResourceWithGitConfig("test-app-git-project", "test-app-git-env", "test-git-app", "main"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-git-app"),
					resource.TestCheckResourceAttr("dokploy_application.test", "custom_git_url", "https://github.com/dokploy/dokploy"),
					resource.TestCheckResourceAttr("dokploy_application.test", "custom_git_branch", "main"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "id"),
				),
			},
			// Update testing - change name and branch
			{
				Config: testAccApplicationResourceWithGitConfig("test-app-git-project", "test-app-git-env", "test-git-app-updated", "canary"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-git-app-updated"),
					resource.TestCheckResourceAttr("dokploy_application.test", "custom_git_branch", "canary"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_application.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"branch", "owner", "repository", "github_id",
					"dockerfile_path", "docker_context_path", "docker_build_stage",
					"deploy_on_create",
				},
			},
		},
	})
}

func testAccApplicationResourceConfig(projectName, envName, appName, dockerImage, title string, replicas int) string {
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
  environment_id = dokploy_environment.test.id
  name           = "%s"
  source_type    = "docker"
  docker_image   = "%s"
  title          = "%s"
  replicas       = %d
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, dockerImage, title, replicas)
}

func testAccApplicationResourceWithGitConfig(projectName, envName, appName, branch string) string {
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
  environment_id     = dokploy_environment.test.id
  name               = "%s"
  source_type        = "git"
  build_type         = "nixpacks"
  custom_git_url     = "https://github.com/dokploy/dokploy"
  custom_git_branch  = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, branch)
}

// TestAccApplicationResourceInferDockerType tests source type inference for docker.
func TestAccApplicationResourceInferDockerType(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without explicit source_type - should infer "docker" from docker_image
			{
				Config: testAccApplicationResourceInferDockerConfig("test-infer-docker-project", "test-infer-docker-env", "test-infer-docker-app"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-infer-docker-app"),
					resource.TestCheckResourceAttr("dokploy_application.test", "source_type", "docker"),
					resource.TestCheckResourceAttr("dokploy_application.test", "docker_image", "nginx:latest"),
				),
			},
		},
	})
}

func testAccApplicationResourceInferDockerConfig(projectName, envName, appName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for infer docker type tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  # source_type omitted - should be inferred as "docker" because docker_image is set
  docker_image   = "nginx:latest"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName)
}

// TestAccApplicationResourceInferGitType tests source type inference for git.
func TestAccApplicationResourceInferGitType(t *testing.T) {
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
				Config: testAccApplicationResourceInferGitConfig("test-infer-git-project", "test-infer-git-env", "test-infer-git-app"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-infer-git-app"),
					resource.TestCheckResourceAttr("dokploy_application.test", "source_type", "git"),
					resource.TestCheckResourceAttr("dokploy_application.test", "custom_git_url", "https://github.com/dokploy/dokploy"),
				),
			},
		},
	})
}

func testAccApplicationResourceInferGitConfig(projectName, envName, appName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for infer git type tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  environment_id     = dokploy_environment.test.id
  name               = "%s"
  # source_type omitted - should be inferred as "git" because custom_git_url is set
  build_type         = "nixpacks"
  custom_git_url     = "https://github.com/dokploy/dokploy"
  custom_git_branch  = "main"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName)
}

// TestAccApplicationResourceExtendedSettings tests more optional fields.
func TestAccApplicationResourceExtendedSettings(t *testing.T) {
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
				Config: testAccApplicationResourceExtendedConfig("test-extended-project", "test-extended-env", "test-extended-app", "Initial description", 1, 256, 128),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-extended-app"),
					resource.TestCheckResourceAttr("dokploy_application.test", "description", "Initial description"),
					resource.TestCheckResourceAttr("dokploy_application.test", "replicas", "1"),
					resource.TestCheckResourceAttr("dokploy_application.test", "memory_limit", "256"),
					resource.TestCheckResourceAttr("dokploy_application.test", "memory_reservation", "128"),
					resource.TestCheckResourceAttr("dokploy_application.test", "env", "APP_ENV=test\nDEBUG=true"),
				),
			},
			// Update extended settings
			{
				Config: testAccApplicationResourceExtendedConfig("test-extended-project", "test-extended-env", "test-extended-app", "Updated description", 2, 512, 256),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("dokploy_application.test", "replicas", "2"),
					resource.TestCheckResourceAttr("dokploy_application.test", "memory_limit", "512"),
					resource.TestCheckResourceAttr("dokploy_application.test", "memory_reservation", "256"),
				),
			},
		},
	})
}

func testAccApplicationResourceExtendedConfig(projectName, envName, appName, description string, replicas, memLimit, memReserve int) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for extended settings tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  environment_id     = dokploy_environment.test.id
  name               = "%s"
  description        = "%s"
  source_type        = "docker"
  docker_image       = "nginx:latest"
  replicas           = %d
  memory_limit       = %d
  memory_reservation = %d
  env                = "APP_ENV=test\nDEBUG=true"
  auto_deploy        = false
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, description, replicas, memLimit, memReserve)
}
