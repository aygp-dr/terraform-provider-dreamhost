terraform {
  required_version = ">= 1.0"
  required_providers {
    dreamhost = {
      source  = "aygp-dr/dreamhost"
      version = "~> 0.1.0"
    }
  }
}

# Configure the DreamHost Provider
# API key can be set via DREAMHOST_API_KEY environment variable
provider "dreamhost" {
  # api_key = var.dreamhost_api_key
}
