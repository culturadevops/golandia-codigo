-- +goose Up
CREATE TABLE productos (
    id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nombre          text    NOT NULL,
    categoria       text    NOT NULL,
    precio_centavos integer NOT NULL CHECK (precio_centavos >= 0)
);

-- +goose Down
DROP TABLE productos;
