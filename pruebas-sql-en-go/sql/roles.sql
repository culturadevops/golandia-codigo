-- Lo corre el superusuario UNA vez, al crear el servidor. Dos bases y dos usuarios.
CREATE ROLE migrador LOGIN PASSWORD 'migrador';   -- dueño del esquema: crea y borra tablas
CREATE ROLE app      LOGIN PASSWORD 'app';        -- la aplicación: solo lee y escribe filas

CREATE DATABASE tienda      OWNER migrador;       -- la de la aplicación
CREATE DATABASE tienda_test OWNER migrador;       -- la de los tests

\connect tienda
GRANT USAGE ON SCHEMA public TO app;
ALTER DEFAULT PRIVILEGES FOR ROLE migrador IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app;

\connect tienda_test
GRANT USAGE ON SCHEMA public TO app;
ALTER DEFAULT PRIVILEGES FOR ROLE migrador IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app;
