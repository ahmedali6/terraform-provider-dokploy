package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAIResource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	if openaiKey == "" {
		t.Skip("OPENAI_API_KEY must be set for AI acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAIResourceConfig("test-ai-config", openaiKey, "gpt-4o"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_ai.test", "name", "test-ai-config"),
					resource.TestCheckResourceAttr("dokploy_ai.test", "api_url", "https://api.openai.com/v1"),
					resource.TestCheckResourceAttr("dokploy_ai.test", "model", "gpt-4o"),
					resource.TestCheckResourceAttr("dokploy_ai.test", "is_enabled", "true"),
					resource.TestCheckResourceAttrSet("dokploy_ai.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_ai.test", "organization_id"),
					resource.TestCheckResourceAttrSet("dokploy_ai.test", "created_at"),
				),
			},
			// Update testing
			{
				Config: testAccAIResourceConfig("test-ai-config-updated", openaiKey, "gpt-4o-mini"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_ai.test", "name", "test-ai-config-updated"),
					resource.TestCheckResourceAttr("dokploy_ai.test", "model", "gpt-4o-mini"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "dokploy_ai.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"}, // API key preserved from config
			},
		},
	})
}

func TestAccAIResourceDisabled(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	if openaiKey == "" {
		t.Skip("OPENAI_API_KEY must be set for AI acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create disabled AI config
			{
				Config: testAccAIResourceConfigDisabled("test-ai-disabled", openaiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_ai.test", "name", "test-ai-disabled"),
					resource.TestCheckResourceAttr("dokploy_ai.test", "is_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAIsDataSource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	if openaiKey == "" {
		t.Skip("OPENAI_API_KEY must be set for AI acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAIsDataSourceConfig(openaiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_ais.all", "ais.#"),
				),
			},
		},
	})
}

func TestAccAIModelsDataSource(t *testing.T) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if host == "" || apiKey == "" {
		t.Skip("DOKPLOY_HOST and DOKPLOY_API_KEY must be set for acceptance tests")
	}

	if openaiKey == "" {
		t.Skip("OPENAI_API_KEY must be set for AI acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAIModelsDataSourceConfig(openaiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dokploy_ai_models.openai", "models.#"),
					// At least one model should be available
					resource.TestCheckResourceAttrSet("data.dokploy_ai_models.openai", "models.0.id"),
				),
			},
		},
	})
}

func testAccAIResourceConfig(name, openaiKey, model string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_ai" "test" {
  name       = "%s"
  api_url    = "https://api.openai.com/v1"
  api_key    = "%s"
  model      = "%s"
  is_enabled = true
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, openaiKey, model)
}

func testAccAIResourceConfigDisabled(name, openaiKey string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

resource "dokploy_ai" "test" {
  name       = "%s"
  api_url    = "https://api.openai.com/v1"
  api_key    = "%s"
  model      = "gpt-4o"
  is_enabled = false
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), name, openaiKey)
}

func testAccAIsDataSourceConfig(openaiKey string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

# Create an AI config first so we have data to read
resource "dokploy_ai" "test" {
  name       = "test-ai-for-datasource"
  api_url    = "https://api.openai.com/v1"
  api_key    = "%s"
  model      = "gpt-4o"
  is_enabled = true
}

data "dokploy_ais" "all" {
  depends_on = [dokploy_ai.test]
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), openaiKey)
}

func testAccAIModelsDataSourceConfig(openaiKey string) string {
	return fmt.Sprintf(`
provider "dokploy" {
  host    = "%s"
  api_key = "%s"
}

data "dokploy_ai_models" "openai" {
  api_url = "https://api.openai.com/v1"
  api_key = "%s"
}
`, os.Getenv("DOKPLOY_HOST"), os.Getenv("DOKPLOY_API_KEY"), openaiKey)
}
