output "log_group_name" {
  description = "Name of the log group"
  value       = null_resource.observability_placeholder.id
}

output "dashboard_url" {
  description = "URL of the monitoring dashboard"
  value       = ""
}
