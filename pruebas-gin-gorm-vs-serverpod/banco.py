# -*- coding: utf-8 -*-
"""banco.py — ¿Gin + GORM o Serverpod? Golandia, vídeo 866. Escribe SALIDA.md. Nada del vídeo se escribe a mano.

La MISMA API en los dos (una tarea por id, los autores con sus tareas, lo mismo con N+1 a mano, y crear una tarea)
contra el MISMO Postgres 18.6, las MISMAS tablas (las crea la migración de Serverpod) y los MISMOS datos
(1.000 autores × 10 tareas, resembrados antes de cada servidor). Servidores con 2 CPU y 1 GiB; Postgres con 2 CPU y
1 GiB; carga con `hey -c 16` y 4 CPU desde otro contenedor de la misma red.

⛔ Solo se mide con el candado tomado (mkdir <proyecto>/_logs/banco.lock): el guion lo comprueba.

    python banco.py            # todo → SALIDA.md
"""
import json
import re
import statistics
import subprocess
import sys
import threading
import time
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")
AQUI = Path(__file__).resolve().parent
CANDADO = Path(r"<proyecto>\_logs\banco.lock")
RED, PG, SRV, PUERTO = "gs-red", "gs-pg", "gs-srv", 18080
HEY = "golandia-revancha-hey:1"
C, FRIO, CALIENTE = 16, 5, 15
DSN = "host=gs-pg user=postgres password=gs dbname=gs port=5432 sslmode=disable"
LINEAS = []


def w(t=""):
    LINEAS.append(t)
    print(t, flush=True)


def sh(cmd, check=False, **kw):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)
    if check and r.returncode:
        raise SystemExit(f"[X] {' '.join(cmd)}\n{r.stdout[-800:]}\n{r.stderr[-800:]}")
    return r


def psql(sql):
    return sh(["docker", "exec", PG, "psql", "-U", "postgres", "-d", "gs", "-tAqc", sql], check=True).stdout.strip()


# ─────────────────────────── los dos servidores ───────────────────────────
GO = {
    "leer": ("GET", "/tareas/4242", None),
    "listar": ("GET", "/autores?limite=20", None),
    "n1": ("GET", "/autores-n1?limite=20", None),
    "escribir": ("POST", "/tareas", '{"titulo":"carga","autorId":7}'),
}
SP = {
    "leer": ("POST", "/api/tarea", '{"id":4242}'),
    "listar": ("POST", "/api/autores", '{"limite":20}'),
    "n1": ("POST", "/api/autoresN1", '{"limite":20}'),
    "escribir": ("POST", "/api/crear", '{"titulo":"carga","autorId":7}'),
}
VARIANTES = [
    ("Gin + GORM (pool por defecto)", "gs-go:1", {"DSN": DSN}, GO),
    ("Gin + GORM (POOL=10)", "gs-go:1", {"DSN": DSN, "POOL": "10"}, GO),
    ("Serverpod", "gs-serverpod:1", {}, SP),
]
ESCENA = {"leer": "leer una tarea por id", "listar": "20 autores con sus tareas (Preload / include)",
          "n1": "20 autores con sus tareas, N+1 a mano", "escribir": "crear una tarea (INSERT)"}


def pedir(rutas, que, host=f"http://127.0.0.1:{PUERTO}", timeout=5):
    metodo, ruta, cuerpo = rutas[que]
    req = urllib.request.Request(host + ruta, method=metodo, data=cuerpo.encode() if cuerpo else None,
                                 headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return r.status, r.read()


def arrancar(imagen, env):
    sh(["docker", "rm", "-f", SRV])
    cmd = ["docker", "run", "-d", "--rm", "--name", SRV, "--network", RED, "-p", f"{PUERTO}:8080",
           "--cpus", "2", "--memory", "1g"]
    for k, v in env.items():
        cmd += ["-e", f"{k}={v}"]
    t0 = time.perf_counter()
    sh(cmd + [imagen], check=True)
    return t0


def esperar_200(rutas, t0, tope=60):
    while time.perf_counter() - t0 < tope:
        try:
            if pedir(rutas, "leer", timeout=1)[0] == 200:
                return (time.perf_counter() - t0) * 1000
        except Exception:
            pass
        time.sleep(0.01)
    raise SystemExit("el servidor no contestó 200")


def resembrar():
    psql('TRUNCATE tarea, autor RESTART IDENTITY; '
         "INSERT INTO autor (nombre) SELECT 'autor-'||g FROM generate_series(1,1000) g; "
         'INSERT INTO tarea (titulo, hecha, "autorId") SELECT \'tarea-\'||a||\'-\'||k, k%3=0, a '
         "FROM generate_series(1,1000) a CROSS JOIN generate_series(1,10) k ORDER BY a, k; ANALYZE autor; ANALYZE tarea;")


def mem_mib():
    t = sh(["docker", "stats", "--no-stream", "--format", "{{.MemUsage}}", SRV]).stdout.strip().split("/")[0].strip()
    n = float(re.match(r"[\d.]+", t).group(0))
    return n / 1024 if t.endswith("KiB") else n * 1024 if t.endswith("GiB") else n


def hey(rutas, que, seg):
    metodo, ruta, cuerpo = rutas[que]
    cmd = ["docker", "run", "--rm", "--network", RED, "--cpus", "4", HEY, "-z", f"{seg}s", "-c", str(C), "-m", metodo]
    if cuerpo:
        cmd += ["-T", "application/json", "-d", cuerpo]
    t = sh(cmd + [f"http://{SRV}:8080{ruta}"]).stdout
    rps = float(re.search(r"Requests/sec:\s+([\d.]+)", t).group(1))
    lat = {q: float(re.search(rf"\s{q}% in ([\d.]+) secs", t).group(1)) * 1000 for q in (50, 99)}
    codigos = re.findall(r"\[(\d+)\]\s+(\d+) responses", t)
    errores = re.search(r"Error distribution:(.*)", t, re.S)
    return rps, lat, codigos, (errores.group(1).strip()[:200] if errores else "")


def limpio(x):
    if isinstance(x, dict):
        return {k: limpio(v) for k, v in x.items() if k != "__className__"}
    if isinstance(x, list):
        return [limpio(v) for v in x]
    return x


def log_pg():
    r = sh(["docker", "logs", PG])
    return (r.stdout + r.stderr).splitlines()


def sentencias(lineas):
    s = [re.sub(r"^.*LOG:  ", "", l) for l in lineas if re.search(r"LOG:  (statement|execute)", l)]
    return [x for x in s if "-- ping" not in x], [x for x in s if "-- ping" in x]


# ─────────────────────────── secciones ───────────────────────────
def versiones():
    w("## Versiones\n")
    w("```")
    w(sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip())
    mod = (AQUI / "go" / "go.mod").read_text(encoding="utf-8")
    for m in ("github.com/gin-gonic/gin", "gorm.io/gorm", "gorm.io/driver/postgres", "github.com/jackc/pgx/v5"):
        v = re.search(rf"{re.escape(m)} (v[\d.]+)", mod)
        w(f"{m} {v.group(1) if v else '?'}")
    lock = (AQUI / "gs_api" / "pubspec.lock").read_text(encoding="utf-8")
    w("serverpod " + re.search(r"\n  serverpod:\n(?:    .*\n)*?    version: \"([\d.]+)\"", lock).group(1))
    w("dart (imagen de build de Serverpod) " + sh(["docker", "run", "--rm", "dart:3.12.2", "dart", "--version"]).stdout.strip())
    w(psql("select version()"))
    w("```\n")


def contar_lineas(rutas):
    archivos = [p for r in rutas for p in (sorted(r.rglob("*")) if r.is_dir() else [r]) if p.is_file()]
    return len(archivos), sum(len(p.read_text(encoding="utf-8", errors="replace").splitlines()) for p in archivos)


def lineas():
    w("## 1 · Lo que escribes tú y lo que se genera\n")
    sp = AQUI / "gs_api" / "gs_api_server"
    yo_go = contar_lineas([AQUI / "go" / "main.go"])
    yo_sp = contar_lineas([sp / "lib/src/modelos", sp / "lib/src/api/api_endpoint.dart"])
    gen_srv = contar_lineas([sp / "lib/src/generated"])
    gen_cli = contar_lineas([AQUI / "gs_api" / "gs_api_client" / "lib"])
    proyecto = [p for p in (AQUI / "gs_api").rglob("*") if p.is_file() and ".dart_tool" not in p.parts]
    w("| | archivos | líneas |")
    w("|---|---|---|")
    w(f"| Gin + GORM, escrito a mano (go/main.go) | {yo_go[0]} | {yo_go[1]} |")
    w(f"| Serverpod, escrito a mano (2 modelos .spy.yaml + el endpoint) | {yo_sp[0]} | {yo_sp[1]} |")
    w(f"| Serverpod, generado en el servidor (lib/src/generated) | {gen_srv[0]} | {gen_srv[1]} |")
    w(f"| Serverpod, generado: el cliente de Dart (gs_api_client/lib) | {gen_cli[0]} | {gen_cli[1]} |")
    w(f"\nEl proyecto de Serverpod entero (lo crea `serverpod create --template server`, sin .dart_tool): "
      f"{len(proyecto)} archivos.\n")
    stub = (AQUI / "_stub_flutter.log").read_text(encoding="utf-8").splitlines()
    w("`serverpod create` y `serverpod generate` (CLI 4.1.2) se niegan a correr sin `flutter` en el PATH aunque el proyecto sea "
      "solo de servidor: «ERROR: Failed to run serverpod. You need to have flutter installed and in your $PATH». Con un "
      "`flutter` de mentira que solo apunta con qué lo llaman, esto es todo lo que le pidió:\n")
    w("```")
    for l in sorted(set(stub)):
        w(f"{l}  (×{stub.count(l)})")
    w("```\n")
    tablas = psql("select count(*) from information_schema.tables where table_schema='public'")
    propias = psql("select count(*) from information_schema.tables where table_schema='public' and table_name not like 'serverpod%'")
    w(f"Tablas que deja la migración de Serverpod: {tablas} ({propias} mías: autor y tarea; el resto, serverpod_*).\n")
    deps_go = sh(["docker", "run", "--rm", "-v", f"{AQUI / 'go'}:/src", "-v", "gs-gomod:/go/pkg/mod", "-w", "/src",
                  "golang:1.27.2", "go", "list", "-m", "all"]).stdout.strip().splitlines()
    lock = (AQUI / "gs_api" / "pubspec.lock").read_text(encoding="utf-8")
    deps_dart = re.findall(r"^  ([a-z0-9_]+):\n    dependency:", lock, re.M)
    w(f"Dependencias: Go `go list -m all` → {len(deps_go) - 1} módulos (sin contar el propio). "
      f"Dart pubspec.lock del workspace → {len(deps_dart)} paquetes.\n")


def construir():
    w("## 2 · Construir la imagen desde cero (`docker build --no-cache`, bases ya descargadas)\n")
    w("| | tiempo | imagen (comprimida) | binario |")
    w("|---|---|---|---|")
    for nombre, tag, df, ctx, binario in (("Gin + GORM", "gs-go:1", "Dockerfile.go", "go", "/servidor"),
                                          ("Serverpod", "gs-serverpod:1", "Dockerfile.serverpod", "gs_api", "/app/bin/server")):
        t0 = time.perf_counter()
        sh(["docker", "build", "--no-cache", "-q", "-t", tag, "-f", str(AQUI / df), str(AQUI / ctx)], check=True)
        t = time.perf_counter() - t0
        tam = int(sh(["docker", "image", "inspect", "-f", "{{.Size}}", tag]).stdout.strip())
        sh(["docker", "rm", "-f", "gs-tmp"])
        sh(["docker", "create", "--name", "gs-tmp", tag], check=True)
        sh(["docker", "cp", f"gs-tmp:{binario}", str(AQUI / "_bin_tmp")], check=True)
        b = (AQUI / "_bin_tmp").stat().st_size
        (AQUI / "_bin_tmp").unlink()
        sh(["docker", "rm", "-f", "gs-tmp"])
        w(f"| {nombre} | {t:.1f} s | {tam / 1048576:.1f} MiB | {b / 1048576:.1f} MiB |")
    w("\nGin + GORM: golang:1.27.2 → scratch (CGO_ENABLED=0, -s -w). Serverpod: el Dockerfile que genera `serverpod create` "
      "(dart:3.12.2 → alpine:3.24.4, `dart build cli`). Los dos tiempos incluyen bajar las dependencias.\n")


def mismo_resultado():
    w("## 0 · ¿Contestan lo mismo? (antes de medir)\n")
    salidas = {}
    for nombre, imagen, env, rutas in (VARIANTES[0], VARIANTES[2]):
        resembrar()
        esperar_200(rutas, arrancar(imagen, env))
        r = {}
        for que in ("leer", "listar", "n1"):
            st, cuerpo = pedir(rutas, que)
            r[que] = limpio(json.loads(cuerpo))
        st, cuerpo = pedir(rutas, "escribir")
        r["escribir"] = (st, json.loads(cuerpo), psql("select count(*) from tarea"),
                         psql("select titulo||' '||hecha||' '||\"autorId\" from tarea order by id desc limit 1"))
        salidas[nombre] = r
        sh(["docker", "rm", "-f", SRV])
    a, b = salidas[VARIANTES[0][0]], salidas["Serverpod"]
    w(f"- leer /tarea 4242 → Gin+GORM {json.dumps(a['leer'], ensure_ascii=False)}")
    w(f"                    → Serverpod {json.dumps(b['leer'], ensure_ascii=False)}")
    for que in ("leer", "listar", "n1"):
        ok = a[que] == b[que]
        tam = len(json.dumps(a[que]))
        w(f"- {ESCENA[que]}: {'IGUAL' if ok else '⛔ DISTINTO'} (JSON de {tam} bytes sin __className__)")
        if not ok:
            raise SystemExit("no contestan lo mismo: no se compara")
    w(f"- crear: Gin+GORM {a['escribir']} · Serverpod {b['escribir']}")
    if a["escribir"][2:] != b["escribir"][2:]:
        raise SystemExit("la escritura no deja lo mismo")
    st, cuerpo = None, None
    w("\nServerpod añade `\"__className__\"` a cada objeto del JSON. Sin ese campo, las respuestas son idénticas.\n")


def consultas():
    w("## 3 · Las consultas que manda cada ORM (log_statement=all en Postgres)\n")
    psql("ALTER SYSTEM SET log_statement='all'")
    psql("SELECT pg_reload_conf()")
    time.sleep(1)
    for nombre, imagen, env, rutas in (VARIANTES[0], VARIANTES[2]):
        resembrar()
        esperar_200(rutas, arrancar(imagen, env))
        pedir(rutas, "leer")
        time.sleep(0.5)
        w(f"### {nombre}\n")
        w("```")
        for que in ("leer", "listar", "n1", "escribir"):
            antes = len(log_pg())
            pedir(rutas, que)
            time.sleep(0.6)
            s, pings = sentencias(log_pg()[antes:])
            w(f"{ESCENA[que]}: {len(s)} consulta(s)" + (f" (+{len(pings)} ping)" if pings else ""))
            for x in s[:3]:
                w("   " + (x if len(x) < 230 else x[:227] + "..."))
            if len(s) > 3:
                w(f"   ... y {len(s) - 3} más iguales")
        antes = len(log_pg())
        time.sleep(65)
        s, pings = sentencias(log_pg()[antes:])
        tablas = sorted(set(re.findall(r'(?:FROM|INTO|UPDATE)\s+"?([a-z_]+)"?', " ".join(s))))
        w(f"65 s en reposo, sin ninguna petición: {len(s)} consulta(s)" + (f" en {', '.join(tablas)}" if tablas else ""))
        w("```\n")
        sh(["docker", "rm", "-f", SRV])
    psql("ALTER SYSTEM SET log_statement='none'")
    psql("SELECT pg_reload_conf()")


def arranque():
    w("## 4 · Arranque hasta el primer 200 (3 arranques, Postgres ya listo)\n")
    w("| | arranques (ms) | mediana | memoria en reposo |")
    w("|---|---|---|---|")
    for nombre, imagen, env, rutas in (VARIANTES[0], VARIANTES[2]):
        ts = []
        for _ in range(3):
            ts.append(esperar_200(rutas, arrancar(imagen, env)))
            time.sleep(5)
            m = mem_mib()
            sh(["docker", "rm", "-f", SRV])
        w(f"| {nombre} | {' · '.join(f'{t:.0f}' for t in ts)} | {statistics.median(ts):.0f} ms | {m:.1f} MiB |")
    w("")


def carga(corrida):
    w(f"## {4 + corrida} · Carga, corrida {corrida} (`hey -c {C}`, {FRIO} s de calentamiento y {CALIENTE} s medidos)\n")
    w("| variante | escena | pet/s | p50 | p99 | memoria bajo carga | códigos |")
    w("|---|---|---|---|---|---|---|")
    for nombre, imagen, env, rutas in VARIANTES:
        resembrar()
        esperar_200(rutas, arrancar(imagen, env))
        for que in ("leer", "listar", "n1", "escribir"):
            hey(rutas, que, FRIO)
            muestras = []
            hilo = threading.Thread(target=lambda: [muestras.append(mem_mib()) or time.sleep(2) for _ in range(5)])
            hilo.start()
            rps, lat, codigos, err = hey(rutas, que, CALIENTE)
            hilo.join()
            cods = " ".join(f"{c}×{n}" for c, n in codigos) + (f" · errores: {err}" if err else "")
            w(f"| {nombre} | {ESCENA[que]} | {rps:,.0f} | {lat[50]:.1f} ms | {lat[99]:.1f} ms | {max(muestras):.1f} MiB | {cods} |"
              .replace(",", "."))
        sh(["docker", "rm", "-f", SRV])
    w("")


def main():
    # el candado lo toma ESTE proceso (mkdir es atómico): si ya existe, otro está midiendo y no se toca nada
    try:
        CANDADO.mkdir()
    except FileExistsError:
        raise SystemExit("⛔ el candado de medición está tomado por otro: espera y repite")
    (CANDADO / "quien.txt").write_text("candado tomado 866 por banco.py", encoding="utf-8")
    if "gs-red" not in sh(["docker", "network", "ls", "--format", "{{.Name}}"]).stdout.split():
        sh(["docker", "network", "create", "--subnet", "10.251.77.0/24", RED], check=True)
    sh(["docker", "rm", "-f", SRV, PG])
    sh(["docker", "run", "-d", "--rm", "--name", PG, "--network", RED, "--cpus", "2", "--memory", "1g",
        "-e", "POSTGRES_PASSWORD=gs", "-e", "POSTGRES_DB=gs", "postgres:18.6"], check=True)
    for _ in range(60):
        if sh(["docker", "exec", PG, "pg_isready", "-U", "postgres", "-d", "gs"]).returncode == 0:
            break
        time.sleep(1)
    time.sleep(2)
    # las tablas las crea la migración de Serverpod (la misma para los dos servidores)
    sh(["docker", "run", "-d", "--rm", "--name", SRV, "--network", RED, "--entrypoint", "./bin/server", "gs-serverpod:1",
        "--mode", "production", "--server-id", "default", "--logging", "normal", "--role", "monolith", "--apply-migrations"],
       check=True)
    for _ in range(120):
        if psql("select count(*) from information_schema.tables where table_name='tarea'") == "1":
            break
        time.sleep(1)
    time.sleep(3)
    sh(["docker", "rm", "-f", SRV])
    try:
        sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
        w("# SALIDA — ¿Gin + GORM o Serverpod? (golandia 866)\n")
        w(f"> Generado por `banco.py` el {sello}. **No se escribe a mano.** Servidores con 2 CPU y 1 GiB; Postgres 18.6 con 2 CPU y "
          f"1 GiB; carga con `hey -c {C}` y 4 CPU desde otro contenedor. Datos: 1.000 autores × 10 tareas, resembrados antes de cada "
          "servidor.\n")
        versiones()
        mismo_resultado()
        lineas()
        construir()
        consultas()
        arranque()
        carga(1)
        carga(2)
    finally:
        sh(["docker", "rm", "-f", SRV, PG])
        (CANDADO / "quien.txt").unlink(missing_ok=True)
        CANDADO.rmdir()
        (AQUI / "SALIDA.md").write_text("\n".join(LINEAS) + "\n", encoding="utf-8", newline="\n")
    print("\n-> SALIDA.md")


if __name__ == "__main__":
    main()
