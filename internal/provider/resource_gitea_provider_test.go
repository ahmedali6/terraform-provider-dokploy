package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGiteaProviderResource(t *testing.T) {
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
				Config: testAccGiteaProviderResourceConfig("test-gitea-provider", "https://gitea.example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "name", "test-gitea-provider"),
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "gitea_url", "https://gitea.example.com"),
					resource.TestCheckResourceAttrSet("dokploy_gitea_provider.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_gitea_provider.test", "git_provider_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccGiteaProviderResourceConfig("test-gitea-provider-updated", "https://gitea.example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "name", "test-gitea-provider-updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_gitea_provider.test",
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

func TestAccGiteaProviderResourceWithGroup(t *testing.T) {
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
				Config: testAccGiteaProviderResourceConfigWithGroup("test-gitea-with-group", "https://gitea.example.com", "my-group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "name", "test-gitea-with-group"),
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "gitea_url", "https://gitea.example.com"),
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "group_name", "my-group"),
					resource.TestCheckResourceAttrSet("dokploy_gitea_provider.test", "id"),
				),
			},
		},
	})
}

func TestAccGiteaProviderResourceWithCustomURL(t *testing.T) {
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
				Config: testAccGiteaProviderResourceConfig("test-gitea-custom", "https://gitea.custom.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "name", "test-gitea-custom"),
					resource.TestCheckResourceAttr("dokploy_gitea_provider.test", "gitea_url", "https://gitea.custom.com"),
					resource.TestCheckResourceAttrSet("dokploy_gitea_provider.test", "id"),
				),
			},
		},
	})
}

func testAccGiteaProviderResourceConfig(name, giteaURL string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_gitea_provider" "test" {
  name      = "%s"
  gitea_url = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, giteaURL)
}

func testAccGiteaProviderResourceConfigWithGroup(name, giteaURL, groupName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_gitea_provider" "test" {
  name       = "%s"
  gitea_url  = "%s"
  group_name = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, giteaURL, groupName)
}
