package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Test certificate and private key (self-signed, for testing only).
const testCertificateData = `-----BEGIN CERTIFICATE-----
MIICpDCCAYwCCQDU+pQ4P4V5xjANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls
b2NhbGhvc3QwHhcNMjQwMTAxMDAwMDAwWhcNMjUwMTAxMDAwMDAwWjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC7
o5e7VvdeF0LZpJPKs9RZpJGsVjLuVGCAxvM3H1GZqYjITL8bVgI4J8i3pHZxYCbV
5TJL0fPYrVKqR9MLqF6vPJqJZ8OLv9g7n7hqfZxyPKLZ6sYHjJFvJqIvxcM2FPkz
kn6VrNmIj+aIavGMDLw0KSM9ZI8m8OHvw7k7HoBiJGZVxF8yTTPVJN0EWGxJvRPA
Q3ckCz3vjwJJFc6Jvk4roUdpkVT7PKLK9V+WDvVsD8IvJD8lveRP7X8H5x7Y9T8L
e7gN0stFLSPY5JR/xFNxkJmTqIqhFyKvbTwxPRmIgpJJZwn1zNaKm6L8c9XcZJzE
V2FJJqFVW5KFZ5p5HzJ3AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKS2hKe0Zdjt
K8RN7OLRveE1XfDYtMQIEhs5b/sDGowpRy9MLMbOVtbPjLo5P8sMy1WQ+xST7K4I
q0nGLuoZwFJfmQCPFweIvMcPfRxS/XkNk2GTGT/JVF9pu8azfjPLIl+YKj0RYjYx
8P8/228ZWspUsXuExleqoQ5N+gB7R5HamsLjPV/u8PCJK8y1TkLmkREIlrV2g/EN
rLzH8D+e8SwjPvy0vxrVbNF5kjZuj4zNedzTNSlPzCJ7fJdniZhBCnpmVyNX2sGf
z8xCrB0Xr5ay5qQ/47dFP+JELsAsHyRbUoJoJrsv9aK97hRAP1JdVDFLzbJ0EQRa
yCJiQs8qx8s=
-----END CERTIFICATE-----`

const testPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC7o5e7VvdeF0LZ
pJPKs9RZpJGsVjLuVGCAxvM3H1GZqYjITL8bVgI4J8i3pHZxYCbV5TJL0fPYrVKq
R9MLqF6vPJqJZ8OLv9g7n7hqfZxyPKLZ6sYHjJFvJqIvxcM2FPkzkn6VrNmIj+aI
avGMDLw0KSM9ZI8m8OHvw7k7HoBiJGZVxF8yTTPVJN0EWGxJvRPAQ3ckCz3vjwJJ
Fc6Jvk4roUdpkVT7PKLK9V+WDvVsD8IvJD8lveRP7X8H5x7Y9T8Le7gN0stFLSPY
5JR/xFNxkJmTqIqhFyKvbTwxPRmIgpJJZwn1zNaKm6L8c9XcZJzEV2FJJqFVW5KF
Z5p5HzJ3AgMBAAECggEABLnGV2BUhHVX0o8q2aZrO8p6dCb3ESUqfMV/cMU7Hngs
X3NqtXF2mULPEO3Zn1BhSaFp7R7aQE2bP4xNXjPyGB6OGFrWDhvr8l9VLdU+2hVm
4UTKPq7s3hLhKWn0Gn2Z1vfAt0UNe8SwchXkWqxRDLkNpbdlBcJHy9O1rOIF9SIi
wSLDeWs0tUbhX7SFnP6qrXPPLlTe6qfNDPqJnMU0ReTL5qAKrfxB0vWnbkSMNOJL
xRKci7zXopCglqL8pSlSqKcbmN3qKLPAE0n/3kV5aHpryqqv0wSlWU5E7Gu/L8p7
c8k0H8Y5HLFqypX/8xv+qVnKAkELDRsEXejy4RqWgQKBgQDwfOb6UE0gET7GHXXB
k8sVnJf3s0P2sqrdgBG7YmLJqCLxf6C7LiYzELaROMbQFnJcfMWgVKR7h6SqV+dO
l3NgMF2N4ZkIuB1b9TRp1ld9IfHrKH8ZaKRkEfR8WFPV8Bv8FLVuGqPy/J3IEFM6
zE8NMOsJZXa5N8djF5ZhKDZHtwKBgQDIMdEiRw0yMexYk4pNb5LW1cR8CqLJD/RV
xIvmJPLc4xo0eHXX3SGcvWtEE5H0NPE8FLWI2vLLXzoqbTKT4xq6h5vM1igNcBs0
vj4w7cHfpW4aYsFrEs0jDNt7xmunD8SOJd4oXT4qwP5pJq1yF3jQ1ub3xZ9HMpHa
q1WQxC+XgQKBgQDqLuE3b3yxO9E0Oqz7fXxuIwDSTlF9F7C8kUvnxf8gP8yVz1h7
ckAZjWYXLy8nXSoMKz3P+TuP4bCrJEWZLqMKaWnzmpLJFHHkKLDsC3NQYAoifgvy
G0LNkzl7jilTYWKcaE+E7yPGSyh1aPBH9zMFBVmMpvPSeyJ1fpzTHO1zqQKBgAuF
C3IFawl1OyLp7p7DuGPp0Y/wPVi/Kwxw0c0aEPKSl6CKSLWXK7Ywb1LZNHqC8TSQ
hnZNH7bVJR9aZ8j7hsBEYpM0sUmFx8xCH1eTMHYKzCvPFH0Ez9I+QWmzD9BDFMF1
tX9tKuDKoN7pAuK7ZQKBB8BEW6JNE5R1J7GVJkgBAoGBANdR/aP2YPH5nQzVNM9Y
vVPL3xHjZMb7LNfOP2L0/iPpeTMHZhMuE/7h2kEQ/ixTB4MZGMP5AZqQDLCZCVIr
1SNghfkxNMIwY8q7JzF2f6hFKLCshH7Eb8m2LOYbCL8fPB5Z6bHYqmx6cjk7xF0j
c5xKbwM2ZDy8LMEPEqZVmXfc
-----END PRIVATE KEY-----`

func TestAccCertificateResource(t *testing.T) {
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
				Config: testAccCertificateResourceConfig("test-certificate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_certificate.test", "name", "test-certificate"),
					resource.TestCheckResourceAttrSet("dokploy_certificate.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_certificate.test", "certificate_path"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_certificate.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_data", "private_key"}, // Sensitive data not verified
			},
		},
	})
}

func TestAccCertificateResourceWithPath(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with custom certificate path
			{
				Config: testAccCertificateResourceConfigWithPath("test-cert-with-path", "custom-cert-path"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_certificate.test", "name", "test-cert-with-path"),
					resource.TestCheckResourceAttr("dokploy_certificate.test", "certificate_path", "custom-cert-path"),
				),
			},
		},
	})
}

func TestAccCertificateDataSource(t *testing.T) {
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
				Config: testAccCertificateDataSourceConfig("test-cert-for-datasource"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.dokploy_certificate.test", "name", "test-cert-for-datasource"),
					resource.TestCheckResourceAttrSet("data.dokploy_certificate.test", "certificate_path"),
					resource.TestCheckResourceAttrSet("data.dokploy_certificate.test", "organization_id"),
				),
			},
		},
	})
}

func TestAccCertificatesDataSource(t *testing.T) {
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
				Config: testAccCertificatesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_certificates.all", "certificates.#"),
				),
			},
		},
	})
}

func testAccCertificateResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_certificate" "test" {
  name             = "%s"
  certificate_data = <<-EOT
%s
EOT
  private_key      = <<-EOT
%s
EOT
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, testCertificateData, testPrivateKey)
}

func testAccCertificateResourceConfigWithPath(name, path string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_certificate" "test" {
  name             = "%s"
  certificate_data = <<-EOT
%s
EOT
  private_key      = <<-EOT
%s
EOT
  certificate_path = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, testCertificateData, testPrivateKey, path)
}

func testAccCertificateDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_certificate" "test" {
  name             = "%s"
  certificate_data = <<-EOT
%s
EOT
  private_key      = <<-EOT
%s
EOT
}

data "dokploy_certificate" "test" {
  id = dokploy_certificate.test.id
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, testCertificateData, testPrivateKey)
}

func testAccCertificatesDataSourceConfig() string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_certificate" "test" {
  name             = "test-cert-for-list"
  certificate_data = <<-EOT
%s
EOT
  private_key      = <<-EOT
%s
EOT
}

data "dokploy_certificates" "all" {
  depends_on = [dokploy_certificate.test]
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), testCertificateData, testPrivateKey)
}
