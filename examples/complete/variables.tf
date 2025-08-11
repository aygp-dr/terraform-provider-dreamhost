variable "domain_name" {
  description = "The domain name to manage DNS records for"
  type        = string
  default     = "example.com"

  validation {
    condition     = can(regex("^[a-zA-Z0-9][a-zA-Z0-9-_.]+[a-zA-Z0-9]$", var.domain_name))
    error_message = "Domain name must be a valid DNS hostname."
  }
}

variable "subdomain" {
  description = "Subdomain prefix for www record"
  type        = string
  default     = "www"

  validation {
    condition     = can(regex("^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$", var.subdomain))
    error_message = "Subdomain must be a valid DNS label."
  }
}

variable "mail_servers" {
  description = "List of mail servers with priorities"
  type = list(object({
    priority = number
    server   = string
  }))
  default = [
    {
      priority = 10
      server   = "mail.example.com"
    },
    {
      priority = 20
      server   = "mail2.example.com"
    }
  ]
}

variable "spf_record" {
  description = "SPF record value for email authentication"
  type        = string
  default     = "v=spf1 include:_spf.dreamhost.com ~all"
}

variable "verification_token" {
  description = "Domain verification token"
  type        = string
  default     = "example-verification-token-replace-me"
  sensitive   = true
}

variable "ipv4_address" {
  description = "IPv4 address for A record"
  type        = string
  default     = "192.0.2.1"

  validation {
    condition     = can(cidrhost("${var.ipv4_address}/32", 0))
    error_message = "Must be a valid IPv4 address."
  }
}

variable "ipv6_address" {
  description = "IPv6 address for AAAA record"
  type        = string
  default     = "2001:db8::1"

  validation {
    condition     = can(cidrhost("${var.ipv6_address}/128", 0))
    error_message = "Must be a valid IPv6 address."
  }
}

variable "srv_records" {
  description = "SRV records for services"
  type = map(object({
    priority = number
    weight   = number
    port     = number
    target   = string
  }))
  default = {
    sip_tcp = {
      priority = 10
      weight   = 60
      port     = 5060
      target   = "sipserver.example.com"
    }
  }
}