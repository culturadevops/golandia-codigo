#!/usr/bin/env bash
# demo.sh — la medición del vídeo «Escribí un escáner de puertos en Go».
#
# ⛔ ÉTICO: todo corre contra un objetivo que levanto yo, en mi propia red de contenedores, con
#    puertos que yo mismo abro. No se toca ninguna máquina ajena.
#
# ⛔ HAY UNA VERDAD CONTRA LA QUE COMPARAR: el objetivo abre 15 puertos conocidos. Así que no se
#    mide solo el tiempo: se mide si el escáner LOS ENCUENTRA. Un escáner rápido que se deja
#    puertos abiertos no es rápido, es inútil.
#
# ⚠️ Los errores NO van a /dev/null: se clasifican, porque distinguir «rechazado» (respuesta
#    normal de un puerto cerrado) de «TIMEOUT» o «SIN-DESCRIPTORES» es la mitad del hallazgo.
set -u
export MSYS_NO_PATHCONV=1
RED=escanred

limpiar() {
  docker rm -f objetivo escaner >/dev/null 2>&1
  docker network rm $RED >/dev/null 2>&1
  return 0
}

limpiar
docker network create $RED >/dev/null

echo "== 1 · EL OBJETIVO: puertos conocidos, para tener una verdad =="
docker run -d --name objetivo --network $RED --network-alias objetivo \
  -v "$PWD/objetivo:/app" -w /app golang:1.26-bookworm \
  bash -c "go build -o /tmp/o . && /tmp/o" >/dev/null
for i in $(seq 1 60); do
  docker logs objetivo 2>&1 | grep -q "OBJETIVO LISTO" && break
  sleep 2
done
docker logs objetivo 2>&1 | sed 's/^/   /'
ESPERADOS=$(docker logs objetivo 2>&1 | grep -oE "^OBJETIVO LISTO: [0-9]+" | grep -oE "[0-9]+")
echo "   >>> la verdad: $ESPERADOS puertos abiertos"

echo
echo "== 2 · COMPILO EL ESCÁNER =="
docker run -d --name escaner --network $RED -v "$PWD:/app" -w /app \
  golang:1.26-bookworm sleep 3600 >/dev/null
docker exec escaner bash -c "go build -o /tmp/escaner . && echo '   ok · '\$(stat -c %s /tmp/escaner)' bytes'"
docker exec escaner bash -c "apt-get update -qq >/dev/null 2>&1; apt-get install -y -qq nmap >/dev/null 2>&1; echo -n '   nmap: '; nmap --version | head -1"
echo -n "   descriptores de fichero del contenedor: "
docker exec escaner bash -c "ulimit -n"

echo
echo "== 3 · 🥇 LA MISMA BÚSQUEDA, CON DISTINTA CONCURRENCIA =="
echo "   (65.535 puertos cada vez · timeout 300 ms · la verdad son $ESPERADOS abiertos)"
echo
for C in 100 500 2000 10000 65535; do
  docker exec escaner /tmp/escaner objetivo $C 300 2>&1 | sed 's/^/   /'
  echo
done

echo "== 4 · ⛔ ¿Y SI LE DOY MÁS TIEMPO A LA VERSIÓN SALVAJE? =="
echo "   (misma concurrencia de 65.535, pero timeout de 3 segundos)"
docker exec escaner /tmp/escaner objetivo 65535 3000 2>&1 | sed 's/^/   /'

echo
echo "== 5 · CONTRA NMAP, sobre el mismo objetivo =="
echo -n "   nmap: "
docker exec escaner bash -c "
  s=\$(date +%s%N)
  r=\$(nmap -p1-65535 -T4 --open objetivo 2>/dev/null | grep -cE '^[0-9]+/tcp')
  e=\$(date +%s%N)
  echo \"\$r abiertos en \$(( (e-s)/1000000 )) ms\"
"

echo
echo "== 6 · SECUENCIAL, para tener el suelo (solo 2.000 puertos: si no, no acaba) =="
docker exec escaner /tmp/escaner objetivo 1 300 2000 2>&1 | sed 's/^/   /'
echo "   ⚠️ ese es de 2.000 puertos, no de 65.535. Para 65.535 serían ~30x más."

limpiar
echo
echo "== fin =="
