#!/bin/bash
# Levanta un k3s, construye y corre el operador PaginaWeb, y deja la salida literal en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-operador/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
docker rm -f k3s-op >/dev/null 2>&1
docker network create operador >/dev/null 2>&1
docker run -d --name k3s-op --network operador --network-alias k3s --privileged rancher/k3s:v1.37.1-k3s1 server \
  --disable traefik --disable metrics-server --tls-san k3s >/dev/null
until docker exec k3s-op kubectl get nodes 2>/dev/null | grep -q " Ready"; do sleep 2; done
until docker exec k3s-op kubectl get serviceaccount default >/dev/null 2>&1; do sleep 2; done
docker exec k3s-op cat /etc/rancher/k3s/k3s.yaml | sed 's#https://127.0.0.1:6443#https://k3s:6443#' > _kubeconfig
{
  echo "# SALIDA — Un operador de Kubernetes en Go (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.26 + kubebuilder v4.16.0 + kubectl v1.37.1, contra k3s (rancher/k3s:v1.37.1-k3s1)."
  echo
  echo '```'
  docker run --rm --network operador -v "$AQUI:/trabajo" golang:1.26 bash -c "go version; bash /trabajo/correr.sh" 2>&1
  echo '```'
} > SALIDA.md
docker rm -f k3s-op >/dev/null 2>&1
rm -f _kubeconfig
cat SALIDA.md
