-- name: PorCategoria :many
SELECT id, nombre, precio_centavos
FROM productos
WHERE categoria = $1
ORDER BY precio_centavos
LIMIT $2;
