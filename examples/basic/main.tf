# Configure the DreamHost Provider
provider "dreamhost" {
  api_key = var.dreamhost_api_key != "" ? var.dreamhost_api_key : null
}

# Create an A record for the domain
resource "dreamhost_dns_record" "main" {
  record = var.domain_name
  type   = "A"
  value  = var.ip_address
}