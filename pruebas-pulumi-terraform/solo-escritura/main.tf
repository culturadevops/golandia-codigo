terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.70.0"
    }
  }
}

provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  endpoints {
    ssm = "http://moto:5000"
    sts = "http://moto:5000"
  }
}

variable "clave" {
  type      = string
  sensitive = true
  ephemeral = true
}

# value_wo: el valor llega a AWS pero no se guarda en el estado (Terraform 1.11+)
resource "aws_ssm_parameter" "clave" {
  name             = "/g680/prod/clave-wo"
  type             = "SecureString"
  value_wo         = var.clave
  value_wo_version = 1
}
