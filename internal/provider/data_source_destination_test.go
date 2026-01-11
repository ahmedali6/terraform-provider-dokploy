package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDestinationDataSource(t *testing.T) {
	// Skip in CI - requires running Dokploy instance with a destination
	t.Skip("Skipping acceptance test - requires running Dokploy instance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "id"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "name"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "storage_provider"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "access_key"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "bucket"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "region"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "organization_id"),
					resource.TestCheckResourceAttrSet("data.dokploy_destination.test", "created_at"),
				),
			},
		},
	})
}

func TestAccDestinationsDataSource(t *testing.T) {
	// Skip in CI - requires running Dokploy instance
	t.Skip("Skipping acceptance test - requires running Dokploy instance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_destinations.test", "destinations.#"),
				),
			},
		},
	})
}

func testAccDestinationDataSourceConfig() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_destination" "test" {
  name              = "test-ds-destination"
  storage_provider  = "aws-s3"
  access_key        = "test-access-key"
  secret_access_key = "test-secret-key"
  bucket            = "test-bucket"
  region            = "us-east-1"
  endpoint          = "https://s3.amazonaws.com"
}

data "dokploy_destination" "test" {
  id = dokploy_destination.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"))
}

func testAccDestinationsDataSourceConfig() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_destination" "test" {
  name              = "test-ds-destinations"
  storage_provider  = "aws-s3"
  access_key        = "test-access-key"
  secret_access_key = "test-secret-key"
  bucket            = "test-bucket"
  region            = "us-east-1"
  endpoint          = "https://s3.amazonaws.com"
}

data "dokploy_destinations" "test" {
  depends_on = [dokploy_destination.test]
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"))
}
