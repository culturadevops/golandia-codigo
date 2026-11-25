source /work/repo/comun.sh
servicio > /tmp/servicio.log 2>&1 & S=$!; esperar_puerto 8090
curl -s "http://127.0.0.1:8090/lento?ms=50" > /dev/null
kill $S; wait $S; true
echo '$ cat servicio.log (los timestamps, recortados)'
sed -E 's/"timestamp":"[^"]*",//' /tmp/servicio.log
echo '$ grep -c lenta servicio.log'; grep -c lenta /tmp/servicio.log
