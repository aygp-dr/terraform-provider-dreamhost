# DreamHost Provider Examples

This directory contains examples demonstrating how to use the DreamHost Terraform Provider.

## Examples

### Basic
Simple example showing basic DNS record management.

```bash
cd basic/
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
terraform init
terraform plan
terraform apply
```

### Complete
Comprehensive example with all DNS record types and best practices.

```bash
cd complete/
export DREAMHOST_API_KEY="your-api-key"
terraform init
terraform plan
terraform apply
```

### Data Sources
Examples of using data sources to query existing DNS records.

```bash
cd data-sources/
export DREAMHOST_API_KEY="your-api-key"
terraform init
terraform apply
```

### Import
Example showing how to import existing DNS records.

```bash
cd import/
export DREAMHOST_API_KEY="your-api-key"
terraform init
# Import existing records
terraform import dreamhost_dns_record.existing "A|example.com|192.0.2.1"
```

## Best Practices

1. **Never commit API keys** - Use environment variables or terraform.tfvars (git-ignored)
2. **Use variables** - Define reusable variables for domain names and IP addresses
3. **Add validation** - Use variable validation blocks to ensure correct input
4. **Document outputs** - Always add descriptions to outputs
5. **Organize files** - Split configuration into:
   - `versions.tf` - Terraform and provider requirements
   - `variables.tf` - Variable definitions
   - `main.tf` - Resource definitions
   - `outputs.tf` - Output definitions
6. **Use for_each** - For multiple similar resources, use for_each instead of count
7. **Add comments** - Document complex configurations

## Terraform Linting

All examples follow Terraform best practices and are formatted with:

```bash
terraform fmt -recursive .
```

You can validate the configuration with:

```bash
terraform validate
```

For additional linting, install TFLint:

```bash
curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
tflint
```

## Variable Validation Examples

The examples demonstrate proper variable validation:

```hcl
variable "domain_name" {
  description = "The domain name to manage DNS records for"
  type        = string
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9][a-zA-Z0-9-_.]+[a-zA-Z0-9]$", var.domain_name))
    error_message = "Domain name must be a valid DNS hostname."
  }
}

variable "ipv4_address" {
  description = "IPv4 address for A record"
  type        = string
  
  validation {
    condition     = can(cidrhost("${var.ipv4_address}/32", 0))
    error_message = "Must be a valid IPv4 address."
  }
}
```

## Dynamic Resource Creation

Use for_each for creating multiple similar resources:

```hcl
variable "dns_records" {
  type = map(object({
    type  = string
    value = string
  }))
}

resource "dreamhost_dns_record" "dynamic" {
  for_each = var.dns_records
  
  record = each.key
  type   = each.value.type
  value  = each.value.value
}
```