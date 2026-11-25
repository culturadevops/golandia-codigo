# -*- coding: utf-8 -*-
"""banco_extra.py — el porqué de tres cifras de SALIDA.md → SALIDA-extra.md (golandia 866). No se escribe a mano.

1. Conexiones nuevas a Postgres durante 10 s de lecturas a tope (log_connections) con el pool por defecto de Go,
   con POOL=10 y con Serverpod.
2. CPU de cada servidor bajo esa carga (docker stats, 100 % = 1 núcleo; el contenedor tiene 2).
3. Las consultas que Serverpod lanza solo, sin peticiones, en 65 s.

⛔ Con el candado de medición tomado. Reutiliza las funciones de banco.py.

    python banco_extra.py
"""
import re
import threading
import time

import banco as K

LINEAS = []


def w(t=""):
    LINEAS.append(t)
    print(t, flush=True)


def cpu():
    t = K.sh(["docker", "stats", "--no-stream", "--format", "{{.CPUPerc}}", K.SRV]).stdout.strip().rstrip("%")
    return float(t)


def main():
    # el candado lo toma ESTE proceso (mkdir es atómico): si ya existe, otro está midiendo y no se toca nada
    try:
        K.CANDADO.mkdir()
    except FileExistsError:
        raise SystemExit("⛔ el candado de medición está tomado por otro: espera y repite")
    (K.CANDADO / "quien.txt").write_text("candado tomado 866 por banco_extra.py", encoding="utf-8")
    K.sh(["docker", "rm", "-f", K.SRV, K.PG])
    K.sh(["docker", "run", "-d", "--rm", "--name", K.PG, "--network", K.RED, "--cpus", "2", "--memory", "1g",
          "-e", "POSTGRES_PASSWORD=gs", "-e", "POSTGRES_DB=gs", "postgres:18.6"], check=True)
    for _ in range(60):
        if K.sh(["docker", "exec", K.PG, "pg_isready", "-U", "postgres", "-d", "gs"]).returncode == 0:
            break
        time.sleep(1)
    time.sleep(2)
    K.sh(["docker", "run", "-d", "--rm", "--name", K.SRV, "--network", K.RED, "--entrypoint", "./bin/server", "gs-serverpod:1",
          "--mode", "production", "--server-id", "default", "--logging", "normal", "--role", "monolith", "--apply-migrations"],
         check=True)
    for _ in range(120):
        if K.psql("select count(*) from information_schema.tables where table_name='tarea'") == "1":
            break
        time.sleep(1)
    time.sleep(3)
    K.sh(["docker", "rm", "-f", K.SRV])
    try:
        w("# SALIDA-extra — el porqué de tres cifras (golandia 866)\n")
        w("> Generado por `banco_extra.py`. **No se escribe a mano.** Mismo montaje que SALIDA.md.\n")
        w("## A · Conexiones nuevas a Postgres y CPU del servidor, 10 s de «leer una tarea por id» a tope (`hey -c 16`)\n")
        w("| variante | pet/s | conexiones nuevas en 10 s | CPU del servidor (100 % = 1 núcleo) |")
        w("|---|---|---|---|")
        K.psql("ALTER SYSTEM SET log_connections = 'receipt,authentication,authorization'")
        K.psql("SELECT pg_reload_conf()")
        for nombre, imagen, env, rutas in K.VARIANTES:
            K.resembrar()
            K.esperar_200(rutas, K.arrancar(imagen, env))
            K.hey(rutas, "leer", K.FRIO)
            antes = len(K.log_pg())
            muestras = []
            hilo = threading.Thread(target=lambda: [muestras.append(cpu()) or time.sleep(1) for _ in range(4)])
            hilo.start()
            rps, lat, codigos, err = K.hey(rutas, "leer", 10)
            hilo.join()
            nuevas = sum(1 for l in K.log_pg()[antes:] if "connection authorized" in l)
            w(f"| {nombre} | {rps:,.0f} | {nuevas} | {statistics_median(muestras):.0f} % (mediana de {len(muestras)}) |".replace(",", "."))
            K.sh(["docker", "rm", "-f", K.SRV])
        K.psql("ALTER SYSTEM SET log_connections = ''")
        K.psql("SELECT pg_reload_conf()")
        w("")
        w("## B · Serverpod sin ninguna petición: lo que manda a Postgres en 65 s (log_statement=all)\n")
        K.psql("ALTER SYSTEM SET log_statement='all'")
        K.psql("SELECT pg_reload_conf()")
        rutas = K.SP
        K.esperar_200(rutas, K.arrancar("gs-serverpod:1", {}))
        time.sleep(2)
        antes = len(K.log_pg())
        time.sleep(65)
        s, _ = K.sentencias(K.log_pg()[antes:])
        w("```")
        cuenta = {}
        for x in s:
            x = re.sub(r"^(execute|statement)[^:]*: ", "", x)
            clave = re.sub(r"\s+", " ", x)[:150]
            cuenta[clave] = cuenta.get(clave, 0) + 1
        for clave, n in sorted(cuenta.items(), key=lambda kv: -kv[1]):
            w(f"{n}×  {clave}")
        w(f"total: {len(s)} consultas en 65 s")
        w("```")
        K.psql("ALTER SYSTEM SET log_statement='none'")
        K.psql("SELECT pg_reload_conf()")
    finally:
        K.sh(["docker", "rm", "-f", K.SRV, K.PG])
        (K.CANDADO / "quien.txt").unlink(missing_ok=True)
        K.CANDADO.rmdir()
        (K.AQUI / "SALIDA-extra.md").write_text("\n".join(LINEAS) + "\n", encoding="utf-8", newline="\n")


def statistics_median(xs):
    xs = sorted(xs)
    return xs[len(xs) // 2] if len(xs) % 2 else (xs[len(xs) // 2 - 1] + xs[len(xs) // 2]) / 2


if __name__ == "__main__":
    main()
