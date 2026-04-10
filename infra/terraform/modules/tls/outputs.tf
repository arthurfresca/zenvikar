output "certificate_arn" {
  description = "ARN of the provisioned TLS certificate"
  value       = null_resource.tls_placeholder.id
}
