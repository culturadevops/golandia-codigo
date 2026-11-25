#!/usr/bin/env bash
# correr.sh — la medición del vídeo «Escribí Docker en Go» (golandia 681).
#
#   bash correr.sh          → salidas/dentro.txt y salidas/docker.txt
#
# mini necesita Linux y root: corre DENTRO de un contenedor --privileged (golang:1.27.2), nunca en tu máquina.
# Cinturón de seguridad para la bomba fork y el comedor de memoria: --pids-limit 300 y --memory 1g en ese contenedor.
# El código se monta de solo lectura; todo lo que se escribe va a un mktemp dentro del contenedor, que se borra con --rm.
set -u
export MSYS_NO_PATHCONV=1
cd "$(dirname "$0")" || exit 1
AQUI=$(pwd -W 2>/dev/null || pwd)
mkdir -p salidas

docker rm -f g681-mini >/dev/null 2>&1
docker run --rm --name g681-mini --privileged --pids-limit 300 --memory 1g \
  -v "$AQUI:/demo:ro" golang:1.27.2 bash /demo/dentro.sh 2>&1 | tee salidas/dentro.txt

# Lo mismo que la sección 11 y 12, pero con Docker de verdad (perfil por defecto), desde el anfitrión
{
  echo "== docker $(docker version --format '{{.Server.Version}}') · alpine:3.24.3 · perfil por defecto =="
  docker run --rm alpine:3.24.3 sh -c 'grep -E "^(CapEff|NoNewPrivs|Seccomp|Seccomp_filters):" /proc/self/status; hostname otro 2>&1 || true; mount -t tmpfs none /mnt 2>&1 || true; unshare -U true 2>&1 || true; ls /dev | tr "\n" " "; echo'
  echo
  echo "== el mismo unshare -U con --security-opt seccomp=unconfined =="
  docker run --rm --security-opt seccomp=unconfined alpine:3.24.3 sh -c 'grep -E "^Seccomp_filters:" /proc/self/status; unshare -U true && echo "unshare -U: permitido"'
  echo
  echo "== docker run --memory 64m --pids-limit 20: los mismos archivos de cgroup =="
  docker run --rm --memory 64m --pids-limit 20 alpine:3.24.3 sh -c 'for f in memory.max memory.swap.max pids.max; do echo "$f=$(cat /sys/fs/cgroup/$f)"; done'
  echo "-- con --memory-swap 64m:"
  docker run --rm --memory 64m --memory-swap 64m alpine:3.24.3 sh -c 'for f in memory.max memory.swap.max; do echo "$f=$(cat /sys/fs/cgroup/$f)"; done'
  echo
  echo "== docker image inspect alpine:3.24.3: capas =="
  docker image inspect alpine:3.24.3 --format '{{len .RootFS.Layers}} capa(s): {{.RootFS.Layers}}'
  echo
  echo "== escribir en un contenedor y leer desde otro =="
  docker run --rm alpine:3.24.3 sh -c 'echo hola > /escrito-por-el-primero && echo "escrito"'
  docker run --rm alpine:3.24.3 cat /escrito-por-el-primero 2>&1 || true
  echo
  echo "== red por defecto =="
  docker run --rm alpine:3.24.3 sh -c 'ip -o -4 addr | cut -c1-60; ip route | head -1; cat /etc/resolv.conf | grep -v "^#" | head -3'
  echo
  echo "== arranque: 3 rondas de 10 veces docker run --rm alpine:3.24.3 true (desde Windows, Docker Desktop) =="
  for ronda in 1 2 3; do
    inicio=$(date +%s%N)
    for i in $(seq 10); do docker run --rm alpine:3.24.3 true; done
    echo "ronda $ronda: $(( ($(date +%s%N) - inicio) / 10000000 )) ms por contenedor"
  done
  echo "carga de la máquina al terminar (otros agentes trabajando): $(docker ps -q | wc -l) contenedores corriendo"
} 2>&1 | tee salidas/docker.txt
