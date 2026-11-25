# -*- coding: utf-8 -*-
"""Demo del vídeo 684 de Golandia: «¿SQL dentro de Go o en archivos .sql?» (respuesta a @Under-War.ht2).

    python correr.py            todo (deja salidas/NN-*.txt; SALIDA.md se arma a partir de ellas)
    python correr.py --limpiar  borra contenedor, red y volúmenes g684-*

Solo crea y borra cosas con el prefijo g684 (contenedor g684-pg, red g684, volúmenes g684-*).
El cambio de columna se hace en una COPIA de app/ dentro de un directorio temporal (tempfile.mkdtemp).
Versiones fijas: postgres:18.6-alpine · golang:1.27.2 · sqlc/sqlc:1.31.1 · goose v3.29.0 · pgx v5.11.0.
"""
import os
import shutil
import subprocess
import sys
import tempfile
import time
from pathlib import Path

AQUI = Path(__file__).resolve().parent
SAL = AQUI / "salidas"
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
PG = "postgres:18.6-alpine"
GO = "golang:1.27.2"
SQLC = "sqlc/sqlc:1.31.1"
GOOSE = "v3.29.0"
RED_ = "g684"
CPG = "g684-pg"
MIG = f"postgres://migrador:migrador@{CPG}:5432/tienda?sslmode=disable"
APPDSN = f"postgres://app:app@{CPG}:5432/tienda?sslmode=disable"
TESTDSN = f"postgres://migrador:migrador@{CPG}:5432/tienda_test?sslmode=disable"


def win(p):
    return str(p)


def sh(*a, check=False, entrada=None):
    r = subprocess.run(list(a), capture_output=True, text=True, env=ENV, encoding="utf-8", errors="replace", input=entrada)
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(a)}:\n{r.stdout}{r.stderr}")
    return r


def golang(script, app=AQUI / "app", env=()):
    a = ["docker", "run", "--rm", "--network", RED_, "-v", f"{win(app)}:/src", "-v", "g684-gomod:/go/pkg/mod",
         "-v", "g684-gocache:/root/.cache/go-build", "-v", "g684-gobin:/gobin", "-e", "PATH=/gobin:/usr/local/go/bin:/usr/bin:/bin",
         "-w", "/src"]
    for e in env:
        a += ["-e", e]
    return sh(*a, GO, "sh", "-c", script)


def psql(usuario, base, sql, clave=None):
    clave = clave or usuario
    return sh("docker", "exec", "-e", f"PGPASSWORD={clave}", CPG, "psql", "-h", "127.0.0.1", "-U", usuario, "-d", base,
              "-v", "ON_ERROR_STOP=0", "-c", sql)


class Salida:
    def __init__(self, nombre):
        self.f = (SAL / f"{nombre}.txt").open("w", encoding="utf-8", newline="\n")

    def cmd(self, texto, r=None):
        self.f.write(f"$ {texto}\n")
        if r is not None:
            out = (r.stdout + r.stderr).rstrip()
            if out:
                self.f.write(out + "\n")
            self.f.write(f"[código de salida {r.returncode}]\n")
        self.f.flush()

    def nota(self, texto):
        self.f.write(texto + "\n")

    def close(self):
        self.f.close()


def limpiar():
    sh("docker", "rm", "-f", CPG)
    sh("docker", "network", "rm", RED_)


def preparar():
    limpiar()
    sh("docker", "network", "create", RED_, check=True)
    sh("docker", "run", "-d", "--name", CPG, "--network", RED_, "-e", "POSTGRES_PASSWORD=super", PG, check=True)
    for _ in range(60):
        if sh("docker", "exec", CPG, "pg_isready", "-h", "127.0.0.1", "-U", "postgres").returncode == 0:
            break
        time.sleep(1)
    time.sleep(1)
    roles = (AQUI / "sql" / "roles.sql").read_text(encoding="utf-8")
    r = sh("docker", "exec", "-i", "-e", "PGPASSWORD=super", CPG, "psql", "-h", "127.0.0.1", "-U", "postgres", "-v", "ON_ERROR_STOP=1",
           "-f", "-", entrada=roles, check=True)
    s = Salida("00-roles")
    s.cmd("psql -U postgres -f sql/roles.sql", r)
    s.close()
    # goose (CLI) con versión fija, sin los drivers que no se usan
    r = golang(f"test -x /gobin/goose || GOBIN=/gobin go install -tags 'no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 "
               f"no_vertica no_ydb' github.com/pressly/goose/v3/cmd/goose@{GOOSE}; goose --version")
    if r.returncode != 0:
        raise SystemExit("no pude instalar goose:\n" + r.stdout + r.stderr)


def versiones():
    s = Salida("01-versiones")
    s.cmd("docker version --format '{{.Server.Version}}'", sh("docker", "version", "--format", "{{.Server.Version}}"))
    s.cmd("SELECT version()", psql("postgres", "postgres", "SELECT version();", "super"))
    s.cmd("go version", golang("go version"))
    s.cmd("goose --version", golang("goose --version"))
    s.cmd("sqlc version", sh("docker", "run", "--rm", SQLC, "version"))
    s.cmd("go list -m all | grep -E 'pgx|goose'", golang("go list -m all | grep -E 'jackc/pgx|pressly/goose'"))
    s.close()


def bases():
    s = Salida("02-bases")
    s.cmd("\\l  (las bases)", psql("postgres", "postgres", "SELECT datname AS base, pg_get_userbyid(datdba) AS duenio FROM pg_database "
                                                           "WHERE datname LIKE 'tienda%' ORDER BY 1;", "super"))
    s.close()


def migraciones():
    s = Salida("03-migraciones")
    g = 'goose -dir migrations postgres "$MIG"'
    for paso in ("status", "up", "status"):
        s.cmd(f"{g} {paso}", golang(f"{g} {paso}", env=[f"MIG={MIG}"]))
    s.cmd("SELECT * FROM goose_db_version", psql("migrador", "tienda", "SELECT id, version_id, is_applied FROM goose_db_version ORDER BY id;"))
    s.cmd(f"{g} down", golang(f"{g} down", env=[f"MIG={MIG}"]))
    s.cmd(f"{g} status", golang(f"{g} status", env=[f"MIG={MIG}"]))
    s.cmd("\\d productos (tras el down: sigue la tabla, sin el índice)", psql("migrador", "tienda", "\\d productos"))
    s.cmd(f"{g} up", golang(f"{g} up", env=[f"MIG={MIG}"]))
    s.cmd("SELECT * FROM goose_db_version", psql("migrador", "tienda", "SELECT id, version_id, is_applied FROM goose_db_version ORDER BY id;"))
    s.close()


def permisos():
    s = Salida("04-permisos")
    semilla = (AQUI / "sql" / "semilla.sql").read_text(encoding="utf-8")
    s.cmd("psql -U app -d tienda -f sql/semilla.sql", psql("app", "tienda", semilla))
    s.cmd("psql -U app: SELECT count(*) FROM productos", psql("app", "tienda", "SELECT count(*) FROM productos;"))
    s.cmd("psql -U app: DROP TABLE productos", psql("app", "tienda", "DROP TABLE productos;"))
    s.cmd("psql -U app: CREATE TABLE copia (id int)", psql("app", "tienda", "CREATE TABLE copia (id int);"))
    s.cmd("psql -U app: TRUNCATE productos", psql("app", "tienda", "TRUNCATE productos;"))
    s.cmd("psql -U app: SELECT count(*) FROM productos", psql("app", "tienda", "SELECT count(*) FROM productos;"))
    s.close()


def tres():
    s = Salida("05-tres-formas")
    s.cmd("DATABASE_URL=postgres://app:…@g684-pg/tienda go run ./cmd/tres", golang("go run ./cmd/tres", env=[f"DATABASE_URL={APPDSN}"]))
    s.cmd("go build -trimpath -ldflags='-s -w' -o /tmp/tres ./cmd/tres && stat -c '%s bytes' /tmp/tres",
          golang("go build -trimpath -ldflags='-s -w' -o /tmp/tres ./cmd/tres && stat -c '%s bytes' /tmp/tres"))
    s.close()


def cambio_columna():
    """Renombro nombre → titulo en una COPIA de app/ (directorio temporal) y en una base aparte (tienda_cambio)."""
    tmp = Path(tempfile.mkdtemp(prefix="g684-"))
    try:
        copia = tmp / "app"
        shutil.copytree(AQUI / "app", copia)
        (copia / "migrations" / "0003_renombrar_nombre.sql").write_text(
            "-- +goose Up\nALTER TABLE productos RENAME COLUMN nombre TO titulo;\n\n"
            "-- +goose Down\nALTER TABLE productos RENAME COLUMN titulo TO nombre;\n", encoding="utf-8", newline="\n")
        s = Salida("06-cambio-columna")
        s.nota("# copia de app/ en un directorio temporal; migración nueva 0003: nombre → titulo")
        s.nota((copia / "migrations" / "0003_renombrar_nombre.sql").read_text(encoding="utf-8").rstrip())
        psql("postgres", "postgres", "DROP DATABASE IF EXISTS tienda_cambio;", "super")
        psql("postgres", "postgres", "CREATE DATABASE tienda_cambio OWNER migrador;", "super")
        dsn = MIG.replace("/tienda?", "/tienda_cambio?")
        g = 'goose -dir migrations postgres "$MIG"'
        s.cmd(f"{g} up", golang(f"{g} up", app=copia, env=[f"MIG={dsn}"]))
        psql("migrador", "tienda_cambio", (AQUI / "sql" / "semilla.sql").read_text(encoding="utf-8").replace("nombre", "titulo"))
        s.nota("\n# 1) sin regenerar nada: el código viejo COMPILA y falla al ejecutarse")
        s.cmd("go build ./... && go run ./cmd/tres", golang("go build ./... && echo 'compila: OK' && go run ./cmd/tres", app=copia,
                                                            env=[f"DATABASE_URL={dsn}"]))
        psql("postgres", "postgres", "DROP DATABASE IF EXISTS tienda_cambio_test;", "super")
        psql("postgres", "postgres", "CREATE DATABASE tienda_cambio_test OWNER migrador;", "super")
        s.nota("\n# 1b) el test de integración, contra una base de tests con las migraciones nuevas")
        s.cmd("TEST_DATABASE_URL=…/tienda_cambio_test go test -count=1 ./consultas/",
              golang("go test -count=1 ./consultas/", app=copia, env=[f"TEST_DATABASE_URL={dsn.replace('tienda_cambio', 'tienda_cambio_test')}"]))
        s.nota("\n# 2) sqlc generate con la consulta vieja")
        s.cmd("sqlc generate", sh("docker", "run", "--rm", "-v", f"{win(copia)}:/src", "-w", "/src", SQLC, "generate"))
        q = (copia / "query.sql").read_text(encoding="utf-8").replace("SELECT id, nombre,", "SELECT id, titulo,")
        (copia / "query.sql").write_text(q, encoding="utf-8", newline="\n")
        s.nota("\n# 3) arreglo query.sql (nombre → titulo) y regenero")
        s.cmd("sqlc generate", sh("docker", "run", "--rm", "-v", f"{win(copia)}:/src", "-w", "/src", SQLC, "generate"))
        s.cmd("go build ./...", golang("go build ./...", app=copia))
        s.close()
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
        psql("postgres", "postgres", "DROP DATABASE IF EXISTS tienda_cambio;", "super")
        psql("postgres", "postgres", "DROP DATABASE IF EXISTS tienda_cambio_test;", "super")


def filtros():
    s = Salida("07-filtros")
    s.cmd("go run ./cmd/filtros", golang("go run ./cmd/filtros", env=[f"DATABASE_URL={APPDSN}"]))
    s.close()


def pool():
    s = Salida("08-pool")
    golang("go build -o /gobin/pool684 ./cmd/pool")
    for modo in ("olvido", "todo", "ctx", "bien"):
        t = time.time()
        r = golang(f"pool684 {modo}", env=[f"DATABASE_URL={APPDSN}"])
        s.cmd(f"./pool {modo}   (go build ./cmd/pool)", r)
        s.nota(f"(reloj del contenedor: {time.time() - t:.1f} s)")
    s.cmd("SELECT count(*) FROM pg_stat_activity WHERE usename = 'app'", psql("postgres", "postgres",
          "SELECT count(*) FROM pg_stat_activity WHERE usename = 'app';", "super"))
    s.close()


def arranque():
    s = Salida("09-arranque")
    s.cmd("go run ./cmd/tres   (sin DATABASE_URL)", golang("go run ./cmd/tres"))
    s.cmd("DATABASE_URL=postgres://app:MAL@g684-pg/tienda go run ./cmd/tres",
          golang("go run ./cmd/tres", env=[f"DATABASE_URL={APPDSN.replace('app:app', 'app:MAL')}"]))
    s.cmd("DATABASE_URL=postgres://app:app@g684-pg:5433/tienda go run ./cmd/tres   (puerto equivocado)",
          golang("go run ./cmd/tres", env=[f"DATABASE_URL={APPDSN.replace(':5432', ':5433')}"]))
    s.close()


def tests():
    s = Salida("10-test-integracion")
    s.cmd("go test -count=1 ./consultas/   (sin TEST_DATABASE_URL)", golang("go test -count=1 -v ./consultas/"))
    s.cmd("TEST_DATABASE_URL=postgres://migrador:…@g684-pg/tienda_test go test -count=1 -v ./consultas/",
          golang("go test -count=1 -v ./consultas/", env=[f"TEST_DATABASE_URL={TESTDSN}"]))
    s.cmd("psql -U app -d tienda: SELECT count(*) FROM productos   (la base de la app, intacta)",
          psql("app", "tienda", "SELECT count(*) FROM productos;"))
    s.close()


TITULOS = {
    "00-roles": "0 · Preparar: dos bases (app y tests) y dos usuarios (migrador y app)",
    "01-versiones": "1 · Versiones exactas",
    "02-bases": "2 · ¿Una base separada? Una para la app, otra para los tests",
    "03-migraciones": "3 · Migraciones en archivos .sql con goose: status, up, tabla de versiones, down",
    "04-permisos": "4 · El usuario de la app no puede tocar el esquema",
    "05-tres-formas": "5 · La misma consulta de tres formas: string, go:embed y sqlc",
    "06-cambio-columna": "6 · Cambio una columna (nombre → titulo): quién avisa y cuándo",
    "07-filtros": "7 · Consulta dinámica: estructura en Go, valores como $n; y la inyección",
    "08-pool": "8 · La trampa de rows.Close() con un pool de 4 conexiones",
    "09-arranque": "9 · DSN por variable de entorno y Ping al arrancar",
    "10-test-integracion": "10 · Test de integración contra la base de tests",
}


def armar_salida():
    partes = ["# SALIDA — ¿SQL dentro de Go o en archivos .sql? (golandia 684)\n",
              f"Generado por `python correr.py` el {time.strftime('%Y-%m-%d %H:%M UTC', time.gmtime())}. **No se edita a mano**: "
              "cada bloque es el contenido de `salidas/<n>.txt`.\n",
              "Postgres 18.6 (postgres:18.6-alpine) · Go 1.27.2 (golang:1.27.2) · sqlc 1.31.1 (sqlc/sqlc:1.31.1) · goose v3.29.0 · "
              "pgx v5.11.0 por database/sql · Docker 29.6.2. Todo en una red de Docker `g684`; el host de la base es `g684-pg`.\n"]
    for f in sorted(SAL.glob("*.txt")):
        partes.append(f"## {TITULOS.get(f.stem, f.stem)}\n\n```\n{f.read_text(encoding='utf-8').rstrip()}\n```\n")
    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8", newline="\n")


def main():
    if "--salida" in sys.argv:
        armar_salida()
        return
    if "--limpiar" in sys.argv:
        limpiar()
        for v in ("g684-gomod", "g684-gocache", "g684-gobin"):
            sh("docker", "volume", "rm", "-f", v)
        return
    SAL.mkdir(exist_ok=True)
    preparar()
    try:
        for paso in (versiones, bases, migraciones, permisos, tres, cambio_columna, filtros, pool, arranque, tests):
            t = time.time()
            paso()
            print(f"{paso.__name__}: {time.time() - t:.1f} s", flush=True)
    finally:
        limpiar()
    armar_salida()


if __name__ == "__main__":
    main()
