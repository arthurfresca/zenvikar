variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "zenvikar"
}

variable "container_image" {
  description = "Container image URI for the application"
  type        = string
  default     = ""
}
