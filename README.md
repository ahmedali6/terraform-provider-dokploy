# Terraform Provider for Dokploy

This is a Terraform provider for [Dokploy](https://dokploy.com/), allowing you to manage Dokploy resources such as projects, applications, databases, and more using Infrastructure as Code.

## Features

### Resources
- **Projects** - Organize your infrastructure
- **Environments** - Manage deployment environments (staging, production, etc.)
- **Applications** - Deploy applications from Git (GitHub, custom Git, etc.)
- **Databases** - Provision databases (PostgreSQL, MySQL, MongoDB, MariaDB, Redis)
- **Compose** - Deploy Docker Compose stacks
- **Domains** - Configure domains and routing
- **Environment Variables** - Manage application configuration
- **SSH Keys** - Handle Git repository authentication
- **Mounts** - Configure volume, bind, and file mounts
- **Ports** - Manage port mappings for non-HTTP services
- **Redirects** - Set up URL redirects and rewrites
- **Registry** - Configure Docker registry credentials

### Data Sources
- **GitHub Providers** - Query configured GitHub integrations
- **Servers** - Retrieve information about Dokploy servers

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- A [Dokploy](https://dokploy.com/) instance with API access

## Using the Provider

### Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    dokploy = {
      source  = "ahmedali6/dokploy"
      version = "~> 0.1"
    }
  }
}

provider "dokploy" {
  host    = "https://your-dokploy-instance.com"
  api_key = var.dokploy_api_key  # Store securely!
}
```

### Quick Example

```hcl
# Create a project
resource "dokploy_project" "main" {
  name        = "my-project"
  description = "My application infrastructure"
}

# Create an environment
resource "dokploy_environment" "production" {
  name       = "production"
  project_id = dokploy_project.main.id
}

# Deploy an application
resource "dokploy_application" "web" {
  name               = "web-app"
  project_id         = dokploy_project.main.id
  environment_id     = dokploy_environment.production.id
  custom_git_url     = "https://github.com/username/repo.git"
  custom_git_branch  = "main"
  build_type         = "dockerfile"
  auto_deploy        = true
}

# Configure a domain
resource "dokploy_domain" "web_domain" {
  application_id      = dokploy_application.web.id
  generate_traefik_me = true
  port                = 3000
  https               = true
}
```

For more examples, see the [examples](./examples/) directory and [documentation](./docs/).

## Building The Provider

1. Clone the repository:
```shell
git clone https://github.com/ahmedali6/terraform-provider-dokploy.git
cd terraform-provider-dokploy
```

2. Build the provider:
```shell
go build .
```

## Documentation

Full documentation is available in the [docs](./docs/) folder, including:
- Resource schemas and examples
- Data source references
- Import instructions

To generate documentation:

```shell
make generate
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine.

### Local Development

1. Build and install locally:
```shell
go install
```

2. Configure Terraform to use your local build:
```shell
# Create or edit ~/.terraformrc
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "ahmedali6/dokploy" = "/path/to/your/go/bin"
  }
  direct {}
}
EOF
```

### Running Tests

First, create a `.env` file from the template:

```shell
cp .env.example .env
# Edit .env and set:
# - DOKPLOY_HOST=https://your-instance.com
# - DOKPLOY_API_KEY=your-api-key
# - TF_ACC=1
```

Run the tests:

```shell
go test -v ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is published under the same license as the original project.
