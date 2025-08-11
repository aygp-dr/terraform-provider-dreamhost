# Terraform Provider DreamHost

The DreamHost Terraform Provider allows you to manage DreamHost DNS records using Infrastructure as Code.

## Features

- ✅ Full CRUD operations for DNS records
- ✅ Support for all major DNS record types (A, AAAA, CNAME, MX, NS, PTR, TXT, SRV, NAPTR)
- ✅ Data sources for querying existing DNS records
- ✅ Import existing DNS records into Terraform state
- ✅ Automatic retry logic for API errors
- ✅ DNS record validation
- ✅ Caching for improved performance

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13
- [Go](https://golang.org/doc/install) >= 1.19 (for building from source)
- DreamHost API key with DNS permissions

## Getting Started

### DreamHost API Key

1. Log in to your DreamHost panel
2. Navigate to [API Keys](https://panel.dreamhost.com/index.cgi?tree=home.api)
3. Create a new API key with "All dns functions" permissions
4. Save the API key securely

### Provider Configuration

```hcl
terraform {
  required_providers {
    dreamhost = {
      version = "~> 0.0.1"
      source  = "hashicorp.com/edu/dreamhost"
    }
  }
}

provider "dreamhost" {
  api_key = var.dreamhost_api_key  # Or use DREAMHOST_API_KEY env var
}
```

## Usage Examples

### Creating DNS Records

```hcl
# A record
resource "dreamhost_dns_record" "example_a" {
  record = "example.com"
  type   = "A"
  value  = "192.0.2.1"
}

# CNAME record
resource "dreamhost_dns_record" "www" {
  record = "www.example.com"
  type   = "CNAME"
  value  = "example.com"
}

# MX record
resource "dreamhost_dns_record" "mail" {
  record = "example.com"
  type   = "MX"
  value  = "10 mail.example.com"
}

# TXT record (for SPF, DKIM, etc.)
resource "dreamhost_dns_record" "spf" {
  record = "example.com"
  type   = "TXT"
  value  = "v=spf1 include:_spf.dreamhost.com ~all"
}
```

### Data Sources

```hcl
# Look up a specific DNS record
data "dreamhost_dns_record" "existing" {
  record = "example.com"
  type   = "A"
}

# List all DNS records with filters
data "dreamhost_dns_records" "all_txt" {
  filter {
    type = "TXT"
  }
}

# Use data source values
resource "dreamhost_dns_record" "backup" {
  record = "backup.example.com"
  type   = "A"
  value  = data.dreamhost_dns_record.existing.value
}
```

### Importing Existing Records

```shell
# Import format: TYPE|RECORD|VALUE
terraform import dreamhost_dns_record.existing "A|example.com|192.0.2.1"
```

## Building from Source

```shell
# Clone the repository
git clone https://github.com/aygp-dr/terraform-provider-dreamhost.git
cd terraform-provider-dreamhost

# Build the provider
go build -o terraform-provider-dreamhost

# Install locally
make install
```

## Testing

```shell
# Run unit tests
make test

# Run acceptance tests (requires DREAMHOST_API_KEY)
make testacc
```

## Documentation

- [Provider Documentation](docs/index.md)
- [DNS Record Resource](docs/resources/dns_record.md)
- [DNS Record Data Source](docs/data-sources/dns_record.md)
- [DNS Records Data Source](docs/data-sources/dns_records.md)
- [Examples](examples/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is distributed under the MIT License. See LICENSE file for details.