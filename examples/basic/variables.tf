variable "dreamhost_api_key" {
  description = "API key for DreamHost provider. Can also be set via DREAMHOST_API_KEY environment variable."
  type        = string
  sensitive   = true
  default     = ""
}

variable "domain_name" {
  description = "The domain name to manage DNS records for"
  type        = string
  default     = "example.com"

  validation {
    condition     = can(regex("^[a-zA-Z0-9][a-zA-Z0-9-_.]+[a-zA-Z0-9]$", var.domain_name))
    error_message = "Domain name must be a valid DNS hostname."
  }
}

variable "ip_address" {
  description = "The IP address for the A record"
  type        = string
  default     = "192.0.2.1"

  validation {
    condition     = can(regex("^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$", var.ip_address))
    error_message = "Must be a valid IPv4 address."
  }
}