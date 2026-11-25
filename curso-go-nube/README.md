# Escribe tu propia nube en Go — material del curso

Código de estos vídeos del canal Golandia:

- [Escribe tu propia nube en Go — Parte 1: cómo habla la nube (curso completo)](https://youtu.be/M1Vi54m1fns) (miembros)
- [Escribe tu propia nube en Go — Parte 2: cómo ejecuta la nube (curso completo)](https://youtu.be/a3m1MtBufEg) (miembros)

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/M1Vi54m1fns**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/a3m1MtBufEg**

---

Canal **Golandia** (YouTube) · curso exclusivo para miembros.

El curso es la saga «escribí mi propia nube en Go» en dos vídeos largos, y su código es **nubecita**, un emulador de S3, DynamoDB, SQS, SNS, Lambda y Logs que entienden la CLI y el SDK oficiales de AWS. Es un único módulo de Go, así que el código va en la raíz y las carpetas `parte-1/` y `parte-2/` solo dicen qué ficheros mirar en cada vídeo. La documentación completa de nubecita (qué hace cada servicio, cómo usarlo desde el SDK y lo que falta) está en [`NUBECITA.md`](NUBECITA.md), que es su `README.md` original.

## Los vídeos

| # | carpeta | vídeo (título en YouTube) | id BD |
|---:|---|---|---:|
| 1 | [`parte-1/`](parte-1/) | Escribe tu propia nube en Go — Parte 1: cómo habla la nube (curso completo) | 175 |
| 2 | [`parte-2/`](parte-2/) | Escribe tu propia nube en Go — Parte 2: cómo ejecuta la nube (curso completo) | 176 |

## Requisitos

- Go 1.24 o superior.
- Docker (para la imagen y para que el lanzador levante funciones).
- Opcional: AWS CLI v2 para hablarle como a AWS.

## Cómo ejecutarlo

```bash
go test ./...
go run .
# o en contenedor:
docker build -t nubecita . && docker run --rm -p 4566:4566 nubecita
```

Credenciales de mentira (`prueba`/`prueba`) y `--endpoint-url http://127.0.0.1:4566`.

## Estructura

```
curso-go-nube/
├── ejemplo-funcion/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── parte-1/
│   └── README.md
├── parte-2/
│   └── README.md
├── .gitignore
├── almacen.go
├── autorizar.go
├── ayudas_test.go
├── bitacora.go
├── bitacora_test.go
├── colas.go
├── disparadores.go
├── disparadores_test.go
├── Dockerfile
├── dynamo.go
├── dynamo_test.go
├── enrutador.go
├── firma.go
├── firma_test.go
├── funciones.go
├── go.mod
├── go.sum
├── lambda.go
├── lambda_test.go
├── lanzador.go
├── lanzador_test.go
├── logs.go
├── logs_test.go
├── main.go
├── NUBECITA.md
├── permisos.go
├── permisos_test.go
├── persistencia.go
├── persistencia_test.go
├── registros.go
├── registros_test.go
├── s3.go
├── s3_test.go
├── sns.go
├── sns_test.go
├── sqs.go
├── sqs_test.go
├── tablas.go
└── temas.go
```

Cada carpeta de lección trae su `README.md` y, si el vídeo enseña terminales, un `EN-PANTALLA.md` con los comandos y salidas tal cual aparecen en el vídeo.
