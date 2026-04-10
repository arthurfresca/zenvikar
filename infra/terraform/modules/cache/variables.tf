variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "zenvikar"
}

variable "node_type" {
  description = "Redis node type/size"
  type        = string
  default     = "cache.t3.micro"
}
