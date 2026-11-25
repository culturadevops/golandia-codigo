# SALIDA — curso «Bases de datos SQL desde Go, para administradores»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-09-28 11:05:47Z`
- go version go1.26.5 windows/amd64
- servidor: `postgres (PostgreSQL) 17.11` (imagen `postgres:17-alpine`)
- driver: `github.com/jackc/pgx/v5` a traves de `database/sql`
- sistema: `Windows 10` · Postgres en Docker

## Leccion 01 — Conectar: sql.Open no conecta

`go run ./01_conectar` · codigo de salida `0`

```
== 1.1 · sql.Open no conecta ==
  clave MAL              sql.Open -> err = <nil>
                         conexiones abiertas: 0
                         db.Ping  -> failed to connect to `user=curso database=tienda`: 127.0.0.1:55432 (127.0.0.1): failed SASL auth: FATAL: password authentication failed for user "curso" (SQLSTATE 28P01)

  servidor que no existe sql.Open -> err = <nil>
                         conexiones abiertas: 0
                         db.Ping  -> failed to connect to `user=curso database=tienda`: 127.0.0.1:1 (127.0.0.1): dial error: dial tcp 127.0.0.1:1: connectex: No se puede establecer una conexión ya que el equipo de destino denegó expresamente dicha conexión.

== 1.2 · Lo correcto: abrir Y hacer Ping al arrancar ==
  antes del Ping:   conexiones abiertas = 0
  despues del Ping: conexiones abiertas = 1
  servidor: PostgreSQL 17.11

== 1.3 · *sql.DB no es una conexion: es un pool ==
  Se abre UNA vez al arrancar y se comparte entre todas las goroutines.
  Abrir un sql.DB por peticion es abrir un pool por peticion.
```

## Leccion 02 — Una fila y sql.ErrNoRows

`go run ./02_una_fila` · codigo de salida `0`

```
== 2.1 · QueryRow + Scan ==
  db-01 -> ip 10.0.2.21, cpu 64%

== 2.2 · La trampa: todo error es «no existe» ==
  ipMal("web-09")                   -> no existe
  ipMal("db-01") con la columna mal -> no existe
  Las dos dicen lo mismo. La segunda es un fallo de codigo, y db-01 SI existe.

== 2.3 · Lo correcto: errors.Is(err, sql.ErrNoRows) ==
  ipBien("web-09")                  -> web-09: no existe
  ipBien("db-01") con la columna mal -> buscando db-01: ERROR: column "ip_addr" does not exist (SQLSTATE 42703)
```

## Leccion 03 — Muchas filas y rows.Err

`go run ./03_muchas_filas` · codigo de salida `0`

```
== 3.1 · El bucle de siempre, sin rows.Err() ==
    fila 1 -> 25
    fila 2 -> 33
    fila 3 -> 50
    fila 4 -> 100
  procesadas 4 filas de 8 -> informe OK

== 3.2 · El mismo bucle, mirando rows.Err() ==
  ERROR: el informe se corto en la fila 5: ERROR: division by zero (SQLSTATE 22012)
```

## Leccion 04 — NULL

`go run ./04_null` · codigo de salida `0`

```
== 4.1 · Scan de NULL a string ==
    web-01    detras del balanceador
    web-02    detras del balanceador
    web-03    detras del balanceador
    db-01     primaria
    db-02     ERROR: sql: Scan error on column index 1, name "notas": converting NULL to string is unsupported

== 4.2 · Tres formas de recibir un NULL ==
  sql.NullString -> Valid=false String=""
  *string        -> <nil>
  COALESCE       -> ""

== 4.3 · NULL no es la cadena vacia ==
  notas IS NULL -> 1   notas = '' -> 1
  WHERE notas <> 'primaria' -> 6 de 8   (la fila con NULL no entra)
```

## Leccion 05 — rows sin cerrar: el pool que se agota

`go run ./05_rows_sin_cerrar` · codigo de salida `0`

```
== 5.1 · Sin rows.Close(), pool de 3 conexiones ==
  llamada 1 -> web-01   (conexiones en uso: 1 de 3)
  llamada 2 -> web-01   (conexiones en uso: 2 de 3)
  llamada 3 -> web-01   (conexiones en uso: 3 de 3)
  llamada 4 -> sigue esperando tras 2 s   (en uso: 3 de 3, esperando turno: 1)
  Ni un error. Sin timeout, esta llamada espera para siempre.

== 5.2 · Con defer rows.Close() ==
  llamada 1 -> web-01   (conexiones en uso: 0 de 3)
  llamada 2 -> web-01   (conexiones en uso: 0 de 3)
  llamada 3 -> web-01   (conexiones en uso: 0 de 3)
  llamada 4 -> web-01   (conexiones en uso: 0 de 3)
  llamada 5 -> web-01   (conexiones en uso: 0 de 3)

== 5.3 · El mismo olvido, con un context que se cancela ==
  llamada 1 -> web-01   (conexiones en uso: 1 de 3)
  llamada 2 -> web-01   (conexiones en uso: 1 de 3)
  llamada 3 -> web-01   (conexiones en uso: 1 de 3)
  llamada 4 -> web-01   (conexiones en uso: 1 de 3)
  llamada 5 -> web-01   (conexiones en uso: 1 de 3)
```

## Leccion 06 — El pool, medido

`go run ./06_pool` · codigo de salida `0`

```
== 6.1 · 40 consultas de 100 ms, a la vez ==
  pool de  1 ->  4103 ms   (esperaron turno: 39 consultas,  79942 ms en total)
  pool de  5 ->   828 ms   (esperaron turno: 35 consultas,  14537 ms en total)
  pool de 20 ->   214 ms   (esperaron turno: 20 consultas,   2169 ms en total)
  pool de 40 ->   115 ms   (esperaron turno:  0 consultas,      0 ms en total)

== 6.2 · El limite que no es tuyo: max_connections ==
  max_connections del servidor: 100
  3 replicas de tu servicio, cada una con un pool de 40 = 120 conexiones:
  consultas que fallaron: 20 de 120
  error: failed to connect to `user=curso database=tienda`: 127.0.0.1:55432 (127.0.0.1): server error: FATAL: sorry, too many clients already (SQLSTATE 53300)

== 6.3 · Los cuatro ajustes ==
  db.SetMaxOpenConns(n)        tope de conexiones de ESTE proceso
  db.SetMaxIdleConns(n)        cuantas se quedan abiertas esperando
  db.SetConnMaxLifetime(d)     renovarlas (balanceadores, failover)
  db.SetConnMaxIdleTime(d)     cerrar las que nadie usa
```

## Leccion 07 — Inyeccion SQL y placeholders

`go run ./07_inyeccion` · codigo de salida `0`

```
== 7.1 · Con fmt.Sprintf ==
  nombre = web-01        -> 1 fila
  la consulta que sale:   SELECT nombre, ip FROM servidores WHERE nombre = 'x' OR '1'='1'
  nombre = x' OR '1'='1  ->  8 filas

== 7.2 · Con placeholder $1 ==
  nombre = web-01        -> 1 fila
  nombre = x' OR '1'='1  ->  0 filas

== 7.3 · Y el DROP TABLE? ==
   ERROR: ERROR: cannot insert multiple commands into a prepared statement (SQLSTATE 42601)
  Te salva el driver (protocolo extendido: una sentencia por consulta), no tu codigo.
  La tabla sigue: 8 filas

== 7.4 · Lo que NO puede ir en un placeholder: el nombre de una tabla ==
  SELECT count(*) FROM $1  -> ERROR: ERROR: syntax error at or near "$1" (SQLSTATE 42601)
  tabla servidores  -> 1 fila   (SELECT count(*) FROM "servidores")
  tabla pg_authid   -> rechazada: no esta en la lista permitida
```

## Leccion 08 — Exec y RowsAffected

`go run ./08_rows_affected` · codigo de salida `0`

```
== 8.1 · Sin mirar RowsAffected ==
  web-01  -> puesto en mantenimiento  (err = <nil>)
  Web-02  -> puesto en mantenimiento  (err = <nil>)
  en la tabla hay 1 servidor(es) en mantenimiento, no 2

== 8.2 · Mirando RowsAffected ==
  web-03  -> puesto en mantenimiento
  Web-02  -> ERROR: Web-02: esperaba cambiar 1 fila y cambie 0

== 8.3 · El otro lado: el filtro vacio que lo cambia TODO ==
  filtro = "" -> 8 filas cambiadas de 8
  RowsAffected lo dice DESPUES. Para abortar a tiempo: transaccion + comprobar + Rollback (leccion 9).
```

## Leccion 09 — Transacciones

`go run ./09_transacciones` · codigo de salida `0`

```
== 9.1 · Una sentencia por db dentro de la transaccion ==
  antes:               norte=100 sur=100 total=200
  trasladar 30 ->      el destino no tiene sitio -> Rollback
  despues:             norte=100 sur=130 total=230

== 9.2 · Todo por tx, con defer tx.Rollback() ==
  antes:               norte=100 sur=100 total=200
  trasladar 30 ->      el destino no tiene sitio -> Rollback
  despues:             norte=100 sur=100 total=200

== 9.3 · La red de la leccion 8: comprobar ANTES del Commit ==
  cambiaria 8 filas y esperaba 1 -> Rollback
  servidores con cpu = 0 en la tabla: 0

== 9.4 · Y con un pool de UNA conexion, usar db dentro de tx se cuelga ==
  sigue colgado tras 2 s: la unica conexion la tiene la transaccion,
  y db.Exec espera a que quede una libre. Sin error.
```

## Leccion 10 — context y timeouts

`go run ./10_context` · codigo de salida `0`

```
== 10.1 · Sin context: esperas lo que tarde ==
  db.Exec(pg_sleep(3))                     -> volvio a los 3.0 s

== 10.2 · Con context.WithTimeout de 1 s ==
  db.ExecContext(ctx, pg_sleep(3))         -> volvio a los 1.0 s
  error: timeout: context deadline exceeded
  consultas pg_sleep aun corriendo en el servidor: 0

== 10.3 · La red por si alguien olvida el context: statement_timeout ==
  sin context, con statement_timeout=1000  -> volvio a los 1.0 s
  error: ERROR: canceling statement due to statement timeout (SQLSTATE 57014)
```

## Leccion 11 — La actualizacion perdida

`go run ./11_actualizacion_perdida` · codigo de salida `0`

```
== 11.1 · Leer en Go, restar en Go, guardar: 50 ventas sobre 100 ==
  intento 1: esperaba stock 50, quedo 98  -> 48 ventas perdidas
  intento 2: esperaba stock 50, quedo 98  -> 48 ventas perdidas
  intento 3: esperaba stock 50, quedo 98  -> 48 ventas perdidas

== 11.2 · Que reste la base de datos: SET stock = stock - 1 ==
  quedo 50   (31 ms)

== 11.3 · Si necesitas leer antes de escribir: SELECT ... FOR UPDATE ==
  quedo 50   (194 ms: las ventas hacen cola en la fila)
```

## Leccion 12 — Bloqueos: la migracion que tumba las lecturas

`go run ./12_bloqueos` · codigo de salida `0`

```
== 12.1 · Una transaccion abierta + un ALTER TABLE + un SELECT ==
  A: BEGIN; UPDATE ... web-01;   y se queda ahi (idle in transaction)
  B: ALTER TABLE servidores ADD COLUMN zona text   (la migracion)
  C: SELECT count(*) FROM servidores                (una lectura cualquiera)

  pid    state                espera bloqueado consulta
  392    idle in transaction  Client           UPDATE servidores SET cpu = cpu WHERE nombre
  393    active               Lock   392       ALTER TABLE servidores ADD COLUMN zona text
  394    active               Lock   393       SELECT count(*) FROM servidores

  C -> timeout: context deadline exceeded
  A hace ROLLBACK ->  B termina: <nil>

== 12.2 · La migracion con lock_timeout: falla rapido y no arrastra a nadie ==
  B con lock_timeout=1s -> 1.0 s: ERROR: canceling statement due to lock timeout (SQLSTATE 55P03)
  C -> 8 filas, err=<nil>   (la lectura ya no espera)

== 12.3 · Y la transaccion olvidada: idle_in_transaction_session_timeout ==
  A, 1,5 s despues -> FATAL: terminating connection due to idle-in-transaction timeout (SQLSTATE 25P03)
```

## Leccion 13 — Carga masiva

`go run ./13_carga_masiva` · codigo de salida `0`

```
== 13.1 · Tres formas de cargar filas (mediana de 3) ==
  fila a fila (autocommit)     10000 filas en  8292 ms        1206 filas/s   (en la tabla: 10000)
  en UNA transaccion           10000 filas en  2592 ms        3857 filas/s   (en la tabla: 10000)
  COPY (pgx.CopyFrom)         500000 filas en   113 ms     4396973 filas/s   (en la tabla: 500000)

  una transaccion frente a fila a fila: 3 veces mas filas por segundo
  COPY frente a fila a fila:            3646 veces mas filas por segundo
```

## Leccion 14 — Migraciones

`go run ./14_migraciones` · codigo de salida `0`

```
== 14.1 · Dos replicas arrancan a la vez, SIN cerrojo ==
  replica 1 -> aplico 0 y FALLO: 001_eventos.sql: ERROR: duplicate key value violates unique constraint "pg_type_typname_nsp_index" (SQLSTATE 23505)
  replica 2 -> aplico 3, ok

== 14.2 · Las mismas dos, con pg_advisory_lock ==
  replica 1 -> aplico 0, ok
  replica 2 -> aplico 3, ok

== 14.3 · La 004 tiene dos sentencias y la segunda esta mal ==
  004_origen.sql:
    ALTER TABLE eventos ADD COLUMN origen text;
    ALTER TABLE evento ADD COLUMN region text;
  sin transaccion, sentencia a sentencia (como hacen muchas herramientas):
    -> ERROR: relation "evento" does not exist (SQLSTATE 42P01)
    columnas de eventos: id, servidor, nivel, mensaje, cuando, origen
    la mitad aplicada, y la 004 sin apuntar: al reintentar, falla la PRIMERA.
  en su transaccion (el migrador de arriba):
    -> 004_origen.sql: ERROR: relation "evento" does not exist (SQLSTATE 42P01)
    columnas de eventos: id, servidor, nivel, mensaje, cuando

== 14.4 · CREATE INDEX CONCURRENTLY dentro de una transaccion ==
  -> ERROR: CREATE INDEX CONCURRENTLY cannot run inside a transaction block (SQLSTATE 25001)
  fuera de la transaccion -> <nil>
```

## Leccion 15 — Fechas: timestamp contra timestamptz

`go run ./15_tiempo` · codigo de salida `0`

```
== 15.1 · Guardo las 12:00 de Ciudad de Mexico en las dos columnas ==
  en Go:       2026-09-28 12:00 CST (-06:00)
  en UTC es:   2026-09-28 18:00 UTC

== 15.2 · Lo que vuelve a Go ==
  timestamp   -> 2026-09-28 12:00 UTC
  timestamptz -> 2026-09-28 18:00 UTC    = 12:00 CST en Mexico
  el mismo instante? false   diferencia: 6h0m0s

== 15.3 · Lo que ve alguien con psql, segun la zona de SU sesion ==
  UTC                  timestamp: 2026-09-28 12:00:00   timestamptz: 2026-09-28 18:00:00+00
  America/Mexico_City  timestamp: 2026-09-28 12:00:00   timestamptz: 2026-09-28 12:00:00-06
  America/Bogota       timestamp: 2026-09-28 12:00:00   timestamptz: 2026-09-28 13:00:00-05
```

## Leccion 16 — El proyecto: pgsalud

### Con problemas

`pgsalud -max-consulta 2s -max-idle-tx 2s` · codigo de salida `1`

```
Conexiones: 4 de 100 (4%) · base de datos: 29 MB
  Consultas lentas:        1
    pid 409     4.3 s  SELECT pg_sleep(12)
  Transacciones olvidadas: 1
    pid 408     4.3 s  UPDATE servidores SET cpu = cpu + 1 WHERE nombre = 'db-01'
  Sesiones bloqueadas:     1
    pid 410     4.3 s · la bloquea el pid 408  UPDATE servidores SET notas = 'reinicio' WHERE nombre = 'db-
level=WARN msg="consultas lentas" cuantas=1
level=WARN msg="transacciones olvidadas" cuantas=1
level=WARN msg="sesiones bloqueadas" cuantas=1
level=WARN msg="hay problemas" total=3
```

### Lo mismo, en JSON

`pgsalud -max-consulta 2s -max-idle-tx 2s -json` · codigo de salida `1`

```
{
  "conexiones": 4,
  "max_connections": 100,
  "porcentaje": 4,
  "consultas_lentas": [
    {
      "pid": 409,
      "estado": "active",
      "segundos": 4.28681,
      "consulta": "SELECT pg_sleep(12)"
    }
  ],
  "idle_in_transaction": [
    {
      "pid": 408,
      "estado": "idle in transaction",
      "segundos": 4.291725,
      "consulta": "UPDATE servidores SET cpu = cpu + 1 WHERE nombre = 'db-01'"
    }
  ],
  "bloqueadas": [
    {
      "pid": 410,
      "estado": "active",
      "segundos": 4.286612,
      "bloqueada_por": "408",
      "consulta": "UPDATE servidores SET notas = 'reinicio' WHERE nombre = 'db-"
    }
  ],
  "tamano_base": "29 MB",
  "problemas": 3
}
level=WARN msg="consultas lentas" cuantas=1
level=WARN msg="transacciones olvidadas" cuantas=1
level=WARN msg="sesiones bloqueadas" cuantas=1
level=WARN msg="hay problemas" total=3
```

### Sin problemas

`pgsalud ` · codigo de salida `0`

```
Conexiones: 1 de 100 (1%) · base de datos: 29 MB
  Consultas lentas:        ninguna
  Transacciones olvidadas: ninguna
  Sesiones bloqueadas:     ninguna
```

### Sin base de datos

`pgsalud -dsn postgres://curso:curso@127.0.0.1:1/tienda?sslmode=disable&connect_timeout=2` · codigo de salida `2`

```
level=ERROR msg="no puedo conectar" err="failed to connect to `user=curso database=tienda`: 127.0.0.1:1 (127.0.0.1): dial error: dial tcp 127.0.0.1:1: connectex: No se puede establecer una conexión ya que el equipo de destino denegó expresamente dicha conexión."
```
