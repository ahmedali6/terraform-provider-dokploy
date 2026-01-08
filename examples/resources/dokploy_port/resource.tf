# Expose PostgreSQL port
resource "dokploy_port" "postgres" {
  application_id = dokploy_application.postgres.id
  published_port = 5432
  target_port    = 5432
  protocol       = "tcp"
  publish_mode   = "host"
}

# Expose Redis port
resource "dokploy_port" "redis" {
  application_id = dokploy_application.redis.id
  published_port = 6379
  target_port    = 6379
  protocol       = "tcp"
  publish_mode   = "ingress"
}

# Expose custom UDP service
resource "dokploy_port" "dns" {
  application_id = dokploy_application.dns_server.id
  published_port = 53
  target_port    = 53
  protocol       = "udp"
  publish_mode   = "host"
}

# Expose gRPC service
resource "dokploy_port" "grpc" {
  application_id = dokploy_application.api.id
  published_port = 50051
  target_port    = 50051
  protocol       = "tcp"
  publish_mode   = "ingress"
}
