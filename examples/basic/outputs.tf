output "dns_record_id" {
  description = "The ID of the created DNS record"
  value       = dreamhost_dns_record.main.id
}

output "dns_record_value" {
  description = "The value of the created DNS record"
  value       = dreamhost_dns_record.main.value
}

output "dns_record_type" {
  description = "The type of the created DNS record"
  value       = dreamhost_dns_record.main.type
}