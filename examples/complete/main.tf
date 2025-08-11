# This file demonstrates a complete DNS configuration for a domain
# Split into versions.tf, variables.tf, main.tf, and outputs.tf for better organization

# Configure the DreamHost Provider
provider "dreamhost" {
  # API key sourced from environment variable DREAMHOST_API_KEY
}

# Root domain A record
resource "dreamhost_dns_record" "root_a" {
  record = var.domain_name
  type   = "A"
  value  = var.ipv4_address
}

# WWW subdomain CNAME record
resource "dreamhost_dns_record" "www_cname" {
  record = "${var.subdomain}.${var.domain_name}"
  type   = "CNAME"
  value  = var.domain_name
}

# MX records for email (using dynamic blocks for multiple servers)
resource "dreamhost_dns_record" "mx" {
  for_each = { for idx, mx in var.mail_servers : idx => mx }

  record = var.domain_name
  type   = "MX"
  value  = "${each.value.priority} ${each.value.server}"
}

# SPF record for email authentication
resource "dreamhost_dns_record" "spf" {
  record = var.domain_name
  type   = "TXT"
  value  = var.spf_record
}

# Domain verification TXT record
resource "dreamhost_dns_record" "verification" {
  record = "_verification.${var.domain_name}"
  type   = "TXT"
  value  = var.verification_token
}

# SRV records for services
resource "dreamhost_dns_record" "srv" {
  for_each = var.srv_records

  record = "_${split("_", each.key)[0]}._${split("_", each.key)[1]}.${var.domain_name}"
  type   = "SRV"
  value  = "${each.value.priority} ${each.value.weight} ${each.value.port} ${each.value.target}"
}

# IPv6 AAAA record
resource "dreamhost_dns_record" "ipv6" {
  record = var.domain_name
  type   = "AAAA"
  value  = var.ipv6_address
}