# Zenvikar Platform — Dev Environment
# Composes all infrastructure modules for the development environment.

terraform {
  required_version = ">= 1.5.0"
}

locals {
  environment  = var.environment
  project_name = var.project_name
}

module "networking" {
  source       = "../../modules/networking"
  environment  = local.environment
  project_name = local.project_name
}

module "dns" {
  source       = "../../modules/dns"
  environment  = local.environment
  project_name = local.project_name
  domain_name  = var.domain_name
}

module "tls" {
  source       = "../../modules/tls"
  environment  = local.environment
  project_name = local.project_name
  domain_name  = var.domain_name
}

module "database" {
  source       = "../../modules/database"
  environment  = local.environment
  project_name = local.project_name
}

module "cache" {
  source       = "../../modules/cache"
  environment  = local.environment
  project_name = local.project_name
}

module "secrets" {
  source       = "../../modules/secrets"
  environment  = local.environment
  project_name = local.project_name
}

module "app_hosting" {
  source       = "../../modules/app-hosting"
  environment  = local.environment
  project_name = local.project_name
}

module "observability" {
  source       = "../../modules/observability"
  environment  = local.environment
  project_name = local.project_name
}
