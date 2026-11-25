# Go 1.27.2 + Terraform 1.16.2 en la misma imagen: el provider y la API se compilan aquí y Terraform los usa.
FROM hashicorp/terraform:1.16.2 AS tf
FROM golang:1.27.2
COPY --from=tf /bin/terraform /usr/local/bin/terraform
