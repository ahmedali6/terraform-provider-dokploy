package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResources(t *testing.T) {
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
				Config: testAccResourcesConfig("TestProjectFull", "staging"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.full", "name", "TestProjectFull"),
					resource.TestCheckResourceAttr("dokploy_environment.staging", "name", "staging"),
					resource.TestCheckResourceAttr("dokploy_application.app", "name", "test-app"),
					resource.TestCheckResourceAttr("dokploy_database.db", "name", "test-db"),
					resource.TestCheckResourceAttr("dokploy_domain.domain", "host", "test-app.example.com"),
					resource.TestCheckResourceAttr("dokploy_ssh_key.key", "name", "test-key"),
				),
			},
		},
	})
}

func testAccResourcesConfig(projectName, envName string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_project" "full" {
  name        = "%s"
  description = "Full integration test project"
}

resource "dokploy_environment" "staging" {
  project_id = dokploy_project.full.id
  name       = "%s"
}

resource "dokploy_ssh_key" "key" {
  name        = "test-key"
  description = "Integration test key"
  private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwOiIqIvWrBVoGBGZ5lAQpPtkfPly5oJyIV6FOIX0W7XQ74G4
U/RTbvgpH/u6y7ax0gQz1F3pAT1xPWTbtWk1JjCuZPZDQtCyGfppvqfpSrMJ0jMq
hGug0FM5uIaqTNYGASWPnHQaTd066Hu/0n7pSIkNn6BfX2GoXKu0AN88ZGdql46U
MwrtiviYY/IYliTHo81sBJ+ROiUN66mZvPp22GjGX7cZy6318talLw6KEhCILL5X
cIfH8JBHTYjGAX0J+NLKWfA89mvzaiV9KuQu2UEOv4Wfx7D0V7o8KwL+vU5hVSTg
SK/WNN40a8vUh7qcKWUfQ5s/yQT9a2H0IhTPpQIDAQABAoIBAHe1lWBqbswY+KC/
A279zmZjLqezMI9E8cgtXKSH0+y5di+6owVOQBxD2VlkoDVcaRq3yuYFZNuFImmD
1ifMYtQRL5etjm2/Dla6o7ZRgp79XaHYaJEGLSQ6ET6mKrZFtjIt+eGP1ubs22p7
qLzaTq/ZhN/C6IxLEOx5DNnhrhtrTxiaS/XDPPZ86c0R/L6++wu8pghp4awDq7T6
sgKD+NNSf/OPPs8U8KuejnDnLM8phEQa3GWKbsK6OMM5OpbsEUorKm1NNG4A7Ta6
6YXzn9cOnIlmrudaLLdH+LcU4nointRzoHVo0F3987ZWyOb5QpqMAFYH4rZzWjyi
A1oIswECgYEA/9YJ/FlGzrhVCbgHxGrcAO5WKAE9jVVdEpJWf6khv1v8gxaQZPiT
ytUDSilTbw5V9maOREq5jeXJt0CzKYv4VTqJWg3N5jOxK4tjMe892Ei/f9mBjTo5
VL9PozSejtPs0xCBHRJd66O4oa6fvZXvsG0L7b/svv6/k8mNMR7PccECgYEAwQgs
eEalBrWe+q7JfYonW90xnBhIKYlpQRv2Dam68mV+19c9nedKfP6lfEy4B3rcUavO
ShkamULmRCBvrPZ+OjisGS6Lg0aOxeIYKJgRYzGVzFwt1GkwuCEbtK8gQwYcDLWw
pGFjAgJ+aZaqDt0U/lTg1s6oG/jFE+xfIeYrjuUCgYAp5fbashBLmJqsrcvv2gRP
zrijcpiPBfTpLrglziAtw7XWDiL1tlQV+s7AHYaBgJqJcQBJpOdAmupvLZRp8Hdq
Fd0at3oKAls4o2pKakD5o+hg2tFBvNBY1cAIHXz+LSzy4BgVF8Xz6ms5Z2zX0q9f
eGxksnLmULg1TuPlsIMOwQKBgQCq9As2RhJ3G7h9ePi3dvgekooSHSsjpi+SWyoR
taT8ccjhbR2Dz8gXZQq4R9Wfwj6HEiozU4JMA4SdB0EAJJlsBK7d6mviSkW9mNwe
b3bOq6ZoA6cO/E4KdD/uSD7BPRLwaqTTH/CoYf8EkktvuHqUOCWb+A/IHgyj9W3X
PtqKxQKBgDJ048knd92Gn010Iidk4ps4CvHF/f/A0aeuTVzRvkHn+QLn26yLtagF
xk/PTaeoEqfZpxIvr6CHF+3evoZ9cIP9pn0oaAlIiIhhLR47nLL8lR1BNyQKk/+g
X5Bc7EtwRsGPq4byz9qcdi6YAYFpWV/YAmHr0d/Lek5Lmjp7LzN9
-----END RSA PRIVATE KEY-----
EOF
  public_key  = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDA6Iioi9asFWgYEZnmUBCk+2R8+XLmgnIhXoU4hfRbtdDvgbhT9FNu+Ckf+7rLtrHSBDPUXekBPXE9ZNu1aTUmMK5k9kNC0LIZ+mm+p+lKswnSMyqEa6DQUzm4hqpM1gYBJY+cdBpN3Troe7/SfulIiQ2foF9fYahcq7QA3zxkZ2qXjpQzCu2K+Jhj8hiWJMejzWwEn5E6JQ3rqZm8+nbYaMZftxnLrfXy1qUvDooSEIgsvldwh8fwkEdNiMYBfQn40spZ8Dz2a/NqJX0q5C7ZQQ6/hZ/HsPRXujwrAv69TmFVJOBIr9Y03jRry9SHupwpZR9Dmz/JBP1rYfQiFM+l jonaspohlmann@Jonass-MacBook-Pro.local"
}

resource "dokploy_application" "app" {
  project_id     = dokploy_project.full.id
  environment_id = dokploy_environment.staging.id
  name           = "test-app"
  repository_url = "https://github.com/dokploy/dokploy" # Using a public repo
  branch         = "main"
  build_type     = "nixpacks"
}

resource "dokploy_database" "db" {
  project_id     = dokploy_project.full.id
  environment_id = dokploy_environment.staging.id
  name           = "test-db"
  type           = "postgres"
  password       = "securepassword123"
  version        = "15"
}

resource "dokploy_domain" "domain" {
  application_id = dokploy_application.app.id
  host           = "test-app.example.com"
  port           = 3000
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), projectName, envName)
}
