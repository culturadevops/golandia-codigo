source /work/repo/comun.sh
prueba() {
	echo "\$ servicio $1   # y a 1 s de una petición de 3 s, kill -TERM"
	servicio $1 > /tmp/s.log 2>&1 & S=$!; esperar_puerto 8090
	T0=$(date +%s.%N)
	(curl -s -S "http://127.0.0.1:8090/lento?ms=3000" 2>&1; echo "[curl: código $?]") > /tmp/c.log &
	C=$!
	sleep 1; kill -TERM $S
	wait $C; wait $S; CS=$?
	T1=$(date +%s.%N)
	echo "  respuesta del cliente:"; sed 's/^/    /' /tmp/c.log
	echo "  el servicio salió con código $CS a los $(awk "BEGIN{printf \"%.1f\", $T1-$T0}") s"
	echo "  últimas líneas del log:"; sed -E 's/"timestamp":"[^"]*",//' /tmp/s.log | tail -2 | sed 's/^/    /'
}
prueba --sin-apagado
prueba ""
