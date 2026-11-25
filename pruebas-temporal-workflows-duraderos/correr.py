# -*- coding: utf-8 -*-
"""correr.py — «Workflows duraderos con Temporal en Go» (golandia). Todo en local, en contenedores desechables.

    python correr.py      → SALIDA.md

Escenarios (los efectos de cada paso se apuntan en _datos/efectos.log, un volumen que comparten todos los procesos):
  A · el flujo como programa normal de Go: se mata a mitad de la factura y se relanza.
  B · el mismo flujo como workflow de Temporal: la pasarela falla 2 veces (reintentos), se mata el worker a mitad de la
      factura (SIGKILL), se espera sin ningún worker y se arranca otro.
  C · cambiar el código de un workflow que está a medias: sin y con workflow.GetVersion.
Prefijo Docker gl-prac-; todo con --rm salvo los contenedores que hay que matar, que se borran al final.
⛔ Lo que diga el vídeo sale de la SALIDA de aquí.
"""
import json
import re
import shutil
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
DATOS = AQUI / "_datos"
RED = "gl-prac-temporal"
SRV = "gl-prac-temporal"
TEMPORAL_IMG = "temporalio/temporal:1.10.1"
GO_IMG = "gl-prac-temporal-go:1"
ADDR = f"{SRV}:7233"
OUT = []
MIOS = []  # contenedores creados (para borrarlos al final)


def sh(args, check=False, **kw):
    r = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(args)}\n{r.stdout}\n{r.stderr}")
    return r


def w(s=""):
    print(s)
    OUT.append(s)


def bloque(texto):
    w("```")
    for linea in texto.rstrip("\n").splitlines():
        w(linea)
    w("```")
    w()


def tcli(*args):
    """La CLI de temporal, dentro del contenedor del servidor."""
    return sh(["docker", "exec", SRV, "temporal", *args, "--address", "127.0.0.1:7233"])


def contenedor(nombre, *cmd, env=None, detach=True):
    sh(["docker", "rm", "-f", nombre])
    args = ["docker", "run", "--name", nombre, "--network", RED, "--ulimit", "core=1",
            "-v", f"{DATOS}:/datos", "-e", f"TEMPORAL_ADDRESS={ADDR}"]
    if detach:
        args.insert(2, "-d")
        MIOS.append(nombre)
    else:
        args.insert(2, "--rm")
    for k, v in (env or {}).items():
        args += ["-e", f"{k}={v}"]
    return sh(args + [GO_IMG, *cmd], check=True)


def logs(nombre, todo=False):
    """La salida de un contenedor. Por defecto sin las líneas DEBUG/INFO del SDK (sí las ERROR/WARN y las del programa)."""
    r = sh(["docker", "logs", nombre])
    t = r.stdout + r.stderr
    if todo:
        return t
    return "\n".join(x for x in t.splitlines() if not re.match(r"\d{4}/\d\d/\d\d \d\d:\d\d:\d\d (DEBUG|INFO) ", x))


def efectos(pedido=None):
    p = DATOS / "efectos.log"
    if not p.exists():
        return []
    ls = p.read_text(encoding="utf-8").splitlines()
    return [x for x in ls if pedido is None or re.search(rf"\b{re.escape(pedido)}\b", x)]


def esperar(cond, seg=60, paso=0.1, que=""):
    t0 = time.time()
    while time.time() - t0 < seg:
        if cond():
            return
        time.sleep(paso)
    raise SystemExit(f"no llegó: {que}")


def cuenta(lineas, que):
    return sum(1 for x in lineas if f" {que} " in f" {x} ")


def resumen(pedido, lineas):
    paginas = [int(m.group(1)) for x in lineas for m in [re.search(r"página (\d+)", x)] if m]
    return (f"reservar: {cuenta(lineas, 'reservar')} · cobrar: {cuenta(lineas, 'cobrar')} · validar: {cuenta(lineas, 'validar')} · "
            f"páginas de factura: {len(paginas)} ({', '.join(map(str, paginas))}) · enviar: {cuenta(lineas, 'enviar')}")


def describe(wid):
    r = tcli("workflow", "describe", "--workflow-id", wid, "--output", "json")
    return json.loads(r.stdout)


def estado(wid):
    return describe(wid)["workflowExecutionInfo"]["status"]


def matar(nombre):
    t = time.time()
    sh(["docker", "kill", "--signal", "KILL", nombre], check=True)
    return t


def main():
    if DATOS.exists():
        shutil.rmtree(DATOS)
    DATOS.mkdir()
    sh(["docker", "network", "create", "--subnet", "10.213.47.0/24", RED])  # sin --subnet: «all predefined address pools have been fully subnetted»
    sh(["docker", "build", "-q", "-t", GO_IMG, str(AQUI)], check=True)

    w("# SALIDA — Workflows duraderos con Temporal en Go (golandia)\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python correr.py`. Todo en local: el servidor de desarrollo "
      "de Temporal en un contenedor y los workers en otros. Los «efectos» de cada paso se apuntan en un archivo compartido "
      "(_datos/efectos.log) para contar cuántas veces se ejecutó de verdad cada uno.\n")
    tv = sh(["docker", "run", "--rm", TEMPORAL_IMG, "--version"]).stdout.strip()
    gv = sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip()
    sdk = re.search(r"go\.temporal\.io/sdk (v[\d.]+)", (AQUI / "go" / "go.mod").read_text(encoding="utf-8")).group(1)
    w(f"- `{tv}` (imagen `{TEMPORAL_IMG}`, `temporal server start-dev`) · `{gv}` · SDK `go.temporal.io/sdk {sdk}`")
    w(f"- Imagen de los binarios de Go (`{GO_IMG}`, FROM scratch): "
      f"{int(sh(['docker', 'image', 'inspect', '-f', '{{.Size}}', GO_IMG]).stdout) / 1048576:.1f} MiB")
    w("- El flujo: Reservar (1 s) → Cobrar (la pasarela falla los intentos 1 y 2) → Facturar (6 páginas de 1 s, con heartbeat "
      "por página) → Enviar. Reintentos: 1 s, ×2, máximo 10 s. HeartbeatTimeout 3 s, StartToCloseTimeout 30 s.\n")

    # ── servidor ──
    sh(["docker", "rm", "-f", SRV])
    MIOS.append(SRV)
    sh(["docker", "run", "-d", "--name", SRV, "--network", RED, "--ulimit", "core=1", TEMPORAL_IMG,
        "server", "start-dev", "--ip", "0.0.0.0", "--headless", "--log-level", "error"], check=True)
    esperar(lambda: tcli("operator", "cluster", "health").returncode == 0, 60, 0.5, "servidor")
    w("## 0 · El servidor\n")
    bloque(tcli("operator", "cluster", "health").stdout)

    # ── A · sin Temporal ──
    w("## A · Sin Temporal: el mismo flujo como un programa normal, y se mata a mitad de la factura\n")
    contenedor("gl-prac-tmp-sin1", "/bin/sintemporal", "pedido-A")
    esperar(lambda: any("página 2" in x for x in efectos("pedido-A")), 30, 0.05, "página 2 de A")
    matar("gl-prac-tmp-sin1")
    w("Primera ejecución (SIGKILL tras la página 2):")
    bloque(logs("gl-prac-tmp-sin1"))
    r = contenedor("gl-prac-tmp-sin2", "/bin/sintemporal", "pedido-A", detach=False)
    w("Se vuelve a lanzar el mismo programa:")
    bloque(r.stdout)
    ea = efectos("pedido-A")
    w("Efectos de pedido-A:")
    bloque("\n".join(ea))
    w(f"**Resumen A:** {resumen('pedido-A', ea)}\n")

    # ── B · con Temporal ──
    w("## B · Con Temporal: reintentos, SIGKILL al worker a mitad de la factura, y otro worker que sigue\n")
    contenedor("gl-prac-tmp-w1", "/bin/worker", env={"WORKER": "worker1"})
    esperar(lambda: "worker listo" in logs("gl-prac-tmp-w1", todo=True), 30, 0.1, "worker1")
    r = contenedor("gl-prac-tmp-ini", "/bin/iniciar", "pedido-B", detach=False)
    w("Se inicia el workflow:")
    bloque(r.stdout)
    esperar(lambda: any("página 2" in x for x in efectos("pedido-B")), 60, 0.05, "página 2 de B")
    q = tcli("workflow", "query", "--workflow-id", "pedido-B", "--type", "estado")
    w("Consulta del estado (query «estado») con worker1 vivo:")
    bloque(q.stdout + q.stderr)
    t_kill = matar("gl-prac-tmp-w1")
    w(f"SIGKILL a worker1 a las {datetime.utcfromtimestamp(t_kill):%H:%M:%S.%f}"[:-3] + " (UTC). Su salida completa:")
    bloque(logs("gl-prac-tmp-w1"))
    time.sleep(5)
    d = describe("pedido-B")
    pa = d.get("pendingActivities", [{}])[0]
    w("5 s después, SIN ningún worker, `temporal workflow describe pedido-B` (campos):")
    hbv = json.dumps(pa.get("heartbeatDetails"), ensure_ascii=False)
    bloque(f"status: {d['workflowExecutionInfo']['status']}\n"
           f"historyLength: {d['workflowExecutionInfo'].get('historyLength')}\n"
           f"pendingActivity: {pa.get('activityType', {}).get('name')} · state {pa.get('state')} · attempt {pa.get('attempt')}\n"
           f"heartbeatDetails: {hbv}\n"
           f"lastFailure: {pa.get('lastFailure', {}).get('message', '—')}")
    tq = time.time()
    q2 = tcli("workflow", "query", "--workflow-id", "pedido-B", "--type", "estado")
    w(f"Y la query sin worker (la responde el código del workflow, que corre en un worker); tardó {time.time() - tq:.1f} s en "
      "responder:")
    bloque((q2.stdout + q2.stderr).strip()[:600])
    contenedor("gl-prac-tmp-w2", "/bin/worker", env={"WORKER": "worker2"})
    t_w2 = time.time()
    esperar(lambda: estado("pedido-B") == "WORKFLOW_EXECUTION_STATUS_COMPLETED", 90, 0.2, "B completo")
    w(f"worker2 arranca {t_w2 - t_kill:.1f} s después del SIGKILL. Su salida:")
    bloque(logs("gl-prac-tmp-w2"))
    res = tcli("workflow", "result", "--workflow-id", "pedido-B")
    w("Resultado:")
    bloque(res.stdout + res.stderr)
    eb = efectos("pedido-B")
    w("Efectos de pedido-B:")
    bloque("\n".join(eb))
    w(f"**Resumen B:** {resumen('pedido-B', eb)}\n")
    h = json.loads(tcli("workflow", "show", "--workflow-id", "pedido-B", "--output", "json").stdout)
    evs = h.get("events", [])
    tipos = [e["eventType"].replace("EVENT_TYPE_", "") for e in evs]
    w(f"El historial de pedido-B: {len(evs)} eventos. Los de actividades, en orden:")
    acts = []
    for e in evs:
        t = e["eventType"].replace("EVENT_TYPE_", "")
        if t == "ACTIVITY_TASK_SCHEDULED":
            acts.append(f"{e['eventId']:>3} SCHEDULED {e['activityTaskScheduledEventAttributes']['activityType']['name']}")
        elif t == "ACTIVITY_TASK_STARTED":
            a = e["activityTaskStartedEventAttributes"]
            acts.append(f"{e['eventId']:>3} STARTED   attempt {a.get('attempt')} · {a.get('identity')}")
        elif t == "ACTIVITY_TASK_COMPLETED":
            acts.append(f"{e['eventId']:>3} COMPLETED · {e['activityTaskCompletedEventAttributes'].get('identity')}")
    bloque("\n".join(acts))
    w("Conteo por tipo de evento: " + ", ".join(f"{t} {tipos.count(t)}" for t in dict.fromkeys(tipos)) + "\n")

    # ⚠️ 2026-11-23, primera corrida: worker2 seguía vivo y se llevó TODO el escenario C (con la versión 1). Se quita aquí.
    sh(["docker", "rm", "-f", "gl-prac-tmp-w2"])

    # ── C · cambiar el código con un workflow a medias ──
    w("## C · Cambiar el código de un workflow que está a medias\n")
    w("(Antes se quita worker2: en C solo hay el worker que se dice en cada paso.)\n")
    contenedor("gl-prac-tmp-w3", "/bin/worker", env={"WORKER": "worker3", "VERSION": "1"})
    esperar(lambda: "worker listo" in logs("gl-prac-tmp-w3", todo=True), 30, 0.1, "worker3")
    contenedor("gl-prac-tmp-ini", "/bin/iniciar", "pedido-C", detach=False)
    esperar(lambda: any("página 1" in x for x in efectos("pedido-C")), 60, 0.05, "página 1 de C")
    matar("gl-prac-tmp-w3")
    w("pedido-C empezó con la versión 1 (worker3) y worker3 murió en la factura. Ahora arranca worker4 con la VERSIÓN 2: "
      "un paso nuevo, Validar, entre Reservar y Cobrar, agregado sin más.\n")
    contenedor("gl-prac-tmp-w4", "/bin/worker", env={"WORKER": "worker4", "VERSION": "2"})
    time.sleep(12)
    l4 = logs("gl-prac-tmp-w4", todo=True)
    nd = [x for x in l4.splitlines() if re.search(r"nondetermin", x, re.I)]
    w(f"worker4, 12 s después: {len(nd)} líneas con «nondeterministic» en su log. La primera (recortada a 600 caracteres):")
    bloque(nd[0][:600] if nd else "(ninguna)")
    w(f"Estado de pedido-C: {estado('pedido-C')} · efectos de pedido-C hasta aquí: {resumen('pedido-C', efectos('pedido-C'))}\n")
    sh(["docker", "rm", "-f", "gl-prac-tmp-w4"])
    w("Se quita worker4 y arranca worker5 con la VERSIÓN 3: el mismo paso, detrás de "
      "`workflow.GetVersion(ctx, \"validar-direccion\", workflow.DefaultVersion, 1)`.\n")
    contenedor("gl-prac-tmp-w5", "/bin/worker", env={"WORKER": "worker5", "VERSION": "3"})
    esperar(lambda: estado("pedido-C") == "WORKFLOW_EXECUTION_STATUS_COMPLETED", 90, 0.2, "C completo")
    contenedor("gl-prac-tmp-ini", "/bin/iniciar", "pedido-D", detach=False)
    esperar(lambda: estado("pedido-D") == "WORKFLOW_EXECUTION_STATUS_COMPLETED", 90, 0.2, "D completo")
    w("worker5:")
    bloque(logs("gl-prac-tmp-w5"))
    w(f"- pedido-C (empezado con la v1, terminado con la v3): {estado('pedido-C')} · {resumen('pedido-C', efectos('pedido-C'))}")
    w(f"- pedido-D (nuevo, entero con la v3): {estado('pedido-D')} · {resumen('pedido-D', efectos('pedido-D'))}")
    w()

    for n in MIOS:
        sh(["docker", "rm", "-f", n])
    sh(["docker", "network", "rm", RED])
    w("Contenedores y red borrados al terminar: " + ", ".join(dict.fromkeys(MIOS)) + f", red {RED}.")
    (AQUI / "SALIDA.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print(f"\n[ok] {AQUI / 'SALIDA.md'}")


if __name__ == "__main__":
    try:
        main()
    finally:
        for n in MIOS:
            sh(["docker", "rm", "-f", n])
