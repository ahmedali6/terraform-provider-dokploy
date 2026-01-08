---
page_title: "dokploy_application Resource - dokploy"
subcategory: ""
description: |-
  Manages a Dokploy application.
---

# dokploy_application (Resource)

Manages a Dokploy application. Applications can be deployed from various sources including:
- **Docker images** - Deploy pre-built container images
- **GitHub** - Deploy from GitHub repositories using GitHub Apps
- **GitLab** - Deploy from GitLab repositories  
- **Bitbucket** - Deploy from Bitbucket repositories
- **Gitea** - Deploy from self-hosted Gitea instances
- **Custom Git** - Deploy from any Git repository via SSH or HTTPS

## Example Usage

### Docker Image Deployment

Deploy a pre-built Docker image from Docker Hub or a private registry.

```terraform
resource "dokploy_project" "myproject" {
  name = "My Application"
}

resource "dokploy_environment" "production" {
  project_id  = dokploy_project.myproject.id
  name        = "Production"
  description = "Production environment"
}

resource "dokploy_application" "nginx" {
  name           = "nginx-app"
  environment_id = dokploy_environment.production.id
  source_type    = "docker"
  docker_image   = "nginx:alpine"
  
  replicas         = 2
  deploy_on_create = true
}
```

### Docker Image with Private Registry

```terraform
resource "dokploy_application" "private_app" {
  name           = "private-app"
  environment_id = dokploy_environment.production.id
  source_type    = "docker"
  docker_image   = "registry.example.com/myorg/myapp:latest"
  registry_url   = "registry.example.com"
  username       = "deploy"
  password       = var.registry_password
  
  deploy_on_create = true
}
```

### GitHub Repository with Nixpacks

Deploy from a GitHub repository using automatic build detection with Nixpacks.

```terraform
resource "dokploy_application" "api" {
  name           = "api"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  # GitHub settings
  github_id   = "your-github-app-installation-id"
  owner       = "myorg"
  repository  = "api"
  branch      = "main"
  
  # Build settings (Nixpacks auto-detects your project type)
  build_type = "nixpacks"
  
  # Runtime settings
  auto_deploy      = true
  deploy_on_create = true
  
  # Environment variables
  env = <<-EOT
    NODE_ENV=production
    DATABASE_URL=${var.database_url}
  EOT
}
```

### GitHub Repository with Dockerfile

```terraform
resource "dokploy_application" "web" {
  name           = "web"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  github_id   = "your-github-app-installation-id"
  owner       = "myorg"
  repository  = "web-frontend"
  branch      = "main"
  build_path  = "apps/web"  # Monorepo path
  
  # Dockerfile build
  build_type          = "dockerfile"
  dockerfile_path     = "./Dockerfile"
  docker_context_path = "."
  docker_build_stage  = "production"  # Multi-stage build target
  
  # Build arguments
  build_args = <<-EOT
    NODE_VERSION=20
    BUILD_DATE=${timestamp()}
  EOT
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### GitLab Repository

```terraform
resource "dokploy_application" "gitlab_app" {
  name           = "gitlab-app"
  environment_id = dokploy_environment.production.id
  source_type    = "gitlab"
  
  # GitLab settings
  gitlab_id         = "your-gitlab-integration-id"
  gitlab_project_id = 12345
  gitlab_owner      = "mygroup"
  gitlab_repository = "myproject"
  gitlab_branch     = "main"
  
  build_type       = "nixpacks"
  auto_deploy      = true
  deploy_on_create = true
}
```

### Bitbucket Repository

```terraform
resource "dokploy_application" "bitbucket_app" {
  name           = "bitbucket-app"
  environment_id = dokploy_environment.production.id
  source_type    = "bitbucket"
  
  # Bitbucket settings
  bitbucket_id         = "your-bitbucket-integration-id"
  bitbucket_owner      = "myworkspace"
  bitbucket_repository = "myrepo"
  bitbucket_branch     = "main"
  
  build_type       = "nixpacks"
  auto_deploy      = true
  deploy_on_create = true
}
```

### Gitea Repository

```terraform
resource "dokploy_application" "gitea_app" {
  name           = "gitea-app"
  environment_id = dokploy_environment.production.id
  source_type    = "gitea"
  
  # Gitea settings
  gitea_id         = "your-gitea-integration-id"
  gitea_owner      = "myorg"
  gitea_repository = "myrepo"
  gitea_branch     = "main"
  
  build_type       = "nixpacks"
  auto_deploy      = true
  deploy_on_create = true
}
```

### Custom Git Repository (SSH)

Deploy from any Git repository using SSH authentication.

```terraform
resource "dokploy_ssh_key" "deploy_key" {
  name        = "deploy-key"
  private_key = var.ssh_private_key
  public_key  = var.ssh_public_key
}

resource "dokploy_application" "custom_git" {
  name           = "custom-app"
  environment_id = dokploy_environment.production.id
  source_type    = "git"
  
  # Custom Git settings
  custom_git_url        = "git@github.com:myorg/private-repo.git"
  custom_git_branch     = "main"
  custom_git_ssh_key_id = dokploy_ssh_key.deploy_key.id
  custom_git_build_path = "."
  enable_submodules     = true
  
  build_type       = "dockerfile"
  dockerfile_path  = "./Dockerfile"
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Static Site Deployment

```terraform
resource "dokploy_application" "docs" {
  name           = "documentation"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  github_id   = "your-github-app-installation-id"
  owner       = "myorg"
  repository  = "docs"
  branch      = "main"
  
  # Static build
  build_type        = "static"
  publish_directory = "build"
  is_static_spa     = true  # Enable SPA routing
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Application with Resource Limits

```terraform
resource "dokploy_application" "resource_limited" {
  name           = "limited-app"
  environment_id = dokploy_environment.production.id
  source_type    = "docker"
  docker_image   = "myapp:latest"
  
  # Resource limits
  memory_limit       = 536870912   # 512MB
  memory_reservation = 268435456   # 256MB
  cpu_limit          = 1000000000  # 1 CPU core
  cpu_reservation    = 500000000   # 0.5 CPU core
  
  replicas = 3
  
  deploy_on_create = true
}
```

### Application with Preview Deployments

Enable automatic preview deployments for pull requests.

```terraform
resource "dokploy_application" "with_previews" {
  name           = "app-with-previews"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  github_id   = "your-github-app-installation-id"
  owner       = "myorg"
  repository  = "web"
  branch      = "main"
  
  build_type = "nixpacks"
  
  # Preview deployment settings
  preview_deployments_enabled = true
  preview_wildcard            = "*.preview.example.com"
  preview_port                = 3000
  preview_https               = true
  preview_certificate_type    = "letsencrypt"
  preview_limit               = 5
  
  preview_env = <<-EOT
    NODE_ENV=preview
    API_URL=https://api-preview.example.com
  EOT
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Application with Rollback Support

```terraform
resource "dokploy_application" "with_rollback" {
  name           = "rollback-app"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  github_id   = "your-github-app-installation-id"
  owner       = "myorg"
  repository  = "api"
  branch      = "main"
  
  build_type = "dockerfile"
  
  # Enable rollback
  rollback_active      = true
  rollback_registry_id = dokploy_registry.internal.id
  
  deploy_on_create = true
}
```

### Application with Remote Build Server

Build on a dedicated build server and push to a registry.

```terraform
resource "dokploy_application" "remote_build" {
  name           = "remote-build-app"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  github_id   = "your-github-app-installation-id"
  owner       = "myorg"
  repository  = "heavy-build"
  branch      = "main"
  
  build_type = "dockerfile"
  
  # Remote build configuration
  build_server_id   = dokploy_server.build.id
  build_registry_id = dokploy_registry.internal.id
  
  deploy_on_create = true
}
```

### Application with Docker Swarm Configuration

Configure advanced Docker Swarm settings using JSON format.

```terraform
resource "dokploy_application" "swarm_app" {
  name           = "swarm-configured-app"
  environment_id = dokploy_environment.production.id
  source_type    = "docker"
  docker_image   = "nginx:alpine"
  
  replicas = 3
  
  # Health check configuration
  health_check_swarm = jsonencode({
    Test     = ["CMD", "curl", "-f", "http://localhost/health"]
    Interval = 30000000000  # 30 seconds in nanoseconds
    Timeout  = 10000000000  # 10 seconds
    Retries  = 3
  })
  
  # Restart policy
  restart_policy_swarm = jsonencode({
    Condition   = "on-failure"
    MaxAttempts = 3
    Delay       = 5000000000  # 5 seconds
    Window      = 60000000000 # 60 seconds
  })
  
  # Update configuration
  update_config_swarm = jsonencode({
    Parallelism   = 1
    Delay         = 10000000000
    FailureAction = "rollback"
    Order         = "start-first"
  })
  
  # Placement constraints
  placement_swarm = jsonencode({
    Constraints = ["node.role == worker"]
  })
  
  # Stop grace period (30 seconds)
  stop_grace_period_swarm = 30000000000
  
  deploy_on_create = true
}
```

### Application with Watch Paths

Trigger deployments only when specific paths change.

```terraform
resource "dokploy_application" "monorepo_app" {
  name           = "monorepo-service"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  github_id   = "your-github-app-installation-id"
  owner       = "myorg"
  repository  = "monorepo"
  branch      = "main"
  build_path  = "services/api"
  
  # Only trigger deployment when these paths change
  watch_paths = [
    "services/api/**",
    "shared/lib/**",
    "package.json"
  ]
  
  build_type       = "nixpacks"
  auto_deploy      = true
  deploy_on_create = true
}
```

### Drop Source Deployment (File Upload)

Deploy using raw Dockerfile content for quick prototyping.

```terraform
resource "dokploy_application" "drop_app" {
  name           = "quick-deploy"
  environment_id = dokploy_environment.production.id
  source_type    = "drop"
  
  # Raw Dockerfile content
  dockerfile = <<-DOCKERFILE
    FROM nginx:alpine
    COPY . /usr/share/nginx/html
    EXPOSE 80
    CMD ["nginx", "-g", "daemon off;"]
  DOCKERFILE
  
  drop_build_path = "/app"
  
  deploy_on_create = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) The environment ID this application belongs to.
- `name` (String) The display name of the application.

### Optional

- `app_name` (String) The app name used for Docker container naming. Auto-generated if not specified.
- `args` (String) Arguments to pass to the command.
- `auto_deploy` (Boolean) Enable automatic deployment on Git push.
- `bitbucket_branch` (String) Bitbucket branch to deploy from.
- `bitbucket_build_path` (String) Build path within the Bitbucket repository.
- `bitbucket_id` (String) Bitbucket integration ID. Required for Bitbucket source type.
- `bitbucket_owner` (String) Bitbucket repository owner/workspace.
- `bitbucket_repository` (String) Bitbucket repository name.
- `branch` (String) Branch to deploy from (GitHub/GitLab/Bitbucket/Gitea).
- `build_args` (String) Build arguments in KEY=VALUE format, one per line.
- `build_path` (String) Build path within the repository for GitHub source.
- `build_registry_id` (String) Registry ID to push build images to.
- `build_secrets` (String, Sensitive) Build secrets in KEY=VALUE format, one per line.
- `build_server_id` (String) Build server ID for remote builds.
- `build_type` (String) Build type: dockerfile, heroku_buildpacks, paketo_buildpacks, nixpacks, static, or railpack.
- `clean_cache` (Boolean) Clean cache before building.
- `command` (String) Custom command to run (overrides Dockerfile CMD).
- `cpu_limit` (Number) CPU limit in nanocores. Example: 1000000000 (1 CPU).
- `cpu_reservation` (Number) CPU reservation in nanocores.
- `create_env_file` (Boolean) Create a .env file in the container.
- `custom_git_branch` (String) Branch to use for custom Git repository.
- `custom_git_build_path` (String) Build path within the custom Git repository.
- `custom_git_ssh_key_id` (String) SSH key ID for accessing the custom Git repository.
- `custom_git_url` (String) Custom Git repository URL (for source_type 'git').
- `deploy_on_create` (Boolean) Trigger a deployment after creating the application.
- `description` (String) A description of the application.
- `docker_build_stage` (String) Target stage for multi-stage Docker builds.
- `docker_context_path` (String) Docker build context path.
- `docker_image` (String) Docker image to use (for source_type 'docker'). Example: 'nginx:alpine'.
- `dockerfile` (String) Raw Dockerfile content (for 'drop' source type or inline Dockerfile).
- `dockerfile_path` (String) Path to the Dockerfile (relative to build path).
- `drop_build_path` (String) Build path for 'drop' source type deployments.
- `enable_submodules` (Boolean) Enable Git submodules support.
- `enabled` (Boolean) Whether the application is enabled.
- `endpoint_spec_swarm` (String) Endpoint specification for Docker Swarm mode (JSON format).
- `env` (String) Environment variables in KEY=VALUE format, one per line.
- `gitea_branch` (String) Gitea branch to deploy from.
- `gitea_build_path` (String) Build path within the Gitea repository.
- `gitea_id` (String) Gitea integration ID. Required for Gitea source type.
- `gitea_owner` (String) Gitea repository owner/organization.
- `gitea_repository` (String) Gitea repository name.
- `github_id` (String) GitHub App installation ID. Required for GitHub source type.
- `gitlab_branch` (String) GitLab branch to deploy from.
- `gitlab_build_path` (String) Build path within the GitLab repository.
- `gitlab_id` (String) GitLab integration ID. Required for GitLab source type.
- `gitlab_owner` (String) GitLab repository owner/group.
- `gitlab_path_namespace` (String) GitLab path namespace (for nested groups).
- `gitlab_project_id` (Number) GitLab project ID.
- `gitlab_repository` (String) GitLab repository name.
- `health_check_swarm` (String) Health check configuration for Docker Swarm mode (JSON format).
- `heroku_version` (String) Heroku buildpack version (for heroku_buildpacks build type).
- `is_static_spa` (Boolean) Whether the static build is a Single Page Application.
- `labels_swarm` (String) Labels for Docker Swarm service (JSON format).
- `memory_limit` (Number) Memory limit in bytes. Example: 536870912 (512MB).
- `memory_reservation` (Number) Memory reservation (soft limit) in bytes.
- `mode_swarm` (String) Service mode for Docker Swarm: replicated or global (JSON format).
- `network_swarm` (String) Network configuration for Docker Swarm mode (JSON array format).
- `owner` (String) Repository owner/organization for GitHub source.
- `password` (String, Sensitive) Password for Docker registry authentication.
- `placement_swarm` (String) Placement constraints for Docker Swarm mode (JSON format).
- `preview_build_args` (String) Build arguments for preview deployments.
- `preview_build_secrets` (String, Sensitive) Build secrets for preview deployments in KEY=VALUE format.
- `preview_certificate_type` (String) Certificate type for preview deployments: letsencrypt, none.
- `preview_custom_cert_resolver` (String) Custom certificate resolver for preview deployments.
- `preview_deployments_enabled` (Boolean) Enable preview deployments for pull requests.
- `preview_env` (String) Environment variables for preview deployments.
- `preview_https` (Boolean) Enable HTTPS for preview deployments.
- `preview_labels` (List of String) Labels for preview deployments.
- `preview_limit` (Number) Maximum number of concurrent preview deployments.
- `preview_path` (String) Path prefix for preview deployment URLs.
- `preview_port` (Number) Port for preview deployment containers.
- `preview_require_collaborator_permissions` (Boolean) Require collaborator permissions to create preview deployments.
- `preview_wildcard` (String) Wildcard domain for preview deployments (e.g., '*.preview.example.com').
- `publish_directory` (String) Publish directory for static builds.
- `railpack_version` (String) Railpack version (for railpack build type).
- `registry_id` (String) Registry ID from Dokploy registry management.
- `registry_url` (String) Docker registry URL. Leave empty for Docker Hub.
- `replicas` (Number) Number of container replicas to run.
- `repository` (String) Repository name for GitHub source (e.g., 'my-repo').
- `restart_policy_swarm` (String) Restart policy configuration for Docker Swarm mode (JSON format).
- `rollback_active` (Boolean) Enable rollback capability.
- `rollback_config_swarm` (String) Rollback configuration for Docker Swarm mode (JSON format).
- `rollback_registry_id` (String) Registry ID to use for rollback images.
- `server_id` (String) Server ID to deploy the application to. If not specified, deploys to the default server.
- `source_type` (String) The source type for the application: github, gitlab, bitbucket, gitea, git, docker, or drop.
- `stop_grace_period_swarm` (Number) Stop grace period in nanoseconds for Docker Swarm mode.
- `subtitle` (String) Display subtitle for the application in the UI.
- `title` (String) Display title for the application in the UI.
- `trigger_type` (String) Trigger type for deployments: 'push' (default) or 'tag'.
- `update_config_swarm` (String) Update configuration for Docker Swarm mode (JSON format).
- `username` (String) Username for Docker registry authentication.
- `watch_paths` (List of String) Paths to watch for changes to trigger deployments.

### Read-Only

- `application_status` (String) Current status of the application: idle, running, done, error.
- `id` (String) The unique identifier of the application.

## Import

Import is supported using the following syntax:

```shell
terraform import dokploy_application.myapp "application-id-123"
```
