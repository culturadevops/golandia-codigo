#!/bin/bash
# Demo del vídeo 561 (golandia): respaldo de Postgres a S3 con Go. Corre en Git Bash; lo crea y lo borra TODO
# con nombre respaldo-demo-* y la etiqueta demo=respaldo-pg-s3. No toca ningún otro contenedor ni volumen.
#   bash canales/golandia/recursos/pruebas-respaldo-postgres-s3/correr.sh > SALIDA-crudo.txt
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
RED=respaldo-demo-red
ET="--label demo=respaldo-pg-s3"

limpiar() {
  docker rm -f respaldo-demo-origen respaldo-demo-destino respaldo-demo-moto >/dev/null 2>&1
  docker network rm $RED >/dev/null 2>&1
}
limpiar
docker network create $ET $RED >/dev/null
docker run -d $ET --name respaldo-demo-origen --network $RED -e POSTGRES_PASSWORD=demo postgres:18 >/dev/null
docker run -d $ET --name respaldo-demo-destino --network $RED -e POSTGRES_PASSWORD=demo postgres:18 >/dev/null
docker run -d $ET --name respaldo-demo-moto --network $RED motoserver/moto:latest >/dev/null
for c in respaldo-demo-origen respaldo-demo-destino; do
  until docker exec $c pg_isready -U postgres -q 2>/dev/null; do sleep 1; done
done
sleep 3
docker build -q -t respaldo-demo . >/dev/null || exit 1

# job: el binario de Go en la imagen de la demo, con el entorno de pg_dump y de S3 (moto)
job() {
  docker run --rm $ET --network $RED \
    -e PGHOST=respaldo-demo-origen -e PGHOST_DESTINO=respaldo-demo-destino -e PGUSER=postgres -e PGPASSWORD=demo \
    -e AWS_ENDPOINT_URL=http://respaldo-demo-moto:5000 -e AWS_REGION=us-east-1 \
    -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e BUCKET=respaldos-demo \
    respaldo-demo sh -c "$1"
}
paso() { echo; echo "\$ $1"; job "$1" 2>&1; }

echo "== versiones"
job 'pg_dump --version; psql --version' 2>&1
docker exec respaldo-demo-moto python -c "import moto; print('moto', moto.__version__)" 2>&1
docker run --rm golang:1.27.2 go version
grep -E 'aws-sdk-go-v2 |transfermanager|service/s3 ' respaldo/go.mod | sed 's/^\t//'
echo "cpus de Docker: $(docker info --format '{{.NCPU}}') · memoria: $(docker info --format '{{.MemTotal}}' | awk '{printf "%.0f GB", $1/1e9}')"

echo; echo "== datos de origen (semilla.sql)"
t0=$(date +%s)
docker exec -i respaldo-demo-origen psql -q -U postgres < semilla.sql >/dev/null
echo "semilla cargada en $(( $(date +%s) - t0 )) s"
docker exec respaldo-demo-origen psql -At -U postgres -d tienda -c "SELECT 'tamaño de la base tienda: ' || pg_size_pretty(pg_database_size('tienda'))"

paso 'respaldo bucket'

echo; echo "== 1. tres formas de comprimir (mismo origen; 3 rondas, la misma clave se sobrescribe)"
for ronda in 1 2 3; do
  echo; echo "-- ronda $ronda"
  for m in "0300 fc" "0301 fc+gzip" "0302 z0+gzip"; do
    set -- $m
    paso "respaldo respaldar -fecha 2026-10-21T$1 -modo $2"
  done
done

echo; echo "== 2. el ingenuo: todo el pg_dump a memoria y después PutObject"
paso 'respaldo ingenuo -fecha 2026-10-21T0303'

echo; echo "== 3. pg_dump muere a la mitad (se corta su conexión a los 2 s)"
CORTE='(sleep 2; psql -qAt -d tienda -c "SELECT count(pg_terminate_backend(pid)) || '"'"' conexión de pg_dump cortada'"'"' FROM pg_stat_activity WHERE application_name = '"'"'pg_dump'"'"'") & '
paso "${CORTE}respaldo sin-wait -fecha 2026-10-21T0400; wait"
paso "${CORTE}respaldo respaldar -fecha 2026-10-21T0401 -modo fc+gzip; wait"
paso 'respaldo listar'
paso 'respaldo restaurar -clave respaldos/tienda/2026-10-21T0400.dump.gz'

echo; echo "== 4. retención: 12 respaldos de inventario repartidos en 10 días"
for f in 2026-10-12T0300 2026-10-13T0300 2026-10-14T0300 2026-10-15T0300 2026-10-16T0300 2026-10-17T0300 \
         2026-10-18T0300 2026-10-18T1500 2026-10-19T0300 2026-10-20T0300 2026-10-21T0300 2026-10-21T1500; do
  job "respaldo respaldar -base inventario -fecha $f -modo fc" >/dev/null 2>&1
done
paso 'respaldo listar -base inventario'
paso 'respaldo retener -base inventario -ultimos 3 -dias 7'
paso 'respaldo retener -base inventario -ultimos 3 -dias 7 -borrar'
paso 'respaldo listar -base inventario'

echo; echo "== 5. el doble de datos: ¿cuánta memoria con streaming y cuánta con el ingenuo?"
docker exec respaldo-demo-origen psql -q -U postgres -d tienda -c "INSERT INTO pedidos SELECT g, 1 + (g % 50000), round((random() * 500)::numeric, 2), timestamptz '2026-01-01' + g * interval '7 seconds', md5(g::text) || md5((g * 7)::text) FROM generate_series(2000001, 4000000) g" >/dev/null
docker exec respaldo-demo-origen psql -At -U postgres -d tienda -c "SELECT 'pedidos: ' || count(*) || ' · tamaño de la base: ' || pg_size_pretty(pg_database_size('tienda')) FROM pedidos"
paso 'respaldo respaldar -fecha 2026-10-21T0300 -modo z0+gzip'
paso 'respaldo ingenuo -fecha 2026-10-21T0303'

echo; echo "== 6. el respaldo solo vale si restaura: a una base nueva, y se compara"
paso 'respaldo restaurar -clave respaldos/tienda/2026-10-21T0300.dump.gz'
paso 'respaldo verificar'

limpiar
echo; echo "== limpieza: contenedores y red respaldo-demo-* borrados"
