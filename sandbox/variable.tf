variable "host" {
  type        = string
  description = "Dokploy host URL"
}

variable "api_key" {
  type        = string
  sensitive   = true
  description = "Dokploy API key"
}

variable "db_password" {
  type        = string
  sensitive   = true
  description = "Database password"
}

variable "github_token" {
  type        = string
  sensitive   = true
  description = "GitHub Personal Access Token"
}

variable "github_owner" {
  type        = string
  description = "GitHub Organization or User"
  default     = "j0bIT"
}

variable "custom_git_url" {
  type        = string
  description = "SSH URL for the custom git repository"
}

variable "compose_service_name" {
  type        = string
  description = "The name of the service within the compose stack to associate with the domain"
}
variable "github_app_id" {
  type        = string
  description = "GitHub App Installation ID for Dokploy"
}

variable "github_repository" {
  type        = string
  description = "GitHub repository name (without owner)"
}