package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServersDataSource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing - all servers
			{
				Config: testAccServersDataSourceConfig(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Just check that servers attribute exists (may be empty list)
					resource.TestCheckResourceAttrSet("data.dokploy_servers.test", "servers.#"),
				),
			},
		},
	})
}

func testAccServersDataSourceConfig(serverType string) string {
	serverTypeConfig := ""
	if serverType != "" {
		serverTypeConfig = fmt.Sprintf(`server_type = "%s"`, serverType)
	}

	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

data "dokploy_servers" "test" {
  %s
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), serverTypeConfig)
}
