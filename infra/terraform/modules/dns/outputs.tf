output "zone_id" {
  description = "ID of the DNS zone"
  value       = null_resource.dns_placeholder.id
}

output "name_servers" {
  description = "Name servers for the DNS zone"
  value       = []
}
