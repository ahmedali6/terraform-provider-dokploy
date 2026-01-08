package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccServerResource tests the server resource.
// Note: This test requires a valid SSH key and a real server to connect to.
// Set TEST_SERVER_IP and TEST_SSH_KEY_ID environment variables to run this test.
func TestAccServerResource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")
	serverIP := os.Getenv("TEST_SERVER_IP")
	sshKeyID := os.Getenv("TEST_SSH_KEY_ID")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	if serverIP == "" || sshKeyID == "" {
		t.Skip("TEST_SERVER_IP and TEST_SSH_KEY_ID must be set for server acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServerResourceConfig("test-server", serverIP, 22, "root", sshKeyID, "deploy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_server.test", "name", "test-server"),
					resource.TestCheckResourceAttr("dokploy_server.test", "ip_address", serverIP),
					resource.TestCheckResourceAttr("dokploy_server.test", "port", "22"),
					resource.TestCheckResourceAttr("dokploy_server.test", "username", "root"),
					resource.TestCheckResourceAttr("dokploy_server.test", "ssh_key_id", sshKeyID),
					resource.TestCheckResourceAttr("dokploy_server.test", "server_type", "deploy"),
					resource.TestCheckResourceAttrSet("dokploy_server.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccServerResourceConfigWithDescription("test-server-updated", "Updated test server", serverIP, 22, "root", sshKeyID, "deploy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_server.test", "name", "test-server-updated"),
					resource.TestCheckResourceAttr("dokploy_server.test", "description", "Updated test server"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dokploy_server.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccServerResourceConfig(name, ipAddress string, port int, username, sshKeyID, serverType string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_server" "test" {
  name        = "%s"
  ip_address  = "%s"
  port        = %d
  username    = "%s"
  ssh_key_id  = "%s"
  server_type = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, ipAddress, port, username, sshKeyID, serverType)
}

func testAccServerResourceConfigWithDescription(name, description, ipAddress string, port int, username, sshKeyID, serverType string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_server" "test" {
  name        = "%s"
  description = "%s"
  ip_address  = "%s"
  port        = %d
  username    = "%s"
  ssh_key_id  = "%s"
  server_type = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, description, ipAddress, port, username, sshKeyID, serverType)
}
