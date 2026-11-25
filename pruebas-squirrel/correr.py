# -*- coding: utf-8 -*-
"""correr.py — el banco de «Squirrel en Go» (golandia 881) → SALIDA.md.

1. Postgres 18.6 desechable (gl-881-pg: --rm, datos en tmpfs, sin volumen) y los datos de esquema.sql (100.000 productos).
2. sqlc generate (sqlc/sqlc:1.31.1) → app/bd/*.go, y el texto que genera.
3. go mod tidy + go vet + go run en golang:1.27.2, en la misma red: las cuatro formas, ejecutadas.
4. Borra el contenedor y la red. ⛔ Las inyecciones de main.go solo tocan este Postgres desechable.

Sin cronómetro: no hace falta el candado.

    python recursos/pruebas-squirrel/correr.py
"""
import subprocess
import time
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
RED, PG, MOD = "gl-881-red", "gl-881-pg", "gl-881-gomod"
URL = f"postgres://postgres:demo@{PG}:5432/postgres?sslmode=disable"
OUT = []


def sh(args, check=True):
    r = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=1200)
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(args)}\n{r.stdout}\n{r.stderr}")
    return (r.stdout + r.stderr).rstrip()


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


sh(["docker", "rm", "-f", PG], check=False)
sh(["docker", "network", "rm", RED], check=False)
sh(["docker", "network", "create", "--subnet", "10.88.1.0/24", RED])
try:
    sh(["docker", "run", "-d", "--rm", "--name", PG, "--network", RED, "--cpus", "2", "--memory", "1g", "--ulimit", "core=1",
        "--tmpfs", "/var/lib/postgresql", "-e", "POSTGRES_PASSWORD=demo", "postgres:18.6-alpine"])
    for _ in range(120):
        if subprocess.run(["docker", "exec", PG, "pg_isready", "-U", "postgres", "-h", "127.0.0.1"], capture_output=True).returncode == 0:
            break
        time.sleep(0.5)
    time.sleep(1)
    sh(["docker", "cp", str(AQUI / "app" / "esquema.sql"), f"{PG}:/tmp/esquema.sql"])
    sh(["docker", "exec", PG, "psql", "-q", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-f", "/tmp/esquema.sql"])

    w("# SALIDA — Squirrel en Go: consultas SQL dinámicas sin concatenar strings (golandia 881)\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python recursos/pruebas-squirrel/correr.py`. Postgres "
      "desechable (contenedor con --rm, datos en tmpfs): las inyecciones solo tocan esa base. ⛔ Todo lo que sale en el vídeo "
      "está aquí.\n")
    w("## Versiones\n")
    w("```")
    w(sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]))
    w("sqlc " + sh(["docker", "run", "--rm", "sqlc/sqlc:1.31.1", "version"]))
    w("```\n")

    w("## sqlc generate\n")
    w("```")
    w("$ sqlc generate")
    w(sh(["docker", "run", "--rm", "-v", f"{AQUI}:/w", "-w", "/w/sqlc", "sqlc/sqlc:1.31.1", "generate"]) or "(sin salida: generado)")
    w("```\n")

    go = ["docker", "run", "--rm", "--network", RED, "--cpus", "2", "--ulimit", "core=1", "-v", f"{AQUI / 'app'}:/app",
          "-v", f"{MOD}:/go/pkg/mod", "-w", "/app", "-e", f"DATABASE_URL={URL}", "golang:1.27.2", "bash", "-c"]
    sh(go + ["go get github.com/Masterminds/squirrel@v1.5.4 github.com/jackc/pgx/v5@v5.13.0 >/dev/null 2>&1; go mod tidy"])
    w("## go.mod\n")
    w("```")
    w((AQUI / "app" / "go.mod").read_text(encoding="utf-8").rstrip())
    w("```\n")
    w("```\n$ go list -m -f '{{.Path}} {{.Version}} {{.Time}}' github.com/Masterminds/squirrel github.com/jackc/pgx/v5")
    w(sh(go + ["go list -m -f '{{.Path}} {{.Version}} {{.Time}}' github.com/Masterminds/squirrel github.com/jackc/pgx/v5"]))
    w("```\n")
    vet = sh(go + ["go vet ./... && echo 'go vet: sin avisos'"], check=False)
    w("```\n$ go vet ./...\n" + vet + "\n```\n")
    w("## go run . (la salida, tal cual)\n")
    w("```")
    w(sh(go + ["go run ."]))
    w("```")
finally:
    sh(["docker", "rm", "-f", PG], check=False)
    sh(["docker", "network", "rm", RED], check=False)

(AQUI / "SALIDA.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
print("[ok] SALIDA.md")
