# SALIDA — Terraform provider en Go (golandia 675, 2026-10-26)

Ejecutado el 2026-10-26 con `python correr.py` en Docker 29.6.2: imagen g675-tf = golang:1.27.2 + el binario de hashicorp/terraform:1.16.2.
Provider: github.com/hashicorp/terraform-plugin-framework v1.19.0 (terraform-plugin-go v0.31.0). API: notas-api, solo biblioteca estándar.
Sin publicar nada: `.terraformrc` con `dev_overrides` → "golandia/notas" = /build/<variante>. TF_CLI_ARGS=-no-color.
Las trampas se activan con `-ldflags "-X main.trampas=<nombre>"` sobre el MISMO código (ver nota.go): clave, 404, lista, vacio.
Todo corre en un mktemp dentro de un contenedor desechable. Las salidas completas, sin editar, están en salidas/*.txt;
aquí solo se reemplaza el bloque repetido «Warning: Provider development overrides are in effect» por una línea.
Las dos notas se crean en paralelo: el id que recibe cada una cambia de una ejecución a otra.

## Resumen medido
- plan en vacío (sin estado): 0 llamadas a la API. plan con 2 recursos: 2 GET. apply de 2 nuevas: 2 POST.
- apply de un cambio de título: 2 GET (refresco) + 1 PUT; `version = 1 -> (known after apply)`; id no cambia (UseStateForUnknown).
- destroy: 2 GET + 2 DELETE.
- drift (PUT a mano): plan → `~ titulo = "Cambiado a mano" -> "Compras"`; plan -refresh-only → «Objects have changed outside of Terraform».
- borrado a mano: Read recibe 404, RemoveResource → plan «1 to add».
- import con bloque: «Plan: 1 to import, 0 to add, 1 to change» porque la clave (la API no la devuelve) sale como `+ clave`; el apply hace un PUT extra (version 1 → 2). Después: No changes.
- terraform import (comando): 2 GET; state show sin clave. Import de un id que no existe (99): «Cannot import non-existent remote object».
- trampa clave: 3 de 3 plans seguidos con «0 to add, 2 to change», cada apply 2 PUT, version sube 1 → 2 → 3 → 4. Plan perpetuo.
- trampa 404: «Planning failed», y destroy falla igual (exit 1). Salida: terraform state rm.
- trampa lista: config ["zeta","casa"], la API devuelve ["casa","zeta"] → 2 errores «Provider produced inconsistent result after apply»; el recurso queda en el estado como tainted y el siguiente plan pide reemplazarlo (-/+).
- trampa vacio: «.texto: was null, but now cty.StringVal("")» → tainted, plan -/+.
- Estado: `terraform state show` pone clave = (sensitive value), pero terraform.tfstate la guarda en claro: "clave": "secreta".
- Código: provider 334 líneas en 4 archivos; API 147 líneas. Provider: 69 módulos en el grafo, 437 paquetes; binario 19.079.431 bytes con -s -w (27.691.565 sin).
- El framework v1.19.0 trae `WriteOnly bool` en schema.StringAttribute (visto en resource/schema/string_attribute.go; no se probó en esta demo).


## 0 · Versiones y tamaños

```
$ go version
go version go1.27.2 linux/amd64
$ terraform version
Terraform v1.16.2
on linux_amd64
$ go list -m (framework)
github.com/hashicorp/terraform-plugin-framework v1.19.0
github.com/hashicorp/terraform-plugin-go v0.31.0
$ go list -m all | wc -l
69
$ go list -deps . | wc -l
437
$ wc -l *.go (provider)
   74 cliente.go
   41 main.go
  162 nota.go
   57 provider.go
  334 total
$ wc -l main.go (api)
147 main.go
$ ls -l (binarios, -s -w)
19079431 /build/bien/terraform-provider-notas
9815554 /build/notas-api
27691565 sin -s -w
```

## 1 · Ciclo completo con el provider correcto: plan, apply, estado, cambio, destroy

```
$ terraform init
Initializing the backend...

Initializing provider plugins...

(aviso de dev_overrides: los providers no se instalan con init)
Terraform has been successfully initialized!
[código de salida 0 · 0.02 s]

$ terraform plan

(aviso de dev_overrides)

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # notas_nota.compras will be created
  + resource "notas_nota" "compras" {
      + clave     = (sensitive value)
      + creada    = (known after apply)
      + etiquetas = [
          + "casa",
          + "zeta",
        ]
      + id        = (known after apply)
      + texto     = "pan, café"
      + titulo    = "Compras"
      + version   = (known after apply)
    }

  # notas_nota.ideas will be created
  + resource "notas_nota" "ideas" {
      + creada  = (known after apply)
      + id      = (known after apply)
      + titulo  = "Ideas para vídeos"
      + version = (known after apply)
        # (1 unchanged attribute hidden)
    }

Plan: 2 to add, 0 to change, 0 to destroy.
[código de salida 0 · 0.08 s]
--- llamadas a notas-api ---

$ terraform apply -auto-approve

(aviso de dev_overrides)

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # notas_nota.compras will be created
  + resource "notas_nota" "compras" {
      + clave     = (sensitive value)
      + creada    = (known after apply)
      + etiquetas = [
          + "casa",
          + "zeta",
        ]
      + id        = (known after apply)
      + texto     = "pan, café"
      + titulo    = "Compras"
      + version   = (known after apply)
    }

  # notas_nota.ideas will be created
  + resource "notas_nota" "ideas" {
      + creada  = (known after apply)
      + id      = (known after apply)
      + titulo  = "Ideas para vídeos"
      + version = (known after apply)
        # (1 unchanged attribute hidden)
    }

Plan: 2 to add, 0 to change, 0 to destroy.
notas_nota.compras: Creating...
notas_nota.ideas: Creating...
notas_nota.compras: Creation complete after 0s [id=2]
notas_nota.ideas: Creation complete after 0s [id=1]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
[código de salida 0 · 0.11 s]
--- llamadas a notas-api ---
POST /notas 201
POST /notas 201

$ terraform state list
notas_nota.compras
notas_nota.ideas
[código de salida 0 · 0.02 s]

$ terraform state show notas_nota.compras
# notas_nota.compras:
resource "notas_nota" "compras" {
    clave     = (sensitive value)
    creada    = "2026-10-26T15:36:31Z"
    etiquetas = [
        "casa",
        "zeta",
    ]
    id        = "2"
    texto     = "pan, café"
    titulo    = "Compras"
    version   = 1
}
[código de salida 0 · 0.04 s]

$ terraform.tfstate (recurso compras)
      "name": "compras",
      "provider": "provider[\"registry.terraform.io/golandia/notas\"]",
      "instances": [
        {
          "schema_version": 0,
          "attributes": {
            "clave": "secreta",
            "creada": "2026-10-26T15:36:31Z",
            "etiquetas": [
              "casa",
              "zeta"
            ],
            "id": "2",
            "texto": "pan, café",
            "titulo": "Compras",
            "version": 1
          },
          "sensitive_attributes": [

$ terraform plan

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.
[código de salida 0 · 0.08 s]
--- llamadas a notas-api ---
GET /notas/2 200
GET /notas/1 200

$ terraform plan

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=1]
notas_nota.compras: Refreshing state... [id=2]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
        id        = "2"
      ~ titulo    = "Compras" -> "Compras del sábado"
      ~ version   = 1 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.
[código de salida 0 · 0.08 s]
--- llamadas a notas-api ---
GET /notas/1 200
GET /notas/2 200

$ terraform apply -auto-approve

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=1]
notas_nota.compras: Refreshing state... [id=2]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
        id        = "2"
      ~ titulo    = "Compras" -> "Compras del sábado"
      ~ version   = 1 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.
notas_nota.compras: Modifying... [id=2]
notas_nota.compras: Modifications complete after 0s [id=2]

Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
[código de salida 0 · 0.12 s]
--- llamadas a notas-api ---
GET /notas/1 200
GET /notas/2 200
PUT /notas/2 200

$ terraform destroy -auto-approve

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # notas_nota.compras will be destroyed
  - resource "notas_nota" "compras" {
      - clave     = (sensitive value) -> null
      - creada    = "2026-10-26T15:36:31Z" -> null
      - etiquetas = [
          - "casa",
          - "zeta",
        ] -> null
      - id        = "2" -> null
      - texto     = "pan, café" -> null
      - titulo    = "Compras del sábado" -> null
      - version   = 2 -> null
    }

  # notas_nota.ideas will be destroyed
  - resource "notas_nota" "ideas" {
      - creada  = "2026-10-26T15:36:31Z" -> null
      - id      = "1" -> null
      - titulo  = "Ideas para vídeos" -> null
      - version = 1 -> null
        # (1 unchanged attribute hidden)
    }

Plan: 0 to add, 0 to change, 2 to destroy.
notas_nota.ideas: Destroying... [id=1]
notas_nota.compras: Destroying... [id=2]
notas_nota.ideas: Destruction complete after 0s
notas_nota.compras: Destruction complete after 0s

Destroy complete! Resources: 2 destroyed.
[código de salida 0 · 0.15 s]
--- llamadas a notas-api ---
GET /notas/1 200
GET /notas/2 200
DELETE /notas/1 204
DELETE /notas/2 204
```

## 2 · Drift: cambio y borrado a mano en la API

```
(apply inicial: 2 notas creadas)

$ curl -X PUT http://127.0.0.1:8080/notas/2 -d '{"titulo":"Cambiado a mano","texto":"pan, café","etiquetas":["casa","zeta"]}'
{"id":"2","titulo":"Cambiado a mano","texto":"pan, café","etiquetas":["casa","zeta"],"creada":"2026-10-26T15:36:33Z","version":2}

$ terraform plan

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=1]
notas_nota.compras: Refreshing state... [id=2]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
        id        = "2"
      ~ titulo    = "Cambiado a mano" -> "Compras"
      ~ version   = 2 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.
[código de salida 0 · 0.08 s]
--- llamadas a notas-api ---
GET /notas/1 200
GET /notas/2 200

$ terraform plan -refresh-only

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=1]
notas_nota.compras: Refreshing state... [id=2]

Note: Objects have changed outside of Terraform

Terraform detected the following changes made outside of Terraform since the
last "terraform apply" which may have affected this plan:

  # notas_nota.compras has changed
  ~ resource "notas_nota" "compras" {
        id        = "2"
      ~ titulo    = "Compras" -> "Cambiado a mano"
      ~ version   = 1 -> 2
        # (4 unchanged attributes hidden)
    }

This is a refresh-only plan, so Terraform will not take any actions to undo
these. If you were expecting these changes then you can apply this plan to
record the updated values in the Terraform state without changing any remote
objects.
[código de salida 0 · 0.08 s]

$ terraform apply -auto-approve

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=1]
notas_nota.compras: Refreshing state... [id=2]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
        id        = "2"
      ~ titulo    = "Cambiado a mano" -> "Compras"
      ~ version   = 2 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.
notas_nota.compras: Modifying... [id=2]
notas_nota.compras: Modifications complete after 0s [id=2]

Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
[código de salida 0 · 0.11 s]
--- llamadas a notas-api ---
GET /notas/2 200
GET /notas/1 200
GET /notas/1 200
GET /notas/2 200
PUT /notas/2 200

$ curl -X DELETE http://127.0.0.1:8080/notas/1 
HTTP 204

$ terraform plan

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # notas_nota.ideas will be created
  + resource "notas_nota" "ideas" {
      + creada  = (known after apply)
      + id      = (known after apply)
      + titulo  = "Ideas para vídeos"
      + version = (known after apply)
        # (1 unchanged attribute hidden)
    }

Plan: 1 to add, 0 to change, 0 to destroy.
[código de salida 0 · 0.09 s]
--- llamadas a notas-api ---
GET /notas/2 200
GET /notas/1 404
```

## 3 · Import (bloque import y comando terraform import)

```
$ curl -X POST http://127.0.0.1:8080/notas -d '{"titulo":"Hecha a mano","texto":"antes de Terraform","etiquetas":["vieja"],"clave":"secreta"}'
{"id":"1","titulo":"Hecha a mano","texto":"antes de Terraform","etiquetas":["vieja"],"creada":"2026-10-26T15:36:35Z","version":1}

$ cat main.tf
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

$ terraform plan

(aviso de dev_overrides)
notas_nota.vieja: Preparing import... [id=1]
notas_nota.vieja: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.vieja will be updated in-place
  # (imported from "1")
  ~ resource "notas_nota" "vieja" {
      + clave     = (sensitive value)
        creada    = "2026-10-26T15:36:35Z"
        etiquetas = [
            "vieja",
        ]
        id        = "1"
        texto     = "antes de Terraform"
        titulo    = "Hecha a mano"
      ~ version   = 1 -> (known after apply)
    }

Plan: 1 to import, 0 to add, 1 to change, 0 to destroy.
[código de salida 0 · 0.07 s]
--- llamadas a notas-api ---
GET /notas/1 200

$ terraform apply -auto-approve

(aviso de dev_overrides)
notas_nota.vieja: Preparing import... [id=1]
notas_nota.vieja: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.vieja will be updated in-place
  # (imported from "1")
  ~ resource "notas_nota" "vieja" {
      + clave     = (sensitive value)
        creada    = "2026-10-26T15:36:35Z"
        etiquetas = [
            "vieja",
        ]
        id        = "1"
        texto     = "antes de Terraform"
        titulo    = "Hecha a mano"
      ~ version   = 1 -> (known after apply)
    }

Plan: 1 to import, 0 to add, 1 to change, 0 to destroy.
notas_nota.vieja: Importing... [id=1]
notas_nota.vieja: Import complete [id=1]
notas_nota.vieja: Modifying... [id=1]
notas_nota.vieja: Modifications complete after 0s [id=1]

Apply complete! Resources: 1 imported, 0 added, 1 changed, 0 destroyed.
[código de salida 0 · 0.10 s]
--- llamadas a notas-api ---
GET /notas/1 200
PUT /notas/1 200

$ terraform plan

(aviso de dev_overrides)
notas_nota.vieja: Refreshing state... [id=1]

No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.
[código de salida 0 · 0.09 s]

== y con el comando terraform import (sin bloque import) ==

$ terraform import notas_nota.vieja 1
notas_nota.vieja: Importing from ID "1"...
notas_nota.vieja: Import prepared!
  Prepared notas_nota for import
notas_nota.vieja: Refreshing state... [id=1]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.

[código de salida 0 · 0.08 s]
--- llamadas a notas-api ---
GET /notas/1 200
GET /notas/1 200

$ terraform state show notas_nota.vieja
# notas_nota.vieja:
resource "notas_nota" "vieja" {
    creada    = "2026-10-26T15:36:35Z"
    etiquetas = [
        "vieja",
    ]
    id        = "1"
    texto     = "antes de Terraform"
    titulo    = "Hecha a mano"
    version   = 2
}
[código de salida 0 · 0.04 s]

== importar un id que no existe ==

$ terraform state rm notas_nota.vieja
Removed notas_nota.vieja
Successfully removed 1 resource instance(s).
[código de salida 0 · 0.02 s]

$ terraform import notas_nota.vieja 99
notas_nota.vieja: Importing from ID "99"...
notas_nota.vieja: Import prepared!
  Prepared notas_nota for import
notas_nota.vieja: Refreshing state... [id=99]

Error: Cannot import non-existent remote object

While attempting to import an existing object to "notas_nota.vieja", the
provider detected that no object exists with the given id. Only pre-existing
objects can be imported; check that the id is correct and that it is
associated with the provider's configured region or endpoint, or use
"terraform apply" to create a new remote object for this resource.

[código de salida 1 · 0.06 s]
--- llamadas a notas-api ---
GET /notas/99 404
```

## 4 · Trampa: Read copia un campo que la API no devuelve (clave) → plan perpetuo

```
$ terraform apply -auto-approve

(aviso de dev_overrides)

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # notas_nota.compras will be created
  + resource "notas_nota" "compras" {
      + clave     = (sensitive value)
      + creada    = (known after apply)
      + etiquetas = [
          + "casa",
          + "zeta",
        ]
      + id        = (known after apply)
      + texto     = "pan, café"
      + titulo    = "Compras"
      + version   = (known after apply)
    }

  # notas_nota.ideas will be created
  + resource "notas_nota" "ideas" {
      + creada  = (known after apply)
      + id      = (known after apply)
      + titulo  = "Ideas para vídeos"
      + version = (known after apply)
        # (1 unchanged attribute hidden)
    }

Plan: 2 to add, 0 to change, 0 to destroy.
notas_nota.ideas: Creating...
notas_nota.compras: Creating...
notas_nota.ideas: Creation complete after 0s [id=1]
notas_nota.compras: Creation complete after 0s [id=2]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
[código de salida 0 · 0.12 s]
--- llamadas a notas-api ---
POST /notas 201
POST /notas 201

== plan número 1 ==

$ terraform plan

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
      + clave     = (sensitive value)
        id        = "2"
      ~ version   = 1 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

  # notas_nota.ideas will be updated in-place
  ~ resource "notas_nota" "ideas" {
        id      = "1"
      ~ version = 1 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 2 to change, 0 to destroy.
[código de salida 0 · 0.08 s]

$ terraform apply -auto-approve

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
      + clave     = (sensitive value)
        id        = "2"
      ~ version   = 1 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

  # notas_nota.ideas will be updated in-place
  ~ resource "notas_nota" "ideas" {
        id      = "1"
      ~ version = 1 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 2 to change, 0 to destroy.
notas_nota.ideas: Modifying... [id=1]
notas_nota.compras: Modifying... [id=2]
notas_nota.compras: Modifications complete after 0s [id=2]
notas_nota.ideas: Modifications complete after 0s [id=1]

Apply complete! Resources: 0 added, 2 changed, 0 destroyed.
[código de salida 0 · 0.11 s]
--- llamadas a notas-api ---
GET /notas/2 200
GET /notas/1 200
GET /notas/1 200
GET /notas/2 200
PUT /notas/1 200
PUT /notas/2 200

== plan número 2 ==

$ terraform plan

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
      + clave     = (sensitive value)
        id        = "2"
      ~ version   = 2 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

  # notas_nota.ideas will be updated in-place
  ~ resource "notas_nota" "ideas" {
        id      = "1"
      ~ version = 2 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 2 to change, 0 to destroy.
[código de salida 0 · 0.09 s]

$ terraform apply -auto-approve

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
      + clave     = (sensitive value)
        id        = "2"
      ~ version   = 2 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

  # notas_nota.ideas will be updated in-place
  ~ resource "notas_nota" "ideas" {
        id      = "1"
      ~ version = 2 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 2 to change, 0 to destroy.
notas_nota.ideas: Modifying... [id=1]
notas_nota.compras: Modifying... [id=2]
notas_nota.compras: Modifications complete after 0s [id=2]
notas_nota.ideas: Modifications complete after 0s [id=1]

Apply complete! Resources: 0 added, 2 changed, 0 destroyed.
[código de salida 0 · 0.12 s]
--- llamadas a notas-api ---
GET /notas/2 200
GET /notas/1 200
GET /notas/1 200
GET /notas/2 200
PUT /notas/1 200
PUT /notas/2 200

== plan número 3 ==

$ terraform plan

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=1]
notas_nota.compras: Refreshing state... [id=2]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
      + clave     = (sensitive value)
        id        = "2"
      ~ version   = 3 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

  # notas_nota.ideas will be updated in-place
  ~ resource "notas_nota" "ideas" {
        id      = "1"
      ~ version = 3 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 2 to change, 0 to destroy.
[código de salida 0 · 0.10 s]

$ terraform apply -auto-approve

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=1]
notas_nota.compras: Refreshing state... [id=2]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # notas_nota.compras will be updated in-place
  ~ resource "notas_nota" "compras" {
      + clave     = (sensitive value)
        id        = "2"
      ~ version   = 3 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

  # notas_nota.ideas will be updated in-place
  ~ resource "notas_nota" "ideas" {
        id      = "1"
      ~ version = 3 -> (known after apply)
        # (4 unchanged attributes hidden)
    }

Plan: 0 to add, 2 to change, 0 to destroy.
notas_nota.ideas: Modifying... [id=1]
notas_nota.compras: Modifying... [id=2]
notas_nota.compras: Modifications complete after 0s [id=2]
notas_nota.ideas: Modifications complete after 0s [id=1]

Apply complete! Resources: 0 added, 2 changed, 0 destroyed.
[código de salida 0 · 0.11 s]
--- llamadas a notas-api ---
GET /notas/2 200
GET /notas/1 200
GET /notas/1 200
GET /notas/2 200
PUT /notas/1 200
PUT /notas/2 200
```

## 5 · Trampa: Read no trata el 404 → ni plan ni destroy

```
(apply inicial: 2 notas creadas)

$ curl -X DELETE http://127.0.0.1:8080/notas/1 
HTTP 204

$ terraform plan

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=1]
notas_nota.ideas: Refreshing state... [id=2]

Planning failed. Terraform encountered an error while generating this plan.

Error: Leyendo la nota

  with notas_nota.compras,
  on main.tf line 9, in resource "notas_nota" "compras":
   9: resource "notas_nota" "compras" {

la nota no existe (404)
[código de salida 1 · 0.08 s]
--- llamadas a notas-api ---
GET /notas/1 404
GET /notas/2 200

$ terraform destroy -auto-approve

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=2]
notas_nota.compras: Refreshing state... [id=1]

Error: Leyendo la nota

  with notas_nota.compras,
  on main.tf line 9, in resource "notas_nota" "compras":
   9: resource "notas_nota" "compras" {

la nota no existe (404)
[código de salida 1 · 0.08 s]

== la salida: sacarla del estado a mano ==

$ terraform state rm notas_nota.compras
Removed notas_nota.compras
Successfully removed 1 resource instance(s).
[código de salida 0 · 0.02 s]

$ terraform plan

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=2]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # notas_nota.compras will be created
  + resource "notas_nota" "compras" {
      + clave     = (sensitive value)
      + creada    = (known after apply)
      + etiquetas = [
          + "casa",
          + "zeta",
        ]
      + id        = (known after apply)
      + texto     = "pan, café"
      + titulo    = "Compras"
      + version   = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.
[código de salida 0 · 0.08 s]
```

## 6 · Trampa: etiquetas como lista y la API las ordena → inconsistent result after apply

```
$ terraform apply -auto-approve

(aviso de dev_overrides)

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # notas_nota.compras will be created
  + resource "notas_nota" "compras" {
      + clave     = (sensitive value)
      + creada    = (known after apply)
      + etiquetas = [
          + "zeta",
          + "casa",
        ]
      + id        = (known after apply)
      + texto     = "pan, café"
      + titulo    = "Compras"
      + version   = (known after apply)
    }

  # notas_nota.ideas will be created
  + resource "notas_nota" "ideas" {
      + creada  = (known after apply)
      + id      = (known after apply)
      + titulo  = "Ideas para vídeos"
      + version = (known after apply)
        # (1 unchanged attribute hidden)
    }

Plan: 2 to add, 0 to change, 0 to destroy.
notas_nota.ideas: Creating...
notas_nota.compras: Creating...
notas_nota.ideas: Creation complete after 0s [id=2]

Error: Provider produced inconsistent result after apply

When applying changes to notas_nota.compras, provider
"provider[\"registry.terraform.io/golandia/notas\"]" produced an unexpected
new value: .etiquetas[0]: was cty.StringVal("zeta"), but now
cty.StringVal("casa").

This is a bug in the provider, which should be reported in the provider's own
issue tracker.

Error: Provider produced inconsistent result after apply

When applying changes to notas_nota.compras, provider
"provider[\"registry.terraform.io/golandia/notas\"]" produced an unexpected
new value: .etiquetas[1]: was cty.StringVal("casa"), but now
cty.StringVal("zeta").

This is a bug in the provider, which should be reported in the provider's own
issue tracker.
[código de salida 1 · 0.10 s]
--- llamadas a notas-api ---
POST /notas 201
POST /notas 201

$ terraform state list
notas_nota.compras
notas_nota.ideas
[código de salida 0 · 0.02 s]

$ terraform plan

(aviso de dev_overrides)
notas_nota.ideas: Refreshing state... [id=2]
notas_nota.compras: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # notas_nota.compras is tainted, so must be replaced
-/+ resource "notas_nota" "compras" {
      ~ creada    = "2026-10-26T15:36:42Z" -> (known after apply)
      ~ etiquetas = [
          - "casa",
            "zeta",
          + "casa",
        ]
      ~ id        = "1" -> (known after apply)
      ~ version   = 1 -> (known after apply)
        # (3 unchanged attributes hidden)
    }

Plan: 1 to add, 0 to change, 1 to destroy.
[código de salida 0 · 0.08 s]
```

## 7 · Trampa: texto opcional sin Default y la API devuelve "" → inconsistent result after apply

```
$ terraform apply -auto-approve

(aviso de dev_overrides)

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # notas_nota.compras will be created
  + resource "notas_nota" "compras" {
      + clave     = (sensitive value)
      + creada    = (known after apply)
      + etiquetas = [
          + "casa",
          + "zeta",
        ]
      + id        = (known after apply)
      + texto     = "pan, café"
      + titulo    = "Compras"
      + version   = (known after apply)
    }

  # notas_nota.ideas will be created
  + resource "notas_nota" "ideas" {
      + creada  = (known after apply)
      + id      = (known after apply)
      + titulo  = "Ideas para vídeos"
      + version = (known after apply)
    }

Plan: 2 to add, 0 to change, 0 to destroy.
notas_nota.ideas: Creating...
notas_nota.compras: Creating...
notas_nota.compras: Creation complete after 0s [id=2]

Error: Provider produced inconsistent result after apply

When applying changes to notas_nota.ideas, provider
"provider[\"registry.terraform.io/golandia/notas\"]" produced an unexpected
new value: .texto: was null, but now cty.StringVal("").

This is a bug in the provider, which should be reported in the provider's own
issue tracker.
[código de salida 1 · 0.10 s]
--- llamadas a notas-api ---
POST /notas 201
POST /notas 201

$ terraform state list
notas_nota.compras
notas_nota.ideas
[código de salida 0 · 0.02 s]

$ terraform plan

(aviso de dev_overrides)
notas_nota.compras: Refreshing state... [id=2]
notas_nota.ideas: Refreshing state... [id=1]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # notas_nota.ideas is tainted, so must be replaced
-/+ resource "notas_nota" "ideas" {
      ~ creada  = "2026-10-26T15:36:44Z" -> (known after apply)
      ~ id      = "1" -> (known after apply)
      ~ version = 1 -> (known after apply)
        # (2 unchanged attributes hidden)
    }

Plan: 1 to add, 0 to change, 1 to destroy.
[código de salida 0 · 0.07 s]
```
