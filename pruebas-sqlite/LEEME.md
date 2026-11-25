# Pruebas — «SQLite en producción con Go: ¿de verdad aguanta?» (vídeo 179)

> **Todo se ejecutó el 2026-09-03** dentro de contenedores oficiales: `golang:alpine`
> (go1.26.5 linux/amd64) y `postgres:17-alpine` (PostgreSQL 17.10), en la misma red de Docker.
> Salida completa y sin retocar en [`salida-real.txt`](salida-real.txt).
>
> ⚠️ Nada se compila en el host: App Control bloquea los binarios sin firmar en esta máquina.

## El método, y por qué es el que es

Un **solo programa** (`banco/main.go`) escribe contra los dos motores con `database/sql`, la
misma tabla y la misma carga. Solo cambian las banderas. Drivers: `modernc.org/sqlite` (Go puro,
sin cgo) y `jackc/pgx`.

⚖️ **La regla de equidad, que es lo que hace utilizable esta medida:** comparar SQLite en
`synchronous=NORMAL` contra Postgres en `synchronous_commit=on` es comparar **dos garantías
distintas**, y así es como se fabrican casi todos los banderines de internet. Aquí el cara a cara
va con **las dos parejas**: estricta contra estricta, y relajada contra relajada.

```sh
cd canales/golandia/recursos/pruebas-sqlite/banco
docker network create bancored
docker run -d --name bancopg --network bancored -e POSTGRES_PASSWORD=secreto postgres:17-alpine

MSYS_NO_PATHCONV=1 docker run --rm --network bancored -v "$(pwd -W):/banco" -w //banco \
    -e GOFLAGS=-mod=mod golang:alpine \
    sh -c "go build -o banco . && sh correr.sh && sh procesos.sh && sh barrido.sh"
```

El fichero de SQLite vive en el disco **del contenedor** (`/datos`), no en un volumen del host:
montar el disco de Windows falsearía la medida.

---

## A · «SQLite no aguanta concurrencia» — no es un límite, es una opción sin poner

| 8 escritores a la vez | escrituras/s | bloqueos en 4 s |
|---|---:|---:|
| diario clásico, **sin** `busy_timeout` | 254 | **584.718** |
| WAL, **sin** `busy_timeout` | 491 | **952.879** |
| WAL, **con** `busy_timeout(5000)` | **1.275** | **0** |

🥇 **El hallazgo:** casi un millón de errores en cuatro segundos, y **desaparecen enteros** al
poner un `busy_timeout`. Sin él, SQLite no espera su turno: vuelve inmediatamente con
`database is locked`. La fama de «no aguanta concurrencia» sale de aquí — de un valor por defecto,
no de una limitación del motor.

## B · El pragma que casi nadie menciona

| 8 escritores · WAL · pool 1 | escrituras/s |
|---|---:|
| `synchronous(FULL)` | 1.297 |
| `synchronous(NORMAL)` | **48.430** |
| `synchronous(OFF)` | 112.118 |

🥇 **Treinta y siete veces**, cambiando una palabra. Esto explica por qué dos artículos sobre
SQLite se contradicen sin que ninguno mienta: están midiendo cosas distintas y ninguno lo dice.
`NORMAL` con WAL es la combinación documentada: se aguanta la caída de la aplicación; lo que se
puede perder es la última transacción si cae **la máquina**.

## C · Cara a cara, con la misma garantía a cada lado

| 8 escritores | escrituras/s |
|---|---:|
| **estricta** — sqlite WAL `FULL` | 1.305 |
| **estricta** — postgres `synchronous_commit=on` | **7.208** |
| **relajada** — sqlite WAL `NORMAL` | 50.348 |
| **relajada** — postgres `synchronous_commit=off` | **87.217** |

Con ocho escritores compitiendo, **Postgres gana en los dos regímenes** (5,5x y 1,7x). Cualquiera
que te diga que SQLite gana en escrituras concurrentes no está igualando las garantías.

## D · Leer mientras alguien escribe (4 escriben + 4 leen)

| | escrituras/s | **lecturas/s** |
|---|---:|---:|
| sqlite diario clásico | 283 | **1** |
| sqlite **WAL** | 1.010 | **156.424** |
| postgres | 3.591 | 24.952 |

🥇 **Uno.** Con el diario clásico los lectores se quedan literalmente sin turno mientras alguien
escribe. Con WAL, los mismos lectores hacen **156.424 por segundo** — seis veces más que Postgres.
WAL no es una optimización opcional: sin él tienes otra base de datos, y es peor.

## E · Un solo escritor — el caso de casi todo el mundo

| 1 escritor | escrituras/s |
|---|---:|
| sqlite WAL `NORMAL` | **51.889** |
| postgres `synchronous_commit=off` | 17.709 |

**2,9 veces a favor de SQLite**, y la razón no es que el motor sea mejor: es que no hay viaje de
red. La consulta es una llamada a función dentro de tu propio proceso.

## F · Dos procesos contra el mismo fichero — y aquí se cae el mito

| | escrituras/s |
|---|---:|
| 1 proceso con 8 escritores | 45.654 |
| 2 procesos con 4 escritores cada uno | 33.447 + 33.283 = **66.730** |

🥇 **Dos procesos suman más que uno.** El candado entre procesos no arruina nada: lo que limitaba
al proceso único era su propio pool de una conexión. O sea que el límite de SQLite **no es
«un solo proceso»**. El límite es **una sola máquina**, porque el fichero no viaja.

## G · La barrida: dónde se cruzan las curvas

| escritores | sqlite | postgres | gana |
|---:|---:|---:|---|
| 1 | **48.962** | 16.636 | sqlite 2,9x |
| 2 | **48.021** | 31.018 | sqlite 1,5x |
| **4** | 47.411 | **57.257** | **postgres 1,2x** ← el cruce |
| 8 | 41.611 | **89.669** | postgres 2,2x |
| 16 | 49.369 | **89.220** | postgres 1,8x |
| 32 | 37.293 | **63.808** | postgres 1,7x |

🥇 **La respuesta del vídeo, en una fila:** el cruce está **entre dos y cuatro escritores
concurrentes**. Y fíjate en la forma de las dos curvas, que es lo que de verdad se aprende aquí:

- **SQLite es plana.** Con uno o con treinta y dos hace lo mismo, alrededor de cincuenta mil.
  No escala con la concurrencia — y tampoco se hunde con ella.
- **Postgres sube** hasta ocho, se estabiliza en dieciséis y **empieza a bajar en treinta y dos**.

---

## Lo que este material NO demuestra

- **No es una prueba de tu carga.** Son inserciones pequeñas en una tabla de tres columnas. Con
  transacciones largas, índices, `JOIN` o filas grandes los números cambian, y la forma de las
  curvas también puede cambiar.
- **No se midió durabilidad real ante un corte de corriente.** Se comparan configuraciones con
  garantías equivalentes según su documentación, no se tiró del cable.
- **No se midió la lectura pura sin escritores**, ni el arranque, ni la memoria.
- Un solo tamaño de máquina y un solo disco. Los 24 núcleos de esta máquina son los que ponen el
  techo de Postgres en treinta y dos escritores; en otra máquina el cruce puede caer en otro sitio.
