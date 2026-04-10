output "service_url" {
  description = "URL of the deployed container service"
  value       = null_resource.app_hosting_placeholder.id
}
