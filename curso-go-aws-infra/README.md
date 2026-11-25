# 🐹 Infraestructura de AWS desde Go: SDK y CDK — material del curso

Código del vídeo [«Infraestructura de AWS desde Go: SDK y CDK para administradores (curso completo)»](https://youtu.be/TfAcsxJjH00) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/TfAcsxJjH00**

---

> 🎬 Acompaña al vídeo **«Infraestructura de AWS desde Go: SDK y CDK para administradores (curso completo)»** del canal **Golandia** (solo miembros).

Código del curso **«Infraestructura de AWS desde Go: SDK y CDK»** de
[Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

> **No es AWS desde una aplicación** (S3, DynamoDB, Lambda: eso está en el curso de Go del
> canal). Es **infraestructura, para quien la administra**: la primera mitad opera lo que ya
> existe con el SDK v2 (EC2, VPC, grupos de seguridad, etiquetas); la segunda la declara como
> código con el CDK en Go. Cada lección enseña cómo se hace y **la trampa que no da error**.

## ⛔ Todo corre contra un emulador, y el curso se niega a otra cosa

Las lecciones crean, paran y borran instancias. `comun.Config()` **no arranca** si
`AWS_ENDPOINT_URL` no apunta a `localhost`. El código es el mismo que usarías contra AWS
(el SDK v2 lee `AWS_ENDPOINT_URL` solo), pero este repo no se ejecuta contra una cuenta real.

## Cómo usar esto

```bash
git clone <este-repo> && cd curso-go-aws-infra
go version                 # 1.24 o superior
docker version             # levanta moto (emulador de AWS) en el puerto 5000
npm i -g aws-cdk           # el CLI de CDK, para las lecciones 9 a 12
python ejecutar_todo.py    # corre las 13 lecciones y regenera SALIDA.md
python ejecutar_todo.py --dejar   # igual, pero deja el emulador corriendo

# o una sola lección, con el emulador levantado:
export AWS_ENDPOINT_URL=http://localhost:5000 AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test
go run ./08_filtro_vacio
```

## Las 13 lecciones

| # | directorio | cómo se hace | la trampa |
|---:|---|---|---|
| 1 | `01_donde_estoy` | STS `GetCallerIdentity` | la región equivocada devuelve 0, no un error |
| 2 | `02_paginacion` | el paginador del SDK | contar reservas; quedarse con la primera página |
| 3 | `03_filtros` | filtros `tag:` | `Env` ≠ `env`: 0 resultados, sin error |
| 4 | `04_crear` | etiquetas EN `RunInstances` | el reintento que duplica; la instancia sin nombre |
| 5 | `05_grupos_seguridad` | `smithy.APIError` | tratar «ya existe» como un fallo |
| 6 | `06_vpc` | la red de 6 recursos | borrar la VPC primero: `DependencyViolation` |
| 7 | `07_dry_run` | `DryRun: true` | el error que significa «adelante» |
| 8 | `08_filtro_vacio` | tres protecciones | la variable vacía para también producción |
| 9 | `09_cdk_synth` | `cdk synth` | una línea = 23 recursos y 2 NAT gateways |
| 10 | `10_cdk_diff` | `cdk diff` | cambiar el id deja el bucket huérfano |
| 11 | `11_cdk_destroy` | `RemovalPolicy` | «destroyed» y el bucket sigue ahí |
| 12 | `12_deriva` | plantilla contra realidad | `cdk diff` no ve lo cambiado a mano |
| 13 | `13_proyecto` | **inventario** | la auditoría de la cuenta, con códigos de salida |

## El proyecto: inventario

```bash
go build -o inventario ./13_proyecto
./inventario -regiones us-east-1,us-west-2 -cuenta 123456789012
./inventario -json
echo $?     # 0 todo bien · 1 hay hallazgos · 2 no pude ni mirar (o la cuenta no es la esperada)
```

Solo **lee**: instancias sin etiqueta `dueno`, puertos de administración abiertos a
`0.0.0.0/0`, volúmenes sueltos, IPs reservadas sin usar y NAT gateways.
`go run ./13_proyecto/escenario` le prepara una cuenta con todo eso.

## Lo que el emulador NO hace como AWS (y cómo lo trata el curso)

- **LocalStack ya no es gratis sin cuenta.** La imagen `localstack/localstack` 2026.8.4 se
  niega a arrancar sin `LOCALSTACK_AUTH_TOKEN` («License activation failed»). Por eso aquí se
  usa **moto** (código abierto), que emula EC2, VPC, STS, S3 y CloudFormation.
- **`ClientToken`**: AWS lo usa para que un `RunInstances` repetido no cree otra instancia;
  moto lo ignora. La lección 4 enseña el patrón que sí se puede ejecutar: buscar por
  etiqueta antes de crear.
- **Actualizar una pila con change set** devuelve un error 500 en moto. Las lecciones usan
  `cdk deploy --method=direct`, que en AWS también funciona.
- **Retain al actualizar**: al quitar un bucket con `DeletionPolicy: Retain` de la plantilla,
  AWS lo deja huérfano; moto lo borra. Por eso la lección 10 enseña el huérfano con
  `cdk diff`, que lo marca como `orphan`, y no desplegando. Al **destruir** la pila, moto sí
  respeta el `Retain` (lección 11).
- **`DetectStackDrift`** no existe en moto (error 500). En AWS úsalo; la lección 12 escribe
  su propio detector para el grupo de seguridad.

## Fuentes

- Precio del NAT gateway: 0,045 USD por hora y 0,045 USD por GB procesado, según
  <https://aws.amazon.com/vpc/pricing/> (ejemplo de US East), consultado el 2026-09-28.
  Mes de 730 horas. Compruébalo para tu región: cambia.

## La regla de la casa

`SALIDA.md` es la salida literal de `ejecutar_todo.py`. Los números del vídeo salen de ahí.
Los identificadores (`i-…`, `sg-…`) cambian en cada ejecución; los conteos, no.

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.24 o superior.
- Docker (emulador de AWS moto).
- Node.js y el CLI de CDK (`npm i -g aws-cdk`) para las lecciones 9 a 12.
- Python 3.10 o superior (para `ejecutar_todo.py`; solo biblioteca estándar salvo que se diga)

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-aws-infra
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-aws-infra/
├── 01_donde_estoy/
│   └── main.go
├── 02_paginacion/
│   └── main.go
├── 03_filtros/
│   └── main.go
├── 04_crear/
│   └── main.go
├── 05_grupos_seguridad/
│   └── main.go
├── 06_vpc/
│   └── main.go
├── 07_dry_run/
│   └── main.go
├── 08_filtro_vacio/
│   └── main.go
├── 09_cdk_synth/
│   └── main.go
├── 10_cdk_diff/
│   └── main.go
├── 11_cdk_destroy/
│   └── main.go
├── 12_deriva/
│   └── main.go
├── 13_proyecto/
│   ├── escenario/
│   └── main.go
├── cdk/
│   ├── cdk.json
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── comun/
│   ├── cdk.go
│   └── comun.go
├── evidencia/
│   └── localstack-sin-token.txt
├── .gitignore
├── ejecutar_todo.py
├── go.mod
├── go.sum
├── README.md
└── SALIDA.md
```
