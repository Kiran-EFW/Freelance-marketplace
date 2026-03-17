# ---------------------------------------------------------------------------
# General
# ---------------------------------------------------------------------------

variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "aws_region" {
  description = "AWS region for infrastructure deployment"
  type        = string
  default     = "ap-south-1" # Mumbai - closest to primary market
}

variable "availability_zones" {
  description = "List of availability zones to use"
  type        = list(string)
  default     = ["ap-south-1a", "ap-south-1b"]
}

# ---------------------------------------------------------------------------
# Networking
# ---------------------------------------------------------------------------

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# ---------------------------------------------------------------------------
# Database (RDS Aurora PostgreSQL)
# ---------------------------------------------------------------------------

variable "postgres_version" {
  description = "Aurora PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

variable "db_master_username" {
  description = "Master username for the PostgreSQL database"
  type        = string
  default     = "seva_admin"
  sensitive   = true
}

variable "db_master_password" {
  description = "Master password for the PostgreSQL database"
  type        = string
  sensitive   = true
}

variable "db_min_capacity" {
  description = "Minimum Aurora Serverless v2 ACU capacity"
  type        = number
  default     = 0.5
}

variable "db_max_capacity" {
  description = "Maximum Aurora Serverless v2 ACU capacity"
  type        = number
  default     = 4.0
}

variable "db_instance_count" {
  description = "Number of database instances (1 for dev, 2+ for prod)"
  type        = number
  default     = 1
}

# ---------------------------------------------------------------------------
# Redis (ElastiCache)
# ---------------------------------------------------------------------------

variable "redis_node_type" {
  description = "ElastiCache node type for Redis"
  type        = string
  default     = "cache.t4g.micro"
}

variable "redis_auth_token" {
  description = "Auth token for Redis (must be 16+ characters)"
  type        = string
  sensitive   = true
}

# ---------------------------------------------------------------------------
# Cloudflare R2 (Object Storage)
# ---------------------------------------------------------------------------

variable "cloudflare_api_token" {
  description = "Cloudflare API token for R2 management"
  type        = string
  sensitive   = true
}

variable "cloudflare_account_id" {
  description = "Cloudflare account ID"
  type        = string
}

variable "r2_location" {
  description = "R2 bucket location hint (apac, weur, enam)"
  type        = string
  default     = "apac"
}

# ---------------------------------------------------------------------------
# Domain
# ---------------------------------------------------------------------------

variable "domain_name" {
  description = "Primary domain name for the Seva platform"
  type        = string
  default     = "seva.app"
}

variable "api_subdomain" {
  description = "Subdomain for the API"
  type        = string
  default     = "api"
}

# ---------------------------------------------------------------------------
# Application
# ---------------------------------------------------------------------------

variable "jwt_secret" {
  description = "Secret key for JWT token signing"
  type        = string
  sensitive   = true
}

variable "otp_secret" {
  description = "Secret key for OTP generation"
  type        = string
  sensitive   = true
}

variable "anthropic_api_key" {
  description = "API key for Anthropic Claude"
  type        = string
  sensitive   = true
}
