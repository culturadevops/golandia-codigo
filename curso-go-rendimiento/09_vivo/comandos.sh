#!/bin/sh
# Lección 9: pprof en un servicio vivo, y por qué no exponerlo.
# Levanta el servicio, le mete carga y le saca un perfil de CPU sin pararlo.
cd "$(dirname "$0")/.." && . ./comun.sh

go build -o "$T/svc" ./09_vivo
go build -o "$T/carga" ./10_carga
IP=$(hostname -i | cut -d' ' -f1) # la IP del contenedor: lo que vería otra máquina

esperar() { for _ in $(seq 50); do curl -s -o /dev/null "$1" && return; sleep 0.1; done; }
codigo() { curl -s -o /dev/null -w '%{http_code}' "$1" || true; }

echo '## 1. Expuesto: pprof en el puerto público'
"$T/svc" -modo expuesto -token sk_demo_no_es_real 2>/dev/null &
PID=$!
esperar http://127.0.0.1:8080/debug/pprof/
echo "\$ curl http://$IP:8080/debug/pprof/          -> HTTP $(codigo "http://$IP:8080/debug/pprof/")"
echo "\$ curl http://$IP:8080/debug/pprof/cmdline"
curl -s "http://$IP:8080/debug/pprof/cmdline" | tr '\0' ' '
echo
kill $PID
wait $PID 2>/dev/null

echo
echo '## 2. Privado: el servicio en 0.0.0.0:8080, pprof en 127.0.0.1:6060'
"$T/svc" -modo privado -token sk_demo_no_es_real 2>/dev/null &
PID=$!
esperar http://127.0.0.1:6060/debug/pprof/
echo "\$ curl http://$IP:8080/debug/pprof/          -> HTTP $(codigo "http://$IP:8080/debug/pprof/")"
echo "\$ curl http://$IP:6060/debug/pprof/          -> $(curl -s -o /dev/null "http://$IP:6060/debug/pprof/" && echo 'responde' || echo 'conexión rechazada')"
echo "\$ curl http://127.0.0.1:6060/debug/pprof/    -> HTTP $(codigo http://127.0.0.1:6060/debug/pprof/)"

echo
echo '## 3. Perfil de CPU de 5 s con el servicio atendiendo tráfico (versión 1)'
"$T/carga" -url http://127.0.0.1:8080/resumen -clientes 4 -duracion 14s >"$T/carga.txt" &
CARGA=$!
sleep 1
echo '$ go tool pprof -top -cum -seconds 5 http://127.0.0.1:6060/debug/pprof/profile'
go tool pprof -top -cum -nodecount=14 -seconds 5 http://127.0.0.1:6060/debug/pprof/profile 2>&1 |
	grep -vE '^(Fetching|Saved|Build ID|Main binary|File:|Time:|Please wait)'
echo
echo '$ curl http://127.0.0.1:6060/debug/pprof/goroutine?debug=1 | head -1'
curl -s 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=1' | head -1
echo
echo '$ go tool pprof -sample_index=inuse_space -top http://127.0.0.1:6060/debug/pprof/heap'
go tool pprof -sample_index=inuse_space -top -nodecount=5 http://127.0.0.1:6060/debug/pprof/heap 2>&1 |
	grep -vE '^(Fetching|Saved|Build ID|Main binary|File:|Time:)'
echo
echo "\$ curl -o \$T/traza.out 'http://127.0.0.1:6060/debug/pprof/trace?seconds=2'"
curl -s -o "$T/traza.out" 'http://127.0.0.1:6060/debug/pprof/trace?seconds=2'
echo "\$ go tool trace -pprof=sched \$T/traza.out > \$T/sched.prof && go tool pprof -top -nodecount=3 \$T/sched.prof"
go tool trace -pprof=sched "$T/traza.out" >"$T/sched.prof" && go tool pprof -top -nodecount=3 "$T/sched.prof" 2>&1 |
	grep -vE '^(Main binary)'
wait $CARGA
kill $PID
wait $PID 2>/dev/null
echo
echo '## La carga mientras se tomaba el perfil'
cat "$T/carga.txt"
