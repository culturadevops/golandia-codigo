-- Datos de la demo. «tienda»: 50.000 clientes y 2.000.000 de pedidos (lo que se respalda de verdad).
-- «inventario»: una base chica para la demo de retención (12 respaldos sin llenar la memoria de moto).
CREATE DATABASE tienda;
CREATE DATABASE inventario;
\c tienda
CREATE TABLE clientes (id int PRIMARY KEY, nombre text NOT NULL, correo text NOT NULL);
CREATE TABLE pedidos (id bigint PRIMARY KEY, cliente_id int NOT NULL REFERENCES clientes(id),
                      monto numeric(10,2) NOT NULL, creado timestamptz NOT NULL, nota text NOT NULL);
SELECT setseed(0.42);
INSERT INTO clientes SELECT g, 'cliente ' || g, 'c' || g || '@ejemplo.com' FROM generate_series(1, 50000) g;
INSERT INTO pedidos SELECT g, 1 + (g % 50000), round((random() * 500)::numeric, 2),
       timestamptz '2026-01-01' + g * interval '7 seconds', md5(g::text) || md5((g * 7)::text)
  FROM generate_series(1, 2000000) g;
\c inventario
CREATE TABLE productos (id int PRIMARY KEY, nombre text, stock int);
INSERT INTO productos SELECT g, 'producto ' || g, g % 90 FROM generate_series(1, 5000) g;
