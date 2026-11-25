-- Datos de la demo de «Squirrel en Go» (golandia 881). Deterministas: generate_series, sin random().
CREATE TABLE productos (
    id        bigint PRIMARY KEY,
    nombre    text    NOT NULL,
    categoria text    NOT NULL,
    precio    numeric(10,2) NOT NULL,
    stock     int     NOT NULL
);
INSERT INTO productos
SELECT i,
       'producto-' || i,
       (ARRAY['cafe','te','galletas','chocolate','pan'])[1 + i % 5],
       round(((i * 7919) % 5000) / 100.0 + 1, 2),
       (i * 31) % 40
FROM generate_series(1, 100000) AS i;
CREATE INDEX productos_categoria ON productos (categoria);
CREATE INDEX productos_precio ON productos (precio);
ANALYZE productos;

-- Para la inyección por ORDER BY: un dato que nadie debería poder leer desde el buscador.
CREATE TABLE usuarios (nombre text PRIMARY KEY, clave text NOT NULL);
INSERT INTO usuarios VALUES ('admin', 'sopa-de-letras');

-- Para la inyección apilada: una tabla de 10 filas que se puede perder (el contenedor es desechable).
CREATE TABLE notas (id int PRIMARY KEY, texto text NOT NULL);
INSERT INTO notas SELECT i, 'nota ' || i FROM generate_series(1, 10) AS i;
