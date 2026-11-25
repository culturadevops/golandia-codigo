-- +goose Up
CREATE TABLE usuarios (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      text        NOT NULL,
    clave_hash text        NOT NULL,
    creado     timestamptz NOT NULL DEFAULT now()
);
-- Único sin mayúsculas: «Ana@x.com» y «ana@x.com» son la misma cuenta.
CREATE UNIQUE INDEX usuarios_email ON usuarios (lower(email));

CREATE TABLE sesiones (
    -- Se guarda el SHA-256 del token, no el token: quien lea la tabla no puede entrar.
    token_hash bytea       PRIMARY KEY,
    usuario_id bigint      NOT NULL REFERENCES usuarios (id) ON DELETE CASCADE,
    vence      timestamptz NOT NULL,
    creada     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX sesiones_usuario ON sesiones (usuario_id);

-- +goose Down
DROP TABLE sesiones;
DROP TABLE usuarios;
