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

// TestAccApplicationResourceTraefikConfig tests the traefik_config attribute.
func TestAccApplicationResourceTraefikConfig(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with traefik_config
			{
				Config: testAccApplicationResourceTraefikConfig("test-traefik-project", "test-traefik-env", "test-traefik-app", "# Custom Traefik config\nhttp:\n  routers:\n    test:\n      rule: Host(`test.example.com`)"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-traefik-app"),
					resource.TestCheckResourceAttrSet("dokploy_application.test", "traefik_config"),
				),
			},
			// Update traefik_config
			{
				Config: testAccApplicationResourceTraefikConfig("test-traefik-project", "test-traefik-env", "test-traefik-app", "# Updated config\nhttp:\n  routers:\n    updated:\n      rule: Host(`updated.example.com`)"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dokploy_application.test", "traefik_config"),
				),
			},
		},
	})
}

func testAccApplicationResourceTraefikConfig(projectName, envName, appName, traefikConfig string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for traefik config tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  source_type    = "docker"
  docker_image   = "nginx:latest"
  traefik_config = <<-EOT
%s
EOT
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName, traefikConfig)
}

// TestAccApplicationResourceMoveEnvironment tests moving an application between environments.
func TestAccApplicationResourceMoveEnvironment(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create in first environment
			{
				Config: testAccApplicationResourceMoveEnvConfig("test-move-project", "env-1", "env-2", "test-move-app", "env-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-move-app"),
					resource.TestCheckResourceAttrPair("dokploy_application.test", "environment_id", "dokploy_environment.env1", "id"),
				),
			},
			// Move to second environment
			{
				Config: testAccApplicationResourceMoveEnvConfig("test-move-project", "env-1", "env-2", "test-move-app", "env-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_application.test", "name", "test-move-app"),
					resource.TestCheckResourceAttrPair("dokploy_application.test", "environment_id", "dokploy_environment.env2", "id"),
				),
			},
		},
	})
}

func testAccApplicationResourceMoveEnvConfig(projectName, env1Name, env2Name, appName, targetEnv string) string {
	envRef := "dokploy_environment.env1.id"
	if targetEnv == "env-2" {
		envRef = "dokploy_environment.env2.id"
	}
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for move environment tests"
}

resource "dokploy_environment" "env1" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_environment" "env2" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  environment_id = %s
  name           = "%s"
  source_type    = "docker"
  docker_image   = "nginx:latest"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, env1Name, env2Name, envRef, appName)
}

// TestAccApplicationDataSource tests the single application data source.
func TestAccApplicationDataSource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig("test-ds-app-project", "test-ds-app-env", "test-ds-app"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.dokploy_application.test", "id", "dokploy_application.test", "id"),
					resource.TestCheckResourceAttr("data.dokploy_application.test", "name", "test-ds-app"),
					resource.TestCheckResourceAttr("data.dokploy_application.test", "source_type", "docker"),
					resource.TestCheckResourceAttr("data.dokploy_application.test", "docker_image", "nginx:latest"),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig(projectName, envName, appName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for application data source tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  source_type    = "docker"
  docker_image   = "nginx:latest"
}

data "dokploy_application" "test" {
  id = dokploy_application.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, appName)
}

// TestAccApplicationsDataSource tests the applications list data source.
func TestAccApplicationsDataSource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationsDataSourceConfig("test-ds-apps-project", "test-ds-apps-env", "test-ds-app-1", "test-ds-app-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_applications.all", "applications.#"),
					resource.TestCheckResourceAttrSet("data.dokploy_applications.by_env", "applications.#"),
				),
			},
		},
	})
}

func testAccApplicationsDataSourceConfig(projectName, envName, app1Name, app2Name string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "test" {
  name        = "%s"
  description = "Test project for applications data source tests"
}

resource "dokploy_environment" "test" {
  project_id = dokploy_project.test.id
  name       = "%s"
}

resource "dokploy_application" "test1" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  source_type    = "docker"
  docker_image   = "nginx:latest"
}

resource "dokploy_application" "test2" {
  environment_id = dokploy_environment.test.id
  name           = "%s"
  source_type    = "docker"
  docker_image   = "nginx:alpine"
}

data "dokploy_applications" "all" {
  depends_on = [dokploy_application.test1, dokploy_application.test2]
}

data "dokploy_applications" "by_env" {
  environment_id = dokploy_environment.test.id
  depends_on     = [dokploy_application.test1, dokploy_application.test2]
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName, app1Name, app2Name)
}
