# Terraform Provider for DreamHost

[![Go Version](https://img.shields.io/github/go-mod/go-version/aygp-dr/terraform-provider-dreamhost)](https://go.dev/)
[![Terraform Version](https://img.shields.io/badge/Terraform-%3E%3D1.0-623CE4?logo=terraform)](https://www.terraform.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/aygp-dr/terraform-provider-dreamhost/graphs/commit-activity)

The DreamHost Terraform Provider enables Infrastructure as Code management of DreamHost DNS records.

## Features

- 🚀 **Full CRUD Operations** - Create, read, update, and delete DNS records
- 📝 **Comprehensive DNS Support** - A, AAAA, CNAME, MX, NS, PTR, TXT, SRV, NAPTR records
- 🔍 **Data Sources** - Query and filter existing DNS records
- ♻️ **Import Support** - Import existing DNS records into Terraform state
- 🔄 **Automatic Retry** - Built-in retry logic for API rate limiting
- ✅ **Validation** - DNS record type and value validation
- ⚡ **Performance** - Intelligent caching with automatic invalidation

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19 (only for building from source)
- DreamHost API key with DNS permissions

## Quick Start

### 1. Obtain DreamHost API Key

1. Log in to your [DreamHost Panel](https://panel.dreamhost.com)
2. Navigate to [API Keys](https://panel.dreamhost.com/index.cgi?tree=home.api)
3. Generate a new key with "All dns functions" permissions
4. Store the key securely

### 2. Configure Provider

```hcl
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
  # API key can be set via:
  # 1. Provider configuration (not recommended)
  # 2. DREAMHOST_API_KEY environment variable (recommended)
  # api_key = var.dreamhost_api_key
}
```

### 3. Manage DNS Records

```hcl
# Create an A record
resource "dreamhost_dns_record" "web" {
  record = "example.com"
  type   = "A"
  value  = "192.0.2.1"
}

# Create a CNAME record
resource "dreamhost_dns_record" "www" {
  record = "www.example.com"
  type   = "CNAME"
  value  = "example.com"
}

# Query existing records
data "dreamhost_dns_records" "all" {
  filter {
    type = "A"
  }
}
```

## Documentation

### Architecture & Design

- [Architecture Overview](docs/ARCHITECTURE.md) - System design, components, and data flow

### Provider Resources

- [`dreamhost_dns_record`](docs/resources/dns_record.md) - Manages DNS records

### Provider Data Sources

- [`dreamhost_dns_record`](docs/data-sources/dns_record.md) - Look up a specific DNS record
- [`dreamhost_dns_records`](docs/data-sources/dns_records.md) - List and filter DNS records

### Examples

- [Basic DNS Management](examples/dns.tf)
- [Complete Configuration](examples/complete/)
- [Data Source Usage](examples/data-sources/)
- [Importing Records](examples/import/)

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    dreamhost = {
      source  = "aygp-dr/dreamhost"
      version = "~> 0.1.0"
    }
  }
}
```

### Building from Source

```bash
# Clone repository
git clone https://github.com/aygp-dr/terraform-provider-dreamhost.git
cd terraform-provider-dreamhost

# Build provider
go build -o terraform-provider-dreamhost

# Install locally
make install
```

## Development

### Prerequisites

- Go 1.19+
- Terraform 1.0+
- Make

### Build

```bash
make build
```

### Test

```bash
# Unit tests
make test

# Acceptance tests (requires DREAMHOST_API_KEY)
export DREAMHOST_API_KEY="your-api-key"
make testacc
```

### Lint

```bash
make lint
```

## Import Existing Resources

Import existing DNS records into Terraform:

```bash
# Format: TYPE|RECORD|VALUE
terraform import dreamhost_dns_record.example "A|example.com|192.0.2.1"
```

## Environment Variables

- `DREAMHOST_API_KEY` - DreamHost API key (recommended over provider configuration)

## Troubleshooting

### Common Issues

**Rate Limiting**
The provider includes automatic retry logic for rate limiting. If you encounter persistent issues, consider:
- Reducing parallel operations
- Adding delays between resource creation

**DNS Propagation**
DNS changes may take time to propagate. The provider waits for changes to be confirmed via the API.

**Import ID Format**
Import IDs must follow the format: `TYPE|RECORD|VALUE`

### Debug Mode

Enable debug output:
```bash
export TF_LOG=DEBUG
terraform apply
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Security

- Never commit API keys to version control
- Use environment variables for sensitive data
- Report security issues privately

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/aygp-dr/terraform-provider-dreamhost/issues)
- 💬 [Discussions](https://github.com/aygp-dr/terraform-provider-dreamhost/discussions)

## License

This provider is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgments

- [DreamHost](https://www.dreamhost.com/) for providing the API
- [HashiCorp](https://www.hashicorp.com/) for Terraform
- Contributors and users of this provider

---

**Note:** This provider is not officially affiliated with or endorsed by DreamHost.