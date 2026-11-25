source /work/repo/comun.sh
servicio > /tmp/servicio.log 2>&1 & S=$!; esperar_puerto 8090
for i in 1 2 3; do curl -s http://127.0.0.1:8090/health > /dev/null; done
curl -s "http://127.0.0.1:8090/lento?ms=10" > /dev/null
echo '$ curl -s localhost:8090/health'; curl -s http://127.0.0.1:8090/health; echo
echo '$ curl -s localhost:8090/metrics'; curl -s http://127.0.0.1:8090/metrics
kill $S; wait $S; true
