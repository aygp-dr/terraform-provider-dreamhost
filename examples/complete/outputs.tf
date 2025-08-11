output "root_domain_a_record" {
  description = "The IPv4 address of the root domain"
  value       = dreamhost_dns_record.root_a.value
}

output "www_cname_record" {
  description = "The CNAME target for the www subdomain"
  value       = dreamhost_dns_record.www_cname.value
}

output "mx_records" {
  description = "All MX records for the domain"
  value = {
    for k, v in dreamhost_dns_record.mx : k => {
      priority = split(" ", v.value)[0]
      server   = split(" ", v.value)[1]
    }
  }
}

output "spf_record" {
  description = "The SPF record for email authentication"
  value       = dreamhost_dns_record.spf.value
}

output "ipv6_address" {
  description = "The IPv6 address of the root domain"
  value       = dreamhost_dns_record.ipv6.value
}

output "srv_records" {
  description = "All SRV records for services"
  value = {
    for k, v in dreamhost_dns_record.srv : k => v.value
  }
}