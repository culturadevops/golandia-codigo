# SALIDA — Squirrel en Go: consultas SQL dinámicas sin concatenar strings (golandia 881)

> Corrida del 2026-11-18 10:55. Regenerable con `python recursos/pruebas-squirrel/correr.py`. Postgres desechable (contenedor con --rm, datos en tmpfs): las inyecciones solo tocan esa base. ⛔ Todo lo que sale en el vídeo está aquí.

## Versiones

```
go version go1.27.2 linux/amd64
sqlc v1.31.1
```

## sqlc generate

```
$ sqlc generate
(sin salida: generado)
```

## go.mod

```
module golandia/squirrel

go 1.27

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/jackc/pgx/v5 v5.13.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
```

```
$ go list -m -f '{{.Path}} {{.Version}} {{.Time}}' github.com/Masterminds/squirrel github.com/jackc/pgx/v5
github.com/Masterminds/squirrel v1.5.4 2023-03-17 12:55:57 +0000 UTC
github.com/jackc/pgx/v5 v5.13.0 2026-11-11 18:26:18 +0000 UTC
```

```
$ go vet ./...
go vet: sin avisos
```

## go run . (la salida, tal cual)

```
Postgres 18.6 · productos: 100000 filas

## 1 · El mismo buscador de tres formas, con entradas normales

### sin filtros

A Sprintf:  SELECT id, nombre, precio FROM productos WHERE 1=1 ORDER BY id LIMIT 5 OFFSET 0
            → 5 filas, ids [1 2 3 4 5]
B a mano:   SELECT id, nombre, precio FROM productos ORDER BY id LIMIT $1 OFFSET $2
            args [5 0] → 5 filas, ids [1 2 3 4 5]
C Squirrel: SELECT id, nombre, precio FROM productos ORDER BY id LIMIT 5 OFFSET 0
            args [] → 5 filas, ids [1 2 3 4 5]
¿lo mismo? sí, las tres

### categoría

A Sprintf:  SELECT id, nombre, precio FROM productos WHERE 1=1 AND categoria = 'cafe' ORDER BY id LIMIT 5 OFFSET 0
            → 5 filas, ids [5 10 15 20 25]
B a mano:   SELECT id, nombre, precio FROM productos WHERE categoria = $1 ORDER BY id LIMIT $2 OFFSET $3
            args [cafe 5 0] → 5 filas, ids [5 10 15 20 25]
C Squirrel: SELECT id, nombre, precio FROM productos WHERE categoria = $1 ORDER BY id LIMIT 5 OFFSET 0
            args [cafe] → 5 filas, ids [5 10 15 20 25]
¿lo mismo? sí, las tres

### categoría + precio + texto + orden + página 2

A Sprintf:  SELECT id, nombre, precio FROM productos WHERE 1=1 AND categoria = 'cafe' AND precio <= 10 AND nombre ILIKE '%99%' ORDER BY precio DESC, id LIMIT 5 OFFSET 5
            → 5 filas, ids [19990 24990 29990 34990 39990]
B a mano:   SELECT id, nombre, precio FROM productos WHERE categoria = $1 AND precio <= $2 AND nombre ILIKE $3 ORDER BY precio DESC, id LIMIT $4 OFFSET $5
            args [cafe 10 %99% 5 5] → 5 filas, ids [19990 24990 29990 34990 39990]
C Squirrel: SELECT id, nombre, precio FROM productos WHERE categoria = $1 AND precio <= $2 AND nombre ILIKE $3 ORDER BY precio DESC, id LIMIT 5 OFFSET 5
            args [cafe 10 %99%] → 5 filas, ids [19990 24990 29990 34990 39990]
¿lo mismo? sí, las tres


## 2 · Entrada maliciosa en «categoría»: nada' OR '1'='1

A Sprintf:  SELECT id, nombre, precio FROM productos WHERE 1=1 AND categoria = 'nada' OR '1'='1' ORDER BY id LIMIT 100000 OFFSET 0
            → 100000 filas (err <nil>)
B a mano:   SELECT id, nombre, precio FROM productos WHERE categoria = $1 ORDER BY id LIMIT $2 OFFSET $3
            args "nada' OR '1'='1" → 0 filas (err <nil>)
C Squirrel: SELECT id, nombre, precio FROM productos WHERE categoria = $1 ORDER BY id LIMIT 100000 OFFSET 0
            args "nada' OR '1'='1" → 0 filas (err <nil>)

## 3 · Inyección apilada con Sprintf: x'; DELETE FROM notas; --

SQL: SELECT id, nombre, precio FROM productos WHERE 1=1 AND categoria = 'x'; DELETE FROM notas; --' ORDER BY id LIMIT 5 OFFSET 0
notas antes: 10
pgx por defecto (protocolo extendido): err = ERROR: cannot insert multiple commands into a prepared statement (SQLSTATE 42601)
notas después: 10
pgx con QueryExecModeSimpleProtocol: err = <nil>
notas después: 0

## 4 · Squirrel sin PlaceholderFormat(sq.Dollar), contra Postgres

SQL: SELECT id, nombre, precio FROM productos WHERE categoria = ? LIMIT 3   args [cafe]
err = ERROR: syntax error at or near "LIMIT" (SQLSTATE 42601)
con psql: SELECT id FROM productos WHERE categoria = $1 LIMIT 3

## 5 · Squirrel .OrderBy(entrada del usuario): el ORDER BY no lleva placeholder

orden = "(CASE WHEN (SELECT substr(clave,1,1) FROM usuarios WHERE nombre='admin')='a' THEN precio END), id"
  SQL: SELECT id, nombre, precio FROM productos WHERE categoria = $1 ORDER BY (CASE WHEN (SELECT substr(clave,1,1) FROM usuarios WHERE nombre='admin')='a' THEN precio END), id LIMIT 5
  → 5 filas, ids [5 10 15 20 25] (err <nil>)
orden = "(CASE WHEN (SELECT substr(clave,1,1) FROM usuarios WHERE nombre='admin')='s' THEN precio END), id"
  SQL: SELECT id, nombre, precio FROM productos WHERE categoria = $1 ORDER BY (CASE WHEN (SELECT substr(clave,1,1) FROM usuarios WHERE nombre='admin')='s' THEN precio END), id LIMIT 5
  → 5 filas, ids [5000 10000 15000 20000 25000] (err <nil>)
referencia, ORDER BY precio, id → 5 filas, ids [5000 10000 15000 20000 25000]
con la lista cerrada, la misma entrada → ORDER BY id

## 6 · Listas: sq.Eq con un slice

sq.Eq{"id": [3 7 42]} → SELECT id FROM productos WHERE id IN ($1,$2,$3)   args [3 7 42]
sq.Eq{"id": []} → SELECT id FROM productos WHERE (1=0)   args []
a mano: SELECT id, nombre, precio FROM productos WHERE id = ANY($1) ORDER BY id LIMIT $2 OFFSET $3   args [[3 7 42] 10 0] → 3 filas, ids [3 7 42]

## 7 · Paginación: OFFSET contra cursor (keyset), página 3 de 5 en 5

OFFSET: SELECT id, nombre, precio FROM productos WHERE categoria = $1 ORDER BY id LIMIT 5 OFFSET 10   args [te] → 5 filas, ids [51 56 61 66 71]
cursor: SELECT id, nombre, precio FROM productos WHERE categoria = $1 AND id > $2 ORDER BY id LIMIT 5   args [te 46] → 5 filas, ids [51 56 61 66 71]
¿la misma página? true

## 8 · sqlc: la consulta fija con sqlc.narg

SQL generado (bd/query.sql.go):
-- name: Buscar :many
SELECT id, nombre, precio FROM productos
WHERE ($1::text IS NULL OR categoria = $1)
  AND ($2::numeric IS NULL OR precio <= $2)
  AND ($3::text IS NULL OR nombre ILIKE '%' || $3 || '%')
ORDER BY id
LIMIT $4

sin filtros → 100000 filas (err <nil>)
categoria=cafe → 20000 filas (err <nil>)
Squirrel, categoria=cafe → 20000 filas

## 9 · sqlc: el orden como parámetro

-- name: BuscarOrden :many
SELECT id, nombre, precio FROM productos
WHERE categoria = $1
ORDER BY $2::text
LIMIT 5

orden="precio" → [5 (46.95) 10 (42.90) 15 (38.85) 20 (34.80) 25 (30.75)] (err <nil>)
ordenado de verdad por precio → [5000 10000 15000 20000 25000]

## 10 · El plan: la consulta de sqlc con todo opcional contra la que arma Squirrel

### plan_cache_mode = force_custom_plan

sqlc · EXPLAIN (COSTS OFF) EXECUTE p('cafe', NULL, NULL, 100000)
  Limit
    ->  Sort
          Sort Key: id
          ->  Bitmap Heap Scan on productos
                Recheck Cond: (categoria = 'cafe'::text)
                ->  Bitmap Index Scan on productos_categoria
                      Index Cond: (categoria = 'cafe'::text)
Squirrel · EXPLAIN (COSTS OFF) EXECUTE p('cafe')
  Limit
    ->  Sort
          Sort Key: id
          ->  Bitmap Heap Scan on productos
                Recheck Cond: (categoria = 'cafe'::text)
                ->  Bitmap Index Scan on productos_categoria
                      Index Cond: (categoria = 'cafe'::text)
sqlc, solo precio <= 1.5 · EXPLAIN (COSTS OFF) EXECUTE p(NULL, 1.5, NULL, 100000)
  Limit
    ->  Sort
          Sort Key: id
          ->  Bitmap Heap Scan on productos
                Recheck Cond: (precio <= 1.5)
                ->  Bitmap Index Scan on productos_precio
                      Index Cond: (precio <= 1.5)
Squirrel, solo precio <= 1.5 · EXPLAIN (COSTS OFF) EXECUTE p(1.5)
  Limit
    ->  Sort
          Sort Key: id
          ->  Bitmap Heap Scan on productos
                Recheck Cond: (precio <= 1.5)
                ->  Bitmap Index Scan on productos_precio
                      Index Cond: (precio <= 1.5)

### plan_cache_mode = force_generic_plan

sqlc · EXPLAIN (COSTS OFF) EXECUTE p('cafe', NULL, NULL, 100000)
  Limit
    ->  Index Scan using productos_pkey on productos
          Filter: ((($1 IS NULL) OR (categoria = $1)) AND (($2 IS NULL) OR (precio <= $2)) AND (($3 IS NULL) OR (nombre ~~* (('%'::text || $3) || '%'::text))))
Squirrel · EXPLAIN (COSTS OFF) EXECUTE p('cafe')
  Limit
    ->  Sort
          Sort Key: id
          ->  Bitmap Heap Scan on productos
                Recheck Cond: (categoria = $1)
                ->  Bitmap Index Scan on productos_categoria
                      Index Cond: (categoria = $1)
sqlc, solo precio <= 1.5 · EXPLAIN (COSTS OFF) EXECUTE p(NULL, 1.5, NULL, 100000)
  Limit
    ->  Index Scan using productos_pkey on productos
          Filter: ((($1 IS NULL) OR (categoria = $1)) AND (($2 IS NULL) OR (precio <= $2)) AND (($3 IS NULL) OR (nombre ~~* (('%'::text || $3) || '%'::text))))
Squirrel, solo precio <= 1.5 · EXPLAIN (COSTS OFF) EXECUTE p(1.5)
  Limit
    ->  Index Scan using productos_pkey on productos
          Filter: (precio <= $1)

### plan_cache_mode = auto

(tras 6 ejecuciones) sqlc · EXPLAIN (COSTS OFF) EXECUTE p('cafe', NULL, NULL, 100000)
  EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY OFF, BUFFERS) 
  Limit (actual rows=20000.00 loops=1)
    Buffers: shared hit=1153
    ->  Index Scan using productos_pkey on productos (actual rows=20000.00 loops=1)
          Filter: ((($1 IS NULL) OR (categoria = $1)) AND (($2 IS NULL) OR (precio <= $2)) AND (($3 IS NULL) OR (nombre ~~* (('%'::text || $3) || '%'::text))))
          Rows Removed by Filter: 80000
          Index Searches: 1
          Buffers: shared hit=1153
(tras 6 ejecuciones) Squirrel · EXPLAIN (COSTS OFF) EXECUTE p('cafe')
  EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY OFF, BUFFERS) 
  Limit (actual rows=20000.00 loops=1)
    Buffers: shared hit=896
    ->  Sort (actual rows=20000.00 loops=1)
          Sort Key: id
          Sort Method: quicksort  Memory: 2175kB
          Buffers: shared hit=896
          ->  Bitmap Heap Scan on productos (actual rows=20000.00 loops=1)
                Recheck Cond: (categoria = 'cafe'::text)
                Heap Blocks: exact=878
                Buffers: shared hit=896
                ->  Bitmap Index Scan on productos_categoria (actual rows=20000.00 loops=1)
                      Index Cond: (categoria = 'cafe'::text)
                      Index Searches: 1
                      Buffers: shared hit=18
(tras 6 ejecuciones) sqlc, solo precio <= 1.5 · EXPLAIN (COSTS OFF) EXECUTE p(NULL, 1.5, NULL, 100000)
  EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY OFF, BUFFERS) 
  Limit (actual rows=1020.00 loops=1)
    Buffers: shared hit=1153
    ->  Index Scan using productos_pkey on productos (actual rows=1020.00 loops=1)
          Filter: ((($1 IS NULL) OR (categoria = $1)) AND (($2 IS NULL) OR (precio <= $2)) AND (($3 IS NULL) OR (nombre ~~* (('%'::text || $3) || '%'::text))))
          Rows Removed by Filter: 98980
          Index Searches: 1
          Buffers: shared hit=1153
(tras 6 ejecuciones) Squirrel, solo precio <= 1.5 · EXPLAIN (COSTS OFF) EXECUTE p(1.5)
  EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY OFF, BUFFERS) 
  Limit (actual rows=1020.00 loops=1)
    Buffers: shared hit=615
    ->  Sort (actual rows=1020.00 loops=1)
          Sort Key: id
          Sort Method: quicksort  Memory: 96kB
          Buffers: shared hit=615
          ->  Bitmap Heap Scan on productos (actual rows=1020.00 loops=1)
                Recheck Cond: (precio <= 1.5)
                Heap Blocks: exact=610
                Buffers: shared hit=615
                ->  Bitmap Index Scan on productos_precio (actual rows=1020.00 loops=1)
                      Index Cond: (precio <= 1.5)
                      Index Searches: 1
                      Buffers: shared hit=5
```
