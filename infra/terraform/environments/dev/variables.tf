variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "zenvikar"
}

variable "domain_name" {
  description = "Base domain name for the platform"
  type        = string
  default     = "dev.zenvikar.com"
}
