CREATE TABLE tareas (
    id     bigint  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    titulo text    NOT NULL CHECK (char_length(titulo) BETWEEN 1 AND 120),
    hecha  boolean NOT NULL DEFAULT false
);
