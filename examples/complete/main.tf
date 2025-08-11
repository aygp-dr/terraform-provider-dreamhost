terraform {
  required_providers {
    dreamhost = {
      version = "0.0.1"
      source  = "hashicorp.com/edu/dreamhost"
    }
  }
}

# Configure the DreamHost Provider
# The API key can also be set via DREAMHOST_API_KEY environment variable
provider "dreamhost" {
  # api_key = var.dreamhost_api_key
}

# Variables for configuration
variable "domain_name" {
  description = "The domain name to manage DNS records for"
  type        = string
  default     = "example.com"
}

variable "subdomain" {
  description = "Subdomain prefix"
  type        = string
  default     = "www"
}

# Create an A record for the root domain
resource "dreamhost_dns_record" "root_a" {
  record = var.domain_name
  type   = "A"
  value  = "192.0.2.1"
}

# Create a www subdomain pointing to the root domain
resource "dreamhost_dns_record" "www_cname" {
  record = "${var.subdomain}.${var.domain_name}"
  type   = "CNAME"
  value  = var.domain_name
}

# Create an MX record for email
resource "dreamhost_dns_record" "mx_primary" {
  record = var.domain_name
  type   = "MX"
  value  = "10 mail.${var.domain_name}"
}

resource "dreamhost_dns_record" "mx_secondary" {
  record = var.domain_name
  type   = "MX"
  value  = "20 mail2.${var.domain_name}"
}

# Create a TXT record for SPF
resource "dreamhost_dns_record" "spf" {
  record = var.domain_name
  type   = "TXT"
  value  = "v=spf1 include:_spf.dreamhost.com ~all"
}

# Create a TXT record for domain verification
resource "dreamhost_dns_record" "verification" {
  record = "_verification.${var.domain_name}"
  type   = "TXT"
  value  = "verification-token-12345"
}

# Create an SRV record for a service
resource "dreamhost_dns_record" "srv_example" {
  record = "_sip._tcp.${var.domain_name}"
  type   = "SRV"
  value  = "10 60 5060 sipserver.${var.domain_name}"
}

# Create AAAA record for IPv6
resource "dreamhost_dns_record" "ipv6" {
  record = var.domain_name
  type   = "AAAA"
  value  = "2001:db8::1"
}

# Output the created records
output "root_domain_a_record" {
  value = dreamhost_dns_record.root_a.value
}

output "www_cname_record" {
  value = dreamhost_dns_record.www_cname.value
}