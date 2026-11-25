#!/usr/bin/env bash
# Corre DENTRO del contenedor g680-tools (lo lanza correr.py). Todo se trabaja en un directorio de mktemp;
# /demo (la carpeta de la demo) solo se lee, salvo /demo/salidas, donde se escriben los resultados.
set -uo pipefail
SAL=/demo/salidas
mkdir -p "$SAL"
WORK=$(mktemp -d /tmp/g680.XXXXXX)
MOTO=http://moto:5000
export TF_VAR_clave=s3cr3to-de-prod TF_IN_AUTOMATION=1
RONDAS=${RONDAS:-3}

seg() { date +%s.%N; }
# mide <archivo-de-log> <comando...>  → imprime los segundos con una decimal
mide() {
  local log=$1; shift
  local a b; a=$(seg); "$@" >"$log" 2>&1; local rc=$?; b=$(seg)
  awk -v a="$a" -v b="$b" -v rc="$rc" 'BEGIN{printf "%.1f%s", b-a, (rc==0||rc==2)?"":" (rc="rc")"}'
}
reset_moto() { curl -s -X POST "$MOTO/moto-api/reset" >/dev/null; }
# líneas de código: sin vacías ni comentarios
loc() { grep -v -E '^\s*($|//|#)' "$@" | wc -l; }

cp -r /demo/pulumi "$WORK/pu"; cp -r /demo/terraform "$WORK/tf"; cp -r /demo/trampa-count "$WORK/tc"; cp -r /demo/solo-escritura "$WORK/se"
rm -f "$WORK/pu/Pulumi.prod.yaml.bak"

# ───────── 0 · versiones ─────────
{
  go version
  echo "pulumi $(pulumi version)"
  terraform version | head -1
  grep -E 'pulumi/(pulumi/sdk|pulumi-aws/sdk)' "$WORK/pu/go.mod" | sed 's/^\s*/go.mod: /'
  echo "provider hashicorp/aws 6.70.0 (main.tf)"
  echo "moto: ${MOTO_VERSION:-?} (motoserver/moto:latest)"
} >"$SAL/00-versiones.txt"

# ───────── 1 · líneas ─────────
{
  echo "Pulumi (Go):   main.go $(loc "$WORK/pu/main.go") · Pulumi.yaml $(loc "$WORK/pu/Pulumi.yaml") · Pulumi.prod.yaml $(loc "$WORK/pu/Pulumi.prod.yaml")"
  echo "               total $(cat "$WORK/pu/main.go" "$WORK/pu/Pulumi.yaml" "$WORK/pu/Pulumi.prod.yaml" | grep -v -E '^\s*($|//|#)' | wc -l)"
  echo "Terraform:     main.tf $(loc "$WORK/tf/main.tf")"
  echo "Tests:         main_test.go $(loc "$WORK/pu/main_test.go") · basico.tftest.hcl $(loc "$WORK/tf/tests/basico.tftest.hcl")"
  echo "go.mod: $(wc -l <"$WORK/pu/go.mod") líneas, $(grep -c '^\s' "$WORK/pu/go.mod") dependencias · go.sum: $(wc -l <"$WORK/pu/go.sum") líneas"
} >"$SAL/01-lineas.txt"

# ───────── 2 · lo que se baja la primera vez ─────────
{
  export PULUMI_HOME="$WORK/pulumi-home-frio"
  echo "pulumi plugin install resource aws 7.51.1: $(mide "$WORK/plugin.log" pulumi plugin install resource aws 7.51.1) s"
  echo "  $(du -sh "$PULUMI_HOME/plugins" | cut -f1) en disco"
  unset PULUMI_HOME
  cd "$WORK/tf" || exit 1
  echo "terraform init (baja hashicorp/aws 6.70.0): $(mide "$WORK/init.log" terraform init -input=false -no-color) s"
  echo "  $(du -sh .terraform | cut -f1) en disco"
  cd "$WORK/pu" || exit 1
  echo "go build en frío (GOCACHE vacío): $(GOCACHE="$WORK/gocache-frio" mide "$WORK/build.log" go build -o "$WORK/prog" .) s"
  echo "go build en caliente:            $(mide "$WORK/build2.log" go build -o "$WORK/prog" .) s"
  echo "  binario del programa: $(stat -c %s "$WORK/prog") bytes"
} >"$SAL/02-primera-vez.txt"

# ───────── 3 · tiempos, $RONDAS rondas, moto limpio en cada una ─────────
cd "$WORK/pu" || exit 1
mkdir -p "$WORK/state"
export PULUMI_BACKEND_URL="file://$WORK/state"
pulumi login "file://$WORK/state" >"$WORK/login.log" 2>&1
pulumi stack init prod --non-interactive >/dev/null 2>&1
pulumi config set --secret clave s3cr3to-de-prod
echo "ronda  herramienta  preview/plan  up/apply  sin-cambios  destroy" >"$SAL/03-tiempos.txt"
for r in $(seq 1 "$RONDAS"); do
  reset_moto
  cd "$WORK/pu" || exit 1
  p1=$(mide "$WORK/pu-preview-$r.log" pulumi preview --non-interactive --color never)
  p2=$(mide "$WORK/pu-up-$r.log" pulumi up --yes --skip-preview --non-interactive --color never)
  p3=$(mide "$WORK/pu-preview2-$r.log" pulumi preview --non-interactive --color never --expect-no-changes)
  p4=$(mide "$WORK/pu-destroy-$r.log" pulumi destroy --yes --skip-preview --non-interactive --color never)
  echo "$r      pulumi       $p1          $p2      $p3        $p4" >>"$SAL/03-tiempos.txt"
  reset_moto
  cd "$WORK/tf" || exit 1
  t1=$(mide "$WORK/tf-plan-$r.log" terraform plan -input=false -no-color -out=plan.bin)
  t2=$(mide "$WORK/tf-apply-$r.log" terraform apply -input=false -no-color plan.bin)
  t3=$(mide "$WORK/tf-plan2-$r.log" terraform plan -input=false -no-color -detailed-exitcode)
  if [ "$r" -lt "$RONDAS" ]; then
    t4=$(mide "$WORK/tf-destroy-$r.log" terraform destroy -auto-approve -input=false -no-color)
  else
    t4="(se queda para el estado)"
  fi
  echo "$r      terraform    $t1          $t2      $t3        $t4" >>"$SAL/03-tiempos.txt"
done
cp "$WORK/pu-preview-1.log" "$SAL/03-pulumi-preview.txt"; cp "$WORK/pu-up-1.log" "$SAL/03-pulumi-up.txt"
cp "$WORK/pu-preview2-1.log" "$SAL/03-pulumi-sin-cambios.txt"
cp "$WORK/tf-plan-1.log" "$SAL/03-terraform-plan.txt"; cp "$WORK/tf-apply-1.log" "$SAL/03-terraform-apply.txt"
cp "$WORK/tf-plan2-1.log" "$SAL/03-terraform-sin-cambios.txt"

# Terraform quedó arriba en la última ronda: se mira su estado, se baja, y luego sube Pulumi para mirar el suyo.

# ───────── 4 · el estado ─────────
cd "$WORK/tf" || exit 1
{
  echo "== Terraform: terraform.tfstate"
  echo "recursos en el estado: $(terraform state list | wc -l)"
  echo "tamaño: $(stat -c %s terraform.tfstate) bytes"
  echo "¿la clave en texto plano?  grep -c s3cr3to-de-prod → $(grep -c s3cr3to-de-prod terraform.tfstate)"
  grep -n '"value": "s3cr3to' terraform.tfstate | head -2
} >"$SAL/04-estado.txt"
terraform destroy -auto-approve -input=false -no-color >"$WORK/tf-destroy-final.log" 2>&1
reset_moto
cd "$WORK/pu" || exit 1
pulumi up --yes --skip-preview --non-interactive --color never >"$WORK/pu-up-estado.log" 2>&1
pulumi stack export --file "$WORK/pu-estado.json"
{
  echo "== Pulumi: pulumi stack export (backend file://)"
  echo "recursos en el estado: $(jq '.deployment.resources|length' "$WORK/pu-estado.json") (incluye la pila y el provider)"
  echo "tamaño: $(stat -c %s "$WORK/pu-estado.json") bytes"
  echo "¿la clave en texto plano?  grep -c s3cr3to-de-prod → $(grep -c s3cr3to-de-prod "$WORK/pu-estado.json")"
  jq -c '.deployment.resources[]|select(.type=="aws:ssm/parameter:Parameter")|.outputs.value' "$WORK/pu-estado.json" | cut -c1-90
  echo "archivos del backend:"; (cd "$WORK/state" && find . -type f | sort | sed 's/^/  /')
} >>"$SAL/04-estado.txt"

# Terraform 1.11+: argumento de solo escritura (value_wo), que no se guarda en el estado
reset_moto
cd "$WORK/se" || exit 1
terraform init -input=false -no-color >/dev/null 2>&1
terraform apply -auto-approve -input=false -no-color >"$WORK/se-apply.log" 2>&1
{
  echo "== Terraform con value_wo (solo escritura)"
  echo "apply: $(grep -E '^Apply complete' "$WORK/se-apply.log")"
  echo "¿la clave en texto plano?  grep -c s3cr3to-de-prod → $(grep -c s3cr3to-de-prod terraform.tfstate)"
  echo "valor en moto: $(curl -s -X POST "$MOTO/" -H 'X-Amz-Target: AmazonSSM.GetParameter' -H 'Content-Type: application/x-amz-json-1.1' -H 'Authorization: AWS4-HMAC-SHA256 Credential=test/20261026/us-east-1/ssm/aws4_request, SignedHeaders=host, Signature=x' -d '{"Name":"/g680/prod/clave-wo","WithDecryption":true}' | jq -r .Parameter.Value)"
} >>"$SAL/04-estado.txt"
terraform destroy -auto-approve -input=false -no-color >/dev/null 2>&1

# ───────── 5 · el bucle: quitar «pedidos» ─────────
cd "$WORK/pu" || exit 1
pulumi config rm colas && pulumi config set --path 'colas[0]' pagos && pulumi config set --path 'colas[1]' avisos
pulumi preview --non-interactive --color never --diff >"$SAL/05-pulumi-quitar-una.txt" 2>&1
pulumi config rm colas && pulumi config set --path 'colas[0]' pedidos && pulumi config set --path 'colas[1]' pagos && pulumi config set --path 'colas[2]' avisos
pulumi destroy --yes --skip-preview --non-interactive --color never >"$WORK/pu-destroy-final.log" 2>&1
reset_moto
cd "$WORK/tc" || exit 1
terraform init -input=false -no-color >/dev/null 2>&1
terraform apply -auto-approve -input=false -no-color >"$WORK/tc-apply.log" 2>&1
terraform plan -input=false -no-color -var 'colas=["pagos","avisos"]' >"$SAL/05-terraform-quitar-una.txt" 2>&1
terraform destroy -auto-approve -input=false -no-color >/dev/null 2>&1

# ───────── 6 · tests ─────────
cd "$WORK/pu" || exit 1
{
  echo "\$ go test -count=1 -v -run 'Prod|Dev|Politica' ."
  echo "(compilación de los tests incluida: $(mide "$WORK/gotest1.log" go test -count=1 -run 'Prod|Dev|Politica' .) s)"
  echo "(otra vez, con caché: $(mide "$WORK/gotest2.log" go test -count=1 -v -run 'Prod|Dev|Politica' .) s)"
  cat "$WORK/gotest2.log"
} >"$SAL/06-go-test.txt"
cd "$WORK/tf" || exit 1
{
  echo "\$ terraform test"
  echo "($(mide "$WORK/tftest.log" terraform test -no-color) s)"
  cat "$WORK/tftest.log"
} >"$SAL/06-terraform-test.txt"

# ───────── 7 · la trampa de Pulumi ─────────
cd "$WORK/pu" || exit 1
go test -count=1 -v -run Trampa . >"$SAL/07-trampa-output.txt" 2>&1

reset_moto
cd / && rm -rf "${WORK:?}"
echo "listo: $SAL"
