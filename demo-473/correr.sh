#!/usr/bin/env bash
# Demo del vídeo 473 «Proyecto de IA 1 en Go» — se EJECUTA y su salida va a SALIDA.md.
# Un servicio en Go (go.mod pide go 1.26.5, usa github.com/google/uuid) sin Dockerfile, y dockerfile-ia (Go, solo librería
# estándar): un LLM local en Ollama escribe el Dockerfile, docker build lo prueba, los errores vuelven al modelo, y al final se
# comprueba GET /salud. Copia en un directorio temporal; borra las imágenes ia-demo:* al terminar.
set -u
AQUI="$(cd "$(dirname "$0")" && pwd)"
T=$(mktemp -d)
cp -r "$AQUI/app" "$T/app"
(cd "$AQUI/dockerfile-ia" && go build -o "$T/dockerfile-ia.exe" .)
echo '$ go version'; go version | cut -d' ' -f3
echo '$ ollama --version'; ollama --version 2>&1 | tail -1
echo
echo '$ ls app && cat app/go.mod'; ls "$T/app"; cat "$T/app/go.mod"
echo
for caso in "qwen2.5-coder:7b cola" "qwen2.5-coder:7b relevantes" "qwen2.5-coder:1.5b relevantes"; do
  set -- $caso
  echo "\$ ./dockerfile-ia -modelo $1 -errores $2 app"
  rm -f "$T/app/Dockerfile"
  (cd "$T" && ./dockerfile-ia.exe -modelo "$1" -errores "$2" app; echo "(código de salida: $?)")
  echo
done
echo '$ # la versión del 7B, revisada por una persona (Dockerfile.revisado)'
cp "$AQUI/Dockerfile.revisado" "$T/app/Dockerfile"
echo '$ docker build -t ia-demo:revisado app && docker run -d -p 18080:8080 ia-demo:revisado && curl /salud'
(cd "$T/app" && docker build -q -t ia-demo:revisado . >/dev/null 2>&1) && echo "  build: ok"
docker rm -f ia-demo >/dev/null 2>&1
docker run -d --name ia-demo -p 18080:8080 ia-demo:revisado >/dev/null
for i in $(seq 1 20); do r=$(curl -s -m 2 http://127.0.0.1:18080/salud) && [ -n "$r" ] && break; sleep 1; done
echo "  GET /salud -> $r"
echo "  usuario de la imagen: $(docker image inspect -f '{{.Config.User}}' ia-demo:revisado)"
echo "  ¿hay shell en la imagen?: $(docker exec ia-demo sh -c true >/dev/null 2>&1 && echo sí || echo no)"
echo "  imagen: $(( $(docker image inspect -f '{{.Size}}' ia-demo:revisado) / 1000000 )) MB"
echo
docker rm -f ia-demo >/dev/null 2>&1
docker images --format '{{.Repository}}:{{.Tag}}' | grep '^ia-demo:' | xargs -r docker rmi -f >/dev/null 2>&1
cd / && rm -rf "$T"
