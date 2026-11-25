terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.70.0"
    }
  }
}

# moto, el emulador de AWS, en la red de Docker de la demo
provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  s3_use_path_style           = true
  endpoints {
    s3  = "http://moto:5000"
    sqs = "http://moto:5000"
    iam = "http://moto:5000"
    ssm = "http://moto:5000"
    sts = "http://moto:5000"
  }
}

variable "entorno" {
  type    = string
  default = "prod"
}

variable "colas" {
  type    = list(string)
  default = ["pedidos", "pagos", "avisos"]
}

variable "clave" {
  type      = string
  sensitive = true
}

resource "aws_s3_bucket" "datos" {
  bucket = "g680-datos-${var.entorno}"
}

# La condición: el versionado solo en prod
resource "aws_s3_bucket_versioning" "datos" {
  count  = var.entorno == "prod" ? 1 : 0
  bucket = aws_s3_bucket.datos.id
  versioning_configuration {
    status = "Enabled"
  }
}

# El bucle: una cola por nombre
resource "aws_sqs_queue" "cola" {
  for_each = toset(var.colas)
  name     = "g680-${each.key}"
}

resource "aws_iam_policy" "app" {
  name = "g680-app-${var.entorno}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["sqs:SendMessage"]
        Resource = [for c in aws_sqs_queue.cola : c.arn]
      },
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject", "s3:PutObject"]
        Resource = ["${aws_s3_bucket.datos.arn}/*"]
      },
    ]
  })
}

resource "aws_ssm_parameter" "clave" {
  name  = "/g680/${var.entorno}/clave"
  type  = "SecureString"
  value = var.clave
}

output "bucket" {
  value = aws_s3_bucket.datos.bucket
}

output "colas" {
  value = { for k, c in aws_sqs_queue.cola : k => c.url }
}

output "politica" {
  value = aws_iam_policy.app.arn
}
