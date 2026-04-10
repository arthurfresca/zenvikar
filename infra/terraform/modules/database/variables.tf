variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "zenvikar"
}

variable "instance_class" {
  description = "Database instance class/size"
  type        = string
  default     = "db.t3.micro"
}
