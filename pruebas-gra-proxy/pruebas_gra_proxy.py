# -*- coding: utf-8 -*-
"""pruebas_gra_proxy.py — golandia 961, «proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust)».

Pingora (Cloudflare) es una LIBRERÍA: no publica un proxy listo para usar. River, el proxy que la ISRG anunció sobre
Pingora, no publica nada desde v0.5.0 (2024-08-30). Por eso, del lado de Rust: Pingap (proxy binario construido sobre
Pingora) y Sōzu (el proxy en Rust de Clever Cloud, en producción desde hace años). Del lado de Go: Caddy y Traefik.
Imágenes OFICIALES: caddy:2.11.9, traefik:v3.8.0, vicanso/pingap:0.16.2, clevercloud/sozu:2.3.7.

Banco: red gl-gra-red (10.95.0.0/24). El MISMO backend (nginx 1.30.6, 4 CPU, 10.95.0.10) para todos; el proxy en
10.95.0.20 con --cpus 2 --memory 1g; el MISMO cliente (hey, 6 CPU, -host banco.local). Antes de medir, la misma respuesta
a través de cada proxy (cuerpo de /chico idéntico y md5 de /grande idéntico); si no, se para. Muestreo del cgroup del
proxy cada 0,5 s (CPU y memoria, la cuenta de docker stats). Línea base: hey directo al backend.

    python investigacion/pruebas_gra_proxy.py <corrida> [--humo]  → recursos/pruebas-gra-proxy/SALIDA-corrida-<N>.md
"""
import hashlib
import re
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_gra_comun as K  # noqa: E402
import muestreo as MU  # noqa: E402

BASE = K.REC / "pruebas-gra-proxy"
BANCO_W = str(BASE).replace("/", "\\")
RED, SUBRED = f"{K.P}-red", "10.95.0.0/24"
IP_B, IP_P = "10.95.0.10", "10.95.0.20"
HUMO = "--humo" in sys.argv
SEG = 5 if HUMO else 20

PROXIES = {
    "caddy": dict(img="caddy:2.11.9", lenguaje="Go", conf=["Caddyfile"],
                  args=["-v", f"{BANCO_W}\\Caddyfile:/etc/caddy/Caddyfile:ro"], cmd=[]),
    "traefik": dict(img="traefik:v3.8.0", lenguaje="Go", conf=["traefik.yml", "traefik-dinamico.yml"],
                    args=["-v", f"{BANCO_W}:/conf:ro"], cmd=["traefik", "--configFile=/conf/traefik.yml"]),
    "pingap": dict(img="vicanso/pingap:0.16.2", lenguaje="Rust", conf=["pingap.toml"],
                   args=["-v", f"{BANCO_W}\\pingap.toml:/conf/pingap.toml:ro"], cmd=["pingap", "-c", "/conf/pingap.toml"]),
    "sozu": dict(img="clevercloud/sozu:2.3.7", lenguaje="Rust", conf=["sozu.toml"],
                 args=["-v", f"{BANCO_W}\\sozu.toml:/conf/sozu.toml:ro"], cmd=["start", "-c", "/conf/sozu.toml"]),
}
CARGAS = [("/chico", 64), ("/chico", 256), ("/grande", 64)]
EXTRA = {"caddy-ka": dict(img="caddy:2.11.9", lenguaje="Go", conf=["Caddyfile-ka"],
                          args=["-v", f"{BANCO_W}\\Caddyfile-ka:/etc/caddy/Caddyfile:ro"], cmd=[])}


def atendidas():
    """Peticiones que el backend lleva atendidas (stub_status de nginx), pedidas sin pasar por el proxy."""
    code, b = curl(f"http://{IP_B}/estado")
    m = re.search(rb"accepts handled requests\s+(\d+)\s+\d+\s+(\d+)", b)
    CONEX[0] = int(m.group(1)) if m else -1
    return int(m.group(2)) if m else -1


CONEX = [0]   # conexiones que nginx lleva aceptadas (la última lectura de atendidas())


class Muestreo(MU.Muestreo):
    """El de la serie, con el contenedor auxiliar llamado gl-gra-muestreo-… (regla del prefijo)."""

    def iniciar(self):
        r = subprocess.run(["docker", "inspect", "-f", "{{.Id}}", self.contenedor], capture_output=True, text=True)
        cid = r.stdout.strip()
        self._nombre = f"{K.P}-muestreo-{int(time.time() * 1000) % 100000}"
        args = ["docker", "run", "--rm", "--ulimit", "core=1", "--name", self._nombre, "--cgroupns=host", "--cpus", "0.25",
                "-v", "/sys/fs/cgroup:/cg:ro", MU.IMAGEN, "sh", "-c", MU._BUCLE, "muestreo", cid, str(self.intervalo), ""]
        self._proc = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, bufsize=1,
                                      env=MU._sin_conversion_de_rutas())
        import threading
        self._hilo = threading.Thread(target=self._leer, daemon=True)
        self._hilo.start()
        fin = time.perf_counter() + 15
        while not self.crudo and time.perf_counter() < fin and self._proc.poll() is None:
            time.sleep(0.02)
        if not self.crudo:
            raise RuntimeError(f"muestreo sin muestras ({self.errores[:3]})")
        return self


def red_y_backend():
    K.rm(f"{K.P}-backend", f"{K.P}-proxy")
    K.sh(["docker", "network", "rm", RED], timeout=60)
    K.sh(["docker", "network", "create", "--subnet", SUBRED, RED], timeout=60, check=True)
    K.sh(["docker", "run", "-d", "--name", f"{K.P}-backend", "--network", RED, "--ip", IP_B, "--cpus", "4", "--memory", "1g",
          "-v", f"{BANCO_W}\\nginx-backend.conf:/etc/nginx/nginx.conf:ro", "nginx:1.30.6-alpine"], timeout=120, check=True)
    # /grande: 100 KiB deterministas
    K.sh(["docker", "exec", f"{K.P}-backend", "sh", "-c",
          "mkdir -p /srv && i=0; : > /srv/grande; while [ $i -lt 1600 ]; do printf '%063d\\n' $i >> /srv/grande; "
          "i=$((i+1)); done; wc -c /srv/grande"], timeout=120, check=True)


def curl(url, archivo=None):
    a = ["docker", "run", "--name", f"{K.P}-curl", "--network", RED, "curlimages/curl:8.17.0", "-s", "-H", "Host: banco.local",
         "-o", "-", "-w", "\n__CODE__%{http_code}", url]
    r = subprocess.run(["docker", "run", "--rm", "--ulimit", "core=1"] + a[2:], capture_output=True, timeout=60)
    cuerpo, _, code = r.stdout.rpartition(b"\n__CODE__")
    return code.decode(), cuerpo


def arrancar(nombre):
    p = PROXIES.get(nombre) or EXTRA[nombre]
    K.rm(f"{K.P}-proxy")
    K.sh(["docker", "run", "-d", "--name", f"{K.P}-proxy", "--network", RED, "--ip", IP_P, "--cpus", "2", "--memory", "1g",
          "--memory-swap", "1g"] + p["args"] + [p["img"]] + p["cmd"], timeout=120, check=True)
    t0 = time.time()
    while time.time() - t0 < 30:
        code, _ = curl(f"http://{IP_P}:8080/chico")
        if code == "200":
            return time.time() - t0
        time.sleep(0.3)
    logs = K.sh(["docker", "logs", f"{K.P}-proxy"]).stdout[-1500:]
    raise SystemExit(f"⛔ {nombre} no respondió 200 en 30 s\n{logs}")


def _ms(v, u):
    return float(v) * {"us": 0.001, "ms": 1.0, "s": 1000.0}[u]


def hey(url, c, seg):
    """La carga, con wrk (6 hilos, 6 CPU, c conexiones con keep-alive). hey se dejó: guarda como mucho 1.000.000 de
    resultados y con más de 50.000 pet/s durante 20 s se queda corto (la línea base dio exactamente 1000000)."""
    r = K.sh(["docker", "run", "--name", f"{K.P}-wrk", "--network", RED, "--cpus", "6", "gl-gra-wrk:1", "-t", "6",
              "-c", str(c), "-d", f"{seg}s", "--timeout", "10s", "--latency", "-H", "Host: banco.local", url],
             timeout=seg + 120)
    t = r.stdout
    rps = re.search(r"Requests/sec:\s+([\d.]+)", t)
    p50 = re.search(r"\s50%\s+([\d.]+)(us|ms|s)", t)
    p99 = re.search(r"\s99%\s+([\d.]+)(us|ms|s)", t)
    tot = re.search(r"(\d+) requests in ([\d.]+)(m|s)", t)
    no2 = re.search(r"Non-2xx or 3xx responses: (\d+)", t)
    sock = re.search(r"Socket errors: (.*)", t)
    total = int(tot.group(1)) if tot else 0
    malas = int(no2.group(1)) if no2 else 0
    return dict(rps=float(rps.group(1)) if rps else 0.0, p50=_ms(*p50.groups()) if p50 else None,
                p99=_ms(*p99.groups()) if p99 else None, total=total, ok=total - malas,
                codes={"2xx": total - malas, "no-2xx": malas}, errores=sock.group(1).strip() if sock else "",
                crudo=t)


def main():
    n = int(sys.argv[1])
    out = K.Salida(BASE / (f"SALIDA-corrida-{n}.md" if not HUMO else "_humo.md"))
    orden = list(PROXIES)
    if n % 2 == 0:
        orden.reverse()
    with K.candado(f"proxy corrida {n}{' humo' if HUMO else ''}"):
        out.w(f"# Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust) · corrida {n}{' (HUMO)' if HUMO else ''}\n")
        out.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado `_logs/banco.lock` tomado y sin `fab producir` "
              f"vivo. Regenerable con `python investigacion/pruebas_gra_proxy.py {n}`. Orden: {' → '.join(orden)}. "
              f"Cada carga: wrk -t6 -d{SEG}s.\n")
        red_y_backend()
        out.w("## 0 · Versiones, imágenes y configuración\n")
        vers = {"caddy": ["caddy", "version"], "traefik": ["traefik", "version"], "pingap": ["pingap", "--version"],
                "sozu": ["--version"]}
        for k, p in PROXIES.items():
            ent = ["--entrypoint", ""] if k != "sozu" else []
            v = K.sh(["docker", "run", "--name", f"{K.P}-ver"] + ent + [p["img"]] + vers[k], timeout=60)
            vt = " ".join((v.stdout + v.stderr).split())[:160]
            tam = int(K.sh(["docker", "image", "inspect", p["img"], "--format", "{{.Size}}"]).stdout.strip()) / 2**20
            lin = sum(len([l for l in (BASE / f).read_text(encoding="utf-8").splitlines()
                           if l.strip() and not l.strip().startswith("#")]) for f in p["conf"])
            out.w(f"- **{k}** ({p['lenguaje']}) `{p['img']}` · imagen {tam:.1f} MiB · config {lin} líneas ({', '.join(p['conf'])})"
                  f" · `{vt}`")
        out.w(f"- backend nginx:1.30.6-alpine (4 CPU) · cliente wrk (Alpine 3.24, 6 hilos, 6 CPU) · ajenos vivos: {K.ajenos()} · CPU de Windows: "
              f"{K.cpu_windows()}\n")

        out.w("## 1 · ¿La misma respuesta a través de cada uno?\n")
        code, ch = curl(f"http://{IP_B}/chico")
        code2, gr = curl(f"http://{IP_B}/grande")
        ref = (ch, hashlib.md5(gr).hexdigest(), len(gr))
        out.w(f"- directo al backend: /chico {code} {len(ch)} bytes · /grande {code2} {len(gr)} bytes md5 {ref[1][:12]}")
        for k in orden:
            t = arrancar(k)
            c1, b1 = curl(f"http://{IP_P}:8080/chico")
            c2, b2 = curl(f"http://{IP_P}:8080/grande")
            ok = c1 == "200" and c2 == "200" and b1 == ref[0] and hashlib.md5(b2).hexdigest() == ref[1]
            out.w(f"- {k}: /chico {c1} {len(b1)} bytes · /grande {c2} {len(b2)} bytes md5 {hashlib.md5(b2).hexdigest()[:12]} · "
                  f"{'✅ igual' if ok else '⛔ DISTINTO'} (respondió a los {t:.1f} s de arrancar)")
            if not ok:
                raise SystemExit(1)
        K.rm(f"{K.P}-proxy")
        out.w("")

        tabla = {}
        out.w("## 2 · Línea base: wrk directo al backend (sin proxy)\n")
        for ruta, c in CARGAS:
            h = hey(f"http://{IP_B}{ruta}", c, SEG)
            out.w(f"- {ruta} c={c}: {h['rps']:.0f} pet/s · p50 {h['p50']:.2f} ms · p99 {h['p99']:.2f} ms · códigos {h['codes']}"
                  f"{' · errores de socket: ' + h['errores'] if h['errores'] else ''}")
        out.w("")

        out.w("## 3 · A tope, cada proxy (2 CPU, 1 GiB)\n")
        for k in orden:
            arrancar(k)
            time.sleep(3)
            with Muestreo(f"{K.P}-proxy") as m:
                time.sleep(3)
            reposo = m.resumen()["mem_max_mib"]
            out.w(f"### {k} · memoria en reposo {reposo:.1f} MiB\n")
            for i, (ruta, c) in enumerate(CARGAS):
                # cada carga con el proxy RECIÉN ARRANCADO: en la corrida 1 (descartada) las conexiones que Caddy abrió y cerró
                # con c=256 dejaron el contenedor tocado y la carga siguiente dio 502. Así ninguna carga hereda la anterior.
                if i:
                    K.rm(f"{K.P}-proxy")
                    time.sleep(2)
                    arrancar(k)
                    time.sleep(2)
                a0 = atendidas()
                k0 = CONEX[0]
                with Muestreo(f"{K.P}-proxy") as m:
                    h = hey(f"http://{IP_P}:8080{ruta}", c, SEG)
                s = m.resumen()
                llegaron = atendidas() - a0 - 1
                conex = CONEX[0] - k0 - 1
                # cada 2xx que vio el cliente tuvo que pasar por el backend (los que quedan en vuelo al cortar: ≤ c)
                cuadra = h["ok"] <= llegaron + 2 and llegaron - h["ok"] <= c + 2
                otros = {"no-2xx": h["codes"]["no-2xx"]} if h["codes"]["no-2xx"] else {}
                out.w(f"- {ruta} c={c}: **{h['rps']:.0f} pet/s** · p50 {h['p50']:.2f} ms · p99 {h['p99']:.2f} ms · CPU media "
                      f"{s['cpu_media']} % (máx {s['cpu_max']}) · memoria máx {s['mem_max_mib']:.1f} MiB · hilos máx "
                      f"{s['hilos_max']} · códigos {h['codes']}{' · ⚠️ ' + str(otros) if otros else ''}"
                      f"{' · errores de socket: ' + h['errores'] if h['errores'] else ''}")
                out.w(f"  - ¿llegó todo al backend? wrk contó {h['total']} respuestas ({h['ok']} 2xx); nginx atendió {llegaron} "
                      f"por {conex} conexiones nuevas del proxy ({'✅' if cuadra else '⛔ NO cuadra'})")
                if not cuadra:
                    out.w("```\n" + h["crudo"].strip() + "\n```")
                    raise SystemExit(1)
                tabla[(k, ruta, c)] = (h, s, reposo)
            K.rm(f"{K.P}-proxy")
            out.w("")

        out.w("## 3b · Caddy con keepalive_idle_conns_per_host 1024 (por defecto 32)\n")
        arrancar("caddy-ka")
        time.sleep(3)
        for i, (ruta, c) in enumerate(CARGAS):
            if i:
                K.rm(f"{K.P}-proxy")
                time.sleep(2)
                arrancar("caddy-ka")
                time.sleep(2)
            a0 = atendidas()
            k0 = CONEX[0]
            with Muestreo(f"{K.P}-proxy") as m:
                h = hey(f"http://{IP_P}:8080{ruta}", c, SEG)
            s = m.resumen()
            llegaron = atendidas() - a0 - 1
            out.w(f"- caddy-ka {ruta} c={c}: **{h['rps']:.0f} pet/s** · p50 {h['p50']:.2f} ms · p99 {h['p99']:.2f} ms · CPU media "
                  f"{s['cpu_media']} % · memoria máx {s['mem_max_mib']:.1f} MiB · códigos {h['codes']} · nginx atendió {llegaron} por {CONEX[0] - k0 - 1} conexiones nuevas")
        K.rm(f"{K.P}-proxy")
        out.w("")

        out.w("## 4 · La tabla\n")
        out.w("| proxy | lenguaje | carga | pet/s | p99 ms | CPU media % | memoria máx MiB | reposo MiB |")
        out.w("|---|---|---|---:|---:|---:|---:|---:|")
        for k in PROXIES:
            for ruta, c in CARGAS:
                h, s, rep = tabla[(k, ruta, c)]
                out.w(f"| {k} | {PROXIES[k]['lenguaje']} | {ruta} c={c} | {h['rps']:.0f} | {h['p99']:.2f} | {s['cpu_media']} | "
                      f"{s['mem_max_mib']:.1f} | {rep:.1f} |")
        K.rm(f"{K.P}-backend")
        K.sh(["docker", "network", "rm", RED], timeout=60)
        out.w(f"\n> Terminó {datetime.now():%Y-%m-%d %H:%M}. CPU de Windows al final: {K.cpu_windows()}.")


if __name__ == "__main__":
    main()
