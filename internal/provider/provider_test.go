package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file from the root of the project
	_ = godotenv.Load("../../.env")
}

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dokploy": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DOKPLOY_HOST"); v == "" {
		t.Fatal("DOKPLOY_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("DOKPLOY_API_KEY"); v == "" {
		t.Fatal("DOKPLOY_API_KEY must be set for acceptance tests")
	}
}
