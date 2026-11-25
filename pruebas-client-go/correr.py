# -*- coding: utf-8 -*-
"""Demo del vídeo «client-go: controla Kubernetes desde Go» (golandia 677).

    python correr.py [escenario ...] [--borrar]
    escenarios: compilar config desplegar pods watch-ingenuo watch-reintento contar conflicto rbac  (por defecto, todos)

Usa un clúster kind propio llamado g677 (lo crea si no existe; --borrar lo elimina al final).
Solo toca el namespace «demo» de ese clúster. kgo corre en Docker (busybox) en la red «kind», con el
kubeconfig interno y KGO_URLS=1 (imprime la URL de cada WATCH); para RBAC corre dentro de un pod con el ServiceAccount «sin-permisos».
Cada escenario deja su salida en salidas/<escenario>.txt; SALIDA.md se escribe a mano a partir de ellas.
"""
import os
import subprocess
import sys
import tempfile
from pathlib import Path

AQUI = Path(__file__).parent
SAL = AQUI / "salidas"
SAL.mkdir(exist_ok=True)
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
CLUSTER = "g677"
CTX = ["--context", f"kind-{CLUSTER}"]
NODO = f"{CLUSTER}-control-plane"
GOIMG = "golang:1.27.2"
TMP = Path(tempfile.mkdtemp(prefix="g677-"))
KCFG = TMP / "kubeconfig"


def win(p):
    r = subprocess.run(["cygpath", "-w", str(p)], capture_output=True, text=True)
    return r.stdout.strip() or str(p)


def sh(*a, check=True, entrada=None):
    r = subprocess.run(list(a), capture_output=True, text=True, env=ENV, encoding="utf-8", errors="replace", input=entrada)
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(a)}:\n{r.stdout}{r.stderr}")
    return (r.stdout + r.stderr).rstrip()


def k(*a, **kw):
    return sh("kubectl", *CTX, *a, **kw)


def guardar(nombre, texto):
    (SAL / f"{nombre}.txt").write_text(texto + "\n", encoding="utf-8")
    print(f"── {nombre}\n{texto}\n")


def kgo(sub):
    return sh("docker", "run", "--rm", "--network", "kind", "-v", f"{win(AQUI / 'kgo')}:/app", "-v", f"{win(TMP)}:/k",
              "-e", "KUBECONFIG=/k/kubeconfig", "-e", "KGO_URLS=1", "busybox:latest", "/app/kgo", sub, check=False)


def preparar():
    if CLUSTER not in sh("kind", "get", "clusters", check=False).split():
        sh("kind", "create", "cluster", "--name", CLUSTER, "--wait", "120s")
    KCFG.write_text(sh("kind", "get", "kubeconfig", "--name", CLUSTER, "--internal"), encoding="utf-8")
    for img in ("docker.io/library/nginx:1.30.5-alpine", "docker.io/library/busybox:1.37"):
        sh("docker", "exec", NODO, "crictl", "pull", img)
    k("create", "namespace", "demo", check=False)
    k("-n", "demo", "delete", "pod", "kgo-rbac", "--ignore-not-found", "--wait=true")  # que no cuente en las listas


def a_una_replica():
    """Cada escenario empieza igual: web con 1 réplica lista."""
    if "web" in k("-n", "demo", "get", "deploy", "-o", "name", check=False):
        k("-n", "demo", "scale", "deploy/web", "--replicas=1")
        k("-n", "demo", "rollout", "status", "deploy/web", "--timeout=90s")
        k("-n", "demo", "wait", "--for=delete", "pod", "-l", "app=web", "--field-selector=status.phase!=Running",
          "--timeout=60s", check=False)
        sh("sh", "-c", f"until [ $(kubectl --context kind-{CLUSTER} -n demo get pods -l app=web --no-headers | wc -l) -eq 1 ]; do sleep 1; done")


def compilar():
    v = win(AQUI / "kgo")
    out = sh("docker", "run", "--rm", "-v", "g677-gomod:/go/pkg/mod", "-v", f"{v}:/src", "-w", "/src", GOIMG, "sh", "-c",
             "go version && CGO_ENABLED=0 go build -ldflags='-s -w' -o kgo . && ls -l kgo | awk '{print \"kgo\", $5, \"bytes\"}' "
             "&& echo \"$(go list -m all | wc -l) módulos en el grafo · $(go list -deps . | wc -l) paquetes\" "
             "&& go list -m k8s.io/client-go k8s.io/api k8s.io/apimachinery")
    guardar("compilar", out)


POD = """apiVersion: v1
kind: Pod
metadata: {name: kgo-rbac, namespace: demo}
spec:
  serviceAccountName: sin-permisos
  restartPolicy: Never
  containers:
  - name: kgo
    image: busybox:1.37
    imagePullPolicy: IfNotPresent
    command: ["sh", "-c", "/kgo/kgo config; /kgo/kgo rbac"]
    volumeMounts: [{name: bin, mountPath: /kgo}]
  volumes: [{name: bin, hostPath: {path: /opt/kgo, type: Directory}}]
"""

ROL = """apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata: {name: leer-pods, namespace: demo}
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata: {name: sin-permisos-lee-pods, namespace: demo}
subjects: [{kind: ServiceAccount, name: sin-permisos, namespace: demo}]
roleRef: {apiGroup: rbac.authorization.k8s.io, kind: Role, name: leer-pods}
"""


def pod_rbac():
    k("-n", "demo", "delete", "pod", "kgo-rbac", "--ignore-not-found", "--wait=true")
    k("apply", "-f", "-", entrada=POD)
    k("-n", "demo", "wait", "--for=jsonpath={.status.phase}=Succeeded", "pod/kgo-rbac", "--timeout=90s", check=False)
    return k("-n", "demo", "logs", "kgo-rbac")


def rbac():
    sh("docker", "exec", NODO, "mkdir", "-p", "/opt/kgo")
    sh("docker", "cp", str(AQUI / "kgo" / "kgo"), f"{NODO}:/opt/kgo/kgo")
    k("-n", "demo", "delete", "rolebinding", "sin-permisos-lee-pods", "--ignore-not-found")
    k("-n", "demo", "delete", "role", "leer-pods", "--ignore-not-found")
    k("-n", "demo", "create", "serviceaccount", "sin-permisos", check=False)
    sa = "system:serviceaccount:demo:sin-permisos"
    txt = ["$ kubectl auth can-i list pods -n demo --as " + sa, k("auth", "can-i", "list", "pods", "-n", "demo", "--as", sa, check=False),
           "--- pod kgo-rbac con serviceAccountName: sin-permisos ---", pod_rbac(),
           "$ kubectl apply -f rol.yaml   (Role leer-pods: get, list, watch en pods · RoleBinding a sin-permisos)",
           k("apply", "-f", "-", entrada=ROL),
           "$ kubectl auth can-i list pods -n demo --as " + sa, k("auth", "can-i", "list", "pods", "-n", "demo", "--as", sa, check=False),
           "--- pod kgo-rbac otra vez ---", pod_rbac()]
    guardar("rbac", "\n".join(txt))


def main():
    pedidos = [a for a in sys.argv[1:] if not a.startswith("--")] or [
        "compilar", "config", "desplegar", "pods", "watch-ingenuo", "watch-reintento", "contar", "conflicto", "rbac"]
    preparar()
    guardar("versiones", "\n".join([sh("kind", "version"), k("version", check=False),
                                    sh("docker", "version", "--format", "Docker {{.Server.Version}}")]))
    for e in pedidos:
        if e == "compilar":
            compilar()
        elif e == "rbac":
            rbac()
        else:
            if e not in ("config", "pods", "desplegar"):
                a_una_replica()
            if e == "desplegar":
                k("-n", "demo", "delete", "deploy", "web", "--ignore-not-found", "--wait=true")
                k("-n", "demo", "wait", "--for=delete", "pod", "-l", "app=web", "--timeout=90s", check=False)
            guardar(e, f"$ kgo {e}\n" + kgo(e))
    if "--borrar" in sys.argv:
        sh("kind", "delete", "cluster", "--name", CLUSTER)
        print("clúster g677 borrado")


if __name__ == "__main__":
    main()
