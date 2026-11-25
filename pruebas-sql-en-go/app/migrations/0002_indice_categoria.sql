-- +goose Up
CREATE INDEX productos_categoria_idx ON productos (categoria);

-- +goose Down
DROP INDEX productos_categoria_idx;
