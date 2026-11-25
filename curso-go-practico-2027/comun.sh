# Lo cargan los demo.sh, que corren DENTRO del contenedor golang:1.27.2 (los lanza ejecutar_todo.py).
#   preparar parteN      copia el proyecto de esa parte a /work/tareas y entra
#   x 'comando'          imprime «$ comando» y lo ejecuta; la salida es la real
#   arrancar             compila la API y la deja corriendo en segundo plano (log en /tmp/api.log)
#   parar                le manda SIGTERM y espera a que termine
#   ver archivo          muestra un archivo del proyecto
export GOFLAGS=-buildvcs=false GOTOOLCHAIN=local
API_PID=""
preparar() {
	rm -rf /work/tareas; mkdir -p /work
	cp -r "/src/$1/tareas" /work/tareas
	cd /work/tareas || exit 1
}
x() {
	echo "\$ $1"
	eval "$1" 2>&1 | tr -d '\r'
	local rc=${PIPESTATUS[0]}
	[ "$rc" -ne 0 ] && echo "[código de salida: $rc]"
	return 0
}
esperar_api() {
	for _ in $(seq 1 100); do
		(: < "/dev/tcp/127.0.0.1/${TAREAS_PUERTO:-8080}") 2>/dev/null && return 0
		sleep 0.1
	done
	echo "[la API no respondió]"; return 1
}
arrancar() {
	go build -o tareas . || return 1
	echo "\$ ./tareas &"
	./tareas >> /tmp/api.log 2>&1 &
	API_PID=$!
	esperar_api
}
parar() {
	echo "\$ kill -TERM \$PID"
	kill -TERM "$API_PID" 2>/dev/null
	wait "$API_PID"
	echo "[la API terminó con código $?]"
}
log_api() {
	echo "\$ cat api.log"
	cat /tmp/api.log
	: > /tmp/api.log
}
ver() { echo "--- $1"; cat "$1"; echo "---"; }
