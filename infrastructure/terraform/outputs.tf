# ---------------------------------------------------------------------------
# Database
# ---------------------------------------------------------------------------

output "database_endpoint" {
  description = "Aurora PostgreSQL cluster endpoint"
  value       = aws_rds_cluster.main.endpoint
}

output "database_reader_endpoint" {
  description = "Aurora PostgreSQL reader endpoint"
  value       = aws_rds_cluster.main.reader_endpoint
}

output "database_port" {
  description = "Database port"
  value       = aws_rds_cluster.main.port
}

output "database_name" {
  description = "Database name"
  value       = aws_rds_cluster.main.database_name
}

output "database_url" {
  description = "Full PostgreSQL connection string"
  value       = "postgresql://${var.db_master_username}:${var.db_master_password}@${aws_rds_cluster.main.endpoint}:${aws_rds_cluster.main.port}/${aws_rds_cluster.main.database_name}?sslmode=require"
  sensitive   = true
}

# ---------------------------------------------------------------------------
# Redis
# ---------------------------------------------------------------------------

output "redis_endpoint" {
  description = "Redis primary endpoint"
  value       = aws_elasticache_replication_group.main.primary_endpoint_address
}

output "redis_port" {
  description = "Redis port"
  value       = aws_elasticache_replication_group.main.port
}

output "redis_url" {
  description = "Full Redis connection string"
  value       = "rediss://:${var.redis_auth_token}@${aws_elasticache_replication_group.main.primary_endpoint_address}:${aws_elasticache_replication_group.main.port}"
  sensitive   = true
}

# ---------------------------------------------------------------------------
# Object Storage
# ---------------------------------------------------------------------------

output "uploads_bucket" {
  description = "R2 bucket name for user uploads"
  value       = cloudflare_r2_bucket.uploads.name
}

output "models_bucket" {
  description = "R2 bucket name for ML model artifacts"
  value       = cloudflare_r2_bucket.models.name
}

output "backups_bucket" {
  description = "R2 bucket name for database backups"
  value       = cloudflare_r2_bucket.backups.name
}

# ---------------------------------------------------------------------------
# Networking
# ---------------------------------------------------------------------------

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = aws_subnet.private[*].id
}

# ---------------------------------------------------------------------------
# Container Registry
# ---------------------------------------------------------------------------

output "ecr_backend_url" {
  description = "ECR repository URL for backend images"
  value       = aws_ecr_repository.backend.repository_url
}

output "ecr_web_url" {
  description = "ECR repository URL for web images"
  value       = aws_ecr_repository.web.repository_url
}

output "ecr_ai_url" {
  description = "ECR repository URL for AI pipeline images"
  value       = aws_ecr_repository.ai.repository_url
}
