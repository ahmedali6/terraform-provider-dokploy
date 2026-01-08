# GitHub Container Registry (ghcr.io)
resource "dokploy_registry" "ghcr" {
  registry_name = "GitHub Container Registry"
  registry_type = "cloud"
  registry_url  = "ghcr.io"
  username      = "myorg"
  password      = var.github_token
  image_prefix  = "ghcr.io/myorg"
}

# Docker Hub
resource "dokploy_registry" "dockerhub" {
  registry_name = "Docker Hub"
  registry_type = "cloud"
  registry_url  = "docker.io"
  username      = "myusername"
  password      = var.dockerhub_token
  image_prefix  = "docker.io/myusername"
}

# AWS ECR
resource "dokploy_registry" "ecr" {
  registry_name = "AWS ECR"
  registry_type = "cloud"
  registry_url  = "123456789012.dkr.ecr.us-east-1.amazonaws.com"
  username      = "AWS"
  password      = var.ecr_token
  image_prefix  = "123456789012.dkr.ecr.us-east-1.amazonaws.com/myapp"
}

# Google Container Registry (GCR)
resource "dokploy_registry" "gcr" {
  registry_name = "Google Container Registry"
  registry_type = "cloud"
  registry_url  = "gcr.io"
  username      = "_json_key"
  password      = var.gcp_service_account_key
  image_prefix  = "gcr.io/my-project"
}

# Registry with specific server
resource "dokploy_registry" "build_server_registry" {
  registry_name = "Build Server Registry"
  registry_type = "cloud"
  registry_url  = "registry.example.com"
  username      = "admin"
  password      = var.registry_password
  image_prefix  = "registry.example.com/myapp"
  server_id     = dokploy_server.build.id
}

# Self-hosted Harbor registry
resource "dokploy_registry" "harbor" {
  registry_name = "Harbor Registry"
  registry_type = "cloud"
  registry_url  = "harbor.example.com"
  username      = "robot$myrobot"
  password      = var.harbor_robot_token
  image_prefix  = "harbor.example.com/library"
}
