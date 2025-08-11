terraform {
  required_version = ">= 1.0"
  required_providers {
    dreamhost = {
      source  = "aygp-dr/dreamhost"
      version = "~> 0.1.0"
    }
  }
}

provider "dreamhost" {
  # api_key = var.dreamhost_api_key
}

# Example: Importing existing DNS records

# First, define the resource in your configuration
resource "dreamhost_dns_record" "existing_a_record" {
  record = "example.com"
  type   = "A"
  value  = "192.0.2.1"
}

resource "dreamhost_dns_record" "existing_cname" {
  record = "www.example.com"
  type   = "CNAME"
  value  = "example.com."
}

resource "dreamhost_dns_record" "existing_mx" {
  record = "example.com"
  type   = "MX"
  value  = "10 mail.example.com"
}

# Import commands to run:
# terraform import dreamhost_dns_record.existing_a_record "A|example.com|192.0.2.1"
# terraform import dreamhost_dns_record.existing_cname "CNAME|www.example.com|example.com."
# terraform import dreamhost_dns_record.existing_mx "MX|example.com|10 mail.example.com"

# The import ID format is: TYPE|RECORD|VALUE
# Where:
#   TYPE   = DNS record type (A, AAAA, CNAME, MX, TXT, etc.)
#   RECORD = The DNS record name
#   VALUE  = The DNS record value

# You can also use data sources to discover existing records before importing
data "dreamhost_dns_records" "discover" {}

output "discovered_records" {
  description = "All discovered DNS records that can be imported"
  value = [for r in data.dreamhost_dns_records.discover.records : {
    import_id = r.id
    command   = "terraform import dreamhost_dns_record.${replace(replace(r.record, ".", "_"), "-", "_")} \"${r.id}\""
    record    = r.record
    type      = r.type
    value     = r.value
  }]
}