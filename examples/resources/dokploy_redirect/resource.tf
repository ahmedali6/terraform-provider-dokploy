# Redirect www to non-www (permanent)
resource "dokploy_redirect" "www_to_non_www" {
  application_id = dokploy_application.myapp.id
  regex          = "^https://www\\.example\\.com/(.*)"
  replacement    = "https://example.com/$1"
  permanent      = true
}

# Redirect HTTP to HTTPS (permanent)
resource "dokploy_redirect" "http_to_https" {
  application_id = dokploy_application.myapp.id
  regex          = "^http://(.+)"
  replacement    = "https://$1"
  permanent      = true
}

# Redirect old blog path to new path (temporary)
resource "dokploy_redirect" "old_blog" {
  application_id = dokploy_application.myapp.id
  regex          = "^https://example\\.com/old-blog/(.*)"
  replacement    = "https://example.com/blog/$1"
  permanent      = false
}

# Redirect legacy API endpoint
resource "dokploy_redirect" "legacy_api" {
  application_id = dokploy_application.api.id
  regex          = "^https://api\\.example\\.com/v1/(.*)"
  replacement    = "https://api.example.com/v2/$1"
  permanent      = true
}

# Redirect subdomain to main domain
resource "dokploy_redirect" "subdomain" {
  application_id = dokploy_application.myapp.id
  regex          = "^https://old\\.example\\.com/(.*)"
  replacement    = "https://example.com/$1"
  permanent      = true
}
