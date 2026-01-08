---
page_title: "dokploy_compose Resource - dokploy"
subcategory: ""
description: |-
  Manages a Docker Compose stack in Dokploy.
---

# dokploy_compose (Resource)

Manages a Docker Compose stack in Dokploy. Deploy multi-container applications using docker-compose files from various sources:
- **Raw** - Inline docker-compose.yml content
- **GitHub** - Deploy from GitHub repositories using GitHub Apps
- **GitLab** - Deploy from GitLab repositories
- **Bitbucket** - Deploy from Bitbucket repositories
- **Gitea** - Deploy from self-hosted Gitea instances
- **Custom Git** - Deploy from any Git repository via SSH or HTTPS

## Example Usage

### Inline Compose File (Raw)

Deploy a compose stack with inline docker-compose.yml content.

```terraform
resource "dokploy_project" "myproject" {
  name = "My Stack"
}

resource "dokploy_environment" "production" {
  project_id  = dokploy_project.myproject.id
  name        = "Production"
  description = "Production environment"
}

resource "dokploy_compose" "wordpress" {
  name           = "wordpress-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "raw"
  
  compose_file_content = <<-EOT
    version: '3.8'
    services:
      wordpress:
        image: wordpress:latest
        ports:
          - "8080:80"
        environment:
          WORDPRESS_DB_HOST: db:3306
          WORDPRESS_DB_USER: wordpress
          WORDPRESS_DB_PASSWORD: wordpress_password
          WORDPRESS_DB_NAME: wordpress
        depends_on:
          - db
          
      db:
        image: mysql:8.0
        environment:
          MYSQL_DATABASE: wordpress
          MYSQL_USER: wordpress
          MYSQL_PASSWORD: wordpress_password
          MYSQL_ROOT_PASSWORD: root_password
        volumes:
          - db_data:/var/lib/mysql
          
    volumes:
      db_data:
  EOT
  
  deploy_on_create = true
}
```

### GitHub Repository

Deploy a compose stack from a GitHub repository.

```terraform
resource "dokploy_compose" "github_stack" {
  name           = "my-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  # GitHub settings
  github_id    = "your-github-app-installation-id"
  owner        = "myorg"
  repository   = "docker-stack"
  branch       = "main"
  compose_path = "./docker-compose.yml"
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### GitLab Repository

```terraform
resource "dokploy_compose" "gitlab_stack" {
  name           = "gitlab-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "gitlab"
  
  # GitLab settings
  gitlab_id         = "your-gitlab-integration-id"
  gitlab_project_id = 12345
  gitlab_owner      = "mygroup"
  gitlab_repository = "docker-stack"
  gitlab_branch     = "main"
  compose_path      = "./docker-compose.yml"
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Bitbucket Repository

```terraform
resource "dokploy_compose" "bitbucket_stack" {
  name           = "bitbucket-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "bitbucket"
  
  # Bitbucket settings
  bitbucket_id         = "your-bitbucket-integration-id"
  bitbucket_owner      = "myworkspace"
  bitbucket_repository = "docker-stack"
  bitbucket_branch     = "main"
  compose_path         = "./docker-compose.yml"
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Gitea Repository

```terraform
resource "dokploy_compose" "gitea_stack" {
  name           = "gitea-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "gitea"
  
  # Gitea settings
  gitea_id         = "your-gitea-integration-id"
  gitea_owner      = "myorg"
  gitea_repository = "docker-stack"
  gitea_branch     = "main"
  compose_path     = "./docker-compose.yml"
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Custom Git Repository (SSH)

Deploy from a private Git repository using SSH authentication.

```terraform
resource "dokploy_ssh_key" "deploy_key" {
  name        = "compose-deploy-key"
  private_key = var.ssh_private_key
  public_key  = var.ssh_public_key
}

resource "dokploy_compose" "private_stack" {
  name           = "private-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "git"
  
  # Custom Git settings
  custom_git_url        = "git@github.com:myorg/private-stack.git"
  custom_git_branch     = "main"
  custom_git_ssh_key_id = dokploy_ssh_key.deploy_key.id
  compose_path          = "./docker-compose.yml"
  enable_submodules     = true
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Compose with Environment Variables

```terraform
resource "dokploy_compose" "with_env" {
  name           = "env-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "github"
  
  github_id    = "your-github-app-installation-id"
  owner        = "myorg"
  repository   = "my-stack"
  branch       = "main"
  compose_path = "./docker-compose.yml"
  
  # Environment variables available to compose file
  env = <<-EOT
    DATABASE_URL=postgresql://user:pass@db:5432/app
    REDIS_URL=redis://redis:6379
    SECRET_KEY=${var.secret_key}
  EOT
  
  auto_deploy      = true
  deploy_on_create = true
}
```

### Compose on Specific Server

Deploy to a specific server in your cluster.

```terraform
resource "dokploy_compose" "server_specific" {
  name           = "server-stack"
  environment_id = dokploy_environment.production.id
  source_type    = "raw"
  server_id      = dokploy_server.worker.id
  
  compose_file_content = <<-EOT
    version: '3.8'
    services:
      worker:
        image: myworker:latest
        deploy:
          replicas: 3
  EOT
  
  deploy_on_create = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) The environment ID this compose stack belongs to.
- `name` (String) The display name of the compose stack.

### Optional

- `app_name` (String) The app name used for Docker service naming. Auto-generated if not specified.
- `auto_deploy` (Boolean) Enable automatic deployment on Git push. Defaults to API default (typically true).
- `bitbucket_branch` (String) Bitbucket branch to deploy from.
- `bitbucket_build_path` (String) Build path within the Bitbucket repository.
- `bitbucket_id` (String) Bitbucket integration ID. Required for Bitbucket source type.
- `bitbucket_owner` (String) Bitbucket repository owner/workspace.
- `bitbucket_repository` (String) Bitbucket repository name.
- `branch` (String) Branch to deploy from (GitHub/GitLab/Bitbucket/Gitea).
- `compose_file_content` (String) Raw docker-compose.yml content (for source_type 'raw').
- `compose_path` (String) Path to the docker-compose.yml file in the repository.
- `custom_git_branch` (String) Branch to use for custom Git repository.
- `custom_git_build_path` (String) Build path within the custom Git repository.
- `custom_git_ssh_key_id` (String) SSH key ID for accessing the custom Git repository.
- `custom_git_url` (String) Custom Git repository URL (for source_type 'git').
- `deploy_on_create` (Boolean) Trigger a deployment after creating the compose stack.
- `description` (String) A description of the compose stack.
- `enable_submodules` (Boolean) Enable Git submodules support.
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
- `owner` (String) Repository owner/organization for GitHub source.
- `repository` (String) Repository name for GitHub source (e.g., 'my-repo').
- `server_id` (String) Server ID to deploy the compose stack to. If not specified, deploys to the default server.
- `source_type` (String) The source type for the compose stack: github, gitlab, bitbucket, gitea, git, or raw.
- `trigger_type` (String) Trigger type for deployments: 'push' (default) or 'tag'.

### Read-Only

- `id` (String) The unique identifier of the compose stack.

## Import

Import is supported using the following syntax:

```shell
terraform import dokploy_compose.wordpress "compose-id-123"
```
