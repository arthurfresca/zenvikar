variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "prod"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "zenvikar"
}

variable "domain_name" {
  description = "Base domain name for the platform"
  type        = string
  default     = "zenvikar.com"
}
