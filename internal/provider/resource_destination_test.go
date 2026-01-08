package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDestinationResource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	// MinIO credentials for testing (if available)
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	// Use default test values if MinIO not configured
	if minioAccessKey == "" {
		minioAccessKey = "test-access-key"
	}
	if minioSecretKey == "" {
		minioSecretKey = "test-secret-key"
	}
	if minioEndpoint == "" {
		minioEndpoint = "https://s3.amazonaws.com"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDestinationResourceConfig("test-destination", "s3", minioAccessKey, minioSecretKey, "test-backup-bucket", "us-east-1", minioEndpoint),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_destination.test", "name", "test-destination"),
					resource.TestCheckResourceAttr("dokploy_destination.test", "storage_provider", "s3"),
					resource.TestCheckResourceAttr("dokploy_destination.test", "access_key", minioAccessKey),
					resource.TestCheckResourceAttr("dokploy_destination.test", "bucket", "test-backup-bucket"),
					resource.TestCheckResourceAttr("dokploy_destination.test", "region", "us-east-1"),
					resource.TestCheckResourceAttr("dokploy_destination.test", "endpoint", minioEndpoint),
					resource.TestCheckResourceAttrSet("dokploy_destination.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDestinationResourceConfig("test-destination-updated", "s3", minioAccessKey, minioSecretKey, "test-backup-bucket-2", "us-west-2", minioEndpoint),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_destination.test", "name", "test-destination-updated"),
					resource.TestCheckResourceAttr("dokploy_destination.test", "bucket", "test-backup-bucket-2"),
					resource.TestCheckResourceAttr("dokploy_destination.test", "region", "us-west-2"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_destination.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_access_key"}, // Secret not returned by API
			},
		},
	})
}

func testAccDestinationResourceConfig(name, provider, accessKey, secretKey, bucket, region, endpoint string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_destination" "test" {
  name              = "%s"
  storage_provider  = "%s"
  access_key        = "%s"
  secret_access_key = "%s"
  bucket            = "%s"
  region            = "%s"
  endpoint          = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, provider, accessKey, secretKey, bucket, region, endpoint)
}
