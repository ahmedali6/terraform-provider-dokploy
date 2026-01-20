package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBitbucketProviderResource(t *testing.T) {
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
				Config: testAccBitbucketProviderResourceConfig("test-bitbucket-provider", "https://bitbucket.org"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "name", "test-bitbucket-provider"),
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "bitbucket_url", "https://bitbucket.org"),
					resource.TestCheckResourceAttrSet("dokploy_bitbucket_provider.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_bitbucket_provider.test", "git_provider_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccBitbucketProviderResourceConfig("test-bitbucket-provider-updated", "https://bitbucket.org"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "name", "test-bitbucket-provider-updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_bitbucket_provider.test",
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

func TestAccBitbucketProviderResourceWithWorkspace(t *testing.T) {
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
				Config: testAccBitbucketProviderResourceConfigWithWorkspace("test-bitbucket-with-workspace", "https://bitbucket.org", "my-workspace"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "name", "test-bitbucket-with-workspace"),
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "bitbucket_url", "https://bitbucket.org"),
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "workspace", "my-workspace"),
					resource.TestCheckResourceAttrSet("dokploy_bitbucket_provider.test", "id"),
				),
			},
		},
	})
}

func TestAccBitbucketProviderResourceWithCustomURL(t *testing.T) {
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
				Config: testAccBitbucketProviderResourceConfig("test-bitbucket-custom", "https://bitbucket.example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "name", "test-bitbucket-custom"),
					resource.TestCheckResourceAttr("dokploy_bitbucket_provider.test", "bitbucket_url", "https://bitbucket.example.com"),
					resource.TestCheckResourceAttrSet("dokploy_bitbucket_provider.test", "id"),
				),
			},
		},
	})
}

func testAccBitbucketProviderResourceConfig(name, bitbucketURL string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_bitbucket_provider" "test" {
  name           = "%s"
  bitbucket_url  = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, bitbucketURL)
}

func testAccBitbucketProviderResourceConfigWithWorkspace(name, bitbucketURL, workspace string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_bitbucket_provider" "test" {
  name           = "%s"
  bitbucket_url  = "%s"
  workspace      = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, bitbucketURL, workspace)
}
