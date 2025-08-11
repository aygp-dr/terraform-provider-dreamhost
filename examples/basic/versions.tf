terraform {
  required_version = ">= 1.0"

  required_providers {
    dreamhost = {
      source  = "aygp-dr/dreamhost"
      version = "~> 0.1.0"
    }
  }
}