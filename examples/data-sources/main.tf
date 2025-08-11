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

# Look up a specific DNS record
data "dreamhost_dns_record" "example_a" {
  record = "example.com"
  type   = "A"
}

# Look up a specific DNS record with value (useful when multiple records of same type exist)
data "dreamhost_dns_record" "specific_mx" {
  record = "example.com"
  type   = "MX"
  value  = "10 mail.example.com"
}

# List all DNS records
data "dreamhost_dns_records" "all" {}

# List DNS records with filters
data "dreamhost_dns_records" "filtered" {
  filter {
    # Filter by record name (supports partial match)
    record = "example.com"
    # Filter by type
    type = "A"
  }
}

# List all TXT records
data "dreamhost_dns_records" "txt_records" {
  filter {
    type = "TXT"
  }
}

# List records for a specific zone
data "dreamhost_dns_records" "zone_records" {
  filter {
    zone = "example.com"
  }
}

# List records matching a value pattern
data "dreamhost_dns_records" "ip_records" {
  filter {
    value = "192.0.2"  # Will match any record with value containing "192.0.2"
  }
}

# Output examples
output "example_a_record_value" {
  description = "The IP address of example.com"
  value       = data.dreamhost_dns_record.example_a.value
}

output "example_a_record_zone" {
  description = "The zone of example.com"
  value       = data.dreamhost_dns_record.example_a.zone
}

output "all_records_count" {
  description = "Total number of DNS records"
  value       = length(data.dreamhost_dns_records.all.records)
}

output "filtered_records" {
  description = "Filtered DNS records"
  value = [for r in data.dreamhost_dns_records.filtered.records : {
    name  = r.record
    type  = r.type
    value = r.value
  }]
}

output "txt_records" {
  description = "All TXT records"
  value = {
    for r in data.dreamhost_dns_records.txt_records.records :
    r.record => r.value
  }
}