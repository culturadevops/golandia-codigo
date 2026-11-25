#!/usr/bin/env bash
# Parte 4 (dentro del contenedor): configuración, logs con slog y apagado ordenado.
# La parte de Docker (imágenes y tamaños) la hace ejecutar_todo.py desde fuera.
source /src/comun.sh
preparar parte4
go build -o tareas . || exit 1
URL_BUENA="$DATABASE_URL"

echo "@@PASO config"
x 'env -u DATABASE_URL ./tareas'
x 'env -u DATABASE_URL TAREAS_PUERTO=ochenta TAREAS_TIMEOUT_BD=2 ./tareas'
x 'go test -run LeerConfig -v .'

echo "@@PASO logs"
export TAREAS_TIMEOUT_BD=5s
arrancar
x "curl -s -X POST localhost:8080/tareas -d '{\"titulo\":\"comprar pan\"}'"
x "curl -s -o /dev/null -w 'HTTP %{http_code}\n' localhost:8080/tareas/99"
x "curl -s -o /dev/null -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{\"titulo\":\"\"}'"
log_api
parar

echo "@@PASO sin-apagado"
# El binario de la parte 2: log.Fatal(http.ListenAndServe(...)), sin manejar señales.
rm -rf /tmp/p2 && cp -r /src/parte2/tareas /tmp/p2 && (cd /tmp/p2 && go build -o /tmp/tareas-parte2 .) || exit 1
go build -o bloquear ./herramientas/bloquear || exit 1
echo '$ ./tareas-parte2 &'
/tmp/tareas-parte2 > /tmp/p2.log 2>&1 &
P2=$!
esperar_api
echo '$ ./bloquear 3 &'
./bloquear 3 > /tmp/bloquear.log 2>&1 &
sleep 0.5
echo "\$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas &"
curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas > /tmp/curl.log 2>&1 &
CURL=$!
sleep 0.5
echo '$ kill -TERM $PID'
kill -TERM $P2
wait $CURL
echo "[curl terminó con código $?]"
cat /tmp/curl.log
wait $P2
echo "[la API terminó con código $?]"
wait
cat /tmp/bloquear.log

echo "@@PASO apagado"
arrancar
: > /tmp/api.log
echo '$ ./bloquear 3 &'
./bloquear 3 > /tmp/bloquear.log 2>&1 &
sleep 0.5
echo "\$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas &"
curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas > /tmp/curl.log 2>&1 &
CURL=$!
sleep 0.5
echo '$ kill -TERM $PID'
kill -TERM "$API_PID"
sleep 0.2
x "curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas"
wait $CURL
cat /tmp/curl.log
wait "$API_PID"
echo "[la API terminó con código $?]"
wait
cat /tmp/bloquear.log
log_api
