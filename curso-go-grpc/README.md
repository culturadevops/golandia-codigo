# 🐹 gRPC en Go: protobuf, streaming y las trampas que no dan error — material del curso

Código del vídeo [«gRPC en Go: protobuf, streaming y las trampas que no dan error (curso completo)»](https://youtu.be/Ztk4MiANu5A) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/Ztk4MiANu5A**

---

> 🎬 Acompaña al vídeo **«gRPC en Go: protobuf, streaming y las trampas que no dan error (curso completo)»** del canal **Golandia** (solo miembros).

Código del curso **«gRPC en Go»** de [Golandia](https://www.youtube.com/@golandia), exclusivo para
miembros del canal.

> gRPC es fácil de arrancar y difícil de tener bien en producción: casi todo lo que sale mal no da
> error. Un número de campo reutilizado cambia los datos en silencio, una llamada sin plazo espera
> para siempre, un pánico tumba el servidor entero y el balanceo manda todo a una réplica. Cada
> lección enseña cómo se hace y, sobre todo, **la trampa que no da error**, medida.

grpc-go **v1.84.0**, protobuf-go **v1.36.12**, Go **1.26**.

## Cómo usar esto

Todo corre en Docker, con Go, `protoc` y `grpcurl` iguales en cualquier máquina:

```bash
python ejecutar_todo.py          # construye la imagen, corre las 9 lecciones y regenera SALIDA.md
python ejecutar_todo.py 03 07    # solo esas
```

Para trabajar a mano dentro de la imagen:

```bash
docker build -t curso-grpc:1 -f Dockerfile.curso .
docker run --rm -it -v "$PWD:/curso" curso-grpc:1 bash
sh generar.sh                    # regenera gen/ desde proto/
go run ./03_plazos
```

`SALIDA.md` lo escribe `ejecutar_todo.py`: es de donde sale cada línea del vídeo.

## Las 9 lecciones

| # | directorio | cómo se hace | la trampa |
|---:|---|---|---|
| 1 | `01_protobuf` | el `.proto`, `reserved`, `optional` | reutilizar un número: `stock=5` se lee como `precio=5`, sin error |
| 2 | `02_cliente` | `grpc.NewClient`, `WaitForReady` | `NewClient` contra un puerto vacío devuelve `err == nil` |
| 3 | `03_plazos` | `context.WithTimeout` | sin plazo se espera para siempre; y el servidor que no mira el `ctx` trabaja para nadie |
| 4 | `04_errores` | `status.Errorf`, interceptor que recupera | `fmt.Errorf` llega como `Unknown`; `nil, nil` llega como OK; un pánico tumba el proceso |
| 5 | `05_streaming` | flujos, `io.EOF`, `cancel()` | `Send` se bloquea si el cliente no lee; sin cancelar, el handler no termina |
| 6 | `06_tamano` | `MaxCallRecvMsgSize`, flujo a trozos | 6 MB: `ResourceExhausted` en el cliente y el servidor sin enterarse |
| 7 | `07_balanceo` | `round_robin`, `MaxConnectionAge` | `pick_first` manda el 100 % a una réplica; la réplica nueva no recibe nada |
| 8 | `08_apagar` | `GracefulStop`, health, interceptores | `Stop()` corta las llamadas en vuelo |
| 9 | `09_rest_o_grpc` | reflexión y `grpcurl` | `curl` no habla gRPC; sin reflexión, `grpcurl` tampoco ve nada |

La lección 7 levanta réplicas en una red de Docker (`curso-grpc-red`) y la orquesta
`ejecutar_todo.py`; al terminar, borra los contenedores y la red.

## Lo que hay que saber de las medidas

- Las latencias de la lección 9 son en la misma máquina, sin red de por medio: comparan el coste
  de cada protocolo, no lo que verás en internet.
- La lección 7 tarda unos 100 s a propósito: el resolvedor DNS de grpc-go no vuelve a preguntar
  más de una vez cada 30 s (`MinResolutionInterval` en `internal/resolver/dns`).

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.26 o superior.
- Docker.
- Python 3.10 o superior (para `ejecutar_todo.py`; solo biblioteca estándar salvo que se diga)

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-grpc
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-grpc/
├── 01_protobuf/
│   └── main.go
├── 02_cliente/
│   └── main.go
├── 03_plazos/
│   └── main.go
├── 04_errores/
│   └── main.go
├── 05_streaming/
│   └── main.go
├── 06_tamano/
│   └── main.go
├── 07_balanceo/
│   └── main.go
├── 08_apagar/
│   └── main.go
├── 09_rest_o_grpc/
│   └── main.go
├── comun/
│   └── comun.go
├── gen/
│   ├── inventario/
│   └── l01/
├── proto/
│   ├── inventario/
│   └── l01/
├── .gitattributes
├── .gitignore
├── Dockerfile.curso
├── ejecutar_todo.py
├── generar.sh
├── go.mod
├── go.sum
├── README.md
└── SALIDA.md
```
