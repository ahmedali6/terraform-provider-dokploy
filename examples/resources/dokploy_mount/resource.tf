# Volume mount example
resource "dokploy_mount" "app_data" {
  service_id   = dokploy_application.myapp.id
  service_type = "application"
  type         = "volume"
  volume_name  = "app-data"
  mount_path   = "/app/data"
}

# Bind mount example
resource "dokploy_mount" "config" {
  service_id   = dokploy_application.myapp.id
  service_type = "application"
  type         = "bind"
  host_path    = "/host/config"
  mount_path   = "/app/config"
}

# File mount example
resource "dokploy_mount" "env_file" {
  service_id   = dokploy_application.myapp.id
  service_type = "application"
  type         = "file"
  file_path    = "/app/.env"
  mount_path   = "/app/.env"
  content      = <<-EOT
    DATABASE_URL=postgresql://user:pass@db:5432/mydb
    REDIS_URL=redis://redis:6379
  EOT
}

# Database volume mount example
resource "dokploy_mount" "postgres_data" {
  service_id   = dokploy_database.postgres.id
  service_type = "postgres"
  type         = "volume"
  volume_name  = "postgres-data"
  mount_path   = "/var/lib/postgresql/data"
}
