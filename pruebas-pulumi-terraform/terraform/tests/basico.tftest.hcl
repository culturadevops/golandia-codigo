# terraform test con un provider falso: no habla con AWS ni con moto
mock_provider "aws" {}

variables {
  clave = "no-importa"
}

run "prod_lleva_versionado" {
  command = apply
  assert {
    condition     = length(aws_s3_bucket_versioning.datos) == 1
    error_message = "prod debe llevar versionado"
  }
  assert {
    condition     = length(aws_sqs_queue.cola) == 3
    error_message = "deben salir 3 colas"
  }
}

run "dev_sin_versionado" {
  command = apply
  variables {
    entorno = "dev"
  }
  assert {
    condition     = length(aws_s3_bucket_versioning.datos) == 0
    error_message = "dev no lleva versionado"
  }
}

run "la_politica_nombra_cada_cola" {
  command = apply
  assert {
    condition     = length(jsondecode(aws_iam_policy.app.policy).Statement[0].Resource) == 3
    error_message = "la política debe nombrar las 3 colas"
  }
}
