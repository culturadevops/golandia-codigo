-- +goose Up
CREATE TABLE tareas (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    usuario_id bigint      NOT NULL REFERENCES usuarios (id) ON DELETE CASCADE,
    titulo     text        NOT NULL CHECK (length(titulo) BETWEEN 1 AND 200),
    prioridad  smallint    NOT NULL DEFAULT 2 CHECK (prioridad BETWEEN 1 AND 3),
    hecha      boolean     NOT NULL DEFAULT false,
    creada     timestamptz NOT NULL DEFAULT now()
);
-- Todas las consultas filtran por usuario y paginan por id.
CREATE INDEX tareas_usuario_id ON tareas (usuario_id, id);

-- +goose Down
DROP TABLE tareas;
