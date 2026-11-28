# -*- coding: utf-8 -*-
"""muestreo.py — CPU %, memoria e hilos de un contenedor EN EL TIEMPO, mientras corre una carga (piloto: revancha de Zig, 2026-11-05).

Por qué no `docker stats`: con `--no-stream` tarda 1-2 s por lectura y en streaming da un dato por segundo; y `docker exec cat ...`
no sirve con imágenes FROM scratch (no hay `cat`). Aquí un contenedor auxiliar (busybox, --cpus 0.25) lee directamente los ficheros
del cgroup v2 del contenedor medido cada `intervalo` segundos (0,5 por defecto) y los imprime; Python los recoge en un hilo.

    cpu.stat      usage_usec        → CPU % = Δusage / Δtiempo (100 % = un núcleo entero; con --cpus 2 el techo es 200 %)
    memory.current − inactive_file  → memoria en MiB, la MISMA cuenta que hace `docker stats`
    pids.current                    → tareas del cgroup = procesos + HILOS (la columna PIDS de `docker stats`)
    cpu.stat      nr_throttled      → cuántas veces el límite de CPU frenó al contenedor
    (opcional, puerto=8080) /proc/<pid>/net/tcp* → conexiones TCP ESTABLECIDAS en ese puerto del contenedor
                                      (el auxiliar corre con --pid=host para ver /proc del proceso medido)

El reloj es /proc/uptime del lado del auxiliar (resolución 10 ms), así el Δtiempo no incluye la latencia de la tubería.

Uso (cualquier banco; `puerto=8080` añade la serie de conexiones establecidas):

    from muestreo import Muestreo
    with Muestreo("srv") as m:          # nombre o id del contenedor YA arrancado
        correr_la_carga()
    m.guardar(carpeta / "serie-go-suma.json", csv=True)
    m.resumen()   # {'cpu_media': .., 'cpu_max': .., 'mem_max_mib': .., 'hilos_max': .., 'muestras': ..}

Comprobado en la revancha de Zig: las peticiones/s con y sin muestreo están en SALIDA.md («¿el muestreo ensucia?»).
"""
import csv
import json
import subprocess
import threading
import time
from pathlib import Path

IMAGEN = "busybox:1.37"

_BUCLE = r"""
ID="$1"; INT="$2"; PUERTO="$3"; D=""
for c in /cg/docker/$ID /cg/system.slice/docker-$ID.scope; do [ -d "$c" ] && D="$c"; done
[ -z "$D" ] && { echo "ERR sin cgroup para $ID"; exit 1; }
while [ -d "$D" ]; do
  set -- $(cat /proc/uptime) ; T=$1
  U=$(sed -n 's/^usage_usec //p' $D/cpu.stat); TH=$(sed -n 's/^nr_throttled //p' $D/cpu.stat)
  M=$(cat $D/memory.current); I=$(sed -n 's/^inactive_file //p' $D/memory.stat); P=$(cat $D/pids.current)
  CX=-1
  if [ -n "$PUERTO" ]; then
    PID=$(head -1 $D/cgroup.procs)
    CX=$(cat /proc/$PID/net/tcp /proc/$PID/net/tcp6 2>/dev/null | awk -v p=":$PUERTO" '$4=="01" && substr($2, length($2)-4)==p' | wc -l)
  fi
  echo "S $T $U $M $I $P $TH $CX"
  sleep "$INT"
done
"""


class Muestreo:
    def __init__(self, contenedor, intervalo=0.5, puerto=None):
        self.contenedor = contenedor
        self.puerto = puerto     # si se da (p. ej. 8080), cuenta también las conexiones establecidas a ese puerto
        self.intervalo = intervalo
        self.crudo = []          # (t_uptime, usage_usec, mem_bytes, inactive_file, pids, nr_throttled, conexiones|-1)
        self.errores = []
        self._proc = None
        self._hilo = None
        self._nombre = None

    # ── ciclo de vida ─────────────────────────────────────────────────────────────────────────────
    def iniciar(self):
        r = subprocess.run(["docker", "inspect", "-f", "{{.Id}}", self.contenedor], capture_output=True, text=True)
        cid = r.stdout.strip()
        if r.returncode != 0 or not cid:
            raise RuntimeError(f"muestreo: no existe el contenedor {self.contenedor!r}")
        self._nombre = f"muestreo-{cid[:12]}-{int(time.time() * 1000) % 100000}"
        args = ["docker", "run", "--rm", "--name", self._nombre, "--cgroupns=host", "--cpus", "0.25",
                "-v", "/sys/fs/cgroup:/cg:ro"] + (["--pid=host"] if self.puerto else []) + [
                IMAGEN, "sh", "-c", _BUCLE, "muestreo", cid, str(self.intervalo), f"{self.puerto:04X}" if self.puerto else ""]
        self._proc = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, bufsize=1,
                                      env=_sin_conversion_de_rutas())
        self._hilo = threading.Thread(target=self._leer, daemon=True)
        self._hilo.start()
        # espera la primera muestra: sin ella, la carga empezaría sin nadie mirando
        fin = time.perf_counter() + 15
        while not self.crudo and time.perf_counter() < fin and self._proc.poll() is None:
            time.sleep(0.02)
        if not self.crudo:
            raise RuntimeError(f"muestreo: el auxiliar no dio muestras ({self.errores[:3]})")
        return self

    def _leer(self):
        for linea in self._proc.stdout:
            p = linea.split()
            if len(p) == 8 and p[0] == "S":
                try:
                    self.crudo.append((float(p[1]), int(p[2]), int(p[3]), int(p[4] or 0), int(p[5]), int(p[6]), int(p[7])))
                except ValueError:
                    self.errores.append(linea.strip())
            elif linea.strip():
                self.errores.append(linea.strip())

    def parar(self):
        if self._nombre:
            subprocess.run(["docker", "rm", "-f", self._nombre], capture_output=True)
        if self._proc:
            try:
                self._proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self._proc.kill()
        if self._hilo:
            self._hilo.join(timeout=5)

    def __enter__(self):
        return self.iniciar()

    def __exit__(self, *exc):
        self.parar()
        return False

    # ── resultados ────────────────────────────────────────────────────────────────────────────────
    def serie(self):
        """[{t, cpu_pct, mem_mib, hilos, frenadas}] con t en segundos desde la primera muestra."""
        out = []
        if not self.crudo:
            return out
        t0 = self.crudo[0][0]
        prev = None
        for t, u, m, i, p, th, cx in self.crudo:
            cpu = None
            if prev and t > prev[0]:
                cpu = (u - prev[1]) / ((t - prev[0]) * 1e6) * 100
            out.append({"t": round(t - t0, 2), "cpu_pct": None if cpu is None else round(cpu, 1),
                        "mem_mib": round((m - i) / 1048576, 2), "hilos": p, "frenadas": th - self.crudo[0][5],
                        **({"conexiones": cx} if cx >= 0 else {})})
            prev = (t, u)
        return out

    def resumen(self, desde=0.0, hasta=None):
        s = [x for x in self.serie() if x["t"] >= desde and (hasta is None or x["t"] <= hasta)]
        cpus = [x["cpu_pct"] for x in s if x["cpu_pct"] is not None]
        if not s:
            return {}
        return {"muestras": len(s), "cpu_media": round(sum(cpus) / len(cpus), 1) if cpus else None,
                "cpu_max": max(cpus) if cpus else None, "mem_max_mib": max(x["mem_mib"] for x in s),
                "hilos_max": max(x["hilos"] for x in s), "frenadas": s[-1]["frenadas"] - s[0]["frenadas"],
                **({"conexiones_max": max(x["conexiones"] for x in s)} if "conexiones" in s[0] else {})}

    def guardar(self, ruta_json, csv_tambien=True, extra=None):
        ruta_json = Path(ruta_json)
        datos = {"contenedor": self.contenedor, "intervalo_s": self.intervalo, "resumen": self.resumen(),
                 "serie": self.serie(), **(extra or {})}
        ruta_json.write_text(json.dumps(datos, indent=1, ensure_ascii=False), encoding="utf-8")
        if csv_tambien:
            with open(ruta_json.with_suffix(".csv"), "w", newline="", encoding="utf-8") as f:
                campos = ["t", "cpu_pct", "mem_mib", "hilos", "frenadas"] + (["conexiones"] if self.puerto else [])
                w = csv.DictWriter(f, fieldnames=campos)
                w.writeheader()
                w.writerows(self.serie())
        return ruta_json


def _sin_conversion_de_rutas():
    """En Git Bash, MSYS convierte «/sys/fs/cgroup» en una ruta de Windows: se apaga para el docker que lanza esto."""
    import os
    e = dict(os.environ)
    e["MSYS_NO_PATHCONV"] = "1"
    return e


if __name__ == "__main__":  # prueba rápida: python investigacion/muestreo.py <contenedor> [segundos]
    import sys
    seg = float(sys.argv[2]) if len(sys.argv) > 2 else 3
    with Muestreo(sys.argv[1], puerto=int(sys.argv[3]) if len(sys.argv) > 3 else None) as m:
        time.sleep(seg)
    for x in m.serie():
        print(x)
    print(m.resumen())
