#!/usr/bin/env bash
# demo.sh — la corrida entera del vídeo 224.
#
# El servidor (kafkita, Go, sin dependencias) en un contenedor; el cliente OFICIAL de Apache Kafka
# en otro; misma red de Docker. Todo lo que se ve aquí lo dice el cliente oficial, no yo.
#
# ⚠️ Se da MARGEN antes de lanzar cada cliente: el DNS de Docker tarda un momento en resolver el
#    nombre del contenedor recién creado, y sin espera el vídeo enseñaría un error que no es el
#    del tema (pasó en el primer sondeo).
set -u
D=$(cd "$(dirname "$0")" && pwd -W)
R=kfred
KAFKA=apache/kafka:latest

arranca() { # $1 = qué anunciamos en Metadata
  docker rm -f kafkita >/dev/null 2>&1
  docker run -d --name kafkita --network $R -v "$D:/src" -w /src -e ANUNCIO="$1" \
    golang:alpine sh -c 'cp -r /src /tmp/w && cd /tmp/w && go run .' >/dev/null
  sleep 14
}
cliente() { docker run --rm -i --network $R $KAFKA "$@"; }

docker network create $R >/dev/null 2>&1

echo "== 1 · LO PRIMERO QUE MANDA UN CLIENTE DE KAFKA =="
echo "   (sonda: un socket que NO contesta nada, solo mira lo que llega)"
docker rm -f sonda >/dev/null 2>&1
docker run -d --name sonda --network $R python:3.12-alpine python -c "
import socket,struct
s=socket.socket(); s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('0.0.0.0',9092)); s.listen(5)
NOM={18:'ApiVersions',3:'Metadata',0:'Produce',1:'Fetch'}
for n in range(3):
    c,_=s.accept(); d=c.recv(4096)
    if len(d)<12: c.close(); continue
    largo,=struct.unpack('>i',d[:4]); api,ver,corr=struct.unpack('>hhi',d[4:12])
    print(f'  peticion {n}: largo={largo} api={api} ({NOM.get(api,\"?\")}) version={ver}',flush=True)
    c.close()
" >/dev/null
sleep 4
cliente /opt/kafka/bin/kafka-topics.sh --bootstrap-server sonda:9092 --list >/dev/null 2>&1 || true
docker logs sonda 2>&1 | head -3
docker rm -f sonda >/dev/null 2>&1
echo "   ^ 49 bytes preguntando «¿que sabes hacer?». Y hasta que no contestas, reintenta"
echo "     en bucle SIN decir que le pasa."

echo
echo "== 2 · LA NEGOCIACION: yo digo qué versiones hablo, y el cliente baja =="
arranca kafkita
cliente /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafkita:9092 --list 2>&1 | tail -2
echo "   ^ el cliente OFICIAL listando temas de 400 lineas de Go."
echo "   y en el registro del servidor se ve la negociacion:"
docker logs kafkita 2>&1 | grep -E 'ApiVersions|Metadata' | head -2 | sed 's/^/     /'
echo "     ^ pide ApiVersions en la v4 y luego manda Metadata en la v3: la mas alta que le dije."

echo
echo "== 3 · PRODUCIR con el productor OFICIAL =="
printf 'hola\nque tal\nadios\n' | cliente /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server kafkita:9092 --topic pruebas 2>&1 | grep -iE 'error|exception' | head -2
echo "   (sin lineas de error arriba = entro bien)"
docker logs kafkita 2>&1 | grep 'guardado' | tail -1 | sed 's/^/   /'
echo "   ^ TRES mensajes en UN solo lote de 98 bytes. La unidad no es el mensaje: es el lote."

echo
echo "== 4 · CONSUMIR con el consumidor OFICIAL =="
cliente /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafkita:9092 \
  --topic pruebas --partition 0 --offset earliest --timeout-ms 8000 2>&1 | grep -v ERROR | tail -5
echo -n "   peticiones de Fetch que le costo: "
docker logs kafkita 2>&1 | grep -c 'Fetch'
echo "   ^ con la espera larga puesta. SIN ella fueron 316.000 en diez segundos."

echo
echo "== 5 · 🥇 LA TRAMPA DE METADATA: el 90% de los «no me conecto a Kafka» =="
echo "   ahora arranco el MISMO servidor pero anunciandome con un nombre que el"
echo "   cliente no puede resolver. Todo lo demas es identico."
arranca no-existe-este-host
cliente /opt/kafka/bin/kafka-console-producer.sh --bootstrap-server kafkita:9092 \
  --topic pruebas </dev/null 2>&1 | grep -iE 'unknown|resolve|disconnect|timed' | head -3
echo "   ^ el cliente se conecta a kafkita PERFECTAMENTE, recibe la respuesta, y a partir de"
echo "     ahi se va al nombre que le dije y se pierde. El puerto estaba abierto todo el rato."
docker logs kafkita 2>&1 | grep -E 'ApiVersions|Metadata' | head -2 | sed 's/^/     /'
echo "     ^ y en el servidor NO hay ni un error: las dos peticiones entraron bien."

docker rm -f kafkita >/dev/null 2>&1
docker network rm $R >/dev/null 2>&1
echo
echo "== limpio =="
