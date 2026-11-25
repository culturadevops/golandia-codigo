#!/bin/bash
# Corre DENTRO de golang:1.26, en la red «operador» junto a un k3s (lo lanza ejecutar.sh). Salida literal → SALIDA.md.
set -u
export KUBECONFIG=/trabajo/_kubeconfig PATH=/w/bin:$PATH GOFLAGS=-buildvcs=false
mkdir -p /w/bin && cd /w
curl -sL -o /w/bin/kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v4.16.0/kubebuilder_linux_amd64
curl -sL -o /w/bin/kubectl https://dl.k8s.io/release/v1.37.1/bin/linux/amd64/kubectl
chmod +x /w/bin/*
k() { kubectl "$@" 2>&1; }

echo "### 1 · El esqueleto con kubebuilder"
mkdir paginas && cd paginas
echo "\$ kubebuilder version"
kubebuilder version 2>&1 | head -1
echo "\$ kubebuilder init --domain golandia.dev --repo paginas"
kubebuilder init --domain golandia.dev --repo paginas >/tmp/init.log 2>&1 && echo "(esqueleto creado)" || tail -5 /tmp/init.log
echo "\$ kubebuilder create api --group web --version v1 --kind PaginaWeb --resource --controller"
kubebuilder create api --group web --version v1 --kind PaginaWeb --resource --controller >/tmp/api.log 2>&1 && echo "(API y controlador creados)" || tail -5 /tmp/api.log
cp /trabajo/codigo/paginaweb_types.go api/v1/paginaweb_types.go
cp /trabajo/codigo/paginaweb_controller.go internal/controller/paginaweb_controller.go
rm -f internal/controller/paginaweb_controller_test.go
echo "\$ wc -l api/v1/paginaweb_types.go internal/controller/paginaweb_controller.go   # lo que escribo yo"
wc -l api/v1/paginaweb_types.go internal/controller/paginaweb_controller.go
echo "\$ find . -name '*.go' | wc -l ; find . -type f | wc -l   # lo que genera kubebuilder"
find . -name '*.go' | wc -l ; find . -type f -not -path './bin/*' | wc -l
echo "\$ make manifests generate"
make manifests generate >/tmp/gen.log 2>&1 && echo "(CRD, RBAC y DeepCopy generados)" || tail -8 /tmp/gen.log
echo "\$ make install   # la CRD en el clúster"
make install >/tmp/install.log 2>&1; grep -E "created|configured|unchanged" /tmp/install.log | tail -2
echo "\$ kubectl get crd | grep paginaweb"
k get crd -o name | grep paginaweb

echo
echo "### 2 · El operador en marcha"
echo "\$ go build -o bin/operador ./cmd/main.go"
go build -o bin/operador ./cmd/main.go 2>&1 | tail -20 && ls -l bin/operador | awk '{print $5, "bytes"}'
./bin/operador --metrics-bind-address=0 --health-probe-bind-address=0 > /tmp/operador.log 2>&1 &
for i in $(seq 1 60); do grep -q "Starting workers" /tmp/operador.log && break; sleep 2; done
grep -q "Starting workers" /tmp/operador.log || { echo "EL OPERADOR NO ARRANCÓ:"; tail -20 /tmp/operador.log; }
grep -oE '"msg":"Starting workers"[^}]*' /tmp/operador.log | head -1
cat > /tmp/pagina.yaml <<'EOF'
apiVersion: web.golandia.dev/v1
kind: PaginaWeb
metadata:
  name: hola
spec:
  replicas: 2
  mensaje: Hola desde un operador en Go
EOF
echo "\$ cat pagina.yaml"
cat /tmp/pagina.yaml
echo "\$ kubectl apply -f pagina.yaml"
k apply -f /tmp/pagina.yaml
k rollout status deployment/hola --timeout=180s >/dev/null
sleep 3
echo "\$ kubectl get paginaweb,deployment,service,configmap -l app=hola -o name ; kubectl get paginaweb"
k get deployment,service,configmap hola -o name
k get paginaweb
echo "\$ kubectl run prueba --rm -i --image=busybox:1.37 --restart=Never -- wget -qO- http://hola"
kubectl run prueba --rm -i --image=busybox:1.37 --restart=Never -- wget -qO- http://hola 2>/dev/null | grep "<h1>" | head -1

echo
echo "### 3 · Cambiar la PaginaWeb: el operador ajusta el Deployment"
echo "\$ kubectl patch paginaweb hola --type merge -p '{\"spec\":{\"replicas\":4}}'"
k patch paginaweb hola --type merge -p '{"spec":{"replicas":4}}'
k rollout status deployment/hola --timeout=180s >/dev/null
sleep 3
echo "\$ kubectl get deployment hola ; kubectl get paginaweb"
k get deployment hola
k get paginaweb

echo
echo "### 4 · Borrar el Deployment a mano: el operador lo vuelve a crear"
echo "\$ kubectl delete deployment hola"
k delete deployment hola
sleep 5
echo "\$ kubectl get deployment hola"
k get deployment hola
grep -oE '"msg":"Deployment"[^}]*' /tmp/operador.log | tail -2

echo
echo "### 5 · La validación de la CRD"
echo "\$ kubectl patch paginaweb hola --type merge -p '{\"spec\":{\"replicas\":50}}'"
k patch paginaweb hola --type merge -p '{"spec":{"replicas":50}}'

echo
echo "### 6 · Borrar la PaginaWeb: se va todo lo suyo"
echo "\$ kubectl delete paginaweb hola"
k delete paginaweb hola
sleep 8
echo "\$ kubectl get deployment,service,configmap hola"
k get deployment,service,configmap hola
kill %1 2>/dev/null
