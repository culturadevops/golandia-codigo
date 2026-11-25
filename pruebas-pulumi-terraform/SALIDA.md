# SALIDA — Pulumi con Go vs Terraform (golandia 680, 2026-10-26)

Ejecutado el 2026-10-26 con `python correr.py` (salidas/*.txt, sin editar). Todo corre en el mismo contenedor Linux
(imagen g680-iac, ver Dockerfile) contra **moto** en la red de Docker `g680`:

```
go version go1.27.2 linux/amd64
pulumi v3.272.0
Terraform v1.16.2
go.mod: github.com/pulumi/pulumi-aws/sdk/v7 v7.51.1
go.mod: github.com/pulumi/pulumi/sdk/v3 v3.272.0
provider hashicorp/aws 6.70.0 (main.tf)
moto: 5.2.4.dev (motoserver/moto:latest)
```

Docker 29.6.2 en Windows 11 (24 CPU). Pulumi con backend local (`file://` en un mktemp) y `PULUMI_CONFIG_PASSPHRASE`; sin cuenta.
La infraestructura, igual en los dos: 1 bucket S3, su versionado **solo si entorno = prod**, **una cola SQS por nombre**
(pedidos, pagos, avisos), 1 política IAM que nombra los ARN de las colas y del bucket, y 1 parámetro SSM SecureString con una clave.

## 1 · Líneas de código (sin vacías ni comentarios)

```
Pulumi (Go):   main.go 90 · Pulumi.yaml 2 · Pulumi.prod.yaml 16
               total 108
Terraform:     main.tf 82
Tests:         main_test.go 63 · basico.tftest.hcl 32
go.mod: 121 líneas, 112 dependencias · go.sum: 342 líneas
```
La configuración del provider (endpoints de moto, credenciales falsas) va en Pulumi.prod.yaml en Pulumi y en el bloque `provider` en main.tf.

## 2 · Lo que se baja y compila la primera vez

```
pulumi plugin install resource aws 7.51.1: 15.1 s
  902M en disco
terraform init (baja hashicorp/aws 6.70.0): 15.0 s
  795M en disco
go build en frío (GOCACHE vacío): 19.8 s
go build en caliente:            0.1 s
  binario del programa: 57096221 bytes
```
(En una pasada anterior el go build en frío dio 26,2 s y 28,6 s: depende de la red y del disco. El orden de magnitud se mantiene.)

## 3 · Tiempos (segundos; 3 rondas, moto vacío en cada una)

```
ronda  herramienta  preview/plan  up/apply  sin-cambios  destroy
1      pulumi       16.9          27.7      2.0        42.7
1      terraform    3.3          27.3      4.6        48.8
2      pulumi       1.7          26.7      1.7        42.5
2      terraform    3.1          27.1      4.3        48.4
3      pulumi       1.7          26.7      1.9        42.6
3      terraform    3.3          27.2      4.5        (se queda para el estado)
```
- El 16,9 s de la ronda 1 de Pulumi incluye **bajar e instalar el plugin aws 7.51.1** (sale en el log: «Downloading plugin aws-7.51.1»)
  y compilar el programa. En caliente: 1,7 s.
- «sin-cambios» = un segundo preview/plan sin tocar nada: Pulumi «8 unchanged»; Terraform «No changes. Your infrastructure matches the configuration.»
- up/apply empatan: el tiempo se lo comen **las 3 colas SQS, 25-26 s en los dos** (el resto, menos de 1 s cada uno):

```
pulumi up:        aws:sqs:Queue cola-avisos created (25s) · aws:s3:Bucket datos created (0.82s) · aws:iam:Policy app created (0.17s)
                  Resources: + 8 created · Duration: 28s
terraform apply:  aws_sqs_queue.cola["avisos"]: Creation complete after 26s · aws_s3_bucket.datos: Creation complete after 1s
                  aws_iam_policy.app: Creation complete after 0s
                  Apply complete! Resources: 7 added, 0 changed, 0 destroyed.
```
  Pulumi cuenta 8 porque suma la pila (pulumi:pulumi:Stack). El provider aws de Pulumi está construido sobre el provider de
  Terraform (github.com/pulumi/pulumi-aws trae hashicorp/terraform-provider-aws como submódulo `upstream`; en una prueba
  previa, un error de Pulumi salió con la traza «sdk-v2/provider2.go … sdk.helper_schema», el SDK de providers de Terraform), así que la espera de las colas es el mismo código.
  ⚠️ Esto es moto, no AWS: los segundos absolutos de AWS real serán otros. Lo que se compara es una herramienta contra la otra.

Preview/plan del primer paso:
```
Pulumi:     Resources: + 8 to create
Terraform:  Plan: 7 to add, 0 to change, 0 to destroy.
```

## 4 · El estado

```
== Terraform: terraform.tfstate
recursos en el estado: 7
tamaño: 11569 bytes
¿la clave en texto plano?  grep -c s3cr3to-de-prod → 1
308:            "value": "s3cr3to-de-prod",
== Pulumi: pulumi stack export (backend file://)
recursos en el estado: 9 (incluye la pila y el provider)
tamaño: 36804 bytes
¿la clave en texto plano?  grep -c s3cr3to-de-prod → 0
{"4dabf18193072939515e22adb298388d":"1b47061264138c4ac30d75fd1eb44270","ciphertext":"v1:Gb
== Terraform con value_wo (solo escritura)
apply: Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
¿la clave en texto plano?  grep -c s3cr3to-de-prod → 0
valor en moto: s3cr3to-de-prod
```
- En Terraform la variable es `sensitive = true` y aun así el valor está en el tfstate, en texto plano (la doc de HashiCorp lo
  advierte: «sensitive» solo lo oculta en la salida).
- En Pulumi, `cfg.RequireSecret` → el valor va cifrado con la passphrase en el estado y en Pulumi.prod.yaml (`secure: v1:…`).
- Terraform 1.11+ tiene argumentos de solo escritura: con `value_wo` + variable `ephemeral = true` (carpeta solo-escritura/),
  el valor llega a moto y el tfstate no lo guarda.
- El backend file:// de Pulumi, tras 7 operaciones (las de las 3 rondas + la subida final), tiene 48 archivos: 24 de datos
  (stacks/, history/, backups/, meta.yaml) y 24 .attrs.

## 5 · El bucle: quitar «pedidos» de la lista

Terraform, trampa-count/: la misma lista con `count` (por posición) y con `for_each` (por nombre), plan con `colas=["pagos","avisos"]`:
```
# aws_sqs_queue.por_nombre["pedidos"] will be destroyed
# aws_sqs_queue.por_posicion[0] must be replaced
~ name = "g680-c-pedidos" -> "g680-c-pagos" # forces replacement
# aws_sqs_queue.por_posicion[1] must be replaced
~ name = "g680-c-pagos" -> "g680-c-avisos" # forces replacement
# aws_sqs_queue.por_posicion[2] will be destroyed
Plan: 2 to add, 0 to change, 4 to destroy.
```
= con count: 3 colas destruidas y 2 creadas (pagos y avisos se recrean: se pierden sus mensajes); con for_each: 1 destruida.

Pulumi (el `for` de main.go nombra cada recurso "cola-"+nombre):
```
~ aws:iam/policy:Policy: (update)
- aws:sqs/queue:Queue: (delete)  [urn=urn:pulumi:prod::g680::aws:sqs/queue:Queue::cola-pedidos]
Resources:
    ~ 1 to update
    - 1 to delete
    2 changes. 6 unchanged
```
(La identidad la da el nombre lógico del recurso. Si en Go lo nombras con el índice, `"cola-"+strconv.Itoa(i)`, tienes la
misma trampa que con count: no lo ejecuté, es la misma regla.)

## 6 · Tests sin nube

```
$ go test -count=1 -v -run 'Prod|Dev|Politica' .
(compilación de los tests incluida: 2.1 s)
(otra vez, con caché: 1.3 s)
=== RUN   TestProdLlevaVersionado
--- PASS: TestProdLlevaVersionado (0.00s)
=== RUN   TestDevSinVersionado
--- PASS: TestDevSinVersionado (0.00s)
=== RUN   TestLaPoliticaNombraCadaCola
--- PASS: TestLaPoliticaNombraCadaCola (0.00s)
PASS
ok  	g680	0.014s
```
```
$ terraform test
(1.2 s)
tests/basico.tftest.hcl... in progress
  run "prod_lleva_versionado"... pass
  run "dev_sin_versionado"... pass
  run "la_politica_nombra_cada_cola"... pass
tests/basico.tftest.hcl... tearing down
tests/basico.tftest.hcl... pass

Success! 3 passed, 0 failed.
```
Los dos sin AWS ni moto: Pulumi con `pulumi.WithMocks` (main_test.go), Terraform con `mock_provider "aws" {}` (Terraform 1.7+).
Los mismos 3 tests. Diferencia: en Go el mock es código (cuenta tipos, guarda la política y la parsea con encoding/json);
en HCL son `assert` con expresiones.

## 7 · La trampa de Pulumi: un Output no es un string

```
=== RUN   TestTrampaOutput
    trampa_test.go:20: fmt.Sprintf   → "{0x3b33c6a8690}/*"
    trampa_test.go:23: pulumi.Sprintf → "arn:aws:mock:::datos/*"
--- PASS: TestTrampaOutput (0.00s)
```
`fmt.Sprintf("%v/*", b.Arn)` compila y no da error: mete la dirección de memoria del Output en el texto. Lo correcto es
`pulumi.Sprintf` o `ApplyT` (main.go arma la política con `pulumi.All(...).ApplyT(...)` y la convierte con `.(pulumi.StringOutput)`).

## Lo que moto NO hizo / no se probó
- moto ignora ClientToken, no hace drift y falla al actualizar con change set: nada de eso entra aquí (no hay CloudFormation;
  Pulumi y Terraform leen el recurso con su propio refresh).
- Las cifras de tiempo son de moto. En AWS real cambian; la comparación es entre herramientas en la misma máquina.
- No se probó Pulumi Cloud ni HCP Terraform (backend remoto, bloqueo de estado): solo backends locales.
