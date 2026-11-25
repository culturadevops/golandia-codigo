# 🐹 Bases de datos SQL desde Go, para administradores — material del curso

Código del vídeo [«Bases de datos SQL desde Go para administradores: 16 trampas que no dan error (curso completo)»](https://youtu.be/4pCeOBWamXM) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/4pCeOBWamXM**

---

> 🎬 Acompaña al vídeo **«Bases de datos SQL desde Go para administradores: 16 trampas que no dan error (curso completo)»** del canal **Golandia** (solo miembros).

Código del curso **«Bases de datos SQL desde Go, para administradores»** de
[Golandia](https://www.youtube.com/@golandia), exclusivo para miembros del canal.

> **No es el curso de «Postgres sin ORM».** Ese enseña las herramientas (database/sql, pgx,
> sqlc, migraciones). Este es para quien **opera** una base de datos desde Go: cada lección
> enseña cómo se hace y, sobre todo, **la trampa que no da error**. Casi todas compilan,
> pasan las pruebas y fallan en producción, en silencio.

## Cómo usar esto

```bash
git clone <este-repo> && cd curso-go-sql-administradores
go version                 # 1.24 o superior
docker version             # levanta un Postgres 17 limpio en el puerto 55432
python ejecutar_todo.py    # corre las 16 lecciones y regenera SALIDA.md
python ejecutar_todo.py --dejar   # igual, pero deja el Postgres corriendo
go run ./12_bloqueos       # o una sola lección (con el Postgres levantado)
```

¿Tu propio Postgres? `export CURSO_PG="postgres://usuario:clave@host:5432/base?sslmode=disable"`.
Las lecciones crean y borran sus tablas: **no lo apuntes a una base de datos que te importe.**

## Las 16 lecciones

| # | directorio | cómo se hace | la trampa |
|---:|---|---|---|
| 1 | `01_conectar` | `sql.Open` + `Ping` | `sql.Open` con la clave mal devuelve `err == nil` |
| 2 | `02_una_fila` | `QueryRow` + `Scan` | tratar todo error como «no existe» |
| 3 | `03_muchas_filas` | `rows.Next` | sin `rows.Err()`, un informe cortado en la fila 5 dice «OK» |
| 4 | `04_null` | `sql.NullString`, `*string` | el NULL que revienta solo en producción |
| 5 | `05_rows_sin_cerrar` | `defer rows.Close()` | el pool que se agota sin un solo error |
| 6 | `06_pool` | `SetMaxOpenConns` medido | 3 réplicas × 40 = 120 > `max_connections` |
| 7 | `07_inyeccion` | placeholders `$1` | `fmt.Sprintf` + `' OR '1'='1` devuelve la tabla entera |
| 8 | `08_rows_affected` | `RowsAffected` | el UPDATE que no encuentra nada no es un error |
| 9 | `09_transacciones` | `Begin`/`Commit`/`defer Rollback` | una sentencia por `db` dentro de la transacción |
| 10 | `10_context` | `ExecContext`, `statement_timeout` | sin context esperas lo que tarde |
| 11 | `11_actualizacion_perdida` | `stock = stock - 1`, `FOR UPDATE` | 50 ventas, y se pierden casi todas |
| 12 | `12_bloqueos` | `lock_timeout` | una transacción olvidada + un ALTER TABLE = nadie lee |
| 13 | `13_carga_masiva` | `pgx.CopyFrom` | el bucle de INSERT, medido |
| 14 | `14_migraciones` | migrador con `pg_advisory_lock` | dos réplicas migrando a la vez; la migración a medias |
| 15 | `15_tiempo` | `timestamptz` | `timestamp` tira la zona: 6 horas de diferencia |
| 16 | `16_proyecto` | **pgsalud** | la revisión de las tres de la mañana, con códigos de salida |

## El proyecto: pgsalud

```bash
go build -o pgsalud ./16_proyecto
./pgsalud -max-consulta 30s -max-idle-tx 1m        # texto
./pgsalud -json                                     # para otro programa
echo $?     # 0 todo bien · 1 hay problemas · 2 no pude ni mirar
```

Busca conexiones cerca de `max_connections`, consultas que no acaban, transacciones
olvidadas (`idle in transaction`) y sesiones bloqueadas, diciendo **qué pid las bloquea**.
`go run ./16_proyecto/escenario` le fabrica esos problemas durante unos segundos.

## La regla de la casa

`SALIDA.md` es la salida literal de `ejecutar_todo.py`, con la versión de Go, la de Postgres
y la fecha. Los números del vídeo salen de ahí. Los tiempos dependen de tu disco y de tu red:
si en tu máquina salen otros, es un dato, dilo.

---

<!-- ficha-repos-miembros -->
## Ficha rápida

### Requisitos

- Go 1.25 o superior.
- Docker (PostgreSQL de laboratorio).
- Python 3.10 o superior (para `ejecutar_todo.py`; solo biblioteca estándar salvo que se diga)

### Cómo ejecutarlo

```bash
git clone <este-repo> && cd curso-go-sql-administradores
python ejecutar_todo.py
```

El detalle de cada lección está arriba, en este mismo README.

### Estructura de carpetas

```
curso-go-sql-administradores/
├── 01_conectar/
│   └── main.go
├── 02_una_fila/
│   └── main.go
├── 03_muchas_filas/
│   └── main.go
├── 04_null/
│   └── main.go
├── 05_rows_sin_cerrar/
│   └── main.go
├── 06_pool/
│   └── main.go
├── 07_inyeccion/
│   └── main.go
├── 08_rows_affected/
│   └── main.go
├── 09_transacciones/
│   └── main.go
├── 10_context/
│   └── main.go
├── 11_actualizacion_perdida/
│   └── main.go
├── 12_bloqueos/
│   └── main.go
├── 13_carga_masiva/
│   └── main.go
├── 14_migraciones/
│   ├── migraciones/
│   └── main.go
├── 15_tiempo/
│   └── main.go
├── 16_proyecto/
│   ├── escenario/
│   └── main.go
├── comun/
│   └── comun.go
├── .gitignore
├── ejecutar_todo.py
├── go.mod
├── go.sum
├── README.md
└── SALIDA.md
```
