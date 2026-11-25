#!/bin/sh
# Corre DENTRO del contenedor (lo lanza correr.py). Uso: escenarios.sh <escenario>
#   versiones | ciclo | drift | import | trampa-clave | trampa-404 | trampa-lista | trampa-vacio
# Todo pasa en un directorio de mktemp; /demo (el código) se monta de solo lectura.
set -u
ESC="$1"
API=http://127.0.0.1:8080
export TF_IN_AUTOMATION=1 TF_CLI_ARGS=-no-color NOTAS_ENDPOINT=$API

compilar() { # compilar <variante> <trampas>
  mkdir -p /build/$1
  (cd /demo/terraform-provider-notas && go build -trimpath -ldflags "-s -w -X main.trampas=$2" -o /build/$1/terraform-provider-notas .)
}

# .terraformrc con dev_overrides: Terraform usa el binario de esa carpeta y no lo busca en el registro.
terraformrc() {
  cat > "$TRABAJO/.terraformrc" <<EOF
provider_installation {
  dev_overrides {
    "golandia/notas" = "/build/$1"
  }
  direct {}
}
EOF
  export TF_CLI_CONFIG_FILE="$TRABAJO/.terraformrc"
}

api() {
  [ -n "${APIPID:-}" ] && kill "$APIPID" 2>/dev/null
  /build/notas-api > /tmp/api.log 2>&1 &
  APIPID=$!
  sleep 0.5
  MARCA=1 # la primera línea es «notas-api en :8080»
}

# Las llamadas que recibió la API desde la última vez.
llamadas() {
  echo "--- llamadas a notas-api ---"
  tail -n +$((MARCA + 1)) /tmp/api.log
  MARCA=$(wc -l < /tmp/api.log)
}

paso() {
  echo
  echo "\$ $*"
  t0=$(date +%s.%N)
  "$@" 2>&1
  rc=$?
  t1=$(date +%s.%N)
  echo "[código de salida $rc · $(awk "BEGIN{printf \"%.2f\", $t1-$t0}") s]"
}

a_mano() { # a_mano <método> <ruta> [json]
  echo
  echo "\$ curl -X $1 $API$2 ${3:+-d '$3'}"
  if [ -n "${3:-}" ]; then curl -s -X "$1" -d "$3" "$API$2"; else curl -s -o /dev/null -w "HTTP %{http_code}\n" -X "$1" "$API$2"; fi
  MARCA=$(wc -l < /tmp/api.log)
}

# El id que la API le dio a un recurso (las dos notas se crean en paralelo: el orden cambia).
id_de() { terraform state show "notas_nota.$1" | awk -F'"' '$1 ~ /^ *id *= *$/ {print $2}'; }

config_dos() {
  cat > "$TRABAJO/main.tf" <<'EOF'
terraform {
  required_providers {
    notas = { source = "golandia/notas" }
  }
}

provider "notas" {}

resource "notas_nota" "compras" {
  titulo    = "Compras"
  texto     = "pan, café"
  etiquetas = ["zeta", "casa"]
  clave     = "secreta"
}

resource "notas_nota" "ideas" {
  titulo = "Ideas para vídeos"
}
EOF
}

TRABAJO=$(mktemp -d)
cd "$TRABAJO" || exit 1
mkdir -p /build
(cd /demo/notas-api && go build -trimpath -o /build/notas-api .)

case "$ESC" in
versiones)
  compilar bien ""
  echo "\$ go version"; go version
  echo "\$ terraform version"; terraform version | head -2
  echo "\$ go list -m (framework)"; (cd /demo/terraform-provider-notas && go list -m github.com/hashicorp/terraform-plugin-framework github.com/hashicorp/terraform-plugin-go)
  echo "\$ go list -m all | wc -l"; (cd /demo/terraform-provider-notas && go list -m all | wc -l)
  echo "\$ go list -deps . | wc -l"; (cd /demo/terraform-provider-notas && go list -deps . | wc -l)
  echo "\$ wc -l *.go (provider)"; (cd /demo/terraform-provider-notas && wc -l *.go)
  echo "\$ wc -l main.go (api)"; (cd /demo/notas-api && wc -l main.go)
  echo "\$ ls -l (binarios, -s -w)"; ls -l /build/bien/terraform-provider-notas /build/notas-api | awk '{print $5, $9}'
  (cd /demo/terraform-provider-notas && go build -trimpath -o /tmp/sin-sw . && ls -l /tmp/sin-sw | awk '{print $5, "sin -s -w"}')
  ;;
ciclo)
  compilar bien ""; terraformrc bien; api; config_dos
  paso terraform init
  paso terraform plan
  llamadas
  paso terraform apply -auto-approve
  llamadas
  paso terraform state list
  paso terraform state show notas_nota.compras
  echo; echo "\$ terraform.tfstate (recurso compras)"
  sed -n '/"name": "compras"/,/sensitive_attributes/p' terraform.tfstate
  paso terraform plan
  llamadas
  sed -i 's/titulo    = "Compras"/titulo    = "Compras del sábado"/' main.tf
  paso terraform plan
  llamadas
  paso terraform apply -auto-approve
  llamadas
  paso terraform destroy -auto-approve
  llamadas
  ;;
drift)
  compilar bien ""; terraformrc bien; api; config_dos
  terraform apply -auto-approve > /dev/null 2>&1 && echo "(apply inicial: 2 notas creadas)"
  MARCA=$(wc -l < /tmp/api.log)
  C=$(id_de compras); I=$(id_de ideas)
  a_mano PUT /notas/$C '{"titulo":"Cambiado a mano","texto":"pan, café","etiquetas":["casa","zeta"]}'
  paso terraform plan
  llamadas
  paso terraform plan -refresh-only
  paso terraform apply -auto-approve
  llamadas
  a_mano DELETE /notas/$I
  paso terraform plan
  llamadas
  ;;
import)
  compilar bien ""; terraformrc bien; api
  a_mano POST /notas '{"titulo":"Hecha a mano","texto":"antes de Terraform","etiquetas":["vieja"],"clave":"secreta"}'
  cat > main.tf <<'EOF'
terraform {
  required_providers {
    notas = { source = "golandia/notas" }
  }
}

provider "notas" {}

import {
  to = notas_nota.vieja
  id = "1"
}

resource "notas_nota" "vieja" {
  titulo    = "Hecha a mano"
  texto     = "antes de Terraform"
  etiquetas = ["vieja"]
  clave     = "secreta"
}
EOF
  echo; echo "\$ cat main.tf"; cat main.tf
  paso terraform plan
  llamadas
  paso terraform apply -auto-approve
  llamadas
  paso terraform plan
  echo; echo "== y con el comando terraform import (sin bloque import) =="
  rm -f terraform.tfstate terraform.tfstate.backup
  sed -i '/^import {/,/^}/d' main.tf
  paso terraform import notas_nota.vieja 1
  llamadas
  paso terraform state show notas_nota.vieja
  echo; echo "== importar un id que no existe =="
  paso terraform state rm notas_nota.vieja
  paso terraform import notas_nota.vieja 99
  llamadas
  ;;
trampa-clave)
  compilar trampa clave; terraformrc trampa; api; config_dos
  paso terraform apply -auto-approve
  llamadas
  for i in 1 2 3; do
    echo; echo "== plan número $i =="
    paso terraform plan
    paso terraform apply -auto-approve
    llamadas
  done
  ;;
trampa-404)
  compilar trampa 404; terraformrc trampa; api; config_dos
  terraform apply -auto-approve > /dev/null 2>&1 && echo "(apply inicial: 2 notas creadas)"
  MARCA=$(wc -l < /tmp/api.log)
  a_mano DELETE /notas/$(id_de compras)
  paso terraform plan
  llamadas
  paso terraform destroy -auto-approve
  echo; echo "== la salida: sacarla del estado a mano =="
  paso terraform state rm notas_nota.compras
  paso terraform plan
  ;;
trampa-lista)
  compilar trampa lista; terraformrc trampa; api; config_dos
  paso terraform apply -auto-approve
  llamadas
  paso terraform state list
  paso terraform plan
  ;;
trampa-vacio)
  compilar trampa vacio; terraformrc trampa; api; config_dos
  paso terraform apply -auto-approve
  llamadas
  paso terraform state list
  paso terraform plan
  ;;
*)
  echo "escenario desconocido: $ESC"; exit 2 ;;
esac
[ -n "${APIPID:-}" ] && kill "$APIPID" 2>/dev/null
exit 0
