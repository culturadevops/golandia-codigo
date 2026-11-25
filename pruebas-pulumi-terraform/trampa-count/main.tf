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
    sqs = "http://moto:5000"
    sts = "http://moto:5000"
  }
}

variable "colas" {
  type    = list(string)
  default = ["pedidos", "pagos", "avisos"]
}

# La trampa: count numera por posición
resource "aws_sqs_queue" "por_posicion" {
  count = length(var.colas)
  name  = "g680-c-${var.colas[count.index]}"
}

# Lo mismo con for_each: numera por nombre
resource "aws_sqs_queue" "por_nombre" {
  for_each = toset(var.colas)
  name     = "g680-f-${each.key}"
}
