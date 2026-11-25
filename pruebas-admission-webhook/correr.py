# -*- coding: utf-8 -*-
"""Demo del vídeo 678 de Golandia: un ValidatingAdmissionWebhook en Go (net/http), en un clúster kind propio.

    python correr.py preparar                 # clúster g678 (si no existe), imagen, certificado, webhook
    python correr.py [escenario ...]          # rechazos review tls caido colgado selector mutante vap latencia
    python correr.py borrar                   # borra el clúster g678

Solo toca el contexto kind-g678 (nunca el contexto actual de kubectl). Las claves del certificado van a un
directorio temporal (mkdtemp) que se borra al terminar. Cada escenario deja su salida en salidas/<escenario>.txt.
"""
import base64
import json
import os
import shutil
import statistics
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
from pathlib import Path

AQUI = Path(__file__).parent
SAL = AQUI / "salidas"
SAL.mkdir(exist_ok=True)
CLUSTER = "g678"
CTX = f"kind-{CLUSTER}"
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
NOMBRE = "pods.admision.golandia"


def sh(*a, entrada=None, check=True):
    t = time.perf_counter()
    r = subprocess.run(list(a), input=entrada, capture_output=True, text=True, env=ENV, encoding="utf-8", errors="replace")
    s = time.perf_counter() - t
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(a)}:\n{r.stdout}{r.stderr}")
    return r.returncode, r.stdout.strip(), r.stderr.strip(), s


def k(*a, entrada=None, check=True):
    return sh("kubectl", "--context", CTX, *a, entrada=entrada, check=check)


def aplicar(obj, *extra, check=True):
    return k("apply", *extra, "-f", "-", entrada=json.dumps(obj), check=check)


def win(p):
    return subprocess.run(["cygpath", "-w", str(p)], capture_output=True, text=True).stdout.strip() or str(p)


class Log:
    def __init__(self, nombre):
        self.f = open(SAL / f"{nombre}.txt", "w", encoding="utf-8")

    def __call__(self, *lineas):
        for l in lineas:
            print(l)
            self.f.write(l + "\n")
        self.f.flush()


# ── Objetos ──

def pod(nombre, imagen, limites=True, ns="default"):
    c = {"name": "web", "image": imagen}
    if limites:
        c["resources"] = {"limits": {"cpu": "200m", "memory": "128Mi"}}
    return {"apiVersion": "v1", "kind": "Pod", "metadata": {"name": nombre, "namespace": ns}, "spec": {"containers": [c]}}


SELECTOR = {"matchExpressions": [{"key": "kubernetes.io/metadata.name", "operator": "NotIn", "values": ["kube-system", "admision"]}]}


def config(ca, politica="Fail", selector=True, timeout=None, tipo="Validating", ruta="/validar"):
    w = {"name": NOMBRE if tipo == "Validating" else "mutar.admision.golandia",
         "admissionReviewVersions": ["v1"], "sideEffects": "None", "failurePolicy": politica,
         "rules": [{"apiGroups": [""], "apiVersions": ["v1"], "operations": ["CREATE", "UPDATE"], "resources": ["pods"]}],
         "clientConfig": {"service": {"namespace": "admision", "name": "admision", "path": ruta}}}
    if ca is not None:
        w["clientConfig"]["caBundle"] = ca
    if selector:
        w["namespaceSelector"] = SELECTOR
    if timeout:
        w["timeoutSeconds"] = timeout
    return {"apiVersion": "admissionregistration.k8s.io/v1", "kind": f"{tipo}WebhookConfiguration",
            "metadata": {"name": "admision" if tipo == "Validating" else "admision-mutar"}, "webhooks": [w]}


def ca_bundle():
    _, b64, _, _ = k("get", "secret", "-n", "admision", "admision-tls", "-o", "jsonpath={.data.ca\\.crt}")
    return b64


def esperar_listo():
    k("rollout", "status", "deploy/admision", "-n", "admision", "--timeout=120s")
    time.sleep(1.5)  # que el endpoint llegue al apiserver


def escala(n):
    k("scale", "deploy/admision", "-n", "admision", f"--replicas={n}")
    if n == 0:
        k("wait", "--for=delete", "pod", "-l", "app=admision", "-n", "admision", "--timeout=60s", check=False)
        time.sleep(1.5)
    else:
        esperar_listo()


def limpiar_pods():
    k("delete", "pod", "--all", "-n", "default", "--now", "--wait=false", check=False)


def intento(L, etiqueta, obj, dry=True):
    """Aplica un pod y apunta lo que ve el usuario y lo que tardó."""
    archivo, _, nota = etiqueta.partition(" ")
    extra = ["--dry-run=server"] if dry else []
    tmp = Path(tempfile.mkdtemp(prefix="g678-pod-"))
    try:
        (tmp / archivo).write_text(json.dumps(obj), encoding="utf-8")
        t = time.perf_counter()
        r = subprocess.run(["kubectl", "--context", CTX, "apply", "-f", archivo, *extra], cwd=tmp, capture_output=True, text=True,
                           env=ENV, encoding="utf-8", errors="replace")
        s = time.perf_counter() - t
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
    rc, out, err = r.returncode, r.stdout.strip(), r.stderr.strip()
    L(f"$ kubectl apply -f {archivo}{' --dry-run=server' if dry else ''}{'   ' + nota.strip() if nota.strip() else ''}",
      out or err, f"  (código {rc}, {s:.2f} s)", "")
    return rc, s


# ── Preparar ──

def preparar():
    L = Log("01-preparar")
    rc, clusters, _, _ = sh("kind", "get", "clusters", check=False)
    if CLUSTER not in clusters.split():
        sh("kind", "create", "cluster", "--name", CLUSTER, "--wait", "120s")
    v = json.loads(k("version", "-o", "json", check=False)[1])
    L(sh("kind", "version")[1], f"kubectl {v['clientVersion']['gitVersion']} (cliente) · servidor {v['serverVersion']['gitVersion']}",
      sh("docker", "version", "--format", "Docker {{.Server.Version}}")[1])
    sh("docker", "build", "-q", "-t", "admision:0.1", str(AQUI / "webhook"))
    sh("kind", "load", "docker-image", "admision:0.1", "--name", CLUSTER)
    tmp = Path(tempfile.mkdtemp(prefix="g678-certs-"))
    try:
        san = "subjectAltName=DNS:admision.admision.svc"
        guion = ("set -e; cd /c; "
                 "openssl req -x509 -newkey rsa:2048 -nodes -keyout ca.key -out ca.crt -days 365 -subj /CN=admision-ca 2>/dev/null; "
                 "openssl req -newkey rsa:2048 -nodes -keyout tls.key -out tls.csr -subj /CN=admision.admision.svc 2>/dev/null; "
                 f"printf '{san}' > san.ext; "
                 "openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt -days 365 -extfile san.ext 2>/dev/null; "
                 "openssl x509 -in tls.crt -noout -subject -issuer -ext subjectAltName; chmod 644 *")
        _, info, _, _ = sh("docker", "run", "--rm", "-v", f"{win(tmp)}:/c", "--entrypoint", "sh", "alpine/openssl:3.5.4", "-c", guion)
        L("$ openssl x509 -in tls.crt -noout -subject -issuer -ext subjectAltName", info, "")
        k("apply", "-f", str(AQUI / "k8s" / "admision.yaml"))
        k("delete", "secret", "admision-tls", "-n", "admision", check=False)
        k("create", "secret", "generic", "admision-tls", "-n", "admision",
          f"--from-file=tls.crt={tmp / 'tls.crt'}", f"--from-file=tls.key={tmp / 'tls.key'}", f"--from-file=ca.crt={tmp / 'ca.crt'}")
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
    k("rollout", "restart", "deploy/admision", "-n", "admision")
    esperar_listo()
    aplicar(config(ca_bundle()))
    L("webhook listo:", k("get", "validatingwebhookconfiguration", "admision")[1])


# ── Escenarios ──

def rechazos():
    L = Log("02-rechazos")
    aplicar(config(ca_bundle()))
    time.sleep(1)
    intento(L, "pod-latest.yaml        # nginx:latest, con limits", pod("web", "nginx:latest"))
    intento(L, "pod-sin-tag.yaml       # nginx, sin etiqueta", pod("web", "nginx"))
    intento(L, "pod-sin-limites.yaml   # nginx:1.30.5-alpine, sin limits", pod("web", "nginx:1.30.5-alpine", limites=False))
    intento(L, "pod-registro.yaml      # registro con puerto: localhost:5000/app", pod("web", "localhost:5000/app"))
    intento(L, "pod-bien.yaml          # nginx:1.30.5-alpine, con limits", pod("web", "nginx:1.30.5-alpine"))
    rc, out, err, s = k("run", "prueba", "--image=nginx:latest", "--dry-run=server", check=False)
    L("$ kubectl run prueba --image=nginx:latest --dry-run=server", out or err, f"  (código {rc}, {s:.2f} s)", "")


def review():
    L = Log("03-review")
    aplicar(config(ca_bundle()))
    intento(L, "pod-latest.yaml", pod("web", "nginx:latest", limites=False))
    time.sleep(1)
    _, logs, _, _ = k("logs", "deploy/admision", "-n", "admision", "--tail=3")
    for l in logs.splitlines():
        if " ENTRADA " in l:
            rev = json.loads(l.split(" ENTRADA ", 1)[1])
            req = rev["request"]
            obj = req["object"]
            obj["metadata"].pop("managedFields", None)
            L("── AdmissionReview de ENTRADA (recortado: object.metadata.managedFields fuera) ──",
              json.dumps({"apiVersion": rev["apiVersion"], "kind": rev["kind"], "request": {
                  kk: req[kk] for kk in ("uid", "kind", "resource", "namespace", "operation", "userInfo", "dryRun") if kk in req}}, indent=2, ensure_ascii=False),
              "object.spec.containers:", json.dumps(obj["spec"]["containers"], indent=2), f"bytes del cuerpo completo: {len(l.split(' ENTRADA ', 1)[1])}", "")
        if " SALIDA " in l:
            L("── AdmissionReview de SALIDA ──", json.dumps(json.loads(l.split(" SALIDA ", 1)[1]), indent=2, ensure_ascii=False), "")
        if " op=" in l:
            L("log: " + l)


def tls():
    L = Log("04-tls")
    aplicar(config(None))
    time.sleep(1)
    L("# configuración SIN caBundle: el apiserver no confía en el certificado autofirmado")
    intento(L, "pod-bien.yaml", pod("web", "nginx:1.30.5-alpine"))
    aplicar(config(ca_bundle()))
    time.sleep(1)
    L("# con caBundle = ca.crt en base64")
    intento(L, "pod-bien.yaml", pod("web", "nginx:1.30.5-alpine"))


def caido():
    L = Log("05-caido")
    ca = ca_bundle()
    escala(0)
    for pol in ("Fail", "Ignore"):
        aplicar(config(ca, politica=pol))
        time.sleep(1)
        L(f"# webhook con 0 réplicas · failurePolicy: {pol}")
        intento(L, "pod-bien.yaml", pod("bien", "nginx:1.30.5-alpine"), dry=False)
        intento(L, "pod-latest.yaml", pod("latest", "nginx:latest"), dry=False)
        L(k("get", "pods", "-n", "default", "-o", "custom-columns=POD:.metadata.name,IMAGEN:.spec.containers[0].image", check=False)[1] or "(no hay pods)", "")
        limpiar_pods()
        time.sleep(2)
    escala(1)
    aplicar(config(ca))


def colgado():
    L = Log("06-colgado")
    ca = ca_bundle()
    k("set", "env", "deploy/admision", "-n", "admision", "DORMIR=8s")
    esperar_listo()
    for pol, to in (("Fail", 3), ("Ignore", 3), ("Fail", None)):
        aplicar(config(ca, politica=pol, timeout=to))
        time.sleep(1)
        L(f"# webhook que tarda 8 s en contestar · failurePolicy: {pol} · timeoutSeconds: {to or 'sin poner (10 por defecto)'}")
        intento(L, "pod-bien.yaml", pod("bien", "nginx:1.30.5-alpine"))
    k("set", "env", "deploy/admision", "-n", "admision", "DORMIR-")
    esperar_listo()
    aplicar(config(ca))


def selector():
    L = Log("07-selector")
    ca = ca_bundle()
    L("# los limits de CoreDNS, tal como los trae kind:",
      k("get", "deploy", "coredns", "-n", "kube-system", "-o", "jsonpath={.spec.template.spec.containers[0].image} {.spec.template.spec.containers[0].resources}")[1], "")
    for con in (True, False):
        aplicar(config(ca, selector=con))
        time.sleep(1)
        uno = k("get", "pods", "-n", "kube-system", "-l", "k8s-app=kube-dns", "-o", "jsonpath={.items[0].metadata.name}")[1]
        k("delete", "pod", "-n", "kube-system", uno, "--wait=true")
        time.sleep(12)
        L(f"# namespaceSelector {'que excluye kube-system y admision' if con else 'QUITADO'} · borro el pod {uno} de CoreDNS y espero 12 s",
          k("get", "deploy", "coredns", "-n", "kube-system")[1])
        _, ev, _, _ = k("get", "events", "-n", "kube-system", "--field-selector", "reason=FailedCreate",
                        "-o", "custom-columns=MENSAJE:.message", "--sort-by=.lastTimestamp", check=False)
        if not con:
            L("eventos FailedCreate:", ev.splitlines()[-1] if ev else "(ninguno)", "")
    # El abrazo mortal: sin selector, Fail, y el webhook caído no puede volver a arrancar.
    aplicar(config(ca, selector=False))
    escala(0)
    k("scale", "deploy/admision", "-n", "admision", "--replicas=1")
    time.sleep(15)
    L("# sin selector + Fail · bajo el webhook a 0 y lo vuelvo a subir a 1 · a los 15 s:",
      k("get", "deploy", "admision", "-n", "admision")[1])
    _, ev, _, _ = k("get", "events", "-n", "admision", "--field-selector", "reason=FailedCreate",
                    "-o", "custom-columns=MENSAJE:.message", "--sort-by=.lastTimestamp", check=False)
    L("eventos FailedCreate:", ev.splitlines()[-1] if ev else "(ninguno)", "")
    t = time.perf_counter()
    k("delete", "validatingwebhookconfiguration", "admision")
    esperar_listo()
    L(f"# salida: borrar la ValidatingWebhookConfiguration · el webhook vuelve a estar listo en {time.perf_counter() - t:.1f} s",
      k("get", "deploy", "admision", "-n", "admision")[1])
    k("rollout", "status", "deploy/coredns", "-n", "kube-system", "--timeout=120s")
    L(k("get", "deploy", "coredns", "-n", "kube-system")[1])
    aplicar(config(ca))


def mutante():
    L = Log("08-mutante")
    ca = ca_bundle()
    aplicar(config(ca))
    aplicar(config(ca, tipo="Mutating", ruta="/mutar"))
    time.sleep(1)
    L("# con el webhook mutante (/mutar) delante del validador (/validar)")
    intento(L, "pod-sin-limites.yaml   # nginx:1.30.5-alpine, sin limits", pod("sinlim", "nginx:1.30.5-alpine", limites=False), dry=False)
    L("$ kubectl get pod sinlim -o jsonpath='{.spec.containers[0].resources}'",
      k("get", "pod", "sinlim", "-o", "jsonpath={.spec.containers[0].resources}", check=False)[1], "")
    intento(L, "pod-latest-sin-limites.yaml", pod("lat", "nginx:latest", limites=False))
    _, logs, _, _ = k("logs", "deploy/admision", "-n", "admision", "--tail=12")
    for l in logs.splitlines():
        if " SALIDA " in l and '"patch"' in l:
            r = json.loads(l.split(" SALIDA ", 1)[1])["response"]
            L("patch (base64 → texto): " + base64.b64decode(r["patch"]).decode())
            break
    k("delete", "mutatingwebhookconfiguration", "admision-mutar", check=False)
    limpiar_pods()


VAP = {"apiVersion": "admissionregistration.k8s.io/v1", "kind": "ValidatingAdmissionPolicy", "metadata": {"name": "admision-cel"},
       "spec": {"failurePolicy": "Fail",
                "matchConstraints": {"resourceRules": [{"apiGroups": [""], "apiVersions": ["v1"], "operations": ["CREATE", "UPDATE"], "resources": ["pods"]}]},
                "validations": [
                    {"expression": "object.spec.containers.all(c, c.image.split('/')[size(c.image.split('/')) - 1].contains(':') && !c.image.endsWith(':latest'))",
                     "message": "fija una versión: nada de :latest ni imágenes sin etiqueta"},
                    {"expression": "object.spec.containers.all(c, has(c.resources.limits) && 'cpu' in c.resources.limits && 'memory' in c.resources.limits)",
                     "message": "cada contenedor necesita limits.cpu y limits.memory"}]}}
VAPB = {"apiVersion": "admissionregistration.k8s.io/v1", "kind": "ValidatingAdmissionPolicyBinding", "metadata": {"name": "admision-cel"},
        "spec": {"policyName": "admision-cel", "validationActions": ["Deny"], "matchResources": {"namespaceSelector": SELECTOR}}}


def vap():
    L = Log("09-vap")
    k("delete", "validatingwebhookconfiguration", "admision", check=False)
    aplicar(VAP)
    aplicar(VAPB)
    time.sleep(3)
    L("# ValidatingAdmissionPolicy (CEL, dentro del apiserver) · sin webhook")
    intento(L, "pod-latest.yaml", pod("web", "nginx:latest"))
    intento(L, "pod-sin-tag.yaml", pod("web", "nginx"))
    intento(L, "pod-sin-limites.yaml", pod("web", "nginx:1.30.5-alpine", limites=False))
    intento(L, "pod-registro.yaml", pod("web", "localhost:5000/app"))
    intento(L, "pod-bien.yaml", pod("web", "nginx:1.30.5-alpine"))
    escala(0)
    L("# con el Deployment del webhook a 0 réplicas, la política sigue funcionando:")
    intento(L, "pod-latest.yaml", pod("web", "nginx:latest"))
    escala(1)
    k("delete", "validatingadmissionpolicybinding", "admision-cel")
    k("delete", "validatingadmissionpolicy", "admision-cel")
    aplicar(config(ca_bundle()))


def medir(url, cuerpo, n):
    tiempos = []
    for i in range(n):
        req = urllib.request.Request(url, data=cuerpo, method="POST", headers={"Content-Type": "application/json"})
        t = time.perf_counter()
        try:
            urllib.request.urlopen(req).read()
        except urllib.error.HTTPError as e:
            raise SystemExit(f"esperaba 201 y llegó {e.code}: {e.read()[:300]}")
        tiempos.append((time.perf_counter() - t) * 1000)
    tiempos.sort()
    return statistics.median(tiempos), tiempos[int(n * 0.95) - 1], statistics.mean(tiempos)


def codigo(url, obj):
    req = urllib.request.Request(url, data=json.dumps(obj).encode(), method="POST", headers={"Content-Type": "application/json"})
    try:
        return urllib.request.urlopen(req).status
    except urllib.error.HTTPError as e:
        return e.code


def montar(variante, ca):
    k("delete", "validatingwebhookconfiguration", "admision", check=False)
    k("delete", "validatingadmissionpolicybinding", "admision-cel", check=False)
    k("delete", "validatingadmissionpolicy", "admision-cel", check=False)
    if variante == "webhook en Go":
        aplicar(config(ca))
    elif variante == "ValidatingAdmissionPolicy":
        aplicar(VAP)
        aplicar(VAPB)
    time.sleep(3)


def metricas():
    """Sumas y cuentas de las métricas de admisión del apiserver (solo CREATE admitidos)."""
    _, txt, _, _ = k("get", "--raw", "/metrics")
    m = {}
    for l in txt.splitlines():
        if l.startswith("#") or 'operation="CREATE"' not in l or "_summary" in l:
            continue
        for clave, pref in (("paso", 'apiserver_admission_step_admission_duration_seconds_{}{{operation="CREATE",rejected="false",type="validate"}}'),
                            ("webhook", 'apiserver_admission_webhook_admission_duration_seconds_{}{{name="pods.admision.golandia",operation="CREATE",rejected="false"')):
            for q in ("sum", "count"):
                if l.startswith(pref.format(q)):
                    m[(clave, q)] = float(l.rsplit(" ", 1)[1])
    return m


def media_us(a, b, clave):
    n = b.get((clave, "count"), 0) - a.get((clave, "count"), 0)
    return (b.get((clave, "sum"), 0) - a.get((clave, "sum"), 0)) / n * 1e6 if n else None


def latencia(n=300, rondas=5):
    L = Log("10-latencia")
    ca = ca_bundle()
    prox = subprocess.Popen(["kubectl", "--context", CTX, "proxy", "--port=8678"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, env=ENV)
    variantes = ["sin nada", "webhook en Go", "ValidatingAdmissionPolicy"]
    esperado = {"sin nada": 201, "webhook en Go": 403, "ValidatingAdmissionPolicy": 422}
    try:
        time.sleep(2)
        url = "http://127.0.0.1:8678/api/v1/namespaces/default/pods?dryRun=All"
        cuerpo = json.dumps(pod("lat", "nginx:1.30.5-alpine")).encode()
        L(f"# {n} POST de un pod válido con dryRun=All por kubectl proxy (pasa por toda la admisión, no guarda nada)",
          f"# {rondas} rondas, el orden rota en cada una; 30 de calentamiento por tanda; antes de medir, un pod :latest comprueba que la regla está puesta")
        meds = {v: [] for v in variantes}
        pasos_us = {}
        for r in range(rondas):
            orden = variantes[r:] + variantes[:r]
            for v in orden:
                montar(v, ca)
                c = codigo(url, pod("lat", "nginx:latest"))
                if c != esperado[v]:
                    raise SystemExit(f"{v}: el pod :latest dio {c}, esperaba {esperado[v]}: el banco no está en el estado que creo")
                medir(url, cuerpo, 30)
                a = metricas()
                med, p95, media = medir(url, cuerpo, n)
                b = metricas()
                meds[v].append(med)
                paso, wh = media_us(a, b, "paso"), media_us(a, b, "webhook")
                L(f"  ronda {r + 1} · {v:<27} (:latest → {c}) cliente: mediana {med:5.2f} ms  p95 {p95:5.2f} ms"
                  f"  · apiserver, paso validate: {paso:6.0f} µs/pod" + (f"  · llamada al webhook: {wh:6.0f} µs" if wh else ""))
                pasos_us.setdefault(v, []).append(paso)
        L("")
        base = statistics.median(meds["sin nada"])
        for v in variantes:
            m = statistics.median(meds[v])
            L(f"  {v:<27} cliente: mediana de las {rondas} rondas {m:5.2f} ms ({m - base:+.2f} ms)"
              f"  · apiserver, paso validate: {statistics.median(pasos_us[v]):5.0f} µs/pod")
        montar("webhook en Go", ca)
        medir(url, cuerpo, 50)
        _, logs, _, _ = k("logs", "deploy/admision", "-n", "admision", "--tail=50")
        dentro = [l.rsplit(" en ", 1)[1] for l in logs.splitlines() if " op=CREATE" in l]
        L("", "# lo que tarda el handler de Go por dentro (últimas 5 líneas del log):", *[f"  {d}" for d in dentro[-5:]])
    finally:
        prox.terminate()


def borrar():
    sh("kind", "delete", "cluster", "--name", CLUSTER)
    print("clúster g678 borrado")


ESC = {"preparar": preparar, "rechazos": rechazos, "review": review, "tls": tls, "caido": caido, "colgado": colgado,
       "selector": selector, "mutante": mutante, "vap": vap, "latencia": latencia, "borrar": borrar}

if __name__ == "__main__":
    pedidos = sys.argv[1:] or ["rechazos", "review", "tls", "caido", "colgado", "selector", "mutante", "vap", "latencia"]
    for e in pedidos:
        print(f"\n═══ {e} ═══")
        ESC[e]()
