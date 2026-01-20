package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGitlabProviderResource(t *testing.T) {
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
				Config: testAccGitlabProviderResourceConfig("test-gitlab-provider", "https://gitlab.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "name", "test-gitlab-provider"),
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "gitlab_url", "https://gitlab.com"),
					resource.TestCheckResourceAttrSet("dokploy_gitlab_provider.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_gitlab_provider.test", "git_provider_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccGitlabProviderResourceConfig("test-gitlab-provider-updated", "https://gitlab.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "name", "test-gitlab-provider-updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_gitlab_provider.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"secret",
					"access_token",
					"refresh_token",
				},
			},
		},
	})
}

func TestAccGitlabProviderResourceWithGroup(t *testing.T) {
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
				Config: testAccGitlabProviderResourceConfigWithGroup("test-gitlab-with-group", "https://gitlab.com", "my-group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "name", "test-gitlab-with-group"),
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "gitlab_url", "https://gitlab.com"),
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "group_name", "my-group"),
					resource.TestCheckResourceAttrSet("dokploy_gitlab_provider.test", "id"),
				),
			},
		},
	})
}

func TestAccGitlabProviderResourceWithCustomURL(t *testing.T) {
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
				Config: testAccGitlabProviderResourceConfig("test-gitlab-custom", "https://gitlab.example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "name", "test-gitlab-custom"),
					resource.TestCheckResourceAttr("dokploy_gitlab_provider.test", "gitlab_url", "https://gitlab.example.com"),
					resource.TestCheckResourceAttrSet("dokploy_gitlab_provider.test", "id"),
				),
			},
		},
	})
}

func testAccGitlabProviderResourceConfig(name, gitlabURL string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_gitlab_provider" "test" {
  name       = "%s"
  gitlab_url = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, gitlabURL)
}

func testAccGitlabProviderResourceConfigWithGroup(name, gitlabURL, groupName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_gitlab_provider" "test" {
  name       = "%s"
  gitlab_url = "%s"
  group_name = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, gitlabURL, groupName)
}
