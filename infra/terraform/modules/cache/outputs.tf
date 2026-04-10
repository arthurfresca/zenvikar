output "endpoint" {
  description = "Redis connection endpoint"
  value       = null_resource.cache_placeholder.id
}

output "port" {
  description = "Redis connection port"
  value       = 6379
}
