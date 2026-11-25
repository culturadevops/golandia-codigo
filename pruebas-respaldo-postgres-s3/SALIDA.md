# SALIDA — Respaldo de Postgres a S3 con Go (golandia, vídeo 561)

> Corrida del 2026-10-21 09:49 con `bash correr.sh`. Postgres 18.6 (origen y destino), moto 5.2.4.dev como S3, Go 1.27.2 + aws-sdk-go-v2 transfermanager v0.4.17. Datos: semilla.sql (50.000 clientes, 2.000.000 de pedidos; en el paso 5, 4.000.000). «memoria pico» = VmHWM del proceso de Go (no incluye a pg_dump, que es otro proceso). Tiempos de reloj, una sola máquina: sirven para comparar entre sí, no como cifra absoluta.

```
== versiones
pg_dump (PostgreSQL) 18.6 (Debian 18.6-1.pgdg13+2)
psql (PostgreSQL) 18.6 (Debian 18.6-1.pgdg13+2)
moto 5.2.4.dev
go version go1.27.2 linux/amd64
github.com/aws/aws-sdk-go-v2 v1.47.3
github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager v0.4.17
github.com/aws/aws-sdk-go-v2/service/s3 v1.114.3
cpus de Docker: 24 · memoria: 16 GB

== datos de origen (semilla.sql)
semilla cargada en 9 s
tamaño de la base tienda: 311 MB

$ respaldo bucket
ok  bucket respaldos-demo

== 1. tres formas de comprimir (mismo origen; 3 rondas, la misma clave se sobrescribe)

-- ronda 1

$ respaldo respaldar -fecha 2026-10-21T0300 -modo fc
ok  respaldos/tienda/2026-10-21T0300.dump
    pg_dump 105.1 MB · subido 105.1 MB · partes 13 · 4.8s · memoria pico 80 MB

$ respaldo respaldar -fecha 2026-10-21T0301 -modo fc+gzip
ok  respaldos/tienda/2026-10-21T0301.dump.gz
    pg_dump 105.1 MB · subido 104.8 MB · partes 13 · 4.8s · memoria pico 81 MB

$ respaldo respaldar -fecha 2026-10-21T0302 -modo z0+gzip
ok  respaldos/tienda/2026-10-21T0302.dump.gz
    pg_dump 228.2 MB · subido 100.8 MB · partes 13 · 3s · memoria pico 66 MB

-- ronda 2

$ respaldo respaldar -fecha 2026-10-21T0300 -modo fc
ok  respaldos/tienda/2026-10-21T0300.dump
    pg_dump 105.1 MB · subido 105.1 MB · partes 13 · 4.8s · memoria pico 64 MB

$ respaldo respaldar -fecha 2026-10-21T0301 -modo fc+gzip
ok  respaldos/tienda/2026-10-21T0301.dump.gz
    pg_dump 105.1 MB · subido 104.8 MB · partes 13 · 4.8s · memoria pico 65 MB

$ respaldo respaldar -fecha 2026-10-21T0302 -modo z0+gzip
ok  respaldos/tienda/2026-10-21T0302.dump.gz
    pg_dump 228.2 MB · subido 100.8 MB · partes 13 · 3.1s · memoria pico 65 MB

-- ronda 3

$ respaldo respaldar -fecha 2026-10-21T0300 -modo fc
ok  respaldos/tienda/2026-10-21T0300.dump
    pg_dump 105.1 MB · subido 105.1 MB · partes 13 · 4.9s · memoria pico 64 MB

$ respaldo respaldar -fecha 2026-10-21T0301 -modo fc+gzip
ok  respaldos/tienda/2026-10-21T0301.dump.gz
    pg_dump 105.1 MB · subido 104.8 MB · partes 13 · 4.9s · memoria pico 66 MB

$ respaldo respaldar -fecha 2026-10-21T0302 -modo z0+gzip
ok  respaldos/tienda/2026-10-21T0302.dump.gz
    pg_dump 228.2 MB · subido 100.8 MB · partes 13 · 3.1s · memoria pico 66 MB

== 2. el ingenuo: todo el pg_dump a memoria y después PutObject

$ respaldo ingenuo -fecha 2026-10-21T0303
ok  respaldos/tienda/2026-10-21T0303.dump
    pg_dump 105.1 MB · subido de una vez · 5.1s · memoria pico 269 MB

== 3. pg_dump muere a la mitad (se corta su conexión a los 2 s)

$ (sleep 2; psql -qAt -d tienda -c "SELECT count(pg_terminate_backend(pid)) || ' conexión de pg_dump cortada' FROM pg_stat_activity WHERE application_name = 'pg_dump'") & respaldo sin-wait -fecha 2026-10-21T0400; wait
1 conexión de pg_dump cortada
ok  respaldos/tienda/2026-10-21T0400.dump.gz  (el job terminó con código 0)

$ (sleep 2; psql -qAt -d tienda -c "SELECT count(pg_terminate_backend(pid)) || ' conexión de pg_dump cortada' FROM pg_stat_activity WHERE application_name = 'pg_dump'") & respaldo respaldar -fecha 2026-10-21T0401 -modo fc+gzip; wait
1 conexión de pg_dump cortada
ERROR: subida abortada: upload multipart failed, upload id: tQvHeBMxXYQkZgBq82GHwxi38uhLEjiEgeHnWC77II9ntMx5lpk65Iw, cause: read multipart upload data failed, pg_dump: exit status 1: pg_dump: error: Dumping the contents of table "pedidos" failed: PQgetCopyData() failed.
pg_dump: detail: Error message from server: server closed the connection unexpectedly
	This probably means the server terminated abnormally
	before or while processing the request.
pg_dump: detail: Command was: COPY public.pedidos (id, cliente_id, monto, creado, nota) TO stdout;

$ respaldo listar
    respaldos/tienda/2026-10-21T0400.dump.gz         45.3 MB
    respaldos/tienda/2026-10-21T0303.dump           105.1 MB
    respaldos/tienda/2026-10-21T0302.dump.gz        100.8 MB
    respaldos/tienda/2026-10-21T0301.dump.gz        104.8 MB
    respaldos/tienda/2026-10-21T0300.dump           105.1 MB
    5 respaldos
    subidas multiparte a medias: 0

$ respaldo restaurar -clave respaldos/tienda/2026-10-21T0400.dump.gz
ERROR: pg_restore: exit status 1: pg_restore: error: could not read from input file: end of file

== 4. retención: 12 respaldos de inventario repartidos en 10 días

$ respaldo listar -base inventario
    respaldos/inventario/2026-10-21T1500.dump          27 KB
    respaldos/inventario/2026-10-21T0300.dump          27 KB
    respaldos/inventario/2026-10-20T0300.dump          27 KB
    respaldos/inventario/2026-10-19T0300.dump          27 KB
    respaldos/inventario/2026-10-18T1500.dump          27 KB
    respaldos/inventario/2026-10-18T0300.dump          27 KB
    respaldos/inventario/2026-10-17T0300.dump          27 KB
    respaldos/inventario/2026-10-16T0300.dump          27 KB
    respaldos/inventario/2026-10-15T0300.dump          27 KB
    respaldos/inventario/2026-10-14T0300.dump          27 KB
    respaldos/inventario/2026-10-13T0300.dump          27 KB
    respaldos/inventario/2026-10-12T0300.dump          27 KB
    12 respaldos
    subidas multiparte a medias: 0

$ respaldo retener -base inventario -ultimos 3 -dias 7
    queda   inventario/2026-10-21T1500.dump          (entre los 3 más nuevos)
    queda   inventario/2026-10-21T0300.dump          (entre los 3 más nuevos)
    queda   inventario/2026-10-20T0300.dump          (entre los 3 más nuevos)
    queda   inventario/2026-10-19T0300.dump          (el último del 2026-10-19)
    queda   inventario/2026-10-18T1500.dump          (el último del 2026-10-18)
    BORRA   inventario/2026-10-18T0300.dump
    queda   inventario/2026-10-17T0300.dump          (el último del 2026-10-17)
    queda   inventario/2026-10-16T0300.dump          (el último del 2026-10-16)
    queda   inventario/2026-10-15T0300.dump          (el último del 2026-10-15)
    BORRA   inventario/2026-10-14T0300.dump
    BORRA   inventario/2026-10-13T0300.dump
    BORRA   inventario/2026-10-12T0300.dump
    quedan 8 · se borran 4
    (ensayo: no se borró nada; usa -borrar)

$ respaldo retener -base inventario -ultimos 3 -dias 7 -borrar
    queda   inventario/2026-10-21T1500.dump          (entre los 3 más nuevos)
    queda   inventario/2026-10-21T0300.dump          (entre los 3 más nuevos)
    queda   inventario/2026-10-20T0300.dump          (entre los 3 más nuevos)
    queda   inventario/2026-10-19T0300.dump          (el último del 2026-10-19)
    queda   inventario/2026-10-18T1500.dump          (el último del 2026-10-18)
    BORRA   inventario/2026-10-18T0300.dump
    queda   inventario/2026-10-17T0300.dump          (el último del 2026-10-17)
    queda   inventario/2026-10-16T0300.dump          (el último del 2026-10-16)
    queda   inventario/2026-10-15T0300.dump          (el último del 2026-10-15)
    BORRA   inventario/2026-10-14T0300.dump
    BORRA   inventario/2026-10-13T0300.dump
    BORRA   inventario/2026-10-12T0300.dump
    quedan 8 · se borran 4
    borrados 4 · errores 0

$ respaldo listar -base inventario
    respaldos/inventario/2026-10-21T1500.dump          27 KB
    respaldos/inventario/2026-10-21T0300.dump          27 KB
    respaldos/inventario/2026-10-20T0300.dump          27 KB
    respaldos/inventario/2026-10-19T0300.dump          27 KB
    respaldos/inventario/2026-10-18T1500.dump          27 KB
    respaldos/inventario/2026-10-17T0300.dump          27 KB
    respaldos/inventario/2026-10-16T0300.dump          27 KB
    respaldos/inventario/2026-10-15T0300.dump          27 KB
    8 respaldos
    subidas multiparte a medias: 0

== 5. el doble de datos: ¿cuánta memoria con streaming y cuánta con el ingenuo?
pedidos: 4000000 · tamaño de la base: 610 MB

$ respaldo respaldar -fecha 2026-10-21T0300 -modo z0+gzip
ok  respaldos/tienda/2026-10-21T0300.dump.gz
    pg_dump 455.3 MB · subido 201.3 MB · partes 24 · 6s · memoria pico 66 MB

$ respaldo ingenuo -fecha 2026-10-21T0303
ok  respaldos/tienda/2026-10-21T0303.dump
    pg_dump 210.0 MB · subido de una vez · 11.1s · memoria pico 525 MB

== 6. el respaldo solo vale si restaura: a una base nueva, y se compara

$ respaldo restaurar -clave respaldos/tienda/2026-10-21T0300.dump.gz
ok  restaurado respaldos/tienda/2026-10-21T0300.dump.gz en respaldo-demo-destino/tienda_restaurada · 19.4s

$ respaldo verificar
    tabla     origen (filas · suma)          restaurada
    clientes  50000 · 1250025000             50000 · 1250025000             igual
    pedidos   4000000 · 1000542247.76        4000000 · 1000542247.76        igual
ok  el respaldo restaura y coincide

== limpieza: contenedores y red respaldo-demo-* borrados
```
