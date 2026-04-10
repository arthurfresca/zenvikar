variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "zenvikar"
}

variable "domain_name" {
  description = "Domain name for the TLS certificate"
  type        = string
  default     = "zenvikar.com"
}
