-- Lo corre el superusuario UNA vez, al crear el servidor (en local lo hace docker-compose).
-- Dos usuarios: «migrador» es dueño del esquema; «app» solo lee y escribe filas.
-- Las claves son de ejemplo: en producción vienen de tu gestor de secretos.
CREATE ROLE migrador LOGIN PASSWORD 'migrador';
CREATE ROLE app      LOGIN PASSWORD 'app';

CREATE DATABASE tareas      OWNER migrador;
CREATE DATABASE tareas_test OWNER migrador;

\connect tareas
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO app;
-- Lo que cree el migrador a partir de ahora, la app lo puede leer y escribir. Nada más.
ALTER DEFAULT PRIVILEGES FOR ROLE migrador IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app;

\connect tareas_test
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO app;
ALTER DEFAULT PRIVILEGES FOR ROLE migrador IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app;
