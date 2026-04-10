output "vpc_id" {
  description = "ID of the provisioned VPC"
  value       = null_resource.networking_placeholder.id
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = []
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = []
}
