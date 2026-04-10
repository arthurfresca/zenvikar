output "endpoint" {
  description = "Database connection endpoint"
  value       = null_resource.database_placeholder.id
}

output "database_name" {
  description = "Name of the provisioned database"
  value       = var.project_name
}
