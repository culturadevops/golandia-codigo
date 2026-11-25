-- name: Buscar :many
-- Los filtros opcionales con sqlc: la consulta es FIJA y cada filtro se «apaga» pasando NULL.
SELECT id, nombre, precio FROM productos
WHERE (sqlc.narg('categoria')::text IS NULL OR categoria = sqlc.narg('categoria'))
  AND (sqlc.narg('precio_max')::numeric IS NULL OR precio <= sqlc.narg('precio_max'))
  AND (sqlc.narg('busca')::text IS NULL OR nombre ILIKE '%' || sqlc.narg('busca') || '%')
ORDER BY id
LIMIT sqlc.arg('limite');

-- name: BuscarOrden :many
-- El orden como parámetro: compila, se ejecuta… y no ordena por la columna.
SELECT id, nombre, precio FROM productos
WHERE categoria = sqlc.arg('categoria')
ORDER BY sqlc.arg('orden')::text
LIMIT 5;
